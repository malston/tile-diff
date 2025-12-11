package releasenotes

import (
	"testing"
)

func TestLoadProductConfig(t *testing.T) {
	config, err := LoadProductConfig("testdata/products.yaml")
	if err != nil {
		t.Fatalf("LoadProductConfig failed: %v", err)
	}

	if config["cf"] != "https://techdocs.broadcom.com/cf/{version}/release-notes.html" {
		t.Errorf("Expected cf URL, got %s", config["cf"])
	}
}

func TestResolveURL(t *testing.T) {
	config := ProductConfig{
		"cf": "https://techdocs.broadcom.com/cf/{version}/release-notes.html",
	}

	url, err := config.ResolveURL("cf", "10.2.5")
	if err != nil {
		t.Fatalf("ResolveURL failed: %v", err)
	}

	expected := "https://techdocs.broadcom.com/cf/10.2.5/release-notes.html"
	if url != expected {
		t.Errorf("Expected %s, got %s", expected, url)
	}
}

func TestResolveURL_ProductNotFound(t *testing.T) {
	config := ProductConfig{}

	_, err := config.ResolveURL("unknown", "1.0.0")
	if err == nil {
		t.Error("Expected error for unknown product")
	}
}
