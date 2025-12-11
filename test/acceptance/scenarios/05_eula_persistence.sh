#!/bin/bash
# ABOUTME: Test EULA acceptance persistence.
# ABOUTME: Verifies EULA is accepted once and remembered for future downloads.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../helpers/common.sh"

test_eula_persistence() {
    log_info "Test: EULA acceptance is persisted"

    if [ -z "${PIVNET_TOKEN:-}" ]; then
        log_warning "PIVNET_TOKEN not set - skipping test"
        return 0
    fi

    # Use a unique product to ensure clean EULA state
    local product="p-healthwatch"

    # First run with --accept-eula
    log_info "First run: accepting EULA..."
    "$TILE_DIFF_BIN" \
        --product-slug "$product" \
        --old-version '2.4.7' \
        --new-version '2.4.8' \
        --product-file "VMware Tanzu® Healthwatch" \
        --pivnet-token "$PIVNET_TOKEN" \
        --accept-eula \
        --non-interactive \
        --cache-dir "$TEST_CACHE_DIR" \
        >/dev/null 2>&1 || true

    # EULA file should exist
    local eula_file="$HOME/.tile-diff/eulas.json"
    assert_file_exists "$eula_file" \
        "EULA acceptance file created"

    # EULA file should contain the product
    if grep -q "$product" "$eula_file"; then
        log_success "✓ EULA file contains product acceptance"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        log_error "✗ EULA file should contain product acceptance"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    # Second run WITHOUT --accept-eula (should work because EULA is remembered)
    log_info "Second run: EULA should be remembered..."
    local output
    output=$("$TILE_DIFF_BIN" \
        --product-slug "$product" \
        --old-version '2.4.7' \
        --new-version '2.4.8' \
        --product-file "VMware Tanzu® Healthwatch" \
        --pivnet-token "$PIVNET_TOKEN" \
        --non-interactive \
        --cache-dir "$TEST_CACHE_DIR" \
        2>&1) || true

    local exit_code=$?

    # Should succeed even without --accept-eula
    assert_success "$exit_code" \
        "Second run succeeds without --accept-eula (EULA remembered)"

    # Should use cached files (not re-download)
    assert_contains "$output" "Using cached file" \
        "Second run uses cached files"
}

test_different_product_requires_new_eula() {
    log_info "Test: Different product requires new EULA acceptance"

    if [ -z "${PIVNET_TOKEN:-}" ]; then
        log_warning "PIVNET_TOKEN not set - skipping test"
        return 0
    fi

    # Use a different product than previous test
    # We'll try with a product that hasn't been accepted yet
    local product="p-spring-cloud-services"

    # Clear EULA for this product if it exists
    local eula_file="$HOME/.tile-diff/eulas.json"
    if [ -f "$eula_file" ]; then
        # Remove this product from EULA file
        local temp_file="/tmp/eula-temp-$$.json"
        jq "del(.products[\"$product\"])" "$eula_file" > "$temp_file" 2>/dev/null || echo '{"products":{}}' > "$temp_file"
        mv "$temp_file" "$eula_file"
    fi

    # Try without --accept-eula (should fail)
    local output
    output=$("$TILE_DIFF_BIN" \
        --product-slug "$product" \
        --old-version '4.0.8' \
        --new-version '4.0.9' \
        --pivnet-token "$PIVNET_TOKEN" \
        --non-interactive \
        --cache-dir "$TEST_CACHE_DIR" \
        2>&1) || true

    local exit_code=$?

    # Should fail
    assert_failure "$exit_code" \
        "New product without EULA acceptance fails in non-interactive mode"

    # Should mention EULA
    assert_contains "$output" "EULA\|--accept-eula" \
        "Error message mentions EULA requirement"
}

# Run tests
setup_test_env

log_info "Running EULA persistence tests..."
test_eula_persistence
test_different_product_requires_new_eula

print_test_summary
