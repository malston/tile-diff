// ABOUTME: High-level comparison orchestration combining all detection algorithms.
// ABOUTME: Provides single entry point for comparing two tile metadata versions.
package compare

import "github.com/malston/tile-diff/pkg/metadata"

// CompareMetadata performs a complete comparison between old and new tile metadata
func CompareMetadata(oldMetadata, newMetadata *metadata.TileMetadata, configurableOnly bool) *ComparisonResults {
	// Build property maps
	oldProps := BuildPropertyMap(oldMetadata.PropertyBlueprints)
	newProps := BuildPropertyMap(newMetadata.PropertyBlueprints)

	// Find all differences
	added := FindNewProperties(oldProps, newProps, configurableOnly)
	removed := FindRemovedProperties(oldProps, newProps, configurableOnly)
	changed := FindChangedProperties(oldProps, newProps, configurableOnly)

	return &ComparisonResults{
		Added:            added,
		Removed:          removed,
		Changed:          changed,
		TotalOldProps:    len(oldMetadata.PropertyBlueprints),
		TotalNewProps:    len(newMetadata.PropertyBlueprints),
		ConfigurableOnly: configurableOnly,
	}
}
