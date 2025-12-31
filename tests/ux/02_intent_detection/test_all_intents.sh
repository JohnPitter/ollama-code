#!/usr/bin/env bash
#
# UX TEST: Intent Detection - All Intent Types
# Tests that all 8 intent types are correctly detected
#

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

OLLAMA_CODE="../../../build/ollama-code.exe"
TEST_NAME="Intent Detection - All Types"
PASSED=0
FAILED=0

echo "========================================="
echo "$TEST_NAME"
echo "========================================="
echo ""

test_intent() {
    local test_id="$1"
    local message="$2"
    local expected_intent="$3"
    local description="$4"

    echo -n "[$test_id] $description... "

    output=$(echo "$message" | timeout 30s $OLLAMA_CODE ask --mode autonomous 2>&1 || true)

    if echo "$output" | grep -q "Intenção: $expected_intent"; then
        echo -e "${GREEN}PASS${NC}"
        ((PASSED++))
        return 0
    else
        echo -e "${RED}FAIL${NC}"
        echo "  Expected: $expected_intent"
        echo "  Got: $(echo "$output" | grep "Intenção:" | head -1)"
        ((FAILED++))
        return 1
    fi
}

# 1. read_file
echo "=== Intent: read_file ==="
test_intent "READ-01" "leia o arquivo README.md" "read_file" "Read file command"
test_intent "READ-02" "mostre o conteúdo de main.go" "read_file" "Show file content"
test_intent "READ-03" "analisa a função ProcessMessage" "read_file" "Analyze function (read only)"
test_intent "READ-04" "explica o que faz agent.go" "read_file" "Explain file"
test_intent "READ-05" "faz code review do handler.go" "read_file" "Code review (read only)"
echo ""

# 2. write_file
echo "=== Intent: write_file ==="
test_intent "WRITE-01" "crie um arquivo test.txt" "write_file" "Create file"
test_intent "WRITE-02" "adicione logging no main.go" "write_file" "Add code to file"
test_intent "WRITE-03" "corrija o bug no handler.go" "write_file" "Fix bug (modify)"
test_intent "WRITE-04" "desenvolve um site HTML" "write_file" "Develop website"
test_intent "WRITE-05" "faz um script python" "write_file" "Create script"
test_intent "WRITE-06" "refatora a função X" "write_file" "Refactor (modify)"
echo ""

# 3. execute_command
echo "=== Intent: execute_command ==="
test_intent "EXEC-01" "rode os testes" "execute_command" "Run tests"
test_intent "EXEC-02" "execute npm install" "execute_command" "Execute npm"
test_intent "EXEC-03" "faça build do projeto" "execute_command" "Build project"
test_intent "EXEC-04" "roda go test" "execute_command" "Run go test"
echo ""

# 4. search_code
echo "=== Intent: search_code ==="
test_intent "SEARCH-01" "busca a função ProcessMessage" "search_code" "Search function"
test_intent "SEARCH-02" "procure por database connection" "search_code" "Search pattern"
test_intent "SEARCH-03" "encontre todos os handlers" "search_code" "Find handlers"
test_intent "SEARCH-04" "onde está a struct User" "search_code" "Find struct"
echo ""

# 5. analyze_project
echo "=== Intent: analyze_project ==="
test_intent "ANALYZE-01" "qual a estrutura do projeto" "analyze_project" "Project structure"
test_intent "ANALYZE-02" "quais arquivos temos" "analyze_project" "List files"
test_intent "ANALYZE-03" "me mostre a arquitetura" "analyze_project" "Show architecture"
test_intent "ANALYZE-04" "analisa o projeto completo" "analyze_project" "Analyze full project"
echo ""

# 6. git_operation
echo "=== Intent: git_operation ==="
test_intent "GIT-01" "commita essas mudanças" "git_operation" "Git commit"
test_intent "GIT-02" "crie uma branch" "git_operation" "Git branch"
test_intent "GIT-03" "mostra o diff" "git_operation" "Git diff"
test_intent "GIT-04" "git status" "git_operation" "Git status"
echo ""

# 7. web_search
echo "=== Intent: web_search ==="
test_intent "WEB-01" "pesquise informações sobre React" "web_search" "Search info online"
test_intent "WEB-02" "busque documentação da API X" "web_search" "Search documentation"
test_intent "WEB-03" "qual a temperatura em São Paulo" "web_search" "Real-time data"
test_intent "WEB-04" "quais as últimas notícias sobre Go" "web_search" "Search news"
echo ""

# 8. question
echo "=== Intent: question ==="
test_intent "QUEST-01" "o que é REST" "question" "Conceptual question"
test_intent "QUEST-02" "como funciona async/await" "question" "How it works"
test_intent "QUEST-03" "explique closures" "question" "Explain concept"
test_intent "QUEST-04" "oi" "question" "Greeting"
test_intent "QUEST-05" "obrigado" "question" "Thanks"
echo ""

# Summary
echo "========================================="
echo "SUMMARY: $TEST_NAME"
echo "========================================="
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo "Total: $((PASSED + FAILED))"
echo "Success Rate: $(awk "BEGIN {printf \"%.1f\", ($PASSED/$((PASSED + FAILED)))*100}")%"

if [ $FAILED -eq 0 ]; then
    echo -e "\n${GREEN}✓ ALL TESTS PASSED${NC}"
    exit 0
else
    echo -e "\n${RED}✗ SOME TESTS FAILED${NC}"
    exit 1
fi
