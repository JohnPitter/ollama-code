package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/johnpitter/ollama-code/internal/intent"
)

// SearchHandler processa busca em código
type SearchHandler struct {
	BaseHandler
}

// NewSearchHandler cria novo handler
func NewSearchHandler() *SearchHandler {
	return &SearchHandler{
		BaseHandler: NewBaseHandler("search"),
	}
}

// Handle processa intent de busca
func (h *SearchHandler) Handle(ctx context.Context, deps *Dependencies, result *intent.DetectionResult) (string, error) {
	// Extrair query dos parâmetros ou da mensagem do usuário
	query, ok := result.Parameters["query"].(string)
	if !ok || query == "" {
		// Fallback: extrair da mensagem do usuário
		query = extractQueryFromMessage(result.UserMessage)
		if query == "" {
			return "", fmt.Errorf("não foi possível determinar o que buscar. " +
				"Exemplo de uso: 'busca a função ProcessMessage' ou 'procure por database connection'")
		}
	}

	pattern, _ := result.Parameters["pattern"].(string)
	if pattern == "" {
		pattern = query
	}

	// Executar via tool registry
	params := map[string]interface{}{
		"query":   query,
		"pattern": pattern,
	}

	toolResult, err := deps.ToolRegistry.Execute(ctx, "code_searcher", params)
	if err != nil {
		return "", fmt.Errorf("erro ao buscar código: %w", err)
	}

	if !toolResult.Success {
		return "", fmt.Errorf("erro: %s", toolResult.Error)
	}

	return toolResult.Message, nil
}

// extractQueryFromMessage extrai o termo de busca da mensagem do usuário
func extractQueryFromMessage(message string) string {
	// Remove palavras comuns de busca para extrair o termo
	message = strings.ToLower(message)

	// Padrões comuns: "busca X", "procure por X", "encontre X", "onde está X"
	replacements := []string{
		"busca a função ", "busca função ", "busca o ", "busca a ", "busca ",
		"procure por ", "procure ", "procura ",
		"encontre a ", "encontre o ", "encontre ",
		"onde está ", "onde está a ", "onde está o ",
		"acha a ", "acha o ", "acha ",
		"search for ", "search ", "find ",
	}

	query := message
	for _, prefix := range replacements {
		if strings.HasPrefix(query, prefix) {
			query = strings.TrimPrefix(query, prefix)
			break
		}
	}

	// Remover aspas se houver
	query = strings.Trim(query, "\"'")
	query = strings.TrimSpace(query)

	// Se a query ficou muito curta (< 2 chars), retornar a mensagem original sem os prefixos comuns
	if len(query) < 2 {
		return message
	}

	return query
}
