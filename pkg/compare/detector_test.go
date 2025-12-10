// ABOUTME: Unit tests for property change detection logic.
// ABOUTME: Validates identification of new, removed, and changed properties.
package compare

import (
	"testing"

	"github.com/malston/tile-diff/pkg/metadata"
)

func TestFindNewProperties(t *testing.T) {
	oldProps := map[string]metadata.PropertyBlueprint{
		"existing_prop": {
			Name:         "existing_prop",
			Type:         "string",
			Configurable: true,
		},
	}

	newProps := map[string]metadata.PropertyBlueprint{
		"existing_prop": {
			Name:         "existing_prop",
			Type:         "string",
			Configurable: true,
		},
		"new_prop": {
			Name:         "new_prop",
			Type:         "boolean",
			Configurable: true,
		},
		"new_system_prop": {
			Name:         "new_system_prop",
			Type:         "string",
			Configurable: false,
		},
	}

	// Test: find all new properties
	allNew := FindNewProperties(oldProps, newProps, false)
	if len(allNew) != 2 {
		t.Errorf("Expected 2 new properties, got %d", len(allNew))
	}

	// Test: find only configurable new properties
	configurableNew := FindNewProperties(oldProps, newProps, true)
	if len(configurableNew) != 1 {
		t.Errorf("Expected 1 configurable new property, got %d", len(configurableNew))
	}
	if configurableNew[0].PropertyName != "new_prop" {
		t.Errorf("Expected new_prop, got %s", configurableNew[0].PropertyName)
	}
	if configurableNew[0].ChangeType != PropertyAdded {
		t.Errorf("Expected ChangeType PropertyAdded, got %v", configurableNew[0].ChangeType)
	}
}

func TestFindNewPropertiesNone(t *testing.T) {
	oldProps := map[string]metadata.PropertyBlueprint{
		"prop1": {Name: "prop1", Type: "string", Configurable: true},
	}
	newProps := map[string]metadata.PropertyBlueprint{
		"prop1": {Name: "prop1", Type: "string", Configurable: true},
	}

	results := FindNewProperties(oldProps, newProps, false)
	if len(results) != 0 {
		t.Errorf("Expected no new properties, got %d", len(results))
	}
}

func TestFindRemovedProperties(t *testing.T) {
	oldProps := map[string]metadata.PropertyBlueprint{
		"existing_prop": {
			Name:         "existing_prop",
			Type:         "string",
			Configurable: true,
		},
		"removed_prop": {
			Name:         "removed_prop",
			Type:         "boolean",
			Configurable: true,
		},
		"removed_system_prop": {
			Name:         "removed_system_prop",
			Type:         "string",
			Configurable: false,
		},
	}

	newProps := map[string]metadata.PropertyBlueprint{
		"existing_prop": {
			Name:         "existing_prop",
			Type:         "string",
			Configurable: true,
		},
	}

	// Test: find all removed properties
	allRemoved := FindRemovedProperties(oldProps, newProps, false)
	if len(allRemoved) != 2 {
		t.Errorf("Expected 2 removed properties, got %d", len(allRemoved))
	}

	// Test: find only configurable removed properties
	configurableRemoved := FindRemovedProperties(oldProps, newProps, true)
	if len(configurableRemoved) != 1 {
		t.Errorf("Expected 1 configurable removed property, got %d", len(configurableRemoved))
	}
	if configurableRemoved[0].PropertyName != "removed_prop" {
		t.Errorf("Expected removed_prop, got %s", configurableRemoved[0].PropertyName)
	}
	if configurableRemoved[0].ChangeType != PropertyRemoved {
		t.Errorf("Expected ChangeType PropertyRemoved, got %v", configurableRemoved[0].ChangeType)
	}
}

func TestFindChangedProperties(t *testing.T) {
	oldProps := map[string]metadata.PropertyBlueprint{
		"unchanged_prop": {
			Name:         "unchanged_prop",
			Type:         "string",
			Configurable: true,
			Optional:     false,
		},
		"type_changed": {
			Name:         "type_changed",
			Type:         "string",
			Configurable: true,
			Optional:     false,
		},
		"optional_changed": {
			Name:         "optional_changed",
			Type:         "string",
			Configurable: true,
			Optional:     false,
		},
		"system_prop_changed": {
			Name:         "system_prop_changed",
			Type:         "string",
			Configurable: false,
			Optional:     false,
		},
	}

	newProps := map[string]metadata.PropertyBlueprint{
		"unchanged_prop": {
			Name:         "unchanged_prop",
			Type:         "string",
			Configurable: true,
			Optional:     false,
		},
		"type_changed": {
			Name:         "type_changed",
			Type:         "integer",
			Configurable: true,
			Optional:     false,
		},
		"optional_changed": {
			Name:         "optional_changed",
			Type:         "string",
			Configurable: true,
			Optional:     true,
		},
		"system_prop_changed": {
			Name:         "system_prop_changed",
			Type:         "integer",
			Configurable: false,
			Optional:     false,
		},
	}

	// Test: find all changed properties
	allChanged := FindChangedProperties(oldProps, newProps, false)
	if len(allChanged) != 3 {
		t.Errorf("Expected 3 changed properties, got %d", len(allChanged))
	}

	// Test: find only configurable changed properties
	configurableChanged := FindChangedProperties(oldProps, newProps, true)
	if len(configurableChanged) != 2 {
		t.Errorf("Expected 2 configurable changed properties, got %d", len(configurableChanged))
	}

	// Verify type change detection
	var typeChangeFound bool
	for _, result := range configurableChanged {
		if result.PropertyName == "type_changed" && result.ChangeType == TypeChanged {
			typeChangeFound = true
			if result.OldProperty.Type != "string" || result.NewProperty.Type != "integer" {
				t.Error("Type change not correctly detected")
			}
		}
	}
	if !typeChangeFound {
		t.Error("Type change not detected")
	}

	// Verify optionality change detection
	var optionalityChangeFound bool
	for _, result := range configurableChanged {
		if result.PropertyName == "optional_changed" && result.ChangeType == OptionalityChanged {
			optionalityChangeFound = true
		}
	}
	if !optionalityChangeFound {
		t.Error("Optionality change not detected")
	}
}
