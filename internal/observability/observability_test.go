package observability

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	logger := NewDefaultLogger()
	if logger == nil {
		t.Fatal("Expected logger to be created")
	}

	logger.Info("test message", "key", "value")
	logger.Debug("debug message")
	logger.Warn("warning message")
	logger.Error("error message")
}

func TestLoggerWithFields(t *testing.T) {
	logger := NewDefaultLogger()
	componentLogger := logger.WithComponent("test")

	componentLogger.Info("message from component")
}

func TestMetricsCollector(t *testing.T) {
	metrics := NewMetricsCollector()

	// Registrar métricas de handler
	metrics.RecordHandlerDuration("test_handler", 100*time.Millisecond)
	metrics.RecordHandlerDuration("test_handler", 200*time.Millisecond)
	metrics.RecordHandlerDuration("test_handler", 150*time.Millisecond)

	stats := metrics.GetHandlerStats("test_handler")
	if stats == nil {
		t.Fatal("Expected stats to be available")
	}

	if stats.Count != 3 {
		t.Errorf("Expected count=3, got %d", stats.Count)
	}

	if stats.Min != 100 {
		t.Errorf("Expected min=100, got %.0f", stats.Min)
	}

	if stats.Max != 200 {
		t.Errorf("Expected max=200, got %.0f", stats.Max)
	}
}

func TestMetricsHandlerError(t *testing.T) {
	metrics := NewMetricsCollector()

	metrics.RecordHandlerDuration("test_handler", 100*time.Millisecond)
	metrics.RecordHandlerError("test_handler")
	metrics.RecordHandlerDuration("test_handler", 200*time.Millisecond)

	errorRate := metrics.GetHandlerErrorRate("test_handler")
	if errorRate != 50.0 {
		t.Errorf("Expected error rate=50%%, got %.1f%%", errorRate)
	}
}

func TestMetricsCacheStats(t *testing.T) {
	metrics := NewMetricsCollector()

	metrics.RecordCacheHit(true)
	metrics.RecordCacheHit(true)
	metrics.RecordCacheHit(false)

	stats := metrics.GetCacheStats()
	if stats.Hits != 2 {
		t.Errorf("Expected 2 hits, got %d", stats.Hits)
	}

	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}

	expectedHitRate := 66.7
	if stats.HitRate < expectedHitRate-0.1 || stats.HitRate > expectedHitRate+0.1 {
		t.Errorf("Expected hit rate ~%.1f%%, got %.1f%%", expectedHitRate, stats.HitRate)
	}
}

func TestTracer(t *testing.T) {
	logger := NewDefaultLogger()
	tracer := NewTracer(logger)

	ctx := context.Background()

	// Criar span raiz
	ctx, rootSpan := tracer.StartSpan(ctx, "root_operation")
	rootSpan.AddTag("type", "test")

	// Simular trabalho
	time.Sleep(10 * time.Millisecond)

	// Criar span filho
	ctx, childSpan := tracer.StartSpan(ctx, "child_operation")
	childSpan.AddTag("type", "sub-test")
	time.Sleep(5 * time.Millisecond)
	tracer.EndSpan(childSpan)

	// Finalizar root
	tracer.EndSpan(rootSpan)

	// Verificar spans
	spans := tracer.GetSpans()
	if len(spans) != 2 {
		t.Errorf("Expected 2 spans, got %d", len(spans))
	}

	// Verificar trace tree
	tree := tracer.GetTraceTree(rootSpan.TraceID)
	if tree == "" {
		t.Error("Expected trace tree to be generated")
	}
}

func TestSpanWithError(t *testing.T) {
	logger := NewDefaultLogger()
	tracer := NewTracer(logger)

	ctx := context.Background()
	ctx, span := tracer.StartSpan(ctx, "failing_operation")

	err := performFailingOperation()
	span.SetError(err)

	tracer.EndSpan(span)

	if span.Error == nil {
		t.Error("Expected span to have error")
	}
}

func TestObservabilityIntegration(t *testing.T) {
	obs := NewDefault()

	// Testar logger
	obs.Logger.Info("test")

	// Testar métricas
	obs.Metrics.RecordHandlerDuration("test", 100*time.Millisecond)

	// Testar tracer
	ctx := context.Background()
	ctx, span := obs.Tracer.StartSpan(ctx, "test")
	obs.Tracer.EndSpan(span)

	// Gerar summary
	summary := obs.PrintSummary()
	if summary == "" {
		t.Error("Expected summary to be generated")
	}
}

func TestPrintSummary(t *testing.T) {
	metrics := NewMetricsCollector()

	// Adicionar algumas métricas
	metrics.RecordHandlerDuration("handler1", 100*time.Millisecond)
	metrics.RecordHandlerDuration("handler1", 200*time.Millisecond)
	metrics.RecordHandlerDuration("handler2", 50*time.Millisecond)

	metrics.RecordToolDuration("tool1", 30*time.Millisecond)

	metrics.RecordCacheHit(true)
	metrics.RecordCacheHit(false)

	summary := metrics.PrintSummary()
	if summary == "" {
		t.Error("Expected summary to be generated")
	}

	t.Logf("Summary:\n%s", summary)
}

// Helper function
func performFailingOperation() error {
	return fmt.Errorf("operation failed")
}
