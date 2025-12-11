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

var productNameMapping = map[string]string{
	"tanzu application service": "cf",
	"tas":                        "cf",
	"cf":                         "cf",
	"mysql":                      "p-mysql",
	"p-mysql":                    "p-mysql",
	"rabbitmq":                   "p-rabbitmq",
	"p-rabbitmq":                 "p-rabbitmq",
}

// IdentifyProduct extracts and normalizes product ID from tile metadata
func IdentifyProduct(metadata map[string]interface{}) string {
	// Try "name" field first
	if name, ok := metadata["name"].(string); ok {
		normalized := strings.ToLower(strings.TrimSpace(name))
		if productID, found := productNameMapping[normalized]; found {
			return productID
		}
		return normalized
	}

	// Try "product_name" field
	if name, ok := metadata["product_name"].(string); ok {
		normalized := strings.ToLower(strings.TrimSpace(name))
		if productID, found := productNameMapping[normalized]; found {
			return productID
		}
		return normalized
	}

	return ""
}
