package upstream

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxResponseBodyBytes int64 = 2 << 20

var client = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			Resolver:  net.DefaultResolver,
		}).DialContext,
	},
	CheckRedirect: checkRedirect,
}

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"

func newRequest(ctx context.Context, rawURL string) (*http.Request, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("unsupported scheme: %s", parsed.Scheme)
	}

	if err := blockPrivateHost(ctx, parsed.Hostname()); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func readBodyLimited(r io.Reader) ([]byte, error) {
	lr := &io.LimitedReader{R: r, N: maxResponseBodyBytes + 1}
	body, err := io.ReadAll(lr)
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > maxResponseBodyBytes {
		return nil, fmt.Errorf("upstream response too large")
	}
	return body, nil
}

func checkRedirect(req *http.Request, via []*http.Request) error {
	if err := blockPrivateHost(req.Context(), req.URL.Hostname()); err != nil {
		return err
	}
	if len(via) >= 10 {
		return fmt.Errorf("too many redirects")
	}
	return nil
}

func blockPrivateHost(ctx context.Context, hostname string) error {
	if hostname == "localhost" || strings.HasSuffix(hostname, ".localhost") || strings.HasSuffix(hostname, ".local") {
		return fmt.Errorf("private hostname blocked: %s", hostname)
	}

	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", hostname)
	if err != nil {
		return fmt.Errorf("DNS lookup failed for %s: %w", hostname, err)
	}

	for _, ip := range ips {
		if isPrivateIP(ip) {
			return fmt.Errorf("private IP blocked: %s", ip)
		}
	}
	return nil
}

func isPrivateIP(ip net.IP) bool {
	for _, n := range privateNets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

var privateNets = []*net.IPNet{
	parseCIDR("10.0.0.0/8"),
	parseCIDR("172.16.0.0/12"),
	parseCIDR("192.168.0.0/16"),
	parseCIDR("127.0.0.0/8"),
	parseCIDR("169.254.0.0/16"),
	parseCIDR("::1/128"),
	parseCIDR("fc00::/7"),
	parseCIDR("fe80::/10"),
}

func parseCIDR(s string) *net.IPNet {
	_, n, _ := net.ParseCIDR(s)
	return n
}

func normalizeAPIURL(apiURL string) string {
	u := strings.TrimRight(apiURL, "/")
	if !strings.HasSuffix(u, "/at/json") {
		u += "/at/json"
	}
	return u
}
