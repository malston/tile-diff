// ABOUTME: Unit tests for tile metadata type definitions.
// ABOUTME: Validates YAML unmarshaling for property blueprints and selectors.
package metadata

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestPropertyBlueprintUnmarshal(t *testing.T) {
	yamlData := `
name: test_property
type: boolean
configurable: true
optional: false
default: true
`
	var pb PropertyBlueprint
	err := yaml.Unmarshal([]byte(yamlData), &pb)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if pb.Name != "test_property" {
		t.Errorf("Expected name 'test_property', got '%s'", pb.Name)
	}
	if pb.Type != "boolean" {
		t.Errorf("Expected type 'boolean', got '%s'", pb.Type)
	}
	if !pb.Configurable {
		t.Error("Expected configurable to be true")
	}
	if pb.Optional {
		t.Error("Expected optional to be false")
	}
}

func TestConstraintsUnmarshal(t *testing.T) {
	// Test constraints as object (min/max)
	yamlData := `
name: count
type: integer
constraints:
  min: 1
  max: 100
`
	var pb PropertyBlueprint
	err := yaml.Unmarshal([]byte(yamlData), &pb)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if pb.Constraints == nil {
		t.Fatal("Expected constraints to be non-nil")
	}

	// Test constraints as array (regex patterns)
	yamlData2 := `
name: pattern_field
type: string
constraints:
  - must_match_regex: ^[^"\\\]]+$
    error_message: cannot contain special characters
`
	var pb2 PropertyBlueprint
	err = yaml.Unmarshal([]byte(yamlData2), &pb2)
	if err != nil {
		t.Fatalf("Failed to unmarshal array constraints: %v", err)
	}

	if pb2.Constraints == nil {
		t.Fatal("Expected array constraints to be non-nil")
	}
}

func TestSelectorWithOptions(t *testing.T) {
	yamlData := `
name: mode
type: selector
option_templates:
  - name: enable
    select_value: enable
    property_blueprints:
      - name: threshold
        type: integer
        default: 100
  - name: disable
    select_value: disable
`
	var pb PropertyBlueprint
	err := yaml.Unmarshal([]byte(yamlData), &pb)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if pb.Type != "selector" {
		t.Errorf("Expected type 'selector', got '%s'", pb.Type)
	}
	if len(pb.OptionTemplates) != 2 {
		t.Errorf("Expected 2 option templates, got %d", len(pb.OptionTemplates))
	}
	if pb.OptionTemplates[0].Name != "enable" {
		t.Errorf("Expected first option name 'enable', got '%s'", pb.OptionTemplates[0].Name)
	}
}
