package handlers

import (
	"context"
	"testing"

	"github.com/johnpitter/ollama-code/internal/intent"
)

func TestFileReadHandler_Success(t *testing.T) {
	handler := NewFileReadHandler()
	deps := NewMockDependencies()

	// Mock tool execution
	toolCalled := false
	deps.ToolRegistry = &MockToolRegistry{
		ExecuteFunc: func(ctx context.Context, toolName string, params map[string]interface{}) (ToolResult, error) {
			toolCalled = true
			AssertEqual(t, "file_reader", toolName, "tool name")
			AssertEqual(t, "test.txt", params["file_path"], "file_path param")
			return MockToolResultSuccess("File content here"), nil
		},
	}

	result := NewMockDetectionResult(intent.IntentReadFile, map[string]interface{}{
		"file_path": "test.txt",
	})

	ctx := context.Background()
	response, err := handler.Handle(ctx, deps, result)

	AssertNoError(t, err)
	AssertNotEmpty(t, response, "response")
	AssertContains(t, response, "File content", "response content")
	AssertToolCalled(t, "file_reader", &toolCalled)
}

func TestFileReadHandler_MissingFilePath(t *testing.T) {
	handler := NewFileReadHandler()
	deps := NewMockDependencies()

	result := NewMockDetectionResult(intent.IntentReadFile, map[string]interface{}{})

	ctx := context.Background()
	_, err := handler.Handle(ctx, deps, result)

	AssertError(t, err, "missing file_path")
	if !ErrorContains(err, "file_path") {
		t.Errorf("Expected error to mention file_path, got: %v", err)
	}
}

func TestFileReadHandler_ToolError(t *testing.T) {
	handler := NewFileReadHandler()
	deps := NewMockDependencies()

	deps.ToolRegistry = CreateMockToolRegistry(
		"file_reader",
		MockToolResultError("file not found"),
		nil,
	)

	result := NewMockDetectionResult(intent.IntentReadFile, map[string]interface{}{
		"file_path": "nonexistent.txt",
	})

	ctx := context.Background()
	response, err := handler.Handle(ctx, deps, result)

	// Handler pode retornar erro ou mensagem de erro
	if err == nil && response == "" {
		t.Error("Expected error or error message")
	}
}

func TestFileReadHandler_UpdatesRecentFiles(t *testing.T) {
	handler := NewFileReadHandler()
	deps := NewMockDependencies()

	deps.ToolRegistry = CreateMockToolRegistry(
		"file_reader",
		MockToolResultSuccess("content"),
		nil,
	)

	result := NewMockDetectionResult(intent.IntentReadFile, map[string]interface{}{
		"file_path": "test.txt",
	})

	ctx := context.Background()
	_, err := handler.Handle(ctx, deps, result)

	AssertNoError(t, err)

	// Verificar se arquivo foi adicionado a recentFiles
	found := false
	for _, f := range deps.RecentFiles {
		if f == "test.txt" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected test.txt to be added to recentFiles")
	}
}
