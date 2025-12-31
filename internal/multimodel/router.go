package multimodel

import (
	"fmt"
	"sync"

	"github.com/johnpitter/ollama-code/internal/llm"
)

// Router gerencia roteamento de requests para modelos específicos
type Router struct {
	config  *Config
	clients map[string]*llm.Client // cache de clients por model name
	baseURL string
	mu      sync.RWMutex
}

// NewRouter cria novo router
func NewRouter(baseURL string, config *Config) *Router {
	return &Router{
		config:  config,
		clients: make(map[string]*llm.Client),
		baseURL: baseURL,
	}
}

// GetClient retorna LLM client para um task type
func (r *Router) GetClient(taskType TaskType) (*llm.Client, error) {
	// Obter modelo para o task type
	modelSpec, err := r.config.GetModel(taskType)
	if err != nil {
		return nil, fmt.Errorf("get model for task type %s: %w", taskType, err)
	}

	// Retornar client (criar se não existir)
	return r.getOrCreateClient(modelSpec.Name), nil
}

// GetClientForModel retorna LLM client para um modelo específico
func (r *Router) GetClientForModel(modelName string) *llm.Client {
	return r.getOrCreateClient(modelName)
}

// getOrCreateClient obtém client do cache ou cria novo
func (r *Router) getOrCreateClient(modelName string) *llm.Client {
	r.mu.RLock()
	client, ok := r.clients[modelName]
	r.mu.RUnlock()

	if ok {
		return client
	}

	// Criar novo client
	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check (outro goroutine pode ter criado enquanto esperávamos lock)
	if client, ok := r.clients[modelName]; ok {
		return client
	}

	// Criar e cachear
	client = llm.NewClient(r.baseURL, modelName)
	r.clients[modelName] = client

	return client
}

// GetModelSpec retorna especificação do modelo para um task type
func (r *Router) GetModelSpec(taskType TaskType) (ModelSpec, error) {
	return r.config.GetModel(taskType)
}

// GetDefaultClient retorna client do modelo padrão
func (r *Router) GetDefaultClient() *llm.Client {
	return r.getOrCreateClient(r.config.DefaultModel.Name)
}

// SetConfig atualiza configuração
func (r *Router) SetConfig(config *Config) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.config = config

	return nil
}

// GetConfig retorna configuração atual
func (r *Router) GetConfig() *Config {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.config.Clone()
}

// IsEnabled retorna se multi-model está habilitado
func (r *Router) IsEnabled() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.config.Enabled
}

// Enable habilita multi-model
func (r *Router) Enable() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.config.Enabled = true
}

// Disable desabilita multi-model
func (r *Router) Disable() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.config.Enabled = false
}

// ClearCache limpa cache de clients
func (r *Router) ClearCache() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.clients = make(map[string]*llm.Client)
}

// GetCachedModels retorna lista de modelos em cache
func (r *Router) GetCachedModels() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	models := make([]string, 0, len(r.clients))
	for modelName := range r.clients {
		models = append(models, modelName)
	}

	return models
}

// Stats retorna estatísticas do router
func (r *Router) Stats() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return map[string]interface{}{
		"enabled":       r.config.Enabled,
		"cached_models": len(r.clients),
		"configured_tasks": len(r.config.Models),
		"default_model": r.config.DefaultModel.Name,
	}
}
