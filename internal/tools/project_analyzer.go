package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProjectAnalyzer ferramenta para analisar estrutura do projeto
type ProjectAnalyzer struct {
	workDir string
}

// NewProjectAnalyzer cria novo analisador de projeto
func NewProjectAnalyzer(workDir string) *ProjectAnalyzer {
	return &ProjectAnalyzer{
		workDir: workDir,
	}
}

// Name retorna nome da ferramenta
func (p *ProjectAnalyzer) Name() string {
	return "project_analyzer"
}

// Description retorna descrição
func (p *ProjectAnalyzer) Description() string {
	return "Analisa estrutura e arquivos do projeto"
}

// RequiresConfirmation indica se requer confirmação
func (p *ProjectAnalyzer) RequiresConfirmation() bool {
	return false
}

// Execute executa a análise
func (p *ProjectAnalyzer) Execute(ctx context.Context, params map[string]interface{}) (Result, error) {
	// Tipo de análise
	analysisType, _ := params["type"].(string)
	if analysisType == "" {
		analysisType = "structure" // Padrão
	}

	switch analysisType {
	case "structure":
		return p.analyzeStructure()
	case "stats":
		return p.analyzeStats()
	case "files":
		return p.listFiles()
	default:
		return NewErrorResult(fmt.Errorf("unknown analysis type: %s", analysisType)), nil
	}
}

// analyzeStructure analisa estrutura de diretórios
func (p *ProjectAnalyzer) analyzeStructure() (Result, error) {
	tree := p.buildDirectoryTree(p.workDir, "", 0, 3) // Max depth 3

	return NewSuccessResult(
		"Estrutura do projeto analisada",
		map[string]interface{}{
			"tree":    tree,
			"rootDir": p.workDir,
		},
	), nil
}

// buildDirectoryTree constrói árvore de diretórios
func (p *ProjectAnalyzer) buildDirectoryTree(dir, prefix string, depth, maxDepth int) []string {
	if depth >= maxDepth {
		return []string{}
	}

	tree := []string{}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return tree
	}

	for i, entry := range entries {
		// Ignorar arquivos ocultos e node_modules
		if strings.HasPrefix(entry.Name(), ".") || entry.Name() == "node_modules" || entry.Name() == "vendor" {
			continue
		}

		isLast := i == len(entries)-1
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		icon := "📄"
		if entry.IsDir() {
			icon = "📁"
		}

		line := fmt.Sprintf("%s%s%s %s", prefix, connector, icon, entry.Name())
		tree = append(tree, line)

		// Recursão para subdiretórios
		if entry.IsDir() {
			newPrefix := prefix
			if isLast {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}

			subTree := p.buildDirectoryTree(
				filepath.Join(dir, entry.Name()),
				newPrefix,
				depth+1,
				maxDepth,
			)
			tree = append(tree, subTree...)
		}
	}

	return tree
}

// analyzeStats analisa estatísticas do projeto
func (p *ProjectAnalyzer) analyzeStats() (Result, error) {
	stats := map[string]interface{}{
		"total_files":   0,
		"total_dirs":    0,
		"total_size":    int64(0),
		"file_types":    make(map[string]int),
		"largest_files": []map[string]interface{}{},
	}

	err := filepath.Walk(p.workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Ignorar diretórios ocultos
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "node_modules" {
				return filepath.SkipDir
			}
			stats["total_dirs"] = stats["total_dirs"].(int) + 1
			return nil
		}

		// Contar arquivo
		stats["total_files"] = stats["total_files"].(int) + 1
		stats["total_size"] = stats["total_size"].(int64) + info.Size()

		// Contar por extensão
		ext := filepath.Ext(info.Name())
		if ext == "" {
			ext = "no_extension"
		}
		fileTypes := stats["file_types"].(map[string]int)
		fileTypes[ext]++

		return nil
	})

	if err != nil {
		return NewErrorResult(fmt.Errorf("walk directory: %w", err)), nil
	}

	return NewSuccessResult(
		"Estatísticas do projeto calculadas",
		stats,
	), nil
}

// listFiles lista todos os arquivos
func (p *ProjectAnalyzer) listFiles() (Result, error) {
	files := []string{}

	err := filepath.Walk(p.workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		relPath, _ := filepath.Rel(p.workDir, path)
		files = append(files, relPath)

		return nil
	})

	if err != nil {
		return NewErrorResult(fmt.Errorf("walk directory: %w", err)), nil
	}

	return NewSuccessResult(
		fmt.Sprintf("Listados %d arquivos", len(files)),
		map[string]interface{}{
			"files": files,
			"count": len(files),
		},
	), nil
}
