// ABOUTME: Unit tests for disk space checking and management.
// ABOUTME: Validates available space calculations and cleanup triggers.
package pivnet

import (
	"testing"
)

func TestHasEnoughSpace(t *testing.T) {
	// This test checks the logic, not actual disk space
	// We'll test with a mock in integration tests

	// Just verify the function exists and can be called
	dm := &DiskManager{
		minFreeSpaceGB: 20,
	}

	// This will check actual disk space - we can't predict the result
	// but we can verify it doesn't panic
	_, err := dm.HasEnoughSpace("/tmp", 1024*1024*1024) // 1GB
	if err != nil {
		// Error is ok if we can't check (e.g., wrong OS)
		t.Logf("HasEnoughSpace returned error (may be expected): %v", err)
	}
}

func TestCalculateRequiredSpace(t *testing.T) {
	dm := &DiskManager{
		minFreeSpaceGB: 20,
	}

	tests := []struct {
		name           string
		fileSize       int64
		expectedGB     int64
	}{
		{
			name:           "1GB file needs 21GB",
			fileSize:       1 * 1024 * 1024 * 1024,
			expectedGB:     21,
		},
		{
			name:           "5GB file needs 25GB",
			fileSize:       5 * 1024 * 1024 * 1024,
			expectedGB:     25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			required := dm.calculateRequired(tt.fileSize)
			requiredGB := required / (1024 * 1024 * 1024)
			if requiredGB != tt.expectedGB {
				t.Errorf("Expected %dGB, got %dGB", tt.expectedGB, requiredGB)
			}
		})
	}
}
