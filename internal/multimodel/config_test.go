package multimodel

import "testing"

// TestNewConfig testa criação de nova configuração
func TestNewConfig(t *testing.T) {
	cfg := NewConfig()

	if cfg == nil {
		t.Fatal("NewConfig returned nil")
	}

	if cfg.Models == nil {
		t.Error("Models map should be initialized")
	}

	if cfg.Enabled {
		t.Error("Config should start disabled")
	}
}

// TestDefaultConfig testa configuração padrão
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	if !cfg.Enabled {
		t.Error("DefaultConfig should be enabled")
	}

	// Verificar que todos os task types têm modelos
	requiredTypes := []TaskType{
		TaskTypeIntent,
		TaskTypeCode,
		TaskTypeSearch,
		TaskTypeAnalysis,
		TaskTypeDefault,
	}

	for _, taskType := range requiredTypes {
		spec, ok := cfg.Models[taskType]
		if !ok {
			t.Errorf("Missing model for task type: %s", taskType)
		}

		if spec.Name == "" {
			t.Errorf("Empty model name for task type: %s", taskType)
		}

		if spec.MaxTokens <= 0 {
			t.Errorf("Invalid MaxTokens for task type %s: %d", taskType, spec.MaxTokens)
		}

		if spec.Temperature < 0 || spec.Temperature > 1 {
			t.Errorf("Invalid Temperature for task type %s: %f", taskType, spec.Temperature)
		}
	}

	// Verificar default model
	if cfg.DefaultModel.Name == "" {
		t.Error("DefaultModel should be configured")
	}
}

// TestConfig_GetModel testa obtenção de modelo
func TestConfig_GetModel(t *testing.T) {
	cfg := DefaultConfig()

	// Obter modelo para intent
	spec, err := cfg.GetModel(TaskTypeIntent)
	if err != nil {
		t.Fatalf("GetModel failed: %v", err)
	}

	if spec.Name == "" {
		t.Error("Model name should not be empty")
	}

	// Verificar que é o modelo correto (fast para intent)
	if spec.Name != "qwen2.5-coder:1.5b" {
		t.Errorf("Expected fast model for intent, got %s", spec.Name)
	}
}

// TestConfig_GetModel_InvalidType testa obtenção com tipo inválido
func TestConfig_GetModel_InvalidType(t *testing.T) {
	cfg := DefaultConfig()

	_, err := cfg.GetModel(TaskType("invalid"))
	if err == nil {
		t.Error("Expected error for invalid task type")
	}
}

// TestConfig_GetModel_DisabledReturnsDefault testa que disabled retorna default
func TestConfig_GetModel_DisabledReturnsDefault(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Disable()

	// Solicitar modelo de intent, mas deve retornar default
	spec, err := cfg.GetModel(TaskTypeIntent)
	if err != nil {
		t.Fatalf("GetModel failed: %v", err)
	}

	if spec.Name != cfg.DefaultModel.Name {
		t.Errorf("Expected default model when disabled, got %s", spec.Name)
	}
}

// TestConfig_SetModel testa configuração de modelo
func TestConfig_SetModel(t *testing.T) {
	cfg := NewConfig()

	customSpec := ModelSpec{
		Name:        "custom-model",
		MaxTokens:   2048,
		Temperature: 0.8,
		Description: "Custom model",
	}

	err := cfg.SetModel(TaskTypeCode, customSpec)
	if err != nil {
		t.Fatalf("SetModel failed: %v", err)
	}

	// Verificar que foi configurado
	spec, ok := cfg.Models[TaskTypeCode]
	if !ok {
		t.Error("Model not found after SetModel")
	}

	if spec.Name != "custom-model" {
		t.Errorf("Expected 'custom-model', got '%s'", spec.Name)
	}
}

// TestConfig_SetModel_InvalidType testa set com tipo inválido
func TestConfig_SetModel_InvalidType(t *testing.T) {
	cfg := NewConfig()

	spec := ModelSpec{Name: "test"}
	err := cfg.SetModel(TaskType("invalid"), spec)
	if err == nil {
		t.Error("Expected error for invalid task type")
	}
}

// TestConfig_SetModel_EmptyName testa set com nome vazio
func TestConfig_SetModel_EmptyName(t *testing.T) {
	cfg := NewConfig()

	spec := ModelSpec{Name: ""}
	err := cfg.SetModel(TaskTypeCode, spec)
	if err == nil {
		t.Error("Expected error for empty model name")
	}
}

// TestConfig_SetModel_UpdatesDefault testa que set default atualiza DefaultModel
func TestConfig_SetModel_UpdatesDefault(t *testing.T) {
	cfg := NewConfig()

	customSpec := ModelSpec{
		Name:        "new-default",
		MaxTokens:   1024,
		Temperature: 0.5,
	}

	err := cfg.SetModel(TaskTypeDefault, customSpec)
	if err != nil {
		t.Fatalf("SetModel failed: %v", err)
	}

	if cfg.DefaultModel.Name != "new-default" {
		t.Errorf("DefaultModel not updated, got %s", cfg.DefaultModel.Name)
	}
}

// TestConfig_EnableDisable testa enable/disable
func TestConfig_EnableDisable(t *testing.T) {
	cfg := NewConfig()

	if cfg.IsEnabled() {
		t.Error("Config should start disabled")
	}

	cfg.Enable()
	if !cfg.IsEnabled() {
		t.Error("Config should be enabled after Enable()")
	}

	cfg.Disable()
	if cfg.IsEnabled() {
		t.Error("Config should be disabled after Disable()")
	}
}

// TestConfig_Validate testa validação
func TestConfig_Validate(t *testing.T) {
	// Configuração válida
	cfg := DefaultConfig()
	err := cfg.Validate()
	if err != nil {
		t.Errorf("Valid config should pass validation: %v", err)
	}
}

// TestConfig_Validate_NoDefaultModel testa validação sem default model
func TestConfig_Validate_NoDefaultModel(t *testing.T) {
	cfg := NewConfig()
	cfg.DefaultModel = ModelSpec{} // Empty

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for missing default model")
	}
}

// TestConfig_Validate_MissingTaskType testa validação com task type faltando
func TestConfig_Validate_MissingTaskType(t *testing.T) {
	cfg := DefaultConfig()

	// Remover um task type
	delete(cfg.Models, TaskTypeIntent)

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for missing task type")
	}
}

// TestConfig_Validate_EmptyModelName testa validação com nome vazio
func TestConfig_Validate_EmptyModelName(t *testing.T) {
	cfg := DefaultConfig()

	// Configurar modelo com nome vazio
	cfg.Models[TaskTypeCode] = ModelSpec{
		Name:        "",
		MaxTokens:   1024,
		Temperature: 0.7,
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for empty model name")
	}
}

// TestConfig_Validate_InvalidMaxTokens testa validação com MaxTokens inválido
func TestConfig_Validate_InvalidMaxTokens(t *testing.T) {
	cfg := DefaultConfig()

	cfg.Models[TaskTypeCode] = ModelSpec{
		Name:        "test",
		MaxTokens:   0,
		Temperature: 0.7,
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for invalid MaxTokens")
	}
}

// TestConfig_Validate_InvalidTemperature testa validação com Temperature inválida
func TestConfig_Validate_InvalidTemperature(t *testing.T) {
	testCases := []struct {
		name        string
		temperature float64
	}{
		{"negative temperature", -0.1},
		{"temperature > 1", 1.1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := DefaultConfig()

			cfg.Models[TaskTypeCode] = ModelSpec{
				Name:        "test",
				MaxTokens:   1024,
				Temperature: tc.temperature,
			}

			err := cfg.Validate()
			if err == nil {
				t.Error("Expected error for invalid Temperature")
			}
		})
	}
}

// TestConfig_Validate_DisabledSkipsValidation testa que disabled não valida models
func TestConfig_Validate_DisabledSkipsValidation(t *testing.T) {
	cfg := NewConfig()
	cfg.Disable()

	// Configurar default model válido
	cfg.DefaultModel = ModelSpec{
		Name:        "default",
		MaxTokens:   1024,
		Temperature: 0.7,
	}

	// Não configurar outros models - quando disabled, não deve validar
	err := cfg.Validate()
	if err != nil {
		t.Errorf("Disabled config should not validate models: %v", err)
	}
}

// TestConfig_Clone testa clonagem de configuração
func TestConfig_Clone(t *testing.T) {
	cfg := DefaultConfig()

	clone := cfg.Clone()

	if clone == nil {
		t.Fatal("Clone returned nil")
	}

	// Verificar que é uma cópia
	if clone == cfg {
		t.Error("Clone should be a different instance")
	}

	// Verificar que enabled foi copiado
	if clone.Enabled != cfg.Enabled {
		t.Error("Enabled not cloned correctly")
	}

	// Verificar que default model foi copiado
	if clone.DefaultModel.Name != cfg.DefaultModel.Name {
		t.Error("DefaultModel not cloned correctly")
	}

	// Verificar que models foi copiado
	if len(clone.Models) != len(cfg.Models) {
		t.Errorf("Expected %d models, got %d", len(cfg.Models), len(clone.Models))
	}

	// Modificar clone não deve afetar original
	clone.Models[TaskTypeCode] = ModelSpec{Name: "modified"}

	originalSpec := cfg.Models[TaskTypeCode]
	if originalSpec.Name == "modified" {
		t.Error("Modifying clone affected original")
	}
}

// TestDefaultConfig_ModelOptimization testa que modelos padrão estão otimizados
func TestDefaultConfig_ModelOptimization(t *testing.T) {
	cfg := DefaultConfig()

	// Intent deve usar modelo rápido
	intentSpec := cfg.Models[TaskTypeIntent]
	if intentSpec.Name != "qwen2.5-coder:1.5b" {
		t.Errorf("Intent should use fast model, got %s", intentSpec.Name)
	}

	// Code deve usar modelo preciso
	codeSpec := cfg.Models[TaskTypeCode]
	if codeSpec.Name != "qwen2.5-coder:7b" {
		t.Errorf("Code should use precise model, got %s", codeSpec.Name)
	}

	// Search deve usar modelo balanceado
	searchSpec := cfg.Models[TaskTypeSearch]
	if searchSpec.Name != "qwen2.5-coder:3b" {
		t.Errorf("Search should use balanced model, got %s", searchSpec.Name)
	}

	// Analysis deve usar modelo preciso
	analysisSpec := cfg.Models[TaskTypeAnalysis]
	if analysisSpec.Name != "qwen2.5-coder:7b" {
		t.Errorf("Analysis should use precise model, got %s", analysisSpec.Name)
	}
}
