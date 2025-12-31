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
	if message == "" {
		return ""
	}

	messageLower := strings.ToLower(message)

	// Padrões comuns de busca com captura do termo
	patterns := []struct {
		prefix string
		extract func(string, string) string
	}{
		// Português
		{"busca a função ", func(msg, p string) string { return strings.TrimPrefix(msg, p) }},
		{"busca função ", func(msg, p string) string { return strings.TrimPrefix(msg, p) }},
		{"busca o ", func(msg, p string) string { return strings.TrimPrefix(msg, p) }},
		{"busca a ", func(msg, p string) string { return strings.TrimPrefix(msg, p) }},
		{"busca ", func(msg, p string) string { return strings.TrimPrefix(msg, p) }},
		{"buscar ", func(msg, p string) string { return strings.TrimPrefix(msg, p) }},
		{"procure por ", func(msg, p string) string { return strings.TrimPrefix(msg, p) }},
		{"procure ", func(msg, p string) string { return strings.TrimPrefix(msg, p) }},
		{"procura ", func(msg, p string) string { return strings.TrimPrefix(msg, p) }},
		{"procurar ", func(msg, p string) string { return strings.TrimPrefix(msg, p) }},
		{"encontre a ", func(msg, p string) string { return strings.TrimPrefix(msg, p) }},
		{"encontre o ", func(msg, p string) string { return strings.TrimPrefix(msg, p) }},
		{"encontre ", func(msg, p string) string { return strings.TrimPrefix(msg, p) }},
		{"encontrar ", func(msg, p string) string { return strings.TrimPrefix(msg, p) }},
		{"onde está ", func(msg, p string) string { return strings.TrimPrefix(msg, p) }},
		{"onde está a ", func(msg, p string) string { return strings.TrimPrefix(msg, p) }},
		{"onde está o ", func(msg, p string) string { return strings.TrimPrefix(msg, p) }},
		{"acha a ", func(msg, p string) string { return strings.TrimPrefix(msg, p) }},
		{"acha o ", func(msg, p string) string { return strings.TrimPrefix(msg, p) }},
		{"acha ", func(msg, p string) string { return strings.TrimPrefix(msg, p) }},
		{"achar ", func(msg, p string) string { return strings.TrimPrefix(msg, p) }},
		// Inglês
		{"search for ", func(msg, p string) string { return strings.TrimPrefix(msg, p) }},
		{"search ", func(msg, p string) string { return strings.TrimPrefix(msg, p) }},
		{"find ", func(msg, p string) string { return strings.TrimPrefix(msg, p) }},
		{"locate ", func(msg, p string) string { return strings.TrimPrefix(msg, p) }},
		{"look for ", func(msg, p string) string { return strings.TrimPrefix(msg, p) }},
	}

	query := messageLower
	for _, pattern := range patterns {
		if strings.HasPrefix(query, pattern.prefix) {
			query = pattern.extract(query, pattern.prefix)
			break
		}
	}

	// Remover aspas e espaços
	query = strings.Trim(query, "\"'` \t\n\r")

	// Se a query ficou muito curta ou igual à mensagem original,
	// tentar usar a mensagem completa como query
	if len(query) < 2 || query == messageLower {
		// Última tentativa: pegar qualquer coisa após verbos de busca
		words := strings.Fields(messageLower)
		if len(words) > 1 {
			// Se primeira palavra é um verbo de busca, pegar resto
			searchVerbs := []string{"busca", "buscar", "procure", "procurar", "encontre", "encontrar",
				                     "acha", "achar", "search", "find", "locate", "look"}
			for _, verb := range searchVerbs {
				if words[0] == verb || words[0] == verb+"r" {
					query = strings.Join(words[1:], " ")
					break
				}
			}
		}

		// Se ainda não tem query válida, usar mensagem inteira como fallback
		if len(query) < 2 {
			query = messageLower
		}
	}

	return strings.TrimSpace(query)
}
