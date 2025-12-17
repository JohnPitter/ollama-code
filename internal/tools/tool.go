package tools

import "context"

// Tool interface para todas as ferramentas
type Tool interface {
	// Name retorna o nome da ferramenta
	Name() string

	// Description retorna descrição da ferramenta
	Description() string

	// Execute executa a ferramenta
	Execute(ctx context.Context, params map[string]interface{}) (Result, error)

	// RequiresConfirmation indica se requer confirmação
	RequiresConfirmation() bool
}

// Result resultado da execução de ferramenta
type Result struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
	Message string                 `json:"message,omitempty"`
}

// NewSuccessResult cria resultado de sucesso
func NewSuccessResult(message string, data map[string]interface{}) Result {
	return Result{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// NewErrorResult cria resultado de erro
func NewErrorResult(err error) Result {
	return Result{
		Success: false,
		Error:   err.Error(),
	}
}
