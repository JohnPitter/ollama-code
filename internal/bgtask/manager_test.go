package bgtask

import (
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestNewManager testa criação do manager
func TestNewManager(t *testing.T) {
	mgr := NewManager()

	if mgr == nil {
		t.Fatal("NewManager returned nil")
	}

	if mgr.tasks == nil {
		t.Error("tasks map should be initialized")
	}

	if len(mgr.tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(mgr.tasks))
	}
}

// TestManager_Start testa iniciar task
func TestManager_Start(t *testing.T) {
	mgr := NewManager()

	var cmd, arg string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		arg = "/c"
	} else {
		cmd = "sh"
		arg = "-c"
	}

	task, err := mgr.Start(cmd, []string{arg, "echo hello"}, "")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if task == nil {
		t.Fatal("Start returned nil task")
	}

	if task.ID == "" {
		t.Error("Task ID should not be empty")
	}

	// Aguardar um pouco para task completar
	time.Sleep(100 * time.Millisecond)
}

// TestManager_Get testa obter task
func TestManager_Get(t *testing.T) {
	mgr := NewManager()

	var cmd, arg string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		arg = "/c"
	} else {
		cmd = "sh"
		arg = "-c"
	}

	task, _ := mgr.Start(cmd, []string{arg, "echo test"}, "")

	retrieved, err := mgr.Get(task.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.ID != task.ID {
		t.Errorf("Expected ID %s, got %s", task.ID, retrieved.ID)
	}
}

// TestManager_Get_NotFound testa get de task inexistente
func TestManager_Get_NotFound(t *testing.T) {
	mgr := NewManager()

	_, err := mgr.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent task")
	}
}

// TestManager_List testa listar tasks
func TestManager_List(t *testing.T) {
	mgr := NewManager()

	var cmd, arg string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		arg = "/c"
	} else {
		cmd = "sh"
		arg = "-c"
	}

	// Criar 3 tasks
	for i := 0; i < 3; i++ {
		mgr.Start(cmd, []string{arg, "echo test"}, "")
	}

	tasks := mgr.List()
	if len(tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(tasks))
	}
}

// TestManager_ListByStatus testa filtrar por status
func TestManager_ListByStatus(t *testing.T) {
	mgr := NewManager()

	var cmd, arg string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		arg = "/c"
	} else {
		cmd = "sh"
		arg = "-c"
	}

	// Criar e aguardar 2 tasks completarem
	for i := 0; i < 2; i++ {
		task, _ := mgr.Start(cmd, []string{arg, "echo done"}, "")
		mgr.Wait(task.ID)
	}

	completed := mgr.ListByStatus(StatusCompleted)
	if len(completed) != 2 {
		t.Errorf("Expected 2 completed tasks, got %d", len(completed))
	}
}

// TestManager_Wait testa aguardar conclusão
func TestManager_Wait(t *testing.T) {
	mgr := NewManager()

	var cmd, arg string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		arg = "/c"
	} else {
		cmd = "sh"
		arg = "-c"
	}

	task, _ := mgr.Start(cmd, []string{arg, "echo waiting"}, "")

	err := mgr.Wait(task.ID)
	if err != nil {
		t.Errorf("Wait failed: %v", err)
	}

	// Task deve estar completa
	finalTask, _ := mgr.Get(task.ID)
	if !finalTask.Status.IsTerminal() {
		t.Errorf("Expected terminal status, got %s", finalTask.Status)
	}
}

// TestManager_WaitWithTimeout testa wait com timeout
func TestManager_WaitWithTimeout(t *testing.T) {
	mgr := NewManager()

	var cmd, arg string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		arg = "/c"
	} else {
		cmd = "sh"
		arg = "-c"
	}

	task, _ := mgr.Start(cmd, []string{arg, "echo quick"}, "")

	err := mgr.WaitWithTimeout(task.ID, 1*time.Second)
	if err != nil {
		t.Errorf("WaitWithTimeout failed: %v", err)
	}
}

// TestManager_GetFullOutput testa obter output completo
func TestManager_GetFullOutput(t *testing.T) {
	mgr := NewManager()

	var cmd, arg string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		arg = "/c"
	} else {
		cmd = "sh"
		arg = "-c"
	}

	task, _ := mgr.Start(cmd, []string{arg, "echo hello world"}, "")
	mgr.Wait(task.ID)

	// Small delay to ensure output buffers are flushed
	time.Sleep(50 * time.Millisecond)

	stdout, _, err := mgr.GetFullOutput(task.ID)
	if err != nil {
		t.Fatalf("GetFullOutput failed: %v", err)
	}

	if !strings.Contains(stdout, "hello world") {
		t.Errorf("Expected 'hello world' in output, got: %s", stdout)
	}
}

// TestManager_GetNewOutput testa obter output novo
func TestManager_GetNewOutput(t *testing.T) {
	mgr := NewManager()

	var cmd, arg string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		arg = "/c"
	} else {
		cmd = "sh"
		arg = "-c"
	}

	task, _ := mgr.Start(cmd, []string{arg, "echo line1 && echo line2"}, "")
	mgr.Wait(task.ID)

	// Small delay to ensure output buffers are flushed
	time.Sleep(50 * time.Millisecond)

	// Primeira leitura - deve retornar todo output
	stdout1, _, err := mgr.GetNewOutput(task.ID)
	if err != nil {
		t.Fatalf("GetNewOutput failed: %v", err)
	}

	if stdout1 == "" {
		t.Error("Expected some output on first read")
	}

	// Segunda leitura - deve retornar vazio
	stdout2, _, err := mgr.GetNewOutput(task.ID)
	if err != nil {
		t.Fatalf("Second GetNewOutput failed: %v", err)
	}

	if stdout2 != "" {
		t.Errorf("Expected empty output on second read, got: %s", stdout2)
	}
}

// TestManager_Cleanup testa limpeza de tasks antigas
func TestManager_Cleanup(t *testing.T) {
	mgr := NewManager()

	var cmd, arg string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		arg = "/c"
	} else {
		cmd = "sh"
		arg = "-c"
	}

	// Criar e completar 3 tasks
	for i := 0; i < 3; i++ {
		task, _ := mgr.Start(cmd, []string{arg, "echo done"}, "")
		mgr.Wait(task.ID)
	}

	// Aguardar um pouco
	time.Sleep(100 * time.Millisecond)

	// Cleanup tasks completadas há mais de 50ms
	removed := mgr.Cleanup(50 * time.Millisecond)
	if removed != 3 {
		t.Errorf("Expected to remove 3 tasks, removed %d", removed)
	}

	// Verificar que foram removidas
	tasks := mgr.List()
	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks after cleanup, got %d", len(tasks))
	}
}

// TestManager_ClearAll testa limpar todas as tasks
func TestManager_ClearAll(t *testing.T) {
	mgr := NewManager()

	var cmd, arg string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		arg = "/c"
	} else {
		cmd = "sh"
		arg = "-c"
	}

	// Criar algumas tasks
	for i := 0; i < 5; i++ {
		mgr.Start(cmd, []string{arg, "echo test"}, "")
	}

	mgr.ClearAll()

	if len(mgr.tasks) != 0 {
		t.Errorf("Expected 0 tasks after ClearAll, got %d", len(mgr.tasks))
	}
}

// TestManager_Stats testa estatísticas
func TestManager_Stats(t *testing.T) {
	mgr := NewManager()

	var cmd, arg string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		arg = "/c"
	} else {
		cmd = "sh"
		arg = "-c"
	}

	// Criar e completar 2 tasks com sucesso
	for i := 0; i < 2; i++ {
		task, _ := mgr.Start(cmd, []string{arg, "echo success"}, "")
		mgr.Wait(task.ID)
	}

	stats := mgr.Stats()

	if stats["total_started"].(int) != 2 {
		t.Errorf("Expected total_started=2, got %v", stats["total_started"])
	}

	if stats["total_completed"].(int) != 2 {
		t.Errorf("Expected total_completed=2, got %v", stats["total_completed"])
	}

	successRate := stats["success_rate"].(float64)
	if successRate != 100.0 {
		t.Errorf("Expected success_rate=100.0, got %v", successRate)
	}
}

// TestManager_TaskWithWorkDir testa task com working directory
func TestManager_TaskWithWorkDir(t *testing.T) {
	mgr := NewManager()

	var cmd, arg string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		arg = "/c"
	} else {
		cmd = "sh"
		arg = "-c"
	}

	// Usar diretório temporário do sistema
	workDir := t.TempDir()

	task, err := mgr.Start(cmd, []string{arg, "echo test"}, workDir)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if task.WorkDir != workDir {
		t.Errorf("Expected WorkDir=%s, got %s", workDir, task.WorkDir)
	}

	mgr.Wait(task.ID)
}

// TestManager_FailingCommand testa comando que falha
func TestManager_FailingCommand(t *testing.T) {
	mgr := NewManager()

	var cmd, arg string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		arg = "/c"
	} else {
		cmd = "sh"
		arg = "-c"
	}

	// Comando que falha (exit 1)
	task, _ := mgr.Start(cmd, []string{arg, "exit 1"}, "")
	err := mgr.Wait(task.ID)

	if err == nil {
		t.Error("Expected error for failing command")
	}

	finalTask, _ := mgr.Get(task.ID)
	if finalTask.Status != StatusFailed {
		t.Errorf("Expected status Failed, got %s", finalTask.Status)
	}

	if finalTask.ExitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", finalTask.ExitCode)
	}
}

// TestManager_Concurrent testa execução concorrente
func TestManager_Concurrent(t *testing.T) {
	mgr := NewManager()

	var cmd, arg string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		arg = "/c"
	} else {
		cmd = "sh"
		arg = "-c"
	}

	// Iniciar 5 tasks em paralelo
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			task, err := mgr.Start(cmd, []string{arg, "echo concurrent"}, "")
			if err != nil {
				t.Errorf("Concurrent start failed: %v", err)
			}
			mgr.Wait(task.ID)
			done <- true
		}()
	}

	// Aguardar todos completarem
	for i := 0; i < 5; i++ {
		<-done
	}

	stats := mgr.Stats()
	if stats["total_started"].(int) != 5 {
		t.Errorf("Expected 5 started tasks, got %v", stats["total_started"])
	}
}
