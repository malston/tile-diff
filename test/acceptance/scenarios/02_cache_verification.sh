#!/bin/bash
# ABOUTME: Test cache functionality.
# ABOUTME: Verifies downloads are cached and reused on subsequent runs.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../helpers/common.sh"

test_cache_stores_downloads() {
    log_info "Test: Cache stores downloads"

    if [ -z "${PIVNET_TOKEN:-}" ]; then
        log_warning "PIVNET_TOKEN not set - skipping test"
        return 0
    fi

    # First download - should populate cache
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

    # Verify cache manifest exists
    assert_file_exists "$TEST_MANIFEST_FILE" \
        "Cache manifest file created"

    # Verify cache has entries
    if cache_has_entries; then
        log_success "✓ Cache has entries"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        log_error "✗ Cache should have entries"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    # Verify actual tile files exist
    local file_count
    file_count=$(find "$TEST_CACHE_DIR" -name "*.pivotal" | wc -l | tr -d ' ')
    if [ "$file_count" -ge 2 ]; then
        log_success "✓ Cache contains tile files ($file_count files)"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        log_error "✗ Cache should contain at least 2 tile files (found $file_count)"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
}

test_cache_reuses_downloads() {
    log_info "Test: Cache reuses downloads on subsequent runs"

    if [ -z "${PIVNET_TOKEN:-}" ]; then
        log_warning "PIVNET_TOKEN not set - skipping test"
        return 0
    fi

    # Record cache state before second run
    local cache_files_before
    cache_files_before=$(find "$TEST_CACHE_DIR" -name "*.pivotal" -type f)

    # Second run - should use cache
    local start_time
    local end_time
    start_time=$(date +%s)

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

    end_time=$(date +%s)
    local duration=$((end_time - start_time))

    # Should complete quickly (cache hit)
    if [ "$duration" -lt 10 ]; then
        log_success "✓ Second run completed quickly (${duration}s - cache hit)"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        log_warning "⚠ Second run took ${duration}s (might be cache miss or slow system)"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))  # Don't fail on timing
    fi

    # Should mention using cached files
    assert_contains "$output" "Using cached file" \
        "Output mentions using cached files"

    # Cache files should not have changed
    local cache_files_after
    cache_files_after=$(find "$TEST_CACHE_DIR" -name "*.pivotal" -type f)

    if [ "$cache_files_before" = "$cache_files_after" ]; then
        log_success "✓ No new downloads (cache reused)"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        log_error "✗ Cache files changed (expected cache reuse)"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
}

# Run tests
setup_test_env

log_info "Running cache verification tests..."
test_cache_stores_downloads
test_cache_reuses_downloads

print_test_summary
