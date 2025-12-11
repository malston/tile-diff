// ABOUTME: HTML parser for extracting features from release notes.
// ABOUTME: Identifies feature sections and descriptions for matching.
package releasenotes

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// Feature represents a feature from release notes
type Feature struct {
	Title       string
	Description string
	Position    int
}

// ParseHTML extracts features from release notes HTML
func ParseHTML(htmlContent string) ([]Feature, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var features []Feature
	position := 0

	var extractFeatures func(*html.Node)
	var currentFeature *Feature

	extractFeatures = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "h2":
				// Save previous feature
				if currentFeature != nil {
					features = append(features, *currentFeature)
				}
				// Start new feature
				position++
				currentFeature = &Feature{
					Title:    extractText(n),
					Position: position,
				}
			case "p", "ul", "li":
				// Add to current feature description
				if currentFeature != nil {
					text := extractText(n)
					if text != "" {
						if currentFeature.Description != "" {
							currentFeature.Description += " "
						}
						currentFeature.Description += text
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractFeatures(c)
		}
	}

	extractFeatures(doc)

	// Don't forget last feature
	if currentFeature != nil {
		features = append(features, *currentFeature)
	}

	return features, nil
}

func extractText(n *html.Node) string {
	if n.Type == html.TextNode {
		return strings.TrimSpace(n.Data)
	}

	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text += extractText(c) + " "
	}
	return strings.TrimSpace(text)
}
