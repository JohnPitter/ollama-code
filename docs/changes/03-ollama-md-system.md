# Sistema OLLAMA.md Hierárquico - Changelog

**Data:** 2024-12-19
**Commit:** `9ca86b4`
**Autor:** Claude AI

## Resumo

Sistema hierárquico de arquivos OLLAMA.md para configuração contextual do agente em múltiplos níveis, permitindo personalização desde configurações organizacionais até preferências locais de subdiretórios.

## Motivação

Projetos precisam de configurações em diferentes níveis:
- **Empresa:** Padrões organizacionais para todos os projetos
- **Projeto:** Configurações específicas do projeto
- **Linguagem:** Convenções por linguagem (Go, Python, etc)
- **Local:** Preferências de subdiretórios específicos

## Arquitetura Hierárquica

### Níveis (Prioridade: 4 > 3 > 2 > 1)

```
4. Local      /projeto/subdir/OLLAMA.md        (Maior prioridade)
3. Language   /projeto/.ollama/go/OLLAMA.md
2. Project    /projeto/OLLAMA.md
1. Enterprise ~/.ollama/OLLAMA.md              (Menor prioridade)
```

**Regra de Merge:**
Arquivos com maior prioridade sobrescrevem configurações de menor prioridade.

## Componentes

### 1. OllamaFile (`internal/ollamamd/ollamamd.go`)

Representa um arquivo OLLAMA.md individual:

```go
type OllamaFile struct {
    Path       string                // Caminho completo
    Level      Level                 // Nível hierárquico
    Content    string                // Conteúdo markdown
    Language   string                // Linguagem (se LevelLanguage)
    LoadedAt   time.Time             // Timestamp de carregamento
    Sections   map[string]string     // Seções indexadas
}
```

**Níveis:**
```go
const (
    LevelEnterprise Level = iota // ~/.ollama/OLLAMA.md
    LevelProject                 // /projeto/OLLAMA.md
    LevelLanguage                // /projeto/.ollama/go/OLLAMA.md
    LevelLocal                   // /projeto/subdir/OLLAMA.md
)
```

**Métodos principais:**

```go
// Carregar arquivo
func (f *OllamaFile) Load() error

// Parsing de seções markdown
func (f *OllamaFile) parseSections()

// Acessar seções
func (f *OllamaFile) GetSection(title string) (string, bool)
func (f *OllamaFile) HasSection(title string) bool

// Extrair diretrizes
func (f *OllamaFile) ExtractGuidelines() []string

// Extrair preferências (key: value)
func (f *OllamaFile) ExtractPreferences() map[string]string

// Prioridade (1-4)
func (f *OllamaFile) Priority() int
```

### 2. OllamaContext (`internal/ollamamd/ollamamd.go`)

Contexto mesclado de múltiplos arquivos:

```go
type OllamaContext struct {
    Files       []*OllamaFile         // Ordenados por prioridade
    Merged      string                // Conteúdo mesclado
    Guidelines  []string              // Diretrizes únicas
    Preferences map[string]string     // Preferências mescladas
    LoadedAt    time.Time             // Timestamp
}
```

**Merge de Guidelines:**
- Extrai de seções: "Guidelines", "Rules", "Best Practices"
- Remove duplicatas
- Mantém ordem de prioridade

**Merge de Preferences:**
- Extrai de seções: "Preferences", "Settings"
- Parse de `key: value`
- Últimos (maior prioridade) sobrescrevem primeiros

### 3. Loader (`internal/ollamamd/loader.go`)

Carregador hierárquico:

```go
type Loader struct {
    workDir string
    homeDir string
}
```

**Métodos principais:**

```go
// Criar loader
func NewLoader(workDir string) *Loader

// Carregar todos os níveis
func (l *Loader) Load() (*OllamaContext, error)

// Descobrir arquivos sem carregar
func (l *Loader) Discover() []string

// Carregar arquivo individual
func (l *Loader) LoadSingle(path string) (*OllamaFile, error)
```

**Carregamento por nível:**

```go
// 1. Enterprise
func (l *Loader) loadEnterprise() *OllamaFile {
    path := filepath.Join(homeDir, ".ollama", "OLLAMA.md")
    // ...
}

// 2. Project
func (l *Loader) loadProject() *OllamaFile {
    path := filepath.Join(workDir, "OLLAMA.md")
    // ...
}

// 3. Language
func (l *Loader) loadLanguage() []*OllamaFile {
    // Procura em .ollama/*/OLLAMA.md
    // Exemplo: .ollama/go/OLLAMA.md, .ollama/python/OLLAMA.md
}

// 4. Local
func (l *Loader) loadLocal() []*OllamaFile {
    // Walk recursivo (max 2 níveis)
    // Ignora: .git, node_modules, vendor, build
}
```

## Formato dos Arquivos OLLAMA.md

### Estrutura Markdown

```markdown
# Guidelines

- Sempre use snake_case para nomes de funções
- Documente todas as funções públicas
- Mantenha funções com menos de 50 linhas

# Preferences

- Language: Go
- Style: Google Style Guide
- Max Line Length: 100

# Best Practices

- Use context.Context para cancelamento
- Sempre feche recursos (defer)
- Prefira composition over inheritance

# Code Standards

## Error Handling

Sempre wrap errors com contexto:

\`\`\`go
return fmt.Errorf("failed to process: %w", err)
\`\`\`

## Testing

Todos os pacotes devem ter >80% coverage.
```

### Parsing de Seções

**Headers detectados:**
```markdown
# Título Nível 1
## Título Nível 2
### Título Nível 3
```

**Extração:**
```go
sections := map[string]string{
    "Guidelines": "- Sempre use...\n- Documente...",
    "Preferences": "- Language: Go\n- Style: ...",
    "Best Practices": "- Use context...",
}
```

### Extração de Guidelines

**Procura seções:**
- "Guidelines", "Diretrizes"
- "Rules", "Regras"
- "Best Practices", "Melhores Práticas"
- "Do's and Don'ts"

**Extrai listas:**
```markdown
- Item de lista
* Outro item
```

**Resultado:**
```go
guidelines := []string{
    "Sempre use snake_case para nomes de funções",
    "Documente todas as funções públicas",
    "Mantenha funções com menos de 50 linhas",
}
```

### Extração de Preferências

**Procura seções:**
- "Preferences", "Preferências"
- "Settings", "Configurações"

**Parse de key: value:**
```markdown
- Language: Go
- Style: Google Style Guide
- Max Line Length: 100
```

**Resultado:**
```go
preferences := map[string]string{
    "Language": "Go",
    "Style": "Google Style Guide",
    "Max Line Length": "100",
}
```

## Fluxo de Carregamento

```
1. NewLoader(workDir)
2. loader.Load()
   ├─ loadEnterprise()    → ~/.ollama/OLLAMA.md
   ├─ loadProject()       → /projeto/OLLAMA.md
   ├─ loadLanguage()      → /projeto/.ollama/go/OLLAMA.md
   └─ loadLocal()         → /projeto/src/OLLAMA.md
3. Ordenar por Priority (1→4)
4. merge(files)
   ├─ Mesclar conteúdo
   ├─ Extrair guidelines (sem duplicatas)
   └─ Mesclar preferences (sobrescrever)
5. Retornar OllamaContext
```

## Integração com Agent

**Agent struct:**
```go
type Agent struct {
    ollamaContext *ollamamd.OllamaContext
    // ...
}
```

**Inicialização:**
```go
func NewAgent(cfg) (*Agent, error) {
    // Carregar OLLAMA.md
    loader := ollamamd.NewLoader(cfg.WorkDir)
    ollamaContext, err := loader.Load()

    if err != nil {
        fmt.Printf("⚠️  Aviso: Não foi possível carregar OLLAMA.md: %v\n", err)
    } else if len(ollamaContext.Files) > 0 {
        fmt.Printf("📋 Carregados %d arquivo(s) OLLAMA.md\n", len(ollamaContext.Files))
    }

    agent := &Agent{
        ollamaContext: ollamaContext,
        // ...
    }

    return agent, nil
}
```

**Uso no prompt:**
```go
func (a *Agent) buildSystemPrompt() string {
    prompt := "Você é um assistente de código.\n\n"

    if a.ollamaContext != nil {
        // Adicionar guidelines
        if len(a.ollamaContext.Guidelines) > 0 {
            prompt += "## Guidelines\n"
            for _, guideline := range a.ollamaContext.Guidelines {
                prompt += "- " + guideline + "\n"
            }
        }

        // Adicionar preferences
        if len(a.ollamaContext.Preferences) > 0 {
            prompt += "\n## Preferences\n"
            for key, value := range a.ollamaContext.Preferences {
                prompt += key + ": " + value + "\n"
            }
        }
    }

    return prompt
}
```

## Exemplos de Uso

### Enterprise Level

**~/.ollama/OLLAMA.md:**
```markdown
# Company Standards

## Code Review

All code must pass review before merge.

## Security

- Never commit secrets
- Use environment variables
- Scan dependencies weekly

# Preferences

- License: MIT
- Copyright: © 2024 MyCompany
```

### Project Level

**/projeto/OLLAMA.md:**
```markdown
# Project: E-commerce API

## Architecture

Using clean architecture with DDD patterns.

## Dependencies

- Go 1.21+
- PostgreSQL 15
- Redis 7

# Preferences

- API Version: v2
- Default Timeout: 30s
```

### Language Level

**/projeto/.ollama/go/OLLAMA.md:**
```markdown
# Go Conventions

## Naming

- Packages: lowercase, single word
- Interfaces: -er suffix (Reader, Writer)
- Errors: Err prefix (ErrNotFound)

## Best Practices

- Use golangci-lint
- Run go vet before commit
- Table-driven tests

# Preferences

- Go Version: 1.21
- Linter: golangci-lint
```

### Local Level

**/projeto/api/handlers/OLLAMA.md:**
```markdown
# API Handlers

## Patterns

All handlers should:
- Validate input
- Use middleware for auth
- Return consistent errors
- Log requests

# Preferences

- Max Request Size: 10MB
- Rate Limit: 100 req/min
```

## Auto-Discovery

**Discover() encontra todos os arquivos:**

```bash
$ loader.Discover()

[
  "/home/user/.ollama/OLLAMA.md",
  "/projeto/OLLAMA.md",
  "/projeto/.ollama/go/OLLAMA.md",
  "/projeto/.ollama/python/OLLAMA.md",
  "/projeto/api/OLLAMA.md",
  "/projeto/api/handlers/OLLAMA.md"
]
```

**Walk configurado:**
- Max profundidade: 2 níveis
- Ignora: `.git`, `node_modules`, `vendor`, `build`, `.*`
- Procura: `OLLAMA.md`

## Merge Example

**Enterprise:**
```markdown
# Guidelines
- Use MIT license
- Code review required
```

**Project:**
```markdown
# Guidelines
- Use clean architecture
- 80% test coverage

# Preferences
- Language: Go
```

**Language (.ollama/go/):**
```markdown
# Guidelines
- Use golangci-lint
- Prefer table-driven tests

# Preferences
- Language: Go
- Linter: golangci-lint
```

**Resultado Mesclado:**

```go
guidelines := []string{
    "Use MIT license",           // Enterprise
    "Code review required",       // Enterprise
    "Use clean architecture",     // Project
    "80% test coverage",          // Project
    "Use golangci-lint",          // Language
    "Prefer table-driven tests",  // Language
}

preferences := map[string]string{
    "Language": "Go",             // Language sobrescreveu Project
    "Linter": "golangci-lint",    // Language (único)
}
```

## Arquivos Criados

1. `internal/ollamamd/ollamamd.go` (213 linhas)
   - Types: Level, OllamaFile, OllamaContext
   - Parsing de seções
   - Extração de guidelines e preferences

2. `internal/ollamamd/loader.go` (323 linhas)
   - Loader hierárquico
   - Auto-discovery
   - Merge inteligente

**Total:** 536 linhas

## Arquivos Modificados

- `internal/agent/agent.go`: Integração do OllamaContext

## Benefícios

✅ **Hierarquia:** Múltiplos níveis de configuração
✅ **Prioridade:** Sistema claro de override
✅ **Modular:** Cada nível independente
✅ **Flexível:** Suporta qualquer estrutura markdown
✅ **Auto-discovery:** Encontra arquivos automaticamente
✅ **Merge inteligente:** Sem duplicatas
✅ **Thread-safe:** Carregamento seguro
✅ **Optional:** Não falha se arquivos não existirem

## Próximos Steps

- [ ] Integrar guidelines no system prompt do LLM
- [ ] Comando `/ollama-context` para visualizar contexto
- [ ] Hot-reload quando arquivos mudarem
- [ ] Validação de sintaxe markdown
- [ ] Template generator (`ollama-code init-config`)
- [ ] Suporte para YAML/JSON além de Markdown
- [ ] Cache de arquivos carregados
- [ ] Métricas de uso das guidelines

## Referências

- Commit: `9ca86b4`
- Inspiração: EditorConfig, .gitignore hierarchy
- Pattern: Chain of Responsibility
