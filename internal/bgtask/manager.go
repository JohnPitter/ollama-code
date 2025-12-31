package bgtask

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Manager gerencia tasks em background
type Manager struct {
	tasks map[string]*Task
	mu    sync.RWMutex

	// Statistics
	totalStarted   int
	totalCompleted int
	totalFailed    int
	totalKilled    int
}

// NewManager cria novo manager
func NewManager() *Manager {
	return &Manager{
		tasks: make(map[string]*Task),
	}
}

// Start inicia uma nova task em background
func (m *Manager) Start(command string, args []string, workDir string) (*Task, error) {
	// Gerar ID único
	taskID := uuid.New().String()

	// Criar task
	task := NewTask(taskID, command, args, workDir)

	// Registrar task
	m.mu.Lock()
	m.tasks[taskID] = task
	m.totalStarted++
	m.mu.Unlock()

	// Iniciar em goroutine
	go m.execute(task)

	return task, nil
}

// execute executa a task em background
func (m *Manager) execute(task *Task) {
	defer task.CloseDone()
	defer m.onTaskComplete(task)

	// Criar comando
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, task.Command, task.Args...)

	if task.WorkDir != "" {
		cmd.Dir = task.WorkDir
	}

	// Configurar pipes para stdout e stderr
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		task.Status = StatusFailed
		task.Error = fmt.Errorf("create stdout pipe: %w", err)
		return
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		task.Status = StatusFailed
		task.Error = fmt.Errorf("create stderr pipe: %w", err)
		return
	}

	// Iniciar comando
	if err := cmd.Start(); err != nil {
		task.Status = StatusFailed
		task.Error = fmt.Errorf("start command: %w", err)
		return
	}

	// Ler stdout em goroutine
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		buf := make([]byte, 4096)
		for {
			n, err := stdoutPipe.Read(buf)
			if n > 0 {
				task.WriteStdout(buf[:n])
			}
			if err != nil {
				break
			}
		}
	}()

	// Ler stderr em goroutine
	go func() {
		defer wg.Done()
		buf := make([]byte, 4096)
		for {
			n, err := stderrPipe.Read(buf)
			if n > 0 {
				task.WriteStderr(buf[:n])
			}
			if err != nil {
				break
			}
		}
	}()

	// Aguardar conclusão
	err = cmd.Wait()
	wg.Wait()

	task.CompletedAt = time.Now()

	if err != nil {
		// Verificar se foi exit error (comando falhou)
		if exitErr, ok := err.(*exec.ExitError); ok {
			task.Status = StatusFailed
			task.ExitCode = exitErr.ExitCode()
			task.Error = fmt.Errorf("command exited with code %d", exitErr.ExitCode())
		} else {
			task.Status = StatusFailed
			task.Error = err
		}
	} else {
		task.Status = StatusCompleted
		task.ExitCode = 0
	}
}

// onTaskComplete chamado quando task completa
func (m *Manager) onTaskComplete(task *Task) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch task.Status {
	case StatusCompleted:
		m.totalCompleted++
	case StatusFailed:
		m.totalFailed++
	case StatusKilled:
		m.totalKilled++
	}
}

// Get retorna task pelo ID
func (m *Manager) Get(taskID string) (*Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	return task, nil
}

// List retorna todas as tasks
func (m *Manager) List() []*Task {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tasks := make([]*Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// ListByStatus retorna tasks por status
func (m *Manager) ListByStatus(status TaskStatus) []*Task {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tasks := make([]*Task, 0)
	for _, task := range m.tasks {
		if task.Status == status {
			tasks = append(tasks, task)
		}
	}

	return tasks
}

// Kill mata uma task em execução
func (m *Manager) Kill(taskID string) error {
	task, err := m.Get(taskID)
	if err != nil {
		return err
	}

	// Verificar se já terminou
	if task.Status.IsTerminal() {
		return fmt.Errorf("task already terminated with status: %s", task.Status)
	}

	// Marcar como killed
	task.Status = StatusKilled
	task.CompletedAt = time.Now()

	// Fechar done channel
	task.CloseDone()

	return nil
}

// Cleanup remove tasks terminadas antigas
func (m *Manager) Cleanup(olderThan time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	removed := 0
	cutoff := time.Now().Add(-olderThan)

	for id, task := range m.tasks {
		// Só remover tasks terminadas
		if !task.Status.IsTerminal() {
			continue
		}

		// Só remover se completas antes do cutoff
		if task.CompletedAt.Before(cutoff) {
			delete(m.tasks, id)
			removed++
		}
	}

	return removed
}

// ClearAll remove todas as tasks (útil para testes)
func (m *Manager) ClearAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tasks = make(map[string]*Task)
}

// Stats retorna estatísticas do manager
func (m *Manager) Stats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"total_tasks":     len(m.tasks),
		"running_tasks":   len(m.ListByStatus(StatusRunning)),
		"completed_tasks": len(m.ListByStatus(StatusCompleted)),
		"failed_tasks":    len(m.ListByStatus(StatusFailed)),
		"killed_tasks":    len(m.ListByStatus(StatusKilled)),
		"total_started":   m.totalStarted,
		"total_completed": m.totalCompleted,
		"total_failed":    m.totalFailed,
		"total_killed":    m.totalKilled,
		"success_rate":    m.getSuccessRate(),
	}
}

// getSuccessRate calcula taxa de sucesso
func (m *Manager) getSuccessRate() float64 {
	total := m.totalCompleted + m.totalFailed + m.totalKilled
	if total == 0 {
		return 0.0
	}

	return float64(m.totalCompleted) / float64(total) * 100
}

// GetNewOutput retorna output novo desde última leitura
func (m *Manager) GetNewOutput(taskID string) (string, string, error) {
	task, err := m.Get(taskID)
	if err != nil {
		return "", "", err
	}

	stdout, stderr := task.GetNewOutput()
	return stdout, stderr, nil
}

// GetFullOutput retorna output completo
func (m *Manager) GetFullOutput(taskID string) (string, string, error) {
	task, err := m.Get(taskID)
	if err != nil {
		return "", "", err
	}

	stdout, stderr := task.GetOutput()
	return stdout, stderr, nil
}

// Wait aguarda conclusão de uma task
func (m *Manager) Wait(taskID string) error {
	task, err := m.Get(taskID)
	if err != nil {
		return err
	}

	// Aguardar conclusão
	<-task.Done()

	if task.Error != nil {
		return task.Error
	}

	return nil
}

// WaitWithTimeout aguarda conclusão com timeout
func (m *Manager) WaitWithTimeout(taskID string, timeout time.Duration) error {
	task, err := m.Get(taskID)
	if err != nil {
		return err
	}

	// Aguardar com timeout
	select {
	case <-task.Done():
		if task.Error != nil {
			return task.Error
		}
		return nil

	case <-time.After(timeout):
		return fmt.Errorf("wait timeout after %v", timeout)
	}
}
