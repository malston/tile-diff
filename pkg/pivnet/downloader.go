// ABOUTME: File download orchestration with progress tracking.
// ABOUTME: Coordinates cache, EULA, disk space, and actual download operations.
package pivnet

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/schollz/progressbar/v3"
)

// Downloader orchestrates file downloads
type Downloader struct {
	client      *Client
	cache       *CacheManager
	eula        *EULAManager
	diskManager *DiskManager
}

// NewDownloader creates a new downloader
func NewDownloader(client *Client, cacheDir, manifestFile, eulaFile string, minFreeSpaceGB int64) *Downloader {
	return &Downloader{
		client:      client,
		cache:       NewCacheManager(cacheDir, manifestFile),
		eula:        NewEULAManager(eulaFile),
		diskManager: NewDiskManager(minFreeSpaceGB),
	}
}

// Download downloads a product file
func (d *Downloader) Download(opts DownloadOptions) (string, error) {
	// Resolve version
	releases, err := d.client.GetReleases(opts.ProductSlug)
	if err != nil {
		return "", err
	}

	resolver := NewResolver(releases, opts.NonInteractive)
	result, err := resolver.Resolve(opts.Version)
	if err != nil {
		return "", err
	}

	// Handle multiple matches
	if result.Selected == nil {
		if opts.NonInteractive {
			return "", fmt.Errorf("multiple releases match - use exact version in non-interactive mode")
		}
		selected, err := PromptForRelease(result.Matches)
		if err != nil {
			return "", err
		}
		result.Selected = selected
	}

	release := result.Selected

	// Get product files
	productFiles, err := d.client.GetProductFiles(opts.ProductSlug, release.ID)
	if err != nil {
		return "", err
	}

	if len(productFiles) == 0 {
		return "", fmt.Errorf("no product files found for release %s", opts.Version)
	}

	// Select product file
	var selectedFile *ProductFile
	if opts.ProductFile != "" {
		// Find specific file
		for _, pf := range productFiles {
			if pf.Name == opts.ProductFile {
				selectedFile = &pf
				break
			}
		}
		if selectedFile == nil {
			return "", fmt.Errorf("product file '%s' not found", opts.ProductFile)
		}
	} else {
		if opts.NonInteractive {
			if len(productFiles) > 1 {
				return "", fmt.Errorf("multiple product files found - specify --product-file in non-interactive mode")
			}
			selectedFile = &productFiles[0]
		} else {
			// Interactive selection
			selected, err := PromptForProductFile(productFiles)
			if err != nil {
				return "", err
			}
			selectedFile = selected
		}
	}

	// Check cache
	cached := d.cache.Get(opts.ProductSlug, release.Version, selectedFile.ID)
	if cached != nil {
		fmt.Printf("Using cached file: %s\n", cached.FilePath)
		return cached.FilePath, nil
	}

	// Check EULA
	if !d.eula.IsAccepted(opts.ProductSlug) {
		eulaURL := fmt.Sprintf("https://network.pivotal.io/products/%s/releases/%d", opts.ProductSlug, release.ID)

		if !opts.AcceptEULA {
			if opts.NonInteractive {
				return "", fmt.Errorf("EULA not accepted for %s (use --accept-eula)", opts.ProductSlug)
			}
			// Interactive EULA prompt
			accepted, err := PromptForEULA(opts.ProductSlug, release.Version, eulaURL)
			if err != nil {
				return "", err
			}
			if !accepted {
				return "", fmt.Errorf("EULA not accepted")
			}
		}

		// Accept EULA via API
		if err := d.client.AcceptEULA(opts.ProductSlug, release.ID); err != nil {
			return "", fmt.Errorf("failed to accept EULA: %w", err)
		}
		d.eula.Accept(opts.ProductSlug, release.Version, eulaURL)
	}

	// Check disk space
	hasSpace, err := d.diskManager.HasEnoughSpace(opts.CacheDir, selectedFile.Size)
	if err != nil {
		return "", fmt.Errorf("failed to check disk space: %w", err)
	}
	if !hasSpace {
		// Try cleanup
		removed, err := d.cache.CleanupOld(7)
		if err != nil {
			return "", fmt.Errorf("insufficient disk space and cleanup failed: %w", err)
		}
		fmt.Printf("Cleaned up %d old cached files\n", removed)

		// Check again
		hasSpace, _ = d.diskManager.HasEnoughSpace(opts.CacheDir, selectedFile.Size)
		if !hasSpace {
			return "", fmt.Errorf("insufficient disk space even after cleanup")
		}
	}

	// Download file
	fmt.Printf("Downloading %s (%s)...\n", selectedFile.Name, formatBytes(selectedFile.Size))

	targetPath := filepath.Join(opts.CacheDir, filepath.Base(selectedFile.AWSObjectKey))
	if err := os.MkdirAll(opts.CacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Download to temp file first
	tempPath := targetPath + ".tmp"
	err = d.downloadFile(opts.ProductSlug, release.ID, selectedFile.ID, tempPath, selectedFile.Size)
	if err != nil {
		os.Remove(tempPath)
		return "", err
	}

	// Move to final location
	if err := os.Rename(tempPath, targetPath); err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to move downloaded file: %w", err)
	}

	// Add to cache
	d.cache.Add(opts.ProductSlug, release.Version, selectedFile.ID, targetPath, selectedFile.Size)

	return targetPath, nil
}

// downloadFile downloads a file with progress bar
func (d *Downloader) downloadFile(productSlug string, releaseID, fileID int, targetPath string, fileSize int64) error {
	// Create output file
	out, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Create progress bar
	bar := progressbar.DefaultBytes(
		fileSize,
		"downloading",
	)

	// Create a multi-writer that writes to both file and progress bar
	multiWriter := io.MultiWriter(out, bar)

	// Download file
	err = d.client.DownloadFile(productSlug, releaseID, fileID, multiWriter)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	return nil
}

// formatBytes formats bytes as human-readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
