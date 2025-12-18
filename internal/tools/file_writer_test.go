package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFileWriter_Create(t *testing.T) {
	tmpDir := t.TempDir()
	writer := NewFileWriter(tmpDir)

	testFile := "test.txt"
	testContent := "Hello, World!"

	result, err := writer.Execute(context.Background(), map[string]interface{}{
		"file_path": testFile,
		"content":   testContent,
		"mode":      "create",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got error: %s", result.Error)
	}

	// Verify file was created
	fullPath := filepath.Join(tmpDir, testFile)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Expected content '%s', got '%s'", testContent, string(content))
	}
}

func TestFileWriter_Append(t *testing.T) {
	tmpDir := t.TempDir()
	writer := NewFileWriter(tmpDir)

	testFile := "test.txt"
	initialContent := "Line 1\n"
	appendContent := "Line 2\n"

	// Create initial file
	fullPath := filepath.Join(tmpDir, testFile)
	if err := os.WriteFile(fullPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create initial file: %v", err)
	}

	// Append
	result, err := writer.Execute(context.Background(), map[string]interface{}{
		"file_path": testFile,
		"content":   appendContent,
		"mode":      "append",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got error: %s", result.Error)
	}

	// Verify content
	content, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	expected := initialContent + appendContent
	if string(content) != expected {
		t.Errorf("Expected content '%s', got '%s'", expected, string(content))
	}
}

func TestFileWriter_Replace(t *testing.T) {
	tmpDir := t.TempDir()
	writer := NewFileWriter(tmpDir)

	testFile := "test.txt"
	initialContent := "Hello World"
	oldText := "World"
	newText := "Go"

	// Create initial file
	fullPath := filepath.Join(tmpDir, testFile)
	if err := os.WriteFile(fullPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create initial file: %v", err)
	}

	// Replace
	result, err := writer.Execute(context.Background(), map[string]interface{}{
		"file_path": testFile,
		"content":   "", // Replace mode doesn't use content, but validation requires it
		"mode":      "replace",
		"old_text":  oldText,
		"new_text":  newText,
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got error: %s", result.Error)
	}

	// Verify content
	content, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	expected := "Hello Go"
	if string(content) != expected {
		t.Errorf("Expected content '%s', got '%s'", expected, string(content))
	}
}

func TestFileWriter_CreateDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	writer := NewFileWriter(tmpDir)

	testFile := "subdir/nested/test.txt"
	testContent := "Test"

	result, err := writer.Execute(context.Background(), map[string]interface{}{
		"file_path": testFile,
		"content":   testContent,
		"mode":      "create",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected success, got error: %s", result.Error)
	}

	// Verify file was created in nested directory
	fullPath := filepath.Join(tmpDir, testFile)
	if _, err := os.Stat(fullPath); err != nil {
		t.Fatalf("File not created in nested directory: %v", err)
	}
}

func TestFileWriter_MissingParameters(t *testing.T) {
	tmpDir := t.TempDir()
	writer := NewFileWriter(tmpDir)

	tests := []struct {
		name   string
		params map[string]interface{}
	}{
		{
			name:   "missing file_path",
			params: map[string]interface{}{"content": "test"},
		},
		{
			name:   "missing content",
			params: map[string]interface{}{"file_path": "test.txt"},
		},
		{
			name: "missing old_text for replace",
			params: map[string]interface{}{
				"file_path": "test.txt",
				"mode":      "replace",
				"new_text":  "new",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := writer.Execute(context.Background(), tt.params)

			if err != nil {
				t.Fatalf("Execute returned error: %v", err)
			}

			if result.Success {
				t.Error("Expected failure due to missing parameters")
			}
		})
	}
}

func TestFileWriter_Name(t *testing.T) {
	writer := NewFileWriter(".")
	if writer.Name() != "file_writer" {
		t.Errorf("Expected name 'file_writer', got '%s'", writer.Name())
	}
}

func TestFileWriter_RequiresConfirmation(t *testing.T) {
	writer := NewFileWriter(".")
	if !writer.RequiresConfirmation() {
		t.Error("FileWriter should require confirmation")
	}
}
