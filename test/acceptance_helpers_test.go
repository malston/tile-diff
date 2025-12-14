// ABOUTME: Shared helper functions for Pivnet acceptance tests.
// ABOUTME: Provides utilities for running tile-diff binary and managing cache.
package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/gomega"
)

var (
	tileDiffBin  string
	testCacheDir string
	testEULAFile string
)

func init() {
	tileDiffBin = getEnvOrDefault("TILE_DIFF_BIN", "./tile-diff")
	testCacheDir = getEnvOrDefault("TEST_CACHE_DIR", "/tmp/tile-diff-test-cache")
	testEULAFile = filepath.Join(testCacheDir, ".pivnet-eula-acceptance")
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// runTileDiff executes the tile-diff binary with given arguments
func runTileDiff(args ...string) (string, error) {
	cmd := exec.Command(tileDiffBin, args...)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PIVNET_CACHE_DIR=%s", testCacheDir),
		fmt.Sprintf("PIVNET_EULA_FILE=%s", testEULAFile),
	)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// setupCacheDir creates and cleans test cache directory
func setupCacheDir() {
	Expect(os.RemoveAll(testCacheDir)).To(Succeed())
	Expect(os.MkdirAll(testCacheDir, 0755)).To(Succeed())
}

// cleanupCacheDir removes test cache directory
func cleanupCacheDir() {
	os.RemoveAll(testCacheDir)
}

// tileExistsInCache checks if a tile file exists in cache
func tileExistsInCache(filename string) bool {
	path := filepath.Join(testCacheDir, filename)
	_, err := os.Stat(path)
	return err == nil
}
