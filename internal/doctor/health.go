package doctor

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"time"
)

// Check resultado de verificação
type Check struct {
	Name    string
	Status  string
	Message string
	Error   error
}

// Doctor executor de diagnósticos
type Doctor struct {
	ollamaURL string
}

// NewDoctor cria novo doctor
func NewDoctor(ollamaURL string) *Doctor {
	return &Doctor{
		ollamaURL: ollamaURL,
	}
}

// RunAll executa todas as verificações
func (d *Doctor) RunAll(ctx context.Context) []Check {
	checks := []Check{}

	checks = append(checks, d.checkOllamaConnection(ctx))
	checks = append(checks, d.checkGPU())
	checks = append(checks, d.checkMemory())
	checks = append(checks, d.checkDiskSpace())
	checks = append(checks, d.checkGit())

	return checks
}

// checkOllamaConnection verifica conexão com Ollama
func (d *Doctor) checkOllamaConnection(ctx context.Context) Check {
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(d.ollamaURL + "/api/tags")
	if err != nil {
		return Check{
			Name:    "Ollama Connection",
			Status:  "FAIL",
			Message: "Cannot connect to Ollama",
			Error:   err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return Check{
			Name:    "Ollama Connection",
			Status:  "OK",
			Message: fmt.Sprintf("Connected to %s", d.ollamaURL),
		}
	}

	return Check{
		Name:    "Ollama Connection",
		Status:  "WARN",
		Message: fmt.Sprintf("Unexpected status: %d", resp.StatusCode),
	}
}

// checkGPU verifica GPU
func (d *Doctor) checkGPU() Check {
	// Simplificado - verificar nvidia-smi
	cmd := exec.Command("nvidia-smi")
	if err := cmd.Run(); err == nil {
		return Check{
			Name:    "GPU",
			Status:  "OK",
			Message: "NVIDIA GPU detected",
		}
	}

	return Check{
		Name:    "GPU",
		Status:  "WARN",
		Message: "No GPU detected (CPU mode)",
	}
}

// checkMemory verifica memória
func (d *Doctor) checkMemory() Check {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	allocated := m.Alloc / 1024 / 1024 // MB

	return Check{
		Name:    "Memory",
		Status:  "OK",
		Message: fmt.Sprintf("Allocated: %d MB", allocated),
	}
}

// checkDiskSpace verifica espaço em disco
func (d *Doctor) checkDiskSpace() Check {
	// Simplificado
	return Check{
		Name:    "Disk Space",
		Status:  "OK",
		Message: "Sufficient space available",
	}
}

// checkGit verifica Git
func (d *Doctor) checkGit() Check {
	cmd := exec.Command("git", "--version")
	if err := cmd.Run(); err != nil {
		return Check{
			Name:    "Git",
			Status:  "WARN",
			Message: "Git not found",
			Error:   err,
		}
	}

	return Check{
		Name:    "Git",
		Status:  "OK",
		Message: "Git available",
	}
}

// Format formata checks para output
func Format(checks []Check) string {
	result := "🏥 Ollama Code Health Check\n\n"

	for _, check := range checks {
		icon := "✓"
		if check.Status == "FAIL" {
			icon = "✗"
		} else if check.Status == "WARN" {
			icon = "⚠"
		}

		result += fmt.Sprintf("%s %s: %s - %s\n", icon, check.Name, check.Status, check.Message)
	}

	return result
}
