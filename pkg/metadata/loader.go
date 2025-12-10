// ABOUTME: High-level loader for tile metadata from .pivotal files.
// ABOUTME: Combines ZIP extraction and YAML parsing into single operation.
package metadata

import (
	"fmt"
	"os"
)

// LoadFromFile loads and parses metadata from a .pivotal file
func LoadFromFile(path string) (*TileMetadata, error) {
	// Read the .pivotal file
	zipData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	// Extract metadata.yml
	yamlData, err := ExtractMetadata(zipData)
	if err != nil {
		return nil, fmt.Errorf("failed to extract metadata from %s: %w", path, err)
	}

	// Parse YAML
	metadata, err := ParseMetadata(yamlData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse metadata from %s: %w", path, err)
	}

	return metadata, nil
}
