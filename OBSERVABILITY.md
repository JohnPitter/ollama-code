# Sistema de Observabilidade

## 📊 Visão Geral

O sistema de observabilidade do ollama-code fornece **logging estruturado**, **métricas de performance** e **distributed tracing** para monitorar e debug

ar a aplicação.

## 🏗️ Componentes

### 1. Logger Estruturado (`observability.Logger`)

Logger baseado em `log/slog` (Go 1.21+) com funcionalidades extras:

```go
// Criar logger
logger := observability.NewLogger(observability.LoggerConfig{
    Level:     observability.LogLevelInfo,
    Format:    observability.LogFormatJSON,  // ou LogFormatText
    AddSource: true,  // Adiciona file:line nos logs
})

// Usar logger
logger.Info("Processando requisição", "user_id", 123, "action", "create")
logger.Error("Erro ao processar", "error", err.Error())
```

**Níveis de log:**
- `LogLevelDebug` - Logs detalhados
- `LogLevelInfo` - Informações gerais
- `LogLevelWarn` - Avisos
- `LogLevelError` - Erros

**Formatos:**
- `LogFormatText` - Texto legível (padrão)
- `LogFormatJSON` - JSON estruturado

### 2. Coletor de Métricas (`observability.MetricsCollector`)

Coleta métricas de performance em memória:

```go
metrics := observability.NewMetricsCollector()

// Registrar métricas
metrics.RecordHandlerDuration("file_read", 150*time.Millisecond)
metrics.RecordToolDuration("file_reader", 50*time.Millisecond)
metrics.RecordLLMDuration(2*time.Second)
metrics.RecordCacheHit(true)

// Obter estatísticas
stats := metrics.GetHandlerStats("file_read")
fmt.Printf("P50: %.0fms, P95: %.0fms, P99: %.0fms\n",
    stats.P50, stats.P95, stats.P99)

// Imprimir sumário
fmt.Println(metrics.PrintSummary())
```

**Métricas coletadas:**
- Duração de handlers (p50, p95, p99)
- Duração de tools
- Duração de requisições LLM
- Duração de detecção de intenção
- Taxa de hit do cache
- Contagem e taxa de erros

### 3. Distributed Tracing (`observability.Tracer`)

Sistema de tracing para acompanhar execução através de múltiplos componentes:

```go
tracer := observability.NewTracer(logger)

// Criar span raiz
ctx, span := tracer.StartSpan(ctx, "handle_request")
span.AddTag("user_id", "123")
defer tracer.EndSpan(span)

// Criar span filho
ctx, childSpan := tracer.StartSpan(ctx, "read_file")
childSpan.AddTag("file", "example.txt")
// ... fazer trabalho ...
tracer.EndSpan(childSpan)

// Visualizar trace
tree := tracer.GetTraceTree(span.TraceID)
fmt.Println(tree)
```

**Output:**
```
└─ handle_request (250ms)
   • started
   └─ read_file (100ms)
      • file opened
      • file read
```

### 4. Middleware de Observabilidade

Wrappers que adicionam observabilidade automaticamente:

```go
obs := observability.NewDefault()

// Wrapper de handler
wrappedHandler := obs.NewHandlerWrapper(handler, "file_read")
response, err := wrappedHandler.Handle(ctx, deps, result)
// Automaticamente loga início/fim, registra métricas, cria spans

// Wrapper de tool
toolWrapper := obs.NewToolWrapper()
err := toolWrapper.WrapToolExecution(ctx, "file_reader", func() error {
    return tool.Execute()
})

// Wrapper de LLM
llmWrapper := obs.NewLLMWrapper()
tokens, err := llmWrapper.WrapLLMRequest(ctx, "qwen2.5-coder", func() (int, error) {
    return llmClient.Complete(ctx, messages)
})
```

## 📦 Integração com DI

### Habilitando Observabilidade

```go
import "github.com/johnpitter/ollama-code/internal/di"

cfg := &di.Config{
    OllamaURL:           "http://localhost:11434",
    Model:               "qwen2.5-coder:7b",
    EnableObservability: true,  // Habilitar observabilidade
    ObservabilityConfig: observability.LoggerConfig{
        Level:     observability.LogLevelInfo,
        Format:    observability.LogFormatJSON,
        AddSource: true,
    },
}

agent, err := di.InitializeAgent(cfg)
// Agent agora tem observabilidade integrada
```

### Acessando Componentes

```go
// Logger
agent.Observability.Logger.Info("Mensagem")

// Métricas
agent.Observability.Metrics.RecordHandlerDuration(...)

// Tracer
ctx, span := agent.Observability.Tracer.StartSpan(ctx, "operation")
defer agent.Observability.Tracer.EndSpan(span)

// Imprimir sumário
fmt.Println(agent.Observability.PrintSummary())
```

## 🎯 Casos de Uso

### 1. Debug de Performance

```go
// Ver quais handlers são mais lentos
summary := agent.Observability.PrintSummary()
fmt.Println(summary)

// Output:
// 📊 Métricas de Performance
//
// 🎯 Handlers:
//   • file_write: 45 execuções (0.0% erros) - p50: 120ms, p95: 250ms, p99: 500ms
//   • file_read: 120 execuções (0.0% erros) - p50: 50ms, p95: 100ms, p99: 150ms
```

### 2. Investigar Erros

```go
// Ver trace de requisição que falhou
trace := agent.Observability.Tracer.GetTraceTree(traceID)
fmt.Println(trace)

// Output:
// └─ handler:file_write (250ms) ❌ file not found
//    └─ tool:file_writer (100ms) ❌ permission denied
```

### 3. Monitorar Cache

```go
cacheStats := agent.Observability.Metrics.GetCacheStats()
fmt.Printf("Cache hit rate: %.1f%%\n", cacheStats.HitRate)

// Output:
// Cache hit rate: 85.3%
```

### 4. Detectar Gargalos

```go
// Ver qual LLM request é mais lento
llmStats := agent.Observability.Metrics.GetLLMStats()
fmt.Printf("LLM p95: %.0fms\n", llmStats.P95)

// Ver qual tool é mais lento
for _, tool := range agent.Observability.Metrics.GetAllTools() {
    stats := agent.Observability.Metrics.GetToolStats(tool)
    fmt.Printf("%s: p95=%.0fms\n", tool, stats.P95)
}
```

## 🔧 Configuração Avançada

### Logger Personalizado

```go
import "os"

logger := observability.NewLogger(observability.LoggerConfig{
    Level:      observability.LogLevelDebug,
    Format:     observability.LogFormatJSON,
    Output:     os.Stderr,  // Log para stderr
    AddSource:  true,       // Mostrar file:line
    TimeFormat: time.RFC3339Nano,
})
```

### Filtros de Log

```go
// Logger com componente específico
componentLogger := logger.WithComponent("file_handler")
componentLogger.Info("Processing file")
// Output: level=INFO component=file_handler msg="Processing file"

// Logger com contexto
ctx := context.WithValue(ctx, traceIDKey, "abc123")
contextLogger := logger.WithContext(ctx)
contextLogger.Info("Processing")
// Output: level=INFO trace_id=abc123 msg="Processing"
```

### Métricas Customizadas

```go
// Resetar métricas
agent.Observability.Metrics.Reset()

// Obter estatísticas específicas
stats := agent.Observability.Metrics.GetHandlerStats("my_handler")
if stats != nil {
    fmt.Printf("Min: %.0fms, Max: %.0fms, Mean: %.0fms\n",
        stats.Min, stats.Max, stats.Mean)
}

// Taxa de erro de handler
errorRate := agent.Observability.Metrics.GetHandlerErrorRate("my_handler")
fmt.Printf("Error rate: %.1f%%\n", errorRate)
```

## 📈 Exemplos Práticos

### Exemplo 1: Handler com Observabilidade

```go
func (h *MyHandler) Handle(ctx context.Context, deps *handlers.Dependencies, result *intent.DetectionResult) (string, error) {
    // Criar span
    ctx, span := deps.Tracer.StartSpan(ctx, "my_handler:execute")
    defer deps.Tracer.EndSpan(span)

    span.AddTag("intent", string(result.Intent))

    // Log início
    deps.Logger.LogHandlerStart(ctx, "my_handler", string(result.Intent))

    start := time.Now()

    // Executar lógica
    response, err := h.doWork(ctx)

    // Registrar métricas
    duration := time.Since(start)
    deps.Metrics.RecordHandlerDuration("my_handler", duration)

    if err != nil {
        span.SetError(err)
        deps.Metrics.RecordHandlerError("my_handler")
        deps.Logger.Error("Handler failed", "error", err.Error())
        return "", err
    }

    // Log fim
    deps.Logger.LogHandlerEnd(ctx, "my_handler", duration, nil)

    return response, nil
}
```

### Exemplo 2: Monitoramento em Produção

```go
// Configurar para produção
cfg := &di.Config{
    EnableObservability: true,
    ObservabilityConfig: observability.LoggerConfig{
        Level:     observability.LogLevelWarn,  // Apenas warns e errors
        Format:    observability.LogFormatJSON, // JSON para parsing
        AddSource: true,                        // Debug info
    },
}

agent, _ := di.InitializeAgent(cfg)

// Periodicamente exportar métricas
go func() {
    ticker := time.NewTicker(1 * time.Minute)
    for range ticker.C {
        summary := agent.Observability.PrintSummary()
        // Enviar para sistema de monitoramento
        sendToMonitoring(summary)
    }
}()
```

### Exemplo 3: Debug de Problema

```go
// Habilitar debug logging temporariamente
cfg := &di.Config{
    EnableObservability: true,
    ObservabilityConfig: observability.LoggerConfig{
        Level:     observability.LogLevelDebug,
        AddSource: true,
    },
}

agent, _ := di.InitializeAgent(cfg)

// Processar requisição problemática
agent.ProcessMessage(ctx, "file write test.txt")

// Ver trace completo
fmt.Println(agent.Observability.Tracer.PrintAllTraces())

// Ver métricas
fmt.Println(agent.Observability.PrintSummary())
```

## 🧪 Testando com Observabilidade

```go
func TestWithObservability(t *testing.T) {
    obs := observability.NewDefault()

    // Criar handler com observabilidade
    handler := NewMyHandler()
    wrapped := obs.NewHandlerWrapper(handler, "test_handler")

    // Executar
    ctx := context.Background()
    _, err := wrapped.Handle(ctx, deps, result)

    // Verificar métricas
    stats := obs.Metrics.GetHandlerStats("test_handler")
    if stats.Count != 1 {
        t.Errorf("Expected 1 execution, got %d", stats.Count)
    }

    // Verificar traces
    spans := obs.Tracer.GetSpans()
    if len(spans) != 1 {
        t.Errorf("Expected 1 span, got %d", len(spans))
    }
}
```

## 📊 Estrutura de Dados

### Stats

```go
type Stats struct {
    Count  int      // Número de medições
    Min    float64  // Valor mínimo
    Max    float64  // Valor máximo
    Mean   float64  // Média
    Median float64  // Mediana
    P50    float64  // Percentil 50
    P95    float64  // Percentil 95
    P99    float64  // Percentil 99
}
```

### CacheStats

```go
type CacheStats struct {
    Hits    int64   // Número de hits
    Misses  int64   // Número de misses
    Total   int64   // Total de acessos
    HitRate float64 // Taxa de hit (%)
}
```

### Span

```go
type Span struct {
    TraceID   string                 // ID do trace
    SpanID    string                 // ID do span
    ParentID  string                 // ID do span pai
    Name      string                 // Nome da operação
    StartTime time.Time              // Início
    EndTime   time.Time              // Fim
    Duration  time.Duration          // Duração
    Tags      map[string]string      // Tags
    Events    []SpanEvent            // Eventos
    Error     error                  // Erro (se houver)
}
```

## 🎓 Boas Práticas

1. **Use níveis apropriados:**
   - Debug: Informações detalhadas de desenvolvimento
   - Info: Eventos importantes da aplicação
   - Warn: Situações anormais mas recuperáveis
   - Error: Erros que precisam atenção

2. **Adicione contexto:**
   ```go
   logger.Info("File processed",
       "file", filename,
       "size", fileSize,
       "duration_ms", duration.Milliseconds(),
   )
   ```

3. **Use spans hierárquicos:**
   ```go
   ctx, rootSpan := tracer.StartSpan(ctx, "process_request")
   defer tracer.EndSpan(rootSpan)

   ctx, childSpan := tracer.StartSpan(ctx, "validate_input")
   defer tracer.EndSpan(childSpan)
   ```

4. **Monitore métricas críticas:**
   - P95/P99 de handlers
   - Taxa de erro
   - Hit rate do cache
   - Latência do LLM

5. **Resete métricas periodicamente:**
   ```go
   // A cada hora, resetar para evitar memory leak
   ticker := time.NewTicker(1 * time.Hour)
   go func() {
       for range ticker.C {
           obs.Reset()
       }
   }()
   ```

## 📚 Referências

- [Go slog](https://pkg.go.dev/log/slog) - Logger estruturado do Go
- [Distributed Tracing](https://opentracing.io/) - Conceitos de tracing
- [RED Method](https://www.weave.works/blog/the-red-method-key-metrics-for-microservices-architecture/) - Rate, Errors, Duration

---

**Implementado em:** 2024-01-22
**Status:** ✅ Completo e Testado
**Testes:** 9/9 passando
