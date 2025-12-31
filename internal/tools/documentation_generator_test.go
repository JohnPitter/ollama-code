package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocumentationGenerator_Name(t *testing.T) {
	dg := NewDocumentationGenerator(".")
	if dg.Name() != "documentation_generator" {
		t.Errorf("Expected name 'documentation_generator', got '%s'", dg.Name())
	}
}

func TestDocumentationGenerator_Description(t *testing.T) {
	dg := NewDocumentationGenerator(".")
	desc := dg.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
	if !strings.Contains(desc, "documentação") {
		t.Error("Description should mention 'documentação'")
	}
}

func TestDocumentationGenerator_RequiresConfirmation(t *testing.T) {
	dg := NewDocumentationGenerator(".")
	if dg.RequiresConfirmation() {
		t.Error("DocumentationGenerator should not require confirmation")
	}
}

func TestDocumentationGenerator_Schema(t *testing.T) {
	dg := NewDocumentationGenerator(".")
	schema := dg.Schema()

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

	if _, exists := props["target"]; !exists {
		t.Error("Schema should have 'target' property")
	}
}

func TestDocumentationGenerator_Execute_InvalidType(t *testing.T) {
	dg := NewDocumentationGenerator(".")
	ctx := context.Background()

	params := map[string]interface{}{
		"type": "invalid_type",
	}

	result, _ := dg.Execute(ctx, params)

	if result.Success {
		t.Error("Result should not be successful for invalid type")
	}

	if !strings.Contains(result.Error, "desconhecido") {
		t.Error("Error should mention unknown type")
	}
}

func TestDocumentationGenerator_Execute_GenerateReadme(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-readme-*")
	defer os.RemoveAll(tmpDir)

	dg := NewDocumentationGenerator(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"type": "readme",
	}

	result, _ := dg.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	// Check if README.md was created
	readmePath := filepath.Join(tmpDir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		t.Error("README.md should have been created")
	}

	// Check content
	content, _ := os.ReadFile(readmePath)

	if len(content) == 0 {
		t.Error("README.md should not be empty")
	}

	// Check for project name (directory name)
	projectName := filepath.Base(tmpDir)
	if !strings.Contains(string(content), projectName) {
		t.Error("README.md should contain project name")
	}
}

func TestDocumentationGenerator_Execute_Auto_GoProject(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-go-doc-*")
	defer os.RemoveAll(tmpDir)

	// Create go.mod
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module test\n\ngo 1.21\n"), 0644); err != nil {
		t.Fatal(err)
	}

	dg := NewDocumentationGenerator(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"type": "auto",
	}

	result, _ := dg.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	// Should detect Go project
	if !strings.Contains(result.Message, "Go detectado") {
		t.Error("Should detect Go project")
	}
}

func TestDocumentationGenerator_Execute_Auto_NodeProject(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-node-doc-*")
	defer os.RemoveAll(tmpDir)

	// Create package.json
	pkgPath := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(pkgPath, []byte(`{"name":"test"}`), 0644); err != nil {
		t.Fatal(err)
	}

	dg := NewDocumentationGenerator(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"type": "auto",
	}

	result, _ := dg.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	// Should detect Node.js project
	if !strings.Contains(result.Message, "Node.js detectado") {
		t.Error("Should detect Node.js project")
	}
}

func TestDocumentationGenerator_Execute_APIDoc(t *testing.T) {
	dg := NewDocumentationGenerator(".")
	ctx := context.Background()

	params := map[string]interface{}{
		"type": "api",
	}

	result, _ := dg.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	// Should mention OpenAPI/Swagger
	if !strings.Contains(result.Message, "OpenAPI") && !strings.Contains(result.Message, "Swagger") {
		t.Error("Should mention OpenAPI or Swagger")
	}
}

func TestDocumentationGenerator_Execute_DefaultType(t *testing.T) {
	dg := NewDocumentationGenerator(".")
	ctx := context.Background()

	// No type specified - should default to "auto"
	params := map[string]interface{}{}

	result, _ := dg.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}
}
