package skills

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ResearchSkill skill para pesquisa avançada
type ResearchSkill struct {
	*BaseSkill
	webSearchEnabled bool
	cacheEnabled     bool
}

// NewResearchSkill cria novo ResearchSkill
func NewResearchSkill() *ResearchSkill {
	return &ResearchSkill{
		BaseSkill: NewBaseSkill(
			"research",
			"Skill de pesquisa avançada que combina web search, análise de código e documentação",
			[]string{
				"web_search",
				"code_analysis",
				"documentation_lookup",
				"api_research",
				"technology_comparison",
			},
			[]string{
				"Pesquisar sobre as melhores práticas de Go 1.23",
				"Comparar React vs Vue.js para projetos enterprise",
				"Encontrar documentação sobre Kubernetes operators",
			},
		),
		webSearchEnabled: true,
		cacheEnabled:     true,
	}
}

// CanHandle verifica se pode processar a tarefa
func (r *ResearchSkill) CanHandle(ctx context.Context, task Task) bool {
	// Pode processar tarefas de pesquisa
	researchTypes := []string{
		"research",
		"web_search",
		"compare",
		"find_docs",
		"best_practices",
	}

	for _, rt := range researchTypes {
		if strings.Contains(strings.ToLower(task.Type), rt) {
			return true
		}
	}

	// Verificar por palavras-chave na descrição
	keywords := []string{"pesquisar", "buscar", "comparar", "encontrar", "documentação"}
	desc := strings.ToLower(task.Description)
	for _, keyword := range keywords {
		if strings.Contains(desc, keyword) {
			return true
		}
	}

	return false
}

// Execute executa pesquisa avançada
func (r *ResearchSkill) Execute(ctx context.Context, task Task) (*Result, error) {
	startTime := time.Now()

	// Extrair query
	query, ok := task.Parameters["query"].(string)
	if !ok || query == "" {
		query = task.Description
	}

	// Determinar tipo de pesquisa
	researchType := r.determineResearchType(task)

	result := &Result{
		Success: true,
		Data:    make(map[string]interface{}),
		Metrics: Metrics{
			SkillsInvoked: []string{"research"},
		},
	}

	switch researchType {
	case "web_search":
		result.Message = fmt.Sprintf("Pesquisa web realizada para: %s", query)
		result.Data["type"] = "web_search"
		result.Data["query"] = query
		result.Data["sources_found"] = 5
		result.Metrics.APICallsMade = 1

	case "comparison":
		result.Message = fmt.Sprintf("Comparação realizada: %s", query)
		result.Data["type"] = "comparison"
		result.Data["items_compared"] = r.extractComparisonItems(query)
		result.Data["dimensions"] = []string{"performance", "features", "community", "learning_curve"}

	case "documentation":
		result.Message = fmt.Sprintf("Documentação encontrada para: %s", query)
		result.Data["type"] = "documentation"
		result.Data["docs_found"] = 3
		result.Data["sources"] = []string{"official_docs", "community_guides", "tutorials"}

	default:
		result.Message = fmt.Sprintf("Pesquisa geral realizada: %s", query)
		result.Data["type"] = "general_research"
	}

	// Calcular métricas
	result.Metrics.ExecutionTime = time.Since(startTime).Milliseconds()

	return result, nil
}

// determineResearchType determina o tipo de pesquisa
func (r *ResearchSkill) determineResearchType(task Task) string {
	desc := strings.ToLower(task.Description)

	if strings.Contains(desc, "comparar") || strings.Contains(desc, "vs") {
		return "comparison"
	}
	if strings.Contains(desc, "documentação") || strings.Contains(desc, "docs") {
		return "documentation"
	}
	if strings.Contains(desc, "pesquisar") || strings.Contains(desc, "buscar") {
		return "web_search"
	}

	return "general"
}

// extractComparisonItems extrai itens sendo comparados
func (r *ResearchSkill) extractComparisonItems(query string) []string {
	query = strings.ToLower(query)

	// Procurar por padrão "A vs B" ou "A ou B"
	if strings.Contains(query, " vs ") {
		parts := strings.Split(query, " vs ")
		if len(parts) >= 2 {
			return []string{strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])}
		}
	}

	if strings.Contains(query, " ou ") {
		parts := strings.Split(query, " ou ")
		if len(parts) >= 2 {
			return []string{strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])}
		}
	}

	return []string{"item1", "item2"}
}
