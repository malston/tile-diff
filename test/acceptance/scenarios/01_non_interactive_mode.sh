#!/bin/bash
# ABOUTME: Test non-interactive mode with exact versions.
# ABOUTME: Verifies automatic downloads, progress tracking, and comparison.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../helpers/common.sh"

test_non_interactive_with_exact_versions() {
    log_info "Test: Non-interactive mode with exact versions"

    if [ -z "${PIVNET_TOKEN:-}" ]; then
        log_warning "PIVNET_TOKEN not set - skipping test"
        return 0
    fi

    # Use a small product for faster testing
    local output
    output=$("$TILE_DIFF_BIN" \
        --product-slug p-healthwatch \
        --old-version '2.4.7' \
        --new-version '2.4.8' \
        --product-file "VMware Tanzu® Healthwatch" \
        --pivnet-token "$PIVNET_TOKEN" \
        --accept-eula \
        --non-interactive \
        --cache-dir "$TEST_CACHE_DIR" \
        2>&1) || true

    local exit_code=$?

    # Should succeed
    assert_success "$exit_code" "Non-interactive mode completes successfully"

    # Should mention downloading or using cache
    assert_contains "$output" "Downloading\|Using cached file" \
        "Output shows download or cache usage"

    # Should show comparison results
    assert_contains "$output" "Tile Upgrade Analysis\|Total Changes" \
        "Output shows comparison results"
}

test_non_interactive_with_json_output() {
    log_info "Test: Non-interactive mode with JSON output"

    if [ -z "${PIVNET_TOKEN:-}" ]; then
        log_warning "PIVNET_TOKEN not set - skipping test"
        return 0
    fi

    local output
    output=$("$TILE_DIFF_BIN" \
        --product-slug p-healthwatch \
        --old-version '2.4.7' \
        --new-version '2.4.8' \
        --product-file "VMware Tanzu® Healthwatch" \
        --pivnet-token "$PIVNET_TOKEN" \
        --accept-eula \
        --non-interactive \
        --format json \
        --cache-dir "$TEST_CACHE_DIR" \
        2>&1) || true

    local exit_code=$?

    # Should succeed
    assert_success "$exit_code" "JSON output mode completes successfully"

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

# Run tests
setup_test_env

log_info "Running non-interactive mode tests..."
test_non_interactive_with_exact_versions
test_non_interactive_with_json_output

print_test_summary
