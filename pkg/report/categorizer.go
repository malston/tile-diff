// ABOUTME: Categorizes property changes by severity and required action.
// ABOUTME: Classifies changes into Required Actions, Warnings, and Informational.
package report

import "github.com/malston/tile-diff/pkg/compare"

// Category represents the severity/type of a change
type Category string

const (
	CategoryRequired      Category = "required"
	CategoryWarning       Category = "warning"
	CategoryInformational Category = "informational"
)

// CategorizedChange represents a change with its severity category
type CategorizedChange struct {
	compare.ComparisonResult
	Category      Category
	Recommendation string
}

// CategorizedChanges holds changes grouped by category
type CategorizedChanges struct {
	RequiredActions []CategorizedChange
	Warnings        []CategorizedChange
	Informational   []CategorizedChange
}

// CategorizeChanges classifies comparison results into severity categories
func CategorizeChanges(changes *compare.ComparisonResults) *CategorizedChanges {
	categorized := &CategorizedChanges{}

	// Categorize added properties
	for _, change := range changes.Added {
		cat := determineCategory(change)
		catChange := CategorizedChange{
			ComparisonResult: change,
			Category:         cat,
			Recommendation:   generateRecommendation(change, cat),
		}

		switch cat {
		case CategoryRequired:
			categorized.RequiredActions = append(categorized.RequiredActions, catChange)
		case CategoryInformational:
			categorized.Informational = append(categorized.Informational, catChange)
		}
	}

	// Categorize removed properties
	for _, change := range changes.Removed {
		cat := determineCategory(change)
		catChange := CategorizedChange{
			ComparisonResult: change,
			Category:         cat,
			Recommendation:   generateRecommendation(change, cat),
		}
		categorized.Warnings = append(categorized.Warnings, catChange)
	}

	// Categorize changed properties
	for _, change := range changes.Changed {
		cat := determineCategory(change)
		catChange := CategorizedChange{
			ComparisonResult: change,
			Category:         cat,
			Recommendation:   generateRecommendation(change, cat),
		}
		categorized.Warnings = append(categorized.Warnings, catChange)
	}

	return categorized
}

// determineCategory determines the severity category for a change
func determineCategory(change compare.ComparisonResult) Category {
	switch change.ChangeType {
	case compare.PropertyAdded:
		// Required if not optional AND no default
		if change.NewProperty != nil && !change.NewProperty.Optional && change.NewProperty.Default == nil {
			return CategoryRequired
		}
		return CategoryInformational

	case compare.PropertyRemoved:
		// Removal of configured property is a warning
		return CategoryWarning

	case compare.TypeChanged, compare.OptionalityChanged:
		// Type or optionality changes are warnings
		return CategoryWarning

	default:
		return CategoryInformational
	}
}

// generateRecommendation creates a recommendation string for a change
func generateRecommendation(change compare.ComparisonResult, category Category) string {
	switch category {
	case CategoryRequired:
		return "Must configure this property before upgrading"
	case CategoryWarning:
		if change.ChangeType == compare.PropertyRemoved {
			return "Property will be ignored after upgrade - review and remove from config"
		}
		return "Review this change and verify compatibility"
	case CategoryInformational:
		return "Optional - review for potential improvements"
	default:
		return ""
	}
}
