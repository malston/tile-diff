// ABOUTME: Builds property maps from tile metadata for comparison.
// ABOUTME: Converts PropertyBlueprint slices to name-keyed maps for efficient lookups.
package compare

import "github.com/malston/tile-diff/pkg/metadata"

// BuildPropertyMap converts a slice of PropertyBlueprints into a map keyed by property name
func BuildPropertyMap(blueprints []metadata.PropertyBlueprint) map[string]metadata.PropertyBlueprint {
	propertyMap := make(map[string]metadata.PropertyBlueprint, len(blueprints))

	for _, blueprint := range blueprints {
		propertyMap[blueprint.Name] = blueprint
	}

	return propertyMap
}
