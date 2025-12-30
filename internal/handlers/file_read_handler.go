package handlers

import (
	"context"
	"fmt"

	"github.com/johnpitter/ollama-code/internal/intent"
)

// FileReadHandler processa leitura de arquivos
type FileReadHandler struct {
	BaseHandler
}

// NewFileReadHandler cria novo handler
func NewFileReadHandler() *FileReadHandler {
	return &FileReadHandler{
		BaseHandler: NewBaseHandler("file_read"),
	}
}

// Handle processa intent de leitura
func (h *FileReadHandler) Handle(ctx context.Context, deps *Dependencies, result *intent.DetectionResult) (string, error) {
	// Extrair parâmetros
	filePath, ok := result.Parameters["file_path"].(string)
	if !ok || filePath == "" {
		return "", fmt.Errorf("file_path não especificado")
	}

	// Executar via tool registry
	params := map[string]interface{}{
		"file_path": filePath,
	}

	toolResult, err := deps.ToolRegistry.Execute(ctx, "file_reader", params)
	if err != nil {
		return "", fmt.Errorf("erro ao ler arquivo: %w", err)
	}

	if !toolResult.Success {
		return "", fmt.Errorf("erro: %s", toolResult.Error)
	}

	// Adicionar arquivo aos recentes
	deps.RecentFiles = append(deps.RecentFiles, filePath)

	return toolResult.Message, nil
}
