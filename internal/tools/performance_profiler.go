package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// PerformanceProfiler analisa performance do código
type PerformanceProfiler struct {
	workDir string
}

// NewPerformanceProfiler cria novo profiler
func NewPerformanceProfiler(workDir string) *PerformanceProfiler {
	return &PerformanceProfiler{
		workDir: workDir,
	}
}

// Name retorna nome da tool
func (p *PerformanceProfiler) Name() string {
	return "performance_profiler"
}

// Description retorna descrição da tool
func (p *PerformanceProfiler) Description() string {
	return "Analisa performance do código (CPU, memória, benchmarks)"
}

// Execute executa profiling
func (p *PerformanceProfiler) Execute(ctx context.Context, params map[string]interface{}) Result {
	profileType, ok := params["type"].(string)
	if !ok {
		profileType = "benchmark"
	}

	switch profileType {
	case "benchmark":
		return p.runBenchmark(params)
	case "cpu":
		return p.profileCPU(params)
	case "memory":
		return p.profileMemory(params)
	case "trace":
		return p.traceExecution(params)
	case "analyze":
		return p.analyzeProfile(params)
	default:
		return Result{
			Success: false,
			Error:   fmt.Sprintf("Tipo de profile desconhecido: %s", profileType),
		}
	}
}

// runBenchmark executa benchmarks
func (p *PerformanceProfiler) runBenchmark(params map[string]interface{}) Result {
	var output strings.Builder
	output.WriteString("⚡ Executando Benchmarks\n\n")

	projectType := p.detectProjectType()

	switch projectType {
	case "go":
		// Run Go benchmarks
		benchPattern, _ := params["pattern"].(string)
		if benchPattern == "" {
			benchPattern = "."
		}

		cmd := exec.Command("go", "test", "-bench="+benchPattern, "-benchmem", "./...")
		cmd.Dir = p.workDir
		result, err := cmd.CombinedOutput()

		if err != nil {
			return Result{
				Success: false,
				Message:  string(result),
				Error:   err.Error(),
			}
		}

		output.WriteString(string(result))

		// Suggest benchstat for comparison
		output.WriteString("\n💡 Para comparar benchmarks, use benchstat:\n")
		output.WriteString("   go install golang.org/x/perf/cmd/benchstat@latest\n")
		output.WriteString("   go test -bench=. -count=5 > old.txt\n")
		output.WriteString("   # (fazer mudanças)\n")
		output.WriteString("   go test -bench=. -count=5 > new.txt\n")
		output.WriteString("   benchstat old.txt new.txt\n")

	case "nodejs":
		// Suggest benchmark.js or similar
		output.WriteString("💡 Para Node.js, use ferramentas como:\n")
		output.WriteString("   - benchmark.js: npm install --save-dev benchmark\n")
		output.WriteString("   - tinybench: npm install --save-dev tinybench\n")
		output.WriteString("   - vitest bench: npx vitest bench\n")

	case "python":
		// Suggest pytest-benchmark
		output.WriteString("💡 Para Python, use pytest-benchmark:\n")
		output.WriteString("   pip install pytest-benchmark\n")
		output.WriteString("   pytest --benchmark-only\n")

	default:
		return Result{
			Success: false,
			Error:   "Tipo de projeto não suportado para benchmarks",
		}
	}

	return Result{
		Success: true,
		Message:  output.String(),
	}
}

// profileCPU executa CPU profiling
func (p *PerformanceProfiler) profileCPU(params map[string]interface{}) Result {
	var output strings.Builder
	output.WriteString("🔥 CPU Profiling\n\n")

	projectType := p.detectProjectType()

	switch projectType {
	case "go":
		output.WriteString("Para CPU profiling em Go:\n\n")
		output.WriteString("1. Durante testes:\n")
		output.WriteString("   go test -cpuprofile=cpu.prof -bench=.\n\n")

		output.WriteString("2. Em aplicação:\n")
		output.WriteString("   import _ \"net/http/pprof\"\n")
		output.WriteString("   go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30\n\n")

		output.WriteString("3. Visualizar profile:\n")
		output.WriteString("   go tool pprof -http=:8080 cpu.prof\n")

		// Check if pprof profiles exist
		if _, err := os.Stat(filepath.Join(p.workDir, "cpu.prof")); err == nil {
			output.WriteString("\n✓ Encontrado cpu.prof - visualize com: go tool pprof -http=:8080 cpu.prof\n")
		}

	case "nodejs":
		output.WriteString("Para CPU profiling em Node.js:\n\n")
		output.WriteString("1. Usar --prof flag:\n")
		output.WriteString("   node --prof app.js\n")
		output.WriteString("   node --prof-process isolate-*.log > processed.txt\n\n")

		output.WriteString("2. Usar clinic.js:\n")
		output.WriteString("   npm install -g clinic\n")
		output.WriteString("   clinic doctor -- node app.js\n")

	case "python":
		output.WriteString("Para CPU profiling em Python:\n\n")
		output.WriteString("1. Usar cProfile:\n")
		output.WriteString("   python -m cProfile -o output.prof script.py\n")
		output.WriteString("   python -m pstats output.prof\n\n")

		output.WriteString("2. Usar py-spy:\n")
		output.WriteString("   pip install py-spy\n")
		output.WriteString("   py-spy record -o profile.svg -- python script.py\n")

	default:
		output.WriteString("⚠️  CPU profiling não configurado para este projeto\n")
	}

	return Result{
		Success: true,
		Message:  output.String(),
	}
}

// profileMemory executa memory profiling
func (p *PerformanceProfiler) profileMemory(params map[string]interface{}) Result {
	var output strings.Builder
	output.WriteString("💾 Memory Profiling\n\n")

	projectType := p.detectProjectType()

	switch projectType {
	case "go":
		output.WriteString("Para Memory profiling em Go:\n\n")
		output.WriteString("1. Durante testes:\n")
		output.WriteString("   go test -memprofile=mem.prof -bench=.\n\n")

		output.WriteString("2. Em aplicação:\n")
		output.WriteString("   go tool pprof http://localhost:6060/debug/pprof/heap\n\n")

		output.WriteString("3. Analisar alocações:\n")
		output.WriteString("   go tool pprof -alloc_space mem.prof\n")
		output.WriteString("   go tool pprof -inuse_space mem.prof\n")

		// Check if memory profiles exist
		if _, err := os.Stat(filepath.Join(p.workDir, "mem.prof")); err == nil {
			output.WriteString("\n✓ Encontrado mem.prof - visualize com: go tool pprof -http=:8080 mem.prof\n")
		}

	case "nodejs":
		output.WriteString("Para Memory profiling em Node.js:\n\n")
		output.WriteString("1. Heap snapshots:\n")
		output.WriteString("   node --inspect app.js\n")
		output.WriteString("   # Use Chrome DevTools > Memory tab\n\n")

		output.WriteString("2. Usar clinic.js:\n")
		output.WriteString("   clinic heapprofiler -- node app.js\n")

	case "python":
		output.WriteString("Para Memory profiling em Python:\n\n")
		output.WriteString("1. Usar memory_profiler:\n")
		output.WriteString("   pip install memory_profiler\n")
		output.WriteString("   python -m memory_profiler script.py\n\n")

		output.WriteString("2. Usar tracemalloc:\n")
		output.WriteString("   import tracemalloc\n")
		output.WriteString("   tracemalloc.start()\n")

	default:
		output.WriteString("⚠️  Memory profiling não configurado para este projeto\n")
	}

	return Result{
		Success: true,
		Message:  output.String(),
	}
}

// traceExecution executa execution tracing
func (p *PerformanceProfiler) traceExecution(params map[string]interface{}) Result {
	var output strings.Builder
	output.WriteString("🔍 Execution Tracing\n\n")

	projectType := p.detectProjectType()

	switch projectType {
	case "go":
		output.WriteString("Para Execution tracing em Go:\n\n")
		output.WriteString("1. Gerar trace:\n")
		output.WriteString("   go test -trace=trace.out\n\n")

		output.WriteString("2. Visualizar trace:\n")
		output.WriteString("   go tool trace trace.out\n")

		// Check if trace exists
		if _, err := os.Stat(filepath.Join(p.workDir, "trace.out")); err == nil {
			output.WriteString("\n✓ Encontrado trace.out - visualize com: go tool trace trace.out\n")
		}

	case "nodejs":
		output.WriteString("Para tracing em Node.js:\n\n")
		output.WriteString("1. Usar --trace-events:\n")
		output.WriteString("   node --trace-events-enabled app.js\n")
		output.WriteString("   # Visualize em chrome://tracing\n")

	default:
		output.WriteString("⚠️  Execution tracing não configurado para este projeto\n")
	}

	return Result{
		Success: true,
		Message:  output.String(),
	}
}

// analyzeProfile analisa profile existente
func (p *PerformanceProfiler) analyzeProfile(params map[string]interface{}) Result {
	var output strings.Builder
	output.WriteString("📊 Análise de Performance\n\n")

	// Check for common profile files
	profiles := []string{
		"cpu.prof", "mem.prof", "trace.out",
		"profile.prof", "heap.prof",
	}

	found := false
	for _, prof := range profiles {
		fullPath := filepath.Join(p.workDir, prof)
		if _, err := os.Stat(fullPath); err == nil {
			output.WriteString(fmt.Sprintf("✓ Encontrado: %s\n", prof))
			found = true

			// Get file info
			info, _ := os.Stat(fullPath)
			output.WriteString(fmt.Sprintf("  Tamanho: %d bytes\n", info.Size()))
			output.WriteString(fmt.Sprintf("  Modificado: %s\n", info.ModTime().Format(time.RFC3339)))
			output.WriteString("\n")
		}
	}

	if !found {
		output.WriteString("⚠️  Nenhum profile encontrado\n")
		output.WriteString("💡 Execute profiling primeiro para gerar dados de análise\n")
	} else {
		output.WriteString("💡 Use 'go tool pprof' para analisar profiles Go\n")
	}

	return Result{
		Success: true,
		Message:  output.String(),
	}
}

// detectProjectType detecta tipo de projeto
func (p *PerformanceProfiler) detectProjectType() string {
	if _, err := os.Stat(filepath.Join(p.workDir, "go.mod")); err == nil {
		return "go"
	}
	if _, err := os.Stat(filepath.Join(p.workDir, "package.json")); err == nil {
		return "nodejs"
	}
	if _, err := os.Stat(filepath.Join(p.workDir, "requirements.txt")); err == nil {
		return "python"
	}
	return "unknown"
}

// Schema retorna schema JSON da tool
func (p *PerformanceProfiler) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"type": map[string]interface{}{
				"type":        "string",
				"description": "Tipo: benchmark, cpu, memory, trace, analyze",
				"enum":        []string{"benchmark", "cpu", "memory", "trace", "analyze"},
			},
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "Padrão de benchmark (para Go)",
			},
		},
	}
}
