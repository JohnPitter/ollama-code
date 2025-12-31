package subagent

import (
	"testing"
	"time"
)

// TestAgentType_IsValid testa validação de tipos de agent
func TestAgentType_IsValid(t *testing.T) {
	testCases := []struct {
		name     string
		agentType AgentType
		expected bool
	}{
		{"Explore is valid", AgentTypeExplore, true},
		{"Plan is valid", AgentTypePlan, true},
		{"Execute is valid", AgentTypeExecute, true},
		{"General is valid", AgentTypeGeneral, true},
		{"Invalid type", AgentType("Invalid"), false},
		{"Empty type", AgentType(""), false},
		{"Random string", AgentType("RandomString"), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.agentType.IsValid()
			if result != tc.expected {
				t.Errorf("Expected IsValid()=%v for %s, got %v", tc.expected, tc.agentType, result)
			}
		})
	}
}

// TestAgentType_String testa conversão para string
func TestAgentType_String(t *testing.T) {
	testCases := []struct {
		agentType AgentType
		expected  string
	}{
		{AgentTypeExplore, "Explore"},
		{AgentTypePlan, "Plan"},
		{AgentTypeExecute, "Execute"},
		{AgentTypeGeneral, "General"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := tc.agentType.String()
			if result != tc.expected {
				t.Errorf("Expected String()=%s, got %s", tc.expected, result)
			}
		})
	}
}

// TestAgentStatus_IsTerminal testa se status é terminal
func TestAgentStatus_IsTerminal(t *testing.T) {
	testCases := []struct {
		name     string
		status   AgentStatus
		expected bool
	}{
		{"Pending is not terminal", StatusPending, false},
		{"Running is not terminal", StatusRunning, false},
		{"Completed is terminal", StatusCompleted, true},
		{"Failed is terminal", StatusFailed, true},
		{"Timeout is terminal", StatusTimeout, true},
		{"Killed is terminal", StatusKilled, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.status.IsTerminal()
			if result != tc.expected {
				t.Errorf("Expected IsTerminal()=%v for %s, got %v", tc.expected, tc.status, result)
			}
		})
	}
}

// TestSubagent_IsSuccess testa verificação de sucesso
func TestSubagent_IsSuccess(t *testing.T) {
	testCases := []struct {
		name     string
		agent    *Subagent
		expected bool
	}{
		{
			name: "Completed without error is success",
			agent: &Subagent{
				Status: StatusCompleted,
				Error:  nil,
			},
			expected: true,
		},
		{
			name: "Completed with error is not success",
			agent: &Subagent{
				Status: StatusCompleted,
				Error:  &mockError{},
			},
			expected: false,
		},
		{
			name: "Failed is not success",
			agent: &Subagent{
				Status: StatusFailed,
				Error:  nil,
			},
			expected: false,
		},
		{
			name: "Running is not success",
			agent: &Subagent{
				Status: StatusRunning,
				Error:  nil,
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.agent.IsSuccess()
			if result != tc.expected {
				t.Errorf("Expected IsSuccess()=%v, got %v", tc.expected, result)
			}
		})
	}
}

// mockError é um erro mock para testes
type mockError struct{}

func (e *mockError) Error() string {
	return "mock error"
}

// TestSubagent_Duration testa cálculo de duração
func TestSubagent_Duration(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name     string
		agent    *Subagent
		expected time.Duration
	}{
		{
			name: "Not started yet",
			agent: &Subagent{
				StartedAt:   time.Time{},
				CompletedAt: time.Time{},
			},
			expected: 0,
		},
		{
			name: "Started but not completed",
			agent: &Subagent{
				StartedAt:   now.Add(-1 * time.Second),
				CompletedAt: time.Time{},
			},
			// Duration should be approximately 1 second (will be slightly more due to test execution time)
			expected: 900 * time.Millisecond, // Allow some tolerance
		},
		{
			name: "Completed",
			agent: &Subagent{
				StartedAt:   now.Add(-2 * time.Second),
				CompletedAt: now,
			},
			expected: 2 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			duration := tc.agent.Duration()

			if tc.name == "Not started yet" {
				if duration != tc.expected {
					t.Errorf("Expected duration=%v, got %v", tc.expected, duration)
				}
			} else if tc.name == "Started but not completed" {
				// For running agents, duration should be at least the expected value
				if duration < tc.expected {
					t.Errorf("Expected duration >= %v, got %v", tc.expected, duration)
				}
			} else {
				// For completed, should be exact (with small tolerance)
				diff := duration - tc.expected
				if diff < 0 {
					diff = -diff
				}
				if diff > 10*time.Millisecond {
					t.Errorf("Expected duration=%v, got %v (diff=%v)", tc.expected, duration, diff)
				}
			}
		})
	}
}

// TestDefaultConfig_Explore testa configuração padrão para Explore
func TestDefaultConfig_Explore(t *testing.T) {
	cfg := DefaultConfig(AgentTypeExplore)

	if cfg.Type != AgentTypeExplore {
		t.Errorf("Expected Type=Explore, got %s", cfg.Type)
	}

	if cfg.Model != "qwen2.5-coder:1.5b" {
		t.Errorf("Expected Model=qwen2.5-coder:1.5b, got %s", cfg.Model)
	}

	if cfg.MaxTokens != 2048 {
		t.Errorf("Expected MaxTokens=2048, got %d", cfg.MaxTokens)
	}

	if cfg.Timeout != 2*time.Minute {
		t.Errorf("Expected Timeout=2m, got %v", cfg.Timeout)
	}

	if cfg.Temperature != 0.7 {
		t.Errorf("Expected Temperature=0.7, got %f", cfg.Temperature)
	}

	if cfg.MaxMemoryMB != 512 {
		t.Errorf("Expected MaxMemoryMB=512, got %d", cfg.MaxMemoryMB)
	}

	if cfg.MaxCPUCores != 1 {
		t.Errorf("Expected MaxCPUCores=1, got %d", cfg.MaxCPUCores)
	}

	if cfg.WorkDir != "." {
		t.Errorf("Expected WorkDir='.', got %s", cfg.WorkDir)
	}
}

// TestDefaultConfig_Plan testa configuração padrão para Plan
func TestDefaultConfig_Plan(t *testing.T) {
	cfg := DefaultConfig(AgentTypePlan)

	if cfg.Type != AgentTypePlan {
		t.Errorf("Expected Type=Plan, got %s", cfg.Type)
	}

	if cfg.Model != "qwen2.5-coder:7b" {
		t.Errorf("Expected Model=qwen2.5-coder:7b, got %s", cfg.Model)
	}

	if cfg.MaxTokens != 8192 {
		t.Errorf("Expected MaxTokens=8192, got %d", cfg.MaxTokens)
	}

	if cfg.Timeout != 10*time.Minute {
		t.Errorf("Expected Timeout=10m, got %v", cfg.Timeout)
	}
}

// TestDefaultConfig_Execute testa configuração padrão para Execute
func TestDefaultConfig_Execute(t *testing.T) {
	cfg := DefaultConfig(AgentTypeExecute)

	if cfg.Type != AgentTypeExecute {
		t.Errorf("Expected Type=Execute, got %s", cfg.Type)
	}

	if cfg.Model != "qwen2.5-coder:7b" {
		t.Errorf("Expected Model=qwen2.5-coder:7b, got %s", cfg.Model)
	}

	if cfg.MaxTokens != 4096 {
		t.Errorf("Expected MaxTokens=4096, got %d", cfg.MaxTokens)
	}

	if cfg.Timeout != 15*time.Minute {
		t.Errorf("Expected Timeout=15m, got %v", cfg.Timeout)
	}

	if cfg.MaxMemoryMB != 1024 {
		t.Errorf("Expected MaxMemoryMB=1024, got %d", cfg.MaxMemoryMB)
	}
}

// TestDefaultConfig_General testa configuração padrão para General
func TestDefaultConfig_General(t *testing.T) {
	cfg := DefaultConfig(AgentTypeGeneral)

	if cfg.Type != AgentTypeGeneral {
		t.Errorf("Expected Type=General, got %s", cfg.Type)
	}

	if cfg.Model != "qwen2.5-coder:7b" {
		t.Errorf("Expected Model=qwen2.5-coder:7b, got %s", cfg.Model)
	}

	if cfg.MaxTokens != 4096 {
		t.Errorf("Expected MaxTokens=4096, got %d", cfg.MaxTokens)
	}

	if cfg.Timeout != 5*time.Minute {
		t.Errorf("Expected Timeout=5m, got %v", cfg.Timeout)
	}
}

// TestDefaultConfig_AllTypes testa que todos os tipos têm config padrão
func TestDefaultConfig_AllTypes(t *testing.T) {
	types := []AgentType{
		AgentTypeExplore,
		AgentTypePlan,
		AgentTypeExecute,
		AgentTypeGeneral,
	}

	for _, agentType := range types {
		t.Run(string(agentType), func(t *testing.T) {
			cfg := DefaultConfig(agentType)

			// Verificar campos comuns
			if cfg.Type != agentType {
				t.Errorf("Expected Type=%s, got %s", agentType, cfg.Type)
			}

			if cfg.Model == "" {
				t.Error("Model should not be empty")
			}

			if cfg.MaxTokens <= 0 {
				t.Error("MaxTokens should be positive")
			}

			if cfg.Temperature < 0 || cfg.Temperature > 1 {
				t.Errorf("Temperature should be between 0 and 1, got %f", cfg.Temperature)
			}

			if cfg.Timeout <= 0 {
				t.Error("Timeout should be positive")
			}

			if cfg.MaxMemoryMB <= 0 {
				t.Error("MaxMemoryMB should be positive")
			}

			if cfg.MaxCPUCores <= 0 {
				t.Error("MaxCPUCores should be positive")
			}
		})
	}
}

// TestAgentConfig_Customization testa que config pode ser customizado
func TestAgentConfig_Customization(t *testing.T) {
	// Começar com config padrão
	cfg := DefaultConfig(AgentTypeExplore)

	// Customizar
	cfg.Model = "custom-model"
	cfg.MaxTokens = 1000
	cfg.Temperature = 0.5
	cfg.Timeout = 30 * time.Second
	cfg.WorkDir = "/custom/path"

	// Verificar customizações
	if cfg.Model != "custom-model" {
		t.Error("Should allow model customization")
	}

	if cfg.MaxTokens != 1000 {
		t.Error("Should allow MaxTokens customization")
	}

	if cfg.Temperature != 0.5 {
		t.Error("Should allow Temperature customization")
	}

	if cfg.Timeout != 30*time.Second {
		t.Error("Should allow Timeout customization")
	}

	if cfg.WorkDir != "/custom/path" {
		t.Error("Should allow WorkDir customization")
	}
}

// TestSubagent_Duration_RealTime testa duração com tempo real
func TestSubagent_Duration_RealTime(t *testing.T) {
	agent := &Subagent{
		StartedAt: time.Now(),
	}

	// Aguardar um pouco
	time.Sleep(100 * time.Millisecond)

	duration := agent.Duration()

	// Duração deve ser pelo menos 100ms
	if duration < 100*time.Millisecond {
		t.Errorf("Expected duration >= 100ms, got %v", duration)
	}

	// Mas não deve ser muito maior (< 200ms para tolerância)
	if duration > 200*time.Millisecond {
		t.Errorf("Expected duration < 200ms, got %v", duration)
	}
}

// TestAgentStatus_AllStatuses testa todos os status possíveis
func TestAgentStatus_AllStatuses(t *testing.T) {
	allStatuses := []AgentStatus{
		StatusPending,
		StatusRunning,
		StatusCompleted,
		StatusFailed,
		StatusTimeout,
		StatusKilled,
	}

	// Verificar que cada status tem um valor
	for _, status := range allStatuses {
		if status == "" {
			t.Errorf("Status should not be empty")
		}

		// Verificar comportamento de IsTerminal
		isTerminal := status.IsTerminal()
		expectedTerminal := status == StatusCompleted ||
			status == StatusFailed ||
			status == StatusTimeout ||
			status == StatusKilled

		if isTerminal != expectedTerminal {
			t.Errorf("Status %s: expected IsTerminal()=%v, got %v", status, expectedTerminal, isTerminal)
		}
	}
}

// TestAgentType_Constants testa que constantes têm valores corretos
func TestAgentType_Constants(t *testing.T) {
	if string(AgentTypeExplore) != "Explore" {
		t.Errorf("AgentTypeExplore should be 'Explore', got '%s'", AgentTypeExplore)
	}

	if string(AgentTypePlan) != "Plan" {
		t.Errorf("AgentTypePlan should be 'Plan', got '%s'", AgentTypePlan)
	}

	if string(AgentTypeExecute) != "Execute" {
		t.Errorf("AgentTypeExecute should be 'Execute', got '%s'", AgentTypeExecute)
	}

	if string(AgentTypeGeneral) != "General" {
		t.Errorf("AgentTypeGeneral should be 'General', got '%s'", AgentTypeGeneral)
	}
}

// TestAgentStatus_Constants testa que constantes de status têm valores corretos
func TestAgentStatus_Constants(t *testing.T) {
	expectedStatuses := map[AgentStatus]string{
		StatusPending:   "pending",
		StatusRunning:   "running",
		StatusCompleted: "completed",
		StatusFailed:    "failed",
		StatusTimeout:   "timeout",
		StatusKilled:    "killed",
	}

	for status, expected := range expectedStatuses {
		if string(status) != expected {
			t.Errorf("Expected status '%s', got '%s'", expected, status)
		}
	}
}
