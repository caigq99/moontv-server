package upstream

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"
)

type SearchResult struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Poster         string   `json:"poster"`
	Episodes       []string `json:"episodes"`
	EpisodesTitles []string `json:"episodes_titles"`
	Source         string   `json:"source"`
	SourceName    string   `json:"source_name"`
	Class          string   `json:"class,omitempty"`
	Year           string   `json:"year"`
	Desc           string   `json:"desc,omitempty"`
	TypeName       string   `json:"type_name,omitempty"`
}

type apiResponse struct {
	List      []apiItem `json:"list"`
	PageCount int       `json:"pagecount"`
}

type apiItem struct {
	VodID       json.Number `json:"vod_id"`
	VodName     string      `json:"vod_name"`
	VodPic      string      `json:"vod_pic"`
	VodPlayURL  string      `json:"vod_play_url"`
	VodClass    string      `json:"vod_class"`
	VodYear     string      `json:"vod_year"`
	VodContent  string      `json:"vod_content"`
	TypeName    string      `json:"type_name"`
}

var yearPattern = regexp.MustCompile(`\d{4}`)
var htmlTagPattern = regexp.MustCompile(`<[^>]+>`)

func SearchFromAPI(apiURL, sourceKey, sourceName, query string, maxPages int) ([]SearchResult, error) {
	base := normalizeAPIURL(apiURL)
	firstURL := fmt.Sprintf("%s?ac=videolist&wd=%s", base, url.QueryEscape(query))
	results, pageCount, err := doSearch(firstURL, sourceKey, sourceName)
	if err != nil {
		return nil, err
	}

	pages := pageCount
	if pages > maxPages {
		pages = maxPages
	}

	for page := 2; page <= pages; page++ {
		pageURL := fmt.Sprintf("%s?ac=videolist&wd=%s&pg=%d", base, url.QueryEscape(query), page)
		pageResults, _, err := doSearch(pageURL, sourceKey, sourceName)
		if err != nil {
			continue
		}
		results = append(results, pageResults...)
	}

	return results, nil
}

func doSearch(apiURL, sourceKey, sourceName string) ([]SearchResult, int, error) {
	req, err := newRequest(apiURL)
	if err != nil {
		return nil, 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		io.Copy(io.Discard, resp.Body)
		return nil, 0, fmt.Errorf("status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	var data apiResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, 0, err
	}

	var results []SearchResult
	for _, item := range data.List {
		episodes, titles := parsePlayURL(item.VodPlayURL)
		if len(episodes) == 0 {
			continue
		}

		year := "unknown"
		if m := yearPattern.FindString(item.VodYear); m != "" {
			year = m
		}

		results = append(results, SearchResult{
			ID:             item.VodID.String(),
			Title:          strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(item.VodName, " ")),
			Poster:         item.VodPic,
			Episodes:       episodes,
			EpisodesTitles: titles,
			Source:         sourceKey,
			SourceName:    sourceName,
			Class:          item.VodClass,
			Year:           year,
			Desc:           cleanHTML(item.VodContent),
			TypeName:       item.TypeName,
		})
	}

	return results, data.PageCount, nil
}

func parsePlayURL(vodPlayURL string) ([]string, []string) {
	if vodPlayURL == "" {
		return nil, nil
	}

	var bestEpisodes, bestTitles []string

	groups := strings.Split(vodPlayURL, "$$$")
	for _, group := range groups {
		var episodes, titles []string
		items := strings.Split(group, "#")
		for _, item := range items {
			parts := strings.SplitN(item, "$", 2)
			if len(parts) == 2 && strings.HasSuffix(parts[1], ".m3u8") {
				titles = append(titles, parts[0])
				episodes = append(episodes, parts[1])
			}
		}
		if len(episodes) > len(bestEpisodes) {
			bestEpisodes = episodes
			bestTitles = titles
		}
	}

	return bestEpisodes, bestTitles
}

func SearchPage(apiURL, sourceKey, sourceName, query string, page int) ([]SearchResult, int, error) {
	base := normalizeAPIURL(apiURL)
	pageURL := fmt.Sprintf("%s?ac=videolist&wd=%s&pg=%d", base, url.QueryEscape(query), page)
	return doSearch(pageURL, sourceKey, sourceName)
}

func cleanHTML(text string) string {
	if text == "" {
		return ""
	}
	cleaned := htmlTagPattern.ReplaceAllString(text, "\n")
	cleaned = regexp.MustCompile(`\n+`).ReplaceAllString(cleaned, "\n")
	cleaned = regexp.MustCompile(`[ \t]+`).ReplaceAllString(cleaned, " ")
	return strings.TrimSpace(cleaned)
}
