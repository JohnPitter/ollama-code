package bgtask

import (
	"testing"
	"time"
)

// TestTaskStatus_IsTerminal testa verificação de status terminal
func TestTaskStatus_IsTerminal(t *testing.T) {
	testCases := []struct {
		name     string
		status   TaskStatus
		expected bool
	}{
		{"Running is not terminal", StatusRunning, false},
		{"Completed is terminal", StatusCompleted, true},
		{"Failed is terminal", StatusFailed, true},
		{"Killed is terminal", StatusKilled, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.status.IsTerminal()
			if result != tc.expected {
				t.Errorf("Expected IsTerminal()=%v for %s, got %v", tc.expected, tc.status, result)
			}
		})
	}
}

// TestNewTask testa criação de task
func TestNewTask(t *testing.T) {
	task := NewTask("test-id", "echo", []string{"hello"}, "/tmp")

	if task == nil {
		t.Fatal("NewTask returned nil")
	}

	if task.ID != "test-id" {
		t.Errorf("Expected ID='test-id', got '%s'", task.ID)
	}

	if task.Command != "echo" {
		t.Errorf("Expected Command='echo', got '%s'", task.Command)
	}

	if len(task.Args) != 1 || task.Args[0] != "hello" {
		t.Errorf("Expected Args=['hello'], got %v", task.Args)
	}

	if task.WorkDir != "/tmp" {
		t.Errorf("Expected WorkDir='/tmp', got '%s'", task.WorkDir)
	}

	if task.Status != StatusRunning {
		t.Errorf("Expected Status=Running, got %s", task.Status)
	}

	if task.stdout == nil {
		t.Error("stdout buffer should be initialized")
	}

	if task.stderr == nil {
		t.Error("stderr buffer should be initialized")
	}
}

// TestTask_GetOutput testa obtenção de output
func TestTask_GetOutput(t *testing.T) {
	task := NewTask("test-id", "echo", []string{"hello"}, "")

	// Escrever alguns dados
	task.WriteStdout([]byte("output line 1\n"))
	task.WriteStdout([]byte("output line 2\n"))
	task.WriteStderr([]byte("error line 1\n"))

	stdout, stderr := task.GetOutput()

	expectedStdout := "output line 1\noutput line 2\n"
	if stdout != expectedStdout {
		t.Errorf("Expected stdout='%s', got '%s'", expectedStdout, stdout)
	}

	expectedStderr := "error line 1\n"
	if stderr != expectedStderr {
		t.Errorf("Expected stderr='%s', got '%s'", expectedStderr, stderr)
	}
}

// TestTask_GetNewOutput testa obtenção de output novo
func TestTask_GetNewOutput(t *testing.T) {
	task := NewTask("test-id", "echo", []string{}, "")

	// Escrever primeira parte
	task.WriteStdout([]byte("line 1\n"))

	// Ler output novo
	stdout1, _ := task.GetNewOutput()
	if stdout1 != "line 1\n" {
		t.Errorf("Expected 'line 1\\n', got '%s'", stdout1)
	}

	// Escrever segunda parte
	task.WriteStdout([]byte("line 2\n"))

	// Ler output novo novamente - deve retornar apenas line 2
	stdout2, _ := task.GetNewOutput()
	if stdout2 != "line 2\n" {
		t.Errorf("Expected 'line 2\\n', got '%s'", stdout2)
	}

	// Ler novamente - deve retornar vazio
	stdout3, _ := task.GetNewOutput()
	if stdout3 != "" {
		t.Errorf("Expected empty string, got '%s'", stdout3)
	}
}

// TestTask_Duration testa cálculo de duração
func TestTask_Duration(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name     string
		task     *Task
		expected time.Duration
	}{
		{
			name: "Not started",
			task: &Task{
				StartedAt:   time.Time{},
				CompletedAt: time.Time{},
			},
			expected: 0,
		},
		{
			name: "Running",
			task: &Task{
				StartedAt:   now.Add(-1 * time.Second),
				CompletedAt: time.Time{},
			},
			expected: 900 * time.Millisecond, // Tolerance
		},
		{
			name: "Completed",
			task: &Task{
				StartedAt:   now.Add(-2 * time.Second),
				CompletedAt: now,
			},
			expected: 2 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			duration := tc.task.Duration()

			if tc.name == "Not started" {
				if duration != tc.expected {
					t.Errorf("Expected duration=%v, got %v", tc.expected, duration)
				}
			} else if tc.name == "Running" {
				// For running, duration should be at least expected
				if duration < tc.expected {
					t.Errorf("Expected duration >= %v, got %v", tc.expected, duration)
				}
			} else {
				// For completed, should be exact (with tolerance)
				diff := duration - tc.expected
				if diff < 0 {
					diff = -diff
				}
				if diff > 10*time.Millisecond {
					t.Errorf("Expected duration=%v, got %v (diff=%v)", tc.expected, duration, diff)
				}
			}
		})
	}
}

// TestTask_IsSuccess testa verificação de sucesso
func TestTask_IsSuccess(t *testing.T) {
	testCases := []struct {
		name     string
		task     *Task
		expected bool
	}{
		{
			name: "Completed with exit 0",
			task: &Task{
				Status:   StatusCompleted,
				ExitCode: 0,
				Error:    nil,
			},
			expected: true,
		},
		{
			name: "Completed with error",
			task: &Task{
				Status:   StatusCompleted,
				ExitCode: 0,
				Error:    &mockError{},
			},
			expected: false,
		},
		{
			name: "Completed with non-zero exit",
			task: &Task{
				Status:   StatusCompleted,
				ExitCode: 1,
				Error:    nil,
			},
			expected: false,
		},
		{
			name: "Failed",
			task: &Task{
				Status:   StatusFailed,
				ExitCode: 1,
				Error:    nil,
			},
			expected: false,
		},
		{
			name: "Running",
			task: &Task{
				Status:   StatusRunning,
				ExitCode: 0,
				Error:    nil,
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.task.IsSuccess()
			if result != tc.expected {
				t.Errorf("Expected IsSuccess()=%v, got %v", tc.expected, result)
			}
		})
	}
}

// mockError é um erro mock para testes
type mockError struct{}

func (e *mockError) Error() string {
	return "mock error"
}

// TestTask_WriteStdout testa escrita em stdout
func TestTask_WriteStdout(t *testing.T) {
	task := NewTask("test-id", "cmd", []string{}, "")

	data1 := []byte("hello ")
	data2 := []byte("world\n")

	task.WriteStdout(data1)
	task.WriteStdout(data2)

	stdout, _ := task.GetOutput()
	expected := "hello world\n"

	if stdout != expected {
		t.Errorf("Expected stdout='%s', got '%s'", expected, stdout)
	}
}

// TestTask_WriteStderr testa escrita em stderr
func TestTask_WriteStderr(t *testing.T) {
	task := NewTask("test-id", "cmd", []string{}, "")

	data1 := []byte("error ")
	data2 := []byte("message\n")

	task.WriteStderr(data1)
	task.WriteStderr(data2)

	_, stderr := task.GetOutput()
	expected := "error message\n"

	if stderr != expected {
		t.Errorf("Expected stderr='%s', got '%s'", expected, stderr)
	}
}

// TestTask_Done testa canal done
func TestTask_Done(t *testing.T) {
	task := NewTask("test-id", "cmd", []string{}, "")

	// Canal não deve estar fechado
	select {
	case <-task.Done():
		t.Error("Done channel should not be closed yet")
	default:
		// OK
	}

	// Fechar canal
	task.CloseDone()

	// Agora deve estar fechado
	select {
	case <-task.Done():
		// OK
	default:
		t.Error("Done channel should be closed")
	}

	// Fechar novamente não deve causar panic
	task.CloseDone()
}

// TestTaskStatus_Constants testa constantes de status
func TestTaskStatus_Constants(t *testing.T) {
	if string(StatusRunning) != "running" {
		t.Errorf("StatusRunning should be 'running', got '%s'", StatusRunning)
	}

	if string(StatusCompleted) != "completed" {
		t.Errorf("StatusCompleted should be 'completed', got '%s'", StatusCompleted)
	}

	if string(StatusFailed) != "failed" {
		t.Errorf("StatusFailed should be 'failed', got '%s'", StatusFailed)
	}

	if string(StatusKilled) != "killed" {
		t.Errorf("StatusKilled should be 'killed', got '%s'", StatusKilled)
	}
}
