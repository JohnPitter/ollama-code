package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// TestRunner executa testes automaticamente
type TestRunner struct {
	workDir string
}

// NewTestRunner cria novo test runner
func NewTestRunner(workDir string) *TestRunner {
	return &TestRunner{
		workDir: workDir,
	}
}

// Name retorna nome da tool
func (t *TestRunner) Name() string {
	return "test_runner"
}

// Description retorna descrição da tool
func (t *TestRunner) Description() string {
	return "Executa testes automaticamente e rastreia cobertura"
}

// Execute executa testes
func (t *TestRunner) Execute(ctx context.Context, params map[string]interface{}) Result {
	action, ok := params["action"].(string)
	if !ok {
		action = "run"
	}

	switch action {
	case "run":
		return t.runTests(params)
	case "coverage":
		return t.runCoverage(params)
	case "watch":
		return t.watchTests()
	case "single":
		testPath, _ := params["test"].(string)
		return t.runSingleTest(testPath)
	default:
		return Result{
			Success: false,
			Error:   fmt.Sprintf("Ação desconhecida: %s", action),
		}
	}
}

// runTests executa todos os testes
func (t *TestRunner) runTests(params map[string]interface{}) Result {
	var output strings.Builder
	output.WriteString("🧪 Executando Testes\n\n")

	projectType := t.detectProjectType()

	var cmd *exec.Cmd
	var testOutput []byte
	var err error

	switch projectType {
	case "go":
		verbose, _ := params["verbose"].(bool)
		args := []string{"test"}
		if verbose {
			args = append(args, "-v")
		}
		args = append(args, "./...")

		cmd = exec.Command("go", args...)
		cmd.Dir = t.workDir
		testOutput, err = cmd.CombinedOutput()

	case "nodejs":
		// Try npm test
		cmd = exec.Command("npm", "test")
		cmd.Dir = t.workDir
		testOutput, err = cmd.CombinedOutput()

	case "python":
		// Try pytest
		cmd = exec.Command("pytest", "-v")
		cmd.Dir = t.workDir
		testOutput, err = cmd.CombinedOutput()

		if err != nil {
			// Fallback to unittest
			cmd = exec.Command("python", "-m", "unittest", "discover")
			cmd.Dir = t.workDir
			testOutput, err = cmd.CombinedOutput()
		}

	default:
		return Result{
			Success: false,
			Error:   "Tipo de projeto não suportado para testes",
		}
	}

	output.WriteString(string(testOutput))

	if err != nil {
		output.WriteString(fmt.Sprintf("\n❌ Testes falharam: %s\n", err.Error()))
		return Result{
			Success: false,
			Message:  output.String(),
			Error:   "Alguns testes falharam",
		}
	}

	output.WriteString("\n✅ Todos os testes passaram!\n")

	return Result{
		Success: true,
		Message:  output.String(),
	}
}

// runCoverage executa testes com cobertura
func (t *TestRunner) runCoverage(params map[string]interface{}) Result {
	var output strings.Builder
	output.WriteString("📊 Executando Testes com Cobertura\n\n")

	projectType := t.detectProjectType()

	var cmd *exec.Cmd
	var err error
	var testOutput []byte

	switch projectType {
	case "go":
		// Run tests with coverage
		cmd = exec.Command("go", "test", "-cover", "-coverprofile=coverage.out", "./...")
		cmd.Dir = t.workDir
		testOutput, _ = cmd.CombinedOutput()
		output.WriteString(string(testOutput))

		if err == nil {
			// Generate HTML report
			cmd = exec.Command("go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html")
			cmd.Dir = t.workDir
			cmd.Run()

			output.WriteString("\n✅ Relatório de cobertura gerado: coverage.html\n")
		}

	case "nodejs":
		// Try jest with coverage
		cmd = exec.Command("npm", "test", "--", "--coverage")
		cmd.Dir = t.workDir
		testOutput, err = cmd.CombinedOutput()
		output.WriteString(string(testOutput))

	case "python":
		// Try pytest-cov
		cmd = exec.Command("pytest", "--cov", "--cov-report=html", "--cov-report=term")
		cmd.Dir = t.workDir
		testOutput, err = cmd.CombinedOutput()
		output.WriteString(string(testOutput))

		if err != nil {
			output.WriteString("\n💡 Instale pytest-cov: pip install pytest-cov\n")
		}

	default:
		return Result{
			Success: false,
			Error:   "Tipo de projeto não suportado",
		}
	}

	if err != nil {
		return Result{
			Success: false,
			Message:  output.String(),
			Error:   err.Error(),
		}
	}

	return Result{
		Success: true,
		Message:  output.String(),
	}
}

// runSingleTest executa um teste específico
func (t *TestRunner) runSingleTest(testPath string) Result {
	var output strings.Builder
	output.WriteString(fmt.Sprintf("🧪 Executando Teste: %s\n\n", testPath))

	projectType := t.detectProjectType()

	var cmd *exec.Cmd

	switch projectType {
	case "go":
		// Extract package and test name
		cmd = exec.Command("go", "test", "-v", "-run", testPath)
		cmd.Dir = t.workDir

	case "nodejs":
		cmd = exec.Command("npm", "test", "--", testPath)
		cmd.Dir = t.workDir

	case "python":
		cmd = exec.Command("pytest", "-v", testPath)
		cmd.Dir = t.workDir

	default:
		return Result{
			Success: false,
			Error:   "Tipo de projeto não suportado",
		}
	}

	testOutput, err := cmd.CombinedOutput()
	output.WriteString(string(testOutput))

	if err != nil {
		return Result{
			Success: false,
			Message:  output.String(),
			Error:   err.Error(),
		}
	}

	return Result{
		Success: true,
		Message:  output.String(),
	}
}

// watchTests executa testes em modo watch
func (t *TestRunner) watchTests() Result {
	var output strings.Builder
	output.WriteString("👀 Modo Watch de Testes\n\n")

	projectType := t.detectProjectType()

	switch projectType {
	case "nodejs":
		output.WriteString("💡 Execute: npm test -- --watch\n")
	case "python":
		output.WriteString("💡 Execute: pytest-watch ou ptw\n")
		output.WriteString("   Instale: pip install pytest-watch\n")
	case "go":
		output.WriteString("💡 Use uma ferramenta como 'gow' para watch mode:\n")
		output.WriteString("   go install github.com/mitranim/gow@latest\n")
		output.WriteString("   gow test ./...\n")
	default:
		output.WriteString("⚠️  Watch mode não configurado para este projeto\n")
	}

	return Result{
		Success: true,
		Message:  output.String(),
	}
}

// detectProjectType detecta tipo de projeto
func (t *TestRunner) detectProjectType() string {
	if _, err := os.Stat(filepath.Join(t.workDir, "go.mod")); err == nil {
		return "go"
	}
	if _, err := os.Stat(filepath.Join(t.workDir, "package.json")); err == nil {
		return "nodejs"
	}
	if _, err := os.Stat(filepath.Join(t.workDir, "requirements.txt")); err == nil {
		return "python"
	}
	return "unknown"
}

// Schema retorna schema JSON da tool
// RequiresConfirmation indica se requer confirmaçãofunc (.*) RequiresConfirmation() bool {	return false}
func (t *TestRunner) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Ação: run, coverage, watch, single",
				"enum":        []string{"run", "coverage", "watch", "single"},
			},
			"test": map[string]interface{}{
				"type":        "string",
				"description": "Caminho do teste específico (para single)",
			},
			"verbose": map[string]interface{}{
				"type":        "boolean",
				"description": "Modo verbose",
			},
		},
	}
}
