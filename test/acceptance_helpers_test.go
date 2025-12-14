// ABOUTME: Shared helper functions for Pivnet acceptance tests.
// ABOUTME: Provides utilities for running tile-diff binary, managing cache, and VCR setup.
package test

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/gomega"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

var (
	tileDiffBin   string
	testCacheDir  string
	testEULAFile  string
	useLivePivnet bool
	vcrMode       string
)

func init() {
	tileDiffBin = getEnvOrDefault("TILE_DIFF_BIN", "./tile-diff")
	testCacheDir = getEnvOrDefault("TEST_CACHE_DIR", "/tmp/tile-diff-test-cache")
	testEULAFile = filepath.Join(testCacheDir, ".pivnet-eula-acceptance")
	useLivePivnet = os.Getenv("ACCEPTANCE_USE_REAL_PIVNET") == "true"
	vcrMode = getEnvOrDefault("VCR_MODE", "replay")
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

// setupVCR creates a VCR recorder for HTTP fixture recording/replay
func setupVCR(cassettePath string) (*recorder.Recorder, *http.Client, error) {
	if useLivePivnet {
		// Live mode - use real HTTP client
		return nil, &http.Client{}, nil
	}

	mode := recorder.ModeReplayOnly
	switch vcrMode {
	case "record":
		mode = recorder.ModeRecordOnly
	case "replay":
		mode = recorder.ModeReplayOnly
	}

	r, err := recorder.NewWithOptions(&recorder.Options{
		CassetteName: cassettePath,
		Mode:         mode,
	})
	if err != nil {
		return nil, nil, err
	}

	// Add hook to mask auth tokens in recordings
	r.AddHook(func(i *cassette.Interaction) error {
		// Mask Authorization header
		delete(i.Request.Headers, "Authorization")
		return nil
	}, recorder.BeforeSaveHook)

	return r, r.GetDefaultClient(), nil
}

// stopVCR stops the VCR recorder if it exists
func stopVCR(r *recorder.Recorder) {
	if r != nil {
		Expect(r.Stop()).To(Succeed())
	}
}
