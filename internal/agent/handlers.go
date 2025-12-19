package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/johnpitter/ollama-code/internal/intent"
	"github.com/johnpitter/ollama-code/internal/llm"
	"github.com/johnpitter/ollama-code/internal/tools"
	"github.com/johnpitter/ollama-code/internal/websearch"
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

	// Extrair parâmetros do resultado da detecção
	filePath, _ := result.Parameters["file_path"].(string)
	content, _ := result.Parameters["content"].(string)
	mode, _ := result.Parameters["mode"].(string)

	// Detectar se é uma correção de arquivo recente
	recentlyModified := a.GetRecentlyModifiedFiles()
	isBugFix := detectBugReport(userMessage)

	if isBugFix && len(recentlyModified) > 0 {
		// Usuário reportou problema em arquivo recente
		return a.handleBugFix(ctx, userMessage, recentlyModified[0])
	}

	// Se conteúdo não foi especificado, significa que o usuário quer que geremos
	if content == "" {
		a.colorBlue.Println("💭 Gerando conteúdo...")

		// Usar LLM para gerar o conteúdo baseado na descrição do usuário
		generationPrompt := fmt.Sprintf(`Você é um assistente de programação. O usuário pediu:

"%s"

TAREFA:
1. Identifique o tipo de arquivo que o usuário quer criar
2. Identifique o nome/caminho do arquivo (se não especificado, sugira um apropriado)
3. Gere o conteúdo completo do arquivo conforme solicitado

Responda APENAS com um JSON no seguinte formato:
{
  "file_path": "caminho/do/arquivo.ext",
  "content": "conteúdo completo do arquivo aqui",
  "mode": "create"
}

IMPORTANTE:
- O campo "content" deve conter TODO o código/conteúdo solicitado
- Use boas práticas de código
- Adicione comentários quando apropriado
- Se for HTML/CSS, crie algo visualmente atraente
- Não inclua explicações fora do JSON`, userMessage)

		llmResponse, err := a.llmClient.Complete(ctx, []llm.Message{
			{Role: "user", Content: generationPrompt},
		}, &llm.CompletionOptions{Temperature: 0.7, MaxTokens: 3000})

		if err != nil {
			return "Erro ao gerar conteúdo", err
		}

		// Extrair JSON da resposta (LLM pode retornar com ```json ou direto)
		jsonStr := strings.TrimSpace(llmResponse)
		jsonStr = strings.TrimPrefix(jsonStr, "```json")
		jsonStr = strings.TrimPrefix(jsonStr, "```")
		jsonStr = strings.TrimSuffix(jsonStr, "```")
		jsonStr = strings.TrimSpace(jsonStr)

		// Parse do JSON
		var parsed map[string]interface{}
		if err := parseJSON(jsonStr, &parsed); err != nil {
			// Fallback: tentar usar a resposta diretamente como conteúdo
			a.colorYellow.Printf("⚠️  Não foi possível fazer parse do JSON, tentando abordagem alternativa...\n")

			// Se não parseou, tenta gerar novamente de forma mais simples
			return a.generateAndWriteFileSimple(ctx, userMessage)
		}

		// Extrair campos do JSON
		if fp, ok := parsed["file_path"].(string); ok && fp != "" {
			filePath = fp
		}
		if c, ok := parsed["content"].(string); ok && c != "" {
			content = c
		}
		if m, ok := parsed["mode"].(string); ok && m != "" {
			mode = m
		}
	}

	// Validações finais
	if filePath == "" {
		return "Erro: não foi possível determinar o caminho do arquivo", nil
	}
	if content == "" && mode != "replace" {
		return "Erro: não foi possível gerar o conteúdo solicitado", nil
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

	// Registrar arquivo como recentemente modificado
	a.AddRecentFile(filePath)

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

	a.colorBlue.Printf("🔍 Buscando por: %s\n", query)

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

	if count == 0 {
		return fmt.Sprintf("Nenhum resultado encontrado para '%s'", query), nil
	}

	// Construir resposta com os resultados
	var response strings.Builder
	response.WriteString(fmt.Sprintf("Encontrados %d resultado(s) para '%s'\n\n", count, query))

	// Mostrar resultados se disponíveis
	if matches, ok := toolResult.Data["matches"].([]interface{}); ok && len(matches) > 0 {
		maxResults := min(len(matches), 10) // Limitar a 10 resultados
		for i := 0; i < maxResults; i++ {
			if match, ok := matches[i].(map[string]interface{}); ok {
				file, _ := match["file"].(string)
				line, _ := match["line"].(int)
				text, _ := match["text"].(string)

				response.WriteString(fmt.Sprintf("📄 %s:%d\n", file, line))
				response.WriteString(fmt.Sprintf("   %s\n\n", strings.TrimSpace(text)))
			}
		}

		if count > 10 {
			response.WriteString(fmt.Sprintf("... e mais %d resultado(s)\n", count-10))
		}
	}

	return response.String(), nil
}

// handleAnalyzeProject processa análise de projeto
func (a *Agent) handleAnalyzeProject(ctx context.Context, result *intent.DetectionResult) (string, error) {
	a.colorBlue.Println("📊 Analisando estrutura do projeto...")

	toolResult, err := a.toolRegistry.Execute(ctx, "project_analyzer", map[string]interface{}{
		"type": "structure",
	})

	if err != nil || !toolResult.Success {
		return fmt.Sprintf("Erro ao analisar projeto: %s", toolResult.Error), nil
	}

	// Construir resposta com informações da análise
	var response strings.Builder
	response.WriteString("📊 Análise da Estrutura do Projeto\n\n")

	// Mostrar informações básicas
	if projectName, ok := toolResult.Data["project_name"].(string); ok {
		response.WriteString(fmt.Sprintf("📦 Projeto: %s\n", projectName))
	}

	if fileCount, ok := toolResult.Data["file_count"].(int); ok {
		response.WriteString(fmt.Sprintf("📄 Arquivos: %d\n", fileCount))
	}

	if dirCount, ok := toolResult.Data["directory_count"].(int); ok {
		response.WriteString(fmt.Sprintf("📁 Diretórios: %d\n", dirCount))
	}

	if languages, ok := toolResult.Data["languages"].([]interface{}); ok && len(languages) > 0 {
		response.WriteString("\n🔤 Linguagens detectadas:\n")
		for _, lang := range languages {
			if langStr, ok := lang.(string); ok {
				response.WriteString(fmt.Sprintf("   • %s\n", langStr))
			}
		}
	}

	if structure, ok := toolResult.Data["structure"].(string); ok && structure != "" {
		response.WriteString(fmt.Sprintf("\n📂 Estrutura:\n%s\n", structure))
	}

	return response.String(), nil
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

	a.colorBlue.Printf("📄 Encontrados %d resultados, buscando conteúdo...\n", len(results))

	// Fazer fetch do conteúdo real dos top 3 resultados
	fetchedContents, err := a.webSearch.FetchContents(ctx, results, 3)
	if err != nil {
		a.colorYellow.Printf("⚠️  Erro ao buscar conteúdo: %v, usando snippets\n", err)
		return a.synthesizeFromSnippets(ctx, userMessage, results)
	}

	// Construir contexto com conteúdo completo
	var contextBuilder strings.Builder
	contextBuilder.WriteString(fmt.Sprintf("Resultados da pesquisa para '%s':\n\n", query))

	validContents := 0
	for i, content := range fetchedContents {
		if content.Error != "" {
			a.colorYellow.Printf("⚠️  Erro ao buscar %s: %s\n", content.URL, content.Error)
			continue
		}
		if content.Content == "" {
			a.colorYellow.Printf("⚠️  Conteúdo vazio de %s\n", content.URL)
			continue
		}
		validContents++
		a.colorGreen.Printf("✓ Conteúdo obtido de %s (%d chars)\n", content.URL, len(content.Content))
		contextBuilder.WriteString(fmt.Sprintf("=== Fonte %d: %s ===\n", i+1, content.Title))
		contextBuilder.WriteString(fmt.Sprintf("URL: %s\n\n", content.URL))
		contextBuilder.WriteString(content.Content)
		contextBuilder.WriteString("\n\n")
	}

	if validContents == 0 {
		a.colorYellow.Printf("⚠️  Nenhum conteúdo válido, usando snippets\n")
		return a.synthesizeFromSnippets(ctx, userMessage, results)
	}

	a.colorGreen.Printf("✓ %d fontes com conteúdo válido\n", validContents)

	// Usar LLM para sintetizar resposta com conteúdo completo
	prompt := fmt.Sprintf(`Você acabou de buscar informações atualizadas na internet. Use SOMENTE as informações dos sites abaixo para responder.

Pergunta: "%s"

%s

IMPORTANTE:
- Use APENAS as informações fornecidas acima
- NÃO diga que não tem acesso à internet ou dados em tempo real
- Você ACABOU de buscar essas informações na web
- Forneça uma resposta direta e objetiva baseada no conteúdo obtido
- Cite as fontes quando relevante`, userMessage, contextBuilder.String())

	a.colorGreen.Println("\n🤖 Assistente:")

	response, err := a.llmClient.CompleteStreaming(ctx, []llm.Message{
		{Role: "user", Content: prompt},
	}, &llm.CompletionOptions{
		Temperature: 0.7,
		MaxTokens:   1500,
	}, func(chunk string) {
		fmt.Print(chunk)
	})

	fmt.Println()

	if err != nil {
		return contextBuilder.String(), nil
	}

	return response, nil
}

// synthesizeFromSnippets sintetiza resposta apenas com snippets (fallback)
func (a *Agent) synthesizeFromSnippets(ctx context.Context, userMessage string, results []websearch.SearchResult) (string, error) {
	a.colorYellow.Println("ℹ️  Usando snippets de pesquisa...")

	resultsText := "Resultados da pesquisa:\n\n"
	validSnippets := 0
	for i, r := range results {
		if i >= 5 {
			break
		}
		if r.Snippet != "" {
			validSnippets++
			resultsText += fmt.Sprintf("%d. %s\n   %s\n   URL: %s\n\n", validSnippets, r.Title, r.Snippet, r.URL)
		}
	}

	if validSnippets == 0 {
		return "Não foi possível obter informações da web no momento. Por favor, tente novamente.", nil
	}

	prompt := fmt.Sprintf(`Você acabou de buscar informações na internet. Use os snippets abaixo para responder.

Pergunta: "%s"

%s

IMPORTANTE:
- Use APENAS as informações dos snippets acima
- NÃO diga que não tem acesso à internet
- Você ACABOU de fazer uma busca web
- Forneça uma resposta direta baseada nos snippets
- Se os snippets não tiverem informação suficiente, diga isso claramente`, userMessage, resultsText)

	a.colorGreen.Println("\n🤖 Assistente:")

	response, err := a.llmClient.CompleteStreaming(ctx, []llm.Message{
		{Role: "user", Content: prompt},
	}, &llm.CompletionOptions{
		Temperature: 0.7,
		MaxTokens:   1000,
	}, func(chunk string) {
		fmt.Print(chunk)
	})

	fmt.Println()

	if err != nil {
		return resultsText, nil
	}

	return response, nil
}

// min retorna o menor de dois inteiros
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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

// parseJSON faz parse de string JSON em um map usando encoding/json
func parseJSON(jsonStr string, result *map[string]interface{}) error {
	err := json.Unmarshal([]byte(jsonStr), result)
	if err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Validar se tem file_path
	if _, ok := (*result)["file_path"]; !ok {
		return fmt.Errorf("JSON missing required field: file_path")
	}

	return nil
}

// generateAndWriteFileSimple método simplificado para gerar e escrever arquivo (fallback)
func (a *Agent) generateAndWriteFileSimple(ctx context.Context, userMessage string) (string, error) {
	a.colorYellow.Println("🔄 Tentando método alternativo de geração...")

	// Prompt mais direto
	prompt := fmt.Sprintf(`O usuário pediu: "%s"

Gere o código/conteúdo completo solicitado.
Comece sua resposta com o nome do arquivo na primeira linha (ex: index.html).
Depois, nas linhas seguintes, coloque todo o conteúdo do arquivo.`, userMessage)

	response, err := a.llmClient.Complete(ctx, []llm.Message{
		{Role: "user", Content: prompt},
	}, &llm.CompletionOptions{Temperature: 0.7, MaxTokens: 3000})

	if err != nil {
		return "Erro ao gerar conteúdo", err
	}

	// Tentar extrair nome do arquivo da primeira linha
	lines := strings.Split(response, "\n")
	if len(lines) < 2 {
		return "Erro: resposta inválida do modelo", nil
	}

	filePath := strings.TrimSpace(lines[0])
	content := strings.Join(lines[1:], "\n")

	// Limpar possíveis marcadores markdown
	filePath = strings.TrimPrefix(filePath, "# ")
	filePath = strings.TrimPrefix(filePath, "Arquivo: ")
	content = strings.TrimPrefix(content, "```html")
	content = strings.TrimPrefix(content, "```css")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	// Validar
	if filePath == "" || content == "" {
		return fmt.Sprintf("Erro: não foi possível gerar arquivo.\nResposta do modelo:\n%s", response), nil
	}

	// Mostrar preview
	preview := fmt.Sprintf("Arquivo: %s\nTamanho: %d bytes\n\nPreview (primeiras linhas):\n%s",
		filePath, len(content), truncate(content, 500))

	a.colorGreen.Printf("\n📄 Conteúdo gerado:\n%s\n\n", preview)

	// Confirmar
	if a.mode.RequiresConfirmation() {
		confirmed, err := a.confirmManager.ConfirmWithPreview(
			"Criar arquivo",
			preview,
		)

		if err != nil || !confirmed {
			return "✗ Operação cancelada pelo usuário", nil
		}
	}

	// Escrever arquivo
	toolResult, err := a.toolRegistry.Execute(ctx, "file_writer", map[string]interface{}{
		"file_path": filePath,
		"content":   content,
		"mode":      "create",
	})

	if err != nil || !toolResult.Success {
		return fmt.Sprintf("Erro ao escrever arquivo: %s", toolResult.Error), nil
	}

	return fmt.Sprintf("✓ %s", toolResult.Message), nil
}

// truncate trunca string para tamanho máximo
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// detectBugReport detecta se usuário está reportando um problema/bug
func detectBugReport(message string) bool {
	msgLower := strings.ToLower(message)

	// Palavras-chave que indicam problema/bug
	bugKeywords := []string{
		"não funcionou", "nao funcionou",
		"não funciona", "nao funciona",
		"erro", "error",
		"bug", "problema",
		"quebrou", "quebrado",
		"falhou", "falha",
		"deu errado",
		"não apareceu", "nao apareceu",
		"não aparece", "nao aparece",
		"conserta", "corrija", "corrige",
		"arruma", "ajusta",
	}

	for _, keyword := range bugKeywords {
		if strings.Contains(msgLower, keyword) {
			return true
		}
	}

	return false
}

// handleBugFix lida com correção de bugs em arquivo existente
func (a *Agent) handleBugFix(ctx context.Context, userMessage, filePath string) (string, error) {
	a.colorYellow.Printf("🔧 Detectado problema em arquivo recente: %s\n", filePath)
	a.colorBlue.Println("📖 Lendo arquivo atual...")

	// Ler arquivo atual
	toolResult, err := a.toolRegistry.Execute(ctx, "file_reader", map[string]interface{}{
		"file_path": filePath,
	})

	if err != nil || !toolResult.Success {
		return fmt.Sprintf("Erro ao ler arquivo para correção: %s", toolResult.Error), nil
	}

	currentContent, ok := toolResult.Data["content"].(string)
	if !ok || currentContent == "" {
		return "Erro: não foi possível ler o conteúdo do arquivo", nil
	}

	a.colorBlue.Println("🔍 Analisando problema e gerando correção...")

	// Prompt para LLM corrigir o problema
	correctionPrompt := fmt.Sprintf(`Você é um assistente de programação especialista em debug.

ARQUIVO ATUAL: %s
%s

PROBLEMA REPORTADO PELO USUÁRIO:
"%s"

TAREFA:
1. Analise o código atual
2. Identifique o problema descrito pelo usuário
3. Corrija o código
4. Retorne o código COMPLETO corrigido

Responda com um JSON:
{
  "analysis": "breve análise do problema encontrado",
  "fixes": "lista de correções aplicadas",
  "code": "código completo corrigido (TUDO, não apenas a parte modificada)"
}`, filePath, currentContent, userMessage)

	llmResponse, err := a.llmClient.Complete(ctx, []llm.Message{
		{Role: "user", Content: correctionPrompt},
	}, &llm.CompletionOptions{Temperature: 0.3, MaxTokens: 4000})

	if err != nil {
		return "Erro ao gerar correção", err
	}

	// Parse JSON response
	jsonStr := strings.TrimSpace(llmResponse)
	jsonStr = strings.TrimPrefix(jsonStr, "```json")
	jsonStr = strings.TrimPrefix(jsonStr, "```")
	jsonStr = strings.TrimSuffix(jsonStr, "```")
	jsonStr = strings.TrimSpace(jsonStr)

	var correction map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &correction); err != nil {
		// Fallback: tentar usar resposta direta
		a.colorYellow.Println("⚠️  Não foi possível fazer parse, tentando abordagem simples...")
		return a.handleBugFixSimple(ctx, userMessage, filePath, currentContent)
	}

	analysis, _ := correction["analysis"].(string)
	fixes, _ := correction["fixes"].(string)
	correctedCode, _ := correction["code"].(string)

	if correctedCode == "" {
		return "Erro: não foi possível gerar código corrigido", nil
	}

	// Mostrar análise
	a.colorGreen.Printf("\n🔍 Análise:\n%s\n\n", analysis)
	a.colorGreen.Printf("✨ Correções aplicadas:\n%s\n\n", fixes)

	// Mostrar diff (primeiras linhas)
	preview := fmt.Sprintf("Arquivo: %s\nTamanho: %d bytes\n\nPreview das correções:\n%s",
		filePath, len(correctedCode), truncate(correctedCode, 500))

	// Confirmar correção
	if a.mode.RequiresConfirmation() {
		confirmed, err := a.confirmManager.ConfirmWithPreview(
			"Aplicar correções",
			preview,
		)

		if err != nil || !confirmed {
			return "✗ Correção cancelada pelo usuário", nil
		}
	}

	// Aplicar correção
	toolResult, err = a.toolRegistry.Execute(ctx, "file_writer", map[string]interface{}{
		"file_path": filePath,
		"content":   correctedCode,
		"mode":      "create", // Sobrescrever
	})

	if err != nil || !toolResult.Success {
		return fmt.Sprintf("Erro ao aplicar correções: %s", toolResult.Error), nil
	}

	return fmt.Sprintf("✓ Arquivo corrigido: %s\n\n🔍 Análise: %s\n✨ Correções: %s",
		filePath, analysis, fixes), nil
}

// handleBugFixSimple método simplificado para correção (fallback)
func (a *Agent) handleBugFixSimple(ctx context.Context, userMessage, filePath, currentContent string) (string, error) {
	a.colorBlue.Println("🔄 Usando método alternativo de correção...")

	prompt := fmt.Sprintf(`Corrija o problema no código abaixo.

ARQUIVO: %s
CÓDIGO ATUAL:
%s

PROBLEMA:
%s

Retorne o código COMPLETO corrigido (não apenas a parte modificada).`, filePath, currentContent, userMessage)

	correctedCode, err := a.llmClient.Complete(ctx, []llm.Message{
		{Role: "user", Content: prompt},
	}, &llm.CompletionOptions{Temperature: 0.3, MaxTokens: 4000})

	if err != nil {
		return "Erro ao gerar correção", err
	}

	// Limpar markdown
	correctedCode = strings.TrimPrefix(correctedCode, "```html")
	correctedCode = strings.TrimPrefix(correctedCode, "```css")
	correctedCode = strings.TrimPrefix(correctedCode, "```javascript")
	correctedCode = strings.TrimPrefix(correctedCode, "```")
	correctedCode = strings.TrimSuffix(correctedCode, "```")
	correctedCode = strings.TrimSpace(correctedCode)

	// Confirmar
	if a.mode.RequiresConfirmation() {
		preview := fmt.Sprintf("Arquivo: %s\nPreview:\n%s", filePath, truncate(correctedCode, 500))
		confirmed, err := a.confirmManager.ConfirmWithPreview("Aplicar correções", preview)

		if err != nil || !confirmed {
			return "✗ Correção cancelada", nil
		}
	}

	// Aplicar
	toolResult, err := a.toolRegistry.Execute(ctx, "file_writer", map[string]interface{}{
		"file_path": filePath,
		"content":   correctedCode,
		"mode":      "create",
	})

	if err != nil || !toolResult.Success {
		return fmt.Sprintf("Erro ao aplicar correções: %s", toolResult.Error), nil
	}

	return fmt.Sprintf("✓ Arquivo corrigido: %s", filePath), nil
}
