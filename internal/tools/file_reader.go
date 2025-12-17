package tools

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileReader ferramenta para ler arquivos
type FileReader struct {
	workDir string
}

// NewFileReader cria novo leitor de arquivos
func NewFileReader(workDir string) *FileReader {
	return &FileReader{
		workDir: workDir,
	}
}

// Name retorna nome da ferramenta
func (f *FileReader) Name() string {
	return "file_reader"
}

// Description retorna descrição
func (f *FileReader) Description() string {
	return "Lê conteúdo de arquivos (texto e imagens)"
}

// RequiresConfirmation indica se requer confirmação
func (f *FileReader) RequiresConfirmation() bool {
	return false // Leitura não precisa confirmação
}

// Execute executa a leitura
func (f *FileReader) Execute(ctx context.Context, params map[string]interface{}) (Result, error) {
	// Obter caminho do arquivo
	filePath, ok := params["file_path"].(string)
	if !ok {
		return NewErrorResult(fmt.Errorf("file_path parameter required")), nil
	}

	// Resolver caminho absoluto
	absPath := filepath.Join(f.workDir, filePath)

	// Verificar se arquivo existe
	info, err := os.Stat(absPath)
	if err != nil {
		return NewErrorResult(fmt.Errorf("file not found: %s", filePath)), nil
	}

	if info.IsDir() {
		return NewErrorResult(fmt.Errorf("%s is a directory, not a file", filePath)), nil
	}

	// Detectar se é imagem
	if f.isImage(absPath) {
		return f.readImage(absPath)
	}

	// Ler arquivo de texto
	return f.readText(absPath)
}

// isImage verifica se é arquivo de imagem
func (f *FileReader) isImage(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	imageExts := []string{".png", ".jpg", ".jpeg", ".gif", ".bmp", ".webp"}

	for _, imgExt := range imageExts {
		if ext == imgExt {
			return true
		}
	}

	return false
}

// readText lê arquivo de texto
func (f *FileReader) readText(path string) (Result, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return NewErrorResult(fmt.Errorf("read file: %w", err)), nil
	}

	return NewSuccessResult(
		fmt.Sprintf("Arquivo lido com sucesso: %s", filepath.Base(path)),
		map[string]interface{}{
			"type":    "text",
			"path":    path,
			"content": string(content),
			"size":    len(content),
		},
	), nil
}

// readImage lê imagem e retorna base64
func (f *FileReader) readImage(path string) (Result, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return NewErrorResult(fmt.Errorf("read image: %w", err)), nil
	}

	// Converter para base64
	base64Image := base64.StdEncoding.EncodeToString(content)

	// Detectar MIME type
	mimeType := f.getMimeType(path)

	return NewSuccessResult(
		fmt.Sprintf("Imagem lida com sucesso: %s", filepath.Base(path)),
		map[string]interface{}{
			"type":      "image",
			"path":      path,
			"base64":    base64Image,
			"mime_type": mimeType,
			"size":      len(content),
		},
	), nil
}

// getMimeType retorna MIME type da imagem
func (f *FileReader) getMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	mimeTypes := map[string]string{
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".bmp":  "image/bmp",
		".webp": "image/webp",
	}

	if mime, ok := mimeTypes[ext]; ok {
		return mime
	}

	return "application/octet-stream"
}
