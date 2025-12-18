package tools

import (
	"context"
	"runtime"
	"testing"
	"time"
)

func TestCommandExecutor_SimpleCommand(t *testing.T) {
	executor := NewCommandExecutor(".", 5*time.Second)

	var command string
	if runtime.GOOS == "windows" {
		command = "echo hello"
	} else {
		command = "echo hello"
	}

	result, err := executor.Execute(context.Background(), map[string]interface{}{
		"command": command,
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got error: %s", result.Error)
	}

	stdout, ok := result.Data["stdout"].(string)
	if !ok {
		t.Fatal("stdout is not a string")
	}

	if stdout == "" {
		t.Error("stdout should not be empty")
	}

	exitCode, ok := result.Data["exit_code"].(int)
	if !ok {
		t.Fatal("exit_code is not an int")
	}

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestCommandExecutor_FailingCommand(t *testing.T) {
	executor := NewCommandExecutor(".", 5*time.Second)

	var command string
	if runtime.GOOS == "windows" {
		command = "cmd /c exit 1"
	} else {
		command = "false"
	}

	result, err := executor.Execute(context.Background(), map[string]interface{}{
		"command": command,
	})

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	// Command should execute but return non-zero exit code
	if !result.Success {
		t.Fatal("Execute should succeed even if command fails")
	}

	exitCode, ok := result.Data["exit_code"].(int)
	if !ok {
		t.Fatal("exit_code is not an int")
	}

	if exitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
}

func TestCommandExecutor_IsDangerous(t *testing.T) {
	executor := NewCommandExecutor(".", 5*time.Second)

	tests := []struct {
		command   string
		dangerous bool
	}{
		{"rm -rf /", true},
		{"del /F /Q C:\\", true},
		{"format C:", true},
		{"dd if=/dev/zero of=/dev/sda", true},
		{"echo hello", false},
		{"ls -la", false},
		{"cat file.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := executor.IsDangerous(tt.command)
			if result != tt.dangerous {
				t.Errorf("IsDangerous(%q) = %v, want %v", tt.command, result, tt.dangerous)
			}
		})
	}
}

func TestCommandExecutor_MissingCommand(t *testing.T) {
	executor := NewCommandExecutor(".", 5*time.Second)

	result, err := executor.Execute(context.Background(), map[string]interface{}{})

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if result.Success {
		t.Error("Expected failure for missing command parameter")
	}
}

func TestCommandExecutor_Name(t *testing.T) {
	executor := NewCommandExecutor(".", 5*time.Second)
	if executor.Name() != "command_executor" {
		t.Errorf("Expected name 'command_executor', got '%s'", executor.Name())
	}
}

func TestCommandExecutor_RequiresConfirmation(t *testing.T) {
	executor := NewCommandExecutor(".", 5*time.Second)
	if !executor.RequiresConfirmation() {
		t.Error("CommandExecutor should require confirmation")
	}
}

func TestCommandExecutor_Description(t *testing.T) {
	executor := NewCommandExecutor(".", 5*time.Second)
	desc := executor.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
}
