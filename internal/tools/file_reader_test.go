package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFileReader_ReadTextFile(t *testing.T) {
	tmpDir := t.TempDir()
	reader := NewFileReader(tmpDir)

	// Create test file
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "Hello, World!\nThis is a test."
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := reader.Execute(context.Background(), map[string]interface{}{
		"file_path": "test.txt",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got error: %s", result.Error)
	}

	fileType, ok := result.Data["type"].(string)
	if !ok || fileType != "text" {
		t.Errorf("Expected type 'text', got %v", result.Data["type"])
	}

	content, ok := result.Data["content"].(string)
	if !ok {
		t.Fatal("Content is not a string")
	}

	if content != testContent {
		t.Errorf("Expected content '%s', got '%s'", testContent, content)
	}
}

func TestFileReader_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	reader := NewFileReader(tmpDir)

	result, err := reader.Execute(context.Background(), map[string]interface{}{
		"file_path": "nonexistent.txt",
	})

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if result.Success {
		t.Error("Expected failure for non-existent file")
	}
}

func TestFileReader_MissingFilePath(t *testing.T) {
	reader := NewFileReader(".")

	result, err := reader.Execute(context.Background(), map[string]interface{}{})

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if result.Success {
		t.Error("Expected failure for missing file_path parameter")
	}
}

func TestFileReader_Name(t *testing.T) {
	reader := NewFileReader(".")
	if reader.Name() != "file_reader" {
		t.Errorf("Expected name 'file_reader', got '%s'", reader.Name())
	}
}

func TestFileReader_RequiresConfirmation(t *testing.T) {
	reader := NewFileReader(".")
	if reader.RequiresConfirmation() {
		t.Error("FileReader should not require confirmation")
	}
}

func TestFileReader_Description(t *testing.T) {
	reader := NewFileReader(".")
	desc := reader.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
}
