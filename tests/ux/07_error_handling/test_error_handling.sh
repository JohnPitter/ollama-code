#!/usr/bin/env bash
#
# UX TEST: Error Handling
# Tests graceful error handling and user-friendly error messages
#

set +e  # Don't exit on test failures

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

OLLAMA_CODE="../../build/ollama-code.exe"
PASSED=0
FAILED=0

echo "========================================="
echo "UX TEST: Error Handling"
echo "========================================="
echo ""

test_error() {
    local test_id="$1"
    local message="$2"
    local should_contain="$3"
    local description="$4"

    echo -n "[$test_id] $description... "

    output=$(timeout 30s $OLLAMA_CODE ask "$message" --mode autonomous 2>&1 || true)

    # Check if error is handled gracefully (no crash, user-friendly message)
    if echo "$output" | grep -qi "$should_contain" && ! echo "$output" | grep -q "panic\\|fatal"; then
        echo -e "${GREEN}PASS${NC}"
        ((PASSED++))
        return 0
    else
        echo -e "${RED}FAIL${NC}"
        echo "  Expected to contain: $should_contain"
        echo "  Output: $(echo "$output" | head -3)"
        ((FAILED++))
        return 1
    fi
}

echo "=== File Not Found ==="
test_error "ERR-01" "leia o arquivo naoexiste123.txt" "não\\|encontrado\\|not found\\|erro" "File not found"
echo ""

echo "=== Invalid Input ==="
test_error "ERR-02" "" "Hello\\|ajudar\\|help" "Empty input (should respond politely)"
test_error "ERR-03" "askdjfhaksjdfh" "desculpe\\|não entendi\\|sorry" "Gibberish input"
echo ""

echo "=== Ambiguous Requests ==="
test_error "ERR-04" "faça algo" "específico\\|detalhe\\|more specific\\|o que" "Vague request"
echo ""

echo "========================================="
echo "SUMMARY"
echo "========================================="
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo "Total: $((PASSED + FAILED))"

[ $FAILED -eq 0 ] && exit 0 || exit 1
