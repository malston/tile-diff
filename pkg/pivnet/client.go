// ABOUTME: Pivnet API client using standard HTTP calls.
// ABOUTME: Provides methods for fetching releases, product files, and downloading files.
package pivnet

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const defaultPivnetHost = "https://network.tanzu.vmware.com"

// Client wraps HTTP client for Pivnet API calls
type Client struct {
	token      string
	httpClient *http.Client
	baseURL    string
}

// pivnetRelease represents a release from the API
type pivnetRelease struct {
	ID      int    `json:"id"`
	Version string `json:"version"`
}

// pivnetReleasesResponse represents the releases list response
type pivnetReleasesResponse struct {
	Releases []pivnetRelease `json:"releases"`
}

// pivnetProductFile represents a product file from the API
type pivnetProductFile struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	AWSObjectKey string `json:"aws_object_key"`
	Size         int    `json:"size"`
}

// pivnetProductFilesResponse represents the product files list response
type pivnetProductFilesResponse struct {
	ProductFiles []pivnetProductFile `json:"product_files"`
}

// pivnetDownloadResponse represents the download endpoint response
type pivnetDownloadResponse struct {
	URL string `json:"url"`
}

// NewClient creates a new Pivnet client
func NewClient(token string) (*Client, error) {
	if token == "" {
		return nil, errors.New("pivnet token cannot be empty")
	}

	// Warn about legacy tokens
	if len(token) <= 32 {
		fmt.Printf("\n⚠️  Warning: Your Pivnet token appears to be a legacy format (%d chars).\n", len(token))
		fmt.Printf("Legacy tokens don't support all API endpoints (especially downloads).\n")
		fmt.Printf("Please get a new UAA token from:\n")
		fmt.Printf("https://network.tanzu.vmware.com/users/dashboard/edit-profile\n\n")
	}

	return &Client{
		token:      token,
		httpClient: &http.Client{},
		baseURL:    defaultPivnetHost,
	}, nil
}

// doRequest performs an authenticated HTTP request
func (c *Client) doRequest(method, path string) (*http.Response, error) {
	req, err := http.NewRequest(method, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

// GetReleases fetches all releases for a product
func (c *Client) GetReleases(productSlug string) ([]Release, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/api/v2/products/%s/releases", productSlug))
	if err != nil {
		return nil, fmt.Errorf("failed to list releases for %s: %w", productSlug, err)
	}
	defer resp.Body.Close()

	var releasesResp pivnetReleasesResponse
	if err := json.NewDecoder(resp.Body).Decode(&releasesResp); err != nil {
		return nil, fmt.Errorf("failed to decode releases response: %w", err)
	}

	result := make([]Release, len(releasesResp.Releases))
	for i, r := range releasesResp.Releases {
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
	resp, err := c.doRequest("GET", fmt.Sprintf("/api/v2/products/%s/releases/%d/product_files", productSlug, releaseID))
	if err != nil {
		return nil, fmt.Errorf("failed to list product files for release %d: %w", releaseID, err)
	}
	defer resp.Body.Close()

	var filesResp pivnetProductFilesResponse
	if err := json.NewDecoder(resp.Body).Decode(&filesResp); err != nil {
		return nil, fmt.Errorf("failed to decode product files response: %w", err)
	}

	result := make([]ProductFile, 0)
	for _, f := range filesResp.ProductFiles {
		// Only include .pivotal files
		if len(f.AWSObjectKey) > 8 && f.AWSObjectKey[len(f.AWSObjectKey)-8:] == ".pivotal" {
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

// AcceptEULA accepts the EULA for a release
func (c *Client) AcceptEULA(productSlug string, releaseID int) error {
	// First, get the EULA ID for this release
	resp, err := c.doRequest("GET", fmt.Sprintf("/api/v2/products/%s/releases/%d", productSlug, releaseID))
	if err != nil {
		return fmt.Errorf("failed to get release details: %w", err)
	}
	defer resp.Body.Close()

	var releaseData struct {
		EULA struct {
			ID int `json:"id"`
		} `json:"eula"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&releaseData); err != nil {
		return fmt.Errorf("failed to decode release response: %w", err)
	}

	// Accept the EULA
	req, err := http.NewRequest("POST", c.baseURL+fmt.Sprintf("/api/v2/products/%s/releases/%d/pivnet_resource_eula_acceptance", productSlug, releaseID), strings.NewReader("{}"))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp2, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to accept EULA: %w", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode < 200 || resp2.StatusCode >= 300 {
		body, _ := io.ReadAll(resp2.Body)
		return fmt.Errorf("EULA acceptance failed with status %d: %s", resp2.StatusCode, string(body))
	}

	return nil
}

// DownloadFile downloads a product file
func (c *Client) DownloadFile(productSlug string, releaseID, fileID int, writer io.Writer) error {
	// Get product file metadata to extract download link
	resp, err := c.doRequest("GET", fmt.Sprintf("/api/v2/products/%s/releases/%d/product_files/%d", productSlug, releaseID, fileID))
	if err != nil {
		return fmt.Errorf("failed to get product file metadata: %w", err)
	}
	defer resp.Body.Close()

	var fileData struct {
		Links struct {
			Download struct {
				Href string `json:"href"`
			} `json:"download"`
		} `json:"_links"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&fileData); err != nil {
		return fmt.Errorf("failed to decode product file response: %w", err)
	}

	if fileData.Links.Download.Href == "" {
		return fmt.Errorf("no download link found in response")
	}

	// Download the file from the download link
	downloadReq, err := http.NewRequest("GET", fileData.Links.Download.Href, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}

	// Add authorization header for the download
	downloadReq.Header.Set("Authorization", "Bearer "+c.token)
	downloadReq.Header.Set("Accept", "application/json")

	downloadResp, err := c.httpClient.Do(downloadReq)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer downloadResp.Body.Close()

	if downloadResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(downloadResp.Body)
		return fmt.Errorf("download failed with status %d: %s", downloadResp.StatusCode, string(body))
	}

	// Copy the file content to the writer
	_, err = io.Copy(writer, downloadResp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
