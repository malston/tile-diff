package report

import (
	"strings"
	"testing"
)

func TestCleanDescription_Deduplication(t *testing.T) {
	input := "This is a test. This is a test. Another sentence. This is a test."
	result := CleanDescription(input)

	// Should contain both sentences at least once (accounting for bullet formatting)
	if !strings.Contains(result, "This is a test") {
		t.Errorf("Expected 'This is a test' to appear at least once, got: %s", result)
	}

	if !strings.Contains(result, "Another sentence") {
		t.Errorf("Expected 'Another sentence' to be preserved, got: %s", result)
	}

	// Count occurrences - should only appear once each (but accounting for bullet formatting)
	// Remove bullet points and formatting for counting
	cleaned := strings.ReplaceAll(result, "•", "")
	cleaned = strings.ReplaceAll(cleaned, "  ", " ")

	// Count periods after "This is a test" to count sentences
	testCount := strings.Count(cleaned, "This is a test.")
	if testCount != 1 {
		t.Errorf("Expected 'This is a test' to appear as a sentence once, got %d times in: %s", testCount, result)
	}
}

func TestCleanDescription_BumpLineSummarization(t *testing.T) {
	input := `Some feature description.
	Bump java-offline-buildpack to version 4.86.1
	Bump nodejs-offline-buildpack to version 1.8.84
	Bump python-offline-buildpack to version 1.8.79
	Bump capi to version 1.218.0
	Bump diego to version 2.122.0`

	result := CleanDescription(input)

	// Should summarize bumps
	if !strings.Contains(result, "Component Updates") {
		t.Error("Expected component updates summary")
	}

	// Should mention buildpacks
	if !strings.Contains(result, "buildpack") {
		t.Error("Expected buildpack summary")
	}

	// Should not have individual bump lines
	if strings.Count(result, "Bump") > 1 {
		t.Error("Expected individual bump lines to be summarized")
	}

	// Should preserve feature description
	if !strings.Contains(result, "feature description") {
		t.Error("Expected feature description to be preserved")
	}
}

func TestCleanDescription_NoiseFiltering(t *testing.T) {
	input := `Real feature content here.
	Products Solutions Support and Services Company How To Buy
	Privacy Supplier Responsibility Terms of Use Site Map
	Content feedback and comments
	For more information, see the docs.
	See the Knowledge Base article.
	Affected versions: 10.2.0, 10.2.1
	Known Issue: This is a known issue.
	Release date: October 28, 2025`

	result := CleanDescription(input)

	// Should preserve real content
	if !strings.Contains(result, "Real feature content") {
		t.Error("Expected real content to be preserved")
	}

	// Should filter out noise
	noisePatterns := []string{
		"Products Solutions",
		"Privacy",
		"Content feedback",
		"For more information",
		"Affected versions",
		"Known Issue",
		"Release date",
	}

	for _, noise := range noisePatterns {
		if strings.Contains(result, noise) {
			t.Errorf("Expected noise pattern '%s' to be filtered out", noise)
		}
	}
}

func TestSplitIntoSentences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "simple sentences",
			input:    "First sentence. Second sentence. Third sentence.",
			expected: 3,
		},
		{
			name:     "with questions",
			input:    "Is this a test? Yes it is! Another sentence.",
			expected: 3,
		},
		{
			name:     "filters short fragments",
			input:    "Real sentence here. A. B. C. Another real sentence.",
			expected: 2, // Only the two real sentences
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitIntoSentences(tt.input)
			if len(result) != tt.expected {
				t.Errorf("Expected %d sentences, got %d: %v", tt.expected, len(result), result)
			}
		})
	}
}

func TestDeduplicateSentences(t *testing.T) {
	input := []string{
		"First sentence",
		"Second sentence",
		"First sentence", // Duplicate
		"FIRST SENTENCE", // Case-insensitive duplicate
		"Third sentence",
	}

	result := deduplicateSentences(input)

	if len(result) != 3 {
		t.Errorf("Expected 3 unique sentences, got %d", len(result))
	}

	// Check order is preserved (first occurrence kept)
	if result[0] != "First sentence" {
		t.Error("Expected first occurrence to be preserved")
	}
}

func TestSeparateBumpLines(t *testing.T) {
	input := []string{
		"Regular feature description",
		"Bump java-offline-buildpack to version 4.86.1",
		"Another description line",
		"Bump capi to version 1.218.0",
		"Bump diego to version 2.122.0",
	}

	bumps, others := separateBumpLines(input)

	if len(bumps) != 3 {
		t.Errorf("Expected 3 bump lines, got %d", len(bumps))
	}

	if len(others) != 2 {
		t.Errorf("Expected 2 other lines, got %d", len(others))
	}
}

func TestSummarizeBumpLines(t *testing.T) {
	tests := []struct {
		name     string
		bumps    []string
		expected []string // strings that should appear in output
	}{
		{
			name: "buildpacks and components",
			bumps: []string{
				"Bump java-offline-buildpack to version 4.86.1",
				"Bump nodejs-offline-buildpack to version 1.8.84",
				"Bump capi to version 1.218.0",
				"Bump diego to version 2.122.0",
			},
			expected: []string{"buildpack", "component", "Component Updates"},
		},
		{
			name: "only buildpacks",
			bumps: []string{
				"Bump java-offline-buildpack to version 4.86.1",
				"Bump nodejs-offline-buildpack to version 1.8.84",
			},
			expected: []string{"buildpack", "Component Updates"},
		},
		{
			name:     "empty",
			bumps:    []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := summarizeBumpLines(tt.bumps)

			for _, exp := range tt.expected {
				if !strings.Contains(result, exp) {
					t.Errorf("Expected summary to contain '%s', got: %s", exp, result)
				}
			}

			if len(tt.bumps) == 0 && result != "" {
				t.Error("Expected empty result for empty input")
			}
		})
	}
}

func TestFilterNoiseLines(t *testing.T) {
	input := []string{
		"Real content with meaningful information here",
		"Products Solutions Support",
		"Privacy Terms of Use",
		"For more information see the docs",
		"Another real piece of content",
		"Content feedback",
		"See this article",
		"Affected versions: 10.2.0",
		"Known issue: there is a problem",
		"Release date: October 28, 2025",
	}

	result := filterNoiseLines(input)

	// Check real content is preserved
	hasFirstContent := false
	hasSecondContent := false
	for _, line := range result {
		if strings.Contains(line, "Real content with meaningful") {
			hasFirstContent = true
		}
		if strings.Contains(line, "Another real piece") {
			hasSecondContent = true
		}
	}

	if !hasFirstContent {
		t.Error("Expected first real content to be preserved")
	}
	if !hasSecondContent {
		t.Error("Expected second real content to be preserved")
	}

	// Check noise is filtered
	for _, line := range result {
		if strings.Contains(line, "Products Solutions") {
			t.Error("Expected 'Products Solutions' to be filtered")
		}
		if strings.Contains(line, "For more information") {
			t.Error("Expected 'For more information' to be filtered")
		}
		if strings.Contains(line, "Affected versions") {
			t.Error("Expected 'Affected versions' to be filtered")
		}
	}
}

func TestFormatCount(t *testing.T) {
	tests := []struct {
		count    int
		noun     string
		expected string
	}{
		{1, "buildpack", "1 buildpack"},
		{2, "buildpack", "2 buildpacks"},
		{5, "component", "5 components"},
		{1, "item", "1 item"},
	}

	for _, tt := range tests {
		result := formatCount(tt.count, tt.noun)
		if result != tt.expected {
			t.Errorf("formatCount(%d, %s) = %s, want %s", tt.count, tt.noun, result, tt.expected)
		}
	}
}

func TestCleanDescription_Integration(t *testing.T) {
	// Test with a realistic messy input similar to what we'd get from release notes
	input := `Canary deployment enhancements: Support a new --instance-steps flag that allows for fine-grained control over a canary deployment rollout.
	Gorouter supports a new X-CF-PROCESS-INSTANCE header for routing http requests to a specific app process.
	Canary deployment enhancements: Support a new --instance-steps flag that allows for fine-grained control over a canary deployment rollout.
	Gorouter supports a new X-CF-PROCESS-INSTANCE header for routing http requests to a specific app process.
	Bump java-offline-buildpack to version 4.86.1
	Bump nodejs-offline-buildpack to version 1.8.84
	Bump python-offline-buildpack to version 1.8.79
	Bump capi to version 1.218.0
	Bump diego to version 2.122.0
	Products Solutions Support and Services Company
	Privacy Terms of Use Site Map
	For more information, see the documentation.`

	result := CleanDescription(input)

	// Should deduplicate canary deployment text
	if strings.Count(result, "Canary deployment") > 1 {
		t.Error("Expected canary deployment to appear only once")
	}

	// Should deduplicate Gorouter text
	if strings.Count(result, "X-CF-PROCESS-INSTANCE") > 1 {
		t.Error("Expected Gorouter text to appear only once")
	}

	// Should summarize bumps
	if strings.Contains(result, "Bump java-offline-buildpack") {
		t.Error("Expected individual bump lines to be summarized")
	}
	if !strings.Contains(result, "Component Updates") {
		t.Error("Expected component updates summary")
	}

	// Should filter noise
	if strings.Contains(result, "Products Solutions") {
		t.Error("Expected footer noise to be filtered")
	}
	if strings.Contains(result, "For more information") {
		t.Error("Expected documentation reference to be filtered")
	}

	// Should format with bullets
	if !strings.Contains(result, "•") {
		t.Error("Expected bullet points in output")
	}
}
