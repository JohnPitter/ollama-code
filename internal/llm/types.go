package llm

// Message representa uma mensagem na conversa
type Message struct {
	Role    string `json:"role"`    // "user", "assistant", "system"
	Content string `json:"content"` // Conteúdo da mensagem
}

// Request estrutura da requisição para Ollama
type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
	Options  Options   `json:"options,omitempty"`
}

// Options opções da requisição
type Options struct {
	Temperature float64 `json:"temperature,omitempty"`
	NumPredict  int     `json:"num_predict,omitempty"`
	TopK        int     `json:"top_k,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
}

// Response estrutura da resposta do Ollama
type Response struct {
	Model     string  `json:"model"`
	CreatedAt string  `json:"created_at"`
	Message   Message `json:"message"`
	Done      bool    `json:"done"`
}

// CompletionOptions opções para completar
type CompletionOptions struct {
	Temperature float64
	MaxTokens   int
	SystemPrompt string
}
