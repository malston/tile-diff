#!/bin/bash
# ABOUTME: Common helper functions for acceptance tests.
# ABOUTME: Provides test utilities, assertions, and cleanup functions.

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Test cache directory (separate from production)
TEST_CACHE_DIR="${TEST_CACHE_DIR:-/tmp/tile-diff-test-cache}"
TEST_EULA_FILE="${TEST_EULA_FILE:-/tmp/tile-diff-test-eulas.json}"
TEST_MANIFEST_FILE="${TEST_MANIFEST_FILE:-${TEST_CACHE_DIR}/manifest.json}"

# Binary location
TILE_DIFF_BIN="${TILE_DIFF_BIN:-./tile-diff}"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $*"
}

# Test assertion functions
assert_success() {
    local exit_code=$1
    local test_name=$2

    TESTS_RUN=$((TESTS_RUN + 1))

    if [ "$exit_code" -eq 0 ]; then
        log_success "✓ $test_name"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        log_error "✗ $test_name (exit code: $exit_code)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

assert_failure() {
    local exit_code=$1
    local test_name=$2

    TESTS_RUN=$((TESTS_RUN + 1))

    if [ "$exit_code" -ne 0 ]; then
        log_success "✓ $test_name (failed as expected)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        log_error "✗ $test_name (should have failed but succeeded)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

assert_contains() {
    local haystack=$1
    local needle=$2
    local test_name=$3

    TESTS_RUN=$((TESTS_RUN + 1))

    if echo "$haystack" | grep -q "$needle"; then
        log_success "✓ $test_name"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        log_error "✗ $test_name"
        log_error "  Expected to find: '$needle'"
        log_error "  In output: '$haystack'"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

assert_file_exists() {
    local file=$1
    local test_name=$2

    TESTS_RUN=$((TESTS_RUN + 1))

    if [ -f "$file" ]; then
        log_success "✓ $test_name"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        log_error "✗ $test_name (file not found: $file)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

assert_file_not_exists() {
    local file=$1
    local test_name=$2

    TESTS_RUN=$((TESTS_RUN + 1))

    if [ ! -f "$file" ]; then
        log_success "✓ $test_name"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        log_error "✗ $test_name (file should not exist: $file)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Cleanup functions
cleanup_test_cache() {
    log_info "Cleaning up test cache..."
    rm -rf "$TEST_CACHE_DIR"
    rm -f "$TEST_EULA_FILE"
}

setup_test_env() {
    log_info "Setting up test environment..."
    cleanup_test_cache
    mkdir -p "$TEST_CACHE_DIR"

    # Verify binary exists
    if [ ! -f "$TILE_DIFF_BIN" ]; then
        log_error "tile-diff binary not found at: $TILE_DIFF_BIN"
        log_error "Run 'make build' first"
        exit 1
    fi

    # Verify PIVNET_TOKEN is set
    if [ -z "${PIVNET_TOKEN:-}" ]; then
        log_warning "PIVNET_TOKEN not set - some tests will be skipped"
    fi
}

# Summary functions
print_test_summary() {
    echo ""
    echo "=================================================="
    echo "Test Summary"
    echo "=================================================="
    echo "Total:  $TESTS_RUN"
    echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
    echo -e "${RED}Failed: $TESTS_FAILED${NC}"
    echo "=================================================="

    if [ "$TESTS_FAILED" -eq 0 ]; then
        echo -e "${GREEN}All tests passed!${NC}"
        return 0
    else
        echo -e "${RED}Some tests failed!${NC}"
        return 1
    fi
}

# Wait for file to exist (for download tests)
wait_for_file() {
    local file=$1
    local timeout=${2:-30}
    local elapsed=0

    while [ ! -f "$file" ] && [ $elapsed -lt $timeout ]; do
        sleep 1
        elapsed=$((elapsed + 1))
    done

    [ -f "$file" ]
}

# Get file size in MB
get_file_size_mb() {
    local file=$1
    local size_bytes
    size_bytes=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null)
    echo $((size_bytes / 1024 / 1024))
}

# Check if cache has entries
cache_has_entries() {
    [ -f "$TEST_MANIFEST_FILE" ] && [ "$(cat "$TEST_MANIFEST_FILE" | grep -c '"file_path"')" -gt 0 ]
}
