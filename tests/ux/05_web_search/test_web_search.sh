#!/usr/bin/env bash
#
# UX TEST: Web Search Functionality
# Tests web search capabilities and result formatting
#

set +e  # Don't exit on test failures

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

OLLAMA_CODE="../../build/ollama-code.exe"
PASSED=0
FAILED=0

echo "========================================="
echo "UX TEST: Web Search"
echo "========================================="
echo ""

test_web_search() {
    local test_id="$1"
    local message="$2"
    local expected_pattern="$3"
    local description="$4"

    echo -n "[$test_id] $description... "

    output=$(timeout 60s $OLLAMA_CODE ask "$message" --mode autonomous 2>&1 || true)

    # Check if output contains sources and expected pattern
    if echo "$output" | grep -q "📚.*Fontes:" && echo "$output" | grep -qi "$expected_pattern"; then
        echo -e "${GREEN}PASS${NC}"
        ((PASSED++))
        return 0
    else
        echo -e "${RED}FAIL${NC}"
        echo "  Expected pattern: $expected_pattern"
        ((FAILED++))
        return 1
    fi
}

echo "=== Real-time Information ==="
test_web_search "WEB-01" "qual a temperatura atual em São Paulo" "temperatura\\|graus\\|°C\\|clima" "Temperature query"
test_web_search "WEB-02" "pesquise sobre Python 3.12" "Python 3.12\\|features\\|release" "Tech search"
echo ""

echo "=== Documentation Search ==="
test_web_search "WEB-03" "busque documentação sobre React hooks" "React\\|hooks\\|useState\\|useEffect" "React docs"
test_web_search "WEB-04" "procure tutorial de Go" "Go\\|golang\\|tutorial\\|learn" "Go tutorial"
echo ""

echo "=== General Information ==="
test_web_search "WEB-05" "pesquise sobre inteligência artificial" "IA\\|AI\\|artificial\\|intelligence" "AI search"
echo ""

echo "========================================="
echo "SUMMARY"
echo "========================================="
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo "Total: $((PASSED + FAILED))"

[ $FAILED -eq 0 ] && exit 0 || exit 1
