// ABOUTME: Cleans and formats feature descriptions from release notes.
// ABOUTME: Handles deduplication, bump line summarization, and noise filtering.
package report

import (
	"fmt"
	"regexp"
	"strings"
)

// CleanDescription processes a raw feature description to make it readable
func CleanDescription(raw string) string {
	if raw == "" {
		return ""
	}

	// Split into sentences
	sentences := splitIntoSentences(raw)

	// Deduplicate sentences
	sentences = deduplicateSentences(sentences)

	// Separate bump lines from other content
	bumpLines, otherLines := separateBumpLines(sentences)

	// Filter out footer/navigation noise
	otherLines = filterNoiseLines(otherLines)

	// Build cleaned output
	var result strings.Builder

	// Add main content with structure
	if len(otherLines) > 0 {
		// Group related content
		grouped := groupRelatedContent(otherLines)
		for _, group := range grouped {
			result.WriteString(group)
			result.WriteString("\n\n")
		}
	}

	// Summarize bump lines if present
	if len(bumpLines) > 0 {
		summary := summarizeBumpLines(bumpLines)
		if result.Len() > 0 {
			result.WriteString("\n")
		}
		result.WriteString(summary)
	}

	return strings.TrimSpace(result.String())
}

// splitIntoSentences breaks text into sentences for deduplication
func splitIntoSentences(text string) []string {
	// First, split by newlines to handle line-based content
	// This is important for content like "Bump X to version Y" which doesn't end with punctuation
	lines := strings.Split(text, "\n")

	var result []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) < 10 {
			continue
		}

		// If the line contains sentence-ending punctuation, split it further
		if strings.ContainsAny(line, ".!?") {
			// Split by sentence boundaries within the line
			sentences := splitBySentenceBoundaries(line)
			result = append(result, sentences...)
		} else {
			// Treat the whole line as a sentence (e.g., "Bump X to version Y")
			result = append(result, line)
		}
	}

	return result
}

// splitBySentenceBoundaries splits a single line by sentence-ending punctuation
func splitBySentenceBoundaries(line string) []string {
	var result []string
	current := 0

	for i := 0; i < len(line); i++ {
		// Check if this is a sentence boundary
		if line[i] == '.' || line[i] == '!' || line[i] == '?' {
			// Look ahead to see if there's whitespace (end of sentence)
			if i+1 >= len(line) || line[i+1] == ' ' || line[i+1] == '\t' {
				sentence := strings.TrimSpace(line[current : i+1])
				if len(sentence) >= 10 { // Filter very short fragments
					result = append(result, sentence)
				}
				// Skip the whitespace after the punctuation
				if i+1 < len(line) && (line[i+1] == ' ' || line[i+1] == '\t') {
					i++
				}
				current = i + 1
			}
		}
	}

	// Don't forget the last sentence if it doesn't end with punctuation
	if current < len(line) {
		sentence := strings.TrimSpace(line[current:])
		if len(sentence) >= 10 {
			result = append(result, sentence)
		}
	}

	return result
}

// deduplicateSentences removes duplicate sentences
func deduplicateSentences(sentences []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, s := range sentences {
		normalized := strings.ToLower(strings.TrimSpace(s))
		if !seen[normalized] {
			seen[normalized] = true
			result = append(result, s)
		}
	}

	return result
}

// separateBumpLines splits bump version lines from other content
func separateBumpLines(sentences []string) (bumps []string, others []string) {
	bumpPattern := regexp.MustCompile(`(?i)bump\s+[\w-]+\s+to\s+version`)

	for _, s := range sentences {
		if bumpPattern.MatchString(s) {
			bumps = append(bumps, s)
		} else {
			others = append(others, s)
		}
	}

	return bumps, others
}

// filterNoiseLines removes footer, navigation, and other noise
func filterNoiseLines(lines []string) []string {
	noisePatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)^(products|solutions|support|company|how to buy)`),
		regexp.MustCompile(`(?i)^(privacy|terms of use|site map|supplier)`),
		regexp.MustCompile(`(?i)^content feedback`),
		regexp.MustCompile(`(?i)^for (more information|component details|the list)`),
		regexp.MustCompile(`(?i)^see (the|this)`),
		regexp.MustCompile(`(?i)^affected versions:`),
		regexp.MustCompile(`(?i)^known issue:`),
		regexp.MustCompile(`(?i)^release date:`),
		regexp.MustCompile(`(?i)^this release includes`),
	}

	var result []string
	for _, line := range lines {
		normalized := strings.TrimSpace(line)

		isNoise := false
		for _, pattern := range noisePatterns {
			if pattern.MatchString(normalized) {
				isNoise = true
				break
			}
		}

		if !isNoise {
			result = append(result, line)
		}
	}

	return result
}

// groupRelatedContent organizes sentences into logical groups
func groupRelatedContent(lines []string) []string {
	if len(lines) == 0 {
		return nil
	}

	// For now, just return as bullet points
	// Future: could add smarter grouping by topic
	var result []string
	for _, line := range lines {
		// Add bullet point if line doesn't start with one
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "•") && !strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "*") {
			line = "• " + line
		}
		result = append(result, line)
	}

	return result
}

// summarizeBumpLines condenses version bump information
func summarizeBumpLines(bumps []string) string {
	if len(bumps) == 0 {
		return ""
	}

	// Count by category
	buildpackCount := 0
	otherCount := 0

	buildpackPattern := regexp.MustCompile(`(?i)buildpack`)

	for _, bump := range bumps {
		if buildpackPattern.MatchString(bump) {
			buildpackCount++
		} else {
			otherCount++
		}
	}

	var summary strings.Builder
	summary.WriteString("Component Updates:\n")

	if buildpackCount > 0 {
		summary.WriteString("  • ")
		summary.WriteString(formatCount(buildpackCount, "buildpack"))
		summary.WriteString(" updated\n")
	}

	if otherCount > 0 {
		summary.WriteString("  • ")
		summary.WriteString(formatCount(otherCount, "component"))
		summary.WriteString(" updated\n")
	}

	return summary.String()
}

// formatCount formats a count with proper singular/plural
func formatCount(count int, noun string) string {
	if count == 1 {
		return "1 " + noun
	}
	return strings.TrimSpace(strings.Join([]string{fmt.Sprintf("%d", count), noun + "s"}, " "))
}
