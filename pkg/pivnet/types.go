// ABOUTME: Core types and configuration for Pivnet integration.
// ABOUTME: Defines config structs, validation logic, and download options.
package pivnet

import (
	"errors"
	"os"
	"path/filepath"
)

// Config holds configuration for Pivnet downloads
type Config struct {
	Token           string // Pivnet API token
	ProductSlug     string // Product slug (e.g., "cf")
	OldVersion      string // Old release version
	NewVersion      string // New release version
	ProductFile     string // Optional: specific product file name
	AcceptEULA      bool   // Accept EULA without prompting
	NonInteractive  bool   // Fail instead of prompting
	CacheDir        string // Cache directory (default: ~/.tile-diff/cache)
}

// DownloadOptions holds options for a single download operation
type DownloadOptions struct {
	ProductSlug    string
	Version        string
	ProductFile    string // Optional
	AcceptEULA     bool
	NonInteractive bool
	CacheDir       string
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Token == "" {
		return errors.New("pivnet token required")
	}
	if c.ProductSlug == "" {
		return errors.New("product slug required")
	}
	if c.OldVersion == "" || c.NewVersion == "" {
		return errors.New("both old and new versions required")
	}
	return nil
}

// DownloadOptions creates download options from config
func (c *Config) DownloadOptions() DownloadOptions {
	cacheDir := c.CacheDir
	if cacheDir == "" {
		home, _ := os.UserHomeDir()
		cacheDir = filepath.Join(home, ".tile-diff", "cache")
	}

	return DownloadOptions{
		ProductSlug:    c.ProductSlug,
		CacheDir:       cacheDir,
		ProductFile:    c.ProductFile,
		AcceptEULA:     c.AcceptEULA,
		NonInteractive: c.NonInteractive,
	}
}

// Release represents a Pivnet product release
type Release struct {
	ID      int
	Version string
}

// ProductFile represents a downloadable file in a release
type ProductFile struct {
	ID           int
	Name         string
	AWSObjectKey string
	Size         int64
}

// CacheEntry represents a cached download
type CacheEntry struct {
	ProductSlug  string
	Version      string
	ProductFileID int
	FilePath     string
	DownloadedAt string
	FileSize     int64
}
