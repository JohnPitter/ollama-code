#!/usr/bin/env bash
#
# UX TEST: File Operations
# Tests file read/write operations and validation
#

set +e  # Don't exit on test failures

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

OLLAMA_CODE="../../build/ollama-code.exe"
TEST_DIR="/tmp/ollama_ux_test_$$"
PASSED=0
FAILED=0

echo "========================================="
echo "UX TEST: File Operations"
echo "========================================="
echo ""

# Setup
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

test_file_op() {
    local test_id="$1"
    local message="$2"
    local validation="$3"
    local description="$4"

    echo -n "[$test_id] $description... "

    output=$(timeout 30s $OLLAMA_CODE ask "$message" --mode autonomous 2>&1 || true)

    if eval "$validation"; then
        echo -e "${GREEN}PASS${NC}"
        ((PASSED++))
        return 0
    else
        echo -e "${RED}FAIL${NC}"
        echo "  Validation failed: $validation"
        ((FAILED++))
        return 1
    fi
}

# Test Cases
echo "=== File Creation ==="
test_file_op "FILE-01" "crie um arquivo test.txt com 'Hello World'" \
    "test -f test.txt" "Create simple file"

test_file_op "FILE-02" "cria um script Python que printa hello" \
    "test -f *.py" "Create Python script"

test_file_op "FILE-03" "faz um arquivo HTML básico" \
    "test -f *.html" "Create HTML file"
echo ""

echo "=== File Reading ==="
echo "Sample content" > sample.txt
test_file_op "READ-01" "leia o arquivo sample.txt" \
    "echo '$output' | grep -q 'Sample content'" "Read existing file"
echo ""

echo "=== File Modification ==="
echo "function old() { return 1; }" > code.js
test_file_op "MOD-01" "adicione comentário no code.js" \
    "grep -q '//' code.js || grep -q '//'" "Add comment to file"
echo ""

# Cleanup
cd - > /dev/null
rm -rf "$TEST_DIR"

# Summary
echo "========================================="
echo "SUMMARY"
echo "========================================="
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo "Total: $((PASSED + FAILED))"

[ $FAILED -eq 0 ] && exit 0 || exit 1
