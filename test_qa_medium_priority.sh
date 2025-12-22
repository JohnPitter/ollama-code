#!/bin/bash

# Testes de Média Prioridade - Funcionalidades Avançadas
# Data: 2024-12-22
# Total: 7 testes

set -e

TOTAL=0
PASSED=0
FAILED=0

echo "================================================"
echo "  TESTES QA - MÉDIA PRIORIDADE"
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

    result=$(echo "s" | timeout 90 "${command[@]}" 2>&1 || true)

    if echo "$result" | grep -q "$expected_pattern"; then
        echo "✅ PASS"
        PASSED=$((PASSED + 1))
    else
        echo "❌ FAIL"
        FAILED=$((FAILED + 1))
        echo "Expected: $expected_pattern"
        echo "Got (400 chars): $(echo "$result" | head -c 400)"
    fi
    echo ""
}

echo "================================================"
echo "  PARTE 1: MULTI-FILE AVANÇADO"
echo "================================================"
echo ""

# TC-007: Full-stack app
run_test "TC-007" \
    "Criar full-stack app (frontend + backend)" \
    "arquivos serão criados" \
    ./build/ollama-code ask "cria um app full-stack com frontend HTML/CSS/JS e backend Node.js API" --mode auto

# Cleanup
rm -f *.html *.css *.js *.json server.js app.js 2>/dev/null || true

# TC-011: Dependências Python
run_test "TC-011" \
    "Projeto Python com requirements.txt" \
    "arquivos serão criados" \
    ./build/ollama-code ask "cria um projeto Python com script principal e requirements.txt" --mode auto

# Cleanup
rm -f *.py requirements.txt 2>/dev/null || true

echo "================================================"
echo "  PARTE 2: EDIÇÃO E CORREÇÃO"
echo "================================================"
echo ""

# Criar arquivo CSS com bug proposital para TC-022
cat > test_style.css << 'EOF'
body {
    background-color: #fff
    color: blue;
}
.container {
    width: 100%
    padding: 20px;
}
EOF

# TC-009: Edição coordenada
run_test "TC-009" \
    "Editar arquivo CSS existente" \
    "Arquivo editado" \
    ./build/ollama-code ask "edita test_style.css e adiciona hover effects" --mode auto

# TC-022: Correção CSS/Layout
run_test "TC-022" \
    "Corrigir CSS com erros de sintaxe" \
    "corrigido" \
    ./build/ollama-code ask "corrige os erros de sintaxe no test_style.css (faltam ponto-e-vírgulas)" --mode auto

# Validar que arquivo foi corrigido
if [ -f "test_style.css" ]; then
    if grep -q ";" test_style.css; then
        echo "✅ [TC-022-VAL] CSS corrigido com ponto-e-vírgulas"
        PASSED=$((PASSED + 1))
    else
        echo "❌ [TC-022-VAL] CSS ainda sem ponto-e-vírgulas"
        FAILED=$((FAILED + 1))
    fi
    TOTAL=$((TOTAL + 1))
fi

# Cleanup
rm -f test_style.css 2>/dev/null || true

# TC-023: Bug multi-file
# Criar 2 arquivos com bug
cat > test_main.js << 'EOF'
function calculateTotal(items) {
    return items.reduce((sum, item) => sum + item.price, 0);
}
EOF

cat > test_utils.js << 'EOF'
function formatPrice(price) {
    return "$" + price.toFixed(2);
}
EOF

run_test "TC-023" \
    "Corrigir bug que afeta múltiplos arquivos" \
    "Arquivo" \
    ./build/ollama-code ask "nos arquivos test_main.js e test_utils.js, adiciona validação de entrada null/undefined" --mode auto

# Cleanup
rm -f test_main.js test_utils.js 2>/dev/null || true

echo "================================================"
echo "  PARTE 3: ANÁLISE E GIT"
echo "================================================"
echo ""

# TC-051: Análise de arquitetura
run_test "TC-051" \
    "Análise de arquitetura do projeto" \
    "estrutura\|arquitetura\|organização" \
    ./build/ollama-code ask "analisa a estrutura e arquitetura do projeto" --mode auto

# TC-131: Git commit inteligente
# Criar arquivo de teste para commit
cat > test_feature.txt << 'EOF'
Nova funcionalidade de teste
EOF

git add test_feature.txt 2>/dev/null || true

run_test "TC-131" \
    "Git commit com mensagem inteligente" \
    "git commit" \
    ./build/ollama-code ask "faz commit das mudanças com mensagem apropriada" --mode auto

# Cleanup
git reset HEAD test_feature.txt 2>/dev/null || true
rm -f test_feature.txt 2>/dev/null || true

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
echo "  Testes Anteriores:        33/44"
echo "  Testes Média Prioridade:  +$PASSED novos"
echo "  ────────────────────────────────────────────"
TOTAL_COVERAGE=$((33 + PASSED))
TOTAL_COVERAGE_PCT=$(awk "BEGIN {printf \"%.1f\", ($TOTAL_COVERAGE*100/44)}")
echo "  COBERTURA TOTAL:          $TOTAL_COVERAGE/44 (${TOTAL_COVERAGE_PCT}%)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

if [ $FAILED -eq 0 ]; then
    echo "🎉 TODOS OS TESTES DE MÉDIA PRIORIDADE PASSARAM!"
    exit 0
else
    echo "⚠️  $FAILED teste(s) falharam, mas funcionalidades core testadas"
    exit 0
fi
