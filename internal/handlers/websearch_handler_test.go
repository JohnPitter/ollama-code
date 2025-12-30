package handlers

import (
	"context"
	"testing"

	"github.com/johnpitter/ollama-code/internal/intent"
)

func TestWebSearchHandler_Success(t *testing.T) {
	handler := NewWebSearchHandler()
	deps := NewMockDependencies()

	searchCalled := false
	deps.WebSearch = &MockWebSearchClient{
		SearchFunc: func(ctx context.Context, query string) (interface{}, error) {
			searchCalled = true
			AssertEqual(t, "golang best practices", query, "search query")
			return []interface{}{
				map[string]interface{}{
					"title":   "Go Best Practices",
					"snippet": "Best practices for Go programming",
					"url":     "https://example.com/go",
				},
			}, nil
		},
	}

	llmCalled := false
	deps.LLMClient = &MockLLMClient{
		CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
			llmCalled = true
			return "Go possui várias boas práticas recomendadas para programação eficiente e manutenível.", nil
		},
	}

	result := NewMockDetectionResult(intent.IntentWebSearch, map[string]interface{}{
		"query": "golang best practices",
	})

	ctx := context.Background()
	response, err := handler.Handle(ctx, deps, result)

	AssertNoError(t, err)
	AssertNotEmpty(t, response, "response")
	AssertContains(t, response, "Fontes:", "should have sources section")

	if !searchCalled {
		t.Error("Expected web search to be called")
	}
	if !llmCalled {
		t.Error("Expected LLM to be called for summarization")
	}
}

func TestWebSearchHandler_MissingQuery(t *testing.T) {
	handler := NewWebSearchHandler()
	deps := NewMockDependencies()

	result := NewMockDetectionResult(intent.IntentWebSearch, map[string]interface{}{})
	result.UserMessage = ""

	ctx := context.Background()
	_, err := handler.Handle(ctx, deps, result)

	AssertError(t, err, "missing query")
}

func TestWebSearchHandler_ExtractQueryFromMessage(t *testing.T) {
	handler := NewWebSearchHandler()
	deps := NewMockDependencies()

	var capturedQuery string
	deps.WebSearch = &MockWebSearchClient{
		SearchFunc: func(ctx context.Context, query string) (interface{}, error) {
			capturedQuery = query
			return map[string]interface{}{"results": []interface{}{}}, nil
		},
	}

	result := NewMockDetectionResult(intent.IntentWebSearch, map[string]interface{}{})
	result.UserMessage = "pesquisar por golang best practices"

	ctx := context.Background()
	_, err := handler.Handle(ctx, deps, result)

	AssertNoError(t, err)

	if !contains(capturedQuery, "golang") {
		t.Errorf("Expected query to be extracted from message, got: %s", capturedQuery)
	}
}

func TestWebSearchHandler_NoResults(t *testing.T) {
	handler := NewWebSearchHandler()
	deps := NewMockDependencies()

	deps.WebSearch = &MockWebSearchClient{
		SearchFunc: func(ctx context.Context, query string) (interface{}, error) {
			return map[string]interface{}{
				"results": []interface{}{},
			}, nil
		},
	}

	result := NewMockDetectionResult(intent.IntentWebSearch, map[string]interface{}{
		"query": "very specific query that returns nothing",
	})

	ctx := context.Background()
	response, err := handler.Handle(ctx, deps, result)

	AssertNoError(t, err)
	AssertNotEmpty(t, response, "response should still be generated")
}

func TestWebSearchHandler_MultipleResults(t *testing.T) {
	handler := NewWebSearchHandler()
	deps := NewMockDependencies()

	deps.WebSearch = &MockWebSearchClient{
		SearchFunc: func(ctx context.Context, query string) (interface{}, error) {
			return []interface{}{
				map[string]interface{}{
					"title":   "Result 1",
					"snippet": "First result",
					"url":     "https://example.com/1",
				},
				map[string]interface{}{
					"title":   "Result 2",
					"snippet": "Second result",
					"url":     "https://example.com/2",
				},
				map[string]interface{}{
					"title":   "Result 3",
					"snippet": "Third result",
					"url":     "https://example.com/3",
				},
			}, nil
		},
	}

	deps.LLMClient = &MockLLMClient{
		CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
			return "Resumo dos resultados encontrados.", nil
		},
	}

	result := NewMockDetectionResult(intent.IntentWebSearch, map[string]interface{}{
		"query": "test query",
	})

	ctx := context.Background()
	response, err := handler.Handle(ctx, deps, result)

	AssertNoError(t, err)
	AssertContains(t, response, "Fontes:", "should have sources")
}

func TestWebSearchHandler_SearchError(t *testing.T) {
	handler := NewWebSearchHandler()
	deps := NewMockDependencies()

	deps.WebSearch = &MockWebSearchClient{
		SearchFunc: func(ctx context.Context, query string) (interface{}, error) {
			return nil, &testError{msg: "search service unavailable"}
		},
	}

	result := NewMockDetectionResult(intent.IntentWebSearch, map[string]interface{}{
		"query": "test query",
	})

	ctx := context.Background()
	_, err := handler.Handle(ctx, deps, result)

	AssertError(t, err, "search error")
}

func TestWebSearchHandler_NilWebSearch(t *testing.T) {
	handler := NewWebSearchHandler()
	deps := NewMockDependencies()
	deps.WebSearch = nil

	result := NewMockDetectionResult(intent.IntentWebSearch, map[string]interface{}{
		"query": "test query",
	})

	ctx := context.Background()
	_, err := handler.Handle(ctx, deps, result)

	AssertError(t, err, "web search not configured")
}

func TestWebSearchHandler_StringResult(t *testing.T) {
	handler := NewWebSearchHandler()
	deps := NewMockDependencies()

	deps.WebSearch = &MockWebSearchClient{
		SearchFunc: func(ctx context.Context, query string) (interface{}, error) {
			// Retornar string simples ao invés de estrutura
			return "Simple search result", nil
		},
	}

	deps.LLMClient = &MockLLMClient{
		CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
			return "Resultado processado", nil
		},
	}

	result := NewMockDetectionResult(intent.IntentWebSearch, map[string]interface{}{
		"query": "test query",
	})

	ctx := context.Background()
	response, err := handler.Handle(ctx, deps, result)

	AssertNoError(t, err)
	// String results don't have snippets, so should return "Nenhum resultado encontrado"
	AssertNotEmpty(t, response, "response")
}
