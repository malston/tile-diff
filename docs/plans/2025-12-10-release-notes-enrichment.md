# Release Notes Enrichment Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Enrich tile-diff reports with feature context from product release notes by fetching, parsing, and matching property changes to documented features.

**Architecture:** Add `pkg/releasenotes` package with registry, fetcher, parser, and matcher components. Integrate with existing `pkg/report` to enrich output with feature grouping and context. Graceful degradation ensures core functionality never breaks.

**Tech Stack:** Go 1.21+, golang.org/x/net/html for parsing, YAML for config, existing tile-diff components

---

## Phase 1: Product Registry & Configuration

### Task 1: Product Config Structure

**Files:**
- Create: `configs/products.yaml`
- Create: `pkg/releasenotes/registry.go`
- Create: `pkg/releasenotes/registry_test.go`

**Step 1: Write the failing test**

```go
// pkg/releasenotes/registry_test.go
package releasenotes

import (
	"testing"
)

func TestLoadProductConfig(t *testing.T) {
	config, err := LoadProductConfig("testdata/products.yaml")
	if err != nil {
		t.Fatalf("LoadProductConfig failed: %v", err)
	}

	if config["cf"] != "https://techdocs.broadcom.com/cf/{version}/release-notes.html" {
		t.Errorf("Expected cf URL, got %s", config["cf"])
	}
}

func TestResolveURL(t *testing.T) {
	config := ProductConfig{
		"cf": "https://techdocs.broadcom.com/cf/{version}/release-notes.html",
	}

	url, err := config.ResolveURL("cf", "10.2.5")
	if err != nil {
		t.Fatalf("ResolveURL failed: %v", err)
	}

	expected := "https://techdocs.broadcom.com/cf/10.2.5/release-notes.html"
	if url != expected {
		t.Errorf("Expected %s, got %s", expected, url)
	}
}

func TestResolveURL_ProductNotFound(t *testing.T) {
	config := ProductConfig{}

	_, err := config.ResolveURL("unknown", "1.0.0")
	if err == nil {
		t.Error("Expected error for unknown product")
	}
}
```

**Step 2: Create test config**

Create `pkg/releasenotes/testdata/products.yaml`:

```yaml
cf: "https://techdocs.broadcom.com/cf/{version}/release-notes.html"
p-mysql: "https://techdocs.broadcom.com/mysql/{version}/release-notes.html"
```

**Step 3: Run test to verify it fails**

```bash
cd .worktrees/release-notes-enrichment
go test ./pkg/releasenotes -v
```

Expected: FAIL - package does not exist

**Step 4: Write minimal implementation**

```go
// pkg/releasenotes/registry.go
// ABOUTME: Product registry for mapping product IDs to release note URLs.
// ABOUTME: Loads configuration and resolves versioned URLs for fetching.
package releasenotes

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// ProductConfig maps product IDs to release notes URL patterns
type ProductConfig map[string]string

// LoadProductConfig loads product configuration from YAML file
func LoadProductConfig(path string) (ProductConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config ProductConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return config, nil
}

// ResolveURL resolves a product ID and version to a release notes URL
func (c ProductConfig) ResolveURL(productID, version string) (string, error) {
	pattern, ok := c[productID]
	if !ok {
		return "", fmt.Errorf("product %s not found in config", productID)
	}

	url := strings.ReplaceAll(pattern, "{version}", version)
	return url, nil
}
```

**Step 5: Run test to verify it passes**

```bash
go test ./pkg/releasenotes -v
```

Expected: PASS

**Step 6: Commit**

```bash
git add pkg/releasenotes/registry.go pkg/releasenotes/registry_test.go pkg/releasenotes/testdata/products.yaml
git commit -m "feat(releasenotes): add product registry with URL resolution"
```

---

### Task 2: Product Identification from Tile

**Files:**
- Modify: `pkg/releasenotes/registry.go`
- Modify: `pkg/releasenotes/registry_test.go`

**Step 1: Write the failing test**

Add to `pkg/releasenotes/registry_test.go`:

```go
func TestIdentifyProduct(t *testing.T) {
	metadata := map[string]interface{}{
		"name": "cf",
	}

	productID := IdentifyProduct(metadata)
	if productID != "cf" {
		t.Errorf("Expected cf, got %s", productID)
	}
}

func TestIdentifyProduct_Normalize(t *testing.T) {
	metadata := map[string]interface{}{
		"name": "Tanzu Application Service",
	}

	productID := IdentifyProduct(metadata)
	if productID != "cf" {
		t.Errorf("Expected cf, got %s", productID)
	}
}

func TestIdentifyProduct_NotFound(t *testing.T) {
	metadata := map[string]interface{}{}

	productID := IdentifyProduct(metadata)
	if productID != "" {
		t.Errorf("Expected empty string, got %s", productID)
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./pkg/releasenotes -v -run TestIdentifyProduct
```

Expected: FAIL - function not defined

**Step 3: Write minimal implementation**

Add to `pkg/releasenotes/registry.go`:

```go
import (
	"strings"
)

var productNameMapping = map[string]string{
	"tanzu application service": "cf",
	"tas":                        "cf",
	"cf":                         "cf",
	"mysql":                      "p-mysql",
	"p-mysql":                    "p-mysql",
	"rabbitmq":                   "p-rabbitmq",
	"p-rabbitmq":                 "p-rabbitmq",
}

// IdentifyProduct extracts and normalizes product ID from tile metadata
func IdentifyProduct(metadata map[string]interface{}) string {
	// Try "name" field first
	if name, ok := metadata["name"].(string); ok {
		normalized := strings.ToLower(strings.TrimSpace(name))
		if productID, found := productNameMapping[normalized]; found {
			return productID
		}
		return normalized
	}

	// Try "product_name" field
	if name, ok := metadata["product_name"].(string); ok {
		normalized := strings.ToLower(strings.TrimSpace(name))
		if productID, found := productNameMapping[normalized]; found {
			return productID
		}
		return normalized
	}

	return ""
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./pkg/releasenotes -v -run TestIdentifyProduct
```

Expected: PASS

**Step 5: Commit**

```bash
git add pkg/releasenotes/registry.go pkg/releasenotes/registry_test.go
git commit -m "feat(releasenotes): add product identification from metadata"
```

---

### Task 3: Create Production Config File

**Files:**
- Create: `configs/products.yaml`

**Step 1: Create production config**

```yaml
# Tanzu Application Service (Cloud Foundry)
cf: "https://techdocs.broadcom.com/us/en/vmware-tanzu/platform/tanzu-application-service/{version}/release-notes.html"

# Tanzu Data Solutions
p-mysql: "https://techdocs.broadcom.com/us/en/vmware-tanzu/data-solutions/tanzu-sql-mysql/{version}/release-notes.html"
p-rabbitmq: "https://techdocs.broadcom.com/us/en/vmware-tanzu/data-solutions/tanzu-rabbitmq/{version}/release-notes.html"

# Add more products as needed
```

**Step 2: Commit**

```bash
git add configs/products.yaml
git commit -m "feat(releasenotes): add production product config"
```

---

## Phase 2: Release Notes Fetching

### Task 4: HTTP Fetcher

**Files:**
- Create: `pkg/releasenotes/fetcher.go`
- Create: `pkg/releasenotes/fetcher_test.go`

**Step 1: Write the failing test**

```go
// pkg/releasenotes/fetcher_test.go
package releasenotes

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchReleaseNotes(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html><body>Release Notes</body></html>"))
	}))
	defer server.Close()

	fetcher := NewFetcher()
	html, err := fetcher.Fetch(server.URL)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if !contains(html, "Release Notes") {
		t.Error("Expected HTML to contain 'Release Notes'")
	}
}

func TestFetchReleaseNotes_404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	fetcher := NewFetcher()
	_, err := fetcher.Fetch(server.URL)
	if err == nil {
		t.Error("Expected error for 404 response")
	}
}

func TestFetchReleaseNotes_Cache(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Write([]byte("<html><body>Release Notes</body></html>"))
	}))
	defer server.Close()

	fetcher := NewFetcher()

	// First fetch
	_, err := fetcher.Fetch(server.URL)
	if err != nil {
		t.Fatalf("First fetch failed: %v", err)
	}

	// Second fetch (should use cache)
	_, err = fetcher.Fetch(server.URL)
	if err != nil {
		t.Fatalf("Second fetch failed: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 HTTP call (cached), got %d", callCount)
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && s != "" && substr != ""
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./pkg/releasenotes -v -run TestFetch
```

Expected: FAIL - function not defined

**Step 3: Write minimal implementation**

```go
// pkg/releasenotes/fetcher.go
// ABOUTME: HTTP client for fetching release notes from documentation sites.
// ABOUTME: Implements in-memory caching and timeout handling.
package releasenotes

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// Fetcher fetches and caches release notes HTML
type Fetcher struct {
	client *http.Client
	cache  map[string]string
}

// NewFetcher creates a new release notes fetcher
func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache: make(map[string]string),
	}
}

// Fetch retrieves release notes HTML from URL (with caching)
func (f *Fetcher) Fetch(url string) (string, error) {
	// Check cache
	if cached, ok := f.cache[url]; ok {
		return cached, nil
	}

	// Fetch from URL
	resp, err := f.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d fetching %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	html := string(body)
	f.cache[url] = html

	return html, nil
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./pkg/releasenotes -v -run TestFetch
```

Expected: PASS

**Step 5: Commit**

```bash
git add pkg/releasenotes/fetcher.go pkg/releasenotes/fetcher_test.go
git commit -m "feat(releasenotes): add HTTP fetcher with caching"
```

---

## Phase 3: HTML Parsing

### Task 5: HTML Parser for Features

**Files:**
- Create: `pkg/releasenotes/parser.go`
- Create: `pkg/releasenotes/parser_test.go`
- Create: `pkg/releasenotes/testdata/sample-release-notes.html`

**Step 1: Create test HTML fixture**

Create `pkg/releasenotes/testdata/sample-release-notes.html`:

```html
<html>
<head><title>TAS 10.2.5 Release Notes</title></head>
<body>
<h1>Release Notes for TAS 10.2.5</h1>

<h2>Enhanced Security Scanning</h2>
<p>This release introduces runtime container vulnerability detection. Configure using the .properties.security_scanner_enabled property.</p>
<ul>
  <li>Enable scanner with security_scanner_enabled</li>
  <li>Set scan interval with scanner_update_interval</li>
</ul>

<h2>Improved Logging</h2>
<p>Application log rate limiting is now available to prevent log flooding.</p>
<p>Properties: app_log_rate_limiting</p>

<h2>Bug Fixes</h2>
<p>Various bug fixes and improvements.</p>
</body>
</html>
```

**Step 2: Write the failing test**

```go
// pkg/releasenotes/parser_test.go
package releasenotes

import (
	"os"
	"testing"
)

func TestParseHTML(t *testing.T) {
	html, err := os.ReadFile("testdata/sample-release-notes.html")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	features, err := ParseHTML(string(html))
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	if len(features) < 2 {
		t.Errorf("Expected at least 2 features, got %d", len(features))
	}

	// Check first feature
	if features[0].Title != "Enhanced Security Scanning" {
		t.Errorf("Expected 'Enhanced Security Scanning', got %s", features[0].Title)
	}

	if !containsString(features[0].Description, "vulnerability detection") {
		t.Error("Expected description to mention vulnerability detection")
	}
}

func TestExtractFeatures_PropertyMentions(t *testing.T) {
	html, _ := os.ReadFile("testdata/sample-release-notes.html")
	features, _ := ParseHTML(string(html))

	// First feature should mention security_scanner_enabled
	if !containsString(features[0].Description, "security_scanner_enabled") {
		t.Error("Expected feature to mention security_scanner_enabled")
	}
}

func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0
}
```

**Step 3: Run test to verify it fails**

```bash
go test ./pkg/releasenotes -v -run TestParseHTML
```

Expected: FAIL - function not defined

**Step 4: Write minimal implementation**

```go
// pkg/releasenotes/parser.go
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
```

**Step 5: Add dependency**

```bash
go get golang.org/x/net/html
go mod tidy
```

**Step 6: Run test to verify it passes**

```bash
go test ./pkg/releasenotes -v -run TestParseHTML
```

Expected: PASS

**Step 7: Commit**

```bash
git add pkg/releasenotes/parser.go pkg/releasenotes/parser_test.go pkg/releasenotes/testdata/ go.mod go.sum
git commit -m "feat(releasenotes): add HTML parser for feature extraction"
```

---

## Phase 4: Property Matching

### Task 6: Direct Matching

**Files:**
- Create: `pkg/releasenotes/matcher.go`
- Create: `pkg/releasenotes/matcher_test.go`

**Step 1: Write the failing test**

```go
// pkg/releasenotes/matcher_test.go
package releasenotes

import (
	"testing"
)

func TestDirectMatch(t *testing.T) {
	features := []Feature{
		{
			Title:       "Enhanced Security",
			Description: "Use .properties.security_scanner_enabled to enable scanning",
			Position:    1,
		},
	}

	properties := []string{
		".properties.security_scanner_enabled",
		".properties.unrelated_property",
	}

	matcher := NewMatcher(features)
	matches := matcher.Match(properties)

	// Should have one match for security_scanner_enabled
	if len(matches) != 1 {
		t.Errorf("Expected 1 match, got %d", len(matches))
	}

	match := matches[".properties.security_scanner_enabled"]
	if match.MatchType != "direct" {
		t.Errorf("Expected direct match, got %s", match.MatchType)
	}

	if match.Confidence != 1.0 {
		t.Errorf("Expected confidence 1.0, got %f", match.Confidence)
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./pkg/releasenotes -v -run TestDirectMatch
```

Expected: FAIL - types not defined

**Step 3: Write minimal implementation**

```go
// pkg/releasenotes/matcher.go
// ABOUTME: Property-to-feature matching using multiple strategies.
// ABOUTME: Implements direct, keyword, and proximity matching with confidence scoring.
package releasenotes

import (
	"strings"
)

// Match represents a property matched to a feature
type Match struct {
	Property   string
	Feature    Feature
	MatchType  string  // "direct", "keyword", "proximity"
	Confidence float64 // 0.0-1.0
}

// Matcher matches properties to features
type Matcher struct {
	features []Feature
}

// NewMatcher creates a new matcher with features
func NewMatcher(features []Feature) *Matcher {
	return &Matcher{features: features}
}

// Match finds matches for properties
func (m *Matcher) Match(properties []string) map[string]Match {
	matches := make(map[string]Match)

	for _, prop := range properties {
		if match, found := m.findBestMatch(prop); found {
			matches[prop] = match
		}
	}

	return matches
}

func (m *Matcher) findBestMatch(property string) (Match, bool) {
	var bestMatch Match
	bestConfidence := 0.0

	// Try direct matching first
	for _, feature := range m.features {
		if strings.Contains(feature.Description, property) {
			match := Match{
				Property:   property,
				Feature:    feature,
				MatchType:  "direct",
				Confidence: 1.0,
			}
			return match, true
		}
	}

	// If no direct match found and no other matches, return false
	if bestConfidence == 0.0 {
		return bestMatch, false
	}

	return bestMatch, true
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./pkg/releasenotes -v -run TestDirectMatch
```

Expected: PASS

**Step 5: Commit**

```bash
git add pkg/releasenotes/matcher.go pkg/releasenotes/matcher_test.go
git commit -m "feat(releasenotes): add direct property matching"
```

---

### Task 7: Keyword Matching

**Files:**
- Modify: `pkg/releasenotes/matcher.go`
- Modify: `pkg/releasenotes/matcher_test.go`

**Step 1: Write the failing test**

Add to `pkg/releasenotes/matcher_test.go`:

```go
func TestKeywordMatch(t *testing.T) {
	features := []Feature{
		{
			Title:       "Application Log Rate Limiting",
			Description: "Prevent log flooding with rate limits on application logs",
			Position:    1,
		},
	}

	properties := []string{
		".properties.app_log_rate_limiting",
	}

	matcher := NewMatcher(features)
	matches := matcher.Match(properties)

	match := matches[".properties.app_log_rate_limiting"]
	if match.MatchType != "keyword" {
		t.Errorf("Expected keyword match, got %s", match.MatchType)
	}

	if match.Confidence < 0.5 {
		t.Errorf("Expected confidence >= 0.5, got %f", match.Confidence)
	}
}

func TestTokenizeProperty(t *testing.T) {
	tests := []struct {
		property string
		expected []string
	}{
		{
			property: ".properties.app_log_rate_limiting",
			expected: []string{"app", "log", "rate", "limiting"},
		},
		{
			property: ".properties.security_scanner_enabled",
			expected: []string{"security", "scanner"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.property, func(t *testing.T) {
			tokens := tokenizeProperty(tt.property)
			if len(tokens) != len(tt.expected) {
				t.Errorf("Expected %d tokens, got %d", len(tt.expected), len(tokens))
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./pkg/releasenotes -v -run TestKeywordMatch
```

Expected: FAIL - logic not implemented

**Step 3: Write implementation**

Add to `pkg/releasenotes/matcher.go`:

```go
var stopwords = map[string]bool{
	"enable":    true,
	"enabled":   true,
	"setting":   true,
	"settings":  true,
	"config":    true,
	"configure": true,
	"new":       true,
	"property":  true,
	"properties": true,
}

func (m *Matcher) findBestMatch(property string) (Match, bool) {
	var bestMatch Match
	bestConfidence := 0.0

	// Try direct matching first
	for _, feature := range m.features {
		if strings.Contains(feature.Description, property) {
			match := Match{
				Property:   property,
				Feature:    feature,
				MatchType:  "direct",
				Confidence: 1.0,
			}
			return match, true
		}
	}

	// Try keyword matching
	tokens := tokenizeProperty(property)
	if len(tokens) >= 2 {
		for _, feature := range m.features {
			score := keywordScore(tokens, feature)
			if score > bestConfidence {
				bestConfidence = score
				bestMatch = Match{
					Property:   property,
					Feature:    feature,
					MatchType:  "keyword",
					Confidence: score,
				}
			}
		}
	}

	// Only return if confidence > 0.5
	if bestConfidence > 0.5 {
		return bestMatch, true
	}

	return bestMatch, false
}

func tokenizeProperty(property string) []string {
	// Remove common prefixes
	prop := strings.TrimPrefix(property, ".properties.")
	prop = strings.TrimPrefix(prop, ".cloud_controller.")

	// Split on underscores and dots
	parts := strings.FieldsFunc(prop, func(r rune) bool {
		return r == '_' || r == '.' || r == '-'
	})

	// Filter stopwords
	var tokens []string
	for _, part := range parts {
		lower := strings.ToLower(part)
		if !stopwords[lower] && len(lower) > 2 {
			tokens = append(tokens, lower)
		}
	}

	return tokens
}

func keywordScore(tokens []string, feature Feature) float64 {
	if len(tokens) == 0 {
		return 0.0
	}

	searchText := strings.ToLower(feature.Title + " " + feature.Description)
	matched := 0

	for _, token := range tokens {
		if strings.Contains(searchText, token) {
			matched++
		}
	}

	// Score: (matched / total) * 0.9
	return (float64(matched) / float64(len(tokens))) * 0.9
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./pkg/releasenotes -v -run "TestKeywordMatch|TestTokenize"
```

Expected: PASS

**Step 5: Commit**

```bash
git add pkg/releasenotes/matcher.go pkg/releasenotes/matcher_test.go
git commit -m "feat(releasenotes): add keyword matching with tokenization"
```

---

## Phase 5: Report Integration

### Task 8: Report Enricher

**Files:**
- Create: `pkg/report/enricher.go`
- Create: `pkg/report/enricher_test.go`

**Step 1: Write the failing test**

```go
// pkg/report/enricher_test.go
package report

import (
	"testing"

	"github.com/malston/tile-diff/pkg/compare"
	"github.com/malston/tile-diff/pkg/metadata"
	"github.com/malston/tile-diff/pkg/releasenotes"
)

func TestEnrichChanges(t *testing.T) {
	features := []releasenotes.Feature{
		{
			Title:       "Enhanced Security",
			Description: "Security scanning feature",
			Position:    1,
		},
	}

	matches := map[string]releasenotes.Match{
		".properties.security_enabled": {
			Property: ".properties.security_enabled",
			Feature:  features[0],
			MatchType: "direct",
			Confidence: 1.0,
		},
	}

	changes := &CategorizedChanges{
		RequiredActions: []CategorizedChange{
			{
				ComparisonResult: compare.ComparisonResult{
					PropertyName: ".properties.security_enabled",
					NewProperty: &metadata.PropertyBlueprint{
						Name: "security_enabled",
						Type: "boolean",
					},
				},
				Category: CategoryRequired,
				Recommendation: "Must configure this property",
			},
		},
	}

	enriched := EnrichChanges(changes, matches)

	if len(enriched.Features) != 1 {
		t.Errorf("Expected 1 feature group, got %d", len(enriched.Features))
	}

	feature := enriched.Features[0]
	if feature.Name != "Enhanced Security" {
		t.Errorf("Expected 'Enhanced Security', got %s", feature.Name)
	}

	if len(feature.Properties) != 1 {
		t.Errorf("Expected 1 property in feature, got %d", len(feature.Properties))
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./pkg/report -v -run TestEnrichChanges
```

Expected: FAIL - types not defined

**Step 3: Write minimal implementation**

```go
// pkg/report/enricher.go
// ABOUTME: Enriches comparison results with feature context from release notes.
// ABOUTME: Groups properties by feature and enhances recommendations.
package report

import (
	"github.com/malston/tile-diff/pkg/releasenotes"
)

// FeatureGroup represents properties grouped by feature
type FeatureGroup struct {
	Name        string
	Description string
	Properties  []string
}

// EnrichedChanges extends CategorizedChanges with feature context
type EnrichedChanges struct {
	*CategorizedChanges
	Features []FeatureGroup
}

// EnrichChanges adds feature context to categorized changes
func EnrichChanges(changes *CategorizedChanges, matches map[string]releasenotes.Match) *EnrichedChanges {
	enriched := &EnrichedChanges{
		CategorizedChanges: changes,
	}

	// Group properties by feature
	featureMap := make(map[string]*FeatureGroup)

	// Process all changes
	for _, change := range changes.RequiredActions {
		if match, ok := matches[change.PropertyName]; ok {
			featureName := match.Feature.Title
			if _, exists := featureMap[featureName]; !exists {
				featureMap[featureName] = &FeatureGroup{
					Name:        match.Feature.Title,
					Description: match.Feature.Description,
					Properties:  []string{},
				}
			}
			featureMap[featureName].Properties = append(
				featureMap[featureName].Properties,
				change.PropertyName,
			)
		}
	}

	// Convert map to slice
	for _, group := range featureMap {
		enriched.Features = append(enriched.Features, *group)
	}

	return enriched
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./pkg/report -v -run TestEnrichChanges
```

Expected: PASS

**Step 5: Commit**

```bash
git add pkg/report/enricher.go pkg/report/enricher_test.go
git commit -m "feat(report): add feature-based enrichment"
```

---

### Task 9: Update Text Report with Feature Grouping

**Files:**
- Modify: `pkg/report/text_report.go`
- Modify: `pkg/report/text_report_test.go`

**Step 1: Write the failing test**

Add to `pkg/report/text_report_test.go`:

```go
func TestGenerateTextReport_WithFeatures(t *testing.T) {
	enriched := &EnrichedChanges{
		CategorizedChanges: &CategorizedChanges{
			RequiredActions: []CategorizedChange{
				{
					ComparisonResult: compare.ComparisonResult{
						PropertyName: ".properties.security_enabled",
					},
					Category: CategoryRequired,
					Recommendation: "Enable security scanning",
				},
			},
		},
		Features: []FeatureGroup{
			{
				Name:        "Enhanced Security",
				Description: "Security scanning feature",
				Properties:  []string{".properties.security_enabled"},
			},
		},
	}

	report := GenerateTextReportWithFeatures(enriched, "6.0.22", "10.2.5")

	if !strings.Contains(report, "📦 Enhanced Security") {
		t.Error("Expected report to contain feature grouping")
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./pkg/report -v -run TestGenerateTextReport_WithFeatures
```

Expected: FAIL - function not defined

**Step 3: Write implementation**

Add to `pkg/report/text_report.go`:

```go
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
```

**Step 4: Run test to verify it passes**

```bash
go test ./pkg/report -v -run TestGenerateTextReport_WithFeatures
```

Expected: PASS

**Step 5: Commit**

```bash
git add pkg/report/text_report.go pkg/report/text_report_test.go
git commit -m "feat(report): add feature grouping to text reports"
```

---

## Phase 6: CLI Integration

### Task 10: Add CLI Flags and Orchestration

**Files:**
- Modify: `cmd/tile-diff/main.go`

**Step 1: Add CLI flags**

```go
// Add to main.go flag definitions
var (
	// ... existing flags ...

	skipReleaseNotes = flag.Bool("skip-release-notes", false, "Skip release notes enrichment")
	releaseNotesURL  = flag.String("release-notes-url", "", "Override release notes URL")
	productID        = flag.String("product-id", "", "Override product ID detection")
	productConfig    = flag.String("product-config", "configs/products.yaml", "Path to product config file")
)
```

**Step 2: Implement enrichment orchestration**

```go
// Add to main() after comparison, before report generation

func enrichWithReleaseNotes(
	comparison *compare.ComparisonResults,
	newVersion string,
	productID string,
	config releasenotes.ProductConfig,
) (map[string]releasenotes.Match, error) {

	// Resolve URL
	url, err := config.ResolveURL(productID, newVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve URL: %w", err)
	}

	// Fetch release notes
	fetcher := releasenotes.NewFetcher()
	html, err := fetcher.Fetch(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release notes: %w", err)
	}

	// Parse features
	features, err := releasenotes.ParseHTML(html)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Collect property names
	var properties []string
	for _, change := range comparison.Added {
		properties = append(properties, change.PropertyName)
	}

	// Match properties to features
	matcher := releasenotes.NewMatcher(features)
	matches := matcher.Match(properties)

	return matches, nil
}

// In main():
func main() {
	// ... existing code ...

	var matches map[string]releasenotes.Match

	if !*skipReleaseNotes {
		// Load product config
		config, err := releasenotes.LoadProductConfig(*productConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not load product config: %v\n", err)
			fmt.Fprintf(os.Stderr, "Continuing with standard report...\n\n")
		} else {
			// Identify product
			prodID := *productID
			if prodID == "" {
				// Extract from metadata (TODO: access metadata map)
				prodID = "cf" // Default for now
			}

			// Try to enrich
			matches, err = enrichWithReleaseNotes(comparisonResults, newTileVersion, prodID, config)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Release notes enrichment failed: %v\n", err)
				fmt.Fprintf(os.Stderr, "Continuing with standard report...\n\n")
			} else {
				if *verbose {
					fmt.Printf("Enriched with %d property matches\n\n", len(matches))
				}
			}
		}
	}

	// Generate report
	categorized := report.CategorizeChanges(comparisonResults)

	if len(matches) > 0 {
		enriched := report.EnrichChanges(categorized, matches)
		output = report.GenerateTextReportWithFeatures(enriched, oldTileVersion, newTileVersion)
	} else {
		output = report.GenerateTextReport(categorized, oldTileVersion, newTileVersion)
	}

	// ... rest of output logic ...
}
```

**Step 3: Test manually**

```bash
cd .worktrees/release-notes-enrichment
make build

./tile-diff \
  --old-tile testdata/old.pivotal \
  --new-tile testdata/new.pivotal \
  --product-id cf \
  --verbose
```

Expected: Should attempt enrichment and fall back gracefully if release notes not found

**Step 4: Commit**

```bash
git add cmd/tile-diff/main.go
git commit -m "feat(cli): integrate release notes enrichment with graceful degradation"
```

---

## Phase 7: Testing & Documentation

### Task 11: Integration Test

**Files:**
- Create: `test/enrichment_test.go`

**Step 1: Write integration test**

```go
// test/enrichment_test.go
package test

import (
	"testing"

	"github.com/malston/tile-diff/pkg/releasenotes"
	"github.com/malston/tile-diff/pkg/report"
)

func TestReleaseNotesEnrichment_EndToEnd(t *testing.T) {
	// Load test config
	config, err := releasenotes.LoadProductConfig("../pkg/releasenotes/testdata/products.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Resolve URL
	url, err := config.ResolveURL("cf", "10.2.5")
	if err != nil {
		t.Fatalf("Failed to resolve URL: %v", err)
	}

	if url == "" {
		t.Error("Expected non-empty URL")
	}

	// This test verifies the full pipeline works
	// Actual HTTP fetching tested in fetcher_test.go with mock server
}
```

**Step 2: Run integration tests**

```bash
go test ./test -v
```

Expected: PASS

**Step 3: Commit**

```bash
git add test/enrichment_test.go
git commit -m "test: add release notes enrichment integration test"
```

---

### Task 12: Update Documentation

**Files:**
- Modify: `README.md`

**Step 1: Update README with new feature**

Add section after existing usage examples:

```markdown
### Release Notes Enrichment

Automatically enrich reports with feature context from product release notes:

```bash
./tile-diff \
  --old-tile srt-6.0.22.pivotal \
  --new-tile srt-10.2.5.pivotal \
  --product-id cf
```

The tool will:
1. Fetch release notes for the target version
2. Match property changes to documented features
3. Group related properties in the report
4. Enhance recommendations with feature context

**Configuration:**

Product release note URLs are configured in `configs/products.yaml`. Add custom products:

```yaml
my-product: "https://docs.example.com/{version}/release-notes.html"
```

**Flags:**

- `--skip-release-notes` - Disable enrichment
- `--release-notes-url <url>` - Override URL for this comparison
- `--product-id <id>` - Override product detection
- `--product-config <path>` - Use custom product config file
```

**Step 2: Commit**

```bash
git add README.md
git commit -m "docs: document release notes enrichment feature"
```

---

### Task 13: Final Testing & Cleanup

**Step 1: Run all tests**

```bash
go test ./... -v
```

Expected: All tests PASS

**Step 2: Run linter**

```bash
go vet ./...
```

Expected: No issues

**Step 3: Build and manual test**

```bash
make build
./tile-diff --help
```

Verify new flags appear in help text

**Step 4: Final commit**

```bash
git add .
git commit -m "chore: final cleanup and testing"
```

---

## Completion Checklist

- [ ] All tests passing
- [ ] Documentation updated
- [ ] Code follows project style
- [ ] Graceful error handling verified
- [ ] Manual testing with real tiles (if available)
- [ ] Ready for code review

---

## Follow-up Tasks (Future)

Not in scope for this implementation:

1. Support for multiple documentation formats
2. LLM-based semantic matching
3. Historical analysis across versions
4. Custom CSS selectors per product
5. Offline/cached release notes
6. JSON report enrichment (use existing format for now)
