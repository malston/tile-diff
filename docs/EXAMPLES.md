# Tile-Diff Examples and Use Cases

Real-world examples showing how to use tile-diff in different scenarios.

## Table of Contents

1. [TAS Upgrade Planning](#example-1-tas-upgrade-planning)
2. [Multi-Environment Upgrade Strategy](#example-2-multi-environment-upgrade-strategy)
3. [CI/CD Integration](#example-3-cicd-integration)
4. [Emergency Hotfix Analysis](#example-4-emergency-hotfix-analysis)
5. [Data Services Tile Upgrade](#example-5-data-services-tile-upgrade)
6. [Configuration Audit](#example-6-configuration-audit)
7. [Upgrade Runbook Generation](#example-7-upgrade-runbook-generation)

---

## Example 1: TAS Upgrade Planning

**Scenario**: Your team is running TAS 6.0.22 in production and needs to upgrade to 10.2.5.

### Step 1: Initial Assessment

```bash
# Download tiles from Tanzu Network
pivnet download-product-files \
  --product-slug elastic-runtime \
  --release-version 6.0.22 \
  --glob "srt-*.pivotal"

pivnet download-product-files \
  --product-slug elastic-runtime \
  --release-version 10.2.5 \
  --glob "srt-*.pivotal"

# Run basic comparison
./tile-diff \
  --old-tile srt-6.0.22-build.2.pivotal \
  --new-tile srt-10.2.5-build.2.pivotal > initial-assessment.txt
```

**Output Analysis:**
```
Properties in old tile: 274
Properties in new tile: 272
Added: 4, Removed: 6, Changed: 0
```

**Decision**: This is a moderate change (10 total property changes). Proceed with detailed analysis.

### Step 2: Production Config Analysis

```bash
# Load Ops Manager credentials
source .envrc  # Contains OM_TARGET, OM_USERNAME, OM_PASSWORD

# Get production product GUID
PROD_GUID=$(om curl -p /api/v0/staged/products | \
  jq -r '.[] | select(.type=="cf") | .guid')

echo "Production CF GUID: $PROD_GUID"

# Run full analysis
./tile-diff \
  --old-tile srt-6.0.22-build.2.pivotal \
  --new-tile srt-10.2.5-build.2.pivotal \
  --product-guid "$PROD_GUID" \
  --ops-manager-url "$OM_TARGET" \
  --username "$OM_USERNAME" \
  --password "$OM_PASSWORD" \
  --skip-ssl-validation > production-upgrade-analysis.txt
```

**Output (Sample):**
```
================================================================================
🚨 REQUIRED ACTIONS (1)
================================================================================

1. .properties.license_key
   Type: tanzu_license_key
   Status: New required property (no default)
   Current: Not set
   Action: Must configure this property before upgrading
   Recommendation: Obtain license key from Tanzu Network

================================================================================
⚠️ WARNINGS (2)
================================================================================

2. .properties.push_apps_manager_offline_tools
   Type: multi_select_options
   Status: Removed in new version
   Current: ["cf-cli-linux", "cf-cli-windows"]
   Action: Property will be ignored after upgrade
   Recommendation: Apps Manager CLI tools now auto-updated via new mechanism

3. .properties.legacy_md5_buildpack_paths_enabled
   Type: boolean
   Status: Removed in new version
   Current: true
   Action: Property will be ignored after upgrade
   Recommendation: New version only supports SHA256, ensure buildpacks compatible

================================================================================
ℹ️ INFORMATIONAL (3)
================================================================================

4. .properties.tanzu_cf_cli_enable_auto_upgrades
   Type: boolean
   Status: New optional property
   Current: Not set
   Default: true
   Recommendation: Enable for automatic CF CLI updates in Apps Manager

5. .properties.push_apps_manager_license_expiration_warning
   Type: boolean
   Status: New optional property
   Current: Not set
   Default: true
   Recommendation: Keep enabled to warn users before license expires

6. .properties.router_enable_log_attempt_details
   Type: boolean
   Status: New optional property
   Current: Not set
   Default: false
   Recommendation: Enable for enhanced router logging (may impact performance)
```

### Step 3: Create Upgrade Checklist

```bash
# Extract required actions for upgrade runbook
grep -A 5 "🚨 REQUIRED ACTIONS" production-upgrade-analysis.txt > upgrade-checklist.txt
grep -A 5 "⚠️ WARNINGS" production-upgrade-analysis.txt >> upgrade-checklist.txt

cat upgrade-checklist.txt
```

### Step 4: Team Review

```markdown
# TAS 6.0.22 → 10.2.5 Upgrade - Required Actions

## Must Do Before Upgrade

1. **Obtain Tanzu License Key**
   - Contact: Broadcom support / TAM
   - Add to `.properties.license_key` in Ops Manager
   - Status: ⏸️ Blocked on license provisioning

## Review Before Upgrade

2. **Offline CLI Tools Removal**
   - Current: Using offline tools feature
   - Impact: Feature removed, new auto-update mechanism
   - Action: Test Apps Manager with new mechanism in staging
   - Decision: ✅ Acceptable, new method preferred

3. **MD5 Buildpack Path Support Removed**
   - Current: Enabled (legacy buildpacks)
   - Impact: Must use SHA256 buildpack references only
   - Action: Audit all custom buildpacks for SHA256 support
   - Status: ⏸️ Buildpack audit in progress

## Optional Improvements

4. **Enable CF CLI Auto-Updates** - Recommend: Yes
5. **License Expiration Warnings** - Recommend: Yes (default)
6. **Enhanced Router Logging** - Recommend: No (performance sensitive)
```

**Team Decision**: Block upgrade until:
1. License key obtained
2. Buildpack audit completed and buildpacks updated if needed

---

## Example 2: Multi-Environment Upgrade Strategy

**Scenario**: Upgrade dev → stage → prod with validation at each step.

### Upgrade Script with Validation

```bash
#!/bin/bash
# upgrade-validator.sh - Validate upgrade readiness across environments

set -e

ENVIRONMENTS=("dev" "staging" "production")
OLD_TILE="$1"
NEW_TILE="$2"

if [ -z "$OLD_TILE" ] || [ -z "$NEW_TILE" ]; then
  echo "Usage: $0 <old-tile> <new-tile>"
  exit 1
fi

for ENV in "${ENVIRONMENTS[@]}"; do
  echo "========================================="
  echo "Analyzing $ENV environment"
  echo "========================================="

  # Load environment-specific credentials
  source ".envrc.${ENV}"

  # Get product GUID for this environment
  GUID=$(om curl -p /api/v0/staged/products | \
    jq -r '.[] | select(.type=="cf") | .guid')

  if [ -z "$GUID" ]; then
    echo "ERROR: No cf product found in $ENV"
    exit 1
  fi

  # Run analysis
  OUTPUT_FILE="analysis-${ENV}.txt"
  ./tile-diff \
    --old-tile "$OLD_TILE" \
    --new-tile "$NEW_TILE" \
    --product-guid "$GUID" \
    --ops-manager-url "$OM_TARGET" \
    --username "$OM_USERNAME" \
    --password "$OM_PASSWORD" \
    --skip-ssl-validation > "$OUTPUT_FILE"

  # Check for required actions
  if grep -q "🚨 REQUIRED ACTIONS" "$OUTPUT_FILE"; then
    REQUIRED_COUNT=$(grep -A 1 "🚨 REQUIRED ACTIONS" "$OUTPUT_FILE" | \
      grep -oP '\(\K[0-9]+')
    echo "⚠️  $ENV has $REQUIRED_COUNT required actions"
  else
    echo "✅ $ENV has no required actions"
  fi

  # Check for warnings
  if grep -q "⚠️ WARNINGS" "$OUTPUT_FILE"; then
    WARNING_COUNT=$(grep -A 1 "⚠️ WARNINGS" "$OUTPUT_FILE" | \
      grep -oP '\(\K[0-9]+')
    echo "⚠️  $ENV has $WARNING_COUNT warnings"
  else
    echo "✅ $ENV has no warnings"
  fi

  echo ""
done

echo "========================================="
echo "Summary"
echo "========================================="
echo "Analysis files generated:"
for ENV in "${ENVIRONMENTS[@]}"; do
  echo "  - analysis-${ENV}.txt"
done
echo ""
echo "Review all files before proceeding with upgrade"
```

### Usage

```bash
./upgrade-validator.sh \
  srt-6.0.22-build.2.pivotal \
  srt-10.2.5-build.2.pivotal
```

**Output:**
```
=========================================
Analyzing dev environment
=========================================
✅ dev has no required actions
✅ dev has no warnings

=========================================
Analyzing staging environment
=========================================
⚠️  staging has 1 required actions
⚠️  staging has 2 warnings

=========================================
Analyzing production environment
=========================================
⚠️  production has 1 required actions
⚠️  production has 3 warnings

=========================================
Summary
=========================================
Analysis files generated:
  - analysis-dev.txt
  - analysis-staging.txt
  - analysis-production.txt

Review all files before proceeding with upgrade
```

**Strategy**: Dev is clean, proceed immediately. Staging and prod need attention - review detailed reports.

---

## Example 3: CI/CD Integration

**Scenario**: Automated upgrade readiness checks in GitHub Actions.

### GitHub Actions Workflow

```yaml
# .github/workflows/tile-upgrade-check.yml
name: Tile Upgrade Readiness Check

on:
  schedule:
    - cron: '0 0 * * 1'  # Weekly on Monday
  workflow_dispatch:
    inputs:
      old_version:
        description: 'Current tile version'
        required: true
      new_version:
        description: 'Target tile version'
        required: true

env:
  PRODUCT: elastic-runtime

jobs:
  check-upgrade:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Build tile-diff
        run: |
          cd tile-diff
          make build

      - name: Download Old Tile
        env:
          PIVNET_TOKEN: ${{ secrets.PIVNET_TOKEN }}
        run: |
          pivnet login --api-token="$PIVNET_TOKEN"
          pivnet download-product-files \
            --product-slug "$PRODUCT" \
            --release-version "${{ github.event.inputs.old_version }}" \
            --glob "*.pivotal" \
            --download-dir ./tiles

      - name: Download New Tile
        env:
          PIVNET_TOKEN: ${{ secrets.PIVNET_TOKEN }}
        run: |
          pivnet download-product-files \
            --product-slug "$PRODUCT" \
            --release-version "${{ github.event.inputs.new_version }}" \
            --glob "*.pivotal" \
            --download-dir ./tiles

      - name: Run Tile-Diff Analysis
        run: |
          cd tile-diff
          ./tile-diff \
            --old-tile ../tiles/*${{ github.event.inputs.old_version }}*.pivotal \
            --new-tile ../tiles/*${{ github.event.inputs.new_version }}*.pivotal \
            --format json > ../analysis.json

      - name: Parse Results
        id: results
        run: |
          REQUIRED=$(jq '.summary.required_actions' analysis.json)
          WARNINGS=$(jq '.summary.warnings' analysis.json)
          TOTAL=$(jq '.summary.added + .summary.removed + .summary.changed' analysis.json)

          echo "required=${REQUIRED}" >> $GITHUB_OUTPUT
          echo "warnings=${WARNINGS}" >> $GITHUB_OUTPUT
          echo "total=${TOTAL}" >> $GITHUB_OUTPUT

      - name: Create Issue if Required Actions Found
        if: steps.results.outputs.required > 0
        uses: actions/github-script@v6
        with:
          script: |
            const required = ${{ steps.results.outputs.required }};
            const warnings = ${{ steps.results.outputs.warnings }};
            const total = ${{ steps.results.outputs.total }};

            const body = `## TAS Upgrade Readiness Alert

            Upgrade from **${{ github.event.inputs.old_version }}** to **${{ github.event.inputs.new_version }}** requires attention:

            - 🚨 **Required Actions**: ${required}
            - ⚠️ **Warnings**: ${warnings}
            - Total Changes: ${total}

            ### Next Steps

            1. Review detailed analysis in workflow artifacts
            2. Plan configuration changes
            3. Test in non-production environment
            4. Schedule production upgrade

            **Workflow Run**: ${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}
            `;

            await github.rest.issues.create({
              owner: context.repo.owner,
              repo: context.repo.repo,
              title: `TAS Upgrade ${{ github.event.inputs.old_version }} → ${{ github.event.inputs.new_version }}: ${required} Required Actions`,
              body: body,
              labels: ['upgrade', 'tas', 'required-action']
            });

      - name: Upload Analysis Report
        uses: actions/upload-artifact@v3
        with:
          name: upgrade-analysis-${{ github.event.inputs.old_version }}-to-${{ github.event.inputs.new_version }}
          path: analysis.json

      - name: Comment on PR if triggered by PR
        if: github.event_name == 'pull_request'
        uses: actions/github-script@v6
        with:
          script: |
            const fs = require('fs');
            const analysis = JSON.parse(fs.readFileSync('analysis.json', 'utf8'));

            const body = `## Tile Upgrade Analysis

            | Metric | Count |
            |--------|-------|
            | Required Actions | ${analysis.summary.required_actions} |
            | Warnings | ${analysis.summary.warnings} |
            | New Properties | ${analysis.summary.added} |
            | Removed Properties | ${analysis.summary.removed} |
            | Changed Properties | ${analysis.summary.changed} |

            ${analysis.summary.required_actions > 0 ? '⚠️ **Action required before merge**' : '✅ No blocking issues'}
            `;

            await github.rest.issues.createComment({
              owner: context.repo.owner,
              repo: context.repo.repo,
              issue_number: context.issue.number,
              body: body
            });

      - name: Fail if Critical Issues
        if: steps.results.outputs.required > 0
        run: |
          echo "::error::Found ${{ steps.results.outputs.required }} required actions"
          exit 1
```

### Trigger Workflow

```bash
# Via GitHub CLI
gh workflow run tile-upgrade-check.yml \
  -f old_version=6.0.22 \
  -f new_version=10.2.5

# Or via UI: Actions → Tile Upgrade Readiness Check → Run workflow
```

---

## Example 4: Emergency Hotfix Analysis

**Scenario**: Critical security patch released, need immediate impact assessment.

### Quick Assessment Script

```bash
#!/bin/bash
# hotfix-assessment.sh - Fast analysis for urgent patches

set -e

OLD_VERSION="${1:-6.0.22}"
NEW_VERSION="${2:-6.0.23}"  # Hotfix version

echo "🚨 EMERGENCY HOTFIX ANALYSIS: $OLD_VERSION → $NEW_VERSION"
echo "=================================================="
echo ""

# Download tiles
echo "Downloading tiles..."
pivnet download-product-files \
  --product-slug elastic-runtime \
  --release-version "$OLD_VERSION" \
  --glob "*.pivotal" \
  --download-dir /tmp &

pivnet download-product-files \
  --product-slug elastic-runtime \
  --release-version "$NEW_VERSION" \
  --glob "*.pivotal" \
  --download-dir /tmp &

wait  # Download in parallel

# Quick analysis without current config (faster)
OLD_TILE=$(ls /tmp/srt-${OLD_VERSION}*.pivotal)
NEW_TILE=$(ls /tmp/srt-${NEW_VERSION}*.pivotal)

echo "Running quick analysis..."
./tile-diff \
  --old-tile "$OLD_TILE" \
  --new-tile "$NEW_TILE" \
  --format json > /tmp/hotfix-analysis.json

# Extract critical info
ADDED=$(jq '.summary.added' /tmp/hotfix-analysis.json)
REMOVED=$(jq '.summary.removed' /tmp/hotfix-analysis.json)
CHANGED=$(jq '.summary.changed' /tmp/hotfix-analysis.json)

echo ""
echo "IMPACT SUMMARY"
echo "=============="
echo "Properties Added: $ADDED"
echo "Properties Removed: $REMOVED"
echo "Properties Changed: $CHANGED"
echo ""

if [ "$ADDED" -eq 0 ] && [ "$REMOVED" -eq 0 ] && [ "$CHANGED" -eq 0 ]; then
  echo "✅ NO CONFIGURATION CHANGES"
  echo "   This appears to be a code-only patch"
  echo "   Safe to apply immediately (after testing)"
elif [ "$ADDED" -gt 0 ]; then
  echo "⚠️  NEW PROPERTIES DETECTED"
  echo "   Review new properties before applying"
  jq -r '.required_actions[]? | "   - " + .property + " (" + .type + ")"' \
    /tmp/hotfix-analysis.json
else
  echo "⚠️  CONFIGURATION CHANGES DETECTED"
  echo "   Full analysis recommended before applying"
fi

echo ""
echo "Full analysis saved to: /tmp/hotfix-analysis.json"
```

### Usage

```bash
./hotfix-assessment.sh 6.0.22 6.0.23
```

**Output:**
```
🚨 EMERGENCY HOTFIX ANALYSIS: 6.0.22 → 6.0.23
==================================================

Downloading tiles...
Running quick analysis...

IMPACT SUMMARY
==============
Properties Added: 0
Properties Removed: 0
Properties Changed: 0

✅ NO CONFIGURATION CHANGES
   This appears to be a code-only patch
   Safe to apply immediately (after testing)

Full analysis saved to: /tmp/hotfix-analysis.json
```

**Decision**: Code-only patch, no config changes needed. Proceed with hotfix in non-prod for validation, then prod.

---

## Example 5: Data Services Tile Upgrade

**Scenario**: Upgrading MySQL for Tanzu from 2.10.0 to 2.11.0.

### Analysis Command

```bash
# Get MySQL product GUID
MYSQL_GUID=$(om curl -p /api/v0/staged/products | \
  jq -r '.[] | select(.type=="p-mysql") | .guid')

echo "MySQL product GUID: $MYSQL_GUID"

# Run analysis
./tile-diff \
  --old-tile pivotal-mysql-2.10.0-build.12.pivotal \
  --new-tile pivotal-mysql-2.11.0-build.8.pivotal \
  --product-guid "$MYSQL_GUID" \
  --ops-manager-url "$OM_TARGET" \
  --username "$OM_USERNAME" \
  --password "$OM_PASSWORD" \
  --skip-ssl-validation > mysql-upgrade-analysis.txt

cat mysql-upgrade-analysis.txt
```

**Sample Output:**
```
================================================================================
🚨 REQUIRED ACTIONS (1)
================================================================================

1. .properties.backup_options.enable.cron_schedule
   Type: string
   Status: Constraint changed (must match cron format)
   Current: "0 0 * * *"
   Action: Validate cron schedule format
   Recommendation: Current format valid, no change needed

================================================================================
ℹ️ INFORMATIONAL (2)
================================================================================

2. .properties.global_recipient_email
   Type: string
   Status: Default changed ("" → "admin@example.com")
   Current: ""
   Recommendation: Consider setting alert recipient email

3. .properties.plan_collection[0].backup_remote_delete_timeout
   Type: integer
   Status: New optional property
   Current: Not set
   Default: 300
   Recommendation: Keep default unless experiencing backup deletion timeouts
```

**Decision**:
- Required action is validation-only (current value already compliant)
- Set global recipient email for better alerting
- Proceed with upgrade

---

## Example 6: Configuration Audit

**Scenario**: Audit current configuration against tile schema to find deprecated settings.

### Audit Script

```bash
#!/bin/bash
# config-audit.sh - Audit current config against latest tile schema

set -e

CURRENT_TILE="$1"
CURRENT_VERSION="$(basename $CURRENT_TILE .pivotal | grep -oP '\d+\.\d+\.\d+')"

# Get current product GUID
PRODUCT_GUID=$(om curl -p /api/v0/staged/products | \
  jq -r '.[] | select(.type=="cf") | .guid')

echo "Configuration Audit for TAS $CURRENT_VERSION"
echo "=============================================="
echo ""

# Run tile-diff against same version to show current config
./tile-diff \
  --old-tile "$CURRENT_TILE" \
  --new-tile "$CURRENT_TILE" \
  --product-guid "$PRODUCT_GUID" \
  --ops-manager-url "$OM_TARGET" \
  --username "$OM_USERNAME" \
  --password "$OM_PASSWORD" \
  --skip-ssl-validation \
  --format json > audit.json

# Analyze current config
echo "Current Configuration Summary:"
echo "------------------------------"
om curl -p "/api/v0/staged/products/${PRODUCT_GUID}/properties" | \
  jq -r '.properties | to_entries[] |
    select(.value.configurable == true and .value.value != null) |
    .key' | wc -l | xargs echo "Configured properties:"

echo ""
echo "Analyzing for potential issues..."
echo ""

# Check for properties with deprecated in name/description
om curl -p "/api/v0/staged/products/${PRODUCT_GUID}/properties" | \
  jq -r '.properties | to_entries[] |
    select(.key | contains("deprecated") or contains("legacy")) |
    "⚠️  Using deprecated property: " + .key'

echo ""
echo "Audit complete. Review properties marked as deprecated."
```

---

## Example 7: Upgrade Runbook Generation

**Scenario**: Generate step-by-step upgrade runbook from tile-diff output.

### Runbook Generator

```bash
#!/bin/bash
# generate-runbook.sh - Create upgrade runbook from tile-diff analysis

set -e

OLD_VERSION="$1"
NEW_VERSION="$2"
OLD_TILE="$3"
NEW_TILE="$4"

if [ $# -ne 4 ]; then
  echo "Usage: $0 <old-version> <new-version> <old-tile> <new-tile>"
  exit 1
fi

RUNBOOK="upgrade-runbook-${OLD_VERSION}-to-${NEW_VERSION}.md"

# Run analysis
./tile-diff \
  --old-tile "$OLD_TILE" \
  --new-tile "$NEW_TILE" \
  --product-guid "$(om curl -p /api/v0/staged/products | jq -r '.[] | select(.type=="cf") | .guid')" \
  --ops-manager-url "$OM_TARGET" \
  --username "$OM_USERNAME" \
  --password "$OM_PASSWORD" \
  --skip-ssl-validation > /tmp/analysis.txt

# Generate runbook
cat > "$RUNBOOK" <<EOF
# TAS Upgrade Runbook: ${OLD_VERSION} → ${NEW_VERSION}

**Generated**: $(date)
**Status**: Draft - Requires Review

## Pre-Upgrade Checklist

### Required Configuration Changes

EOF

# Extract required actions
grep -A 100 "🚨 REQUIRED ACTIONS" /tmp/analysis.txt | \
  grep -E "^[0-9]+\." | \
  sed 's/^/- [ ] /' >> "$RUNBOOK"

cat >> "$RUNBOOK" <<EOF

### Warnings to Review

EOF

# Extract warnings
grep -A 100 "⚠️ WARNINGS" /tmp/analysis.txt | \
  grep -E "^[0-9]+\." | \
  sed 's/^/- [ ] Review: /' >> "$RUNBOOK"

cat >> "$RUNBOOK" <<EOF

## Upgrade Steps

### Phase 1: Preparation (Dev Environment)

1. [ ] Backup current configuration
   \`\`\`bash
   om staged-config -p cf > config-backup-${OLD_VERSION}.yml
   \`\`\`

2. [ ] Download new tile
   \`\`\`bash
   pivnet download-product-files \
     --product-slug elastic-runtime \
     --release-version ${NEW_VERSION} \
     --glob "*.pivotal"
   \`\`\`

3. [ ] Address all required configuration changes (see above)

4. [ ] Upload new tile to Ops Manager Dev

5. [ ] Stage new tile version

6. [ ] Apply configuration changes in Ops Manager

7. [ ] Review Changes in Ops Manager

8. [ ] Apply Changes (trigger deployment)

9. [ ] Validation:
   - [ ] CF API responding
   - [ ] Apps running normally
   - [ ] Logs flowing
   - [ ] Metrics collecting

### Phase 2: Staging Environment

Repeat Phase 1 steps in staging environment.

**Additional Validations:**
- [ ] Performance testing
- [ ] Integration testing
- [ ] Security scanning

### Phase 3: Production Environment

**Pre-Production:**
- [ ] Review all test results from dev/staging
- [ ] Confirm backup strategy
- [ ] Schedule maintenance window
- [ ] Notify stakeholders

**Production Upgrade:**
- [ ] Repeat Phase 1 steps in production
- [ ] Monitor closely during deployment
- [ ] Validate all services post-upgrade

## Rollback Plan

If upgrade fails:

1. Revert to ${OLD_VERSION} tile
2. Restore configuration from backup
3. Apply changes
4. Validate system recovery

## Post-Upgrade

- [ ] Monitor system for 24 hours
- [ ] Review logs for errors
- [ ] Update documentation
- [ ] Communicate completion to stakeholders

## Appendix: Full Analysis

See \`analysis-${OLD_VERSION}-to-${NEW_VERSION}.txt\` for complete details.

---

**Approved By**: _______________
**Date**: _______________
EOF

echo "Runbook generated: $RUNBOOK"
echo ""
echo "Next steps:"
echo "1. Review and customize runbook"
echo "2. Get approval from team lead"
echo "3. Schedule upgrade window"
echo "4. Execute runbook in dev environment first"
```

### Usage

```bash
./generate-runbook.sh \
  6.0.22 \
  10.2.5 \
  srt-6.0.22-build.2.pivotal \
  srt-10.2.5-build.2.pivotal
```

**Output**: `upgrade-runbook-6.0.22-to-10.2.5.md` ready for team review and execution.

---

## Tips for Effective Use

1. **Always test in non-prod first**: Use dev environment for initial analysis
2. **Version control everything**: Commit analysis files and runbooks to git
3. **Automate where possible**: Use scripts and CI/CD for consistency
4. **Cross-reference release notes**: Tile-diff shows what, release notes explain why
5. **Document decisions**: Record why you chose specific configurations
6. **Share knowledge**: Use runbooks to standardize team procedures

## Need Help?

- **User Guide**: [USER_GUIDE.md](USER_GUIDE.md) - Comprehensive documentation
- **Quick Start**: [QUICKSTART.md](QUICKSTART.md) - Get started in 5 minutes
- **Issues**: https://github.com/malston/tile-diff/issues - Report bugs or request features
