#!/bin/bash
# ABOUTME: List all available test scenarios and their descriptions.
# ABOUTME: Useful for understanding what tests are available.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCENARIOS_DIR="$SCRIPT_DIR/scenarios"

echo "Available Test Scenarios:"
echo "========================="
echo ""

for scenario in "$SCENARIOS_DIR"/*.sh; do
    if [ -f "$scenario" ]; then
        scenario_name=$(basename "$scenario" .sh)

        # Extract description from ABOUTME comments
        description=$(grep "^# ABOUTME:" "$scenario" | head -2 | sed 's/^# ABOUTME: //' | tr '\n' ' ' | sed 's/  */ /g')

        echo "📋 $scenario_name"
        echo "   $description"
        echo ""
    fi
done

echo "Usage:"
echo "======"
echo "Run all tests:              make test-acceptance"
echo "Run with verbose output:    make test-acceptance-verbose"
echo "Run specific scenario:      ./test/acceptance/scenarios/<scenario>.sh"
echo ""
