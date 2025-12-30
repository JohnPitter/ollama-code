package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/johnpitter/ollama-code/internal/intent"
	"github.com/johnpitter/ollama-code/internal/websearch"
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

	// Check if we should summarize with LLM (for factual queries)
	if h.shouldSummarize(result.UserMessage, query) {
		return h.summarizeWithLLM(ctx, deps, query, searchResults)
	}

	// Format results for presentation (fallback)
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
		// Structured results with "results" key
		if items, ok := v["results"].([]interface{}); ok {
			h.formatResultItems(&sb, items)
		} else {
			// Fallback to general map formatting
			for key, value := range v {
				sb.WriteString(fmt.Sprintf("**%s:** %v\n", key, value))
			}
		}
	case []interface{}:
		// Array of results
		h.formatResultItems(&sb, v)
	default:
		// Fallback - usar reflection para tentar extrair dados
		sb.WriteString(fmt.Sprintf("%v", results))
	}

	if sb.Len() == len(fmt.Sprintf("🔍 Resultados da busca por: %s\n\n", query)) {
		sb.WriteString("Nenhum resultado encontrado.")
	}

	return sb.String()
}

// formatResultItems formats an array of result items
func (h *WebSearchHandler) formatResultItems(sb *strings.Builder, items []interface{}) {
	for i, item := range items {
		// Try to extract fields using map interface
		if result, ok := item.(map[string]interface{}); ok {
			sb.WriteString(fmt.Sprintf("%d. ", i+1))

			if title, ok := result["title"].(string); ok {
				sb.WriteString(fmt.Sprintf("**%s**\n", title))
			} else if title, ok := result["Title"].(string); ok {
				sb.WriteString(fmt.Sprintf("**%s**\n", title))
			}

			if snippet, ok := result["snippet"].(string); ok {
				sb.WriteString(fmt.Sprintf("   %s\n", snippet))
			} else if snippet, ok := result["Snippet"].(string); ok {
				sb.WriteString(fmt.Sprintf("   %s\n", snippet))
			}

			if url, ok := result["url"].(string); ok {
				sb.WriteString(fmt.Sprintf("   🔗 %s\n", url))
			} else if url, ok := result["URL"].(string); ok {
				sb.WriteString(fmt.Sprintf("   🔗 %s\n", url))
			}

			sb.WriteString("\n")
		} else {
			// Fallback: usar reflection via fmt.Sprintf com %+v para ver campos
			itemStr := fmt.Sprintf("%+v", item)

			// Tentar extrair Title, URL, Snippet manualmente
			title := h.extractField(itemStr, "Title:")
			url := h.extractField(itemStr, "URL:")
			snippet := h.extractField(itemStr, "Snippet:")

			if title != "" || url != "" {
				sb.WriteString(fmt.Sprintf("%d. ", i+1))
				if title != "" {
					sb.WriteString(fmt.Sprintf("**%s**\n", title))
				}
				if snippet != "" {
					sb.WriteString(fmt.Sprintf("   %s\n", snippet))
				}
				if url != "" {
					sb.WriteString(fmt.Sprintf("   🔗 %s\n", url))
				}
				sb.WriteString("\n")
			}
		}
	}
}

// extractField extracts a field value from a struct string representation
func (h *WebSearchHandler) extractField(structStr, fieldName string) string {
	idx := strings.Index(structStr, fieldName)
	if idx == -1 {
		return ""
	}

	start := idx + len(fieldName)
	end := strings.IndexAny(structStr[start:], " }")
	if end == -1 {
		return strings.TrimSpace(structStr[start:])
	}

	return strings.TrimSpace(structStr[start : start+end])
}

// shouldSummarize determines if we should use LLM to summarize results
func (h *WebSearchHandler) shouldSummarize(userMessage, query string) bool {
	// Always summarize for better user experience
	// This provides context-aware answers instead of raw search results
	return true
}

// summarizeWithLLM uses LLM to create a summary from search results
func (h *WebSearchHandler) summarizeWithLLM(ctx context.Context, deps *Dependencies, query string, searchResults interface{}) (string, error) {
	// Extract URLs, snippets and sources from results
	var urls []string
	var snippets []string
	var sources []string

	switch v := searchResults.(type) {
	case []interface{}:
		for _, item := range v {
			if result, ok := item.(map[string]interface{}); ok {
				if url, ok := result["url"].(string); ok && url != "" {
					urls = append(urls, url)
					if title, ok := result["title"].(string); ok && title != "" {
						sources = append(sources, fmt.Sprintf("- %s: %s", title, url))
					} else {
						sources = append(sources, fmt.Sprintf("- %s", url))
					}
				}
				if snippet, ok := result["snippet"].(string); ok && snippet != "" {
					snippets = append(snippets, snippet)
				}
			}
		}
	}

	if len(urls) == 0 {
		return "Nenhum resultado encontrado para processar.", nil
	}

	// Fetch actual content from top 2 pages to get real data (temperature, etc.)
	contentSnippets := h.fetchPageContents(ctx, urls, 2)

	// Fallback: if fetch failed, use original snippets
	if len(contentSnippets) == 0 {
		contentSnippets = snippets
		if len(contentSnippets) == 0 {
			return "Não foi possível obter informações detalhadas.", nil
		}
	}

	// Build prompt for LLM with actual page content
	var prompt strings.Builder
	prompt.WriteString("Com base no conteúdo das páginas abaixo, forneça um resumo objetivo e direto respondendo à pergunta do usuário.\n\n")
	prompt.WriteString(fmt.Sprintf("Pergunta: %s\n\n", query))
	prompt.WriteString("Conteúdo das páginas:\n\n")

	for i, content := range contentSnippets {
		prompt.WriteString(fmt.Sprintf("=== Fonte %d ===\n%s\n\n", i+1, content))
	}

	prompt.WriteString("\n\nInstruções:\n")
	prompt.WriteString("- Forneça um resumo conciso e objetivo (2-4 frases)\n")
	prompt.WriteString("- Responda diretamente à pergunta com DADOS ESPECÍFICOS (números, valores, etc.)\n")
	prompt.WriteString("- Se for sobre clima/temperatura, SEMPRE mencione os graus, sensação térmica, etc.\n")
	prompt.WriteString("- Se for sobre valores/preços, mencione os números\n")
	prompt.WriteString("- Use informações específicas do conteúdo acima\n")
	prompt.WriteString("- Seja natural e conversacional\n")
	prompt.WriteString("- NÃO mencione 'de acordo com os resultados' ou 'segundo a busca'\n")
	prompt.WriteString("- NÃO adicione fontes (serão adicionadas automaticamente)\n\n")
	prompt.WriteString("Resumo:")

	// Call LLM
	summary, err := deps.LLMClient.Complete(ctx, prompt.String())
	if err != nil {
		// Fallback to formatted results if LLM fails
		return h.formatSearchResults(query, searchResults), nil
	}

	// Build final response with summary + sources
	var response strings.Builder
	response.WriteString(strings.TrimSpace(summary))
	response.WriteString("\n\n")

	if len(sources) > 0 {
		response.WriteString("📚 **Fontes:**\n")
		for i, source := range sources {
			if i >= 3 { // Limit to top 3 sources
				break
			}
			response.WriteString(source)
			response.WriteString("\n")
		}
	}

	return response.String(), nil
}

// fetchPageContents fetches actual content from web pages
func (h *WebSearchHandler) fetchPageContents(ctx context.Context, urls []string, maxPages int) []string {
	if maxPages <= 0 || maxPages > len(urls) {
		maxPages = len(urls)
	}

	// Limit URLs to fetch
	urlsToFetch := urls[:maxPages]

	// Create content fetcher
	fetcher := websearch.NewContentFetcher()

	// Fetch content in parallel
	fetchedContents := fetcher.FetchMultiple(ctx, urlsToFetch, 2)

	// Extract text content
	var contents []string
	for _, fetched := range fetchedContents {
		if fetched.Error == "" && fetched.Content != "" {
			// Limit content size to avoid overwhelming the LLM
			content := fetched.Content
			if len(content) > 2000 {
				content = content[:2000] + "..."
			}
			contents = append(contents, content)
		}
	}

	return contents
}
