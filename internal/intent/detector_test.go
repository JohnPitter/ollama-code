package intent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/johnpitter/ollama-code/internal/llm"
)

func TestNewDetector(t *testing.T) {
	client := llm.NewClient("http://localhost:11434", "test-model")
	detector := NewDetector(client)

	if detector == nil {
		t.Fatal("Detector should not be nil")
	}

	if detector.llmClient == nil {
		t.Error("LLM client should not be nil")
	}
}

func TestDetect_ReadFile(t *testing.T) {
	// Mock server that returns read_file intent
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := DetectionResult{
			Intent:     IntentReadFile,
			Confidence: 0.95,
			Parameters: map[string]interface{}{
				"file_path": "config.json",
			},
			Reasoning: "User wants to read a file",
		}

		jsonResponse, _ := json.Marshal(response)
		mockResp := llm.Response{
			Message: llm.Message{
				Content: string(jsonResponse),
			},
			Done: true,
		}

		json.NewEncoder(w).Encode(mockResp)
	}))
	defer server.Close()

	client := llm.NewClient(server.URL, "test-model")
	detector := NewDetector(client)

	result, err := detector.Detect(context.Background(), "read config.json", "/tmp", []string{})

	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if result.Intent != IntentReadFile {
		t.Errorf("Expected intent 'read_file', got '%s'", result.Intent)
	}

	if result.Confidence != 0.95 {
		t.Errorf("Expected confidence 0.95, got %.2f", result.Confidence)
	}

	filePath, ok := result.Parameters["file_path"].(string)
	if !ok {
		t.Error("Expected file_path parameter as string")
	}

	if filePath != "config.json" {
		t.Errorf("Expected file_path 'config.json', got '%s'", filePath)
	}
}

func TestDetect_WriteFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := DetectionResult{
			Intent:     IntentWriteFile,
			Confidence: 0.90,
			Parameters: map[string]interface{}{
				"file_path": "output.txt",
				"content":   "Hello World",
				"mode":      "create",
			},
			Reasoning: "User wants to create a file",
		}

		jsonResponse, _ := json.Marshal(response)
		mockResp := llm.Response{
			Message: llm.Message{Content: string(jsonResponse)},
			Done:    true,
		}

		json.NewEncoder(w).Encode(mockResp)
	}))
	defer server.Close()

	client := llm.NewClient(server.URL, "test-model")
	detector := NewDetector(client)

	result, err := detector.Detect(context.Background(), "create output.txt with 'Hello World'", "/tmp", []string{})

	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if result.Intent != IntentWriteFile {
		t.Errorf("Expected intent 'write_file', got '%s'", result.Intent)
	}

	if result.Parameters["file_path"] != "output.txt" {
		t.Errorf("Unexpected file_path: %v", result.Parameters["file_path"])
	}

	if result.Parameters["content"] != "Hello World" {
		t.Errorf("Unexpected content: %v", result.Parameters["content"])
	}
}

func TestDetect_ExecuteCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := DetectionResult{
			Intent:     IntentExecuteCommand,
			Confidence: 0.98,
			Parameters: map[string]interface{}{
				"command": "ls -la",
			},
			Reasoning: "User wants to execute a shell command",
		}

		jsonResponse, _ := json.Marshal(response)
		mockResp := llm.Response{
			Message: llm.Message{Content: string(jsonResponse)},
			Done:    true,
		}

		json.NewEncoder(w).Encode(mockResp)
	}))
	defer server.Close()

	client := llm.NewClient(server.URL, "test-model")
	detector := NewDetector(client)

	result, err := detector.Detect(context.Background(), "execute ls -la", "/tmp", []string{})

	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if result.Intent != IntentExecuteCommand {
		t.Errorf("Expected intent 'execute_command', got '%s'", result.Intent)
	}

	if result.Parameters["command"] != "ls -la" {
		t.Errorf("Unexpected command: %v", result.Parameters["command"])
	}
}

func TestDetect_Question(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := DetectionResult{
			Intent:     IntentQuestion,
			Confidence: 1.0,
			Parameters: map[string]interface{}{},
			Reasoning:  "Simple question",
		}

		jsonResponse, _ := json.Marshal(response)
		mockResp := llm.Response{
			Message: llm.Message{Content: string(jsonResponse)},
			Done:    true,
		}

		json.NewEncoder(w).Encode(mockResp)
	}))
	defer server.Close()

	client := llm.NewClient(server.URL, "test-model")
	detector := NewDetector(client)

	result, err := detector.Detect(context.Background(), "What is Go?", "/tmp", []string{})

	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if result.Intent != IntentQuestion {
		t.Errorf("Expected intent 'question', got '%s'", result.Intent)
	}

	if result.Confidence != 1.0 {
		t.Errorf("Expected confidence 1.0, got %.2f", result.Confidence)
	}
}

func TestDetect_WebSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := DetectionResult{
			Intent:     IntentWebSearch,
			Confidence: 0.92,
			Parameters: map[string]interface{}{
				"query": "golang best practices",
			},
			Reasoning: "User wants to search the web",
		}

		jsonResponse, _ := json.Marshal(response)
		mockResp := llm.Response{
			Message: llm.Message{Content: string(jsonResponse)},
			Done:    true,
		}

		json.NewEncoder(w).Encode(mockResp)
	}))
	defer server.Close()

	client := llm.NewClient(server.URL, "test-model")
	detector := NewDetector(client)

	result, err := detector.Detect(context.Background(), "search golang best practices", "/tmp", []string{})

	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if result.Intent != IntentWebSearch {
		t.Errorf("Expected intent 'web_search', got '%s'", result.Intent)
	}

	if result.Parameters["query"] != "golang best practices" {
		t.Errorf("Unexpected query: %v", result.Parameters["query"])
	}
}

func TestDetect_WithRecentFiles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify that recent files were included in prompt
		var req llm.Request
		json.NewDecoder(r.Body).Decode(&req)

		// The user message should contain recent files context
		userMessage := req.Messages[len(req.Messages)-1].Content
		if !contains(userMessage, "file1.go") || !contains(userMessage, "file2.go") {
			t.Error("Recent files should be included in context")
		}

		response := DetectionResult{
			Intent:     IntentQuestion,
			Confidence: 1.0,
			Parameters: map[string]interface{}{},
			Reasoning:  "Question",
		}

		jsonResponse, _ := json.Marshal(response)
		mockResp := llm.Response{
			Message: llm.Message{Content: string(jsonResponse)},
			Done:    true,
		}

		json.NewEncoder(w).Encode(mockResp)
	}))
	defer server.Close()

	client := llm.NewClient(server.URL, "test-model")
	detector := NewDetector(client)

	recentFiles := []string{"file1.go", "file2.go", "file3.go"}
	result, err := detector.Detect(context.Background(), "What are these files?", "/tmp", recentFiles)

	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}
}

func TestDetect_InvalidJSON_Fallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return invalid JSON
		mockResp := llm.Response{
			Message: llm.Message{Content: "This is not JSON"},
			Done:    true,
		}

		json.NewEncoder(w).Encode(mockResp)
	}))
	defer server.Close()

	client := llm.NewClient(server.URL, "test-model")
	detector := NewDetector(client)

	result, err := detector.Detect(context.Background(), "Hello", "/tmp", []string{})

	// Should not error, but fallback to question
	if err != nil {
		t.Fatalf("Detect should not fail with invalid JSON, got: %v", err)
	}

	if result.Intent != IntentQuestion {
		t.Errorf("Expected fallback to 'question', got '%s'", result.Intent)
	}

	if result.Confidence != 0.5 {
		t.Errorf("Expected fallback confidence 0.5, got %.2f", result.Confidence)
	}

	if result.Reasoning != "Fallback: não foi possível detectar intenção específica" {
		t.Errorf("Unexpected reasoning: %s", result.Reasoning)
	}
}

func TestParseResponse_WithMarkdown(t *testing.T) {
	client := llm.NewClient("http://localhost:11434", "test-model")
	detector := NewDetector(client)

	// Response wrapped in markdown code block
	response := "```json\n{\n  \"intent\": \"read_file\",\n  \"confidence\": 0.95,\n  \"parameters\": {\"file_path\": \"test.txt\"},\n  \"reasoning\": \"Test\"\n}\n```"

	result, err := detector.parseResponse(response)

	if err != nil {
		t.Fatalf("parseResponse failed: %v", err)
	}

	if result.Intent != IntentReadFile {
		t.Errorf("Expected intent 'read_file', got '%s'", result.Intent)
	}
}

func TestParseResponse_PlainJSON(t *testing.T) {
	client := llm.NewClient("http://localhost:11434", "test-model")
	detector := NewDetector(client)

	response := `{"intent": "question", "confidence": 1.0, "parameters": {}, "reasoning": "Test"}`

	result, err := detector.parseResponse(response)

	if err != nil {
		t.Fatalf("parseResponse failed: %v", err)
	}

	if result.Intent != IntentQuestion {
		t.Errorf("Expected intent 'question', got '%s'", result.Intent)
	}

	if result.Confidence != 1.0 {
		t.Errorf("Expected confidence 1.0, got %.2f", result.Confidence)
	}
}

func TestParseResponse_InvalidJSON(t *testing.T) {
	client := llm.NewClient("http://localhost:11434", "test-model")
	detector := NewDetector(client)

	response := "This is not JSON at all"

	_, err := detector.parseResponse(response)

	if err == nil {
		t.Fatal("Expected error for invalid JSON")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
