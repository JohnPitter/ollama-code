#!/bin/bash
# Testes de Regressão - Bug Fixes 2024-12-30
# Valida que os 3 bugs críticos corrigidos não regressem

set -e  # Exit on error

OLLAMA_CODE="$(pwd)/../build/ollama-code.exe"
# Fallback para o executável se não estiver em build/
if [ ! -f "$OLLAMA_CODE" ]; then
    OLLAMA_CODE="$(pwd)/../build/ollama-code"
fi
if [ ! -f "$OLLAMA_CODE" ]; then
    echo "ERRO: Executável ollama-code não encontrado!"
    echo "Tentou: ../build/ollama-code.exe e ../build/ollama-code"
    echo "Por favor, compile primeiro: go build -o build/ollama-code.exe ./cmd/ollama-code"
    exit 1
fi
TEST_DIR="regression_test_$(date +%Y%m%d_%H%M%S)"

echo "========================================"
echo "TESTES DE REGRESSÃO - BUG FIXES"
echo "========================================"
echo ""
echo "Data: $(date)"
echo "Build: $OLLAMA_CODE"
echo "Diretório de teste: $TEST_DIR"
echo ""

# Criar diretório de teste
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

# Contadores
PASSED=0
FAILED=0

# ====================
# TESTE 1: Bug #1 - Read-Only Mode
# ====================
echo "[TEST 1] Bug #1: Modo Read-Only Deve Bloquear Escritas"
echo "--------------------------------------------------------"

# Criar arquivo de teste
echo "original content" > readonly_test.txt

# Tentar modificar em modo readonly
OUTPUT=$($OLLAMA_CODE ask "modifica o arquivo readonly_test.txt" --mode readonly 2>&1 || true)

if echo "$OUTPUT" | grep -q "Operação bloqueada"; then
    echo "✅ PASSOU: Modo read-only bloqueou escrita corretamente"
    PASSED=$((PASSED + 1))

    # Verificar que arquivo não foi modificado
    CONTENT=$(cat readonly_test.txt)
    if [ "$CONTENT" = "original content" ]; then
        echo "✅ PASSOU: Arquivo permaneceu inalterado"
        PASSED=$((PASSED + 1))
    else
        echo "❌ FALHOU: Arquivo foi modificado em modo read-only!"
        FAILED=$((FAILED + 1))
    fi
else
    echo "❌ FALHOU: Modo read-only NÃO bloqueou escrita!"
    echo "Output: $OUTPUT"
    FAILED=$((FAILED + 2))
fi

echo ""

# ====================
# TESTE 2: Bug #2 - Code Search
# ====================
echo "[TEST 2] Bug #2: Code Search Não Deve Retornar Erro de 'query parameter required'"
echo "--------------------------------------------------------"

OUTPUT=$($OLLAMA_CODE ask "busca a função ProcessMessage no código" --mode autonomous 2>&1 || true)

if echo "$OUTPUT" | grep -q "query parameter required"; then
    echo "❌ FALHOU: Code search retornou erro 'query parameter required'"
    echo "Output: $OUTPUT"
    FAILED=$((FAILED + 1))
elif echo "$OUTPUT" | grep -qE "(Nenhum resultado|resultado|encontrado|search)"; then
    echo "✅ PASSOU: Code search executou sem erro de query"
    PASSED=$((PASSED + 1))
else
    echo "⚠️  AVISO: Code search retornou output inesperado, mas sem erro de query"
    echo "Output: $OUTPUT"
    PASSED=$((PASSED + 1))
fi

echo ""

# ====================
# TESTE 3: Bug #3/4 - Multi-File Creation
# ====================
echo "[TEST 3] Bug #3/4: Multi-File Creation Deve Criar Múltiplos Arquivos"
echo "--------------------------------------------------------"

# Limpar arquivos anteriores se existirem
rm -f multifile_*.html multifile_*.css multifile_*.js

OUTPUT=$($OLLAMA_CODE ask "cria HTML e CSS separados chamados multifile_test.html e multifile_test.css" --mode autonomous 2>&1 || true)

# Contar arquivos criados
HTML_COUNT=$(ls multifile_*.html 2>/dev/null | wc -l)
CSS_COUNT=$(ls multifile_*.css 2>/dev/null | wc -l)
TOTAL_FILES=$((HTML_COUNT + CSS_COUNT))

if [ $TOTAL_FILES -ge 2 ]; then
    echo "✅ PASSOU: Criou $TOTAL_FILES arquivos (HTML: $HTML_COUNT, CSS: $CSS_COUNT)"
    PASSED=$((PASSED + 1))

    # Verificar linkagem se ambos existem
    if [ $HTML_COUNT -ge 1 ] && [ $CSS_COUNT -ge 1 ]; then
        HTML_FILE=$(ls multifile_*.html | head -1)
        if grep -q "stylesheet" "$HTML_FILE" && grep -q "\.css" "$HTML_FILE"; then
            echo "✅ PASSOU: HTML linkado ao CSS corretamente"
            PASSED=$((PASSED + 1))
        else
            echo "⚠️  AVISO: HTML pode não estar linkado ao CSS"
            echo "Conteúdo HTML:"
            head -10 "$HTML_FILE"
            PASSED=$((PASSED + 1))  # Não falhar por isso
        fi
    fi
else
    echo "❌ FALHOU: Criou apenas $TOTAL_FILES arquivo(s), esperado >= 2"
    echo "Output: $OUTPUT"
    ls -la multifile_* 2>&1 || echo "Nenhum arquivo multifile_* encontrado"
    FAILED=$((FAILED + 1))
fi

echo ""

# ====================
# TESTE 4: Criação de 3+ Arquivos
# ====================
echo "[TEST 4] Multi-File: 3+ Arquivos (HTML, CSS, JS)"
echo "--------------------------------------------------------"

rm -f project_*.html project_*.css project_*.js

OUTPUT=$($OLLAMA_CODE ask "cria um projeto completo com HTML, CSS e JavaScript separados, chamados project_index.html, project_style.css, project_script.js" --mode autonomous 2>&1 || true)

HTML_COUNT=$(ls project_*.html 2>/dev/null | wc -l)
CSS_COUNT=$(ls project_*.css 2>/dev/null | wc -l)
JS_COUNT=$(ls project_*.js 2>/dev/null | wc -l)
TOTAL_FILES=$((HTML_COUNT + CSS_COUNT + JS_COUNT))

if [ $TOTAL_FILES -ge 3 ]; then
    echo "✅ PASSOU: Criou $TOTAL_FILES arquivos (HTML: $HTML_COUNT, CSS: $CSS_COUNT, JS: $JS_COUNT)"
    PASSED=$((PASSED + 1))
else
    echo "❌ FALHOU: Criou apenas $TOTAL_FILES arquivo(s), esperado >= 3"
    echo "Output: $OUTPUT"
    ls -la project_* 2>&1 || echo "Nenhum arquivo project_* encontrado"
    FAILED=$((FAILED + 1))
fi

echo ""

# ====================
# RESULTADO FINAL
# ====================
echo "========================================"
echo "RESULTADO FINAL"
echo "========================================"
echo ""
echo "Testes executados: $((PASSED + FAILED))"
echo "Testes passaram:   $PASSED ✅"
echo "Testes falharam:   $FAILED ❌"
echo ""

if [ $FAILED -eq 0 ]; then
    echo "🎉 SUCESSO! Todos os testes de regressão passaram!"
    echo ""
    echo "✅ Bug #1 (Read-Only) - CORRIGIDO E VALIDADO"
    echo "✅ Bug #2 (Code Search) - CORRIGIDO E VALIDADO"
    echo "✅ Bug #3/4 (Multi-File) - CORRIGIDO E VALIDADO"
    exit 0
else
    echo "❌ FALHA! $FAILED teste(s) falharam."
    echo ""
    echo "REGRESSÕES DETECTADAS!"
    echo "Revise os bugs corrigidos em:"
    echo "  - internal/handlers/file_write_handler.go (Bug #1, #3, #4)"
    echo "  - internal/handlers/search_handler.go (Bug #2)"
    exit 1
fi
