package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTestRunner_Name(t *testing.T) {
	tr := NewTestRunner(".")
	if tr.Name() != "test_runner" {
		t.Errorf("Expected name 'test_runner', got '%s'", tr.Name())
	}
}

func TestTestRunner_Description(t *testing.T) {
	tr := NewTestRunner(".")
	desc := tr.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
	if !strings.Contains(desc, "testes") {
		t.Error("Description should mention 'testes'")
	}
}

func TestTestRunner_RequiresConfirmation(t *testing.T) {
	tr := NewTestRunner(".")
	if tr.RequiresConfirmation() {
		t.Error("TestRunner should not require confirmation")
	}
}

func TestTestRunner_Schema(t *testing.T) {
	tr := NewTestRunner(".")
	schema := tr.Schema()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema properties should be a map")
	}

	if _, exists := props["action"]; !exists {
		t.Error("Schema should have 'action' property")
	}

	if _, exists := props["test"]; !exists {
		t.Error("Schema should have 'test' property")
	}

	if _, exists := props["verbose"]; !exists {
		t.Error("Schema should have 'verbose' property")
	}
}

func TestTestRunner_Execute_InvalidAction(t *testing.T) {
	tr := NewTestRunner(".")
	ctx := context.Background()

	params := map[string]interface{}{
		"action": "invalid_action",
	}

	result, _ := tr.Execute(ctx, params)

	if result.Success {
		t.Error("Result should not be successful for invalid action")
	}

	if !strings.Contains(result.Error, "desconhecida") {
		t.Error("Error should mention unknown action")
	}
}

func TestTestRunner_DetectProjectType(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected string
	}{
		{
			name:     "Node.js project",
			files:    []string{"package.json"},
			expected: "nodejs",
		},
		{
			name:     "Go project",
			files:    []string{"go.mod"},
			expected: "go",
		},
		{
			name:     "Python project",
			files:    []string{"requirements.txt"},
			expected: "python",
		},
		{
			name:     "Unknown project",
			files:    []string{},
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "test-*")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			for _, file := range tt.files {
				path := filepath.Join(tmpDir, file)
				if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
					t.Fatal(err)
				}
			}

			tr := NewTestRunner(tmpDir)
			result := tr.detectProjectType()

			if result != tt.expected {
				t.Errorf("Expected project type '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestTestRunner_Execute_Watch(t *testing.T) {
	tr := NewTestRunner(".")
	ctx := context.Background()

	params := map[string]interface{}{
		"action": "watch",
	}

	result, _ := tr.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	// Should return watch mode information
	if !strings.Contains(result.Message, "Watch") {
		t.Error("Should return watch mode information")
	}
}

func TestTestRunner_Execute_Run_UnsupportedProject(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-unknown-*")
	defer os.RemoveAll(tmpDir)

	tr := NewTestRunner(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"action": "run",
	}

	result, _ := tr.Execute(ctx, params)

	if result.Success {
		t.Error("Result should not be successful for unsupported project")
	}

	if !strings.Contains(result.Error, "não suportado") {
		t.Error("Error should mention unsupported project type")
	}
}

func TestTestRunner_Execute_Coverage_UnsupportedProject(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-cov-*")
	defer os.RemoveAll(tmpDir)

	tr := NewTestRunner(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"action": "coverage",
	}

	result, _ := tr.Execute(ctx, params)

	if result.Success {
		t.Error("Result should not be successful for unsupported project")
	}
}

func TestTestRunner_Execute_Single(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-single-*")
	defer os.RemoveAll(tmpDir)

	tr := NewTestRunner(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"action": "single",
		"test":   "TestExample",
	}

	result, _ := tr.Execute(ctx, params)

	// May fail for unsupported project, but should not panic
	if !result.Success {
		t.Logf("Single test failed as expected: %s", result.Error)
	}
}

func TestTestRunner_Execute_DefaultAction(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-default-*")
	defer os.RemoveAll(tmpDir)

	tr := NewTestRunner(tmpDir)
	ctx := context.Background()

	// No action specified - should default to "run"
	params := map[string]interface{}{}

	result, _ := tr.Execute(ctx, params)

	// Should fail for unknown project, but not panic
	if !result.Success {
		t.Logf("Default action failed as expected: %s", result.Error)
	}
}
