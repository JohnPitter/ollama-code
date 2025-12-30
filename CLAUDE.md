# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Ollama Code is a 100% local AI coding assistant that runs entirely on the user's computer using Ollama models. It features web search, code analysis, file operations, and specialized skills for research and API testing.

## Development Principles

### Core Principles (Prompt Mestre)

Follow these 12 principles when developing features or making changes:

1. **Clean Architecture** - Maintain separation of concerns, use Handler Pattern, avoid God objects, follow SOLID principles
2. **Performance Based on Big O Notation** - Analyze algorithm complexity, optimize critical paths, avoid O(n²) when O(n log n) is possible
3. **Mitigated Against Major CVEs** - Check for common vulnerabilities (injection, XSS, path traversal), validate all inputs, sanitize outputs
4. **Service Resilience and Cache Usage** - Implement proper error handling, use the cache manager for expensive operations, design for failure recovery
5. **Modern Context-Based Design** - Use context.Context for cancellation and timeouts, pass context through the call chain
6. **Functionality Guaranteed Through Test Pyramid** - Unit tests (majority), integration tests (moderate), E2E tests (few but critical)
7. **Security** - Validate file paths, sanitize user input, never execute arbitrary code without confirmation, check permissions
8. **Observability** - Use structured logging, record metrics, create spans for distributed tracing, enable debugging in production
9. **Design System Principles** - Consistent error messages, uniform output formatting, predictable user experience
10. **Create a Plan and Build in Phases/Subphases** - Break down complex features, implement incrementally, validate each phase
11. **Document Changes** - Update relevant documentation files, add entries to `changes/` directory for significant features
12. **Functional Build with CHANGELOG.md** - Ensure `make build` succeeds after changes, document all changes in CHANGELOG.md

### Agent Behavior Guidelines

When acting as an AI agent working on this codebase:

1. **Long-Running Commands** - If a command takes too long to execute, cancel it or run it as a subprocess. Don't block the main thread.
2. **Try Alternative Approaches** - If a solution doesn't work, research alternatives on the internet. Don't repeat the same failing approach.
3. **Token Economy** - Focus on implementation over summaries. Be concise in explanations, verbose in code quality.

## Build, Test, and Development Commands

### Building
```bash
# Development build (fast)
make dev

# Optimized build
make build

# Build for all platforms
make build-all

# Release build (maximum optimization)
make release
```

### Testing
```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tools tests only
make test-tools

# Run all checks (lint + vet + test)
make check
```

### Code Quality
```bash
# Run linter
make lint

# Run go vet
make vet

# Format code
make fmt
```

### Running
```bash
# Run directly without building
make run

# Or use go run
go run ./cmd/ollama-code chat
```

### Installation
```bash
# Install globally (requires sudo on Unix)
make install

# Install to ~/bin (no sudo)
make install-local
```

## Performance and Troubleshooting

### GPU Overload and CPU Fallback

**Problem:** Ollama consuming too much GPU, causing system slowdowns or freezes.

**Understanding:** Ollama Code is a **client** - it doesn't control GPU/CPU directly. GPU/CPU usage is managed by the **Ollama server**.

#### Solution 1: Force Ollama to Use CPU

Stop Ollama and restart with CPU-only mode:

```bash
# Linux/Mac
CUDA_VISIBLE_DEVICES="" ollama serve

# Windows (PowerShell)
$env:CUDA_VISIBLE_DEVICES=""; ollama serve

# Windows (CMD)
set CUDA_VISIBLE_DEVICES= && ollama serve
```

#### Solution 2: Use Lighter Models

Switch to smaller, faster models when GPU is limited:

```bash
# Instead of qwen2.5-coder:7b (7B parameters)
ollama pull qwen2.5-coder:1.5b  # Much faster, less accurate

# Or use the default small model
ollama pull qwen2.5-coder:0.5b  # Fastest, basic tasks only
```

Configure in `~/.ollama/config.json`:
```json
{
  "model": "qwen2.5-coder:1.5b"
}
```

#### Solution 3: Limit Ollama GPU Memory

Configure Ollama to use less GPU memory:

```bash
# Limit to 4GB GPU memory
OLLAMA_MAX_LOADED_MODELS=1 OLLAMA_NUM_PARALLEL=1 ollama serve
```

### Performance Monitoring

Ollama Code includes an **Observability System** that tracks performance:

- **Handler durations:** Time spent in each intent handler
- **Tool durations:** Time spent in each tool execution
- **LLM call durations:** Time spent waiting for Ollama responses
- **Cache hit rate:** Effectiveness of caching

Enable observability in code:
```go
cfg := &di.Config{
    EnableObservability: true,
    ObservabilityConfig: observability.LoggerConfig{
        Level:  observability.LogLevelInfo,
        Format: observability.LogFormatJSON,
    },
}
```

View metrics summary:
```bash
# Metrics are logged to stderr during execution
# Look for "Metrics Summary" at the end of execution
```

### Common Performance Issues

#### Issue: Slow LLM Responses (>30s)

**Causes:**
1. GPU overloaded with other processes
2. Model too large for available VRAM
3. Ollama server not responding

**Solutions:**
1. Check GPU usage: `nvidia-smi` (Linux) or Task Manager (Windows)
2. Kill other GPU processes
3. Switch to CPU mode (see above)
4. Use smaller model

#### Issue: Timeouts or Hangs

**Causes:**
1. Ollama server crashed or not running
2. Network issues (if using remote Ollama)
3. Out of memory (GPU or system RAM)

**Solutions:**
1. Restart Ollama: `ollama serve`
2. Check Ollama status: `ollama list`
3. Check logs: `journalctl -u ollama` (Linux) or Event Viewer (Windows)
4. Increase timeout in code (default: 2 minutes)

#### Issue: High Memory Usage

**Causes:**
1. Large context (many files in session)
2. Multiple models loaded in Ollama
3. Cache not being cleared

**Solutions:**
1. Limit context size - avoid loading entire large files
2. Unload unused models: `ollama rm <model>`
3. Restart Ollama to clear cache

### Benchmarking

Typical performance targets:

| Operation | Expected Time | Action if Slower |
|-----------|--------------|------------------|
| Simple file read | < 1s | Check disk I/O |
| Simple file write | < 2s | Check disk I/O |
| Code search | < 3s | Check index size |
| LLM intent detection | 1-5s | Use smaller model |
| LLM code generation | 5-30s | Use smaller model or CPU |
| Web search | 3-10s | Check internet connection |

## Architecture Overview

This codebase recently underwent a major refactoring implementing **Handler Pattern**, **Manual Dependency Injection**, and **Observability**. Understanding this architecture is critical.

### Core Architecture Patterns

#### 1. Handler Pattern (Phase 1 - Completed)
The old `handlers.go` God object (2282 lines) was refactored into individual handler files using the Handler Pattern:

**Location:** `internal/handlers/`

**Key files:**
- `handler.go` - Handler interface and Dependencies struct
- `registry.go` - HandlerRegistry (thread-safe, routes intents to handlers)
- Individual handlers: `file_read_handler.go`, `file_write_handler.go`, `search_handler.go`, `execute_handler.go`, `question_handler.go`, `git_handler.go`, `analyze_handler.go`, `websearch_handler.go`
- `adapters.go` - Adapters to bridge real implementations with interfaces

**Handler Interface:**
```go
type Handler interface {
    Handle(ctx context.Context, deps *Dependencies, result *intent.DetectionResult) (string, error)
}
```

**Dependencies Struct:**
Handlers receive a `Dependencies` struct instead of the entire Agent. This struct contains only what handlers need through **interfaces**, enabling loose coupling and testability.

**HandlerRegistry:**
Thread-safe registry that routes intents to appropriate handlers. Supports dynamic registration and default handlers.

#### 2. Manual Dependency Injection (Phase 2 - Completed)
Located in `internal/di/`, this provides organized dependency creation without external frameworks.

**Why Manual DI:**
- Google Wire was **archived in 2024** and is not recommended for new projects
- Manual DI is idiomatic in Go
- No external dependencies
- Easy to debug and maintain

**Key files:**
- `config.go` - Config struct
- `providers.go` - 25 provider functions for components
- `agent.go` - InitializeAgent function that wires everything together

**Usage:**
```go
cfg := &di.Config{
    OllamaURL: "http://localhost:11434",
    Model:     "qwen2.5-coder:7b",
}
agent, err := di.InitializeAgent(cfg)
```

Providers can be used individually in tests for mocking.

#### 3. Observability System (Phase 3 - Completed)
Located in `internal/observability/`, provides structured logging, metrics, and distributed tracing.

**Components:**
- `logger.go` - Structured logger using Go's `log/slog`
- `metrics.go` - In-memory metrics collector (P50/P95/P99, error rates, cache hit rate)
- `tracing.go` - Distributed tracing with spans
- `middleware.go` - Wrappers for handlers, tools, and LLM calls

**Key features:**
- Handler execution metrics
- Tool execution metrics
- LLM request latency tracking
- Cache hit/miss tracking
- Error rate monitoring
- Hierarchical trace visualization

**Enable observability:**
```go
cfg := &di.Config{
    EnableObservability: true,
    ObservabilityConfig: observability.LoggerConfig{
        Level:  observability.LogLevelInfo,
        Format: observability.LogFormatJSON,
    },
}
```

### Package Structure

**Core packages:**
- `internal/agent/` - Main orchestration logic
- `internal/handlers/` - Intent handlers (refactored from God object)
- `internal/validators/` - Shared validation logic (filename, JSON, code cleaning)
- `internal/di/` - Manual dependency injection providers
- `internal/observability/` - Logging, metrics, and tracing

**Supporting packages:**
- `internal/llm/` - Ollama client
- `internal/intent/` - Intent detection
- `internal/tools/` - Tool registry and implementations (15 tools)
- `internal/skills/` - Specialized skills (research, API, code analysis)
- `internal/websearch/` - Web search orchestrator and HTML fetcher
- `internal/ollamamd/` - Hierarchical OLLAMA.md context system

**Other packages:**
- `internal/config/` - Configuration management
- `internal/modes/` - Operation modes (readonly, interactive, autonomous)
- `internal/session/` - Session management
- `internal/cache/` - Caching with TTL
- `internal/confirmation/` - User confirmation prompts
- `internal/statusline/` - Rich status line display
- `internal/commands/` - Built-in slash commands

### Agent Flow

```
User Input → Agent.ProcessMessage()
           → Intent.Detect()
           → Agent.handleIntent()
           → HandlerRegistry.Handle()
           → Specific Handler executes
           → Response to User
```

The Agent builds a `Dependencies` struct and delegates to the HandlerRegistry, which routes to the appropriate handler based on detected intent.

### OLLAMA.md System

Hierarchical configuration system with 4 levels (enterprise, project, language, local). Files at different levels are merged, with more specific levels overriding general ones.

**Levels:**
1. Enterprise: `~/.ollama/OLLAMA.md`
2. Project: `/project/OLLAMA.md`
3. Language: `/project/.ollama/go/OLLAMA.md`
4. Local: `/project/subdir/OLLAMA.md`

## Important Development Practices

### Adding New Handlers

When adding a new handler:
1. Create a new file in `internal/handlers/` (e.g., `myfeature_handler.go`)
2. Implement the `Handler` interface
3. Add tests in `myfeature_handler_test.go`
4. Create a provider in `internal/di/providers.go`
5. Register it in `ProvideHandlerRegistry()` in `internal/di/providers.go`
6. Update the `Agent` struct in `internal/agent/agent.go` if needed

### Adding New Tools

Tools live in `internal/tools/`. When adding a tool:
1. Implement the tool in a new file
2. Register it in `ProvideToolRegistry()` in `internal/di/providers.go`
3. Add comprehensive tests
4. Document in `internal/tools/README.md`

### Code Organization Conventions

**Import order:**
```go
import (
    // Standard library
    "context"
    "fmt"

    // External dependencies
    "github.com/fatih/color"

    // Internal packages
    "github.com/johnpitter/ollama-code/internal/agent"
)
```

**Naming:**
- Packages: lowercase, single word (`websearch`, `ollamamd`)
- Exported types: PascalCase (`Agent`, `HandlerRegistry`)
- Unexported functions: camelCase (`buildDependencies`, `processMessage`)
- Interfaces: often -er suffix (`Handler`, `Manager`)
- Errors: Err prefix (`ErrNotFound`)

### Testing Philosophy

- Use table-driven tests where appropriate
- Test files live alongside code (`*_test.go`)
- Target >80% coverage for new code
- Use the DI providers in tests to create components in isolation
- Mock dependencies through interfaces in `Dependencies` struct

### Commit Message Format

Follow Conventional Commits:
- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation only
- `refactor:` - Code refactoring
- `test:` - Adding/fixing tests
- `chore:` - Build, dependencies, tooling

## Critical Context for Making Changes

### The Refactoring History

This project went through 3 major refactoring phases:
1. **Handler Pattern** - Eliminated 2282-line God object
2. **Manual DI** - Organized dependency creation (rejected Wire as it's archived)
3. **Observability** - Added comprehensive monitoring

**Current test count:** 210 tests passing

When making changes, maintain this architecture. Do not reintroduce tight coupling or God objects.

### Validators Package

Common validation logic was extracted to `internal/validators/`:
- `filename.go` - Filename validation and extraction
- `json.go` - JSON extraction and parsing
- `code.go` - Code cleaning and language detection

Use these validators instead of duplicating validation logic.

### Agent Fields are Public

Agent struct fields are **public** (PascalCase) to support DI. This is intentional. The public API (methods) remains unchanged.

### Backward Compatibility

Both `agent.NewAgent()` and `di.InitializeAgent()` work. The former is the traditional approach, the latter uses organized providers. Both are valid.

## Recent Quality Assurance

The project recently completed extensive QA testing achieving **100% test coverage** (44/44 tests passing) across high, medium, and low priority scenarios. See `docs/QA_100_PERCENT_COVERAGE_2024-12-22.md` for details.

## Advanced Tools Implementation

The project includes 7 advanced tools (100% coverage):
- Advanced Refactoring (AST-based)
- Background Task Manager
- Code Formatter
- Dependency Manager
- Documentation Generator
- Performance Profiler
- Security Scanner
- Test Runner
- Git Helper

These tools have comprehensive tests and are documented in `docs/ADVANCED_TOOLS_USAGE.md`.

## Web Search Capabilities

The web search system uses a hybrid approach:
- DuckDuckGo for search queries
- HTML fetching for real content
- LLM summarization of results

See `changes/01-web-search-hybrid.md` for implementation details.

## When Working on This Codebase

1. **Understand the architecture** - Handler Pattern + Manual DI + Observability
2. **Use the DI providers** - Don't create dependencies manually
3. **Follow the established patterns** - Don't reinvent wheels
4. **Add observability** - Use logger, metrics, and tracing in new code
5. **Write tests** - Use table-driven tests and DI providers for mocks
6. **Run checks before committing** - `make check` runs lint + vet + test
7. **Consult architecture docs** - See `docs/architecture/` for detailed explanations

## Key Documentation Files

- `docs/architecture/ARCHITECTURE_REFACTORING.md` - Handler Pattern details
- `docs/architecture/MANUAL_DI.md` - DI implementation details
- `docs/architecture/OBSERVABILITY.md` - Observability system guide
- `internal/di/README.md` - DI package documentation
- `internal/tools/README.md` - Tools documentation
- `docs/guides/CONTRIBUTING.md` - Contribution guidelines

## Dependencies

Minimal external dependencies:
- `github.com/fatih/color` - Terminal colors
- `github.com/spf13/cobra` - CLI framework

Go version: 1.21+
