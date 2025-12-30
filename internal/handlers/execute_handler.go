package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/johnpitter/ollama-code/internal/intent"
)

// ExecuteHandler processa execução de comandos
type ExecuteHandler struct {
	BaseHandler
}

// NewExecuteHandler cria novo handler
func NewExecuteHandler() *ExecuteHandler {
	return &ExecuteHandler{
		BaseHandler: NewBaseHandler("execute"),
	}
}

// Handle processa intent de execução de comando
func (h *ExecuteHandler) Handle(ctx context.Context, deps *Dependencies, result *intent.DetectionResult) (string, error) {
	// Extrair parâmetros
	command, ok := result.Parameters["command"].(string)
	if !ok || command == "" {
		return "", fmt.Errorf("comando não especificado")
	}

	// 📝 Criar TODO para tracking
	var todoID string
	if deps.TodoManager != nil {
		id, err := deps.TodoManager.Add(
			fmt.Sprintf("Executando comando: %s", h.truncateCommand(command)),
			fmt.Sprintf("Executando %s", h.truncateCommand(command)),
		)
		if err == nil {
			todoID = id
		}
	}

	// Verificar se comando é perigoso
	if h.isDangerousCommand(command) {
		if !deps.Mode.RequiresConfirmation() {
			// Cancelar TODO
			if todoID != "" && deps.TodoManager != nil {
				deps.TodoManager.Delete(todoID)
			}
			return "", fmt.Errorf("comando perigoso requer modo interativo: %s", command)
		}

		// Pedir confirmação (com opção de cancelar)
		confirmed, err := deps.ConfirmManager.Confirm(
			fmt.Sprintf("⚠️  Comando potencialmente perigoso. Executar: %s ?", command),
		)
		if err != nil || !confirmed {
			// Cancelar TODO
			if todoID != "" && deps.TodoManager != nil {
				deps.TodoManager.Delete(todoID)
			}
			return "Comando cancelado pelo usuário", nil
		}
	}

	// Executar via tool registry
	params := map[string]interface{}{
		"command": command,
	}

	toolResult, err := deps.ToolRegistry.Execute(ctx, "command_executor", params)
	if err != nil {
		// ❌ Falha na execução
		if todoID != "" && deps.TodoManager != nil {
			deps.TodoManager.Delete(todoID)
		}
		return "", fmt.Errorf("erro ao executar comando: %w", err)
	}

	if !toolResult.Success {
		// ❌ Comando retornou erro
		if todoID != "" && deps.TodoManager != nil {
			deps.TodoManager.Delete(todoID)
		}
		return "", fmt.Errorf("erro: %s", toolResult.Error)
	}

	// ✅ Completar TODO
	if todoID != "" && deps.TodoManager != nil {
		deps.TodoManager.Complete(todoID)
	}

	return toolResult.Message, nil
}

// truncateCommand trunca comando para exibição
func (h *ExecuteHandler) truncateCommand(command string) string {
	if len(command) > 50 {
		return command[:47] + "..."
	}
	return command
}

// isDangerousCommand verifica se comando é perigoso
func (h *ExecuteHandler) isDangerousCommand(command string) bool {
	dangerousPatterns := []string{
		"rm -rf",
		"rm -fr",
		"mkfs",
		"dd if=",
		":(){ :|:& };:",
		"> /dev/",
		"chmod -R 777",
		"chown -R",
	}

	commandLower := strings.ToLower(command)

	for _, pattern := range dangerousPatterns {
		if strings.Contains(commandLower, strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}
