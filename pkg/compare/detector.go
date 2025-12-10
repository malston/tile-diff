// ABOUTME: Property change detection algorithms.
// ABOUTME: Identifies new, removed, and changed properties between tile versions.
package compare

import (
	"fmt"

	"github.com/malston/tile-diff/pkg/metadata"
)

// FindNewProperties identifies properties in newProps that don't exist in oldProps
func FindNewProperties(oldProps, newProps map[string]metadata.PropertyBlueprint, configurableOnly bool) []ComparisonResult {
	var results []ComparisonResult

	for name, newProp := range newProps {
		// Skip if property exists in old version
		if _, exists := oldProps[name]; exists {
			continue
		}

		// Skip non-configurable if filtering
		if configurableOnly && !newProp.Configurable {
			continue
		}

		result := ComparisonResult{
			PropertyName: name,
			ChangeType:   PropertyAdded,
			OldProperty:  nil,
			NewProperty:  &newProp,
			Description:  fmt.Sprintf("New property: %s (type: %s)", name, newProp.Type),
		}
		results = append(results, result)
	}

	return results
}

// FindRemovedProperties identifies properties in oldProps that don't exist in newProps
func FindRemovedProperties(oldProps, newProps map[string]metadata.PropertyBlueprint, configurableOnly bool) []ComparisonResult {
	var results []ComparisonResult

	for name, oldProp := range oldProps {
		// Skip if property still exists in new version
		if _, exists := newProps[name]; exists {
			continue
		}

		// Skip non-configurable if filtering
		if configurableOnly && !oldProp.Configurable {
			continue
		}

		result := ComparisonResult{
			PropertyName: name,
			ChangeType:   PropertyRemoved,
			OldProperty:  &oldProp,
			NewProperty:  nil,
			Description:  fmt.Sprintf("Removed property: %s (was type: %s)", name, oldProp.Type),
		}
		results = append(results, result)
	}

	return results
}
