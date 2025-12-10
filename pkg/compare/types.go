// ABOUTME: Data structures for comparison results.
// ABOUTME: Defines types for categorizing and describing property changes between versions.
package compare

import "github.com/malston/tile-diff/pkg/metadata"

// ChangeType represents the type of change detected
type ChangeType string

const (
	PropertyAdded      ChangeType = "added"
	PropertyRemoved    ChangeType = "removed"
	TypeChanged        ChangeType = "type_changed"
	OptionalityChanged ChangeType = "optionality_changed"
)

// ComparisonResult represents a single property difference between versions
type ComparisonResult struct {
	PropertyName string
	ChangeType   ChangeType
	OldProperty  *metadata.PropertyBlueprint
	NewProperty  *metadata.PropertyBlueprint
	Description  string
}

// ComparisonResults holds all comparison results
type ComparisonResults struct {
	Added            []ComparisonResult
	Removed          []ComparisonResult
	Changed          []ComparisonResult
	TotalOldProps    int
	TotalNewProps    int
	ConfigurableOnly bool
}
