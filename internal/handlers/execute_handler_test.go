package handlers

import (
	"context"
	"testing"

	"github.com/johnpitter/ollama-code/internal/intent"
)

func TestExecuteHandler_Success(t *testing.T) {
	handler := NewExecuteHandler()
	deps := NewMockDependencies()

	toolCalled := false
	deps.ToolRegistry = &MockToolRegistry{
		ExecuteFunc: func(ctx context.Context, toolName string, params map[string]interface{}) (ToolResult, error) {
			toolCalled = true
			AssertEqual(t, "command_executor", toolName, "tool name")
			AssertEqual(t, "ls -la", params["command"], "command param")
			return MockToolResultSuccess("file1.txt\nfile2.txt"), nil
		},
	}

	result := NewMockDetectionResult(intent.IntentExecuteCommand, map[string]interface{}{
		"command": "ls -la",
	})

	ctx := context.Background()
	response, err := handler.Handle(ctx, deps, result)

	AssertNoError(t, err)
	AssertNotEmpty(t, response, "response")
	AssertToolCalled(t, "command_executor", &toolCalled)
}

func TestExecuteHandler_MissingCommand(t *testing.T) {
	handler := NewExecuteHandler()
	deps := NewMockDependencies()

	result := NewMockDetectionResult(intent.IntentExecuteCommand, map[string]interface{}{})

	ctx := context.Background()
	_, err := handler.Handle(ctx, deps, result)

	AssertError(t, err, "missing command")
}

func TestExecuteHandler_DangerousCommand_Blocked(t *testing.T) {
	handler := NewExecuteHandler()
	deps := NewMockDependencies()

	// Mode precisa retornar true para RequiresConfirmation
	deps.Mode = &MockOperationMode{
		RequiresConfirmationFunc: func() bool {
			return true
		},
	}

	// Mock retornando false (usuário rejeitou)
	deps.ConfirmManager = &MockConfirmationManager{
		ConfirmFunc: func(message string) (bool, error) {
			// Verificar que confirmação foi solicitada
			if !contains(message, "perigoso") && !contains(message, "danger") {
				t.Error("Expected confirmation message to mention danger")
			}
			return false, nil // Usuário rejeita
		},
	}

	dangerousCommands := []string{
		"rm -rf /",
		"mkfs.ext4 /dev/sda",
		"dd if=/dev/zero of=/dev/sda",
		":(){ :|:& };:",
		"> /dev/sda",
		"chmod -R 777 /",
	}

	for _, cmd := range dangerousCommands {
		result := NewMockDetectionResult(intent.IntentExecuteCommand, map[string]interface{}{
			"command": cmd,
		})

		ctx := context.Background()
		response, err := handler.Handle(ctx, deps, result)

		// Handler retorna mensagem de cancelamento, não erro
		AssertNoError(t, err)
		AssertContains(t, response, "cancelado", "cancellation message")
	}
}

func TestExecuteHandler_DangerousCommand_Confirmed(t *testing.T) {
	handler := NewExecuteHandler()
	deps := NewMockDependencies()

	// Mode precisa retornar true para RequiresConfirmation
	deps.Mode = &MockOperationMode{
		RequiresConfirmationFunc: func() bool {
			return true
		},
	}

	confirmed := false
	deps.ConfirmManager = &MockConfirmationManager{
		ConfirmFunc: func(message string) (bool, error) {
			confirmed = true
			return true, nil // Usuário confirma
		},
	}

	toolCalled := false
	deps.ToolRegistry = &MockToolRegistry{
		ExecuteFunc: func(ctx context.Context, toolName string, params map[string]interface{}) (ToolResult, error) {
			toolCalled = true
			return MockToolResultSuccess("executed"), nil
		},
	}

	result := NewMockDetectionResult(intent.IntentExecuteCommand, map[string]interface{}{
		"command": "rm -rf /tmp/test",
	})

	ctx := context.Background()
	response, err := handler.Handle(ctx, deps, result)

	AssertNoError(t, err)
	AssertNotEmpty(t, response, "response")

	if !confirmed {
		t.Error("Expected confirmation to be requested")
	}
	AssertToolCalled(t, "command_executor", &toolCalled)
}

func TestExecuteHandler_SafeCommand_NoConfirmation(t *testing.T) {
	handler := NewExecuteHandler()
	deps := NewMockDependencies()

	confirmCalled := false
	deps.ConfirmManager = &MockConfirmationManager{
		ConfirmFunc: func(message string) (bool, error) {
			confirmCalled = true
			return true, nil
		},
	}

	toolCalled := false
	deps.ToolRegistry = &MockToolRegistry{
		ExecuteFunc: func(ctx context.Context, toolName string, params map[string]interface{}) (ToolResult, error) {
			toolCalled = true
			return MockToolResultSuccess("executed"), nil
		},
	}

	safeCommands := []string{
		"ls -la",
		"cat file.txt",
		"echo hello",
		"pwd",
		"git status",
	}

	for _, cmd := range safeCommands {
		confirmCalled = false
		toolCalled = false

		result := NewMockDetectionResult(intent.IntentExecuteCommand, map[string]interface{}{
			"command": cmd,
		})

		ctx := context.Background()
		_, err := handler.Handle(ctx, deps, result)

		AssertNoError(t, err)

		if confirmCalled {
			t.Errorf("Safe command should not require confirmation: %s", cmd)
		}
		AssertToolCalled(t, "command_executor", &toolCalled)
	}
}

func TestExecuteHandler_ToolError(t *testing.T) {
	handler := NewExecuteHandler()
	deps := NewMockDependencies()

	deps.ToolRegistry = CreateMockToolRegistry(
		"command_executor",
		MockToolResultError("command failed"),
		nil,
	)

	result := NewMockDetectionResult(intent.IntentExecuteCommand, map[string]interface{}{
		"command": "false",
	})

	ctx := context.Background()
	response, err := handler.Handle(ctx, deps, result)

	// Handler pode retornar erro ou mensagem de erro
	if err == nil && response == "" {
		t.Error("Expected error or error message")
	}
}
