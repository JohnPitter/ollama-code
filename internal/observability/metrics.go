package observability

import (
	"fmt"
	"sync"
	"time"
)

// MetricType tipo de métrica
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
)

// Metric representa uma métrica
type Metric struct {
	Name   string
	Type   MetricType
	Value  float64
	Labels map[string]string
	Time   time.Time
}

// MetricsCollector coleta métricas
type MetricsCollector struct {
	mu      sync.RWMutex
	metrics map[string]*Metric

	// Histograms específicos
	handlerDurations  map[string][]float64
	toolDurations     map[string][]float64
	llmDurations      []float64
	intentDurations   []float64

	// Contadores
	handlerCounts     map[string]int64
	handlerErrors     map[string]int64
	toolCounts        map[string]int64
	cacheHits         int64
	cacheMisses       int64
}

// NewMetricsCollector cria novo coletor de métricas
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics:          make(map[string]*Metric),
		handlerDurations: make(map[string][]float64),
		toolDurations:    make(map[string][]float64),
		llmDurations:     make([]float64, 0, 1000),
		intentDurations:  make([]float64, 0, 1000),
		handlerCounts:    make(map[string]int64),
		handlerErrors:    make(map[string]int64),
		toolCounts:       make(map[string]int64),
	}
}

// RecordHandlerDuration registra duração de handler
func (m *MetricsCollector) RecordHandlerDuration(handler string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	durationMs := float64(duration.Milliseconds())
	m.handlerDurations[handler] = append(m.handlerDurations[handler], durationMs)
	m.handlerCounts[handler]++

	// Manter apenas últimas 1000 medições
	if len(m.handlerDurations[handler]) > 1000 {
		m.handlerDurations[handler] = m.handlerDurations[handler][1:]
	}
}

// RecordHandlerError registra erro de handler
func (m *MetricsCollector) RecordHandlerError(handler string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlerErrors[handler]++
}

// RecordToolDuration registra duração de tool
func (m *MetricsCollector) RecordToolDuration(tool string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	durationMs := float64(duration.Milliseconds())
	m.toolDurations[tool] = append(m.toolDurations[tool], durationMs)
	m.toolCounts[tool]++

	if len(m.toolDurations[tool]) > 1000 {
		m.toolDurations[tool] = m.toolDurations[tool][1:]
	}
}

// RecordLLMDuration registra duração de requisição LLM
func (m *MetricsCollector) RecordLLMDuration(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	durationMs := float64(duration.Milliseconds())
	m.llmDurations = append(m.llmDurations, durationMs)

	if len(m.llmDurations) > 1000 {
		m.llmDurations = m.llmDurations[1:]
	}
}

// RecordIntentDuration registra duração de detecção de intenção
func (m *MetricsCollector) RecordIntentDuration(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	durationMs := float64(duration.Milliseconds())
	m.intentDurations = append(m.intentDurations, durationMs)

	if len(m.intentDurations) > 1000 {
		m.intentDurations = m.intentDurations[1:]
	}
}

// RecordCacheHit registra hit/miss de cache
func (m *MetricsCollector) RecordCacheHit(hit bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if hit {
		m.cacheHits++
	} else {
		m.cacheMisses++
	}
}

// GetHandlerStats retorna estatísticas de handler
func (m *MetricsCollector) GetHandlerStats(handler string) *Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	durations, exists := m.handlerDurations[handler]
	if !exists || len(durations) == 0 {
		return nil
	}

	return calculateStats(durations)
}

// GetToolStats retorna estatísticas de tool
func (m *MetricsCollector) GetToolStats(tool string) *Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	durations, exists := m.toolDurations[tool]
	if !exists || len(durations) == 0 {
		return nil
	}

	return calculateStats(durations)
}

// GetLLMStats retorna estatísticas de LLM
func (m *MetricsCollector) GetLLMStats() *Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.llmDurations) == 0 {
		return nil
	}

	return calculateStats(m.llmDurations)
}

// GetIntentStats retorna estatísticas de detecção de intenção
func (m *MetricsCollector) GetIntentStats() *Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.intentDurations) == 0 {
		return nil
	}

	return calculateStats(m.intentDurations)
}

// GetCacheStats retorna estatísticas de cache
func (m *MetricsCollector) GetCacheStats() *CacheStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := m.cacheHits + m.cacheMisses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(m.cacheHits) / float64(total) * 100
	}

	return &CacheStats{
		Hits:    m.cacheHits,
		Misses:  m.cacheMisses,
		Total:   total,
		HitRate: hitRate,
	}
}

// GetAllHandlers retorna lista de handlers com métricas
func (m *MetricsCollector) GetAllHandlers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	handlers := make([]string, 0, len(m.handlerCounts))
	for handler := range m.handlerCounts {
		handlers = append(handlers, handler)
	}
	return handlers
}

// GetAllTools retorna lista de tools com métricas
func (m *MetricsCollector) GetAllTools() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tools := make([]string, 0, len(m.toolCounts))
	for tool := range m.toolCounts {
		tools = append(tools, tool)
	}
	return tools
}

// GetHandlerCount retorna contagem de execuções de handler
func (m *MetricsCollector) GetHandlerCount(handler string) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.handlerCounts[handler]
}

// GetHandlerErrorCount retorna contagem de erros de handler
func (m *MetricsCollector) GetHandlerErrorCount(handler string) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.handlerErrors[handler]
}

// GetHandlerErrorRate retorna taxa de erro de handler
func (m *MetricsCollector) GetHandlerErrorRate(handler string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := m.handlerCounts[handler]
	if count == 0 {
		return 0
	}

	errors := m.handlerErrors[handler]
	return float64(errors) / float64(count) * 100
}

// Reset reseta todas as métricas
func (m *MetricsCollector) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics = make(map[string]*Metric)
	m.handlerDurations = make(map[string][]float64)
	m.toolDurations = make(map[string][]float64)
	m.llmDurations = make([]float64, 0, 1000)
	m.intentDurations = make([]float64, 0, 1000)
	m.handlerCounts = make(map[string]int64)
	m.handlerErrors = make(map[string]int64)
	m.toolCounts = make(map[string]int64)
	m.cacheHits = 0
	m.cacheMisses = 0
}

// PrintSummary imprime sumário de métricas
func (m *MetricsCollector) PrintSummary() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summary := "📊 Métricas de Performance\n\n"

	// Handlers
	if len(m.handlerCounts) > 0 {
		summary += "🎯 Handlers:\n"
		for handler, count := range m.handlerCounts {
			errors := m.handlerErrors[handler]
			errorRate := 0.0
			if count > 0 {
				errorRate = float64(errors) / float64(count) * 100
			}

			stats := m.GetHandlerStats(handler)
			if stats != nil {
				summary += fmt.Sprintf("  • %s: %d execuções (%.1f%% erros) - p50: %.0fms, p95: %.0fms, p99: %.0fms\n",
					handler, count, errorRate, stats.P50, stats.P95, stats.P99)
			}
		}
		summary += "\n"
	}

	// Tools
	if len(m.toolCounts) > 0 {
		summary += "🔧 Tools:\n"
		for tool, count := range m.toolCounts {
			stats := m.GetToolStats(tool)
			if stats != nil {
				summary += fmt.Sprintf("  • %s: %d execuções - p50: %.0fms, p95: %.0fms\n",
					tool, count, stats.P50, stats.P95)
			}
		}
		summary += "\n"
	}

	// LLM
	if len(m.llmDurations) > 0 {
		stats := m.GetLLMStats()
		summary += fmt.Sprintf("🤖 LLM: %d requisições - p50: %.0fms, p95: %.0fms, p99: %.0fms\n\n",
			len(m.llmDurations), stats.P50, stats.P95, stats.P99)
	}

	// Cache
	cacheStats := m.GetCacheStats()
	if cacheStats.Total > 0 {
		summary += fmt.Sprintf("💾 Cache: %.1f%% hit rate (%d hits, %d misses)\n",
			cacheStats.HitRate, cacheStats.Hits, cacheStats.Misses)
	}

	return summary
}

// Stats estatísticas calculadas
type Stats struct {
	Count  int
	Min    float64
	Max    float64
	Mean   float64
	Median float64
	P50    float64
	P95    float64
	P99    float64
}

// CacheStats estatísticas de cache
type CacheStats struct {
	Hits    int64
	Misses  int64
	Total   int64
	HitRate float64
}

// calculateStats calcula estatísticas de uma série
func calculateStats(values []float64) *Stats {
	if len(values) == 0 {
		return nil
	}

	// Copiar e ordenar
	sorted := make([]float64, len(values))
	copy(sorted, values)

	// Bubble sort simples (ok para ~1000 valores)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Calcular estatísticas
	sum := 0.0
	for _, v := range sorted {
		sum += v
	}

	return &Stats{
		Count:  len(sorted),
		Min:    sorted[0],
		Max:    sorted[len(sorted)-1],
		Mean:   sum / float64(len(sorted)),
		Median: percentile(sorted, 0.5),
		P50:    percentile(sorted, 0.5),
		P95:    percentile(sorted, 0.95),
		P99:    percentile(sorted, 0.99),
	}
}

// percentile calcula percentil de série ordenada
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}

	index := int(float64(len(sorted)-1) * p)
	return sorted[index]
}
