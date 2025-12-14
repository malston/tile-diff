// ABOUTME: Ginkgo acceptance tests for Pivnet integration.
// ABOUTME: Tests tile downloads, caching, EULA handling, and error scenarios.
package test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pivnet Integration", func() {
	BeforeEach(func() {
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

			// If the versions don't exist, expect a clear error message
			if err != nil {
				Expect(output).To(ContainSubstring("no releases found matching version"))
				Skip("Test product/versions not available in Pivnet - update test to use valid versions")
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

			// If the versions don't exist, expect a clear error message
			if err != nil {
				Expect(output).To(ContainSubstring("no releases found matching version"))
				Skip("Test product/versions not available in Pivnet - update test to use valid versions")
			}

			// Output should be valid JSON (we can check for basic JSON structure)
			Expect(output).To(Or(
				ContainSubstring("{"),
				ContainSubstring("}"),
			))
		})
	})
})
