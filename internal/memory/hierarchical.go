package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Level nível de memória
type Level int

const (
	// LevelEnterprise política enterprise (~/.claude-enterprise/POLICY.md)
	LevelEnterprise Level = iota

	// LevelProject memória do projeto (./CLAUDE.md)
	LevelProject

	// LevelRules regras do projeto (./.claude/rules/*.md)
	LevelRules

	// LevelUser memória do usuário (~/.claude/CLAUDE.md)
	LevelUser

	// LevelLocal memória local (./CLAUDE.local.md)
	LevelLocal
)

// Memory sistema de memória hierárquica
type Memory struct {
	homeDir     string
	projectDir  string
	memories    map[Level]string
	initialized bool
}

// NewMemory cria novo sistema de memória
func NewMemory(projectDir string) *Memory {
	homeDir, _ := os.UserHomeDir()

	return &Memory{
		homeDir:    homeDir,
		projectDir: projectDir,
		memories:   make(map[Level]string),
	}
}

// Load carrega todas as memórias
func (m *Memory) Load() error {
	// 1. Enterprise Policy
	enterprisePath := filepath.Join(m.homeDir, ".claude-enterprise", "POLICY.md")
	if content, err := os.ReadFile(enterprisePath); err == nil {
		m.memories[LevelEnterprise] = string(content)
	}

	// 2. Project Memory
	projectPath := filepath.Join(m.projectDir, "CLAUDE.md")
	if content, err := os.ReadFile(projectPath); err == nil {
		m.memories[LevelProject] = string(content)
	}

	// 3. Project Rules
	rulesDir := filepath.Join(m.projectDir, ".claude", "rules")
	if rulesContent, err := m.loadRules(rulesDir); err == nil {
		m.memories[LevelRules] = rulesContent
	}

	// 4. User Memory
	userPath := filepath.Join(m.homeDir, ".claude", "CLAUDE.md")
	if content, err := os.ReadFile(userPath); err == nil {
		m.memories[LevelUser] = string(content)
	}

	// 5. Local Memory
	localPath := filepath.Join(m.projectDir, "CLAUDE.local.md")
	if content, err := os.ReadFile(localPath); err == nil {
		m.memories[LevelLocal] = string(content)
	}

	m.initialized = true
	return nil
}

// Get retorna memória de um nível
func (m *Memory) Get(level Level) string {
	if !m.initialized {
		m.Load()
	}
	return m.memories[level]
}

// GetAll retorna todas as memórias concatenadas
func (m *Memory) GetAll() string {
	if !m.initialized {
		m.Load()
	}

	var result strings.Builder

	// Ordem hierárquica: Enterprise → Project → Rules → User → Local
	levels := []Level{
		LevelEnterprise,
		LevelProject,
		LevelRules,
		LevelUser,
		LevelLocal,
	}

	for _, level := range levels {
		if content := m.memories[level]; content != "" {
			result.WriteString(fmt.Sprintf("\n# %s\n\n", m.levelName(level)))
			result.WriteString(content)
			result.WriteString("\n")
		}
	}

	return result.String()
}

// Set define memória de um nível
func (m *Memory) Set(level Level, content string) error {
	m.memories[level] = content

	// Persistir
	path := m.levelPath(level)
	if path == "" {
		return fmt.Errorf("cannot persist level %d", level)
	}

	// Criar diretório se necessário
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// Append adiciona à memória de um nível
func (m *Memory) Append(level Level, content string) error {
	current := m.Get(level)
	return m.Set(level, current+"\n"+content)
}

// Clear limpa memória de um nível
func (m *Memory) Clear(level Level) error {
	return m.Set(level, "")
}

// loadRules carrega todos os arquivos de regras
func (m *Memory) loadRules(rulesDir string) (string, error) {
	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		return "", err
	}

	var result strings.Builder

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		path := filepath.Join(rulesDir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		result.WriteString(fmt.Sprintf("\n## Rule: %s\n\n", entry.Name()))
		result.WriteString(string(content))
		result.WriteString("\n")
	}

	return result.String(), nil
}

// levelPath retorna caminho do arquivo de memória
func (m *Memory) levelPath(level Level) string {
	switch level {
	case LevelEnterprise:
		return filepath.Join(m.homeDir, ".claude-enterprise", "POLICY.md")
	case LevelProject:
		return filepath.Join(m.projectDir, "CLAUDE.md")
	case LevelRules:
		return "" // Rules não tem arquivo único
	case LevelUser:
		return filepath.Join(m.homeDir, ".claude", "CLAUDE.md")
	case LevelLocal:
		return filepath.Join(m.projectDir, "CLAUDE.local.md")
	default:
		return ""
	}
}

// levelName retorna nome amigável do nível
func (m *Memory) levelName(level Level) string {
	switch level {
	case LevelEnterprise:
		return "Enterprise Policy"
	case LevelProject:
		return "Project Memory"
	case LevelRules:
		return "Project Rules"
	case LevelUser:
		return "User Memory"
	case LevelLocal:
		return "Local Memory"
	default:
		return "Unknown"
	}
}

// GetSystemPrompt retorna system prompt com todas as memórias
func (m *Memory) GetSystemPrompt() string {
	all := m.GetAll()
	if all == "" {
		return ""
	}

	return fmt.Sprintf(`# Context from Memory

The following context is loaded from hierarchical memory (Enterprise → Project → Rules → User → Local):

%s

Please follow these guidelines and context when responding.`, all)
}

// Stats retorna estatísticas das memórias
func (m *Memory) Stats() map[string]interface{} {
	stats := make(map[string]interface{})

	for level := LevelEnterprise; level <= LevelLocal; level++ {
		content := m.Get(level)
		stats[m.levelName(level)] = map[string]interface{}{
			"loaded": content != "",
			"size":   len(content),
			"lines":  strings.Count(content, "\n"),
		}
	}

	return stats
}
