package tools

import (
	"context"
	"fmt"
	"sync"
)

// Registry registro de ferramentas
type Registry struct {
	tools map[string]Tool
	mu    sync.RWMutex
}

// NewRegistry cria novo registro
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register registra uma ferramenta
func (r *Registry) Register(tool Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := tool.Name()
	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool %s already registered", name)
	}

	r.tools[name] = tool
	return nil
}

// Get obtém ferramenta por nome
func (r *Registry) Get(name string) (Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool %s not found", name)
	}

	return tool, nil
}

// List lista todas as ferramentas
func (r *Registry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}

	return tools
}

// Execute executa uma ferramenta
func (r *Registry) Execute(ctx context.Context, toolName string, params map[string]interface{}) (Result, error) {
	tool, err := r.Get(toolName)
	if err != nil {
		return NewErrorResult(err), err
	}

	return tool.Execute(ctx, params)
}
