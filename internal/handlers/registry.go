package handlers

import (
	"context"
	"fmt"
	"sync"

	"github.com/johnpitter/ollama-code/internal/intent"
)

// Registry gerencia handlers de intents
type Registry struct {
	handlers       map[intent.Intent]Handler
	defaultHandler Handler
	mu             sync.RWMutex
}

// NewRegistry cria novo registry
func NewRegistry() *Registry {
	return &Registry{
		handlers: make(map[intent.Intent]Handler),
	}
}

// Register registra um handler para um intent
func (r *Registry) Register(intentType intent.Intent, handler Handler) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.handlers[intentType]; exists {
		return fmt.Errorf("handler already registered for intent: %s", intentType)
	}

	r.handlers[intentType] = handler
	return nil
}

// RegisterDefault registra handler padrão (fallback)
func (r *Registry) RegisterDefault(handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defaultHandler = handler
}

// Handle processa um intent usando o handler apropriado
func (r *Registry) Handle(ctx context.Context, deps *Dependencies, result *intent.DetectionResult) (string, error) {
	r.mu.RLock()
	handler, exists := r.handlers[result.Intent]

	// Se não existe, usar default
	if !exists {
		handler = r.defaultHandler
		if handler == nil {
			r.mu.RUnlock()
			return "", fmt.Errorf("no handler registered for intent: %s", result.Intent)
		}
	}
	r.mu.RUnlock()

	// Executar handler
	return handler.Handle(ctx, deps, result)
}

// GetHandler retorna handler para um intent
func (r *Registry) GetHandler(intentType intent.Intent) (Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	handler, exists := r.handlers[intentType]
	return handler, exists
}

// List retorna todos os intents registrados
func (r *Registry) List() []intent.Intent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	intents := make([]intent.Intent, 0, len(r.handlers))
	for intentType := range r.handlers {
		intents = append(intents, intentType)
	}

	return intents
}
