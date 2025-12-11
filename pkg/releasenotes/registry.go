// ABOUTME: Product registry for mapping product IDs to release note URLs.
// ABOUTME: Loads configuration and resolves versioned URLs for fetching.
package releasenotes

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// ProductConfig maps product IDs to release notes URL patterns
type ProductConfig map[string]string

// LoadProductConfig loads product configuration from YAML file
func LoadProductConfig(path string) (ProductConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config ProductConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return config, nil
}

// ResolveURL resolves a product ID and version to a release notes URL
func (c ProductConfig) ResolveURL(productID, version string) (string, error) {
	pattern, ok := c[productID]
	if !ok {
		return "", fmt.Errorf("product %s not found in config", productID)
	}

	url := strings.ReplaceAll(pattern, "{version}", version)
	return url, nil
}
