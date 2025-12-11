# Acceptance Tests Quick Start

## Run Tests in 2 Steps

### 1. Set your Pivnet token
```bash
export PIVNET_TOKEN="your-pivnet-api-token"
```

Get your token from: https://network.tanzu.vmware.com/users/dashboard/edit-profile

### 2. Run the tests
```bash
make test-acceptance
```

The binary will be built automatically if needed.

## Expected Output

```
==================================================
  tile-diff Pivnet Integration Acceptance Tests
==================================================

[INFO] Environment:
[INFO]   Binary: ./tile-diff
[INFO]   Cache Dir: /tmp/tile-diff-test-cache
[INFO]   EULA File: /tmp/tile-diff-test-eulas.json
[INFO]   Pivnet Token: ✓ Set

[INFO] Found 5 test scenario(s)

========================================
Running: 01_non_interactive_mode
========================================
[SUCCESS] ✓ Non-interactive mode completes successfully
[SUCCESS] ✓ Output shows download or cache usage
[SUCCESS] ✓ Output shows comparison results
[SUCCESS] ✓ JSON output mode completes successfully
[SUCCESS] ✓ Output is valid JSON

========================================
Running: 02_cache_verification
========================================
[SUCCESS] ✓ Cache manifest file created
[SUCCESS] ✓ Cache has entries
[SUCCESS] ✓ Cache contains tile files (2 files)
[SUCCESS] ✓ Second run completed quickly (3s - cache hit)
[SUCCESS] ✓ Output mentions using cached files
[SUCCESS] ✓ No new downloads (cache reused)

========================================
Running: 03_error_handling
========================================
[SUCCESS] ✓ Invalid token causes failure
[SUCCESS] ✓ Error message mentions authentication failure
[SUCCESS] ✓ Ambiguous version in non-interactive mode fails
[SUCCESS] ✓ Error message suggests using exact version
[SUCCESS] ✓ Missing product-file flag fails in non-interactive mode
[SUCCESS] ✓ Error message mentions --product-file flag
[SUCCESS] ✓ Missing EULA acceptance fails in non-interactive mode
[SUCCESS] ✓ Error message mentions EULA
[SUCCESS] ✓ Mixed mode flags are rejected
[SUCCESS] ✓ Error message mentions mode conflict
[SUCCESS] ✓ Missing required flags causes failure
[SUCCESS] ✓ Error message mentions missing required flags

========================================
Running: 04_local_files_mode
========================================
[SUCCESS] ✓ Local files mode completes successfully
[SUCCESS] ✓ Output shows comparison results
[SUCCESS] ✓ No Pivnet references in local files mode
[SUCCESS] ✓ Local files JSON mode completes successfully
[SUCCESS] ✓ Output is valid JSON
[SUCCESS] ✓ Missing local file causes failure
[SUCCESS] ✓ Error message mentions missing file

========================================
Running: 05_eula_persistence
========================================
[SUCCESS] ✓ EULA acceptance file created
[SUCCESS] ✓ EULA file contains product acceptance
[SUCCESS] ✓ Second run succeeds without --accept-eula (EULA remembered)
[SUCCESS] ✓ Second run uses cached files
[SUCCESS] ✓ New product without EULA acceptance fails in non-interactive mode
[SUCCESS] ✓ Error message mentions EULA requirement

==================================================
Final Summary
==================================================
Total Scenarios: 5
Passed: 5
Failed: 0

[SUCCESS] All scenarios passed!
==================================================
```

## What Gets Tested

✅ **Non-Interactive Mode** - Automatic downloads with exact versions
✅ **Cache System** - Downloads are cached and reused
✅ **Error Handling** - Clear error messages for invalid inputs
✅ **Local Files** - Backward compatibility with local tiles
✅ **EULA Persistence** - EULA acceptance is remembered

## Verbose Mode

To see detailed output from each test:
```bash
make test-acceptance-verbose
```

## Run Individual Scenario

```bash
# Non-interactive mode tests
./test/acceptance/scenarios/01_non_interactive_mode.sh

# Cache tests
./test/acceptance/scenarios/02_cache_verification.sh

# Error handling tests
./test/acceptance/scenarios/03_error_handling.sh

# Local files tests
./test/acceptance/scenarios/04_local_files_mode.sh

# EULA tests
./test/acceptance/scenarios/05_eula_persistence.sh
```

## Cleanup

Test artifacts are stored in temporary directories:
```bash
rm -rf /tmp/tile-diff-test-cache
rm -f /tmp/tile-diff-test-eulas.json
```

Your production cache (`~/.tile-diff/`) is not affected.

## Troubleshooting

### "PIVNET_TOKEN not set - skipping test"
Export your Pivnet token:
```bash
export PIVNET_TOKEN="your-token-here"
```

### "tile-diff binary not found"
This shouldn't happen when using `make test-acceptance` (builds automatically).

If running test scripts directly:
```bash
make build
./test/acceptance/scenarios/01_non_interactive_mode.sh
```

### Tests are slow
First run downloads tiles (~100-200MB each). Subsequent runs use cache and complete in seconds.

## CI/CD

For automated testing in CI/CD:
```bash
#!/bin/bash
set -e

export PIVNET_TOKEN="${CI_PIVNET_TOKEN}"
make test-acceptance  # Builds automatically
```

## Next Steps

- See [README.md](README.md) for detailed documentation
- Add new test scenarios in `scenarios/`
- Use helper functions from `helpers/common.sh`
