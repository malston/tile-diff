# Tile-Diff Quick Start Guide

Get up and running with tile-diff in 5 minutes.

## Installation

```bash
git clone https://github.com/malston/tile-diff.git
cd tile-diff
make build
```

Verify installation:

```bash
./tile-diff --help
```

## Quick Start Scenarios

### Scenario 1: Basic Tile Comparison (2 minutes)

**What you need:**

- Two `.pivotal` files (old and new versions)

**Command:**

```bash
./tile-diff \
  --old-tile /path/to/srt-6.0.22-build.2.pivotal \
  --new-tile /path/to/srt-10.2.5-build.2.pivotal
```

**What you get:**

- List of new properties
- List of removed properties
- List of changed properties
- Summary counts

**Example output:**

```
✨ New Properties (4):
  + tanzu_cf_cli_enable_auto_upgrades (boolean)
  + push_apps_manager_license_expiration_warning (boolean)
  + license_key (tanzu_license_key)
  + router_enable_log_attempt_details (boolean)

🗑️  Removed Properties (6):
  - push_apps_manager_offline_tools (multi_select_options)
  - nats_enabled_endpoints (selector)
  ...

Summary:
  Properties in old tile: 274
  Properties in new tile: 272
  Added: 4, Removed: 6, Changed: 0
```

**Use this when:**

- Initial upgrade scope assessment
- You don't have access to Ops Manager yet
- Comparing tiles for documentation purposes

---

### Scenario 2: Full Analysis with Your Current Config (5 minutes)

**What you need:**

- Two `.pivotal` files
- Ops Manager URL
- Ops Manager credentials
- Product GUID

**Step 1: Get your product GUID**

```bash
# Using om CLI
om curl -p /api/v0/staged/products | jq -r '.[] | select(.type=="cf") | .guid'

# Output example: cf-abc123def456
```

**Step 2: Run full analysis**

```bash
./tile-diff \
  --old-tile /path/to/srt-6.0.22-build.2.pivotal \
  --new-tile /path/to/srt-10.2.5-build.2.pivotal \
  --product-guid cf-abc123def456 \
  --ops-manager-url https://opsman.example.com \
  --username admin \
  --password your-password \
  --skip-ssl-validation
```

**What you get:**

- **🚨 Required Actions**: Must be done before upgrade
- **⚠️ Warnings**: Should be reviewed
- **ℹ️ Informational**: Nice to know
- Specific recommendations for each change
- Current values from your deployment

**Example output:**

```
================================================================================
🚨 REQUIRED ACTIONS (2)
================================================================================

1. .properties.new_security_feature
   Type: boolean
   Status: New required property (no default)
   Current: Not set
   Action: Must configure this property before upgrading
   Recommendation: Set to 'true' for enhanced security

2. .properties.memory_limit
   Type: integer
   Status: Constraint changed (min: 512 → 1024)
   Current: 768
   Action: Current value violates new minimum constraint
   Recommendation: Update to at least 1024 MB
```

**Use this when:**

- Planning actual upgrade execution
- Creating pre-upgrade checklists
- Validating upgrade readiness

---

### Scenario 3: JSON Output for Automation (3 minutes)

**Command:**

```bash
./tile-diff \
  --old-tile /path/to/old.pivotal \
  --new-tile /path/to/new.pivotal \
  --format json > analysis.json
```

**Parse results:**

```bash
# Check for required actions
jq '.summary.required_actions' analysis.json

# List all new properties
jq '.required_actions[] | select(.status | contains("New"))' analysis.json

# Count warnings
jq '.summary.warnings' analysis.json
```

**Use this when:**

- Integrating with CI/CD pipelines
- Automating upgrade readiness checks
- Building custom dashboards or reports

---

## Common Commands

### Compare tiles with env variables for credentials

```bash
export OM_TARGET="https://opsman.example.com"
export OM_USERNAME="admin"
export OM_PASSWORD="your-password"

./tile-diff \
  --old-tile old.pivotal \
  --new-tile new.pivotal \
  --ops-manager-url "$OM_TARGET" \
  --username "$OM_USERNAME" \
  --password "$OM_PASSWORD" \
  --skip-ssl-validation
```

### Save output to file

```bash
./tile-diff \
  --old-tile old.pivotal \
  --new-tile new.pivotal > upgrade-analysis.txt
```

### Check specific product (non-cf)

```bash
# First, find the product GUID
om curl -p /api/v0/staged/products | jq -r '.[] | select(.type=="p-mysql")'

# Then run analysis
./tile-diff \
  --old-tile p-mysql-2.10.0.pivotal \
  --new-tile p-mysql-2.11.0.pivotal \
  --product-guid p-mysql-xyz789 \
  --ops-manager-url "$OM_TARGET" \
  --username "$OM_USERNAME" \
  --password "$OM_PASSWORD"
```

---

## Understanding the Output

### Property Change Icons

| Icon | Meaning |
|------|---------|
| ✨ | New property added |
| 🗑️ | Property removed |
| 🔄 | Property changed (type, constraints, etc.) |
| 🚨 | Required action needed |
| ⚠️ | Warning - review recommended |
| ℹ️ | Informational - optional |

### Categories Explained

**🚨 Required Actions**

- New properties that have no default and are not optional
- Properties where your current value violates new constraints
- Properties that changed from optional to required

**Action:** Must address before upgrade

**⚠️ Warnings**

- Properties you're using that are removed in new version
- Properties where type changed
- Properties with significant constraint changes

**Action:** Review and plan for changes

**ℹ️ Informational**

- New optional properties with defaults
- Removed properties you never configured
- Default value changes

**Action:** Review for potential improvements to your config

---

## Quick Troubleshooting

### "Error: --old-tile and --new-tile are required"

**Fix:** Provide both tile paths:

```bash
./tile-diff --old-tile path1.pivotal --new-tile path2.pivotal
```

### "Error loading old tile: failed to extract"

**Fix:** Verify file is a valid `.pivotal` file:

```bash
file old.pivotal  # Should show: Zip archive data
unzip -t old.pivotal  # Test archive integrity
```

### "Error: failed to connect to Ops Manager"

**Fixes:**

1. Check URL is correct and accessible: `curl -k https://your-opsman-url`
2. Verify credentials
3. Add `--skip-ssl-validation` for self-signed certs
4. Check product GUID exists: `om curl -p /api/v0/staged/products`

### "No output or shows identical tiles"

**Possible causes:**

1. Comparing same version twice
2. Only system-managed properties changed (not shown in output)

**Verify:** Check the property counts in output - should be different if versions differ

---

## Next Steps

- **Detailed guide**: See [USER_GUIDE.md](USER_GUIDE.md) for comprehensive documentation
- **Real examples**: See [EXAMPLES.md](EXAMPLES.md) for real-world upgrade scenarios
- **Integration**: Add tile-diff to your upgrade workflow
- **Feedback**: Report issues at <https://github.com/malston/tile-diff/issues>

## Pro Tips

1. **Always test in non-prod first**: Run against dev/staging Ops Manager before production
2. **Save your analysis**: Version control the output for historical reference
3. **Use JSON for automation**: Parse JSON output in scripts and CI/CD
4. **Secure credentials**: Use environment variables, never hardcode passwords
5. **Cross-reference release notes**: Tile-diff shows what changed, release notes explain why

---

## Cheat Sheet

```bash
# Basic comparison
./tile-diff --old-tile old.pivotal --new-tile new.pivotal

# With current config
./tile-diff \
  --old-tile old.pivotal \
  --new-tile new.pivotal \
  --product-guid $(om curl -p /api/v0/staged/products | jq -r '.[] | select(.type=="cf") | .guid') \
  --ops-manager-url "$OM_TARGET" \
  --username "$OM_USERNAME" \
  --password "$OM_PASSWORD" \
  --skip-ssl-validation

# JSON output
./tile-diff --old-tile old.pivotal --new-tile new.pivotal --format json

# Check for required actions (exit 1 if any found)
./tile-diff ... --format json | jq -e '.summary.required_actions == 0'
```
