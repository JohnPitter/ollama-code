package multimodel

// TaskType representa tipos de tarefas que podem usar modelos diferentes
type TaskType string

const (
	// TaskTypeIntent - Detecção de intenções (modelo rápido)
	TaskTypeIntent TaskType = "intent"

	// TaskTypeCode - Geração de código (modelo preciso)
	TaskTypeCode TaskType = "code"

	// TaskTypeSearch - Web search e summarization (modelo balanceado)
	TaskTypeSearch TaskType = "search"

	// TaskTypeAnalysis - Análise de código (modelo preciso)
	TaskTypeAnalysis TaskType = "analysis"

	// TaskTypeDefault - Tarefas gerais (modelo padrão)
	TaskTypeDefault TaskType = "default"
)

// IsValid verifica se o task type é válido
func (t TaskType) IsValid() bool {
	switch t {
	case TaskTypeIntent, TaskTypeCode, TaskTypeSearch, TaskTypeAnalysis, TaskTypeDefault:
		return true
	default:
		return false
	}
}

// String retorna representação em string
func (t TaskType) String() string {
	return string(t)
}

// ModelSpec especificação de um modelo
type ModelSpec struct {
	Name        string  // Nome do modelo (ex: "qwen2.5-coder:1.5b")
	MaxTokens   int     // Tokens máximos
	Temperature float64 // Temperatura
	Description string  // Descrição do propósito
}
