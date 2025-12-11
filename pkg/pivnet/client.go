// ABOUTME: Pivnet API client using go-pivnet SDK.
// ABOUTME: Provides methods for fetching releases, product files, and downloading files.
package pivnet

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/pivotal-cf/go-pivnet/v7"
	"github.com/pivotal-cf/go-pivnet/v7/download"
	"github.com/pivotal-cf/go-pivnet/v7/logshim"
)

// Client wraps go-pivnet SDK client
type Client struct {
	pivnetClient pivnet.Client
}

// NewClient creates a new Pivnet client using the SDK
func NewClient(token string) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("pivnet token cannot be empty")
	}

	// Use SDK's token service that handles both legacy and UAA tokens
	config := pivnet.ClientConfig{
		Host:      "https://network.tanzu.vmware.com",
		UserAgent: "tile-diff",
	}

	// Create logger (discard output - we handle our own logging)
	stdout := log.New(io.Discard, "", 0)
	stderr := log.New(io.Discard, "", 0)
	logger := logshim.NewLogShim(stdout, stderr, false)

	// Create token service that handles both legacy and UAA tokens
	tokenService := pivnet.NewAccessTokenOrLegacyToken(
		token,
		config.Host,
		false, // skipSSL
		config.UserAgent,
	)

	// Create SDK client
	pivnetClient := pivnet.NewClient(tokenService, config, logger)

	return &Client{
		pivnetClient: pivnetClient,
	}, nil
}

// GetReleases fetches all releases for a product
func (c *Client) GetReleases(productSlug string) ([]Release, error) {
	releases, err := c.pivnetClient.Releases.List(productSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to list releases for %s: %w", productSlug, err)
	}

	result := make([]Release, len(releases))
	for i, r := range releases {
		result[i] = Release{
			ID:      r.ID,
			Version: r.Version,
		}
	}

	return result, nil
}

// GetRelease fetches a specific release by version
func (c *Client) GetRelease(productSlug, version string) (*Release, error) {
	releases, err := c.GetReleases(productSlug)
	if err != nil {
		return nil, err
	}

	for _, r := range releases {
		if r.Version == version {
			return &r, nil
		}
	}

	return nil, fmt.Errorf("release version %s not found for product %s", version, productSlug)
}

// GetProductFiles fetches product files for a release
func (c *Client) GetProductFiles(productSlug string, releaseID int) ([]ProductFile, error) {
	// Get product files from the release
	productFiles, err := c.pivnetClient.ProductFiles.ListForRelease(productSlug, releaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to list product files: %w", err)
	}

	// Also get files from file groups (some products organize files this way)
	fileGroups, err := c.pivnetClient.FileGroups.ListForRelease(productSlug, releaseID)
	if err != nil {
		// File groups are optional, don't fail if they don't exist
		fileGroups = []pivnet.FileGroup{}
	}

	// Combine files from file groups
	for _, fg := range fileGroups {
		productFiles = append(productFiles, fg.ProductFiles...)
	}

	if len(productFiles) == 0 {
		return nil, fmt.Errorf("no product files found for release")
	}

	result := make([]ProductFile, 0, len(productFiles))
	for _, pf := range productFiles {
		// Only include files that are actual tiles (not docs, etc)
		if pf.FileType == "Software" || pf.AWSObjectKey != "" {
			result = append(result, ProductFile{
				ID:           pf.ID,
				Name:         pf.Name,
				AWSObjectKey: pf.AWSObjectKey,
				Size:         int64(pf.Size),
			})
		}
	}

	return result, nil
}

// AcceptEULA accepts the EULA for a release
func (c *Client) AcceptEULA(productSlug string, releaseID int) error {
	return c.pivnetClient.EULA.Accept(productSlug, releaseID)
}

// DownloadFile downloads a product file
func (c *Client) DownloadFile(productSlug string, releaseID, fileID int, writer io.Writer) error {
	// The SDK expects an *os.File for download, but we need to write to an io.Writer
	// We'll need to handle this by using a temp file or direct stream
	// For now, check if writer is already an *os.File
	file, ok := writer.(*os.File)
	if !ok {
		return fmt.Errorf("writer must be an *os.File for SDK download")
	}

	// Create file info for the download
	fileInfo, err := download.NewFileInfo(file)
	if err != nil {
		return fmt.Errorf("failed to create file info: %w", err)
	}

	// Download using SDK
	err = c.pivnetClient.ProductFiles.DownloadForRelease(
		fileInfo,
		productSlug,
		releaseID,
		fileID,
		io.Discard, // progress writer - we handle progress elsewhere
	)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	return nil
}
