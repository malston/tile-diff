// ABOUTME: File download orchestration with progress tracking.
// ABOUTME: Coordinates cache, EULA, disk space, and actual download operations.
package pivnet

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Downloader orchestrates file downloads
type Downloader struct {
	client      *Client
	cache       *CacheManager
	eula        *EULAManager
	diskManager *DiskManager
	quiet       bool // Suppress progress output
}

// NewDownloader creates a new downloader
func NewDownloader(client *Client, cacheDir, manifestFile, eulaFile string, minFreeSpaceGB int64, quiet bool) *Downloader {
	return &Downloader{
		client:      client,
		cache:       NewCacheManager(cacheDir, manifestFile),
		eula:        NewEULAManager(eulaFile),
		diskManager: NewDiskManager(minFreeSpaceGB),
		quiet:       quiet,
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

	// Check EULA (per-release acceptance required)
	if !d.eula.IsAcceptedForRelease(opts.ProductSlug, release.Version) {
		eulaURL := fmt.Sprintf("https://network.tanzu.vmware.com/products/%s/releases/%d", opts.ProductSlug, release.ID)

		if opts.NonInteractive && !opts.AcceptEULA {
			return "", fmt.Errorf("EULA not accepted for %s.\n\nPlease accept the EULA at:\n%s\n\nThen run this command again, or use --accept-eula to acknowledge you've accepted it.", opts.ProductSlug, eulaURL)
		}

		if !opts.AcceptEULA {
			// Interactive EULA prompt
			accepted, err := PromptForEULA(opts.ProductSlug, release.Version, eulaURL)
			if err != nil {
				return "", err
			}
			if !accepted {
				return "", fmt.Errorf("EULA not accepted")
			}
		}

		// Try to accept EULA via API
		// Note: This only works for VMware/Broadcom employees
		err := d.client.AcceptEULA(opts.ProductSlug, release.ID)
		if err != nil {
			// API acceptance failed - handle based on interactive mode
			if opts.NonInteractive {
				// Non-interactive: Assume user has manually accepted EULA via web
				// Mark as accepted locally and proceed (download will fail if not actually accepted)
				d.eula.Accept(opts.ProductSlug, release.Version, eulaURL)
				fmt.Printf("Note: API EULA acceptance unavailable. Proceeding with download...\n")
			} else {
				// Interactive: Prompt user to accept manually via web
				fmt.Printf("\n⚠️  EULA must be accepted manually\n")
				fmt.Printf("API EULA acceptance is only available for Broadcom/VMware employees.\n")
				fmt.Printf("\nPlease:\n")
				fmt.Printf("1. Open this URL in your browser: %s\n", eulaURL)
				fmt.Printf("2. Accept the EULA\n")
				fmt.Printf("3. Press Enter here to continue...\n\n")
				fmt.Scanln()

				// Don't mark as accepted yet - let the download verify it
				// If the download succeeds, we'll mark it then
				fmt.Printf("Proceeding with download (EULA acceptance will be verified)...\n")
			}
		} else {
			// API acceptance succeeded (Broadcom/VMware employee)
			d.eula.Accept(opts.ProductSlug, release.Version, eulaURL)
			fmt.Printf("EULA accepted via API for %s\n", opts.ProductSlug)
		}
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

	// Get actual file size (ListForRelease doesn't return sizes)
	fileSize, err := d.client.GetProductFileSize(opts.ProductSlug, release.ID, selectedFile.ID)
	if err != nil {
		return "", fmt.Errorf("failed to get file size: %w", err)
	}

	// Download file
	fmt.Printf("Downloading %s (%s)...\n", selectedFile.Name, formatBytes(fileSize))

	targetPath := filepath.Join(opts.CacheDir, filepath.Base(selectedFile.AWSObjectKey))
	if err := os.MkdirAll(opts.CacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Download to temp file first
	tempPath := targetPath + ".tmp"
	err = d.downloadFile(opts.ProductSlug, release.ID, selectedFile.ID, tempPath, fileSize)
	if err != nil {
		os.Remove(tempPath)
		return "", err
	}

	// Verify temp file exists before attempting rename
	if _, err := os.Stat(tempPath); err != nil {
		return "", fmt.Errorf("downloaded temp file missing before rename: %w", err)
	}

	// Move to final location
	if err := os.Rename(tempPath, targetPath); err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to move downloaded file: %w", err)
	}

	// Add to cache
	d.cache.Add(opts.ProductSlug, release.Version, selectedFile.ID, targetPath, fileSize)

	// Mark EULA as accepted now that download succeeded
	// (If it wasn't already marked via API acceptance)
	if !d.eula.IsAcceptedForRelease(opts.ProductSlug, release.Version) {
		eulaURL := fmt.Sprintf("https://network.tanzu.vmware.com/products/%s/releases/%d", opts.ProductSlug, release.ID)
		d.eula.Accept(opts.ProductSlug, release.Version, eulaURL)
		fmt.Printf("EULA acceptance recorded for %s %s\n", opts.ProductSlug, release.Version)
	}

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

	// Create progress writer (suppress if quiet mode)
	// The SDK has its own progress bar (cheggaaa/pb) that outputs formatted text
	var progressWriter io.Writer
	if d.quiet {
		progressWriter = io.Discard
	} else {
		progressWriter = os.Stderr
	}

	// Download file - SDK's built-in progress bar will display to progressWriter
	err = d.client.DownloadFile(productSlug, releaseID, fileID, out, progressWriter)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	// Sync file to disk before closing
	if err := out.Sync(); err != nil {
		return fmt.Errorf("failed to sync file to disk: %w", err)
	}

	// Verify file size matches expected
	fileInfo, err := out.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat downloaded file: %w", err)
	}
	if fileInfo.Size() != fileSize {
		return fmt.Errorf("downloaded file size mismatch: expected %d bytes, got %d bytes", fileSize, fileInfo.Size())
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
