#!/bin/bash
# ABOUTME: Test local files mode (backward compatibility).
# ABOUTME: Verifies that local tile file comparison still works as before.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../helpers/common.sh"

setup_local_test_tiles() {
    # We'll use the cache from previous tests or download fresh tiles
    # to have real .pivotal files to test with
    if [ -z "${PIVNET_TOKEN:-}" ]; then
        log_warning "PIVNET_TOKEN not set - skipping local files test"
        return 1
    fi

    log_info "Setting up test tiles for local files mode..."

    # Download two small tiles if not in cache
    "$TILE_DIFF_BIN" \
        --product-slug p-healthwatch \
        --old-version '2.4.7' \
        --new-version '2.4.8' \
        --product-file "VMware Tanzu® Healthwatch" \
        --pivnet-token "$PIVNET_TOKEN" \
        --accept-eula \
        --non-interactive \
        --cache-dir "$TEST_CACHE_DIR" \
        >/dev/null 2>&1 || true

    # Find the downloaded tiles
    OLD_TILE=$(find "$TEST_CACHE_DIR" -name "*2.4.7*.pivotal" | head -1)
    NEW_TILE=$(find "$TEST_CACHE_DIR" -name "*2.4.8*.pivotal" | head -1)

    if [ -z "$OLD_TILE" ] || [ -z "$NEW_TILE" ]; then
        log_error "Could not find test tiles in cache"
        return 1
    fi

    log_info "Using tiles:"
    log_info "  Old: $OLD_TILE"
    log_info "  New: $NEW_TILE"

    return 0
}

test_local_files_basic() {
    log_info "Test: Local files mode basic comparison"

    if ! setup_local_test_tiles; then
        return 0
    fi

    local output
    output=$("$TILE_DIFF_BIN" \
        --old-tile "$OLD_TILE" \
        --new-tile "$NEW_TILE" \
        2>&1) || true

    local exit_code=$?

    # Should succeed
    assert_success "$exit_code" "Local files mode completes successfully"

    # Should show comparison results
    assert_contains "$output" "Tile Upgrade Analysis\|Total Changes" \
        "Output shows comparison results"

    # Should NOT mention Pivnet
    if echo "$output" | grep -qi "pivnet\|downloading\|cache"; then
        log_error "✗ Local files mode should not mention Pivnet/downloading"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    else
        log_success "✓ No Pivnet references in local files mode"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    fi
}

test_local_files_with_json() {
    log_info "Test: Local files mode with JSON output"

    if ! setup_local_test_tiles; then
        return 0
    fi

    local output
    output=$("$TILE_DIFF_BIN" \
        --old-tile "$OLD_TILE" \
        --new-tile "$NEW_TILE" \
        --format json \
        2>&1) || true

    local exit_code=$?

    # Should succeed
    assert_success "$exit_code" "Local files JSON mode completes successfully"

    # Should output valid JSON
    if echo "$output" | jq . >/dev/null 2>&1; then
        log_success "✓ Output is valid JSON"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        log_error "✗ Output is not valid JSON"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
}

test_local_files_missing() {
    log_info "Test: Local files mode with missing file"

    local output
    output=$("$TILE_DIFF_BIN" \
        --old-tile /tmp/nonexistent-tile.pivotal \
        --new-tile /tmp/another-nonexistent.pivotal \
        2>&1) || true

    local exit_code=$?

    # Should fail
    assert_failure "$exit_code" "Missing local file causes failure"

    # Should have error about missing file
    assert_contains "$output" "not found\|no such file\|cannot open" \
        "Error message mentions missing file"
}

# Run tests
setup_test_env

log_info "Running local files mode tests..."
test_local_files_basic
test_local_files_with_json
test_local_files_missing

print_test_summary
