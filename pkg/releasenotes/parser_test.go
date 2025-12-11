package releasenotes

import (
	"os"
	"strings"
	"testing"
)

func TestParseHTML(t *testing.T) {
	html, err := os.ReadFile("testdata/sample-release-notes.html")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	features, err := ParseHTML(string(html))
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	if len(features) < 2 {
		t.Errorf("Expected at least 2 features, got %d", len(features))
	}

	// Check first feature
	if features[0].Title != "Enhanced Security Scanning" {
		t.Errorf("Expected 'Enhanced Security Scanning', got %s", features[0].Title)
	}

	if !containsString(features[0].Description, "vulnerability detection") {
		t.Error("Expected description to mention vulnerability detection")
	}
}

func TestExtractFeatures_PropertyMentions(t *testing.T) {
	html, _ := os.ReadFile("testdata/sample-release-notes.html")
	features, _ := ParseHTML(string(html))

	// First feature should mention security_scanner_enabled
	if !containsString(features[0].Description, "security_scanner_enabled") {
		t.Error("Expected feature to mention security_scanner_enabled")
	}
}

func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}
