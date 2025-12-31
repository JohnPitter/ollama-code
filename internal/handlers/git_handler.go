package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/johnpitter/ollama-code/internal/intent"
	"github.com/johnpitter/ollama-code/internal/confirmation"
)

// GitHandler processa operações Git
type GitHandler struct {
	BaseHandler
}

// NewGitHandler cria novo handler
func NewGitHandler() *GitHandler {
	return &GitHandler{
		BaseHandler: NewBaseHandler("git"),
	}
}

// Handle processa intent de operação Git
func (h *GitHandler) Handle(ctx context.Context, deps *Dependencies, result *intent.DetectionResult) (string, error) {
	// Extrair parâmetros
	operation, ok := result.Parameters["operation"].(string)
	if !ok || operation == "" {
		operation = "status" // Default
	}

	// 🔍 Se operação precisa de confirmação interativa, usar AskQuestion
	if h.needsInteraction(operation) && deps.Mode.RequiresConfirmation() {
		confirmedOp, err := h.askGitOperation(deps, operation, result.UserMessage)
		if err != nil {
			return "", err
		}
		if confirmedOp == "" {
			return "Operação cancelada", nil
		}
		operation = confirmedOp
	}

	// Executar via tool registry
	params := map[string]interface{}{
		"operation": operation,
	}

	// Adicionar parâmetros extras se houver
	for key, value := range result.Parameters {
		if key != "operation" {
			params[key] = value
		}
	}

	toolResult, err := deps.ToolRegistry.Execute(ctx, "git_operations", params)
	if err != nil {
		return "", fmt.Errorf("erro na operação git: %w", err)
	}

	if !toolResult.Success {
		return "", fmt.Errorf("erro: %s", toolResult.Error)
	}

	// Formatar resultado com output do comando git
	return h.formatGitResult(toolResult, operation), nil
}

// formatGitResult formata resultado da operação git com output
func (h *GitHandler) formatGitResult(result ToolResult, operation string) string {
	output := result.Message + "\n\n"

	// Obter output do comando
	gitOutput, hasOutput := result.Data["output"].(string)

	if !hasOutput || gitOutput == "" {
		output += "✓ Operação concluída (sem output)\n"
		return output
	}

	// Formatar baseado na operação
	switch operation {
	case "status":
		output += "📊 Status do repositório:\n\n"
		if strings.TrimSpace(gitOutput) == "" {
			output += "✓ Nada a commitar, diretório de trabalho limpo\n"
		} else {
			output += fmt.Sprintf("```\n%s\n```\n", gitOutput)
		}

	case "log":
		output += "📜 Histórico de commits:\n\n"
		output += fmt.Sprintf("```\n%s\n```\n", gitOutput)

	case "diff":
		output += "🔍 Diferenças:\n\n"
		if strings.TrimSpace(gitOutput) == "" {
			output += "✓ Nenhuma alteração detectada\n"
		} else {
			// Limitar diff para primeiras 50 linhas
			lines := strings.Split(gitOutput, "\n")
			limit := 50
			if len(lines) < limit {
				limit = len(lines)
			}

			output += "```diff\n"
			for i := 0; i < limit; i++ {
				output += lines[i] + "\n"
			}
			if len(lines) > limit {
				output += fmt.Sprintf("\n... e mais %d linhas\n", len(lines)-limit)
			}
			output += "```\n"
		}

	case "branch":
		output += "🌿 Branches:\n\n"
		output += fmt.Sprintf("```\n%s\n```\n", gitOutput)

	default:
		// Output genérico para outros comandos
		output += fmt.Sprintf("Output:\n\n```\n%s\n```\n", gitOutput)
	}

	return output
}

// needsInteraction verifica se operação precisa de confirmação
func (h *GitHandler) needsInteraction(operation string) bool {
	interactiveOps := []string{
		"reset", "rebase", "cherry-pick", "revert",
		"merge", "stash drop", "branch -D",
	}

	opLower := strings.ToLower(operation)
	for _, op := range interactiveOps {
		if strings.Contains(opLower, op) {
			return true
		}
	}

	return false
}

// askGitOperation usa AskQuestion para confirmar operação Git
func (h *GitHandler) askGitOperation(deps *Dependencies, operation, userMessage string) (string, error) {
	if deps.ConfirmManager == nil {
		return operation, nil
	}

	// Criar question
	question := confirmation.Question{
		Question: fmt.Sprintf("Confirmar operação Git: %s?", operation),
		Header:   "Git Op",
		Options: []confirmation.Option{
			{
				Label:       "Executar",
				Description: fmt.Sprintf("Executar '%s' como especificado", operation),
			},
			{
				Label:       "Cancelar",
				Description: "Cancelar esta operação",
			},
		},
		MultiSelect: false,
	}

	// Perguntar
	answerInterface, err := deps.ConfirmManager.AskQuestion(question)
	if err != nil {
		return "", err
	}

	answer, ok := answerInterface.(*confirmation.Answer)
	if !ok || answer == nil {
		return "", fmt.Errorf("resposta inválida")
	}

	// Se escolheu "Cancelar"
	if answer.SelectedLabel == "Cancelar" {
		return "", nil
	}

	// Se escolheu "Other" com input customizado
	if answer.SelectedLabel == "Other" && answer.CustomInput != "" {
		return answer.CustomInput, nil
	}

	// Se escolheu "Executar", retornar operação original
	return operation, nil
}
