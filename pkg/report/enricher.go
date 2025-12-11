// ABOUTME: Enriches comparison results with feature context from release notes.
// ABOUTME: Groups properties by feature and enhances recommendations.
package report

import (
	"sort"

	"github.com/malston/tile-diff/pkg/releasenotes"
)

// FeatureGroup represents properties grouped by feature
type FeatureGroup struct {
	Name        string
	Description string
	Properties  []string
}

// EnrichedChanges extends CategorizedChanges with feature context
type EnrichedChanges struct {
	*CategorizedChanges
	Features []FeatureGroup
}

// EnrichChanges adds feature context to categorized changes
func EnrichChanges(changes *CategorizedChanges, matches map[string]releasenotes.Match) *EnrichedChanges {
	enriched := &EnrichedChanges{
		CategorizedChanges: changes,
	}

	// Group properties by feature
	featureMap := make(map[string]*FeatureGroup)

	// Helper function to process changes
	processChanges := func(changeList []CategorizedChange) {
		for _, change := range changeList {
			if match, ok := matches[change.PropertyName]; ok {
				featureName := match.Feature.Title
				if _, exists := featureMap[featureName]; !exists {
					featureMap[featureName] = &FeatureGroup{
						Name:        match.Feature.Title,
						Description: match.Feature.Description,
						Properties:  []string{},
					}
				}
				featureMap[featureName].Properties = append(
					featureMap[featureName].Properties,
					change.PropertyName,
				)
			}
		}
	}

	// Process all three categories
	processChanges(changes.RequiredActions)
	processChanges(changes.Warnings)
	processChanges(changes.Informational)

	// Convert map to slice with deterministic ordering
	featureNames := make([]string, 0, len(featureMap))
	for name := range featureMap {
		featureNames = append(featureNames, name)
	}
	sort.Strings(featureNames)

	for _, name := range featureNames {
		enriched.Features = append(enriched.Features, *featureMap[name])
	}

	return enriched
}
