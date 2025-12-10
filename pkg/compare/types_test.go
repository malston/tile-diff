// ABOUTME: Unit tests for comparison result data structures.
// ABOUTME: Validates comparison result type definitions and constructors.
package compare

import (
	"testing"

	"github.com/malston/tile-diff/pkg/metadata"
)

func TestComparisonResult(t *testing.T) {
	oldProp := metadata.PropertyBlueprint{
		Name:         "test_property",
		Type:         "string",
		Configurable: true,
	}
	newProp := metadata.PropertyBlueprint{
		Name:         "test_property",
		Type:         "integer",
		Configurable: true,
	}

	result := ComparisonResult{
		PropertyName: "test_property",
		ChangeType:   TypeChanged,
		OldProperty:  &oldProp,
		NewProperty:  &newProp,
		Description:  "Type changed from string to integer",
	}

	if result.PropertyName != "test_property" {
		t.Errorf("Expected PropertyName 'test_property', got '%s'", result.PropertyName)
	}
	if result.ChangeType != TypeChanged {
		t.Errorf("Expected ChangeType TypeChanged, got %v", result.ChangeType)
	}
	if result.OldProperty.Type != "string" {
		t.Errorf("Expected OldProperty type 'string', got '%s'", result.OldProperty.Type)
	}
	if result.NewProperty.Type != "integer" {
		t.Errorf("Expected NewProperty type 'integer', got '%s'", result.NewProperty.Type)
	}
}

func TestChangeTypes(t *testing.T) {
	// Verify all change type constants are defined
	changeTypes := []ChangeType{
		PropertyAdded,
		PropertyRemoved,
		TypeChanged,
		OptionalityChanged,
	}

	for _, ct := range changeTypes {
		if ct == "" {
			t.Error("ChangeType should not be empty string")
		}
	}
}
