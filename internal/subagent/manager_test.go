package subagent

import (
	"context"
	"errors"
	"testing"
	"time"
)

// Mock executor for testing
func mockExecutorSuccess(ctx context.Context, agent *Subagent) (string, error) {
	// Simulate some work
	time.Sleep(10 * time.Millisecond)
	return "success result", nil
}

func mockExecutorFail(ctx context.Context, agent *Subagent) (string, error) {
	time.Sleep(10 * time.Millisecond)
	return "", errors.New("execution failed")
}

func mockExecutorSlow(ctx context.Context, agent *Subagent) (string, error) {
	// Simulate slow execution (will be cancelled by timeout)
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(10 * time.Second):
		return "slow result", nil
	}
}

// TestNewManager testa criação do manager
func TestNewManager(t *testing.T) {
	executor := mockExecutorSuccess
	manager := NewManager(executor)

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	if manager.maxConcurrent != 5 {
		t.Errorf("Expected maxConcurrent=5, got %d", manager.maxConcurrent)
	}

	if len(manager.agents) != 0 {
		t.Errorf("Expected empty agents map, got %d agents", len(manager.agents))
	}
}

// TestSpawn_Success testa spawn de agent com sucesso
func TestSpawn_Success(t *testing.T) {
	manager := NewManager(mockExecutorSuccess)

	cfg := DefaultConfig(AgentTypeExplore)
	cfg.Prompt = "test prompt"

	agent, err := manager.Spawn(cfg)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	if agent == nil {
		t.Fatal("Spawn returned nil agent")
	}

	if agent.ID == "" {
		t.Error("Agent ID is empty")
	}

	if agent.Type != AgentTypeExplore {
		t.Errorf("Expected type Explore, got %s", agent.Type)
	}

	if agent.Status != StatusPending && agent.Status != StatusRunning {
		t.Errorf("Expected status pending or running, got %s", agent.Status)
	}
}

// TestSpawn_InvalidType testa spawn com tipo inválido
func TestSpawn_InvalidType(t *testing.T) {
	manager := NewManager(mockExecutorSuccess)

	cfg := AgentConfig{
		Type:   AgentType("Invalid"),
		Prompt: "test",
	}

	_, err := manager.Spawn(cfg)
	if err == nil {
		t.Error("Expected error for invalid agent type")
	}

	if manager.totalSpawned != 0 {
		t.Errorf("totalSpawned should be 0, got %d", manager.totalSpawned)
	}
}

// TestSpawn_MaxConcurrentReached testa limite de agents simultâneos
func TestSpawn_MaxConcurrentReached(t *testing.T) {
	// Usar executor lento para manter agents ativos
	manager := NewManager(mockExecutorSlow)
	manager.SetMaxConcurrent(2)

	cfg := DefaultConfig(AgentTypeGeneral)
	cfg.Prompt = "test"
	cfg.Timeout = 5 * time.Second

	// Spawn 2 agents (limite)
	agent1, err := manager.Spawn(cfg)
	if err != nil {
		t.Fatalf("First spawn failed: %v", err)
	}

	agent2, err := manager.Spawn(cfg)
	if err != nil {
		t.Fatalf("Second spawn failed: %v", err)
	}

	// Terceiro deve falhar
	_, err = manager.Spawn(cfg)
	if err == nil {
		t.Error("Expected error when max concurrent reached")
	}

	// Cleanup
	manager.Kill(agent1.ID)
	manager.Kill(agent2.ID)
}

// TestWait_Success testa aguardar conclusão com sucesso
func TestWait_Success(t *testing.T) {
	manager := NewManager(mockExecutorSuccess)

	cfg := DefaultConfig(AgentTypeGeneral)
	cfg.Prompt = "test prompt"

	agent, err := manager.Spawn(cfg)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	result, err := manager.Wait(agent.ID)
	if err != nil {
		t.Errorf("Wait failed: %v", err)
	}

	if result != "success result" {
		t.Errorf("Expected 'success result', got '%s'", result)
	}

	// Verificar status final
	finalAgent, _ := manager.Get(agent.ID)
	if finalAgent.Status != StatusCompleted {
		t.Errorf("Expected status completed, got %s", finalAgent.Status)
	}
}

// TestWait_Failure testa aguardar conclusão com falha
func TestWait_Failure(t *testing.T) {
	manager := NewManager(mockExecutorFail)

	cfg := DefaultConfig(AgentTypeGeneral)
	cfg.Prompt = "test prompt"

	agent, err := manager.Spawn(cfg)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	_, err = manager.Wait(agent.ID)
	if err == nil {
		t.Error("Expected error from failed agent")
	}

	// Verificar status final
	finalAgent, _ := manager.Get(agent.ID)
	if finalAgent.Status != StatusFailed {
		t.Errorf("Expected status failed, got %s", finalAgent.Status)
	}
}

// TestWait_NotFound testa aguardar agent inexistente
func TestWait_NotFound(t *testing.T) {
	manager := NewManager(mockExecutorSuccess)

	_, err := manager.Wait("nonexistent-id")
	if err == nil {
		t.Error("Expected error for nonexistent agent")
	}
}

// TestWaitWithTimeout_Success testa wait com timeout
func TestWaitWithTimeout_Success(t *testing.T) {
	manager := NewManager(mockExecutorSuccess)

	cfg := DefaultConfig(AgentTypeGeneral)
	cfg.Prompt = "test"

	agent, err := manager.Spawn(cfg)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	result, err := manager.WaitWithTimeout(agent.ID, 1*time.Second)
	if err != nil {
		t.Errorf("WaitWithTimeout failed: %v", err)
	}

	if result != "success result" {
		t.Errorf("Expected 'success result', got '%s'", result)
	}
}

// TestWaitWithTimeout_Timeout testa timeout no wait customizado
func TestWaitWithTimeout_Timeout(t *testing.T) {
	manager := NewManager(mockExecutorSlow)

	cfg := DefaultConfig(AgentTypeGeneral)
	cfg.Prompt = "test"
	cfg.Timeout = 10 * time.Second // Agent timeout longo

	agent, err := manager.Spawn(cfg)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	// Wait com timeout curto (vai timeout antes do agent terminar)
	_, err = manager.WaitWithTimeout(agent.ID, 50*time.Millisecond)
	if err == nil {
		t.Error("Expected timeout error")
	}

	// Cleanup
	manager.Kill(agent.ID)
}

// TestKill_Running testa matar agent em execução
func TestKill_Running(t *testing.T) {
	manager := NewManager(mockExecutorSlow)

	cfg := DefaultConfig(AgentTypeGeneral)
	cfg.Prompt = "test"
	cfg.Timeout = 10 * time.Second

	agent, err := manager.Spawn(cfg)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	// Aguardar um pouco para garantir que está running
	time.Sleep(20 * time.Millisecond)

	err = manager.Kill(agent.ID)
	if err != nil {
		t.Errorf("Kill failed: %v", err)
	}

	// Verificar status final
	finalAgent, _ := manager.Get(agent.ID)
	if finalAgent.Status != StatusKilled {
		t.Errorf("Expected status killed, got %s", finalAgent.Status)
	}
}

// TestKill_AlreadyCompleted testa matar agent já completo
func TestKill_AlreadyCompleted(t *testing.T) {
	manager := NewManager(mockExecutorSuccess)

	cfg := DefaultConfig(AgentTypeGeneral)
	cfg.Prompt = "test"

	agent, err := manager.Spawn(cfg)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	// Aguardar conclusão
	manager.Wait(agent.ID)

	// Tentar matar agent já completo
	err = manager.Kill(agent.ID)
	if err == nil {
		t.Error("Expected error when killing completed agent")
	}
}

// TestGet testa buscar agent por ID
func TestGet(t *testing.T) {
	manager := NewManager(mockExecutorSuccess)

	cfg := DefaultConfig(AgentTypeGeneral)
	cfg.Prompt = "test"

	agent, err := manager.Spawn(cfg)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	retrieved, err := manager.Get(agent.ID)
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}

	if retrieved.ID != agent.ID {
		t.Errorf("Expected ID %s, got %s", agent.ID, retrieved.ID)
	}
}

// TestList testa listar todos os agents
func TestList(t *testing.T) {
	manager := NewManager(mockExecutorSuccess)

	// Spawn 3 agents
	for i := 0; i < 3; i++ {
		cfg := DefaultConfig(AgentTypeGeneral)
		cfg.Prompt = "test"
		manager.Spawn(cfg)
	}

	agents := manager.List()
	if len(agents) != 3 {
		t.Errorf("Expected 3 agents, got %d", len(agents))
	}
}

// TestListByStatus testa filtrar agents por status
func TestListByStatus(t *testing.T) {
	manager := NewManager(mockExecutorSuccess)

	// Spawn e aguardar conclusão de 2 agents
	for i := 0; i < 2; i++ {
		cfg := DefaultConfig(AgentTypeGeneral)
		cfg.Prompt = "test"
		agent, _ := manager.Spawn(cfg)
		manager.Wait(agent.ID)
	}

	// Spawn 1 agent lento (vai ficar running)
	cfg := DefaultConfig(AgentTypeGeneral)
	cfg.Prompt = "test"
	cfg.Timeout = 10 * time.Second
	slowManager := NewManager(mockExecutorSlow)
	slowAgent, _ := slowManager.Spawn(cfg)

	// Aguardar para garantir que está running
	time.Sleep(20 * time.Millisecond)

	// Buscar completed
	completed := manager.ListByStatus(StatusCompleted)
	if len(completed) != 2 {
		t.Errorf("Expected 2 completed agents, got %d", len(completed))
	}

	// Cleanup
	slowManager.Kill(slowAgent.ID)
}

// TestListByType testa filtrar agents por tipo
func TestListByType(t *testing.T) {
	manager := NewManager(mockExecutorSuccess)

	// Spawn diferentes tipos
	types := []AgentType{AgentTypeExplore, AgentTypePlan, AgentTypeExplore}
	for _, agentType := range types {
		cfg := DefaultConfig(agentType)
		cfg.Prompt = "test"
		manager.Spawn(cfg)
	}

	explores := manager.ListByType(AgentTypeExplore)
	if len(explores) != 2 {
		t.Errorf("Expected 2 Explore agents, got %d", len(explores))
	}

	plans := manager.ListByType(AgentTypePlan)
	if len(plans) != 1 {
		t.Errorf("Expected 1 Plan agent, got %d", len(plans))
	}
}

// TestCleanup testa limpeza de agents antigos
func TestCleanup(t *testing.T) {
	manager := NewManager(mockExecutorSuccess)

	// Spawn e aguardar conclusão de 3 agents
	for i := 0; i < 3; i++ {
		cfg := DefaultConfig(AgentTypeGeneral)
		cfg.Prompt = "test"
		agent, _ := manager.Spawn(cfg)
		manager.Wait(agent.ID)
	}

	// Aguardar um pouco
	time.Sleep(100 * time.Millisecond)

	// Cleanup agents completados há mais de 50ms
	removed := manager.Cleanup(50 * time.Millisecond)
	if removed != 3 {
		t.Errorf("Expected to remove 3 agents, removed %d", removed)
	}

	// Verificar que foram removidos
	agents := manager.List()
	if len(agents) != 0 {
		t.Errorf("Expected 0 agents after cleanup, got %d", len(agents))
	}
}

// TestCleanup_OnlyTerminated testa que cleanup não remove agents ativos
func TestCleanup_OnlyTerminated(t *testing.T) {
	manager := NewManager(mockExecutorSlow)

	// Spawn agent lento
	cfg := DefaultConfig(AgentTypeGeneral)
	cfg.Prompt = "test"
	cfg.Timeout = 10 * time.Second
	agent, _ := manager.Spawn(cfg)

	// Aguardar para garantir que está running
	time.Sleep(20 * time.Millisecond)

	// Tentar cleanup
	removed := manager.Cleanup(0)
	if removed != 0 {
		t.Errorf("Expected 0 removals (agent still running), got %d", removed)
	}

	// Cleanup
	manager.Kill(agent.ID)
}

// TestClearAll testa limpar todos os agents
func TestClearAll(t *testing.T) {
	manager := NewManager(mockExecutorSuccess)

	// Spawn vários agents
	for i := 0; i < 5; i++ {
		cfg := DefaultConfig(AgentTypeGeneral)
		cfg.Prompt = "test"
		manager.Spawn(cfg)
	}

	manager.ClearAll()

	if len(manager.agents) != 0 {
		t.Errorf("Expected 0 agents after ClearAll, got %d", len(manager.agents))
	}

	if manager.activeAgents != 0 {
		t.Errorf("Expected 0 active agents, got %d", manager.activeAgents)
	}
}

// TestStats testa estatísticas do manager
func TestStats(t *testing.T) {
	manager := NewManager(mockExecutorSuccess)

	// Spawn e aguardar 2 agents com sucesso
	for i := 0; i < 2; i++ {
		cfg := DefaultConfig(AgentTypeGeneral)
		cfg.Prompt = "test"
		agent, _ := manager.Spawn(cfg)
		manager.Wait(agent.ID)
	}

	// Spawn 1 agent com falha
	failManager := NewManager(mockExecutorFail)
	cfg := DefaultConfig(AgentTypeGeneral)
	cfg.Prompt = "test"
	failAgent, _ := failManager.Spawn(cfg)
	failManager.Wait(failAgent.ID)

	stats := manager.Stats()

	if stats["total_spawned"].(int) != 2 {
		t.Errorf("Expected total_spawned=2, got %v", stats["total_spawned"])
	}

	if stats["total_completed"].(int) != 2 {
		t.Errorf("Expected total_completed=2, got %v", stats["total_completed"])
	}

	if stats["success_rate"].(float64) != 100.0 {
		t.Errorf("Expected success_rate=100.0, got %v", stats["success_rate"])
	}

	// Check fail manager stats
	failStats := failManager.Stats()
	if failStats["total_failed"].(int) != 1 {
		t.Errorf("Expected total_failed=1, got %v", failStats["total_failed"])
	}

	if failStats["success_rate"].(float64) != 0.0 {
		t.Errorf("Expected success_rate=0.0, got %v", failStats["success_rate"])
	}
}

// TestSetMaxConcurrent testa configurar limite de concorrência
func TestSetMaxConcurrent(t *testing.T) {
	manager := NewManager(mockExecutorSuccess)

	manager.SetMaxConcurrent(10)
	if manager.maxConcurrent != 10 {
		t.Errorf("Expected maxConcurrent=10, got %d", manager.maxConcurrent)
	}

	// Testar valor inválido (< 1)
	manager.SetMaxConcurrent(0)
	if manager.maxConcurrent != 1 {
		t.Errorf("Expected maxConcurrent=1 (minimum), got %d", manager.maxConcurrent)
	}
}

// TestAgentTimeout testa timeout de agent
func TestAgentTimeout(t *testing.T) {
	manager := NewManager(mockExecutorSlow)

	cfg := DefaultConfig(AgentTypeGeneral)
	cfg.Prompt = "test"
	cfg.Timeout = 50 * time.Millisecond // Timeout muito curto

	agent, err := manager.Spawn(cfg)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	// Aguardar timeout
	_, err = manager.Wait(agent.ID)
	if err == nil {
		t.Error("Expected timeout error")
	}

	// Verificar status
	finalAgent, _ := manager.Get(agent.ID)
	if finalAgent.Status != StatusTimeout {
		t.Errorf("Expected status timeout, got %s", finalAgent.Status)
	}
}

// TestConcurrentSpawns testa spawns concorrentes
func TestConcurrentSpawns(t *testing.T) {
	manager := NewManager(mockExecutorSuccess)
	manager.SetMaxConcurrent(10)

	// Spawn 5 agents em paralelo
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			cfg := DefaultConfig(AgentTypeGeneral)
			cfg.Prompt = "test"
			agent, err := manager.Spawn(cfg)
			if err != nil {
				t.Errorf("Concurrent spawn failed: %v", err)
			}
			manager.Wait(agent.ID)
			done <- true
		}()
	}

	// Aguardar todos completarem
	for i := 0; i < 5; i++ {
		<-done
	}

	stats := manager.Stats()
	if stats["total_spawned"].(int) != 5 {
		t.Errorf("Expected 5 spawned agents, got %v", stats["total_spawned"])
	}

	if stats["total_completed"].(int) != 5 {
		t.Errorf("Expected 5 completed agents, got %v", stats["total_completed"])
	}
}
