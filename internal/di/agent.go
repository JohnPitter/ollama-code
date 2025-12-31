package di

import (
	"sync"

	"github.com/fatih/color"
	"github.com/johnpitter/ollama-code/internal/agent"
	"github.com/johnpitter/ollama-code/internal/llm"
)

// InitializeAgent inicializa o Agent com todas as dependências usando Manual DI
func InitializeAgent(cfg *Config) (*agent.Agent, error) {
	// Core dependencies
	llmClient := ProvideLLMClient(cfg)
	intentDetector := ProvideIntentDetector(llmClient)

	// Managers (opcionais)
	sessionManager := ProvideSessionManager(cfg)
	cacheManager := ProvideCacheManager(cfg)
	statusLine := ProvideStatusLine(cfg)
	todoManager := ProvideTodoManager(cfg)
	differ := ProvideDiffer()
	previewer := ProvidePreviewer()

	// Subagent System
	subagentExecutor := ProvideSubagentExecutor(cfg)
	subagentManager := ProvideSubagentManager(subagentExecutor)

	// Multi-Model System
	multiModelRouter := ProvideMultiModelRouter(cfg)

	// Ollama context
	ollamaContext, err := ProvideOllamaContext(cfg)
	if err != nil {
		// Já logado dentro do provider
		// OLLAMA.md é opcional, continuamos mesmo com erro
	}

	// Registries
	toolRegistry := ProvideToolRegistry(cfg)
	commandRegistry := ProvideCommandRegistry()
	skillRegistry := ProvideSkillRegistry()

	// Outros managers
	confirmManager := ProvideConfirmationManager()
	webSearch := ProvideWebSearchOrchestrator()

	// Handlers
	fileReadHandler := ProvideFileReadHandler()
	fileWriteHandler := ProvideFileWriteHandler()
	searchHandler := ProvideSearchHandler()
	executeHandler := ProvideExecuteHandler()
	questionHandler := ProvideQuestionHandler()
	gitHandler := ProvideGitHandler()
	analyzeHandler := ProvideAnalyzeHandler()
	webSearchHandler := ProvideWebSearchHandler()

	// Handler registry
	handlerRegistry := ProvideHandlerRegistry(
		fileReadHandler,
		fileWriteHandler,
		searchHandler,
		executeHandler,
		questionHandler,
		gitHandler,
		analyzeHandler,
		webSearchHandler,
	)

	// Criar Agent com todas as dependências
	agentInstance := &agent.Agent{
		LLMClient:        llmClient,
		IntentDetector:   intentDetector,
		ToolRegistry:     toolRegistry,
		CommandRegistry:  commandRegistry,
		SkillRegistry:    skillRegistry,
		ConfirmManager:   confirmManager,
		WebSearch:        webSearch,
		SessionManager:   sessionManager,
		Cache:            cacheManager,
		StatusLine:       statusLine,
		OllamaContext:    ollamaContext,
		HandlerRegistry:  handlerRegistry,
		TodoManager:      todoManager,
		Differ:           differ,
		Previewer:        previewer,
		SubagentManager:  subagentManager,
		MultiModelRouter: multiModelRouter,
		Mode:             cfg.Mode,
		WorkDir:          cfg.WorkDir,
		History:          []llm.Message{},
		RecentFiles:      []string{},
		Mu:               sync.Mutex{},
		ColorGreen:       color.New(color.FgGreen, color.Bold),
		ColorBlue:        color.New(color.FgBlue, color.Bold),
		ColorYellow:      color.New(color.FgYellow),
		ColorRed:         color.New(color.FgRed),
	}

	return agentInstance, nil
}
