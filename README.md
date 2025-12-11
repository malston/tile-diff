# tile-diff

A command-line tool to compare Ops Manager product tile configurations between versions, identifying configuration changes required during upgrades.

## Overview

When upgrading product tiles (e.g., TAS from 6.0.22 to 10.2.5), operators need to understand:
- What new configuration properties must be set
- What existing properties are no longer supported
- What properties have changed constraints or defaults
- Whether current configuration values remain valid

This tool automates that analysis by comparing tile metadata and cross-referencing with your current Ops Manager configuration.

## Status

✅ **Phase 1 Complete** - Extraction & parsing

✅ **Phase 2 Complete** - Property comparison

✅ **Phase 3 Complete** - Actionable reports

Full upgrade analysis with:
- Current config cross-reference
- Change categorization (Required/Warning/Info)
- Formatted reports (text and JSON)
- Specific recommendations per change

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

#### Compare Tiles with Actionable Report

```bash
./tile-diff \
  --old-tile srt-6.0.22.pivotal \
  --new-tile srt-10.2.5.pivotal \
  --product-guid cf-xxxxx \
  --ops-manager-url https://opsman.tas.vcf.lab \
  --username admin \
  --password your-password \
  --skip-ssl-validation
```

#### JSON Output

```bash
./tile-diff \
  --old-tile srt-6.0.22.pivotal \
  --new-tile srt-10.2.5.pivotal \
  --format json
```

#### Basic Comparison (without current config)

```bash
./tile-diff --old-tile srt-6.0.22.pivotal --new-tile srt-10.2.5.pivotal
```

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
