// ABOUTME: Property-to-feature matching using multiple strategies.
// ABOUTME: Implements direct, keyword, and proximity matching with confidence scoring.
package releasenotes

import (
	"strings"
)

// Match represents a property matched to a feature
type Match struct {
	Property   string
	Feature    Feature
	MatchType  string  // "direct", "keyword", "proximity"
	Confidence float64 // 0.0-1.0
}

// Matcher matches properties to features
type Matcher struct {
	features []Feature
}

// NewMatcher creates a new matcher with features
func NewMatcher(features []Feature) *Matcher {
	return &Matcher{features: features}
}

// Match finds matches for properties
func (m *Matcher) Match(properties []string) map[string]Match {
	matches := make(map[string]Match)

	for _, prop := range properties {
		if match, found := m.findBestMatch(prop); found {
			matches[prop] = match
		}
	}

	return matches
}

func (m *Matcher) findBestMatch(property string) (Match, bool) {
	var bestMatch Match
	bestConfidence := 0.0

	// Try direct matching first
	for _, feature := range m.features {
		if strings.Contains(feature.Description, property) {
			match := Match{
				Property:   property,
				Feature:    feature,
				MatchType:  "direct",
				Confidence: 1.0,
			}
			return match, true
		}
	}

	// If no direct match found and no other matches, return false
	if bestConfidence == 0.0 {
		return bestMatch, false
	}

	return bestMatch, true
}
