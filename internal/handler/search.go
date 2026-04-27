package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/moontv/server/internal/cache"
	"github.com/moontv/server/internal/model"
	"github.com/moontv/server/internal/repository"
	"github.com/moontv/server/pkg/filter"
	"github.com/moontv/server/pkg/response"
	"github.com/moontv/server/pkg/upstream"
)

const maxConcurrentSourceSearches = 4

type SearchHandler struct {
	Cache *cache.SearchCache
}

type sourceResult struct {
	Source     string                  `json:"source"`
	SourceName string                 `json:"source_name"`
	Results    []upstream.SearchResult `json:"results"`
	PageCount int                     `json:"page_count"`
}

type sourceErrorEvent struct {
	Source     string `json:"source"`
	SourceName string `json:"source_name"`
	ErrorType  string `json:"error_type"`
}

func (h *SearchHandler) Search(c *gin.Context) {
	userID := getUserID(c)
	query := c.Query("q")
	if query == "" {
		response.Fail(c, http.StatusBadRequest, response.ErrInvalidParam, "missing query parameter 'q'")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	yellowFilter := c.DefaultQuery("yellow_filter", "true") != "false"

	sources, err := repository.GetEnabledSourcesByUserID(userID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ErrInternal, "failed to get sources")
		return
	}

	results := h.searchSources(c.Request.Context(), userID, sources, query, page, yellowFilter)
	response.OK(c, results)
}

func (h *SearchHandler) SearchSSE(c *gin.Context) {
	ctx := c.Request.Context()
	userID := getUserID(c)
	query := c.Query("q")
	if query == "" {
		response.Fail(c, http.StatusBadRequest, response.ErrInvalidParam, "missing query parameter 'q'")
		return
	}
	yellowFilter := c.DefaultQuery("yellow_filter", "true") != "false"

	sources, err := repository.GetEnabledSourcesByUserID(userID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ErrInternal, "failed to get sources")
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	sem := make(chan struct{}, maxConcurrentSourceSearches)

	type sseEvent struct {
		event string
		data  []byte
	}

	ch := make(chan sseEvent, len(sources))
	for _, src := range sources {
		go func(src model.Source) {
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				return
			}

			results, pageCount, searchErr := h.searchOneSource(ctx, userID, src, query, 1)
			if searchErr != nil {
				data, _ := json.Marshal(sourceErrorEvent{
					Source:     src.Key,
					SourceName: src.Name,
					ErrorType:  classifySearchError(searchErr),
				})
				select {
				case ch <- sseEvent{event: "source_error", data: data}:
				case <-ctx.Done():
				}
				return
			}
			filtered := filterYellow(results, yellowFilter)
			data, _ := json.Marshal(sourceResult{
				Source:     src.Key,
				SourceName: src.Name,
				Results:    filtered,
				PageCount:  pageCount,
			})
			select {
			case ch <- sseEvent{event: "message", data: data}:
			case <-ctx.Done():
			}
		}(src)
	}

	received := 0
	for received < len(sources) {
		select {
		case evt := <-ch:
			received++
			if evt.event == "source_error" {
				fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", evt.event, evt.data)
			} else {
				fmt.Fprintf(c.Writer, "data: %s\n\n", evt.data)
			}
			c.Writer.(http.Flusher).Flush()
		case <-ctx.Done():
			return
		}
	}

	fmt.Fprintf(c.Writer, "event: done\ndata: {}\n\n")
	c.Writer.(http.Flusher).Flush()
}

func (h *SearchHandler) Detail(c *gin.Context) {
	userID := getUserID(c)
	sourceKey := c.Query("source")
	id := c.Query("id")
	if sourceKey == "" || id == "" {
		response.Fail(c, http.StatusBadRequest, response.ErrInvalidParam, "missing 'source' or 'id'")
		return
	}

	src, err := repository.GetSourceByKey(userID, sourceKey)
	if err != nil {
		src, err = repository.GetGlobalSourceByKey(sourceKey)
		if err != nil {
			response.Fail(c, http.StatusNotFound, response.ErrNotFound, "source not found")
			return
		}
	}

	result, err := upstream.GetDetail(c.Request.Context(), src.APIUrl, src.DetailUrl, src.Key, src.Name, id)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ErrInternal, "failed to get detail")
		return
	}

	response.OK(c, result)
}

func (h *SearchHandler) Suggest(c *gin.Context) {
	userID := getUserID(c)
	query := c.Query("q")
	if query == "" {
		response.Fail(c, http.StatusBadRequest, response.ErrInvalidParam, "missing query parameter 'q'")
		return
	}

	sources, err := repository.GetEnabledSourcesByUserID(userID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ErrInternal, "failed to get sources")
		return
	}

	limit := 3
	if len(sources) < limit {
		limit = len(sources)
	}

	results := h.searchSources(c.Request.Context(), userID, sources[:limit], query, 1, true)

	seen := make(map[string]bool)
	var titles []string
	for _, sr := range results {
		for _, r := range sr.Results {
			if !seen[r.Title] {
				seen[r.Title] = true
				titles = append(titles, r.Title)
			}
			if len(titles) >= 10 {
				break
			}
		}
		if len(titles) >= 10 {
			break
		}
	}

	response.OK(c, titles)
}

func (h *SearchHandler) searchSources(ctx context.Context, userID uint, sources []model.Source, query string, page int, yellowFilter bool) []sourceResult {
	var mu sync.Mutex
	var wg sync.WaitGroup
	var results []sourceResult
	sem := make(chan struct{}, maxConcurrentSourceSearches)

	for _, src := range sources {
		wg.Add(1)
		go func(src model.Source) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				return
			}

			searchResults, pageCount, err := h.searchOneSource(ctx, userID, src, query, page)
			if err != nil {
				return
			}
			filtered := filterYellow(searchResults, yellowFilter)
			mu.Lock()
			results = append(results, sourceResult{
				Source:     src.Key,
				SourceName: src.Name,
				Results:    filtered,
				PageCount:  pageCount,
			})
			mu.Unlock()
		}(src)
	}

	wg.Wait()
	return results
}

func (h *SearchHandler) searchOneSource(ctx context.Context, userID uint, src model.Source, query string, page int) ([]upstream.SearchResult, int, error) {
	if cached, pageCount, ok := h.Cache.Get(userID, src.Key, query, page); ok {
		return cached, pageCount, nil
	}

	results, pageCount, err := upstream.SearchPage(ctx, src.APIUrl, src.Key, src.Name, query, page)
	if err != nil {
		return nil, 0, err
	}

	h.Cache.Set(userID, src.Key, query, page, results, pageCount)
	return results, pageCount, nil
}

func classifySearchError(err error) string {
	if err == nil {
		return "unknown"
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return "timeout"
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "timeout"
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "private ip blocked") || strings.Contains(msg, "private hostname blocked"):
		return "private_ip_blocked"
	case strings.Contains(msg, "response too large"):
		return "response_too_large"
	case strings.Contains(msg, "invalid character") || strings.Contains(msg, "cannot unmarshal") || strings.Contains(msg, "unexpected end of json input"):
		return "parse_error"
	case strings.Contains(msg, "status ") || strings.Contains(msg, "dns lookup failed") || strings.Contains(msg, "unsupported scheme"):
		return "upstream_error"
	default:
		return "unknown"
	}
}

func filterYellow(results []upstream.SearchResult, enabled bool) []upstream.SearchResult {
	if !enabled {
		return results
	}
	var filtered []upstream.SearchResult
	for _, r := range results {
		if !filter.IsYellow(r.TypeName) {
			filtered = append(filtered, r)
		}
	}
	return filtered
}
