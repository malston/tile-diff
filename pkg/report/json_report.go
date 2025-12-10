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
