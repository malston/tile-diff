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
		// Skip download tests unless explicitly enabled
		// This prevents expensive downloads in CI and local testing
		if os.Getenv("ENABLE_DOWNLOAD_TESTS") == "" {
			Skip("ENABLE_DOWNLOAD_TESTS not set - skipping download tests (set to '1' to enable)")
		}
		setupCacheDir()
	})

	AfterEach(func() {
		cleanupCacheDir()
	})

	Describe("Non-Interactive Mode", func() {
		// NOTE: These tests use harbor-container-registry 2.10.2 -> 2.10.3 as examples.
		// If these specific versions are no longer available in Pivnet, the tests
		// will fail. This is expected behavior - the tests are designed to work
		// against live Pivnet data like the bash acceptance tests do.
		// Update the product-slug and versions as needed to match available releases.

		It("downloads tiles with exact versions and all required flags", Label("slow", "downloads"), func() {
			output, err := runTileDiff(
				"--product-slug", "harbor-container-registry",
				"--old-version", "2.10.2",
				"--new-version", "2.10.3",
				"--product-file", "VMware Harbor Container Registry for Tanzu",
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
				ContainSubstring("Comparison Results"),
				ContainSubstring("Summary"),
			))
		})

		It("supports JSON output format", Label("slow", "downloads"), func() {
			output, err := runTileDiff(
				"--product-slug", "harbor-container-registry",
				"--old-version", "2.10.2",
				"--new-version", "2.10.3",
				"--product-file", "VMware Harbor Container Registry for Tanzu",
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

		It("stores downloaded tiles in cache", Label("slow", "downloads"), func() {
			// First download - should populate cache
			output, err := runTileDiff(
				"--product-slug", "harbor-container-registry",
				"--old-version", "2.10.2",
				"--new-version", "2.10.3",
				"--product-file", "VMware Harbor Container Registry for Tanzu",
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

		It("reuses cached tiles on subsequent runs", Label("slow", "downloads"), func() {
			// First run to populate cache
			_, err := runTileDiff(
				"--product-slug", "harbor-container-registry",
				"--old-version", "2.10.2",
				"--new-version", "2.10.3",
				"--product-file", "VMware Harbor Container Registry for Tanzu",
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
				"--product-slug", "harbor-container-registry",
				"--old-version", "2.10.2",
				"--new-version", "2.10.3",
				"--product-file", "VMware Harbor Container Registry for Tanzu",
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

	Describe("EULA Handling", func() {
		// NOTE: Tests EULA acceptance requirement and persistence across runs

		It("requires EULA acceptance for first download", func() {
			// Try without --accept-eula flag (should fail in non-interactive mode)
			output, err := runTileDiff(
				"--product-slug", "harbor-container-registry",
				"--old-version", "2.10.2",
				"--new-version", "2.10.3",
				"--product-file", "VMware Harbor Container Registry for Tanzu",
				"--pivnet-token", os.Getenv("PIVNET_TOKEN"),
				"--non-interactive",
			)

			if err != nil {
				if strings.Contains(output, "no releases found matching version") {
					Skip("Test product/versions not available - update test")
				}
			}

			// Should fail
			Expect(err).To(HaveOccurred(), "Should fail without EULA acceptance")

			// Should mention EULA or --accept-eula in error
			Expect(output).To(Or(
				ContainSubstring("EULA"),
				ContainSubstring("--accept-eula"),
			), "Error should mention EULA requirement")
		})

		It("persists EULA acceptance for subsequent downloads", Label("slow", "downloads"), func() {
			// First run with --accept-eula
			output, err := runTileDiff(
				"--product-slug", "harbor-container-registry",
				"--old-version", "2.10.2",
				"--new-version", "2.10.3",
				"--product-file", "VMware Harbor Container Registry for Tanzu",
				"--pivnet-token", os.Getenv("PIVNET_TOKEN"),
				"--accept-eula",
				"--non-interactive",
			)

			if err != nil {
				if strings.Contains(output, "no releases found matching version") {
					Skip("Test product/versions not available - update test")
				}
				Fail(fmt.Sprintf("First run failed: %v\nOutput: %s", err, output))
			}

			// Second run WITHOUT --accept-eula (should work because EULA is remembered)
			output, err = runTileDiff(
				"--product-slug", "harbor-container-registry",
				"--old-version", "2.10.2",
				"--new-version", "2.10.3",
				"--product-file", "VMware Harbor Container Registry for Tanzu",
				"--pivnet-token", os.Getenv("PIVNET_TOKEN"),
				"--non-interactive",
			)

			// Should succeed even without --accept-eula
			Expect(err).NotTo(HaveOccurred(), "Second run should succeed with remembered EULA")

			// Should use cached files
			Expect(output).To(ContainSubstring("Using cached file"), "Should use cached files")
		})
	})

	Describe("Error Handling", func() {
		// NOTE: Tests proper error messages for invalid inputs and failure scenarios

		It("fails with meaningful error for invalid Pivnet token", func() {
			output, err := runTileDiff(
				"--product-slug", "harbor-container-registry",
				"--old-version", "2.10.2",
				"--new-version", "2.10.3",
				"--product-file", "VMware Harbor Container Registry for Tanzu",
				"--pivnet-token", "invalid-token-12345",
				"--accept-eula",
				"--non-interactive",
			)

			if err != nil {
				if strings.Contains(output, "no releases found matching version") {
					Skip("Test product/versions not available - update test")
				}
			}

			// Should fail
			Expect(err).To(HaveOccurred(), "Invalid token should cause failure")

			// Should have meaningful error message about authentication
			Expect(output).To(Or(
				ContainSubstring("401"),
				ContainSubstring("authentication"),
				ContainSubstring("invalid token"),
				ContainSubstring("unauthorized"),
			), "Error should mention authentication failure")
		})

		It("fails when product files are ambiguous without --product-file flag", func() {
			output, err := runTileDiff(
				"--product-slug", "cf",
				"--old-version", "6.0.22+LTS-T",
				"--new-version", "6.0.23+LTS-T",
				"--pivnet-token", os.Getenv("PIVNET_TOKEN"),
				"--accept-eula",
				"--non-interactive",
			)

			// Should fail
			Expect(err).To(HaveOccurred(), "Missing product-file should cause failure")

			// Should mention --product-file flag
			Expect(output).To(Or(
				ContainSubstring("product-file"),
				ContainSubstring("--product-file"),
			), "Error should mention --product-file requirement")
		})

		It("fails when mixing local files and Pivnet flags", func() {
			output, err := runTileDiff(
				"--old-tile", "/tmp/old.pivotal",
				"--new-tile", "/tmp/new.pivotal",
				"--product-slug", "cf",
				"--old-version", "6.0",
			)

			// Should fail
			Expect(err).To(HaveOccurred(), "Mixed mode should be rejected")

			// Should mention mode conflict
			Expect(output).To(Or(
				ContainSubstring("Cannot mix"),
				MatchRegexp("local.*pivnet"),
				MatchRegexp("--old-tile.*--product-slug"),
			), "Error should mention mode conflict")
		})
	})

	Describe("Local Files Mode", Label("slow", "downloads"), func() {
		// NOTE: Tests backward compatibility with local tile file comparison

		var oldTilePath, newTilePath string

		BeforeEach(func() {
			// Download tiles to use as local test files
			output, err := runTileDiff(
				"--product-slug", "harbor-container-registry",
				"--old-version", "2.10.2",
				"--new-version", "2.10.3",
				"--product-file", "VMware Harbor Container Registry for Tanzu",
				"--pivnet-token", os.Getenv("PIVNET_TOKEN"),
				"--accept-eula",
				"--non-interactive",
			)

			if err != nil {
				if strings.Contains(output, "no releases found matching version") {
					Skip("Cannot download test tiles - update test")
				}
				Skip("Cannot download test tiles for local files mode")
			}

			// Find the downloaded tiles in cache
			files, err := os.ReadDir(testCacheDir)
			Expect(err).NotTo(HaveOccurred(), "Should be able to read cache directory")

			for _, file := range files {
				if strings.Contains(file.Name(), "2.10.2") && strings.HasSuffix(file.Name(), ".pivotal") {
					oldTilePath = fmt.Sprintf("%s/%s", testCacheDir, file.Name())
				}
				if strings.Contains(file.Name(), "2.10.3") && strings.HasSuffix(file.Name(), ".pivotal") {
					newTilePath = fmt.Sprintf("%s/%s", testCacheDir, file.Name())
				}
			}

			if oldTilePath == "" || newTilePath == "" {
				Skip("Could not find test tiles in cache")
			}
		})

		It("compares local tile files without Pivnet", func() {
			output, err := runTileDiff(
				"--old-tile", oldTilePath,
				"--new-tile", newTilePath,
			)

			// Should succeed
			Expect(err).NotTo(HaveOccurred(), "Local files mode should complete successfully")

			// Should show comparison results
			Expect(output).To(Or(
				ContainSubstring("Comparison Results"),
				ContainSubstring("Summary"),
			), "Should show comparison results")

			// Should NOT mention Pivnet operations
			Expect(output).NotTo(Or(
				ContainSubstring("Downloading"),
				ContainSubstring("Pivnet"),
			), "Should not mention Pivnet operations")
		})

		It("supports JSON output in local files mode", func() {
			output, err := runTileDiff(
				"--old-tile", oldTilePath,
				"--new-tile", newTilePath,
				"--format", "json",
			)

			// Should succeed
			Expect(err).NotTo(HaveOccurred(), "Local files JSON mode should complete successfully")

			// Validate JSON output
			var result interface{}
			err = json.Unmarshal([]byte(output), &result)
			Expect(err).NotTo(HaveOccurred(), "Output should be valid JSON")
		})

		It("fails with meaningful error when local file is missing", func() {
			output, err := runTileDiff(
				"--old-tile", "/tmp/nonexistent-tile.pivotal",
				"--new-tile", "/tmp/another-nonexistent.pivotal",
			)

			// Should fail
			Expect(err).To(HaveOccurred(), "Missing local file should cause failure")

			// Should have error about missing file
			Expect(output).To(Or(
				ContainSubstring("not found"),
				ContainSubstring("no such file"),
				ContainSubstring("cannot open"),
			), "Error should mention missing file")
		})
	})
})
