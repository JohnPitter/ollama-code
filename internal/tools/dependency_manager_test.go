package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDependencyManager_Name(t *testing.T) {
	dm := NewDependencyManager(".")
	if dm.Name() != "dependency_manager" {
		t.Errorf("Expected name 'dependency_manager', got '%s'", dm.Name())
	}
}

func TestDependencyManager_Description(t *testing.T) {
	dm := NewDependencyManager(".")
	desc := dm.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
	if !strings.Contains(desc, "dependências") {
		t.Error("Description should mention 'dependências'")
	}
}

func TestDependencyManager_RequiresConfirmation(t *testing.T) {
	dm := NewDependencyManager(".")
	if dm.RequiresConfirmation() {
		t.Error("DependencyManager should not require confirmation by default")
	}
}

func TestDependencyManager_Schema(t *testing.T) {
	dm := NewDependencyManager(".")
	schema := dm.Schema()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}

	if schema["type"] != "object" {
		t.Error("Schema type should be 'object'")
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema properties should be a map")
	}

	if _, exists := props["operation"]; !exists {
		t.Error("Schema should have 'operation' property")
	}

	if _, exists := props["package"]; !exists {
		t.Error("Schema should have 'package' property")
	}
}

func TestDependencyManager_DetectProjectType_Go(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-go-*")
	defer os.RemoveAll(tmpDir)

	// Create go.mod
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	dm := NewDependencyManager(tmpDir)
	projectType := dm.detectProjectType()

	if projectType != "go" {
		t.Errorf("Expected project type 'go', got '%s'", projectType)
	}
}

func TestDependencyManager_DetectProjectType_NodeJS(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-node-*")
	defer os.RemoveAll(tmpDir)

	// Create package.json
	pkgPath := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(pkgPath, []byte(`{"name":"test"}`), 0644); err != nil {
		t.Fatal(err)
	}

	dm := NewDependencyManager(tmpDir)
	projectType := dm.detectProjectType()

	if projectType != "nodejs" {
		t.Errorf("Expected project type 'nodejs', got '%s'", projectType)
	}
}

func TestDependencyManager_DetectProjectType_Python(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-python-*")
	defer os.RemoveAll(tmpDir)

	// Create requirements.txt
	reqPath := filepath.Join(tmpDir, "requirements.txt")
	if err := os.WriteFile(reqPath, []byte("requests==2.28.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	dm := NewDependencyManager(tmpDir)
	projectType := dm.detectProjectType()

	if projectType != "python" {
		t.Errorf("Expected project type 'python', got '%s'", projectType)
	}
}

func TestDependencyManager_DetectProjectType_Unknown(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-unknown-*")
	defer os.RemoveAll(tmpDir)

	dm := NewDependencyManager(tmpDir)
	projectType := dm.detectProjectType()

	if projectType != "unknown" {
		t.Errorf("Expected project type 'unknown', got '%s'", projectType)
	}
}

func TestDependencyManager_Execute_InvalidOperation(t *testing.T) {
	dm := NewDependencyManager(".")
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "invalid_operation",
	}

	result, _ := dm.Execute(ctx, params)


	if result.Success {
		t.Error("Result should not be successful for invalid operation")
	}

	if result.Error == "" {
		t.Error("Result should have error message")
	}

	if !strings.Contains(result.Error, "desconhecida") {
		t.Error("Error should mention unknown operation")
	}
}

func TestDependencyManager_Execute_Check_UnknownProject(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-check-*")
	defer os.RemoveAll(tmpDir)

	dm := NewDependencyManager(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "check",
	}

	result, _ := dm.Execute(ctx, params)


	if result.Success {
		t.Error("Result should not be successful for unknown project type")
	}
}

func TestDependencyManager_Execute_Install_MissingPackage(t *testing.T) {
	dm := NewDependencyManager(".")
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "install",
		// Missing "package" parameter
	}

	result, _ := dm.Execute(ctx, params)


	if result.Success {
		t.Error("Result should not be successful when package is missing")
	}

	if !strings.Contains(result.Error, "não especificado") {
		t.Error("Error should mention package not specified")
	}
}

func TestDependencyManager_Execute_DefaultOperation(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-default-*")
	defer os.RemoveAll(tmpDir)

	dm := NewDependencyManager(tmpDir)
	ctx := context.Background()

	// No operation specified - should default to "check"
	params := map[string]interface{}{}

	result, _ := dm.Execute(ctx, params)


	// Should attempt to check (will fail for unknown project, but shouldn't panic)
	if result.Success {
		t.Log("Check succeeded (unexpected but ok)")
	} else {
		t.Logf("Check failed as expected: %s", result.Error)
	}
}

func TestDependencyManager_Execute_Update_UnknownProject(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-update-*")
	defer os.RemoveAll(tmpDir)

	dm := NewDependencyManager(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "update",
	}

	result, _ := dm.Execute(ctx, params)


	if result.Success {
		t.Error("Result should not be successful for unknown project")
	}
}

func TestDependencyManager_Execute_Audit_UnknownProject(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-audit-*")
	defer os.RemoveAll(tmpDir)

	dm := NewDependencyManager(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "audit",
	}

	result, _ := dm.Execute(ctx, params)


	if result.Success {
		t.Error("Result should not be successful for unknown project")
	}
}
