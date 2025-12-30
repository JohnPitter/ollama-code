package agent

import (
	"testing"

	"github.com/johnpitter/ollama-code/internal/llm"
	"github.com/johnpitter/ollama-code/internal/modes"
)

func TestNewAgent(t *testing.T) {
	cfg := Config{
		OllamaURL: "http://localhost:11434",
		Model:     "qwen2.5-coder:7b",
		Mode:      modes.ModeInteractive,
		WorkDir:   ".",
	}

	agent, err := NewAgent(cfg)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	if agent == nil {
		t.Fatal("Agent is nil")
	}

	// Verify defaults
	if agent.LLMClient == nil {
		t.Error("LLM client not initialized")
	}
	if agent.IntentDetector == nil {
		t.Error("Intent detector not initialized")
	}
	if agent.ToolRegistry == nil {
		t.Error("Tool registry not initialized")
	}
	if agent.CommandRegistry == nil {
		t.Error("Command registry not initialized")
	}
	if agent.ConfirmManager == nil {
		t.Error("Confirmation manager not initialized")
	}
}

func TestNewAgent_DefaultValues(t *testing.T) {
	cfg := Config{}

	agent, err := NewAgent(cfg)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Check defaults
	if agent.Mode != modes.ModeInteractive {
		t.Errorf("Expected default mode interactive, got %s", agent.Mode)
	}

	if agent.WorkDir == "" {
		t.Error("Work directory should not be empty")
	}
}

func TestNewAgent_WithSessions(t *testing.T) {
	cfg := Config{
		OllamaURL:      "http://localhost:11434",
		Model:          "qwen2.5-coder:7b",
		EnableSessions: true,
	}

	agent, err := NewAgent(cfg)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	if agent.SessionManager == nil {
		t.Error("Session manager should be initialized when enabled")
	}
}

func TestNewAgent_WithCache(t *testing.T) {
	cfg := Config{
		OllamaURL:   "http://localhost:11434",
		Model:       "qwen2.5-coder:7b",
		EnableCache: true,
	}

	agent, err := NewAgent(cfg)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	if agent.Cache == nil {
		t.Error("Cache should be initialized when enabled")
	}
}

func TestGetMode(t *testing.T) {
	cfg := Config{
		Mode: modes.ModeReadOnly,
	}

	agent, _ := NewAgent(cfg)

	if agent.GetMode() != modes.ModeReadOnly {
		t.Errorf("Expected mode readonly, got %s", agent.GetMode())
	}
}

func TestSetMode(t *testing.T) {
	agent, _ := NewAgent(Config{})

	agent.SetMode(modes.ModeAutonomous)

	if agent.GetMode() != modes.ModeAutonomous {
		t.Errorf("Expected mode autonomous, got %s", agent.GetMode())
	}
}

func TestGetWorkDir(t *testing.T) {
	cfg := Config{
		WorkDir: "/test/path",
	}

	agent, _ := NewAgent(cfg)

	if agent.GetWorkDir() != "/test/path" {
		t.Errorf("Expected work dir /test/path, got %s", agent.GetWorkDir())
	}
}

func TestSetWorkDir(t *testing.T) {
	agent, _ := NewAgent(Config{})

	// Test with current directory (should work)
	err := agent.SetWorkDir(".")
	if err != nil {
		t.Errorf("Failed to set work dir to current directory: %v", err)
	}

	// Test with invalid directory (should fail)
	err = agent.SetWorkDir("/nonexistent/invalid/path/12345")
	if err == nil {
		t.Error("Expected error for invalid directory, got nil")
	}
}

func TestClearHistory(t *testing.T) {
	agent, _ := NewAgent(Config{})

	// Add some history manually
	agent.History = append(agent.History, llm.Message{
		Role:    "user",
		Content: "test",
	})

	if len(agent.History) == 0 {
		t.Error("History should not be empty")
	}

	agent.ClearHistory()

	if len(agent.History) != 0 {
		t.Errorf("History should be empty after clear, got %d messages", len(agent.History))
	}
}

func TestGetHistory(t *testing.T) {
	agent, _ := NewAgent(Config{})

	history := agent.GetHistory()

	if history == nil {
		t.Error("History should not be nil")
	}

	if len(history) != 0 {
		t.Error("Initial history should be empty")
	}
}

func TestGetCommandRegistry(t *testing.T) {
	agent, _ := NewAgent(Config{})

	registry := agent.GetCommandRegistry()

	if registry == nil {
		t.Error("Command registry should not be nil")
	}

	// Verify built-in commands are registered
	commands := registry.List()
	if len(commands) == 0 {
		t.Error("Command registry should have built-in commands")
	}
}

func TestGetSessionManager(t *testing.T) {
	// Without sessions
	agent1, _ := NewAgent(Config{EnableSessions: false})
	if agent1.GetSessionManager() != nil {
		t.Error("Session manager should be nil when disabled")
	}

	// With sessions
	agent2, _ := NewAgent(Config{EnableSessions: true})
	if agent2.GetSessionManager() == nil {
		t.Error("Session manager should not be nil when enabled")
	}
}

func TestGetCache(t *testing.T) {
	// Without cache
	agent1, _ := NewAgent(Config{EnableCache: false})
	if agent1.GetCache() != nil {
		t.Error("Cache should be nil when disabled")
	}

	// With cache
	agent2, _ := NewAgent(Config{EnableCache: true})
	if agent2.GetCache() == nil {
		t.Error("Cache should not be nil when enabled")
	}
}
