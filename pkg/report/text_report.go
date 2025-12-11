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
				if change.NewProperty.Default != nil {
					report.WriteString(fmt.Sprintf("   Default: %v\n", change.NewProperty.Default))
				}
			}
			report.WriteString(fmt.Sprintf("   Note: %s\n", change.Recommendation))
			report.WriteString("\n")
		}
	}

	return report.String()
}
