package observability

import (
	"context"
	"time"

	"github.com/johnpitter/ollama-code/internal/intent"
)

// HandlerWrapper envolve handler com observabilidade
type HandlerWrapper struct {
	handler Handler
	logger  *Logger
	metrics *MetricsCollector
	tracer  *Tracer
	name    string
}

// Handler interface para compatibilidade com handlers
type Handler interface {
	Handle(ctx context.Context, deps interface{}, result *intent.DetectionResult) (string, error)
}

// NewHandlerWrapper cria wrapper com observabilidade
func NewHandlerWrapper(handler Handler, name string, logger *Logger, metrics *MetricsCollector, tracer *Tracer) *HandlerWrapper {
	return &HandlerWrapper{
		handler: handler,
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
		name:    name,
	}
}

// Handle executa handler com observabilidade
func (h *HandlerWrapper) Handle(ctx context.Context, deps interface{}, result *intent.DetectionResult) (string, error) {
	// Criar span
	ctx, span := h.tracer.StartSpan(ctx, "handler:"+h.name)
	defer h.tracer.EndSpan(span)

	span.AddTag("handler", h.name)
	span.AddTag("intent", string(result.Intent))
	span.AddTag("confidence", formatFloat(result.Confidence))

	// Log início
	h.logger.LogHandlerStart(ctx, h.name, string(result.Intent))

	// Medir tempo
	start := time.Now()

	// Executar handler
	response, err := h.handler.Handle(ctx, deps, result)

	// Calcular duração
	duration := time.Since(start)

	// Registrar métricas
	h.metrics.RecordHandlerDuration(h.name, duration)
	if err != nil {
		h.metrics.RecordHandlerError(h.name)
		span.SetError(err)
	}

	// Log fim
	h.logger.LogHandlerEnd(ctx, h.name, duration, err)

	return response, err
}

// ToolWrapper envolve execução de tool com observabilidade
type ToolWrapper struct {
	logger  *Logger
	metrics *MetricsCollector
	tracer  *Tracer
}

// NewToolWrapper cria wrapper de tool
func NewToolWrapper(logger *Logger, metrics *MetricsCollector, tracer *Tracer) *ToolWrapper {
	return &ToolWrapper{
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}
}

// WrapToolExecution envolve execução de tool
func (t *ToolWrapper) WrapToolExecution(ctx context.Context, toolName string, fn func() error) error {
	// Criar span
	ctx, span := t.tracer.StartSpan(ctx, "tool:"+toolName)
	defer t.tracer.EndSpan(span)

	span.AddTag("tool", toolName)

	// Medir tempo
	start := time.Now()

	// Executar
	err := fn()

	// Calcular duração
	duration := time.Since(start)

	// Registrar métricas
	t.metrics.RecordToolDuration(toolName, duration)
	success := err == nil

	// Log
	t.logger.LogToolExecution(ctx, toolName, duration, success)

	if err != nil {
		span.SetError(err)
	}

	return err
}

// LLMWrapper envolve requisições LLM com observabilidade
type LLMWrapper struct {
	logger  *Logger
	metrics *MetricsCollector
	tracer  *Tracer
}

// NewLLMWrapper cria wrapper de LLM
func NewLLMWrapper(logger *Logger, metrics *MetricsCollector, tracer *Tracer) *LLMWrapper {
	return &LLMWrapper{
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}
}

// WrapLLMRequest envolve requisição LLM
func (l *LLMWrapper) WrapLLMRequest(ctx context.Context, model string, fn func() (int, error)) (int, error) {
	// Criar span
	ctx, span := l.tracer.StartSpan(ctx, "llm:request")
	defer l.tracer.EndSpan(span)

	span.AddTag("model", model)

	// Medir tempo
	start := time.Now()

	// Executar
	tokens, err := fn()

	// Calcular duração
	duration := time.Since(start)

	// Registrar métricas
	l.metrics.RecordLLMDuration(duration)

	// Log
	l.logger.LogLLMRequest(ctx, model, tokens, duration)

	if err != nil {
		span.SetError(err)
	}

	span.AddTag("tokens", formatInt(tokens))

	return tokens, err
}

// IntentWrapper envolve detecção de intenção com observabilidade
type IntentWrapper struct {
	logger  *Logger
	metrics *MetricsCollector
	tracer  *Tracer
}

// NewIntentWrapper cria wrapper de intent
func NewIntentWrapper(logger *Logger, metrics *MetricsCollector, tracer *Tracer) *IntentWrapper {
	return &IntentWrapper{
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}
}

// WrapIntentDetection envolve detecção de intenção
func (i *IntentWrapper) WrapIntentDetection(ctx context.Context, fn func() (*intent.DetectionResult, error)) (*intent.DetectionResult, error) {
	// Criar span
	ctx, span := i.tracer.StartSpan(ctx, "intent:detection")
	defer i.tracer.EndSpan(span)

	// Medir tempo
	start := time.Now()

	// Executar
	result, err := fn()

	// Calcular duração
	duration := time.Since(start)

	// Registrar métricas
	i.metrics.RecordIntentDuration(duration)

	if err != nil {
		span.SetError(err)
		return result, err
	}

	// Log
	i.logger.LogIntentDetection(ctx, string(result.Intent), result.Confidence, duration)

	span.AddTag("intent", string(result.Intent))
	span.AddTag("confidence", formatFloat(result.Confidence))

	return result, nil
}

// CacheWrapper envolve operações de cache com observabilidade
type CacheWrapper struct {
	logger  *Logger
	metrics *MetricsCollector
}

// NewCacheWrapper cria wrapper de cache
func NewCacheWrapper(logger *Logger, metrics *MetricsCollector) *CacheWrapper {
	return &CacheWrapper{
		logger:  logger,
		metrics: metrics,
	}
}

// WrapCacheGet envolve get de cache
func (c *CacheWrapper) WrapCacheGet(ctx context.Context, key string, fn func() (interface{}, bool)) (interface{}, bool) {
	value, hit := fn()

	// Registrar métricas
	c.metrics.RecordCacheHit(hit)

	// Log
	c.logger.LogCacheHit(ctx, key, hit)

	return value, hit
}

// Helpers

func formatFloat(f float64) string {
	return formatFloatPrec(f, 2)
}

func formatFloatPrec(f float64, prec int) string {
	format := "%." + formatInt(prec) + "f"
	return formatValue(format, f)
}

func formatInt(i int) string {
	return formatValue("%d", i)
}

func formatValue(format string, value interface{}) string {
	// Implementação simplificada
	switch v := value.(type) {
	case int:
		if format == "%d" {
			return itoa(v)
		}
	case float64:
		// Conversão básica de float para string
		return ftoa(v, 2)
	}
	return ""
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}

	neg := i < 0
	if neg {
		i = -i
	}

	var buf [20]byte
	pos := len(buf)

	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}

	if neg {
		pos--
		buf[pos] = '-'
	}

	return string(buf[pos:])
}

func ftoa(f float64, prec int) string {
	// Conversão simplificada
	intPart := int(f)
	fracPart := int((f - float64(intPart)) * pow10(prec))

	if fracPart < 0 {
		fracPart = -fracPart
	}

	result := itoa(intPart) + "."

	// Adicionar zeros à esquerda se necessário
	zeros := prec - len(itoa(fracPart))
	for i := 0; i < zeros; i++ {
		result += "0"
	}

	result += itoa(fracPart)
	return result
}

func pow10(n int) float64 {
	result := 1.0
	for i := 0; i < n; i++ {
		result *= 10
	}
	return result
}
