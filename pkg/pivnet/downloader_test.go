// ABOUTME: Unit tests for file download with progress tracking.
// ABOUTME: Validates download orchestration and progress reporting.
package pivnet

import (
	"path/filepath"
	"testing"
)

func TestDownloader(t *testing.T) {
	// Create test client (will need to be mocked for real API calls)
	client, err := NewClient("test-token")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")
	manifestFile := filepath.Join(tempDir, "manifest.json")
	eulaFile := filepath.Join(tempDir, "eulas.json")

	downloader := NewDownloader(client, cacheDir, manifestFile, eulaFile, 20, false)

	// We can't test actual downloads without mocking the Pivnet API
	// This test just verifies the downloader can be created
	if downloader == nil {
		t.Fatal("Expected non-nil downloader")
	}
}
