package multimodel

import (
	"testing"
)

// TestNewRouter testa criação de router
func TestNewRouter(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter("http://localhost:11434", cfg)

	if router == nil {
		t.Fatal("NewRouter returned nil")
	}

	if router.config == nil {
		t.Error("Router config should be set")
	}

	if router.baseURL != "http://localhost:11434" {
		t.Errorf("Expected baseURL='http://localhost:11434', got '%s'", router.baseURL)
	}

	if router.clients == nil {
		t.Error("Clients map should be initialized")
	}
}

// TestRouter_GetClient testa obtenção de client
func TestRouter_GetClient(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter("http://localhost:11434", cfg)

	client, err := router.GetClient(TaskTypeIntent)
	if err != nil {
		t.Fatalf("GetClient failed: %v", err)
	}

	if client == nil {
		t.Error("Client should not be nil")
	}

	// Verificar que client foi cacheado
	if len(router.clients) != 1 {
		t.Errorf("Expected 1 cached client, got %d", len(router.clients))
	}
}

// TestRouter_GetClient_InvalidTaskType testa get client com tipo inválido
func TestRouter_GetClient_InvalidTaskType(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter("http://localhost:11434", cfg)

	_, err := router.GetClient(TaskType("invalid"))
	if err == nil {
		t.Error("Expected error for invalid task type")
	}
}

// TestRouter_GetClient_Caching testa cache de clients
func TestRouter_GetClient_Caching(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter("http://localhost:11434", cfg)

	// Obter client duas vezes
	client1, err := router.GetClient(TaskTypeIntent)
	if err != nil {
		t.Fatalf("First GetClient failed: %v", err)
	}

	client2, err := router.GetClient(TaskTypeIntent)
	if err != nil {
		t.Fatalf("Second GetClient failed: %v", err)
	}

	// Deve ser o mesmo client (cached)
	if client1 != client2 {
		t.Error("Clients should be the same instance (cached)")
	}

	// Deve ter apenas 1 entry no cache
	if len(router.clients) != 1 {
		t.Errorf("Expected 1 cached client, got %d", len(router.clients))
	}
}

// TestRouter_GetClient_DifferentModels testa cache de modelos diferentes
func TestRouter_GetClient_DifferentModels(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter("http://localhost:11434", cfg)

	// Obter clients para diferentes task types
	client1, _ := router.GetClient(TaskTypeIntent)
	client2, _ := router.GetClient(TaskTypeCode)

	// Devem ser clients diferentes (modelos diferentes)
	if client1 == client2 {
		t.Error("Clients for different models should be different instances")
	}

	// Deve ter 2 entries no cache
	if len(router.clients) != 2 {
		t.Errorf("Expected 2 cached clients, got %d", len(router.clients))
	}
}

// TestRouter_GetClientForModel testa obtenção por nome de modelo
func TestRouter_GetClientForModel(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter("http://localhost:11434", cfg)

	client := router.GetClientForModel("custom-model")

	if client == nil {
		t.Error("Client should not be nil")
	}

	// Verificar que foi cacheado
	cachedModels := router.GetCachedModels()
	found := false
	for _, model := range cachedModels {
		if model == "custom-model" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Model should be in cache")
	}
}

// TestRouter_GetModelSpec testa obtenção de spec
func TestRouter_GetModelSpec(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter("http://localhost:11434", cfg)

	spec, err := router.GetModelSpec(TaskTypeIntent)
	if err != nil {
		t.Fatalf("GetModelSpec failed: %v", err)
	}

	if spec.Name == "" {
		t.Error("Model spec name should not be empty")
	}

	// Deve ser modelo rápido para intent
	if spec.Name != "qwen2.5-coder:1.5b" {
		t.Errorf("Expected fast model for intent, got %s", spec.Name)
	}
}

// TestRouter_GetDefaultClient testa obtenção de client padrão
func TestRouter_GetDefaultClient(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter("http://localhost:11434", cfg)

	client := router.GetDefaultClient()

	if client == nil {
		t.Error("Default client should not be nil")
	}
}

// TestRouter_SetConfig testa atualização de configuração
func TestRouter_SetConfig(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter("http://localhost:11434", cfg)

	newCfg := NewConfig()
	newCfg.DefaultModel = ModelSpec{
		Name:        "new-default",
		MaxTokens:   1024,
		Temperature: 0.5,
	}

	err := router.SetConfig(newCfg)
	if err != nil {
		t.Fatalf("SetConfig failed: %v", err)
	}

	// Verificar que config foi atualizado
	currentCfg := router.GetConfig()
	if currentCfg.DefaultModel.Name != "new-default" {
		t.Errorf("Config not updated, got %s", currentCfg.DefaultModel.Name)
	}
}

// TestRouter_SetConfig_Invalid testa set config inválido
func TestRouter_SetConfig_Invalid(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter("http://localhost:11434", cfg)

	invalidCfg := NewConfig()
	invalidCfg.Enabled = true
	// Não configurar models - inválido

	err := router.SetConfig(invalidCfg)
	if err == nil {
		t.Error("Expected error for invalid config")
	}
}

// TestRouter_GetConfig testa obtenção de configuração
func TestRouter_GetConfig(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter("http://localhost:11434", cfg)

	retrievedCfg := router.GetConfig()

	if retrievedCfg == nil {
		t.Error("GetConfig returned nil")
	}

	// Deve ser uma cópia (não o original)
	if retrievedCfg == router.config {
		t.Error("GetConfig should return a clone")
	}
}

// TestRouter_EnableDisable testa enable/disable
func TestRouter_EnableDisable(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter("http://localhost:11434", cfg)

	if !router.IsEnabled() {
		t.Error("Router should start enabled (using DefaultConfig)")
	}

	router.Disable()
	if router.IsEnabled() {
		t.Error("Router should be disabled after Disable()")
	}

	router.Enable()
	if !router.IsEnabled() {
		t.Error("Router should be enabled after Enable()")
	}
}

// TestRouter_ClearCache testa limpeza de cache
func TestRouter_ClearCache(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter("http://localhost:11434", cfg)

	// Criar alguns clients
	router.GetClient(TaskTypeIntent)
	router.GetClient(TaskTypeCode)

	if len(router.clients) != 2 {
		t.Errorf("Expected 2 cached clients, got %d", len(router.clients))
	}

	// Limpar cache
	router.ClearCache()

	if len(router.clients) != 0 {
		t.Errorf("Expected 0 cached clients after clear, got %d", len(router.clients))
	}
}

// TestRouter_GetCachedModels testa listagem de modelos em cache
func TestRouter_GetCachedModels(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter("http://localhost:11434", cfg)

	// Criar clients para 3 task types
	router.GetClient(TaskTypeIntent)
	router.GetClient(TaskTypeCode)
	router.GetClient(TaskTypeSearch)

	cachedModels := router.GetCachedModels()

	// Devem ser 3 modelos diferentes
	// (intent=1.5b, code=7b, search=3b)
	expectedCount := 3
	if len(cachedModels) != expectedCount {
		t.Errorf("Expected %d cached models, got %d", expectedCount, len(cachedModels))
	}
}

// TestRouter_Stats testa estatísticas
func TestRouter_Stats(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter("http://localhost:11434", cfg)

	// Criar alguns clients
	router.GetClient(TaskTypeIntent)
	router.GetClient(TaskTypeCode)

	stats := router.Stats()

	if stats == nil {
		t.Fatal("Stats returned nil")
	}

	// Verificar campos
	if enabled, ok := stats["enabled"].(bool); !ok || !enabled {
		t.Error("Expected enabled=true")
	}

	if cachedModels, ok := stats["cached_models"].(int); !ok || cachedModels != 2 {
		t.Errorf("Expected cached_models=2, got %v", stats["cached_models"])
	}

	if configuredTasks, ok := stats["configured_tasks"].(int); !ok || configuredTasks != 5 {
		t.Errorf("Expected configured_tasks=5, got %v", stats["configured_tasks"])
	}

	if defaultModel, ok := stats["default_model"].(string); !ok || defaultModel == "" {
		t.Error("Expected default_model to be set")
	}
}

// TestRouter_Concurrent testa acesso concorrente
func TestRouter_Concurrent(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter("http://localhost:11434", cfg)

	// Executar gets concorrentes
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := router.GetClient(TaskTypeIntent)
			if err != nil {
				t.Errorf("Concurrent GetClient failed: %v", err)
			}
			done <- true
		}()
	}

	// Aguardar todos completarem
	for i := 0; i < 10; i++ {
		<-done
	}

	// Deve ter apenas 1 client cacheado
	if len(router.clients) != 1 {
		t.Errorf("Expected 1 cached client, got %d", len(router.clients))
	}
}

// TestRouter_DisabledUsesDefault testa que disabled sempre usa default
func TestRouter_DisabledUsesDefault(t *testing.T) {
	cfg := DefaultConfig()
	router := NewRouter("http://localhost:11434", cfg)

	// Desabilitar
	router.Disable()

	// Obter client para intent
	client, err := router.GetClient(TaskTypeIntent)
	if err != nil {
		t.Fatalf("GetClient failed: %v", err)
	}

	// Obter default client
	defaultClient := router.GetDefaultClient()

	// Quando disabled, intent deve retornar o mesmo que default
	// (mesmo modelo)
	if client.GetModel() != defaultClient.GetModel() {
		t.Errorf("Expected default model when disabled, got %s", client.GetModel())
	}
}
