package observability

// Observability agrega todos os componentes de observabilidade
type Observability struct {
	Logger  *Logger
	Metrics *MetricsCollector
	Tracer  *Tracer
}

// New cria nova instância de observabilidade
func New(config LoggerConfig) *Observability {
	logger := NewLogger(config)
	metrics := NewMetricsCollector()
	tracer := NewTracer(logger)

	return &Observability{
		Logger:  logger,
		Metrics: metrics,
		Tracer:  tracer,
	}
}

// NewDefault cria instância com configuração padrão
func NewDefault() *Observability {
	return New(LoggerConfig{
		Level:     LogLevelInfo,
		Format:    LogFormatText,
		AddSource: false,
	})
}

// PrintSummary imprime sumário completo
func (o *Observability) PrintSummary() string {
	summary := ""

	// Métricas
	summary += o.Metrics.PrintSummary()
	summary += "\n"

	// Traces (se houver)
	if len(o.Tracer.GetSpans()) > 0 {
		summary += o.Tracer.PrintAllTraces()
	}

	return summary
}

// Reset reseta todas as métricas e traces
func (o *Observability) Reset() {
	o.Metrics.Reset()
	o.Tracer.Reset()
}

// NewHandlerWrapper cria wrapper de handler
func (o *Observability) NewHandlerWrapper(handler Handler, name string) *HandlerWrapper {
	return NewHandlerWrapper(handler, name, o.Logger, o.Metrics, o.Tracer)
}

// NewToolWrapper cria wrapper de tool
func (o *Observability) NewToolWrapper() *ToolWrapper {
	return NewToolWrapper(o.Logger, o.Metrics, o.Tracer)
}

// NewLLMWrapper cria wrapper de LLM
func (o *Observability) NewLLMWrapper() *LLMWrapper {
	return NewLLMWrapper(o.Logger, o.Metrics, o.Tracer)
}

// NewIntentWrapper cria wrapper de intent
func (o *Observability) NewIntentWrapper() *IntentWrapper {
	return NewIntentWrapper(o.Logger, o.Metrics, o.Tracer)
}

// NewCacheWrapper cria wrapper de cache
func (o *Observability) NewCacheWrapper() *CacheWrapper {
	return NewCacheWrapper(o.Logger, o.Metrics)
}
