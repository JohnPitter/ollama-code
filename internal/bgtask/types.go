package bgtask

import (
	"bytes"
	"sync"
	"time"
)

// TaskStatus representa o status de uma task em background
type TaskStatus string

const (
	StatusRunning   TaskStatus = "running"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
	StatusKilled    TaskStatus = "killed"
)

// IsTerminal verifica se o status é terminal (não vai mudar mais)
func (s TaskStatus) IsTerminal() bool {
	switch s {
	case StatusCompleted, StatusFailed, StatusKilled:
		return true
	default:
		return false
	}
}

// Task representa uma tarefa em background
type Task struct {
	ID          string
	Command     string
	Args        []string
	WorkDir     string
	Status      TaskStatus
	ExitCode    int
	Error       error
	StartedAt   time.Time
	CompletedAt time.Time

	// Output streaming
	stdout      *bytes.Buffer
	stderr      *bytes.Buffer
	stdoutRead  int // Bytes já lidos do stdout
	stderrRead  int // Bytes já lidos do stderr
	outputMu    sync.RWMutex

	// Process control
	done chan struct{}
}

// NewTask cria nova task
func NewTask(id, command string, args []string, workDir string) *Task {
	return &Task{
		ID:         id,
		Command:    command,
		Args:       args,
		WorkDir:    workDir,
		Status:     StatusRunning,
		StartedAt:  time.Now(),
		stdout:     &bytes.Buffer{},
		stderr:     &bytes.Buffer{},
		stdoutRead: 0,
		stderrRead: 0,
		done:       make(chan struct{}),
	}
}

// GetOutput retorna output completo (stdout + stderr)
func (t *Task) GetOutput() (string, string) {
	t.outputMu.RLock()
	defer t.outputMu.RUnlock()

	return t.stdout.String(), t.stderr.String()
}

// GetNewOutput retorna apenas output novo desde última leitura
func (t *Task) GetNewOutput() (string, string) {
	t.outputMu.Lock()
	defer t.outputMu.Unlock()

	// Ler bytes novos do stdout
	stdoutBytes := t.stdout.Bytes()
	newStdout := string(stdoutBytes[t.stdoutRead:])
	t.stdoutRead = len(stdoutBytes)

	// Ler bytes novos do stderr
	stderrBytes := t.stderr.Bytes()
	newStderr := string(stderrBytes[t.stderrRead:])
	t.stderrRead = len(stderrBytes)

	return newStdout, newStderr
}

// WriteStdout escreve no stdout buffer
func (t *Task) WriteStdout(data []byte) {
	t.outputMu.Lock()
	defer t.outputMu.Unlock()
	t.stdout.Write(data)
}

// WriteStderr escreve no stderr buffer
func (t *Task) WriteStderr(data []byte) {
	t.outputMu.Lock()
	defer t.outputMu.Unlock()
	t.stderr.Write(data)
}

// Duration retorna duração da execução
func (t *Task) Duration() time.Duration {
	if t.StartedAt.IsZero() {
		return 0
	}

	if t.CompletedAt.IsZero() {
		return time.Since(t.StartedAt)
	}

	return t.CompletedAt.Sub(t.StartedAt)
}

// IsSuccess verifica se completou com sucesso
func (t *Task) IsSuccess() bool {
	return t.Status == StatusCompleted && t.ExitCode == 0 && t.Error == nil
}

// Done retorna canal que é fechado quando task termina
func (t *Task) Done() <-chan struct{} {
	return t.done
}

// CloseDone fecha o canal done
func (t *Task) CloseDone() {
	select {
	case <-t.done:
		// Already closed
	default:
		close(t.done)
	}
}
