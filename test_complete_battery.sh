#!/bin/bash

# Bateria Completa de Testes QA
# Data: 2024-12-21
# Objetivo: Testar TODOS os bugs corrigidos (#1, #4, #6, #7, #8, #9, #11, #12, #13)

set -e

TOTAL=0
PASSED=0
FAILED=0

echo "================================================"
echo "  BATERIA COMPLETA DE TESTES QA"
echo "  Data: $(date '+%Y-%m-%d %H:%M:%S')"
echo "================================================"
echo ""

# Função auxiliar para testar
run_test() {
    local test_id=$1
    local test_desc=$2
    local expected_pattern=$3
    shift 3
    local command=("$@")

    TOTAL=$((TOTAL + 1))
    echo "───────────────────────────────────────────────"
    echo "[$test_id] $test_desc"

    # Executar comando
    result=$(echo "s" | timeout 30 "${command[@]}" 2>&1 || true)

    # Verificar se contém padrão esperado
    if echo "$result" | grep -q "$expected_pattern"; then
        echo "✅ PASS"
        PASSED=$((PASSED + 1))
    else
        echo "❌ FAIL"
        FAILED=$((FAILED + 1))
        echo "Esperado: $expected_pattern"
        echo "Output (400 chars): $(echo "$result" | head -c 400)"
    fi
    echo ""
}

# ================================================
# BUGS CORRIGIDOS - REGRESSÃO
# ================================================
echo "================================================"
echo "  PARTE 1: TESTES DE REGRESSÃO"
echo "================================================"
echo ""

# BUG #1
run_test "BUG1" \
    "Multi-file creation" \
    "Projeto criado" \
    ./build/ollama-code ask "Cria arquivos test1.txt e test2.txt" --mode auto
rm -f test1.txt test2.txt 2>/dev/null || true

# BUG #4
run_test "BUG4" \
    "JSON extraction & valid filename" \
    "Arquivo criado/atualizado" \
    ./build/ollama-code ask "Cria arquivo helper.go" --mode auto
rm -f helper.go 2>/dev/null || true

# BUG #6
echo "test content" > temp_file.txt
run_test "BUG6" \
    "File overwrite protection" \
    "Editando arquivo existente" \
    ./build/ollama-code ask "Atualiza temp_file.txt" --mode auto
rm -f temp_file.txt 2>/dev/null || true

# BUG #9
run_test "BUG9-1" \
    "Dotfile .env creation" \
    "Arquivo criado/atualizado" \
    ./build/ollama-code ask "Cria .env" --mode auto
rm -f .env 2>/dev/null || true

run_test "BUG9-2" \
    "Dotfile .gitignore creation" \
    "Arquivo criado/atualizado" \
    ./build/ollama-code ask "Cria .gitignore" --mode auto
rm -f .gitignore 2>/dev/null || true

# BUG #12
echo "buggy code" > fix_me.go
run_test "BUG12" \
    "Keyword 'corrige' detection" \
    "Editando arquivo existente" \
    ./build/ollama-code ask "Corrige fix_me.go" --mode auto
rm -f fix_me.go 2>/dev/null || true

# ================================================
# BUGS NOVOS - SESSÃO ATUAL
# ================================================
echo "================================================"
echo "  PARTE 2: BUGS CORRIGIDOS NA SESSÃO ATUAL"
echo "================================================"
echo ""

# BUG #7
run_test "BUG7-1" \
    "Git status" \
    "Operação git 'status'" \
    ./build/ollama-code ask "Mostra status do git" --mode auto

run_test "BUG7-2" \
    "Git diff" \
    "Operação git 'diff'" \
    ./build/ollama-code ask "Mostra mudanças do git" --mode auto

run_test "BUG7-3" \
    "Git log" \
    "Operação git 'log'" \
    ./build/ollama-code ask "Mostra histórico de commits" --mode auto

# BUG #8
run_test "BUG8-1" \
    "Integration hint (JS)" \
    "Dica" \
    ./build/ollama-code ask "Cria app.js e conecta no index.html" --mode auto
rm -f app.js 2>/dev/null || true

run_test "BUG8-2" \
    "Integration hint (CSS)" \
    "Dica" \
    ./build/ollama-code ask "Cria styles.css e conecta no index.html" --mode auto
rm -f styles.css 2>/dev/null || true

# BUG #11
run_test "BUG11-1" \
    "Multi-file read (comma)" \
    "Lendo 2 arquivos" \
    ./build/ollama-code ask "Lê go.mod, main.go" --mode auto

run_test "BUG11-2" \
    "Multi-file read (e)" \
    "Lendo 2 arquivos" \
    ./build/ollama-code ask "Lê go.mod e main.go" --mode auto

run_test "BUG11-3" \
    "Multi-file read + analysis" \
    "Analisando arquivos" \
    ./build/ollama-code ask "Lê go.mod e main.go e me diz a relação" --mode auto

# BUG #13
run_test "BUG13-1" \
    "Location hint (Go file)" \
    "Dica de organização" \
    ./build/ollama-code ask "Cria arquivo tools.go" --mode auto
rm -f tools.go 2>/dev/null || true

run_test "BUG13-2" \
    "Location hint (main.go)" \
    "cmd/" \
    ./build/ollama-code ask "Cria main.go" --mode auto
rm -f main.go 2>/dev/null || true

run_test "BUG13-3" \
    "Location hint (test file)" \
    "internal/" \
    ./build/ollama-code ask "Cria utils_test.go" --mode auto
rm -f utils_test.go 2>/dev/null || true

# ================================================
# FUNCIONALIDADES BÁSICAS
# ================================================
echo "================================================"
echo "  PARTE 3: FUNCIONALIDADES BÁSICAS"
echo "================================================"
echo ""

run_test "BASIC-1" \
    "File read" \
    "go 1." \
    ./build/ollama-code ask "Lê go.mod" --mode auto

run_test "BASIC-2" \
    "Code search" \
    "package main" \
    ./build/ollama-code ask "Busca package main" --mode auto

# ================================================
# RESULTADOS FINAIS
# ================================================
echo "================================================"
echo "  RESULTADOS FINAIS"
echo "================================================"
echo ""
echo "Total de testes: $TOTAL"
PASS_PCT=$(awk "BEGIN {printf \"%.1f\", ($PASSED*100/$TOTAL)}")
FAIL_PCT=$(awk "BEGIN {printf \"%.1f\", ($FAILED*100/$TOTAL)}")
echo ""
echo "✅ Passou: $PASSED (${PASS_PCT}%)"
echo "❌ Falhou: $FAILED (${FAIL_PCT}%)"
echo ""

# Breakdown por categoria
echo "Breakdown por categoria:"
echo "  Regressão (BUG #1, #4, #6, #9, #12): 6 testes"
echo "  Novos (BUG #7, #8, #11, #13): 11 testes"
echo "  Básicos: 2 testes"
echo "  Total: 19 testes"
echo ""

if [ $FAILED -eq 0 ]; then
    echo "🎉 TODOS OS TESTES PASSARAM!"
    echo ""
    echo "Taxa de Sucesso: 100%"
    echo "Bugs Corrigidos: 10/14 (71.4%)"
    echo "Bugs Pendentes: 4/14 (28.6%)"
    exit 0
else
    echo "⚠️  ALGUNS TESTES FALHARAM - INVESTIGAR"
    exit 1
fi
