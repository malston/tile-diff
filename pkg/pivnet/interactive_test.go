// ABOUTME: Unit tests for interactive selection prompts.
// ABOUTME: Validates prompt generation and selection logic (UI tested manually).
package pivnet

import (
	"testing"
)

func TestPromptForRelease(t *testing.T) {
	// We can't easily test the actual interactive prompt in unit tests
	// but we can test the helper functions

	releases := []Release{
		{ID: 1, Version: "6.0.22+LTS-T"},
		{ID: 2, Version: "6.0.21+LTS-T"},
		{ID: 3, Version: "6.0.20+LTS-T"},
	}

	// Test that we can build prompt options
	options := buildReleaseOptions(releases)
	if len(options) != 3 {
		t.Errorf("Expected 3 options, got %d", len(options))
	}
	if options[0] != "6.0.22+LTS-T (latest)" {
		t.Errorf("Expected first option to be marked as latest, got %s", options[0])
	}
}

func TestPromptForProductFile(t *testing.T) {
	files := []ProductFile{
		{ID: 1, Name: "TAS for VMs", Size: 2 * 1024 * 1024 * 1024},
		{ID: 2, Name: "Small Footprint TAS", Size: 1 * 1024 * 1024 * 1024},
	}

	options := buildProductFileOptions(files)
	if len(options) != 2 {
		t.Errorf("Expected 2 options, got %d", len(options))
	}
}
