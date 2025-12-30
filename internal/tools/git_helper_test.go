package tools

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestGitHelper_Name(t *testing.T) {
	gh := NewGitHelper(".")
	if gh.Name() != "git_helper" {
		t.Errorf("Expected name 'git_helper', got '%s'", gh.Name())
	}
}

func TestGitHelper_Description(t *testing.T) {
	gh := NewGitHelper(".")
	desc := gh.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
	if !strings.Contains(desc, "Git") {
		t.Error("Description should mention 'Git'")
	}
}

func TestGitHelper_RequiresConfirmation(t *testing.T) {
	gh := NewGitHelper(".")
	if gh.RequiresConfirmation() {
		t.Error("GitHelper should not require confirmation")
	}
}

func TestGitHelper_Schema(t *testing.T) {
	gh := NewGitHelper(".")
	schema := gh.Schema()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema properties should be a map")
	}

	if _, exists := props["action"]; !exists {
		t.Error("Schema should have 'action' property")
	}
}

func TestGitHelper_Execute_InvalidAction(t *testing.T) {
	gh := NewGitHelper(".")
	ctx := context.Background()

	params := map[string]interface{}{
		"action": "invalid_action",
	}

	result, _ := gh.Execute(ctx, params)

	if result.Success {
		t.Error("Result should not be successful for invalid action")
	}

	if !strings.Contains(result.Error, "desconhecida") {
		t.Error("Error should mention unknown action")
	}
}

func TestGitHelper_GetStatus_NotGitRepo(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-notgit-*")
	defer os.RemoveAll(tmpDir)

	gh := NewGitHelper(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"action": "status",
	}

	result, _ := gh.Execute(ctx, params)

	if result.Success {
		t.Error("Result should not be successful for non-git repo")
	}

	if !strings.Contains(result.Error, "Git") {
		t.Error("Error should mention Git repository")
	}
}

func TestGitHelper_GetStatus_GitRepo(t *testing.T) {
	// Only run if we're in a git repo
	if !isGitRepoAvailable() {
		t.Skip("Not in a git repository, skipping test")
	}

	gh := NewGitHelper(".")
	ctx := context.Background()

	params := map[string]interface{}{
		"action": "status",
	}

	result, _ := gh.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful in git repo, got error: %s", result.Error)
	}

	if !strings.Contains(result.Message, "Branch atual") {
		t.Error("Status should show current branch")
	}
}

func TestGitHelper_AnalyzeCommits(t *testing.T) {
	if !isGitRepoAvailable() {
		t.Skip("Not in a git repository, skipping test")
	}

	gh := NewGitHelper(".")
	ctx := context.Background()

	params := map[string]interface{}{
		"action": "analyze_commits",
		"count":  5.0,
	}

	result, _ := gh.Execute(ctx, params)

	// Skip if git command fails (may have no commits or other issues)
	if !result.Success {
		if strings.Contains(result.Error, "Erro ao obter commits") {
			t.Skip("Git repository may have no commits, skipping test")
		}
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	if result.Success && !strings.Contains(result.Message, "Análise de Commits") {
		t.Error("Should show commit analysis header")
	}
}

func TestGitHelper_SuggestBranch_WithDescription(t *testing.T) {
	gh := NewGitHelper(".")
	ctx := context.Background()

	params := map[string]interface{}{
		"action":      "suggest_branch",
		"type":        "feature",
		"description": "Add User Authentication",
	}

	result, _ := gh.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	if !strings.Contains(result.Message, "feature/") {
		t.Error("Should suggest feature branch")
	}

	if !strings.Contains(result.Message, "add-user-authentication") {
		t.Error("Should include sanitized description")
	}
}

func TestGitHelper_SuggestBranch_NoDescription(t *testing.T) {
	gh := NewGitHelper(".")
	ctx := context.Background()

	params := map[string]interface{}{
		"action": "suggest_branch",
		"type":   "bugfix",
	}

	result, _ := gh.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	if !strings.Contains(result.Message, "bugfix/") {
		t.Error("Should suggest bugfix branch")
	}
}

func TestGitHelper_DetectConflicts(t *testing.T) {
	if !isGitRepoAvailable() {
		t.Skip("Not in a git repository, skipping test")
	}

	gh := NewGitHelper(".")
	ctx := context.Background()

	params := map[string]interface{}{
		"action": "detect_conflicts",
	}

	result, _ := gh.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	// Should either show conflicts or no conflicts
	hasConflictMessage := strings.Contains(result.Message, "Conflitos") ||
		strings.Contains(result.Message, "Nenhum conflito")

	if !hasConflictMessage {
		t.Error("Should show conflict detection result")
	}
}

func TestGitHelper_GenerateCommitMessage_NoStaged(t *testing.T) {
	if !isGitRepoAvailable() {
		t.Skip("Not in a git repository, skipping test")
	}

	// Create temp git repo to ensure clean state
	tmpDir, _ := os.MkdirTemp("", "test-git-*")
	defer os.RemoveAll(tmpDir)

	// Init git repo
	cmd1 := exec.Command("git", "init")
	cmd1.Dir = tmpDir
	cmd1.Run()

	cmd2 := exec.Command("git", "config", "user.name", "Test")
	cmd2.Dir = tmpDir
	cmd2.Run()

	cmd3 := exec.Command("git", "config", "user.email", "test@test.com")
	cmd3.Dir = tmpDir
	cmd3.Run()

	gh := NewGitHelper(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"action": "generate_commit_message",
	}

	result, _ := gh.Execute(ctx, params)

	if result.Success {
		t.Error("Result should fail when no files are staged")
	}

	if !strings.Contains(result.Error, "staged") {
		t.Error("Error should mention no staged files")
	}
}

func TestGitHelper_GetHistory(t *testing.T) {
	if !isGitRepoAvailable() {
		t.Skip("Not in a git repository, skipping test")
	}

	gh := NewGitHelper(".")
	ctx := context.Background()

	params := map[string]interface{}{
		"action": "history",
		"count":  5.0,
	}

	result, _ := gh.Execute(ctx, params)

	// Skip if git command fails (may have no commits or other issues)
	if !result.Success {
		if strings.Contains(result.Error, "Erro ao obter histórico") {
			t.Skip("Git repository may have no commits, skipping test")
		}
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	if result.Success && !strings.Contains(result.Message, "Histórico") {
		t.Error("Should show history header")
	}
}

func TestGitHelper_GetUncommitted(t *testing.T) {
	if !isGitRepoAvailable() {
		t.Skip("Not in a git repository, skipping test")
	}

	gh := NewGitHelper(".")
	ctx := context.Background()

	params := map[string]interface{}{
		"action": "uncommitted",
	}

	result, _ := gh.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	// Should show either changes or no changes
	hasMessage := strings.Contains(result.Message, "Alterações") ||
		strings.Contains(result.Message, "Nenhuma")

	if !hasMessage {
		t.Error("Should show uncommitted changes status")
	}
}

func TestGitHelper_GetBranchInfo(t *testing.T) {
	if !isGitRepoAvailable() {
		t.Skip("Not in a git repository, skipping test")
	}

	gh := NewGitHelper(".")
	ctx := context.Background()

	params := map[string]interface{}{
		"action": "branch_info",
	}

	result, _ := gh.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	if !strings.Contains(result.Message, "Branch atual") {
		t.Error("Should show current branch")
	}
}

func TestGitHelper_IsGitRepo(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() string
		expected bool
	}{
		{
			name: "Valid git repo",
			setup: func() string {
				tmpDir, _ := os.MkdirTemp("", "test-git-valid-*")
				cmd := exec.Command("git", "init")
				cmd.Dir = tmpDir
				cmd.Run()
				return tmpDir
			},
			expected: true,
		},
		{
			name: "Not a git repo",
			setup: func() string {
				tmpDir, _ := os.MkdirTemp("", "test-git-invalid-*")
				return tmpDir
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup()
			defer os.RemoveAll(dir)

			gh := NewGitHelper(dir)
			result := gh.isGitRepo()

			if result != tt.expected {
				t.Errorf("isGitRepo() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Helper function to check if git is available
func isGitRepoAvailable() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}
