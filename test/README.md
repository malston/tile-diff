# Test Documentation

This directory contains the test suite for tile-diff, organized into unit tests, integration tests, and acceptance tests.

## Test Organization

```sh
test/
├── acceptance_suite_test.go      # Ginkgo test suite bootstrap
├── acceptance_test.go            # Pivnet integration acceptance tests
├── acceptance_helpers_test.go    # Shared helper functions for tests
├── comparison_test.go            # Unit tests for comparison logic
├── enrichment_test.go            # Unit tests for enrichment logic
├── integration_test.go           # Integration tests with real tile files
└── fixtures/                     # Test fixtures and sample data
```

## Test Types

### Unit Tests

Unit tests validate individual packages and components in isolation:

- **Location**: `./pkg/*/...` (within each package)
- **Framework**: Standard Go testing
- **Coverage**: Metadata extraction, comparison logic, report generation
- **Run with**: `make test` or `go test -v ./pkg/...`
- **Dependencies**: None - uses mock data and fixtures

**Example:**

```bash
# Run all unit tests with coverage
make test

# Run tests for specific package
go test -v ./pkg/compare/
go test -v ./pkg/metadata/
```

### Integration Tests

Integration tests verify functionality using real tile files:

- **Location**: `test/integration_test.go`, `test/comparison_test.go`
- **Framework**: Standard Go testing with build tags
- **Coverage**: Real tile metadata extraction, full comparison workflows
- **Run with**: `go test -v -tags=integration ./test/...`
- **Dependencies**: Requires actual `.pivotal` tile files

**Example:**

```bash
# Run integration tests (requires tile files in /tmp/elastic-runtime/)
go test -v -tags=integration ./test/...
```

### Acceptance Tests (Ginkgo)

Modern acceptance tests using Ginkgo v2 BDD framework:

- **Location**: `test/acceptance_test.go`
- **Framework**: Ginkgo v2 + Gomega
- **Coverage**: End-to-end Pivnet integration, caching, EULA handling, error scenarios
- **Run with**: `make acceptance-test` or `ginkgo -v ./test`
- **Dependencies**: Requires `PIVNET_TOKEN` environment variable

**Test Suites:**

1. **Non-Interactive Mode** - Downloads, version matching, output formats
2. **Cache Verification** - Tile caching and reuse behavior
3. **EULA Handling** - EULA acceptance and persistence
4. **Error Handling** - Invalid tokens, ambiguous products, mixed modes
5. **Local Files Mode** - Backward compatibility with local tiles

**Example:**

```bash
# Set Pivnet token
export PIVNET_TOKEN="your-pivnet-api-token"

# Run all acceptance tests
make acceptance-test

# Or run fast tests only (skips slow downloads)
make acceptance-test-fast-with-token PIVNET_TOKEN=your-token

# Run specific test suite
ginkgo -v --focus="Cache Verification" ./test

# Run with debugging output
ginkgo -v -trace ./test

# Skip slow tests using labels
ginkgo -v --label-filter='!slow' ./test
```

## Test Configuration

### Environment Variables

- `PIVNET_TOKEN` - Required for acceptance tests that download from Pivnet
- `ENABLE_DOWNLOAD_TESTS` - Set to `1` to enable expensive download tests (⚠️ downloads multi-GB files)
- `TILE_DIFF_BIN` - Path to tile-diff binary (default: `./tile-diff`)
- `TEST_CACHE_DIR` - Cache directory for test runs (default: `/tmp/tile-diff-test-cache`)

**⚠️ Download Tests Warning:**
Download tests are **SKIPPED by default** to prevent:

- ISP bandwidth quota consumption
- CI/CD pipeline slowdowns
- Expensive Pivnet API usage

Only enable when specifically testing download functionality:

```bash
export ENABLE_DOWNLOAD_TESTS=1
make acceptance-test
```

### Test Helpers

The `acceptance_helpers_test.go` file provides shared utilities:

- `runTileDiff(args...)` - Execute tile-diff binary with test configuration
- `setupCacheDir()` - Create clean test cache directory
- `cleanupCacheDir()` - Remove test cache directory
- `tileExistsInCache(filename)` - Check if tile is cached

## Running Tests

### Quick Start

```bash
# 1. Build the binary
make build

# 2. Run unit tests
make test

# 3. Run acceptance tests (with Pivnet token)
export PIVNET_TOKEN="your-token"
make acceptance-test

# 4. Run all tests
make test-all
```

### Continuous Integration

For CI environments, use non-interactive mode and pre-accepted EULAs:

```bash
# Set token from secrets
export PIVNET_TOKEN="${CI_PIVNET_TOKEN}"

# Run unit tests (always)
make test

# Run acceptance tests (if token available)
if [ -n "$PIVNET_TOKEN" ]; then
  make acceptance-test
fi
```

### Test Fixtures

The `fixtures/` directory contains sample data for tests:

- Sample tile metadata
- Mock API responses
- Example configuration files

## Test Patterns

### Ginkgo Test Structure

```go
var _ = Describe("Feature Name", func() {
    BeforeEach(func() {
        // Setup for each test
        setupCacheDir()
    })

    AfterEach(func() {
        // Cleanup after each test
        cleanupCacheDir()
    })

    Context("Specific Scenario", func() {
        It("describes expected behavior", func() {
            output, err := runTileDiff(args...)
            Expect(err).NotTo(HaveOccurred())
            Expect(output).To(ContainSubstring("expected"))
        })
    })
})
```

### Handling Test Data Changes

Acceptance tests use live Pivnet data and may require updates:

```go
// NOTE: These tests use p-redis 3.2.0 -> 3.2.1 as examples.
// If these specific versions are no longer available in Pivnet, the tests
// will fail. This is expected behavior - update the product-slug and
// versions as needed to match available releases.
```

If tests fail due to missing product versions:

1. Check Pivnet for available versions
2. Update test to use available product/versions
3. Update --product-file flag if product has multiple files
4. Document the change in test comments

### Skipping Tests

Tests are automatically skipped when dependencies are missing:

```go
if os.Getenv("PIVNET_TOKEN") == "" {
    Skip("PIVNET_TOKEN not set - skipping live Pivnet tests")
}
```

## Test Coverage

Generate coverage reports:

```bash
# Generate coverage for unit tests
make test-coverage

# Open HTML coverage report
open coverage.html

# View coverage in terminal
go tool cover -func=coverage.txt
```

## Debugging Tests

### Verbose Output

```bash
# Ginkgo verbose mode
ginkgo -v ./test

# With trace for debugging
ginkgo -v -trace ./test

# Standard go test verbose
go test -v ./pkg/...
```

### Running Single Tests

```bash
# Ginkgo focus on specific test
ginkgo --focus="downloads tiles with exact versions" ./test

# Standard go test
go test -v -run TestMetadataExtraction ./pkg/metadata/
```

### Preserving Test Cache

By default, tests clean up their cache. To preserve for debugging:

```bash
# Comment out cleanupCacheDir() in AfterEach
# Or manually inspect: ls -la /tmp/tile-diff-test-cache/
```

## Best Practices

1. **Unit tests should be fast** - Use mocks and fixtures, no external dependencies
2. **Integration tests verify real behavior** - Use actual tile files when available
3. **Acceptance tests validate user workflows** - Test complete scenarios end-to-end
4. **Clean up test artifacts** - Use BeforeEach/AfterEach for setup/teardown
5. **Handle missing dependencies gracefully** - Skip tests rather than fail
6. **Document test data assumptions** - Note when tests rely on specific Pivnet versions
7. **Use descriptive test names** - Focus on behavior, not implementation

## Troubleshooting

### "PIVNET_TOKEN not set"

Acceptance tests require a Pivnet token:

```bash
export PIVNET_TOKEN="your-pivnet-api-token"
```

### "tile-diff binary not found"

Build the binary first:

```bash
make build
```

### "no releases found matching version"

Test product/version no longer available in Pivnet - update test to use available versions.

### Integration test failures

Integration tests need actual tile files:

```bash
# Download tiles manually or use tile-diff to cache them
./tile-diff --product-slug cf --old-version 6.0.22 --new-version 10.2.5
```

## Contributing

When adding new tests:

1. **Unit tests** - Add to relevant package in `pkg/*/`
2. **Ginkgo acceptance tests** - Add to `test/acceptance_test.go`
3. **Integration tests** - Add to `test/integration_test.go` with `// +build integration` tag
4. **Test helpers** - Add shared utilities to `test/acceptance_helpers_test.go`
5. **Update documentation** - Keep this README current with test changes

Run the full test suite before submitting:

```bash
make test-all
```
