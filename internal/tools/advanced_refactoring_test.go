package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdvancedRefactoring_Name(t *testing.T) {
	ar := NewAdvancedRefactoring(".")
	if ar.Name() != "advanced_refactoring" {
		t.Errorf("Expected name 'advanced_refactoring', got '%s'", ar.Name())
	}
}

func TestAdvancedRefactoring_Description(t *testing.T) {
	ar := NewAdvancedRefactoring(".")
	desc := ar.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
	if !strings.Contains(desc, "Refatorações") {
		t.Error("Description should mention 'Refatorações'")
	}
}

func TestAdvancedRefactoring_RequiresConfirmation(t *testing.T) {
	ar := NewAdvancedRefactoring(".")
	if ar.RequiresConfirmation() {
		t.Error("AdvancedRefactoring should not require confirmation")
	}
}

func TestAdvancedRefactoring_Schema(t *testing.T) {
	ar := NewAdvancedRefactoring(".")
	schema := ar.Schema()

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

	if _, exists := props["old_name"]; !exists {
		t.Error("Schema should have 'old_name' property")
	}

	if _, exists := props["new_name"]; !exists {
		t.Error("Schema should have 'new_name' property")
	}
}

func TestAdvancedRefactoring_Execute_InvalidType(t *testing.T) {
	ar := NewAdvancedRefactoring(".")
	ctx := context.Background()

	params := map[string]interface{}{
		"type": "invalid_refactor_type",
	}

	result, _ := ar.Execute(ctx, params)


	if result.Success {
		t.Error("Result should not be successful for invalid type")
	}

	if !strings.Contains(result.Error, "desconhecido") {
		t.Error("Error should mention unknown type")
	}
}

func TestAdvancedRefactoring_Execute_MissingType(t *testing.T) {
	ar := NewAdvancedRefactoring(".")
	ctx := context.Background()

	params := map[string]interface{}{}

	result, _ := ar.Execute(ctx, params)


	if result.Success {
		t.Error("Result should not be successful when type is missing")
	}

	if !strings.Contains(result.Error, "não especificado") {
		t.Error("Error should mention type not specified")
	}
}

func TestAdvancedRefactoring_RenameSymbol(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-rename-*")
	defer os.RemoveAll(tmpDir)

	// Create test file
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package main

func oldFunction() {
	println("Hello")
}

func main() {
	oldFunction()
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	ar := NewAdvancedRefactoring(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"type":     "rename",
		"old_name": "oldFunction",
		"new_name": "newFunction",
	}

	result, _ := ar.Execute(ctx, params)


	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	// Read modified file
	modifiedContent, _ := os.ReadFile(testFile)

	// Check that old name was replaced
	if strings.Contains(string(modifiedContent), "oldFunction") {
		t.Error("Old function name should have been replaced")
	}

	// Check that new name exists
	if !strings.Contains(string(modifiedContent), "newFunction") {
		t.Error("New function name should exist in file")
	}
}

func TestAdvancedRefactoring_RenameSymbol_MissingParams(t *testing.T) {
	ar := NewAdvancedRefactoring(".")
	ctx := context.Background()

	params := map[string]interface{}{
		"type": "rename",
		// Missing old_name and new_name
	}

	result, _ := ar.Execute(ctx, params)


	if result.Success {
		t.Error("Result should not be successful when parameters are missing")
	}

	if !strings.Contains(result.Error, "obrigatórios") {
		t.Error("Error should mention required parameters")
	}
}

func TestAdvancedRefactoring_FindDuplicates(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-dup-*")
	defer os.RemoveAll(tmpDir)

	// Create files with duplicate code
	duplicateCode := `func doSomething() {
	x := 1
	y := 2
	z := x + y
	return z
}
`

	file1 := filepath.Join(tmpDir, "file1.go")
	if err := os.WriteFile(file1, []byte(duplicateCode), 0644); err != nil {
		t.Fatal(err)
	}

	file2 := filepath.Join(tmpDir, "file2.go")
	if err := os.WriteFile(file2, []byte(duplicateCode), 0644); err != nil {
		t.Fatal(err)
	}

	ar := NewAdvancedRefactoring(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"type": "find_duplicates",
	}

	result, _ := ar.Execute(ctx, params)


	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	// Should detect duplicates
	if !strings.Contains(result.Message, "duplicad") {
		t.Error("Should detect duplicate code")
	}
}

func TestAdvancedRefactoring_ExtractMethod(t *testing.T) {
	ar := NewAdvancedRefactoring(".")
	ctx := context.Background()

	// Test with missing parameters
	params := map[string]interface{}{
		"type": "extract_method",
	}

	result, _ := ar.Execute(ctx, params)

	if result.Success {
		t.Error("Result should fail when parameters are missing")
	}

	if !strings.Contains(result.Error, "obrigatórios") {
		t.Error("Error should mention required parameters")
	}
}

func TestAdvancedRefactoring_ExtractClass(t *testing.T) {
	ar := NewAdvancedRefactoring(".")
	ctx := context.Background()

	// Test with missing parameters
	params := map[string]interface{}{
		"type": "extract_class",
	}

	result, _ := ar.Execute(ctx, params)

	if result.Success {
		t.Error("Result should fail when parameters are missing")
	}

	if !strings.Contains(result.Error, "obrigatórios") {
		t.Error("Error should mention required parameters")
	}
}

func TestAdvancedRefactoring_Inline(t *testing.T) {
	ar := NewAdvancedRefactoring(".")
	ctx := context.Background()

	// Test with missing parameters
	params := map[string]interface{}{
		"type": "inline",
	}

	result, _ := ar.Execute(ctx, params)

	if result.Success {
		t.Error("Result should fail when parameters are missing")
	}

	if !strings.Contains(result.Error, "obrigatórios") {
		t.Error("Error should mention required parameters")
	}
}

func TestAdvancedRefactoring_Move(t *testing.T) {
	ar := NewAdvancedRefactoring(".")
	ctx := context.Background()

	// Test with missing parameters
	params := map[string]interface{}{
		"type": "move",
	}

	result, _ := ar.Execute(ctx, params)

	if result.Success {
		t.Error("Result should fail when parameters are missing")
	}

	if !strings.Contains(result.Error, "obrigatórios") {
		t.Error("Error should mention required parameters")
	}
}

func TestAdvancedRefactoring_IsCodeFile(t *testing.T) {
	tests := []struct {
		ext      string
		expected bool
	}{
		{".go", true},
		{".js", true},
		{".py", true},
		{".java", true},
		{".txt", false},
		{".md", false},
		{".exe", false},
		{"", false},
	}

	for _, tt := range tests {
		result := isCodeFile(tt.ext)
		if result != tt.expected {
			t.Errorf("isCodeFile(%s) = %v, expected %v", tt.ext, result, tt.expected)
		}
	}
}
