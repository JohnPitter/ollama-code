# Changelog

Todas as mudanças notáveis neste projeto serão documentadas neste arquivo.

O formato é baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.0.0/),
e este projeto adere ao [Semantic Versioning](https://semver.org/lang/pt-BR/).

## [Unreleased]

### Added

#### Automated Regression Tests
- **Adicionado:** Suite de testes de regressão automatizados em `scripts/test_regression.sh`
- **Cobertura:** 6 testes validando os 3 bugs críticos corrigidos:
  - **Test 1-2:** Bug #1 (Read-Only Mode) - Valida bloqueio e integridade do arquivo
  - **Test 3:** Bug #2 (Code Search) - Valida ausência do erro "query parameter required"
  - **Test 4-5:** Bug #3/4 (Multi-File) - Valida criação de 2 e 3 arquivos
  - **Test 6:** Linkagem entre arquivos HTML/CSS/JS
- **Execução:** `cd scripts && bash test_regression.sh`
- **Resultado:** 🎉 100% de sucesso (6/6 testes passaram)
- **Impacto:** Previne regressões futuras dos bugs críticos
- **Arquivos adicionados:**
  - `scripts/test_regression.sh` - Suite de testes E2E automatizados

#### Performance Monitoring and Troubleshooting Documentation
- **Adicionado:** Seção completa "Performance and Troubleshooting" no `CLAUDE.md`
- **Conteúdo:**
  - **GPU Overload e CPU Fallback:** 3 soluções para forçar Ollama a usar CPU quando GPU está sobrecarregada
    1. Forçar modo CPU com `CUDA_VISIBLE_DEVICES=""`
    2. Usar modelos mais leves (1.5b, 0.5b ao invés de 7b)
    3. Limitar memória GPU com variáveis de ambiente
  - **Performance Monitoring:** Documentação do sistema de Observability existente
  - **Common Issues:** 3 problemas comuns com causas e soluções:
    - Slow LLM Responses (>30s)
    - Timeouts or Hangs
    - High Memory Usage
  - **Benchmarking:** Tabela com tempos esperados para cada operação
- **Impacto:** Usuários agora sabem como lidar com problemas de performance e GPU
- **Limitação Arquitetural:** Ollama Code é um **client** e não controla GPU/CPU diretamente - isso é gerenciado pelo Ollama server
- **Arquivos alterados:**
  - `CLAUDE.md` - Adicionada seção "Performance and Troubleshooting" (linhas 98-227)

### Fixed

#### Bug #1: Modo Read-Only Não Bloqueava Escritas 🔴 CRÍTICO
- **Problema:** O modo `--mode readonly` permitia modificações de arquivos, violando a política de segurança
- **Impacto:** Usuários podiam inadvertidamente modificar arquivos quando esperavam apenas leitura
- **Causa:** Faltava verificação de `AllowsWrites()` no início do `FileWriteHandler.Handle()`
- **Correção:** Adicionada verificação de read-only no início do método `Handle()` em `internal/handlers/file_write_handler.go`:
  ```go
  if !deps.Mode.AllowsWrites() {
      return "❌ Operação bloqueada: modo somente leitura (read-only)..."
  }
  ```
- **Arquivos alterados:**
  - `internal/handlers/file_write_handler.go` - Adicionada verificação de read-only (linhas 32-38)
  - `internal/handlers/handler.go` - Adicionado método `AllowsWrites()` à interface `OperationMode`
  - `internal/handlers/adapters.go` - Implementado `AllowsWrites()` no `OperationModeAdapter`
- **Testes afetados:** TC-080

#### Bug #2: Ferramenta de Busca em Código Quebrada 🔴 CRÍTICO
- **Problema:** Erro "query parameter required" ao buscar código, impossibilitando uso da funcionalidade
- **Impacto:** Funcionalidade de busca em código completamente indisponível
- **Causa:** Intent detector nem sempre populava o parâmetro `query`
- **Correção:** Adicionada função de fallback `extractQueryFromMessage()` em `internal/handlers/search_handler.go`:
  ```go
  query, ok := result.Parameters["query"].(string)
  if !ok || query == "" {
      query = extractQueryFromMessage(result.UserMessage)
  }
  ```
- **Funcionalidade:** Extrai query de padrões comuns em português e inglês:
  - "busca a função X", "procure por X", "encontre X"
  - "search for X", "find X"
- **Arquivos alterados:**
  - `internal/handlers/search_handler.go` - Adicionada função `extractQueryFromMessage()` (linhas 58-91)
  - `internal/handlers/search_handler.go` - Adicionado import `strings`
- **Testes afetados:** TC-040, TC-041

#### Bug #3 & #4: Multi-File Creation Não Funcionava 🟡 ALTO
- **Problema #3:** Projetos complexos falhavam com erro "caminho do arquivo não especificado"
- **Problema #4:** Solicitar múltiplos arquivos (HTML, CSS, JS) criava apenas 1 arquivo
- **Impacto:** Impossível criar projetos estruturados com arquivos separados
- **Causa:** **REGRESSÃO** - Funcionalidade implementada em 19/12/2024 foi perdida durante refatoração para Handler Pattern
- **Correção:** Re-implementadas 3 funções em `internal/handlers/file_write_handler.go`:
  1. **`detectMultiFileRequest()`** - Detecta 20+ palavras-chave indicando multi-file:
     - "separados", "múltiplos arquivos", "HTML, CSS e JS"
     - "projeto completo", "full-stack", "frontend e backend"
  2. **`handleMultiFileWrite()`** - Cria múltiplos arquivos coordenados:
     - Gera prompt específico para LLM retornar JSON com array de arquivos
     - Parseia resposta e cria cada arquivo sequencialmente
     - Confirma UMA VEZ com usuário (não para cada arquivo)
     - Retorna resumo com lista de sucessos e falhas
  3. **`buildMultiFilePrompt()`** - Constrói prompt com instruções explícitas:
     - Criar TODOS os arquivos solicitados
     - Linkar arquivos corretamente (HTML → CSS/JS)
     - Usar caminhos relativos
- **Funcionalidade:** Agora suporta criação de projetos multi-file:
  - ✅ HTML + CSS + JavaScript separados
  - ✅ Linkagem automática entre arquivos
  - ✅ Estrutura profissional de projeto
- **Arquivos alterados:**
  - `internal/handlers/file_write_handler.go` - Adicionadas 3 funções (linhas 45-48, 208-426)
- **Testes afetados:** TC-004, TC-006, TC-007, TC-011

### Technical Details

#### Arquitetura
- **Handler Pattern:** Mantido com handlers individuais
- **Manual DI:** Providers não alterados
- **Observability:** Compatível com sistema existente

#### Compatibilidade
- ✅ Retrocompatível com comandos antigos
- ✅ Não quebra criação de arquivo único
- ✅ Detecção automática de modo (single vs multi-file)
- ✅ Fallbacks robustos em caso de erro

#### Performance
- **Multi-File Creation:** O(n) onde n = número de arquivos
- **Detecção de Keywords:** O(k) onde k = número de keywords (~20)
- **Impacto:** Mínimo, operações I/O dominam

### Testing

#### Testes Manuais Executados
- ✅ TC-080: Modo read-only agora bloqueia escritas corretamente
- ✅ TC-040: Busca de código funciona com query extraída da mensagem
- ✅ TC-004: Multi-file cria 3 arquivos (HTML, CSS, JS) linkados

#### Validação
```bash
# Teste read-only
./build/ollama-code.exe ask "modifica arquivo.txt" --mode readonly
# ✅ Resultado: Operação bloqueada

# Teste code search
./build/ollama-code.exe ask "busca a função ProcessMessage"
# ✅ Resultado: Busca executada com sucesso

# Teste multi-file
./build/ollama-code.exe ask "cria HTML e CSS separados" --mode autonomous
# ✅ Resultado: 3 arquivos criados (index.html, style.css, script.js)
# ✅ Linkagem: HTML tem <link> para CSS e <script> para JS
```

### Breaking Changes
Nenhuma mudança incompatível. Todas as alterações são correções de bugs mantendo compatibilidade total.

### Migration Guide
Não é necessária migração. Basta recompilar:
```bash
go build -o build/ollama-code.exe ./cmd/ollama-code
```

---

## [0.3.0] - 2024-12-22

### Added
- ✅ 100% QA Test Coverage (44/44 testes passando)
- ✅ 7 Ferramentas Avançadas (Advanced Refactoring, Background Tasks, Code Formatter, etc.)
- ✅ Observability System (Logging, Metrics, Tracing)
- ✅ Manual Dependency Injection
- ✅ Handler Pattern (refatoração de 2282 linhas God object)

Veja `docs/QA_100_PERCENT_COVERAGE_2024-12-22.md` para detalhes.

---

## [0.2.0] - 2024-12-19

### Added
- Multi-file creation support (originalmente implementado)
- Web search com DuckDuckGo + HTML fetching
- OLLAMA.md hierarchical context system (4 níveis)

### Fixed
- 14 bugs identificados e corrigidos no QA inicial

Veja `changes/08-multi-file-creation.md` para implementação original.

---

## [0.1.0] - 2024-12-15

### Added
- Initial release
- Basic LLM integration via Ollama
- Intent detection system
- File read/write operations
- Code search functionality
- Git operations
- Interactive and autonomous modes

---

## Notas de Versão

### Convenções de Versionamento
- **MAJOR:** Mudanças incompatíveis na API
- **MINOR:** Nova funcionalidade compatível
- **PATCH:** Correções de bugs compatíveis

### Categorias de Mudanças
- **Added:** Novas funcionalidades
- **Changed:** Alterações em funcionalidades existentes
- **Deprecated:** Funcionalidades que serão removidas
- **Removed:** Funcionalidades removidas
- **Fixed:** Correções de bugs
- **Security:** Correções de vulnerabilidades

### Prioridades de Bugs
- 🔴 **CRÍTICO:** Violação de segurança, funcionalidade quebrada
- 🟡 **ALTO:** Funcionalidade importante não funciona
- 🟢 **MÉDIO:** Comportamento incorreto mas tem workaround
- 🔵 **BAIXO:** Problema menor, cosmético

---

**Responsável pelas correções:** Claude Code
**Data das correções:** 30 de Dezembro de 2024
**Build testado:** ollama-code.exe (Windows 11, Go 1.21+)
**Modelo LLM:** qwen2.5-coder:7b via Ollama 0.13.5
