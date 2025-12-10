// ABOUTME: Unit tests for property map building from metadata.
// ABOUTME: Validates conversion of PropertyBlueprint slices to keyed maps.
package compare

import (
	"testing"

	"github.com/malston/tile-diff/pkg/metadata"
)

func TestBuildPropertyMap(t *testing.T) {
	blueprints := []metadata.PropertyBlueprint{
		{
			Name:         "first_property",
			Type:         "string",
			Configurable: true,
			Optional:     false,
		},
		{
			Name:         "second_property",
			Type:         "integer",
			Configurable: true,
			Optional:     true,
		},
		{
			Name:         "system_property",
			Type:         "string",
			Configurable: false,
			Optional:     false,
		},
	}

	propertyMap := BuildPropertyMap(blueprints)

	// Should have all 3 properties
	if len(propertyMap) != 3 {
		t.Errorf("Expected 3 properties in map, got %d", len(propertyMap))
	}

	// Check first property exists and has correct data
	first, exists := propertyMap["first_property"]
	if !exists {
		t.Error("Expected first_property to exist in map")
	}
	if first.Type != "string" {
		t.Errorf("Expected first_property type 'string', got '%s'", first.Type)
	}
	if !first.Configurable {
		t.Error("Expected first_property to be configurable")
	}

	// Check system property
	system, exists := propertyMap["system_property"]
	if !exists {
		t.Error("Expected system_property to exist in map")
	}
	if system.Configurable {
		t.Error("Expected system_property to not be configurable")
	}
}

func TestBuildPropertyMapEmpty(t *testing.T) {
	blueprints := []metadata.PropertyBlueprint{}
	propertyMap := BuildPropertyMap(blueprints)

	if len(propertyMap) != 0 {
		t.Errorf("Expected empty map, got %d properties", len(propertyMap))
	}
}
