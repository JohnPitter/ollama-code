package observability

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"
)

// LogLevel representa o nível de log
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// LogFormat representa o formato de saída
type LogFormat string

const (
	LogFormatJSON LogFormat = "json"
	LogFormatText LogFormat = "text"
)

// LoggerConfig configuração do logger
type LoggerConfig struct {
	Level      LogLevel
	Format     LogFormat
	Output     io.Writer
	AddSource  bool // Adiciona file:line nos logs
	TimeFormat string
}

// Logger wrapper para slog com funcionalidades extras
type Logger struct {
	slog   *slog.Logger
	config LoggerConfig
}

// NewLogger cria novo logger estruturado
func NewLogger(config LoggerConfig) *Logger {
	// Defaults
	if config.Output == nil {
		config.Output = os.Stdout
	}
	if config.TimeFormat == "" {
		config.TimeFormat = time.RFC3339
	}

	// Converter level
	var level slog.Level
	switch config.Level {
	case LogLevelDebug:
		level = slog.LevelDebug
	case LogLevelInfo:
		level = slog.LevelInfo
	case LogLevelWarn:
		level = slog.LevelWarn
	case LogLevelError:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Opções do slog
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: config.AddSource,
	}

	// Criar handler baseado no formato
	var handler slog.Handler
	if config.Format == LogFormatJSON {
		handler = slog.NewJSONHandler(config.Output, opts)
	} else {
		handler = slog.NewTextHandler(config.Output, opts)
	}

	return &Logger{
		slog:   slog.New(handler),
		config: config,
	}
}

// NewDefaultLogger cria logger com configuração padrão
func NewDefaultLogger() *Logger {
	return NewLogger(LoggerConfig{
		Level:     LogLevelInfo,
		Format:    LogFormatText,
		Output:    os.Stdout,
		AddSource: false,
	})
}

// Debug log em nível debug
func (l *Logger) Debug(msg string, args ...any) {
	l.slog.Debug(msg, args...)
}

// Info log em nível info
func (l *Logger) Info(msg string, args ...any) {
	l.slog.Info(msg, args...)
}

// Warn log em nível warn
func (l *Logger) Warn(msg string, args ...any) {
	l.slog.Warn(msg, args...)
}

// Error log em nível error
func (l *Logger) Error(msg string, args ...any) {
	l.slog.Error(msg, args...)
}

// With cria logger com campos adicionais
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		slog:   l.slog.With(args...),
		config: l.config,
	}
}

// WithContext cria logger com contexto
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// Extrair trace_id se existir
	if traceID := ctx.Value(traceIDKey); traceID != nil {
		return l.With("trace_id", traceID)
	}
	return l
}

// WithComponent cria logger para um componente específico
func (l *Logger) WithComponent(component string) *Logger {
	return l.With("component", component)
}

// LogHandlerStart log início de handler
func (l *Logger) LogHandlerStart(ctx context.Context, handlerName string, intent string) {
	l.WithContext(ctx).Info("handler_start",
		"handler", handlerName,
		"intent", intent,
	)
}

// LogHandlerEnd log fim de handler
func (l *Logger) LogHandlerEnd(ctx context.Context, handlerName string, duration time.Duration, err error) {
	logger := l.WithContext(ctx).With(
		"handler", handlerName,
		"duration_ms", duration.Milliseconds(),
	)

	if err != nil {
		logger.Error("handler_error", "error", err.Error())
	} else {
		logger.Info("handler_complete")
	}
}

// LogToolExecution log execução de tool
func (l *Logger) LogToolExecution(ctx context.Context, toolName string, duration time.Duration, success bool) {
	logger := l.WithContext(ctx).With(
		"tool", toolName,
		"duration_ms", duration.Milliseconds(),
		"success", success,
	)
	logger.Debug("tool_execution")
}

// LogLLMRequest log requisição ao LLM
func (l *Logger) LogLLMRequest(ctx context.Context, model string, tokens int, duration time.Duration) {
	l.WithContext(ctx).Info("llm_request",
		"model", model,
		"tokens", tokens,
		"duration_ms", duration.Milliseconds(),
	)
}

// LogIntentDetection log detecção de intenção
func (l *Logger) LogIntentDetection(ctx context.Context, intent string, confidence float64, duration time.Duration) {
	l.WithContext(ctx).Info("intent_detection",
		"intent", intent,
		"confidence", confidence,
		"duration_ms", duration.Milliseconds(),
	)
}

// LogCacheHit log hit de cache
func (l *Logger) LogCacheHit(ctx context.Context, key string, hit bool) {
	l.WithContext(ctx).Debug("cache_access",
		"key", key,
		"hit", hit,
	)
}

// Contexto keys
type contextKey string

const traceIDKey contextKey = "trace_id"
