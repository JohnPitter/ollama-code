package statusline

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

// StatusLine gerenciador de linha de status
type StatusLine struct {
	model        string
	mode         string
	workDir      string
	tokensUsed   int
	totalTokens  int
	responseTime time.Duration
	activeTask   string
	enabled      bool

	// Colors
	cyan    *color.Color
	yellow  *color.Color
	green   *color.Color
	magenta *color.Color
	gray    *color.Color
}

// Config configuração da status line
type Config struct {
	Model      string
	Mode       string
	WorkDir    string
	MaxTokens  int
	ShowTokens bool
	ShowTime   bool
	ShowTask   bool
	Enabled    bool
}

// New cria nova status line
func New(cfg Config) *StatusLine {
	return &StatusLine{
		model:       cfg.Model,
		mode:        cfg.Mode,
		workDir:     cfg.WorkDir,
		totalTokens: cfg.MaxTokens,
		enabled:     cfg.Enabled,
		cyan:        color.New(color.FgCyan),
		yellow:      color.New(color.FgYellow),
		green:       color.New(color.FgGreen),
		magenta:     color.New(color.FgMagenta),
		gray:        color.New(color.FgHiBlack),
	}
}

// Update atualiza informações da status line
func (s *StatusLine) Update(tokensUsed int, responseTime time.Duration, task string) {
	s.tokensUsed = tokensUsed
	s.responseTime = responseTime
	s.activeTask = task
}

// SetTask define a tarefa ativa
func (s *StatusLine) SetTask(task string) {
	s.activeTask = task
}

// ClearTask limpa a tarefa ativa
func (s *StatusLine) ClearTask() {
	s.activeTask = ""
}

// Render renderiza a status line
func (s *StatusLine) Render() string {
	if !s.enabled {
		return ""
	}

	var parts []string

	// Model indicator
	modelShort := s.getShortModelName()
	parts = append(parts, s.cyan.Sprintf("⚡ %s", modelShort))

	// Mode indicator
	modeIcon := s.getModeIcon()
	parts = append(parts, s.yellow.Sprintf("%s %s", modeIcon, s.mode))

	// Tokens (se houver uso)
	if s.tokensUsed > 0 {
		tokenPercent := float64(s.tokensUsed) / float64(s.totalTokens) * 100
		tokenColor := s.getTokenColor(tokenPercent)
		parts = append(parts, tokenColor.Sprintf("📊 %d/%dk (%.0f%%)",
			s.tokensUsed, s.totalTokens/1000, tokenPercent))
	}

	// Response time
	if s.responseTime > 0 {
		timeStr := s.formatDuration(s.responseTime)
		parts = append(parts, s.green.Sprintf("⏱️  %s", timeStr))
	}

	// Active task
	if s.activeTask != "" {
		taskShort := s.truncate(s.activeTask, 30)
		parts = append(parts, s.magenta.Sprintf("🔧 %s", taskShort))
	}

	// Working directory
	dirShort := s.getShortDir()
	parts = append(parts, s.gray.Sprintf("📁 %s", dirShort))

	return strings.Join(parts, s.gray.Sprint(" │ "))
}

// Display exibe a status line
func (s *StatusLine) Display() {
	if !s.enabled {
		return
	}
	fmt.Println(s.Render())
}

// DisplayInline exibe a status line inline (sem newline)
func (s *StatusLine) DisplayInline() {
	if !s.enabled {
		return
	}
	fmt.Print(s.Render())
}

// Helper functions

func (s *StatusLine) getShortModelName() string {
	// Extrair nome curto do modelo
	parts := strings.Split(s.model, ":")
	if len(parts) > 0 {
		modelName := parts[0]
		// Remover prefixo comum
		modelName = strings.TrimPrefix(modelName, "qwen2.5-coder")
		if modelName == "" {
			modelName = "qwen2.5"
		}
		return modelName
	}
	return s.model
}

func (s *StatusLine) getModeIcon() string {
	switch s.mode {
	case "readonly":
		return "👁️"
	case "interactive":
		return "🤝"
	case "autonomous":
		return "🤖"
	default:
		return "❓"
	}
}

func (s *StatusLine) getTokenColor(percent float64) *color.Color {
	if percent < 50 {
		return s.green
	} else if percent < 80 {
		return s.yellow
	}
	return color.New(color.FgRed)
}

func (s *StatusLine) formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	} else if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}

func (s *StatusLine) getShortDir() string {
	// Pegar apenas o último diretório
	parts := strings.Split(strings.ReplaceAll(s.workDir, "\\", "/"), "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return s.workDir
}

func (s *StatusLine) truncate(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}

// Progress cria uma barra de progresso
func (s *StatusLine) Progress(current, total int, label string) string {
	if !s.enabled {
		return ""
	}

	percent := float64(current) / float64(total)
	barWidth := 20
	filled := int(percent * float64(barWidth))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	return fmt.Sprintf("%s [%s] %d/%d (%.0f%%)",
		label, bar, current, total, percent*100)
}

// Spinner retorna caracteres de spinner animado
func (s *StatusLine) Spinner(step int) string {
	if !s.enabled {
		return ""
	}

	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	return frames[step%len(frames)]
}
