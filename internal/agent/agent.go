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
	"github.com/johnpitter/ollama-code/internal/handlers"
	"github.com/johnpitter/ollama-code/internal/intent"
	"github.com/johnpitter/ollama-code/internal/llm"
	"github.com/johnpitter/ollama-code/internal/modes"
	"github.com/johnpitter/ollama-code/internal/observability"
	"github.com/johnpitter/ollama-code/internal/ollamamd"
	"github.com/johnpitter/ollama-code/internal/session"
	"github.com/johnpitter/ollama-code/internal/skills"
	"github.com/johnpitter/ollama-code/internal/statusline"
	"github.com/johnpitter/ollama-code/internal/tools"
	"github.com/johnpitter/ollama-code/internal/websearch"
)

// Agent agente principal
type Agent struct {
	LLMClient       *llm.Client
	IntentDetector  *intent.Detector
	ToolRegistry    *tools.Registry
	CommandRegistry *commands.Registry
	SkillRegistry   *skills.Registry
	ConfirmManager  *confirmation.Manager
	WebSearch       *websearch.Orchestrator
	SessionManager  *session.Manager
	Cache           *cache.Manager
	StatusLine      *statusline.StatusLine
	OllamaContext   *ollamamd.OllamaContext
	HandlerRegistry *handlers.Registry
	Observability   *observability.Observability
	Mode            modes.OperationMode
	WorkDir         string
	History         []llm.Message
	RecentFiles     []string // Arquivos criados/modificados recentemente
	Mu              sync.Mutex

	// Colors
	ColorGreen  *color.Color
	ColorBlue   *color.Color
	ColorYellow *color.Color
	ColorRed    *color.Color
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
	// Registrar ferramentas avançadas do QA Plan
	toolRegistry.Register(tools.NewDependencyManager(cfg.WorkDir))
	toolRegistry.Register(tools.NewDocumentationGenerator(cfg.WorkDir))
	toolRegistry.Register(tools.NewSecurityScanner(cfg.WorkDir))
	toolRegistry.Register(tools.NewAdvancedRefactoring(cfg.WorkDir))
	toolRegistry.Register(tools.NewTestRunner(cfg.WorkDir))
	toolRegistry.Register(tools.NewBackgroundTaskManager(cfg.WorkDir))
	toolRegistry.Register(tools.NewPerformanceProfiler(cfg.WorkDir))

	// Registrar novas integrações
	toolRegistry.Register(tools.NewGitHelper(cfg.WorkDir))
	toolRegistry.Register(tools.NewCodeFormatter(cfg.WorkDir))

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

	// Criar HandlerRegistry
	handlerRegistry := handlers.NewRegistry()

	// Registrar handlers
	handlerRegistry.Register(intent.IntentReadFile, handlers.NewFileReadHandler())
	handlerRegistry.Register(intent.IntentWriteFile, handlers.NewFileWriteHandler())
	handlerRegistry.Register(intent.IntentExecuteCommand, handlers.NewExecuteHandler())
	handlerRegistry.Register(intent.IntentSearchCode, handlers.NewSearchHandler())
	handlerRegistry.Register(intent.IntentAnalyzeProject, handlers.NewAnalyzeHandler())
	handlerRegistry.Register(intent.IntentGitOperation, handlers.NewGitHandler())
	handlerRegistry.Register(intent.IntentWebSearch, handlers.NewWebSearchHandler())

	// Registrar default handler (Question)
	handlerRegistry.RegisterDefault(handlers.NewQuestionHandler())

	agent := &Agent{
		LLMClient:       llmClient,
		IntentDetector:  intentDetector,
		ToolRegistry:    toolRegistry,
		CommandRegistry: commands.NewRegistry(),
		SkillRegistry:   skillRegistry,
		ConfirmManager:  confirmation.NewManager(),
		WebSearch:       websearch.NewOrchestrator(),
		SessionManager:  sessionMgr,
		Cache:           cacheMgr,
		StatusLine:      statusLineMgr,
		OllamaContext:   ollamaContext,
		HandlerRegistry: handlerRegistry,
		Mode:            cfg.Mode,
		WorkDir:         cfg.WorkDir,
		History:         []llm.Message{},
		RecentFiles:     []string{},
		ColorGreen:      color.New(color.FgGreen, color.Bold),
		ColorBlue:       color.New(color.FgBlue, color.Bold),
		ColorYellow:     color.New(color.FgYellow),
		ColorRed:        color.New(color.FgRed),
	}

	return agent, nil
}

// GetSessionManager retorna o gerenciador de sessões
func (a *Agent) GetSessionManager() *session.Manager {
	return a.SessionManager
}

// GetCache retorna o gerenciador de cache
func (a *Agent) GetCache() *cache.Manager {
	return a.Cache
}

// GetCommandRegistry retorna o registry de comandos
func (a *Agent) GetCommandRegistry() *commands.Registry {
	return a.CommandRegistry
}

// GetSkillRegistry retorna o registry de skills
func (a *Agent) GetSkillRegistry() *skills.Registry {
	return a.SkillRegistry
}

// ProcessMessage processa mensagem do usuário
func (a *Agent) ProcessMessage(ctx context.Context, userMessage string) error {
	// Adicionar mensagem ao histórico
	a.Mu.Lock()
	a.History = append(a.History, llm.Message{
		Role:    "user",
		Content: userMessage,
	})
	a.Mu.Unlock()

	// Detectar intenção com histórico da conversa
	a.ColorBlue.Println("\n🔍 Detectando intenção...")

	recentFiles := a.getRecentFiles()
	detectionResult, err := a.IntentDetector.DetectWithHistory(ctx, userMessage, a.WorkDir, recentFiles, a.History)
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
	a.Mu.Lock()
	a.History = append(a.History, llm.Message{
		Role:    "assistant",
		Content: response,
	})
	a.Mu.Unlock()

	// Mostrar resposta (se não foi mostrada em streaming)
	if detectionResult.Intent != intent.IntentQuestion && response != "" {
		a.ColorGreen.Println("\n🤖 Assistente:")
		fmt.Println(response)
		fmt.Println()
	}

	return nil
}

// handleIntent processa a intenção detectada
func (a *Agent) handleIntent(ctx context.Context, result *intent.DetectionResult, userMessage string) (string, error) {
	// Atualizar DetectionResult com userMessage
	result.UserMessage = userMessage

	// Construir dependencies
	deps := a.buildDependencies()

	// Delegar para HandlerRegistry
	response, err := a.HandlerRegistry.Handle(ctx, deps, result)
	if err != nil {
		return "", err
	}

	// Atualizar recentFiles se o handler modificou
	if len(deps.RecentFiles) > len(a.RecentFiles) {
		a.Mu.Lock()
		a.RecentFiles = deps.RecentFiles
		a.Mu.Unlock()
	}

	return response, nil
}

// getRecentFiles obtém lista de arquivos recentes no diretório
func (a *Agent) getRecentFiles() []string {
	files := []string{}

	entries, err := os.ReadDir(a.WorkDir)
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
	a.Mu.Lock()
	defer a.Mu.Unlock()
	return append([]llm.Message{}, a.History...)
}

// ClearHistory limpa histórico
func (a *Agent) ClearHistory() {
	a.Mu.Lock()
	defer a.Mu.Unlock()
	a.History = []llm.Message{}
}

// SetMode altera modo de operação
func (a *Agent) SetMode(mode modes.OperationMode) {
	a.Mode = mode
}

// GetMode retorna modo atual
func (a *Agent) GetMode() modes.OperationMode {
	return a.Mode
}

// GetWorkDir retorna diretório de trabalho
func (a *Agent) GetWorkDir() string {
	return a.WorkDir
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

	a.WorkDir = absDir
	return nil
}

// AddRecentFile adiciona arquivo à lista de arquivos recentes
func (a *Agent) AddRecentFile(filePath string) {
	a.Mu.Lock()
	defer a.Mu.Unlock()

	// Adicionar no início da lista
	a.RecentFiles = append([]string{filePath}, a.RecentFiles...)

	// Manter apenas últimos 10 arquivos
	if len(a.RecentFiles) > 10 {
		a.RecentFiles = a.RecentFiles[:10]
	}
}

// GetRecentlyModifiedFiles retorna arquivos recentemente modificados
func (a *Agent) GetRecentlyModifiedFiles() []string {
	a.Mu.Lock()
	defer a.Mu.Unlock()
	return append([]string{}, a.RecentFiles...)
}

// buildDependencies cria Dependencies struct a partir do Agent
func (a *Agent) buildDependencies() *handlers.Dependencies {
	// Converter history para handlers.Message
	history := a.GetHistory()
	handlerHistory := make([]handlers.Message, len(history))
	for i, msg := range history {
		handlerHistory[i] = handlers.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	return &handlers.Dependencies{
		ToolRegistry:    handlers.NewToolRegistryAdapter(a.ToolRegistry),
		CommandRegistry: handlers.NewCommandRegistryAdapter(a.CommandRegistry),
		SkillRegistry:   handlers.NewSkillRegistryAdapter(a.SkillRegistry),
		ConfirmManager:  handlers.NewConfirmationManagerAdapter(a.ConfirmManager),
		SessionManager:  handlers.NewSessionManagerAdapter(a.SessionManager),
		CacheManager:    handlers.NewCacheManagerAdapter(a.Cache),
		LLMClient:       handlers.NewLLMClientAdapter(a.LLMClient),
		WebSearch:       handlers.NewWebSearchClientAdapter(a.WebSearch),
		IntentDetector:  handlers.NewIntentDetectorAdapter(a.IntentDetector),
		Mode:            handlers.NewOperationModeAdapter(a.Mode),
		WorkDir:         a.WorkDir,
		History:         handlerHistory,
		RecentFiles:     a.GetRecentlyModifiedFiles(),
	}
}
