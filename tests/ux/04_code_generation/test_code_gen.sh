#!/usr/bin/env bash
#
# UX TEST: Code Generation
# Tests code generation across different languages and frameworks
#

set +e  # Don't exit on test failures

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

OLLAMA_CODE="../../build/ollama-code.exe"
TEST_DIR="/tmp/ollama_codegen_test_$$"
PASSED=0
FAILED=0

echo "========================================="
echo "UX TEST: Code Generation"
echo "========================================="
echo ""

mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

test_code_gen() {
    local test_id="$1"
    local message="$2"
    local pattern="$3"
    local description="$4"

    echo -n "[$test_id] $description... "

    output=$(timeout 45s $OLLAMA_CODE ask "$message" --mode autonomous 2>&1 || true)

    # Check if code was generated and contains expected pattern
    if echo "$output" | grep -q "Arquivo criado" && find . -type f -name "$pattern" | head -1 | xargs cat | grep -q "."; then
        echo -e "${GREEN}PASS${NC}"
        ((PASSED++))
        return 0
    else
        echo -e "${RED}FAIL${NC}"
        ((FAILED++))
        return 1
    fi
}

echo "=== Python ==="
test_code_gen "PY-01" "cria um script Python que calcula fatorial" "*.py" "Python factorial"
test_code_gen "PY-02" "faz um servidor HTTP simples em Python" "*.py" "Python HTTP server"
echo ""

echo "=== JavaScript/HTML/CSS ==="
test_code_gen "JS-01" "cria uma página HTML com botão" "*.html" "HTML button page"
test_code_gen "JS-02" "faz um arquivo CSS com dark mode" "*.css" "CSS dark mode"
test_code_gen "JS-03" "cria JavaScript que valida email" "*.js" "JS email validator"
echo ""

echo "=== Go ==="
test_code_gen "GO-01" "cria um programa Go que lê arquivo" "*.go" "Go file reader"
test_code_gen "GO-02" "faz uma API REST simples em Go" "*.go" "Go REST API"
echo ""

echo "=== Multi-file Projects ==="
test_code_gen "MULTI-01" "cria HTML, CSS e JS separados para um contador" "*.html" "Multi-file counter"
echo ""

cd - > /dev/null
rm -rf "$TEST_DIR"

echo "========================================="
echo "SUMMARY"
echo "========================================="
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo "Total: $((PASSED + FAILED))"

[ $FAILED -eq 0 ] && exit 0 || exit 1
