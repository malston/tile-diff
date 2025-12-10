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

- [Implementation Specification](docs/2025-12-10-tile-diff-implementation-spec.md)

### Phase 1: Data Extraction & Parsing
- [Implementation Plan](docs/plans/2025-12-10-phase-1-mvp.md)
- [Completion Report](docs/phase-1-completion.md)

### Phase 2: Property Comparison Logic
- [Implementation Plan](docs/plans/2025-12-10-phase-2-comparison.md)
- [Completion Report](docs/phase-2-completion.md)

### Phase 3: Actionable Reports
- [Implementation Plan](docs/plans/2025-12-10-phase-3-reports.md)
- [Completion Report](docs/phase-3-completion.md)

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
tile-diff - Ops Manager Product Tile Comparison
================================================

Loading old tile: srt-6.0.22.pivotal
  Found 274 properties (184 configurable)
Loading new tile: srt-10.2.5.pivotal
  Found 272 properties (182 configurable)

Comparing tiles...

Comparison Results:
===================

✨ New Properties (15):
  + .properties.new_feature_flag (boolean)
  + .properties.enhanced_logging_level (string)
  + .properties.retry_configuration (integer)
  ...

🗑️  Removed Properties (8):
  - .properties.deprecated_setting (boolean)
  - .properties.legacy_timeout (integer)
  ...

🔄 Changed Properties (5):
  ~ .properties.memory_limit: Type changed from string to integer
  ~ .properties.optional_field: Now required (was optional)
  ...

Summary:
  Properties in old tile: 274 (184 configurable)
  Properties in new tile: 272 (182 configurable)
  Added: 15, Removed: 8, Changed: 5
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
