package handlers

import (
	"context"
	"fmt"

	"github.com/johnpitter/ollama-code/internal/intent"
)

// GitHandler processa operações Git
type GitHandler struct {
	BaseHandler
}

// NewGitHandler cria novo handler
func NewGitHandler() *GitHandler {
	return &GitHandler{
		BaseHandler: NewBaseHandler("git"),
	}
}

// Handle processa intent de operação Git
func (h *GitHandler) Handle(ctx context.Context, deps *Dependencies, result *intent.DetectionResult) (string, error) {
	// Extrair parâmetros
	operation, ok := result.Parameters["operation"].(string)
	if !ok || operation == "" {
		operation = "status" // Default
	}

	// Executar via tool registry
	params := map[string]interface{}{
		"operation": operation,
	}

	// Adicionar parâmetros extras se houver
	for key, value := range result.Parameters {
		if key != "operation" {
			params[key] = value
		}
	}

	toolResult, err := deps.ToolRegistry.Execute(ctx, "git_operations", params)
	if err != nil {
		return "", fmt.Errorf("erro na operação git: %w", err)
	}

	if !toolResult.Success {
		return "", fmt.Errorf("erro: %s", toolResult.Error)
	}

	return toolResult.Message, nil
}
