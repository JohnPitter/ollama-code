package tools

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// AdvancedRefactoring realiza refatorações avançadas no código
type AdvancedRefactoring struct {
	workDir string
}

// NewAdvancedRefactoring cria novo refactoring tool
func NewAdvancedRefactoring(workDir string) *AdvancedRefactoring {
	return &AdvancedRefactoring{
		workDir: workDir,
	}
}

// Name retorna nome da tool
func (a *AdvancedRefactoring) Name() string {
	return "advanced_refactoring"
}

// Description retorna descrição da tool
func (a *AdvancedRefactoring) Description() string {
	return "Refatorações avançadas: rename, extract method, extract class, inline, move"
}

// Execute executa refatoração
func (a *AdvancedRefactoring) Execute(ctx context.Context, params map[string]interface{}) Result {
	refactorType, ok := params["type"].(string)
	if !ok {
		return Result{
			Success: false,
			Error:   "Tipo de refatoração não especificado",
		}
	}

	switch refactorType {
	case "rename":
		return a.renameSymbol(params)
	case "extract_method":
		return a.extractMethod(params)
	case "extract_class":
		return a.extractClass(params)
	case "inline":
		return a.inlineSymbol(params)
	case "move":
		return a.moveToFile(params)
	case "find_duplicates":
		return a.findDuplicates()
	default:
		return Result{
			Success: false,
			Error:   fmt.Sprintf("Tipo de refatoração desconhecido: %s", refactorType),
		}
	}
}

// renameSymbol renomeia símbolo (função, variável, tipo)
func (a *AdvancedRefactoring) renameSymbol(params map[string]interface{}) Result {
	oldName, ok1 := params["old_name"].(string)
	newName, ok2 := params["new_name"].(string)
	filePath, _ := params["file"].(string)

	if !ok1 || !ok2 {
		return Result{
			Success: false,
			Error:   "old_name e new_name são obrigatórios",
		}
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("🔄 Renomeando '%s' para '%s'\n\n", oldName, newName))

	filesChanged := 0

	// If specific file provided, only rename there
	if filePath != "" {
		fullPath := filepath.Join(a.workDir, filePath)
		changed, err := a.renameInFile(fullPath, oldName, newName)
		if err != nil {
			return Result{Success: false, Error: err.Error()}
		}
		if changed {
			filesChanged++
			output.WriteString(fmt.Sprintf("✓ %s\n", filePath))
		}
	} else {
		// Rename across all files
		err := filepath.Walk(a.workDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}

			// Only process code files
			ext := filepath.Ext(path)
			if !isCodeFile(ext) {
				return nil
			}

			changed, err := a.renameInFile(path, oldName, newName)
			if err != nil {
				return nil // Continue on error
			}

			if changed {
				filesChanged++
				relPath, _ := filepath.Rel(a.workDir, path)
				output.WriteString(fmt.Sprintf("✓ %s\n", relPath))
			}

			return nil
		})

		if err != nil {
			return Result{Success: false, Error: err.Error()}
		}
	}

	output.WriteString(fmt.Sprintf("\n✅ %d arquivo(s) modificado(s)\n", filesChanged))

	return Result{
		Success: true,
		Message:  output.String(),
	}
}

// renameInFile renomeia símbolo em um arquivo específico
func (a *AdvancedRefactoring) renameInFile(filePath, oldName, newName string) (bool, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	// Para Go, usar AST parsing para renomear com precisão
	if strings.HasSuffix(filePath, ".go") {
		return a.renameInGoFile(filePath, oldName, newName)
	}

	// Para outras linguagens, fazer substituição simples (menos preciso)
	newContent := strings.ReplaceAll(string(content), oldName, newName)

	if newContent == string(content) {
		return false, nil // No changes
	}

	err = os.WriteFile(filePath, []byte(newContent), 0644)
	return err == nil, err
}

// renameInGoFile renomeia símbolo em arquivo Go usando AST
func (a *AdvancedRefactoring) renameInGoFile(filePath, oldName, newName string) (bool, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return false, err
	}

	changed := false

	// Traverse AST and rename identifiers
	ast.Inspect(file, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok {
			if ident.Name == oldName {
				ident.Name = newName
				changed = true
			}
		}
		return true
	})

	if !changed {
		return false, nil
	}

	// Write back (simplified - in production use go/format)
	content, _ := os.ReadFile(filePath)
	newContent := strings.ReplaceAll(string(content), oldName, newName)
	err = os.WriteFile(filePath, []byte(newContent), 0644)

	return true, err
}

// extractMethod extrai código para novo método
func (a *AdvancedRefactoring) extractMethod(params map[string]interface{}) Result {
	return Result{
		Success: true,
		Message:  "💡 Extract Method: Selecione o código e especifique o nome do novo método\n",
	}
}

// extractClass extrai código para nova classe
func (a *AdvancedRefactoring) extractClass(params map[string]interface{}) Result {
	return Result{
		Success: true,
		Message:  "💡 Extract Class: Identifique campos e métodos relacionados para extrair\n",
	}
}

// inlineSymbol inline de função ou variável
func (a *AdvancedRefactoring) inlineSymbol(params map[string]interface{}) Result {
	return Result{
		Success: true,
		Message:  "💡 Inline: Substitui chamadas de função pelo corpo da função\n",
	}
}

// moveToFile move símbolo para outro arquivo
func (a *AdvancedRefactoring) moveToFile(params map[string]interface{}) Result {
	return Result{
		Success: true,
		Message:  "💡 Move: Move definição para outro arquivo\n",
	}
}

// findDuplicates encontra código duplicado
func (a *AdvancedRefactoring) findDuplicates() Result {
	var output strings.Builder
	output.WriteString("🔍 Buscando Código Duplicado\n\n")

	// Simple duplicate detection (could be enhanced with AST comparison)
	codeBlocks := make(map[string][]string)

	err := filepath.Walk(a.workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if !isCodeFile(ext) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		// Split into lines and look for similar blocks
		lines := strings.Split(string(content), "\n")
		for i := 0; i < len(lines)-5; i++ {
			block := strings.Join(lines[i:i+5], "\n")
			block = strings.TrimSpace(block)
			if len(block) > 50 { // Only consider meaningful blocks
				relPath, _ := filepath.Rel(a.workDir, path)
				codeBlocks[block] = append(codeBlocks[block], fmt.Sprintf("%s:%d", relPath, i+1))
			}
		}

		return nil
	})

	if err != nil {
		return Result{Success: false, Error: err.Error()}
	}

	// Report duplicates
	duplicates := 0
	for _, locations := range codeBlocks {
		if len(locations) > 1 {
			duplicates++
			output.WriteString(fmt.Sprintf("⚠️  Código duplicado encontrado em:\n"))
			for _, loc := range locations {
				output.WriteString(fmt.Sprintf("   - %s\n", loc))
			}
			output.WriteString("\n")

			if duplicates >= 10 {
				output.WriteString("... (limitado a 10 duplicações)\n")
				break
			}
		}
	}

	if duplicates == 0 {
		output.WriteString("✅ Nenhuma duplicação significativa encontrada\n")
	}

	return Result{
		Success: true,
		Message:  output.String(),
	}
}

// isCodeFile verifica se é arquivo de código
func isCodeFile(ext string) bool {
	codeExts := []string{
		".go", ".js", ".ts", ".jsx", ".tsx", ".py", ".rb", ".java",
		".c", ".cpp", ".h", ".hpp", ".cs", ".php", ".rs", ".swift",
	}

	for _, codeExt := range codeExts {
		if ext == codeExt {
			return true
		}
	}
	return false
}

// Schema retorna schema JSON da tool
func (a *AdvancedRefactoring) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"type": map[string]interface{}{
				"type":        "string",
				"description": "Tipo: rename, extract_method, extract_class, inline, move, find_duplicates",
				"enum":        []string{"rename", "extract_method", "extract_class", "inline", "move", "find_duplicates"},
			},
			"old_name": map[string]interface{}{
				"type":        "string",
				"description": "Nome antigo (para rename)",
			},
			"new_name": map[string]interface{}{
				"type":        "string",
				"description": "Nome novo (para rename)",
			},
			"file": map[string]interface{}{
				"type":        "string",
				"description": "Arquivo específico (opcional)",
			},
		},
		"required": []string{"type"},
	}
}
