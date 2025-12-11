// ABOUTME: Main entry point for tile-diff CLI tool.
// ABOUTME: Handles command-line arguments and orchestrates comparison workflow.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/malston/tile-diff/pkg/api"
	"github.com/malston/tile-diff/pkg/compare"
	"github.com/malston/tile-diff/pkg/metadata"
	"github.com/malston/tile-diff/pkg/releasenotes"
	"github.com/malston/tile-diff/pkg/report"
)

// enrichWithReleaseNotes orchestrates the release notes enrichment process
func enrichWithReleaseNotes(
	comparison *compare.ComparisonResults,
	newVersion string,
	productID string,
	config releasenotes.ProductConfig,
	urlOverride string,
) (map[string]releasenotes.Match, error) {

	// Resolve URL
	var url string
	var err error
	if urlOverride != "" {
		url = urlOverride
	} else {
		url, err = config.ResolveURL(productID, newVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve URL: %w", err)
		}
	}

	// Fetch release notes
	fetcher := releasenotes.NewFetcher()
	html, err := fetcher.Fetch(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release notes: %w", err)
	}

	// Parse features
	features, err := releasenotes.ParseHTML(html)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Collect property names
	var properties []string
	for _, change := range comparison.Added {
		properties = append(properties, change.PropertyName)
	}

	// Match properties to features
	matcher := releasenotes.NewMatcher(features)
	matches := matcher.Match(properties)

	return matches, nil
}

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

	// Release notes enrichment flags
	skipReleaseNotes := flag.Bool("skip-release-notes", false, "Skip release notes enrichment")
	releaseNotesURL := flag.String("release-notes-url", "", "Override release notes URL")
	productID := flag.String("product-id", "", "Override product ID detection")
	productConfig := flag.String("product-config", "configs/products.yaml", "Path to product config file")
	verbose := flag.Bool("verbose", false, "Enable verbose output")

	flag.Parse()

	// Validate required flags
	if *oldTile == "" || *newTile == "" {
		fmt.Fprintf(os.Stderr, "Error: --old-tile and --new-tile are required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	fmt.Printf("tile-diff - Ops Manager Product Tile Comparison\n")
	fmt.Printf("================================================\n\n")

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

	// Compare metadata
	fmt.Printf("\nComparing tiles...\n")
	results := compare.CompareMetadata(oldMetadata, newMetadata, true)

	// Try release notes enrichment
	var matches map[string]releasenotes.Match
	if !*skipReleaseNotes {
		if *verbose {
			fmt.Printf("\nAttempting release notes enrichment...\n")
		}

		// Load product config
		config, err := releasenotes.LoadProductConfig(*productConfig)
		if err != nil {
			if *verbose {
				fmt.Fprintf(os.Stderr, "Warning: Could not load product config: %v\n", err)
				fmt.Fprintf(os.Stderr, "Continuing with standard report...\n\n")
			}
		} else {
			// Determine product ID
			prodID := *productID
			if prodID == "" {
				// Default to "cf" for now - could be enhanced to auto-detect from metadata
				prodID = "cf"
			}

			// Extract version from new tile filename
			// For now, use a simple approach - this could be enhanced
			newVersion := "10.2.5" // Default version - should be extracted from metadata or filename

			// Try to enrich
			matches, err = enrichWithReleaseNotes(results, newVersion, prodID, config, *releaseNotesURL)
			if err != nil {
				if *verbose {
					fmt.Fprintf(os.Stderr, "Warning: Release notes enrichment failed: %v\n", err)
					fmt.Fprintf(os.Stderr, "Continuing with standard report...\n\n")
				}
			} else {
				if *verbose {
					fmt.Printf("Enriched with %d property matches\n\n", len(matches))
				}
			}
		}
	}

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
			// Use enriched report if we have matches
			var textReport string
			if len(matches) > 0 {
				enriched := report.EnrichChanges(categorized, matches)
				textReport = report.GenerateTextReportWithFeatures(enriched, *oldTile, *newTile)
			} else {
				textReport = report.GenerateTextReport(categorized, *oldTile, *newTile)
			}
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
