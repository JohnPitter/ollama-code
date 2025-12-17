package hardware

import (
	"fmt"

	"github.com/johnpitter/ollama-code/internal/config"
)

// Preset tipo de preset de configuração
type Preset string

const (
	// PresetCompatibility compatibilidade máxima (funciona em qualquer hardware)
	PresetCompatibility Preset = "compatibility"

	// PresetPerformance balanceamento entre performance e compatibilidade
	PresetPerformance Preset = "performance"

	// PresetUltra máxima performance (requer hardware potente)
	PresetUltra Preset = "ultra"
)

// Optimizer otimizador de configuração baseado em hardware
type Optimizer struct {
	detector *Detector
}

// NewOptimizer cria novo otimizador
func NewOptimizer() *Optimizer {
	return &Optimizer{
		detector: NewDetector(),
	}
}

// DetectAndOptimize detecta hardware e retorna configuração otimizada
func (o *Optimizer) DetectAndOptimize() (*config.Config, *Specs, Preset, error) {
	// Detectar hardware
	specs, err := o.detector.Detect()
	if err != nil {
		return nil, nil, "", fmt.Errorf("detect hardware: %w", err)
	}

	// Determinar preset baseado no hardware
	preset := o.DeterminePreset(specs)

	// Gerar configuração otimizada
	cfg := o.GenerateConfig(specs, preset)

	return cfg, specs, preset, nil
}

// DeterminePreset determina melhor preset baseado no hardware
func (o *Optimizer) DeterminePreset(specs *Specs) Preset {
	tier := specs.GetPerformanceTier()

	switch tier {
	case "high-end":
		return PresetUltra
	case "mid-range":
		return PresetPerformance
	default:
		return PresetCompatibility
	}
}

// GenerateConfig gera configuração otimizada
func (o *Optimizer) GenerateConfig(specs *Specs, preset Preset) *config.Config {
	switch preset {
	case PresetUltra:
		return o.generateUltraConfig(specs)
	case PresetPerformance:
		return o.generatePerformanceConfig(specs)
	case PresetCompatibility:
		return o.generateCompatibilityConfig(specs)
	default:
		return config.DefaultConfig()
	}
}

// generateCompatibilityConfig gera config para compatibilidade
func (o *Optimizer) generateCompatibilityConfig(specs *Specs) *config.Config {
	cfg := config.DefaultConfig()

	// Modelo menor para compatibilidade
	cfg.Ollama.Model = "qwen2.5-coder:7b-instruct-q4_K_M"
	cfg.Ollama.Temperature = 0.7
	cfg.Ollama.MaxTokens = 2048

	// GPU settings conservadores
	if specs.HasNVIDIAGPU {
		cfg.Ollama.GPULayers = 20 // Apenas algumas layers na GPU
		cfg.Ollama.NumGPU = 1
		cfg.Ollama.MaxVRAM = int(minInt64(specs.GPUMemory/2, 4096)) // Usa no máximo 50% da VRAM ou 4GB
	} else {
		cfg.Ollama.GPULayers = 0 // CPU only
		cfg.Ollama.NumGPU = 0
	}

	cfg.Ollama.NumParallel = 1 // Uma requisição por vez
	cfg.Ollama.FlashAttention = false

	// App settings conservadores
	cfg.App.Mode = "interactive"
	cfg.App.EnableCheckpoints = false // Economizar disco
	cfg.App.CheckpointRetention = 7   // 7 dias apenas
	cfg.App.MaxCheckpoints = 10

	// Performance conservadora
	cfg.Performance.CacheTTL = 5
	cfg.Performance.EnableCache = true
	cfg.Performance.MaxConcurrentTools = 1
	cfg.Performance.CommandTimeout = 30

	return cfg
}

// generatePerformanceConfig gera config para performance balanceada
func (o *Optimizer) generatePerformanceConfig(specs *Specs) *config.Config {
	cfg := config.DefaultConfig()

	// Modelo balanceado
	if specs.TotalRAM >= 32768 && specs.GPUMemory >= 12288 {
		cfg.Ollama.Model = "qwen2.5-coder:32b-instruct-q6_K"
	} else if specs.TotalRAM >= 16384 {
		cfg.Ollama.Model = "qwen2.5-coder:14b-instruct-q5_K_M"
	} else {
		cfg.Ollama.Model = "qwen2.5-coder:7b-instruct-q5_K_M"
	}

	cfg.Ollama.Temperature = 0.7
	cfg.Ollama.MaxTokens = 4096

	// GPU settings balanceados
	if specs.HasNVIDIAGPU {
		cfg.Ollama.GPULayers = 35 // Maioria na GPU
		cfg.Ollama.NumGPU = minInt(specs.GPUCount, 2)
		cfg.Ollama.MaxVRAM = int(minInt64(specs.GPUMemory*80/100, 12288)) // 80% da VRAM ou 12GB
		cfg.Ollama.FlashAttention = true
	} else {
		cfg.Ollama.GPULayers = 0
		cfg.Ollama.NumGPU = 0
	}

	cfg.Ollama.NumParallel = 2

	// App settings balanceados
	cfg.App.Mode = "interactive"
	cfg.App.EnableCheckpoints = true
	cfg.App.EnableSessions = true
	cfg.App.EnableMemory = true
	cfg.App.CheckpointRetention = 14 // 2 semanas
	cfg.App.MaxCheckpoints = 50

	// Performance balanceada
	cfg.Performance.CacheTTL = 15
	cfg.Performance.EnableCache = true
	cfg.Performance.MaxConcurrentTools = 2
	cfg.Performance.CommandTimeout = 60

	return cfg
}

// generateUltraConfig gera config para máxima performance
func (o *Optimizer) generateUltraConfig(specs *Specs) *config.Config {
	cfg := config.DefaultConfig()

	// Modelo máximo
	cfg.Ollama.Model = "qwen2.5-coder:32b-instruct-q6_K"
	cfg.Ollama.Temperature = 0.7
	cfg.Ollama.MaxTokens = 8192

	// GPU settings agressivos
	if specs.HasNVIDIAGPU {
		cfg.Ollama.GPULayers = 999 // Todas as layers
		cfg.Ollama.NumGPU = specs.GPUCount
		cfg.Ollama.MaxVRAM = int(specs.GPUMemory * 95 / 100) // 95% da VRAM
		cfg.Ollama.FlashAttention = true
	} else {
		// Mesmo sem GPU, usa CPU otimizado
		cfg.Ollama.GPULayers = 0
		cfg.Ollama.NumGPU = 0
	}

	cfg.Ollama.NumParallel = minInt(specs.CPUCores/4, 8) // Até 8 paralelos

	// App settings completos
	cfg.App.Mode = "interactive"
	cfg.App.EnableCheckpoints = true
	cfg.App.EnableSessions = true
	cfg.App.EnableMemory = true
	cfg.App.CheckpointRetention = 30 // 30 dias
	cfg.App.MaxCheckpoints = 100

	// Performance máxima
	cfg.Performance.CacheTTL = 30
	cfg.Performance.EnableCache = true
	cfg.Performance.MaxConcurrentTools = minInt(specs.CPUCores/2, 5)
	cfg.Performance.CommandTimeout = 120

	return cfg
}

// minInt64 retorna menor valor int64
func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// minInt retorna menor valor int
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetPresetDescription retorna descrição do preset
func GetPresetDescription(preset Preset) string {
	switch preset {
	case PresetCompatibility:
		return "Compatibilidade - Funciona em qualquer hardware (modelo 7B, uso mínimo de recursos)"
	case PresetPerformance:
		return "Performance - Balanceamento entre velocidade e compatibilidade (modelo 14B-32B)"
	case PresetUltra:
		return "Ultra - Máxima performance, requer hardware potente (modelo 32B, todas as otimizações)"
	default:
		return "Unknown preset"
	}
}

// GetRecommendedModel retorna modelo recomendado para o preset
func GetRecommendedModel(preset Preset, hasGPU bool, vram int64, ram int64) string {
	switch preset {
	case PresetCompatibility:
		return "qwen2.5-coder:7b-instruct-q4_K_M"

	case PresetPerformance:
		if ram >= 32768 && hasGPU && vram >= 12288 {
			return "qwen2.5-coder:32b-instruct-q6_K"
		} else if ram >= 16384 {
			return "qwen2.5-coder:14b-instruct-q5_K_M"
		}
		return "qwen2.5-coder:7b-instruct-q5_K_M"

	case PresetUltra:
		if ram >= 64000 && hasGPU && vram >= 24576 {
			return "qwen2.5-coder:72b-instruct-q4_K_M"
		}
		return "qwen2.5-coder:32b-instruct-q6_K"

	default:
		return "qwen2.5-coder:7b-instruct-q5_K_M"
	}
}

// PrintOptimizationReport imprime relatório de otimização
func PrintOptimizationReport(specs *Specs, preset Preset, cfg *config.Config) string {
	report := fmt.Sprintf(`
╔════════════════════════════════════════════════════════════╗
║          OLLAMA CODE - HARDWARE DETECTION REPORT           ║
╚════════════════════════════════════════════════════════════╝

🖥️  HARDWARE DETECTED:
   CPU: %s
   Cores/Threads: %d / %d
   RAM: %d MB total (%d MB available)
   GPU: %s
   VRAM: %d MB (%d GPU(s))
   Disk Space: %d GB available
   OS: %s / %s

⚡ PERFORMANCE TIER: %s

🎯 PRESET SELECTED: %s
   %s

⚙️  OPTIMIZED CONFIGURATION:
   Model: %s
   Temperature: %.1f
   Max Tokens: %d
   GPU Layers: %d
   Max VRAM: %d MB
   Parallel Requests: %d
   Flash Attention: %v

   Mode: %s
   Checkpoints: %v (retention: %d days, max: %d)
   Sessions: %v
   Memory: %v

   Cache: %v (TTL: %d min)
   Max Concurrent Tools: %d
   Command Timeout: %d sec

💡 RECOMMENDATIONS:
`,
		specs.CPUModel,
		specs.CPUCores, specs.CPUThreads,
		specs.TotalRAM, specs.AvailableRAM,
		specs.GPUModel,
		specs.GPUMemory, specs.GPUCount,
		specs.DiskSpace,
		specs.OS, specs.Arch,

		specs.GetPerformanceTier(),

		preset,
		GetPresetDescription(preset),

		cfg.Ollama.Model,
		cfg.Ollama.Temperature,
		cfg.Ollama.MaxTokens,
		cfg.Ollama.GPULayers,
		cfg.Ollama.MaxVRAM,
		cfg.Ollama.NumParallel,
		cfg.Ollama.FlashAttention,

		cfg.App.Mode,
		cfg.App.EnableCheckpoints, cfg.App.CheckpointRetention, cfg.App.MaxCheckpoints,
		cfg.App.EnableSessions,
		cfg.App.EnableMemory,

		cfg.Performance.EnableCache, cfg.Performance.CacheTTL,
		cfg.Performance.MaxConcurrentTools,
		cfg.Performance.CommandTimeout,
	)

	// Adicionar recomendações específicas
	if !specs.HasNVIDIAGPU {
		report += "   ⚠️  No NVIDIA GPU detected - using CPU mode (slower)\n"
		report += "   💡 Consider using a system with NVIDIA GPU for better performance\n"
	} else if specs.GPUMemory < 8192 {
		report += "   ⚠️  GPU has limited VRAM - consider upgrading to 12GB+ for larger models\n"
	}

	if specs.TotalRAM < 16384 {
		report += "   ⚠️  Limited RAM - recommend 16GB+ for better performance\n"
	}

	if specs.DiskSpace < 50 {
		report += "   ⚠️  Low disk space - ensure 50GB+ free for models and checkpoints\n"
	}

	report += "\n✅ Configuration optimized for your hardware!\n"
	report += "   Config saved to: ~/.ollama-code/config.json\n\n"

	return report
}
