package upstream

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

var m3u8Pattern = regexp.MustCompile(`https?://[^"'\s]+?\.m3u8`)

func GetDetail(ctx context.Context, apiURL, detailURL, sourceKey, sourceName, id string) (*SearchResult, error) {
	if detailURL != "" {
		return getDetailFromHTML(ctx, detailURL, sourceKey, sourceName, id)
	}
	return getDetailFromAPI(ctx, apiURL, sourceKey, sourceName, id)
}

func getDetailFromAPI(ctx context.Context, apiURL, sourceKey, sourceName, id string) (*SearchResult, error) {
	base := normalizeAPIURL(apiURL)
	reqURL := fmt.Sprintf("%s?ac=videolist&ids=%s", base, id)
	req, err := newRequest(ctx, reqURL)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	body, err := readBodyLimited(resp.Body)
	if err != nil {
		return nil, err
	}

	var data apiResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	if len(data.List) == 0 {
		return nil, fmt.Errorf("no detail found")
	}

	item := data.List[0]
	episodes, titles := parsePlayURL(item.VodPlayURL)

	if len(episodes) == 0 && item.VodContent != "" {
		matches := m3u8Pattern.FindAllString(item.VodContent, -1)
		for _, m := range matches {
			episodes = append(episodes, strings.TrimPrefix(m, "$"))
		}
	}

	year := "unknown"
	if m := yearPattern.FindString(item.VodYear); m != "" {
		year = m
	}

	return &SearchResult{
		ID:             id,
		Title:          item.VodName,
		Poster:         item.VodPic,
		Episodes:       episodes,
		EpisodesTitles: titles,
		Source:         sourceKey,
		SourceName:     sourceName,
		Class:          item.VodClass,
		Year:           year,
		Desc:           cleanHTML(item.VodContent),
		TypeName:       item.TypeName,
	}, nil
}

func getDetailFromHTML(ctx context.Context, detailURL, sourceKey, sourceName, id string) (*SearchResult, error) {
	reqURL := fmt.Sprintf("%s/index.php/vod/detail/id/%s.html", detailURL, id)
	req, err := newRequest(ctx, reqURL)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	body, err := readBodyLimited(resp.Body)
	if err != nil {
		return nil, err
	}

	html := string(body)

	pattern := regexp.MustCompile(`\$(https?://[^"'\s]+?\.m3u8)`)
	matches := pattern.FindAllStringSubmatch(html, -1)
	seen := make(map[string]bool)
	var episodes []string
	for _, m := range matches {
		link := m[1]
		if idx := strings.Index(link, "("); idx > 0 {
			link = link[:idx]
		}
		if !seen[link] {
			seen[link] = true
			episodes = append(episodes, link)
		}
	}

	titles := make([]string, len(episodes))
	for i := range episodes {
		titles[i] = fmt.Sprintf("%d", i+1)
	}

	titleMatch := regexp.MustCompile(`<h1[^>]*>([^<]+)</h1>`).FindStringSubmatch(html)
	title := ""
	if len(titleMatch) > 1 {
		title = strings.TrimSpace(titleMatch[1])
	}

	coverMatch := regexp.MustCompile(`(https?://[^"'\s]+?\.jpg)`).FindString(html)

	yearMatch := regexp.MustCompile(`>(\d{4})<`).FindStringSubmatch(html)
	year := "unknown"
	if len(yearMatch) > 1 {
		year = yearMatch[1]
	}

	return &SearchResult{
		ID:             id,
		Title:          title,
		Poster:         coverMatch,
		Episodes:       episodes,
		EpisodesTitles: titles,
		Source:         sourceKey,
		SourceName:     sourceName,
		Year:           year,
	}, nil
}
