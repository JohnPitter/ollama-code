package subagent

import (
	"context"
	"time"
)

// AgentType define o tipo de subagent
type AgentType string

const (
	// AgentTypeExplore - Especializado em busca e exploração de código
	AgentTypeExplore AgentType = "Explore"

	// AgentTypePlan - Especializado em planejamento e análise
	AgentTypePlan AgentType = "Plan"

	// AgentTypeExecute - Especializado em execução de tarefas
	AgentTypeExecute AgentType = "Execute"

	// AgentTypeGeneral - Agent genérico sem especialização
	AgentTypeGeneral AgentType = "General"
)

// AgentStatus representa o status de um subagent
type AgentStatus string

const (
	StatusPending   AgentStatus = "pending"
	StatusRunning   AgentStatus = "running"
	StatusCompleted AgentStatus = "completed"
	StatusFailed    AgentStatus = "failed"
	StatusTimeout   AgentStatus = "timeout"
	StatusKilled    AgentStatus = "killed"
)

// Subagent representa um subagent em execução
type Subagent struct {
	ID          string
	Type        AgentType
	Prompt      string
	Model       string
	Status      AgentStatus
	Result      string
	Error       error
	CreatedAt   time.Time
	StartedAt   time.Time
	CompletedAt time.Time

	// Context isolation
	WorkDir     string
	MaxTokens   int
	Temperature float64
	Timeout     time.Duration

	// Resource limits
	MaxMemoryMB int
	MaxCPUCores int

	// Communication
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

// AgentConfig configuração para criar um subagent
type AgentConfig struct {
	Type        AgentType
	Prompt      string
	Model       string
	WorkDir     string
	MaxTokens   int
	Temperature float64
	Timeout     time.Duration
	MaxMemoryMB int
	MaxCPUCores int
}

// DefaultConfig retorna configuração padrão para um agent type
func DefaultConfig(agentType AgentType) AgentConfig {
	cfg := AgentConfig{
		Type:        agentType,
		WorkDir:     ".",
		MaxTokens:   4096,
		Temperature: 0.7,
		Timeout:     5 * time.Minute,
		MaxMemoryMB: 512,
		MaxCPUCores: 1,
	}

	// Configurações específicas por tipo
	switch agentType {
	case AgentTypeExplore:
		cfg.Model = "qwen2.5-coder:1.5b" // Modelo rápido para exploração
		cfg.MaxTokens = 2048
		cfg.Timeout = 2 * time.Minute

	case AgentTypePlan:
		cfg.Model = "qwen2.5-coder:7b" // Modelo preciso para planejamento
		cfg.MaxTokens = 8192
		cfg.Timeout = 10 * time.Minute

	case AgentTypeExecute:
		cfg.Model = "qwen2.5-coder:7b" // Modelo preciso para execução
		cfg.MaxTokens = 4096
		cfg.Timeout = 15 * time.Minute
		cfg.MaxMemoryMB = 1024

	case AgentTypeGeneral:
		cfg.Model = "qwen2.5-coder:7b"
		cfg.MaxTokens = 4096
		cfg.Timeout = 5 * time.Minute
	}

	return cfg
}

// IsValid verifica se o agent type é válido
func (at AgentType) IsValid() bool {
	switch at {
	case AgentTypeExplore, AgentTypePlan, AgentTypeExecute, AgentTypeGeneral:
		return true
	default:
		return false
	}
}

// String retorna representação em string do agent type
func (at AgentType) String() string {
	return string(at)
}

// IsTerminal verifica se o status é terminal (não vai mudar mais)
func (s AgentStatus) IsTerminal() bool {
	switch s {
	case StatusCompleted, StatusFailed, StatusTimeout, StatusKilled:
		return true
	default:
		return false
	}
}

// IsSuccess verifica se o agent completou com sucesso
func (s *Subagent) IsSuccess() bool {
	return s.Status == StatusCompleted && s.Error == nil
}

// Duration retorna a duração da execução do agent
func (s *Subagent) Duration() time.Duration {
	if s.StartedAt.IsZero() {
		return 0
	}

	if s.CompletedAt.IsZero() {
		return time.Since(s.StartedAt)
	}

	return s.CompletedAt.Sub(s.StartedAt)
}
