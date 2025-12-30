package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// Span representa um span de tracing
type Span struct {
	TraceID   string
	SpanID    string
	ParentID  string
	Name      string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Tags      map[string]string
	Events    []SpanEvent
	Error     error
}

// SpanEvent evento dentro de um span
type SpanEvent struct {
	Time    time.Time
	Name    string
	Attributes map[string]string
}

// Tracer gerencia tracing
type Tracer struct {
	spans  []*Span
	logger *Logger
}

// NewTracer cria novo tracer
func NewTracer(logger *Logger) *Tracer {
	return &Tracer{
		spans:  make([]*Span, 0, 100),
		logger: logger,
	}
}

// StartSpan inicia novo span
func (t *Tracer) StartSpan(ctx context.Context, name string) (context.Context, *Span) {
	span := &Span{
		TraceID:   getOrCreateTraceID(ctx),
		SpanID:    generateID(),
		Name:      name,
		StartTime: time.Now(),
		Tags:      make(map[string]string),
		Events:    make([]SpanEvent, 0),
	}

	// Se existe span pai no contexto
	if parentSpan := GetSpanFromContext(ctx); parentSpan != nil {
		span.ParentID = parentSpan.SpanID
		span.TraceID = parentSpan.TraceID
	}

	// Adicionar span ao contexto
	ctx = context.WithValue(ctx, spanKey, span)
	ctx = context.WithValue(ctx, traceIDKey, span.TraceID)

	return ctx, span
}

// EndSpan finaliza span
func (t *Tracer) EndSpan(span *Span) {
	if span == nil {
		return
	}

	span.EndTime = time.Now()
	span.Duration = span.EndTime.Sub(span.StartTime)

	// Adicionar à lista de spans
	t.spans = append(t.spans, span)

	// Log span completion
	if t.logger != nil {
		attrs := []any{
			"trace_id", span.TraceID,
			"span_id", span.SpanID,
			"duration_ms", span.Duration.Milliseconds(),
		}

		if span.Error != nil {
			attrs = append(attrs, "error", span.Error.Error())
			t.logger.Error("span_complete", attrs...)
		} else {
			t.logger.Debug("span_complete", attrs...)
		}
	}

	// Manter apenas últimos 1000 spans
	if len(t.spans) > 1000 {
		t.spans = t.spans[1:]
	}
}

// AddTag adiciona tag ao span
func (s *Span) AddTag(key, value string) {
	if s.Tags == nil {
		s.Tags = make(map[string]string)
	}
	s.Tags[key] = value
}

// AddEvent adiciona evento ao span
func (s *Span) AddEvent(name string, attributes map[string]string) {
	s.Events = append(s.Events, SpanEvent{
		Time:       time.Now(),
		Name:       name,
		Attributes: attributes,
	})
}

// SetError marca span com erro
func (s *Span) SetError(err error) {
	s.Error = err
	s.AddTag("error", "true")
}

// GetSpans retorna todos os spans
func (t *Tracer) GetSpans() []*Span {
	return t.spans
}

// GetSpansByTrace retorna spans de um trace específico
func (t *Tracer) GetSpansByTrace(traceID string) []*Span {
	spans := make([]*Span, 0)
	for _, span := range t.spans {
		if span.TraceID == traceID {
			spans = append(spans, span)
		}
	}
	return spans
}

// GetTraceTree retorna árvore de spans de um trace
func (t *Tracer) GetTraceTree(traceID string) string {
	spans := t.GetSpansByTrace(traceID)
	if len(spans) == 0 {
		return fmt.Sprintf("No spans found for trace %s", traceID)
	}

	// Encontrar root span (sem parent)
	var root *Span
	spanMap := make(map[string]*Span)
	for _, span := range spans {
		spanMap[span.SpanID] = span
		if span.ParentID == "" {
			root = span
		}
	}

	if root == nil {
		return "No root span found"
	}

	// Construir árvore
	return buildTree(root, spanMap, 0)
}

// buildTree constrói representação em texto da árvore
func buildTree(span *Span, spanMap map[string]*Span, depth int) string {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	errorStr := ""
	if span.Error != nil {
		errorStr = fmt.Sprintf(" ❌ %v", span.Error)
	}

	result := fmt.Sprintf("%s└─ %s (%.0fms)%s\n", indent, span.Name, float64(span.Duration.Milliseconds()), errorStr)

	// Adicionar eventos
	for _, event := range span.Events {
		result += fmt.Sprintf("%s   • %s\n", indent, event.Name)
	}

	// Adicionar filhos
	for _, childSpan := range spanMap {
		if childSpan.ParentID == span.SpanID {
			result += buildTree(childSpan, spanMap, depth+1)
		}
	}

	return result
}

// PrintAllTraces imprime todos os traces
func (t *Tracer) PrintAllTraces() string {
	// Agrupar por trace_id
	traces := make(map[string][]*Span)
	for _, span := range t.spans {
		traces[span.TraceID] = append(traces[span.TraceID], span)
	}

	result := fmt.Sprintf("📍 Total de traces: %d\n\n", len(traces))

	for traceID := range traces {
		result += fmt.Sprintf("Trace ID: %s\n", traceID)
		result += t.GetTraceTree(traceID)
		result += "\n"
	}

	return result
}

// Reset reseta todos os spans
func (t *Tracer) Reset() {
	t.spans = make([]*Span, 0, 100)
}

// Context keys
type ctxKey string

const spanKey ctxKey = "span"

// GetSpanFromContext extrai span do contexto
func GetSpanFromContext(ctx context.Context) *Span {
	if span, ok := ctx.Value(spanKey).(*Span); ok {
		return span
	}
	return nil
}

// getOrCreateTraceID obtém ou cria trace ID do contexto
func getOrCreateTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return generateID()
}

// generateID gera ID aleatório
func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
