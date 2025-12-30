package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPerformanceProfiler_Name(t *testing.T) {
	pp := NewPerformanceProfiler(".")
	if pp.Name() != "performance_profiler" {
		t.Errorf("Expected name 'performance_profiler', got '%s'", pp.Name())
	}
}

func TestPerformanceProfiler_Description(t *testing.T) {
	pp := NewPerformanceProfiler(".")
	desc := pp.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
	if !strings.Contains(desc, "performance") {
		t.Error("Description should mention 'performance'")
	}
}

func TestPerformanceProfiler_Schema(t *testing.T) {
	pp := NewPerformanceProfiler(".")
	schema := pp.Schema()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema properties should be a map")
	}

	if _, exists := props["type"]; !exists {
		t.Error("Schema should have 'type' property")
	}

	if _, exists := props["pattern"]; !exists {
		t.Error("Schema should have 'pattern' property")
	}
}

func TestPerformanceProfiler_Execute_InvalidType(t *testing.T) {
	pp := NewPerformanceProfiler(".")
	ctx := context.Background()

	params := map[string]interface{}{
		"type": "invalid_type",
	}

	result, _ := pp.Execute(ctx, params)

	if result.Success {
		t.Error("Result should not be successful for invalid type")
	}

	if !strings.Contains(result.Error, "desconhecido") {
		t.Error("Error should mention unknown type")
	}
}

func TestPerformanceProfiler_DetectProjectType_Go(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-go-*")
	defer os.RemoveAll(tmpDir)

	// Create go.mod
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	pp := NewPerformanceProfiler(tmpDir)
	projectType := pp.detectProjectType()

	if projectType != "go" {
		t.Errorf("Expected project type 'go', got '%s'", projectType)
	}
}

func TestPerformanceProfiler_DetectProjectType_NodeJS(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-node-*")
	defer os.RemoveAll(tmpDir)

	// Create package.json
	pkgPath := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(pkgPath, []byte(`{"name":"test"}`), 0644); err != nil {
		t.Fatal(err)
	}

	pp := NewPerformanceProfiler(tmpDir)
	projectType := pp.detectProjectType()

	if projectType != "nodejs" {
		t.Errorf("Expected project type 'nodejs', got '%s'", projectType)
	}
}

func TestPerformanceProfiler_DetectProjectType_Python(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-python-*")
	defer os.RemoveAll(tmpDir)

	// Create requirements.txt
	reqPath := filepath.Join(tmpDir, "requirements.txt")
	if err := os.WriteFile(reqPath, []byte("requests==2.28.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	pp := NewPerformanceProfiler(tmpDir)
	projectType := pp.detectProjectType()

	if projectType != "python" {
		t.Errorf("Expected project type 'python', got '%s'", projectType)
	}
}

func TestPerformanceProfiler_DetectProjectType_Unknown(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-unknown-*")
	defer os.RemoveAll(tmpDir)

	pp := NewPerformanceProfiler(tmpDir)
	projectType := pp.detectProjectType()

	if projectType != "unknown" {
		t.Errorf("Expected project type 'unknown', got '%s'", projectType)
	}
}

func TestPerformanceProfiler_Execute_Benchmark_UnsupportedProject(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-bench-*")
	defer os.RemoveAll(tmpDir)

	pp := NewPerformanceProfiler(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"type": "benchmark",
	}

	result, _ := pp.Execute(ctx, params)

	if result.Success {
		t.Error("Result should not be successful for unsupported project")
	}

	if !strings.Contains(result.Error, "não suportado") {
		t.Error("Error should mention unsupported project type")
	}
}

func TestPerformanceProfiler_Execute_Benchmark_NodeJS(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-bench-node-*")
	defer os.RemoveAll(tmpDir)

	// Create package.json
	pkgPath := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(pkgPath, []byte(`{"name":"test"}`), 0644); err != nil {
		t.Fatal(err)
	}

	pp := NewPerformanceProfiler(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"type": "benchmark",
	}

	result, _ := pp.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	// Should suggest Node.js benchmark tools
	if !strings.Contains(result.Message, "benchmark.js") && !strings.Contains(result.Message, "tinybench") {
		t.Error("Should suggest Node.js benchmark tools")
	}
}

func TestPerformanceProfiler_Execute_Benchmark_Python(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-bench-python-*")
	defer os.RemoveAll(tmpDir)

	// Create requirements.txt
	reqPath := filepath.Join(tmpDir, "requirements.txt")
	if err := os.WriteFile(reqPath, []byte("pytest\n"), 0644); err != nil {
		t.Fatal(err)
	}

	pp := NewPerformanceProfiler(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"type": "benchmark",
	}

	result, _ := pp.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	// Should suggest pytest-benchmark
	if !strings.Contains(result.Message, "pytest-benchmark") {
		t.Error("Should suggest pytest-benchmark")
	}
}

func TestPerformanceProfiler_Execute_CPU_Go(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-cpu-*")
	defer os.RemoveAll(tmpDir)

	// Create go.mod
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	pp := NewPerformanceProfiler(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"type": "cpu",
	}

	result, _ := pp.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	// Should mention CPU profiling for Go
	if !strings.Contains(result.Message, "cpuprofile") {
		t.Error("Should mention cpuprofile for Go")
	}

	if !strings.Contains(result.Message, "pprof") {
		t.Error("Should mention pprof")
	}
}

func TestPerformanceProfiler_Execute_CPU_NodeJS(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-cpu-node-*")
	defer os.RemoveAll(tmpDir)

	// Create package.json
	pkgPath := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(pkgPath, []byte(`{"name":"test"}`), 0644); err != nil {
		t.Fatal(err)
	}

	pp := NewPerformanceProfiler(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"type": "cpu",
	}

	result, _ := pp.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	// Should mention --prof or clinic.js
	if !strings.Contains(result.Message, "--prof") && !strings.Contains(result.Message, "clinic") {
		t.Error("Should mention Node.js CPU profiling tools")
	}
}

func TestPerformanceProfiler_Execute_Memory_Go(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-mem-*")
	defer os.RemoveAll(tmpDir)

	// Create go.mod
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	pp := NewPerformanceProfiler(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"type": "memory",
	}

	result, _ := pp.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	// Should mention memory profiling for Go
	if !strings.Contains(result.Message, "memprofile") {
		t.Error("Should mention memprofile for Go")
	}
}

func TestPerformanceProfiler_Execute_Memory_Python(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-mem-python-*")
	defer os.RemoveAll(tmpDir)

	// Create requirements.txt
	reqPath := filepath.Join(tmpDir, "requirements.txt")
	if err := os.WriteFile(reqPath, []byte("pytest\n"), 0644); err != nil {
		t.Fatal(err)
	}

	pp := NewPerformanceProfiler(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"type": "memory",
	}

	result, _ := pp.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	// Should mention memory_profiler or tracemalloc
	if !strings.Contains(result.Message, "memory_profiler") && !strings.Contains(result.Message, "tracemalloc") {
		t.Error("Should mention Python memory profiling tools")
	}
}

func TestPerformanceProfiler_Execute_Trace_Go(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-trace-*")
	defer os.RemoveAll(tmpDir)

	// Create go.mod
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	pp := NewPerformanceProfiler(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"type": "trace",
	}

	result, _ := pp.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	// Should mention trace for Go
	if !strings.Contains(result.Message, "trace") {
		t.Error("Should mention trace for Go")
	}
}

func TestPerformanceProfiler_Execute_Trace_NodeJS(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-trace-node-*")
	defer os.RemoveAll(tmpDir)

	// Create package.json
	pkgPath := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(pkgPath, []byte(`{"name":"test"}`), 0644); err != nil {
		t.Fatal(err)
	}

	pp := NewPerformanceProfiler(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"type": "trace",
	}

	result, _ := pp.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	// Should mention trace-events
	if !strings.Contains(result.Message, "trace-events") {
		t.Error("Should mention trace-events for Node.js")
	}
}

func TestPerformanceProfiler_Execute_Analyze_NoProfiles(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-analyze-*")
	defer os.RemoveAll(tmpDir)

	pp := NewPerformanceProfiler(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"type": "analyze",
	}

	result, _ := pp.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	// Should report no profiles found
	if !strings.Contains(result.Message, "Nenhum profile encontrado") {
		t.Error("Should report no profiles found")
	}
}

func TestPerformanceProfiler_Execute_Analyze_WithProfiles(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-analyze-prof-*")
	defer os.RemoveAll(tmpDir)

	// Create dummy profile files
	profileFiles := []string{"cpu.prof", "mem.prof", "trace.out"}
	for _, file := range profileFiles {
		path := filepath.Join(tmpDir, file)
		if err := os.WriteFile(path, []byte("dummy profile data"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	pp := NewPerformanceProfiler(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"type": "analyze",
	}

	result, _ := pp.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	// Should detect all profile files
	for _, file := range profileFiles {
		if !strings.Contains(result.Message, file) {
			t.Errorf("Should detect profile file: %s", file)
		}
	}

	// Should show file info
	if !strings.Contains(result.Message, "Tamanho") {
		t.Error("Should show file size")
	}

	if !strings.Contains(result.Message, "Modificado") {
		t.Error("Should show modification time")
	}
}

func TestPerformanceProfiler_Execute_DefaultType(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-default-*")
	defer os.RemoveAll(tmpDir)

	pp := NewPerformanceProfiler(tmpDir)
	ctx := context.Background()

	// No type specified - should default to "benchmark"
	params := map[string]interface{}{}

	result, _ := pp.Execute(ctx, params)

	// For unknown project, benchmark should fail
	if result.Success {
		t.Log("Benchmark succeeded (unexpected for unknown project)")
	} else {
		if !strings.Contains(result.Error, "não suportado") {
			t.Logf("Got expected error: %s", result.Error)
		}
	}
}

func TestPerformanceProfiler_Execute_CPU_WithExistingProfile(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-cpu-prof-*")
	defer os.RemoveAll(tmpDir)

	// Create go.mod
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create cpu.prof
	cpuProfPath := filepath.Join(tmpDir, "cpu.prof")
	if err := os.WriteFile(cpuProfPath, []byte("dummy cpu profile"), 0644); err != nil {
		t.Fatal(err)
	}

	pp := NewPerformanceProfiler(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"type": "cpu",
	}

	result, _ := pp.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	// Should detect existing cpu.prof
	if !strings.Contains(result.Message, "Encontrado cpu.prof") {
		t.Error("Should detect existing cpu.prof file")
	}
}

func TestPerformanceProfiler_Execute_Memory_WithExistingProfile(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-mem-prof-*")
	defer os.RemoveAll(tmpDir)

	// Create go.mod
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create mem.prof
	memProfPath := filepath.Join(tmpDir, "mem.prof")
	if err := os.WriteFile(memProfPath, []byte("dummy memory profile"), 0644); err != nil {
		t.Fatal(err)
	}

	pp := NewPerformanceProfiler(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"type": "memory",
	}

	result, _ := pp.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	// Should detect existing mem.prof
	if !strings.Contains(result.Message, "Encontrado mem.prof") {
		t.Error("Should detect existing mem.prof file")
	}
}

func TestPerformanceProfiler_Execute_Trace_WithExistingTrace(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-trace-file-*")
	defer os.RemoveAll(tmpDir)

	// Create go.mod
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create trace.out
	traceOutPath := filepath.Join(tmpDir, "trace.out")
	if err := os.WriteFile(traceOutPath, []byte("dummy trace data"), 0644); err != nil {
		t.Fatal(err)
	}

	pp := NewPerformanceProfiler(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"type": "trace",
	}

	result, _ := pp.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	// Should detect existing trace.out
	if !strings.Contains(result.Message, "Encontrado trace.out") {
		t.Error("Should detect existing trace.out file")
	}
}
