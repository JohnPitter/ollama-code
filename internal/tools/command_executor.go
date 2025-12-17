package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// CommandExecutor ferramenta para executar comandos shell
type CommandExecutor struct {
	workDir string
	timeout time.Duration
}

// NewCommandExecutor cria novo executor de comandos
func NewCommandExecutor(workDir string, timeout time.Duration) *CommandExecutor {
	if timeout == 0 {
		timeout = 60 * time.Second // Default: 60s
	}

	return &CommandExecutor{
		workDir: workDir,
		timeout: timeout,
	}
}

// Name retorna nome da ferramenta
func (c *CommandExecutor) Name() string {
	return "command_executor"
}

// Description retorna descrição
func (c *CommandExecutor) Description() string {
	return "Executa comandos shell"
}

// RequiresConfirmation indica se requer confirmação
func (c *CommandExecutor) RequiresConfirmation() bool {
	return true // Comandos requerem confirmação
}

// Execute executa o comando
func (c *CommandExecutor) Execute(ctx context.Context, params map[string]interface{}) (Result, error) {
	// Obter comando
	command, ok := params["command"].(string)
	if !ok {
		return NewErrorResult(fmt.Errorf("command parameter required")), nil
	}

	// Criar contexto com timeout
	execCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Executar comando de acordo com o SO
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(execCtx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(execCtx, "sh", "-c", command)
	}

	cmd.Dir = c.workDir

	// Capturar stdout e stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Executar
	startTime := time.Now()
	err := cmd.Run()
	duration := time.Since(startTime)

	// Preparar resultado
	result := map[string]interface{}{
		"command":      command,
		"exit_code":    0,
		"stdout":       stdout.String(),
		"stderr":       stderr.String(),
		"duration_ms":  duration.Milliseconds(),
		"working_dir":  c.workDir,
	}

	// Se houve erro
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result["exit_code"] = exitErr.ExitCode()
		} else {
			return NewErrorResult(fmt.Errorf("execute command: %w", err)), nil
		}
	}

	message := fmt.Sprintf("Comando executado: %s (exit code: %d)", command, result["exit_code"])

	return NewSuccessResult(message, result), nil
}

// IsDangerous verifica se comando é potencialmente perigoso
func (c *CommandExecutor) IsDangerous(command string) bool {
	dangerousPatterns := []string{
		"rm -rf",
		"del /f",
		"format",
		"mkfs",
		"dd if=",
		"> /dev/",
		":(){ :|:& };:", // Fork bomb
	}

	commandLower := strings.ToLower(command)

	for _, pattern := range dangerousPatterns {
		if strings.Contains(commandLower, strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}
