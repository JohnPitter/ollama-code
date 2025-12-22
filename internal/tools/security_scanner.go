package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// SecurityScanner escaneia código em busca de vulnerabilidades
type SecurityScanner struct {
	workDir string
}

// NewSecurityScanner cria novo scanner de segurança
func NewSecurityScanner(workDir string) *SecurityScanner {
	return &SecurityScanner{
		workDir: workDir,
	}
}

// Name retorna nome da tool
func (s *SecurityScanner) Name() string {
	return "security_scanner"
}

// Description retorna descrição da tool
func (s *SecurityScanner) Description() string {
	return "Escaneia código em busca de vulnerabilidades (CVE, SAST, secrets)"
}

// Execute executa scan de segurança
func (s *SecurityScanner) Execute(ctx context.Context, params map[string]interface{}) Result {
	scanType, ok := params["type"].(string)
	if !ok {
		scanType = "all"
	}

	switch scanType {
	case "all":
		return s.scanAll()
	case "secrets":
		return s.scanSecrets()
	case "sast":
		return s.scanSAST()
	case "dependencies":
		return s.scanDependencies()
	default:
		return Result{
			Success: false,
			Error:   fmt.Sprintf("Tipo de scan desconhecido: %s", scanType),
		}
	}
}

// scanAll executa todos os tipos de scan
func (s *SecurityScanner) scanAll() Result {
	var output strings.Builder
	output.WriteString("🔒 Scan de Segurança Completo\n\n")

	// Secrets scan
	output.WriteString("=== 1. Busca por Secrets ===\n")
	secretsResult := s.scanSecrets()
	output.WriteString(secretsResult.Message)
	output.WriteString("\n")

	// SAST scan
	output.WriteString("=== 2. SAST (Static Analysis) ===\n")
	sastResult := s.scanSAST()
	output.WriteString(sastResult.Message)
	output.WriteString("\n")

	// Dependencies scan
	output.WriteString("=== 3. Vulnerabilidades em Dependências ===\n")
	depsResult := s.scanDependencies()
	output.WriteString(depsResult.Message)

	return Result{
		Success: true,
		Message:  output.String(),
	}
}

// scanSecrets busca por secrets no código
func (s *SecurityScanner) scanSecrets() Result {
	var output strings.Builder
	var findings []string

	// Padrões comuns de secrets
	patterns := map[string]*regexp.Regexp{
		"API Key":           regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[:=]\s*['"]?([a-zA-Z0-9_\-]{20,})['"]?`),
		"AWS Access Key":    regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
		"Password":          regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[:=]\s*['"]([^'"\s]{8,})['"]`),
		"Private Key":       regexp.MustCompile(`-----BEGIN (RSA|DSA|EC|OPENSSH) PRIVATE KEY-----`),
		"JWT Token":         regexp.MustCompile(`eyJ[a-zA-Z0-9_-]+\.eyJ[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+`),
		"GitHub Token":      regexp.MustCompile(`ghp_[a-zA-Z0-9]{36}`),
		"Generic Secret":    regexp.MustCompile(`(?i)(secret|token)\s*[:=]\s*['"]?([a-zA-Z0-9_\-]{20,})['"]?`),
	}

	// Scan files
	err := filepath.Walk(s.workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip directories and binary files
		if info.IsDir() {
			// Skip common directories
			name := filepath.Base(path)
			if name == ".git" || name == "node_modules" || name == "vendor" || name == "build" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only scan text files
		ext := filepath.Ext(path)
		if !isTextFile(ext) {
			return nil
		}

		// Read file
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		// Check patterns
		for secretType, pattern := range patterns {
			matches := pattern.FindAllStringSubmatch(string(content), -1)
			for range matches {
				relPath, _ := filepath.Rel(s.workDir, path)
				finding := fmt.Sprintf("⚠️  %s encontrado em: %s", secretType, relPath)
				findings = append(findings, finding)
			}
		}

		return nil
	})

	if err != nil {
		return Result{
			Success: false,
			Error:   err.Error(),
		}
	}

	if len(findings) == 0 {
		output.WriteString("✅ Nenhum secret encontrado\n")
	} else {
		output.WriteString(fmt.Sprintf("⚠️  %d possíveis secrets encontrados:\n\n", len(findings)))
		for _, finding := range findings {
			output.WriteString(finding + "\n")
		}
		output.WriteString("\n💡 Revise estes achados e considere usar variáveis de ambiente\n")
	}

	return Result{
		Success: true,
		Message:  output.String(),
	}
}

// scanSAST executa análise estática de segurança
func (s *SecurityScanner) scanSAST() Result {
	var output strings.Builder

	projectType := detectProjectType(s.workDir)

	switch projectType {
	case "go":
		// Use gosec for Go
		cmd := exec.Command("gosec", "./...")
		cmd.Dir = s.workDir
		result, err := cmd.CombinedOutput()

		if err != nil {
			// gosec might not be installed
			output.WriteString("💡 Instale gosec: go install github.com/securego/gosec/v2/cmd/gosec@latest\n\n")

			// Use go vet as fallback
			cmd = exec.Command("go", "vet", "./...")
			cmd.Dir = s.workDir
			vetResult, _ := cmd.CombinedOutput()
			output.WriteString("Resultado do go vet:\n")
			output.WriteString(string(vetResult))
		} else {
			output.WriteString(string(result))
		}

	case "nodejs":
		// Use eslint with security plugin
		cmd := exec.Command("npx", "eslint", ".", "--ext", ".js,.ts")
		cmd.Dir = s.workDir
		result, _ := cmd.CombinedOutput()
		output.WriteString(string(result))

		output.WriteString("\n💡 Para análise mais profunda, instale: npm install --save-dev eslint-plugin-security\n")

	case "python":
		// Use bandit
		cmd := exec.Command("bandit", "-r", ".")
		cmd.Dir = s.workDir
		result, err := cmd.CombinedOutput()

		if err != nil {
			output.WriteString("💡 Instale bandit: pip install bandit\n")
		} else {
			output.WriteString(string(result))
		}

	default:
		output.WriteString("⚠️  SAST automático não disponível para este tipo de projeto\n")
		output.WriteString("💡 Considere usar ferramentas como SonarQube ou Semgrep\n")
	}

	return Result{
		Success: true,
		Message:  output.String(),
	}
}

// scanDependencies escaneia vulnerabilidades em dependências
func (s *SecurityScanner) scanDependencies() Result {
	var output strings.Builder

	projectType := detectProjectType(s.workDir)

	switch projectType {
	case "nodejs":
		cmd := exec.Command("npm", "audit")
		cmd.Dir = s.workDir
		result, _ := cmd.CombinedOutput()
		output.WriteString(string(result))

	case "go":
		// Try govulncheck
		cmd := exec.Command("govulncheck", "./...")
		cmd.Dir = s.workDir
		result, err := cmd.CombinedOutput()

		if err != nil {
			output.WriteString("💡 Instale govulncheck: go install golang.org/x/vuln/cmd/govulncheck@latest\n")
		} else {
			output.WriteString(string(result))
		}

	case "python":
		// Try safety
		cmd := exec.Command("safety", "check")
		cmd.Dir = s.workDir
		result, err := cmd.CombinedOutput()

		if err != nil {
			output.WriteString("💡 Instale safety: pip install safety\n")
		} else {
			output.WriteString(string(result))
		}

	default:
		output.WriteString("⚠️  Scan de dependências não disponível para este tipo de projeto\n")
	}

	return Result{
		Success: true,
		Message:  output.String(),
	}
}

// isTextFile verifica se extensão é de arquivo de texto
func isTextFile(ext string) bool {
	textExts := []string{
		".go", ".js", ".ts", ".jsx", ".tsx", ".py", ".rb", ".java",
		".c", ".cpp", ".h", ".hpp", ".cs", ".php", ".rs", ".swift",
		".kt", ".scala", ".sh", ".bash", ".yaml", ".yml", ".json",
		".xml", ".html", ".css", ".scss", ".sql", ".md", ".txt",
		".env", ".config", ".conf",
	}

	for _, textExt := range textExts {
		if ext == textExt {
			return true
		}
	}
	return false
}

// detectProjectType detecta tipo de projeto
func detectProjectType(workDir string) string {
	if _, err := os.Stat(filepath.Join(workDir, "package.json")); err == nil {
		return "nodejs"
	}
	if _, err := os.Stat(filepath.Join(workDir, "go.mod")); err == nil {
		return "go"
	}
	if _, err := os.Stat(filepath.Join(workDir, "requirements.txt")); err == nil {
		return "python"
	}
	return "unknown"
}

// Schema retorna schema JSON da tool
func (s *SecurityScanner) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"type": map[string]interface{}{
				"type":        "string",
				"description": "Tipo de scan: all, secrets, sast, dependencies",
				"enum":        []string{"all", "secrets", "sast", "dependencies"},
			},
		},
	}
}
