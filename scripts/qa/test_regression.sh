#!/bin/bash

# Bateria de Testes de Regressão
# Data: 2024-12-21
# Objetivo: Validar que bugs anteriormente corrigidos ainda passam

set -e

TOTAL=0
PASSED=0
FAILED=0

echo "================================================"
echo "  TESTES DE REGRESSÃO"
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
    echo ""

    # Executar comando
    result=$(echo "s" | timeout 30 "${command[@]}" 2>&1 || true)

    # Verificar se contém padrão esperado
    if echo "$result" | grep -q "$expected_pattern"; then
        echo "✅ PASS"
        PASSED=$((PASSED + 1))
    else
        echo "❌ FAIL"
        FAILED=$((FAILED + 1))
        echo ""
        echo "Esperado pattern: $expected_pattern"
        echo "Output obtido (primeiras 400 chars):"
        echo "$result" | head -c 400
    fi
    echo ""
}

# ================================================
# BUG #1: Multi-file Creation
# ================================================
echo "=== BUG #1: Multi-file Creation ==="
echo ""

run_test "REG-BUG1" \
    "Multi-file creation deve funcionar" \
    "Projeto criado" \
    ./build/ollama-code ask "Cria arquivos teste1.txt e teste2.txt" --mode auto

# Limpar
rm -f teste1.txt teste2.txt 2>/dev/null || true

# ================================================
# BUG #4: JSON Extraction
# ================================================
echo "=== BUG #4: LLM Text vs JSON ==="
echo ""

run_test "REG-BUG4" \
    "Deve criar arquivo com nome válido" \
    "Arquivo criado/atualizado" \
    ./build/ollama-code ask "Cria um arquivo chamado utils.go" --mode auto

# Limpar
rm -f utils.go 2>/dev/null || true

# ================================================
# BUG #6: File Overwrite Protection
# ================================================
echo "=== BUG #6: File Overwrite Protection ==="
echo ""

# Criar arquivo existente primeiro
echo "conteúdo original" > existing_file.txt

run_test "REG-BUG6" \
    "Deve detectar edit ao invés de overwrite" \
    "Editando arquivo existente" \
    ./build/ollama-code ask "Atualiza existing_file.txt com novo conteúdo" --mode auto

# Limpar
rm -f existing_file.txt 2>/dev/null || true

# ================================================
# BUG #9: Dotfiles
# ================================================
echo "=== BUG #9: Dotfiles Should Work ==="
echo ""

run_test "REG-BUG9-1" \
    "Deve criar .env" \
    "Arquivo criado/atualizado" \
    ./build/ollama-code ask "Cria um arquivo .env" --mode auto

run_test "REG-BUG9-2" \
    "Deve criar .gitignore" \
    "Arquivo criado/atualizado" \
    ./build/ollama-code ask "Cria um arquivo .gitignore" --mode auto

# Limpar
rm -f .env .gitignore 2>/dev/null || true

# ================================================
# BUG #12: Keyword "corrige"
# ================================================
echo "=== BUG #12: Keyword 'corrige' Detection ==="
echo ""

# Criar arquivo existente
echo "codigo com bug" > buggy_file.go

run_test "REG-BUG12" \
    "Deve detectar edit com palavra 'corrige'" \
    "Editando arquivo existente" \
    ./build/ollama-code ask "Corrige o arquivo buggy_file.go" --mode auto

# Limpar
rm -f buggy_file.go 2>/dev/null || true

# ================================================
# Testes Básicos de Funcionalidade
# ================================================
echo "=== Testes Básicos ==="
echo ""

run_test "BASIC-READ" \
    "Leitura de arquivo deve funcionar" \
    "go 1." \
    ./build/ollama-code ask "Lê o arquivo go.mod" --mode auto

run_test "BASIC-SEARCH" \
    "Busca de código deve funcionar" \
    "package main" \
    ./build/ollama-code ask "Busca por 'package main' no código" --mode auto

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
echo "Passou: $PASSED (${PASS_PCT}%)"
echo "Falhou: $FAILED (${FAIL_PCT}%)"
echo ""

if [ $FAILED -eq 0 ]; then
    echo "🎉 NENHUMA REGRESSÃO DETECTADA!"
    exit 0
else
    echo "⚠️  REGRESSÕES DETECTADAS - INVESTIGAR"
    exit 1
fi
