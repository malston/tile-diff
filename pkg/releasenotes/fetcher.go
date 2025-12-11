// ABOUTME: HTTP client for fetching release notes from documentation sites.
// ABOUTME: Implements in-memory caching and timeout handling.
package releasenotes

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// Fetcher fetches and caches release notes HTML
type Fetcher struct {
	client *http.Client
	cache  map[string]string
}

// NewFetcher creates a new release notes fetcher
func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache: make(map[string]string),
	}
}

// Fetch retrieves release notes HTML from URL (with caching)
func (f *Fetcher) Fetch(url string) (string, error) {
	// Check cache
	if cached, ok := f.cache[url]; ok {
		return cached, nil
	}

	// Fetch from URL
	resp, err := f.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d fetching %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	html := string(body)
	f.cache[url] = html

	return html, nil
}
