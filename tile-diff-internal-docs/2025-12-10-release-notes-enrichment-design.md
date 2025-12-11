# Release Notes Enrichment Design

**Date:** 2025-12-10
**Status:** Draft
**Author:** Claude & Mark

## Overview

Enhance tile-diff reports by cross-referencing property changes with product release notes. This provides operators with feature context, grouping related properties and enriching recommendations with information from official documentation.

## Problem Statement

Current tile-diff reports show *what* changed (new properties, removed properties, type changes) but not *why* those changes exist or what features they enable. Operators must manually cross-reference release notes to understand the purpose of new properties.

Example: A report shows three new required properties but doesn't indicate they're all part of a single "Enhanced Security Scanning" feature.

## Goals

1. **Automatic enrichment** - Fetch and parse release notes for the target tile version
2. **Feature grouping** - Group related properties together when they belong to the same feature
3. **Context enhancement** - Add feature descriptions and context to property recommendations
4. **Graceful degradation** - Never break core functionality if enrichment fails

## Non-Goals

- LLM-based inference of property-to-feature relationships (too complex for v1)
- Support for all possible documentation formats (focus on Broadcom HTML structure)
- Historical analysis across multiple versions (only target version)

## Architecture

### Component Overview

Release notes enrichment sits between comparison and reporting:

```
┌─────────────────┐
│ Tile Comparison │
│   (existing)    │
└────────┬────────┘
         │
         v
┌─────────────────────────────┐
│  Release Notes Enrichment   │
│  ┌────────────────────────┐ │
│  │ 1. Product Registry    │ │
│  │ 2. Fetcher             │ │
│  │ 3. Parser              │ │
│  │ 4. Matcher             │ │
│  └────────────────────────┘ │
└────────┬────────────────────┘
         │
         v
┌─────────────────┐
│ Report Generator│
│   (enhanced)    │
└─────────────────┘
```

### New Packages

**`pkg/releasenotes/registry.go`**
- Loads product configuration mapping product IDs to release note URLs
- Resolves URLs with version substitution
- Product identification from tile metadata

**`pkg/releasenotes/fetcher.go`**
- Downloads release notes HTML from Broadcom docs
- In-memory caching for session
- Timeout and error handling

**`pkg/releasenotes/parser.go`**
- Parses HTML to extract features and descriptions
- Identifies feature sections (headings and content)
- Preserves structure for proximity matching

**`pkg/releasenotes/matcher.go`**
- Links properties to features using multiple strategies
- Confidence scoring for match quality
- Ranking to select best matches

**`pkg/report/enricher.go`**
- Adds feature context to comparison results
- Groups properties by feature
- Enhances recommendations

## Product Configuration

### Config File Format

`configs/products.yaml`:

```yaml
# Product ID to release notes URL pattern
# {version} placeholder replaced with tile version

cf: "https://techdocs.broadcom.com/us/en/vmware-tanzu/platform/tanzu-application-service/{version}/release-notes.html"
p-mysql: "https://techdocs.broadcom.com/us/en/vmware-tanzu/data-solutions/tanzu-sql-mysql/{version}/release-notes.html"
p-rabbitmq: "https://techdocs.broadcom.com/us/en/vmware-tanzu/data-solutions/tanzu-rabbitmq/{version}/release-notes.html"
```

### Product Identification

1. Extract `product_name` or `name` field from tile `metadata/metadata.yml`
2. Normalize to product ID (lowercase, hyphens)
3. Lookup in product config
4. Fallback to CLI flag `--product-id` if not found
5. Skip enrichment if product unknown (with warning)

### URL Resolution

- Replace `{version}` placeholder with new tile version
- Handle version format normalization (e.g., "10.2.5" → "10-2-5" if needed)
- Validate URL format before fetching

## Matching Strategy

### Data Structures

```go
type Feature struct {
    Title       string
    Description string
    Position    int  // For proximity matching
}

type Match struct {
    Property   string
    Feature    Feature
    MatchType  string  // "direct", "keyword", "proximity"
    Confidence float64 // 0.0-1.0
}
```

### Matching Types

**1. Direct Matching (Confidence: 1.0)**

Property name appears verbatim in release notes text.

Example: Feature text contains ".properties.security_scanner_enabled"

**2. Keyword Matching (Confidence: 0.3-0.9)**

Tokenize property names and match against feature descriptions:

1. Split property name: `app_log_rate_limiting` → `["app", "log", "rate", "limiting"]`
2. Filter stopwords: `["enable", "enabled", "setting", "config", "new"]`
3. Match keywords against feature text
4. Score: `(matched_keywords / total_keywords) × 0.9`
5. Minimum 2 keywords required

**3. Proximity Matching (Confidence: 0.2-0.5)**

Version-based association when direct/keyword matching fails:

- Property added in v10.2.5
- Feature section mentions "New in 10.2.5"
- Associate properties within same section
- Lower confidence ("possibly related")

### Match Selection

- Each property gets 0-N matches
- Rank by confidence score
- Accept top match if confidence > 0.5
- Otherwise, leave property ungrouped

## Report Enrichment

### Text Report Format

**Feature Grouping:**

```
================================================================================
🚨 REQUIRED ACTIONS
================================================================================

📦 Enhanced Security Scanning (2 properties)
   Introduced in v10.2.5 to provide runtime container vulnerability detection

1. .properties.security_scanner_enabled
   Type: boolean
   Status: New required property (no default)
   Current: Not set
   Feature: Part of Enhanced Security Scanning
   Action: Must configure this property before upgrading
   Recommendation: Enable for production environments to detect CVEs

2. .properties.scanner_update_interval
   Type: integer
   Status: New required property (no default)
   Current: Not set
   Feature: Part of Enhanced Security Scanning
   Action: Must configure this property before upgrading
   Recommendation: Set to 3600 (hourly) for active monitoring

-- Ungrouped Properties --

3. .properties.unrelated_property
   Type: string
   ...
```

**Individual Property Context:**

- **Feature:** line showing which feature it belongs to
- Enhanced **Recommendation:** Original + release note context

### JSON Report Format

```json
{
  "metadata": {
    "old_version": "6.0.22",
    "new_version": "10.2.5",
    "release_notes_url": "https://...",
    "enrichment_status": "success"
  },
  "features": [
    {
      "name": "Enhanced Security Scanning",
      "description": "Runtime container vulnerability detection...",
      "properties": [
        ".properties.security_scanner_enabled",
        ".properties.scanner_update_interval"
      ],
      "source_url": "https://..."
    }
  ],
  "changes": {
    "required_actions": [
      {
        "property": ".properties.security_scanner_enabled",
        "type": "boolean",
        "status": "New required property (no default)",
        "feature_name": "Enhanced Security Scanning",
        "feature_description": "Runtime container vulnerability detection...",
        "recommendation": "Enable for production environments to detect CVEs"
      }
    ]
  }
}
```

## Error Handling

### Graceful Degradation

**Principle:** Release note enrichment is optional. Core comparison functionality must never fail due to enrichment issues.

### Failure Scenarios

**1. Product Config Not Found**

```
Warning: No release notes URL configured for product 'p-redis'
Add to configs/products.yaml or use --release-notes-url flag
Continuing with standard report...
```

- Log warning
- Continue with standard report

**2. Release Notes Fetch Failed**

```
Warning: Could not fetch release notes from https://...
Error: HTTP 404 Not Found
Continuing with standard report...
```

Common causes:
- Version doesn't exist yet on docs site
- URL pattern incorrect
- Network connectivity issues

**3. HTML Parsing Failed**

```
Warning: Could not parse release notes HTML
Error: Unexpected document structure
Continuing with standard report...
```

Broadcom may change HTML structure. Fail gracefully.

**4. No Matches Found**

```
Info: Fetched release notes but found no property matches
Release notes URL: https://...
```

This is acceptable - not all properties mentioned explicitly.

### CLI Flags

- `--skip-release-notes` - Explicitly disable enrichment
- `--release-notes-url <url>` - Override URL for current comparison
- `--product-id <id>` - Override product identification
- `--verbose` - Show matching details (confidence scores, etc.)

## Testing Strategy

### Unit Tests

**`registry_test.go`**
- Load product config
- Resolve URLs with version substitution
- Handle missing products
- Test product identification from metadata

**`parser_test.go`**
- Parse HTML fixtures
- Extract feature sections correctly
- Handle malformed HTML

**`matcher_test.go`**
- Direct matching (exact property names)
- Keyword matching with various property name formats
- Proximity matching based on version
- Confidence scoring

**`enricher_test.go`**
- Add feature context to comparison results
- Group properties by feature
- Enhance recommendations

### Integration Tests

**Test with Real Data:**
- Check in HTML snapshot from Broadcom TAS release notes
- Verify actual property matching against real release notes
- Test end-to-end enrichment flow

**Test Graceful Degradation:**
- Missing product config
- 404 responses
- Parse failures
- No matches found

### Test Fixtures

```
testdata/
├── products.yaml                      # Test product config
├── cf-10.2.5-release-notes.html      # Real HTML snapshot
├── expected-matches.json              # Expected matching results
└── malformed.html                     # Edge cases
```

## Implementation Phases

### Phase 1: Infrastructure (Week 1)

1. Product registry and config loading
2. URL resolution and fetching
3. Basic HTML parsing (extract text)
4. Error handling framework

**Deliverable:** Can fetch release notes for configured products

### Phase 2: Matching Logic (Week 2)

1. Direct property name matching
2. Keyword tokenization and matching
3. Proximity matching
4. Confidence scoring

**Deliverable:** Can match properties to features with confidence scores

### Phase 3: Report Integration (Week 3)

1. Report enricher
2. Feature grouping in text reports
3. JSON format updates
4. CLI flags

**Deliverable:** Enriched reports with feature context

### Phase 4: Polish (Week 4)

1. Add common products to config
2. Documentation
3. Integration tests
4. Performance optimization (caching)

**Deliverable:** Production-ready feature

## Future Enhancements

**Not in scope for v1, consider later:**

- Support for multiple documentation formats (Markdown, PDF)
- LLM-based semantic matching for better accuracy
- Historical analysis (compare release notes across version range)
- Custom parsing rules per product (CSS selectors config)
- Offline mode (pre-downloaded release notes)
- Release notes caching across sessions

## Success Criteria

1. **Accuracy:** 80%+ of explicitly mentioned properties matched correctly
2. **Reliability:** Never breaks core comparison functionality
3. **Usability:** Feature grouping helps operators understand changes
4. **Extensibility:** Users can add new products to config
5. **Performance:** Enrichment adds < 5 seconds to comparison time

## Open Questions

1. Should we cache release notes HTML to disk for offline use?
   - **Decision:** No for v1, keep it simple with in-memory caching

2. How to handle release notes that span multiple pages?
   - **Decision:** Assume single-page format for v1, enhance later if needed

3. Should confidence thresholds be configurable?
   - **Decision:** Hardcode for v1 (0.5 threshold), make configurable if needed

## Approval

- [ ] Design reviewed by: Mark
- [ ] Ready for implementation: ___________
