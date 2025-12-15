// ABOUTME: HTTP client for Ops Manager API interactions.
// ABOUTME: Handles UAA authentication and property retrieval.
package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client represents an Ops Manager API client
type Client struct {
	baseURL     string
	username    string
	password    string
	httpClient  *http.Client
	accessToken string
	useBasicAuth bool
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

// uaaTokenResponse represents the UAA token response
type uaaTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// authenticate gets a UAA access token or falls back to basic auth
func (c *Client) authenticate() error {
	tokenURL := fmt.Sprintf("%s/uaa/oauth/token", c.baseURL)

	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("username", c.username)
	data.Set("password", c.password)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	req.SetBasicAuth("opsman", "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// UAA endpoint not available, fall back to basic auth
		c.useBasicAuth = true
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// UAA not available, fall back to basic auth
		c.useBasicAuth = true
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp uaaTokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	c.accessToken = tokenResp.AccessToken
	return nil
}

// GetProperties retrieves properties for a staged product
func (c *Client) GetProperties(productGUID string) (*PropertiesResponse, error) {
	// Authenticate if we don't have a token and not using basic auth
	if c.accessToken == "" && !c.useBasicAuth {
		if err := c.authenticate(); err != nil {
			return nil, fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	url := fmt.Sprintf("%s/api/v0/staged/products/%s/properties", c.baseURL, productGUID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Use basic auth or bearer token based on what's available
	if c.useBasicAuth {
		req.SetBasicAuth(c.username, c.password)
	} else {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
	}
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

// GetStagedProducts retrieves all staged products
func (c *Client) GetStagedProducts() ([]StagedProduct, error) {
	// Authenticate if we don't have a token and not using basic auth
	if c.accessToken == "" && !c.useBasicAuth {
		if err := c.authenticate(); err != nil {
			return nil, fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	url := fmt.Sprintf("%s/api/v0/staged/products", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Use basic auth or bearer token based on what's available
	if c.useBasicAuth {
		req.SetBasicAuth(c.username, c.password)
	} else {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
	}
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

	var products []StagedProduct
	err = json.NewDecoder(resp.Body).Decode(&products)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return products, nil
}

// FindProductGUID finds a product GUID by product slug/type
func (c *Client) FindProductGUID(productSlug string) (string, error) {
	products, err := c.GetStagedProducts()
	if err != nil {
		return "", err
	}

	for _, product := range products {
		if product.Type == productSlug {
			return product.GUID, nil
		}
	}

	return "", fmt.Errorf("no staged product found with type '%s'", productSlug)
}
