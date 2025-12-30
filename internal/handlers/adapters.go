package handlers

import (
	"context"

	"github.com/johnpitter/ollama-code/internal/cache"
	"github.com/johnpitter/ollama-code/internal/commands"
	"github.com/johnpitter/ollama-code/internal/confirmation"
	"github.com/johnpitter/ollama-code/internal/intent"
	"github.com/johnpitter/ollama-code/internal/llm"
	"github.com/johnpitter/ollama-code/internal/modes"
	"github.com/johnpitter/ollama-code/internal/session"
	"github.com/johnpitter/ollama-code/internal/skills"
	"github.com/johnpitter/ollama-code/internal/todos"
	"github.com/johnpitter/ollama-code/internal/tools"
	"github.com/johnpitter/ollama-code/internal/websearch"
)

// ToolRegistryAdapter adapta tools.Registry para handlers.ToolRegistry
type ToolRegistryAdapter struct {
	registry *tools.Registry
}

func NewToolRegistryAdapter(registry *tools.Registry) *ToolRegistryAdapter {
	return &ToolRegistryAdapter{registry: registry}
}

func (a *ToolRegistryAdapter) Execute(ctx context.Context, toolName string, params map[string]interface{}) (ToolResult, error) {
	result, err := a.registry.Execute(ctx, toolName, params)
	if err != nil {
		return ToolResult{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	return ToolResult{
		Success: result.Success,
		Message: result.Message,
		Error:   result.Error,
		Data:    result.Data,
	}, nil
}

func (a *ToolRegistryAdapter) Get(name string) (interface{}, error) {
	tool, err := a.registry.Get(name)
	if err != nil {
		return nil, err
	}
	return tool, nil
}

// CommandRegistryAdapter adapta commands.Registry para handlers.CommandRegistry
type CommandRegistryAdapter struct {
	registry *commands.Registry
}

func NewCommandRegistryAdapter(registry *commands.Registry) *CommandRegistryAdapter {
	return &CommandRegistryAdapter{registry: registry}
}

func (a *CommandRegistryAdapter) Execute(ctx context.Context, cmdName string, args []string) (string, error) {
	return a.registry.Execute(ctx, cmdName, args)
}

func (a *CommandRegistryAdapter) IsCommand(text string) bool {
	return a.registry.IsCommand(text)
}

// SkillRegistryAdapter adapta skills.Registry para handlers.SkillRegistry
type SkillRegistryAdapter struct {
	registry *skills.Registry
}

func NewSkillRegistryAdapter(registry *skills.Registry) *SkillRegistryAdapter {
	return &SkillRegistryAdapter{registry: registry}
}

func (a *SkillRegistryAdapter) FindSkill(ctx context.Context, message string) (interface{}, error) {
	// skills.Registry pode não ter FindSkill, então retornamos nil
	// Este é um stub para manter compatibilidade
	return nil, nil
}

// ConfirmationManagerAdapter adapta confirmation.Manager para handlers.ConfirmationManager
type ConfirmationManagerAdapter struct {
	manager *confirmation.Manager
}

func NewConfirmationManagerAdapter(manager *confirmation.Manager) *ConfirmationManagerAdapter {
	return &ConfirmationManagerAdapter{manager: manager}
}

func (a *ConfirmationManagerAdapter) Confirm(message string) (bool, error) {
	return a.manager.Confirm("Confirmação", message)
}

func (a *ConfirmationManagerAdapter) ConfirmWithPreview(message, preview string) (bool, error) {
	return a.manager.ConfirmWithPreview(message, preview)
}

// SessionManagerAdapter adapta session.Manager para handlers.SessionManager
type SessionManagerAdapter struct {
	manager *session.Manager
}

func NewSessionManagerAdapter(manager *session.Manager) *SessionManagerAdapter {
	return &SessionManagerAdapter{manager: manager}
}

func (a *SessionManagerAdapter) SaveMessage(role, content string) error {
	if a.manager == nil {
		return nil // Session manager é opcional
	}
	// session.Manager pode não ter SaveMessage, stub
	return nil
}

// CacheManagerAdapter adapta cache.Manager para handlers.CacheManager
type CacheManagerAdapter struct {
	manager *cache.Manager
}

func NewCacheManagerAdapter(manager *cache.Manager) *CacheManagerAdapter {
	return &CacheManagerAdapter{manager: manager}
}

func (a *CacheManagerAdapter) Get(key string) (interface{}, bool) {
	if a.manager == nil {
		return nil, false
	}
	return a.manager.Get(key)
}

func (a *CacheManagerAdapter) Set(key string, value interface{}) {
	if a.manager != nil {
		a.manager.Set(key, value)
	}
}

// TodoManagerAdapter adapta todos.Manager para handlers.TodoManager
type TodoManagerAdapter struct {
	manager *todos.Manager
}

func NewTodoManagerAdapter(manager *todos.Manager) *TodoManagerAdapter {
	return &TodoManagerAdapter{manager: manager}
}

func (a *TodoManagerAdapter) Add(content, activeForm string) (string, error) {
	if a.manager == nil {
		return "", nil // TODO manager é opcional
	}
	return a.manager.Add(content, activeForm)
}

func (a *TodoManagerAdapter) Update(id string, status interface{}) error {
	if a.manager == nil {
		return nil
	}
	todoStatus, ok := status.(todos.TodoStatus)
	if !ok {
		return nil
	}
	return a.manager.Update(id, todoStatus)
}

func (a *TodoManagerAdapter) Complete(id string) error {
	if a.manager == nil {
		return nil
	}
	return a.manager.Complete(id)
}

func (a *TodoManagerAdapter) SetInProgress(id string) error {
	if a.manager == nil {
		return nil
	}
	return a.manager.SetInProgress(id)
}

func (a *TodoManagerAdapter) List() interface{} {
	if a.manager == nil {
		return nil
	}
	return a.manager.List()
}

func (a *TodoManagerAdapter) ListByStatus(status interface{}) interface{} {
	if a.manager == nil {
		return nil
	}
	todoStatus, ok := status.(todos.TodoStatus)
	if !ok {
		return nil
	}
	return a.manager.ListByStatus(todoStatus)
}

func (a *TodoManagerAdapter) Summary() map[interface{}]int {
	if a.manager == nil {
		return nil
	}
	summary := a.manager.Summary()
	result := make(map[interface{}]int)
	for k, v := range summary {
		result[k] = v
	}
	return result
}

func (a *TodoManagerAdapter) Clear() error {
	if a.manager == nil {
		return nil
	}
	return a.manager.Clear()
}

func (a *TodoManagerAdapter) Delete(id string) error {
	if a.manager == nil {
		return nil
	}
	return a.manager.Delete(id)
}

func (a *TodoManagerAdapter) Count() int {
	if a.manager == nil {
		return 0
	}
	return a.manager.Count()
}

// LLMClientAdapter adapta llm.Client para handlers.LLMClient
type LLMClientAdapter struct {
	client *llm.Client
}

func NewLLMClientAdapter(client *llm.Client) *LLMClientAdapter {
	return &LLMClientAdapter{client: client}
}

func (a *LLMClientAdapter) Complete(ctx context.Context, prompt string) (string, error) {
	messages := []llm.Message{
		{Role: "user", Content: prompt},
	}
	return a.client.Complete(ctx, messages, nil)
}

func (a *LLMClientAdapter) CompleteWithHistory(ctx context.Context, messages []Message) (string, error) {
	llmMessages := make([]llm.Message, len(messages))
	for i, msg := range messages {
		llmMessages[i] = llm.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
	return a.client.Complete(ctx, llmMessages, nil)
}

func (a *LLMClientAdapter) CompleteStreaming(ctx context.Context, messages []Message, opts interface{}, callback func(string)) (string, error) {
	llmMessages := make([]llm.Message, len(messages))
	for i, msg := range messages {
		llmMessages[i] = llm.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Se opts for *llm.CompletionOptions, converter
	var completionOpts *llm.CompletionOptions
	if o, ok := opts.(*llm.CompletionOptions); ok {
		completionOpts = o
	}

	return a.client.CompleteStreaming(ctx, llmMessages, completionOpts, callback)
}

// WebSearchClientAdapter adapta websearch.Orchestrator para handlers.WebSearchClient
type WebSearchClientAdapter struct {
	orchestrator *websearch.Orchestrator
}

func NewWebSearchClientAdapter(orchestrator *websearch.Orchestrator) *WebSearchClientAdapter {
	return &WebSearchClientAdapter{orchestrator: orchestrator}
}

func (a *WebSearchClientAdapter) Search(ctx context.Context, query string) (interface{}, error) {
	results, err := a.orchestrator.Search(ctx, query, []string{})
	if err != nil {
		return nil, err
	}

	// Convert []SearchResult to []interface{} with map[string]interface{} items
	// This makes it easier for the handler to format
	converted := make([]interface{}, len(results))
	for i, result := range results {
		converted[i] = map[string]interface{}{
			"title":   result.Title,
			"url":     result.URL,
			"snippet": result.Snippet,
			"source":  result.Source,
		}
	}

	return converted, nil
}

// IntentDetectorAdapter adapta intent.Detector para handlers.IntentDetector
type IntentDetectorAdapter struct {
	detector *intent.Detector
}

func NewIntentDetectorAdapter(detector *intent.Detector) *IntentDetectorAdapter {
	return &IntentDetectorAdapter{detector: detector}
}

func (a *IntentDetectorAdapter) DetectWithHistory(ctx context.Context, message string, history []Message) (*intent.DetectionResult, error) {
	llmMessages := make([]llm.Message, len(history))
	for i, msg := range history {
		llmMessages[i] = llm.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Detector original precisa de workDir e recentFiles, vamos passar valores vazios
	return a.detector.DetectWithHistory(ctx, message, "", []string{}, llmMessages)
}

// OperationModeAdapter adapta modes.OperationMode para handlers.OperationMode
type OperationModeAdapter struct {
	mode modes.OperationMode
}

func NewOperationModeAdapter(mode modes.OperationMode) *OperationModeAdapter {
	return &OperationModeAdapter{mode: mode}
}

func (a *OperationModeAdapter) String() string {
	return string(a.mode)
}

func (a *OperationModeAdapter) AllowsWrites() bool {
	return a.mode.AllowsWrites()
}

func (a *OperationModeAdapter) RequiresConfirmation() bool {
	return a.mode.RequiresConfirmation()
}
