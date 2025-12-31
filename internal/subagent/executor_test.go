package subagent

import (
	"strings"
	"testing"
)

// TestNewExecutor testa criação do executor
func TestNewExecutor(t *testing.T) {
	ollamaURL := "http://localhost:11434"
	executor := NewExecutor(ollamaURL)

	if executor == nil {
		t.Fatal("NewExecutor returned nil")
	}

	if executor.ollamaURL != ollamaURL {
		t.Errorf("Expected ollamaURL=%s, got %s", ollamaURL, executor.ollamaURL)
	}
}

// TestBuildPrompt_Explore testa prompt para agent Explore
func TestBuildPrompt_Explore(t *testing.T) {
	executor := NewExecutor("http://localhost:11434")

	agent := &Subagent{
		Type:    AgentTypeExplore,
		Prompt:  "Find all authentication functions",
		WorkDir: "/project/src",
	}

	prompt := executor.buildPrompt(agent)

	// Verificar que contém instruções de Explore
	if !strings.Contains(prompt, "Explore agent") {
		t.Error("Prompt should identify agent as Explore")
	}

	if !strings.Contains(prompt, "code exploration and search") {
		t.Error("Prompt should mention code exploration")
	}

	// Verificar que contém o task prompt do usuário
	if !strings.Contains(prompt, "Find all authentication functions") {
		t.Error("Prompt should contain user's task")
	}

	// Verificar que contém working directory
	if !strings.Contains(prompt, "/project/src") {
		t.Error("Prompt should contain working directory")
	}
}

// TestBuildPrompt_Plan testa prompt para agent Plan
func TestBuildPrompt_Plan(t *testing.T) {
	executor := NewExecutor("http://localhost:11434")

	agent := &Subagent{
		Type:   AgentTypePlan,
		Prompt: "Plan refactoring of auth module",
	}

	prompt := executor.buildPrompt(agent)

	if !strings.Contains(prompt, "Plan agent") {
		t.Error("Prompt should identify agent as Plan")
	}

	if !strings.Contains(prompt, "planning and architectural analysis") {
		t.Error("Prompt should mention planning")
	}

	if !strings.Contains(prompt, "Break down complex tasks") {
		t.Error("Prompt should mention breaking down tasks")
	}

	if !strings.Contains(prompt, "Plan refactoring of auth module") {
		t.Error("Prompt should contain user's task")
	}
}

// TestBuildPrompt_Execute testa prompt para agent Execute
func TestBuildPrompt_Execute(t *testing.T) {
	executor := NewExecutor("http://localhost:11434")

	agent := &Subagent{
		Type:   AgentTypeExecute,
		Prompt: "Implement unit tests for UserService",
	}

	prompt := executor.buildPrompt(agent)

	if !strings.Contains(prompt, "Execute agent") {
		t.Error("Prompt should identify agent as Execute")
	}

	if !strings.Contains(prompt, "task execution") {
		t.Error("Prompt should mention task execution")
	}

	if !strings.Contains(prompt, "Execute tasks efficiently") {
		t.Error("Prompt should mention efficiency")
	}

	if !strings.Contains(prompt, "Implement unit tests for UserService") {
		t.Error("Prompt should contain user's task")
	}
}

// TestBuildPrompt_General testa prompt para agent General
func TestBuildPrompt_General(t *testing.T) {
	executor := NewExecutor("http://localhost:11434")

	agent := &Subagent{
		Type:   AgentTypeGeneral,
		Prompt: "Help me understand this code",
	}

	prompt := executor.buildPrompt(agent)

	if !strings.Contains(prompt, "general-purpose coding assistant") {
		t.Error("Prompt should identify agent as general-purpose")
	}

	if !strings.Contains(prompt, "Help me understand this code") {
		t.Error("Prompt should contain user's task")
	}
}

// TestBuildPrompt_NoWorkDir testa prompt sem working directory
func TestBuildPrompt_NoWorkDir(t *testing.T) {
	executor := NewExecutor("http://localhost:11434")

	agent := &Subagent{
		Type:    AgentTypeExplore,
		Prompt:  "test",
		WorkDir: "",
	}

	prompt := executor.buildPrompt(agent)

	// Não deve conter "Working directory:" quando WorkDir está vazio
	if strings.Contains(prompt, "Working directory:") {
		t.Error("Prompt should not contain working directory when empty")
	}

	// Teste com WorkDir = "."
	agent.WorkDir = "."
	prompt = executor.buildPrompt(agent)

	if strings.Contains(prompt, "Working directory:") {
		t.Error("Prompt should not contain working directory when it's '.'")
	}
}

// TestPostProcess_Explore testa pós-processamento para Explore
func TestPostProcess_Explore(t *testing.T) {
	executor := NewExecutor("http://localhost:11434")

	agent := &Subagent{Type: AgentTypeExplore}

	// Response sem estrutura (mais de 3 linhas para trigger header addition)
	response := "Found 3 authentication functions\nThey use JWT tokens\nLocated in auth package\nAll properly secured"

	result := executor.postProcess(agent, response)

	// Deve adicionar header "## Exploration Results"
	if !strings.Contains(result, "## Exploration Results") {
		t.Error("PostProcess should add '## Exploration Results' header for Explore agents")
	}

	// Deve conter o conteúdo original
	if !strings.Contains(result, "Found 3 authentication functions") {
		t.Error("PostProcess should preserve original content")
	}
}

// TestPostProcess_Explore_AlreadyStructured testa Explore já estruturado
func TestPostProcess_Explore_AlreadyStructured(t *testing.T) {
	executor := NewExecutor("http://localhost:11434")

	agent := &Subagent{Type: AgentTypeExplore}

	// Response já com estrutura markdown
	response := "## Found Files\n\n- auth.go\n- user.go"

	result := executor.postProcess(agent, response)

	// Não deve adicionar header duplicado
	headerCount := strings.Count(result, "##")
	if headerCount > 1 {
		t.Error("PostProcess should not add duplicate headers when response already has structure")
	}
}

// TestPostProcess_Plan testa pós-processamento para Plan
func TestPostProcess_Plan(t *testing.T) {
	executor := NewExecutor("http://localhost:11434")

	agent := &Subagent{Type: AgentTypePlan}

	// Response sem numeração
	response := "Create new auth module\nMigrate existing code\nWrite tests\nDeploy changes"

	result := executor.postProcess(agent, response)

	// Deve adicionar header "## Plan"
	if !strings.Contains(result, "## Plan") {
		t.Error("PostProcess should add '## Plan' header for Plan agents")
	}

	// Deve adicionar numeração
	if !strings.Contains(result, "1.") {
		t.Error("PostProcess should add step numbering")
	}

	// Deve ter múltiplos steps
	stepCount := strings.Count(result, ".")
	if stepCount < 4 { // Pelo menos 4 steps (1. 2. 3. 4.)
		t.Errorf("PostProcess should number all steps, found %d periods", stepCount)
	}
}

// TestPostProcess_Plan_AlreadyNumbered testa Plan já numerado
func TestPostProcess_Plan_AlreadyNumbered(t *testing.T) {
	executor := NewExecutor("http://localhost:11434")

	agent := &Subagent{Type: AgentTypePlan}

	// Response já numerado
	response := "1. Create module\n2. Write tests\n3. Deploy"

	result := executor.postProcess(agent, response)

	// Não deve re-numerar
	// Verifica que não duplicou numeração
	if strings.Contains(result, "1. 1.") {
		t.Error("PostProcess should not re-number already numbered steps")
	}
}

// TestPostProcess_Execute testa pós-processamento para Execute
func TestPostProcess_Execute(t *testing.T) {
	executor := NewExecutor("http://localhost:11434")

	agent := &Subagent{Type: AgentTypeExecute}

	response := "  Task completed successfully  \n\n"

	result := executor.postProcess(agent, response)

	// Deve remover espaços extras
	if result != "Task completed successfully" {
		t.Errorf("Expected trimmed result, got '%s'", result)
	}
}

// TestPostProcess_General testa pós-processamento para General
func TestPostProcess_General(t *testing.T) {
	executor := NewExecutor("http://localhost:11434")

	agent := &Subagent{Type: AgentTypeGeneral}

	response := "  General response  "

	result := executor.postProcess(agent, response)

	// Deve apenas fazer trim
	if result != "General response" {
		t.Errorf("Expected trimmed result, got '%s'", result)
	}

	// Não deve adicionar estrutura extra
	if strings.Contains(result, "##") {
		t.Error("General agent should not add special structure")
	}
}

// TestPostProcess_WhitespaceHandling testa manipulação de espaços em branco
func TestPostProcess_WhitespaceHandling(t *testing.T) {
	executor := NewExecutor("http://localhost:11434")

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "leading whitespace",
			input:    "   content",
			expected: "content",
		},
		{
			name:     "trailing whitespace",
			input:    "content   ",
			expected: "content",
		},
		{
			name:     "multiple newlines",
			input:    "\n\ncontent\n\n",
			expected: "content",
		},
		{
			name:     "tabs and spaces",
			input:    "\t  content  \t",
			expected: "content",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			agent := &Subagent{Type: AgentTypeGeneral}
			result := executor.postProcess(agent, tc.input)

			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// TestCreateExecutorFunc testa criação de ExecutorFunc
func TestCreateExecutorFunc(t *testing.T) {
	executor := NewExecutor("http://localhost:11434")

	executorFunc := executor.CreateExecutorFunc()

	if executorFunc == nil {
		t.Fatal("CreateExecutorFunc returned nil")
	}

	// Verificar que é uma função válida do tipo ExecutorFunc
	// Não executamos porque requer LLM real, mas verificamos que o tipo está correto
	var _ ExecutorFunc = executorFunc
}

// TestBuildPrompt_AllTypes testa todos os tipos de agent
func TestBuildPrompt_AllTypes(t *testing.T) {
	executor := NewExecutor("http://localhost:11434")

	types := []AgentType{
		AgentTypeExplore,
		AgentTypePlan,
		AgentTypeExecute,
		AgentTypeGeneral,
	}

	for _, agentType := range types {
		t.Run(string(agentType), func(t *testing.T) {
			agent := &Subagent{
				Type:   agentType,
				Prompt: "test prompt",
			}

			prompt := executor.buildPrompt(agent)

			// Verificar que prompt não está vazio
			if prompt == "" {
				t.Error("Prompt should not be empty")
			}

			// Verificar que contém task
			if !strings.Contains(prompt, "test prompt") {
				t.Error("Prompt should contain task")
			}

			// Verificar que contém "Task:"
			if !strings.Contains(prompt, "Task:") {
				t.Error("Prompt should contain 'Task:' section")
			}
		})
	}
}

// TestPostProcess_Plan_ComplexScenario testa cenário complexo de numeração
func TestPostProcess_Plan_ComplexScenario(t *testing.T) {
	executor := NewExecutor("http://localhost:11434")

	agent := &Subagent{Type: AgentTypePlan}

	// Response com linhas vazias e headers
	response := `Planning the refactoring

Create new structure
Migrate old code

Test everything
Deploy`

	result := executor.postProcess(agent, response)

	// Deve ter header
	if !strings.Contains(result, "## Plan") {
		t.Error("Should add Plan header")
	}

	// Deve numerar apenas linhas não vazias e não headers
	lines := strings.Split(result, "\n")
	numberedCount := 0
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "1.") ||
			strings.HasPrefix(strings.TrimSpace(line), "2.") ||
			strings.HasPrefix(strings.TrimSpace(line), "3.") ||
			strings.HasPrefix(strings.TrimSpace(line), "4.") {
			numberedCount++
		}
	}

	if numberedCount < 4 {
		t.Errorf("Expected at least 4 numbered steps, got %d", numberedCount)
	}
}

// TestPostProcess_Explore_ShortResponse testa response curto
func TestPostProcess_Explore_ShortResponse(t *testing.T) {
	executor := NewExecutor("http://localhost:11434")

	agent := &Subagent{Type: AgentTypeExplore}

	// Response muito curto (poucas linhas)
	response := "Found file\nDone"

	result := executor.postProcess(agent, response)

	// Não deve adicionar header para responses muito curtos (< 3 linhas)
	// Baseado na lógica: if len(lines) > 3
	if strings.Contains(result, "## Exploration Results") {
		t.Error("Should not add header for very short responses")
	}

	// Deve apenas fazer trim
	expected := "Found file\nDone"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}
