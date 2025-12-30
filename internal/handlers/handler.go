package handlers

import (
	"context"

	"github.com/johnpitter/ollama-code/internal/intent"
)

// Handler processa um intent específico
type Handler interface {
	// Handle processa o intent e retorna resultado
	Handle(ctx context.Context, deps *Dependencies, result *intent.DetectionResult) (string, error)
}

// Dependencies agrupa as dependências que handlers podem precisar
// Isso evita que handlers precisem do Agent completo (reduz acoplamento)
type Dependencies struct {
	// Registries
	ToolRegistry    ToolRegistry
	CommandRegistry CommandRegistry
	SkillRegistry   SkillRegistry

	// Managers
	ConfirmManager  ConfirmationManager
	SessionManager  SessionManager
	CacheManager    CacheManager
	TodoManager     TodoManager
	DiffManager     DiffManager
	PreviewManager  PreviewManager

	// Clients
	LLMClient      LLMClient
	WebSearch      WebSearchClient
	IntentDetector IntentDetector

	// State
	Mode        OperationMode
	WorkDir     string
	History     []Message
	RecentFiles []string
}

// Interfaces para desacoplamento (não depender de implementações concretas)

type ToolRegistry interface {
	Execute(ctx context.Context, toolName string, params map[string]interface{}) (ToolResult, error)
	Get(name string) (interface{}, error)
}

type CommandRegistry interface {
	Execute(ctx context.Context, cmdName string, args []string) (string, error)
	IsCommand(text string) bool
}

type SkillRegistry interface {
	FindSkill(ctx context.Context, message string) (interface{}, error)
}

type ConfirmationManager interface {
	Confirm(message string) (bool, error)
	ConfirmWithPreview(message, preview string) (bool, error)
	AskQuestion(question interface{}) (interface{}, error)
	AskQuestions(questionSet interface{}) (map[string]interface{}, error)
}

type SessionManager interface {
	SaveMessage(role, content string) error
}

type CacheManager interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
}

type TodoManager interface {
	Add(content, activeForm string) (string, error)
	Update(id string, status interface{}) error
	Complete(id string) error
	SetInProgress(id string) error
	List() interface{}
	ListByStatus(status interface{}) interface{}
	Summary() map[interface{}]int
	Clear() error
	Delete(id string) error
	Count() int
}

type DiffManager interface {
	ComputeDiff(filePath, oldContent, newContent string) interface{}
	ApplyEdit(filePath, content string, editRange interface{}) (string, interface{}, error)
	Rollback(filePath string) (string, error)
	GetHistory(filePath string) interface{}
	ClearHistory()
}

type PreviewManager interface {
	Preview(diff interface{}) string
	PreviewRange(filePath, oldContent string, editRange interface{}) string
	CompactPreview(diff interface{}) string
}

type LLMClient interface {
	Complete(ctx context.Context, prompt string) (string, error)
	CompleteWithHistory(ctx context.Context, messages []Message) (string, error)
	CompleteStreaming(ctx context.Context, messages []Message, opts interface{}, callback func(string)) (string, error)
}

type WebSearchClient interface {
	Search(ctx context.Context, query string) (interface{}, error)
}

type IntentDetector interface {
	DetectWithHistory(ctx context.Context, message string, history []Message) (*intent.DetectionResult, error)
}

type OperationMode interface {
	String() string
	AllowsWrites() bool
	RequiresConfirmation() bool
}

type Message struct {
	Role    string
	Content string
}

type ToolResult struct {
	Success bool
	Message string
	Error   string
	Data    map[string]interface{}
}

// BaseHandler fornece funcionalidade comum para handlers
type BaseHandler struct {
	name string
}

// NewBaseHandler cria um novo base handler
func NewBaseHandler(name string) BaseHandler {
	return BaseHandler{name: name}
}

// Name retorna o nome do handler
func (h *BaseHandler) Name() string {
	return h.name
}
