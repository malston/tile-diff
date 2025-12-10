# tile-diff

A command-line tool to compare Tanzu Application Service (TAS) tile configurations between versions, identifying configuration changes required during upgrades.

## Overview

When upgrading TAS tiles (e.g., from 6.0.22 to 10.2.5), operators need to understand:
- What new configuration properties must be set
- What existing properties are no longer supported
- What properties have changed constraints or defaults
- Whether current configuration values remain valid

This tool automates that analysis by comparing tile metadata and cross-referencing with your current Ops Manager configuration.

## Status

✅ **Phase 1 MVP - Complete**

Core extraction and parsing functionality implemented.

✅ **Phase 2 - Complete**

Property comparison logic implemented:
- Identify new properties in target version
- Identify removed properties
- Detect type and optionality changes
- Display categorized comparison results

🚧 **Phase 3 - In Planning**

Next: Add current config cross-reference and generate actionable reports.

## Documentation

- [Implementation Specification](docs/2025-12-10-tile-diff-implementation-spec.md)
- [Phase 1 Implementation Plan](docs/plans/2025-12-10-phase-1-mvp.md)

## Quick Start

### Build

```bash
make build
```

### Run Phase 1 MVP

Compare two tile versions (metadata only):

```bash
./tile-diff --old-tile srt-6.0.22.pivotal --new-tile srt-10.2.5.pivotal
```

Include current Ops Manager configuration:

```bash
./tile-diff \
  --old-tile srt-6.0.22.pivotal \
  --new-tile srt-10.2.5.pivotal \
  --product-guid cf-85da7fd88e99806e5d08 \
  --ops-manager-url https://opsman.tas.vcf.lab \
  --username admin \
  --password your-password \
  --skip-ssl-validation
```

### Example Output

```
tile-diff Phase 1 MVP
=====================

Loading old tile: srt-6.0.22.pivotal
  Found 450 properties
Loading new tile: srt-10.2.5.pivotal
  Found 520 properties

Configurable properties:
  Old tile: 422
  New tile: 485

Querying Ops Manager API...
  Found 599 total properties
  Configurable: 422
  Currently configured: ~156

==================================================
Phase 1 MVP: Complete ✓
==================================================

Data sources validated:
  ✓ Old tile metadata extraction
  ✓ New tile metadata extraction
  ✓ Ops Manager API current configuration

Next phase: Implement comparison logic
```

## Development

### Requirements

- Go 1.21+
- `om` CLI (Ops Manager CLI)
- Access to TAS tile `.pivotal` files
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
