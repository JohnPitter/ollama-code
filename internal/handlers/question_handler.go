package handlers

import (
	"context"
	"fmt"

	"github.com/johnpitter/ollama-code/internal/intent"
)

// QuestionHandler processa perguntas gerais (handler padrão)
type QuestionHandler struct {
	BaseHandler
}

// NewQuestionHandler cria novo handler
func NewQuestionHandler() *QuestionHandler {
	return &QuestionHandler{
		BaseHandler: NewBaseHandler("question"),
	}
}

// Handle processa perguntas gerais usando LLM
func (h *QuestionHandler) Handle(ctx context.Context, deps *Dependencies, result *intent.DetectionResult) (string, error) {
	userMessage := result.UserMessage

	// Construir histórico com contexto limitado (últimas 10 mensagens)
	messages := make([]Message, 0)

	// Adicionar contexto do working directory
	systemMsg := Message{
		Role:    "system",
		Content: fmt.Sprintf("You are a helpful coding assistant. Working directory: %s", deps.WorkDir),
	}
	messages = append(messages, systemMsg)

	// Adicionar histórico recente
	historyStart := 0
	if len(deps.History) > 10 {
		historyStart = len(deps.History) - 10
	}
	messages = append(messages, deps.History[historyStart:]...)

	// Adicionar mensagem do usuário
	messages = append(messages, Message{
		Role:    "user",
		Content: userMessage,
	})

	// Completar com LLM
	response, err := deps.LLMClient.CompleteWithHistory(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("erro ao processar pergunta: %w", err)
	}

	return response, nil
}
