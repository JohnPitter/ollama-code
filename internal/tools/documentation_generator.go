package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DocumentationGenerator gera documentação automaticamente
type DocumentationGenerator struct {
	workDir string
}

// NewDocumentationGenerator cria novo gerador de documentação
func NewDocumentationGenerator(workDir string) *DocumentationGenerator {
	return &DocumentationGenerator{
		workDir: workDir,
	}
}

// Name retorna nome da tool
func (d *DocumentationGenerator) Name() string {
	return "documentation_generator"
}

// Description retorna descrição da tool
func (d *DocumentationGenerator) Description() string {
	return "Gera documentação automática (GoDoc, JSDoc, README.md)"
}

// Execute executa geração de documentação
func (d *DocumentationGenerator) Execute(ctx context.Context, params map[string]interface{}) Result {
	docType, ok := params["type"].(string)
	if !ok {
		docType = "auto"
	}

	target, _ := params["target"].(string)

	switch docType {
	case "auto":
		return d.generateAuto()
	case "godoc":
		return d.generateGoDoc(target)
	case "jsdoc":
		return d.generateJSDoc(target)
	case "readme":
		return d.generateReadme()
	case "api":
		return d.generateAPIDoc()
	default:
		return Result{
			Success: false,
			Error:   fmt.Sprintf("Tipo de documentação desconhecido: %s", docType),
		}
	}
}

// generateAuto detecta tipo de projeto e gera documentação apropriada
func (d *DocumentationGenerator) generateAuto() Result {
	var output strings.Builder
	output.WriteString("📚 Gerando Documentação Automática\n\n")

	// Check for Go project
	if _, err := os.Stat(filepath.Join(d.workDir, "go.mod")); err == nil {
		output.WriteString("✓ Projeto Go detectado\n")
		result := d.generateGoDoc("")
		if result.Success {
			output.WriteString(result.Message)
		}
	}

	// Check for Node.js project
	if _, err := os.Stat(filepath.Join(d.workDir, "package.json")); err == nil {
		output.WriteString("✓ Projeto Node.js detectado\n")
		result := d.generateJSDoc("")
		if result.Success {
			output.WriteString(result.Message)
		}
	}

	// Always try to generate README if missing
	readmePath := filepath.Join(d.workDir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		output.WriteString("✓ Gerando README.md\n")
		result := d.generateReadme()
		if result.Success {
			output.WriteString(result.Message)
		}
	}

	return Result{
		Success: true,
		Message:  output.String(),
	}
}

// generateGoDoc gera documentação Go
func (d *DocumentationGenerator) generateGoDoc(target string) Result {
	var output strings.Builder
	output.WriteString("📖 Documentação Go (GoDoc)\n\n")

	// Check if godoc is installed
	cmd := exec.Command("go", "doc", "-all")
	if target != "" {
		cmd = exec.Command("go", "doc", target)
	}
	cmd.Dir = d.workDir
	result, err := cmd.CombinedOutput()

	if err != nil {
		// Try to install godoc if not available
		installCmd := exec.Command("go", "install", "golang.org/x/tools/cmd/godoc@latest")
		installCmd.Run()

		output.WriteString("💡 Sugestão: Execute 'godoc -http=:6060' para visualizar docs em http://localhost:6060\n\n")
		output.WriteString(string(result))
	} else {
		output.WriteString(string(result))
	}

	return Result{
		Success: true,
		Message:  output.String(),
	}
}

// generateJSDoc gera documentação JavaScript/TypeScript
func (d *DocumentationGenerator) generateJSDoc(target string) Result {
	var output strings.Builder
	output.WriteString("📖 Documentação JavaScript/TypeScript (JSDoc)\n\n")

	// Check if jsdoc is installed
	cmd := exec.Command("npx", "jsdoc", "-r", ".")
	if target != "" {
		cmd = exec.Command("npx", "jsdoc", target)
	}
	cmd.Dir = d.workDir
	result, err := cmd.CombinedOutput()

	if err != nil {
		output.WriteString("💡 Instale JSDoc: npm install --save-dev jsdoc\n")
		output.WriteString(fmt.Sprintf("Erro: %s\n", err.Error()))
	} else {
		output.WriteString("✅ Documentação gerada em ./out/\n")
		output.WriteString(string(result))
	}

	return Result{
		Success: true,
		Message:  output.String(),
	}
}

// generateReadme gera README.md básico
func (d *DocumentationGenerator) generateReadme() Result {
	readmePath := filepath.Join(d.workDir, "README.md")

	// Get project name from directory
	projectName := filepath.Base(d.workDir)

	content := fmt.Sprintf(`# %s

## Descrição

[Adicione aqui uma descrição do projeto]

## Instalação

[Adicione aqui instruções de instalação]

## Uso

[Adicione aqui exemplos de uso]

## Contribuição

[Adicione aqui instruções para contribuir]

## Licença

[Adicione aqui informações de licença]

---
*Documentação gerada automaticamente por Ollama Code*
`, projectName)

	err := os.WriteFile(readmePath, []byte(content), 0644)
	if err != nil {
		return Result{
			Success: false,
			Error:   fmt.Sprintf("Erro ao criar README.md: %s", err.Error()),
		}
	}

	return Result{
		Success: true,
		Message:  fmt.Sprintf("✅ README.md criado em: %s\n", readmePath),
	}
}

// generateAPIDoc gera documentação de API
func (d *DocumentationGenerator) generateAPIDoc() Result {
	var output strings.Builder
	output.WriteString("📖 Documentação de API\n\n")

	// Check for OpenAPI/Swagger files
	swaggerFiles := []string{
		"swagger.yaml", "swagger.json",
		"openapi.yaml", "openapi.json",
		"api/swagger.yaml", "api/openapi.yaml",
	}

	found := false
	for _, file := range swaggerFiles {
		fullPath := filepath.Join(d.workDir, file)
		if _, err := os.Stat(fullPath); err == nil {
			output.WriteString(fmt.Sprintf("✓ Encontrado: %s\n", file))
			found = true

			// Try to generate docs with swagger-ui or redoc
			output.WriteString("💡 Use swagger-ui ou redoc para visualizar:\n")
			output.WriteString("   npx @redocly/cli preview-docs " + file + "\n")
		}
	}

	if !found {
		output.WriteString("⚠️  Nenhum arquivo OpenAPI/Swagger encontrado\n")
		output.WriteString("💡 Crie um arquivo swagger.yaml ou openapi.yaml para documentar sua API\n")
	}

	return Result{
		Success: true,
		Message:  output.String(),
	}
}

// Schema retorna schema JSON da tool
// RequiresConfirmation indica se requer confirmaçãofunc (.*) RequiresConfirmation() bool {	return false}
func (d *DocumentationGenerator) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"type": map[string]interface{}{
				"type":        "string",
				"description": "Tipo de documentação: auto, godoc, jsdoc, readme, api",
				"enum":        []string{"auto", "godoc", "jsdoc", "readme", "api"},
			},
			"target": map[string]interface{}{
				"type":        "string",
				"description": "Alvo específico (package, arquivo, etc)",
			},
		},
	}
}
