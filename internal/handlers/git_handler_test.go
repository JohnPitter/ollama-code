package handlers

import (
	"context"
	"testing"

	"github.com/johnpitter/ollama-code/internal/intent"
)

func TestGitHandler_Success(t *testing.T) {
	handler := NewGitHandler()
	deps := NewMockDependencies()

	toolCalled := false
	deps.ToolRegistry = &MockToolRegistry{
		ExecuteFunc: func(ctx context.Context, toolName string, params map[string]interface{}) (ToolResult, error) {
			toolCalled = true
			AssertEqual(t, "git_operations", toolName, "tool name")
			AssertEqual(t, "status", params["operation"], "operation param")
			return MockToolResultSuccess("On branch main"), nil
		},
	}

	result := NewMockDetectionResult(intent.IntentGitOperation, map[string]interface{}{
		"operation": "status",
	})

	ctx := context.Background()
	response, err := handler.Handle(ctx, deps, result)

	AssertNoError(t, err)
	AssertNotEmpty(t, response, "response")
	AssertContains(t, response, "branch main", "git status output")
	AssertToolCalled(t, "git_operations", &toolCalled)
}

func TestGitHandler_MissingOperation(t *testing.T) {
	handler := NewGitHandler()
	deps := NewMockDependencies()

	toolCalled := false
	deps.ToolRegistry = &MockToolRegistry{
		ExecuteFunc: func(ctx context.Context, toolName string, params map[string]interface{}) (ToolResult, error) {
			toolCalled = true
			// Handler should default to "status"
			AssertEqual(t, "status", params["operation"], "default operation")
			return MockToolResultSuccess("On branch main"), nil
		},
	}

	result := NewMockDetectionResult(intent.IntentGitOperation, map[string]interface{}{})

	ctx := context.Background()
	response, err := handler.Handle(ctx, deps, result)

	AssertNoError(t, err)
	AssertNotEmpty(t, response, "response")
	AssertToolCalled(t, "git_operations", &toolCalled)
}

func TestGitHandler_CommitOperation(t *testing.T) {
	handler := NewGitHandler()
	deps := NewMockDependencies()

	toolCalled := false
	deps.ToolRegistry = &MockToolRegistry{
		ExecuteFunc: func(ctx context.Context, toolName string, params map[string]interface{}) (ToolResult, error) {
			toolCalled = true
			AssertEqual(t, "commit", params["operation"], "operation")
			AssertEqual(t, "feat: add feature", params["message"], "commit message")
			return MockToolResultSuccess("Commit created"), nil
		},
	}

	result := NewMockDetectionResult(intent.IntentGitOperation, map[string]interface{}{
		"operation": "commit",
		"message":   "feat: add feature",
	})

	ctx := context.Background()
	response, err := handler.Handle(ctx, deps, result)

	AssertNoError(t, err)
	AssertNotEmpty(t, response, "response")
	AssertToolCalled(t, "git_operations", &toolCalled)
}
