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

// RequiresConfirmation indica se requer confirmação
func (a *AdvancedRefactoring) RequiresConfirmation() bool {
	return false
}

// Execute executa refatoração
func (a *AdvancedRefactoring) Execute(ctx context.Context, params map[string]interface{}) (Result, error) {
	refactorType, ok := params["type"].(string)
	if !ok {
		return Result{
			Success: false,
			Error:   "Tipo de refatoração não especificado",
		}, nil
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
		}, nil
	}
}

// renameSymbol renomeia símbolo (função, variável, tipo)
func (a *AdvancedRefactoring) renameSymbol(params map[string]interface{}) (Result, error) {
	oldName, ok1 := params["old_name"].(string)
	newName, ok2 := params["new_name"].(string)
	filePath, _ := params["file"].(string)

	if !ok1 || !ok2 {
		return Result{
			Success: false,
			Error:   "old_name e new_name são obrigatórios",
		}, nil
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("🔄 Renomeando '%s' para '%s'\n\n", oldName, newName))

	filesChanged := 0

	// If specific file provided, only rename there
	if filePath != "" {
		fullPath := filepath.Join(a.workDir, filePath)
		changed, err := a.renameInFile(fullPath, oldName, newName)
		if err != nil {
			return Result{Success: false, Error: err.Error()}, nil
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
			return Result{Success: false, Error: err.Error()}, nil
		}
	}

	output.WriteString(fmt.Sprintf("\n✅ %d arquivo(s) modificado(s)\n", filesChanged))

	return Result{
		Success: true,
		Message:  output.String(),
	}, nil
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
func (a *AdvancedRefactoring) extractMethod(params map[string]interface{}) (Result, error) {
	filePath, _ := params["file"].(string)
	methodName, _ := params["method_name"].(string)
	startLine, startOk := params["start_line"].(float64)
	endLine, endOk := params["end_line"].(float64)

	if filePath == "" || methodName == "" || !startOk || !endOk {
		return Result{
			Success: false,
			Error:   "Parâmetros obrigatórios: file, method_name, start_line, end_line",
		}, nil
	}

	fullPath := filepath.Join(a.workDir, filePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return Result{Success: false, Error: err.Error()}, nil
	}

	lines := strings.Split(string(content), "\n")
	start := int(startLine) - 1
	end := int(endLine)

	if start < 0 || end > len(lines) || start >= end {
		return Result{
			Success: false,
			Error:   "Intervalo de linhas inválido",
		}, nil
	}

	// Extract code block
	extractedCode := lines[start:end]

	// Detect indentation
	indent := ""
	for _, line := range extractedCode {
		if len(strings.TrimSpace(line)) > 0 {
			indent = line[:len(line)-len(strings.TrimLeft(line, "\t "))]
			break
		}
	}

	// Create new method
	var newMethod strings.Builder
	newMethod.WriteString(fmt.Sprintf("\n%sfunc %s() {\n", indent, methodName))
	for _, line := range extractedCode {
		newMethod.WriteString(line + "\n")
	}
	newMethod.WriteString(fmt.Sprintf("%s}\n", indent))

	// Replace extracted code with method call
	methodCall := fmt.Sprintf("%s%s()", indent, methodName)

	// Build new content
	var newLines []string
	newLines = append(newLines, lines[:start]...)
	newLines = append(newLines, methodCall)
	newLines = append(newLines, lines[end:]...)

	// Add new method at the end of file (before last })
	// Find last non-empty line
	lastIndex := len(newLines) - 1
	for lastIndex >= 0 && strings.TrimSpace(newLines[lastIndex]) == "" {
		lastIndex--
	}

	if lastIndex >= 0 {
		// Insert method before last closing brace or at end
		insertPoint := lastIndex
		if strings.TrimSpace(newLines[lastIndex]) == "}" {
			insertPoint = lastIndex
		}

		result := append(newLines[:insertPoint], newMethod.String())
		result = append(result, newLines[insertPoint:]...)
		newLines = result
	}

	// Write back
	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(fullPath, []byte(newContent), 0644); err != nil {
		return Result{Success: false, Error: err.Error()}, nil
	}

	return Result{
		Success: true,
		Message:  fmt.Sprintf("✅ Método '%s' extraído com sucesso em %s\n", methodName, filePath),
	}, nil
}

// extractClass extrai código para nova classe/struct
func (a *AdvancedRefactoring) extractClass(params map[string]interface{}) (Result, error) {
	sourceFile, _ := params["source_file"].(string)
	className, _ := params["class_name"].(string)
	fieldsRaw, _ := params["fields"].([]interface{})

	if sourceFile == "" || className == "" || len(fieldsRaw) == 0 {
		return Result{
			Success: false,
			Error:   "Parâmetros obrigatórios: source_file, class_name, fields (array de nomes de campos)",
		}, nil
	}

	// Convert fields
	var fields []string
	for _, f := range fieldsRaw {
		if fieldStr, ok := f.(string); ok {
			fields = append(fields, fieldStr)
		}
	}

	sourcePath := filepath.Join(a.workDir, sourceFile)
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return Result{Success: false, Error: err.Error()}, nil
	}

	// For Go files, create new struct
	if strings.HasSuffix(sourceFile, ".go") {
		var output strings.Builder
		output.WriteString(fmt.Sprintf("📦 Extract Class: %s\n\n", className))
		output.WriteString("Novo arquivo sugerido:\n\n")

		// Get package name from source
		lines := strings.Split(string(content), "\n")
		packageName := "main"
		for _, line := range lines {
			if strings.HasPrefix(line, "package ") {
				packageName = strings.TrimSpace(strings.TrimPrefix(line, "package "))
				break
			}
		}

		// Generate new struct file
		output.WriteString(fmt.Sprintf("// Arquivo: %s.go\n", strings.ToLower(className)))
		output.WriteString(fmt.Sprintf("package %s\n\n", packageName))
		output.WriteString(fmt.Sprintf("// %s representa...\n", className))
		output.WriteString(fmt.Sprintf("type %s struct {\n", className))

		// Extract field definitions from source
		for _, field := range fields {
			for _, line := range lines {
				if strings.Contains(line, field) && (strings.Contains(line, "string") || strings.Contains(line, "int") || strings.Contains(line, "bool") || strings.Contains(line, "float")) {
					output.WriteString(fmt.Sprintf("\t%s\n", strings.TrimSpace(line)))
					break
				}
			}
		}

		output.WriteString("}\n\n")

		// Constructor suggestion
		output.WriteString(fmt.Sprintf("// New%s cria nova instância\n", className))
		output.WriteString(fmt.Sprintf("func New%s() *%s {\n", className, className))
		output.WriteString(fmt.Sprintf("\treturn &%s{}\n", className))
		output.WriteString("}\n")

		return Result{
			Success: true,
			Message:  output.String(),
		}, nil
	}

	return Result{
		Success: true,
		Message:  fmt.Sprintf("💡 Extract Class suportado para arquivos .go. Arquivo: %s\n", sourceFile),
	}, nil
}

// inlineSymbol inline de função ou variável
func (a *AdvancedRefactoring) inlineSymbol(params map[string]interface{}) (Result, error) {
	filePath, _ := params["file"].(string)
	symbolName, _ := params["symbol"].(string)

	if filePath == "" || symbolName == "" {
		return Result{
			Success: false,
			Error:   "Parâmetros obrigatórios: file, symbol",
		}, nil
	}

	fullPath := filepath.Join(a.workDir, filePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return Result{Success: false, Error: err.Error()}, nil
	}

	lines := strings.Split(string(content), "\n")

	// Find function definition for Go files
	if strings.HasSuffix(filePath, ".go") {
		var funcBody []string
		var funcStart, funcEnd int
		inFunction := false
		braceCount := 0

		// Find the function definition
		for i, line := range lines {
			if strings.Contains(line, fmt.Sprintf("func %s(", symbolName)) ||
			   strings.Contains(line, fmt.Sprintf("func %s (", symbolName)) {
				funcStart = i
				inFunction = true
				braceCount = strings.Count(line, "{") - strings.Count(line, "}")
				continue
			}

			if inFunction {
				braceCount += strings.Count(line, "{") - strings.Count(line, "}")
				funcBody = append(funcBody, line)

				if braceCount == 0 {
					funcEnd = i
					break
				}
			}
		}

		if len(funcBody) == 0 {
			return Result{
				Success: false,
				Error:   fmt.Sprintf("Função '%s' não encontrada", symbolName),
			}, nil
		}

		// Remove first and last lines (opening and closing braces)
		if len(funcBody) > 1 {
			funcBody = funcBody[:len(funcBody)-1]
		}

		// Remove leading/trailing empty lines
		for len(funcBody) > 0 && strings.TrimSpace(funcBody[0]) == "" {
			funcBody = funcBody[1:]
		}
		for len(funcBody) > 0 && strings.TrimSpace(funcBody[len(funcBody)-1]) == "" {
			funcBody = funcBody[:len(funcBody)-1]
		}

		// Find and replace all function calls
		replacements := 0
		var newLines []string
		skipUntil := -1

		for i, line := range lines {
			// Skip the function definition itself
			if i >= funcStart && i <= funcEnd {
				skipUntil = funcEnd
				continue
			}

			if i <= skipUntil {
				continue
			}

			// Check if this line calls the function
			if strings.Contains(line, fmt.Sprintf("%s()", symbolName)) {
				// Get indentation of the call
				indent := line[:len(line)-len(strings.TrimLeft(line, "\t "))]

				// Replace with function body
				for _, bodyLine := range funcBody {
					newLines = append(newLines, indent+strings.TrimSpace(bodyLine))
				}
				replacements++
			} else {
				newLines = append(newLines, line)
			}
		}

		if replacements == 0 {
			return Result{
				Success: false,
				Error:   fmt.Sprintf("Nenhuma chamada para '%s()' encontrada", symbolName),
			}, nil
		}

		// Write back
		newContent := strings.Join(newLines, "\n")
		if err := os.WriteFile(fullPath, []byte(newContent), 0644); err != nil {
			return Result{Success: false, Error: err.Error()}, nil
		}

		return Result{
			Success: true,
			Message:  fmt.Sprintf("✅ Função '%s' inline executado com sucesso (%d chamadas substituídas)\n", symbolName, replacements),
		}, nil
	}

	return Result{
		Success: false,
		Error:   "Inline suportado apenas para arquivos .go no momento",
	}, nil
}

// moveToFile move símbolo para outro arquivo
func (a *AdvancedRefactoring) moveToFile(params map[string]interface{}) (Result, error) {
	sourceFile, _ := params["source_file"].(string)
	targetFile, _ := params["target_file"].(string)
	symbolName, _ := params["symbol"].(string)

	if sourceFile == "" || targetFile == "" || symbolName == "" {
		return Result{
			Success: false,
			Error:   "Parâmetros obrigatórios: source_file, target_file, symbol",
		}, nil
	}

	sourcePath := filepath.Join(a.workDir, sourceFile)
	targetPath := filepath.Join(a.workDir, targetFile)

	// Read source file
	sourceContent, err := os.ReadFile(sourcePath)
	if err != nil {
		return Result{Success: false, Error: fmt.Sprintf("Erro ao ler arquivo fonte: %v", err)}, nil
	}

	sourceLines := strings.Split(string(sourceContent), "\n")

	// Find and extract the symbol definition
	var symbolLines []string
	var symbolStart, symbolEnd int
	found := false
	inSymbol := false
	braceCount := 0

	// For Go files
	if strings.HasSuffix(sourceFile, ".go") {
		for i, line := range sourceLines {
			// Check for function, type, const, or var
			if !inSymbol {
				if strings.Contains(line, fmt.Sprintf("func %s", symbolName)) ||
				   strings.Contains(line, fmt.Sprintf("type %s ", symbolName)) ||
				   strings.Contains(line, fmt.Sprintf("const %s ", symbolName)) ||
				   strings.Contains(line, fmt.Sprintf("var %s ", symbolName)) {
					symbolStart = i
					inSymbol = true
					found = true

					// Include any comment lines above
					for j := i - 1; j >= 0; j-- {
						if strings.HasPrefix(strings.TrimSpace(sourceLines[j]), "//") {
							symbolStart = j
						} else if strings.TrimSpace(sourceLines[j]) == "" {
							continue
						} else {
							break
						}
					}

					symbolLines = append(symbolLines, line)
					braceCount = strings.Count(line, "{") - strings.Count(line, "}")

					// Check if it's a single-line definition
					if !strings.Contains(line, "func") || braceCount == 0 {
						symbolEnd = i
						break
					}
					continue
				}
			}

			if inSymbol {
				symbolLines = append(symbolLines, line)
				braceCount += strings.Count(line, "{") - strings.Count(line, "}")

				if braceCount == 0 {
					symbolEnd = i
					break
				}
			}
		}
	}

	if !found {
		return Result{
			Success: false,
			Error:   fmt.Sprintf("Símbolo '%s' não encontrado em %s", symbolName, sourceFile),
		}, nil
	}

	// Remove symbol from source file
	var newSourceLines []string
	for i, line := range sourceLines {
		if i < symbolStart || i > symbolEnd {
			newSourceLines = append(newSourceLines, line)
		}
	}

	// Clean up extra blank lines
	var cleanedSourceLines []string
	lastWasEmpty := false
	for _, line := range newSourceLines {
		isEmpty := strings.TrimSpace(line) == ""
		if isEmpty && lastWasEmpty {
			continue
		}
		cleanedSourceLines = append(cleanedSourceLines, line)
		lastWasEmpty = isEmpty
	}

	// Write updated source file
	newSourceContent := strings.Join(cleanedSourceLines, "\n")
	if err := os.WriteFile(sourcePath, []byte(newSourceContent), 0644); err != nil {
		return Result{Success: false, Error: fmt.Sprintf("Erro ao escrever arquivo fonte: %v", err)}, nil
	}

	// Add symbol to target file
	var targetContent string
	var targetLines []string

	// Check if target file exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		// Create new file with same package
		packageName := "main"
		for _, line := range sourceLines {
			if strings.HasPrefix(line, "package ") {
				packageName = strings.TrimSpace(strings.TrimPrefix(line, "package "))
				break
			}
		}

		targetLines = append(targetLines, fmt.Sprintf("package %s", packageName))
		targetLines = append(targetLines, "")
	} else {
		// Read existing target file
		targetContentBytes, err := os.ReadFile(targetPath)
		if err != nil {
			return Result{Success: false, Error: fmt.Sprintf("Erro ao ler arquivo destino: %v", err)}, nil
		}
		targetContent = string(targetContentBytes)
		targetLines = strings.Split(targetContent, "\n")
	}

	// Find insertion point (after package and imports)
	insertPoint := len(targetLines)
	for i, line := range targetLines {
		if strings.HasPrefix(line, "package ") {
			insertPoint = i + 1
			// Skip import block if exists
			for j := i + 1; j < len(targetLines); j++ {
				if strings.HasPrefix(strings.TrimSpace(targetLines[j]), "import") {
					for k := j; k < len(targetLines); k++ {
						if strings.TrimSpace(targetLines[k]) == ")" {
							insertPoint = k + 1
							break
						}
					}
					break
				}
				if strings.TrimSpace(targetLines[j]) != "" {
					break
				}
			}
			break
		}
	}

	// Insert symbol
	var finalLines []string
	finalLines = append(finalLines, targetLines[:insertPoint]...)
	finalLines = append(finalLines, "")
	finalLines = append(finalLines, symbolLines...)
	if insertPoint < len(targetLines) {
		finalLines = append(finalLines, targetLines[insertPoint:]...)
	}

	// Write target file
	newTargetContent := strings.Join(finalLines, "\n")
	if err := os.WriteFile(targetPath, []byte(newTargetContent), 0644); err != nil {
		return Result{Success: false, Error: fmt.Sprintf("Erro ao escrever arquivo destino: %v", err)}, nil
	}

	return Result{
		Success: true,
		Message:  fmt.Sprintf("✅ Símbolo '%s' movido de %s para %s\n", symbolName, sourceFile, targetFile),
	}, nil
}

// findDuplicates encontra código duplicado
func (a *AdvancedRefactoring) findDuplicates() (Result, error) {
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
		return Result{Success: false, Error: err.Error()}, nil
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
	}, nil
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
			// Rename parameters
			"old_name": map[string]interface{}{
				"type":        "string",
				"description": "Nome antigo (para rename)",
			},
			"new_name": map[string]interface{}{
				"type":        "string",
				"description": "Nome novo (para rename)",
			},
			// Extract Method parameters
			"method_name": map[string]interface{}{
				"type":        "string",
				"description": "Nome do novo método (para extract_method)",
			},
			"start_line": map[string]interface{}{
				"type":        "number",
				"description": "Linha inicial do código a extrair (para extract_method)",
			},
			"end_line": map[string]interface{}{
				"type":        "number",
				"description": "Linha final do código a extrair (para extract_method)",
			},
			// Extract Class parameters
			"class_name": map[string]interface{}{
				"type":        "string",
				"description": "Nome da nova classe/struct (para extract_class)",
			},
			"source_file": map[string]interface{}{
				"type":        "string",
				"description": "Arquivo fonte (para extract_class, move)",
			},
			"fields": map[string]interface{}{
				"type":        "array",
				"description": "Lista de campos a extrair (para extract_class)",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			// Inline parameters
			"symbol": map[string]interface{}{
				"type":        "string",
				"description": "Nome do símbolo para inline ou move",
			},
			// Move parameters
			"target_file": map[string]interface{}{
				"type":        "string",
				"description": "Arquivo destino (para move)",
			},
			// General file parameter
			"file": map[string]interface{}{
				"type":        "string",
				"description": "Arquivo específico (para rename, extract_method, inline)",
			},
		},
		"required": []string{"type"},
	}
}
