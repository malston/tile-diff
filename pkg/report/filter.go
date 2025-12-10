// ABOUTME: Filters comparison results to show only relevant changes.
// ABOUTME: Determines which property changes affect the current deployed configuration.
package report

import "github.com/malston/tile-diff/pkg/compare"

// FilterRelevantChanges filters comparison results to include only changes relevant to current config
func FilterRelevantChanges(allChanges *compare.ComparisonResults, currentConfig *CurrentConfig) *compare.ComparisonResults {
	filtered := &compare.ComparisonResults{
		TotalOldProps:    allChanges.TotalOldProps,
		TotalNewProps:    allChanges.TotalNewProps,
		ConfigurableOnly: allChanges.ConfigurableOnly,
	}

	// Filter added properties (always relevant)
	filtered.Added = allChanges.Added

	// Filter removed properties (only if configured)
	for _, change := range allChanges.Removed {
		if isChangeRelevant(change.ChangeType, change.PropertyName, currentConfig) {
			filtered.Removed = append(filtered.Removed, change)
		}
	}

	// Filter changed properties (only if configured)
	for _, change := range allChanges.Changed {
		if isChangeRelevant(change.ChangeType, change.PropertyName, currentConfig) {
			filtered.Changed = append(filtered.Changed, change)
		}
	}

	return filtered
}

// isChangeRelevant determines if a property change is relevant to current config
func isChangeRelevant(changeType compare.ChangeType, propertyName string, currentConfig *CurrentConfig) bool {
	// New properties are always relevant (might become required)
	if changeType == compare.PropertyAdded {
		return true
	}

	// For removals and changes, check if property is currently configured
	prop, exists := currentConfig.Properties[propertyName]
	if !exists {
		return false
	}

	return prop.IsConfigured()
}
