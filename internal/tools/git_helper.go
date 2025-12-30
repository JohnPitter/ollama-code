package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// GitHelper fornece operações avançadas de Git
type GitHelper struct {
	workDir string
}

// NewGitHelper cria novo Git Helper
func NewGitHelper(workDir string) *GitHelper {
	return &GitHelper{
		workDir: workDir,
	}
}

// Name retorna nome da tool
func (g *GitHelper) Name() string {
	return "git_helper"
}

// Description retorna descrição da tool
func (g *GitHelper) Description() string {
	return "Operações avançadas de Git: análise de commits, sugestões de branches, detecção de conflitos"
}

// RequiresConfirmation indica se requer confirmação
func (g *GitHelper) RequiresConfirmation() bool {
	return false
}

// Execute executa operação de Git
func (g *GitHelper) Execute(ctx context.Context, params map[string]interface{}) (Result, error) {
	action, ok := params["action"].(string)
	if !ok {
		action = "status"
	}

	switch action {
	case "status":
		return g.getStatus()
	case "analyze_commits":
		return g.analyzeCommits(params)
	case "suggest_branch":
		return g.suggestBranch(params)
	case "detect_conflicts":
		return g.detectConflicts()
	case "generate_commit_message":
		return g.generateCommitMessage()
	case "history":
		return g.getHistory(params)
	case "uncommitted":
		return g.getUncommitted()
	case "branch_info":
		return g.getBranchInfo()
	default:
		return Result{
			Success: false,
			Error:   fmt.Sprintf("Ação desconhecida: %s", action),
		}, nil
	}
}

// getStatus obtém status do repositório
func (g *GitHelper) getStatus() (Result, error) {
	// Check if git repo
	if !g.isGitRepo() {
		return Result{
			Success: false,
			Error:   "Não é um repositório Git",
		}, nil
	}

	// Get current branch
	branch, _ := g.runGitCommand("branch", "--show-current")

	// Get status
	status, _ := g.runGitCommand("status", "--short")

	// Get remote info
	remote, _ := g.runGitCommand("remote", "-v")

	output := fmt.Sprintf(`📊 Status do Git

Branch atual: %s

Arquivos modificados:
%s

Remotes:
%s
`, strings.TrimSpace(branch), status, remote)

	return Result{
		Success: true,
		Message: output,
	}, nil
}

// analyzeCommits analisa commits recentes
func (g *GitHelper) analyzeCommits(params map[string]interface{}) (Result, error) {
	// Get number of commits to analyze
	count := 10
	if c, ok := params["count"].(float64); ok {
		count = int(c)
	}

	// Get commit log
	log, err := g.runGitCommand("log", fmt.Sprintf("-%d", count), "--pretty=format:%h|%an|%ar|%s")
	if err != nil {
		return Result{
			Success: false,
			Error:   "Erro ao obter commits: " + err.Error(),
		}, nil
	}

	lines := strings.Split(strings.TrimSpace(log), "\n")

	output := fmt.Sprintf("📝 Análise de Commits (últimos %d)\n\n", count)

	// Analyze commit patterns
	authors := make(map[string]int)
	prefixes := make(map[string]int)

	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) != 4 {
			continue
		}

		hash, author, time, message := parts[0], parts[1], parts[2], parts[3]

		// Count authors
		authors[author]++

		// Detect commit type prefix (fix:, feat:, etc.)
		if idx := strings.Index(message, ":"); idx > 0 && idx < 10 {
			prefix := message[:idx]
			prefixes[prefix]++
		}

		output += fmt.Sprintf("  %s - %s (%s)\n    %s\n\n", hash, author, time, message)
	}

	// Statistics
	output += "📊 Estatísticas:\n\n"
	output += "Commits por autor:\n"
	for author, count := range authors {
		output += fmt.Sprintf("  %s: %d commits\n", author, count)
	}

	if len(prefixes) > 0 {
		output += "\nTipos de commit detectados:\n"
		for prefix, count := range prefixes {
			output += fmt.Sprintf("  %s: %d\n", prefix, count)
		}
	}

	return Result{
		Success: true,
		Message: output,
	}, nil
}

// suggestBranch sugere nome de branch
func (g *GitHelper) suggestBranch(params map[string]interface{}) (Result, error) {
	taskType, _ := params["type"].(string)
	description, _ := params["description"].(string)

	if taskType == "" {
		taskType = "feature"
	}

	// Get current branch to understand naming convention
	currentBranch, _ := g.runGitCommand("branch", "--show-current")
	currentBranch = strings.TrimSpace(currentBranch)

	// Suggest branch name
	var suggestion string
	if description != "" {
		// Sanitize description
		desc := strings.ToLower(description)
		desc = strings.ReplaceAll(desc, " ", "-")
		desc = strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
				return r
			}
			return -1
		}, desc)

		suggestion = fmt.Sprintf("%s/%s", taskType, desc)
	} else {
		suggestion = fmt.Sprintf("%s/new-task", taskType)
	}

	output := fmt.Sprintf(`🌿 Sugestão de Branch

Branch atual: %s
Tipo: %s
Sugestão: %s

Convenções comuns:
  - feature/nome-funcionalidade
  - bugfix/correcao-bug
  - hotfix/correcao-urgente
  - release/versao
  - docs/documentacao

Para criar:
  git checkout -b %s
`, currentBranch, taskType, suggestion, suggestion)

	return Result{
		Success: true,
		Message: output,
	}, nil
}

// detectConflicts detecta possíveis conflitos
func (g *GitHelper) detectConflicts() (Result, error) {
	// Check for merge conflicts
	conflicts, _ := g.runGitCommand("diff", "--name-only", "--diff-filter=U")

	if conflicts == "" {
		// No active conflicts, check for potential conflicts with remote
		// Get current branch
		branch, _ := g.runGitCommand("branch", "--show-current")
		branch = strings.TrimSpace(branch)

		// Fetch to get latest remote info
		g.runGitCommand("fetch", "--quiet")

		// Check divergence
		ahead, _ := g.runGitCommand("rev-list", "--count", fmt.Sprintf("origin/%s..HEAD", branch))
		behind, _ := g.runGitCommand("rev-list", "--count", fmt.Sprintf("HEAD..origin/%s", branch))

		aheadCount := strings.TrimSpace(ahead)
		behindCount := strings.TrimSpace(behind)

		output := "✅ Nenhum conflito detectado\n\n"

		if aheadCount != "0" {
			output += fmt.Sprintf("⬆️  Você está %s commits à frente da origin/%s\n", aheadCount, branch)
		}
		if behindCount != "0" {
			output += fmt.Sprintf("⬇️  Você está %s commits atrás da origin/%s\n", behindCount, branch)
			output += "\n⚠️  Recomendação: Execute 'git pull' para atualizar\n"
		}

		return Result{
			Success: true,
			Message: output,
		}, nil
	}

	// Active conflicts found
	conflictFiles := strings.Split(strings.TrimSpace(conflicts), "\n")

	output := fmt.Sprintf("⚠️  Conflitos Detectados (%d arquivos)\n\n", len(conflictFiles))

	for _, file := range conflictFiles {
		output += fmt.Sprintf("  ❌ %s\n", file)
	}

	output += "\n💡 Para resolver:\n"
	output += "  1. Edite os arquivos em conflito\n"
	output += "  2. Remova os marcadores de conflito (<<<<, ====, >>>>)\n"
	output += "  3. git add <arquivo>\n"
	output += "  4. git commit\n"

	return Result{
		Success: true,
		Message: output,
	}, nil
}

// generateCommitMessage gera sugestão de mensagem de commit
func (g *GitHelper) generateCommitMessage() (Result, error) {
	// Get staged files
	staged, _ := g.runGitCommand("diff", "--cached", "--name-only")

	if staged == "" {
		return Result{
			Success: false,
			Error:   "Nenhum arquivo staged para commit",
		}, nil
	}

	stagedFiles := strings.Split(strings.TrimSpace(staged), "\n")

	// Get diff stats
	stats, _ := g.runGitCommand("diff", "--cached", "--stat")

	// Analyze changes
	var commitType string
	var scope string

	// Detect type based on file patterns
	hasTests := false
	hasDocs := false
	hasConfig := false

	for _, file := range stagedFiles {
		if strings.Contains(file, "_test.go") || strings.Contains(file, ".test.") {
			hasTests = true
		}
		if strings.Contains(file, "README") || strings.Contains(file, ".md") {
			hasDocs = true
		}
		if strings.Contains(file, "config") || strings.Contains(file, ".json") || strings.Contains(file, ".yaml") {
			hasConfig = true
		}
	}

	// Suggest commit type
	if hasTests {
		commitType = "test"
		scope = "adicionar testes"
	} else if hasDocs {
		commitType = "docs"
		scope = "atualizar documentação"
	} else if hasConfig {
		commitType = "chore"
		scope = "configuração"
	} else {
		commitType = "feat/fix"
		scope = "implementação"
	}

	output := fmt.Sprintf(`💬 Sugestão de Mensagem de Commit

Arquivos staged (%d):
%s

Estatísticas:
%s

Tipo sugerido: %s
Escopo: %s

Sugestões de mensagem:
  %s: %s

Formato Conventional Commits:
  <tipo>[escopo opcional]: <descrição>

  Tipos comuns:
  - feat: Nova funcionalidade
  - fix: Correção de bug
  - docs: Documentação
  - test: Testes
  - refactor: Refatoração
  - chore: Manutenção
`, len(stagedFiles), staged, stats, commitType, scope, commitType, scope)

	return Result{
		Success: true,
		Message: output,
	}, nil
}

// getHistory obtém histórico detalhado
func (g *GitHelper) getHistory(params map[string]interface{}) (Result, error) {
	count := 20
	if c, ok := params["count"].(float64); ok {
		count = int(c)
	}

	file, _ := params["file"].(string)

	var args []string
	args = append(args, "log", fmt.Sprintf("-%d", count), "--pretty=format:%h|%an|%ad|%s", "--date=short")

	if file != "" {
		args = append(args, "--", file)
	}

	log, err := g.runGitCommand(args...)
	if err != nil {
		return Result{
			Success: false,
			Error:   "Erro ao obter histórico: " + err.Error(),
		}, nil
	}

	lines := strings.Split(strings.TrimSpace(log), "\n")

	var output string
	if file != "" {
		output = fmt.Sprintf("📜 Histórico do Arquivo: %s\n\n", file)
	} else {
		output = fmt.Sprintf("📜 Histórico de Commits (últimos %d)\n\n", count)
	}

	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) != 4 {
			continue
		}
		hash, author, date, message := parts[0], parts[1], parts[2], parts[3]
		output += fmt.Sprintf("%s  %s  %s\n  %s\n\n", hash, date, author, message)
	}

	return Result{
		Success: true,
		Message: output,
	}, nil
}

// getUncommitted obtém arquivos não commitados
func (g *GitHelper) getUncommitted() (Result, error) {
	// Get status with porcelain format
	status, err := g.runGitCommand("status", "--porcelain")
	if err != nil {
		return Result{
			Success: false,
			Error:   "Erro ao obter status: " + err.Error(),
		}, nil
	}

	if status == "" {
		return Result{
			Success: true,
			Message: "✅ Nenhuma alteração não commitada\n",
		}, nil
	}

	lines := strings.Split(strings.TrimSpace(status), "\n")

	staged := []string{}
	modified := []string{}
	untracked := []string{}

	for _, line := range lines {
		if len(line) < 3 {
			continue
		}

		statusCode := line[:2]
		file := strings.TrimSpace(line[3:])

		switch {
		case statusCode[0] != ' ' && statusCode[0] != '?':
			staged = append(staged, file)
		case statusCode == "??":
			untracked = append(untracked, file)
		default:
			modified = append(modified, file)
		}
	}

	output := "📝 Alterações Não Commitadas\n\n"

	if len(staged) > 0 {
		output += fmt.Sprintf("✅ Staged (%d):\n", len(staged))
		for _, file := range staged {
			output += fmt.Sprintf("  + %s\n", file)
		}
		output += "\n"
	}

	if len(modified) > 0 {
		output += fmt.Sprintf("📝 Modificados (%d):\n", len(modified))
		for _, file := range modified {
			output += fmt.Sprintf("  M %s\n", file)
		}
		output += "\n"
	}

	if len(untracked) > 0 {
		output += fmt.Sprintf("❓ Não rastreados (%d):\n", len(untracked))
		for _, file := range untracked {
			output += fmt.Sprintf("  ? %s\n", file)
		}
	}

	return Result{
		Success: true,
		Message: output,
	}, nil
}

// getBranchInfo obtém informações sobre branches
func (g *GitHelper) getBranchInfo() (Result, error) {
	// Get all branches
	branches, err := g.runGitCommand("branch", "-a", "-v")
	if err != nil {
		return Result{
			Success: false,
			Error:   "Erro ao obter branches: " + err.Error(),
		}, nil
	}

	// Get current branch
	current, _ := g.runGitCommand("branch", "--show-current")
	current = strings.TrimSpace(current)

	output := fmt.Sprintf("🌿 Informações de Branches\n\nBranch atual: %s\n\nTodas as branches:\n%s\n", current, branches)

	return Result{
		Success: true,
		Message: output,
	}, nil
}

// Helper methods
func (g *GitHelper) isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = g.workDir
	err := cmd.Run()
	return err == nil
}

func (g *GitHelper) runGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.workDir
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// Schema retorna schema JSON da tool
func (g *GitHelper) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Ação: status, analyze_commits, suggest_branch, detect_conflicts, generate_commit_message, history, uncommitted, branch_info",
				"enum":        []string{"status", "analyze_commits", "suggest_branch", "detect_conflicts", "generate_commit_message", "history", "uncommitted", "branch_info"},
			},
			"count": map[string]interface{}{
				"type":        "number",
				"description": "Número de commits para analisar/histórico (padrão: 10 para analyze, 20 para history)",
			},
			"type": map[string]interface{}{
				"type":        "string",
				"description": "Tipo de branch para suggest_branch: feature, bugfix, hotfix, release, docs",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Descrição da tarefa para suggest_branch",
			},
			"file": map[string]interface{}{
				"type":        "string",
				"description": "Arquivo para obter histórico específico",
			},
		},
	}
}
