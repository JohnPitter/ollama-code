package intent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/johnpitter/ollama-code/internal/llm"
)

// Detector detecta intenções usando LLM
type Detector struct {
	llmClient *llm.Client
}

// NewDetector cria novo detector
func NewDetector(llmClient *llm.Client) *Detector {
	return &Detector{
		llmClient: llmClient,
	}
}

// Detect detecta a intenção de uma mensagem
func (d *Detector) Detect(ctx context.Context, userMessage, currentDir string, recentFiles []string) (*DetectionResult, error) {
	return d.DetectWithHistory(ctx, userMessage, currentDir, recentFiles, []llm.Message{})
}

// DetectWithHistory detecta a intenção usando histórico de mensagens anteriores
func (d *Detector) DetectWithHistory(ctx context.Context, userMessage, currentDir string, recentFiles []string, history []llm.Message) (*DetectionResult, error) {
	// Preparar contexto de arquivos
	filesContext := "nenhum"
	if len(recentFiles) > 0 {
		filesContext = strings.Join(recentFiles, ", ")
	}

	// Preparar contexto de conversa
	conversationContext := ""
	if len(history) > 0 {
		// Pegar últimas 4 mensagens (2 trocas) para contexto
		startIdx := len(history) - 4
		if startIdx < 0 {
			startIdx = 0
		}

		conversationContext = "\n\nHistórico recente da conversa:"
		for i := startIdx; i < len(history); i++ {
			role := "Usuário"
			if history[i].Role == "assistant" {
				role = "Assistente"
			}
			// Truncar mensagens muito longas
			content := history[i].Content
			if len(content) > 200 {
				content = content[:200] + "..."
			}
			conversationContext += fmt.Sprintf("\n%s: %s", role, content)
		}
	}

	// Criar prompt do usuário
	userPrompt := fmt.Sprintf(UserPromptTemplate, currentDir, filesContext, conversationContext, userMessage)

	// Chamar LLM
	messages := []llm.Message{
		{
			Role:    "user",
			Content: userPrompt,
		},
	}

	opts := &llm.CompletionOptions{
		Temperature:  0.1, // Baixa temperatura para respostas consistentes
		MaxTokens:    500,
		SystemPrompt: SystemPrompt,
	}

	response, err := d.llmClient.Complete(ctx, messages, opts)
	if err != nil {
		return nil, fmt.Errorf("llm completion: %w", err)
	}

	// Parse JSON response
	result, err := d.parseResponse(response)
	if err != nil {
		// Se falhar parsing, retornar intenção de pergunta como fallback
		return &DetectionResult{
			Intent:     IntentQuestion,
			Confidence: 0.5,
			Parameters: map[string]interface{}{},
			Reasoning:  "Fallback: não foi possível detectar intenção específica",
		}, nil
	}

	return result, nil
}

// parseResponse faz parse da resposta JSON
func (d *Detector) parseResponse(response string) (*DetectionResult, error) {
	// Limpar possível markdown
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	var result DetectionResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}

	return &result, nil
}

// DetectSimple detecta intenção de forma simplificada (sem contexto)
func (d *Detector) DetectSimple(ctx context.Context, userMessage string) (*DetectionResult, error) {
	return d.Detect(ctx, userMessage, ".", []string{})
}
