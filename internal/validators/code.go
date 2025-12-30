package validators

import (
	"fmt"
	"path/filepath"
	"strings"
)

// CodeCleaner limpa e formata código
type CodeCleaner struct{}

// NewCodeCleaner cria novo cleaner
func NewCodeCleaner() *CodeCleaner {
	return &CodeCleaner{}
}

// Clean remove markdown e formata código
func (c *CodeCleaner) Clean(content, filePath string) string {
	lang := c.DetectLanguage(filePath)

	// Remover markdown code blocks
	content = c.removeMarkdownBlocks(content, lang)

	// Remover espaços extras no início e fim
	content = strings.TrimSpace(content)

	// Normalizar line endings
	content = c.normalizeLineEndings(content)

	return content
}

// removeMarkdownBlocks remove blocos de markdown
func (c *CodeCleaner) removeMarkdownBlocks(content, lang string) string {
	// Padrões de markdown code blocks
	patterns := []string{
		"```" + lang + "\n",
		"```" + lang,
		"```\n",
		"```",
	}

	for _, pattern := range patterns {
		content = strings.TrimPrefix(content, pattern)
		content = strings.TrimSuffix(content, "```")
	}

	return strings.TrimSpace(content)
}

// normalizeLineEndings normaliza line endings para \n
func (c *CodeCleaner) normalizeLineEndings(content string) string {
	// Substituir \r\n por \n
	content = strings.ReplaceAll(content, "\r\n", "\n")

	// Remover \r sozinho
	content = strings.ReplaceAll(content, "\r", "\n")

	return content
}

// DetectLanguage detecta linguagem do código pelo filepath
func (c *CodeCleaner) DetectLanguage(filePath string) string {
	ext := filepath.Ext(filePath)

	languageMap := map[string]string{
		".go":   "go",
		".js":   "javascript",
		".jsx":  "javascript",
		".ts":   "typescript",
		".tsx":  "typescript",
		".py":   "python",
		".java": "java",
		".rs":   "rust",
		".c":    "c",
		".cpp":  "cpp",
		".h":    "c",
		".hpp":  "cpp",
		".rb":   "ruby",
		".php":  "php",
		".cs":   "csharp",
		".sh":   "bash",
		".bash": "bash",
		".sql":  "sql",
		".html": "html",
		".css":  "css",
		".yaml": "yaml",
		".yml":  "yaml",
		".json": "json",
		".xml":  "xml",
		".md":   "markdown",
	}

	if lang, ok := languageMap[ext]; ok {
		return lang
	}

	return ""
}

// RemoveComments remove comentários do código (básico)
func (c *CodeCleaner) RemoveComments(content, language string) string {
	switch language {
	case "go", "java", "c", "cpp", "javascript", "typescript", "rust":
		return c.removeCStyleComments(content)
	case "python", "bash", "ruby":
		return c.removePythonStyleComments(content)
	default:
		return content
	}
}

// removeCStyleComments remove comentários estilo C
func (c *CodeCleaner) removeCStyleComments(content string) string {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))

	inBlockComment := false

	for _, line := range lines {
		// Verificar início de bloco de comentário
		if strings.Contains(line, "/*") {
			inBlockComment = true
		}

		// Se não estiver em bloco de comentário
		if !inBlockComment {
			// Remover comentários de linha
			if idx := strings.Index(line, "//"); idx >= 0 {
				line = line[:idx]
			}

			// Adicionar linha se não estiver vazia
			if strings.TrimSpace(line) != "" {
				result = append(result, line)
			}
		}

		// Verificar fim de bloco de comentário
		if strings.Contains(line, "*/") {
			inBlockComment = false
		}
	}

	return strings.Join(result, "\n")
}

// removePythonStyleComments remove comentários estilo Python
func (c *CodeCleaner) removePythonStyleComments(content string) string {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		// Remover comentários de linha
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = line[:idx]
		}

		// Adicionar linha se não estiver vazia
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// AddLineNumbers adiciona números de linha ao código
func (c *CodeCleaner) AddLineNumbers(content string) string {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))

	for i, line := range lines {
		result = append(result, fmt.Sprintf("%4d | %s", i+1, line))
	}

	return strings.Join(result, "\n")
}
