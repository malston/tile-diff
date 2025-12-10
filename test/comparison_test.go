// +build integration

// ABOUTME: Integration tests for tile comparison using real tile files.
// ABOUTME: Run with: go test -tags=integration ./test/...
package test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/malston/tile-diff/pkg/compare"
	"github.com/malston/tile-diff/pkg/metadata"
	"github.com/malston/tile-diff/pkg/report"
)

func TestRealTileComparison(t *testing.T) {
	oldTilePath := "/tmp/elastic-runtime/srt-6.0.22-build.2.pivotal"
	newTilePath := "/tmp/elastic-runtime/srt-10.2.5-build.2.pivotal"

	// Load old tile
	oldMetadata, err := metadata.LoadFromFile(oldTilePath)
	if err != nil {
		t.Skipf("Skipping integration test (old tile not found): %v", err)
		return
	}

	// Load new tile
	newMetadata, err := metadata.LoadFromFile(newTilePath)
	if err != nil {
		t.Skipf("Skipping integration test (new tile not found): %v", err)
		return
	}

	// Compare
	results := compare.CompareMetadata(oldMetadata, newMetadata, true)

	t.Logf("Total properties - Old: %d, New: %d", results.TotalOldProps, results.TotalNewProps)
	t.Logf("Added: %d, Removed: %d, Changed: %d",
		len(results.Added), len(results.Removed), len(results.Changed))

	// Verify we found some differences (tiles are different versions)
	totalChanges := len(results.Added) + len(results.Removed) + len(results.Changed)
	if totalChanges == 0 {
		t.Error("Expected some property differences between versions")
	}

	// Log sample changes
	if len(results.Added) > 0 {
		t.Logf("Sample added property: %s", results.Added[0].PropertyName)
	}
	if len(results.Removed) > 0 {
		t.Logf("Sample removed property: %s", results.Removed[0].PropertyName)
	}
	if len(results.Changed) > 0 {
		t.Logf("Sample changed property: %s - %s",
			results.Changed[0].PropertyName, results.Changed[0].Description)
	}
}

func TestActionableReportGeneration(t *testing.T) {
	oldTilePath := "/tmp/elastic-runtime/srt-6.0.22-build.2.pivotal"
	newTilePath := "/tmp/elastic-runtime/srt-10.2.5-build.2.pivotal"

	// Load tiles
	oldMetadata, err := metadata.LoadFromFile(oldTilePath)
	if err != nil {
		t.Skipf("Skipping integration test (old tile not found): %v", err)
		return
	}

	newMetadata, err := metadata.LoadFromFile(newTilePath)
	if err != nil {
		t.Skipf("Skipping integration test (new tile not found): %v", err)
		return
	}

	// Compare
	results := compare.CompareMetadata(oldMetadata, newMetadata, true)

	// Create mock current config
	mockConfig := &report.CurrentConfig{
		Properties: make(map[string]report.ConfiguredProperty),
	}

	// Filter and categorize
	filtered := report.FilterRelevantChanges(results, mockConfig)
	categorized := report.CategorizeChanges(filtered)

	// Generate reports
	textReport := report.GenerateTextReport(categorized, "6.0.22", "10.2.5")
	jsonReport := report.GenerateJSONReport(categorized, "6.0.22", "10.2.5")

	// Verify text report
	if !strings.Contains(textReport, "Upgrade Analysis") {
		t.Error("Expected 'Upgrade Analysis' in text report")
	}

	// Verify JSON report
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonReport), &jsonData); err != nil {
		t.Errorf("Invalid JSON report: %v", err)
	}

	t.Logf("Generated text report (%d bytes)", len(textReport))
	t.Logf("Generated JSON report (%d bytes)", len(jsonReport))
	t.Logf("Required Actions: %d", len(categorized.RequiredActions))
	t.Logf("Warnings: %d", len(categorized.Warnings))
	t.Logf("Informational: %d", len(categorized.Informational))
}
