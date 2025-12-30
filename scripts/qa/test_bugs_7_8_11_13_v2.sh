#!/bin/bash

# Bateria de Testes para BUGs #7, #8, #11, #13
# Data: 2024-12-21
# Objetivo: Validar correções implementadas

set -e

TOTAL=0
PASSED=0
FAILED=0

echo "================================================"
echo "  TESTES DOS BUGS #7, #8, #11, #13"
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
    echo "Comando: ${command[@]}"
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
        echo "Output obtido (primeiras 500 chars):"
        echo "$result" | head -c 500
    fi
    echo ""
}

# ================================================
# BUG #7: Git Operations
# ================================================
echo "=== BUG #7: Git Operations ==="
echo ""

run_test "BUG7-T1" \
    "Git status deve funcionar" \
    "Operação git 'status'" \
    ./build/ollama-code ask "Mostra o status do git" --mode auto

run_test "BUG7-T2" \
    "Git diff deve funcionar" \
    "Operação git 'diff'" \
    ./build/ollama-code ask "Mostra as mudanças do git" --mode auto

run_test "BUG7-T3" \
    "Git log deve funcionar" \
    "Operação git 'log'" \
    ./build/ollama-code ask "Mostra o histórico de commits" --mode auto

# ================================================
# BUG #8: File Integration (Hints)
# ================================================
echo "=== BUG #8: File Integration Hints ==="
echo ""

run_test "BUG8-T1" \
    "Hint de integração para arquivo JS" \
    "Dica" \
    ./build/ollama-code ask "Cria um arquivo app.js e conecta no index.html" --mode auto

run_test "BUG8-T2" \
    "Hint de integração para arquivo CSS" \
    "Dica" \
    ./build/ollama-code ask "Cria um arquivo styles.css e conecta no index.html" --mode auto

# Limpar arquivos de teste
rm -f app.js styles.css 2>/dev/null || true

# ================================================
# BUG #11: Multi-file Read
# ================================================
echo "=== BUG #11: Multi-file Read ==="
echo ""

run_test "BUG11-T1" \
    "Ler múltiplos arquivos com vírgula" \
    "Lendo 2 arquivos" \
    ./build/ollama-code ask "Lê go.mod, main.go" --mode auto

run_test "BUG11-T2" \
    "Ler múltiplos arquivos com e" \
    "Lendo 2 arquivos" \
    ./build/ollama-code ask "Lê go.mod e main.go" --mode auto

run_test "BUG11-T3" \
    "Análise automática de múltiplos arquivos" \
    "Analisando arquivos" \
    ./build/ollama-code ask "Lê go.mod e main.go e me diz a relação entre eles" --mode auto

# ================================================
# BUG #13: Location Hints
# ================================================
echo "=== BUG #13: Location Hints ==="
echo ""

run_test "BUG13-T1" \
    "Hint para arquivo .go na raiz" \
    "Dica de organização" \
    ./build/ollama-code ask "Cria um arquivo utils.go" --mode auto

run_test "BUG13-T2" \
    "Hint para main.go sugere cmd/" \
    "cmd/" \
    ./build/ollama-code ask "Cria um arquivo main.go" --mode auto

run_test "BUG13-T3" \
    "Hint para arquivo de teste" \
    "internal/" \
    ./build/ollama-code ask "Cria um arquivo helper_test.go" --mode auto

# Limpar arquivos de teste
rm -f utils.go main.go helper_test.go 2>/dev/null || true

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
    echo "🎉 TODOS OS TESTES PASSARAM!"
    exit 0
else
    echo "⚠️  ALGUNS TESTES FALHARAM"
    exit 1
fi
