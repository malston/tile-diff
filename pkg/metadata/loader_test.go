package metadata

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromFile(t *testing.T) {
	// Create a temporary test .pivotal file
	tmpDir := t.TempDir()
	pivotalPath := filepath.Join(tmpDir, "test.pivotal")

	// Create ZIP with metadata
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	metadataContent := []byte(`property_blueprints:
  - name: test_property
    type: string
    configurable: true
    optional: false
  - name: another_property
    type: integer
    configurable: false
`)

	f, err := w.Create("metadata/metadata.yml")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write(metadataContent)
	if err != nil {
		t.Fatal(err)
	}
	w.Close()

	// Write to file
	err = os.WriteFile(pivotalPath, buf.Bytes(), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test loading
	metadata, err := LoadFromFile(pivotalPath)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	if len(metadata.PropertyBlueprints) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(metadata.PropertyBlueprints))
	}

	if metadata.PropertyBlueprints[0].Name != "test_property" {
		t.Errorf("Expected first property name 'test_property', got '%s'",
			metadata.PropertyBlueprints[0].Name)
	}
}

func TestLoadFromFileNotFound(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/file.pivotal")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}
