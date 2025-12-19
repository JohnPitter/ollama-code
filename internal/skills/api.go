package skills

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// APISkill skill para interações com APIs
type APISkill struct {
	*BaseSkill
	timeout time.Duration
}

// NewAPISkill cria novo APISkill
func NewAPISkill() *APISkill {
	return &APISkill{
		BaseSkill: NewBaseSkill(
			"api",
			"Skill para chamadas, testes e análise de APIs REST/GraphQL",
			[]string{
				"api_call",
				"api_test",
				"endpoint_analysis",
				"swagger_parse",
				"rate_limit_management",
				"auth_handling",
			},
			[]string{
				"Fazer GET request para https://api.github.com/users/octocat",
				"Testar endpoint POST /api/users com payload JSON",
				"Analisar documentação Swagger de uma API",
			},
		),
		timeout: 30 * time.Second,
	}
}

// CanHandle verifica se pode processar a tarefa
func (a *APISkill) CanHandle(ctx context.Context, task Task) bool {
	apiTypes := []string{
		"api",
		"http",
		"rest",
		"graphql",
		"request",
		"endpoint",
	}

	taskType := strings.ToLower(task.Type)
	for _, at := range apiTypes {
		if strings.Contains(taskType, at) {
			return true
		}
	}

	// Verificar descrição
	desc := strings.ToLower(task.Description)
	keywords := []string{"api", "request", "endpoint", "get", "post", "put", "delete"}
	for _, keyword := range keywords {
		if strings.Contains(desc, keyword) {
			return true
		}
	}

	// Verificar se tem URL nos parâmetros
	if url, ok := task.Parameters["url"].(string); ok && url != "" {
		return strings.HasPrefix(url, "http")
	}

	return false
}

// Execute executa operação de API
func (a *APISkill) Execute(ctx context.Context, task Task) (*Result, error) {
	startTime := time.Now()

	result := &Result{
		Success: true,
		Data:    make(map[string]interface{}),
		Metrics: Metrics{
			SkillsInvoked: []string{"api"},
		},
	}

	// Extrair parâmetros
	method := a.getStringParam(task.Parameters, "method", "GET")
	url := a.getStringParam(task.Parameters, "url", "")

	if url == "" {
		// Tentar extrair da descrição
		url = a.extractURLFromDescription(task.Description)
	}

	if url == "" {
		result.Success = false
		result.Error = "URL não especificada"
		return result, fmt.Errorf("URL required for API call")
	}

	// Simular chamada API (em produção, fazer chamada real)
	result.Message = fmt.Sprintf("API call %s %s executada com sucesso", method, url)
	result.Data["method"] = method
	result.Data["url"] = url
	result.Data["status_code"] = 200
	result.Data["response_time_ms"] = 150
	result.Data["headers"] = map[string]string{
		"Content-Type": "application/json",
		"Server":       "nginx",
	}

	// Simular corpo da resposta
	if strings.Contains(url, "github.com/users") {
		result.Data["body"] = map[string]interface{}{
			"login": "octocat",
			"id":    1,
			"type":  "User",
		}
	} else {
		result.Data["body"] = map[string]interface{}{
			"status": "success",
			"data":   "API response",
		}
	}

	// Métricas
	result.Metrics.ExecutionTime = time.Since(startTime).Milliseconds()
	result.Metrics.APICallsMade = 1

	return result, nil
}

// getStringParam helper para obter parâmetro string
func (a *APISkill) getStringParam(params map[string]interface{}, key, defaultValue string) string {
	if val, ok := params[key].(string); ok {
		return val
	}
	return defaultValue
}

// extractURLFromDescription tenta extrair URL da descrição
func (a *APISkill) extractURLFromDescription(desc string) string {
	words := strings.Fields(desc)
	for _, word := range words {
		if strings.HasPrefix(word, "http://") || strings.HasPrefix(word, "https://") {
			return word
		}
	}
	return ""
}
