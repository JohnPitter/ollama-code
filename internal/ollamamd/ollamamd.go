package ollamamd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Level representa o nível hierárquico de um arquivo CLAUDE.md
type Level int

const (
	LevelEnterprise Level = iota // ~/.claude/CLAUDE.md (configurações organizacionais)
	LevelProject                 // /projeto/CLAUDE.md (configurações do projeto)
	LevelLanguage                // /projeto/.claude/go/CLAUDE.md (específico de linguagem)
	LevelLocal                   // /projeto/subdir/CLAUDE.md (configurações locais)
)

// String retorna nome do nível
func (l Level) String() string {
	switch l {
	case LevelEnterprise:
		return "enterprise"
	case LevelProject:
		return "project"
	case LevelLanguage:
		return "language"
	case LevelLocal:
		return "local"
	default:
		return "unknown"
	}
}

// OllamaFile representa um arquivo OLLAMA.md
type OllamaFile struct {
	Path       string    // Caminho completo do arquivo
	Level      Level     // Nível hierárquico
	Content    string    // Conteúdo do arquivo
	Language   string    // Linguagem (se Level == LevelLanguage)
	LoadedAt   time.Time // Timestamp de quando foi carregado
	Sections   map[string]string // Seções do arquivo (título -> conteúdo)
}

// OllamaContext contexto mesclado de múltiplos arquivos OLLAMA.md
type OllamaContext struct {
	Files       []*OllamaFile // Arquivos carregados (ordenados por prioridade)
	Merged      string        // Conteúdo mesclado
	Guidelines  []string      // Diretrizes extraídas
	Preferences map[string]string // Preferências extraídas
	LoadedAt    time.Time     // Timestamp da última atualização
}

// NewOllamaFile cria novo OllamaFile
func NewOllamaFile(path string, level Level) *OllamaFile {
	return &OllamaFile{
		Path:     path,
		Level:    level,
		LoadedAt: time.Now(),
		Sections: make(map[string]string),
	}
}

// Load carrega conteúdo do arquivo
func (cf *OllamaFile) Load() error {
	content, err := os.ReadFile(cf.Path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", cf.Path, err)
	}

	cf.Content = string(content)
	cf.LoadedAt = time.Now()

	// Parse sections
	cf.parseSections()

	return nil
}

// parseSections extrai seções do markdown
func (cf *OllamaFile) parseSections() {
	lines := strings.Split(cf.Content, "\n")
	var currentSection string
	var sectionContent strings.Builder

	for _, line := range lines {
		// Detectar headers markdown (# Header)
		if strings.HasPrefix(line, "#") {
			// Salvar seção anterior se existir
			if currentSection != "" {
				cf.Sections[currentSection] = strings.TrimSpace(sectionContent.String())
				sectionContent.Reset()
			}

			// Extrair título da seção (remover #)
			currentSection = strings.TrimSpace(strings.TrimLeft(line, "#"))
		} else if currentSection != "" {
			sectionContent.WriteString(line + "\n")
		}
	}

	// Salvar última seção
	if currentSection != "" {
		cf.Sections[currentSection] = strings.TrimSpace(sectionContent.String())
	}
}

// GetSection retorna conteúdo de uma seção específica
func (cf *OllamaFile) GetSection(title string) (string, bool) {
	content, exists := cf.Sections[title]
	return content, exists
}

// HasSection verifica se tem uma seção específica
func (cf *OllamaFile) HasSection(title string) bool {
	_, exists := cf.Sections[title]
	return exists
}

// ExtractGuidelines extrai diretrizes do conteúdo
func (cf *OllamaFile) ExtractGuidelines() []string {
	guidelines := []string{}

	// Procurar por seções comuns de guidelines
	commonSections := []string{
		"Guidelines", "Diretrizes", "Rules", "Regras",
		"Best Practices", "Melhores Práticas",
		"Do's and Don'ts", "Fazer e Não Fazer",
	}

	for _, section := range commonSections {
		if content, ok := cf.GetSection(section); ok {
			// Extrair linhas que começam com - ou *
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "*") {
					guideline := strings.TrimLeft(trimmed, "-*")
					guideline = strings.TrimSpace(guideline)
					if guideline != "" {
						guidelines = append(guidelines, guideline)
					}
				}
			}
		}
	}

	return guidelines
}

// ExtractPreferences extrai preferências do conteúdo
func (cf *OllamaFile) ExtractPreferences() map[string]string {
	prefs := make(map[string]string)

	// Procurar por seção de preferências
	preferenceSections := []string{
		"Preferences", "Preferências", "Settings", "Configurações",
	}

	for _, section := range preferenceSections {
		if content, ok := cf.GetSection(section); ok {
			// Extrair pares chave: valor
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				if strings.Contains(line, ":") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						key := strings.TrimSpace(parts[0])
						value := strings.TrimSpace(parts[1])
						// Remover - ou * do início
						key = strings.TrimLeft(key, "-*")
						key = strings.TrimSpace(key)
						if key != "" && value != "" {
							prefs[key] = value
						}
					}
				}
			}
		}
	}

	return prefs
}

// Priority retorna prioridade do arquivo (maior = mais prioritário)
func (cf *OllamaFile) Priority() int {
	// Local tem maior prioridade, Enterprise menor
	switch cf.Level {
	case LevelLocal:
		return 4
	case LevelLanguage:
		return 3
	case LevelProject:
		return 2
	case LevelEnterprise:
		return 1
	default:
		return 0
	}
}

// Exists verifica se o arquivo existe
func (cf *OllamaFile) Exists() bool {
	_, err := os.Stat(cf.Path)
	return err == nil
}

// GetLanguageFromPath extrai linguagem do caminho (ex: .claude/go/CLAUDE.md -> "go")
func GetLanguageFromPath(path string) string {
	// Procurar por padrão .claude/<lang>/CLAUDE.md
	if strings.Contains(path, ".claude") {
		parts := strings.Split(filepath.ToSlash(path), "/")
		for i, part := range parts {
			if part == ".claude" && i+1 < len(parts) {
				return parts[i+1]
			}
		}
	}
	return ""
}
