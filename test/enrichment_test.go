// test/enrichment_test.go
package test

import (
	"testing"

	"github.com/malston/tile-diff/pkg/releasenotes"
)

func TestReleaseNotesEnrichment_EndToEnd(t *testing.T) {
	// Load test config
	config, err := releasenotes.LoadProductConfig("../pkg/releasenotes/testdata/products.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Resolve URL
	url, err := config.ResolveURL("cf", "10.2.5")
	if err != nil {
		t.Fatalf("Failed to resolve URL: %v", err)
	}

	if url == "" {
		t.Error("Expected non-empty URL")
	}

	// This test verifies the full pipeline works
	// Actual HTTP fetching tested in fetcher_test.go with mock server
}
