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
- **Auto-Detect Product GUID**: Automatically finds product GUID from Ops Manager using product slug
- **Formatted Reports Always**: Get professional upgrade analysis reports even without Ops Manager credentials
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

**Option 1: Auto-detect product GUID (Recommended)**

```bash
./tile-diff \
  --product-slug harbor-container-registry \
  --old-version 2.11.0 \
  --new-version 2.13.2 \
  --ops-manager-url https://opsman.example.com \
  --username admin \
  --password your-password \
  --skip-ssl-validation
```

The tool will automatically query Ops Manager to find the product GUID based on the product slug.

**Option 2: Explicit product GUID**

```bash
./tile-diff \
  --old-tile srt-6.0.22.pivotal \
  --new-tile srt-10.2.5.pivotal \
  --product-guid cf-abc123xyz \
  --ops-manager-url https://opsman.example.com \
  --username admin \
  --password your-password \
  --skip-ssl-validation
```

**Option 3: Formatted report without credentials**

```bash
./tile-diff \
  --old-tile harbor-2.11.0.pivotal \
  --new-tile harbor-2.13.2.pivotal
```

Even without Ops Manager credentials, you'll get a professional formatted upgrade analysis report showing all potential changes.

#### JSON Output

```bash
./tile-diff \
  --product-slug cf \
  --old-version 6.0.22 \
  --new-version 10.2.5 \
  --format json
```

### Getting a Pivnet API Token

1. Go to the Broadcom Support Portal: <https://support.broadcom.com/>
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
tile-diff - Ops Manager Product Tile Comparison
================================================

Mode: Pivnet Download
Product: cf
Versions: 6.0.22 -> 10.2.5

Resolving and downloading old tile (6.0.22)...
✓ Old tile: /Users/user/.tile-diff/cache/srt-6.0.22.pivotal

Resolving and downloading new tile (10.2.5)...
✓ New tile: /Users/user/.tile-diff/cache/srt-10.2.5.pivotal

Loading old tile: /Users/user/.tile-diff/cache/srt-6.0.22.pivotal
  Found 274 properties
Loading new tile: /Users/user/.tile-diff/cache/srt-10.2.5.pivotal
  Found 272 properties

Comparing tiles...

Comparison Results:
===================

✨ New Properties (8):
  + .properties.new_security_setting (boolean)
  + .properties.enhanced_monitoring (string)
  ...

🗑️  Removed Properties (4):
  - .properties.deprecated_timeout (integer)
  - .properties.old_feature (string)
  ...

🔄 Changed Properties (2):
  ~ .properties.authentication_method: Type changed from string to selector
  ~ .properties.memory_limit: Constraints changed

Summary:
  Properties in old tile: 274
  Properties in new tile: 272
  Added: 8, Removed: 4, Changed: 2

Configurable properties:
  Old tile: 184
  New tile: 182

Querying Ops Manager API...
  Found 599 total properties
  Configurable: 184
  Currently configured: ~156

Generating actionable report...

================================================================================
                  Ops Manager Tile Upgrade Analysis
================================================================================

Old Version: /Users/user/.tile-diff/cache/srt-6.0.22.pivotal
New Version: /Users/user/.tile-diff/cache/srt-10.2.5.pivotal

Total Changes: 12
  Required Actions: 2
  Warnings: 4
  Informational: 6

================================================================================
🚨 REQUIRED ACTIONS
================================================================================

These changes MUST be addressed before upgrading:

1. .properties.new_security_setting
   Type: boolean
   Action: Must configure this property before upgrading

2. .properties.authentication_method
   Type: selector
   Action: Must configure this property before upgrading

================================================================================
⚠️  WARNINGS
================================================================================

These changes should be reviewed:

3. .properties.deprecated_timeout
   Change: Property removed in new version
   Recommendation: Property will be ignored after upgrade - review and remove from config

4. .properties.memory_limit
   Change: Constraints changed
   Recommendation: Review this change and verify compatibility

================================================================================
ℹ️  INFORMATIONAL
================================================================================

New optional features available:

5. .properties.new_optional_feature
   Type: boolean
   Default: false
   Note: Optional - review for potential improvements

6. .properties.enhanced_monitoring
   Type: string
   Default: basic
   Note: Optional - review for potential improvements

```

## Development

### Requirements

- Go 1.21+
- `om` CLI (Ops Manager CLI)
- Access to product tile `.pivotal` files
- Access to Ops Manager API (for current config comparison)
- Ginkgo v2 (for running acceptance tests)

### Setup

```bash
# Clone the repository
git clone https://github.com/malston/tile-diff.git
cd tile-diff

# Install dependencies
go mod download

# Install Ginkgo (if not already installed)
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Build
make build

# Run unit tests
make test

# Run acceptance tests (requires PIVNET_TOKEN)
export PIVNET_TOKEN="your-pivnet-api-token"
make acceptance-test

# Run all tests (unit + acceptance)
make test-all
```

### Running Tests

The project includes comprehensive test coverage at multiple levels:

#### Unit Tests

Unit tests cover individual packages and run without external dependencies:

```bash
# Run all unit tests with coverage
make test

# Generate HTML coverage report
make test-coverage

# Run tests for a specific package
go test -v ./pkg/compare/...
```

#### Acceptance Tests

Acceptance tests use Ginkgo v2 and verify end-to-end functionality against live Pivnet:

```bash
# Set your Pivnet API token
export PIVNET_TOKEN="your-pivnet-api-token"

# Run Ginkgo acceptance tests
make acceptance-test

# Or run Ginkgo directly with verbose output
ginkgo -v ./test

# Run specific test suites
ginkgo -v --focus="Cache Verification" ./test
ginkgo -v --focus="EULA Handling" ./test
```

**Note:** Acceptance tests require a valid `PIVNET_TOKEN` environment variable. Tests will be skipped if the token is not set.

#### Integration Tests

Integration tests use real tile files and are tagged separately:

```bash
# Run integration tests (requires actual tile files)
go test -v -tags=integration ./test/...
```

See [test/README.md](test/README.md) for detailed documentation on test structure and organization.

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
