package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/johnpitter/ollama-code/internal/commands"
	"github.com/johnpitter/ollama-code/internal/confirmation"
	"github.com/johnpitter/ollama-code/internal/intent"
	"github.com/johnpitter/ollama-code/internal/llm"
	"github.com/johnpitter/ollama-code/internal/modes"
	"github.com/johnpitter/ollama-code/internal/tools"
	"github.com/johnpitter/ollama-code/internal/websearch"
)

// Agent agente principal
type Agent struct {
	llmClient       *llm.Client
	intentDetector  *intent.Detector
	toolRegistry    *tools.Registry
	commandRegistry *commands.Registry
	confirmManager  *confirmation.Manager
	webSearch       *websearch.Orchestrator
	mode            modes.OperationMode
	workDir         string
	history         []llm.Message
	mu              sync.Mutex

	// Colors
	colorGreen  *color.Color
	colorBlue   *color.Color
	colorYellow *color.Color
	colorRed    *color.Color
}

// Config configuração do agente
type Config struct {
	OllamaURL   string
	Model       string
	Mode        modes.OperationMode
	WorkDir     string
	Temperature float64
	MaxTokens   int
}

// NewAgent cria novo agente
func NewAgent(cfg Config) (*Agent, error) {
	// Default values
	if cfg.OllamaURL == "" {
		cfg.OllamaURL = "http://localhost:11434"
	}
	if cfg.Model == "" {
		cfg.Model = "qwen2.5-coder:32b-instruct-q6_K"
	}
	if cfg.WorkDir == "" {
		cfg.WorkDir, _ = os.Getwd()
	}
	if cfg.Mode == "" {
		cfg.Mode = modes.ModeInteractive
	}

	// Criar LLM client
	llmClient := llm.NewClient(cfg.OllamaURL, cfg.Model)

	// Criar detector de intenções
	intentDetector := intent.NewDetector(llmClient)

	// Criar registry de ferramentas
	toolRegistry := tools.NewRegistry()

	// Registrar ferramentas
	toolRegistry.Register(tools.NewFileReader(cfg.WorkDir))
	toolRegistry.Register(tools.NewFileWriter(cfg.WorkDir))
	toolRegistry.Register(tools.NewCommandExecutor(cfg.WorkDir, 60*time.Second))
	toolRegistry.Register(tools.NewCodeSearcher(cfg.WorkDir))
	toolRegistry.Register(tools.NewProjectAnalyzer(cfg.WorkDir))
	toolRegistry.Register(tools.NewGitOperations(cfg.WorkDir))

	agent := &Agent{
		llmClient:       llmClient,
		intentDetector:  intentDetector,
		toolRegistry:    toolRegistry,
		commandRegistry: commands.NewRegistry(),
		confirmManager:  confirmation.NewManager(),
		webSearch:       websearch.NewOrchestrator(),
		mode:            cfg.Mode,
		workDir:         cfg.WorkDir,
		history:         []llm.Message{},
		colorGreen:      color.New(color.FgGreen, color.Bold),
		colorBlue:       color.New(color.FgBlue, color.Bold),
		colorYellow:     color.New(color.FgYellow),
		colorRed:        color.New(color.FgRed),
	}

	return agent, nil
}

// GetCommandRegistry retorna o registry de comandos
func (a *Agent) GetCommandRegistry() *commands.Registry {
	return a.commandRegistry
}

// ProcessMessage processa mensagem do usuário
func (a *Agent) ProcessMessage(ctx context.Context, userMessage string) error {
	// Adicionar mensagem ao histórico
	a.mu.Lock()
	a.history = append(a.history, llm.Message{
		Role:    "user",
		Content: userMessage,
	})
	a.mu.Unlock()

	// Detectar intenção
	a.colorBlue.Println("\n🔍 Detectando intenção...")

	recentFiles := a.getRecentFiles()
	detectionResult, err := a.intentDetector.Detect(ctx, userMessage, a.workDir, recentFiles)
	if err != nil {
		return fmt.Errorf("detect intent: %w", err)
	}

	fmt.Printf("Intenção: %s (confiança: %.0f%%)\n", detectionResult.Intent, detectionResult.Confidence*100)

	// Processar de acordo com a intenção
	response, err := a.handleIntent(ctx, detectionResult, userMessage)
	if err != nil {
		return fmt.Errorf("handle intent: %w", err)
	}

	// Adicionar resposta ao histórico
	a.mu.Lock()
	a.history = append(a.history, llm.Message{
		Role:    "assistant",
		Content: response,
	})
	a.mu.Unlock()

	// Mostrar resposta
	a.colorGreen.Println("\n🤖 Assistente:")
	fmt.Println(response)
	fmt.Println()

	return nil
}

// handleIntent processa a intenção detectada
func (a *Agent) handleIntent(ctx context.Context, result *intent.DetectionResult, userMessage string) (string, error) {
	switch result.Intent {
	case intent.IntentReadFile:
		return a.handleReadFile(ctx, result)

	case intent.IntentWriteFile:
		return a.handleWriteFile(ctx, result, userMessage)

	case intent.IntentExecuteCommand:
		return a.handleExecuteCommand(ctx, result)

	case intent.IntentSearchCode:
		return a.handleSearchCode(ctx, result)

	case intent.IntentAnalyzeProject:
		return a.handleAnalyzeProject(ctx, result)

	case intent.IntentGitOperation:
		return a.handleGitOperation(ctx, result)

	case intent.IntentWebSearch:
		return a.handleWebSearch(ctx, result)

	case intent.IntentQuestion:
		return a.handleQuestion(ctx, userMessage)

	default:
		return a.handleQuestion(ctx, userMessage)
	}
}

// getRecentFiles obtém lista de arquivos recentes no diretório
func (a *Agent) getRecentFiles() []string {
	files := []string{}

	entries, err := os.ReadDir(a.workDir)
	if err != nil {
		return files
	}

	for _, entry := range entries {
		if entry.IsDir() || entry.Name()[0] == '.' {
			continue
		}
		files = append(files, entry.Name())
		if len(files) >= 10 {
			break
		}
	}

	return files
}

// GetHistory retorna histórico de mensagens
func (a *Agent) GetHistory() []llm.Message {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]llm.Message{}, a.history...)
}

// ClearHistory limpa histórico
func (a *Agent) ClearHistory() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.history = []llm.Message{}
}

// SetMode altera modo de operação
func (a *Agent) SetMode(mode modes.OperationMode) {
	a.mode = mode
}

// GetMode retorna modo atual
func (a *Agent) GetMode() modes.OperationMode {
	return a.mode
}

// GetWorkDir retorna diretório de trabalho
func (a *Agent) GetWorkDir() string {
	return a.workDir
}

// SetWorkDir altera diretório de trabalho
func (a *Agent) SetWorkDir(dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	info, err := os.Stat(absDir)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", absDir)
	}

	a.workDir = absDir
	return nil
}
