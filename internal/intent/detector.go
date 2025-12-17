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
	// Preparar contexto de arquivos
	filesContext := "nenhum"
	if len(recentFiles) > 0 {
		filesContext = strings.Join(recentFiles, ", ")
	}

	// Criar prompt do usuário
	userPrompt := fmt.Sprintf(UserPromptTemplate, currentDir, filesContext, userMessage)

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
