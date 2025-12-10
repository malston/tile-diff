// ABOUTME: Defines data structures for TAS tile metadata parsing.
// ABOUTME: Maps YAML property_blueprints to Go structs for comparison.
package metadata

// PropertyBlueprint represents a single property definition from tile metadata
type PropertyBlueprint struct {
	Name            string              `yaml:"name"`
	Type            string              `yaml:"type"`
	Configurable    bool                `yaml:"configurable"`
	Optional        bool                `yaml:"optional"`
	Default         interface{}         `yaml:"default,omitempty"`
	Constraints     *Constraints        `yaml:"constraints,omitempty"`
	OptionTemplates []OptionTemplate    `yaml:"option_templates,omitempty"`
}

// Constraints defines validation rules for property values
type Constraints struct {
	Min *int `yaml:"min,omitempty"`
	Max *int `yaml:"max,omitempty"`
}

// OptionTemplate represents a selector option with nested properties
type OptionTemplate struct {
	Name               string              `yaml:"name"`
	SelectValue        string              `yaml:"select_value"`
	PropertyBlueprints []PropertyBlueprint `yaml:"property_blueprints,omitempty"`
}

// TileMetadata represents the top-level metadata structure
type TileMetadata struct {
	PropertyBlueprints []PropertyBlueprint `yaml:"property_blueprints"`
}
