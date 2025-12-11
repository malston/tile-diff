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

// IsAccepted checks if EULA has been accepted for a product release
func (m *EULAManager) IsAccepted(productSlug string) bool {
	// For backwards compatibility, check both product-only and any product-release keys
	if _, exists := m.records[productSlug]; exists {
		return true
	}

	// Check if any release of this product has accepted EULA
	for key := range m.records {
		if len(key) >= len(productSlug) && key[:len(productSlug)] == productSlug {
			return true
		}
	}
	return false
}

// IsAcceptedForRelease checks if EULA has been accepted for a specific release
func (m *EULAManager) IsAcceptedForRelease(productSlug, version string) bool {
	key := fmt.Sprintf("%s-%s", productSlug, version)
	_, exists := m.records[key]
	return exists
}

// Accept records EULA acceptance for a product release
func (m *EULAManager) Accept(productSlug, version, eulaURL string) error {
	// Key by product-version to track per-release acceptance
	key := fmt.Sprintf("%s-%s", productSlug, version)
	m.records[key] = EULARecord{
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
