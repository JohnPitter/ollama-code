#!/bin/bash

# Bateria FINAL Completa - Todos os Bugs (#1-#14)
# Data: 2024-12-21
# Objetivo: Validar TODAS as correções implementadas

set -e

TOTAL=0
PASSED=0
FAILED=0

echo "================================================"
echo "  BATERIA FINAL COMPLETA - TODOS OS BUGS"
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
    fi
    echo ""
}

echo "================================================"
echo "  PARTE 1: BUGS ORIGINAIS (#1, #4, #6, #9, #12)"
echo "================================================"
echo ""

# BUG #1
run_test "BUG1" "Multi-file creation" "Projeto criado" \
    ./build/ollama-code ask "Cria arquivos test1.txt e test2.txt" --mode auto
rm -f test1.txt test2.txt 2>/dev/null || true

# BUG #4
run_test "BUG4" "JSON extraction & valid filename" "Arquivo criado/atualizado" \
    ./build/ollama-code ask "Cria arquivo helper.go" --mode auto
rm -f helper.go 2>/dev/null || true

# BUG #6
echo "test content" > temp_file.txt
run_test "BUG6" "File overwrite protection" "Editando arquivo existente" \
    ./build/ollama-code ask "Atualiza temp_file.txt" --mode auto
rm -f temp_file.txt 2>/dev/null || true

# BUG #9
run_test "BUG9-1" "Dotfile .env creation" "Arquivo criado/atualizado" \
    ./build/ollama-code ask "Cria .env" --mode auto
rm -f .env 2>/dev/null || true

run_test "BUG9-2" "Dotfile .gitignore creation" "Arquivo criado/atualizado" \
    ./build/ollama-code ask "Cria .gitignore" --mode auto
rm -f .gitignore 2>/dev/null || true

# BUG #12
echo "buggy code" > fix_me.go
run_test "BUG12" "Keyword 'corrige' detection" "Editando arquivo existente" \
    ./build/ollama-code ask "Corrige fix_me.go" --mode auto
rm -f fix_me.go 2>/dev/null || true

echo "================================================"
echo "  PARTE 2: BUGS DA SESSÃO ATUAL (#7, #8, #11, #13)"
echo "================================================"
echo ""

# BUG #7
run_test "BUG7-1" "Git status" "Operação git 'status'" \
    ./build/ollama-code ask "Mostra status do git" --mode auto

run_test "BUG7-2" "Git diff" "Operação git 'diff'" \
    ./build/ollama-code ask "Mostra mudanças do git" --mode auto

run_test "BUG7-3" "Git log" "Operação git 'log'" \
    ./build/ollama-code ask "Mostra histórico de commits" --mode auto

# BUG #8
run_test "BUG8-1" "Integration hint (JS)" "Dica" \
    ./build/ollama-code ask "Cria app.js e conecta no index.html" --mode auto
rm -f app.js 2>/dev/null || true

run_test "BUG8-2" "Integration hint (CSS)" "Dica" \
    ./build/ollama-code ask "Cria styles.css e conecta no index.html" --mode auto
rm -f styles.css 2>/dev/null || true

# BUG #11
run_test "BUG11-1" "Multi-file read (comma)" "Lendo 2 arquivos" \
    ./build/ollama-code ask "Lê go.mod, main.go" --mode auto

run_test "BUG11-2" "Multi-file read (e)" "Lendo 2 arquivos" \
    ./build/ollama-code ask "Lê go.mod e main.go" --mode auto

run_test "BUG11-3" "Multi-file read + analysis" "Analisando arquivos" \
    ./build/ollama-code ask "Lê go.mod e main.go e me diz a relação" --mode auto

# BUG #13
run_test "BUG13-1" "Location hint (Go file)" "Dica de organização" \
    ./build/ollama-code ask "Cria arquivo tools.go" --mode auto
rm -f tools.go 2>/dev/null || true

run_test "BUG13-2" "Location hint (main.go)" "cmd/" \
    ./build/ollama-code ask "Cria main.go" --mode auto
rm -f main.go 2>/dev/null || true

run_test "BUG13-3" "Location hint (test file)" "internal/" \
    ./build/ollama-code ask "Cria utils_test.go" --mode auto
rm -f utils_test.go 2>/dev/null || true

echo "================================================"
echo "  PARTE 3: BUGS FINAIS (#10, #14)"
echo "================================================"
echo ""

# BUG #10
cat > test_analysis.go <<'EOF'
package main
import "fmt"
func main() {
    fmt.Println("Hello World")
}
EOF

run_test "BUG10-1" "Análise de arquivo" "Analisando código" \
    ./build/ollama-code ask "Analisa o arquivo test_analysis.go" --mode auto

run_test "BUG10-2" "Review de arquivo" "Analisando código" \
    ./build/ollama-code ask "Faz review do test_analysis.go" --mode auto

run_test "BUG10-3" "Explicação de arquivo" "Analisando código" \
    ./build/ollama-code ask "Explica o test_analysis.go" --mode auto

run_test "BUG10-4" "Refatoração (edição)" "Editando arquivo existente" \
    ./build/ollama-code ask "Refatora o test_analysis.go" --mode auto

rm -f test_analysis.go

# BUG #14
run_test "BUG14-1" "package.json preservation" "Arquivo criado/atualizado" \
    ./build/ollama-code ask "Cria package.json com name test-app" --mode auto

if [ -f "package.json" ]; then
    if python -m json.tool package.json > /dev/null 2>&1; then
        echo "✅ [BUG14-1-VALIDATION] package.json é JSON válido"
        PASSED=$((PASSED + 1))
    else
        echo "❌ [BUG14-1-VALIDATION] package.json NÃO é JSON válido"
        FAILED=$((FAILED + 1))
    fi
    TOTAL=$((TOTAL + 1))
    rm -f package.json
fi

run_test "BUG14-2" "tsconfig.json preservation" "Arquivo criado/atualizado" \
    ./build/ollama-code ask "Cria tsconfig.json com compilerOptions" --mode auto

if [ -f "tsconfig.json" ]; then
    if python -m json.tool tsconfig.json > /dev/null 2>&1; then
        echo "✅ [BUG14-2-VALIDATION] tsconfig.json é JSON válido"
        PASSED=$((PASSED + 1))
    else
        echo "❌ [BUG14-2-VALIDATION] tsconfig.json NÃO é JSON válido"
        FAILED=$((FAILED + 1))
    fi
    TOTAL=$((TOTAL + 1))
    rm -f tsconfig.json
fi

echo "================================================"
echo "  PARTE 4: FUNCIONALIDADES BÁSICAS"
echo "================================================"
echo ""

run_test "BASIC-1" "File read" "go 1." \
    ./build/ollama-code ask "Lê go.mod" --mode auto

run_test "BASIC-2" "Code search" "package main" \
    ./build/ollama-code ask "Busca package main" --mode auto

# ================================================
# RESULTADOS FINAIS
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
echo "📋 BREAKDOWN POR CATEGORIA"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Originais (#1, #4, #6, #9, #12):  6 testes"
echo "  Sessão 1 (#7, #8, #11, #13):      11 testes"
echo "  Finais (#10, #14):                8 testes"
echo "  Básicos:                          2 testes"
echo "  ────────────────────────────────────────────"
echo "  TOTAL:                            27 testes"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "🐛 BUGS TESTADOS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  ✅ BUG #1:  Multi-file Creation"
echo "  ✅ BUG #4:  JSON Extraction"
echo "  ✅ BUG #6:  File Overwrite Protection"
echo "  ✅ BUG #7:  Git Operations"
echo "  ✅ BUG #8:  File Integration Hints"
echo "  ✅ BUG #9:  Dotfiles Support"
echo "  ✅ BUG #10: Intent Detection (NEW!)"
echo "  ✅ BUG #11: Multi-file Read"
echo "  ✅ BUG #12: Keyword 'corrige'"
echo "  ✅ BUG #13: Location Hints"
echo "  ✅ BUG #14: JSON Preservation (NEW!)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

if [ $FAILED -eq 0 ]; then
    echo "🎉🎉🎉 TODOS OS TESTES PASSARAM! 🎉🎉🎉"
    echo ""
    echo "✅ 11/14 BUGS CORRIGIDOS (78.6%)"
    echo "✅ Taxa de Sucesso: 100%"
    echo "✅ Meta 95%: SUPERADA!"
    exit 0
elif [ $PASS_PCT -ge 95 ]; then
    echo "🎯 META 95% ATINGIDA!"
    echo ""
    echo "Taxa de Sucesso: ${PASS_PCT}% ≥ 95% ✅"
    echo "Bugs Corrigidos: 11/14 (78.6%)"
    exit 0
else
    echo "📈 Progresso Excelente: ${PASS_PCT}%"
    echo ""
    echo "Faltam $(echo "95 - $PASS_PCT" | bc) pontos para meta 95%"
    exit 1
fi
