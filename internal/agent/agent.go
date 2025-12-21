package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fatih/color"
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

// Agent agente principal
type Agent struct {
	llmClient       *llm.Client
	intentDetector  *intent.Detector
	toolRegistry    *tools.Registry
	commandRegistry *commands.Registry
	skillRegistry   *skills.Registry
	confirmManager  *confirmation.Manager
	webSearch       *websearch.Orchestrator
	sessionManager  *session.Manager
	cache           *cache.Manager
	statusLine      *statusline.StatusLine
	ollamaContext   *ollamamd.OllamaContext
	mode            modes.OperationMode
	workDir         string
	history         []llm.Message
	recentFiles     []string // Arquivos criados/modificados recentemente
	mu              sync.Mutex

	// Colors
	colorGreen  *color.Color
	colorBlue   *color.Color
	colorYellow *color.Color
	colorRed    *color.Color
}

// Config configuração do agente
type Config struct {
	OllamaURL        string
	Model            string
	Mode             modes.OperationMode
	WorkDir          string
	Temperature      float64
	MaxTokens        int
	EnableSessions   bool
	EnableCache      bool
	EnableStatusLine bool
	CacheTTL         time.Duration
}

// NewAgent cria novo agente
func NewAgent(cfg Config) (*Agent, error) {
	// Default values
	if cfg.OllamaURL == "" {
		cfg.OllamaURL = "http://localhost:11434"
	}
	if cfg.Model == "" {
		cfg.Model = "qwen2.5-coder:7b"
	}
	if cfg.WorkDir == "" {
		cfg.WorkDir, _ = os.Getwd()
	}
	if cfg.Mode == "" {
		cfg.Mode = modes.ModeInteractive
	}
	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = 5 * time.Minute // Default 5 minutes
	}

	// Criar LLM client
	llmClient := llm.NewClient(cfg.OllamaURL, cfg.Model)

	// Criar detector de intenções
	intentDetector := intent.NewDetector(llmClient)

	// Session manager (opcional)
	var sessionMgr *session.Manager
	if cfg.EnableSessions {
		homeDir, _ := os.UserHomeDir()
		sessionMgr = session.NewManager(homeDir)
	}

	// Cache (opcional)
	var cacheMgr *cache.Manager
	if cfg.EnableCache {
		cacheMgr = cache.NewManager(cfg.CacheTTL)
	}

	// Status Line (opcional)
	var statusLineMgr *statusline.StatusLine
	if cfg.EnableStatusLine {
		maxTokens := cfg.MaxTokens
		if maxTokens == 0 {
			maxTokens = 4096
		}
		statusLineMgr = statusline.New(statusline.Config{
			Model:      cfg.Model,
			Mode:       string(cfg.Mode),
			WorkDir:    cfg.WorkDir,
			MaxTokens:  maxTokens,
			ShowTokens: true,
			ShowTime:   true,
			ShowTask:   true,
			Enabled:    true,
		})
	}

	// Criar registry de ferramentas
	toolRegistry := tools.NewRegistry()

	// Registrar ferramentas
	toolRegistry.Register(tools.NewFileReader(cfg.WorkDir))
	toolRegistry.Register(tools.NewFileWriter(cfg.WorkDir))
	toolRegistry.Register(tools.NewCommandExecutor(cfg.WorkDir, 60*time.Second))
	toolRegistry.Register(tools.NewCodeSearcher(cfg.WorkDir))
	toolRegistry.Register(tools.NewProjectAnalyzer(cfg.WorkDir))
	toolRegistry.Register(tools.NewGitOperations(cfg.WorkDir))

	// Criar registry de skills
	skillRegistry := skills.NewRegistry()

	// Registrar skills especializados
	skillRegistry.Register(skills.NewResearchSkill())
	skillRegistry.Register(skills.NewAPISkill())
	skillRegistry.Register(skills.NewCodeAnalysisSkill())

	// Carregar contexto OLLAMA.md hierárquico
	ollamaMDLoader := ollamamd.NewLoader(cfg.WorkDir)
	ollamaContext, err := ollamaMDLoader.Load()
	if err != nil {
		// Log mas não falhe - OLLAMA.md é opcional
		fmt.Printf("⚠️  Aviso: Não foi possível carregar OLLAMA.md: %v\n", err)
	} else if len(ollamaContext.Files) > 0 {
		fmt.Printf("📋 Carregados %d arquivo(s) OLLAMA.md\n", len(ollamaContext.Files))
	}

	agent := &Agent{
		llmClient:       llmClient,
		intentDetector:  intentDetector,
		toolRegistry:    toolRegistry,
		commandRegistry: commands.NewRegistry(),
		skillRegistry:   skillRegistry,
		confirmManager:  confirmation.NewManager(),
		webSearch:       websearch.NewOrchestrator(),
		sessionManager:  sessionMgr,
		cache:           cacheMgr,
		statusLine:      statusLineMgr,
		ollamaContext:   ollamaContext,
		mode:            cfg.Mode,
		workDir:         cfg.WorkDir,
		history:         []llm.Message{},
		recentFiles:     []string{},
		colorGreen:      color.New(color.FgGreen, color.Bold),
		colorBlue:       color.New(color.FgBlue, color.Bold),
		colorYellow:     color.New(color.FgYellow),
		colorRed:        color.New(color.FgRed),
	}

	return agent, nil
}

// GetSessionManager retorna o gerenciador de sessões
func (a *Agent) GetSessionManager() *session.Manager {
	return a.sessionManager
}

// GetCache retorna o gerenciador de cache
func (a *Agent) GetCache() *cache.Manager {
	return a.cache
}

// GetCommandRegistry retorna o registry de comandos
func (a *Agent) GetCommandRegistry() *commands.Registry {
	return a.commandRegistry
}

// GetSkillRegistry retorna o registry de skills
func (a *Agent) GetSkillRegistry() *skills.Registry {
	return a.skillRegistry
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

	// Detectar intenção com histórico da conversa
	a.colorBlue.Println("\n🔍 Detectando intenção...")

	recentFiles := a.getRecentFiles()
	detectionResult, err := a.intentDetector.DetectWithHistory(ctx, userMessage, a.workDir, recentFiles, a.history)
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

	// Mostrar resposta (se não foi mostrada em streaming)
	if detectionResult.Intent != intent.IntentQuestion && response != "" {
		a.colorGreen.Println("\n🤖 Assistente:")
		fmt.Println(response)
		fmt.Println()
	}

	return nil
}

// handleIntent processa a intenção detectada
func (a *Agent) handleIntent(ctx context.Context, result *intent.DetectionResult, userMessage string) (string, error) {
	switch result.Intent {
	case intent.IntentReadFile:
		return a.handleReadFile(ctx, result, userMessage)

	case intent.IntentWriteFile:
		return a.handleWriteFile(ctx, result, userMessage)

	case intent.IntentExecuteCommand:
		return a.handleExecuteCommand(ctx, result)

	case intent.IntentSearchCode:
		return a.handleSearchCode(ctx, result)

	case intent.IntentAnalyzeProject:
		return a.handleAnalyzeProject(ctx, result)

	case intent.IntentGitOperation:
		return a.handleGitOperation(ctx, result, userMessage)

	case intent.IntentWebSearch:
		return a.handleWebSearch(ctx, result, userMessage)

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

// AddRecentFile adiciona arquivo à lista de arquivos recentes
func (a *Agent) AddRecentFile(filePath string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Adicionar no início da lista
	a.recentFiles = append([]string{filePath}, a.recentFiles...)

	// Manter apenas últimos 10 arquivos
	if len(a.recentFiles) > 10 {
		a.recentFiles = a.recentFiles[:10]
	}
}

// GetRecentlyModifiedFiles retorna arquivos recentemente modificados
func (a *Agent) GetRecentlyModifiedFiles() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]string{}, a.recentFiles...)
}
