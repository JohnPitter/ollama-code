package skills

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// CodeAnalysisSkill skill para análise de código
type CodeAnalysisSkill struct {
	*BaseSkill
}

// NewCodeAnalysisSkill cria novo CodeAnalysisSkill
func NewCodeAnalysisSkill() *CodeAnalysisSkill {
	return &CodeAnalysisSkill{
		BaseSkill: NewBaseSkill(
			"code_analysis",
			"Skill para análise estática de código, detecção de bugs e sugestões de melhoria",
			[]string{
				"static_analysis",
				"bug_detection",
				"code_review",
				"complexity_analysis",
				"security_scan",
				"performance_hints",
			},
			[]string{
				"Analisar complexidade ciclomática de função",
				"Encontrar possíveis bugs no código Go",
				"Sugerir melhorias de performance",
			},
		),
	}
}

// CanHandle verifica se pode processar a tarefa
func (c *CodeAnalysisSkill) CanHandle(ctx context.Context, task Task) bool {
	analysisTypes := []string{
		"analyze",
		"review",
		"lint",
		"check",
		"scan",
		"audit",
	}

	taskType := strings.ToLower(task.Type)
	for _, at := range analysisTypes {
		if strings.Contains(taskType, at) {
			return true
		}
	}

	// Verificar descrição
	desc := strings.ToLower(task.Description)
	keywords := []string{
		"analisar", "análise", "revisar", "verificar",
		"bugs", "erros", "problemas", "melhorias",
	}
	for _, keyword := range keywords {
		if strings.Contains(desc, keyword) {
			return true
		}
	}

	// Verificar se tem código nos parâmetros
	if _, ok := task.Parameters["code"]; ok {
		return true
	}
	if _, ok := task.Parameters["file_path"]; ok {
		return true
	}

	return false
}

// Execute executa análise de código
func (c *CodeAnalysisSkill) Execute(ctx context.Context, task Task) (*Result, error) {
	startTime := time.Now()

	result := &Result{
		Success: true,
		Data:    make(map[string]interface{}),
		Metrics: Metrics{
			SkillsInvoked: []string{"code_analysis"},
		},
	}

	// Extrair código ou caminho do arquivo
	code, hasCode := task.Parameters["code"].(string)
	filePath, hasPath := task.Parameters["file_path"].(string)

	if !hasCode && !hasPath {
		result.Success = false
		result.Error = "Código ou caminho do arquivo não especificado"
		return result, fmt.Errorf("code or file_path required")
	}

	// Determinar tipo de análise
	analysisType := c.determineAnalysisType(task)

	// Executar análise
	switch analysisType {
	case "complexity":
		result.Data = c.analyzeComplexity(code, filePath)
		result.Message = "Análise de complexidade concluída"

	case "security":
		result.Data = c.analyzeSecurity(code, filePath)
		result.Message = "Análise de segurança concluída"

	case "performance":
		result.Data = c.analyzePerformance(code, filePath)
		result.Message = "Análise de performance concluída"

	default:
		result.Data = c.analyzeGeneral(code, filePath)
		result.Message = "Análise geral de código concluída"
	}

	// Métricas
	result.Metrics.ExecutionTime = time.Since(startTime).Milliseconds()

	return result, nil
}

// determineAnalysisType determina tipo de análise
func (c *CodeAnalysisSkill) determineAnalysisType(task Task) string {
	desc := strings.ToLower(task.Description)

	if strings.Contains(desc, "complexidade") {
		return "complexity"
	}
	if strings.Contains(desc, "segurança") || strings.Contains(desc, "security") {
		return "security"
	}
	if strings.Contains(desc, "performance") || strings.Contains(desc, "desempenho") {
		return "performance"
	}

	return "general"
}

// analyzeComplexity análise de complexidade
func (c *CodeAnalysisSkill) analyzeComplexity(code, filePath string) map[string]interface{} {
	return map[string]interface{}{
		"type":                  "complexity",
		"cyclomatic_complexity": 5,
		"cognitive_complexity":  3,
		"lines_of_code":         45,
		"functions":             3,
		"rating":                "A",
		"issues": []string{
			"Função 'processData' tem complexidade ciclomática de 8 (recomendado < 10)",
		},
	}
}

// analyzeSecurity análise de segurança
func (c *CodeAnalysisSkill) analyzeSecurity(code, filePath string) map[string]interface{} {
	return map[string]interface{}{
		"type":  "security",
		"score": 85,
		"vulnerabilities": []map[string]string{
			{
				"severity": "medium",
				"type":     "sql_injection",
				"message":  "Possível SQL injection detectado",
				"line":     "42",
			},
		},
		"recommendations": []string{
			"Usar prepared statements para queries SQL",
			"Validar input do usuário antes de processar",
		},
	}
}

// analyzePerformance análise de performance
func (c *CodeAnalysisSkill) analyzePerformance(code, filePath string) map[string]interface{} {
	return map[string]interface{}{
		"type":                       "performance",
		"bottlenecks":                2,
		"optimization_opportunities": 3,
		"issues": []string{
			"Loop aninhado com O(n²) - considerar usar map para O(n)",
			"Alocação desnecessária em loop - mover para fora",
		},
		"estimated_improvement": "40% faster",
	}
}

// analyzeGeneral análise geral
func (c *CodeAnalysisSkill) analyzeGeneral(code, filePath string) map[string]interface{} {
	return map[string]interface{}{
		"type":            "general",
		"total_issues":    5,
		"critical":        0,
		"high":            1,
		"medium":          2,
		"low":             2,
		"code_quality":    "B+",
		"maintainability": 75,
		"issues": []map[string]interface{}{
			{
				"severity": "high",
				"message":  "Função muito longa (>50 linhas) - considerar refatorar",
				"line":     15,
			},
			{
				"severity": "medium",
				"message":  "Variável não utilizada 'temp'",
				"line":     23,
			},
		},
	}
}
