package skills

import (
	"context"
	"fmt"
)

// Skill representa uma habilidade especializada do agente
type Skill interface {
	// Name retorna o nome único do skill
	Name() string

	// Description retorna descrição do que o skill faz
	Description() string

	// Capabilities retorna lista de capacidades que o skill oferece
	Capabilities() []string

	// CanHandle verifica se o skill pode processar uma tarefa
	CanHandle(ctx context.Context, task Task) bool

	// Execute executa o skill com os parâmetros fornecidos
	Execute(ctx context.Context, task Task) (*Result, error)

	// Examples retorna exemplos de uso do skill
	Examples() []string
}

// Task representa uma tarefa a ser executada por um skill
type Task struct {
	Type        string                 // Tipo da tarefa (ex: "api_call", "cloud_deploy")
	Description string                 // Descrição da tarefa
	Parameters  map[string]interface{} // Parâmetros da tarefa
	Context     map[string]interface{} // Contexto adicional
}

// Result representa o resultado da execução de um skill
type Result struct {
	Success bool                   // Se a execução foi bem-sucedida
	Data    map[string]interface{} // Dados retornados
	Message string                 // Mensagem descritiva
	Error   string                 // Mensagem de erro (se houver)
	Metrics Metrics                // Métricas da execução
}

// Metrics métricas de execução
type Metrics struct {
	ExecutionTime int64  // Tempo de execução em ms
	TokensUsed    int    // Tokens usados (se aplicável)
	APICallsMade  int    // Chamadas API feitas
	CacheHits     int    // Hits de cache
	SkillsInvoked []string // Skills invocados durante execução
}

// BaseSkill implementação base com funcionalidades comuns
type BaseSkill struct {
	name         string
	description  string
	capabilities []string
	examples     []string
}

// NewBaseSkill cria um novo BaseSkill
func NewBaseSkill(name, description string, capabilities, examples []string) *BaseSkill {
	return &BaseSkill{
		name:         name,
		description:  description,
		capabilities: capabilities,
		examples:     examples,
	}
}

// Name implementa Skill.Name
func (b *BaseSkill) Name() string {
	return b.name
}

// Description implementa Skill.Description
func (b *BaseSkill) Description() string {
	return b.description
}

// Capabilities implementa Skill.Capabilities
func (b *BaseSkill) Capabilities() []string {
	return b.capabilities
}

// Examples implementa Skill.Examples
func (b *BaseSkill) Examples() []string {
	return b.examples
}

// CanHandle implementação padrão (deve ser sobrescrita)
func (b *BaseSkill) CanHandle(ctx context.Context, task Task) bool {
	return false
}

// Execute implementação padrão (deve ser sobrescrita)
func (b *BaseSkill) Execute(ctx context.Context, task Task) (*Result, error) {
	return nil, fmt.Errorf("skill %s não implementou Execute()", b.name)
}
