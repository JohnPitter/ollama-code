package subagent

import (
	"context"
	"fmt"
	"strings"

	"github.com/johnpitter/ollama-code/internal/llm"
)

// Executor executa subagents usando LLM
type Executor struct {
	llmClient *llm.Client
	ollamaURL string
}

// NewExecutor cria novo executor
func NewExecutor(ollamaURL string) *Executor {
	return &Executor{
		ollamaURL: ollamaURL,
	}
}

// Execute executa um subagent
func (e *Executor) Execute(ctx context.Context, agent *Subagent) (string, error) {
	// Criar LLM client específico para este agent (com modelo customizado)
	client := llm.NewClient(e.ollamaURL, agent.Model)

	// Construir prompt baseado no tipo de agent
	prompt := e.buildPrompt(agent)

	// Construir opções de completion
	opts := &llm.CompletionOptions{
		Temperature: agent.Temperature,
		MaxTokens:   agent.MaxTokens,
	}

	// Executar com LLM
	messages := []llm.Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	// Executar com context do agent (para timeout/cancel)
	response, err := client.Complete(ctx, messages, opts)
	if err != nil {
		return "", fmt.Errorf("llm execution failed: %w", err)
	}

	// Post-processar resposta baseado no tipo
	result := e.postProcess(agent, response)

	return result, nil
}

// buildPrompt constrói prompt baseado no tipo de agent
func (e *Executor) buildPrompt(agent *Subagent) string {
	var sb strings.Builder

	// System prompt baseado no tipo
	switch agent.Type {
	case AgentTypeExplore:
		sb.WriteString("You are an Explore agent specialized in code exploration and search.\n")
		sb.WriteString("Your task is to:\n")
		sb.WriteString("1. Analyze the codebase thoroughly\n")
		sb.WriteString("2. Find relevant files, functions, and patterns\n")
		sb.WriteString("3. Provide concise summaries of findings\n")
		sb.WriteString("4. Focus on accuracy and completeness\n\n")

	case AgentTypePlan:
		sb.WriteString("You are a Plan agent specialized in planning and architectural analysis.\n")
		sb.WriteString("Your task is to:\n")
		sb.WriteString("1. Break down complex tasks into steps\n")
		sb.WriteString("2. Analyze dependencies and requirements\n")
		sb.WriteString("3. Provide clear, actionable plans\n")
		sb.WriteString("4. Consider edge cases and potential issues\n\n")

	case AgentTypeExecute:
		sb.WriteString("You are an Execute agent specialized in task execution.\n")
		sb.WriteString("Your task is to:\n")
		sb.WriteString("1. Execute tasks efficiently\n")
		sb.WriteString("2. Provide clear status updates\n")
		sb.WriteString("3. Handle errors gracefully\n")
		sb.WriteString("4. Return concrete results\n\n")

	case AgentTypeGeneral:
		sb.WriteString("You are a general-purpose coding assistant.\n\n")
	}

	// Context do agent
	if agent.WorkDir != "" && agent.WorkDir != "." {
		sb.WriteString(fmt.Sprintf("Working directory: %s\n\n", agent.WorkDir))
	}

	// Task prompt
	sb.WriteString("Task:\n")
	sb.WriteString(agent.Prompt)
	sb.WriteString("\n\n")

	// Instruções finais
	sb.WriteString("Provide a clear, concise response focused on the task above.\n")

	return sb.String()
}

// postProcess pós-processa a resposta do LLM
func (e *Executor) postProcess(agent *Subagent, response string) string {
	// Remover espaços extras
	result := strings.TrimSpace(response)

	// Para agents Explore, tentar estruturar melhor
	if agent.Type == AgentTypeExplore {
		// Se a resposta não tem marcadores, adicionar estrutura básica
		if !strings.Contains(result, "## ") && !strings.Contains(result, "- ") {
			lines := strings.Split(result, "\n")
			if len(lines) > 3 {
				// Adicionar cabeçalho
				result = "## Exploration Results\n\n" + result
			}
		}
	}

	// Para agents Plan, garantir que tem estrutura de steps
	if agent.Type == AgentTypePlan {
		if !strings.Contains(result, "Step ") && !strings.Contains(result, "1.") {
			// Tentar adicionar numeração básica se não tiver
			lines := strings.Split(result, "\n")
			if len(lines) > 2 {
				var numbered strings.Builder
				numbered.WriteString("## Plan\n\n")
				step := 1
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line == "" {
						numbered.WriteString("\n")
						continue
					}
					if !strings.HasPrefix(line, "#") {
						numbered.WriteString(fmt.Sprintf("%d. %s\n", step, line))
						step++
					} else {
						numbered.WriteString(line + "\n")
					}
				}
				result = numbered.String()
			}
		}
	}

	return result
}

// CreateExecutorFunc cria uma ExecutorFunc a partir deste executor
func (e *Executor) CreateExecutorFunc() ExecutorFunc {
	return func(ctx context.Context, agent *Subagent) (string, error) {
		return e.Execute(ctx, agent)
	}
}
