// ABOUTME: Unit tests for EULA acceptance tracking.
// ABOUTME: Validates EULA persistence and acceptance logic.
package pivnet

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEULAManager(t *testing.T) {
	// Create temp directory for testing
	tempDir := t.TempDir()
	eulaFile := filepath.Join(tempDir, "eulas.json")

	mgr := NewEULAManager(eulaFile)

	// Test: EULA not accepted initially
	if mgr.IsAccepted("cf") {
		t.Error("Expected EULA not accepted for cf")
	}

	// Test: Accept EULA
	err := mgr.Accept("cf", "6.0.22+LTS-T", "https://example.com/eula")
	if err != nil {
		t.Fatalf("Failed to accept EULA: %v", err)
	}

	// Test: EULA now accepted
	if !mgr.IsAccepted("cf") {
		t.Error("Expected EULA accepted for cf after Accept()")
	}

	// Test: Different product not accepted
	if mgr.IsAccepted("p-redis") {
		t.Error("Expected EULA not accepted for p-redis")
	}

	// Test: Persistence - create new manager with same file
	mgr2 := NewEULAManager(eulaFile)
	if !mgr2.IsAccepted("cf") {
		t.Error("Expected EULA still accepted after reload")
	}
}

func TestEULAManagerWithExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	eulaFile := filepath.Join(tempDir, "eulas.json")

	// Write existing EULA file
	existingData := `{
  "cf": {
    "accepted_at": "2024-12-10T10:00:00Z",
    "release_version": "6.0.22+LTS-T",
    "eula_url": "https://example.com/eula"
  }
}`
	err := os.WriteFile(eulaFile, []byte(existingData), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	mgr := NewEULAManager(eulaFile)
	if !mgr.IsAccepted("cf") {
		t.Error("Expected EULA accepted from existing file")
	}
}
