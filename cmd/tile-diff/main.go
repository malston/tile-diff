// ABOUTME: Main entry point for tile-diff CLI tool.
// ABOUTME: Handles command-line arguments and orchestrates comparison workflow.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/malston/tile-diff/pkg/metadata"
)

func main() {
	// Define flags
	oldTile := flag.String("old-tile", "", "Path to old .pivotal file (required)")
	newTile := flag.String("new-tile", "", "Path to new .pivotal file (required)")
	productGUID := flag.String("product-guid", "", "Product GUID in Ops Manager (optional)")
	opsManagerURL := flag.String("ops-manager-url", "", "Ops Manager URL (optional)")
	username := flag.String("username", "", "Ops Manager username (optional)")
	password := flag.String("password", "", "Ops Manager password (optional)")
	skipSSL := flag.Bool("skip-ssl-validation", false, "Skip SSL certificate validation")

	flag.Parse()

	// Validate required flags
	if *oldTile == "" || *newTile == "" {
		fmt.Fprintf(os.Stderr, "Error: --old-tile and --new-tile are required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	fmt.Printf("tile-diff Phase 1 MVP\n")
	fmt.Printf("=====================\n\n")

	// Load old tile metadata
	fmt.Printf("Loading old tile: %s\n", *oldTile)
	oldMetadata, err := metadata.LoadFromFile(*oldTile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading old tile: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  Found %d properties\n", len(oldMetadata.PropertyBlueprints))

	// Load new tile metadata
	fmt.Printf("Loading new tile: %s\n", *newTile)
	newMetadata, err := metadata.LoadFromFile(*newTile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading new tile: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  Found %d properties\n", len(newMetadata.PropertyBlueprints))

	// Count configurable properties
	oldConfigurable := countConfigurable(oldMetadata.PropertyBlueprints)
	newConfigurable := countConfigurable(newMetadata.PropertyBlueprints)

	fmt.Printf("\nConfigurable properties:\n")
	fmt.Printf("  Old tile: %d\n", oldConfigurable)
	fmt.Printf("  New tile: %d\n", newConfigurable)

	// API client placeholder
	if *productGUID != "" && *opsManagerURL != "" {
		fmt.Printf("\n[Next: Add API client integration]\n")
		fmt.Printf("Product GUID: %s\n", *productGUID)
		fmt.Printf("Ops Manager: %s\n", *opsManagerURL)
		// Keep these to avoid unused variable errors
		_ = username
		_ = password
		_ = skipSSL
	}

	fmt.Printf("\nPhase 1 MVP: Data extraction complete ✓\n")
}

func countConfigurable(blueprints []metadata.PropertyBlueprint) int {
	count := 0
	for _, bp := range blueprints {
		if bp.Configurable {
			count++
		}
	}
	return count
}
