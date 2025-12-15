// ABOUTME: Main entry point for tile-diff CLI tool.
// ABOUTME: Handles command-line arguments and orchestrates comparison workflow.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/malston/tile-diff/pkg/api"
	"github.com/malston/tile-diff/pkg/compare"
	"github.com/malston/tile-diff/pkg/metadata"
	"github.com/malston/tile-diff/pkg/om"
	"github.com/malston/tile-diff/pkg/pivnet"
	"github.com/malston/tile-diff/pkg/releasenotes"
	"github.com/malston/tile-diff/pkg/report"
)

// EnrichmentResult contains the results of release notes enrichment
type EnrichmentResult struct {
	Matches    map[string]releasenotes.Match
	Features   []releasenotes.Feature
	Properties []string
}

// printMatchingDebugInfo outputs detailed matching information
func printMatchingDebugInfo(result *EnrichmentResult) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("PROPERTY-TO-FEATURE MATCHING DEBUG")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Printf("\nTotal Features Found: %d\n", len(result.Features))
	fmt.Printf("Total Properties to Match: %d\n", len(result.Properties))
	fmt.Printf("Successful Matches: %d\n", len(result.Matches))
	fmt.Printf("Unmatched Properties: %d\n\n", len(result.Properties)-len(result.Matches))

	// Show all features
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println("FEATURES EXTRACTED FROM RELEASE NOTES")
	fmt.Println(strings.Repeat("-", 80))
	for i, feature := range result.Features {
		fmt.Printf("\n[%d] %s\n", i+1, feature.Title)
		// Show truncated description
		desc := feature.Description
		if len(desc) > 150 {
			desc = desc[:150] + "..."
		}
		fmt.Printf("    Description: %s\n", desc)
	}

	// Show matched properties
	if len(result.Matches) > 0 {
		fmt.Println("\n" + strings.Repeat("-", 80))
		fmt.Println("MATCHED PROPERTIES")
		fmt.Println(strings.Repeat("-", 80))

		// Group by feature for easier reading
		featureMatches := make(map[string][]releasenotes.Match)
		for _, match := range result.Matches {
			featureMatches[match.Feature.Title] = append(featureMatches[match.Feature.Title], match)
		}

		for featureName, matches := range featureMatches {
			fmt.Printf("\n📦 %s (%d properties)\n", featureName, len(matches))
			for _, match := range matches {
				fmt.Printf("   ✓ %s\n", match.Property)
				fmt.Printf("      Match Type: %s\n", match.MatchType)
				fmt.Printf("      Confidence: %.2f\n", match.Confidence)
			}
		}
	}

	// Show unmatched properties
	unmatched := findUnmatchedProperties(result.Properties, result.Matches)
	if len(unmatched) > 0 {
		fmt.Println("\n" + strings.Repeat("-", 80))
		fmt.Println("UNMATCHED PROPERTIES")
		fmt.Println(strings.Repeat("-", 80))
		fmt.Println("\nThese properties could not be matched to any release note feature:")
		for _, prop := range unmatched {
			fmt.Printf("   ✗ %s\n", prop)
		}
		fmt.Println("\nTip: Unmatched properties may indicate:")
		fmt.Println("  - Property names don't appear in release notes")
		fmt.Println("  - Keyword matching threshold (0.5) not met")
		fmt.Println("  - Property is an internal implementation detail")
	}

	fmt.Println("\n" + strings.Repeat("=", 80) + "\n")
}

// findUnmatchedProperties returns properties that weren't matched
func findUnmatchedProperties(properties []string, matches map[string]releasenotes.Match) []string {
	var unmatched []string
	for _, prop := range properties {
		if _, found := matches[prop]; !found {
			unmatched = append(unmatched, prop)
		}
	}
	return unmatched
}

// enrichWithReleaseNotes orchestrates the release notes enrichment process
func enrichWithReleaseNotes(
	comparison *compare.ComparisonResults,
	newVersion string,
	productID string,
	config releasenotes.ProductConfig,
	urlOverride string,
) (*EnrichmentResult, error) {

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

	return &EnrichmentResult{
		Matches:    matches,
		Features:   features,
		Properties: properties,
	}, nil
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

	// Pivnet-related flags
	productSlug := flag.String("product-slug", "", "Pivnet product slug (e.g., 'cf')")
	oldVersion := flag.String("old-version", "", "Old release version")
	newVersion := flag.String("new-version", "", "New release version")
	pivnetToken := flag.String("pivnet-token", "", "Pivnet API token (or use PIVNET_TOKEN env var)")
	productFile := flag.String("product-file", "", "Specific product file name (optional)")
	acceptEULA := flag.Bool("accept-eula", false, "Accept EULAs without prompting")
	nonInteractive := flag.Bool("non-interactive", false, "Fail instead of prompting for input")
	cacheDir := flag.String("cache-dir", "", "Download cache directory (default: ~/.tile-diff/cache)")

	// Release notes enrichment flags
	skipReleaseNotes := flag.Bool("skip-release-notes", false, "Skip release notes enrichment")
	releaseNotesURL := flag.String("release-notes-url", "", "Override release notes URL")
	productID := flag.String("product-id", "", "Override product ID detection")
	productConfig := flag.String("product-config", "configs/products.yaml", "Path to product config file")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	debugMatching := flag.Bool("debug-matching", false, "Show detailed property-to-feature matching information")

	flag.Parse()

	// Track if we're in JSON mode to suppress non-JSON output
	jsonMode := *reportFormat == "json"

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

		if *reportFormat != "json" {
			fmt.Printf("tile-diff - Ops Manager Product Tile Comparison\n")
			fmt.Printf("================================================\n\n")
			fmt.Printf("Mode: Pivnet Download\n")
			fmt.Printf("Product: %s\n", *productSlug)
			fmt.Printf("Versions: %s -> %s\n\n", *oldVersion, *newVersion)
		}

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
			// Check environment variable before using default
			cacheDirectory = os.Getenv("PIVNET_CACHE_DIR")
			if cacheDirectory == "" {
				cacheDirectory = filepath.Join(home, ".tile-diff", "cache")
			}
		}
		manifestFile := filepath.Join(cacheDirectory, "manifest.json")

		// Check environment variable for EULA file, otherwise place in cache dir
		eulaFile := os.Getenv("PIVNET_EULA_FILE")
		if eulaFile == "" {
			eulaFile = filepath.Join(cacheDirectory, "eula_acceptance.json")
		}

		// Create downloader (quiet mode in JSON to suppress progress output)
		downloader := pivnet.NewDownloader(client, cacheDirectory, manifestFile, eulaFile, 20, jsonMode)

		// Download old tile
		if !jsonMode {
			fmt.Printf("Resolving and downloading old tile (%s)...\n", *oldVersion)
		}
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
		if !jsonMode {
			fmt.Printf("✓ Old tile: %s\n\n", oldTilePath)
		}

		// Download new tile
		if !jsonMode {
			fmt.Printf("Resolving and downloading new tile (%s)...\n", *newVersion)
		}
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
		if !jsonMode {
			fmt.Printf("✓ New tile: %s\n\n", newTilePath)
		}

	} else {
		// Local files mode
		if *oldTile == "" || *newTile == "" {
			fmt.Fprintf(os.Stderr, "Error: Local mode requires both --old-tile and --new-tile\n\n")
			flag.Usage()
			os.Exit(1)
		}
		oldTilePath = *oldTile
		newTilePath = *newTile

		if *reportFormat != "json" {
			fmt.Printf("tile-diff - Ops Manager Product Tile Comparison\n")
			fmt.Printf("================================================\n\n")
		}
	}

	// Load old tile metadata
	if !jsonMode {
		fmt.Printf("Loading old tile: %s\n", oldTilePath)
	}
	oldMetadata, err := metadata.LoadFromFile(oldTilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading old tile: %v\n", err)
		os.Exit(1)
	}
	if !jsonMode {
		fmt.Printf("  Found %d properties\n", len(oldMetadata.PropertyBlueprints))
	}

	// Load new tile metadata
	if !jsonMode {
		fmt.Printf("Loading new tile: %s\n", newTilePath)
	}
	newMetadata, err := metadata.LoadFromFile(newTilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading new tile: %v\n", err)
		os.Exit(1)
	}
	if !jsonMode {
		fmt.Printf("  Found %d properties\n", len(newMetadata.PropertyBlueprints))
	}

	// Compare metadata
	if !jsonMode {
		fmt.Printf("\nComparing tiles...\n")
	}
	results := compare.CompareMetadata(oldMetadata, newMetadata, true)

	// Extract and compare configuration templates using om config-template
	var configComparison *om.ConfigComparison
	if err := om.CheckOMAvailable(); err == nil {
		if !jsonMode {
			fmt.Printf("\nExtracting configuration templates...\n")
		}

		oldConfig, err := om.ExtractConfig(oldTilePath)
		if err != nil {
			if *verbose {
				fmt.Fprintf(os.Stderr, "Warning: Failed to extract config from old tile: %v\n", err)
			}
		}

		newConfig, err := om.ExtractConfig(newTilePath)
		if err != nil {
			if *verbose {
				fmt.Fprintf(os.Stderr, "Warning: Failed to extract config from new tile: %v\n", err)
			}
		}

		if oldConfig != nil && newConfig != nil {
			configComparison = om.CompareConfigs(oldConfig, newConfig)
			if !jsonMode {
				fmt.Printf("  Found %d configuration changes\n",
					len(configComparison.Added)+len(configComparison.Removed)+len(configComparison.Changed))
			}
		}
	} else if *verbose {
		fmt.Fprintf(os.Stderr, "Note: om CLI not available - skipping config template comparison\n")
		fmt.Fprintf(os.Stderr, "Install from: https://github.com/pivotal-cf/om\n")
	}

	// Try release notes enrichment
	var matches map[string]releasenotes.Match
	var enrichmentResult *EnrichmentResult
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
			enrichmentResult, err = enrichWithReleaseNotes(results, newVersion, prodID, config, *releaseNotesURL)
			if err != nil {
				if *verbose {
					fmt.Fprintf(os.Stderr, "Warning: Release notes enrichment failed: %v\n", err)
					fmt.Fprintf(os.Stderr, "Continuing with standard report...\n\n")
				}
			} else {
				matches = enrichmentResult.Matches
				if *verbose {
					fmt.Printf("Enriched with %d property matches\n\n", len(matches))
				}

				// Print debug matching info if requested
				if *debugMatching {
					printMatchingDebugInfo(enrichmentResult)
				}
			}
		}
	}

	// Only show comparison results in text mode (not JSON)
	if !jsonMode {
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

		// Display configuration changes from om config-template
		if configComparison != nil {
			fmt.Printf("\nConfiguration Changes (from om config-template):\n")
			fmt.Printf("=================================================\n\n")

			// Display added config properties
			if len(configComparison.Added) > 0 {
				fmt.Printf("✨ New Configuration Properties (%d):\n", len(configComparison.Added))
				for _, change := range configComparison.Added {
					fmt.Printf("  + %s\n", change.PropertyName)
				}
				fmt.Println()
			}

			// Display removed config properties
			if len(configComparison.Removed) > 0 {
				fmt.Printf("🗑️  Removed Configuration Properties (%d):\n", len(configComparison.Removed))
				for _, change := range configComparison.Removed {
					fmt.Printf("  - %s\n", change.PropertyName)
				}
				fmt.Println()
			}

			// Display changed config properties
			if len(configComparison.Changed) > 0 {
				fmt.Printf("🔄 Changed Configuration Properties (%d):\n", len(configComparison.Changed))
				for _, change := range configComparison.Changed {
					fmt.Printf("  ~ %s: %s\n", change.PropertyName, change.Description)
				}
				fmt.Println()
			}

			// Config summary
			if len(configComparison.Added) == 0 && len(configComparison.Removed) == 0 && len(configComparison.Changed) == 0 {
				fmt.Printf("No configuration changes detected\n\n")
			}
		}
	}

	// Load current configuration if API credentials provided
	if *productGUID != "" && *opsManagerURL != "" && *username != "" && *password != "" {
		if !jsonMode {
			fmt.Printf("\nQuerying Ops Manager API...\n")
		}
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
