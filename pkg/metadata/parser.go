// ABOUTME: Parses YAML metadata into Go structures.
// ABOUTME: Handles property blueprints and nested selector options.
package metadata

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// ParseMetadata parses YAML metadata content into structured format
func ParseMetadata(yamlData []byte) (*TileMetadata, error) {
	var metadata TileMetadata
	err := yaml.Unmarshal(yamlData, &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to parse metadata YAML: %w", err)
	}

	return &metadata, nil
}
