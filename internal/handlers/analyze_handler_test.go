package handlers

import (
	"context"
	"testing"

	"github.com/johnpitter/ollama-code/internal/intent"
)

func TestAnalyzeHandler_Success(t *testing.T) {
	handler := NewAnalyzeHandler()
	deps := NewMockDependencies()

	toolCalled := false
	deps.ToolRegistry = &MockToolRegistry{
		ExecuteFunc: func(ctx context.Context, toolName string, params map[string]interface{}) (ToolResult, error) {
			toolCalled = true
			AssertEqual(t, "project_analyzer", toolName, "tool name")
			return MockToolResultSuccess("Project structure analyzed"), nil
		},
	}

	result := NewMockDetectionResult(intent.IntentAnalyzeProject, map[string]interface{}{})

	ctx := context.Background()
	response, err := handler.Handle(ctx, deps, result)

	AssertNoError(t, err)
	AssertNotEmpty(t, response, "response")
	AssertContains(t, response, "analyzed", "analysis output")
	AssertToolCalled(t, "project_analyzer", &toolCalled)
}

func TestAnalyzeHandler_WithPath(t *testing.T) {
	handler := NewAnalyzeHandler()
	deps := NewMockDependencies()

	toolCalled := false
	deps.ToolRegistry = &MockToolRegistry{
		ExecuteFunc: func(ctx context.Context, toolName string, params map[string]interface{}) (ToolResult, error) {
			toolCalled = true
			if path, ok := params["path"].(string); ok {
				AssertEqual(t, "internal/", path, "path param")
			}
			return MockToolResultSuccess("Analyzed internal/ directory"), nil
		},
	}

	result := NewMockDetectionResult(intent.IntentAnalyzeProject, map[string]interface{}{
		"path": "internal/",
	})

	ctx := context.Background()
	response, err := handler.Handle(ctx, deps, result)

	AssertNoError(t, err)
	AssertNotEmpty(t, response, "response")
	AssertToolCalled(t, "project_analyzer", &toolCalled)
}

func TestAnalyzeHandler_ToolError(t *testing.T) {
	handler := NewAnalyzeHandler()
	deps := NewMockDependencies()

	deps.ToolRegistry = CreateMockToolRegistry(
		"project_analyzer",
		MockToolResultError("analysis failed"),
		nil,
	)

	result := NewMockDetectionResult(intent.IntentAnalyzeProject, map[string]interface{}{})

	ctx := context.Background()
	response, err := handler.Handle(ctx, deps, result)

	// Handler pode retornar erro ou mensagem de erro
	if err == nil && response == "" {
		t.Error("Expected error or error message")
	}
}
