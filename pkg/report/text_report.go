// ABOUTME: Generates human-readable text reports for tile upgrades.
// ABOUTME: Formats categorized changes with sections and recommendations.
package report

import (
	"fmt"
	"strings"
)

const separator = "================================================================================\n"

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

// GenerateTextReportWithFeatures generates a text report with feature grouping
func GenerateTextReportWithFeatures(enriched *EnrichedChanges, oldVersion, newVersion string) string {
	var sb strings.Builder

	writeHeader(&sb, oldVersion, newVersion)
	writeSummary(&sb, enriched.CategorizedChanges)

	// Write required actions with feature grouping
	if len(enriched.RequiredActions) > 0 {
		sb.WriteString("\n")
		sb.WriteString(separator)
		sb.WriteString("🚨 REQUIRED ACTIONS\n")
		sb.WriteString(separator)
		sb.WriteString("\n")

		// Group by feature
		featureProps := buildFeaturePropertyMap(enriched)
		ungrouped := findUngroupedProperties(enriched.RequiredActions, featureProps)

		// Write feature groups first
		for _, feature := range enriched.Features {
			writeFeatureGroup(&sb, feature, enriched.RequiredActions)
		}

		// Write ungrouped properties
		if len(ungrouped) > 0 {
			sb.WriteString("\n-- Ungrouped Properties --\n\n")
			for _, change := range ungrouped {
				writePropertyDetail(&sb, change, 0)
			}
		}
	}

	// Write warnings and informational (existing logic)
	writeWarnings(&sb, enriched.Warnings)
	writeInformational(&sb, enriched.Informational)

	return sb.String()
}

func writeHeader(sb *strings.Builder, oldVersion, newVersion string) {
	sb.WriteString(separator)
	sb.WriteString("                        TAS Tile Upgrade Analysis\n")
	sb.WriteString(separator)
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Old Version: %s\n", oldVersion))
	sb.WriteString(fmt.Sprintf("New Version: %s\n\n", newVersion))
}

func writeSummary(sb *strings.Builder, changes *CategorizedChanges) {
	totalChanges := len(changes.RequiredActions) + len(changes.Warnings) + len(changes.Informational)
	sb.WriteString(fmt.Sprintf("Total Changes: %d\n", totalChanges))
	sb.WriteString(fmt.Sprintf("  Required Actions: %d\n", len(changes.RequiredActions)))
	sb.WriteString(fmt.Sprintf("  Warnings: %d\n", len(changes.Warnings)))
	sb.WriteString(fmt.Sprintf("  Informational: %d\n\n", len(changes.Informational)))
}

func writeWarnings(sb *strings.Builder, warnings []CategorizedChange) {
	if len(warnings) > 0 {
		sb.WriteString(separator)
		sb.WriteString("⚠️  WARNINGS\n")
		sb.WriteString(separator)
		sb.WriteString("\n")
		sb.WriteString("These changes should be reviewed:\n\n")

		for i, change := range warnings {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, change.PropertyName))
			sb.WriteString(fmt.Sprintf("   Change: %s\n", change.Description))
			sb.WriteString(fmt.Sprintf("   Recommendation: %s\n", change.Recommendation))
			sb.WriteString("\n")
		}
	}
}

func writeInformational(sb *strings.Builder, informational []CategorizedChange) {
	if len(informational) > 0 {
		sb.WriteString(separator)
		sb.WriteString("ℹ️  INFORMATIONAL\n")
		sb.WriteString(separator)
		sb.WriteString("\n")
		sb.WriteString("New optional features available:\n\n")

		for i, change := range informational {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, change.PropertyName))
			if change.NewProperty != nil {
				sb.WriteString(fmt.Sprintf("   Type: %s\n", change.NewProperty.Type))
				if change.NewProperty.Default != nil {
					sb.WriteString(fmt.Sprintf("   Default: %v\n", change.NewProperty.Default))
				}
			}
			sb.WriteString(fmt.Sprintf("   Note: %s\n", change.Recommendation))
			sb.WriteString("\n")
		}
	}
}

func buildFeaturePropertyMap(enriched *EnrichedChanges) map[string]string {
	featureProps := make(map[string]string)
	for _, feature := range enriched.Features {
		for _, prop := range feature.Properties {
			featureProps[prop] = feature.Name
		}
	}
	return featureProps
}

func findUngroupedProperties(changes []CategorizedChange, featureProps map[string]string) []CategorizedChange {
	var ungrouped []CategorizedChange
	for _, change := range changes {
		if _, grouped := featureProps[change.PropertyName]; !grouped {
			ungrouped = append(ungrouped, change)
		}
	}
	return ungrouped
}

func writeFeatureGroup(sb *strings.Builder, feature FeatureGroup, changes []CategorizedChange) {
	sb.WriteString(fmt.Sprintf("📦 %s (%d properties)\n", feature.Name, len(feature.Properties)))
	sb.WriteString(fmt.Sprintf("   %s\n\n", feature.Description))

	for _, prop := range feature.Properties {
		for _, change := range changes {
			if change.PropertyName == prop {
				writePropertyDetail(sb, change, 0)
			}
		}
	}
	sb.WriteString("\n")
}

func writePropertyDetail(sb *strings.Builder, change CategorizedChange, indent int) {
	indentStr := strings.Repeat(" ", indent)
	sb.WriteString(fmt.Sprintf("%s• %s\n", indentStr, change.PropertyName))
	if change.NewProperty != nil {
		sb.WriteString(fmt.Sprintf("%s  Type: %s\n", indentStr, change.NewProperty.Type))
	}
	sb.WriteString(fmt.Sprintf("%s  Action: %s\n", indentStr, change.Recommendation))
	sb.WriteString("\n")
}
