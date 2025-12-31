package websearch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// ContentFetcher busca e extrai conteúdo de URLs
type ContentFetcher struct {
	client *http.Client
}

// FetchedContent conteúdo extraído de uma URL
type FetchedContent struct {
	URL     string
	Title   string
	Content string
	Error   string
}

// NewContentFetcher cria novo fetcher
func NewContentFetcher() *ContentFetcher {
	return &ContentFetcher{
		client: &http.Client{
			Timeout: 15 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 5 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
	}
}

// FetchContent busca e extrai conteúdo de uma URL
func (f *ContentFetcher) FetchContent(ctx context.Context, url string) FetchedContent {
	result := FetchedContent{URL: url}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		result.Error = fmt.Sprintf("create request: %v", err)
		return result
	}

	// Headers para parecer um browser real
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7")

	resp, err := f.client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("fetch failed: %v", err)
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		result.Error = fmt.Sprintf("status %d", resp.StatusCode)
		return result
	}

	// Ler corpo (limitar a 1MB para evitar problemas)
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		result.Error = fmt.Sprintf("read body: %v", err)
		return result
	}

	html := string(body)

	// Extrair título
	result.Title = f.extractTitle(html)

	// Extrair conteúdo principal
	result.Content = f.extractMainContent(html)

	return result
}

// extractTitle extrai o título da página
func (f *ContentFetcher) extractTitle(html string) string {
	// Procurar por <title>...</title>
	titleRegex := regexp.MustCompile(`<title[^>]*>(.*?)</title>`)
	matches := titleRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		title := matches[1]
		// Decodificar HTML entities básicas
		title = strings.ReplaceAll(title, "&amp;", "&")
		title = strings.ReplaceAll(title, "&lt;", "<")
		title = strings.ReplaceAll(title, "&gt;", ">")
		title = strings.ReplaceAll(title, "&quot;", "\"")
		title = strings.ReplaceAll(title, "&#39;", "'")
		return strings.TrimSpace(title)
	}
	return "Sem título"
}

// extractMainContent extrai o conteúdo principal removendo tags HTML e ruído
func (f *ContentFetcher) extractMainContent(html string) string {
	// Remover scripts e styles
	html = f.removeTag(html, "script")
	html = f.removeTag(html, "style")
	html = f.removeTag(html, "nav")
	html = f.removeTag(html, "header")
	html = f.removeTag(html, "footer")
	html = f.removeTag(html, "aside")

	// Tentar encontrar conteúdo principal
	// Procurar por tags comuns de conteúdo: article, main, div.content, etc
	mainContent := ""

	// Tentar <article>
	articleRegex := regexp.MustCompile(`(?s)<article[^>]*>(.*?)</article>`)
	if matches := articleRegex.FindStringSubmatch(html); len(matches) > 1 {
		mainContent = matches[1]
	}

	// Se não achou, tentar <main>
	if mainContent == "" {
		mainRegex := regexp.MustCompile(`(?s)<main[^>]*>(.*?)</main>`)
		if matches := mainRegex.FindStringSubmatch(html); len(matches) > 1 {
			mainContent = matches[1]
		}
	}

	// Se ainda não achou, pegar body inteiro
	if mainContent == "" {
		bodyRegex := regexp.MustCompile(`(?s)<body[^>]*>(.*?)</body>`)
		if matches := bodyRegex.FindStringSubmatch(html); len(matches) > 1 {
			mainContent = matches[1]
		} else {
			mainContent = html
		}
	}

	// Remover todas as tags HTML
	text := f.stripHTML(mainContent)

	// Limpar espaços em branco excessivos
	text = f.cleanWhitespace(text)

	// Limitar tamanho (primeiros 3000 caracteres são geralmente suficientes)
	if len(text) > 3000 {
		text = text[:3000] + "..."
	}

	return text
}

// removeTag remove todas as ocorrências de uma tag específica
func (f *ContentFetcher) removeTag(html, tag string) string {
	regex := regexp.MustCompile(fmt.Sprintf(`(?s)<%s[^>]*>.*?</%s>`, tag, tag))
	return regex.ReplaceAllString(html, "")
}

// stripHTML remove todas as tags HTML
func (f *ContentFetcher) stripHTML(html string) string {
	// Remover tags
	regex := regexp.MustCompile(`<[^>]+>`)
	text := regex.ReplaceAllString(html, " ")

	// Decodificar entidades HTML comuns
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&mdash;", "—")
	text = strings.ReplaceAll(text, "&ndash;", "–")

	return text
}

// cleanWhitespace limpa espaços em branco excessivos
func (f *ContentFetcher) cleanWhitespace(text string) string {
	// Substituir múltiplos espaços por um único
	spaceRegex := regexp.MustCompile(`\s+`)
	text = spaceRegex.ReplaceAllString(text, " ")

	// Substituir múltiplas quebras de linha por no máximo duas
	lineRegex := regexp.MustCompile(`\n{3,}`)
	text = lineRegex.ReplaceAllString(text, "\n\n")

	return strings.TrimSpace(text)
}

// FetchMultiple busca conteúdo de múltiplas URLs em paralelo
func (f *ContentFetcher) FetchMultiple(ctx context.Context, urls []string, maxConcurrent int) []FetchedContent {
	if maxConcurrent <= 0 {
		maxConcurrent = 3
	}

	results := make([]FetchedContent, len(urls))
	semaphore := make(chan struct{}, maxConcurrent)

	type result struct {
		index   int
		content FetchedContent
	}

	resultChan := make(chan result, len(urls))

	// Lançar goroutines para fetch paralelo
	for i, url := range urls {
		go func(idx int, u string) {
			semaphore <- struct{}{}        // Adquirir semaphore
			defer func() { <-semaphore }() // Liberar semaphore

			content := f.FetchContent(ctx, u)
			resultChan <- result{index: idx, content: content}
		}(i, url)
	}

	// Coletar resultados
	for i := 0; i < len(urls); i++ {
		res := <-resultChan
		results[res.index] = res.content
	}

	close(resultChan)

	return results
}
