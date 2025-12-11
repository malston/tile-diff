# Phase 3: Actionable Reports Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Generate actionable upgrade reports that show operators what they must do, should review, and can ignore.

**Architecture:** Cross-reference comparison results with current Ops Manager configuration to filter relevant changes. Categorize changes by severity (Required Actions, Warnings, Informational). Generate formatted reports with specific recommendations for each change.

**Tech Stack:** Go 1.21+, existing metadata/api/compare packages, TDD with table-driven tests

---

## Task 1: Current Config Parser

**Files:**
- Create: `pkg/report/config_parser.go`
- Create: `pkg/report/config_parser_test.go`

**Step 1: Write the failing test**

Create `pkg/report/config_parser_test.go`:

```go
// ABOUTME: Unit tests for parsing Ops Manager current configuration.
// ABOUTME: Validates extraction of configured properties from API response.
package report

import (
	"testing"

	"github.com/malston/tile-diff/pkg/api"
)

func TestParseCurrentConfig(t *testing.T) {
	apiResponse := &api.PropertiesResponse{
		Properties: map[string]api.Property{
			".properties.configured_prop": {
				Type:         "string",
				Configurable: true,
				Credential:   false,
				Value:        "custom_value",
				Optional:     false,
			},
			".properties.default_prop": {
				Type:         "boolean",
				Configurable: true,
				Credential:   false,
				Value:        true,
				Optional:     true,
			},
			".properties.system_prop": {
				Type:         "string",
				Configurable: false,
				Credential:   false,
				Value:        "system_value",
				Optional:     false,
			},
		},
	}

	config := ParseCurrentConfig(apiResponse)

	// Should have all 3 properties
	if len(config.Properties) != 3 {
		t.Errorf("Expected 3 properties, got %d", len(config.Properties))
	}

	// Check configured property exists
	prop, exists := config.Properties["configured_prop"]
	if !exists {
		t.Error("Expected configured_prop to exist")
	}
	if prop.Value != "custom_value" {
		t.Errorf("Expected value 'custom_value', got '%v'", prop.Value)
	}

	// Count configurable properties
	configurableCount := 0
	for _, p := range config.Properties {
		if p.Configurable {
			configurableCount++
		}
	}
	if configurableCount != 2 {
		t.Errorf("Expected 2 configurable properties, got %d", configurableCount)
	}
}

func TestIsPropertyConfigured(t *testing.T) {
	tests := []struct {
		name       string
		prop       ConfiguredProperty
		expected   bool
	}{
		{
			name: "non-optional with value",
			prop: ConfiguredProperty{
				Name:         "test",
				Configurable: true,
				Optional:     false,
				Value:        "value",
			},
			expected: true,
		},
		{
			name: "optional with non-nil value",
			prop: ConfiguredProperty{
				Name:         "test",
				Configurable: true,
				Optional:     true,
				Value:        "value",
			},
			expected: true,
		},
		{
			name: "optional with nil value",
			prop: ConfiguredProperty{
				Name:         "test",
				Configurable: true,
				Optional:     true,
				Value:        nil,
			},
			expected: false,
		},
		{
			name: "non-configurable",
			prop: ConfiguredProperty{
				Name:         "test",
				Configurable: false,
				Optional:     false,
				Value:        "value",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.prop.IsConfigured()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
cd /Users/markalston/workspace/tile-diff
go test ./pkg/report/... -v
```

Expected: FAIL - package report is not in GOROOT or GOPATH

**Step 3: Write minimal implementation**

Create `pkg/report/config_parser.go`:

```go
// ABOUTME: Parses current Ops Manager configuration from API responses.
// ABOUTME: Extracts configured properties for cross-reference with comparison results.
package report

import (
	"strings"

	"github.com/malston/tile-diff/pkg/api"
)

// ConfiguredProperty represents a property from current Ops Manager configuration
type ConfiguredProperty struct {
	Name         string
	Type         string
	Configurable bool
	Credential   bool
	Optional     bool
	Value        interface{}
}

// CurrentConfig represents the current Ops Manager configuration
type CurrentConfig struct {
	Properties map[string]ConfiguredProperty
}

// ParseCurrentConfig converts API response to CurrentConfig
func ParseCurrentConfig(apiResponse *api.PropertiesResponse) *CurrentConfig {
	config := &CurrentConfig{
		Properties: make(map[string]ConfiguredProperty),
	}

	for fullPath, prop := range apiResponse.Properties {
		// Extract property name from full path (e.g., ".properties.name" -> "name")
		name := extractPropertyName(fullPath)

		config.Properties[name] = ConfiguredProperty{
			Name:         name,
			Type:         prop.Type,
			Configurable: prop.Configurable,
			Credential:   prop.Credential,
			Optional:     prop.Optional,
			Value:        prop.Value,
		}
	}

	return config
}

// IsConfigured returns true if the property is actually configured (not just present)
func (p *ConfiguredProperty) IsConfigured() bool {
	// Non-configurable properties don't count
	if !p.Configurable {
		return false
	}

	// Optional properties with nil values are not configured
	if p.Optional && p.Value == nil {
		return false
	}

	return true
}

// extractPropertyName extracts the property name from a full API path
func extractPropertyName(fullPath string) string {
	// Remove common prefixes
	name := strings.TrimPrefix(fullPath, ".properties.")
	name = strings.TrimPrefix(name, ".cloud_controller.")
	name = strings.TrimPrefix(name, ".diego_brain.")
	name = strings.TrimPrefix(name, ".mysql.")
	name = strings.TrimPrefix(name, ".")

	return name
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test ./pkg/report/... -v
```

Expected: PASS - Both tests pass

**Step 5: Commit**

```bash
git add pkg/report/config_parser.go pkg/report/config_parser_test.go
git commit -m "feat(report): add current config parser"
```

---

## Task 2: Change Relevance Filter

**Files:**
- Create: `pkg/report/filter.go`
- Create: `pkg/report/filter_test.go`

**Step 1: Write the failing test**

Create `pkg/report/filter_test.go`:

```go
// ABOUTME: Unit tests for filtering changes by relevance to current config.
// ABOUTME: Validates which property changes affect the deployed configuration.
package report

import (
	"testing"

	"github.com/malston/tile-diff/pkg/compare"
	"github.com/malston/tile-diff/pkg/metadata"
)

func TestFilterRelevantChanges(t *testing.T) {
	currentConfig := &CurrentConfig{
		Properties: map[string]ConfiguredProperty{
			"configured_prop": {
				Name:         "configured_prop",
				Configurable: true,
				Value:        "value",
			},
			"default_prop": {
				Name:         "default_prop",
				Configurable: true,
				Optional:     true,
				Value:        nil, // Using default
			},
		},
	}

	oldProp := metadata.PropertyBlueprint{Name: "configured_prop", Type: "string"}
	newProp := metadata.PropertyBlueprint{Name: "configured_prop", Type: "integer"}

	allChanges := &compare.ComparisonResults{
		Added: []compare.ComparisonResult{
			{PropertyName: "new_prop", ChangeType: compare.PropertyAdded},
		},
		Removed: []compare.ComparisonResult{
			{PropertyName: "configured_prop", ChangeType: compare.PropertyRemoved, OldProperty: &oldProp},
			{PropertyName: "unused_prop", ChangeType: compare.PropertyRemoved},
		},
		Changed: []compare.ComparisonResult{
			{PropertyName: "configured_prop", ChangeType: compare.TypeChanged, OldProperty: &oldProp, NewProperty: &newProp},
		},
	}

	filtered := FilterRelevantChanges(allChanges, currentConfig)

	// New properties are always relevant
	if len(filtered.Added) != 1 {
		t.Errorf("Expected 1 added property, got %d", len(filtered.Added))
	}

	// Only configured properties that are removed are relevant
	if len(filtered.Removed) != 1 {
		t.Errorf("Expected 1 relevant removed property, got %d", len(filtered.Removed))
	}
	if filtered.Removed[0].PropertyName != "configured_prop" {
		t.Errorf("Expected configured_prop, got %s", filtered.Removed[0].PropertyName)
	}

	// Changed properties affecting configured values are relevant
	if len(filtered.Changed) != 1 {
		t.Errorf("Expected 1 relevant changed property, got %d", len(filtered.Changed))
	}
}

func TestIsChangeRelevant(t *testing.T) {
	tests := []struct {
		name          string
		changeType    compare.ChangeType
		propertyName  string
		isConfigured  bool
		expected      bool
	}{
		{
			name:         "added property always relevant",
			changeType:   compare.PropertyAdded,
			propertyName: "new_prop",
			isConfigured: false,
			expected:     true,
		},
		{
			name:         "removed configured property relevant",
			changeType:   compare.PropertyRemoved,
			propertyName: "configured_prop",
			isConfigured: true,
			expected:     true,
		},
		{
			name:         "removed unconfigured property not relevant",
			changeType:   compare.PropertyRemoved,
			propertyName: "unused_prop",
			isConfigured: false,
			expected:     false,
		},
		{
			name:         "changed configured property relevant",
			changeType:   compare.TypeChanged,
			propertyName: "configured_prop",
			isConfigured: true,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &CurrentConfig{
				Properties: make(map[string]ConfiguredProperty),
			}
			if tt.isConfigured {
				config.Properties[tt.propertyName] = ConfiguredProperty{
					Name:         tt.propertyName,
					Configurable: true,
					Value:        "value",
				}
			}

			result := isChangeRelevant(tt.changeType, tt.propertyName, config)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./pkg/report/... -v -run TestFilter
```

Expected: FAIL - undefined: FilterRelevantChanges

**Step 3: Write minimal implementation**

Create `pkg/report/filter.go`:

```go
// ABOUTME: Filters comparison results to show only relevant changes.
// ABOUTME: Determines which property changes affect the current deployed configuration.
package report

import "github.com/malston/tile-diff/pkg/compare"

// FilterRelevantChanges filters comparison results to include only changes relevant to current config
func FilterRelevantChanges(allChanges *compare.ComparisonResults, currentConfig *CurrentConfig) *compare.ComparisonResults {
	filtered := &compare.ComparisonResults{
		TotalOldProps:    allChanges.TotalOldProps,
		TotalNewProps:    allChanges.TotalNewProps,
		ConfigurableOnly: allChanges.ConfigurableOnly,
	}

	// Filter added properties (always relevant)
	filtered.Added = allChanges.Added

	// Filter removed properties (only if configured)
	for _, change := range allChanges.Removed {
		if isChangeRelevant(change.ChangeType, change.PropertyName, currentConfig) {
			filtered.Removed = append(filtered.Removed, change)
		}
	}

	// Filter changed properties (only if configured)
	for _, change := range allChanges.Changed {
		if isChangeRelevant(change.ChangeType, change.PropertyName, currentConfig) {
			filtered.Changed = append(filtered.Changed, change)
		}
	}

	return filtered
}

// isChangeRelevant determines if a property change is relevant to current config
func isChangeRelevant(changeType compare.ChangeType, propertyName string, currentConfig *CurrentConfig) bool {
	// New properties are always relevant (might become required)
	if changeType == compare.PropertyAdded {
		return true
	}

	// For removals and changes, check if property is currently configured
	prop, exists := currentConfig.Properties[propertyName]
	if !exists {
		return false
	}

	return prop.IsConfigured()
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test ./pkg/report/... -v
```

Expected: PASS - All tests pass

**Step 5: Commit**

```bash
git add pkg/report/filter.go pkg/report/filter_test.go
git commit -m "feat(report): add change relevance filter"
```

---

## Task 3: Change Categorization

**Files:**
- Create: `pkg/report/categorizer.go`
- Create: `pkg/report/categorizer_test.go`

**Step 1: Write the failing test**

Create `pkg/report/categorizer_test.go`:

```go
// ABOUTME: Unit tests for categorizing changes by severity.
// ABOUTME: Validates classification into Required Actions, Warnings, and Informational.
package report

import (
	"testing"

	"github.com/malston/tile-diff/pkg/compare"
	"github.com/malston/tile-diff/pkg/metadata"
)

func TestCategorizeChanges(t *testing.T) {
	newReqProp := metadata.PropertyBlueprint{
		Name:     "new_required",
		Type:     "string",
		Optional: false,
	}
	newOptProp := metadata.PropertyBlueprint{
		Name:     "new_optional",
		Type:     "boolean",
		Optional: true,
	}
	removedProp := metadata.PropertyBlueprint{
		Name: "removed_prop",
		Type: "string",
	}
	oldTypeProp := metadata.PropertyBlueprint{
		Name: "type_changed",
		Type: "string",
	}
	newTypeProp := metadata.PropertyBlueprint{
		Name: "type_changed",
		Type: "integer",
	}

	changes := &compare.ComparisonResults{
		Added: []compare.ComparisonResult{
			{PropertyName: "new_required", NewProperty: &newReqProp, ChangeType: compare.PropertyAdded},
			{PropertyName: "new_optional", NewProperty: &newOptProp, ChangeType: compare.PropertyAdded},
		},
		Removed: []compare.ComparisonResult{
			{PropertyName: "removed_prop", OldProperty: &removedProp, ChangeType: compare.PropertyRemoved},
		},
		Changed: []compare.ComparisonResult{
			{PropertyName: "type_changed", OldProperty: &oldTypeProp, NewProperty: &newTypeProp, ChangeType: compare.TypeChanged},
		},
	}

	categorized := CategorizeChanges(changes)

	// New required properties = Required Actions
	if len(categorized.RequiredActions) != 1 {
		t.Errorf("Expected 1 required action, got %d", len(categorized.RequiredActions))
	}
	if categorized.RequiredActions[0].PropertyName != "new_required" {
		t.Error("Expected new_required in Required Actions")
	}

	// Removed and type changed = Warnings
	if len(categorized.Warnings) != 2 {
		t.Errorf("Expected 2 warnings, got %d", len(categorized.Warnings))
	}

	// New optional = Informational
	if len(categorized.Informational) != 1 {
		t.Errorf("Expected 1 informational, got %d", len(categorized.Informational))
	}
	if categorized.Informational[0].PropertyName != "new_optional" {
		t.Error("Expected new_optional in Informational")
	}
}

func TestDetermineCategory(t *testing.T) {
	tests := []struct {
		name       string
		change     compare.ComparisonResult
		expected   Category
	}{
		{
			name: "new required property",
			change: compare.ComparisonResult{
				ChangeType:  compare.PropertyAdded,
				NewProperty: &metadata.PropertyBlueprint{Optional: false},
			},
			expected: CategoryRequired,
		},
		{
			name: "new optional property",
			change: compare.ComparisonResult{
				ChangeType:  compare.PropertyAdded,
				NewProperty: &metadata.PropertyBlueprint{Optional: true},
			},
			expected: CategoryInformational,
		},
		{
			name: "removed property",
			change: compare.ComparisonResult{
				ChangeType: compare.PropertyRemoved,
			},
			expected: CategoryWarning,
		},
		{
			name: "type changed",
			change: compare.ComparisonResult{
				ChangeType: compare.TypeChanged,
			},
			expected: CategoryWarning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineCategory(tt.change)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./pkg/report/... -v -run TestCategorize
```

Expected: FAIL - undefined: CategorizeChanges

**Step 3: Write minimal implementation**

Create `pkg/report/categorizer.go`:

```go
// ABOUTME: Categorizes property changes by severity and required action.
// ABOUTME: Classifies changes into Required Actions, Warnings, and Informational.
package report

import "github.com/malston/tile-diff/pkg/compare"

// Category represents the severity/type of a change
type Category string

const (
	CategoryRequired      Category = "required"
	CategoryWarning       Category = "warning"
	CategoryInformational Category = "informational"
)

// CategorizedChange represents a change with its severity category
type CategorizedChange struct {
	compare.ComparisonResult
	Category      Category
	Recommendation string
}

// CategorizedChanges holds changes grouped by category
type CategorizedChanges struct {
	RequiredActions []CategorizedChange
	Warnings        []CategorizedChange
	Informational   []CategorizedChange
}

// CategorizeChanges classifies comparison results into severity categories
func CategorizeChanges(changes *compare.ComparisonResults) *CategorizedChanges {
	categorized := &CategorizedChanges{}

	// Categorize added properties
	for _, change := range changes.Added {
		cat := determineCategory(change)
		catChange := CategorizedChange{
			ComparisonResult: change,
			Category:         cat,
			Recommendation:   generateRecommendation(change, cat),
		}

		switch cat {
		case CategoryRequired:
			categorized.RequiredActions = append(categorized.RequiredActions, catChange)
		case CategoryInformational:
			categorized.Informational = append(categorized.Informational, catChange)
		}
	}

	// Categorize removed properties
	for _, change := range changes.Removed {
		cat := determineCategory(change)
		catChange := CategorizedChange{
			ComparisonResult: change,
			Category:         cat,
			Recommendation:   generateRecommendation(change, cat),
		}
		categorized.Warnings = append(categorized.Warnings, catChange)
	}

	// Categorize changed properties
	for _, change := range changes.Changed {
		cat := determineCategory(change)
		catChange := CategorizedChange{
			ComparisonResult: change,
			Category:         cat,
			Recommendation:   generateRecommendation(change, cat),
		}
		categorized.Warnings = append(categorized.Warnings, catChange)
	}

	return categorized
}

// determineCategory determines the severity category for a change
func determineCategory(change compare.ComparisonResult) Category {
	switch change.ChangeType {
	case compare.PropertyAdded:
		// Required if not optional
		if change.NewProperty != nil && !change.NewProperty.Optional {
			return CategoryRequired
		}
		return CategoryInformational

	case compare.PropertyRemoved:
		// Removal of configured property is a warning
		return CategoryWarning

	case compare.TypeChanged, compare.OptionalityChanged:
		// Type or optionality changes are warnings
		return CategoryWarning

	default:
		return CategoryInformational
	}
}

// generateRecommendation creates a recommendation string for a change
func generateRecommendation(change compare.ComparisonResult, category Category) string {
	switch category {
	case CategoryRequired:
		return "Must configure this property before upgrading"
	case CategoryWarning:
		if change.ChangeType == compare.PropertyRemoved {
			return "Property will be ignored after upgrade - review and remove from config"
		}
		return "Review this change and verify compatibility"
	case CategoryInformational:
		return "Optional - review for potential improvements"
	default:
		return ""
	}
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test ./pkg/report/... -v
```

Expected: PASS - All tests pass

**Step 5: Commit**

```bash
git add pkg/report/categorizer.go pkg/report/categorizer_test.go
git commit -m "feat(report): add change categorization by severity"
```

---

## Task 4: Text Report Generator

**Files:**
- Create: `pkg/report/text_report.go`
- Create: `pkg/report/text_report_test.go`

**Step 1: Write the failing test**

Create `pkg/report/text_report_test.go`:

```go
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
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./pkg/report/... -v -run TestGenerateText
```

Expected: FAIL - undefined: GenerateTextReport

**Step 3: Write minimal implementation**

Create `pkg/report/text_report.go`:

```go
// ABOUTME: Generates human-readable text reports for tile upgrades.
// ABOUTME: Formats categorized changes with sections and recommendations.
package report

import (
	"fmt"
	"strings"
)

// GenerateTextReport creates a formatted text report from categorized changes
func GenerateTextReport(categorized *CategorizedChanges, oldVersion, newVersion string) string {
	var report strings.Builder

	// Header
	report.WriteString("================================================================================\n")
	report.WriteString("                        TAS Tile Upgrade Analysis\n")
	report.WriteString("================================================================================\n\n")
	report.WriteString(fmt.Sprintf("Old Version: %s\n", oldVersion))
	report.WriteString(fmt.Sprintf("New Version: %s\n\n", newVersion))

	// Summary
	totalChanges := len(categorized.RequiredActions) + len(categorized.Warnings) + len(categorized.Informational)
	report.WriteString(fmt.Sprintf("Total Changes: %d\n", totalChanges))
	report.WriteString(fmt.Sprintf("  Required Actions: %d\n", len(categorized.RequiredActions)))
	report.WriteString(fmt.Sprintf("  Warnings: %d\n", len(categorized.Warnings)))
	report.WriteString(fmt.Sprintf("  Informational: %d\n\n", len(categorized.Informational)))

	// Required Actions
	if len(categorized.RequiredActions) > 0 {
		report.WriteString("================================================================================\n")
		report.WriteString("🚨 REQUIRED ACTIONS\n")
		report.WriteString("================================================================================\n\n")
		report.WriteString("These changes MUST be addressed before upgrading:\n\n")

		for i, change := range categorized.RequiredActions {
			report.WriteString(fmt.Sprintf("%d. %s\n", i+1, change.PropertyName))
			report.WriteString(fmt.Sprintf("   Type: %s\n", change.NewProperty.Type))
			report.WriteString(fmt.Sprintf("   Action: %s\n", change.Recommendation))
			report.WriteString("\n")
		}
	}

	// Warnings
	if len(categorized.Warnings) > 0 {
		report.WriteString("================================================================================\n")
		report.WriteString("⚠️  WARNINGS\n")
		report.WriteString("================================================================================\n\n")
		report.WriteString("These changes should be reviewed:\n\n")

		for i, change := range categorized.Warnings {
			report.WriteString(fmt.Sprintf("%d. %s\n", i+1, change.PropertyName))
			report.WriteString(fmt.Sprintf("   Change: %s\n", change.Description))
			report.WriteString(fmt.Sprintf("   Recommendation: %s\n", change.Recommendation))
			report.WriteString("\n")
		}
	}

	// Informational
	if len(categorized.Informational) > 0 {
		report.WriteString("================================================================================\n")
		report.WriteString("ℹ️  INFORMATIONAL\n")
		report.WriteString("================================================================================\n\n")
		report.WriteString("New optional features available:\n\n")

		for i, change := range categorized.Informational {
			report.WriteString(fmt.Sprintf("%d. %s\n", i+1, change.PropertyName))
			if change.NewProperty != nil {
				report.WriteString(fmt.Sprintf("   Type: %s\n", change.NewProperty.Type))
			}
			report.WriteString(fmt.Sprintf("   Note: %s\n", change.Recommendation))
			report.WriteString("\n")
		}
	}

	return report.String()
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test ./pkg/report/... -v
```

Expected: PASS - All tests pass

**Step 5: Commit**

```bash
git add pkg/report/text_report.go pkg/report/text_report_test.go
git commit -m "feat(report): add text report generator"
```

---

## Task 5: JSON Report Generator

**Files:**
- Create: `pkg/report/json_report.go`
- Create: `pkg/report/json_report_test.go`

**Step 1: Write the failing test**

Create `pkg/report/json_report_test.go`:

```go
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
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./pkg/report/... -v -run TestGenerateJSON
```

Expected: FAIL - undefined: GenerateJSONReport

**Step 3: Write minimal implementation**

Create `pkg/report/json_report.go`:

```go
// ABOUTME: Generates machine-readable JSON reports for tile upgrades.
// ABOUTME: Provides structured data format for automation and tooling integration.
package report

import (
	"encoding/json"
)

// JSONReport represents the JSON report structure
type JSONReport struct {
	OldVersion      string                `json:"old_version"`
	NewVersion      string                `json:"new_version"`
	Summary         JSONSummary           `json:"summary"`
	RequiredActions []JSONChange          `json:"required_actions"`
	Warnings        []JSONChange          `json:"warnings"`
	Informational   []JSONChange          `json:"informational"`
}

// JSONSummary contains summary statistics
type JSONSummary struct {
	TotalChanges    int `json:"total_changes"`
	RequiredActions int `json:"required_actions"`
	Warnings        int `json:"warnings"`
	Informational   int `json:"informational"`
}

// JSONChange represents a single change in JSON format
type JSONChange struct {
	PropertyName   string `json:"property_name"`
	ChangeType     string `json:"change_type"`
	Category       string `json:"category"`
	Description    string `json:"description"`
	Recommendation string `json:"recommendation"`
	PropertyType   string `json:"property_type,omitempty"`
}

// GenerateJSONReport creates a JSON-formatted report from categorized changes
func GenerateJSONReport(categorized *CategorizedChanges, oldVersion, newVersion string) string {
	report := JSONReport{
		OldVersion: oldVersion,
		NewVersion: newVersion,
		Summary: JSONSummary{
			TotalChanges:    len(categorized.RequiredActions) + len(categorized.Warnings) + len(categorized.Informational),
			RequiredActions: len(categorized.RequiredActions),
			Warnings:        len(categorized.Warnings),
			Informational:   len(categorized.Informational),
		},
	}

	// Convert required actions
	for _, change := range categorized.RequiredActions {
		report.RequiredActions = append(report.RequiredActions, toJSONChange(change))
	}

	// Convert warnings
	for _, change := range categorized.Warnings {
		report.Warnings = append(report.Warnings, toJSONChange(change))
	}

	// Convert informational
	for _, change := range categorized.Informational {
		report.Informational = append(report.Informational, toJSONChange(change))
	}

	jsonBytes, _ := json.MarshalIndent(report, "", "  ")
	return string(jsonBytes)
}

// toJSONChange converts a CategorizedChange to JSONChange
func toJSONChange(change CategorizedChange) JSONChange {
	jsonChange := JSONChange{
		PropertyName:   change.PropertyName,
		ChangeType:     string(change.ChangeType),
		Category:       string(change.Category),
		Description:    change.Description,
		Recommendation: change.Recommendation,
	}

	if change.NewProperty != nil {
		jsonChange.PropertyType = change.NewProperty.Type
	} else if change.OldProperty != nil {
		jsonChange.PropertyType = change.OldProperty.Type
	}

	return jsonChange
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test ./pkg/report/... -v
```

Expected: PASS - All tests pass

**Step 5: Commit**

```bash
git add pkg/report/json_report.go pkg/report/json_report_test.go
git commit -m "feat(report): add JSON report generator"
```

---

## Task 6: CLI Integration

**Files:**
- Modify: `cmd/tile-diff/main.go`

**Step 1: Add report generation to CLI**

Modify `cmd/tile-diff/main.go` to add after comparison (around where comparison results are displayed):

```go
import (
	// ... existing imports ...
	"github.com/malston/tile-diff/pkg/report"
)

// Add new flags at the top with other flag definitions
var (
	reportFormat = flag.String("format", "text", "Output format: text or json")
	// ... existing flags ...
)

// After comparison logic, add:

// Generate report if API config is available
if *productGUID != "" && *opsManagerURL != "" && *username != "" && *password != "" {
	fmt.Printf("\nGenerating actionable report...\n")

	// Get current config
	client := api.NewClient(*opsManagerURL, *username, *password, *skipSSL)
	apiConfig, err := client.GetProperties(*productGUID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not fetch current config: %v\n", err)
	} else {
		// Parse current config
		currentConfig := report.ParseCurrentConfig(apiConfig)

		// Filter relevant changes
		filtered := report.FilterRelevantChanges(results, currentConfig)

		// Categorize changes
		categorized := report.CategorizeChanges(filtered)

		// Generate report based on format
		switch *reportFormat {
		case "json":
			jsonReport := report.GenerateJSONReport(categorized, "old", "new")
			fmt.Println(jsonReport)
		default:
			textReport := report.GenerateTextReport(categorized, "old", "new")
			fmt.Println(textReport)
		}
	}
} else {
	fmt.Println("\nNote: Provide Ops Manager credentials for actionable report with current config analysis")
}
```

**Step 2: Build and test**

Run:
```bash
cd /Users/markalston/workspace/tile-diff
make build
./tile-diff --old-tile /tmp/elastic-runtime/srt-6.0.22-build.2.pivotal \
  --new-tile /tmp/elastic-runtime/srt-10.2.5-build.2.pivotal \
  --product-guid cf-xxxxx \
  --ops-manager-url https://opsman.tas.vcf.lab \
  --username admin \
  --password pass \
  --skip-ssl-validation
```

Expected: Displays formatted actionable report with Required Actions, Warnings, Informational

**Step 3: Test JSON format**

Run:
```bash
./tile-diff --old-tile /tmp/elastic-runtime/srt-6.0.22-build.2.pivotal \
  --new-tile /tmp/elastic-runtime/srt-10.2.5-build.2.pivotal \
  --format json \
  --product-guid cf-xxxxx \
  --ops-manager-url https://opsman.tas.vcf.lab \
  --username admin \
  --password pass \
  --skip-ssl-validation
```

Expected: Displays JSON-formatted report

**Step 4: Commit**

```bash
git add cmd/tile-diff/main.go
git commit -m "feat(cli): integrate actionable report generation

Adds --format flag (text or json) and generates categorized reports:
- Parses current Ops Manager config
- Filters changes by relevance
- Categorizes into Required/Warning/Info
- Displays formatted actionable reports

Completes Phase 3: Actionable reporting"
```

---

## Task 7: Integration Test

**Files:**
- Modify: `test/comparison_test.go`

**Step 1: Add report integration test**

Add to `test/comparison_test.go`:

```go
func TestActionableReportGeneration(t *testing.T) {
	oldTilePath := "/tmp/elastic-runtime/srt-6.0.22-build.2.pivotal"
	newTilePath := "/tmp/elastic-runtime/srt-10.2.5-build.2.pivotal"

	// Load tiles
	oldMetadata, err := metadata.LoadFromFile(oldTilePath)
	if err != nil {
		t.Skipf("Skipping integration test (old tile not found): %v", err)
		return
	}

	newMetadata, err := metadata.LoadFromFile(newTilePath)
	if err != nil {
		t.Skipf("Skipping integration test (new tile not found): %v", err)
		return
	}

	// Compare
	results := compare.CompareMetadata(oldMetadata, newMetadata, true)

	// Create mock current config
	mockConfig := &report.CurrentConfig{
		Properties: make(map[string]report.ConfiguredProperty),
	}

	// Filter and categorize
	filtered := report.FilterRelevantChanges(results, mockConfig)
	categorized := report.CategorizeChanges(filtered)

	// Generate reports
	textReport := report.GenerateTextReport(categorized, "6.0.22", "10.2.5")
	jsonReport := report.GenerateJSONReport(categorized, "6.0.22", "10.2.5")

	// Verify text report
	if !strings.Contains(textReport, "Upgrade Analysis") {
		t.Error("Expected 'Upgrade Analysis' in text report")
	}

	// Verify JSON report
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonReport), &jsonData); err != nil {
		t.Errorf("Invalid JSON report: %v", err)
	}

	t.Logf("Generated text report (%d bytes)", len(textReport))
	t.Logf("Generated JSON report (%d bytes)", len(jsonReport))
	t.Logf("Required Actions: %d", len(categorized.RequiredActions))
	t.Logf("Warnings: %d", len(categorized.Warnings))
	t.Logf("Informational: %d", len(categorized.Informational))
}
```

**Step 2: Run integration test**

Run:
```bash
go test -tags=integration -v ./test/... -run TestActionable
```

Expected: PASS - Shows report generation stats

**Step 3: Commit**

```bash
git add test/comparison_test.go
git commit -m "test: add integration test for actionable reports"
```

---

## Task 8: Documentation Update

**Files:**
- Modify: `README.md`
- Create: `docs/phase-3-completion.md`

**Step 1: Update README**

Modify README.md Status section:

```markdown
## Status

✅ **Phase 1 Complete** - Extraction & parsing
✅ **Phase 2 Complete** - Property comparison
✅ **Phase 3 Complete** - Actionable reports

Full upgrade analysis with:
- Current config cross-reference
- Change categorization (Required/Warning/Info)
- Formatted reports (text and JSON)
- Specific recommendations per change

## Quick Start

### Compare Tiles with Actionable Report

```bash
./tile-diff \
  --old-tile srt-6.0.22.pivotal \
  --new-tile srt-10.2.5.pivotal \
  --product-guid cf-xxxxx \
  --ops-manager-url https://opsman.tas.vcf.lab \
  --username admin \
  --password your-password \
  --skip-ssl-validation
```

### JSON Output

```bash
./tile-diff \
  --old-tile srt-6.0.22.pivotal \
  --new-tile srt-10.2.5.pivotal \
  --format json
```
```

**Step 2: Create completion report**

Create `docs/phase-3-completion.md`:

```markdown
# Phase 3 Completion Report

**Date:** 2025-12-10
**Status:** ✅ Complete

## Summary

Phase 3 successfully implements actionable reporting with current config cross-reference.

## Deliverables

### Report Package (`pkg/report/`)
- `config_parser.go`: Parse Ops Manager API config
- `filter.go`: Filter changes by relevance
- `categorizer.go`: Categorize by severity
- `text_report.go`: Human-readable reports
- `json_report.go`: Machine-readable reports
- Full unit test coverage

### CLI Integration
- `--format` flag (text or json)
- Automatic current config fetch
- Categorized output

## Features

- **Required Actions**: Must-do items before upgrade
- **Warnings**: Changes needing review
- **Informational**: Optional new features
- **Recommendations**: Specific guidance per change

## Sample Output

```
================================================================================
                        TAS Tile Upgrade Analysis
================================================================================

Old Version: 6.0.22
New Version: 10.2.5

Total Changes: 10
  Required Actions: 2
  Warnings: 3
  Informational: 5

================================================================================
🚨 REQUIRED ACTIONS
================================================================================

1. new_security_property
   Type: boolean
   Action: Must configure this property before upgrading
...
```

## Success Criteria

All met:
- ✅ Parse current Ops Manager config
- ✅ Filter changes by relevance
- ✅ Categorize changes by severity
- ✅ Generate text reports
- ✅ Generate JSON reports
- ✅ CLI integration with format flag
- ✅ All tests passing
```

**Step 3: Commit**

```bash
git add README.md docs/phase-3-completion.md
git commit -m "docs: update documentation for Phase 3 completion"
```

---

## Task 9: Final Tag

**Step 1: Run all tests**

Run:
```bash
cd /Users/markalston/workspace/tile-diff
make test
```

Expected: All tests pass

**Step 2: Create git tag**

Run:
```bash
git tag -a v0.3.0-phase3 -m "Phase 3: Actionable reports complete

Features:
- Current config cross-reference
- Change categorization (Required/Warning/Info)
- Text and JSON report formats
- Specific recommendations

Full production-ready upgrade analysis tool"
```

**Step 3: Verify**

Run:
```bash
git tag -l -n5
```

Expected: Shows all three version tags

---

## Success Criteria

Phase 3 is successful if:

1. **Current Config Parsing**: Can parse Ops Manager API properties ✓
2. **Relevance Filtering**: Shows only changes affecting deployed config ✓
3. **Categorization**: Classifies into Required/Warning/Info ✓
4. **Text Reports**: Generates human-readable formatted reports ✓
5. **JSON Reports**: Generates machine-readable structured reports ✓
6. **CLI Integration**: Format flag works, displays reports ✓
7. **Tests Pass**: All unit and integration tests pass ✓
8. **Real Data Works**: Successfully analyzes actual TAS tiles ✓

**Deliverable Met:** CLI generates actionable upgrade report showing what operator must do, should review, and can ignore.
