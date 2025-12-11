// ABOUTME: Pivnet API client wrapper for interacting with Pivotal Network.
// ABOUTME: Provides methods for fetching releases, product files, and managing downloads.
package pivnet

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/go-pivnet/v7"
	"github.com/pivotal-cf/go-pivnet/v7/logger"
)

// Client wraps the Pivnet API client
type Client struct {
	pivnetClient pivnet.Client
}

// simpleLogger implements the logger.Logger interface
type simpleLogger struct{}

func (l *simpleLogger) Debug(action string, data ...logger.Data) {}
func (l *simpleLogger) Info(action string, data ...logger.Data)  {}

// NewClient creates a new Pivnet client
func NewClient(token string) (*Client, error) {
	if token == "" {
		return nil, errors.New("pivnet token cannot be empty")
	}

	config := pivnet.ClientConfig{
		Host:              pivnet.DefaultHost,
		SkipSSLValidation: false,
	}

	tokenService := pivnet.NewAccessTokenOrLegacyToken(token, config.Host, config.SkipSSLValidation)
	pivnetClient := pivnet.NewClient(tokenService, config, &simpleLogger{})

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
	files, err := c.pivnetClient.ProductFiles.ListForRelease(productSlug, releaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to list product files for release %d: %w", releaseID, err)
	}

	result := make([]ProductFile, 0)
	for _, f := range files {
		// Only include .pivotal files
		if len(f.AWSObjectKey) > 0 && (len(f.AWSObjectKey) > 8 && f.AWSObjectKey[len(f.AWSObjectKey)-8:] == ".pivotal") {
			result = append(result, ProductFile{
				ID:           f.ID,
				Name:         f.Name,
				AWSObjectKey: f.AWSObjectKey,
				Size:         int64(f.Size),
			})
		}
	}

	return result, nil
}
