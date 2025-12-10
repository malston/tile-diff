// ABOUTME: Unit tests for filtering changes by relevance to current config.
// ABOUTME: Validates which property changes affect the deployed configuration.
package report

import (
	"testing"

	"github.com/malston/tile-diff/pkg/compare"
	"github.com/malston/tile-diff/pkg/metadata"
)

func TestFilterRelevantChanges(t *testing.T) {
	currentConfig := &CurrentConfig{
		Properties: map[string]ConfiguredProperty{
			"configured_prop": {
				Name:         "configured_prop",
				Configurable: true,
				Value:        "value",
			},
			"default_prop": {
				Name:         "default_prop",
				Configurable: true,
				Optional:     true,
				Value:        nil, // Using default
			},
		},
	}

	oldProp := metadata.PropertyBlueprint{Name: "configured_prop", Type: "string"}
	newProp := metadata.PropertyBlueprint{Name: "configured_prop", Type: "integer"}

	allChanges := &compare.ComparisonResults{
		Added: []compare.ComparisonResult{
			{PropertyName: "new_prop", ChangeType: compare.PropertyAdded},
		},
		Removed: []compare.ComparisonResult{
			{PropertyName: "configured_prop", ChangeType: compare.PropertyRemoved, OldProperty: &oldProp},
			{PropertyName: "unused_prop", ChangeType: compare.PropertyRemoved},
		},
		Changed: []compare.ComparisonResult{
			{PropertyName: "configured_prop", ChangeType: compare.TypeChanged, OldProperty: &oldProp, NewProperty: &newProp},
		},
	}

	filtered := FilterRelevantChanges(allChanges, currentConfig)

	// New properties are always relevant
	if len(filtered.Added) != 1 {
		t.Errorf("Expected 1 added property, got %d", len(filtered.Added))
	}

	// Only configured properties that are removed are relevant
	if len(filtered.Removed) != 1 {
		t.Errorf("Expected 1 relevant removed property, got %d", len(filtered.Removed))
	}
	if filtered.Removed[0].PropertyName != "configured_prop" {
		t.Errorf("Expected configured_prop, got %s", filtered.Removed[0].PropertyName)
	}

	// Changed properties affecting configured values are relevant
	if len(filtered.Changed) != 1 {
		t.Errorf("Expected 1 relevant changed property, got %d", len(filtered.Changed))
	}
}

func TestIsChangeRelevant(t *testing.T) {
	tests := []struct {
		name          string
		changeType    compare.ChangeType
		propertyName  string
		isConfigured  bool
		expected      bool
	}{
		{
			name:         "added property always relevant",
			changeType:   compare.PropertyAdded,
			propertyName: "new_prop",
			isConfigured: false,
			expected:     true,
		},
		{
			name:         "removed configured property relevant",
			changeType:   compare.PropertyRemoved,
			propertyName: "configured_prop",
			isConfigured: true,
			expected:     true,
		},
		{
			name:         "removed unconfigured property not relevant",
			changeType:   compare.PropertyRemoved,
			propertyName: "unused_prop",
			isConfigured: false,
			expected:     false,
		},
		{
			name:         "changed configured property relevant",
			changeType:   compare.TypeChanged,
			propertyName: "configured_prop",
			isConfigured: true,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &CurrentConfig{
				Properties: make(map[string]ConfiguredProperty),
			}
			if tt.isConfigured {
				config.Properties[tt.propertyName] = ConfiguredProperty{
					Name:         tt.propertyName,
					Configurable: true,
					Value:        "value",
				}
			}

			result := isChangeRelevant(tt.changeType, tt.propertyName, config)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
