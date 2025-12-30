package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodeFormatter_Name(t *testing.T) {
	cf := NewCodeFormatter(".")
	if cf.Name() != "code_formatter" {
		t.Errorf("Expected name 'code_formatter', got '%s'", cf.Name())
	}
}

func TestCodeFormatter_Description(t *testing.T) {
	cf := NewCodeFormatter(".")
	desc := cf.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
	if !strings.Contains(desc, "Formata") {
		t.Error("Description should mention 'Formata'")
	}
}

func TestCodeFormatter_RequiresConfirmation(t *testing.T) {
	cf := NewCodeFormatter(".")
	if cf.RequiresConfirmation() {
		t.Error("CodeFormatter should not require confirmation")
	}
}

func TestCodeFormatter_Schema(t *testing.T) {
	cf := NewCodeFormatter(".")
	schema := cf.Schema()

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

	if _, exists := props["language"]; !exists {
		t.Error("Schema should have 'language' property")
	}
}

func TestCodeFormatter_Execute_InvalidAction(t *testing.T) {
	cf := NewCodeFormatter(".")
	ctx := context.Background()

	params := map[string]interface{}{
		"action": "invalid_action",
	}

	result, _ := cf.Execute(ctx, params)

	if result.Success {
		t.Error("Result should not be successful for invalid action")
	}

	if !strings.Contains(result.Error, "desconhecida") {
		t.Error("Error should mention unknown action")
	}
}

func TestCodeFormatter_DetectLanguage(t *testing.T) {
	cf := NewCodeFormatter(".")

	tests := []struct {
		file     string
		expected string
	}{
		{"main.go", "go"},
		{"app.js", "javascript"},
		{"component.jsx", "javascript"},
		{"types.ts", "typescript"},
		{"App.tsx", "typescript"},
		{"script.py", "python"},
		{"lib.rs", "rust"},
		{"Main.java", "java"},
		{"program.c", "c"},
		{"program.cpp", "cpp"},
		{"unknown.txt", ""},
	}

	for _, tt := range tests {
		result := cf.detectLanguage(tt.file)
		if result != tt.expected {
			t.Errorf("detectLanguage(%s) = %s, expected %s", tt.file, result, tt.expected)
		}
	}
}

func TestCodeFormatter_DetectFormatters(t *testing.T) {
	cf := NewCodeFormatter(".")
	ctx := context.Background()

	params := map[string]interface{}{
		"action": "detect",
	}

	result, _ := cf.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	if !strings.Contains(result.Message, "Formatadores") {
		t.Error("Should show formatters detection")
	}

	// Should mention at least gofmt (since we're in a Go project)
	if !strings.Contains(result.Message, "gofmt") {
		t.Error("Should mention gofmt")
	}
}

func TestCodeFormatter_FormatGo_FileNotFound(t *testing.T) {
	cf := NewCodeFormatter(".")
	ctx := context.Background()

	params := map[string]interface{}{
		"action":   "format",
		"language": "go",
		"file":     "nonexistent.go",
	}

	result, _ := cf.Execute(ctx, params)

	// gofmt will fail on nonexistent file
	if result.Success {
		t.Error("Result should not be successful for nonexistent file")
	}
}

func TestCodeFormatter_FormatGo_ValidFile(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-format-*")
	defer os.RemoveAll(tmpDir)

	// Create unformatted Go file
	testFile := filepath.Join(tmpDir, "test.go")
	unformattedCode := `package main

func  main(  ){
	x:=1
		y:=2
	z:=x+y
		println(z)
}
`
	if err := os.WriteFile(testFile, []byte(unformattedCode), 0644); err != nil {
		t.Fatal(err)
	}

	cf := NewCodeFormatter(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"action":   "format",
		"language": "go",
		"file":     testFile,
	}

	result, _ := cf.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	// Read formatted file
	formatted, _ := os.ReadFile(testFile)

	// Check if formatting was applied (spaces should be normalized)
	if strings.Contains(string(formatted), "func  main") {
		t.Error("File should be formatted (extra spaces removed)")
	}
}

func TestCodeFormatter_CheckGo(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-check-*")
	defer os.RemoveAll(tmpDir)

	// Create properly formatted Go file
	testFile := filepath.Join(tmpDir, "test.go")
	formattedCode := `package main

func main() {
	x := 1
	y := 2
	z := x + y
	println(z)
}
`
	if err := os.WriteFile(testFile, []byte(formattedCode), 0644); err != nil {
		t.Fatal(err)
	}

	cf := NewCodeFormatter(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"action":   "check",
		"language": "go",
		"file":     testFile,
	}

	result, _ := cf.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful for formatted file, got: %s", result.Message)
	}
}

func TestCodeFormatter_FormatUnsupportedLanguage(t *testing.T) {
	cf := NewCodeFormatter(".")
	ctx := context.Background()

	params := map[string]interface{}{
		"action":   "format",
		"language": "cobol",
		"file":     "test.cob",
	}

	result, _ := cf.Execute(ctx, params)

	if result.Success {
		t.Error("Result should not be successful for unsupported language")
	}

	if !strings.Contains(result.Error, "não suportada") {
		t.Error("Error should mention unsupported language")
	}
}

func TestCodeFormatter_FormatJavaScript_NoPrettier(t *testing.T) {
	cf := NewCodeFormatter(".")
	ctx := context.Background()

	// Test assumes prettier might not be installed
	params := map[string]interface{}{
		"action":   "format",
		"language": "javascript",
		"file":     "test.js",
	}

	result, _ := cf.Execute(ctx, params)

	// Will either succeed if prettier is installed, or fail with appropriate message
	if !result.Success {
		if !strings.Contains(result.Error, "Prettier") && !strings.Contains(result.Error, "Erro ao formatar") {
			t.Error("Error should mention Prettier or formatting error")
		}
	}
}

func TestCodeFormatter_FormatPython_NoBlack(t *testing.T) {
	cf := NewCodeFormatter(".")
	ctx := context.Background()

	params := map[string]interface{}{
		"action":   "format",
		"language": "python",
		"file":     "test.py",
	}

	result, _ := cf.Execute(ctx, params)

	// Will either succeed if black/autopep8 is installed, or fail with appropriate message
	if !result.Success {
		if !strings.Contains(result.Error, "Black") && !strings.Contains(result.Error, "autopep8") && !strings.Contains(result.Error, "Erro ao formatar") {
			t.Error("Error should mention Black/autopep8 or formatting error")
		}
	}
}

func TestCodeFormatter_CheckUnsupportedLanguage(t *testing.T) {
	cf := NewCodeFormatter(".")
	ctx := context.Background()

	params := map[string]interface{}{
		"action":   "check",
		"language": "rust",
		"file":     "test.rs",
	}

	result, _ := cf.Execute(ctx, params)

	if result.Success {
		t.Error("Result should not be successful for unsupported check language")
	}

	if !strings.Contains(result.Error, "não suportada") {
		t.Error("Error should mention unsupported verification")
	}
}

func TestCodeFormatter_IsCommandAvailable(t *testing.T) {
	cf := NewCodeFormatter(".")

	// gofmt should always be available in Go environment
	if !cf.isCommandAvailable("gofmt") {
		t.Error("gofmt should be available")
	}

	// This command should not exist
	if cf.isCommandAvailable("nonexistent-formatter-xyz-123") {
		t.Error("Nonexistent command should not be available")
	}
}
