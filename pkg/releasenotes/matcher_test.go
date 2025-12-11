package releasenotes

import (
	"testing"
)

func TestDirectMatch(t *testing.T) {
	features := []Feature{
		{
			Title:       "Enhanced Security",
			Description: "Use .properties.security_scanner_enabled to enable scanning",
			Position:    1,
		},
	}

	properties := []string{
		".properties.security_scanner_enabled",
		".properties.unrelated_property",
	}

	matcher := NewMatcher(features)
	matches := matcher.Match(properties)

	// Should have one match for security_scanner_enabled
	if len(matches) != 1 {
		t.Errorf("Expected 1 match, got %d", len(matches))
	}

	match := matches[".properties.security_scanner_enabled"]
	if match.MatchType != "direct" {
		t.Errorf("Expected direct match, got %s", match.MatchType)
	}

	if match.Confidence != 1.0 {
		t.Errorf("Expected confidence 1.0, got %f", match.Confidence)
	}
}
