package api

import (
	"encoding/json"
	"testing"
)

func TestPropertyResponseUnmarshal(t *testing.T) {
	jsonData := []byte(`{
		"type": "boolean",
		"configurable": true,
		"credential": false,
		"value": true,
		"optional": false
	}`)

	var prop Property
	err := json.Unmarshal(jsonData, &prop)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if prop.Type != "boolean" {
		t.Errorf("Expected type 'boolean', got '%s'", prop.Type)
	}
	if !prop.Configurable {
		t.Error("Expected configurable to be true")
	}
	if prop.Optional {
		t.Error("Expected optional to be false")
	}
}

func TestPropertiesResponseUnmarshal(t *testing.T) {
	jsonData := []byte(`{
		"properties": {
			".properties.test_prop": {
				"type": "string",
				"configurable": true,
				"credential": false,
				"value": "test_value",
				"optional": false
			},
			".properties.selector_prop": {
				"type": "selector",
				"configurable": true,
				"credential": false,
				"value": "enable",
				"optional": false,
				"selected_option": "enable"
			}
		}
	}`)

	var response PropertiesResponse
	err := json.Unmarshal(jsonData, &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(response.Properties) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(response.Properties))
	}

	testProp, exists := response.Properties[".properties.test_prop"]
	if !exists {
		t.Error("Expected .properties.test_prop to exist")
	}
	if testProp.Value != "test_value" {
		t.Errorf("Expected value 'test_value', got '%v'", testProp.Value)
	}

	selectorProp, exists := response.Properties[".properties.selector_prop"]
	if !exists {
		t.Error("Expected .properties.selector_prop to exist")
	}
	if selectorProp.SelectedOption == nil || *selectorProp.SelectedOption != "enable" {
		t.Error("Expected selected_option to be 'enable'")
	}
}
