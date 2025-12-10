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

🚧 **Phase 1 MVP - In Development**

Currently implementing core extraction and parsing functionality.

## Documentation

- [Implementation Specification](docs/2025-12-10-tile-diff-implementation-spec.md)
- [Phase 1 Implementation Plan](docs/phase-1-implementation-plan.md) (coming soon)

## Quick Start

*(Coming soon)*

```bash
# Compare two tile versions
tile-diff compare \
  --old-tile srt-6.0.22.pivotal \
  --new-tile srt-10.2.5.pivotal

# Include current Ops Manager configuration
tile-diff compare \
  --old-tile srt-6.0.22.pivotal \
  --new-tile srt-10.2.5.pivotal \
  --product-guid cf-xxxxx
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
