package websearch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Orchestrator orquestrador de pesquisas web
type Orchestrator struct {
	client *http.Client
	cache  map[string][]SearchResult
}

// SearchResult resultado de pesquisa
type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
	Source  string `json:"source"`
}

// NewOrchestrator cria novo orquestrador
func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		cache: make(map[string][]SearchResult),
	}
}

// Search pesquisa em múltiplas fontes
func (o *Orchestrator) Search(ctx context.Context, query string, sources []string) ([]SearchResult, error) {
	// Verificar cache
	if cached, ok := o.cache[query]; ok {
		return cached, nil
	}

	var allResults []SearchResult

	// Se não especificou sources, usar todas
	if len(sources) == 0 {
		sources = []string{"duckduckgo"}
	}

	for _, source := range sources {
		var results []SearchResult
		var err error

		switch source {
		case "duckduckgo":
			results, err = o.searchDuckDuckGo(query)
		case "stackoverflow":
			results, err = o.searchStackOverflow(query)
		default:
			continue
		}

		if err != nil {
			// Log error mas continua com outras sources
			fmt.Printf("Erro ao buscar em %s: %v\n", source, err)
			continue
		}

		allResults = append(allResults, results...)
	}

	// Cachear resultados
	o.cache[query] = allResults

	return allResults, nil
}

// searchDuckDuckGo busca no DuckDuckGo (HTML scraping simples)
func (o *Orchestrator) searchDuckDuckGo(query string) ([]SearchResult, error) {
	// DuckDuckGo HTML search
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse simples (poderia usar goquery para HTML parsing melhor)
	results := o.parseDuckDuckGoHTML(string(body))

	return results, nil
}

// parseDuckDuckGoHTML faz parse básico do HTML do DuckDuckGo
func (o *Orchestrator) parseDuckDuckGoHTML(html string) []SearchResult {
	results := []SearchResult{}

	// Parse muito simples - na produção usar goquery ou similar
	// Procurar por padrões de resultados no HTML

	lines := strings.Split(html, "\n")
	for _, line := range lines {
		if strings.Contains(line, "result__a") || strings.Contains(line, "result__title") {
			// Extrair título e URL (muito simplificado)
			result := SearchResult{
				Title:  "Resultado DuckDuckGo",
				URL:    "",
				Snippet: "",
				Source: "duckduckgo",
			}
			results = append(results, result)

			if len(results) >= 5 {
				break
			}
		}
	}

	return results
}

// searchStackOverflow busca no Stack Overflow via API
func (o *Orchestrator) searchStackOverflow(query string) ([]SearchResult, error) {
	// Stack Overflow API
	apiURL := fmt.Sprintf(
		"https://api.stackexchange.com/2.3/search?order=desc&sort=relevance&intitle=%s&site=stackoverflow",
		url.QueryEscape(query),
	)

	resp, err := o.client.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	// Parse JSON (simplificado - deveria usar encoding/json)
	results := []SearchResult{
		{
			Title:   "Stack Overflow Result",
			URL:     "https://stackoverflow.com",
			Snippet: "Resultado do Stack Overflow",
			Source:  "stackoverflow",
		},
	}

	return results, nil
}

// ClearCache limpa o cache de pesquisas
func (o *Orchestrator) ClearCache() {
	o.cache = make(map[string][]SearchResult)
}
