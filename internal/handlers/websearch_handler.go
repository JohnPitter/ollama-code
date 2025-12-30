package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/johnpitter/ollama-code/internal/intent"
)

// WebSearchHandler handles web search operations
type WebSearchHandler struct {
	BaseHandler
}

// NewWebSearchHandler creates a new web search handler
func NewWebSearchHandler() *WebSearchHandler {
	return &WebSearchHandler{
		BaseHandler: NewBaseHandler("websearch"),
	}
}

// Handle executes web search operations
func (h *WebSearchHandler) Handle(ctx context.Context, deps *Dependencies, result *intent.DetectionResult) (string, error) {
	// Extract query from parameters
	query, ok := result.Parameters["query"].(string)
	if !ok || query == "" {
		// Try to extract from user message if not in parameters
		query = h.extractQueryFromMessage(result.UserMessage)
		if query == "" {
			return "", fmt.Errorf("query de busca não especificado")
		}
	}

	// Check if web search client is available
	if deps.WebSearch == nil {
		return "", fmt.Errorf("cliente de busca web não configurado")
	}

	// Execute web search
	searchResults, err := deps.WebSearch.Search(ctx, query)
	if err != nil {
		return "", fmt.Errorf("erro ao buscar: %w", err)
	}

	// Format results for presentation
	response := h.formatSearchResults(query, searchResults)

	return response, nil
}

// extractQueryFromMessage tries to extract search query from user message
func (h *WebSearchHandler) extractQueryFromMessage(message string) string {
	// Remove common search keywords
	message = strings.ToLower(message)
	keywords := []string{"pesquisar", "buscar", "procurar", "search", "find", "lookup"}

	for _, keyword := range keywords {
		if strings.Contains(message, keyword) {
			// Extract text after keyword
			parts := strings.SplitN(message, keyword, 2)
			if len(parts) > 1 {
				query := strings.TrimSpace(parts[1])
				// Remove common prepositions
				query = strings.TrimPrefix(query, "por")
				query = strings.TrimPrefix(query, "sobre")
				query = strings.TrimPrefix(query, "for")
				query = strings.TrimPrefix(query, "about")
				return strings.TrimSpace(query)
			}
		}
	}

	return ""
}

// formatSearchResults formats search results into a readable response
func (h *WebSearchHandler) formatSearchResults(query string, results interface{}) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("🔍 Resultados da busca por: %s\n\n", query))

	// Handle different result formats
	switch v := results.(type) {
	case string:
		// Simple string result
		sb.WriteString(v)
	case map[string]interface{}:
		// Structured results
		if items, ok := v["results"].([]interface{}); ok {
			for i, item := range items {
				if result, ok := item.(map[string]interface{}); ok {
					sb.WriteString(fmt.Sprintf("%d. ", i+1))

					if title, ok := result["title"].(string); ok {
						sb.WriteString(fmt.Sprintf("**%s**\n", title))
					}

					if snippet, ok := result["snippet"].(string); ok {
						sb.WriteString(fmt.Sprintf("   %s\n", snippet))
					}

					if url, ok := result["url"].(string); ok {
						sb.WriteString(fmt.Sprintf("   🔗 %s\n", url))
					}

					sb.WriteString("\n")
				}
			}
		} else {
			// Fallback to general map formatting
			for key, value := range v {
				sb.WriteString(fmt.Sprintf("**%s:** %v\n", key, value))
			}
		}
	case []interface{}:
		// Array of results
		for i, item := range v {
			sb.WriteString(fmt.Sprintf("%d. %v\n", i+1, item))
		}
	default:
		// Fallback
		sb.WriteString(fmt.Sprintf("%v", results))
	}

	return sb.String()
}
