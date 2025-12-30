package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TaskStatus representa status de uma tarefa
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

// BackgroundTask representa uma tarefa em background
type BackgroundTask struct {
	ID        string
	Name      string
	Status    TaskStatus
	Progress  float64
	Result    string
	Error     string
	StartTime time.Time
	EndTime   time.Time
}

// BackgroundTaskManager gerencia tarefas em background
type BackgroundTaskManager struct {
	workDir     string
	tasks       map[string]*BackgroundTask
	mu          sync.RWMutex
	taskCounter int64
	storageFile string
}

// NewBackgroundTaskManager cria novo gerenciador de tarefas
func NewBackgroundTaskManager(workDir string) *BackgroundTaskManager {
	storageFile := filepath.Join(workDir, ".ollama-code", "background_tasks.json")

	btm := &BackgroundTaskManager{
		workDir:     workDir,
		tasks:       make(map[string]*BackgroundTask),
		storageFile: storageFile,
	}

	// Load existing tasks from disk
	btm.loadTasks()

	return btm
}

// Name retorna nome da tool
func (b *BackgroundTaskManager) Name() string {
	return "background_task"
}

// Description retorna descrição da tool
func (b *BackgroundTaskManager) Description() string {
	return "Gerencia tarefas assíncronas em background"
}

// RequiresConfirmation indica se requer confirmação
func (b *BackgroundTaskManager) RequiresConfirmation() bool {
	return false
}

// Execute executa operação de background task
func (b *BackgroundTaskManager) Execute(ctx context.Context, params map[string]interface{}) (Result, error) {
	action, ok := params["action"].(string)
	if !ok {
		action = "list"
	}

	switch action {
	case "start":
		taskName, _ := params["task"].(string)
		return b.startTask(taskName, params)
	case "status":
		taskID, _ := params["task_id"].(string)
		return b.getStatus(taskID)
	case "list":
		return b.listTasks()
	case "cancel":
		taskID, _ := params["task_id"].(string)
		return b.cancelTask(taskID)
	case "result":
		taskID, _ := params["task_id"].(string)
		return b.getResult(taskID)
	default:
		return Result{
			Success: false,
			Error:   fmt.Sprintf("Ação desconhecida: %s", action),
		}, nil
	}
}

// startTask inicia nova tarefa em background
func (b *BackgroundTaskManager) startTask(taskName string, params map[string]interface{}) (Result, error) {
	if taskName == "" {
		return Result{
			Success: false,
			Error:   "Nome da tarefa não especificado",
		}, nil
	}

	// Generate task ID with counter to ensure uniqueness
	b.mu.Lock()
	b.taskCounter++
	taskID := fmt.Sprintf("task_%d_%d", time.Now().UnixNano(), b.taskCounter)

	task := &BackgroundTask{
		ID:        taskID,
		Name:      taskName,
		Status:    TaskStatusPending,
		Progress:  0,
		StartTime: time.Now(),
	}

	b.tasks[taskID] = task
	b.mu.Unlock()

	// Auto-save new task
	go b.saveTasks()

	// Start task in goroutine
	go b.executeTask(taskID, taskName, params)

	return Result{
		Success: true,
		Message:  fmt.Sprintf("✅ Tarefa iniciada: %s (ID: %s)\n", taskName, taskID),
	}, nil
}

// executeTask executa tarefa específica
func (b *BackgroundTaskManager) executeTask(taskID, taskName string, params map[string]interface{}) {
	b.updateTaskStatus(taskID, TaskStatusRunning)

	// Simulate different types of tasks
	switch taskName {
	case "long_test":
		b.runLongTest(taskID)
	case "build":
		b.runBuild(taskID)
	case "deploy":
		b.runDeploy(taskID)
	case "analysis":
		b.runAnalysis(taskID)
	default:
		b.updateTaskError(taskID, fmt.Sprintf("Tarefa desconhecida: %s", taskName))
		return
	}
}

// runLongTest simula teste longo
func (b *BackgroundTaskManager) runLongTest(taskID string) {
	steps := 10
	for i := 0; i < steps; i++ {
		time.Sleep(1 * time.Second)
		progress := float64(i+1) / float64(steps) * 100
		b.updateTaskProgress(taskID, progress)
	}

	b.updateTaskComplete(taskID, "Teste longo concluído com sucesso")
}

// runBuild simula build
func (b *BackgroundTaskManager) runBuild(taskID string) {
	phases := []string{"Compilando", "Linkando", "Otimizando", "Empacotando"}

	for i, phase := range phases {
		time.Sleep(2 * time.Second)
		progress := float64(i+1) / float64(len(phases)) * 100
		b.updateTaskProgress(taskID, progress)
		b.updateTaskResult(taskID, fmt.Sprintf("Fase: %s", phase))
	}

	b.updateTaskComplete(taskID, "Build concluído com sucesso")
}

// runDeploy simula deploy
func (b *BackgroundTaskManager) runDeploy(taskID string) {
	phases := []string{"Preparando", "Uploading", "Configurando", "Validando"}

	for i, phase := range phases {
		time.Sleep(3 * time.Second)
		progress := float64(i+1) / float64(len(phases)) * 100
		b.updateTaskProgress(taskID, progress)
		b.updateTaskResult(taskID, fmt.Sprintf("Fase: %s", phase))
	}

	b.updateTaskComplete(taskID, "Deploy concluído com sucesso")
}

// runAnalysis simula análise
func (b *BackgroundTaskManager) runAnalysis(taskID string) {
	phases := []string{"Escaneando", "Analisando", "Gerando relatório"}

	for i, phase := range phases {
		time.Sleep(2 * time.Second)
		progress := float64(i+1) / float64(len(phases)) * 100
		b.updateTaskProgress(taskID, progress)
		b.updateTaskResult(taskID, fmt.Sprintf("Fase: %s", phase))
	}

	b.updateTaskComplete(taskID, "Análise concluída - 0 issues encontrados")
}

// getStatus obtém status de tarefa
func (b *BackgroundTaskManager) getStatus(taskID string) (Result, error) {
	b.mu.RLock()
	task, exists := b.tasks[taskID]
	b.mu.RUnlock()

	if !exists {
		return Result{
			Success: false,
			Error:   fmt.Sprintf("Tarefa não encontrada: %s", taskID),
		}, nil
	}

	output := fmt.Sprintf(`📊 Status da Tarefa

ID: %s
Nome: %s
Status: %s
Progresso: %.1f%%
Iniciado: %s
`, task.ID, task.Name, task.Status, task.Progress, task.StartTime.Format("15:04:05"))

	if task.Status == TaskStatusCompleted || task.Status == TaskStatusFailed {
		output += fmt.Sprintf("Finalizado: %s\n", task.EndTime.Format("15:04:05"))
		duration := task.EndTime.Sub(task.StartTime)
		output += fmt.Sprintf("Duração: %s\n", duration.Round(time.Second))
	}

	if task.Result != "" {
		output += fmt.Sprintf("\nResultado: %s\n", task.Result)
	}

	if task.Error != "" {
		output += fmt.Sprintf("\nErro: %s\n", task.Error)
	}

	return Result{
		Success: true,
		Message:  output,
	}, nil
}

// listTasks lista todas as tarefas
func (b *BackgroundTaskManager) listTasks() (Result, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(b.tasks) == 0 {
		return Result{
			Success: true,
			Message:  "Nenhuma tarefa em execução\n",
		}, nil
	}

	var output string
	output = "📋 Tarefas em Background\n\n"

	for _, task := range b.tasks {
		statusIcon := "⏳"
		switch task.Status {
		case TaskStatusRunning:
			statusIcon = "🔄"
		case TaskStatusCompleted:
			statusIcon = "✅"
		case TaskStatusFailed:
			statusIcon = "❌"
		}

		output += fmt.Sprintf("%s [%s] %s - %.0f%%\n", statusIcon, task.ID[:12], task.Name, task.Progress)
	}

	return Result{
		Success: true,
		Message:  output,
	}, nil
}

// cancelTask cancela tarefa
func (b *BackgroundTaskManager) cancelTask(taskID string) (Result, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	task, exists := b.tasks[taskID]
	if !exists {
		return Result{
			Success: false,
			Error:   fmt.Sprintf("Tarefa não encontrada: %s", taskID),
		}, nil
	}

	if task.Status != TaskStatusRunning {
		return Result{
			Success: false,
			Error:   "Tarefa não está em execução",
		}, nil
	}

	task.Status = TaskStatusFailed
	task.Error = "Cancelado pelo usuário"
	task.EndTime = time.Now()

	// Auto-save after cancel
	go b.saveTasks()

	return Result{
		Success: true,
		Message:  fmt.Sprintf("✅ Tarefa cancelada: %s\n", taskID),
	}, nil
}

// getResult obtém resultado de tarefa
func (b *BackgroundTaskManager) getResult(taskID string) (Result, error) {
	b.mu.RLock()
	task, exists := b.tasks[taskID]
	b.mu.RUnlock()

	if !exists {
		return Result{
			Success: false,
			Error:   fmt.Sprintf("Tarefa não encontrada: %s", taskID),
		}, nil
	}

	if task.Status != TaskStatusCompleted {
		return Result{
			Success: false,
			Error:   fmt.Sprintf("Tarefa ainda não concluída (status: %s)", task.Status),
		}, nil
	}

	return Result{
		Success: true,
		Message:  task.Result,
	}, nil
}

// Helper methods to update task state
func (b *BackgroundTaskManager) updateTaskStatus(taskID string, status TaskStatus) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if task, exists := b.tasks[taskID]; exists {
		task.Status = status
	}
	// Auto-save after update (run in background to avoid blocking)
	go b.saveTasks()
}

func (b *BackgroundTaskManager) updateTaskProgress(taskID string, progress float64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if task, exists := b.tasks[taskID]; exists {
		task.Progress = progress
	}
	// Auto-save after update
	go b.saveTasks()
}

func (b *BackgroundTaskManager) updateTaskResult(taskID string, result string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if task, exists := b.tasks[taskID]; exists {
		task.Result = result
	}
	// Auto-save after update
	go b.saveTasks()
}

func (b *BackgroundTaskManager) updateTaskComplete(taskID string, result string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if task, exists := b.tasks[taskID]; exists {
		task.Status = TaskStatusCompleted
		task.Progress = 100
		task.Result = result
		task.EndTime = time.Now()
	}
	// Auto-save after update
	go b.saveTasks()
}

func (b *BackgroundTaskManager) updateTaskError(taskID string, errMsg string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if task, exists := b.tasks[taskID]; exists {
		task.Status = TaskStatusFailed
		task.Error = errMsg
		task.EndTime = time.Now()
	}
	// Auto-save after update
	go b.saveTasks()
}

// saveTasks persiste tasks em disco
func (b *BackgroundTaskManager) saveTasks() error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Create directory if it doesn't exist
	dir := filepath.Dir(b.storageFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Serialize tasks to JSON
	data, err := json.MarshalIndent(b.tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}

	// Write to file
	if err := os.WriteFile(b.storageFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write tasks file: %w", err)
	}

	return nil
}

// loadTasks carrega tasks do disco
func (b *BackgroundTaskManager) loadTasks() error {
	// Check if file exists
	if _, err := os.Stat(b.storageFile); os.IsNotExist(err) {
		return nil // No tasks to load, not an error
	}

	// Read file
	data, err := os.ReadFile(b.storageFile)
	if err != nil {
		return fmt.Errorf("failed to read tasks file: %w", err)
	}

	// Deserialize tasks
	b.mu.Lock()
	defer b.mu.Unlock()

	if err := json.Unmarshal(data, &b.tasks); err != nil {
		return fmt.Errorf("failed to unmarshal tasks: %w", err)
	}

	// Update task counter to avoid ID conflicts
	for range b.tasks {
		// Extract counter from task ID (format: task_timestamp_counter)
		// This is a simple approach, could be improved
		b.taskCounter++
	}

	return nil
}

// Schema retorna schema JSON da tool
func (b *BackgroundTaskManager) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Ação: start, status, list, cancel, result",
				"enum":        []string{"start", "status", "list", "cancel", "result"},
			},
			"task": map[string]interface{}{
				"type":        "string",
				"description": "Nome da tarefa (para start): long_test, build, deploy, analysis",
			},
			"task_id": map[string]interface{}{
				"type":        "string",
				"description": "ID da tarefa (para status, cancel, result)",
			},
		},
	}
}
