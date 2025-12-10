package metadata

import (
	"archive/zip"
	"bytes"
	"testing"
)

func TestExtractMetadataFromZip(t *testing.T) {
	// Create a test ZIP with metadata.yml
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	metadataContent := []byte(`property_blueprints:
  - name: test_prop
    type: string
    configurable: true
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

	// Test extraction
	zipBytes := buf.Bytes()
	content, err := ExtractMetadata(zipBytes)
	if err != nil {
		t.Fatalf("ExtractMetadata failed: %v", err)
	}

	if len(content) == 0 {
		t.Error("Expected non-empty metadata content")
	}

	// Verify content contains our test data
	if !bytes.Contains(content, []byte("test_prop")) {
		t.Error("Expected metadata to contain 'test_prop'")
	}
}

func TestExtractMetadataNotFound(t *testing.T) {
	// Create a ZIP without metadata.yml
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	f, err := w.Create("some_other_file.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write([]byte("content"))
	if err != nil {
		t.Fatal(err)
	}
	w.Close()

	zipBytes := buf.Bytes()
	_, err = ExtractMetadata(zipBytes)
	if err == nil {
		t.Error("Expected error when metadata.yml not found")
	}
}

func TestExtractMetadataInvalidZip(t *testing.T) {
	// Test with invalid ZIP data
	_, err := ExtractMetadata([]byte("not a zip file"))
	if err == nil {
		t.Error("Expected error for invalid ZIP data")
	}
}

func TestParseMetadata(t *testing.T) {
	yamlData := []byte(`property_blueprints:
  - name: first_property
    type: boolean
    configurable: true
    optional: false
  - name: second_property
    type: integer
    configurable: true
    optional: true
    default: 100
    constraints:
      min: 10
      max: 1000
`)

	metadata, err := ParseMetadata(yamlData)
	if err != nil {
		t.Fatalf("ParseMetadata failed: %v", err)
	}

	if len(metadata.PropertyBlueprints) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(metadata.PropertyBlueprints))
	}

	first := metadata.PropertyBlueprints[0]
	if first.Name != "first_property" {
		t.Errorf("Expected name 'first_property', got '%s'", first.Name)
	}
	if !first.Configurable {
		t.Error("Expected first property to be configurable")
	}

	second := metadata.PropertyBlueprints[1]
	if second.Name != "second_property" {
		t.Errorf("Expected name 'second_property', got '%s'", second.Name)
	}
	if second.Constraints == nil {
		t.Error("Expected second property to have constraints")
	}
}

func TestParseMetadataInvalid(t *testing.T) {
	yamlData := []byte(`invalid yaml: [[[`)

	_, err := ParseMetadata(yamlData)
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}
