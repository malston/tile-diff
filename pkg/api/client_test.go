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
		// Verify request
		if r.URL.Path != "/api/v0/staged/products/test-guid/properties" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}

		// Check auth header
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

	prop, exists := properties.Properties[".properties.test"]
	if !exists {
		t.Error("Expected .properties.test to exist")
	}
	if prop.Value != true {
		t.Errorf("Expected value true, got %v", prop.Value)
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
