package subagent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Manager gerencia subagents
type Manager struct {
	agents map[string]*Subagent
	mu     sync.RWMutex

	// Executor function - será injetado pelo DI
	executor ExecutorFunc

	// Resource tracking
	activeAgents   int
	maxConcurrent  int
	totalSpawned   int
	totalCompleted int
	totalFailed    int
}

// ExecutorFunc função que executa um subagent
type ExecutorFunc func(ctx context.Context, agent *Subagent) (string, error)

// NewManager cria novo manager de subagents
func NewManager(executor ExecutorFunc) *Manager {
	return &Manager{
		agents:        make(map[string]*Subagent),
		executor:      executor,
		maxConcurrent: 5, // Limite de agents simultâneos
	}
}

// Spawn cria e inicia um novo subagent
func (m *Manager) Spawn(cfg AgentConfig) (*Subagent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validar tipo de agent
	if !cfg.Type.IsValid() {
		return nil, fmt.Errorf("invalid agent type: %s", cfg.Type)
	}

	// Verificar limite de agents simultâneos
	if m.activeAgents >= m.maxConcurrent {
		return nil, fmt.Errorf("max concurrent agents reached (%d)", m.maxConcurrent)
	}

	// Criar agent
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)

	agent := &Subagent{
		ID:          uuid.New().String(),
		Type:        cfg.Type,
		Prompt:      cfg.Prompt,
		Model:       cfg.Model,
		Status:      StatusPending,
		WorkDir:     cfg.WorkDir,
		MaxTokens:   cfg.MaxTokens,
		Temperature: cfg.Temperature,
		Timeout:     cfg.Timeout,
		MaxMemoryMB: cfg.MaxMemoryMB,
		MaxCPUCores: cfg.MaxCPUCores,
		CreatedAt:   time.Now(),
		ctx:         ctx,
		cancel:      cancel,
		done:        make(chan struct{}),
	}

	// Registrar agent
	m.agents[agent.ID] = agent
	m.activeAgents++
	m.totalSpawned++

	// Iniciar execução em goroutine
	go m.execute(agent)

	return agent, nil
}

// execute executa um subagent em background
func (m *Manager) execute(agent *Subagent) {
	defer close(agent.done)
	defer m.onAgentComplete(agent)

	// Marcar como running
	agent.Status = StatusRunning
	agent.StartedAt = time.Now()

	// Executar agent
	result, err := m.executor(agent.ctx, agent)

	// Verificar se foi cancelado/timeout
	select {
	case <-agent.ctx.Done():
		if agent.ctx.Err() == context.DeadlineExceeded {
			agent.Status = StatusTimeout
			agent.Error = fmt.Errorf("agent timeout after %v", agent.Timeout)
		} else {
			agent.Status = StatusKilled
			agent.Error = fmt.Errorf("agent killed")
		}
		return
	default:
	}

	// Processar resultado
	agent.CompletedAt = time.Now()
	if err != nil {
		agent.Status = StatusFailed
		agent.Error = err
		return
	}

	agent.Status = StatusCompleted
	agent.Result = result
}

// onAgentComplete chamado quando um agent completa
func (m *Manager) onAgentComplete(agent *Subagent) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.activeAgents--

	if agent.Status == StatusCompleted {
		m.totalCompleted++
	} else {
		m.totalFailed++
	}
}

// Wait aguarda conclusão de um agent
func (m *Manager) Wait(agentID string) (string, error) {
	agent, err := m.Get(agentID)
	if err != nil {
		return "", err
	}

	// Aguardar conclusão
	<-agent.done

	if agent.Error != nil {
		return "", agent.Error
	}

	return agent.Result, nil
}

// WaitWithTimeout aguarda conclusão com timeout customizado
func (m *Manager) WaitWithTimeout(agentID string, timeout time.Duration) (string, error) {
	agent, err := m.Get(agentID)
	if err != nil {
		return "", err
	}

	// Aguardar com timeout
	select {
	case <-agent.done:
		if agent.Error != nil {
			return "", agent.Error
		}
		return agent.Result, nil

	case <-time.After(timeout):
		return "", fmt.Errorf("wait timeout after %v", timeout)
	}
}

// Kill mata um agent em execução
func (m *Manager) Kill(agentID string) error {
	agent, err := m.Get(agentID)
	if err != nil {
		return err
	}

	// Verificar se já terminou
	if agent.Status.IsTerminal() {
		return fmt.Errorf("agent already terminated with status: %s", agent.Status)
	}

	// Cancelar context
	agent.cancel()

	// Aguardar conclusão
	<-agent.done

	return nil
}

// Get retorna um agent pelo ID
func (m *Manager) Get(agentID string) (*Subagent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agent, ok := m.agents[agentID]
	if !ok {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}

	return agent, nil
}

// List retorna todos os agents
func (m *Manager) List() []*Subagent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agents := make([]*Subagent, 0, len(m.agents))
	for _, agent := range m.agents {
		agents = append(agents, agent)
	}

	return agents
}

// ListByStatus retorna agents por status
func (m *Manager) ListByStatus(status AgentStatus) []*Subagent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agents := make([]*Subagent, 0)
	for _, agent := range m.agents {
		if agent.Status == status {
			agents = append(agents, agent)
		}
	}

	return agents
}

// ListByType retorna agents por tipo
func (m *Manager) ListByType(agentType AgentType) []*Subagent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agents := make([]*Subagent, 0)
	for _, agent := range m.agents {
		if agent.Type == agentType {
			agents = append(agents, agent)
		}
	}

	return agents
}

// Cleanup remove agents terminados antigos
func (m *Manager) Cleanup(olderThan time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	removed := 0
	cutoff := time.Now().Add(-olderThan)

	for id, agent := range m.agents {
		// Só remover agents terminados
		if !agent.Status.IsTerminal() {
			continue
		}

		// Só remover se completos antes do cutoff
		if agent.CompletedAt.Before(cutoff) {
			delete(m.agents, id)
			removed++
		}
	}

	return removed
}

// ClearAll remove todos os agents (útil para testes)
func (m *Manager) ClearAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.agents = make(map[string]*Subagent)
	m.activeAgents = 0
}

// Stats retorna estatísticas do manager
func (m *Manager) Stats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"total_agents":     len(m.agents),
		"active_agents":    m.activeAgents,
		"total_spawned":    m.totalSpawned,
		"total_completed":  m.totalCompleted,
		"total_failed":     m.totalFailed,
		"max_concurrent":   m.maxConcurrent,
		"success_rate":     m.getSuccessRate(),
		"pending_agents":   len(m.ListByStatus(StatusPending)),
		"running_agents":   len(m.ListByStatus(StatusRunning)),
		"completed_agents": len(m.ListByStatus(StatusCompleted)),
		"failed_agents":    len(m.ListByStatus(StatusFailed)),
		"timeout_agents":   len(m.ListByStatus(StatusTimeout)),
		"killed_agents":    len(m.ListByStatus(StatusKilled)),
	}
}

// getSuccessRate calcula taxa de sucesso
func (m *Manager) getSuccessRate() float64 {
	total := m.totalCompleted + m.totalFailed
	if total == 0 {
		return 0.0
	}

	return float64(m.totalCompleted) / float64(total) * 100
}

// SetMaxConcurrent define número máximo de agents simultâneos
func (m *Manager) SetMaxConcurrent(max int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if max < 1 {
		max = 1
	}

	m.maxConcurrent = max
}
