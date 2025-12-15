// ABOUTME: Compares configuration templates from two tiles.
// ABOUTME: Identifies added, removed, and changed configuration properties.
package om

import (
	"fmt"
)

// ConfigChange represents a change in configuration
type ConfigChange struct {
	PropertyName string
	ChangeType   string // "added", "removed", "changed"
	Description  string
	OldValue     interface{}
	NewValue     interface{}
}

// ConfigComparison holds the results of comparing two configs
type ConfigComparison struct {
	Added   []ConfigChange
	Removed []ConfigChange
	Changed []ConfigChange
}

// CompareConfigs compares two ProductConfig structures
func CompareConfigs(oldConfig, newConfig *ProductConfig) *ConfigComparison {
	result := &ConfigComparison{
		Added:   []ConfigChange{},
		Removed: []ConfigChange{},
		Changed: []ConfigChange{},
	}

	oldProps := oldConfig.ProductProperties
	newProps := newConfig.ProductProperties

	// Find added and changed properties
	for name, newProp := range newProps {
		if oldProp, exists := oldProps[name]; exists {
			// Property exists in both - check if value template changed
			oldVal := fmt.Sprintf("%v", oldProp.Value)
			newVal := fmt.Sprintf("%v", newProp.Value)
			if oldVal != newVal {
				result.Changed = append(result.Changed, ConfigChange{
					PropertyName: name,
					ChangeType:   "changed",
					Description:  fmt.Sprintf("Default value changed"),
					OldValue:     oldProp.Value,
					NewValue:     newProp.Value,
				})
			}
		} else {
			// New property
			result.Added = append(result.Added, ConfigChange{
				PropertyName: name,
				ChangeType:   "added",
				Description:  "New configurable property",
				NewValue:     newProp.Value,
			})
		}
	}

	// Find removed properties
	for name, oldProp := range oldProps {
		if _, exists := newProps[name]; !exists {
			result.Removed = append(result.Removed, ConfigChange{
				PropertyName: name,
				ChangeType:   "removed",
				Description:  "Property removed",
				OldValue:     oldProp.Value,
			})
		}
	}

	return result
}
