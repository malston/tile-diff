// ABOUTME: Unit tests for categorizing changes by severity.
// ABOUTME: Validates classification into Required Actions, Warnings, and Informational.
package report

import (
	"testing"

	"github.com/malston/tile-diff/pkg/compare"
	"github.com/malston/tile-diff/pkg/metadata"
)

func TestCategorizeChanges(t *testing.T) {
	newReqProp := metadata.PropertyBlueprint{
		Name:     "new_required",
		Type:     "string",
		Optional: false,
	}
	newOptProp := metadata.PropertyBlueprint{
		Name:     "new_optional",
		Type:     "boolean",
		Optional: true,
	}
	removedProp := metadata.PropertyBlueprint{
		Name: "removed_prop",
		Type: "string",
	}
	oldTypeProp := metadata.PropertyBlueprint{
		Name: "type_changed",
		Type: "string",
	}
	newTypeProp := metadata.PropertyBlueprint{
		Name: "type_changed",
		Type: "integer",
	}

	changes := &compare.ComparisonResults{
		Added: []compare.ComparisonResult{
			{PropertyName: "new_required", NewProperty: &newReqProp, ChangeType: compare.PropertyAdded},
			{PropertyName: "new_optional", NewProperty: &newOptProp, ChangeType: compare.PropertyAdded},
		},
		Removed: []compare.ComparisonResult{
			{PropertyName: "removed_prop", OldProperty: &removedProp, ChangeType: compare.PropertyRemoved},
		},
		Changed: []compare.ComparisonResult{
			{PropertyName: "type_changed", OldProperty: &oldTypeProp, NewProperty: &newTypeProp, ChangeType: compare.TypeChanged},
		},
	}

	categorized := CategorizeChanges(changes)

	// New required properties = Required Actions
	if len(categorized.RequiredActions) != 1 {
		t.Errorf("Expected 1 required action, got %d", len(categorized.RequiredActions))
	}
	if categorized.RequiredActions[0].PropertyName != "new_required" {
		t.Error("Expected new_required in Required Actions")
	}

	// Removed and type changed = Warnings
	if len(categorized.Warnings) != 2 {
		t.Errorf("Expected 2 warnings, got %d", len(categorized.Warnings))
	}

	// New optional = Informational
	if len(categorized.Informational) != 1 {
		t.Errorf("Expected 1 informational, got %d", len(categorized.Informational))
	}
	if categorized.Informational[0].PropertyName != "new_optional" {
		t.Error("Expected new_optional in Informational")
	}
}

func TestDetermineCategory(t *testing.T) {
	tests := []struct {
		name       string
		change     compare.ComparisonResult
		expected   Category
	}{
		{
			name: "new required property without default",
			change: compare.ComparisonResult{
				ChangeType:  compare.PropertyAdded,
				NewProperty: &metadata.PropertyBlueprint{Optional: false},
			},
			expected: CategoryRequired,
		},
		{
			name: "new required property with default",
			change: compare.ComparisonResult{
				ChangeType:  compare.PropertyAdded,
				NewProperty: &metadata.PropertyBlueprint{
					Optional: false,
					Default:  true,
				},
			},
			expected: CategoryInformational,
		},
		{
			name: "new optional property",
			change: compare.ComparisonResult{
				ChangeType:  compare.PropertyAdded,
				NewProperty: &metadata.PropertyBlueprint{Optional: true},
			},
			expected: CategoryInformational,
		},
		{
			name: "removed property",
			change: compare.ComparisonResult{
				ChangeType: compare.PropertyRemoved,
			},
			expected: CategoryWarning,
		},
		{
			name: "type changed",
			change: compare.ComparisonResult{
				ChangeType: compare.TypeChanged,
			},
			expected: CategoryWarning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineCategory(tt.change)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
