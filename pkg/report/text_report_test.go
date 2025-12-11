// ABOUTME: Unit tests for text report generation.
// ABOUTME: Validates formatted text output with sections and recommendations.
package report

import (
	"strings"
	"testing"

	"github.com/malston/tile-diff/pkg/compare"
	"github.com/malston/tile-diff/pkg/metadata"
)

func TestGenerateTextReport(t *testing.T) {
	newReqProp := metadata.PropertyBlueprint{Name: "new_required", Type: "string", Optional: false}
	newOptProp := metadata.PropertyBlueprint{Name: "new_optional", Type: "boolean", Optional: true}

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
		Warnings: []CategorizedChange{
			{
				ComparisonResult: compare.ComparisonResult{
					PropertyName: "removed_prop",
					ChangeType:   compare.PropertyRemoved,
				},
				Category:       CategoryWarning,
				Recommendation: "Review and remove",
			},
		},
		Informational: []CategorizedChange{
			{
				ComparisonResult: compare.ComparisonResult{
					PropertyName: "new_optional",
					ChangeType:   compare.PropertyAdded,
					NewProperty:  &newOptProp,
				},
				Category:       CategoryInformational,
				Recommendation: "Optional",
			},
		},
	}

	report := GenerateTextReport(categorized, "6.0.22", "10.2.5")

	// Check for header
	if !strings.Contains(report, "Upgrade Analysis") {
		t.Error("Expected 'Upgrade Analysis' in report")
	}

	// Check for version info
	if !strings.Contains(report, "6.0.22") || !strings.Contains(report, "10.2.5") {
		t.Error("Expected version information in report")
	}

	// Check for Required Actions section
	if !strings.Contains(report, "REQUIRED ACTIONS") {
		t.Error("Expected 'REQUIRED ACTIONS' section")
	}
	if !strings.Contains(report, "new_required") {
		t.Error("Expected 'new_required' in Required Actions")
	}

	// Check for Warnings section
	if !strings.Contains(report, "WARNINGS") {
		t.Error("Expected 'WARNINGS' section")
	}
	if !strings.Contains(report, "removed_prop") {
		t.Error("Expected 'removed_prop' in Warnings")
	}

	// Check for Informational section
	if !strings.Contains(report, "INFORMATIONAL") {
		t.Error("Expected 'INFORMATIONAL' section")
	}
	if !strings.Contains(report, "new_optional") {
		t.Error("Expected 'new_optional' in Informational")
	}

	// Check for recommendations
	if !strings.Contains(report, "Must configure") {
		t.Error("Expected recommendation text in report")
	}
}

func TestGenerateTextReport_ShowsDefaultValues(t *testing.T) {
	boolDefault := true
	stringDefault := "default-value"

	categorized := &CategorizedChanges{
		Informational: []CategorizedChange{
			{
				ComparisonResult: compare.ComparisonResult{
					PropertyName: "optional_bool",
					ChangeType:   compare.PropertyAdded,
					NewProperty: &metadata.PropertyBlueprint{
						Name:     "optional_bool",
						Type:     "boolean",
						Optional: true,
						Default:  boolDefault,
					},
				},
				Category:       CategoryInformational,
				Recommendation: "Optional",
			},
			{
				ComparisonResult: compare.ComparisonResult{
					PropertyName: "optional_string",
					ChangeType:   compare.PropertyAdded,
					NewProperty: &metadata.PropertyBlueprint{
						Name:     "optional_string",
						Type:     "string",
						Optional: true,
						Default:  stringDefault,
					},
				},
				Category:       CategoryInformational,
				Recommendation: "Optional",
			},
			{
				ComparisonResult: compare.ComparisonResult{
					PropertyName: "optional_no_default",
					ChangeType:   compare.PropertyAdded,
					NewProperty: &metadata.PropertyBlueprint{
						Name:     "optional_no_default",
						Type:     "integer",
						Optional: true,
						Default:  nil,
					},
				},
				Category:       CategoryInformational,
				Recommendation: "Optional",
			},
		},
	}

	report := GenerateTextReport(categorized, "6.0.22", "10.2.5")

	// Check that default values are shown
	if !strings.Contains(report, "Default: true") {
		t.Error("Expected 'Default: true' for boolean property with default")
	}
	if !strings.Contains(report, "Default: default-value") {
		t.Error("Expected 'Default: default-value' for string property with default")
	}

	// Check that properties without defaults don't show Default line
	lines := strings.Split(report, "\n")
	inNoDefaultSection := false
	for _, line := range lines {
		if strings.Contains(line, "optional_no_default") {
			inNoDefaultSection = true
		}
		if inNoDefaultSection && strings.Contains(line, "Note:") {
			// We've reached the end of this property section
			break
		}
		if inNoDefaultSection && strings.Contains(line, "Default:") {
			t.Error("Property without default should not show Default line")
		}
	}
}
