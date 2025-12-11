# Debug Matching Feature

## Overview

The `--debug-matching` flag provides detailed information about how tile properties are matched to release note features. This is useful for:

- Understanding why certain properties are grouped together
- Tuning the matching algorithm
- Debugging missing or incorrect matches
- Analyzing the quality of release notes parsing

## Usage

```bash
./tile-diff \
  --old-tile /path/to/old.pivotal \
  --new-tile /path/to/new.pivotal \
  --ops-manager-url https://opsman.example.com \
  --username admin \
  --password $OM_PASSWORD \
  --skip-ssl-validation \
  --release-notes-url https://docs.example.com/release-notes \
  --debug-matching
```

## Output Format

The debug output is divided into three main sections:

### 1. Summary Statistics

```
================================================================================
PROPERTY-TO-FEATURE MATCHING DEBUG
================================================================================

Total Features Found: 15
Total Properties to Match: 3
Successful Matches: 2
Unmatched Properties: 1
```

### 2. Features Extracted from Release Notes

Shows all `<h2>` sections found in the release notes:

```
--------------------------------------------------------------------------------
FEATURES EXTRACTED FROM RELEASE NOTES
--------------------------------------------------------------------------------

[1] v10.2.0
    Description: This release introduces runtime container vulnerability detection. Configure using the .properties.security_scanner_enabled property...

[2] v10.2.4
    Description: Feature improvements and bug fixes...

[3] Enhanced Security Scanning
    Description: New security scanning capabilities with real-time detection...
```

### 3. Matched Properties

Groups properties by the feature they matched to:

```
--------------------------------------------------------------------------------
MATCHED PROPERTIES
--------------------------------------------------------------------------------

📦 v10.2.0 (2 properties)
   ✓ .properties.security_scanner_enabled
      Match Type: direct
      Confidence: 1.00
   ✓ .properties.scanner_update_interval
      Match Type: keyword
      Confidence: 0.67
```

**Match Types:**
- **direct**: Property name appears verbatim in feature description
- **keyword**: Multiple keywords from property name found in feature description

**Confidence Scores:**
- `1.00`: Direct match (property name found exactly)
- `0.51-0.90`: Keyword match (percentage of keywords found × 0.9)
- `≤0.50`: Below threshold, not matched

### 4. Unmatched Properties

Lists properties that couldn't be matched to any feature:

```
--------------------------------------------------------------------------------
UNMATCHED PROPERTIES
--------------------------------------------------------------------------------

These properties could not be matched to any release note feature:
   ✗ .properties.internal_cache_size

Tip: Unmatched properties may indicate:
  - Property names don't appear in release notes
  - Keyword matching threshold (0.5) not met
  - Property is an internal implementation detail
```

## Matching Algorithm Details

The matcher uses a two-phase approach:

### Phase 1: Direct Matching (Confidence: 1.0)

Looks for exact property name in feature description.

**Example:**
- Property: `.properties.security_scanner_enabled`
- Feature description: "Configure using the .properties.security_scanner_enabled property"
- Result: ✓ Direct match

### Phase 2: Keyword Matching (Confidence: 0.0-0.9)

If no direct match, tokenizes the property name and searches for keywords:

1. **Tokenization:**
   - Remove common prefixes: `.properties.`, `.cloud_controller.`
   - Split on underscores, dots, dashes
   - Filter stopwords: `enable`, `enabled`, `setting`, `new`, `property`, etc.
   - Filter short tokens (< 3 chars)

2. **Scoring:**
   - Count how many tokens appear in feature title + description
   - Score = (matched tokens / total tokens) × 0.9
   - Only accept if score > 0.5

**Example:**
- Property: `.properties.app_log_rate_limiting`
- Tokens: `["app", "log", "rate", "limiting"]`
- Feature: "Application log rate limiting is now available"
- Matched: 4/4 tokens
- Score: (4/4) × 0.9 = 0.90
- Result: ✓ Keyword match

## Interpreting Results

### High Match Rate (>80%)

Good! Your release notes and property naming are well-aligned.

### Medium Match Rate (50-80%)

Consider:
- Are property names descriptive enough?
- Do release notes mention property names explicitly?
- Is the 0.5 confidence threshold appropriate?

### Low Match Rate (<50%)

Possible issues:
- Release notes don't describe configuration properties
- Property names are implementation details not documented
- HTML structure doesn't match expected format (h2 tags for features)

## Tuning the Matcher

If you need to adjust matching behavior, see:

- `pkg/releasenotes/matcher.go` - Main matching logic
- Adjust `stopwords` list for domain-specific terms
- Modify confidence threshold (currently 0.5)
- Add custom tokenization rules

## Example Workflow

1. Run with `--debug-matching` to see current matches
2. Identify unmatched properties
3. Check if property names appear in release notes
4. Consider:
   - Adding property names to release notes
   - Adjusting tokenization rules
   - Lowering confidence threshold (carefully)
5. Re-run and verify improvements

## Tips

- Use `--verbose` along with `--debug-matching` for maximum detail
- Pipe output to a file for easier analysis: `./tile-diff ... --debug-matching > debug.txt 2>&1`
- Compare matched vs. unmatched properties to identify patterns
- Direct matches (confidence 1.0) are most reliable
