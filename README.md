# tile-diff

A command-line tool to compare Ops Manager product tile configurations between versions, identifying configuration changes required during upgrades.

## Overview

When upgrading product tiles (e.g., TAS from 6.0.22 to 10.2.5), operators need to understand:
- What new configuration properties must be set
- What existing properties are no longer supported
- What properties have changed constraints or defaults
- Whether current configuration values remain valid

This tool automates that analysis by comparing tile metadata and cross-referencing with your current Ops Manager configuration.

## Features

- **Automatic Tile Downloads**: Download tiles directly from Pivotal Network without manual steps
- **Smart Caching**: Reuse downloaded tiles across comparisons
- **Interactive Selection**: Fuzzy version matching with user-friendly prompts
- **CI-Friendly**: Non-interactive mode for automated workflows
- **Smart Property Comparison**: Automatically detects new, removed, and changed properties between tile versions
- **Current Config Analysis**: Cross-references changes with your actual Ops Manager configuration
- **Intelligent Categorization**: Classifies changes as Required Actions, Warnings, or Informational
- **Actionable Recommendations**: Provides specific guidance for each configuration change
- **Multiple Output Formats**: Human-readable text reports and machine-readable JSON
- **Constraint Validation**: Checks if current values meet new requirements

## Documentation

- **[Quick Start Guide](docs/QUICKSTART.md)** - Get up and running in 5 minutes
- **[User Guide](docs/USER_GUIDE.md)** - Comprehensive usage documentation
- **[Examples](docs/EXAMPLES.md)** - Real-world scenarios and use cases

## Quick Start

### Build

```bash
make build
```

### Usage

#### Download Tiles from Pivnet (Recommended)

**Interactive Mode:**
```bash
export PIVNET_TOKEN="your-pivnet-api-token"

./tile-diff \
  --product-slug cf \
  --old-version 6.0 \
  --new-version 10.2.5
```

tile-diff will:
- Resolve version strings (prompts if multiple matches)
- Show available product files (e.g., TAS vs Small Footprint)
- Handle EULA acceptance (one-time per product)
- Cache downloads for reuse

**Non-Interactive Mode (CI/Scripts):**
```bash
./tile-diff \
  --product-slug cf \
  --old-version '6.0.22+LTS-T' \
  --new-version '10.2.5+LTS-T' \
  --product-file "TAS for VMs" \
  --pivnet-token "$PIVNET_TOKEN" \
  --accept-eula \
  --non-interactive
```

#### Use Local Files

If you've already downloaded tiles:
```bash
./tile-diff \
  --old-tile srt-6.0.22.pivotal \
  --new-tile srt-10.2.5.pivotal
```

#### Compare with Current Ops Manager Config

```bash
./tile-diff \
  --product-slug cf \
  --old-version 6.0.22 \
  --new-version 10.2.5 \
  --product-guid cf-xxxxx \
  --ops-manager-url https://opsman.tas.vcf.lab \
  --username admin \
  --password your-password \
  --skip-ssl-validation
```

#### JSON Output

```bash
./tile-diff \
  --product-slug cf \
  --old-version 6.0.22 \
  --new-version 10.2.5 \
  --format json
```

### Getting a Pivnet API Token

1. Go to the Broadcom Support Portal: https://support.broadcom.com/
2. Sign in with your Broadcom account
3. Navigate to your API token settings
4. Copy your "UAA API Token" (64+ characters, not the legacy 20-char token)
5. Use it via flag or environment variable:

```bash
# Via environment variable (recommended)
export PIVNET_TOKEN="your-token-here"
./tile-diff --product-slug cf --old-version 6.0 --new-version 10.2

# Via flag
./tile-diff --pivnet-token "your-token-here" --product-slug cf --old-version 6.0 --new-version 10.2
```

### EULA Acceptance

**Important:** For legal reasons, EULAs must be accepted through the Broadcom Support Portal web interface. The API does not support programmatic EULA acceptance for regular users (only for Broadcom/VMware employees).

**First-time download workflow:**

1. Run tile-diff command
2. If EULA not accepted, you'll be directed to the web URL
3. Accept the EULA in your browser
4. Press Enter to continue the download

**Interactive mode:**
```bash
./tile-diff --product-slug cf --old-version 6.0 --new-version 10.2
# Will pause and show EULA URL if needed
```

**Non-interactive mode:**
```bash
# For CI/CD: Accept EULAs through web first, then use --accept-eula to acknowledge
./tile-diff \
  --product-slug cf \
  --old-version 6.0.22 \
  --new-version 10.2.5 \
  --accept-eula \
  --non-interactive
```

Once accepted for a product, the acceptance is remembered locally and you won't be prompted again.

### Example Output

```
================================================================================
                    Ops Manager Tile Upgrade Analysis
================================================================================

Old Version: srt-6.0.22
New Version: srt-10.2.5

Loading old tile...
  Found 274 properties (184 configurable)
Loading new tile...
  Found 272 properties (182 configurable)

Querying Ops Manager API...
  Found 599 total properties
  Currently configured: 156

Analyzing changes...

Total Changes: 12
  Required Actions: 2
  Warnings: 4
  Informational: 6

================================================================================
🚨 REQUIRED ACTIONS
================================================================================

1. .properties.new_security_setting
   Type: boolean
   Status: New required property (no default)
   Current: Not set
   Action: Must configure this property before upgrading
   Recommendation: Set to 'true' for enhanced security

2. .properties.authentication_method
   Type: selector
   Status: Type changed from string to selector
   Current: "basic"
   Action: Update to use new selector format
   Recommendation: Choose 'oauth' option for modern authentication

================================================================================
⚠️  WARNINGS
================================================================================

3. .properties.deprecated_timeout
   Type: integer
   Status: Removed in new version
   Current: 30
   Action: Property will be ignored after upgrade
   Recommendation: Review if this setting impacts your deployment

4. .properties.memory_limit
   Type: integer
   Status: Constraints changed (min: 512 → 1024)
   Current: 768
   Action: Current value 768 is below new minimum 1024
   Recommendation: Update to at least 1024 MB

================================================================================
ℹ️  INFORMATIONAL
================================================================================

5. .properties.new_optional_feature
   Type: boolean
   Status: New optional property
   Current: Not set
   Default: false
   Recommendation: Enable for improved logging capabilities

6. .properties.enhanced_monitoring
   Type: string
   Status: New optional property
   Current: Not set
   Default: "basic"
   Recommendation: Consider 'advanced' for production environments

================================================================================
Summary: 2 required actions must be completed before upgrade
================================================================================
```

## Development

### Requirements

- Go 1.21+
- `om` CLI (Ops Manager CLI)
- Access to product tile `.pivotal` files
- Access to Ops Manager API (for current config comparison)

### Setup

```bash
# Clone the repository
git clone https://github.com/malston/tile-diff.git
cd tile-diff

# Install dependencies
go mod download

# Build
make build

# Run tests
make test
```

### Project Structure

```
tile-diff/
├── cmd/
│   └── tile-diff/        # CLI entry point
├── pkg/
│   ├── metadata/         # Tile metadata extraction
│   ├── api/              # Ops Manager API client
│   └── compare/          # Comparison logic
├── docs/                 # Documentation
└── Makefile              # Build tasks
```

## License

MIT
