#!/bin/bash
# ABOUTME: Main acceptance test runner for Pivnet integration.
# ABOUTME: Orchestrates all test scenarios and provides comprehensive reporting.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/helpers/common.sh"

# Test configuration
RUN_ALL_TESTS=${RUN_ALL_TESTS:-true}
SKIP_SLOW_TESTS=${SKIP_SLOW_TESTS:-false}
VERBOSE=${VERBOSE:-false}

print_banner() {
    echo ""
    echo "=================================================="
    echo "  tile-diff Pivnet Integration Acceptance Tests"
    echo "=================================================="
    echo ""
}

print_environment() {
    log_info "Environment:"
    log_info "  Binary: $TILE_DIFF_BIN"
    log_info "  Cache Dir: $TEST_CACHE_DIR"
    log_info "  EULA File: $TEST_EULA_FILE"
    if [ -n "${PIVNET_TOKEN:-}" ]; then
        log_info "  Pivnet Token: ✓ Set"
    else
        log_warning "  Pivnet Token: ✗ Not set (some tests will be skipped)"
    fi
    echo ""
}

run_test_scenario() {
    local scenario=$1
    local scenario_name=$(basename "$scenario" .sh)

    echo ""
    log_info "========================================"
    log_info "Running: $scenario_name"
    log_info "========================================"

    if [ "$VERBOSE" = "true" ]; then
        bash "$scenario"
    else
        bash "$scenario" 2>&1 | grep -E "^\[|^=" || true
    fi

    local exit_code=$?

    if [ $exit_code -eq 0 ]; then
        log_success "Scenario passed: $scenario_name"
    else
        log_error "Scenario failed: $scenario_name"
        return 1
    fi

    return 0
}

main() {
    print_banner
    print_environment

    # Build binary if it doesn't exist
    if [ ! -f "$TILE_DIFF_BIN" ]; then
        log_info "Building tile-diff binary..."
        cd "$SCRIPT_DIR/../.." && make build
    fi

    local scenarios_dir="$SCRIPT_DIR/scenarios"
    local scenarios=()
    local failed_scenarios=()

    # Collect test scenarios
    if [ "$RUN_ALL_TESTS" = "true" ]; then
        while IFS= read -r scenario; do
            scenarios+=("$scenario")
        done < <(find "$scenarios_dir" -name "*.sh" | sort)
    else
        # Run specific scenarios
        for scenario in "$@"; do
            if [ -f "$scenarios_dir/$scenario.sh" ]; then
                scenarios+=("$scenarios_dir/$scenario.sh")
            else
                log_error "Scenario not found: $scenario"
            fi
        done
    fi

    if [ ${#scenarios[@]} -eq 0 ]; then
        log_error "No test scenarios found"
        exit 1
    fi

    log_info "Found ${#scenarios[@]} test scenario(s)"

    # Run scenarios
    local scenario_count=0
    for scenario in "${scenarios[@]}"; do
        scenario_count=$((scenario_count + 1))

        if ! run_test_scenario "$scenario"; then
            failed_scenarios+=("$(basename "$scenario" .sh)")
        fi
    done

    # Print final summary
    echo ""
    echo "=================================================="
    echo "Final Summary"
    echo "=================================================="
    echo "Total Scenarios: $scenario_count"
    echo "Passed: $((scenario_count - ${#failed_scenarios[@]}))"
    echo "Failed: ${#failed_scenarios[@]}"

    if [ ${#failed_scenarios[@]} -gt 0 ]; then
        echo ""
        log_error "Failed scenarios:"
        for scenario in "${failed_scenarios[@]}"; do
            echo "  - $scenario"
        done
        echo ""
        echo "=================================================="
        exit 1
    else
        echo ""
        log_success "All scenarios passed!"
        echo "=================================================="
        exit 0
    fi
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --skip-slow)
            SKIP_SLOW_TESTS=true
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS] [SCENARIOS...]"
            echo ""
            echo "Options:"
            echo "  -v, --verbose      Show verbose output"
            echo "  --skip-slow        Skip slow tests (large downloads)"
            echo "  -h, --help         Show this help message"
            echo ""
            echo "Environment Variables:"
            echo "  PIVNET_TOKEN       Pivotal Network API token (required for most tests)"
            echo "  TEST_CACHE_DIR     Test cache directory (default: /tmp/tile-diff-test-cache)"
            echo "  TILE_DIFF_BIN      Path to tile-diff binary (default: ./tile-diff)"
            echo ""
            echo "Examples:"
            echo "  $0                                # Run all tests"
            echo "  $0 --verbose                      # Run all tests with verbose output"
            echo "  $0 01_non_interactive_mode        # Run specific test"
            echo "  $0 --skip-slow                    # Skip slow tests"
            exit 0
            ;;
        *)
            RUN_ALL_TESTS=false
            break
            ;;
    esac
done

main "$@"
