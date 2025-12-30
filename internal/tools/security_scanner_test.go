package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSecurityScanner_Name(t *testing.T) {
	ss := NewSecurityScanner(".")
	if ss.Name() != "security_scanner" {
		t.Errorf("Expected name 'security_scanner', got '%s'", ss.Name())
	}
}

func TestSecurityScanner_Description(t *testing.T) {
	ss := NewSecurityScanner(".")
	desc := ss.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
	if !strings.Contains(desc, "vulnerabilidades") {
		t.Error("Description should mention 'vulnerabilidades'")
	}
}

func TestSecurityScanner_RequiresConfirmation(t *testing.T) {
	ss := NewSecurityScanner(".")
	if ss.RequiresConfirmation() {
		t.Error("SecurityScanner should not require confirmation")
	}
}

func TestSecurityScanner_Schema(t *testing.T) {
	ss := NewSecurityScanner(".")
	schema := ss.Schema()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema properties should be a map")
	}

	if _, exists := props["type"]; !exists {
		t.Error("Schema should have 'type' property")
	}
}

func TestSecurityScanner_Execute_InvalidType(t *testing.T) {
	ss := NewSecurityScanner(".")
	ctx := context.Background()

	params := map[string]interface{}{
		"type": "invalid_scan_type",
	}

	result, _ := ss.Execute(ctx, params)


	if result.Success {
		t.Error("Result should not be successful for invalid type")
	}

	if !strings.Contains(result.Error, "desconhecido") {
		t.Error("Error should mention unknown type")
	}
}

func TestSecurityScanner_ScanSecrets_NoSecrets(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-secrets-*")
	defer os.RemoveAll(tmpDir)

	// Create clean Go file
	cleanFile := filepath.Join(tmpDir, "main.go")
	cleanContent := `package main

func main() {
	println("Hello, World!")
}
`
	if err := os.WriteFile(cleanFile, []byte(cleanContent), 0644); err != nil {
		t.Fatal(err)
	}

	ss := NewSecurityScanner(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"type": "secrets",
	}

	result, _ := ss.Execute(ctx, params)


	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	// Should report no secrets found
	if !strings.Contains(result.Message, "Nenhum secret encontrado") {
		t.Error("Should report no secrets found")
	}
}

func TestSecurityScanner_ScanSecrets_WithAPIKey(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-apikey-*")
	defer os.RemoveAll(tmpDir)

	// Create file with API key
	secretFile := filepath.Join(tmpDir, "config.js")
	secretContent := `const config = {
	apiKey: "sk_test_1234567890abcdefghijklmnop"
};
`
	if err := os.WriteFile(secretFile, []byte(secretContent), 0644); err != nil {
		t.Fatal(err)
	}

	ss := NewSecurityScanner(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"type": "secrets",
	}

	result, _ := ss.Execute(ctx, params)


	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	// Should detect potential secret
	if !strings.Contains(result.Message, "encontrado") {
		t.Error("Should detect API key pattern")
	}
}

func TestSecurityScanner_ScanAll(t *testing.T) {
	ss := NewSecurityScanner(".")
	ctx := context.Background()

	params := map[string]interface{}{
		"type": "all",
	}

	result, _ := ss.Execute(ctx, params)


	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	// Should include all scan types
	if !strings.Contains(result.Message, "Busca por Secrets") {
		t.Error("Should include secrets scan")
	}
	if !strings.Contains(result.Message, "SAST") {
		t.Error("Should include SAST scan")
	}
	if !strings.Contains(result.Message, "Vulnerabilidades em Dependências") {
		t.Error("Should include dependencies scan")
	}
}

func TestSecurityScanner_DetectProjectType(t *testing.T) {
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

			result := detectProjectType(tmpDir)

			if result != tt.expected {
				t.Errorf("Expected project type '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestSecurityScanner_IsTextFile(t *testing.T) {
	tests := []struct {
		ext      string
		expected bool
	}{
		{".go", true},
		{".js", true},
		{".py", true},
		{".txt", true},
		{".md", true},
		{".exe", false},
		{".bin", false},
		{".dll", false},
		{"", false},
	}

	for _, tt := range tests {
		result := isTextFile(tt.ext)
		if result != tt.expected {
			t.Errorf("isTextFile(%s) = %v, expected %v", tt.ext, result, tt.expected)
		}
	}
}

func TestSecurityScanner_Execute_DefaultType(t *testing.T) {
	ss := NewSecurityScanner(".")
	ctx := context.Background()

	// No type specified - should default to "all"
	params := map[string]interface{}{}

	result, _ := ss.Execute(ctx, params)


	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}
}
