package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// CodeSearcher ferramenta para buscar código
type CodeSearcher struct {
	workDir string
}

// NewCodeSearcher cria novo buscador de código
func NewCodeSearcher(workDir string) *CodeSearcher {
	return &CodeSearcher{
		workDir: workDir,
	}
}

// Name retorna nome da ferramenta
func (c *CodeSearcher) Name() string {
	return "code_searcher"
}

// Description retorna descrição
func (c *CodeSearcher) Description() string {
	return "Busca código no projeto usando ripgrep ou grep"
}

// RequiresConfirmation indica se requer confirmação
func (c *CodeSearcher) RequiresConfirmation() bool {
	return false // Busca não precisa confirmação
}

// Execute executa a busca
func (c *CodeSearcher) Execute(ctx context.Context, params map[string]interface{}) (Result, error) {
	// Obter query de busca
	query, ok := params["query"].(string)
	if !ok {
		return NewErrorResult(fmt.Errorf("query parameter required")), nil
	}

	// Padrão de arquivo (opcional)
	filePattern, _ := params["file_pattern"].(string)

	// Tentar ripgrep primeiro, depois grep
	if c.hasRipgrep() {
		return c.searchWithRipgrep(query, filePattern)
	}

	return c.searchWithGrep(query, filePattern)
}

// hasRipgrep verifica se ripgrep está disponível
func (c *CodeSearcher) hasRipgrep() bool {
	_, err := exec.LookPath("rg")
	return err == nil
}

// searchWithRipgrep busca usando ripgrep
func (c *CodeSearcher) searchWithRipgrep(query, filePattern string) (Result, error) {
	args := []string{
		"--json",
		"--max-count", "50", // Limitar a 50 matches
		query,
	}

	if filePattern != "" {
		args = append(args, "--glob", filePattern)
	}

	cmd := exec.Command("rg", args...)
	cmd.Dir = c.workDir

	output, err := cmd.Output()
	if err != nil {
		// Exit code 1 = sem resultados (não é erro)
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return NewSuccessResult(
				"Nenhum resultado encontrado",
				map[string]interface{}{
					"matches": []string{},
					"count":   0,
				},
			), nil
		}
		return NewErrorResult(fmt.Errorf("ripgrep: %w", err)), nil
	}

	matches := c.parseRipgrepJSON(string(output))

	return NewSuccessResult(
		fmt.Sprintf("Encontrados %d resultados", len(matches)),
		map[string]interface{}{
			"matches": matches,
			"count":   len(matches),
			"tool":    "ripgrep",
		},
	), nil
}

// searchWithGrep busca usando grep padrão
func (c *CodeSearcher) searchWithGrep(query, filePattern string) (Result, error) {
	pattern := "*"
	if filePattern != "" {
		pattern = filePattern
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// Windows: usar findstr
		cmd = exec.Command("findstr", "/S", "/N", "/I", query, pattern)
	} else {
		// Unix: usar grep
		cmd = exec.Command("grep", "-r", "-n", "-i", query, ".")
	}

	cmd.Dir = c.workDir

	output, err := cmd.Output()
	if err != nil {
		// Exit code 1 = sem resultados
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return NewSuccessResult(
				"Nenhum resultado encontrado",
				map[string]interface{}{
					"matches": []string{},
					"count":   0,
				},
			), nil
		}
		return NewErrorResult(fmt.Errorf("grep: %w", err)), nil
	}

	matches := c.parseGrepOutput(string(output))

	return NewSuccessResult(
		fmt.Sprintf("Encontrados %d resultados", len(matches)),
		map[string]interface{}{
			"matches": matches,
			"count":   len(matches),
			"tool":    "grep",
		},
	), nil
}

// parseRipgrepJSON faz parse do output JSON do ripgrep
func (c *CodeSearcher) parseRipgrepJSON(output string) []map[string]interface{} {
	matches := []map[string]interface{}{}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Parse simples (poderia usar encoding/json)
		if strings.Contains(line, `"type":"match"`) {
			matches = append(matches, map[string]interface{}{
				"line": line,
			})
		}
	}

	return matches
}

// parseGrepOutput faz parse do output do grep
func (c *CodeSearcher) parseGrepOutput(output string) []map[string]interface{} {
	matches := []map[string]interface{}{}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Formato: arquivo:linha:conteúdo
		parts := strings.SplitN(line, ":", 3)
		if len(parts) >= 3 {
			matches = append(matches, map[string]interface{}{
				"file":    parts[0],
				"line":    parts[1],
				"content": parts[2],
			})
		}
	}

	return matches
}

// SearchFiles busca arquivos por nome
func (c *CodeSearcher) SearchFiles(pattern string) ([]string, error) {
	var files []string

	err := filepath.Walk(c.workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Ignorar diretórios ocultos e node_modules, etc
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		// Verificar se nome do arquivo match
		matched, _ := filepath.Match(pattern, info.Name())
		if matched {
			relPath, _ := filepath.Rel(c.workDir, path)
			files = append(files, relPath)
		}

		return nil
	})

	return files, err
}
