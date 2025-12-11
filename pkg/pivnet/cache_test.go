// ABOUTME: Unit tests for download cache management.
// ABOUTME: Validates cache storage, retrieval, and cleanup logic.
package pivnet

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCacheManager(t *testing.T) {
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")
	manifestFile := filepath.Join(tempDir, "cache-manifest.json")

	mgr := NewCacheManager(cacheDir, manifestFile)

	// Test: Get non-existent entry
	entry := mgr.Get("cf", "6.0.22+LTS-T", 12345)
	if entry != nil {
		t.Error("Expected nil for non-existent cache entry")
	}

	// Test: Add entry
	testFile := filepath.Join(cacheDir, "test.pivotal")
	os.MkdirAll(cacheDir, 0755)
	os.WriteFile(testFile, []byte("test"), 0644)

	err := mgr.Add("cf", "6.0.22+LTS-T", 12345, testFile, 4)
	if err != nil {
		t.Fatalf("Failed to add cache entry: %v", err)
	}

	// Test: Get existing entry
	entry = mgr.Get("cf", "6.0.22+LTS-T", 12345)
	if entry == nil {
		t.Fatal("Expected cache entry, got nil")
	}
	if entry.FilePath != testFile {
		t.Errorf("Expected FilePath %s, got %s", testFile, entry.FilePath)
	}

	// Test: Persistence
	mgr2 := NewCacheManager(cacheDir, manifestFile)
	entry2 := mgr2.Get("cf", "6.0.22+LTS-T", 12345)
	if entry2 == nil {
		t.Error("Expected cache entry after reload")
	}
}

func TestCacheCleanup(t *testing.T) {
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")
	manifestFile := filepath.Join(tempDir, "cache-manifest.json")
	os.MkdirAll(cacheDir, 0755)

	mgr := NewCacheManager(cacheDir, manifestFile)

	// Add old entry (8 days ago)
	oldFile := filepath.Join(cacheDir, "old.pivotal")
	os.WriteFile(oldFile, []byte("old"), 0644)
	oldTime := time.Now().Add(-8 * 24 * time.Hour).Format(time.RFC3339)
	mgr.entries["cf-old-1"] = &CacheEntry{
		ProductSlug:   "cf",
		Version:       "old",
		ProductFileID: 1,
		FilePath:      oldFile,
		DownloadedAt:  oldTime,
		FileSize:      3,
	}

	// Add recent entry (2 days ago)
	newFile := filepath.Join(cacheDir, "new.pivotal")
	os.WriteFile(newFile, []byte("new"), 0644)
	newTime := time.Now().Add(-2 * 24 * time.Hour).Format(time.RFC3339)
	mgr.entries["cf-new-2"] = &CacheEntry{
		ProductSlug:   "cf",
		Version:       "new",
		ProductFileID: 2,
		FilePath:      newFile,
		DownloadedAt:  newTime,
		FileSize:      3,
	}

	mgr.save()

	// Run cleanup (7 day threshold)
	removed, err := mgr.CleanupOld(7)
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	if removed != 1 {
		t.Errorf("Expected 1 file removed, got %d", removed)
	}

	// Verify old file deleted
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("Expected old file to be deleted")
	}

	// Verify new file still exists
	if _, err := os.Stat(newFile); err != nil {
		t.Error("Expected new file to still exist")
	}
}
