package handlers

import (
	"context"
	"fmt"
	"testing"

	"github.com/johnpitter/ollama-code/internal/intent"
)

func TestFileWriteHandler_WithContent(t *testing.T) {
	handler := NewFileWriteHandler()
	deps := NewMockDependencies()

	toolCalled := false
	deps.ToolRegistry = &MockToolRegistry{
		ExecuteFunc: func(ctx context.Context, toolName string, params map[string]interface{}) (ToolResult, error) {
			toolCalled = true
			AssertEqual(t, "file_writer", toolName, "tool name")
			AssertEqual(t, "test.txt", params["file_path"], "file_path param")
			AssertEqual(t, "hello world", params["content"], "content param")
			return MockToolResultSuccess("File written"), nil
		},
	}

	result := NewMockDetectionResult(intent.IntentWriteFile, map[string]interface{}{
		"file_path": "test.txt",
		"content":   "hello world",
	})

	ctx := context.Background()
	response, err := handler.Handle(ctx, deps, result)

	AssertNoError(t, err)
	AssertNotEmpty(t, response, "response")
	AssertToolCalled(t, "file_writer", &toolCalled)
}

func TestFileWriteHandler_GenerateWithLLM(t *testing.T) {
	handler := NewFileWriteHandler()
	deps := NewMockDependencies()

	llmCalled := false
	deps.LLMClient = &MockLLMClient{
		CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
			llmCalled = true
			// Verificar que prompt contém instruções
			if !contains(prompt, "file") && !contains(prompt, "generate") {
				t.Error("Expected prompt to contain generation instructions")
			}
			// Retornar JSON com file_path e content
			return `{
				"file_path": "generated.txt",
				"content": "// Generated content\npackage main"
			}`, nil
		},
	}

	toolCalled := false
	deps.ToolRegistry = &MockToolRegistry{
		ExecuteFunc: func(ctx context.Context, toolName string, params map[string]interface{}) (ToolResult, error) {
			toolCalled = true
			AssertEqual(t, "file_writer", toolName, "tool name")
			return MockToolResultSuccess("File written"), nil
		},
	}

	result := NewMockDetectionResult(intent.IntentWriteFile, map[string]interface{}{
		// Sem content - deve gerar com LLM
		"file_path": "generated.txt",
	})
	result.UserMessage = "create a main.go file"

	ctx := context.Background()
	response, err := handler.Handle(ctx, deps, result)

	AssertNoError(t, err)
	AssertNotEmpty(t, response, "response")

	if !llmCalled {
		t.Error("Expected LLM to be called for content generation")
	}
	AssertToolCalled(t, "file_writer", &toolCalled)
}

func TestFileWriteHandler_MissingFilePath(t *testing.T) {
	handler := NewFileWriteHandler()
	deps := NewMockDependencies()

	result := NewMockDetectionResult(intent.IntentWriteFile, map[string]interface{}{
		"content": "hello",
	})

	ctx := context.Background()
	_, err := handler.Handle(ctx, deps, result)

	AssertError(t, err, "missing file_path")
}

func TestFileWriteHandler_InvalidFilename(t *testing.T) {
	handler := NewFileWriteHandler()
	deps := NewMockDependencies()

	invalidNames := []string{
		"test<file>.txt",
		"test>file.txt",
		"test:file.txt",
		"test|file.txt",
		"test?file.txt",
		"test*file.txt",
	}

	for _, name := range invalidNames {
		result := NewMockDetectionResult(intent.IntentWriteFile, map[string]interface{}{
			"file_path": name,
			"content":   "test",
		})

		ctx := context.Background()
		_, err := handler.Handle(ctx, deps, result)

		if err == nil {
			t.Errorf("Expected error for invalid filename: %s", name)
		}
	}
}

func TestFileWriteHandler_WithConfirmation(t *testing.T) {
	handler := NewFileWriteHandler()
	deps := NewMockDependencies()

	// Mode que requer confirmação
	deps.Mode = &MockOperationMode{
		RequiresConfirmationFunc: func() bool {
			return true
		},
	}

	confirmed := false
	deps.ConfirmManager = &MockConfirmationManager{
		ConfirmWithPreviewFunc: func(message, preview string) (bool, error) {
			confirmed = true
			// Verificar que preview contém conteúdo
			if !contains(preview, "hello") {
				t.Error("Expected preview to contain file content")
			}
			return true, nil
		},
	}

	toolCalled := false
	deps.ToolRegistry = &MockToolRegistry{
		ExecuteFunc: func(ctx context.Context, toolName string, params map[string]interface{}) (ToolResult, error) {
			toolCalled = true
			return MockToolResultSuccess("File written"), nil
		},
	}

	result := NewMockDetectionResult(intent.IntentWriteFile, map[string]interface{}{
		"file_path": "test.txt",
		"content":   "hello world",
	})

	ctx := context.Background()
	response, err := handler.Handle(ctx, deps, result)

	AssertNoError(t, err)
	AssertNotEmpty(t, response, "response")

	if !confirmed {
		t.Error("Expected confirmation to be requested")
	}
	AssertToolCalled(t, "file_writer", &toolCalled)
}

func TestFileWriteHandler_ConfirmationRejected(t *testing.T) {
	handler := NewFileWriteHandler()
	deps := NewMockDependencies()

	deps.Mode = &MockOperationMode{
		RequiresConfirmationFunc: func() bool {
			return true
		},
	}

	deps.ConfirmManager = &MockConfirmationManager{
		ConfirmWithPreviewFunc: func(message, preview string) (bool, error) {
			return false, nil // Usuário rejeita
		},
	}

	result := NewMockDetectionResult(intent.IntentWriteFile, map[string]interface{}{
		"file_path": "test.txt",
		"content":   "hello world",
	})

	ctx := context.Background()
	response, err := handler.Handle(ctx, deps, result)

	// Handler retorna mensagem de cancelamento, não erro
	AssertNoError(t, err)
	AssertContains(t, response, "cancelada", "cancellation message")
}

func TestFileWriteHandler_CleanMarkdown(t *testing.T) {
	handler := NewFileWriteHandler()
	deps := NewMockDependencies()

	var capturedContent string
	deps.ToolRegistry = &MockToolRegistry{
		ExecuteFunc: func(ctx context.Context, toolName string, params map[string]interface{}) (ToolResult, error) {
			capturedContent = params["content"].(string)
			return MockToolResultSuccess("File written"), nil
		},
	}

	// Content com markdown code block
	result := NewMockDetectionResult(intent.IntentWriteFile, map[string]interface{}{
		"file_path": "test.go",
		"content":   "```go\npackage main\n```",
	})

	ctx := context.Background()
	_, err := handler.Handle(ctx, deps, result)

	AssertNoError(t, err)

	// Verificar que markdown foi removido
	if contains(capturedContent, "```") {
		t.Error("Expected markdown code blocks to be removed")
	}
	if !contains(capturedContent, "package main") {
		t.Error("Expected actual content to be preserved")
	}
}

func TestFileWriteHandler_LLMGenerationError(t *testing.T) {
	handler := NewFileWriteHandler()
	deps := NewMockDependencies()

	deps.LLMClient = &MockLLMClient{
		CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
			return "", fmt.Errorf("LLM service unavailable")
		},
	}

	result := NewMockDetectionResult(intent.IntentWriteFile, map[string]interface{}{
		"file_path": "test.txt",
		// Sem content - tentará gerar com LLM
	})

	ctx := context.Background()
	_, err := handler.Handle(ctx, deps, result)

	AssertError(t, err, "LLM generation failure")
}
