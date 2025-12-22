#!/bin/bash

# Testes para BUGs #10 e #14
# Data: 2024-12-21
# Objetivo: Validar correções de intent detection e cleanCodeContent

set -e

TOTAL=0
PASSED=0
FAILED=0

echo "================================================"
echo "  TESTES DOS BUGS #10 E #14"
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
        echo "Output (500 chars): $(echo "$result" | head -c 500)"
    fi
    echo ""
}

# ================================================
# BUG #10: Intent Detection
# ================================================
echo "=== BUG #10: Intent Detection (Análise/Review/Refatoração) ==="
echo ""

# Criar arquivo de teste
cat > test_analysis.go <<'EOF'
package main

import "fmt"

func main() {
    fmt.Println("Hello World")
}
EOF

run_test "BUG10-1" \
    "Análise de arquivo deve usar análise (não edição)" \
    "Analisando código" \
    ./build/ollama-code ask "Analisa o arquivo test_analysis.go" --mode auto

run_test "BUG10-2" \
    "Review deve usar análise" \
    "Analisando código" \
    ./build/ollama-code ask "Faz review do test_analysis.go" --mode auto

run_test "BUG10-3" \
    "Explicação deve usar análise" \
    "Analisando código" \
    ./build/ollama-code ask "Explica o test_analysis.go" --mode auto

run_test "BUG10-4" \
    "Refatoração deve usar edição (não análise)" \
    "Editando arquivo existente" \
    ./build/ollama-code ask "Refatora o test_analysis.go" --mode auto

# Limpar
rm -f test_analysis.go

# ================================================
# BUG #14: cleanCodeContent() Remove Chaves JSONs
# ================================================
echo "=== BUG #14: cleanCodeContent() Deve Preservar JSONs ==="
echo ""

run_test "BUG14-1" \
    "Criação de package.json deve preservar chaves" \
    "Arquivo criado/atualizado" \
    ./build/ollama-code ask "Cria package.json com name test-app" --mode auto

# Verificar se package.json foi criado corretamente
if [ -f "package.json" ]; then
    # Tentar fazer parse do JSON
    if python -m json.tool package.json > /dev/null 2>&1; then
        echo "✅ package.json é JSON válido"
        PASSED=$((PASSED + 1))
    else
        echo "❌ package.json NÃO é JSON válido"
        FAILED=$((FAILED + 1))
        echo "Conteúdo:"
        cat package.json
    fi
    TOTAL=$((TOTAL + 1))
    rm -f package.json
else
    echo "❌ package.json não foi criado"
    FAILED=$((FAILED + 1))
    TOTAL=$((TOTAL + 1))
fi

run_test "BUG14-2" \
    "Criação de tsconfig.json deve preservar chaves" \
    "Arquivo criado/atualizado" \
    ./build/ollama-code ask "Cria tsconfig.json com compilerOptions" --mode auto

if [ -f "tsconfig.json" ]; then
    if python -m json.tool tsconfig.json > /dev/null 2>&1; then
        echo "✅ tsconfig.json é JSON válido"
        PASSED=$((PASSED + 1))
    else
        echo "❌ tsconfig.json NÃO é JSON válido"
        FAILED=$((FAILED + 1))
    fi
    TOTAL=$((TOTAL + 1))
    rm -f tsconfig.json
else
    echo "❌ tsconfig.json não foi criado"
    FAILED=$((FAILED + 1))
    TOTAL=$((TOTAL + 1))
fi

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

echo "Breakdown:"
echo "  BUG #10 (Intent Detection): 4 testes"
echo "  BUG #14 (JSON Preservation): 4 testes"
echo "  Total: 8 testes"
echo ""

if [ $FAILED -eq 0 ]; then
    echo "🎉 TODOS OS TESTES PASSARAM!"
    echo ""
    echo "BUG #10: ✅ CORRIGIDO"
    echo "BUG #14: ✅ CORRIGIDO"
    exit 0
else
    echo "⚠️  ALGUNS TESTES FALHARAM"
    exit 1
fi
