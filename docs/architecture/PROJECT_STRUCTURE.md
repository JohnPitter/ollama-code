# 📁 Estrutura do Projeto Ollama Code

Este documento descreve a organização completa do projeto.

## 🌳 Árvore de Diretórios

```
ollama-code/
├── cmd/                    # Aplicações executáveis
│   └── ollama-code/        # CLI principal
│       └── main.go
│
├── internal/               # Código interno (não exportado)
│   ├── agent/              # Agente principal e lógica de orquestração
│   │   ├── agent.go        # Core do agente
│   │   ├── agent_test.go   # Testes do agente
│   │   └── handlers.go     # Handlers de intenções
│   │
│   ├── skills/             # Sistema de Skills especializados
│   │   ├── skill.go        # Interface e tipos base
│   │   ├── registry.go     # Registry de skills
│   │   ├── research.go     # ResearchSkill
│   │   ├── api.go          # APISkill
│   │   └── codeanalysis.go # CodeAnalysisSkill
│   │
│   ├── ollamamd/           # Sistema OLLAMA.md hierárquico
│   │   ├── ollamamd.go     # Tipos e OllamaFile
│   │   └── loader.go       # Loader hierárquico
│   │
│   ├── websearch/          # Sistema de pesquisa web
│   │   ├── orchestrator.go # Orquestrador de buscas
│   │   └── fetcher.go      # Fetch de conteúdo HTML
│   │
│   ├── llm/                # Cliente Ollama
│   │   ├── client.go       # Cliente HTTP para Ollama
│   │   └── client_test.go  # Testes do cliente
│   │
│   ├── intent/             # Detecção de intenções
│   │   ├── detector.go     # Detector de intenções
│   │   └── detector_test.go
│   │
│   ├── tools/              # Ferramentas do agente
│   │   ├── registry.go     # Registry de ferramentas
│   │   ├── file_ops.go     # Operações de arquivo
│   │   ├── command_exec.go # Execução de comandos
│   │   └── *_test.go       # Testes
│   │
│   ├── session/            # Gerenciamento de sessões
│   │   ├── manager.go      # Manager de sessões
│   │   └── manager_test.go
│   │
│   ├── cache/              # Sistema de cache
│   │   ├── manager.go
│   │   └── manager_test.go
│   │
│   ├── confirmation/       # Sistema de confirmações
│   │   ├── manager.go
│   │   └── manager_test.go
│   │
│   ├── config/             # Configuração
│   │   ├── config.go
│   │   └── config_test.go
│   │
│   ├── statusline/         # Status line rico
│   │   └── statusline.go
│   │
│   ├── commands/           # Comandos built-in
│   │   ├── registry.go
│   │   └── builtins.go
│   │
│   ├── modes/              # Modos de operação
│   │   └── modes.go
│   │
│   ├── hardware/           # Detecção de hardware
│   │   ├── detector.go
│   │   ├── detector_test.go
│   │   └── optimizer.go
│   │
│   ├── checkpoint/         # Sistema de checkpoints
│   │   ├── manager.go
│   │   └── types.go
│   │
│   ├── hooks/              # Sistema de hooks
│   │   └── manager.go
│   │
│   ├── doctor/             # Health checks
│   │   └── health.go
│   │
│   ├── memory/             # Gerenciamento de memória
│   ├── output/             # Formatação de output
│   ├── background/         # Tarefas em background
│   ├── sandbox/            # Execução sandboxed
│   └── telemetry/          # Telemetria (opcional)
│
├── docs/                   # Documentação
│   ├── archive/            # Documentos históricos
│   │   ├── ENTERPRISE_FEATURES.md
│   │   ├── IMPLEMENTATION_PLAN.md
│   │   ├── PHASE2_COMPLETE.md
│   │   ├── PRODUCTION_READINESS.md
│   │   └── FINAL_REPORT.md
│   │
│   ├── development/        # Docs para desenvolvedores
│   │   ├── CONFIG.md       # Configuração avançada
│   │   ├── INSTALL.md      # Instalação detalhada
│   │   └── ROADMAP.md      # Roadmap do projeto
│   │
│   └── user-guide/         # Guias do usuário
│       └── (a criar)
│
├── changes/                # Changelog detalhado
│   ├── 01-web-search-hybrid.md
│   ├── 02-agent-skills.md
│   └── 03-ollama-md-system.md
│
├── scripts/                # Scripts utilitários
│   ├── download-models-direct.sh
│   ├── download-models-direct.ps1
│   └── ollama-optimized-setup.sh
│
├── build/                  # Binários compilados (ignorado no git)
│   └── ollama-code
│
├── .claude/                # Configuração Claude Code
│   └── settings.local.json
│
├── .git/                   # Controle de versão
├── .gitignore              # Arquivos ignorados
│
├── build.sh                # Script de build (Linux/Mac)
├── build.bat               # Script de build (Windows)
├── Makefile                # Make targets
│
├── go.mod                  # Módulo Go
├── go.sum                  # Checksums de dependências
│
├── README.md               # README principal (noob-friendly)
├── README.old.md           # Backup do README anterior
├── CONTRIBUTING.md         # Guia de contribuição
├── PROJECT_STRUCTURE.md    # Este arquivo
├── LICENSE                 # Licença MIT
└── config.example.json     # Exemplo de configuração
```

## 📦 Módulos Principais

### 1. Agent (`internal/agent/`)

**Responsabilidade:** Orquestração principal do assistente

**Componentes:**
- `agent.go`: Core do agente, inicialização, processamento
- `handlers.go`: Handlers específicos para cada tipo de intenção
- `agent_test.go`: Testes unitários

**Interações:**
- Usa LLM para comunicação com modelos
- Usa Intent para detectar intenções
- Usa Tools para executar ações
- Usa Skills para tarefas especializadas
- Usa WebSearch para buscar na web

### 2. Skills (`internal/skills/`)

**Responsabilidade:** Habilidades especializadas modulares

**Componentes:**
- `skill.go`: Interface Skill, tipos Task/Result
- `registry.go`: Gerenciamento de skills
- `research.go`: Pesquisa avançada
- `api.go`: Chamadas API
- `codeanalysis.go`: Análise de código

**Padrão:** Strategy + Registry

### 3. OLLAMA.md (`internal/ollamamd/`)

**Responsabilidade:** Configuração hierárquica contextual

**Componentes:**
- `ollamamd.go`: Tipos OllamaFile e OllamaContext
- `loader.go`: Carregamento e merge hierárquico

**Níveis:**
1. Enterprise (~/.ollama/OLLAMA.md)
2. Project (/projeto/OLLAMA.md)
3. Language (/projeto/.ollama/go/OLLAMA.md)
4. Local (/projeto/subdir/OLLAMA.md)

### 4. Web Search (`internal/websearch/`)

**Responsabilidade:** Pesquisa e fetch de conteúdo web

**Componentes:**
- `orchestrator.go`: Orquestra buscas (DuckDuckGo)
- `fetcher.go`: Fetch de conteúdo HTML real

**Fluxo:**
```
Query → DuckDuckGo → URLs → Fetch HTML → Parse → Clean → LLM
```

### 5. LLM (`internal/llm/`)

**Responsabilidade:** Comunicação com Ollama

**Componentes:**
- `client.go`: Cliente HTTP
- Suporte para streaming
- Gerenciamento de contexto

### 6. Intent (`internal/intent/`)

**Responsabilidade:** Detectar intenção do usuário

**Intenções:**
- `question`: Pergunta simples
- `read_file`: Ler arquivo
- `write_file`: Escrever arquivo
- `execute_command`: Executar comando
- `web_search`: Buscar na web

### 7. Tools (`internal/tools/`)

**Responsabilidade:** Ferramentas disponíveis para o agente

**Tools:**
- FileReader: Leitura de arquivos
- FileWriter: Escrita de arquivos
- CommandExecutor: Execução de comandos
- CodeSearcher: Busca em código
- ProjectAnalyzer: Análise de projetos
- GitOperations: Operações git

## 🔄 Fluxo de Execução

```
1. main.go
   ↓
2. Agent.NewAgent()
   ├─ Carrega LLM Client
   ├─ Carrega Intent Detector
   ├─ Carrega Tools Registry
   ├─ Carrega Skills Registry
   ├─ Carrega OLLAMA.md Context
   └─ Carrega Web Search
   ↓
3. Agent.ProcessMessage(userInput)
   ↓
4. Intent.Detect(userInput)
   ↓
5. Agent.handle{Intent}()
   ├─ handleQuestion() → LLM
   ├─ handleReadFile() → Tools
   ├─ handleWebSearch() → WebSearch + LLM
   └─ etc.
   ↓
6. Response → User
```

## 🧪 Testes

**Localização:** `*_test.go` ao lado do código

**Cobertura Atual:**
- Total: 90 testes
- LLM: 77.8%
- Intent: 91.7%
- Confirmation: 87.5%
- Tools: Vários testes
- Session: Vários testes

**Executar:**
```bash
go test ./...                    # Todos os testes
go test ./internal/llm/          # Pacote específico
go test -v ./...                 # Verbose
go test -cover ./...             # Com coverage
```

## 📝 Documentação

**Estrutura:**

```
docs/
├── archive/         # Histórico (antigos planos, reports)
├── development/     # Para desenvolvedores (CONFIG, INSTALL, ROADMAP)
└── user-guide/      # Para usuários (a criar)

changes/            # Changelog detalhado de cada feature

README.md          # Principal (noob-friendly)
CONTRIBUTING.md    # Como contribuir
PROJECT_STRUCTURE.md # Este arquivo
```

## 🔧 Build e Deploy

**Build:**
```bash
./build.sh              # Linux/Mac
./build.bat             # Windows
make build              # Via Makefile
```

**Output:** `build/ollama-code`

**Targets do Makefile:**
- `make build`: Compila
- `make test`: Roda testes
- `make clean`: Limpa build
- `make install`: Instala no sistema

## 🎯 Convenções

### Nomenclatura

- **Packages:** minúsculas, uma palavra (`websearch`, `ollamamd`)
- **Files:** snake_case (`ollama_md.go`, `code_analysis.go`)
- **Types:** PascalCase (`OllamaFile`, `ResearchSkill`)
- **Functions:** camelCase (`loadEnterprise`, `processMessage`)
- **Constants:** PascalCase (`LevelEnterprise`, `ModeReadOnly`)

### Organização de Imports

```go
import (
    // Standard library
    "context"
    "fmt"
    "strings"

    // External
    "github.com/fatih/color"

    // Internal
    "github.com/johnpitter/ollama-code/internal/llm"
    "github.com/johnpitter/ollama-code/internal/skills"
)
```

### Estrutura de Arquivos

```go
// 1. Package declaration
package agent

// 2. Imports
import (...)

// 3. Constants
const (
    DefaultTimeout = 30 * time.Second
)

// 4. Types
type Agent struct {...}

// 5. Constructors
func NewAgent(...) *Agent {...}

// 6. Methods (receiver functions)
func (a *Agent) ProcessMessage(...) {...}

// 7. Helper functions (non-exported)
func processInternal(...) {...}
```

## 🚀 Próximos Passos

- [ ] Adicionar docs/user-guide/
- [ ] Criar exemplos práticos em examples/
- [ ] Adicionar integração contínua (GitHub Actions)
- [ ] Melhorar cobertura de testes (meta: 90%)
- [ ] Documentação de API (GoDoc)
- [ ] Benchmarks de performance

## 📞 Contato

- Issues: https://github.com/johnpitter/ollama-code/issues
- Discussions: https://github.com/johnpitter/ollama-code/discussions

---

Última atualização: 2024-12-19
