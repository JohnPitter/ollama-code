# 🚀 OLLAMA CODE - Plano de Implementação Completo

**Linguagem:** Go 1.21+
**Hardware Alvo:** Intel i9 14ª gen (24 cores), 64GB RAM, NVIDIA RTX Ada 2000 (16GB VRAM), 1TB NVMe SSD
**Modelo AI:** qwen2.5-coder:32b-instruct-q6_K (~12GB VRAM)

---

## 🎯 OBJETIVO

Criar um assistente de código AI em Go que:
- ✅ Entende **linguagem natural** (sem comandos `/read`, `/exec`)
- ✅ Detecta **intenções automaticamente** usando LLM
- ✅ Executa **ferramentas** de forma autônoma
- ✅ Oferece **3 modos de operação** (readonly, interactive, autonomous)
- ✅ **Pesquisa na internet** quando necessário
- ✅ **Performance máxima** - startup <15ms, streaming em tempo real

---

## 📋 PRÉ-REQUISITOS E SETUP

### 1. Instalar Go

**Linux:**
```bash
wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version
```

**Windows:**
```powershell
# Baixar instalador: https://go.dev/dl/go1.21.6.windows-amd64.msi
# Executar instalador
# Verificar:
go version
```

**macOS:**
```bash
brew install go@1.21
go version
```

### 2. Instalar Ollama

**Linux:**
```bash
curl -fsSL https://ollama.ai/install.sh | sh

# Iniciar serviço
sudo systemctl enable ollama
sudo systemctl start ollama
```

**Windows:**
```powershell
# Baixar de: https://ollama.ai/download/windows
# Executar instalador OllamaSetup.exe
# Ollama inicia automaticamente
```

**macOS:**
```bash
# Baixar de: https://ollama.ai/download/mac
# Ou via brew:
brew install ollama

# Iniciar serviço
brew services start ollama
```

### 3. Configurar Ollama para RTX Ada 2000

**Linux/macOS** (`~/.config/ollama/env.conf`):
```bash
# GPU Configuration
export OLLAMA_GPU_LAYERS=999
export OLLAMA_NUM_GPU=1
export CUDA_VISIBLE_DEVICES=0

# Performance
export OLLAMA_MAX_LOADED_MODELS=2
export OLLAMA_NUM_PARALLEL=4
export OLLAMA_FLASH_ATTENTION=1

# Memory
export OLLAMA_MAX_VRAM=16384
export OLLAMA_HOST=127.0.0.1:11434
```

**Windows:**

**COM admin** (PowerShell como Admin):
```powershell
[System.Environment]::SetEnvironmentVariable('OLLAMA_GPU_LAYERS', '999', 'Machine')
[System.Environment]::SetEnvironmentVariable('OLLAMA_NUM_GPU', '1', 'Machine')
[System.Environment]::SetEnvironmentVariable('OLLAMA_MAX_LOADED_MODELS', '2', 'Machine')
[System.Environment]::SetEnvironmentVariable('OLLAMA_NUM_PARALLEL', '4', 'Machine')
[System.Environment]::SetEnvironmentVariable('OLLAMA_FLASH_ATTENTION', '1', 'Machine')

# Reiniciar serviço Ollama
Restart-Service Ollama
```

**SEM admin** (PowerShell normal - Use 'User' em vez de 'Machine'):
```powershell
# Definir para seu usuário apenas
[System.Environment]::SetEnvironmentVariable('OLLAMA_GPU_LAYERS', '999', 'User')
[System.Environment]::SetEnvironmentVariable('OLLAMA_NUM_GPU', '1', 'User')
[System.Environment]::SetEnvironmentVariable('OLLAMA_MAX_LOADED_MODELS', '2', 'User')
[System.Environment]::SetEnvironmentVariable('OLLAMA_NUM_PARALLEL', '4', 'User')
[System.Environment]::SetEnvironmentVariable('OLLAMA_FLASH_ATTENTION', '1', 'User')
[System.Environment]::SetEnvironmentVariable('OLLAMA_MAX_VRAM', '16384', 'User')

# Fechar e reabrir PowerShell
```

**Alternativa - Script .bat** (não precisa admin):

Crie `ollama-config.bat`:
```batch
@echo off
set OLLAMA_GPU_LAYERS=999
set OLLAMA_NUM_GPU=1
set OLLAMA_MAX_LOADED_MODELS=2
set OLLAMA_NUM_PARALLEL=4
set OLLAMA_FLASH_ATTENTION=1
set OLLAMA_MAX_VRAM=16384
echo Ollama configurado! Execute 'ollama serve' neste terminal.
cmd
```

### 4. Baixar Modelos

**Método 1: Ollama Pull (requer rede livre)**
```bash
ollama pull qwen2.5-coder:32b-instruct-q6_K
ollama pull nomic-embed-text
```

**Método 2: Download Direto (bypass proxy corporativo)**

Use os scripts fornecidos:
```bash
# Linux/macOS
chmod +x download-models-direct.sh
./download-models-direct.sh

# Windows
.\download-models-direct.ps1
```

**Método 3: Download Manual**

Links diretos para download:
```
QWen2.5-Coder 32B (19GB):
https://huggingface.co/Qwen/Qwen2.5-Coder-32B-Instruct-GGUF/resolve/main/qwen2.5-coder-32b-instruct-q6_k.gguf

Nomic Embed Text (274MB):
https://huggingface.co/nomic-ai/nomic-embed-text-v1.5/resolve/main/model.safetensors
```

Após download, importar:
```bash
# Criar Modelfile
cat > Modelfile << 'EOF'
FROM ./qwen2.5-coder-32b-instruct-q6_k.gguf
TEMPLATE """{{ if .System }}<|im_start|>system
{{ .System }}<|im_end|>
{{ end }}{{ if .Prompt }}<|im_start|>user
{{ .Prompt }}<|im_end|>
<|im_start|>assistant
{{ end }}"""
PARAMETER stop "<|im_start|>"
PARAMETER stop "<|im_end|>"
PARAMETER temperature 0.7
PARAMETER num_gpu 999
EOF

# Importar
ollama create qwen2.5-coder:32b-instruct-q6_K -f Modelfile
```

### 5. Verificar Setup

```bash
# Testar Ollama
curl http://localhost:11434/api/tags

# Testar modelo
ollama run qwen2.5-coder:32b-instruct-q6_K "Write a hello world in Go"

# Verificar GPU (Linux/Windows)
nvidia-smi

# Verificar Go
go version
```

---

## 🏗️ ARQUITETURA DO SISTEMA

```
┌─────────────────────────────────────────────────────────────┐
│                    USUÁRIO (Linguagem Natural)               │
│  "Cria um servidor REST em Go"                              │
│  "Corrija os erros nesse arquivo"                           │
│  "Pesquise como fazer autenticação JWT"                     │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│              CONVERSATIONAL AGENT                            │
│  • Mantém histórico da conversa                             │
│  • Gerencia contexto do workspace                           │
│  • Coordena todo o fluxo                                    │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│           INTENT DETECTION (LLM-Powered)                     │
│  Analisa mensagem e detecta intenções:                      │
│  ├─ READ_FILE: "mostra o main.go"                          │
│  ├─ WRITE_FILE: "cria um servidor"                         │
│  ├─ EXECUTE_CMD: "roda os testes"                          │
│  ├─ SEARCH_CODE: "onde está a função X?"                   │
│  ├─ WEB_SEARCH: "como fazer JWT?" → internet               │
│  └─ MULTI_INTENT: Múltiplas ações em sequência            │
└────────────────────────┬────────────────────────────────────┘
                         │
         ┌───────────────┴───────────────┐
         │                               │
         ▼                               ▼
┌──────────────────┐            ┌──────────────────┐
│  TOOL EXECUTION  │            │  WEB SEARCH      │
│                  │            │  ORCHESTRATOR    │
│  • FileReader    │            │                  │
│  • FileWriter    │            │  • DuckDuckGo    │
│  • CommandExec   │            │  • StackOverflow │
│  • CodeSearcher  │            │  • GitHub        │
│  • GitOps        │            │  • Synthesizer   │
│  • Analyzer      │            │    (LLM)         │
└────────┬─────────┘            └────────┬─────────┘
         │                               │
         └───────────────┬───────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│         CONFIRMATION MANAGER (Based on Mode)                 │
│                                                              │
│  READ-ONLY:     Só leitura, bloqueia escrita               │
│  INTERACTIVE:   Confirma ações destrutivas (padrão)        │
│  AUTONOMOUS:    Tudo automático                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 📁 ESTRUTURA DE ARQUIVOS

```
ollama-code/
├── cmd/
│   └── ollama-code/
│       └── main.go                   # Entry point da aplicação
│
├── internal/
│   ├── agent/
│   │   ├── agent.go                  # Agente principal
│   │   ├── conversation.go           # Gerenciamento de conversa
│   │   ├── context.go                # Contexto do workspace
│   │   └── operation_mode.go         # Modos de operação
│   │
│   ├── intent/
│   │   ├── detector.go               # Detecção de intenções
│   │   ├── classifier.go             # Classificação de tipos
│   │   └── prompts.go                # System prompts
│   │
│   ├── tools/
│   │   ├── tool.go                   # Interface Tool
│   │   ├── file_ops.go               # Leitura/escrita de arquivos
│   │   ├── command_exec.go           # Execução de comandos
│   │   ├── code_search.go            # Busca em código (ripgrep)
│   │   ├── git_ops.go                # Operações Git
│   │   ├── project_analyzer.go       # Análise de projeto
│   │   └── registry.go               # Registro de ferramentas
│   │
│   ├── websearch/
│   │   ├── orchestrator.go           # Coordenador de buscas
│   │   ├── cache.go                  # Cache de resultados
│   │   ├── providers/
│   │   │   ├── provider.go           # Interface
│   │   │   ├── duckduckgo.go         # DuckDuckGo API
│   │   │   ├── stackoverflow.go      # Stack Overflow API
│   │   │   └── github.go             # GitHub Search API
│   │   └── processor/
│   │       ├── extractor.go          # Extrai conteúdo relevante
│   │       ├── cleaner.go            # Remove HTML/ads
│   │       └── synthesizer.go        # Sintetiza com LLM
│   │
│   ├── llm/
│   │   ├── ollama.go                 # Client Ollama
│   │   ├── streaming.go              # Streaming de respostas
│   │   └── cache.go                  # Cache de contexto
│   │
│   └── confirmation/
│       ├── manager.go                # Gerenciador de confirmações
│       └── policies.go               # Políticas de aprovação
│
├── pkg/
│   └── workspace/
│       ├── scanner.go                # Scanner de workspace
│       ├── indexer.go                # Indexação de arquivos
│       └── watcher.go                # File watcher
│
├── configs/
│   └── default.yaml                  # Configuração padrão
│
├── scripts/
│   ├── download-models-direct.sh     # Download de modelos (Linux)
│   └── download-models-direct.ps1    # Download de modelos (Windows)
│
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── IMPLEMENTATION_PLAN.md            # Este arquivo
```

---

## 🔨 IMPLEMENTAÇÃO POR FASES

### FASE 1: Fundação (2 dias)

#### 1.1. Setup do Projeto
```bash
mkdir ollama-code
cd ollama-code

# Inicializar módulo Go
go mod init ollama-code

# Estrutura de diretórios
mkdir -p cmd/ollama-code
mkdir -p internal/{agent,intent,tools,websearch,llm,confirmation}
mkdir -p internal/websearch/{providers,processor}
mkdir -p pkg/workspace
mkdir -p configs
```

#### 1.2. Dependências (`go.mod`)
```go
module ollama-code

go 1.21

require (
    github.com/fatih/color v1.16.0
    github.com/spf13/cobra v1.8.0
    gopkg.in/yaml.v3 v3.0.1
)
```

#### 1.3. LLM Client (`internal/llm/ollama.go`)

```go
package llm

import (
    "bufio"
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
)

type OllamaClient struct {
    baseURL string
    model   string
    client  *http.Client
}

func NewOllamaClient(baseURL, model string) *OllamaClient {
    return &OllamaClient{
        baseURL: baseURL,
        model:   model,
        client:  &http.Client{},
    }
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type ChatRequest struct {
    Model    string    `json:"model"`
    Messages []Message `json:"messages"`
    Stream   bool      `json:"stream"`
    Options  Options   `json:"options,omitempty"`
}

type Options struct {
    Temperature float64 `json:"temperature,omitempty"`
    NumPredict  int     `json:"num_predict,omitempty"`
    NumGPU      int     `json:"num_gpu,omitempty"`
}

type ChatResponse struct {
    Model   string  `json:"model"`
    Message Message `json:"message"`
    Done    bool    `json:"done"`
}

// CompleteStreaming - Gera resposta com streaming
func (o *OllamaClient) CompleteStreaming(
    ctx context.Context,
    messages []Message,
    onChunk func(string),
) (string, error) {
    reqBody := ChatRequest{
        Model:    o.model,
        Messages: messages,
        Stream:   true,
        Options: Options{
            Temperature: 0.7,
            NumGPU:      999, // Usar toda GPU
        },
    }

    jsonData, _ := json.Marshal(reqBody)

    req, err := http.NewRequestWithContext(
        ctx,
        "POST",
        o.baseURL+"/api/chat",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return "", err
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := o.client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var fullResponse strings.Builder
    scanner := bufio.NewScanner(resp.Body)

    for scanner.Scan() {
        var chatResp ChatResponse
        if err := json.Unmarshal(scanner.Bytes(), &chatResp); err != nil {
            continue
        }

        chunk := chatResp.Message.Content
        fullResponse.WriteString(chunk)

        if onChunk != nil {
            onChunk(chunk) // Stream output em tempo real
        }

        if chatResp.Done {
            break
        }
    }

    return fullResponse.String(), scanner.Err()
}
```

#### 1.4. Agent Básico (`internal/agent/agent.go`)

```go
package agent

import (
    "context"
    "fmt"
    "sync"

    "ollama-code/internal/llm"
    "ollama-code/internal/tools"
    "ollama-code/internal/intent"
    "ollama-code/internal/confirmation"
    "ollama-code/internal/websearch"
)

type Agent struct {
    llm            *llm.OllamaClient
    toolRegistry   *tools.Registry
    intentDetector *intent.Detector
    confirmMgr     *confirmation.Manager
    webSearch      *websearch.Orchestrator

    // Estado
    conversation   []llm.Message
    workspaceDir   string
    config         *Config

    mu sync.RWMutex
}

type Config struct {
    OllamaURL       string
    Model           string
    OperationConfig *OperationConfig
}

func NewAgent(cfg *Config) (*Agent, error) {
    llmClient := llm.NewOllamaClient(cfg.OllamaURL, cfg.Model)

    toolRegistry := tools.NewRegistry()
    toolRegistry.Register(tools.NewFileReader())
    toolRegistry.Register(tools.NewFileWriter())
    toolRegistry.Register(tools.NewCommandExecutor())
    toolRegistry.Register(tools.NewCodeSearcher())
    toolRegistry.Register(tools.NewProjectAnalyzer())
    toolRegistry.Register(tools.NewGitOperations())

    intentDetector := intent.NewDetector(llmClient)
    confirmMgr := confirmation.NewManager(cfg.OperationConfig)
    webSearch := websearch.NewOrchestrator(llmClient)

    return &Agent{
        llm:            llmClient,
        toolRegistry:   toolRegistry,
        intentDetector: intentDetector,
        confirmMgr:     confirmMgr,
        webSearch:      webSearch,
        conversation:   make([]llm.Message, 0),
        workspaceDir:   mustGetwd(),
        config:         cfg,
    }, nil
}

// ProcessMessage - MÉTODO PRINCIPAL
func (a *Agent) ProcessMessage(ctx context.Context, userMessage string) error {
    // 1. Adicionar mensagem do usuário
    a.addMessage("user", userMessage)

    // 2. Verificar se precisa web search
    if a.shouldSearchWeb(userMessage) {
        if err := a.performWebSearch(ctx, userMessage); err != nil {
            fmt.Printf("⚠️  Aviso: falha na pesquisa web: %v\n", err)
        }
    }

    // 3. Construir contexto completo
    contextMessage := a.buildContext(userMessage)

    // 4. Detectar intenções
    intents, err := a.intentDetector.Detect(ctx, contextMessage)
    if err != nil {
        return err
    }

    // 5. Executar ferramentas
    for _, intent := range intents {
        if err := a.executeIntent(ctx, intent); err != nil {
            return err
        }
    }

    // 6. Gerar resposta com streaming
    fmt.Println("\n🤖 Assistente:\n")

    response, err := a.llm.CompleteStreaming(
        ctx,
        a.conversation,
        func(chunk string) {
            fmt.Print(chunk) // Stream em tempo real
        },
    )
    if err != nil {
        return err
    }

    fmt.Println()

    // 7. Adicionar resposta ao histórico
    a.addMessage("assistant", response)

    return nil
}

func (a *Agent) executeIntent(ctx context.Context, intent *intent.Intent) error {
    tool := a.toolRegistry.Get(intent.ToolName)
    if tool == nil {
        return fmt.Errorf("ferramenta não encontrada: %s", intent.ToolName)
    }

    // Verificar se precisa confirmação
    action := confirmation.Action{
        Type:        confirmation.ActionType(intent.Type),
        Description: intent.Description,
        Tool:        tool.Name(),
        Params:      intent.Params,
    }

    approved, err := a.confirmMgr.ShouldApprove(action)
    if err != nil || !approved {
        return fmt.Errorf("ação não aprovada")
    }

    // Executar ferramenta
    result, err := tool.Execute(ctx, intent.Params)
    if err != nil {
        return err
    }

    // Adicionar resultado ao contexto
    a.addMessage("system", fmt.Sprintf("Resultado da ferramenta %s:\n%s", tool.Name(), result.Output))

    return nil
}

func (a *Agent) addMessage(role, content string) {
    a.mu.Lock()
    defer a.mu.Unlock()

    a.conversation = append(a.conversation, llm.Message{
        Role:    role,
        Content: content,
    })
}
```

#### 1.5. CLI Principal (`cmd/ollama-code/main.go`)

```go
package main

import (
    "bufio"
    "context"
    "fmt"
    "os"
    "strings"

    "github.com/spf13/cobra"
    "ollama-code/internal/agent"
)

func main() {
    ctx := context.Background()

    var operationMode string

    rootCmd := &cobra.Command{
        Use:   "ollama-code",
        Short: "AI-powered code assistant",
    }

    rootCmd.PersistentFlags().StringVarP(&operationMode, "mode", "m", "interactive",
        "Operation mode: readonly, interactive, autonomous")

    chatCmd := &cobra.Command{
        Use:   "chat",
        Short: "Start interactive chat",
        Run: func(cmd *cobra.Command, args []string) {
            var opConfig *agent.OperationConfig

            switch strings.ToLower(operationMode) {
            case "readonly":
                opConfig = agent.NewReadOnlyConfig()
                fmt.Println("🔒 Modo: READ-ONLY")
            case "autonomous":
                opConfig = agent.NewAutonomousConfig()
                fmt.Println("🚀 Modo: AUTONOMOUS")
            default:
                opConfig = agent.NewInteractiveConfig()
                fmt.Println("⚠️  Modo: INTERACTIVE")
            }

            cfg := &agent.Config{
                OllamaURL:       "http://localhost:11434",
                Model:           "qwen2.5-coder:32b-instruct-q6_K",
                OperationConfig: opConfig,
            }

            ag, err := agent.NewAgent(cfg)
            if err != nil {
                fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
                os.Exit(1)
            }

            runChat(ctx, ag)
        },
    }

    rootCmd.AddCommand(chatCmd)
    rootCmd.Execute()
}

func runChat(ctx context.Context, ag *agent.Agent) {
    fmt.Println("\n🤖 Ollama Code - AI Assistant")
    fmt.Println("Digite 'exit' para sair\n")

    reader := bufio.NewReader(os.Stdin)

    for {
        fmt.Print("Você: ")
        message, _ := reader.ReadString('\n')
        message = strings.TrimSpace(message)

        if message == "exit" || message == "quit" {
            fmt.Println("\n👋 Até logo!")
            break
        }

        if message == "" {
            continue
        }

        if err := ag.ProcessMessage(ctx, message); err != nil {
            fmt.Printf("❌ Erro: %v\n", err)
        }

        fmt.Println()
    }
}
```

---

### FASE 2: Intent Detection (2 dias)

#### 2.1. Intent Types (`internal/intent/types.go`)

```go
package intent

type IntentType string

const (
    IntentReadFile       IntentType = "READ_FILE"
    IntentWriteFile      IntentType = "WRITE_FILE"
    IntentExecuteCommand IntentType = "EXECUTE_COMMAND"
    IntentSearchCode     IntentType = "SEARCH_CODE"
    IntentAnalyzeProject IntentType = "ANALYZE_PROJECT"
    IntentGitOperation   IntentType = "GIT_OPERATION"
    IntentWebSearch      IntentType = "WEB_SEARCH"
    IntentExplain        IntentType = "EXPLAIN"
)

type Intent struct {
    Type        IntentType         `json:"type"`
    ToolName    string             `json:"tool_name"`
    Description string             `json:"description"`
    Params      map[string]any     `json:"params"`
    Confidence  float64            `json:"confidence"`
}
```

#### 2.2. Intent Detector (`internal/intent/detector.go`)

```go
package intent

import (
    "context"
    "encoding/json"
    "strings"
)

type Detector struct {
    llm LLMClient
}

type LLMClient interface {
    CompleteStreaming(ctx context.Context, messages []Message, onChunk func(string)) (string, error)
}

func NewDetector(llm LLMClient) *Detector {
    return &Detector{llm: llm}
}

func (d *Detector) Detect(ctx context.Context, fullContext string) ([]*Intent, error) {
    systemPrompt := d.getSystemPrompt()

    messages := []Message{
        {Role: "system", Content: systemPrompt},
        {Role: "user", Content: fullContext},
    }

    var result string
    response, err := d.llm.CompleteStreaming(ctx, messages, func(chunk string) {
        result += chunk
    })

    if err != nil {
        // Fallback para detecção heurística
        return d.fallbackDetection(fullContext), nil
    }

    // Parse JSON
    var intents []*Intent
    if err := json.Unmarshal([]byte(response), &intents); err != nil {
        return d.fallbackDetection(fullContext), nil
    }

    return intents, nil
}

func (d *Detector) getSystemPrompt() string {
    return `Você é um sistema de detecção de intenções para um assistente de código.

Analise a mensagem do usuário e retorne um JSON array de intenções.

Formato de saída (JSON apenas, sem texto extra):
[
  {
    "type": "READ_FILE" | "WRITE_FILE" | "EXECUTE_COMMAND" | "SEARCH_CODE" | "WEB_SEARCH" | "ANALYZE_PROJECT" | "GIT_OPERATION" | "EXPLAIN",
    "tool_name": "file_reader" | "file_writer" | "command_executor" | "code_searcher" | "web_searcher" | "project_analyzer" | "git_ops",
    "description": "Descrição da ação",
    "params": { "key": "value" },
    "confidence": 0.0-1.0
  }
]

Exemplos:

User: "Mostra o arquivo main.go"
Output: [{"type": "READ_FILE", "tool_name": "file_reader", "description": "Ler arquivo main.go", "params": {"path": "main.go"}, "confidence": 0.95}]

User: "Cria um servidor HTTP em Go"
Output: [{"type": "WRITE_FILE", "tool_name": "file_writer", "description": "Criar servidor HTTP", "params": {"path": "server.go"}, "confidence": 0.90}]

User: "Como fazer autenticação JWT em Go?"
Output: [{"type": "WEB_SEARCH", "tool_name": "web_searcher", "description": "Buscar informações sobre JWT em Go", "params": {"query": "JWT authentication Go", "intent": "EXAMPLE"}, "confidence": 0.85}]

User: "Leia main.go e corrija os erros"
Output: [
  {"type": "READ_FILE", "tool_name": "file_reader", "description": "Ler main.go", "params": {"path": "main.go"}, "confidence": 0.95},
  {"type": "WRITE_FILE", "tool_name": "file_writer", "description": "Corrigir erros em main.go", "params": {"path": "main.go"}, "confidence": 0.85}
]

Retorne APENAS o JSON array.`
}

func (d *Detector) fallbackDetection(context string) []*Intent {
    lower := strings.ToLower(context)
    intents := []*Intent{}

    // Heurística simples
    if strings.Contains(lower, "leia") || strings.Contains(lower, "mostra") {
        if filename := extractFilename(context); filename != "" {
            intents = append(intents, &Intent{
                Type:        IntentReadFile,
                ToolName:    "file_reader",
                Description: "Ler arquivo",
                Params:      map[string]any{"path": filename},
                Confidence:  0.7,
            })
        }
    }

    if strings.Contains(lower, "cria") || strings.Contains(lower, "escreva") {
        intents = append(intents, &Intent{
            Type:        IntentWriteFile,
            ToolName:    "file_writer",
            Description: "Escrever arquivo",
            Params:      map[string]any{},
            Confidence:  0.6,
        })
    }

    if strings.Contains(lower, "pesquise") || strings.Contains(lower, "como fazer") {
        intents = append(intents, &Intent{
            Type:        IntentWebSearch,
            ToolName:    "web_searcher",
            Description: "Pesquisar na internet",
            Params:      map[string]any{"query": context},
            Confidence:  0.75,
        })
    }

    return intents
}

func extractFilename(text string) string {
    // Extrai nomes de arquivo simples (ex: "main.go", "server.js")
    words := strings.Fields(text)
    for _, word := range words {
        if strings.Contains(word, ".") && !strings.HasPrefix(word, ".") {
            return strings.Trim(word, ".,!?\"'")
        }
    }
    return ""
}
```

---

### FASE 3: Tools System (3 dias)

#### 3.1. Tool Interface (`internal/tools/tool.go`)

```go
package tools

import "context"

type Tool interface {
    Name() string
    Description() string
    Execute(ctx context.Context, params map[string]any) (Result, error)
}

type Result struct {
    Success bool
    Output  string
    Data    map[string]any
    Error   error
}

type Registry struct {
    tools map[string]Tool
}

func NewRegistry() *Registry {
    return &Registry{tools: make(map[string]Tool)}
}

func (r *Registry) Register(tool Tool) {
    r.tools[tool.Name()] = tool
}

func (r *Registry) Get(name string) Tool {
    return r.tools[name]
}
```

#### 3.2. File Operations (`internal/tools/file_ops.go`)

```go
package tools

import (
    "context"
    "encoding/base64"
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

// FileReader - Lê arquivos de texto E imagens
type FileReader struct{}

func NewFileReader() *FileReader {
    return &FileReader{}
}

func (f *FileReader) Name() string {
    return "file_reader"
}

func (f *FileReader) Description() string {
    return "Lê conteúdo de arquivos (texto e imagens)"
}

func (f *FileReader) Execute(ctx context.Context, params map[string]any) (Result, error) {
    path, ok := params["path"].(string)
    if !ok {
        return Result{Success: false}, fmt.Errorf("parâmetro 'path' obrigatório")
    }

    absPath, _ := filepath.Abs(path)

    // Detectar tipo de arquivo
    ext := strings.ToLower(filepath.Ext(absPath))

    // Imagens suportadas
    imageExts := []string{".png", ".jpg", ".jpeg", ".gif", ".bmp", ".webp"}
    isImage := false
    for _, imgExt := range imageExts {
        if ext == imgExt {
            isImage = true
            break
        }
    }

    if isImage {
        return f.readImage(absPath)
    }

    // Ler arquivo de texto normal
    content, err := os.ReadFile(absPath)
    if err != nil {
        return Result{Success: false, Error: err}, err
    }

    return Result{
        Success: true,
        Output:  string(content),
        Data:    map[string]any{
            "path": absPath,
            "size": len(content),
            "type": "text",
        },
    }, nil
}

func (f *FileReader) readImage(path string) (Result, error) {
    // Ler imagem como base64 para enviar ao LLM
    content, err := os.ReadFile(path)
    if err != nil {
        return Result{Success: false, Error: err}, err
    }

    // Converter para base64
    base64Image := base64.StdEncoding.EncodeToString(content)

    return Result{
        Success: true,
        Output:  fmt.Sprintf("Imagem lida: %s (%d bytes)", filepath.Base(path), len(content)),
        Data: map[string]any{
            "path":        path,
            "size":        len(content),
            "type":        "image",
            "base64":      base64Image,
            "mime_type":   getMimeType(path),
        },
    }, nil
}

func getMimeType(path string) string {
    ext := strings.ToLower(filepath.Ext(path))
    mimeTypes := map[string]string{
        ".png":  "image/png",
        ".jpg":  "image/jpeg",
        ".jpeg": "image/jpeg",
        ".gif":  "image/gif",
        ".bmp":  "image/bmp",
        ".webp": "image/webp",
    }

    if mime, ok := mimeTypes[ext]; ok {
        return mime
    }
    return "application/octet-stream"
}

// FileWriter
type FileWriter struct{}

func NewFileWriter() *FileWriter {
    return &FileWriter{}
}

func (f *FileWriter) Name() string {
    return "file_writer"
}

func (f *FileWriter) Description() string {
    return "Escreve conteúdo em arquivos"
}

func (f *FileWriter) Execute(ctx context.Context, params map[string]any) (Result, error) {
    path, ok := params["path"].(string)
    if !ok {
        return Result{Success: false}, fmt.Errorf("parâmetro 'path' obrigatório")
    }

    content, ok := params["content"].(string)
    if !ok {
        return Result{Success: false}, fmt.Errorf("parâmetro 'content' obrigatório")
    }

    dir := filepath.Dir(path)
    os.MkdirAll(dir, 0755)

    if err := os.WriteFile(path, []byte(content), 0644); err != nil {
        return Result{Success: false, Error: err}, err
    }

    return Result{
        Success: true,
        Output:  fmt.Sprintf("Arquivo criado: %s", path),
        Data:    map[string]any{"path": path, "size": len(content)},
    }, nil
}
```

*(Implementar similarmente: CommandExecutor, CodeSearcher, ProjectAnalyzer, GitOperations)*

---

### FASE 4: Operation Modes (2 dias)

#### 4.1. Operation Modes (`internal/agent/operation_mode.go`)

```go
package agent

type OperationMode int

const (
    ModeReadOnly OperationMode = iota
    ModeInteractive
    ModeAutonomous
)

type OperationConfig struct {
    Mode                 OperationMode
    AutoApproveRead      bool
    AutoApproveWrite     bool
    AutoApproveExecute   bool
    ShowPreview          bool
    RequireDoubleConfirm []string
    BlockedCommands      []string
    LogActions           bool
}

func NewReadOnlyConfig() *OperationConfig {
    return &OperationConfig{
        Mode:                 ModeReadOnly,
        AutoApproveRead:      true,
        AutoApproveWrite:     false,
        AutoApproveExecute:   false,
        ShowPreview:          true,
        BlockedCommands:      []string{"rm", "git commit", "git push"},
        LogActions:           true,
    }
}

func NewInteractiveConfig() *OperationConfig {
    return &OperationConfig{
        Mode:                 ModeInteractive,
        AutoApproveRead:      true,
        AutoApproveWrite:     false,
        AutoApproveExecute:   false,
        ShowPreview:          true,
        RequireDoubleConfirm: []string{"rm", "git push"},
        BlockedCommands:      []string{"rm -rf /"},
        LogActions:           true,
    }
}

func NewAutonomousConfig() *OperationConfig {
    return &OperationConfig{
        Mode:                 ModeAutonomous,
        AutoApproveRead:      true,
        AutoApproveWrite:     true,
        AutoApproveExecute:   true,
        ShowPreview:          false,
        RequireDoubleConfirm: []string{"sudo"},
        BlockedCommands:      []string{"rm -rf /", "git push --force origin main"},
        LogActions:           true,
    }
}
```

#### 4.2. Confirmation Manager (`internal/confirmation/manager.go`)

```go
package confirmation

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

type ActionType string

type Action struct {
    Type        ActionType
    Description string
    Tool        string
    Params      map[string]any
    Command     string
}

type Manager struct {
    config *OperationConfig
}

func NewManager(config *OperationConfig) *Manager {
    return &Manager{config: config}
}

func (m *Manager) ShouldApprove(action Action) (bool, error) {
    // Verificar comandos bloqueados
    if m.isBlocked(action) {
        return false, fmt.Errorf("ação bloqueada: %s", action.Description)
    }

    switch m.config.Mode {
    case ModeReadOnly:
        return m.approveReadOnly(action)
    case ModeInteractive:
        return m.approveInteractive(action)
    case ModeAutonomous:
        return m.approveAutonomous(action)
    }

    return false, nil
}

func (m *Manager) approveReadOnly(action Action) (bool, error) {
    readOnlyActions := []ActionType{"READ_FILE", "SEARCH_CODE", "ANALYZE_PROJECT", "WEB_SEARCH"}

    for _, allowed := range readOnlyActions {
        if action.Type == allowed {
            return true, nil
        }
    }

    return false, fmt.Errorf("ação bloqueada em modo READ-ONLY")
}

func (m *Manager) approveInteractive(action Action) (bool, error) {
    if m.config.AutoApproveRead && action.Type == "READ_FILE" {
        return true, nil
    }

    // Pedir confirmação
    return m.requestConfirmation(action)
}

func (m *Manager) approveAutonomous(action Action) (bool, error) {
    // Tudo aprovado automaticamente (exceto bloqueados)
    return true, nil
}

func (m *Manager) requestConfirmation(action Action) (bool, error) {
    fmt.Printf("\n🔔 Confirmação necessária:\n")
    fmt.Printf("   Ação: %s\n", action.Description)
    fmt.Printf("   Tipo: %s\n", action.Type)

    fmt.Print("\nExecutar? [y/N]: ")

    reader := bufio.NewReader(os.Stdin)
    response, _ := reader.ReadString('\n')
    response = strings.ToLower(strings.TrimSpace(response))

    return response == "y" || response == "yes", nil
}

func (m *Manager) isBlocked(action Action) bool {
    for _, blocked := range m.config.BlockedCommands {
        if strings.Contains(action.Command, blocked) {
            return true
        }
    }
    return false
}
```

---

### FASE 5: Web Search (2 dias)

#### 5.1. Web Search Orchestrator (`internal/websearch/orchestrator.go`)

```go
package websearch

import (
    "context"
    "sync"
    "time"
)

type SearchQuery struct {
    Query      string
    Intent     string
    Language   string
    MaxResults int
    UseCache   bool
}

type SearchResult struct {
    Title       string
    URL         string
    Snippet     string
    Content     string
    Source      string
    Relevance   float64
    PublishedAt time.Time
}

type Orchestrator struct {
    providers []SearchProvider
    cache     *Cache
    llm       LLMClient
}

func NewOrchestrator(llm LLMClient) *Orchestrator {
    return &Orchestrator{
        providers: []SearchProvider{
            NewDuckDuckGoProvider(),
            NewStackOverflowProvider(),
        },
        cache: NewCache(),
        llm:   llm,
    }
}

func (o *Orchestrator) Search(ctx context.Context, query *SearchQuery) ([]*SearchResult, error) {
    // Verificar cache
    if query.UseCache {
        if cached := o.cache.Get(query.Query); cached != nil {
            return cached, nil
        }
    }

    // Executar buscas em paralelo
    results := make([]*SearchResult, 0)
    var wg sync.WaitGroup
    resultsChan := make(chan []*SearchResult, len(o.providers))

    for _, provider := range o.providers {
        wg.Add(1)
        go func(p SearchProvider) {
            defer wg.Done()
            if res, err := p.Search(ctx, query); err == nil {
                resultsChan <- res
            }
        }(provider)
    }

    go func() {
        wg.Wait()
        close(resultsChan)
    }()

    for res := range resultsChan {
        results = append(results, res...)
    }

    // Salvar cache
    if query.UseCache {
        o.cache.Set(query.Query, results, 24*time.Hour)
    }

    return results, nil
}

func (o *Orchestrator) Synthesize(ctx context.Context, results []*SearchResult, originalQuery string) (string, error) {
    // Usar LLM para sintetizar resultados
    // (Implementação similar ao synthesizer detalhado anteriormente)
    return "Síntese dos resultados...", nil
}
```

*(Implementar providers: DuckDuckGo, StackOverflow, GitHub)*

---

### FASE 6: Build & Otimização (2 dias)

#### 6.1. Makefile

```makefile
BINARY_NAME=ollama-code
BUILD_DIR=./bin

.PHONY: all build build-optimized install test clean run

all: build

build:
    @echo "Building $(BINARY_NAME)..."
    @mkdir -p $(BUILD_DIR)
    @go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/ollama-code/main.go

build-optimized:
    @echo "Building optimized..."
    @mkdir -p $(BUILD_DIR)
    @CGO_ENABLED=0 go build \
        -ldflags="-s -w -extldflags '-static'" \
        -trimpath \
        -o $(BUILD_DIR)/$(BINARY_NAME) \
        cmd/ollama-code/main.go
    @echo "Binary size: $(shell du -h $(BUILD_DIR)/$(BINARY_NAME) | cut -f1)"

install: build-optimized
    @echo "Installing to /usr/local/bin..."
    @sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
    @echo "✓ Installed!"

test:
    @go test -v ./...

run: build
    @$(BUILD_DIR)/$(BINARY_NAME) chat

clean:
    @rm -rf $(BUILD_DIR)
```

---

## 🎯 EXEMPLOS DE USO

### Modo READ-ONLY
```bash
ollama-code chat --mode readonly

Você: Analisa esse projeto
🤖: [Lê arquivos e analisa]

Você: Corrige os erros
❌ Ação bloqueada: Escrita não permitida em modo READ-ONLY
```

### Modo INTERACTIVE (padrão)
```bash
ollama-code chat

Você: Cria um servidor HTTP em Go

🤖: Vou criar o arquivo server.go

🔔 Confirmação necessária:
   Ação: Criar arquivo server.go
   Tipo: WRITE_FILE

Executar? [y/N]: y

✅ Arquivo criado: server.go
```

### Modo AUTONOMOUS
```bash
ollama-code chat --mode autonomous

Você: Cria um projeto REST completo com testes

[10:23:45] ✓ Criado: main.go (145 linhas)
[10:23:46] ✓ Criado: handlers/user.go (89 linhas)
[10:23:47] ✓ Criado: tests/user_test.go (123 linhas)
[10:23:48] ⚙️  Executando: go mod tidy
[10:23:49] ⚙️  Executando: go test ./...
[10:23:52] ✅ Todos os testes passaram
```

### Web Search
```bash
Você: Como corrigir erro "permission denied" no Docker?

🌐 Pesquisando na internet...
✓ Encontrei 3 fontes relevantes

🤖: O erro "permission denied" geralmente ocorre quando...

[Solução detalhada com exemplos]

📚 Fontes:
[1] Stack Overflow - https://stackoverflow.com/...
[2] Docker Docs - https://docs.docker.com/...
```

---

## 📊 MÉTRICAS DE PERFORMANCE

**Esperadas no hardware alvo:**
```
Startup time:     < 15ms
Memory (base):    ~10MB
Binary size:      ~8MB (otimizado)
LLM inference:    ~30-40 tokens/s (GPU)
File ops:         < 10ms
Web search:       ~2-5s (com cache: <100ms)
```

---

## ✅ CHECKLIST FINAL

- [ ] Go 1.21+ instalado
- [ ] Ollama instalado e configurado
- [ ] Modelo baixado (qwen2.5-coder:32b)
- [ ] GPU configurada (999 layers)
- [ ] Estrutura de diretórios criada
- [ ] LLM Client implementado
- [ ] Agent principal implementado
- [ ] Intent Detection funcionando
- [ ] 6+ ferramentas implementadas
- [ ] 3 modos de operação funcionando
- [ ] Web Search integrado
- [ ] Confirmation Manager funcionando
- [ ] CLI com Cobra funcionando
- [ ] Build otimizado (<10MB)
- [ ] Testes básicos passando
- [ ] README completo

---

## 🚀 BUILD E EXECUÇÃO

```bash
# Clone ou crie o projeto
mkdir ollama-code && cd ollama-code

# Desenvolver
go mod init ollama-code
go mod tidy

# Build
make build

# Executar
./bin/ollama-code chat

# Build otimizado
make build-optimized

# Instalar globalmente
make install

# Usar de qualquer lugar
ollama-code chat --mode interactive
```

---

## 📚 FUNCIONALIDADES ENTERPRISE

Para uso corporativo resiliente, veja funcionalidades adicionais em:
**[ENTERPRISE_FEATURES.md](ENTERPRISE_FEATURES.md)**

Funcionalidades críticas enterprise (adicionar após base):
1. ✅ **Checkpoints & Rewind** - Recuperação de estado
2. ✅ **Session Management** - Múltiplas sessões, resumir conversas
3. ✅ **Hierarchical Memory** - Sistema CLAUDE.md em 5 níveis
4. ✅ **Slash Commands** - 40+ comandos customizáveis
5. ✅ **Hooks System** - Pre/Post execution hooks
6. ✅ **Telemetry** - Monitoramento OpenTelemetry
7. ✅ **Sandboxing** - Isolamento Linux/macOS/Windows
8. ✅ **Diagnostics** - /doctor para health checks
9. ✅ **Background Tasks** - Execução assíncrona
10. ✅ **CI/CD Integration** - GitHub Actions, GitLab

**Estimativa adicional:** 24 dias (priorizar conforme necessidade)

---

## 🎯 PRÓXIMOS PASSOS

**Implementação Base (10-12 dias):**
1. Setup & LLM Client
2. Intent Detection
3. Tools System
4. Operation Modes
5. Web Search
6. Build & Otimização

**Funcionalidades Enterprise (24 dias):**
- Ver [ENTERPRISE_FEATURES.md](ENTERPRISE_FEATURES.md) para detalhes completos
- Priorizar: Session Management → Memory → Checkpoints → Hooks

**Pós-Implementação:**
1. Testes unitários abrangentes
2. Documentação de API completa
3. Benchmarks de performance
4. Guias de uso corporativo

---

**Estimativa total base:** 10-12 dias de desenvolvimento focado

**Estimativa total enterprise:** 34-36 dias (base + enterprise completo)

**Resultado:** CLI profissional enterprise-grade, altamente resiliente, que funciona como Claude Code, 100% local, pronto para uso corporativo diário.
