package tools

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// CodeFormatter formata código automaticamente
type CodeFormatter struct {
	workDir string
}

// NewCodeFormatter cria novo formatador de código
func NewCodeFormatter(workDir string) *CodeFormatter {
	return &CodeFormatter{
		workDir: workDir,
	}
}

// Name retorna nome da tool
func (c *CodeFormatter) Name() string {
	return "code_formatter"
}

// Description retorna descrição da tool
func (c *CodeFormatter) Description() string {
	return "Formata código automaticamente para múltiplas linguagens (Go, JS, Python, etc.)"
}

// RequiresConfirmation indica se requer confirmação
func (c *CodeFormatter) RequiresConfirmation() bool {
	return false
}

// Execute executa operação de formatação
func (c *CodeFormatter) Execute(ctx context.Context, params map[string]interface{}) (Result, error) {
	action, ok := params["action"].(string)
	if !ok {
		action = "format"
	}

	switch action {
	case "format":
		return c.formatCode(params)
	case "check":
		return c.checkFormatting(params)
	case "detect":
		return c.detectFormatters()
	default:
		return Result{
			Success: false,
			Error:   fmt.Sprintf("Ação desconhecida: %s", action),
		}, nil
	}
}

// formatCode formata código
func (c *CodeFormatter) formatCode(params map[string]interface{}) (Result, error) {
	language, _ := params["language"].(string)
	file, _ := params["file"].(string)
	path, _ := params["path"].(string)

	// Auto-detect language if not specified
	if language == "" && file != "" {
		language = c.detectLanguage(file)
	}

	if language == "" && path == "" && file == "" {
		// Format all files in workDir
		return c.formatAllFiles()
	}

	var formatter string
	var args []string
	var description string

	switch language {
	case "go":
		formatter = "gofmt"
		if file != "" {
			args = []string{"-w", file}
			description = "Formatando arquivo Go: " + file
		} else if path != "" {
			args = []string{"-w", path}
			description = "Formatando arquivos Go em: " + path
		} else {
			args = []string{"-w", "."}
			description = "Formatando todos os arquivos Go"
		}

	case "javascript", "js", "typescript", "ts":
		// Try prettier first, fallback to standard
		if c.isCommandAvailable("prettier") {
			formatter = "prettier"
			if file != "" {
				args = []string{"--write", file}
			} else if path != "" {
				args = []string{"--write", path + "/**/*.{js,ts,jsx,tsx}"}
			} else {
				args = []string{"--write", "**/*.{js,ts,jsx,tsx}"}
			}
			description = "Formatando com Prettier"
		} else {
			return Result{
				Success: false,
				Error:   "Prettier não encontrado. Instale com: npm install -g prettier",
			}, nil
		}

	case "python", "py":
		// Try black first, fallback to autopep8
		if c.isCommandAvailable("black") {
			formatter = "black"
			if file != "" {
				args = []string{file}
			} else if path != "" {
				args = []string{path}
			} else {
				args = []string{"."}
			}
			description = "Formatando com Black"
		} else if c.isCommandAvailable("autopep8") {
			formatter = "autopep8"
			if file != "" {
				args = []string{"-i", file}
			} else {
				args = []string{"-i", "-r", "."}
			}
			description = "Formatando com autopep8"
		} else {
			return Result{
				Success: false,
				Error:   "Black ou autopep8 não encontrado. Instale com: pip install black",
			}, nil
		}

	case "rust", "rs":
		formatter = "rustfmt"
		if file != "" {
			args = []string{file}
		} else {
			// Use cargo fmt for rust projects
			formatter = "cargo"
			args = []string{"fmt"}
		}
		description = "Formatando Rust"

	case "java":
		if c.isCommandAvailable("google-java-format") {
			formatter = "google-java-format"
			if file != "" {
				args = []string{"-i", file}
			} else {
				return Result{
					Success: false,
					Error:   "Especifique um arquivo para formatar Java",
				}, nil
			}
			description = "Formatando Java"
		} else {
			return Result{
				Success: false,
				Error:   "google-java-format não encontrado",
			}, nil
		}

	case "c", "cpp", "c++":
		if c.isCommandAvailable("clang-format") {
			formatter = "clang-format"
			if file != "" {
				args = []string{"-i", file}
			} else {
				return Result{
					Success: false,
					Error:   "Especifique um arquivo para formatar C/C++",
				}, nil
			}
			description = "Formatando C/C++"
		} else {
			return Result{
				Success: false,
				Error:   "clang-format não encontrado",
			}, nil
		}

	default:
		return Result{
			Success: false,
			Error:   fmt.Sprintf("Linguagem não suportada: %s", language),
		}, nil
	}

	// Execute formatter
	cmd := exec.Command(formatter, args...)
	cmd.Dir = c.workDir
	output, err := cmd.CombinedOutput()

	if err != nil {
		return Result{
			Success: false,
			Error:   fmt.Sprintf("Erro ao formatar: %s\n%s", err.Error(), string(output)),
		}, nil
	}

	return Result{
		Success: true,
		Message: fmt.Sprintf("✅ %s\n\nSaída:\n%s", description, string(output)),
	}, nil
}

// checkFormatting verifica se código está formatado
func (c *CodeFormatter) checkFormatting(params map[string]interface{}) (Result, error) {
	language, _ := params["language"].(string)
	file, _ := params["file"].(string)

	if language == "" && file != "" {
		language = c.detectLanguage(file)
	}

	var checker string
	var args []string

	switch language {
	case "go":
		checker = "gofmt"
		if file != "" {
			args = []string{"-l", file}
		} else {
			args = []string{"-l", "."}
		}

	case "javascript", "js", "typescript", "ts":
		if c.isCommandAvailable("prettier") {
			checker = "prettier"
			if file != "" {
				args = []string{"--check", file}
			} else {
				args = []string{"--check", "**/*.{js,ts,jsx,tsx}"}
			}
		} else {
			return Result{
				Success: false,
				Error:   "Prettier não encontrado",
			}, nil
		}

	case "python", "py":
		if c.isCommandAvailable("black") {
			checker = "black"
			if file != "" {
				args = []string{"--check", file}
			} else {
				args = []string{"--check", "."}
			}
		} else {
			return Result{
				Success: false,
				Error:   "Black não encontrado",
			}, nil
		}

	default:
		return Result{
			Success: false,
			Error:   fmt.Sprintf("Verificação não suportada para: %s", language),
		}, nil
	}

	cmd := exec.Command(checker, args...)
	cmd.Dir = c.workDir
	output, err := cmd.CombinedOutput()

	if err != nil {
		return Result{
			Success: false,
			Message: fmt.Sprintf("⚠️  Arquivos precisam de formatação:\n%s", string(output)),
		}, nil
	}

	return Result{
		Success: true,
		Message: "✅ Todos os arquivos estão formatados corretamente\n",
	}, nil
}

// detectFormatters detecta formatadores disponíveis
func (c *CodeFormatter) detectFormatters() (Result, error) {
	formatters := []struct {
		name     string
		language string
		command  string
	}{
		{"gofmt", "Go", "gofmt"},
		{"Prettier", "JS/TS", "prettier"},
		{"Black", "Python", "black"},
		{"autopep8", "Python", "autopep8"},
		{"rustfmt", "Rust", "rustfmt"},
		{"clang-format", "C/C++", "clang-format"},
		{"google-java-format", "Java", "google-java-format"},
	}

	output := "🔍 Formatadores Disponíveis:\n\n"

	available := []string{}
	missing := []string{}

	for _, formatter := range formatters {
		if c.isCommandAvailable(formatter.command) {
			available = append(available, fmt.Sprintf("  ✅ %s (%s)", formatter.name, formatter.language))
		} else {
			missing = append(missing, fmt.Sprintf("  ❌ %s (%s)", formatter.name, formatter.language))
		}
	}

	if len(available) > 0 {
		output += "Instalados:\n" + strings.Join(available, "\n") + "\n\n"
	}

	if len(missing) > 0 {
		output += "Não instalados:\n" + strings.Join(missing, "\n") + "\n\n"
		output += "💡 Dicas de instalação:\n"
		output += "  - Prettier: npm install -g prettier\n"
		output += "  - Black: pip install black\n"
		output += "  - rustfmt: rustup component add rustfmt\n"
		output += "  - clang-format: apt/brew install clang-format\n"
	}

	return Result{
		Success: true,
		Message: output,
	}, nil
}

// formatAllFiles formata todos os arquivos no diretório
func (c *CodeFormatter) formatAllFiles() (Result, error) {
	results := []string{}

	// Try formatting Go files
	if c.isCommandAvailable("gofmt") {
		goFiles, _ := filepath.Glob(filepath.Join(c.workDir, "**/*.go"))
		if len(goFiles) > 0 {
			result, _ := c.formatCode(map[string]interface{}{"language": "go"})
			if result.Success {
				results = append(results, "✅ Go files formatted")
			}
		}
	}

	// Try formatting JS/TS files
	if c.isCommandAvailable("prettier") {
		result, _ := c.formatCode(map[string]interface{}{"language": "javascript"})
		if result.Success {
			results = append(results, "✅ JS/TS files formatted")
		}
	}

	// Try formatting Python files
	if c.isCommandAvailable("black") {
		pyFiles, _ := filepath.Glob(filepath.Join(c.workDir, "**/*.py"))
		if len(pyFiles) > 0 {
			result, _ := c.formatCode(map[string]interface{}{"language": "python"})
			if result.Success {
				results = append(results, "✅ Python files formatted")
			}
		}
	}

	if len(results) == 0 {
		return Result{
			Success: false,
			Error:   "Nenhum formatador disponível ou nenhum arquivo para formatar",
		}, nil
	}

	return Result{
		Success: true,
		Message: "📝 Formatação Completa:\n\n" + strings.Join(results, "\n") + "\n",
	}, nil
}

// Helper methods
func (c *CodeFormatter) detectLanguage(file string) string {
	ext := filepath.Ext(file)
	switch ext {
	case ".go":
		return "go"
	case ".js", ".jsx":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".py":
		return "python"
	case ".rs":
		return "rust"
	case ".java":
		return "java"
	case ".c", ".h":
		return "c"
	case ".cpp", ".hpp", ".cc", ".cxx":
		return "cpp"
	default:
		return ""
	}
}

func (c *CodeFormatter) isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// Schema retorna schema JSON da tool
func (c *CodeFormatter) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Ação: format, check, detect",
				"enum":        []string{"format", "check", "detect"},
			},
			"language": map[string]interface{}{
				"type":        "string",
				"description": "Linguagem: go, javascript, python, rust, java, c, cpp",
				"enum":        []string{"go", "javascript", "js", "typescript", "ts", "python", "py", "rust", "rs", "java", "c", "cpp", "c++"},
			},
			"file": map[string]interface{}{
				"type":        "string",
				"description": "Arquivo específico para formatar",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Caminho para formatar (diretório)",
			},
		},
	}
}
