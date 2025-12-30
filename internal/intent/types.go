package intent

// Intent tipo de intenção detectada
type Intent string

const (
	// IntentReadFile ler arquivo(s)
	IntentReadFile Intent = "read_file"

	// IntentWriteFile criar/editar arquivo
	IntentWriteFile Intent = "write_file"

	// IntentExecuteCommand executar comando shell
	IntentExecuteCommand Intent = "execute_command"

	// IntentSearchCode buscar código no projeto
	IntentSearchCode Intent = "search_code"

	// IntentAnalyzeProject analisar estrutura do projeto
	IntentAnalyzeProject Intent = "analyze_project"

	// IntentGitOperation operação git
	IntentGitOperation Intent = "git_operation"

	// IntentWebSearch pesquisa na internet
	IntentWebSearch Intent = "web_search"

	// IntentQuestion apenas responder pergunta
	IntentQuestion Intent = "question"

	// IntentUnknown intenção desconhecida
	IntentUnknown Intent = "unknown"
)

// DetectionResult resultado da detecção
type DetectionResult struct {
	Intent      Intent                 `json:"intent"`
	Confidence  float64                `json:"confidence"`
	Parameters  map[string]interface{} `json:"parameters"`
	Reasoning   string                 `json:"reasoning"`
	UserMessage string                 `json:"user_message,omitempty"`
}
