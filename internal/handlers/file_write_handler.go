package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/johnpitter/ollama-code/internal/intent"
	"github.com/johnpitter/ollama-code/internal/validators"
)

// FileWriteHandler processa escrita de arquivos
type FileWriteHandler struct {
	BaseHandler
	fileValidator *validators.FileValidator
	jsonValidator *validators.JSONValidator
	codeCleaner   *validators.CodeCleaner
}

// NewFileWriteHandler cria novo handler
func NewFileWriteHandler() *FileWriteHandler {
	return &FileWriteHandler{
		BaseHandler:   NewBaseHandler("file_write"),
		fileValidator: validators.NewFileValidator(),
		jsonValidator: validators.NewJSONValidator(),
		codeCleaner:   validators.NewCodeCleaner(),
	}
}

// Handle processa intent de escrita
func (h *FileWriteHandler) Handle(ctx context.Context, deps *Dependencies, result *intent.DetectionResult) (string, error) {
	// 🔒 VERIFICAR MODO READ-ONLY PRIMEIRO
	if !deps.Mode.AllowsWrites() {
		return "❌ Operação bloqueada: modo somente leitura (read-only)\n" +
			"Para permitir modificações, use:\n" +
			"  --mode interactive  (pede confirmação)\n" +
			"  --mode autonomous   (executa automaticamente)", nil
	}

	// Extrair parâmetros
	filePath, _ := result.Parameters["file_path"].(string)
	content, _ := result.Parameters["content"].(string)
	userMessage := result.UserMessage

	// 📦 DETECTAR REQUISIÇÃO DE MÚLTIPLOS ARQUIVOS
	if h.detectMultiFileRequest(userMessage) {
		return h.handleMultiFileWrite(ctx, deps, userMessage)
	}

	// Se não tem conteúdo direto, precisar gerar via LLM
	if content == "" {
		return h.generateAndWrite(ctx, deps, userMessage, filePath, result)
	}

	// Validar filename
	if !h.fileValidator.IsValid(filePath) {
		return "", fmt.Errorf("nome de arquivo inválido: %s", filePath)
	}

	// Limpar content (remover markdown, etc)
	content = h.codeCleaner.Clean(content, filePath)

	// 📝 Criar TODO para tracking
	var todoID string
	if deps.TodoManager != nil {
		id, err := deps.TodoManager.Add(
			fmt.Sprintf("Escrevendo arquivo: %s", filePath),
			fmt.Sprintf("Escrevendo %s", filePath),
		)
		if err == nil {
			todoID = id
		}
	}

	// Confirmar com usuário se necessário
	if deps.Mode.RequiresConfirmation() {
		preview := content

		// 🎨 Se o arquivo existe e temos DiffManager, mostrar diff colorizado
		if deps.DiffManager != nil && deps.PreviewManager != nil {
			// Tentar ler arquivo existente
			readParams := map[string]interface{}{
				"file_path": filePath,
			}
			readResult, readErr := deps.ToolRegistry.Execute(ctx, "file_reader", readParams)

			if readErr == nil && readResult.Success {
				oldContent, ok := readResult.Data["content"].(string)
				if ok && oldContent != "" {
					// Computar diff
					diffResult := deps.DiffManager.ComputeDiff(filePath, oldContent, content)
					if diffResult != nil {
						// Gerar preview colorizado
						preview = deps.PreviewManager.Preview(diffResult)
					}
				}
			}
		}

		// Fallback para preview simples
		if len(preview) > 500 && !strings.Contains(preview, "📄 Arquivo:") {
			preview = preview[:500] + "\n...(truncated)"
		}

		confirmed, err := deps.ConfirmManager.ConfirmWithPreview(
			fmt.Sprintf("Escrever arquivo %s?", filePath),
			preview,
		)
		if err != nil || !confirmed {
			return "Operação cancelada", nil
		}
	}

	// Executar escrita via tool
	params := map[string]interface{}{
		"file_path": filePath,
		"content":   content,
	}

	toolResult, err := deps.ToolRegistry.Execute(ctx, "file_writer", params)
	if err != nil {
		// ❌ Marcar TODO como failed (se houver)
		if todoID != "" && deps.TodoManager != nil {
			deps.TodoManager.Delete(todoID)
		}
		return "", fmt.Errorf("erro ao escrever arquivo: %w", err)
	}

	if !toolResult.Success {
		// ❌ Marcar TODO como failed
		if todoID != "" && deps.TodoManager != nil {
			deps.TodoManager.Delete(todoID)
		}
		return "", fmt.Errorf("erro: %s", toolResult.Error)
	}

	// ✅ Completar TODO
	if todoID != "" && deps.TodoManager != nil {
		deps.TodoManager.Complete(todoID)
	}

	// Adicionar aos arquivos recentes
	deps.RecentFiles = append(deps.RecentFiles, filePath)

	return toolResult.Message, nil
}

// generateAndWrite gera conteúdo via LLM e escreve
func (h *FileWriteHandler) generateAndWrite(ctx context.Context, deps *Dependencies, userMessage, suggestedPath string, result *intent.DetectionResult) (string, error) {
	// Construir prompt para geração
	prompt := h.buildGenerationPrompt(userMessage, suggestedPath, deps)

	// Completar com LLM
	response, err := deps.LLMClient.Complete(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("erro ao gerar conteúdo: %w", err)
	}

	// Tentar extrair JSON da resposta
	parsed, err := h.jsonValidator.Parse(response)
	if err != nil {
		// Se não conseguiu fazer parse como JSON, tratar como conteúdo direto
		return h.writeDirectContent(ctx, deps, response, suggestedPath)
	}

	// Extrair file_path e content do JSON
	filePath, ok := parsed["file_path"].(string)
	if !ok || filePath == "" {
		filePath = suggestedPath
		if filePath == "" {
			// Tentar inferir do userMessage
			filePath = h.fileValidator.ExtractFilename(userMessage)
			if filePath == "" {
				return "", fmt.Errorf("não foi possível determinar o caminho do arquivo")
			}
		}
	}

	content, ok := parsed["content"].(string)
	if !ok || content == "" {
		// Se não tem content no JSON, usar a resposta completa
		content = response
	}

	// Recursivamente chamar Handle com parâmetros completos
	newResult := &intent.DetectionResult{
		Intent:      intent.IntentWriteFile,
		UserMessage: userMessage,
		Parameters: map[string]interface{}{
			"file_path": filePath,
			"content":   content,
		},
	}

	return h.Handle(ctx, deps, newResult)
}

// buildGenerationPrompt constrói prompt para geração de conteúdo
func (h *FileWriteHandler) buildGenerationPrompt(userMessage, suggestedPath string, deps *Dependencies) string {
	var prompt strings.Builder

	prompt.WriteString("Generate file content based on the following request:\n\n")
	prompt.WriteString(fmt.Sprintf("User request: %s\n\n", userMessage))

	if suggestedPath != "" {
		prompt.WriteString(fmt.Sprintf("Suggested file path: %s\n\n", suggestedPath))
	}

	prompt.WriteString("Working directory: " + deps.WorkDir + "\n\n")

	// Adicionar contexto de arquivos recentes se houver
	if len(deps.RecentFiles) > 0 {
		prompt.WriteString("Recent files:\n")
		for _, file := range deps.RecentFiles {
			if len(deps.RecentFiles) > 5 {
				// Limitar a 5 mais recentes
				break
			}
			prompt.WriteString(fmt.Sprintf("- %s\n", file))
		}
		prompt.WriteString("\n")
	}

	prompt.WriteString("Output a JSON object with 'file_path' and 'content' fields.\n")
	prompt.WriteString("Example:\n")
	prompt.WriteString("{\n")
	prompt.WriteString("  \"file_path\": \"example.go\",\n")
	prompt.WriteString("  \"content\": \"package main\\n\\nfunc main() {\\n  // code here\\n}\"\n")
	prompt.WriteString("}\n")

	return prompt.String()
}

// writeDirectContent escreve conteúdo direto (fallback quando não é JSON)
func (h *FileWriteHandler) writeDirectContent(ctx context.Context, deps *Dependencies, content, suggestedPath string) (string, error) {
	filePath := suggestedPath
	if filePath == "" {
		return "", fmt.Errorf("caminho do arquivo não especificado")
	}

	// Limpar conteúdo
	content = h.codeCleaner.Clean(content, filePath)

	// Recursivamente chamar Handle
	newResult := &intent.DetectionResult{
		Intent: intent.IntentWriteFile,
		Parameters: map[string]interface{}{
			"file_path": filePath,
			"content":   content,
		},
	}

	return h.Handle(ctx, deps, newResult)
}

// detectMultiFileRequest detecta se a mensagem solicita múltiplos arquivos
func (h *FileWriteHandler) detectMultiFileRequest(message string) bool {
	msgLower := strings.ToLower(message)

	multiFileKeywords := []string{
		"separados", "separadas",
		"múltiplos arquivos", "multiplos arquivos",
		"vários arquivos", "varios arquivos",
		"html, css e javascript", "html, css e js",
		"html e css separados", "html e css separadas",
		"html, css", "css, js", "html, js",
		"arquivo html e css", "arquivo css e js",
		"com estrutura de pastas",
		"projeto completo",
		"full-stack",
		"frontend e backend",
		"cliente e servidor",
		"3 arquivos", "três arquivos",
		"multiple files", "separate files",
	}

	for _, keyword := range multiFileKeywords {
		if strings.Contains(msgLower, keyword) {
			return true
		}
	}

	return false
}

// handleMultiFileWrite cria múltiplos arquivos coordenados
func (h *FileWriteHandler) handleMultiFileWrite(ctx context.Context, deps *Dependencies, userMessage string) (string, error) {
	// Construir prompt específico para multi-file
	prompt := h.buildMultiFilePrompt(userMessage, deps)

	// Completar com LLM
	response, err := deps.LLMClient.Complete(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("erro ao gerar múltiplos arquivos: %w", err)
	}

	// Parse JSON response
	parsed, err := h.jsonValidator.Parse(response)
	if err != nil {
		// Fallback: tentar como arquivo único
		return h.generateAndWrite(ctx, deps, userMessage, "", &intent.DetectionResult{
			Intent:      intent.IntentWriteFile,
			UserMessage: userMessage,
			Parameters:  map[string]interface{}{},
		})
	}

	// Extrair array de arquivos
	filesRaw, ok := parsed["files"]
	if !ok {
		// Fallback: tentar como arquivo único
		return h.generateAndWrite(ctx, deps, userMessage, "", &intent.DetectionResult{
			Intent:      intent.IntentWriteFile,
			UserMessage: userMessage,
			Parameters:  map[string]interface{}{},
		})
	}

	filesArray, ok := filesRaw.([]interface{})
	if !ok || len(filesArray) == 0 {
		return "", fmt.Errorf("formato de resposta inválido: esperado array de arquivos")
	}

	// Confirmar com usuário se necessário (UMA VEZ para todo o projeto)
	if deps.Mode.RequiresConfirmation() {
		fileList := make([]string, 0, len(filesArray))
		for _, fileRaw := range filesArray {
			fileMap, ok := fileRaw.(map[string]interface{})
			if !ok {
				continue
			}
			filePath, _ := fileMap["file_path"].(string)
			if filePath != "" {
				fileList = append(fileList, filePath)
			}
		}

		confirmed, err := deps.ConfirmManager.Confirm(
			fmt.Sprintf("Criar %d arquivo(s)?\n  - %s",
				len(fileList),
				strings.Join(fileList, "\n  - ")),
		)
		if err != nil || !confirmed {
			return "Operação cancelada", nil
		}
	}

	// Criar cada arquivo
	var created []string
	var failed []string

	for _, fileRaw := range filesArray {
		fileMap, ok := fileRaw.(map[string]interface{})
		if !ok {
			failed = append(failed, "arquivo com formato inválido")
			continue
		}

		filePath, _ := fileMap["file_path"].(string)
		content, _ := fileMap["content"].(string)

		if filePath == "" || content == "" {
			failed = append(failed, fmt.Sprintf("%s (falta file_path ou content)", filePath))
			continue
		}

		// Validar filename
		if !h.fileValidator.IsValid(filePath) {
			failed = append(failed, fmt.Sprintf("%s (nome inválido)", filePath))
			continue
		}

		// Limpar content
		content = h.codeCleaner.Clean(content, filePath)

		// Executar escrita via tool
		params := map[string]interface{}{
			"file_path": filePath,
			"content":   content,
		}

		toolResult, err := deps.ToolRegistry.Execute(ctx, "file_writer", params)
		if err != nil {
			failed = append(failed, fmt.Sprintf("%s (erro: %v)", filePath, err))
			continue
		}

		if !toolResult.Success {
			failed = append(failed, fmt.Sprintf("%s (erro: %s)", filePath, toolResult.Error))
			continue
		}

		// Adicionar aos arquivos recentes
		deps.RecentFiles = append(deps.RecentFiles, filePath)
		created = append(created, filePath)
	}

	// Construir mensagem de resultado
	var result strings.Builder
	result.WriteString(fmt.Sprintf("✓ Projeto multi-file criado!\n\n"))

	if len(created) > 0 {
		result.WriteString(fmt.Sprintf("Arquivos criados (%d):\n", len(created)))
		for _, file := range created {
			result.WriteString(fmt.Sprintf("  ✓ %s\n", file))
		}
	}

	if len(failed) > 0 {
		result.WriteString(fmt.Sprintf("\nFalhas (%d):\n", len(failed)))
		for _, file := range failed {
			result.WriteString(fmt.Sprintf("  ✗ %s\n", file))
		}
	}

	if len(created) == 0 {
		return "", fmt.Errorf("nenhum arquivo foi criado")
	}

	return result.String(), nil
}

// buildMultiFilePrompt constrói prompt específico para geração de múltiplos arquivos
func (h *FileWriteHandler) buildMultiFilePrompt(userMessage string, deps *Dependencies) string {
	var prompt strings.Builder

	prompt.WriteString("Generate multiple coordinated files based on the following request:\n\n")
	prompt.WriteString(fmt.Sprintf("User request: %s\n\n", userMessage))
	prompt.WriteString("Working directory: " + deps.WorkDir + "\n\n")

	// Adicionar contexto de arquivos recentes se houver
	if len(deps.RecentFiles) > 0 {
		prompt.WriteString("Recent files:\n")
		count := 0
		for i := len(deps.RecentFiles) - 1; i >= 0 && count < 5; i-- {
			prompt.WriteString(fmt.Sprintf("- %s\n", deps.RecentFiles[i]))
			count++
		}
		prompt.WriteString("\n")
	}

	prompt.WriteString("Output a JSON object with a 'files' array. Each file must have 'file_path' and 'content'.\n\n")
	prompt.WriteString("IMPORTANT RULES:\n")
	prompt.WriteString("1. Create ALL files requested by the user\n")
	prompt.WriteString("2. If user asks for 'HTML, CSS and JavaScript separated': create 3 files\n")
	prompt.WriteString("3. HTML must reference CSS with <link rel=\"stylesheet\" href=\"...\">\n")
	prompt.WriteString("4. HTML must reference JS with <script src=\"...\"></script>\n")
	prompt.WriteString("5. Use appropriate file names (e.g., index.html, style.css, script.js)\n")
	prompt.WriteString("6. Each file must have COMPLETE and functional content\n")
	prompt.WriteString("7. Files must be correctly linked to each other\n")
	prompt.WriteString("8. Use relative paths for linking\n\n")

	prompt.WriteString("Example output:\n")
	prompt.WriteString("{\n")
	prompt.WriteString("  \"files\": [\n")
	prompt.WriteString("    {\n")
	prompt.WriteString("      \"file_path\": \"index.html\",\n")
	prompt.WriteString("      \"content\": \"<!DOCTYPE html>\\n<html>\\n<head>\\n  <link rel=\\\"stylesheet\\\" href=\\\"style.css\\\">\\n</head>\\n<body>\\n  <h1>Hello</h1>\\n  <script src=\\\"script.js\\\"></script>\\n</body>\\n</html>\"\n")
	prompt.WriteString("    },\n")
	prompt.WriteString("    {\n")
	prompt.WriteString("      \"file_path\": \"style.css\",\n")
	prompt.WriteString("      \"content\": \"body { font-family: Arial; }\"\n")
	prompt.WriteString("    },\n")
	prompt.WriteString("    {\n")
	prompt.WriteString("      \"file_path\": \"script.js\",\n")
	prompt.WriteString("      \"content\": \"console.log('Hello');\"\n")
	prompt.WriteString("    }\n")
	prompt.WriteString("  ]\n")
	prompt.WriteString("}\n\n")

	prompt.WriteString("Now generate the files:\n")

	return prompt.String()
}
