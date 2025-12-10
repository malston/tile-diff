// ABOUTME: Unit tests for parsing Ops Manager current configuration.
// ABOUTME: Validates extraction of configured properties from API response.
package report

import (
	"testing"

	"github.com/malston/tile-diff/pkg/api"
)

func TestParseCurrentConfig(t *testing.T) {
	apiResponse := &api.PropertiesResponse{
		Properties: map[string]api.Property{
			".properties.configured_prop": {
				Type:         "string",
				Configurable: true,
				Credential:   false,
				Value:        "custom_value",
				Optional:     false,
			},
			".properties.default_prop": {
				Type:         "boolean",
				Configurable: true,
				Credential:   false,
				Value:        true,
				Optional:     true,
			},
			".properties.system_prop": {
				Type:         "string",
				Configurable: false,
				Credential:   false,
				Value:        "system_value",
				Optional:     false,
			},
		},
	}

	config := ParseCurrentConfig(apiResponse)

	// Should have all 3 properties
	if len(config.Properties) != 3 {
		t.Errorf("Expected 3 properties, got %d", len(config.Properties))
	}

	// Check configured property exists
	prop, exists := config.Properties["configured_prop"]
	if !exists {
		t.Error("Expected configured_prop to exist")
	}
	if prop.Value != "custom_value" {
		t.Errorf("Expected value 'custom_value', got '%v'", prop.Value)
	}

	// Count configurable properties
	configurableCount := 0
	for _, p := range config.Properties {
		if p.Configurable {
			configurableCount++
		}
	}
	if configurableCount != 2 {
		t.Errorf("Expected 2 configurable properties, got %d", configurableCount)
	}
}

func TestIsPropertyConfigured(t *testing.T) {
	tests := []struct {
		name       string
		prop       ConfiguredProperty
		expected   bool
	}{
		{
			name: "non-optional with value",
			prop: ConfiguredProperty{
				Name:         "test",
				Configurable: true,
				Optional:     false,
				Value:        "value",
			},
			expected: true,
		},
		{
			name: "optional with non-nil value",
			prop: ConfiguredProperty{
				Name:         "test",
				Configurable: true,
				Optional:     true,
				Value:        "value",
			},
			expected: true,
		},
		{
			name: "optional with nil value",
			prop: ConfiguredProperty{
				Name:         "test",
				Configurable: true,
				Optional:     true,
				Value:        nil,
			},
			expected: false,
		},
		{
			name: "non-configurable",
			prop: ConfiguredProperty{
				Name:         "test",
				Configurable: false,
				Optional:     false,
				Value:        "value",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.prop.IsConfigured()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
