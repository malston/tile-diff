# Pivnet Integration Manual Test Procedure

## Prerequisites
- Valid Pivnet API token
- Internet connection
- At least 25GB free disk space

## Test 1: Interactive Mode - Complete Flow

```bash
export PIVNET_TOKEN="your-token-here"

./tile-diff \
  --product-slug cf \
  --old-version 6.0 \
  --new-version 10.2

# Expected:
# - Prompts for version selection (multiple 6.0.x versions)
# - Prompts for product file selection (TAS vs Small Footprint)
# - Prompts for EULA acceptance
# - Downloads both tiles with progress bars
# - Completes comparison
```

## Test 2: Non-Interactive Mode with Exact Versions

```bash
./tile-diff \
  --product-slug cf \
  --old-version '6.0.22+LTS-T' \
  --new-version '10.2.5+LTS-T' \
  --product-file "TAS for VMs" \
  --pivnet-token "$PIVNET_TOKEN" \
  --accept-eula \
  --non-interactive

# Expected:
# - No prompts
# - Downloads both tiles
# - Completes comparison
```

## Test 3: Cache Verification

```bash
# Run test 2 again
./tile-diff \
  --product-slug cf \
  --old-version '6.0.22+LTS-T' \
  --new-version '10.2.5+LTS-T' \
  --product-file "TAS for VMs" \
  --pivnet-token "$PIVNET_TOKEN" \
  --accept-eula \
  --non-interactive

# Expected:
# - Both tiles loaded from cache (instant)
# - No downloads
```

## Test 4: EULA Persistence

```bash
# First run - should prompt or require --accept-eula
./tile-diff \
  --product-slug p-redis \
  --old-version 3.2.0 \
  --new-version 3.2.1 \
  --pivnet-token "$PIVNET_TOKEN"

# Second run - should NOT prompt (EULA remembered)
./tile-diff \
  --product-slug p-redis \
  --old-version 3.2.1 \
  --new-version 3.2.2 \
  --pivnet-token "$PIVNET_TOKEN"
```

## Test 5: Error Handling - Invalid Token

```bash
./tile-diff \
  --product-slug cf \
  --old-version 6.0.22 \
  --new-version 10.2.5 \
  --pivnet-token "invalid-token" \
  --non-interactive

# Expected:
# - Clear error about invalid token
```

## Test 6: Error Handling - Ambiguous Version

```bash
./tile-diff \
  --product-slug cf \
  --old-version 6.0 \
  --new-version 10.2 \
  --pivnet-token "$PIVNET_TOKEN" \
  --non-interactive

# Expected:
# - Error listing matching versions
# - Suggestion to use exact version
```

## Test 7: Local Files Mode Still Works

```bash
./tile-diff \
  --old-tile /path/to/old.pivotal \
  --new-tile /path/to/new.pivotal

# Expected:
# - Works as before (no Pivnet interaction)
```

## Verification Checklist

- [ ] Interactive version selection works
- [ ] Interactive product file selection works
- [ ] Interactive EULA acceptance works
- [ ] Non-interactive mode works with all flags
- [ ] Cache stores and retrieves files correctly
- [ ] EULA acceptance is persisted
- [ ] Download progress bars display correctly
- [ ] Disk space check prevents downloads when full
- [ ] Cache cleanup removes old files
- [ ] Error messages are clear and actionable
- [ ] Local files mode still works
- [ ] Mixed mode (local + pivnet) is rejected
