# Tile-Diff User Guide

## Overview

`tile-diff` is a command-line tool that analyzes configuration changes between different versions of Tanzu Application Service (TAS) and other Ops Manager product tiles. It helps platform operators understand what configuration changes are required when upgrading tiles.

## What Problem Does It Solve?

When upgrading product tiles in Ops Manager, operators face several challenges:

- **Unknown Requirements**: What new properties must be configured?
- **Breaking Changes**: What existing properties are no longer supported?
- **Validation**: Will current configuration values work with the new version?
- **Risk Assessment**: What changes require immediate attention vs. optional updates?

Tile-diff automates this analysis, providing actionable reports that categorize changes by severity and provide specific recommendations.

## Core Features

### 1. Property Schema Comparison

Compares property definitions between two tile versions to identify:

- **New Properties**: Properties added in the new version
- **Removed Properties**: Properties that no longer exist
- **Changed Properties**: Properties with modified constraints, types, or defaults

### 2. Smart Categorization

Automatically categorizes changes into three severity levels:

- **🚨 Required Actions**: Changes that MUST be addressed before/during upgrade
  - New required properties without defaults
  - Properties that became non-optional
  - Constraint violations with current values

- **⚠️ Warnings**: Changes that should be reviewed but may not block upgrade
  - Removed properties currently in use
  - Type changes (rare but possible)
  - Changed constraints that don't violate current values

- **ℹ️ Informational**: Changes that are nice to know
  - New optional properties with defaults
  - Default value changes
  - Removed properties not currently in use

### 3. Current Configuration Analysis

When provided with Ops Manager credentials, tile-diff:

- Retrieves your current tile configuration via API
- Cross-references changes with what you've actually configured
- Filters out irrelevant changes (e.g., removed properties you never set)
- Validates current values against new constraints
- Provides context-aware recommendations

### 4. Multiple Output Formats

- **Text Reports**: Human-readable reports optimized for terminal display
- **JSON Reports**: Machine-readable output for automation and CI/CD pipelines

### 5. Actionable Recommendations

Each identified change includes:

- Property name and full path
- Property type and characteristics
- Current value (if available)
- Specific action required
- Contextual recommendation for your environment

## Installation

### Prerequisites

- Go 1.21 or later
- Access to product tile `.pivotal` files
- (Optional) Ops Manager API access for current config analysis

### Building from Source

```bash
git clone https://github.com/malston/tile-diff.git
cd tile-diff
make build
```

The compiled binary will be at `./tile-diff`.

### Verification

```bash
./tile-diff --help
```

You should see the available command-line flags.

## Basic Usage

### Scenario 1: Basic Comparison (Schema-Only)

Compare two tile versions without current configuration analysis:

```bash
./tile-diff \
  --old-tile /path/to/srt-6.0.22.pivotal \
  --new-tile /path/to/srt-10.2.5.pivotal
```

This shows all property changes between versions but doesn't indicate which affect your deployment.

**Use case**: Initial assessment of upgrade scope before diving into specifics.

### Scenario 2: Full Analysis with Current Config

Include your current Ops Manager configuration for contextualized analysis:

```bash
./tile-diff \
  --old-tile /path/to/srt-6.0.22.pivotal \
  --new-tile /path/to/srt-10.2.5.pivotal \
  --product-guid cf-abc123xyz \
  --ops-manager-url https://opsman.tas.example.com \
  --username admin \
  --password your-password \
  --skip-ssl-validation
```

This provides a filtered view showing only changes that affect your current deployment.

**Use case**: Pre-upgrade checklist - know exactly what you need to configure.

### Scenario 3: JSON Output for Automation

Generate machine-readable output for CI/CD pipelines:

```bash
./tile-diff \
  --old-tile /path/to/srt-6.0.22.pivotal \
  --new-tile /path/to/srt-10.2.5.pivotal \
  --format json > upgrade-analysis.json
```

**Use case**: Automated upgrade readiness checks in CI/CD workflows.

### Scenario 4: Auto-Detect Product GUID

Let tile-diff automatically find your product GUID from Ops Manager:

```bash
./tile-diff \
  --product-slug harbor-container-registry \
  --old-version 2.11.0 \
  --new-version 2.13.2 \
  --ops-manager-url https://opsman.tas.vcf.lab \
  --username admin \
  --password your-password \
  --skip-ssl-validation
```

**How it works**:
1. Queries Ops Manager API for all staged products
2. Matches the product slug to find the GUID
3. Uses that GUID to fetch current configuration
4. Generates filtered, actionable report

**Use case**: Simplify workflows when you know the product slug but not the GUID.

### Scenario 5: Formatted Report Without Credentials

Get professional upgrade analysis even without Ops Manager access:

```bash
./tile-diff \
  --old-tile /path/to/harbor-2.11.0.pivotal \
  --new-tile /path/to/harbor-2.13.2.pivotal
```

**Output**: Same professional "Ops Manager Tile Upgrade Analysis" format showing:
- Total changes summary
- Required Actions (properties that must be configured)
- Warnings (properties to review)
- Informational (optional new features)

**Difference from full analysis**:
- Shows ALL potential changes (no filtering by current config)
- Useful for initial assessment before accessing Ops Manager
- Perfect for offline analysis or planning sessions

**Use case**: Initial upgrade assessment, offline planning, or when Ops Manager access is restricted.

## Command-Line Reference

### Required Flags

| Flag | Description |
|------|-------------|
| `--old-tile` | Path to current version `.pivotal` file |
| `--new-tile` | Path to target version `.pivotal` file |

### Optional Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--product-guid` | Product GUID in Ops Manager | None |
| `--ops-manager-url` | Ops Manager URL | None |
| `--username` | Ops Manager username | None |
| `--password` | Ops Manager password | None |
| `--skip-ssl-validation` | Skip SSL certificate validation | false |
| `--format` | Output format: `text` or `json` | `text` |

### Finding Your Product GUID

```bash
# Using om CLI
om curl -p /api/v0/staged/products | jq -r '.[] | select(.type=="cf") | .guid'

# Or from Ops Manager UI
# Installation Dashboard → Tile → Settings → GUID shown in URL
```

## Understanding the Output

### Text Report Structure

```
tile-diff - Ops Manager Product Tile Comparison
================================================

Loading old tile: srt-6.0.22.pivotal
  Found 274 properties
Loading new tile: srt-10.2.5.pivotal
  Found 272 properties

Comparison Results:
===================

✨ New Properties (4):
  + property_name (type)

🗑️  Removed Properties (6):
  - property_name (type)

🔄 Changed Properties (0):
  ~ property_name: description

Summary:
  Properties in old tile: 274
  Properties in new tile: 272
  Added: 4, Removed: 6, Changed: 0
```

#### With Current Config Analysis

When Ops Manager credentials are provided, you get an enhanced report:

```
================================================================================
🚨 REQUIRED ACTIONS (2)
================================================================================

1. .properties.new_security_feature
   Type: boolean
   Status: New required property (no default)
   Current: Not set
   Action: Must configure this property before upgrading
   Recommendation: Set to 'true' for enhanced security scanning

2. .properties.memory_limit
   Type: integer
   Status: Constraint changed (min: 512 → 1024)
   Current: 768
   Action: Current value violates new minimum constraint
   Recommendation: Update to at least 1024 MB

================================================================================
⚠️  WARNINGS (3)
================================================================================

3. .properties.deprecated_setting
   Type: string
   Status: Removed in new version
   Current: "legacy-mode"
   Action: Property will be ignored after upgrade
   Recommendation: Review new configuration options

================================================================================
ℹ️  INFORMATIONAL (5)
================================================================================

4. .properties.enhanced_logging
   Type: boolean
   Status: New optional property
   Current: Not set
   Default: false
   Recommendation: Enable for improved observability

================================================================================
Summary: 2 required actions must be completed before upgrade
================================================================================
```

### JSON Report Structure

```json
{
  "metadata": {
    "old_version": "6.0.22",
    "new_version": "10.2.5",
    "analysis_date": "2025-12-11T10:30:00Z",
    "product_guid": "cf-abc123xyz",
    "current_config_available": true
  },
  "summary": {
    "total_old_properties": 274,
    "total_new_properties": 272,
    "added": 4,
    "removed": 6,
    "changed": 0,
    "required_actions": 2,
    "warnings": 3,
    "informational": 5
  },
  "required_actions": [
    {
      "property": ".properties.new_security_feature",
      "type": "boolean",
      "status": "New required property (no default)",
      "current_value": null,
      "action": "Must configure this property before upgrading",
      "recommendation": "Set to 'true' for enhanced security scanning"
    }
  ],
  "warnings": [...],
  "informational": [...]
}
```

## Common Workflows

### Pre-Upgrade Assessment

**Goal**: Understand upgrade scope before downloading tiles

1. Identify target version from release notes
2. Download both old and new `.pivotal` files
3. Run basic comparison (schema-only)
4. Review summary counts to assess scope

```bash
./tile-diff --old-tile current.pivotal --new-tile target.pivotal
```

### Upgrade Readiness Check

**Goal**: Create actionable checklist for upgrade execution

1. Run full analysis with current config
2. Export report for team review
3. Address all Required Actions
4. Document decisions for Warnings

```bash
./tile-diff \
  --old-tile current.pivotal \
  --new-tile target.pivotal \
  --product-guid $(get_product_guid) \
  --ops-manager-url $OPS_MANAGER_URL \
  --username $USERNAME \
  --password $PASSWORD > upgrade-checklist.txt
```

### Continuous Monitoring

**Goal**: Track configuration drift in CI/CD

1. Run tile-diff in JSON mode
2. Parse output programmatically
3. Fail pipeline if required actions > 0
4. Generate reports for operations team

```bash
#!/bin/bash
RESULT=$(./tile-diff --old-tile old.pivotal --new-tile new.pivotal --format json)
REQUIRED=$(echo "$RESULT" | jq '.summary.required_actions')

if [ "$REQUIRED" -gt 0 ]; then
  echo "ERROR: $REQUIRED required actions detected"
  exit 1
fi
```

## Best Practices

### 1. Always Test in Non-Production First

- Run tile-diff against staging/dev Ops Manager instances
- Validate recommendations before applying to production
- Compare results across environments (dev, stage, prod)

### 2. Version Control Your Analysis

```bash
# Save analysis for historical reference
./tile-diff \
  --old-tile v6.0.22.pivotal \
  --new-tile v10.2.5.pivotal > docs/upgrades/6.0.22-to-10.2.5-analysis.txt

git add docs/upgrades/
git commit -m "docs: upgrade analysis for 6.0.22 to 10.2.5"
```

### 3. Cross-Reference with Release Notes

Tile-diff shows **what** changed, release notes explain **why**:

1. Run tile-diff for technical details
2. Read release notes for feature context
3. Correlate property names with features
4. Make informed decisions on optional properties

### 4. Handle Credentials Securely

**Don't do this:**
```bash
./tile-diff --password mysecretpassword  # Password in shell history!
```

**Do this instead:**
```bash
# Use environment variables
export OPS_MANAGER_PASSWORD=$(op read "op://vault/opsman/password")
./tile-diff --password "$OPS_MANAGER_PASSWORD"

# Or use .envrc with direnv
# Add to .envrc:
# export OPS_MANAGER_PASSWORD=$(op read "op://vault/opsman/password")
```

### 5. Document Your Decisions

For each Required Action and Warning:

- Document the decision made (set to X, left as default, etc.)
- Explain rationale (security requirement, team policy, etc.)
- Note who approved the decision
- Include in upgrade runbook

## Troubleshooting

### Error: "failed to extract metadata.yml"

**Cause**: Invalid or corrupted `.pivotal` file

**Solution**:
1. Verify file is a valid `.pivotal` file (it's a ZIP archive)
2. Check file isn't truncated: `unzip -t file.pivotal`
3. Re-download from Tanzu Network if corrupted

### Error: "failed to connect to Ops Manager API"

**Cause**: Network, authentication, or SSL issues

**Solutions**:
- Verify Ops Manager URL is correct and accessible
- Check username/password are correct
- Use `--skip-ssl-validation` for self-signed certs
- Verify product GUID exists: `om curl -p /api/v0/staged/products`

### Warning: "Provide Ops Manager credentials for actionable report"

**Not an error**: This means you ran schema-only comparison

**Solution**: Add Ops Manager flags if you want current config analysis

### Output Shows No Changes

**Possible causes**:
1. Comparing identical tile versions
2. Only non-configurable properties changed
3. All changes filtered out (if using current config)

**Verify**:
```bash
# Check property counts in output
# Should show different total properties if versions differ
```

## Integration Examples

### Bash Script with Error Handling

```bash
#!/bin/bash
set -e

OLD_TILE="$1"
NEW_TILE="$2"

if [ -z "$OLD_TILE" ] || [ -z "$NEW_TILE" ]; then
  echo "Usage: $0 <old-tile> <new-tile>"
  exit 1
fi

# Get product GUID
PRODUCT_GUID=$(om curl -p /api/v0/staged/products | jq -r '.[] | select(.type=="cf") | .guid')

if [ -z "$PRODUCT_GUID" ]; then
  echo "ERROR: Could not find cf product"
  exit 1
fi

# Run analysis
OUTPUT_FILE="upgrade-analysis-$(date +%Y%m%d).txt"
./tile-diff \
  --old-tile "$OLD_TILE" \
  --new-tile "$NEW_TILE" \
  --product-guid "$PRODUCT_GUID" \
  --ops-manager-url "$OM_TARGET" \
  --username "$OM_USERNAME" \
  --password "$OM_PASSWORD" \
  --skip-ssl-validation > "$OUTPUT_FILE"

echo "Analysis saved to $OUTPUT_FILE"

# Check for required actions
if grep -q "🚨 REQUIRED ACTIONS" "$OUTPUT_FILE"; then
  echo "WARNING: Required actions detected!"
  exit 1
fi
```

### GitHub Actions Workflow

```yaml
name: Tile Upgrade Analysis

on:
  schedule:
    - cron: '0 0 * * 1'  # Weekly on Monday
  workflow_dispatch:

jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Download Old Tile
        run: |
          pivnet download-product-files \
            --product-slug elastic-runtime \
            --release-version ${{ env.OLD_VERSION }} \
            --glob "*.pivotal"

      - name: Download New Tile
        run: |
          pivnet download-product-files \
            --product-slug elastic-runtime \
            --release-version ${{ env.NEW_VERSION }} \
            --glob "*.pivotal"

      - name: Run Tile-Diff
        run: |
          ./tile-diff \
            --old-tile old.pivotal \
            --new-tile new.pivotal \
            --format json > analysis.json

      - name: Check Required Actions
        run: |
          REQUIRED=$(jq '.summary.required_actions' analysis.json)
          if [ "$REQUIRED" -gt 0 ]; then
            echo "::warning::$REQUIRED required actions detected"
          fi

      - name: Upload Analysis
        uses: actions/upload-artifact@v3
        with:
          name: upgrade-analysis
          path: analysis.json
```

## FAQ

### Q: Do I need Ops Manager access to use tile-diff?

**A**: No. Tile-diff works in two modes:
- **Schema-only mode**: Compare tiles without current config (no credentials needed)
- **Full analysis mode**: Include current config analysis (requires Ops Manager API access)

### Q: Can I compare tiles for different products?

**A**: Yes, tile-diff works with any Ops Manager product tile, not just TAS. The tool analyzes property schemas generically.

### Q: How do I get the `.pivotal` files?

**A**: Download from Tanzu Network (formerly Pivotal Network):
1. Log in to the Broadcom Support Portal: https://support.broadcom.com/
2. Navigate to your product (e.g., Tanzu Application Service)
3. Select version and download `.pivotal` file

### Q: Will tile-diff modify my Ops Manager configuration?

**A**: No. Tile-diff only **reads** configuration via API. It never modifies anything.

### Q: Can I automate configuration updates based on the report?

**A**: The report provides analysis, not automation. You should:
1. Review the report
2. Make informed decisions on each change
3. Apply changes manually via Ops Manager UI or `om configure-product` CLI

### Q: What if a property shows as "Required Action" but I don't know what value to use?

**A**:
1. Check tile-diff recommendations (it provides context-aware suggestions)
2. Read product release notes for feature details
3. Consult product documentation
4. Test in non-production environment
5. Contact support if unclear

### Q: Can I compare more than two versions at once?

**A**: Not currently. Tile-diff compares two versions (A → B). For multi-version analysis:
```bash
./tile-diff --old-tile v6.pivotal --new-tile v8.pivotal
./tile-diff --old-tile v8.pivotal --new-tile v10.pivotal
```

## Support and Contributing

- **Issues**: https://github.com/malston/tile-diff/issues
- **Documentation**: https://github.com/malston/tile-diff/docs
- **Contributing**: See CONTRIBUTING.md in repository

## Next Steps

1. **Try it out**: Run basic comparison on your tiles
2. **Read examples**: See EXAMPLES.md for real-world scenarios
3. **Integrate**: Add to your upgrade workflow
4. **Share feedback**: Help improve the tool
