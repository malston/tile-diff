#!/bin/bash
# ABOUTME: Test error handling scenarios.
# ABOUTME: Verifies proper error messages for invalid inputs and failures.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../helpers/common.sh"

test_invalid_token() {
    log_info "Test: Invalid Pivnet token"

    local output
    output=$("$TILE_DIFF_BIN" \
        --product-slug p-healthwatch \
        --old-version '2.4.7' \
        --new-version '2.4.8' \
        --pivnet-token "invalid-token-12345" \
        --accept-eula \
        --non-interactive \
        --cache-dir "$TEST_CACHE_DIR" \
        2>&1) || true

    local exit_code=$?

    # Should fail
    assert_failure "$exit_code" "Invalid token causes failure"

    # Should have meaningful error message
    assert_contains "$output" "401\|authentication\|invalid token\|unauthorized" \
        "Error message mentions authentication failure"
}

test_ambiguous_version_non_interactive() {
    log_info "Test: Ambiguous version in non-interactive mode"

    if [ -z "${PIVNET_TOKEN:-}" ]; then
        log_warning "PIVNET_TOKEN not set - skipping test"
        return 0
    fi

    local output
    output=$("$TILE_DIFF_BIN" \
        --product-slug cf \
        --old-version '6.0' \
        --new-version '10.2' \
        --pivnet-token "$PIVNET_TOKEN" \
        --accept-eula \
        --non-interactive \
        --cache-dir "$TEST_CACHE_DIR" \
        2>&1) || true

    local exit_code=$?

    # Should fail
    assert_failure "$exit_code" "Ambiguous version in non-interactive mode fails"

    # Should suggest using exact version
    assert_contains "$output" "multiple\|exact version" \
        "Error message suggests using exact version"
}

test_missing_product_file_non_interactive() {
    log_info "Test: Multiple product files without --product-file flag"

    if [ -z "${PIVNET_TOKEN:-}" ]; then
        log_warning "PIVNET_TOKEN not set - skipping test"
        return 0
    fi

    local output
    output=$("$TILE_DIFF_BIN" \
        --product-slug cf \
        --old-version '6.0.22+LTS-T' \
        --new-version '6.0.23+LTS-T' \
        --pivnet-token "$PIVNET_TOKEN" \
        --accept-eula \
        --non-interactive \
        --cache-dir "$TEST_CACHE_DIR" \
        2>&1) || true

    local exit_code=$?

    # Should fail
    assert_failure "$exit_code" "Missing product-file flag fails in non-interactive mode"

    # Should mention --product-file flag
    assert_contains "$output" "product-file\|--product-file" \
        "Error message mentions --product-file flag"
}

test_missing_eula_non_interactive() {
    log_info "Test: Missing EULA acceptance in non-interactive mode"

    if [ -z "${PIVNET_TOKEN:-}" ]; then
        log_warning "PIVNET_TOKEN not set - skipping test"
        return 0
    fi

    # Use a fresh cache to avoid cached EULA acceptance
    local temp_cache="/tmp/tile-diff-eula-test-$$"
    mkdir -p "$temp_cache"

    local output
    output=$("$TILE_DIFF_BIN" \
        --product-slug p-healthwatch \
        --old-version '2.4.7' \
        --new-version '2.4.8' \
        --product-file "VMware Tanzu® Healthwatch" \
        --pivnet-token "$PIVNET_TOKEN" \
        --non-interactive \
        --cache-dir "$temp_cache" \
        2>&1) || true

    local exit_code=$?
    rm -rf "$temp_cache"

    # Should fail
    assert_failure "$exit_code" "Missing EULA acceptance fails in non-interactive mode"

    # Should mention EULA
    assert_contains "$output" "EULA\|--accept-eula" \
        "Error message mentions EULA"
}

test_mixed_mode_rejection() {
    log_info "Test: Mixed mode (local files + pivnet flags) is rejected"

    local output
    output=$("$TILE_DIFF_BIN" \
        --old-tile /tmp/old.pivotal \
        --new-tile /tmp/new.pivotal \
        --product-slug cf \
        --old-version '6.0' \
        2>&1) || true

    local exit_code=$?

    # Should fail
    assert_failure "$exit_code" "Mixed mode flags are rejected"

    # Should mention mode conflict
    assert_contains "$output" "Cannot mix\|local.*pivnet\|--old-tile.*--product-slug" \
        "Error message mentions mode conflict"
}

test_missing_required_flags() {
    log_info "Test: Missing required flags for Pivnet mode"

    local output
    output=$("$TILE_DIFF_BIN" \
        --product-slug cf \
        --old-version '6.0' \
        2>&1) || true

    local exit_code=$?

    # Should fail
    assert_failure "$exit_code" "Missing required flags causes failure"

    # Should mention missing flags
    assert_contains "$output" "new-version\|required" \
        "Error message mentions missing required flags"
}

# Run tests
setup_test_env

log_info "Running error handling tests..."
test_invalid_token
test_ambiguous_version_non_interactive
test_missing_product_file_non_interactive
test_missing_eula_non_interactive
test_mixed_mode_rejection
test_missing_required_flags

print_test_summary
