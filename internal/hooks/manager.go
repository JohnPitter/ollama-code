package hooks

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
)

// Hook tipo de hook
type Hook string

const (
	HookPreToolExec  Hook = "pre_tool_exec"
	HookPostToolExec Hook = "post_tool_exec"
	HookPreCommit    Hook = "pre_commit"
	HookPostCommit   Hook = "post_commit"
)

// Manager gerenciador de hooks
type Manager struct {
	hooks map[Hook][]HookFunc
	mu    sync.RWMutex
}

// HookFunc função de hook
type HookFunc func(ctx context.Context, data map[string]interface{}) error

// NewManager cria novo gerenciador
func NewManager() *Manager {
	return &Manager{
		hooks: make(map[Hook][]HookFunc),
	}
}

// Register registra hook
func (m *Manager) Register(hook Hook, fn HookFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.hooks[hook] = append(m.hooks[hook], fn)
}

// Execute executa hooks
func (m *Manager) Execute(ctx context.Context, hook Hook, data map[string]interface{}) error {
	m.mu.RLock()
	funcs := m.hooks[hook]
	m.mu.RUnlock()

	for _, fn := range funcs {
		if err := fn(ctx, data); err != nil {
			return fmt.Errorf("hook %s failed: %w", hook, err)
		}
	}

	return nil
}

// ExecuteScript executa script de hook
func (m *Manager) ExecuteScript(scriptPath string, data map[string]interface{}) error {
	cmd := exec.Command("sh", scriptPath)
	return cmd.Run()
}
