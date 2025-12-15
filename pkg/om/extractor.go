// ABOUTME: Runs om config-template to extract configuration from tiles.
// ABOUTME: Generates configuration templates without needing deployment.
package om

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// ExtractConfig runs om config-template on a tile and returns the parsed product config
func ExtractConfig(tilePath string) (*ProductConfig, error) {
	// Create temporary directory for extraction
	tempDir, err := os.MkdirTemp("", "tile-diff-om-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Run om config-template
	cmd := exec.Command("om", "config-template",
		"--product-path", tilePath,
		"--output-directory", tempDir,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("om config-template failed: %w\nOutput: %s", err, string(output))
	}

	// Find product.yml in the nested directory structure
	// om config-template creates: <output-dir>/<product-name>/<version>/product.yml
	productYAMLPath, err := findProductYAML(tempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to find product.yml: %w", err)
	}

	data, err := os.ReadFile(productYAMLPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read product.yml: %w", err)
	}

	var config ProductConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse product.yml: %w", err)
	}

	return &config, nil
}

// findProductYAML searches for product.yml in the nested directory structure
// created by om config-template: <output-dir>/<product-name>/<version>/product.yml
func findProductYAML(baseDir string) (string, error) {
	var productYAMLPath string

	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() == "product.yml" {
			productYAMLPath = path
			return filepath.SkipAll // Found it, stop walking
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	if productYAMLPath == "" {
		return "", fmt.Errorf("product.yml not found in %s", baseDir)
	}

	return productYAMLPath, nil
}

// CheckOMAvailable checks if om CLI is available
func CheckOMAvailable() error {
	cmd := exec.Command("om", "version")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("om CLI not found - please install from https://github.com/pivotal-cf/om")
	}
	return nil
}
