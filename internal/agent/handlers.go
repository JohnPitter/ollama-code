package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/johnpitter/ollama-code/internal/intent"
	"github.com/johnpitter/ollama-code/internal/llm"
	"github.com/johnpitter/ollama-code/internal/tools"
	"github.com/johnpitter/ollama-code/internal/websearch"
)

// handleReadFile processa leitura de arquivo
func (a *Agent) handleReadFile(ctx context.Context, result *intent.DetectionResult, userMessage string) (string, error) {
	filePath, ok := result.Parameters["file_path"].(string)
	if !ok || filePath == "" {
		return "Erro: caminho do arquivo não especificado", nil
	}

	// Detectar se há múltiplos arquivos mencionados
	fileList := extractMultipleFiles(filePath)
	if len(fileList) > 1 {
		// Processar múltiplos arquivos
		return a.handleMultiFileRead(ctx, fileList, userMessage)
	}

	// Processar arquivo único (comportamento original)
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

		// Detectar se usuário quer análise/explicação/review
		msgLower := strings.ToLower(userMessage)
		needsAnalysis := strings.Contains(msgLower, "analisa") ||
			strings.Contains(msgLower, "explica") ||
			strings.Contains(msgLower, "review") ||
			strings.Contains(msgLower, "revisa") ||
			strings.Contains(msgLower, "examina") ||
			strings.Contains(msgLower, "o que faz")

		if needsAnalysis {
			// Usar LLM para analisar/explicar
			a.colorBlue.Print("🔍 Analisando código")

			analysisPrompt := fmt.Sprintf(`Você é um assistente de programação expert. O usuário pediu:

"%s"

Arquivo: %s

Conteúdo:
%s

Sua tarefa: Responder à pergunta do usuário de forma clara e objetiva. Se o usuário pediu para:
- "analisa" → identifique a função/propósito do código, possíveis problemas, melhorias
- "explica" → explique o que o código faz de forma clara
- "review" → faça uma análise crítica apontando pontos fortes e fracos
- "examina" → examine em detalhes a estrutura e lógica

Responda em português de forma direta e técnica.`, userMessage, filePath, truncate(content, 3000))

			dotCount := 0
			response, err := a.llmClient.CompleteStreaming(ctx, []llm.Message{
				{Role: "user", Content: analysisPrompt},
			}, &llm.CompletionOptions{Temperature: 0.3, MaxTokens: 2000}, func(chunk string) {
				if dotCount < 30 {
					fmt.Print(".")
					dotCount++
				}
			})
			fmt.Println()

			if err != nil {
				return fmt.Sprintf("Erro ao analisar código: %v", err), nil
			}

			return response, nil
		}

		// Apenas mostrar conteúdo
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

	// Detectar se é uma edição de arquivo existente
	isEdit, editFilePath := detectEditRequest(userMessage)
	if isEdit && editFilePath != "" {
		// Usuário quer editar arquivo existente
		return a.handleFileEdit(ctx, userMessage, editFilePath)
	}

	// Detectar se é uma correção de arquivo recente
	recentlyModified := a.GetRecentlyModifiedFiles()
	isBugFix := detectBugReport(userMessage)

	if isBugFix && len(recentlyModified) > 0 {
		// Usuário reportou problema em arquivo recente
		return a.handleBugFix(ctx, userMessage, recentlyModified[0])
	}

	// Detectar se é uma requisição de múltiplos arquivos
	isMultiFile := detectMultiFileRequest(userMessage)
	if isMultiFile {
		return a.handleMultiFileWrite(ctx, userMessage)
	}

	// Se conteúdo não foi especificado, significa que o usuário quer que geremos
	if content == "" {
		a.colorBlue.Print("💭 Gerando conteúdo")

		// Usar LLM para gerar o conteúdo baseado na descrição do usuário
		generationPrompt := fmt.Sprintf(`Você é um assistente de programação. O usuário pediu:

"%s"

IMPORTANTE: Responda APENAS com JSON puro, SEM texto adicional antes ou depois.
Não escreva "Aqui está", "Claro", ou qualquer explicação.
Retorne SOMENTE o JSON abaixo:

{
  "file_path": "nome_do_arquivo.ext",
  "content": "código completo aqui",
  "mode": "create"
}

Regras:
- Primeira linha deve ser { (abre chave JSON)
- Última linha deve ser } (fecha chave JSON)
- file_path deve ser nome de arquivo válido (ex: index.html, style.css, main.py)
- Gere código funcional e completo no campo content
- Use boas práticas de programação
- NÃO adicione texto explicativo fora do JSON`, userMessage)

		// Usar streaming com indicador de progresso
		dotCount := 0
		llmResponse, err := a.llmClient.CompleteStreaming(ctx, []llm.Message{
			{Role: "user", Content: generationPrompt},
		}, &llm.CompletionOptions{Temperature: 0.7, MaxTokens: 2000}, func(chunk string) {
			// Mostrar progresso com pontos
			if dotCount < 30 {
				fmt.Print(".")
				dotCount++
			}
		})
		fmt.Println() // nova linha após progresso

		if err != nil {
			return "Erro ao gerar conteúdo", err
		}

		// Extrair JSON da resposta (LLM pode retornar com ```json, texto antes/depois, etc)
		jsonStr := extractJSON(llmResponse)

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
			// Limpar possíveis wrappers e artefatos
			content = cleanCodeContent(content, filePath)
		}
		if m, ok := parsed["mode"].(string); ok && m != "" {
			mode = m
		}
	}

	// Validações finais
	if filePath == "" {
		return "Erro: não foi possível determinar o caminho do arquivo", nil
	}

	// Validar nome de arquivo
	if !isValidFilename(filePath) {
		return fmt.Sprintf("Erro: nome de arquivo inválido: '%s'\nNome deve ser válido (ex: index.html, style.css)", filePath), nil
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

	// Verificar se usuário mencionou integração e sugerir
	integrationHint := generateIntegrationHint(userMessage, filePath)

	// Verificar se arquivo foi criado na raiz e sugerir melhor localização
	locationHint := generateLocationHint(filePath, a.workDir)

	// Formatar resposta
	response := fmt.Sprintf("✓ %s", toolResult.Message)
	if integrationHint != "" {
		response += "\n\n" + integrationHint
	}
	if locationHint != "" {
		response += "\n\n" + locationHint
	}

	return response, nil
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
func (a *Agent) handleGitOperation(ctx context.Context, result *intent.DetectionResult, userMessage string) (string, error) {
	if !a.mode.AllowsWrites() {
		return "❌ Operação bloqueada: modo somente leitura ativo", nil
	}

	// Tentar obter operation dos parâmetros
	operation, ok := result.Parameters["operation"].(string)

	// Se não veio nos parâmetros, inferir da mensagem do usuário
	if !ok || operation == "" {
		operation = detectGitOperation(userMessage)
	}

	// Garantir que operation está nos parâmetros para o tool
	params := make(map[string]interface{})
	for k, v := range result.Parameters {
		params[k] = v
	}
	params["operation"] = operation

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

	toolResult, err := a.toolRegistry.Execute(ctx, "git_operations", params)

	if err != nil || !toolResult.Success {
		return fmt.Sprintf("Erro na operação git: %s", toolResult.Error), nil
	}

	// Mostrar output se disponível
	if output, ok := toolResult.Data["output"].(string); ok && output != "" {
		return fmt.Sprintf("Operação git '%s':\n\n%s", operation, output), nil
	}

	return fmt.Sprintf("Operação git '%s' executada com sucesso", operation), nil
}

// detectGitOperation detecta qual operação git o usuário quer executar
func detectGitOperation(message string) string {
	msgLower := strings.ToLower(message)

	// Detectar operação específica por keywords
	if strings.Contains(msgLower, "diff") ||
	   strings.Contains(msgLower, "diferença") ||
	   strings.Contains(msgLower, "diferenças") ||
	   strings.Contains(msgLower, "mudança") ||
	   strings.Contains(msgLower, "mudanças") ||
	   strings.Contains(msgLower, "alteraç") ||
	   strings.Contains(msgLower, "changed") {
		return "diff"
	}

	if strings.Contains(msgLower, "log") ||
	   strings.Contains(msgLower, "histórico") ||
	   strings.Contains(msgLower, "commits") ||
	   strings.Contains(msgLower, "history") {
		return "log"
	}

	if strings.Contains(msgLower, "add") ||
	   strings.Contains(msgLower, "staged") ||
	   (strings.Contains(msgLower, "adiciona") && strings.Contains(msgLower, "git")) {
		return "add"
	}

	if strings.Contains(msgLower, "commit") ||
	   (strings.Contains(msgLower, "salva") && strings.Contains(msgLower, "git")) ||
	   (strings.Contains(msgLower, "grava") && strings.Contains(msgLower, "git")) {
		return "commit"
	}

	if strings.Contains(msgLower, "branch") ||
	   strings.Contains(msgLower, "ramo") ||
	   strings.Contains(msgLower, "ramificação") {
		return "branch"
	}

	// Default: status (operação mais segura e informativa)
	return "status"
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

	_, err = a.llmClient.CompleteStreaming(ctx, []llm.Message{
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

	// Resposta já foi impressa via streaming, retornar vazio para evitar duplicação
	return "", nil
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

	_, err := a.llmClient.CompleteStreaming(ctx, []llm.Message{
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

	// Resposta já foi impressa via streaming, retornar vazio para evitar duplicação
	return "", nil
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
	a.colorYellow.Print("🔄 Método alternativo")

	// Prompt mais direto e explícito
	prompt := fmt.Sprintf(`O usuário pediu: "%s"

IMPORTANTE:
- Linha 1: APENAS o nome do arquivo (ex: index.html ou style.css ou main.py)
- Linhas seguintes: código completo

NÃO escreva explicações, apenas:
Linha 1: nome_do_arquivo.ext
Linha 2+: código`, userMessage)

	// Usar streaming com progresso
	dotCount := 0
	response, err := a.llmClient.CompleteStreaming(ctx, []llm.Message{
		{Role: "user", Content: prompt},
	}, &llm.CompletionOptions{Temperature: 0.7, MaxTokens: 2000}, func(chunk string) {
		if dotCount < 20 {
			fmt.Print(".")
			dotCount++
		}
	})
	fmt.Println()

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

	// Limpar possíveis marcadores markdown do filename
	filePath = strings.TrimPrefix(filePath, "# ")
	filePath = strings.TrimPrefix(filePath, "## ")
	filePath = strings.TrimPrefix(filePath, "### ")
	filePath = strings.TrimPrefix(filePath, "Arquivo: ")
	filePath = strings.TrimPrefix(filePath, "Nome: ")
	filePath = strings.TrimSpace(filePath)

	// Limpar wrappers e artefatos do content
	content = cleanCodeContent(content, filePath)

	// Validar nome de arquivo
	if !isValidFilename(filePath) {
		return fmt.Sprintf("Erro: nome de arquivo inválido: '%s'\nResposta completa:\n%s", filePath, truncate(response, 500)), nil
	}

	// Validar conteúdo
	if content == "" {
		return fmt.Sprintf("Erro: conteúdo vazio.\nResposta do modelo:\n%s", response), nil
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

// extractJSON extrai JSON de texto que pode conter lixo ao redor
func extractJSON(text string) string {
	// Remover markdown code blocks comuns
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	// Procurar por JSON usando índices de { e }
	// Encontra o primeiro { e o último } balanceado
	firstBrace := strings.Index(text, "{")
	if firstBrace == -1 {
		return text // Sem JSON encontrado, retorna original
	}

	// Contar chaves para encontrar o } correto
	braceCount := 0
	lastBrace := -1
	for i := firstBrace; i < len(text); i++ {
		if text[i] == '{' {
			braceCount++
		} else if text[i] == '}' {
			braceCount--
			if braceCount == 0 {
				lastBrace = i
				break
			}
		}
	}

	if lastBrace == -1 {
		return text // JSON incompleto, retorna original
	}

	// Extrair JSON puro
	jsonStr := text[firstBrace : lastBrace+1]
	return strings.TrimSpace(jsonStr)
}

// isValidFilename verifica se string é um nome de arquivo válido
func isValidFilename(filename string) bool {
	// Limpar espaços
	filename = strings.TrimSpace(filename)

	// Verificações básicas
	if filename == "" || len(filename) > 255 {
		return false
	}

	// Deve ter extensão
	if !strings.Contains(filename, ".") {
		return false
	}

	// Não deve conter caracteres inválidos do Windows
	invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*", "\n", "\r"}
	for _, char := range invalidChars {
		if strings.Contains(filename, char) {
			return false
		}
	}

	// Não deve conter caminhos absolutos ou relativos complexos
	if strings.Contains(filename, "..") || strings.HasPrefix(filename, "/") || strings.HasPrefix(filename, "\\") {
		return false
	}

	// Não deve começar com espaço
	if strings.HasPrefix(filename, " ") {
		return false
	}

	// Não deve ser apenas "." sozinho (dotfiles como .env são permitidos)
	if filename == "." {
		return false
	}

	// Não deve conter frases (espaços demais indicam texto, não filename)
	spaceCount := strings.Count(filename, " ")
	if spaceCount > 2 {
		return false
	}

	return true
}

// detectEditRequest detecta se usuário quer editar arquivo existente e retorna nome do arquivo
func detectEditRequest(message string) (bool, string) {
	msgLower := strings.ToLower(message)

	// Keywords que indicam edição de arquivo existente
	editKeywords := []string{
		"adiciona",
		"adiciona no",
		"edita",
		"edita o",
		"modifica",
		"modifica o",
		"atualiza",
		"atualiza o",
		"muda",
		"muda o",
		"altera",
		"altera o",
		"insere",
		"insere no",
		"corrige",
		"corrige o",
		"conserta",
		"conserta o",
		"arruma",
		"arruma o",
		"resolve",
		"resolve o",
		"fix",
	}

	// Verificar se mensagem contém keyword de edição
	isEdit := false
	for _, keyword := range editKeywords {
		if strings.Contains(msgLower, keyword) {
			isEdit = true
			break
		}
	}

	if !isEdit {
		return false, ""
	}

	// Tentar extrair nome do arquivo
	// Procurar por palavras que parecem nome de arquivo (tem extensão válida)
	words := strings.Fields(message)
	var foundFile string

	for i, word := range words {
		// Limpar pontuação
		cleanWord := strings.Trim(word, ".,;:!?\"'")

		// Se encontrou "arquivo" ou "no" ou "em", próxima palavra pode ser o nome
		if strings.ToLower(word) == "arquivo" || strings.ToLower(word) == "no" || strings.ToLower(word) == "em" {
			if i+1 < len(words) {
				potentialFile := strings.Trim(words[i+1], ".,;:!?\"'")
				if isValidFilename(potentialFile) {
					foundFile = potentialFile
					break
				}
			}
		}

		// Também procurar por nomes de arquivo diretamente
		if isValidFilename(cleanWord) {
			foundFile = cleanWord
			break
		}
	}

	// Só retorna true se encontrou TANTO keyword de edição QUANTO nome de arquivo
	if isEdit && foundFile != "" {
		return true, foundFile
	}

	return false, ""
}

// handleFileEdit lida com edição de arquivo existente fazendo merge inteligente
func (a *Agent) handleFileEdit(ctx context.Context, userMessage, filePath string) (string, error) {
	a.colorYellow.Printf("✏️  Editando arquivo existente: %s\n", filePath)

	// 1. Ler arquivo atual
	a.colorBlue.Println("📖 Lendo conteúdo atual...")
	toolResult, err := a.toolRegistry.Execute(ctx, "file_reader", map[string]interface{}{
		"file_path": filePath,
	})

	if err != nil || !toolResult.Success {
		// Se arquivo não existe, criar novo
		a.colorYellow.Printf("⚠️  Arquivo não existe, será criado como novo\n")
		return a.handleWriteFile(ctx, &intent.DetectionResult{
			Intent: intent.IntentWriteFile,
			Parameters: map[string]interface{}{
				"file_path": filePath,
			},
		}, userMessage)
	}

	currentContent := toolResult.Data

	// 2. Usar LLM para fazer merge inteligente
	a.colorBlue.Print("🔄 Mesclando mudanças")

	mergePrompt := fmt.Sprintf(`Você é um assistente de programação. O usuário tem um arquivo com o seguinte conteúdo:

<arquivo_atual>
%s
</arquivo_atual>

O usuário pediu: "%s"

Sua tarefa: Editar o arquivo PRESERVANDO o código existente e adicionando/modificando conforme solicitado.

IMPORTANTE: Retorne APENAS o código completo do arquivo editado, SEM explicações, SEM JSON, SEM markdown.
Primeira linha deve ser a primeira linha do código.
Última linha deve ser a última linha do código.

Regras:
- PRESERVE todo código existente que não precisa ser alterado
- ADICIONE o novo código no local apropriado
- MANTENHA a estrutura e formatação do arquivo
- NÃO remova funções/métodos existentes a menos que explicitamente solicitado
- Se adicionar função, coloque após funções existentes
- Mantenha imports/includes existentes`, currentContent, userMessage)

	dotCount := 0
	newContent, err := a.llmClient.CompleteStreaming(ctx, []llm.Message{
		{Role: "user", Content: mergePrompt},
	}, &llm.CompletionOptions{Temperature: 0.3, MaxTokens: 4000}, func(chunk string) {
		if dotCount < 30 {
			fmt.Print(".")
			dotCount++
		}
	})
	fmt.Println()

	if err != nil {
		return "Erro ao mesclar mudanças", err
	}

	// Limpar possíveis markdown code blocks e wrappers
	newContent = cleanCodeContent(newContent, filePath)

	// 3. Mostrar diff (preview das mudanças)
	a.colorGreen.Printf("\n📝 Mudanças detectadas:\n")
	fmt.Printf("Arquivo: %s\n", filePath)
	fmt.Printf("Tamanho original: %d bytes\n", len(currentContent))
	fmt.Printf("Tamanho novo: %d bytes\n", len(newContent))

	// 4. Confirmar se necessário
	if a.mode.RequiresConfirmation() {
		preview := fmt.Sprintf("Arquivo: %s\nTamanho: %d → %d bytes\n\nNovo conteúdo:\n%s",
			filePath, len(currentContent), len(newContent), truncate(newContent, 500))

		confirmed, err := a.confirmManager.ConfirmWithPreview(
			"Salvar mudanças",
			preview,
		)

		if err != nil || !confirmed {
			return "✗ Operação cancelada pelo usuário", nil
		}
	}

	// 5. Salvar arquivo editado
	saveResult, err := a.toolRegistry.Execute(ctx, "file_writer", map[string]interface{}{
		"file_path": filePath,
		"content":   newContent,
		"mode":      "create", // Sobrescreve mas preservamos conteúdo via merge
	})

	if err != nil || !saveResult.Success {
		return fmt.Sprintf("Erro ao salvar arquivo: %s", saveResult.Error), nil
	}

	// Registrar como recentemente modificado
	a.AddRecentFile(filePath)

	return fmt.Sprintf("✓ Arquivo editado com sucesso: %s", filePath), nil
}

// cleanCodeContent remove wrappers JSON, markdown e outros artefatos do código gerado
// Recebe o filename para detectar tipo de arquivo e evitar limpar JSONs válidos
func cleanCodeContent(content string, filename string) string {
	content = strings.TrimSpace(content)

	// Detectar extensão do arquivo
	isJSON := strings.HasSuffix(strings.ToLower(filename), ".json") ||
		strings.HasSuffix(strings.ToLower(filename), ".jsonc")

	// 1. Remover JSON wrapper se presente: {"content": "código"}
	if strings.HasPrefix(content, "{") && strings.Contains(content, `"content":`) {
		// Tentar extrair content do JSON
		startIdx := strings.Index(content, `"content":`)
		if startIdx != -1 {
			// Pular até o valor
			startIdx += len(`"content":`)
			content = content[startIdx:]
			content = strings.TrimSpace(content)
			// Remover aspas iniciais
			content = strings.TrimPrefix(content, `"`)
			// Encontrar fim do valor (última aspas antes de })
			endIdx := strings.LastIndex(content, `"`)
			if endIdx != -1 {
				content = content[:endIdx]
			}
			// Decodificar escapes (\n → newline)
			content = strings.ReplaceAll(content, `\n`, "\n")
			content = strings.ReplaceAll(content, `\t`, "\t")
			content = strings.ReplaceAll(content, `\"`, `"`)
		}
	}

	content = strings.TrimSpace(content)

	// 2. Remover markdown code blocks (```language ... ```)
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	// 3. Remover nome de linguagem na primeira linha se presente
	lines := strings.Split(content, "\n")
	if len(lines) > 0 {
		firstLine := strings.ToLower(strings.TrimSpace(lines[0]))
		// Lista de linguagens comuns que podem aparecer
		languages := []string{"go", "python", "javascript", "java", "rust", "cpp", "c", "html", "css", "json", "yaml", "bash", "sh"}
		for _, lang := range languages {
			if firstLine == lang || firstLine == "```"+lang {
				// Remover primeira linha
				lines = lines[1:]
				break
			}
		}
		content = strings.Join(lines, "\n")
	}

	content = strings.TrimSpace(content)

	// 4. Remover chaves extras se arquivo começar e terminar com { }
	// (possível resíduo de JSON wrapper)
	// IMPORTANTE: NÃO fazer isso para arquivos .json pois são estruturalmente válidos
	if !isJSON && strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
		// Verificar se não é código válido (struct, objeto, etc)
		// Se segunda linha não é código, é provável que seja wrapper
		testLines := strings.Split(content, "\n")
		if len(testLines) > 1 {
			secondLine := strings.TrimSpace(testLines[1])
			// Se segunda linha não parece código (não tem keywords), é wrapper
			if !strings.Contains(secondLine, "package") &&
				!strings.Contains(secondLine, "import") &&
				!strings.Contains(secondLine, "func") &&
				!strings.Contains(secondLine, "class") &&
				!strings.Contains(secondLine, "def") &&
				!strings.Contains(secondLine, "const") &&
				!strings.Contains(secondLine, "var") &&
				!strings.Contains(secondLine, "let") {
				// É wrapper, remover primeira e última linha
				if len(testLines) > 2 {
					content = strings.Join(testLines[1:len(testLines)-1], "\n")
				}
			}
		}
	}

	return strings.TrimSpace(content)
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

// detectMultiFileRequest detecta se usuário quer criar múltiplos arquivos
func detectMultiFileRequest(message string) bool {
	msgLower := strings.ToLower(message)

	// Se mensagem contém keywords de integração, NÃO é multi-file
	// (usuário quer criar um arquivo e conectar em outro existente)
	integrationKeywords := []string{
		"conecta no", "conecta ao", "conecta em", "conecta com",
		"adiciona no", "adiciona ao", "adiciona em",
		"integra no", "integra ao", "integra em", "integra com",
		"inclui no", "inclui em",
		"linka no", "linka ao", "linka em",
		"importa no", "importa em",
	}
	for _, keyword := range integrationKeywords {
		if strings.Contains(msgLower, keyword) {
			return false
		}
	}

	// Keywords explícitas de multi-file
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
	}

	for _, keyword := range multiFileKeywords {
		if strings.Contains(msgLower, keyword) {
			return true
		}
	}

	// Padrão: "arquivos" (plural) + " e "
	if strings.Contains(msgLower, "arquivos") && strings.Contains(msgLower, " e ") {
		return true
	}

	// Padrão: "files" (plural) + " and "
	if strings.Contains(msgLower, "files") && strings.Contains(msgLower, " and ") {
		return true
	}

	// Padrão: número + "arquivos" (ex: "3 arquivos", "dois arquivos")
	numberKeywords := []string{"2 arquivos", "3 arquivos", "4 arquivos", "5 arquivos",
		"dois arquivos", "três arquivos", "tres arquivos", "quatro arquivos", "cinco arquivos"}
	for _, keyword := range numberKeywords {
		if strings.Contains(msgLower, keyword) {
			return true
		}
	}

	// Padrão: contar extensões de arquivo distintas (se >= 2, é multi-file)
	extensions := make(map[string]bool)
	words := strings.Fields(message)
	for _, word := range words {
		if strings.Contains(word, ".") {
			ext := strings.ToLower(filepath.Ext(word))
			if ext != "" && len(ext) <= 10 { // extensões válidas têm no máximo ~10 chars
				extensions[ext] = true
			}
		}
	}
	if len(extensions) >= 2 {
		return true
	}

	return false
}

// generateIntegrationHint gera sugestão de integração se usuário mencionou conectar/integrar arquivos
func generateIntegrationHint(userMessage, createdFile string) string {
	msgLower := strings.ToLower(userMessage)

	// Keywords de integração
	integrationKeywords := []string{
		"conecta no", "conecta ao", "conecta em", "conecta com",
		"adiciona no", "adiciona ao", "adiciona em",
		"integra no", "integra ao", "integra em", "integra com",
		"inclui no", "inclui em",
		"linka no", "linka ao", "linka em",
		"importa no", "importa em",
	}

	// Verificar se mensagem contém keyword de integração
	hasIntegration := false
	for _, keyword := range integrationKeywords {
		if strings.Contains(msgLower, keyword) {
			hasIntegration = true
			break
		}
	}

	if !hasIntegration {
		return ""
	}

	// Tentar extrair arquivo de destino
	targetFile := extractTargetFile(msgLower, integrationKeywords)
	if targetFile == "" {
		return ""
	}

	// Gerar sugestão baseada na extensão do arquivo criado
	ext := strings.ToLower(filepath.Ext(createdFile))
	baseName := filepath.Base(createdFile)

	switch ext {
	case ".js":
		return fmt.Sprintf("💡 Dica: Para usar %s no %s, adicione:\n   <script src=\"%s\"></script>", baseName, targetFile, baseName)
	case ".css":
		return fmt.Sprintf("💡 Dica: Para usar %s no %s, adicione:\n   <link rel=\"stylesheet\" href=\"%s\">", baseName, targetFile, baseName)
	case ".jsx", ".tsx":
		return fmt.Sprintf("💡 Dica: Para importar %s no %s, adicione:\n   import Component from './%s';", baseName, targetFile, baseName)
	case ".ts":
		importName := strings.TrimSuffix(baseName, ext)
		return fmt.Sprintf("💡 Dica: Para importar %s no %s, adicione:\n   import { %s } from './%s';", baseName, targetFile, importName, importName)
	case ".go":
		return fmt.Sprintf("💡 Dica: Para usar %s no %s, certifique-se de que ambos estão no mesmo package ou importe adequadamente.", baseName, targetFile)
	case ".py":
		importName := strings.TrimSuffix(baseName, ext)
		return fmt.Sprintf("💡 Dica: Para importar %s no %s, adicione:\n   from %s import *", baseName, targetFile, importName)
	}

	return ""
}

// extractTargetFile extrai nome do arquivo de destino da mensagem
func extractTargetFile(msgLower string, integrationKeywords []string) string {
	for _, keyword := range integrationKeywords {
		if strings.Contains(msgLower, keyword) {
			parts := strings.Split(msgLower, keyword)
			if len(parts) > 1 {
				afterKeyword := strings.TrimSpace(parts[1])
				words := strings.Fields(afterKeyword)

				// Procurar por nome de arquivo (contém extensão comum)
				for _, word := range words {
					word = strings.Trim(word, ".,;:\"'")
					if strings.Contains(word, ".html") ||
						strings.Contains(word, ".htm") ||
						strings.Contains(word, ".js") ||
						strings.Contains(word, ".jsx") ||
						strings.Contains(word, ".tsx") ||
						strings.Contains(word, ".ts") ||
						strings.Contains(word, ".css") ||
						strings.Contains(word, ".go") ||
						strings.Contains(word, ".py") ||
						strings.Contains(word, ".java") ||
						strings.Contains(word, ".php") {
						return word
					}
				}
			}
		}
	}
	return ""
}

// extractMultipleFiles extrai lista de arquivos de uma string
func extractMultipleFiles(filePath string) []string {
	// Limpar espaços
	filePath = strings.TrimSpace(filePath)

	var files []string

	// Estratégia 1: Separar por vírgulas
	if strings.Contains(filePath, ",") {
		parts := strings.Split(filePath, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				files = append(files, part)
			}
		}
		return files
	}

	// Estratégia 2: Separar por " e " ou " and "
	if strings.Contains(filePath, " e ") || strings.Contains(filePath, " and ") {
		// Substituir " e " por vírgula
		filePath = strings.ReplaceAll(filePath, " e ", ",")
		filePath = strings.ReplaceAll(filePath, " and ", ",")
		parts := strings.Split(filePath, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				files = append(files, part)
			}
		}
		return files
	}

	// Estratégia 3: Separar por espaços (apenas se houver múltiplas extensões de arquivo)
	if strings.Contains(filePath, " ") {
		parts := strings.Fields(filePath)
		// Contar quantas partes parecem ser arquivos (têm extensão)
		fileCount := 0
		for _, part := range parts {
			if strings.Contains(part, ".") && !strings.HasPrefix(part, ".") {
				fileCount++
			}
		}

		// Se temos múltiplos arquivos, retornar lista
		if fileCount > 1 {
			for _, part := range parts {
				if strings.Contains(part, ".") && !strings.HasPrefix(part, ".") {
					files = append(files, part)
				}
			}
			return files
		}
	}

	// Caso padrão: retornar como arquivo único
	return []string{filePath}
}

// generateLocationHint sugere melhor localização se arquivo foi criado na raiz
func generateLocationHint(filePath, workDir string) string {
	// Ignorar se não for arquivo na raiz (já tem caminho)
	baseName := filepath.Base(filePath)
	if filePath != baseName {
		// Arquivo já tem caminho (ex: src/main.go)
		return ""
	}

	// Detectar tipo de projeto
	projectType := detectProjectType(workDir)
	if projectType == "" {
		// Sem estrutura detectável
		return ""
	}

	// Sugerir localização baseada no tipo de arquivo e projeto
	suggestions := suggestFileLocation(baseName, projectType, workDir)
	if len(suggestions) == 0 {
		return ""
	}

	hint := "💡 Dica de organização: Este arquivo poderia estar melhor em:\n"
	for _, suggestion := range suggestions {
		hint += fmt.Sprintf("   📁 %s\n", suggestion)
	}
	hint += "\nConsidere mover o arquivo para manter o projeto organizado."

	return hint
}

// detectProjectType detecta tipo de projeto examinando arquivos marcadores
func detectProjectType(workDir string) string {
	// Go project
	if fileExists(filepath.Join(workDir, "go.mod")) {
		return "go"
	}

	// Node.js project
	if fileExists(filepath.Join(workDir, "package.json")) {
		return "nodejs"
	}

	// Python project
	if fileExists(filepath.Join(workDir, "requirements.txt")) ||
	   fileExists(filepath.Join(workDir, "setup.py")) ||
	   fileExists(filepath.Join(workDir, "pyproject.toml")) {
		return "python"
	}

	// Rust project
	if fileExists(filepath.Join(workDir, "Cargo.toml")) {
		return "rust"
	}

	// Java/Maven project
	if fileExists(filepath.Join(workDir, "pom.xml")) {
		return "java-maven"
	}

	// Java/Gradle project
	if fileExists(filepath.Join(workDir, "build.gradle")) ||
	   fileExists(filepath.Join(workDir, "build.gradle.kts")) {
		return "java-gradle"
	}

	return ""
}

// fileExists verifica se arquivo existe
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// suggestFileLocation sugere localizações apropriadas baseado no tipo de projeto
func suggestFileLocation(filename, projectType, workDir string) []string {
	ext := strings.ToLower(filepath.Ext(filename))
	var suggestions []string

	switch projectType {
	case "go":
		// Estrutura Go padrão
		if strings.HasSuffix(filename, "_test.go") {
			// Arquivos de teste vão no mesmo diretório do código
			suggestions = append(suggestions, "internal/"+strings.TrimSuffix(filename, "_test.go")+"/")
		} else if strings.Contains(filename, "main.go") {
			// Executáveis vão em cmd/
			if dirExists(filepath.Join(workDir, "cmd")) {
				suggestions = append(suggestions, "cmd/<nome-do-app>/main.go")
			}
		} else {
			// Código interno vai em internal/
			if dirExists(filepath.Join(workDir, "internal")) {
				suggestions = append(suggestions, "internal/<package>/"+filename)
			}
			// Código público vai em pkg/
			if dirExists(filepath.Join(workDir, "pkg")) {
				suggestions = append(suggestions, "pkg/<package>/"+filename)
			}
		}

	case "nodejs":
		// Estrutura Node.js comum
		if ext == ".js" || ext == ".ts" || ext == ".jsx" || ext == ".tsx" {
			if dirExists(filepath.Join(workDir, "src")) {
				suggestions = append(suggestions, "src/"+filename)
			}
			if strings.Contains(filename, "test") || strings.Contains(filename, "spec") {
				suggestions = append(suggestions, "test/"+filename)
			}
		} else if ext == ".json" && filename != "package.json" {
			suggestions = append(suggestions, "config/"+filename)
		}

	case "python":
		// Estrutura Python comum
		if ext == ".py" {
			if strings.Contains(filename, "test_") {
				suggestions = append(suggestions, "tests/"+filename)
			} else {
				if dirExists(filepath.Join(workDir, "src")) {
					suggestions = append(suggestions, "src/"+filename)
				}
				// Nome do package baseado no diretório
				pkgName := filepath.Base(workDir)
				suggestions = append(suggestions, pkgName+"/"+filename)
			}
		}

	case "rust":
		// Estrutura Rust padrão
		if ext == ".rs" {
			if filename == "main.rs" {
				suggestions = append(suggestions, "src/main.rs")
			} else if filename == "lib.rs" {
				suggestions = append(suggestions, "src/lib.rs")
			} else {
				suggestions = append(suggestions, "src/"+filename)
			}
		}

	case "java-maven", "java-gradle":
		// Estrutura Java padrão
		if ext == ".java" {
			if strings.Contains(filename, "Test") {
				suggestions = append(suggestions, "src/test/java/<package>/"+filename)
			} else {
				suggestions = append(suggestions, "src/main/java/<package>/"+filename)
			}
		}
	}

	return suggestions
}

// dirExists verifica se diretório existe
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// handleMultiFileRead processa leitura de múltiplos arquivos
func (a *Agent) handleMultiFileRead(ctx context.Context, fileList []string, userMessage string) (string, error) {
	a.colorBlue.Printf("📚 Lendo %d arquivos...\n", len(fileList))

	var results []string
	var failedFiles []string

	for _, filePath := range fileList {
		filePath = strings.TrimSpace(filePath)
		if filePath == "" {
			continue
		}

		a.colorBlue.Printf("   📄 %s\n", filePath)

		// Ler arquivo usando o tool
		toolResult, err := a.toolRegistry.Execute(ctx, "file_reader", map[string]interface{}{
			"file_path": filePath,
		})

		if err != nil || !toolResult.Success {
			a.colorYellow.Printf("   ⚠️  Erro ao ler %s: %s\n", filePath, toolResult.Error)
			failedFiles = append(failedFiles, filePath)
			continue
		}

		// Extrair conteúdo
		fileType, _ := toolResult.Data["type"].(string)
		if fileType == "text" {
			content, ok := toolResult.Data["content"].(string)
			if ok {
				// Truncar se muito longo
				if len(content) > 1000 {
					content = content[:1000] + "\n... (truncado)"
				}
				results = append(results, fmt.Sprintf("=== %s ===\n%s\n", filePath, content))
			}
		}
	}

	if len(results) == 0 {
		return fmt.Sprintf("❌ Nenhum arquivo foi lido com sucesso.\n\nArquivos com falha: %s", strings.Join(failedFiles, ", ")), nil
	}

	// Construir resposta
	response := fmt.Sprintf("✓ Lidos %d de %d arquivos:\n\n", len(results), len(fileList))
	response += strings.Join(results, "\n")

	if len(failedFiles) > 0 {
		response += fmt.Sprintf("\n\n⚠️  %d arquivo(s) com falha: %s", len(failedFiles), strings.Join(failedFiles, ", "))
	}

	// Detectar se usuário quer análise/comparação
	msgLower := strings.ToLower(userMessage)
	needsAnalysis := strings.Contains(msgLower, "relação") ||
		strings.Contains(msgLower, "compara") ||
		strings.Contains(msgLower, "diferença") ||
		strings.Contains(msgLower, "analisa") ||
		strings.Contains(msgLower, "explica") ||
		strings.Contains(msgLower, "me diz")

	if needsAnalysis && len(results) > 0 {
		a.colorBlue.Print("\n🔍 Analisando arquivos")

		analysisPrompt := fmt.Sprintf(`Você é um assistente de programação expert. O usuário pediu:

"%s"

Conteúdo dos arquivos:
%s

Sua tarefa: Responder à pergunta do usuário de forma clara e objetiva sobre esses arquivos.

Responda em português de forma direta e técnica.`, userMessage, response)

		dotCount := 0
		llmResponse, err := a.llmClient.CompleteStreaming(ctx, []llm.Message{
			{Role: "user", Content: analysisPrompt},
		}, &llm.CompletionOptions{Temperature: 0.3, MaxTokens: 2000}, func(chunk string) {
			if dotCount < 30 {
				fmt.Print(".")
				dotCount++
			}
		})
		fmt.Println()

		if err == nil {
			return llmResponse, nil
		}
	}

	return response, nil
}

// handleMultiFileWrite processa criação de múltiplos arquivos
func (a *Agent) handleMultiFileWrite(ctx context.Context, userMessage string) (string, error) {
	a.colorBlue.Println("📦 Detectada requisição de múltiplos arquivos...")
	a.colorBlue.Print("💭 Gerando projeto")

	// Prompt para LLM gerar múltiplos arquivos (simplificado)
	multiFilePrompt := fmt.Sprintf(`Você é um assistente de programação. O usuário pediu:

"%s"

Responda APENAS com JSON:
{
  "files": [
    {"file_path": "index.html", "content": "código HTML completo"},
    {"file_path": "style.css", "content": "código CSS completo"},
    {"file_path": "script.js", "content": "código JS completo"}
  ]
}

Regras:
- Crie TODOS os arquivos solicitados
- HTML deve ter <link rel="stylesheet" href="..."> e <script src="...">
- Código funcional e completo
- Não inclua explicações fora do JSON`, userMessage)

	// Usar streaming com indicador de progresso
	dotCount := 0
	llmResponse, err := a.llmClient.CompleteStreaming(ctx, []llm.Message{
		{Role: "user", Content: multiFilePrompt},
	}, &llm.CompletionOptions{Temperature: 0.7, MaxTokens: 3000}, func(chunk string) {
		// Mostrar progresso com pontos
		if dotCount < 30 {
			fmt.Print(".")
			dotCount++
		}
	})
	fmt.Println() // nova linha após progresso

	if err != nil {
		return "Erro ao gerar arquivos", err
	}

	// Parse JSON (não usar parseJSON pois valida file_path que não existe em multi-file)
	jsonStr := strings.TrimSpace(llmResponse)
	jsonStr = strings.TrimPrefix(jsonStr, "```json")
	jsonStr = strings.TrimPrefix(jsonStr, "```")
	jsonStr = strings.TrimSuffix(jsonStr, "```")
	jsonStr = strings.TrimSpace(jsonStr)

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		a.colorYellow.Printf("⚠️  Erro ao parsear JSON de múltiplos arquivos: %v\n", err)
		a.colorYellow.Println("⚠️  Tentando criar arquivo único...")
		// Fallback para criação de arquivo único
		return a.generateAndWriteFileSimple(ctx, userMessage)
	}

	// Extrair array de arquivos
	filesInterface, ok := parsed["files"]
	if !ok {
		a.colorYellow.Println("⚠️  Campo 'files' não encontrado, tentando arquivo único...")
		return a.generateAndWriteFileSimple(ctx, userMessage)
	}

	filesArray, ok := filesInterface.([]interface{})
	if !ok {
		a.colorYellow.Println("⚠️  'files' não é um array, tentando arquivo único...")
		return a.generateAndWriteFileSimple(ctx, userMessage)
	}

	if len(filesArray) == 0 {
		return "Erro: nenhum arquivo foi gerado", nil
	}

	a.colorGreen.Printf("📁 %d arquivos serão criados:\n", len(filesArray))

	// Processar cada arquivo
	var createdFiles []string
	var failedFiles []string

	for i, fileInterface := range filesArray {
		fileMap, ok := fileInterface.(map[string]interface{})
		if !ok {
			a.colorYellow.Printf("⚠️  Arquivo %d tem formato inválido, pulando...\n", i+1)
			continue
		}

		filePath, ok := fileMap["file_path"].(string)
		if !ok || filePath == "" {
			a.colorYellow.Printf("⚠️  Arquivo %d sem caminho válido, pulando...\n", i+1)
			continue
		}

		content, ok := fileMap["content"].(string)
		if !ok || content == "" {
			a.colorYellow.Printf("⚠️  Arquivo %s sem conteúdo, pulando...\n", filePath)
			failedFiles = append(failedFiles, filePath)
			continue
		}

		a.colorBlue.Printf("   - %s (%d bytes)\n", filePath, len(content))

		// Pedir confirmação se necessário (apenas uma vez para o projeto todo)
		if a.mode.RequiresConfirmation() && i == 0 {
			filesList := ""
			for _, f := range filesArray {
				if fm, ok := f.(map[string]interface{}); ok {
					if fp, ok := fm["file_path"].(string); ok {
						filesList += fmt.Sprintf("   - %s\n", fp)
					}
				}
			}

			preview := fmt.Sprintf("Projeto com %d arquivos:\n%s", len(filesArray), filesList)
			confirmed, err := a.confirmManager.ConfirmWithPreview("Criar projeto multi-file", preview)

			if err != nil || !confirmed {
				return "✗ Operação cancelada pelo usuário", nil
			}
		}

		// Criar arquivo
		toolResult, err := a.toolRegistry.Execute(ctx, "file_writer", map[string]interface{}{
			"file_path": filePath,
			"content":   content,
			"mode":      "create",
		})

		if err != nil || !toolResult.Success {
			a.colorRed.Printf("✗ Erro ao criar %s: %s\n", filePath, toolResult.Error)
			failedFiles = append(failedFiles, filePath)
		} else {
			a.colorGreen.Printf("✓ %s criado\n", filePath)
			createdFiles = append(createdFiles, filePath)
			a.AddRecentFile(filePath)
		}
	}

	// Resumo
	summary := fmt.Sprintf("\n✓ Projeto criado com %d arquivo(s):", len(createdFiles))
	for _, file := range createdFiles {
		summary += fmt.Sprintf("\n   - %s", file)
	}

	if len(failedFiles) > 0 {
		summary += fmt.Sprintf("\n\n⚠️  %d arquivo(s) falharam:", len(failedFiles))
		for _, file := range failedFiles {
			summary += fmt.Sprintf("\n   - %s", file)
		}
	}

	return summary, nil
}
