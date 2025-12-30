package handlers

import (
	"context"
	"fmt"

	"github.com/johnpitter/ollama-code/internal/intent"
)

// AnalyzeHandler processa análise de projeto
type AnalyzeHandler struct {
	BaseHandler
}

// NewAnalyzeHandler cria novo handler
func NewAnalyzeHandler() *AnalyzeHandler {
	return &AnalyzeHandler{
		BaseHandler: NewBaseHandler("analyze"),
	}
}

// Handle processa intent de análise
func (h *AnalyzeHandler) Handle(ctx context.Context, deps *Dependencies, result *intent.DetectionResult) (string, error) {
	// Extrair parâmetros
	target, _ := result.Parameters["target"].(string)
	if target == "" {
		target = deps.WorkDir // Analisar diretório atual
	}

	// Executar via tool registry
	params := map[string]interface{}{
		"target": target,
	}

	toolResult, err := deps.ToolRegistry.Execute(ctx, "project_analyzer", params)
	if err != nil {
		return "", fmt.Errorf("erro ao analisar projeto: %w", err)
	}

	if !toolResult.Success {
		return "", fmt.Errorf("erro: %s", toolResult.Error)
	}

	return toolResult.Message, nil
}
