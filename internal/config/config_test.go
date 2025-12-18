package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configJSON := `{
		"ollama": {
			"url": "http://localhost:11434",
			"model": "qwen2.5-coder:7b",
			"temperature": 0.7,
			"max_tokens": 2048,
			"gpu_layers": 20,
			"num_gpu": 1,
			"max_vram": 4096
		},
		"app": {
			"mode": "interactive",
			"work_dir": ".",
			"enable_checkpoints": true
		},
		"performance": {
			"cache_ttl": 5,
			"enable_cache": true,
			"max_concurrent_tools": 2,
			"command_timeout": 30
		}
	}`

	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Test load
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify values
	if cfg.Ollama.URL != "http://localhost:11434" {
		t.Errorf("Expected URL 'http://localhost:11434', got '%s'", cfg.Ollama.URL)
	}
	if cfg.Ollama.Model != "qwen2.5-coder:7b" {
		t.Errorf("Expected model 'qwen2.5-coder:7b', got '%s'", cfg.Ollama.Model)
	}
	if cfg.App.Mode != "interactive" {
		t.Errorf("Expected mode 'interactive', got '%s'", cfg.App.Mode)
	}
	if !cfg.App.EnableCheckpoints {
		t.Error("Expected EnableCheckpoints to be true")
	}
	if cfg.Performance.CacheTTL != 5 {
		t.Errorf("Expected CacheTTL 5, got %d", cfg.Performance.CacheTTL)
	}
}

func TestLoadConfig_NonExistent(t *testing.T) {
	// Load creates default config if file doesn't exist, so we test that
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Expected Load to create default config, got error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Expected default config, got nil")
	}

	// Verify default values were applied
	if cfg.Ollama.URL == "" {
		t.Error("Expected default URL to be set")
	}
	if cfg.Ollama.Model == "" {
		t.Error("Expected default model to be set")
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := &Config{
		Ollama: OllamaConfig{
			URL:         "http://localhost:11434",
			Model:       "qwen2.5-coder:7b",
			Temperature: 0.7,
			MaxTokens:   2048,
			GPULayers:   20,
			NumGPU:      1,
			MaxVRAM:     4096,
		},
		App: AppConfig{
			Mode:              "interactive",
			WorkDir:           ".",
			EnableCheckpoints: true,
		},
		Performance: PerformanceConfig{
			CacheTTL:           5,
			EnableCache:        true,
			MaxConcurrentTools: 2,
			CommandTimeout:     30,
		},
	}

	// Save
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("Config file not created: %v", err)
	}

	// Load back and verify
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loaded.Ollama.Model != cfg.Ollama.Model {
		t.Errorf("Expected model '%s', got '%s'", cfg.Ollama.Model, loaded.Ollama.Model)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Ollama: OllamaConfig{
					URL:   "http://localhost:11434",
					Model: "qwen2.5-coder:7b",
				},
				App: AppConfig{
					Mode: "interactive",
				},
			},
			wantErr: false,
		},
		{
			name: "missing URL",
			config: &Config{
				Ollama: OllamaConfig{
					Model: "qwen2.5-coder:7b",
				},
				App: AppConfig{
					Mode: "interactive",
				},
			},
			wantErr: true,
		},
		{
			name: "missing model",
			config: &Config{
				Ollama: OllamaConfig{
					URL: "http://localhost:11434",
				},
				App: AppConfig{
					Mode: "interactive",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid mode",
			config: &Config{
				Ollama: OllamaConfig{
					URL:   "http://localhost:11434",
					Model: "qwen2.5-coder:7b",
				},
				App: AppConfig{
					Mode: "invalid",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
