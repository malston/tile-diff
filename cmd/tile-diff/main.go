// ABOUTME: Main entry point for tile-diff CLI tool.
// ABOUTME: Handles command-line arguments and orchestrates comparison workflow.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/malston/tile-diff/pkg/api"
	"github.com/malston/tile-diff/pkg/compare"
	"github.com/malston/tile-diff/pkg/metadata"
	"github.com/malston/tile-diff/pkg/pivnet"
	"github.com/malston/tile-diff/pkg/report"
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
	reportFormat := flag.String("format", "text", "Output format: text or json")

	// Pivnet-related flags
	productSlug := flag.String("product-slug", "", "Pivnet product slug (e.g., 'cf')")
	oldVersion := flag.String("old-version", "", "Old release version")
	newVersion := flag.String("new-version", "", "New release version")
	pivnetToken := flag.String("pivnet-token", "", "Pivnet API token (or use PIVNET_TOKEN env var)")
	productFile := flag.String("product-file", "", "Specific product file name (optional)")
	acceptEULA := flag.Bool("accept-eula", false, "Accept EULAs without prompting")
	nonInteractive := flag.Bool("non-interactive", false, "Fail instead of prompting for input")
	cacheDir := flag.String("cache-dir", "", "Download cache directory (default: ~/.tile-diff/cache)")

	flag.Parse()

	// Detect mode: local files or Pivnet download
	usingLocalFiles := *oldTile != "" || *newTile != ""
	usingPivnet := *productSlug != "" || *oldVersion != "" || *newVersion != ""

	if usingLocalFiles && usingPivnet {
		fmt.Fprintf(os.Stderr, "Error: Cannot mix local and Pivnet modes\n")
		fmt.Fprintf(os.Stderr, "Use either:\n")
		fmt.Fprintf(os.Stderr, "  --old-tile + --new-tile (local files)\n")
		fmt.Fprintf(os.Stderr, "OR\n")
		fmt.Fprintf(os.Stderr, "  --product-slug + --old-version + --new-version (Pivnet download)\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if !usingLocalFiles && !usingPivnet {
		fmt.Fprintf(os.Stderr, "Error: Must provide either local files or Pivnet download options\n\n")
		flag.Usage()
		os.Exit(1)
	}

	var oldTilePath, newTilePath string

	if usingPivnet {
		// Validate Pivnet flags
		if *productSlug == "" || *oldVersion == "" || *newVersion == "" {
			fmt.Fprintf(os.Stderr, "Error: Pivnet mode requires --product-slug, --old-version, and --new-version\n\n")
			flag.Usage()
			os.Exit(1)
		}

		// Get Pivnet token from flag or env var
		token := *pivnetToken
		if token == "" {
			token = os.Getenv("PIVNET_TOKEN")
		}
		if token == "" {
			fmt.Fprintf(os.Stderr, "Error: Pivnet token required\n")
			fmt.Fprintf(os.Stderr, "Provide via --pivnet-token or PIVNET_TOKEN env var\n\n")
			flag.Usage()
			os.Exit(1)
		}

		fmt.Printf("tile-diff - Ops Manager Product Tile Comparison\n")
		fmt.Printf("================================================\n\n")
		fmt.Printf("Mode: Pivnet Download\n")
		fmt.Printf("Product: %s\n", *productSlug)
		fmt.Printf("Versions: %s -> %s\n\n", *oldVersion, *newVersion)

		// Create Pivnet client
		client, err := pivnet.NewClient(token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to create Pivnet client: %v\n", err)
			os.Exit(1)
		}

		// Setup paths
		home, _ := os.UserHomeDir()
		cacheDirectory := *cacheDir
		if cacheDirectory == "" {
			cacheDirectory = filepath.Join(home, ".tile-diff", "cache")
		}
		manifestFile := filepath.Join(home, ".tile-diff", "cache-manifest.json")
		eulaFile := filepath.Join(home, ".tile-diff", "eulas.json")

		// Create downloader
		downloader := pivnet.NewDownloader(client, cacheDirectory, manifestFile, eulaFile, 20)

		// Download old tile
		fmt.Printf("Resolving and downloading old tile (%s)...\n", *oldVersion)
		oldOpts := pivnet.DownloadOptions{
			ProductSlug:    *productSlug,
			Version:        *oldVersion,
			ProductFile:    *productFile,
			AcceptEULA:     *acceptEULA,
			NonInteractive: *nonInteractive,
			CacheDir:       cacheDirectory,
		}
		oldTilePath, err = downloader.Download(oldOpts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error downloading old tile: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Old tile: %s\n\n", oldTilePath)

		// Download new tile
		fmt.Printf("Resolving and downloading new tile (%s)...\n", *newVersion)
		newOpts := pivnet.DownloadOptions{
			ProductSlug:    *productSlug,
			Version:        *newVersion,
			ProductFile:    *productFile,
			AcceptEULA:     *acceptEULA,
			NonInteractive: *nonInteractive,
			CacheDir:       cacheDirectory,
		}
		newTilePath, err = downloader.Download(newOpts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error downloading new tile: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ New tile: %s\n\n", newTilePath)

	} else {
		// Local files mode
		if *oldTile == "" || *newTile == "" {
			fmt.Fprintf(os.Stderr, "Error: Local mode requires both --old-tile and --new-tile\n\n")
			flag.Usage()
			os.Exit(1)
		}
		oldTilePath = *oldTile
		newTilePath = *newTile

		fmt.Printf("tile-diff - Ops Manager Product Tile Comparison\n")
		fmt.Printf("================================================\n\n")
	}

	// Load old tile metadata
	fmt.Printf("Loading old tile: %s\n", oldTilePath)
	oldMetadata, err := metadata.LoadFromFile(oldTilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading old tile: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  Found %d properties\n", len(oldMetadata.PropertyBlueprints))

	// Load new tile metadata
	fmt.Printf("Loading new tile: %s\n", newTilePath)
	newMetadata, err := metadata.LoadFromFile(newTilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading new tile: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  Found %d properties\n", len(newMetadata.PropertyBlueprints))

	// Compare metadata
	fmt.Printf("\nComparing tiles...\n")
	results := compare.CompareMetadata(oldMetadata, newMetadata, true)

	fmt.Printf("\nComparison Results:\n")
	fmt.Printf("===================\n\n")

	// Display added properties
	if len(results.Added) > 0 {
		fmt.Printf("✨ New Properties (%d):\n", len(results.Added))
		for _, result := range results.Added {
			fmt.Printf("  + %s (%s)\n", result.PropertyName, result.NewProperty.Type)
		}
		fmt.Println()
	}

	// Display removed properties
	if len(results.Removed) > 0 {
		fmt.Printf("🗑️  Removed Properties (%d):\n", len(results.Removed))
		for _, result := range results.Removed {
			fmt.Printf("  - %s (%s)\n", result.PropertyName, result.OldProperty.Type)
		}
		fmt.Println()
	}

	// Display changed properties
	if len(results.Changed) > 0 {
		fmt.Printf("🔄 Changed Properties (%d):\n", len(results.Changed))
		for _, result := range results.Changed {
			fmt.Printf("  ~ %s: %s\n", result.PropertyName, result.Description)
		}
		fmt.Println()
	}

	// Summary
	fmt.Printf("Summary:\n")
	fmt.Printf("  Properties in old tile: %d\n", results.TotalOldProps)
	fmt.Printf("  Properties in new tile: %d\n", results.TotalNewProps)
	fmt.Printf("  Added: %d, Removed: %d, Changed: %d\n",
		len(results.Added), len(results.Removed), len(results.Changed))

	// Count configurable properties
	oldConfigurable := countConfigurable(oldMetadata.PropertyBlueprints)
	newConfigurable := countConfigurable(newMetadata.PropertyBlueprints)

	fmt.Printf("\nConfigurable properties:\n")
	fmt.Printf("  Old tile: %d\n", oldConfigurable)
	fmt.Printf("  New tile: %d\n", newConfigurable)

	// Load current configuration if API credentials provided
	if *productGUID != "" && *opsManagerURL != "" && *username != "" && *password != "" {
		fmt.Printf("\nQuerying Ops Manager API...\n")
		client := api.NewClient(*opsManagerURL, *username, *password, *skipSSL)

		properties, err := client.GetProperties(*productGUID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching properties from Ops Manager: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("  Found %d total properties\n", len(properties.Properties))

		// Count configurable properties in current config
		currentConfigurable := 0
		for _, prop := range properties.Properties {
			if prop.Configurable {
				currentConfigurable++
			}
		}
		fmt.Printf("  Configurable: %d\n", currentConfigurable)

		// Count properties with non-default values (approximate)
		configured := 0
		for _, prop := range properties.Properties {
			if prop.Configurable && prop.Value != nil {
				configured++
			}
		}
		fmt.Printf("  Currently configured: ~%d\n", configured)

		// Generate actionable report
		fmt.Printf("\nGenerating actionable report...\n")

		// Parse current config
		currentConfig := report.ParseCurrentConfig(properties)

		// Filter relevant changes
		filtered := report.FilterRelevantChanges(results, currentConfig)

		// Categorize changes
		categorized := report.CategorizeChanges(filtered)

		// Generate report based on format
		fmt.Println()
		switch *reportFormat {
		case "json":
			jsonReport := report.GenerateJSONReport(categorized, *oldTile, *newTile)
			fmt.Println(jsonReport)
		default:
			textReport := report.GenerateTextReport(categorized, *oldTile, *newTile)
			fmt.Println(textReport)
		}
	} else if *productGUID != "" {
		fmt.Printf("\nSkipping Ops Manager API (credentials not provided)\n")
		fmt.Printf("To include current configuration, provide:\n")
		fmt.Printf("  --ops-manager-url, --username, --password\n")
	} else {
		fmt.Println("\nNote: Provide Ops Manager credentials for actionable report with current config analysis")
	}
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
