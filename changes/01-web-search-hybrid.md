# Web Search Híbrido - Changelog

**Data:** 2024-12-19
**Commits:** `978b2f0`, `3cc6b3e`
**Autor:** Claude AI

## Resumo

Implementação de um sistema híbrido de web search que não apenas busca resultados no DuckDuckGo, mas também faz fetch do conteúdo real das páginas, extrai texto limpo e sintetiza respostas usando o LLM com informações atualizadas da internet.

## Problema Original

O sistema de web search anterior tinha limitações:
- Apenas retornava snippets dos resultados de busca
- Não acessava o conteúdo real das páginas
- LLM respondia "não tenho acesso à internet" mesmo após buscar
- Parsing do HTML do DuckDuckGo tinha bugs (loop infinito)
- Mensagens duplicadas na interface

## Solução Implementada

### 1. ContentFetcher (`internal/websearch/fetcher.go`)

**Novo arquivo** para buscar e processar conteúdo HTML:

```go
type ContentFetcher struct {
    client *http.Client
}

type FetchedContent struct {
    URL     string
    Title   string
    Content string
    Error   string
}
```

**Funcionalidades:**
- `FetchContent(ctx, url)`: Busca HTML e extrai conteúdo principal
- `FetchMultiple(ctx, urls, maxConcurrent)`: Fetch paralelo com semaphore
- `extractTitle()`: Extrai título da página
- `extractMainContent()`: Extrai conteúdo de `<article>`, `<main>` ou `<body>`
- Remove scripts, styles, nav, header, footer, ads
- Strip HTML tags e limpa whitespace
- Limita a 3000 caracteres por página

**Técnicas de Web Scraping:**
- User-Agent spoofing para evitar detecção de bots
- Timeout de 15s por requisição
- Rate limiting com semaphore (max 3 concurrent)
- Decodificação de HTML entities
- Limpeza de múltiplos espaços e newlines

### 2. Melhorias no Orchestrator (`internal/websearch/orchestrator.go`)

**Adicionado:**
- Campo `fetcher *ContentFetcher`
- Método `FetchContents(ctx, results, maxResults)`: Wrapper para fetch paralelo
- Função `extractRealURL(rawURL)`: Decodifica URLs do DuckDuckGo

**URLs do DuckDuckGo:**
O DuckDuckGo codifica URLs como:
```
//duckduckgo.com/l/?uddg=https%3A%2F%2Fexample.com
```

A função `extractRealURL` decodifica para:
```
https://example.com
```

**Timeout reduzido:**
- De 30s para 10s para evitar travamentos
- Context com timeout em todas requisições

**Fix do Loop Infinito:**
```go
// ANTES (bug):
titleStart = strings.Index(html[titleStart:], `class="result__a"`)
titleStart += len(`class="result__a"`) // índice relativo!

// DEPOIS (correto):
searchPos := 0
foundIndex := strings.Index(html[searchPos:], `class="result__a"`)
titleStart := searchPos + foundIndex // índice absoluto
searchPos = titleStart + 1 // avançar
```

**Validações adicionadas:**
- Ignorar URLs vazias
- Ignorar URLs sem `http://` ou `https://`
- Ignorar resultados sem título

### 3. Handler Aprimorado (`internal/agent/handlers.go`)

**handleWebSearch() reescrito:**

```go
func (a *Agent) handleWebSearch(ctx, result, userMessage) {
    // 1. Buscar no DuckDuckGo
    results := a.webSearch.Search(ctx, query, ["duckduckgo"])

    // 2. Fazer fetch de conteúdo real (top 3)
    fetchedContents := a.webSearch.FetchContents(ctx, results, 3)

    // 3. Construir contexto com conteúdo completo
    for content := range fetchedContents {
        contextBuilder += content.Title + "\n" + content.Content
    }

    // 4. Sintetizar com LLM usando streaming
    response := a.llmClient.CompleteStreaming(ctx, prompt, callback)
}
```

**Prompt melhorado para o LLM:**
```
Você acabou de buscar informações atualizadas na internet.
Use SOMENTE as informações dos sites abaixo para responder.

IMPORTANTE:
- Use APENAS as informações fornecidas acima
- NÃO diga que não tem acesso à internet ou dados em tempo real
- Você ACABOU de buscar essas informações na web
- Forneça uma resposta direta e objetiva baseada no conteúdo obtido
```

**Fallback robusto:**
Se o fetch de conteúdo falhar, cai back para `synthesizeFromSnippets()`:
- Usa apenas snippets dos resultados de busca
- Streaming com LLM
- Prompt adaptado para snippets
- Retorna mensagem clara se snippets vazios

**Logs detalhados:**
- `📄 Encontrados N resultados, buscando conteúdo...`
- `✓ Conteúdo obtido de URL (N chars)`
- `⚠️ Erro ao buscar URL: erro`
- `⚠️ Conteúdo vazio de URL`
- `✓ N fontes com conteúdo válido`
- `ℹ️ Usando snippets de pesquisa...`

### 4. Fix de Mensagens Duplicadas

**Problema:**
`handleQuestion()` printava durante streaming, depois `agent.go` printava novamente.

**Solução:**
```go
// agent.go - só printar se NÃO for question
if detectionResult.Intent != intent.IntentQuestion {
    a.colorGreen.Println("\n🤖 Assistente:")
    fmt.Println(response)
}

// handlers.go - printar header ANTES do streaming
a.colorGreen.Println("\n🤖 Assistente:")
response, err := a.llmClient.CompleteStreaming(ctx, messages, opts, callback)
fmt.Println() // newline após streaming
```

## Resultados

### Teste 1: Temperatura em Recife
```bash
$ ollama-code ask "qual a temperatura em recife hoje"

🔍 Detectando intenção...
Intenção: web_search (confiança: 95%)
🌐 Pesquisando na web: qual a temperatura em recife hoje
📄 Encontrados 5 resultados, buscando conteúdo...
✓ Conteúdo obtido de https://www.climatempo.com.br/... (3003 chars)
✓ Conteúdo obtido de https://www.tempo.com/recife.htm (3003 chars)
✓ Conteúdo obtido de https://tempoagora.uol.com.br/... (3003 chars)
✓ 3 fontes com conteúdo válido

🤖 Assistente:
A temperatura atual em Recife é de 27°C, com sensação térmica de 30°C.
A previsão indica pancadas de chuva durante o dia e tempo firme à noite.

Fonte: Climatempo, Tempo.com
```

### Teste 2: Go 1.23 Features
```bash
$ ollama-code ask "o que há de novo no go 1.23"

🌐 Pesquisando na web: o que há de novo no go 1.23
📄 Encontrados 5 resultados, buscando conteúdo...
✓ Conteúdo obtido de https://go.dev/doc/go1.23 (3003 chars)
✓ 3 fontes com conteúdo válido

🤖 Assistente:
No Go 1.23, foram introduzidas várias funcionalidades:

1. Range over Function Types:
   - Loop for-range aceita funções iteradoras
   - Tipos: func(func() bool), func(func(K) bool), func(func(K, V) bool)

2. Generic Type Aliases (preview):
   - Suporte experimental para aliases genéricos
   - GOEXPERIMENT=aliastypeparams

3. Novos Pacotes:
   - iter: definições de iteradores
   - Melhorias em slices e maps

4. Melhorias no Compilador:
   - Eliminação de operações redundantes
   - Binários menores e mais eficientes

5. Garbage Collection:
   - Pausas menores e mais previsíveis
```

## Métricas

- **Tempo de resposta:** ~5-10s (busca + fetch + LLM)
- **Páginas buscadas:** 5 resultados do DuckDuckGo
- **Conteúdo extraído:** Top 3 URLs, ~3000 chars cada
- **Concorrência:** 3 fetches paralelos
- **Taxa de sucesso:** ~80% (depende de acesso aos sites)

## Arquivos Modificados

1. `internal/websearch/fetcher.go` (NOVO) - 220 linhas
2. `internal/websearch/orchestrator.go` - Adicionado FetchContents(), extractRealURL()
3. `internal/agent/handlers.go` - Reescrito handleWebSearch(), synthesizeFromSnippets()
4. `internal/agent/agent.go` - Fix de mensagens duplicadas

## Limitações Conhecidas

1. **Rate Limiting:** Alguns sites bloqueiam ou limitam requests
2. **JavaScript:** Conteúdo renderizado por JS não é capturado
3. **Paywalls:** Sites com paywall retornam conteúdo limitado
4. **Tamanho:** Limitado a 3000 chars por página (pode perder contexto)
5. **Velocidade:** 10s timeout pode ser lento para alguns sites

## Melhorias Futuras

- [ ] Suporte para JavaScript rendering (Playwright/Selenium)
- [ ] Cache de conteúdo fetched
- [ ] Retry com backoff exponencial
- [ ] Suporte para mais search engines (Google, Bing)
- [ ] Detecção de paywall e fallback
- [ ] Compressão de texto longo com summarização
- [ ] Métricas de qualidade do conteúdo extraído

## Referências

- Commit inicial: `978b2f0`
- Bug fix loop: `3cc6b3e`
- Issue: awesome-claude-code web search
