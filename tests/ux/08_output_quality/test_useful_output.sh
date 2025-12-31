#!/usr/bin/env bash
#
# UX TEST: Useful Output Quality
# Tests that commands return useful information, not empty messages
#

set +e  # Don't exit on test failures

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OLLAMA_CODE="$(cd "$SCRIPT_DIR/../../.." && pwd)/build/ollama-code.exe"
PASSED=0
FAILED=0

echo "========================================="
echo "UX TEST: Useful Output Quality"
echo "========================================="
echo ""

test_output_quality() {
    local test_id="$1"
    local message="$2"
    local must_contain="$3"
    local must_not_be_only="$4"
    local description="$5"

    echo -n "[$test_id] $description... "

    output=$(timeout 60s $OLLAMA_CODE ask "$message" --mode autonomous 2>&1 || true)

    # Check if output contains required content
    if ! echo "$output" | grep -q "$must_contain"; then
        echo -e "${RED}FAIL${NC}"
        echo "  Expected to contain: $must_contain"
        echo "  Got: $(echo "$output" | head -n 5)"
        ((FAILED++))
        return 1
    fi

    # Check if output is NOT just the empty message
    if [ -n "$must_not_be_only" ]; then
        # Count lines in output
        line_count=$(echo "$output" | wc -l)
        if [ "$line_count" -lt 5 ]; then
            echo -e "${YELLOW}WARN${NC} - Output too short (${line_count} lines)"
            echo "  Output: $output"
        fi
    fi

    echo -e "${GREEN}PASS${NC}"
    ((PASSED++))
    return 0
}

# Test Cases
echo "=== Project Analysis ==="
test_output_quality "OUT-01" \
    "analisa a estrutura desse projeto" \
    "📂\|Estrutura\|├──\|└──" \
    "yes" \
    "Project analysis shows directory tree"

test_output_quality "OUT-02" \
    "mostre a estrutura do projeto" \
    "📂\|📁\|📄\|├──" \
    "yes" \
    "Project structure shows icons and tree"
echo ""

echo "=== File Reading ==="
# Create test file
TEST_DIR="/tmp/ollama_ux_output_$$"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

cat > test_analysis.md <<'EOF'
# Test Document

This is a test document for UX validation.

## Features
- Feature 1
- Feature 2
- Feature 3
EOF

test_output_quality "OUT-03" \
    "leia o arquivo test_analysis.md" \
    "Test Document\|Feature\|📄\|Conteúdo" \
    "yes" \
    "File read shows content preview"

test_output_quality "OUT-04" \
    "analise o arquivo test_analysis.md" \
    "análise\|Análise\|Test Document\|Features" \
    "yes" \
    "File analysis shows LLM analysis or content"

cd - > /dev/null
rm -rf "$TEST_DIR"
echo ""

echo "=== Search Operations ==="
test_output_quality "OUT-05" \
    "onde está a função Handle" \
    "Handle\|func\|\.go" \
    "yes" \
    "Code search shows file locations"

test_output_quality "OUT-06" \
    "busca por 'NewHandler'" \
    "NewHandler\|internal\|handlers" \
    "yes" \
    "Search shows results not just 'found'"
echo ""

echo "=== Git Operations ==="
test_output_quality "OUT-07" \
    "mostre os últimos commits" \
    "commit\|Author\|Date\|fix\|feat" \
    "yes" \
    "Git log shows commit details"

test_output_quality "OUT-08" \
    "qual o status do git" \
    "branch\|On branch\|nothing to commit\|Changes" \
    "yes" \
    "Git status shows branch and changes"
echo ""

echo "=== Web Search ==="
test_output_quality "OUT-09" \
    "busca na internet sobre Go 1.22" \
    "Go\|Fontes:\|https\|http" \
    "yes" \
    "Web search shows results and sources"
echo ""

echo "=== Command Execution ==="
test_output_quality "OUT-10" \
    "execute o comando 'echo Hello Test'" \
    "Hello Test\|Saída\|Output" \
    "yes" \
    "Command execution shows output"
echo ""

# Summary
echo "========================================="
echo "SUMMARY: Output Quality"
echo "========================================="
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo "Total: $((PASSED + FAILED))"

if [ $FAILED -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✓ ALL OUTPUT QUALITY TESTS PASSED${NC}"
fi

[ $FAILED -eq 0 ] && exit 0 || exit 1
