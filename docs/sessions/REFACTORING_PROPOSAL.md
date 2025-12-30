# 🔧 Proposta de Refatoração - Ollama Code

## Data: 22/12/2024

Este documento apresenta uma proposta detalhada de refatoração baseada na análise arquitetural realizada.

---

## 🎯 Objetivo

Transform

ar a arquitetura atual em uma estrutura mais **modular**, **testável** e **maintainable** sem quebrar funcionalidades existentes.

---

## 📋 Fase 1: Handler Pattern (CRÍTICO)

### Problema Atual

```go
// internal/agent/handlers.go - 2282 linhas
func (a *Agent) handleIntent(ctx context.Context, result *intent.DetectionResult) (string, error) {
    switch result.Intent {
    case intent.IntentReadFile:
        return a.handleReadFile(ctx, result)
    case intent.IntentWriteFile:
        return a.handleWriteFile(ctx, result)
    case intent.IntentExecuteCommand:
        return a.handleExecuteCommand(ctx, result)
    // ... 11+ outros cases
    default:
        return a.handleQuestion(ctx, result)
    }
}

// Cada handler tem 40-200 linhas
func (a *Agent) handleReadFile(ctx context.Context, result *intent.DetectionResult) (string, error) {
    // 106 linhas de código
}

func (a *Agent) handleWriteFile(ctx context.Context, result *intent.DetectionResult) (string, error) {
    // 191 linhas de código
}
```

### Solução Proposta

#### Passo 1: Criar Handler Interface

```go
// internal/handlers/handler.go
package handlers

import (
    "context"
    "github.com/johnpitter/ollama-code/internal/agent"
    "github.com/johnpitter/ollama-code/internal/intent"
)

// Handler interface para processar intents específicos
type Handler interface {
    // Handle processa o intent e retorna resultado
    Handle(ctx context.Context, ag *agent.Agent, result *intent.DetectionResult) (string, error)
}

// BaseHandler fornece funcionalidades comuns
type BaseHandler struct {
    name string
}

func (h *BaseHandler) Name() string {
    return h.name
}
```

#### Passo 2: Criar Handler Registry

```go
// internal/handlers/registry.go
package handlers

import (
    "context"
    "fmt"
    "sync"

    "github.com/johnpitter/ollama-code/internal/agent"
    "github.com/johnpitter/ollama-code/internal/intent"
)

// Registry gerencia handlers de intents
type Registry struct {
    handlers map[intent.Intent]Handler
    mu       sync.RWMutex
}

// NewRegistry cria novo registry
func NewRegistry() *Registry {
    return &Registry{
        handlers: make(map[intent.Intent]Handler),
    }
}

// Register registra um handler para um intent
func (r *Registry) Register(intent intent.Intent, handler Handler) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    if _, exists := r.handlers[intent]; exists {
        return fmt.Errorf("handler already registered for intent: %s", intent)
    }

    r.handlers[intent] = handler
    return nil
}

// Handle processa um intent usando o handler apropriado
func (r *Registry) Handle(ctx context.Context, ag *agent.Agent, result *intent.DetectionResult) (string, error) {
    r.mu.RLock()
    handler, exists := r.handlers[result.Intent]
    r.mu.RUnlock()

    if !exists {
        // Fallback para handler padrão (question)
        return r.handleDefault(ctx, ag, result)
    }

    return handler.Handle(ctx, ag, result)
}

func (r *Registry) handleDefault(ctx context.Context, ag *agent.Agent, result *intent.DetectionResult) (string, error) {
    // Implementação do question handler como fallback
    return fmt.Sprintf("Intent não suportado: %s", result.Intent), nil
}
```

#### Passo 3: Implementar Handlers Específicos

```go
// internal/handlers/file_read_handler.go
package handlers

import (
    "context"
    "fmt"

    "github.com/johnpitter/ollama-code/internal/agent"
    "github.com/johnpitter/ollama-code/internal/intent"
)

// FileReadHandler processa leitura de arquivos
type FileReadHandler struct {
    BaseHandler
}

// NewFileReadHandler cria novo handler
func NewFileReadHandler() *FileReadHandler {
    return &FileReadHandler{
        BaseHandler: BaseHandler{name: "file_read"},
    }
}

// Handle processa intent de leitura
func (h *FileReadHandler) Handle(ctx context.Context, ag *agent.Agent, result *intent.DetectionResult) (string, error) {
    // Extrair parâmetros
    filePath, ok := result.Parameters["file_path"].(string)
    if !ok || filePath == "" {
        return "", fmt.Errorf("file_path não especificado")
    }

    // Executar via tool registry
    params := map[string]interface{}{
        "file_path": filePath,
    }

    toolResult, err := ag.GetToolRegistry().Execute(ctx, "file_reader", params)
    if err != nil {
        return "", fmt.Errorf("erro ao ler arquivo: %w", err)
    }

    if !toolResult.Success {
        return "", fmt.Errorf("erro: %s", toolResult.Error)
    }

    return toolResult.Message, nil
}
```

```go
// internal/handlers/file_write_handler.go
package handlers

import (
    "context"
    "encoding/json"
    "fmt"
    "strings"

    "github.com/johnpitter/ollama-code/internal/agent"
    "github.com/johnpitter/ollama-code/internal/intent"
    "github.com/johnpitter/ollama-code/internal/validators"
)

// FileWriteHandler processa escrita de arquivos
type FileWriteHandler struct {
    BaseHandler
    validator *validators.FileValidator
    cleaner   *validators.CodeCleaner
}

// NewFileWriteHandler cria novo handler
func NewFileWriteHandler() *FileWriteHandler {
    return &FileWriteHandler{
        BaseHandler: BaseHandler{name: "file_write"},
        validator:   validators.NewFileValidator(),
        cleaner:     validators.NewCodeCleaner(),
    }
}

// Handle processa intent de escrita
func (h *FileWriteHandler) Handle(ctx context.Context, ag *agent.Agent, result *intent.DetectionResult) (string, error) {
    // Extrair parâmetros
    filePath, _ := result.Parameters["file_path"].(string)
    content, _ := result.Parameters["content"].(string)
    userMessage := result.UserMessage

    // Se não tem content, precisa gerar via LLM
    if content == "" {
        return h.generateAndWrite(ctx, ag, userMessage, filePath)
    }

    // Validar filename
    if !h.validator.IsValidFilename(filePath) {
        return "", fmt.Errorf("nome de arquivo inválido: %s", filePath)
    }

    // Limpar content
    content = h.cleaner.CleanCode(content, filePath)

    // Confirmar com usuário
    if ag.GetMode().RequiresConfirmation() {
        confirmed, err := ag.GetConfirmationManager().ConfirmWithPreview(
            fmt.Sprintf("Escrever arquivo %s?", filePath),
            content,
        )
        if err != nil || !confirmed {
            return "Operação cancelada", nil
        }
    }

    // Executar escrita
    params := map[string]interface{}{
        "file_path": filePath,
        "content":   content,
    }

    toolResult, err := ag.GetToolRegistry().Execute(ctx, "file_writer", params)
    if err != nil {
        return "", fmt.Errorf("erro ao escrever arquivo: %w", err)
    }

    if !toolResult.Success {
        return "", fmt.Errorf("erro: %s", toolResult.Error)
    }

    return toolResult.Message, nil
}

func (h *FileWriteHandler) generateAndWrite(ctx context.Context, ag *agent.Agent, userMessage, suggestedPath string) (string, error) {
    // Gerar conteúdo via LLM
    prompt := fmt.Sprintf("Generate file content based on: %s\nOutput JSON with 'file_path' and 'content' fields.", userMessage)

    response, err := ag.GetLLMClient().Complete(ctx, prompt)
    if err != nil {
        return "", err
    }

    // Parse JSON
    var parsed map[string]interface{}
    if err := json.Unmarshal([]byte(response), &parsed); err != nil {
        // Fallback se não for JSON válido
        return h.generateSimple(ctx, ag, userMessage, suggestedPath)
    }

    filePath := parsed["file_path"].(string)
    content := parsed["content"].(string)

    // Recursivamente chamar Handle com parâmetros completos
    result := &intent.DetectionResult{
        Intent: intent.IntentWriteFile,
        Parameters: map[string]interface{}{
            "file_path": filePath,
            "content":   content,
        },
    }

    return h.Handle(ctx, ag, result)
}

func (h *FileWriteHandler) generateSimple(ctx context.Context, ag *agent.Agent, userMessage, suggestedPath string) (string, error) {
    // Implementação de fallback simples
    return "", fmt.Errorf("não foi possível gerar conteúdo")
}
```

#### Passo 4: Refatorar Agent para Usar Handler Registry

```go
// internal/agent/agent.go (MODIFICADO)
package agent

import (
    "github.com/johnpitter/ollama-code/internal/handlers"
    // ... outros imports
)

type Agent struct {
    llmClient       *llm.Client
    intentDetector  *intent.Detector
    toolRegistry    *tools.Registry
    handlerRegistry *handlers.Registry  // NOVO
    // ... outros campos
}

func NewAgent(cfg Config) (*Agent, error) {
    // ... código existente ...

    // Criar handler registry
    handlerRegistry := handlers.NewRegistry()

    // Registrar handlers
    handlerRegistry.Register(intent.IntentReadFile, handlers.NewFileReadHandler())
    handlerRegistry.Register(intent.IntentWriteFile, handlers.NewFileWriteHandler())
    handlerRegistry.Register(intent.IntentExecuteCommand, handlers.NewExecuteHandler())
    handlerRegistry.Register(intent.IntentSearchCode, handlers.NewSearchHandler())
    handlerRegistry.Register(intent.IntentAnalyzeProject, handlers.NewAnalyzeHandler())
    handlerRegistry.Register(intent.IntentGitOperation, handlers.NewGitHandler())
    handlerRegistry.Register(intent.IntentWebSearch, handlers.NewWebSearchHandler())

    return &Agent{
        // ... campos existentes ...
        handlerRegistry: handlerRegistry,
    }, nil
}

// ProcessMessage SIMPLIFICADO
func (a *Agent) ProcessMessage(ctx context.Context, message string) error {
    a.mu.Lock()
    a.history = append(a.history, llm.Message{
        Role:    "user",
        Content: message,
    })
    a.mu.Unlock()

    // Detectar intent
    result, err := a.intentDetector.DetectWithHistory(ctx, message, a.history)
    if err != nil {
        return err
    }

    // Delegar para handler registry
    response, err := a.handlerRegistry.Handle(ctx, a, result)
    if err != nil {
        return err
    }

    fmt.Println(response)

    a.mu.Lock()
    a.history = append(a.history, llm.Message{
        Role:    "assistant",
        Content: response,
    })
    a.mu.Unlock()

    return nil
}

// Getters para handlers acessarem dependências
func (a *Agent) GetToolRegistry() *tools.Registry {
    return a.toolRegistry
}

func (a *Agent) GetLLMClient() *llm.Client {
    return a.llmClient
}

func (a *Agent) GetConfirmationManager() *confirmation.Manager {
    return a.confirmManager
}

func (a *Agent) GetMode() modes.OperationMode {
    return a.mode
}
```

### Benefícios da Refatoração

#### Antes (handlers.go):
- ❌ 2282 linhas em 1 arquivo
- ❌ 14+ responsabilidades misturadas
- ❌ Difícil de testar (precisa mockar Agent completo)
- ❌ Difícil de estender (modificar arquivo gigante)
- ❌ Code review complexo

#### Depois (handler pattern):
- ✅ 8-10 arquivos (~200-300 linhas cada)
- ✅ 1 responsabilidade por arquivo
- ✅ Fácil de testar (mockar interface Handler)
- ✅ Fácil de estender (criar novo Handler)
- ✅ Code review simples

---

## 📦 Fase 2: Validators Package

### Problema Atual

Código duplicado em múltiplos lugares:

```go
// handlers.go linha 197
func extractJSON(content string) string {
    // 15 linhas de regex
}

// handlers.go linha 859
// Mesma lógica repetida

// handlers.go linha 216
func cleanCodeContent(content, filePath string) string {
    // 30 linhas de limpeza
}

// handlers.go linha 814
// Mesma função chamada novamente
```

### Solução Proposta

```go
// internal/validators/filename.go
package validators

import (
    "path/filepath"
    "strings"
)

type FileValidator struct{}

func NewFileValidator() *FileValidator {
    return &FileValidator{}
}

// IsValidFilename verifica se o filename é válido
func (v *FileValidator) IsValidFilename(name string) bool {
    if name == "" {
        return false
    }

    // Não pode ter certos caracteres
    invalid := []string{"<", ">", ":", "\"", "|", "?", "*"}
    for _, char := range invalid {
        if strings.Contains(name, char) {
            return false
        }
    }

    // Deve ter extensão válida
    ext := filepath.Ext(name)
    return ext != ""
}

// ExtractTargetFile extrai nome do arquivo de uma mensagem
func (v *FileValidator) ExtractTargetFile(message string) string {
    // Implementação de detecção
    return ""
}
```

```go
// internal/validators/json.go
package validators

import (
    "encoding/json"
    "regexp"
)

type JSONValidator struct {
    jsonRegex *regexp.Regexp
}

func NewJSONValidator() *JSONValidator {
    return &JSONValidator{
        jsonRegex: regexp.MustCompile(`\{[\s\S]*\}`),
    }
}

// ExtractJSON extrai JSON de uma string
func (v *JSONValidator) ExtractJSON(content string) string {
    match := v.jsonRegex.FindString(content)
    if match == "" {
        return ""
    }
    return match
}

// ParseJSON faz parse de JSON com fallback
func (v *JSONValidator) ParseJSON(content string) (map[string]interface{}, error) {
    var result map[string]interface{}

    // Tentar extrair JSON se não for válido
    if err := json.Unmarshal([]byte(content), &result); err != nil {
        extracted := v.ExtractJSON(content)
        if extracted == "" {
            return nil, err
        }

        if err := json.Unmarshal([]byte(extracted), &result); err != nil {
            return nil, err
        }
    }

    return result, nil
}
```

```go
// internal/validators/code.go
package validators

import (
    "path/filepath"
    "strings"
)

type CodeCleaner struct{}

func NewCodeCleaner() *CodeCleaner {
    return &CodeCleaner{}
}

// CleanCode remove markdown e formata código
func (c *CodeCleaner) CleanCode(content, filePath string) string {
    ext := filepath.Ext(filePath)

    // Remover markdown code blocks
    content = strings.TrimPrefix(content, "```"+ext)
    content = strings.TrimPrefix(content, "```")
    content = strings.TrimSuffix(content, "```")

    // Remover espaços extras
    content = strings.TrimSpace(content)

    return content
}

// DetectLanguage detecta linguagem do código
func (c *CodeCleaner) DetectLanguage(filePath string) string {
    ext := filepath.Ext(filePath)

    languageMap := map[string]string{
        ".go":   "go",
        ".js":   "javascript",
        ".ts":   "typescript",
        ".py":   "python",
        ".java": "java",
        ".rs":   "rust",
    }

    return languageMap[ext]
}
```

### Uso nos Handlers

```go
// internal/handlers/file_write_handler.go (REFATORADO)
package handlers

import (
    "github.com/johnpitter/ollama-code/internal/validators"
)

type FileWriteHandler struct {
    validator *validators.FileValidator
    jsonVal   *validators.JSONValidator
    cleaner   *validators.CodeCleaner
}

func NewFileWriteHandler() *FileWriteHandler {
    return &FileWriteHandler{
        validator: validators.NewFileValidator(),
        jsonVal:   validators.NewJSONValidator(),
        cleaner:   validators.NewCodeCleaner(),
    }
}

func (h *FileWriteHandler) Handle(ctx context.Context, ag *agent.Agent, result *intent.DetectionResult) (string, error) {
    // Usar validators
    if !h.validator.IsValidFilename(filePath) {
        return "", fmt.Errorf("invalid filename")
    }

    content = h.cleaner.CleanCode(content, filePath)

    parsed, err := h.jsonVal.ParseJSON(responseFromLLM)
    // ...
}
```

---

## 🔄 Fase 3: Dependency Injection

### Problema Atual

```go
// internal/agent/agent.go
func NewAgent(cfg Config) (*Agent, error) {
    // Cria todas as dependências diretamente
    llmClient := llm.NewClient(cfg.OllamaURL, cfg.Model)
    intentDetector := intent.NewDetector(llmClient)
    toolRegistry := tools.NewRegistry()
    // ... mais 10+ dependências
}
```

**Problemas**:
- Impossível injetar mocks
- Difícil testar
- Acoplamento forte

### Solução Proposta

```go
// internal/agent/dependencies.go (NOVO)
package agent

import (
    "github.com/johnpitter/ollama-code/internal/cache"
    "github.com/johnpitter/ollama-code/internal/commands"
    "github.com/johnpitter/ollama-code/internal/confirmation"
    "github.com/johnpitter/ollama-code/internal/handlers"
    "github.com/johnpitter/ollama-code/internal/intent"
    "github.com/johnpitter/ollama-code/internal/llm"
    "github.com/johnpitter/ollama-code/internal/session"
    "github.com/johnpitter/ollama-code/internal/skills"
    "github.com/johnpitter/ollama-code/internal/tools"
    "github.com/johnpitter/ollama-code/internal/websearch"
)

// Dependencies agrupa todas as dependências do Agent
type Dependencies struct {
    LLMClient        llm.ClientInterface
    IntentDetector   intent.DetectorInterface
    ToolRegistry     tools.RegistryInterface
    HandlerRegistry  handlers.RegistryInterface
    CommandRegistry  commands.RegistryInterface
    SkillRegistry    skills.RegistryInterface
    ConfirmManager   confirmation.ManagerInterface
    WebSearch        websearch.OrchestratorInterface
    SessionManager   session.ManagerInterface
    CacheManager     cache.ManagerInterface
}

// Builder para facilitar criação
type DependenciesBuilder struct {
    deps Dependencies
}

func NewDependenciesBuilder() *DependenciesBuilder {
    return &DependenciesBuilder{
        deps: Dependencies{},
    }
}

func (b *DependenciesBuilder) WithLLMClient(client llm.ClientInterface) *DependenciesBuilder {
    b.deps.LLMClient = client
    return b
}

func (b *DependenciesBuilder) WithIntentDetector(detector intent.DetectorInterface) *DependenciesBuilder {
    b.deps.IntentDetector = detector
    return b
}

// ... mais builders ...

func (b *DependenciesBuilder) Build() (*Dependencies, error) {
    // Validar que todas as dependências foram fornecidas
    if b.deps.LLMClient == nil {
        return nil, fmt.Errorf("LLMClient is required")
    }

    // ... validar outras ...

    return &b.deps, nil
}
```

```go
// internal/agent/agent.go (REFATORADO)
func NewAgent(deps *Dependencies, cfg Config) (*Agent, error) {
    return &Agent{
        llmClient:       deps.LLMClient,
        intentDetector:  deps.IntentDetector,
        toolRegistry:    deps.ToolRegistry,
        handlerRegistry: deps.HandlerRegistry,
        // ... usar deps injetadas
        mode:    cfg.Mode,
        workDir: cfg.WorkDir,
    }, nil
}
```

```go
// cmd/ollama-code/main.go (REFATORADO)
func main() {
    // ... código de config ...

    // Criar dependências
    llmClient := llm.NewClient(cfg.OllamaURL, cfg.Model)
    intentDetector := intent.NewDetector(llmClient)
    toolRegistry := tools.NewRegistry()
    handlerRegistry := handlers.NewRegistry()

    // Usar builder
    deps, err := agent.NewDependenciesBuilder().
        WithLLMClient(llmClient).
        WithIntentDetector(intentDetector).
        WithToolRegistry(toolRegistry).
        WithHandlerRegistry(handlerRegistry).
        Build()

    if err != nil {
        log.Fatal(err)
    }

    // Criar agent com dependências
    ag, err := agent.NewAgent(deps, agentCfg)
    // ...
}
```

### Benefícios para Testes

```go
// internal/agent/agent_test.go (NOVO - POSSÍVEL!)
package agent_test

import (
    "testing"
    "github.com/johnpitter/ollama-code/internal/agent"
    "github.com/johnpitter/ollama-code/internal/mocks"
)

func TestAgent_ProcessMessage(t *testing.T) {
    // Criar mocks
    mockLLM := mocks.NewMockLLMClient()
    mockIntent := mocks.NewMockIntentDetector()
    mockTools := mocks.NewMockToolRegistry()
    mockHandlers := mocks.NewMockHandlerRegistry()

    // Configurar comportamento esperado
    mockIntent.On("DetectWithHistory", ...).Return(&intent.DetectionResult{
        Intent: intent.IntentReadFile,
        Parameters: map[string]interface{}{
            "file_path": "test.go",
        },
    }, nil)

    mockHandlers.On("Handle", ...).Return("File content", nil)

    // Criar dependencies com mocks
    deps, _ := agent.NewDependenciesBuilder().
        WithLLMClient(mockLLM).
        WithIntentDetector(mockIntent).
        WithToolRegistry(mockTools).
        WithHandlerRegistry(mockHandlers).
        Build()

    // Criar agent
    ag, err := agent.NewAgent(deps, agent.Config{})
    if err != nil {
        t.Fatal(err)
    }

    // Testar
    err = ag.ProcessMessage(context.Background(), "read test.go")
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }

    // Verificar chamadas
    mockIntent.AssertExpectations(t)
    mockHandlers.AssertExpectations(t)
}
```

---

## 📊 Comparação Antes/Depois

### Estrutura de Arquivos

#### ANTES:
```
internal/agent/
├── agent.go (376 linhas)
└── handlers.go (2282 linhas) ← PROBLEMA
```

#### DEPOIS:
```
internal/
├── agent/
│   ├── agent.go (200 linhas) ← Simplificado
│   └── dependencies.go (100 linhas) ← Novo
├── handlers/
│   ├── handler.go (50 linhas) ← Interface
│   ├── registry.go (100 linhas) ← Registry
│   ├── file_read_handler.go (150 linhas)
│   ├── file_write_handler.go (250 linhas)
│   ├── execute_handler.go (100 linhas)
│   ├── search_handler.go (150 linhas)
│   ├── git_handler.go (100 linhas)
│   ├── analyze_handler.go (120 linhas)
│   ├── question_handler.go (80 linhas)
│   └── helpers.go (100 linhas)
└── validators/
    ├── filename.go (80 linhas)
    ├── json.go (100 linhas)
    └── code.go (120 linhas)
```

### Métricas

| Métrica | Antes | Depois | Melhoria |
|---------|-------|--------|----------|
| Maior arquivo | 2282 linhas | 250 linhas | ✅ 89% |
| Total de arquivos | 2 | 16 | ✅ Modular |
| Linhas por arquivo (média) | 1329 | 135 | ✅ 90% |
| Testabilidade | Impossível | Fácil | ✅ 100% |
| Code duplication | 4-5 padrões | 0 | ✅ 100% |
| Acoplamento Agent | 12+ deps | 6 deps | ✅ 50% |

---

## 🚀 Plano de Migração

### Semana 1: Setup e Infraestrutura
- [ ] Criar pacote `handlers/`
- [ ] Criar pacote `validators/`
- [ ] Definir interfaces (Handler, Registry)
- [ ] Implementar HandlerRegistry

### Semana 2: Migrar Handlers Simples
- [ ] FileReadHandler
- [ ] SearchHandler
- [ ] GitHandler
- [ ] QuestionHandler

### Semana 3: Migrar Handlers Complexos
- [ ] FileWriteHandler
- [ ] ExecuteHandler
- [ ] AnalyzeHandler
- [ ] WebSearchHandler

### Semana 4: Cleanup e Testes
- [ ] Remover handlers.go antigo
- [ ] Atualizar agent.go
- [ ] Adicionar testes unitários
- [ ] Atualizar documentação

### Semana 5: Validators e DI
- [ ] Implementar validators
- [ ] Refatorar handlers para usar validators
- [ ] Implementar dependency injection
- [ ] Atualizar testes

---

## ✅ Critérios de Sucesso

- [ ] handlers.go não existe mais
- [ ] Todos os handlers têm < 300 linhas
- [ ] 80%+ de code coverage em handlers
- [ ] Zero code duplication
- [ ] Agent com < 10 dependências diretas
- [ ] Todos os testes passando
- [ ] Build limpo sem warnings

---

## 🎯 Conclusão

Esta refatoração transformará o código de:

**ANTES**: Monolítico, difícil de manter, impossível de testar

**DEPOIS**: Modular, fácil de manter, totalmente testável

Sem quebrar funcionalidades existentes! 🎉

---

**Próximo Passo**: Começar implementação da Semana 1
