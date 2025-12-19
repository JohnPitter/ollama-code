package agent

import (
	"context"
	"fmt"

	"github.com/johnpitter/ollama-code/internal/intent"
	"github.com/johnpitter/ollama-code/internal/llm"
	"github.com/johnpitter/ollama-code/internal/tools"
)

// handleReadFile processa leitura de arquivo
func (a *Agent) handleReadFile(ctx context.Context, result *intent.DetectionResult) (string, error) {
	filePath, ok := result.Parameters["file_path"].(string)
	if !ok || filePath == "" {
		return "Erro: caminho do arquivo não especificado", nil
	}

	// Executar ferramenta
	toolResult, err := a.toolRegistry.Execute(ctx, "file_reader", map[string]interface{}{
		"file_path": filePath,
	})

	if err != nil || !toolResult.Success {
		return fmt.Sprintf("Erro ao ler arquivo: %s", toolResult.Error), nil
	}

	// Formatar resposta com validação de tipo
	fileType, ok := toolResult.Data["type"].(string)
	if !ok {
		return "Erro: tipo de arquivo inválido", nil
	}

	if fileType == "text" {
		content, ok := toolResult.Data["content"].(string)
		if !ok {
			return "Erro: conteúdo do arquivo em formato inválido", nil
		}
		return fmt.Sprintf("Conteúdo do arquivo %s:\n\n```\n%s\n```", filePath, content), nil
	}

	return fmt.Sprintf("Arquivo %s lido com sucesso (tipo: %s)", filePath, fileType), nil
}

// handleWriteFile processa escrita de arquivo
func (a *Agent) handleWriteFile(ctx context.Context, result *intent.DetectionResult, userMessage string) (string, error) {
	// Verificar se modo permite escritas
	if !a.mode.AllowsWrites() {
		return "❌ Operação bloqueada: modo somente leitura ativo", nil
	}

	// Extrair parâmetros do resultado da detecção ou usar LLM
	filePath, _ := result.Parameters["file_path"].(string)
	content, _ := result.Parameters["content"].(string)
	mode, _ := result.Parameters["mode"].(string)

	// Se parâmetros não vieram da detecção, usar LLM para extrair da mensagem
	if filePath == "" || content == "" {
		llmResponse, err := a.llmClient.Complete(ctx, []llm.Message{
			{
				Role: "system",
				Content: "Você é um assistente que extrai informações de solicitações de escrita de arquivo. " +
					"Responda APENAS com JSON no formato: {\"file_path\": \"caminho\", \"content\": \"conteúdo\", \"mode\": \"create|append|replace\"}",
			},
			{
				Role: "user",
				Content: fmt.Sprintf("Extraia os parâmetros desta solicitação: %s", userMessage),
			},
		}, &llm.CompletionOptions{Temperature: 0.1})

		if err != nil {
			return "Erro ao processar requisição de escrita", err
		}

		// Parse do JSON (simplificado - em produção usar encoding/json)
		// Por enquanto, assumir que veio nos parâmetros ou pedir confirmação com o que temos
		if filePath == "" {
			return fmt.Sprintf("Erro: não consegui identificar o caminho do arquivo. Resposta LLM: %s", llmResponse), nil
		}
	}

	// Validações
	if filePath == "" {
		return "Erro: caminho do arquivo não especificado", nil
	}
	if content == "" && mode != "replace" {
		return "Erro: conteúdo não especificado", nil
	}
	if mode == "" {
		mode = "create" // Padrão
	}

	// Preparar parâmetros para a ferramenta
	params := map[string]interface{}{
		"file_path": filePath,
		"content":   content,
		"mode":      mode,
	}

	// Se for replace, adicionar old_text e new_text
	if mode == "replace" {
		if oldText, ok := result.Parameters["old_text"].(string); ok {
			params["old_text"] = oldText
		}
		if newText, ok := result.Parameters["new_text"].(string); ok {
			params["new_text"] = newText
		}
	}

	// Pedir confirmação se necessário
	if a.mode.RequiresConfirmation() {
		preview := fmt.Sprintf("Arquivo: %s\nModo: %s\nTamanho: %d bytes", filePath, mode, len(content))
		if mode == "create" && len(content) < 500 {
			preview += fmt.Sprintf("\n\nConteúdo:\n%s", content)
		}

		confirmed, err := a.confirmManager.ConfirmWithPreview(
			"Escrever arquivo",
			preview,
		)

		if err != nil || !confirmed {
			return "✗ Operação cancelada pelo usuário", nil
		}
	}

	// Executar ferramenta
	toolResult, err := a.toolRegistry.Execute(ctx, "file_writer", params)

	if err != nil || !toolResult.Success {
		return fmt.Sprintf("Erro ao escrever arquivo: %s", toolResult.Error), nil
	}

	// Formatar resposta
	return fmt.Sprintf("✓ %s", toolResult.Message), nil
}

// handleExecuteCommand processa execução de comando
func (a *Agent) handleExecuteCommand(ctx context.Context, result *intent.DetectionResult) (string, error) {
	// Verificar se modo permite
	if !a.mode.AllowsWrites() {
		return "❌ Operação bloqueada: modo somente leitura ativo", nil
	}

	command, ok := result.Parameters["command"].(string)
	if !ok || command == "" {
		return "Erro: comando não especificado", nil
	}

	// Verificar se é perigoso
	cmdTool, err := a.toolRegistry.Get("command_executor")
	if err != nil {
		return "Erro interno: ferramenta command_executor não encontrada", nil
	}
	cmdExecutor, ok := cmdTool.(*tools.CommandExecutor)
	if !ok {
		return "Erro interno: tipo de ferramenta inválido", nil
	}
	if cmdExecutor.IsDangerous(command) {
		if a.mode.RequiresConfirmation() {
			confirmed, err := a.confirmManager.ConfirmDangerousAction(
				"Executar comando perigoso",
				fmt.Sprintf("Comando: %s\n\nEste comando pode ser destrutivo!", command),
			)

			if err != nil || !confirmed {
				return "✗ Comando cancelado por segurança", nil
			}
		}
	} else if a.mode.RequiresConfirmation() {
		confirmed, err := a.confirmManager.Confirm(
			"Executar comando",
			fmt.Sprintf("Comando: %s", command),
		)

		if err != nil || !confirmed {
			return "✗ Operação cancelada", nil
		}
	}

	// Executar
	toolResult, err := a.toolRegistry.Execute(ctx, "command_executor", map[string]interface{}{
		"command": command,
	})

	if err != nil || !toolResult.Success {
		return fmt.Sprintf("Erro ao executar comando: %s", toolResult.Error), nil
	}

	// Validar tipo dos resultados
	stdout, ok := toolResult.Data["stdout"].(string)
	if !ok {
		stdout = ""
	}
	stderr, ok := toolResult.Data["stderr"].(string)
	if !ok {
		stderr = ""
	}
	exitCode, ok := toolResult.Data["exit_code"].(int)
	if !ok {
		exitCode = -1
	}

	response := fmt.Sprintf("Comando executado (exit code: %d)\n\n", exitCode)
	if stdout != "" {
		response += fmt.Sprintf("Output:\n%s\n", stdout)
	}
	if stderr != "" {
		response += fmt.Sprintf("Errors:\n%s\n", stderr)
	}

	return response, nil
}

// handleSearchCode processa busca de código
func (a *Agent) handleSearchCode(ctx context.Context, result *intent.DetectionResult) (string, error) {
	query, ok := result.Parameters["query"].(string)
	if !ok || query == "" {
		return "Erro: termo de busca não especificado", nil
	}

	toolResult, err := a.toolRegistry.Execute(ctx, "code_searcher", map[string]interface{}{
		"query": query,
	})

	if err != nil || !toolResult.Success {
		return fmt.Sprintf("Erro ao buscar código: %s", toolResult.Error), nil
	}

	count, ok := toolResult.Data["count"].(int)
	if !ok {
		count = 0
	}
	return fmt.Sprintf("Encontrados %d resultados para '%s'", count, query), nil
}

// handleAnalyzeProject processa análise de projeto
func (a *Agent) handleAnalyzeProject(ctx context.Context, result *intent.DetectionResult) (string, error) {
	toolResult, err := a.toolRegistry.Execute(ctx, "project_analyzer", map[string]interface{}{
		"type": "structure",
	})

	if err != nil || !toolResult.Success {
		return fmt.Sprintf("Erro ao analisar projeto: %s", toolResult.Error), nil
	}

	return "Estrutura do projeto analisada com sucesso", nil
}

// handleGitOperation processa operação git
func (a *Agent) handleGitOperation(ctx context.Context, result *intent.DetectionResult) (string, error) {
	if !a.mode.AllowsWrites() {
		return "❌ Operação bloqueada: modo somente leitura ativo", nil
	}

	operation, ok := result.Parameters["operation"].(string)
	if !ok {
		operation = "status"
	}

	// Confirmação para operações destrutivas
	if operation != "status" && operation != "diff" && operation != "log" {
		if a.mode.RequiresConfirmation() {
			confirmed, err := a.confirmManager.Confirm(
				fmt.Sprintf("Operação git: %s", operation),
				"",
			)

			if err != nil || !confirmed {
				return "✗ Operação cancelada", nil
			}
		}
	}

	toolResult, err := a.toolRegistry.Execute(ctx, "git_operations", result.Parameters)

	if err != nil || !toolResult.Success {
		return fmt.Sprintf("Erro na operação git: %s", toolResult.Error), nil
	}

	return fmt.Sprintf("Operação git '%s' executada com sucesso", operation), nil
}

// handleWebSearch processa pesquisa web
func (a *Agent) handleWebSearch(ctx context.Context, result *intent.DetectionResult, userMessage string) (string, error) {
	query, ok := result.Parameters["query"].(string)
	if !ok || query == "" {
		// Fallback: usar a mensagem do usuário como query
		query = userMessage
	}

	a.colorBlue.Printf("🌐 Pesquisando na web: %s\n", query)

	results, err := a.webSearch.Search(ctx, query, []string{"duckduckgo"})
	if err != nil {
		return fmt.Sprintf("Erro ao pesquisar: %v", err), nil
	}

	if len(results) == 0 {
		return "Nenhum resultado encontrado na web.", nil
	}

	// Formatar resultados para o LLM sintetizar
	resultsText := fmt.Sprintf("Encontrei %d resultados para '%s':\n\n", len(results), query)
	for i, r := range results {
		if i >= 3 { // Limitar a 3 resultados
			break
		}
		resultsText += fmt.Sprintf("%d. %s\n   %s\n   URL: %s\n\n", i+1, r.Title, r.Snippet, r.URL)
	}

	// Usar LLM para sintetizar a resposta baseada nos resultados
	prompt := fmt.Sprintf(`Com base nos resultados da pesquisa abaixo, responda à pergunta do usuário: "%s"

%s

Forneça uma resposta clara e concisa baseada nas informações encontradas. Se relevante, cite as fontes.`, userMessage, resultsText)

	messages := []llm.Message{
		{Role: "user", Content: prompt},
	}

	response, err := a.llmClient.Complete(ctx, messages, &llm.CompletionOptions{
		Temperature: 0.7,
		MaxTokens:   1000,
	})
	if err != nil {
		// Se falhar, retornar apenas os resultados formatados
		return resultsText, nil
	}

	return response, nil
}

// handleQuestion processa pergunta simples
func (a *Agent) handleQuestion(ctx context.Context, userMessage string) (string, error) {
	// Usar LLM para responder
	messages := append(a.GetHistory(), llm.Message{
		Role:    "user",
		Content: userMessage,
	})

	// Mostrar header antes de começar streaming
	a.colorGreen.Println("\n🤖 Assistente:")

	response, err := a.llmClient.CompleteStreaming(ctx, messages, &llm.CompletionOptions{
		Temperature: 0.7,
		MaxTokens:   2000,
	}, func(chunk string) {
		fmt.Print(chunk)
	})

	// Adicionar newline após streaming
	fmt.Println()

	if err != nil {
		return "", fmt.Errorf("llm completion: %w", err)
	}

	return response, nil
}
