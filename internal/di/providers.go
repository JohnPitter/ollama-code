package di

import (
	"fmt"
	"os"
	"time"

	"github.com/johnpitter/ollama-code/internal/cache"
	"github.com/johnpitter/ollama-code/internal/commands"
	"github.com/johnpitter/ollama-code/internal/confirmation"
	"github.com/johnpitter/ollama-code/internal/diff"
	"github.com/johnpitter/ollama-code/internal/handlers"
	"github.com/johnpitter/ollama-code/internal/intent"
	"github.com/johnpitter/ollama-code/internal/llm"
	"github.com/johnpitter/ollama-code/internal/modes"
	"github.com/johnpitter/ollama-code/internal/multimodel"
	"github.com/johnpitter/ollama-code/internal/observability"
	"github.com/johnpitter/ollama-code/internal/ollamamd"
	"github.com/johnpitter/ollama-code/internal/session"
	"github.com/johnpitter/ollama-code/internal/skills"
	"github.com/johnpitter/ollama-code/internal/statusline"
	"github.com/johnpitter/ollama-code/internal/subagent"
	"github.com/johnpitter/ollama-code/internal/todos"
	"github.com/johnpitter/ollama-code/internal/tools"
	"github.com/johnpitter/ollama-code/internal/websearch"
)

// Config representa a configuração da aplicação
type Config struct {
	OllamaURL            string
	Model                string
	Mode                 modes.OperationMode
	WorkDir              string
	Temperature          float64
	MaxTokens            int
	EnableSessions       bool
	EnableCache          bool
	EnableStatusLine     bool
	EnableObservability  bool
	EnableTodos          bool
	EnableMultiModel     bool
	CacheTTL             time.Duration
	ObservabilityConfig  observability.LoggerConfig
}

// ProvideLLMClient fornece LLM client
func ProvideLLMClient(cfg *Config) *llm.Client {
	return llm.NewClient(cfg.OllamaURL, cfg.Model)
}

// ProvideIntentDetector fornece detector de intenções
func ProvideIntentDetector(client *llm.Client) *intent.Detector {
	return intent.NewDetector(client)
}

// ProvideSessionManager fornece session manager (opcional)
func ProvideSessionManager(cfg *Config) *session.Manager {
	if !cfg.EnableSessions {
		return nil
	}
	homeDir, _ := os.UserHomeDir()
	return session.NewManager(homeDir)
}

// ProvideCacheManager fornece cache manager (opcional)
func ProvideCacheManager(cfg *Config) *cache.Manager {
	if !cfg.EnableCache {
		return nil
	}
	return cache.NewManager(cfg.CacheTTL)
}

// ProvideStatusLine fornece status line (opcional)
func ProvideStatusLine(cfg *Config) *statusline.StatusLine {
	if !cfg.EnableStatusLine {
		return nil
	}

	maxTokens := cfg.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	return statusline.New(statusline.Config{
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

// ProvideToolRegistry fornece registry de ferramentas
func ProvideToolRegistry(cfg *Config) *tools.Registry {
	registry := tools.NewRegistry()

	// Ferramentas básicas
	registry.Register(tools.NewFileReader(cfg.WorkDir))
	registry.Register(tools.NewFileWriter(cfg.WorkDir))
	registry.Register(tools.NewCommandExecutor(cfg.WorkDir, 60*time.Second))
	registry.Register(tools.NewCodeSearcher(cfg.WorkDir))
	registry.Register(tools.NewProjectAnalyzer(cfg.WorkDir))
	registry.Register(tools.NewGitOperations(cfg.WorkDir))

	// Ferramentas avançadas
	registry.Register(tools.NewDependencyManager(cfg.WorkDir))
	registry.Register(tools.NewDocumentationGenerator(cfg.WorkDir))
	registry.Register(tools.NewSecurityScanner(cfg.WorkDir))
	registry.Register(tools.NewAdvancedRefactoring(cfg.WorkDir))
	registry.Register(tools.NewTestRunner(cfg.WorkDir))
	registry.Register(tools.NewBackgroundTaskManager(cfg.WorkDir))
	registry.Register(tools.NewPerformanceProfiler(cfg.WorkDir))

	// Novas integrações
	registry.Register(tools.NewGitHelper(cfg.WorkDir))
	registry.Register(tools.NewCodeFormatter(cfg.WorkDir))

	return registry
}

// ProvideCommandRegistry fornece registry de comandos
func ProvideCommandRegistry() *commands.Registry {
	return commands.NewRegistry()
}

// ProvideSkillRegistry fornece registry de skills
func ProvideSkillRegistry() *skills.Registry {
	registry := skills.NewRegistry()

	registry.Register(skills.NewResearchSkill())
	registry.Register(skills.NewAPISkill())
	registry.Register(skills.NewCodeAnalysisSkill())

	return registry
}

// ProvideConfirmationManager fornece confirmation manager
func ProvideConfirmationManager() *confirmation.Manager {
	return confirmation.NewManager()
}

// ProvideWebSearchOrchestrator fornece web search orchestrator
func ProvideWebSearchOrchestrator() *websearch.Orchestrator {
	return websearch.NewOrchestrator()
}

// ProvideOllamaContext fornece contexto OLLAMA.md
func ProvideOllamaContext(cfg *Config) (*ollamamd.OllamaContext, error) {
	loader := ollamamd.NewLoader(cfg.WorkDir)
	ollamaContext, err := loader.Load()
	if err != nil {
		// Log mas não falhe - OLLAMA.md é opcional
		fmt.Printf("⚠️  Aviso: Não foi possível carregar OLLAMA.md: %v\n", err)
		return &ollamamd.OllamaContext{}, nil
	}

	if len(ollamaContext.Files) > 0 {
		fmt.Printf("📋 Carregados %d arquivo(s) OLLAMA.md\n", len(ollamaContext.Files))
	}

	return ollamaContext, nil
}

// Handler Providers

// ProvideFileReadHandler fornece file read handler
func ProvideFileReadHandler() *handlers.FileReadHandler {
	return handlers.NewFileReadHandler()
}

// ProvideFileWriteHandler fornece file write handler
func ProvideFileWriteHandler() *handlers.FileWriteHandler {
	return handlers.NewFileWriteHandler()
}

// ProvideSearchHandler fornece search handler
func ProvideSearchHandler() *handlers.SearchHandler {
	return handlers.NewSearchHandler()
}

// ProvideExecuteHandler fornece execute handler
func ProvideExecuteHandler() *handlers.ExecuteHandler {
	return handlers.NewExecuteHandler()
}

// ProvideQuestionHandler fornece question handler
func ProvideQuestionHandler() *handlers.QuestionHandler {
	return handlers.NewQuestionHandler()
}

// ProvideGitHandler fornece git handler
func ProvideGitHandler() *handlers.GitHandler {
	return handlers.NewGitHandler()
}

// ProvideAnalyzeHandler fornece analyze handler
func ProvideAnalyzeHandler() *handlers.AnalyzeHandler {
	return handlers.NewAnalyzeHandler()
}

// ProvideWebSearchHandler fornece websearch handler
func ProvideWebSearchHandler() *handlers.WebSearchHandler {
	return handlers.NewWebSearchHandler()
}

// ProvideMode fornece o modo de operação
func ProvideMode(cfg *Config) modes.OperationMode {
	return cfg.Mode
}

// ProvideWorkDir fornece o diretório de trabalho
func ProvideWorkDir(cfg *Config) string {
	return cfg.WorkDir
}

// ProvideHandlerRegistry fornece handler registry
func ProvideHandlerRegistry(
	fileReadHandler *handlers.FileReadHandler,
	fileWriteHandler *handlers.FileWriteHandler,
	searchHandler *handlers.SearchHandler,
	executeHandler *handlers.ExecuteHandler,
	questionHandler *handlers.QuestionHandler,
	gitHandler *handlers.GitHandler,
	analyzeHandler *handlers.AnalyzeHandler,
	webSearchHandler *handlers.WebSearchHandler,
) *handlers.Registry {
	registry := handlers.NewRegistry()

	// Registrar handlers
	registry.Register(intent.IntentReadFile, fileReadHandler)
	registry.Register(intent.IntentWriteFile, fileWriteHandler)
	registry.Register(intent.IntentExecuteCommand, executeHandler)
	registry.Register(intent.IntentSearchCode, searchHandler)
	registry.Register(intent.IntentAnalyzeProject, analyzeHandler)
	registry.Register(intent.IntentGitOperation, gitHandler)
	registry.Register(intent.IntentWebSearch, webSearchHandler)

	// Default handler
	registry.RegisterDefault(questionHandler)

	return registry
}

// ProvideObservability fornece sistema de observabilidade
func ProvideObservability(cfg *Config) *observability.Observability {
	if !cfg.EnableObservability {
		return nil
	}

	// Usar config personalizado se fornecido, senão usar padrão
	loggerConfig := cfg.ObservabilityConfig
	if loggerConfig.Level == "" {
		loggerConfig = observability.LoggerConfig{
			Level:     observability.LogLevelInfo,
			Format:    observability.LogFormatText,
			AddSource: false,
		}
	}

	return observability.New(loggerConfig)
}

// ProvideTodoManager fornece TODO manager
func ProvideTodoManager(cfg *Config) *todos.Manager {
	if !cfg.EnableTodos {
		return nil
	}

	// Tentar usar file storage, fallback para memory storage
	storage, err := todos.DefaultFileStorage()
	if err != nil {
		fmt.Printf("⚠️  Aviso: Usando TODO storage em memória: %v\n", err)
		return todos.NewManager()
	}

	return todos.NewManagerWithStorage(storage)
}

// ProvideDiffer fornece diff manager
func ProvideDiffer() *diff.Differ {
	return diff.NewDiffer()
}

// ProvidePreviewer fornece preview manager
func ProvidePreviewer() *diff.Previewer {
	return diff.NewPreviewer()
}

// ProvideSubagentExecutor fornece executor de subagents
func ProvideSubagentExecutor(cfg *Config) *subagent.Executor {
	return subagent.NewExecutor(cfg.OllamaURL)
}

// ProvideSubagentManager fornece manager de subagents
func ProvideSubagentManager(executor *subagent.Executor) *subagent.Manager {
	// Criar ExecutorFunc a partir do Executor
	executorFunc := executor.CreateExecutorFunc()
	return subagent.NewManager(executorFunc)
}

// ProvideMultiModelRouter fornece router de multi-model
func ProvideMultiModelRouter(cfg *Config) *multimodel.Router {
	var mmConfig *multimodel.Config

	if cfg.EnableMultiModel {
		mmConfig = multimodel.DefaultConfig()
	} else {
		// Multi-model desabilitado - criar config que sempre usa default
		mmConfig = multimodel.NewConfig()
		mmConfig.DefaultModel = multimodel.ModelSpec{
			Name:        cfg.Model,
			MaxTokens:   cfg.MaxTokens,
			Temperature: cfg.Temperature,
			Description: "Default model",
		}
		mmConfig.Disable()
	}

	return multimodel.NewRouter(cfg.OllamaURL, mmConfig)
}
