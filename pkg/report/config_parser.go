// ABOUTME: Parses current Ops Manager configuration from API responses.
// ABOUTME: Extracts configured properties for cross-reference with comparison results.
package report

import (
	"strings"

	"github.com/malston/tile-diff/pkg/api"
)

// ConfiguredProperty represents a property from current Ops Manager configuration
type ConfiguredProperty struct {
	Name         string
	Type         string
	Configurable bool
	Credential   bool
	Optional     bool
	Value        interface{}
}

// CurrentConfig represents the current Ops Manager configuration
type CurrentConfig struct {
	Properties map[string]ConfiguredProperty
}

// ParseCurrentConfig converts API response to CurrentConfig
func ParseCurrentConfig(apiResponse *api.PropertiesResponse) *CurrentConfig {
	config := &CurrentConfig{
		Properties: make(map[string]ConfiguredProperty),
	}

	for fullPath, prop := range apiResponse.Properties {
		// Extract property name from full path (e.g., ".properties.name" -> "name")
		name := extractPropertyName(fullPath)

		config.Properties[name] = ConfiguredProperty{
			Name:         name,
			Type:         prop.Type,
			Configurable: prop.Configurable,
			Credential:   prop.Credential,
			Optional:     prop.Optional,
			Value:        prop.Value,
		}
	}

	return config
}

// IsConfigured returns true if the property is actually configured (not just present)
func (p *ConfiguredProperty) IsConfigured() bool {
	// Non-configurable properties don't count
	if !p.Configurable {
		return false
	}

	// Optional properties with nil values are not configured
	if p.Optional && p.Value == nil {
		return false
	}

	return true
}

// extractPropertyName extracts the property name from a full API path
func extractPropertyName(fullPath string) string {
	// Remove common prefixes
	name := strings.TrimPrefix(fullPath, ".properties.")
	name = strings.TrimPrefix(name, ".cloud_controller.")
	name = strings.TrimPrefix(name, ".diego_brain.")
	name = strings.TrimPrefix(name, ".mysql.")
	name = strings.TrimPrefix(name, ".")

	return name
}
