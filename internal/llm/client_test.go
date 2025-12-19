package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:11434", "qwen2.5-coder:7b")

	if client == nil {
		t.Fatal("Client should not be nil")
	}

	if client.baseURL != "http://localhost:11434" {
		t.Errorf("Expected baseURL 'http://localhost:11434', got '%s'", client.baseURL)
	}

	if client.model != "qwen2.5-coder:7b" {
		t.Errorf("Expected model 'qwen2.5-coder:7b', got '%s'", client.model)
	}

	if client.httpClient == nil {
		t.Error("HTTP client should not be nil")
	}

	if client.httpClient.Timeout != 300*time.Second {
		t.Errorf("Expected timeout 300s, got %v", client.httpClient.Timeout)
	}
}

func TestGetModel(t *testing.T) {
	client := NewClient("http://localhost:11434", "test-model")

	if client.GetModel() != "test-model" {
		t.Errorf("Expected model 'test-model', got '%s'", client.GetModel())
	}
}

func TestSetModel(t *testing.T) {
	client := NewClient("http://localhost:11434", "old-model")

	client.SetModel("new-model")

	if client.GetModel() != "new-model" {
		t.Errorf("Expected model 'new-model', got '%s'", client.GetModel())
	}
}

func TestComplete_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/api/chat" {
			t.Errorf("Expected path '/api/chat', got '%s'", r.URL.Path)
		}

		// Decode request
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		// Verify request fields
		if req.Model != "test-model" {
			t.Errorf("Expected model 'test-model', got '%s'", req.Model)
		}

		if req.Stream {
			t.Error("Expected Stream to be false")
		}

		if len(req.Messages) != 1 {
			t.Errorf("Expected 1 message, got %d", len(req.Messages))
		}

		// Send response
		resp := Response{
			Message: Message{
				Role:    "assistant",
				Content: "Test response",
			},
			Done: true,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-model")

	messages := []Message{
		{Role: "user", Content: "Test question"},
	}

	response, err := client.Complete(context.Background(), messages, &CompletionOptions{
		Temperature: 0.7,
		MaxTokens:   100,
	})

	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}

	if response != "Test response" {
		t.Errorf("Expected 'Test response', got '%s'", response)
	}
}

func TestComplete_WithSystemPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req Request
		json.NewDecoder(r.Body).Decode(&req)

		// Verify system prompt was prepended
		if len(req.Messages) != 2 {
			t.Errorf("Expected 2 messages (system + user), got %d", len(req.Messages))
		}

		if req.Messages[0].Role != "system" {
			t.Errorf("First message should be system, got '%s'", req.Messages[0].Role)
		}

		if req.Messages[0].Content != "You are a helpful assistant" {
			t.Errorf("Unexpected system prompt: %s", req.Messages[0].Content)
		}

		resp := Response{
			Message: Message{Role: "assistant", Content: "OK"},
			Done:    true,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-model")

	messages := []Message{
		{Role: "user", Content: "Hello"},
	}

	_, err := client.Complete(context.Background(), messages, &CompletionOptions{
		SystemPrompt: "You are a helpful assistant",
	})

	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}
}

func TestComplete_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"model not found"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-model")

	messages := []Message{
		{Role: "user", Content: "Test"},
	}

	_, err := client.Complete(context.Background(), messages, nil)

	if err == nil {
		t.Fatal("Expected error for 404 status")
	}

	if !strings.Contains(err.Error(), "404") {
		t.Errorf("Error should mention status 404, got: %v", err)
	}
}

func TestComplete_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		resp := Response{Message: Message{Content: "Response"}, Done: true}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-model")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	messages := []Message{{Role: "user", Content: "Test"}}

	_, err := client.Complete(ctx, messages, nil)

	if err == nil {
		t.Fatal("Expected error due to context cancellation")
	}
}

func TestCompleteStreaming_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify streaming is enabled
		var req Request
		json.NewDecoder(r.Body).Decode(&req)

		if !req.Stream {
			t.Error("Expected Stream to be true")
		}

		// Send streaming responses
		chunks := []string{"Hello", " ", "World"}
		encoder := json.NewEncoder(w)

		for i, chunk := range chunks {
			resp := Response{
				Message: Message{Content: chunk},
				Done:    i == len(chunks)-1,
			}
			encoder.Encode(resp)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-model")

	var receivedChunks []string
	onChunk := func(chunk string) {
		receivedChunks = append(receivedChunks, chunk)
	}

	messages := []Message{{Role: "user", Content: "Test"}}

	fullResponse, err := client.CompleteStreaming(context.Background(), messages, nil, onChunk)

	if err != nil {
		t.Fatalf("CompleteStreaming failed: %v", err)
	}

	if fullResponse != "Hello World" {
		t.Errorf("Expected 'Hello World', got '%s'", fullResponse)
	}

	if len(receivedChunks) != 3 {
		t.Errorf("Expected 3 chunks, got %d", len(receivedChunks))
	}

	expectedChunks := []string{"Hello", " ", "World"}
	for i, expected := range expectedChunks {
		if receivedChunks[i] != expected {
			t.Errorf("Chunk %d: expected '%s', got '%s'", i, expected, receivedChunks[i])
		}
	}
}

func TestCompleteStreaming_NoCallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encoder := json.NewEncoder(w)
		resp := Response{
			Message: Message{Content: "Test response"},
			Done:    true,
		}
		encoder.Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-model")
	messages := []Message{{Role: "user", Content: "Test"}}

	// Should not panic with nil callback
	fullResponse, err := client.CompleteStreaming(context.Background(), messages, nil, nil)

	if err != nil {
		t.Fatalf("CompleteStreaming failed: %v", err)
	}

	if fullResponse != "Test response" {
		t.Errorf("Expected 'Test response', got '%s'", fullResponse)
	}
}

func TestCompleteStreaming_EmptyChunks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encoder := json.NewEncoder(w)

		// Send empty chunks (should not trigger callback)
		chunks := []struct {
			content string
			done    bool
		}{
			{"", false},
			{"Hello", false},
			{"", false},
			{" World", true},
		}

		for _, chunk := range chunks {
			resp := Response{
				Message: Message{Content: chunk.content},
				Done:    chunk.done,
			}
			encoder.Encode(resp)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-model")

	var callbackCount int
	onChunk := func(chunk string) {
		callbackCount++
	}

	messages := []Message{{Role: "user", Content: "Test"}}

	fullResponse, err := client.CompleteStreaming(context.Background(), messages, nil, onChunk)

	if err != nil {
		t.Fatalf("CompleteStreaming failed: %v", err)
	}

	// Empty chunks should not trigger callback
	if callbackCount != 2 {
		t.Errorf("Expected 2 callback calls (non-empty chunks), got %d", callbackCount)
	}

	if fullResponse != "Hello World" {
		t.Errorf("Expected 'Hello World', got '%s'", fullResponse)
	}
}
