// ABOUTME: Unit tests for high-level comparison orchestration.
// ABOUTME: Validates end-to-end comparison workflow combining all detectors.
package compare

import (
	"testing"

	"github.com/malston/tile-diff/pkg/metadata"
)

func TestCompareMetadata(t *testing.T) {
	oldMetadata := &metadata.TileMetadata{
		PropertyBlueprints: []metadata.PropertyBlueprint{
			{Name: "existing_prop", Type: "string", Configurable: true},
			{Name: "removed_prop", Type: "boolean", Configurable: true},
			{Name: "changed_prop", Type: "string", Configurable: true},
			{Name: "system_prop", Type: "string", Configurable: false},
		},
	}

	newMetadata := &metadata.TileMetadata{
		PropertyBlueprints: []metadata.PropertyBlueprint{
			{Name: "existing_prop", Type: "string", Configurable: true},
			{Name: "new_prop", Type: "integer", Configurable: true},
			{Name: "changed_prop", Type: "integer", Configurable: true},
			{Name: "system_prop", Type: "string", Configurable: false},
		},
	}

	results := CompareMetadata(oldMetadata, newMetadata, true)

	// Verify counts
	if results.TotalOldProps != 4 {
		t.Errorf("Expected TotalOldProps 4, got %d", results.TotalOldProps)
	}
	if results.TotalNewProps != 4 {
		t.Errorf("Expected TotalNewProps 4, got %d", results.TotalNewProps)
	}

	// Should have 1 added (configurable only)
	if len(results.Added) != 1 {
		t.Errorf("Expected 1 added property, got %d", len(results.Added))
	}

	// Should have 1 removed (configurable only)
	if len(results.Removed) != 1 {
		t.Errorf("Expected 1 removed property, got %d", len(results.Removed))
	}

	// Should have 1 changed (type change)
	if len(results.Changed) != 1 {
		t.Errorf("Expected 1 changed property, got %d", len(results.Changed))
	}
}
