package validators

import (
	"path/filepath"
	"strings"
)

// FileValidator valida nomes de arquivos
type FileValidator struct{}

// NewFileValidator cria novo validador
func NewFileValidator() *FileValidator {
	return &FileValidator{}
}

// IsValid verifica se o filename é válido
func (v *FileValidator) IsValid(name string) bool {
	if name == "" {
		return false
	}

	// Não pode ter certos caracteres
	invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return false
		}
	}

	// Deve ter extensão válida (opcional, mas recomendado)
	ext := filepath.Ext(name)
	if ext == "" {
		// Sem extensão é válido (pode ser diretório ou script)
		return true
	}

	// Extensões comuns de código
	validExtensions := []string{
		".go", ".js", ".ts", ".jsx", ".tsx",
		".py", ".java", ".c", ".cpp", ".h", ".hpp",
		".rs", ".rb", ".php", ".cs", ".swift",
		".kt", ".scala", ".sh", ".bash",
		".yaml", ".yml", ".json", ".xml",
		".md", ".txt", ".sql", ".html", ".css",
	}

	for _, validExt := range validExtensions {
		if strings.EqualFold(ext, validExt) {
			return true
		}
	}

	// Se não está na lista, ainda aceita (pode ser extensão menos comum)
	return true
}

// SanitizePath sanitiza um path removendo caracteres perigosos
func (v *FileValidator) SanitizePath(path string) string {
	// Normalizar separadores
	path = filepath.Clean(path)

	// Remover tentativas de path traversal
	path = strings.ReplaceAll(path, "..", "")

	return path
}

// ExtractFilename extrai nome do arquivo de uma mensagem
func (v *FileValidator) ExtractFilename(message string) string {
	// Procurar por padrões comuns: "file.go", "path/to/file.go"
	words := strings.Fields(message)

	for _, word := range words {
		// Remove aspas se houver
		word = strings.Trim(word, "\"'`")

		// Se tem extensão, provavelmente é um arquivo
		if filepath.Ext(word) != "" {
			return word
		}
	}

	return ""
}

// IsDirectory verifica se o path é um diretório
func (v *FileValidator) IsDirectory(path string) bool {
	// Se termina com / ou \ é diretório
	return strings.HasSuffix(path, "/") || strings.HasSuffix(path, "\\")
}
