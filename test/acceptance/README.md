# Pivnet Integration Acceptance Tests

Automated acceptance tests for the Pivnet integration feature.

## Prerequisites

- Built `tile-diff` binary (run `make build`)
- Valid Pivnet API token (set `PIVNET_TOKEN` environment variable)
- Internet connection
- `jq` installed (for JSON validation tests)

## Quick Start

```bash
# Export your Pivnet token
export PIVNET_TOKEN="your-token-here"

# Run all acceptance tests
./test/acceptance/run_acceptance_tests.sh
```

## Test Scenarios

### 01. Non-Interactive Mode
Tests non-interactive downloads with exact versions and JSON output.

**What it tests:**
- Non-interactive mode completes successfully
- Downloads or uses cache appropriately
- JSON output is valid

**Run individually:**
```bash
./test/acceptance/scenarios/01_non_interactive_mode.sh
```

### 02. Cache Verification
Tests that downloads are cached and reused on subsequent runs.

**What it tests:**
- Cache stores downloaded tiles
- Cache manifest is created
- Subsequent runs reuse cached files
- Cache hits complete quickly

**Run individually:**
```bash
./test/acceptance/scenarios/02_cache_verification.sh
```

### 03. Error Handling
Tests various error scenarios and validates error messages.

**What it tests:**
- Invalid token produces proper error
- Ambiguous versions fail in non-interactive mode
- Missing --product-file flag fails when needed
- Missing EULA acceptance fails in non-interactive mode
- Mixed mode (local + pivnet) is rejected
- Missing required flags produce helpful errors

**Run individually:**
```bash
./test/acceptance/scenarios/03_error_handling.sh
```

### 04. Local Files Mode
Tests backward compatibility with local tile files.

**What it tests:**
- Local files mode still works
- JSON output works with local files
- Missing files produce proper errors
- No Pivnet references in output

**Run individually:**
```bash
./test/acceptance/scenarios/04_local_files_mode.sh
```

### 05. EULA Persistence
Tests that EULA acceptance is remembered across runs.

**What it tests:**
- EULA acceptance creates persistence file
- Subsequent runs don't require --accept-eula
- Different products require new EULA acceptance

**Run individually:**
```bash
./test/acceptance/scenarios/05_eula_persistence.sh
```

## Running Tests

### Run All Tests
```bash
export PIVNET_TOKEN="your-token-here"
./test/acceptance/run_acceptance_tests.sh
```

### Run with Verbose Output
```bash
./test/acceptance/run_acceptance_tests.sh --verbose
```

### Run Specific Scenario
```bash
./test/acceptance/run_acceptance_tests.sh 01_non_interactive_mode
```

### Run Without Pivnet Token (Limited Tests)
```bash
# Only error handling and some basic tests will run
./test/acceptance/run_acceptance_tests.sh
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PIVNET_TOKEN` | Pivotal Network API token | _(required for most tests)_ |
| `TEST_CACHE_DIR` | Test cache directory | `/tmp/tile-diff-test-cache` |
| `TEST_EULA_FILE` | Test EULA file | `/tmp/tile-diff-test-eulas.json` |
| `TILE_DIFF_BIN` | Path to tile-diff binary | `./tile-diff` |

## Test Output

Tests use color-coded output:
- 🔵 **BLUE**: Informational messages
- 🟢 **GREEN**: Success messages
- 🔴 **RED**: Error messages
- 🟡 **YELLOW**: Warning messages

Each test reports:
- Test name
- Pass/Fail status
- Details on failures

Summary includes:
- Total tests run
- Tests passed
- Tests failed

## Cleanup

The test suite uses separate cache and EULA files to avoid interfering with your production environment.

To clean up test artifacts:
```bash
rm -rf /tmp/tile-diff-test-cache
rm -f /tmp/tile-diff-test-eulas.json
```

Test files will be automatically cleaned up between test runs.

## CI/CD Integration

The test suite is designed for CI/CD pipelines:

```bash
#!/bin/bash
set -e

# Build
make build

# Run tests
export PIVNET_TOKEN="${PIVNET_TOKEN}"
./test/acceptance/run_acceptance_tests.sh

# Exit code indicates success/failure
```

### GitHub Actions Example
```yaml
- name: Run Acceptance Tests
  env:
    PIVNET_TOKEN: ${{ secrets.PIVNET_TOKEN }}
  run: |
    make build
    ./test/acceptance/run_acceptance_tests.sh
```

## Test Data

Tests use small products to minimize download time:
- `p-healthwatch` (versions 2.4.7 and 2.4.8)
- Files are typically 50-200MB each

First run will download tiles, subsequent runs use cache and complete quickly.

## Troubleshooting

### Tests Skipped
If you see "PIVNET_TOKEN not set - skipping test", export your token:
```bash
export PIVNET_TOKEN="your-token-here"
```

### Network Errors
Some tests require internet access. Ensure you can reach `network.tanzu.vmware.com`.

### Cache Issues
If cache tests fail, manually clean the cache:
```bash
rm -rf /tmp/tile-diff-test-cache
```

### EULA Issues
If EULA tests fail, check the EULA file:
```bash
cat ~/.tile-diff/eulas.json
```

## Adding New Tests

1. Create a new scenario file in `scenarios/`:
   ```bash
   touch test/acceptance/scenarios/06_your_test.sh
   chmod +x test/acceptance/scenarios/06_your_test.sh
   ```

2. Use the common helper:
   ```bash
   source "$SCRIPT_DIR/../helpers/common.sh"
   ```

3. Follow the pattern:
   ```bash
   test_your_feature() {
       log_info "Test: Your feature description"

       # Test implementation
       local output
       output=$("$TILE_DIFF_BIN" --your-flags 2>&1) || true
       local exit_code=$?

       # Assertions
       assert_success "$exit_code" "Your test name"
       assert_contains "$output" "expected text" "Verification message"
   }

   setup_test_env
   test_your_feature
   print_test_summary
   ```

4. Test runs automatically with `run_acceptance_tests.sh`

## Helper Functions

The test suite provides helper functions in `helpers/common.sh`:

### Assertions
- `assert_success <exit_code> <test_name>`
- `assert_failure <exit_code> <test_name>`
- `assert_contains <haystack> <needle> <test_name>`
- `assert_file_exists <file> <test_name>`
- `assert_file_not_exists <file> <test_name>`

### Logging
- `log_info <message>`
- `log_success <message>`
- `log_error <message>`
- `log_warning <message>`

### Utilities
- `setup_test_env` - Initialize test environment
- `cleanup_test_cache` - Clean up test cache
- `print_test_summary` - Print test results
- `wait_for_file <file> [timeout]` - Wait for file to exist
- `get_file_size_mb <file>` - Get file size in MB
- `cache_has_entries` - Check if cache has entries
