// ABOUTME: Download cache management for tile files.
// ABOUTME: Handles caching, retrieval, and time-based cleanup of downloaded tiles.
package pivnet

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CacheManager manages the download cache
type CacheManager struct {
	cacheDir     string
	manifestFile string
	entries      map[string]*CacheEntry
}

// NewCacheManager creates a new cache manager
func NewCacheManager(cacheDir, manifestFile string) *CacheManager {
	mgr := &CacheManager{
		cacheDir:     cacheDir,
		manifestFile: manifestFile,
		entries:      make(map[string]*CacheEntry),
	}
	mgr.load()
	return mgr
}

// Get retrieves a cache entry if it exists
func (m *CacheManager) Get(productSlug, version string, productFileID int) *CacheEntry {
	key := m.cacheKey(productSlug, version, productFileID)
	entry, exists := m.entries[key]
	if !exists {
		return nil
	}

	// Verify file still exists
	if _, err := os.Stat(entry.FilePath); err != nil {
		// File gone - remove from cache
		delete(m.entries, key)
		m.save()
		return nil
	}

	return entry
}

// Add adds a file to the cache
func (m *CacheManager) Add(productSlug, version string, productFileID int, filePath string, fileSize int64) error {
	key := m.cacheKey(productSlug, version, productFileID)
	m.entries[key] = &CacheEntry{
		ProductSlug:   productSlug,
		Version:       version,
		ProductFileID: productFileID,
		FilePath:      filePath,
		DownloadedAt:  time.Now().UTC().Format(time.RFC3339),
		FileSize:      fileSize,
	}
	return m.save()
}

// CleanupOld removes cache entries older than maxAgeDays
func (m *CacheManager) CleanupOld(maxAgeDays int) (int, error) {
	threshold := time.Now().Add(-time.Duration(maxAgeDays) * 24 * time.Hour)
	removed := 0

	for key, entry := range m.entries {
		downloadedAt, err := time.Parse(time.RFC3339, entry.DownloadedAt)
		if err != nil {
			continue
		}

		if downloadedAt.Before(threshold) {
			// Remove file
			if err := os.Remove(entry.FilePath); err != nil {
				if !os.IsNotExist(err) {
					return removed, fmt.Errorf("failed to remove %s: %w", entry.FilePath, err)
				}
			}
			delete(m.entries, key)
			removed++
		}
	}

	if removed > 0 {
		m.save()
	}

	return removed, nil
}

// cacheKey generates a unique cache key
func (m *CacheManager) cacheKey(productSlug, version string, productFileID int) string {
	return fmt.Sprintf("%s-%s-%d", productSlug, version, productFileID)
}

// load reads cache manifest from disk
func (m *CacheManager) load() error {
	data, err := os.ReadFile(m.manifestFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &m.entries)
}

// save writes cache manifest to disk
func (m *CacheManager) save() error {
	dir := filepath.Dir(m.manifestFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	data, err := json.MarshalIndent(m.entries, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.manifestFile, data, 0644)
}
