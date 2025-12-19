package ollamamd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Loader carrega e mescla arquivos CLAUDE.md hierárquicos
type Loader struct {
	workDir string
	homeDir string
}

// NewLoader cria novo loader
func NewLoader(workDir string) *Loader {
	homeDir, _ := os.UserHomeDir()
	return &Loader{
		workDir: workDir,
		homeDir: homeDir,
	}
}

// Load carrega todos os arquivos CLAUDE.md aplicáveis e retorna contexto mesclado
func (l *Loader) Load() (*OllamaContext, error) {
	files := []*OllamaFile{}

	// 1. Enterprise level (~/.claude/CLAUDE.md)
	if enterpriseFile := l.loadEnterprise(); enterpriseFile != nil {
		files = append(files, enterpriseFile)
	}

	// 2. Project level (workDir/CLAUDE.md)
	if projectFile := l.loadProject(); projectFile != nil {
		files = append(files, projectFile)
	}

	// 3. Language level (workDir/.claude/<lang>/CLAUDE.md)
	languageFiles := l.loadLanguage()
	files = append(files, languageFiles...)

	// 4. Local level (workDir subdirs)
	localFiles := l.loadLocal()
	files = append(files, localFiles...)

	// Ordenar por prioridade (menor prioridade primeiro, será sobrescrito por maior)
	sort.Slice(files, func(i, j int) bool {
		return files[i].Priority() < files[j].Priority()
	})

	// Mesclar arquivos
	context := l.merge(files)

	return context, nil
}

// loadEnterprise carrega arquivo enterprise
func (l *Loader) loadEnterprise() *OllamaFile {
	if l.homeDir == "" {
		return nil
	}

	path := filepath.Join(l.homeDir, ".claude", "CLAUDE.md")
	file := NewOllamaFile(path, LevelEnterprise)

	if file.Exists() {
		if err := file.Load(); err == nil {
			return file
		}
	}

	return nil
}

// loadProject carrega arquivo do projeto
func (l *Loader) loadProject() *OllamaFile {
	path := filepath.Join(l.workDir, "CLAUDE.md")
	file := NewOllamaFile(path, LevelProject)

	if file.Exists() {
		if err := file.Load(); err == nil {
			return file
		}
	}

	return nil
}

// loadLanguage carrega arquivos específicos de linguagem
func (l *Loader) loadLanguage() []*OllamaFile {
	files := []*OllamaFile{}

	claudeDir := filepath.Join(l.workDir, ".claude")
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		return files
	}

	// Listar subdiretórios em .claude/
	entries, err := os.ReadDir(claudeDir)
	if err != nil {
		return files
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Verificar se tem CLAUDE.md dentro
		langPath := filepath.Join(claudeDir, entry.Name(), "CLAUDE.md")
		file := NewOllamaFile(langPath, LevelLanguage)
		file.Language = entry.Name()

		if file.Exists() {
			if err := file.Load(); err == nil {
				files = append(files, file)
			}
		}
	}

	return files
}

// loadLocal carrega arquivos CLAUDE.md em subdiretórios
func (l *Loader) loadLocal() []*OllamaFile {
	files := []*OllamaFile{}

	// Procurar por CLAUDE.md em subdiretórios (não recursivo profundo, max 2 níveis)
	err := filepath.Walk(l.workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Ignorar erros e continuar
		}

		// Pular diretórios escondidos e node_modules, vendor, etc
		if info.IsDir() {
			name := info.Name()
			if name[0] == '.' || name == "node_modules" || name == "vendor" || name == "build" {
				return filepath.SkipDir
			}

			// Limitar profundidade
			rel, _ := filepath.Rel(l.workDir, path)
			depth := strings.Count(rel, string(os.PathSeparator))
			if depth > 2 {
				return filepath.SkipDir
			}

			return nil
		}

		// Verificar se é CLAUDE.md
		if info.Name() == "CLAUDE.md" {
			// Ignorar se for o projeto root (já carregado)
			if filepath.Dir(path) == l.workDir {
				return nil
			}

			file := NewOllamaFile(path, LevelLocal)
			if err := file.Load(); err == nil {
				files = append(files, file)
			}
		}

		return nil
	})

	if err != nil {
		// Log error but continue
		fmt.Printf("Error walking directory: %v\n", err)
	}

	return files
}

// merge mescla múltiplos arquivos em um contexto único
func (l *Loader) merge(files []*OllamaFile) *OllamaContext {
	context := &OllamaContext{
		Files:       files,
		LoadedAt:    time.Now(),
		Guidelines:  []string{},
		Preferences: make(map[string]string),
	}

	var mergedContent strings.Builder
	guidelineSet := make(map[string]bool)

	// Processar arquivos em ordem de prioridade
	for _, file := range files {
		// Adicionar header indicando origem
		mergedContent.WriteString(fmt.Sprintf("\n<!-- From: %s (%s) -->\n", file.Path, file.Level))
		mergedContent.WriteString(file.Content)
		mergedContent.WriteString("\n\n")

		// Extrair e mesclar guidelines (sem duplicatas)
		for _, guideline := range file.ExtractGuidelines() {
			if !guidelineSet[guideline] {
				context.Guidelines = append(context.Guidelines, guideline)
				guidelineSet[guideline] = true
			}
		}

		// Mesclar preferências (arquivos com maior prioridade sobrescrevem)
		for key, value := range file.ExtractPreferences() {
			context.Preferences[key] = value
		}
	}

	context.Merged = strings.TrimSpace(mergedContent.String())

	return context
}

// LoadSingle carrega um único arquivo CLAUDE.md
func (l *Loader) LoadSingle(path string) (*OllamaFile, error) {
	// Determinar nível baseado no caminho
	level := l.determineLevel(path)

	file := NewOllamaFile(path, level)
	if !file.Exists() {
		return nil, fmt.Errorf("file does not exist: %s", path)
	}

	if err := file.Load(); err != nil {
		return nil, err
	}

	return file, nil
}

// determineLevel determina o nível hierárquico baseado no caminho
func (l *Loader) determineLevel(path string) Level {
	absPath, _ := filepath.Abs(path)

	// Enterprise: ~/.claude/CLAUDE.md
	if l.homeDir != "" && strings.Contains(absPath, filepath.Join(l.homeDir, ".claude")) {
		return LevelEnterprise
	}

	// Language: .claude/<lang>/CLAUDE.md
	if strings.Contains(absPath, ".claude") && filepath.Dir(filepath.Dir(absPath)) != filepath.Dir(absPath) {
		return LevelLanguage
	}

	// Project: workDir/CLAUDE.md
	if filepath.Dir(absPath) == l.workDir {
		return LevelProject
	}

	// Local: qualquer outro
	return LevelLocal
}

// Discover encontra todos os arquivos CLAUDE.md sem carregá-los
func (l *Loader) Discover() []string {
	paths := []string{}

	// Enterprise
	if l.homeDir != "" {
		enterprisePath := filepath.Join(l.homeDir, ".claude", "CLAUDE.md")
		if _, err := os.Stat(enterprisePath); err == nil {
			paths = append(paths, enterprisePath)
		}
	}

	// Project
	projectPath := filepath.Join(l.workDir, "CLAUDE.md")
	if _, err := os.Stat(projectPath); err == nil {
		paths = append(paths, projectPath)
	}

	// Language files
	claudeDir := filepath.Join(l.workDir, ".claude")
	if entries, err := os.ReadDir(claudeDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				langPath := filepath.Join(claudeDir, entry.Name(), "CLAUDE.md")
				if _, err := os.Stat(langPath); err == nil {
					paths = append(paths, langPath)
				}
			}
		}
	}

	// Local files (walk subdirectories)
	filepath.Walk(l.workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		if info.Name() == "CLAUDE.md" && filepath.Dir(path) != l.workDir {
			paths = append(paths, path)
		}

		return nil
	})

	return paths
}
