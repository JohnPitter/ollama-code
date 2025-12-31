package validators

import (
	"encoding/json"
	"regexp"
	"strings"
)

// JSONValidator valida e extrai JSON
type JSONValidator struct {
	jsonRegex      *regexp.Regexp
	jsonBlockRegex *regexp.Regexp
}

// NewJSONValidator cria novo validador
func NewJSONValidator() *JSONValidator {
	return &JSONValidator{
		jsonRegex:      regexp.MustCompile(`\{[\s\S]*?\}`),
		jsonBlockRegex: regexp.MustCompile("```json\\s*([\\s\\S]*?)```"),
	}
}

// Extract extrai JSON de uma string
func (v *JSONValidator) Extract(content string) string {
	// Primeiro, tentar extrair de bloco markdown
	if matches := v.jsonBlockRegex.FindStringSubmatch(content); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Segundo, procurar por JSON puro
	match := v.jsonRegex.FindString(content)
	if match == "" {
		return ""
	}

	return match
}

// Parse faz parse de JSON com fallback para extrair
func (v *JSONValidator) Parse(content string) (map[string]interface{}, error) {
	var result map[string]interface{}

	// Tentar parse direto
	err := json.Unmarshal([]byte(content), &result)
	if err == nil {
		return result, nil
	}

	// Tentar extrair primeiro
	extracted := v.Extract(content)
	if extracted == "" {
		return nil, err // Retorna erro original
	}

	// Tentar parse do extraído
	if err := json.Unmarshal([]byte(extracted), &result); err != nil {
		return nil, err
	}

	return result, nil
}

// IsValid verifica se string é JSON válido
func (v *JSONValidator) IsValid(content string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(content), &js) == nil
}

// Prettify formata JSON de forma legível
func (v *JSONValidator) Prettify(content string) (string, error) {
	var obj interface{}

	if err := json.Unmarshal([]byte(content), &obj); err != nil {
		return "", err
	}

	pretty, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", err
	}

	return string(pretty), nil
}

// ParseArray faz parse de JSON array
func (v *JSONValidator) ParseArray(content string) ([]interface{}, error) {
	var result []interface{}

	// Tentar parse direto
	err := json.Unmarshal([]byte(content), &result)
	if err == nil {
		return result, nil
	}

	// Tentar extrair primeiro
	extracted := v.Extract(content)
	if extracted == "" {
		return nil, err
	}

	if err := json.Unmarshal([]byte(extracted), &result); err != nil {
		return nil, err
	}

	return result, nil
}
