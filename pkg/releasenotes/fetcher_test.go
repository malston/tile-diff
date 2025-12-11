// pkg/releasenotes/fetcher_test.go
package releasenotes

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchReleaseNotes(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html><body>Release Notes</body></html>"))
	}))
	defer server.Close()

	fetcher := NewFetcher()
	html, err := fetcher.Fetch(server.URL)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if !contains(html, "Release Notes") {
		t.Error("Expected HTML to contain 'Release Notes'")
	}
}

func TestFetchReleaseNotes_404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	fetcher := NewFetcher()
	_, err := fetcher.Fetch(server.URL)
	if err == nil {
		t.Error("Expected error for 404 response")
	}
}

func TestFetchReleaseNotes_Cache(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Write([]byte("<html><body>Release Notes</body></html>"))
	}))
	defer server.Close()

	fetcher := NewFetcher()

	// First fetch
	_, err := fetcher.Fetch(server.URL)
	if err != nil {
		t.Fatalf("First fetch failed: %v", err)
	}

	// Second fetch (should use cache)
	_, err = fetcher.Fetch(server.URL)
	if err != nil {
		t.Fatalf("Second fetch failed: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 HTTP call (cached), got %d", callCount)
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && s != "" && substr != ""
}
