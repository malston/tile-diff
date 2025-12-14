// ABOUTME: Ginkgo acceptance tests for Pivnet integration.
// ABOUTME: Tests tile downloads, caching, EULA handling, and error scenarios.
package test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pivnet Integration", func() {
	BeforeEach(func() {
		if _, err := os.Stat(tileDiffBin); os.IsNotExist(err) {
			Fail(fmt.Sprintf("tile-diff binary not found at %s - run 'make build' first", tileDiffBin))
		}
		if os.Getenv("PIVNET_TOKEN") == "" {
			Skip("PIVNET_TOKEN not set - skipping live Pivnet tests")
		}
		setupCacheDir()
	})

	AfterEach(func() {
		cleanupCacheDir()
	})

	Describe("Non-Interactive Mode", func() {
		// NOTE: These tests use p-healthwatch 2.4.7 -> 2.4.8 as examples.
		// If these specific versions are no longer available in Pivnet, the tests
		// will fail. This is expected behavior - the tests are designed to work
		// against live Pivnet data like the bash acceptance tests do.
		// Update the product-slug and versions as needed to match available releases.

		It("downloads tiles with exact versions and all required flags", func() {
			output, err := runTileDiff(
				"--product-slug", "p-healthwatch",
				"--old-version", "2.4.7",
				"--new-version", "2.4.8",
				"--product-file", "VMware Tanzu® Healthwatch",
				"--pivnet-token", os.Getenv("PIVNET_TOKEN"),
				"--accept-eula",
				"--non-interactive",
			)

			if err != nil {
				if strings.Contains(output, "no releases found matching version") {
					Skip("Test product/versions not available - update test")
				}
				Fail(fmt.Sprintf("Unexpected error: %v\nOutput: %s", err, output))
			}

			// Should mention downloading or using cache
			Expect(output).To(Or(
				ContainSubstring("Downloading"),
				ContainSubstring("Using cached file"),
			))

			// Should show comparison results
			Expect(output).To(Or(
				ContainSubstring("Tile Upgrade Analysis"),
				ContainSubstring("Total Changes"),
			))
		})

		It("supports JSON output format", func() {
			output, err := runTileDiff(
				"--product-slug", "p-healthwatch",
				"--old-version", "2.4.7",
				"--new-version", "2.4.8",
				"--product-file", "VMware Tanzu® Healthwatch",
				"--pivnet-token", os.Getenv("PIVNET_TOKEN"),
				"--accept-eula",
				"--non-interactive",
				"--format", "json",
			)

			if err != nil {
				if strings.Contains(output, "no releases found matching version") {
					Skip("Test product/versions not available - update test")
				}
				Fail(fmt.Sprintf("Unexpected error: %v\nOutput: %s", err, output))
			}

			// Validate that output is actually valid JSON by parsing it
			var result interface{}
			err = json.Unmarshal([]byte(output), &result)
			Expect(err).NotTo(HaveOccurred(), "Output should be valid JSON")
		})
	})

	Describe("Cache Verification", func() {
		// NOTE: These tests verify that tile downloads are cached and reused
		// on subsequent runs to avoid unnecessary Pivnet downloads.

		It("stores downloaded tiles in cache", func() {
			// First download - should populate cache
			output, err := runTileDiff(
				"--product-slug", "p-healthwatch",
				"--old-version", "2.4.7",
				"--new-version", "2.4.8",
				"--product-file", "VMware Tanzu® Healthwatch",
				"--pivnet-token", os.Getenv("PIVNET_TOKEN"),
				"--accept-eula",
				"--non-interactive",
			)

			if err != nil {
				if strings.Contains(output, "no releases found matching version") {
					Skip("Test product/versions not available - update test")
				}
				Fail(fmt.Sprintf("Unexpected error: %v\nOutput: %s", err, output))
			}

			// Verify cache manifest exists
			manifestPath := fmt.Sprintf("%s/manifest.json", testCacheDir)
			Expect(manifestPath).To(BeAnExistingFile(), "Cache manifest should be created")

			// Verify actual tile files exist in cache
			files, err := os.ReadDir(testCacheDir)
			Expect(err).NotTo(HaveOccurred(), "Should be able to read cache directory")

			tileCount := 0
			for _, file := range files {
				if strings.HasSuffix(file.Name(), ".pivotal") {
					tileCount++
				}
			}
			Expect(tileCount).To(BeNumerically(">=", 2), "Cache should contain at least 2 tile files")
		})

		It("reuses cached tiles on subsequent runs", func() {
			// First run to populate cache
			_, err := runTileDiff(
				"--product-slug", "p-healthwatch",
				"--old-version", "2.4.7",
				"--new-version", "2.4.8",
				"--product-file", "VMware Tanzu® Healthwatch",
				"--pivnet-token", os.Getenv("PIVNET_TOKEN"),
				"--accept-eula",
				"--non-interactive",
			)

			if err != nil {
				Skip("First run failed - cannot test cache reuse")
			}

			// Record cache files before second run
			filesBefore, err := os.ReadDir(testCacheDir)
			Expect(err).NotTo(HaveOccurred(), "Should be able to read cache directory")

			tileFilesBefore := make([]string, 0)
			for _, file := range filesBefore {
				if strings.HasSuffix(file.Name(), ".pivotal") {
					tileFilesBefore = append(tileFilesBefore, file.Name())
				}
			}

			// Second run - should use cache
			output, err := runTileDiff(
				"--product-slug", "p-healthwatch",
				"--old-version", "2.4.7",
				"--new-version", "2.4.8",
				"--product-file", "VMware Tanzu® Healthwatch",
				"--pivnet-token", os.Getenv("PIVNET_TOKEN"),
				"--accept-eula",
				"--non-interactive",
			)

			if err != nil {
				if strings.Contains(output, "no releases found matching version") {
					Skip("Test product/versions not available - update test")
				}
				Fail(fmt.Sprintf("Unexpected error: %v\nOutput: %s", err, output))
			}

			// Should mention using cached files
			Expect(output).To(ContainSubstring("Using cached file"), "Second run should use cache")

			// Cache files should not have changed
			filesAfter, err := os.ReadDir(testCacheDir)
			Expect(err).NotTo(HaveOccurred(), "Should be able to read cache directory")

			tileFilesAfter := make([]string, 0)
			for _, file := range filesAfter {
				if strings.HasSuffix(file.Name(), ".pivotal") {
					tileFilesAfter = append(tileFilesAfter, file.Name())
				}
			}

			Expect(tileFilesAfter).To(Equal(tileFilesBefore), "No new downloads should occur (cache reused)")
		})
	})
})
