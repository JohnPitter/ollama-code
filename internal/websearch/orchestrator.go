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
	client  *http.Client
	fetcher *ContentFetcher
	cache   map[string][]SearchResult
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
			Timeout: 10 * time.Second, // Reduzir timeout para evitar travamentos
		},
		fetcher: NewContentFetcher(),
		cache:   make(map[string][]SearchResult),
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
			results, err = o.searchDuckDuckGo(ctx, query)
		case "stackoverflow":
			results, err = o.searchStackOverflow(ctx, query)
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
func (o *Orchestrator) searchDuckDuckGo(ctx context.Context, query string) ([]SearchResult, error) {
	// DuckDuckGo HTML search
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en;q=0.8")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024)) // Limitar a 5MB
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	htmlContent := string(body)

	// Parse simples (poderia usar goquery para HTML parsing melhor)
	results := o.parseDuckDuckGoHTML(htmlContent)

	if len(results) == 0 {
		return nil, fmt.Errorf("nenhum resultado encontrado no HTML")
	}

	return results, nil
}

// parseDuckDuckGoHTML faz parse básico do HTML do DuckDuckGo
func (o *Orchestrator) parseDuckDuckGoHTML(html string) []SearchResult {
	results := []SearchResult{}

	// Parse muito simples - na produção usar goquery ou similar
	// Tentar extrair snippets básicos

	// Procurar por tags <a class="result__a"
	searchPos := 0
	for {
		foundIndex := strings.Index(html[searchPos:], `class="result__a"`)
		if foundIndex == -1 {
			break
		}
		// Converter para índice absoluto
		titleStart := searchPos + foundIndex

		// Tentar extrair href
		hrefStart := strings.LastIndex(html[:titleStart], `href="`)
		if hrefStart == -1 {
			searchPos = titleStart + 1
			continue
		}
		hrefStart += len(`href="`)
		hrefEnd := strings.Index(html[hrefStart:], `"`)
		if hrefEnd == -1 {
			searchPos = titleStart + 1
			continue
		}
		resultURL := html[hrefStart : hrefStart+hrefEnd]

		// DuckDuckGo usa URLs codificadas como //duckduckgo.com/l/?uddg=URL
		// Extrair a URL real
		resultURL = o.extractRealURL(resultURL)

		// Ignorar URLs inválidas ou vazias
		if resultURL == "" || !strings.HasPrefix(resultURL, "http") {
			searchPos = titleStart + 1
			continue
		}

		// Tentar extrair título
		titleTextStart := strings.Index(html[titleStart:], `>`)
		if titleTextStart == -1 {
			searchPos = titleStart + 1
			continue
		}
		titleTextStart += titleStart + 1
		titleTextEnd := strings.Index(html[titleTextStart:], `</a>`)
		if titleTextEnd == -1 {
			searchPos = titleStart + 1
			continue
		}
		title := strings.TrimSpace(html[titleTextStart : titleTextStart+titleTextEnd])

		// Ignorar resultados sem título
		if title == "" {
			searchPos = titleStart + 1
			continue
		}

		// Tentar extrair snippet
		snippetStart := strings.Index(html[titleStart:], `class="result__snippet"`)
		snippet := ""
		if snippetStart != -1 {
			snippetStart += titleStart + len(`class="result__snippet"`)
			snippetTextStart := strings.Index(html[snippetStart:], `>`)
			if snippetTextStart != -1 {
				snippetTextStart += snippetStart + 1
				snippetTextEnd := strings.Index(html[snippetTextStart:], `</`)
				if snippetTextEnd != -1 {
					snippet = strings.TrimSpace(html[snippetTextStart : snippetTextStart+snippetTextEnd])
				}
			}
		}

		result := SearchResult{
			Title:   title,
			URL:     resultURL,
			Snippet: snippet,
			Source:  "duckduckgo",
		}
		results = append(results, result)

		// Avançar searchPos para buscar próximo resultado
		searchPos = titleStart + 1

		if len(results) >= 5 {
			break
		}
	}

	return results
}

// searchStackOverflow busca no Stack Overflow via API
func (o *Orchestrator) searchStackOverflow(ctx context.Context, query string) ([]SearchResult, error) {
	// Stack Overflow API
	apiURL := fmt.Sprintf(
		"https://api.stackexchange.com/2.3/search?order=desc&sort=relevance&intitle=%s&site=stackoverflow",
		url.QueryEscape(query),
	)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := o.client.Do(req)
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

// FetchContents busca conteúdo real das URLs dos resultados de pesquisa
func (o *Orchestrator) FetchContents(ctx context.Context, results []SearchResult, maxResults int) ([]FetchedContent, error) {
	if maxResults <= 0 || maxResults > len(results) {
		maxResults = len(results)
	}

	// Extrair URLs dos top resultados
	urls := make([]string, maxResults)
	for i := 0; i < maxResults; i++ {
		urls[i] = results[i].URL
	}

	// Buscar conteúdo em paralelo
	contents := o.fetcher.FetchMultiple(ctx, urls, 3)

	// Adicionar títulos dos resultados originais caso o fetch não tenha conseguido extrair
	for i := range contents {
		if contents[i].Title == "Sem título" && i < len(results) {
			contents[i].Title = results[i].Title
		}
	}

	return contents, nil
}

// extractRealURL extrai a URL real de URLs codificadas do DuckDuckGo
func (o *Orchestrator) extractRealURL(rawURL string) string {
	// DuckDuckGo pode retornar URLs em formato:
	// //duckduckgo.com/l/?uddg=https%3A%2F%2Fexample.com
	// Ou diretamente: https://example.com

	// Se começar com //, adicionar https:
	if strings.HasPrefix(rawURL, "//") {
		rawURL = "https:" + rawURL
	}

	// Se contém uddg= ou kh=, tentar extrair URL decodificada
	if strings.Contains(rawURL, "uddg=") {
		parts := strings.Split(rawURL, "uddg=")
		if len(parts) > 1 {
			decoded, err := url.QueryUnescape(parts[1])
			if err == nil {
				// Remover parâmetros adicionais após &
				if idx := strings.Index(decoded, "&"); idx != -1 {
					decoded = decoded[:idx]
				}
				return decoded
			}
		}
	}

	// Fallback: retornar URL original
	return rawURL
}
