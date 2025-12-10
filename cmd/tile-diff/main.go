// ABOUTME: Main entry point for tile-diff CLI tool.
// ABOUTME: Handles command-line arguments and orchestrates comparison workflow.
package main

import (
	"flag"
	"fmt"
	"os"
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
	fmt.Printf("Old tile: %s\n", *oldTile)
	fmt.Printf("New tile: %s\n", *newTile)

	if *productGUID != "" {
		fmt.Printf("Product GUID: %s\n", *productGUID)
		fmt.Printf("Ops Manager: %s\n", *opsManagerURL)
		fmt.Printf("Username: %s\n", *username)
		fmt.Printf("Skip SSL: %t\n", *skipSSL)
		// Password is intentionally not printed
		_ = password
	}

	fmt.Printf("\n[Phase 1 - Next: Add metadata loading]\n")
}
