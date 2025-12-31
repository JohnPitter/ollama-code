package handlers

import (
	"context"
	"fmt"
	"strings"

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

	// Verificar se usuário pediu análise do arquivo
	userAskedForAnalysis := h.detectAnalysisIntent(result.UserMessage)

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

	// Formatar resposta com conteúdo
	return h.formatReadResult(ctx, deps, toolResult, filePath, userAskedForAnalysis), nil
}

// detectAnalysisIntent detecta se usuário pediu análise do arquivo
func (h *FileReadHandler) detectAnalysisIntent(message string) bool {
	messageStr := fmt.Sprintf("%v", message)
	messageLower := strings.ToLower(messageStr)

	analysisKeywords := []string{
		"analise", "analisa", "analyze", "analyse",
		"explique", "explica", "explain",
		"revise", "revisa", "review",
		"o que faz", "what does",
	}

	for _, keyword := range analysisKeywords {
		if strings.Contains(messageLower, keyword) {
			return true
		}
	}
	return false
}

// formatReadResult formata resultado da leitura com conteúdo
func (h *FileReadHandler) formatReadResult(ctx context.Context, deps *Dependencies, result ToolResult, filePath string, needsAnalysis bool) string {
	output := result.Message + "\n\n"

	// Obter conteúdo do arquivo
	content, hasContent := result.Data["content"].(string)
	fileType, _ := result.Data["type"].(string)

	if !hasContent || fileType == "image" {
		// Para imagens, apenas indicar que foi lida
		output += "📷 Arquivo de imagem carregado\n"
		return output
	}

	// Se usuário pediu análise, usar LLM para analisar
	if needsAnalysis && deps.LLMClient != nil {
		output += "📝 Análise do arquivo:\n\n"

		analysisPrompt := fmt.Sprintf(`Você é um assistente de código. Analise o seguinte arquivo e forneça:
1. Resumo do que o arquivo faz
2. Principais componentes/funções/classes
3. Tecnologias/linguagens usadas
4. Observações importantes

Arquivo: %s
Conteúdo:
%s

Forneça uma análise concisa e útil.`, filePath, content)

		response, err := deps.LLMClient.Complete(ctx, analysisPrompt)
		if err == nil {
			output += response + "\n"
		} else {
			// Fallback: mostrar preview se análise falhar
			output += h.formatContentPreview(content)
		}
	} else {
		// Mostrar preview do conteúdo
		output += h.formatContentPreview(content)
	}

	return output
}

// formatContentPreview formata preview do conteúdo (primeiras 20 linhas)
func (h *FileReadHandler) formatContentPreview(content string) string {
	lines := splitLines(content)

	preview := "📄 Conteúdo do arquivo:\n\n"
	preview += "```\n"

	// Mostrar até 20 linhas
	limit := 20
	if len(lines) < limit {
		limit = len(lines)
	}

	for i := 0; i < limit; i++ {
		preview += fmt.Sprintf("%4d | %s\n", i+1, lines[i])
	}

	if len(lines) > limit {
		preview += fmt.Sprintf("\n... e mais %d linhas\n", len(lines)-limit)
	}

	preview += "```\n"
	preview += fmt.Sprintf("\nTotal: %d linhas\n", len(lines))

	return preview
}

// splitLines divide string em linhas
func splitLines(s string) []string {
	if s == "" {
		return []string{}
	}

	lines := []string{}
	current := ""

	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, current)
			current = ""
		} else if s[i] != '\r' {
			current += string(s[i])
		}
	}

	if current != "" {
		lines = append(lines, current)
	}

	return lines
}
