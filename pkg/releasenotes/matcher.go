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

var stopwords = map[string]bool{
	"enable":     true,
	"enabled":    true,
	"setting":    true,
	"settings":   true,
	"config":     true,
	"configure":  true,
	"new":        true,
	"property":   true,
	"properties": true,
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

	// Try keyword matching
	tokens := tokenizeProperty(property)
	if len(tokens) >= 2 {
		for _, feature := range m.features {
			score := keywordScore(tokens, feature)
			if score > bestConfidence {
				bestConfidence = score
				bestMatch = Match{
					Property:   property,
					Feature:    feature,
					MatchType:  "keyword",
					Confidence: score,
				}
			}
		}
	}

	// Only return if confidence > 0.5
	if bestConfidence > 0.5 {
		return bestMatch, true
	}

	return bestMatch, false
}

func tokenizeProperty(property string) []string {
	// Remove common prefixes
	prop := strings.TrimPrefix(property, ".properties.")
	prop = strings.TrimPrefix(prop, ".cloud_controller.")

	// Split on underscores and dots
	parts := strings.FieldsFunc(prop, func(r rune) bool {
		return r == '_' || r == '.' || r == '-'
	})

	// Filter stopwords
	var tokens []string
	for _, part := range parts {
		lower := strings.ToLower(part)
		if !stopwords[lower] && len(lower) > 2 {
			tokens = append(tokens, lower)
		}
	}

	return tokens
}

func keywordScore(tokens []string, feature Feature) float64 {
	if len(tokens) == 0 {
		return 0.0
	}

	searchText := strings.ToLower(feature.Title + " " + feature.Description)
	matched := 0

	for _, token := range tokens {
		if strings.Contains(searchText, token) {
			matched++
		}
	}

	// Score: (matched / total) * 0.9
	return (float64(matched) / float64(len(tokens))) * 0.9
}
