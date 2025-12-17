package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileWriter ferramenta para escrever arquivos
type FileWriter struct {
	workDir string
}

// NewFileWriter cria novo escritor de arquivos
func NewFileWriter(workDir string) *FileWriter {
	return &FileWriter{
		workDir: workDir,
	}
}

// Name retorna nome da ferramenta
func (f *FileWriter) Name() string {
	return "file_writer"
}

// Description retorna descrição
func (f *FileWriter) Description() string {
	return "Cria ou edita arquivos"
}

// RequiresConfirmation indica se requer confirmação
func (f *FileWriter) RequiresConfirmation() bool {
	return true // Escrita requer confirmação
}

// Execute executa a escrita
func (f *FileWriter) Execute(ctx context.Context, params map[string]interface{}) (Result, error) {
	// Obter parâmetros
	filePath, ok := params["file_path"].(string)
	if !ok {
		return NewErrorResult(fmt.Errorf("file_path parameter required")), nil
	}

	content, ok := params["content"].(string)
	if !ok {
		return NewErrorResult(fmt.Errorf("content parameter required")), nil
	}

	// Modo de operação: create, append, replace
	mode, _ := params["mode"].(string)
	if mode == "" {
		mode = "create" // Padrão
	}

	// Resolver caminho absoluto
	absPath := filepath.Join(f.workDir, filePath)

	// Criar diretórios se necessário
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return NewErrorResult(fmt.Errorf("create directories: %w", err)), nil
	}

	// Executar de acordo com o modo
	switch mode {
	case "create":
		return f.createFile(absPath, content)
	case "append":
		return f.appendFile(absPath, content)
	case "replace":
		return f.replaceInFile(absPath, params)
	default:
		return NewErrorResult(fmt.Errorf("unknown mode: %s", mode)), nil
	}
}

// createFile cria novo arquivo ou sobrescreve
func (f *FileWriter) createFile(path, content string) (Result, error) {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return NewErrorResult(fmt.Errorf("write file: %w", err)), nil
	}

	return NewSuccessResult(
		fmt.Sprintf("Arquivo criado/atualizado: %s", filepath.Base(path)),
		map[string]interface{}{
			"path": path,
			"size": len(content),
			"mode": "create",
		},
	), nil
}

// appendFile adiciona conteúdo ao final
func (f *FileWriter) appendFile(path, content string) (Result, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return NewErrorResult(fmt.Errorf("open file: %w", err)), nil
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return NewErrorResult(fmt.Errorf("append to file: %w", err)), nil
	}

	return NewSuccessResult(
		fmt.Sprintf("Conteúdo adicionado a: %s", filepath.Base(path)),
		map[string]interface{}{
			"path": path,
			"size": len(content),
			"mode": "append",
		},
	), nil
}

// replaceInFile substitui texto no arquivo
func (f *FileWriter) replaceInFile(path string, params map[string]interface{}) (Result, error) {
	oldText, ok := params["old_text"].(string)
	if !ok {
		return NewErrorResult(fmt.Errorf("old_text parameter required for replace mode")), nil
	}

	newText, ok := params["new_text"].(string)
	if !ok {
		return NewErrorResult(fmt.Errorf("new_text parameter required for replace mode")), nil
	}

	// Ler arquivo atual
	content, err := os.ReadFile(path)
	if err != nil {
		return NewErrorResult(fmt.Errorf("read file: %w", err)), nil
	}

	// Substituir
	newContent := strings.ReplaceAll(string(content), oldText, newText)

	// Escrever de volta
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return NewErrorResult(fmt.Errorf("write file: %w", err)), nil
	}

	return NewSuccessResult(
		fmt.Sprintf("Texto substituído em: %s", filepath.Base(path)),
		map[string]interface{}{
			"path":         path,
			"replacements": strings.Count(string(content), oldText),
			"mode":         "replace",
		},
	), nil
}
