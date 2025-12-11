#!/bin/bash
# ABOUTME: Quick smoke test to validate basic functionality.
# ABOUTME: Runs without PIVNET_TOKEN to test CLI validation and error handling.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/helpers/common.sh"

echo "=================================================="
echo "  tile-diff Smoke Test"
echo "=================================================="
echo ""

test_binary_exists() {
    log_info "Test: Binary exists and is executable"

    if [ -f "$TILE_DIFF_BIN" ] && [ -x "$TILE_DIFF_BIN" ]; then
        log_success "✓ Binary exists at $TILE_DIFF_BIN"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        log_error "✗ Binary not found or not executable: $TILE_DIFF_BIN"
        log_error "  Run 'make build' first"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
        exit 1
    fi
}

test_help_flag() {
    log_info "Test: --help flag works"

    local output
    output=$("$TILE_DIFF_BIN" --help 2>&1) || true

    # Help is allowed to exit 0 or non-zero, we just care about output
    assert_contains "$output" "Usage\|tile-diff" \
        "Help output shows usage information"
}

test_version_validation() {
    log_info "Test: Version validation (missing required flags)"

    local output
    local exit_code=0
    output=$("$TILE_DIFF_BIN" --product-slug cf 2>&1) || exit_code=$?

    assert_failure "$exit_code" "Missing flags causes failure"
    assert_contains "$output" "required\|version" \
        "Error mentions missing required flags"
}

test_mode_conflict_detection() {
    log_info "Test: Mode conflict detection"

    local output
    local exit_code=0
    output=$("$TILE_DIFF_BIN" \
        --old-tile /tmp/old.pivotal \
        --product-slug cf \
        2>&1) || exit_code=$?

    assert_failure "$exit_code" "Mode conflict is detected"
    assert_contains "$output" "Cannot mix\|local.*pivnet" \
        "Error explains mode conflict"
}

test_local_mode_missing_files() {
    log_info "Test: Local mode with missing files"

    local output
    local exit_code=0
    output=$("$TILE_DIFF_BIN" \
        --old-tile /tmp/nonexistent-$$.pivotal \
        --new-tile /tmp/also-nonexistent-$$.pivotal \
        2>&1) || exit_code=$?

    assert_failure "$exit_code" "Missing files cause failure"
    assert_contains "$output" "not found\|no such file\|cannot" \
        "Error mentions missing file"
}

test_invalid_format() {
    log_info "Test: Invalid format flag"

    local output
    local exit_code=0
    output=$("$TILE_DIFF_BIN" \
        --old-tile /tmp/old.pivotal \
        --new-tile /tmp/new.pivotal \
        --format invalid-format \
        2>&1) || exit_code=$?

    assert_failure "$exit_code" "Invalid format causes failure"
}

# Run tests
log_info "Running smoke tests (no PIVNET_TOKEN required)..."
echo ""

test_binary_exists
test_help_flag
test_version_validation
test_mode_conflict_detection
test_local_mode_missing_files
test_invalid_format

echo ""
print_test_summary
