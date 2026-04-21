package upstream

import (
	"net/http"
	"strings"
	"time"
)

var client = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	},
}

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"

func newRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func normalizeAPIURL(apiURL string) string {
	u := strings.TrimRight(apiURL, "/")
	if !strings.HasSuffix(u, "/at/json") {
		u += "/at/json"
	}
	return u
}
