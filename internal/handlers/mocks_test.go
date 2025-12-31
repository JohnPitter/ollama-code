package handlers

import (
	"context"
	"fmt"

	"github.com/johnpitter/ollama-code/internal/intent"
)

// MockToolRegistry mock para ToolRegistry
type MockToolRegistry struct {
	ExecuteFunc func(ctx context.Context, toolName string, params map[string]interface{}) (ToolResult, error)
	GetFunc     func(name string) (interface{}, error)
}

func (m *MockToolRegistry) Execute(ctx context.Context, toolName string, params map[string]interface{}) (ToolResult, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, toolName, params)
	}
	return ToolResult{Success: true, Message: "mock result"}, nil
}

func (m *MockToolRegistry) Get(name string) (interface{}, error) {
	if m.GetFunc != nil {
		return m.GetFunc(name)
	}
	return nil, nil
}

// MockCommandRegistry mock para CommandRegistry
type MockCommandRegistry struct {
	ExecuteFunc   func(ctx context.Context, cmdName string, args []string) (string, error)
	IsCommandFunc func(text string) bool
}

func (m *MockCommandRegistry) Execute(ctx context.Context, cmdName string, args []string) (string, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, cmdName, args)
	}
	return "mock command result", nil
}

func (m *MockCommandRegistry) IsCommand(text string) bool {
	if m.IsCommandFunc != nil {
		return m.IsCommandFunc(text)
	}
	return false
}

// MockSkillRegistry mock para SkillRegistry
type MockSkillRegistry struct {
	FindSkillFunc func(ctx context.Context, message string) (interface{}, error)
}

func (m *MockSkillRegistry) FindSkill(ctx context.Context, message string) (interface{}, error) {
	if m.FindSkillFunc != nil {
		return m.FindSkillFunc(ctx, message)
	}
	return nil, nil
}

// MockConfirmationManager mock para ConfirmationManager
type MockConfirmationManager struct {
	ConfirmFunc            func(message string) (bool, error)
	ConfirmWithPreviewFunc func(message, preview string) (bool, error)
	AskQuestionFunc        func(question interface{}) (interface{}, error)
	AskQuestionsFunc       func(questionSet interface{}) (map[string]interface{}, error)
}

func (m *MockConfirmationManager) Confirm(message string) (bool, error) {
	if m.ConfirmFunc != nil {
		return m.ConfirmFunc(message)
	}
	return true, nil
}

func (m *MockConfirmationManager) ConfirmWithPreview(message, preview string) (bool, error) {
	if m.ConfirmWithPreviewFunc != nil {
		return m.ConfirmWithPreviewFunc(message, preview)
	}
	return true, nil
}

func (m *MockConfirmationManager) AskQuestion(question interface{}) (interface{}, error) {
	if m.AskQuestionFunc != nil {
		return m.AskQuestionFunc(question)
	}
	return nil, nil
}

func (m *MockConfirmationManager) AskQuestions(questionSet interface{}) (map[string]interface{}, error) {
	if m.AskQuestionsFunc != nil {
		return m.AskQuestionsFunc(questionSet)
	}
	return nil, nil
}

// MockSessionManager mock para SessionManager
type MockSessionManager struct {
	SaveMessageFunc func(role, content string) error
}

func (m *MockSessionManager) SaveMessage(role, content string) error {
	if m.SaveMessageFunc != nil {
		return m.SaveMessageFunc(role, content)
	}
	return nil
}

// MockCacheManager mock para CacheManager
type MockCacheManager struct {
	GetFunc func(key string) (interface{}, bool)
	SetFunc func(key string, value interface{})
}

func (m *MockCacheManager) Get(key string) (interface{}, bool) {
	if m.GetFunc != nil {
		return m.GetFunc(key)
	}
	return nil, false
}

func (m *MockCacheManager) Set(key string, value interface{}) {
	if m.SetFunc != nil {
		m.SetFunc(key, value)
	}
}

// MockLLMClient mock para LLMClient
type MockLLMClient struct {
	CompleteFunc            func(ctx context.Context, prompt string) (string, error)
	CompleteWithHistoryFunc func(ctx context.Context, messages []Message) (string, error)
	CompleteStreamingFunc   func(ctx context.Context, messages []Message, opts interface{}, callback func(string)) (string, error)
}

func (m *MockLLMClient) Complete(ctx context.Context, prompt string) (string, error) {
	if m.CompleteFunc != nil {
		return m.CompleteFunc(ctx, prompt)
	}
	return "mock llm response", nil
}

func (m *MockLLMClient) CompleteWithHistory(ctx context.Context, messages []Message) (string, error) {
	if m.CompleteWithHistoryFunc != nil {
		return m.CompleteWithHistoryFunc(ctx, messages)
	}
	return "mock llm response with history", nil
}

func (m *MockLLMClient) CompleteStreaming(ctx context.Context, messages []Message, opts interface{}, callback func(string)) (string, error) {
	if m.CompleteStreamingFunc != nil {
		return m.CompleteStreamingFunc(ctx, messages, opts, callback)
	}
	return "mock streaming response", nil
}

// MockWebSearchClient mock para WebSearchClient
type MockWebSearchClient struct {
	SearchFunc func(ctx context.Context, query string) (interface{}, error)
}

func (m *MockWebSearchClient) Search(ctx context.Context, query string) (interface{}, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, query)
	}
	return map[string]interface{}{
		"results": []interface{}{
			map[string]interface{}{
				"title":   "Mock Result",
				"snippet": "Mock snippet",
				"url":     "https://example.com",
			},
		},
	}, nil
}

// MockIntentDetector mock para IntentDetector
type MockIntentDetector struct {
	DetectWithHistoryFunc func(ctx context.Context, message string, history []Message) (*intent.DetectionResult, error)
}

func (m *MockIntentDetector) DetectWithHistory(ctx context.Context, message string, history []Message) (*intent.DetectionResult, error) {
	if m.DetectWithHistoryFunc != nil {
		return m.DetectWithHistoryFunc(ctx, message, history)
	}
	return &intent.DetectionResult{
		Intent:     intent.IntentQuestion,
		Confidence: 0.9,
		Parameters: map[string]interface{}{},
	}, nil
}

// MockOperationMode mock para OperationMode
type MockOperationMode struct {
	StringFunc               func() string
	RequiresConfirmationFunc func() bool
	AllowsWritesFunc         func() bool
}

func (m *MockOperationMode) String() string {
	if m.StringFunc != nil {
		return m.StringFunc()
	}
	return "interactive"
}

func (m *MockOperationMode) RequiresConfirmation() bool {
	if m.RequiresConfirmationFunc != nil {
		return m.RequiresConfirmationFunc()
	}
	return false
}

func (m *MockOperationMode) AllowsWrites() bool {
	if m.AllowsWritesFunc != nil {
		return m.AllowsWritesFunc()
	}
	return true // Default allows writes
}

// NewMockDependencies cria Dependencies com mocks
func NewMockDependencies() *Dependencies {
	return &Dependencies{
		ToolRegistry:    &MockToolRegistry{},
		CommandRegistry: &MockCommandRegistry{},
		SkillRegistry:   &MockSkillRegistry{},
		ConfirmManager:  &MockConfirmationManager{},
		SessionManager:  &MockSessionManager{},
		CacheManager:    &MockCacheManager{},
		LLMClient:       &MockLLMClient{},
		WebSearch:       &MockWebSearchClient{},
		IntentDetector:  &MockIntentDetector{},
		Mode:            &MockOperationMode{},
		WorkDir:         "/tmp/test",
		History:         []Message{},
		RecentFiles:     []string{},
	}
}

// Helper para criar DetectionResult
func NewMockDetectionResult(intentType intent.Intent, params map[string]interface{}) *intent.DetectionResult {
	if params == nil {
		params = make(map[string]interface{})
	}

	return &intent.DetectionResult{
		Intent:      intentType,
		Confidence:  0.9,
		Parameters:  params,
		UserMessage: "test message",
	}
}

// AssertToolCalled verifica se tool foi chamado
func AssertToolCalled(t TestingT, toolName string, called *bool) {
	if !*called {
		t.Errorf("Expected tool '%s' to be called", toolName)
	}
}

// TestingT interface mínima para testes
type TestingT interface {
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Helper()
}

// ErrorContains verifica se erro contém substring
func ErrorContains(err error, substr string) bool {
	if err == nil {
		return false
	}
	return contains(err.Error(), substr)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// AssertError verifica se há erro
func AssertError(t TestingT, err error, message string) {
	t.Helper()
	if err == nil {
		t.Fatalf("Expected error: %s", message)
	}
}

// AssertNoError verifica se não há erro
func AssertNoError(t TestingT, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

// AssertEqual verifica igualdade
func AssertEqual(t TestingT, expected, actual interface{}, message string) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s: expected %v, got %v", message, expected, actual)
	}
}

// AssertNotEmpty verifica se string não está vazia
func AssertNotEmpty(t TestingT, s string, message string) {
	t.Helper()
	if s == "" {
		t.Errorf("%s: expected non-empty string", message)
	}
}

// AssertContains verifica se string contém substring
func AssertContains(t TestingT, s, substr string, message string) {
	t.Helper()
	if !contains(s, substr) {
		t.Errorf("%s: expected '%s' to contain '%s'", message, s, substr)
	}
}

// MockToolResultSuccess cria ToolResult de sucesso
func MockToolResultSuccess(message string) ToolResult {
	return ToolResult{
		Success: true,
		Message: message,
		Data:    make(map[string]interface{}),
	}
}

// MockToolResultError cria ToolResult de erro
func MockToolResultError(errMsg string) ToolResult {
	return ToolResult{
		Success: false,
		Error:   errMsg,
		Data:    make(map[string]interface{}),
	}
}

// CreateMockToolRegistry cria registry com tool específico
func CreateMockToolRegistry(toolName string, result ToolResult, err error) *MockToolRegistry {
	return &MockToolRegistry{
		ExecuteFunc: func(ctx context.Context, name string, params map[string]interface{}) (ToolResult, error) {
			if name == toolName {
				return result, err
			}
			return MockToolResultError(fmt.Sprintf("tool not found: %s", name)), fmt.Errorf("tool not found")
		},
	}
}

// testError erro simples para testes
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
