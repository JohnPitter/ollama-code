package di

import (
	"os"
	"time"

	"github.com/johnpitter/ollama-code/internal/agent"
	"github.com/johnpitter/ollama-code/internal/modes"
)

// NewConfigFromAgent cria Config do DI a partir de agent.Config
func NewConfigFromAgent(agentCfg agent.Config) *Config {
	// Default values
	if agentCfg.OllamaURL == "" {
		agentCfg.OllamaURL = "http://localhost:11434"
	}
	if agentCfg.Model == "" {
		agentCfg.Model = "qwen2.5-coder:7b"
	}
	if agentCfg.WorkDir == "" {
		agentCfg.WorkDir, _ = os.Getwd()
	}
	if agentCfg.Mode == "" {
		agentCfg.Mode = modes.ModeInteractive
	}
	if agentCfg.CacheTTL == 0 {
		agentCfg.CacheTTL = 5 * time.Minute
	}

	return &Config{
		OllamaURL:        agentCfg.OllamaURL,
		Model:            agentCfg.Model,
		Mode:             agentCfg.Mode,
		WorkDir:          agentCfg.WorkDir,
		Temperature:      agentCfg.Temperature,
		MaxTokens:        agentCfg.MaxTokens,
		EnableSessions:   agentCfg.EnableSessions,
		EnableCache:      agentCfg.EnableCache,
		EnableStatusLine: agentCfg.EnableStatusLine,
		CacheTTL:         agentCfg.CacheTTL,
	}
}
