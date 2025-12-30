package handlers

import (
	"context"
	"testing"

	"github.com/johnpitter/ollama-code/internal/intent"
)

func TestSearchHandler_Success(t *testing.T) {
	handler := NewSearchHandler()
	deps := NewMockDependencies()

	toolCalled := false
	deps.ToolRegistry = &MockToolRegistry{
		ExecuteFunc: func(ctx context.Context, toolName string, params map[string]interface{}) (ToolResult, error) {
			toolCalled = true
			AssertEqual(t, "code_searcher", toolName, "tool name")
			AssertEqual(t, "func main", params["pattern"], "pattern param")
			return MockToolResultSuccess("Found 3 matches"), nil
		},
	}

	result := NewMockDetectionResult(intent.IntentSearchCode, map[string]interface{}{
		"query": "func main",
	})

	ctx := context.Background()
	response, err := handler.Handle(ctx, deps, result)

	AssertNoError(t, err)
	AssertNotEmpty(t, response, "response")
	AssertToolCalled(t, "code_searcher", &toolCalled)
}

func TestSearchHandler_MissingPattern(t *testing.T) {
	handler := NewSearchHandler()
	deps := NewMockDependencies()

	result := NewMockDetectionResult(intent.IntentSearchCode, map[string]interface{}{})
	result.UserMessage = "" // Empty message to test error case

	ctx := context.Background()
	_, err := handler.Handle(ctx, deps, result)

	AssertError(t, err, "não foi possível determinar o que buscar")
}

func TestSearchHandler_WithPath(t *testing.T) {
	handler := NewSearchHandler()
	deps := NewMockDependencies()

	toolCalled := false
	deps.ToolRegistry = &MockToolRegistry{
		ExecuteFunc: func(ctx context.Context, toolName string, params map[string]interface{}) (ToolResult, error) {
			toolCalled = true
			AssertEqual(t, "TODO", params["pattern"], "pattern param")
			return MockToolResultSuccess("Found matches"), nil
		},
	}

	result := NewMockDetectionResult(intent.IntentSearchCode, map[string]interface{}{
		"query": "TODO",
		"path":  "internal/",
	})

	ctx := context.Background()
	_, err := handler.Handle(ctx, deps, result)

	AssertNoError(t, err)
	AssertToolCalled(t, "code_searcher", &toolCalled)
}
