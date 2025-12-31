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

	// Formatar resultado com dados estruturados
	return h.formatAnalysisResult(toolResult), nil
}

// formatAnalysisResult formata resultado da análise com dados estruturados
func (h *AnalyzeHandler) formatAnalysisResult(result ToolResult) string {
	output := result.Message + "\n\n"

	// Se houver árvore de diretórios, exibi-la
	if tree, ok := result.Data["tree"].([]string); ok && len(tree) > 0 {
		output += "📂 Estrutura de Diretórios:\n\n"
		for _, line := range tree {
			output += line + "\n"
		}
		output += "\n"
	}

	// Se houver estatísticas, exibi-las
	if totalFiles, ok := result.Data["total_files"].(int); ok {
		output += fmt.Sprintf("📊 Estatísticas:\n")
		output += fmt.Sprintf("  • Arquivos: %d\n", totalFiles)

		if totalDirs, ok := result.Data["total_dirs"].(int); ok {
			output += fmt.Sprintf("  • Diretórios: %d\n", totalDirs)
		}

		if totalSize, ok := result.Data["total_size"].(int64); ok {
			output += fmt.Sprintf("  • Tamanho total: %.2f MB\n", float64(totalSize)/(1024*1024))
		}

		if fileTypes, ok := result.Data["file_types"].(map[string]int); ok && len(fileTypes) > 0 {
			output += "\n  📄 Tipos de arquivo:\n"
			for ext, count := range fileTypes {
				output += fmt.Sprintf("    %s: %d\n", ext, count)
			}
		}
	}

	// Se houver lista de arquivos, exibir resumo
	if files, ok := result.Data["files"].([]string); ok && len(files) > 0 {
		output += fmt.Sprintf("\n📄 Total de %d arquivos listados\n", len(files))

		// Mostrar primeiros 20 arquivos
		limit := 20
		if len(files) < limit {
			limit = len(files)
		}

		output += "\nPrimeiros arquivos:\n"
		for i := 0; i < limit; i++ {
			output += fmt.Sprintf("  • %s\n", files[i])
		}

		if len(files) > limit {
			output += fmt.Sprintf("\n... e mais %d arquivos\n", len(files)-limit)
		}
	}

	return output
}
