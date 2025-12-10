package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetProperties(t *testing.T) {
	// Mock API server
	mockResponse := PropertiesResponse{
		Properties: map[string]Property{
			".properties.test": {
				Type:         "boolean",
				Configurable: true,
				Credential:   false,
				Value:        true,
				Optional:     false,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle UAA token endpoint
		if r.URL.Path == "/uaa/oauth/token" {
			if r.Method != "POST" {
				t.Errorf("Expected POST for token endpoint, got %s", r.Method)
			}

			// Verify basic auth for UAA client
			username, _, ok := r.BasicAuth()
			if !ok || username != "opsman" {
				t.Error("Expected opsman client credentials")
			}

			// Return token
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"access_token": "test-token-12345",
				"token_type":   "bearer",
			})
			return
		}

		// Handle properties endpoint
		if r.URL.Path != "/api/v0/staged/products/test-guid/properties" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}

		// Check bearer token auth
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-token-12345" {
			t.Errorf("Expected bearer token auth, got: %s", authHeader)
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Test client
	client := NewClient(server.URL, "admin", "password", true)
	properties, err := client.GetProperties("test-guid")
	if err != nil {
		t.Fatalf("GetProperties failed: %v", err)
	}

	if len(properties.Properties) != 1 {
		t.Errorf("Expected 1 property, got %d", len(properties.Properties))
	}

	prop, exists := properties.Properties[".properties.test"]
	if !exists {
		t.Error("Expected .properties.test to exist")
	}
	if prop.Value != true {
		t.Errorf("Expected value true, got %v", prop.Value)
	}
}

func TestGetPropertiesBasicAuthFallback(t *testing.T) {
	// Mock API server that doesn't have UAA endpoint
	mockResponse := PropertiesResponse{
		Properties: map[string]Property{
			".properties.test": {
				Type:         "boolean",
				Configurable: true,
				Credential:   false,
				Value:        true,
				Optional:     false,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// UAA endpoint returns 404
		if r.URL.Path == "/uaa/oauth/token" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Handle properties endpoint with basic auth
		if r.URL.Path != "/api/v0/staged/products/test-guid/properties" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		// Check basic auth
		username, password, ok := r.BasicAuth()
		if !ok || username != "admin" || password != "password" {
			t.Error("Expected basic auth credentials")
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Test client
	client := NewClient(server.URL, "admin", "password", true)
	properties, err := client.GetProperties("test-guid")
	if err != nil {
		t.Fatalf("GetProperties failed: %v", err)
	}

	if len(properties.Properties) != 1 {
		t.Errorf("Expected 1 property, got %d", len(properties.Properties))
	}
}

func TestGetPropertiesHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient(server.URL, "admin", "wrong", true)
	_, err := client.GetProperties("test-guid")
	if err == nil {
		t.Error("Expected error for 401 response")
	}
}
