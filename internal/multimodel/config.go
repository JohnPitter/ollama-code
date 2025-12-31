package multimodel

import "fmt"

// Config configuração de modelos múltiplos
type Config struct {
	// Models mapeamento de TaskType -> ModelSpec
	Models map[TaskType]ModelSpec

	// DefaultModel modelo padrão caso task type não seja encontrado
	DefaultModel ModelSpec

	// Enabled se multi-model está habilitado
	Enabled bool
}

// NewConfig cria nova configuração multi-model
func NewConfig() *Config {
	return &Config{
		Models:  make(map[TaskType]ModelSpec),
		Enabled: false,
	}
}

// DefaultConfig retorna configuração padrão otimizada
func DefaultConfig() *Config {
	cfg := &Config{
		Enabled: true,
		Models: map[TaskType]ModelSpec{
			// Intent detection - modelo rápido e leve
			TaskTypeIntent: {
				Name:        "qwen2.5-coder:1.5b",
				MaxTokens:   512,
				Temperature: 0.3,
				Description: "Fast model for intent detection",
			},

			// Code generation - modelo preciso
			TaskTypeCode: {
				Name:        "qwen2.5-coder:7b",
				MaxTokens:   4096,
				Temperature: 0.7,
				Description: "Precise model for code generation",
			},

			// Web search summarization - modelo balanceado
			TaskTypeSearch: {
				Name:        "qwen2.5-coder:3b",
				MaxTokens:   2048,
				Temperature: 0.5,
				Description: "Balanced model for web search summarization",
			},

			// Code analysis - modelo preciso
			TaskTypeAnalysis: {
				Name:        "qwen2.5-coder:7b",
				MaxTokens:   8192,
				Temperature: 0.5,
				Description: "Precise model for code analysis",
			},

			// Default - modelo padrão
			TaskTypeDefault: {
				Name:        "qwen2.5-coder:7b",
				MaxTokens:   4096,
				Temperature: 0.7,
				Description: "Default general-purpose model",
			},
		},
	}

	// Definir modelo padrão
	cfg.DefaultModel = cfg.Models[TaskTypeDefault]

	return cfg
}

// GetModel retorna modelo para um task type
func (c *Config) GetModel(taskType TaskType) (ModelSpec, error) {
	if !c.Enabled {
		return c.DefaultModel, nil
	}

	if !taskType.IsValid() {
		return ModelSpec{}, fmt.Errorf("invalid task type: %s", taskType)
	}

	model, ok := c.Models[taskType]
	if !ok {
		// Fallback para default
		return c.DefaultModel, nil
	}

	return model, nil
}

// SetModel define modelo para um task type
func (c *Config) SetModel(taskType TaskType, spec ModelSpec) error {
	if !taskType.IsValid() {
		return fmt.Errorf("invalid task type: %s", taskType)
	}

	if spec.Name == "" {
		return fmt.Errorf("model name cannot be empty")
	}

	c.Models[taskType] = spec

	// Se for default, atualizar DefaultModel também
	if taskType == TaskTypeDefault {
		c.DefaultModel = spec
	}

	return nil
}

// Enable habilita multi-model
func (c *Config) Enable() {
	c.Enabled = true
}

// Disable desabilita multi-model (sempre usa default)
func (c *Config) Disable() {
	c.Enabled = false
}

// IsEnabled retorna se multi-model está habilitado
func (c *Config) IsEnabled() bool {
	return c.Enabled
}

// Validate valida configuração
func (c *Config) Validate() error {
	if c.DefaultModel.Name == "" {
		return fmt.Errorf("default model must be configured")
	}

	if c.Enabled {
		// Validar que todos os task types têm modelos configurados
		requiredTypes := []TaskType{
			TaskTypeIntent,
			TaskTypeCode,
			TaskTypeSearch,
			TaskTypeAnalysis,
			TaskTypeDefault,
		}

		for _, taskType := range requiredTypes {
			model, ok := c.Models[taskType]
			if !ok {
				return fmt.Errorf("missing model for task type: %s", taskType)
			}

			if model.Name == "" {
				return fmt.Errorf("empty model name for task type: %s", taskType)
			}

			if model.MaxTokens <= 0 {
				return fmt.Errorf("invalid MaxTokens for task type %s: %d", taskType, model.MaxTokens)
			}

			if model.Temperature < 0 || model.Temperature > 1 {
				return fmt.Errorf("invalid Temperature for task type %s: %f", taskType, model.Temperature)
			}
		}
	}

	return nil
}

// Clone cria cópia profunda da configuração
func (c *Config) Clone() *Config {
	clone := &Config{
		Enabled:      c.Enabled,
		DefaultModel: c.DefaultModel,
		Models:       make(map[TaskType]ModelSpec, len(c.Models)),
	}

	for taskType, spec := range c.Models {
		clone.Models[taskType] = spec
	}

	return clone
}
