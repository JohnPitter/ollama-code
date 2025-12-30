package handlers

import (
	"context"
	"testing"

	"github.com/johnpitter/ollama-code/internal/intent"
)

func TestQuestionHandler_Success(t *testing.T) {
	handler := NewQuestionHandler()
	deps := NewMockDependencies()

	llmCalled := false
	deps.LLMClient = &MockLLMClient{
		CompleteWithHistoryFunc: func(ctx context.Context, messages []Message) (string, error) {
			llmCalled = true
			// Verificar que mensagem do usuário está no histórico
			found := false
			for _, msg := range messages {
				if msg.Role == "user" && contains(msg.Content, "test question") {
					found = true
					break
				}
			}
			if !found {
				t.Error("Expected user message in history")
			}
			return "Answer to your question", nil
		},
	}

	result := NewMockDetectionResult(intent.IntentQuestion, map[string]interface{}{})
	result.UserMessage = "test question"

	ctx := context.Background()
	response, err := handler.Handle(ctx, deps, result)

	AssertNoError(t, err)
	AssertNotEmpty(t, response, "response")
	AssertContains(t, response, "Answer", "LLM response")

	if !llmCalled {
		t.Error("Expected LLM to be called")
	}
}

func TestQuestionHandler_WithHistory(t *testing.T) {
	handler := NewQuestionHandler()
	deps := NewMockDependencies()

	// Adicionar histórico
	deps.History = []Message{
		{Role: "user", Content: "previous question"},
		{Role: "assistant", Content: "previous answer"},
	}

	historyIncluded := false
	deps.LLMClient = &MockLLMClient{
		CompleteWithHistoryFunc: func(ctx context.Context, messages []Message) (string, error) {
			// Verificar que histórico foi incluído
			if len(messages) >= 3 {
				historyIncluded = true
			}
			return "Response with context", nil
		},
	}

	result := NewMockDetectionResult(intent.IntentQuestion, map[string]interface{}{})
	result.UserMessage = "follow-up question"

	ctx := context.Background()
	_, err := handler.Handle(ctx, deps, result)

	AssertNoError(t, err)

	if !historyIncluded {
		t.Error("Expected conversation history to be included")
	}
}

func TestQuestionHandler_EmptyMessage(t *testing.T) {
	handler := NewQuestionHandler()
	deps := NewMockDependencies()

	llmCalled := false
	deps.LLMClient = &MockLLMClient{
		CompleteWithHistoryFunc: func(ctx context.Context, messages []Message) (string, error) {
			llmCalled = true
			// Handler doesn't validate empty messages, just passes to LLM
			return "I don't understand the question", nil
		},
	}

	result := NewMockDetectionResult(intent.IntentQuestion, map[string]interface{}{})
	result.UserMessage = ""

	ctx := context.Background()
	response, err := handler.Handle(ctx, deps, result)

	AssertNoError(t, err)
	AssertNotEmpty(t, response, "response")

	if !llmCalled {
		t.Error("Expected LLM to be called even with empty message")
	}
}

func TestQuestionHandler_LLMError(t *testing.T) {
	handler := NewQuestionHandler()
	deps := NewMockDependencies()

	deps.LLMClient = &MockLLMClient{
		CompleteWithHistoryFunc: func(ctx context.Context, messages []Message) (string, error) {
			return "", &testError{msg: "LLM service error"}
		},
	}

	result := NewMockDetectionResult(intent.IntentQuestion, map[string]interface{}{})
	result.UserMessage = "test question"

	ctx := context.Background()
	_, err := handler.Handle(ctx, deps, result)

	AssertError(t, err, "LLM error")
}

func TestQuestionHandler_SystemPrompt(t *testing.T) {
	handler := NewQuestionHandler()
	deps := NewMockDependencies()

	systemPromptFound := false
	deps.LLMClient = &MockLLMClient{
		CompleteWithHistoryFunc: func(ctx context.Context, messages []Message) (string, error) {
			// Verificar que existe uma mensagem de sistema
			for _, msg := range messages {
				if msg.Role == "system" {
					systemPromptFound = true
					break
				}
			}
			return "Response", nil
		},
	}

	result := NewMockDetectionResult(intent.IntentQuestion, map[string]interface{}{})
	result.UserMessage = "test question"

	ctx := context.Background()
	_, err := handler.Handle(ctx, deps, result)

	AssertNoError(t, err)

	if !systemPromptFound {
		t.Error("Expected system prompt to be included")
	}
}
