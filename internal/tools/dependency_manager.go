package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DependencyManager gerencia dependências de projetos
type DependencyManager struct {
	workDir string
}

// NewDependencyManager cria novo gerenciador de dependências
func NewDependencyManager(workDir string) *DependencyManager {
	return &DependencyManager{
		workDir: workDir,
	}
}

// Name retorna nome da tool
func (d *DependencyManager) Name() string {
	return "dependency_manager"
}

// Description retorna descrição da tool
func (d *DependencyManager) Description() string {
	return "Gerencia dependências do projeto (install, update, check vulnerabilities)"
}

// Execute executa operação de gerenciamento de dependências
func (d *DependencyManager) Execute(ctx context.Context, params map[string]interface{}) Result {
	operation, ok := params["operation"].(string)
	if !ok {
		operation = "check"
	}

	projectType := d.detectProjectType()

	switch operation {
	case "check":
		return d.checkDependencies(projectType)
	case "install":
		pkg, _ := params["package"].(string)
		return d.installDependency(projectType, pkg)
	case "update":
		return d.updateDependencies(projectType)
	case "audit":
		return d.auditSecurity(projectType)
	default:
		return Result{
			Success: false,
			Message:  "",
			Error:   fmt.Sprintf("Operação desconhecida: %s", operation),
		}
	}
}

// detectProjectType detecta tipo de projeto
func (d *DependencyManager) detectProjectType() string {
	// Check for Node.js
	if _, err := os.Stat(filepath.Join(d.workDir, "package.json")); err == nil {
		return "nodejs"
	}

	// Check for Go
	if _, err := os.Stat(filepath.Join(d.workDir, "go.mod")); err == nil {
		return "go"
	}

	// Check for Python
	if _, err := os.Stat(filepath.Join(d.workDir, "requirements.txt")); err == nil {
		return "python"
	}
	if _, err := os.Stat(filepath.Join(d.workDir, "pyproject.toml")); err == nil {
		return "python"
	}

	// Check for Rust
	if _, err := os.Stat(filepath.Join(d.workDir, "Cargo.toml")); err == nil {
		return "rust"
	}

	return "unknown"
}

// checkDependencies lista dependências atuais
func (d *DependencyManager) checkDependencies(projectType string) Result {
	var cmd *exec.Cmd
	var output strings.Builder

	output.WriteString(fmt.Sprintf("📦 Tipo de Projeto: %s\n\n", projectType))

	switch projectType {
	case "nodejs":
		// List installed packages
		cmd = exec.Command("npm", "list", "--depth=0")
		cmd.Dir = d.workDir
		result, err := cmd.CombinedOutput()
		if err != nil {
			// npm list retorna erro se houver dependências faltando, mas ainda mostra a lista
			output.WriteString(string(result))
		} else {
			output.WriteString(string(result))
		}

		// Check for outdated
		cmd = exec.Command("npm", "outdated")
		cmd.Dir = d.workDir
		outdated, _ := cmd.CombinedOutput()
		if len(outdated) > 0 {
			output.WriteString("\n⚠️  Dependências Desatualizadas:\n")
			output.WriteString(string(outdated))
		}

	case "go":
		// List dependencies
		cmd = exec.Command("go", "list", "-m", "all")
		cmd.Dir = d.workDir
		result, err := cmd.CombinedOutput()
		if err != nil {
			return Result{Success: false, Error: err.Error()}
		}
		output.WriteString(string(result))

	case "python":
		// List installed packages
		cmd = exec.Command("pip", "list")
		cmd.Dir = d.workDir
		result, err := cmd.CombinedOutput()
		if err != nil {
			return Result{Success: false, Error: err.Error()}
		}
		output.WriteString(string(result))

	default:
		return Result{
			Success: false,
			Error:   "Tipo de projeto não suportado ou não detectado",
		}
	}

	return Result{
		Success: true,
		Message:  output.String(),
	}
}

// installDependency instala nova dependência
func (d *DependencyManager) installDependency(projectType, pkg string) Result {
	if pkg == "" {
		return Result{Success: false, Error: "Nome do pacote não especificado"}
	}

	var cmd *exec.Cmd

	switch projectType {
	case "nodejs":
		cmd = exec.Command("npm", "install", pkg, "--save")
	case "go":
		cmd = exec.Command("go", "get", pkg)
	case "python":
		cmd = exec.Command("pip", "install", pkg)
	default:
		return Result{Success: false, Error: "Tipo de projeto não suportado"}
	}

	cmd.Dir = d.workDir
	output, err := cmd.CombinedOutput()

	if err != nil {
		return Result{
			Success: false,
			Message:  string(output),
			Error:   err.Error(),
		}
	}

	return Result{
		Success: true,
		Message:  fmt.Sprintf("✅ Pacote '%s' instalado com sucesso\n\n%s", pkg, string(output)),
	}
}

// updateDependencies atualiza todas as dependências
func (d *DependencyManager) updateDependencies(projectType string) Result {
	var cmd *exec.Cmd

	switch projectType {
	case "nodejs":
		cmd = exec.Command("npm", "update")
	case "go":
		cmd = exec.Command("go", "get", "-u", "./...")
	case "python":
		cmd = exec.Command("pip", "install", "--upgrade", "-r", "requirements.txt")
	default:
		return Result{Success: false, Error: "Tipo de projeto não suportado"}
	}

	cmd.Dir = d.workDir
	output, err := cmd.CombinedOutput()

	if err != nil {
		return Result{
			Success: false,
			Message:  string(output),
			Error:   err.Error(),
		}
	}

	return Result{
		Success: true,
		Message:  fmt.Sprintf("✅ Dependências atualizadas\n\n%s", string(output)),
	}
}

// auditSecurity verifica vulnerabilidades
func (d *DependencyManager) auditSecurity(projectType string) Result {
	var cmd *exec.Cmd
	var output strings.Builder

	output.WriteString("🔒 Auditoria de Segurança\n\n")

	switch projectType {
	case "nodejs":
		cmd = exec.Command("npm", "audit")
		cmd.Dir = d.workDir
		result, err := cmd.CombinedOutput()

		// npm audit retorna erro se houver vulnerabilidades
		output.WriteString(string(result))

		if err != nil {
			output.WriteString("\n⚠️  Vulnerabilidades encontradas! Execute 'npm audit fix' para corrigir.")
		} else {
			output.WriteString("\n✅ Nenhuma vulnerabilidade encontrada")
		}

	case "go":
		// Go não tem audit nativo, mas podemos sugerir govulncheck
		output.WriteString("💡 Para Go, recomendamos instalar govulncheck:\n")
		output.WriteString("   go install golang.org/x/vuln/cmd/govulncheck@latest\n")
		output.WriteString("   govulncheck ./...\n")

	case "python":
		// Python pode usar safety
		cmd = exec.Command("pip", "install", "safety")
		cmd.Dir = d.workDir
		cmd.Run() // Instala se não estiver

		cmd = exec.Command("safety", "check")
		cmd.Dir = d.workDir
		result, _ := cmd.CombinedOutput()
		output.WriteString(string(result))

	default:
		return Result{Success: false, Error: "Tipo de projeto não suportado"}
	}

	return Result{
		Success: true,
		Message:  output.String(),
	}
}

// Schema retorna schema JSON da tool
func (d *DependencyManager) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "Operação: check, install, update, audit",
				"enum":        []string{"check", "install", "update", "audit"},
			},
			"package": map[string]interface{}{
				"type":        "string",
				"description": "Nome do pacote (para install)",
			},
		},
		"required": []string{"operation"},
	}
}
