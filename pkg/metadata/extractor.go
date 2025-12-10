// ABOUTME: Extracts metadata.yml from .pivotal ZIP archives.
// ABOUTME: Provides functions to read and parse tile metadata files.
package metadata

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
)

// ExtractMetadata extracts metadata/metadata.yml from a .pivotal ZIP archive
func ExtractMetadata(zipData []byte) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, fmt.Errorf("failed to read ZIP archive: %w", err)
	}

	// Look for metadata/metadata.yml
	for _, f := range reader.File {
		if f.Name == "metadata/metadata.yml" {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open metadata.yml: %w", err)
			}
			defer rc.Close()

			content, err := io.ReadAll(rc)
			if err != nil {
				return nil, fmt.Errorf("failed to read metadata.yml: %w", err)
			}

			return content, nil
		}
	}

	return nil, fmt.Errorf("metadata/metadata.yml not found in archive")
}
