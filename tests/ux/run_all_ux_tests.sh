#!/usr/bin/env bash
#
# Master UX/UI Regression Test Suite
# Runs all UX tests and generates consolidated report
#

set +e  # Don't exit on test failures - we want to run all suites

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="$SCRIPT_DIR/logs"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
REPORT_FILE="$LOG_DIR/ux_test_report_$TIMESTAMP.log"

# Test categories
TESTS=(
    "01_conversation_flow/test_greetings.sh"
    "02_intent_detection/test_all_intents.sh"
    "03_file_operations/test_file_ops.sh"
    "04_code_generation/test_code_gen.sh"
    "05_web_search/test_web_search.sh"
    "07_error_handling/test_error_handling.sh"
    "08_output_quality/test_useful_output.sh"
)

# Create log directory
mkdir -p "$LOG_DIR"

echo "==========================================" | tee "$REPORT_FILE"
echo "OLLAMA CODE - UX/UI REGRESSION TEST SUITE" | tee -a "$REPORT_FILE"
echo "==========================================" | tee -a "$REPORT_FILE"
echo "Started: $(date)" | tee -a "$REPORT_FILE"
echo "" | tee -a "$REPORT_FILE"

TOTAL_PASSED=0
TOTAL_FAILED=0
SUITE_PASSED=0
SUITE_FAILED=0

# Run each test suite
for test in "${TESTS[@]}"; do
    test_path="$SCRIPT_DIR/$test"
    test_name=$(basename "$test" .sh)
    category=$(dirname "$test")

    echo -e "${BLUE}Running: $category/$test_name${NC}" | tee -a "$REPORT_FILE"
    echo "----------------------------------------" | tee -a "$REPORT_FILE"

    if [ -f "$test_path" ]; then
        chmod +x "$test_path"

        # Run test and capture output
        log_file="$LOG_DIR/${category//\//_}_$test_name.log"

        if bash "$test_path" 2>&1 | tee "$log_file" | tee -a "$REPORT_FILE"; then
            echo -e "${GREEN}✓ $test_name PASSED${NC}" | tee -a "$REPORT_FILE"
            ((SUITE_PASSED++))

            # Extract passed/failed counts if available
            if grep -q "Passed:" "$log_file"; then
                passed=$(grep "Passed:" "$log_file" | tail -1 | grep -oP '\d+' | head -1 || echo "0")
                failed=$(grep "Failed:" "$log_file" | tail -1 | grep -oP '\d+' | head -1 || echo "0")
                TOTAL_PASSED=$((TOTAL_PASSED + passed))
                TOTAL_FAILED=$((TOTAL_FAILED + failed))
            fi
        else
            echo -e "${RED}✗ $test_name FAILED${NC}" | tee -a "$REPORT_FILE"
            ((SUITE_FAILED++))

            # Extract counts even from failed suite
            if grep -q "Passed:" "$log_file"; then
                passed=$(grep "Passed:" "$log_file" | tail -1 | grep -oP '\d+' | head -1 || echo "0")
                failed=$(grep "Failed:" "$log_file" | tail -1 | grep -oP '\d+' | head -1 || echo "0")
                TOTAL_PASSED=$((TOTAL_PASSED + passed))
                TOTAL_FAILED=$((TOTAL_FAILED + failed))
            fi
        fi
    else
        echo -e "${YELLOW}⚠ Test not found: $test_path${NC}" | tee -a "$REPORT_FILE"
    fi

    echo "" | tee -a "$REPORT_FILE"
done

# Final Summary
echo "==========================================" | tee -a "$REPORT_FILE"
echo "FINAL SUMMARY" | tee -a "$REPORT_FILE"
echo "==========================================" | tee -a "$REPORT_FILE"
echo "" | tee -a "$REPORT_FILE"

echo "Test Suites:" | tee -a "$REPORT_FILE"
echo -e "  Passed: ${GREEN}$SUITE_PASSED${NC}" | tee -a "$REPORT_FILE"
echo -e "  Failed: ${RED}$SUITE_FAILED${NC}" | tee -a "$REPORT_FILE"
echo "  Total: $((SUITE_PASSED + SUITE_FAILED))" | tee -a "$REPORT_FILE"
echo "" | tee -a "$REPORT_FILE"

echo "Individual Tests:" | tee -a "$REPORT_FILE"
echo -e "  Passed: ${GREEN}$TOTAL_PASSED${NC}" | tee -a "$REPORT_FILE"
echo -e "  Failed: ${RED}$TOTAL_FAILED${NC}" | tee -a "$REPORT_FILE"
echo "  Total: $((TOTAL_PASSED + TOTAL_FAILED))" | tee -a "$REPORT_FILE"
echo "" | tee -a "$REPORT_FILE"

if [ $((TOTAL_PASSED + TOTAL_FAILED)) -gt 0 ]; then
    success_rate=$(awk "BEGIN {printf \"%.1f\", ($TOTAL_PASSED/($TOTAL_PASSED + $TOTAL_FAILED))*100}")
    echo "Success Rate: ${success_rate}%" | tee -a "$REPORT_FILE"
fi

echo "" | tee -a "$REPORT_FILE"
echo "Completed: $(date)" | tee -a "$REPORT_FILE"
echo "Report saved: $REPORT_FILE" | tee -a "$REPORT_FILE"
echo "" | tee -a "$REPORT_FILE"

if [ $SUITE_FAILED -eq 0 ] && [ $TOTAL_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓✓✓ ALL UX TESTS PASSED! ✓✓✓${NC}" | tee -a "$REPORT_FILE"
    exit 0
else
    echo -e "${RED}✗✗✗ SOME UX TESTS FAILED ✗✗✗${NC}" | tee -a "$REPORT_FILE"
    echo "" | tee -a "$REPORT_FILE"
    echo "Failed test logs:" | tee -a "$REPORT_FILE"
    ls -1 "$LOG_DIR"/*.log 2>/dev/null | while read log; do
        if grep -q "FAIL" "$log" 2>/dev/null; then
            echo "  - $(basename "$log")" | tee -a "$REPORT_FILE"
        fi
    done
    exit 1
fi
