// +build integration

// ABOUTME: Integration tests for tile-diff using real tile files.
// ABOUTME: Run with: go test -tags=integration ./test/...
package test

import (
	"testing"

	"github.com/malston/tile-diff/pkg/metadata"
)

// TestMetadataExtractionRealTile tests loading from actual .pivotal file
// Set TILE_PATH environment variable to test with real tile
func TestMetadataExtractionRealTile(t *testing.T) {
	tilePath := "/tmp/elastic-runtime/srt-6.0.22-build.2.pivotal"

	metadata, err := metadata.LoadFromFile(tilePath)
	if err != nil {
		t.Skipf("Skipping integration test (tile not found): %v", err)
		return
	}

	// Verify we got reasonable data
	if len(metadata.PropertyBlueprints) < 100 {
		t.Errorf("Expected at least 100 properties, got %d", len(metadata.PropertyBlueprints))
	}

	// Count configurable properties
	configurable := 0
	for _, bp := range metadata.PropertyBlueprints {
		if bp.Configurable {
			configurable++
		}
	}

	t.Logf("Total properties: %d", len(metadata.PropertyBlueprints))
	t.Logf("Configurable properties: %d", configurable)

	if configurable == 0 {
		t.Error("Expected some configurable properties")
	}
}
