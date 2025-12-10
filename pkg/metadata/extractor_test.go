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
