// ABOUTME: HTTP client for Ops Manager API interactions.
// ABOUTME: Handles authentication and property retrieval.
package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client represents an Ops Manager API client
type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
}

// NewClient creates a new Ops Manager API client
func NewClient(baseURL, username, password string, skipSSLValidation bool) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipSSLValidation,
		},
	}

	return &Client{
		baseURL:  baseURL,
		username: username,
		password: password,
		httpClient: &http.Client{
			Transport: transport,
		},
	}
}

// GetProperties retrieves properties for a staged product
func (c *Client) GetProperties(productGUID string) (*PropertiesResponse, error) {
	url := fmt.Sprintf("%s/api/v0/staged/products/%s/properties", c.baseURL, productGUID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var properties PropertiesResponse
	err = json.NewDecoder(resp.Body).Decode(&properties)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &properties, nil
}
