package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// GitOperations ferramenta para operações git
type GitOperations struct {
	workDir string
}

// NewGitOperations cria novo executor de operações git
func NewGitOperations(workDir string) *GitOperations {
	return &GitOperations{
		workDir: workDir,
	}
}

// Name retorna nome da ferramenta
func (g *GitOperations) Name() string {
	return "git_operations"
}

// Description retorna descrição
func (g *GitOperations) Description() string {
	return "Executa operações git (status, diff, commit, etc)"
}

// RequiresConfirmation indica se requer confirmação
func (g *GitOperations) RequiresConfirmation() bool {
	return true // Operações git requerem confirmação
}

// Execute executa a operação git
func (g *GitOperations) Execute(ctx context.Context, params map[string]interface{}) (Result, error) {
	// Tipo de operação
	operation, ok := params["operation"].(string)
	if !ok {
		return NewErrorResult(fmt.Errorf("operation parameter required")), nil
	}

	switch operation {
	case "status":
		return g.gitStatus()
	case "diff":
		return g.gitDiff()
	case "log":
		return g.gitLog(params)
	case "add":
		return g.gitAdd(params)
	case "commit":
		return g.gitCommit(params)
	case "branch":
		return g.gitBranch(params)
	default:
		return NewErrorResult(fmt.Errorf("unknown operation: %s", operation)), nil
	}
}

// gitStatus executa git status
func (g *GitOperations) gitStatus() (Result, error) {
	output, err := g.runGitCommand("status", "--short")
	if err != nil {
		return NewErrorResult(err), nil
	}

	return NewSuccessResult(
		"Git status executado",
		map[string]interface{}{
			"output": output,
		},
	), nil
}

// gitDiff executa git diff
func (g *GitOperations) gitDiff() (Result, error) {
	output, err := g.runGitCommand("diff")
	if err != nil {
		return NewErrorResult(err), nil
	}

	return NewSuccessResult(
		"Git diff executado",
		map[string]interface{}{
			"output": output,
		},
	), nil
}

// gitLog executa git log
func (g *GitOperations) gitLog(params map[string]interface{}) (Result, error) {
	limit, _ := params["limit"].(float64)
	if limit == 0 {
		limit = 10
	}

	output, err := g.runGitCommand("log", "--oneline", fmt.Sprintf("-n%d", int(limit)))
	if err != nil {
		return NewErrorResult(err), nil
	}

	return NewSuccessResult(
		"Git log executado",
		map[string]interface{}{
			"output": output,
		},
	), nil
}

// gitAdd executa git add
func (g *GitOperations) gitAdd(params map[string]interface{}) (Result, error) {
	files, _ := params["files"].(string)
	if files == "" {
		files = "."
	}

	output, err := g.runGitCommand("add", files)
	if err != nil {
		return NewErrorResult(err), nil
	}

	return NewSuccessResult(
		fmt.Sprintf("Arquivos adicionados: %s", files),
		map[string]interface{}{
			"output": output,
			"files":  files,
		},
	), nil
}

// gitCommit executa git commit
func (g *GitOperations) gitCommit(params map[string]interface{}) (Result, error) {
	message, ok := params["message"].(string)
	if !ok {
		return NewErrorResult(fmt.Errorf("commit message required")), nil
	}

	output, err := g.runGitCommand("commit", "-m", message)
	if err != nil {
		return NewErrorResult(err), nil
	}

	return NewSuccessResult(
		"Commit criado com sucesso",
		map[string]interface{}{
			"output":  output,
			"message": message,
		},
	), nil
}

// gitBranch executa operações de branch
func (g *GitOperations) gitBranch(params map[string]interface{}) (Result, error) {
	action, _ := params["action"].(string)

	switch action {
	case "list":
		output, err := g.runGitCommand("branch", "-a")
		if err != nil {
			return NewErrorResult(err), nil
		}
		return NewSuccessResult("Branches listadas", map[string]interface{}{"output": output}), nil

	case "create":
		branchName, ok := params["name"].(string)
		if !ok {
			return NewErrorResult(fmt.Errorf("branch name required")), nil
		}
		output, err := g.runGitCommand("branch", branchName)
		if err != nil {
			return NewErrorResult(err), nil
		}
		return NewSuccessResult(
			fmt.Sprintf("Branch criada: %s", branchName),
			map[string]interface{}{"output": output, "branch": branchName},
		), nil

	case "checkout":
		branchName, ok := params["name"].(string)
		if !ok {
			return NewErrorResult(fmt.Errorf("branch name required")), nil
		}
		output, err := g.runGitCommand("checkout", branchName)
		if err != nil {
			return NewErrorResult(err), nil
		}
		return NewSuccessResult(
			fmt.Sprintf("Checkout para: %s", branchName),
			map[string]interface{}{"output": output, "branch": branchName},
		), nil

	default:
		return NewErrorResult(fmt.Errorf("unknown branch action: %s", action)), nil
	}
}

// runGitCommand executa comando git
func (g *GitOperations) runGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git %s: %w\nstderr: %s", strings.Join(args, " "), err, stderr.String())
	}

	return stdout.String(), nil
}

// IsGitRepository verifica se é repositório git
func (g *GitOperations) IsGitRepository() bool {
	_, err := g.runGitCommand("rev-parse", "--git-dir")
	return err == nil
}
