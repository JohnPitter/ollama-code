#!/usr/bin/env bash
#
# UX TEST: Conversation Flow - Greetings and Courtesy
# Tests that greetings and courtesy messages are correctly classified as 'question'
# NOT as 'web_search' or other intents
#

set +e  # Don't exit on test failures - we want to run all tests

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

OLLAMA_CODE="../../build/ollama-code.exe"
TEST_NAME="Conversation Flow - Greetings"
PASSED=0
FAILED=0

echo "========================================="
echo "$TEST_NAME"
echo "========================================="
echo ""

# Function to test single message
test_message() {
    local test_id="$1"
    local message="$2"
    local expected_intent="$3"
    local description="$4"

    echo -n "[$test_id] $description... "

    # Execute command and capture output
    output=$(timeout 30s $OLLAMA_CODE ask "$message" --mode autonomous 2>&1 || true)

    # Check if expected intent was detected
    if echo "$output" | grep -q "Intenção: $expected_intent"; then
        echo -e "${GREEN}PASS${NC}"
        ((PASSED++))
        return 0
    else
        echo -e "${RED}FAIL${NC}"
        echo "  Expected: $expected_intent"
        echo "  Output: $output" | head -5
        ((FAILED++))
        return 1
    fi
}

# Test Cases - Greetings (should be 'question')
echo "=== Greetings ==="
test_message "GREET-01" "oi" "question" "Simple greeting 'oi'"
test_message "GREET-02" "olá" "question" "Simple greeting 'olá'"
test_message "GREET-03" "oi, tudo bem?" "question" "Greeting with question"
test_message "GREET-04" "olá! como vai?" "question" "Greeting with follow-up"
echo ""

# Test Cases - Thanks (should be 'question')
echo "=== Thanks and Acknowledgments ==="
test_message "THANKS-01" "obrigado" "question" "Simple thanks"
test_message "THANKS-02" "obrigado!" "question" "Thanks with exclamation"
test_message "THANKS-03" "valeu" "question" "Informal thanks"
test_message "THANKS-04" "muito obrigado pela ajuda" "question" "Extended thanks"
echo ""

# Test Cases - Confirmations (should be 'question')
echo "=== Confirmations ==="
test_message "CONFIRM-01" "ok" "question" "Simple ok"
test_message "CONFIRM-02" "certo" "question" "Agreement"
test_message "CONFIRM-03" "entendi" "question" "Understanding"
test_message "CONFIRM-04" "show" "question" "Informal ok"
test_message "CONFIRM-05" "blz" "question" "Very informal ok"
echo ""

# Test Cases - State (should be 'question')
echo "=== State Messages ==="
test_message "STATE-01" "estou bem" "question" "Positive state"
test_message "STATE-02" "tudo certo" "question" "All good"
test_message "STATE-03" "tudo ótimo" "question" "Everything great"
test_message "STATE-04" "estou bem e você?" "question" "State with question back"
echo ""

# Test Cases - Farewells (should be 'question')
echo "=== Farewells ==="
test_message "BYE-01" "tchau" "question" "Simple bye"
test_message "BYE-02" "até logo" "question" "See you later"
test_message "BYE-03" "até mais" "question" "Until later"
test_message "BYE-04" "valeu, até mais!" "question" "Thanks and bye"
echo ""

# Summary
echo "========================================="
echo "SUMMARY: $TEST_NAME"
echo "========================================="
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo "Total: $((PASSED + FAILED))"

if [ $FAILED -eq 0 ]; then
    echo -e "\n${GREEN}✓ ALL TESTS PASSED${NC}"
    exit 0
else
    echo -e "\n${RED}✗ SOME TESTS FAILED${NC}"
    exit 1
fi
