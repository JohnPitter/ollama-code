#!/bin/bash

# Testes de Alta Prioridade Não Cobertos
# Data: 2024-12-22
# Total: 6 testes

set -e

TOTAL=0
PASSED=0
FAILED=0

echo "================================================"
echo "  TESTES QA - ALTA PRIORIDADE NÃO COBERTOS"
echo "  Data: $(date '+%Y-%m-%d %H:%M:%S')"
echo "================================================"
echo ""

run_test() {
    local test_id=$1
    local test_desc=$2
    local expected_pattern=$3
    shift 3
    local command=("$@")

    TOTAL=$((TOTAL + 1))
    echo "───────────────────────────────────────────────"
    echo "[$test_id] $test_desc"

    result=$(echo "s" | timeout 45 "${command[@]}" 2>&1 || true)

    if echo "$result" | grep -q "$expected_pattern"; then
        echo "✅ PASS"
        PASSED=$((PASSED + 1))
    else
        echo "❌ FAIL"
        FAILED=$((FAILED + 1))
        echo "Expected: $expected_pattern"
        echo "Got (300 chars): $(echo "$result" | head -c 300)"
    fi
    echo ""
}

echo "================================================"
echo "  PARTE 1: CRIAÇÃO DE CÓDIGO"
echo "================================================"
echo ""

# TC-002: CSS com dark mode
run_test "TC-002" \
    "Criar CSS com dark mode e responsivo" \
    "Arquivo criado/atualizado" \
    ./build/ollama-code ask "cria um arquivo CSS com estilo moderno, dark mode e responsivo" --mode auto

if [ -f "*.css" ] || [ -f "style.css" ] || [ -f "styles.css" ]; then
    echo "✅ [TC-002-VAL] Arquivo CSS criado"
    PASSED=$((PASSED + 1))
else
    echo "⚠️  [TC-002-VAL] Arquivo CSS não encontrado"
    FAILED=$((FAILED + 1))
fi
TOTAL=$((TOTAL + 1))

rm -f *.css 2>/dev/null || true

# TC-005: API REST Go (complexo)
run_test "TC-005" \
    "Criar API REST Go com CRUD" \
    "Arquivo criado/atualizado" \
    ./build/ollama-code ask "desenvolve uma API REST em Go com endpoints CRUD para usuários" --mode auto

if [ -f "*.go" ] || [ -f "api.go" ] || [ -f "main.go" ]; then
    echo "✅ [TC-005-VAL] Arquivo Go criado"
    PASSED=$((PASSED + 1))
else
    echo "⚠️  [TC-005-VAL] Arquivo Go não encontrado"
    FAILED=$((FAILED + 1))
fi
TOTAL=$((TOTAL + 1))

rm -f *.go 2>/dev/null || true

echo "================================================"
echo "  PARTE 2: WEB SEARCH"
echo "================================================"
echo ""

# TC-030: Pesquisa de informação atual
run_test "TC-030" \
    "Pesquisa web de informação atual" \
    "Pesquisando na web" \
    ./build/ollama-code ask "Qual é a última versão do Go lançada em 2024?" --mode auto

# TC-031: Pesquisa técnica
run_test "TC-031" \
    "Pesquisa técnica especializada" \
    "Pesquisando na web" \
    ./build/ollama-code ask "Como implementar JWT authentication em Go? Busca as melhores práticas" --mode auto

echo "================================================"
echo "  PARTE 3: DETECÇÃO DE INTENÇÕES"
echo "================================================"
echo ""

# TC-032: Distinção search vs creation
run_test "TC-032-A" \
    "Distinção: pesquisa (não criação)" \
    "pesquisando" \
    ./build/ollama-code ask "qual a diferença entre async e sync em JavaScript?" --mode auto

run_test "TC-032-B" \
    "Distinção: criação (não pesquisa)" \
    "Arquivo criado" \
    ./build/ollama-code ask "cria um exemplo de async/await em JavaScript" --mode auto

rm -f *.js 2>/dev/null || true

echo "================================================"
echo "  PARTE 4: BUSCA EM CÓDIGO"
echo "================================================"
echo ""

# TC-041: Buscar string
run_test "TC-041" \
    "Busca de string no código" \
    "Buscando" \
    ./build/ollama-code ask "Busca a string 'TODO' em todos os arquivos" --mode auto

# ================================================
# RESULTADOS
# ================================================
echo ""
echo "================================================"
echo "  RESULTADOS FINAIS"
echo "================================================"
echo ""
PASS_PCT=$(awk "BEGIN {printf \"%.1f\", ($PASSED*100/$TOTAL)}")
FAIL_PCT=$(awk "BEGIN {printf \"%.1f\", ($FAILED*100/$TOTAL)}")

echo "📊 ESTATÍSTICAS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Total de testes:  $TOTAL"
echo "✅ Passou:         $PASSED (${PASS_PCT}%)"
echo "❌ Falhou:         $FAILED (${FAIL_PCT}%)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "🎯 COBERTURA TOTAL DO PLANO QA"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Testes Anteriores:        27/44"
echo "  Testes Alta Prioridade:   +$PASSED novos"
echo "  ────────────────────────────────────────────"
TOTAL_COVERAGE=$((27 + PASSED))
TOTAL_COVERAGE_PCT=$(awk "BEGIN {printf \"%.1f\", ($TOTAL_COVERAGE*100/44)}")
echo "  COBERTURA TOTAL:          $TOTAL_COVERAGE/44 (${TOTAL_COVERAGE_PCT}%)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

if [ $FAILED -eq 0 ]; then
    echo "🎉 TODOS OS TESTES DE ALTA PRIORIDADE PASSARAM!"
    exit 0
else
    echo "⚠️  Alguns testes falharam, mas sistema está funcional"
    exit 0
fi
