package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config configuração completa da aplicação
type Config struct {
	// Ollama settings
	Ollama OllamaConfig `json:"ollama"`

	// Application settings
	App AppConfig `json:"app"`

	// Performance settings
	Performance PerformanceConfig `json:"performance,omitempty"`
}

// OllamaConfig configurações do Ollama
type OllamaConfig struct {
	URL            string  `json:"url"`                       // URL do servidor Ollama
	Model          string  `json:"model"`                     // Modelo padrão
	Temperature    float64 `json:"temperature,omitempty"`     // Temperatura (0.0-1.0)
	MaxTokens      int     `json:"max_tokens,omitempty"`      // Max tokens por resposta
	GPULayers      int     `json:"gpu_layers,omitempty"`      // Número de layers na GPU
	NumGPU         int     `json:"num_gpu,omitempty"`         // Número de GPUs
	MaxVRAM        int     `json:"max_vram,omitempty"`        // Max VRAM em MB
	NumParallel    int     `json:"num_parallel,omitempty"`    // Requisições paralelas
	FlashAttention bool    `json:"flash_attention,omitempty"` // Usar flash attention
}

// AppConfig configurações da aplicação
type AppConfig struct {
	Mode                string `json:"mode"`                           // Modo padrão (readonly, interactive, autonomous)
	WorkDir             string `json:"work_dir,omitempty"`             // Diretório de trabalho padrão
	OutputStyle         string `json:"output_style,omitempty"`         // Estilo de output
	EnableColors        bool   `json:"enable_colors"`                  // Usar cores no terminal
	EnableCheckpoints   bool   `json:"enable_checkpoints"`             // Habilitar checkpoints automáticos
	EnableSessions      bool   `json:"enable_sessions"`                // Habilitar sessões
	EnableMemory        bool   `json:"enable_memory"`                  // Habilitar memória hierárquica
	CheckpointRetention int    `json:"checkpoint_retention,omitempty"` // Dias de retenção
	MaxCheckpoints      int    `json:"max_checkpoints,omitempty"`      // Máximo de checkpoints
	LogLevel            string `json:"log_level,omitempty"`            // Nível de log (debug, info, warn, error)
	LogFile             string `json:"log_file,omitempty"`             // Arquivo de log
}

// PerformanceConfig configurações de performance
type PerformanceConfig struct {
	CacheTTL           int  `json:"cache_ttl,omitempty"`            // TTL do cache em minutos
	EnableCache        bool `json:"enable_cache"`                   // Habilitar cache
	MaxConcurrentTools int  `json:"max_concurrent_tools,omitempty"` // Max tools paralelas
	CommandTimeout     int  `json:"command_timeout,omitempty"`      // Timeout de comandos em segundos
}

// DefaultConfig retorna configuração padrão
func DefaultConfig() *Config {
	return &Config{
		Ollama: OllamaConfig{
			URL:            "http://localhost:11434",
			Model:          "qwen2.5-coder:7b",
			Temperature:    0.7,
			MaxTokens:      4096,
			GPULayers:      35,
			NumGPU:         1,
			MaxVRAM:        8192,
			NumParallel:    2,
			FlashAttention: true,
		},
		App: AppConfig{
			Mode:                "interactive",
			OutputStyle:         "default",
			EnableColors:        true,
			EnableCheckpoints:   true,
			EnableSessions:      true,
			EnableMemory:        true,
			CheckpointRetention: 30,
			MaxCheckpoints:      100,
			LogLevel:            "info",
		},
		Performance: PerformanceConfig{
			CacheTTL:           15,
			EnableCache:        true,
			MaxConcurrentTools: 3,
			CommandTimeout:     60,
		},
	}
}

// Load carrega configuração de arquivo
func Load(configPath string) (*Config, error) {
	// Se arquivo não existe, criar com defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cfg := DefaultConfig()
		if err := cfg.Save(configPath); err != nil {
			return nil, fmt.Errorf("create default config: %w", err)
		}
		return cfg, nil
	}

	// Ler arquivo
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	// Parse JSON
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config JSON: %w", err)
	}

	return &cfg, nil
}

// Save salva configuração em arquivo
func (c *Config) Save(configPath string) error {
	// Criar diretório se necessário
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	// Serializar para JSON
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	// Escrever arquivo
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

// GetConfigPath retorna caminho padrão do arquivo de configuração
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".ollama-code", "config.json"), nil
}

// LoadDefault carrega configuração do local padrão
func LoadDefault() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	return Load(configPath)
}

// LoadOrOptimize carrega config existente ou cria otimizado baseado em hardware
func LoadOrOptimize() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// Se config existe, carregar
	if _, err := os.Stat(configPath); err == nil {
		return Load(configPath)
	}

	// Config não existe - retornar nil para permitir otimização
	return nil, nil
}

// Validate valida configuração
func (c *Config) Validate() error {
	if c.Ollama.URL == "" {
		return fmt.Errorf("ollama.url is required")
	}

	if c.Ollama.Model == "" {
		return fmt.Errorf("ollama.model is required")
	}

	validModes := map[string]bool{"readonly": true, "interactive": true, "autonomous": true}
	if !validModes[c.App.Mode] {
		return fmt.Errorf("invalid mode: %s (must be readonly, interactive, or autonomous)", c.App.Mode)
	}

	if c.Ollama.Temperature < 0 || c.Ollama.Temperature > 1 {
		return fmt.Errorf("ollama.temperature must be between 0 and 1")
	}

	return nil
}

// Merge mescla configuração com outra (other sobrescreve)
func (c *Config) Merge(other *Config) {
	if other.Ollama.URL != "" {
		c.Ollama.URL = other.Ollama.URL
	}
	if other.Ollama.Model != "" {
		c.Ollama.Model = other.Ollama.Model
	}
	if other.App.Mode != "" {
		c.App.Mode = other.App.Mode
	}
	// ... outros campos conforme necessário
}
