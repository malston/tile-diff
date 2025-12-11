// ABOUTME: EULA acceptance tracking and persistence.
// ABOUTME: Manages one-time EULA acceptance per product with file-based storage.
package pivnet

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// EULARecord represents an accepted EULA
type EULARecord struct {
	AcceptedAt     string `json:"accepted_at"`
	ReleaseVersion string `json:"release_version"`
	EULAURL        string `json:"eula_url"`
}

// EULAManager manages EULA acceptances
type EULAManager struct {
	filePath string
	records  map[string]EULARecord
}

// NewEULAManager creates a new EULA manager
func NewEULAManager(filePath string) *EULAManager {
	mgr := &EULAManager{
		filePath: filePath,
		records:  make(map[string]EULARecord),
	}
	mgr.load()
	return mgr
}

// IsAccepted checks if EULA has been accepted for a product
func (m *EULAManager) IsAccepted(productSlug string) bool {
	_, exists := m.records[productSlug]
	return exists
}

// Accept records EULA acceptance for a product
func (m *EULAManager) Accept(productSlug, version, eulaURL string) error {
	m.records[productSlug] = EULARecord{
		AcceptedAt:     time.Now().UTC().Format(time.RFC3339),
		ReleaseVersion: version,
		EULAURL:        eulaURL,
	}
	return m.save()
}

// load reads EULA records from disk
func (m *EULAManager) load() error {
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet - that's ok
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &m.records)
}

// save writes EULA records to disk
func (m *EULAManager) save() error {
	// Ensure directory exists
	dir := filepath.Dir(m.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create EULA directory: %w", err)
	}

	data, err := json.MarshalIndent(m.records, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.filePath, data, 0644)
}
