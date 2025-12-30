# 🏗️ Análise Arquitetural - Ollama Code

## Data da Análise: 22/12/2024

---

## 📊 Resumo Executivo

| Aspecto | Status | Severidade |
|---------|--------|-----------|
| Estrutura de Pacotes | ✓ Bom | ✅ |
| Padrões Arquiteturais | ✓ Bem Implementados | ✅ |
| Circular Dependencies | ✓ Nenhuma | ✅ |
| **God Objects** | **6 arquivos** | 🔴 **CRÍTICO** |
| **handlers.go** | **2282 linhas** | 🔴 **CRÍTICO** |
| Separation of Concerns | ❌ Violado | 🔴 |
| SOLID Violations | 3 de 5 | 🟠 |
| Error Handling | Inconsistente | 🟠 |
| Interface Segregation | Falta Interfaces | 🟠 |
| Dependency Injection | Não usa | 🟠 |
| Code Duplication | 4-5 padrões | 🟠 |
| Testing Capability | Difícil | 🟠 |

---

## 🔍 Problemas Críticos Identificados

### 1. 🔴 GOD OBJECT: handlers.go (2282 linhas)

**Localização**: `internal/agent/handlers.go`

**Problema**: Um único arquivo contém 14+ handlers e 30+ funções helper, violando gravemente o princípio Single Responsibility.

**Responsabilidades Misturadas**:
- Leitura de arquivos (handleReadFile, handleMultiFileRead, handleReadFileWithAnalysis)
- Escrita de arquivos (handleWriteFile, handleMultiFileWrite, handleFileEdit)
- Execução de comandos (handleExecuteCommand)
- Busca em código (handleSearchCode)
- Análise de projeto (handleAnalyzeProject)
- Operações Git (handleGitOperation)
- Busca web (handleWebSearch)
- Processamento de perguntas (handleQuestion)
- Bug fixes (handleBugFix, handleBugFixSimple)

**Impacto**:
- Difícil de testar (necessita mockar Agent completo)
- Difícil de manter (mudanças podem afetar múltiplos handlers)
- Difícil de estender (adicionar novo handler requer modificar arquivo gigante)
- Code review complexo

---

### 2. 🔴 ACOPLAMENTO EXCESSIVO DO AGENT

**Localização**: `internal/agent/agent.go`

**Dependências Diretas** (12+):
```go
import (
    "github.com/johnpitter/ollama-code/internal/cache"
    "github.com/johnpitter/ollama-code/internal/commands"
    "github.com/johnpitter/ollama-code/internal/confirmation"
    "github.com/johnpitter/ollama-code/internal/intent"
    "github.com/johnpitter/ollama-code/internal/llm"
    "github.com/johnpitter/ollama-code/internal/modes"
    "github.com/johnpitter/ollama-code/internal/ollamamd"
    "github.com/johnpitter/ollama-code/internal/session"
    "github.com/johnpitter/ollama-code/internal/skills"
    "github.com/johnpitter/ollama-code/internal/statusline"
    "github.com/johnpitter/ollama-code/internal/tools"
    "github.com/johnpitter/ollama-code/internal/websearch"
)
```

**Consequências**:
- Mudança em qualquer dependência afeta Agent
- Difícil testar em isolamento
- Impossível substituir implementações (sem DI)

---

### 3. 🔴 SEPARATION OF CONCERNS VIOLADO

**Problema**: Agent é tanto ORQUESTRADOR quanto EXECUTOR.

```go
// Agent atua como orquestrador:
ProcessMessage() → detectIntent() → handleIntent()

// MAS também executa lógica de negócio:
handleReadFile()  // Lê arquivos
handleWriteFile() // Escreve arquivos
handleExecuteCommand() // Executa comandos
```

**Deveria ser**:
- Agent = Orquestrador (routing)
- Handlers = Executores (lógica de negócio)

---

## ⚠️ Violações de SOLID

### ❌ S - Single Responsibility Principle

**Violações**:

1. **handlers.go** - 14+ responsabilidades em 1 arquivo
2. **Agent struct** - 50 campos, gerencia tudo:
   - LLM Client
   - Intent Detection
   - Tool/Command/Skill Registries
   - Confirmation/Session/Cache Managers
   - Web Search
   - Status Line
   - Colors
   - History
   - Recent files

3. **advanced_refactoring.go** (825 linhas) - Combina:
   - Análise de código
   - Refatoração automática
   - Pattern matching
   - Sugestão de melhorias

### ⚠️ O - Open/Closed Principle

**Parcialmente cumprido**:
- ✅ Tools/Skills usam interfaces → Extensível
- ❌ handlers.go usa switch/case → Não extensível

```go
// Não extensível sem modificar código:
switch result.Intent {
case intent.IntentReadFile:
    return a.handleReadFile(...)
case intent.IntentWriteFile:
    return a.handleWriteFile(...)
// ... mais 12 cases
}
```

### ⚠️ I - Interface Segregation Principle

**Problemas**:

1. **Tool interface muito genérica**:
```go
type Tool interface {
    Name() string
    Description() string
    Execute(ctx context.Context, params map[string]interface{}) (Result, error)
    RequiresConfirmation() bool
}
```
- `map[string]interface{}` → Sem type safety

2. **Agent expõe múltiplos getters** (violação):
```go
GetSessionManager()
GetCache()
GetCommandRegistry()
GetSkillRegistry()
GetHistory()
```

### ⚠️ D - Dependency Inversion Principle

**Problema**: Agent cria todas as dependências (sem inversão):

```go
func NewAgent(cfg Config) (*Agent, error) {
    llmClient := llm.NewClient(...)          // Criação direta
    intentDetector := intent.NewDetector(...) // Criação direta
    toolRegistry := tools.NewRegistry()      // Criação direta
    // ...
}
```

**Consequência**: Impossível injetar mocks para testes.

---

## 🐘 God Objects Identificados

| Arquivo | Linhas | Responsabilidades | Severidade |
|---------|--------|-------------------|------------|
| `agent/handlers.go` | 2282 | 14 handlers + 30 helpers | 🔴 |
| `agent/agent.go` | 376 | Orquestração + Estado | 🟠 |
| `tools/advanced_refactoring.go` | 825 | Análise + Refatoração | 🟠 |
| `tools/background_task.go` | 484 | Execução async + Queue | 🟠 |
| `tools/git_helper.go` | 554 | 15+ operações Git | 🟠 |
| `tools/code_formatter.go` | 439 | Formatação + Estilo | 🟠 |

---

## 🔄 Código Duplicado

### 1. JSON Parsing (4+ locais)
```go
// handlers.go linha 197
func extractJSON(content string) string

// handlers.go linha 859
// Mesma lógica repetida

// handlers.go linha 1516
// Similar

// handlers.go linha 2173
// Similar
```

### 2. Code Cleaning (3+ chamadas)
```go
// handlers.go linhas 216, 814, 1243
content = cleanCodeContent(content, filePath)
```

### 3. Detecção de Tipo de Projeto (múltiplos handlers)
```go
// handlers.go linha 1884-1940
// Lógica duplicada de detecção
```

### 4. Markdown Removal (múltiplos locais)
```go
// handlers.go linha 1363
content = strings.TrimPrefix(content, "```")

// handlers.go linha 1596
content = strings.TrimPrefix(content, "```html")
```

---

## 🚫 Falta de Interfaces

### Interfaces que DEVERIAM existir:

#### 1. Agent Interface
```go
// FALTA
type Agent interface {
    ProcessMessage(ctx context.Context, msg string) error
    GetMode() modes.OperationMode
    SetMode(modes.OperationMode) error
}
```

**Benefício**: Facilitar testes e mocks

#### 2. Handler Interface
```go
// FALTA
type Handler interface {
    CanHandle(intent intent.Intent) bool
    Handle(ctx context.Context, result *intent.DetectionResult) (string, error)
}
```

**Benefício**: Adicionar handlers sem modificar Agent

#### 3. Manager Interface
```go
// FALTA
type Manager interface {
    Initialize(ctx context.Context) error
    Execute(ctx context.Context, action string, params map[string]interface{}) (Result, error)
    Cleanup() error
}
```

**Benefício**: Gerenciadores intercambiáveis

---

## 📦 Problemas Arquiteturais Específicos

### PA1: handlers.go é um Switch Statement Gigante

Fluxo atual:
```
ProcessMessage()
    ↓
detectIntent()
    ↓
handleIntent() → switch/case (14+ cases)
    ↓
handleReadFile() / handleWriteFile() / ... (14 handlers)
```

**Problema**: Não extensível, viola Open/Closed Principle.

### PA2: Ausência de Singleton Pattern

```go
// Agent é único mas cria múltiplos gerenciadores únicos
session, cache, confirmation, ...
```

Sem clara separação de responsabilidades.

### PA3: Fluxo de Dados Não Estruturado

```
User Input → ProcessMessage() → DetectIntent() → handleIntent() →
toolRegistry.Execute() → Output
```

Sem pipeline pattern ou middleware.

### PA4: Error Handling Inconsistente

```go
// Alguns retornam nil com mensagem:
return "Erro: ...", nil

// Outros retornam erro real:
return "", fmt.Errorf("...")
```

Difícil determinar se foi erro ou comportamento normal.

### PA5: Context Não Respeitado

```go
// Muitas funções recebem ctx mas não o usam:
func (a *Agent) handleReadFile(ctx context.Context, ...) (string, error)
// ctx não é passado adiante
```

### PA6: Sem Timeout Policies

```go
// Pode travar indefinidamente:
response, err := a.llmClient.CompleteStreaming(ctx, messages, opts, callback)
```

---

## ✅ Pontos Positivos

### 1. Estrutura de Pacotes
- ✅ Segue convenção Go (`cmd/`, `internal/`)
- ✅ Nomes descritivos
- ✅ Organização lógica

### 2. Padrões Bem Implementados
- ✅ Registry Pattern (tools, skills, commands)
- ✅ Strategy Pattern (skills)
- ✅ Manager Pattern (session, cache, confirmation)
- ✅ Detector/Analyzer Pattern (intent detection)

### 3. Sem Circular Dependencies
- ✅ Nenhuma circular dependency detectada
- ✅ Grafo de dependências acíclico

### 4. Thread-Safety
- ✅ Uso de `sync.RWMutex` em Registries
- ✅ Acesso concorrente protegido

---

## 🛠️ Plano de Refatoração

### 🔴 PRIORIDADE ALTA (Crítico)

#### 1. Quebrar handlers.go em 4-5 arquivos

**Estrutura proposta**:
```
internal/agent/handlers/
├── handler.go              # Interface + Router
├── file_read_handler.go    # handleReadFile, handleReadFileWithAnalysis
├── file_write_handler.go   # handleWriteFile, handleMultiFileWrite, handleFileEdit
├── execute_handler.go      # handleExecuteCommand
├── search_handler.go       # handleSearchCode, handleWebSearch
├── git_handler.go          # handleGitOperation
├── analyze_handler.go      # handleAnalyzeProject
├── question_handler.go     # handleQuestion
└── helpers.go              # Funções utilitárias
```

**Benefícios**:
- Arquivos menores (200-400 linhas cada)
- Responsabilidades claras
- Facilita testes
- Facilita code review

#### 2. Criar Handler Interface

```go
package handlers

type Handler interface {
    Handle(ctx context.Context,
           agent *Agent,
           result *intent.DetectionResult) (string, error)
}
```

**Implementações**:
```go
type FileReadHandler struct { ... }
func (h *FileReadHandler) Handle(...) (string, error)

type FileWriteHandler struct { ... }
func (h *FileWriteHandler) Handle(...) (string, error)

// ... etc para cada handler
```

#### 3. Criar Handler Registry

```go
type HandlerRegistry struct {
    handlers map[intent.Intent]Handler
    mu       sync.RWMutex
}

func (r *HandlerRegistry) Register(intent intent.Intent, h Handler)
func (r *HandlerRegistry) Handle(ctx context.Context, result *intent.DetectionResult) (string, error)
```

**Uso**:
```go
// Substituir switch/case por:
return handlerRegistry.Handle(ctx, result)
```

#### 4. Remover Campos Desnecessários de Agent

Mover responsabilidades:
- Colors → ColorManager (novo)
- RecentFiles → FileTracker (novo)
- History → SessionManager (já existe!)

---

### 🟠 PRIORIDADE MÉDIA

#### 5. Consolidar Detecção de Intent

**Estrutura proposta**:
```
internal/detection/
├── patterns.go      # Regex patterns
├── keywords.go      # Listas de keywords centralizadas
└── detector.go      # Lógica de detecção
```

**Consolidar**:
```go
// Todas as keywords em um lugar:
var ReadKeywords = []string{"read", "show", "cat", "view", ...}
var WriteKeywords = []string{"write", "create", "make", ...}
var EditKeywords = []string{"edit", "modify", "update", ...}
```

#### 6. Criar Validators Package

**Estrutura proposta**:
```
internal/validators/
├── filename.go      # isValidFilename, extractTargetFile
├── json.go         # extractJSON, parseJSON
└── code.go         # cleanCodeContent, detectLanguage
```

**Benefícios**:
- Reutilização de código
- Testes isolados
- Sem duplicação

#### 7. Implementar Pipeline Pattern

```go
type Pipeline interface {
    AddStep(Step) Pipeline
    Execute(ctx context.Context, data interface{}) (interface{}, error)
}

// Exemplo para file writing:
pipeline.
    AddStep(ValidateFilePath).
    AddStep(GenerateContent).
    AddStep(CleanContent).
    AddStep(ConfirmUser).
    AddStep(WriteFile).
    Execute(ctx, request)
```

---

### 🟢 PRIORIDADE BAIXA

#### 8. Implementar Dependency Injection

```go
type AgentDependencies struct {
    LLM            llm.Client
    IntentDetector intent.Detector
    ToolRegistry   tools.Registry
    Cache          cache.Manager
    Session        session.Manager
    // ... etc
}

func NewAgent(deps *AgentDependencies) *Agent {
    // Injeta ao invés de criar
}
```

#### 9. Criar Agent Interface

```go
type Agent interface {
    ProcessMessage(ctx context.Context, msg string) error
    GetMode() modes.OperationMode
    SetMode(modes.OperationMode) error
}
```

#### 10. Adicionar Observability

```go
type ObservableHandler struct {
    inner   Handler
    logger  Logger
    metrics MetricsCollector
}

func (h *ObservableHandler) Handle(ctx context.Context, ...) (string, error) {
    start := time.Now()
    defer h.metrics.RecordDuration("handler.duration", time.Since(start))

    h.logger.Info("handling request", "intent", result.Intent)
    return h.inner.Handle(ctx, agent, result)
}
```

---

## 📊 Métricas de Impacto

### Situação Atual

| Métrica | Valor Atual | Problema |
|---------|-------------|----------|
| Linhas em handlers.go | 2282 | 🔴 Muito alto |
| Campos em Agent | 50 | 🔴 Muito alto |
| Dependências de Agent | 12+ | 🟠 Alto |
| Arquivos > 500 linhas | 6 | 🟠 Alto |
| Testes de Agent | Difícil | 🔴 Impossível mockar |
| Code duplication | 4-5 padrões | 🟠 Médio |

### Após Refatoração (Estimado)

| Métrica | Valor Esperado | Melhoria |
|---------|----------------|----------|
| Linhas em handlers.go | 0 (quebrado) | ✅ 100% |
| Arquivos handler | 8-10 (~250 linhas cada) | ✅ Modular |
| Campos em Agent | ~25 | ✅ 50% redução |
| Dependências diretas | ~6 | ✅ 50% redução |
| Testes de handlers | Fácil | ✅ Mockável |
| Code duplication | 0-1 | ✅ 80% redução |

---

## 🎯 Roadmap de Implementação

### Fase 1: Refatoração Crítica (1-2 semanas)
- [ ] Criar Handler interface
- [ ] Criar HandlerRegistry
- [ ] Quebrar handlers.go em 8 arquivos
- [ ] Migrar Agent para usar HandlerRegistry
- [ ] Testes unitários para cada handler

### Fase 2: Consolidação (1 semana)
- [ ] Criar validators package
- [ ] Consolidar keywords/patterns em detection package
- [ ] Remover código duplicado
- [ ] Adicionar testes de validators

### Fase 3: Dependency Injection (1 semana)
- [ ] Criar AgentDependencies struct
- [ ] Refatorar NewAgent para aceitar dependencies
- [ ] Adicionar builder pattern para facilitar criação
- [ ] Atualizar testes para usar DI

### Fase 4: Observability (Opcional)
- [ ] Adicionar logging estruturado
- [ ] Implementar metrics collection
- [ ] Adicionar tracing

---

## 🔍 Comparação com Golang-Standards

### ✅ Segue:
- `cmd/` para executáveis
- `internal/` para código privado
- Nomes de pacotes no singular
- Estrutura hierárquica

### ❌ Não Segue:
- Arquivos muito grandes (handlers.go)
- Falta `pkg/` para código reutilizável
- Sem `examples/`
- Falta de interfaces
- Mistura domain com infrastructure

---

## 📝 Conclusão

O projeto **Ollama Code** tem uma **boa fundação arquitetural** com:
- ✅ Estrutura de pacotes adequada
- ✅ Padrões bem implementados (Registry, Manager)
- ✅ Sem circular dependencies
- ✅ Thread-safety

**Porém**, sofre de problemas de **escalabilidade**:
- 🔴 God object (handlers.go com 2282 linhas)
- 🔴 Acoplamento excessivo (Agent com 12+ dependências)
- 🔴 Violações de SOLID (principalmente SRP)
- 🟠 Falta de interfaces para testabilidade

Esses problemas **não impedem o funcionamento**, mas tornam:
- ❌ Manutenção difícil
- ❌ Testes complexos
- ❌ Extensão não trivial
- ❌ Code review demorado

### Recomendação

Executar **Fase 1 do Roadmap** (refatoração crítica) para:
1. Quebrar handlers.go em 8 arquivos menores
2. Implementar Handler pattern
3. Reduzir acoplamento de Agent
4. Melhorar testabilidade

Isso resolverá **80% dos problemas** identificados sem quebrar funcionalidade existente.

---

**Data de Conclusão da Análise**: 22/12/2024
**Próximo Passo**: Implementar Fase 1 do Roadmap
