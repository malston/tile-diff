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

	if match.Property != ".properties.security_scanner_enabled" {
		t.Errorf("Expected property .properties.security_scanner_enabled, got %s", match.Property)
	}

	if match.Feature.Title != "Enhanced Security" {
		t.Errorf("Expected feature title 'Enhanced Security', got %s", match.Feature.Title)
	}
}

func TestKeywordMatch(t *testing.T) {
	features := []Feature{
		{
			Title:       "Application Log Rate Limiting",
			Description: "Prevent log flooding with rate limits on application logs",
			Position:    1,
		},
	}

	properties := []string{
		".properties.app_log_rate_limiting",
	}

	matcher := NewMatcher(features)
	matches := matcher.Match(properties)

	match := matches[".properties.app_log_rate_limiting"]
	if match.MatchType != "keyword" {
		t.Errorf("Expected keyword match, got %s", match.MatchType)
	}

	expectedConfidence := 0.9
	if match.Confidence != expectedConfidence {
		t.Errorf("Expected confidence %f, got %f", expectedConfidence, match.Confidence)
	}
}

func TestTokenizeProperty(t *testing.T) {
	tests := []struct {
		property string
		expected []string
	}{
		{
			property: ".properties.app_log_rate_limiting",
			expected: []string{"app", "log", "rate", "limiting"},
		},
		{
			property: ".properties.security_scanner_enabled",
			expected: []string{"security", "scanner"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.property, func(t *testing.T) {
			tokens := tokenizeProperty(tt.property)
			if len(tokens) != len(tt.expected) {
				t.Errorf("Expected %d tokens, got %d", len(tt.expected), len(tokens))
			}
			for i, token := range tokens {
				if i >= len(tt.expected) || token != tt.expected[i] {
					t.Errorf("Token %d: expected %q, got %q", i, tt.expected[i], token)
				}
			}
		})
	}
}
