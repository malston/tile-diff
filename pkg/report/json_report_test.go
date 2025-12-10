// ABOUTME: Unit tests for JSON report generation.
// ABOUTME: Validates machine-readable JSON output format.
package report

import (
	"encoding/json"
	"testing"

	"github.com/malston/tile-diff/pkg/compare"
	"github.com/malston/tile-diff/pkg/metadata"
)

func TestGenerateJSONReport(t *testing.T) {
	newReqProp := metadata.PropertyBlueprint{Name: "new_required", Type: "string", Optional: false}

	categorized := &CategorizedChanges{
		RequiredActions: []CategorizedChange{
			{
				ComparisonResult: compare.ComparisonResult{
					PropertyName: "new_required",
					ChangeType:   compare.PropertyAdded,
					NewProperty:  &newReqProp,
				},
				Category:       CategoryRequired,
				Recommendation: "Must configure",
			},
		},
	}

	jsonReport := GenerateJSONReport(categorized, "6.0.22", "10.2.5")

	// Verify it's valid JSON
	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonReport), &result)
	if err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	// Check version info
	if result["old_version"] != "6.0.22" {
		t.Error("Expected old_version in JSON")
	}
	if result["new_version"] != "10.2.5" {
		t.Error("Expected new_version in JSON")
	}

	// Check summary
	summary, ok := result["summary"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected summary object in JSON")
	}
	if summary["required_actions"] != 1.0 {
		t.Error("Expected 1 required action in summary")
	}

	// Check required_actions array
	actions, ok := result["required_actions"].([]interface{})
	if !ok {
		t.Fatal("Expected required_actions array in JSON")
	}
	if len(actions) != 1 {
		t.Errorf("Expected 1 action, got %d", len(actions))
	}
}
