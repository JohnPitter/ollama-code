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
	if !ok {
		return "Erro: caminho do arquivo não especificado", nil
	}

	// Executar ferramenta
	toolResult, err := a.toolRegistry.Execute(ctx, "file_reader", map[string]interface{}{
		"file_path": filePath,
	})

	if err != nil || !toolResult.Success {
		return fmt.Sprintf("Erro ao ler arquivo: %s", toolResult.Error), nil
	}

	// Formatar resposta
	if toolResult.Data["type"] == "text" {
		content := toolResult.Data["content"].(string)
		return fmt.Sprintf("Conteúdo do arquivo %s:\n\n```\n%s\n```", filePath, content), nil
	}

	return fmt.Sprintf("Arquivo %s lido com sucesso (tipo: %s)", filePath, toolResult.Data["type"]), nil
}

// handleWriteFile processa escrita de arquivo
func (a *Agent) handleWriteFile(ctx context.Context, result *intent.DetectionResult, userMessage string) (string, error) {
	// Verificar se modo permite escritas
	if !a.mode.AllowsWrites() {
		return "❌ Operação bloqueada: modo somente leitura ativo", nil
	}

	// Obter parâmetros do LLM
	_, err := a.llmClient.Complete(ctx, []llm.Message{
		{
			Role: "user",
			Content: fmt.Sprintf(`Extraia os seguintes parâmetros desta mensagem e retorne em formato JSON:
- file_path: caminho do arquivo
- content: conteúdo a escrever
- mode: "create", "append" ou "replace"

Mensagem: "%s"

Responda APENAS com JSON.`, userMessage),
		},
	}, &llm.CompletionOptions{Temperature: 0.1})

	if err != nil {
		return "Erro ao processar requisição de escrita", err
	}

	// Pedir confirmação se necessário
	if a.mode.RequiresConfirmation() {
		confirmed, err := a.confirmManager.ConfirmWithPreview(
			"Escrever arquivo",
			fmt.Sprintf("Arquivo: %v\nModo: %v", result.Parameters["file_path"], result.Parameters["mode"]),
		)

		if err != nil || !confirmed {
			return "✗ Operação cancelada pelo usuário", nil
		}
	}

	// Executar (simplificado - deveria fazer parse do JSON)
	return "Funcionalidade de escrita de arquivo em desenvolvimento", nil
}

// handleExecuteCommand processa execução de comando
func (a *Agent) handleExecuteCommand(ctx context.Context, result *intent.DetectionResult) (string, error) {
	// Verificar se modo permite
	if !a.mode.AllowsWrites() {
		return "❌ Operação bloqueada: modo somente leitura ativo", nil
	}

	command, ok := result.Parameters["command"].(string)
	if !ok {
		return "Erro: comando não especificado", nil
	}

	// Verificar se é perigoso
	cmdTool, _ := a.toolRegistry.Get("command_executor")
	cmdExecutor, ok := cmdTool.(*tools.CommandExecutor)
	if !ok {
		return "Erro interno: ferramenta não encontrada", nil
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

	stdout := toolResult.Data["stdout"].(string)
	stderr := toolResult.Data["stderr"].(string)
	exitCode := toolResult.Data["exit_code"].(int)

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
	if !ok {
		return "Erro: termo de busca não especificado", nil
	}

	toolResult, err := a.toolRegistry.Execute(ctx, "code_searcher", map[string]interface{}{
		"query": query,
	})

	if err != nil || !toolResult.Success {
		return fmt.Sprintf("Erro ao buscar código: %s", toolResult.Error), nil
	}

	count := toolResult.Data["count"].(int)
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
func (a *Agent) handleWebSearch(ctx context.Context, result *intent.DetectionResult) (string, error) {
	query, ok := result.Parameters["query"].(string)
	if !ok {
		return "Erro: termo de busca não especificado", nil
	}

	a.colorBlue.Printf("🌐 Pesquisando na web: %s\n", query)

	results, err := a.webSearch.Search(ctx, query, []string{"duckduckgo"})
	if err != nil {
		return fmt.Sprintf("Erro ao pesquisar: %v", err), nil
	}

	return fmt.Sprintf("Encontrados %d resultados para '%s'", len(results), query), nil
}

// handleQuestion processa pergunta simples
func (a *Agent) handleQuestion(ctx context.Context, userMessage string) (string, error) {
	// Usar LLM para responder
	messages := append(a.GetHistory(), llm.Message{
		Role:    "user",
		Content: userMessage,
	})

	response, err := a.llmClient.CompleteStreaming(ctx, messages, &llm.CompletionOptions{
		Temperature: 0.7,
		MaxTokens:   2000,
	}, func(chunk string) {
		fmt.Print(chunk)
	})

	if err != nil {
		return "", fmt.Errorf("llm completion: %w", err)
	}

	return response, nil
}
