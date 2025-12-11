// ABOUTME: Tests for enriching comparison results with feature context.
// ABOUTME: Verifies feature grouping and property matching from release notes.
package report

import (
	"testing"

	"github.com/malston/tile-diff/pkg/compare"
	"github.com/malston/tile-diff/pkg/metadata"
	"github.com/malston/tile-diff/pkg/releasenotes"
)

func TestEnrichChanges(t *testing.T) {
	features := []releasenotes.Feature{
		{
			Title:       "Enhanced Security",
			Description: "Security scanning feature",
			Position:    1,
		},
	}

	matches := map[string]releasenotes.Match{
		".properties.security_enabled": {
			Property: ".properties.security_enabled",
			Feature:  features[0],
			MatchType: "direct",
			Confidence: 1.0,
		},
	}

	changes := &CategorizedChanges{
		RequiredActions: []CategorizedChange{
			{
				ComparisonResult: compare.ComparisonResult{
					PropertyName: ".properties.security_enabled",
					NewProperty: &metadata.PropertyBlueprint{
						Name: "security_enabled",
						Type: "boolean",
					},
				},
				Category: CategoryRequired,
				Recommendation: "Must configure this property",
			},
		},
	}

	enriched := EnrichChanges(changes, matches)

	if len(enriched.Features) != 1 {
		t.Errorf("Expected 1 feature group, got %d", len(enriched.Features))
	}

	feature := enriched.Features[0]
	if feature.Name != "Enhanced Security" {
		t.Errorf("Expected 'Enhanced Security', got %s", feature.Name)
	}

	if len(feature.Properties) != 1 {
		t.Errorf("Expected 1 property in feature, got %d", len(feature.Properties))
	}
}
