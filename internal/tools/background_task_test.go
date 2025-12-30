package tools

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestBackgroundTaskManager_Name(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-btm-*")
	defer os.RemoveAll(tmpDir)

	btm := NewBackgroundTaskManager(tmpDir)
	if btm.Name() != "background_task" {
		t.Errorf("Expected name 'background_task', got '%s'", btm.Name())
	}
}

func TestBackgroundTaskManager_Description(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-btm-*")
	defer os.RemoveAll(tmpDir)

	btm := NewBackgroundTaskManager(tmpDir)
	desc := btm.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
	if !strings.Contains(desc, "assíncronas") {
		t.Error("Description should mention 'assíncronas'")
	}
}

func TestBackgroundTaskManager_Schema(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-btm-*")
	defer os.RemoveAll(tmpDir)

	btm := NewBackgroundTaskManager(tmpDir)
	schema := btm.Schema()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema properties should be a map")
	}

	if _, exists := props["action"]; !exists {
		t.Error("Schema should have 'action' property")
	}

	if _, exists := props["task"]; !exists {
		t.Error("Schema should have 'task' property")
	}

	if _, exists := props["task_id"]; !exists {
		t.Error("Schema should have 'task_id' property")
	}
}

func TestBackgroundTaskManager_Execute_InvalidAction(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-btm-*")
	defer os.RemoveAll(tmpDir)

	btm := NewBackgroundTaskManager(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"action": "invalid_action",
	}

	result, _ := btm.Execute(ctx, params)

	if result.Success {
		t.Error("Result should not be successful for invalid action")
	}

	if !strings.Contains(result.Error, "desconhecida") {
		t.Error("Error should mention unknown action")
	}
}

func TestBackgroundTaskManager_Execute_StartTask_MissingName(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-btm-*")
	defer os.RemoveAll(tmpDir)

	btm := NewBackgroundTaskManager(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"action": "start",
		// Missing task name
	}

	result, _ := btm.Execute(ctx, params)

	if result.Success {
		t.Error("Result should not be successful when task name is missing")
	}

	if !strings.Contains(result.Error, "não especificado") {
		t.Error("Error should mention task name not specified")
	}
}

func TestBackgroundTaskManager_Execute_StartTask_Success(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-btm-*")
	defer os.RemoveAll(tmpDir)

	btm := NewBackgroundTaskManager(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"action": "start",
		"task":   "long_test",
	}

	result, _ := btm.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	if !strings.Contains(result.Message, "Tarefa iniciada") {
		t.Error("Message should confirm task started")
	}

	// Verify task was added to the list
	listResult, _ := btm.Execute(ctx, map[string]interface{}{"action": "list"})
	if !listResult.Success {
		t.Error("List should be successful")
	}

	if !strings.Contains(listResult.Message, "long_test") {
		t.Error("List should contain the started task")
	}
}

func TestBackgroundTaskManager_Execute_ListEmpty(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-btm-*")
	defer os.RemoveAll(tmpDir)

	btm := NewBackgroundTaskManager(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"action": "list",
	}

	result, _ := btm.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	if !strings.Contains(result.Message, "Nenhuma tarefa") {
		t.Error("Should report no tasks when empty")
	}
}

func TestBackgroundTaskManager_Execute_ListWithTasks(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-btm-*")
	defer os.RemoveAll(tmpDir)

	btm := NewBackgroundTaskManager(tmpDir)
	ctx := context.Background()

	// Start a task
	_, _ = btm.Execute(ctx, map[string]interface{}{
		"action": "start",
		"task":   "analysis",
	})

	// List tasks
	result, _ := btm.Execute(ctx, map[string]interface{}{"action": "list"})

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	if !strings.Contains(result.Message, "analysis") {
		t.Error("List should contain the started task")
	}
}

func TestBackgroundTaskManager_Execute_Status_NonExistentTask(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-btm-*")
	defer os.RemoveAll(tmpDir)

	btm := NewBackgroundTaskManager(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"action":  "status",
		"task_id": "nonexistent_task_id",
	}

	result, _ := btm.Execute(ctx, params)

	if result.Success {
		t.Error("Result should not be successful for non-existent task")
	}

	if !strings.Contains(result.Error, "não encontrada") {
		t.Error("Error should mention task not found")
	}
}

func TestBackgroundTaskManager_Execute_Status_ExistingTask(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-btm-*")
	defer os.RemoveAll(tmpDir)

	btm := NewBackgroundTaskManager(tmpDir)
	ctx := context.Background()

	// Start a task
	startResult, _ := btm.Execute(ctx, map[string]interface{}{
		"action": "start",
		"task":   "analysis",
	})

	if !startResult.Success {
		t.Fatal("Failed to start task")
	}

	// Extract task ID from the list
	_, _ = btm.Execute(ctx, map[string]interface{}{"action": "list"})

	// Get the actual task ID from the internal map
	btm.mu.RLock()
	var taskID string
	for id := range btm.tasks {
		taskID = id
		break
	}
	btm.mu.RUnlock()

	if taskID == "" {
		t.Fatal("No task ID found")
	}

	// Get status
	statusResult, _ := btm.Execute(ctx, map[string]interface{}{
		"action":  "status",
		"task_id": taskID,
	})

	if !statusResult.Success {
		t.Errorf("Status should be successful, got error: %s", statusResult.Error)
	}

	if !strings.Contains(statusResult.Message, "Status da Tarefa") {
		t.Error("Status message should contain task status")
	}
}

func TestBackgroundTaskManager_Execute_Cancel_NonExistentTask(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-btm-*")
	defer os.RemoveAll(tmpDir)

	btm := NewBackgroundTaskManager(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"action":  "cancel",
		"task_id": "nonexistent_task_id",
	}

	result, _ := btm.Execute(ctx, params)

	if result.Success {
		t.Error("Result should not be successful for non-existent task")
	}

	if !strings.Contains(result.Error, "não encontrada") {
		t.Error("Error should mention task not found")
	}
}

func TestBackgroundTaskManager_Execute_Cancel_RunningTask(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-btm-*")
	defer os.RemoveAll(tmpDir)

	btm := NewBackgroundTaskManager(tmpDir)
	ctx := context.Background()

	// Start a long task
	_, _ = btm.Execute(ctx, map[string]interface{}{
		"action": "start",
		"task":   "long_test",
	})

	// Get task ID
	btm.mu.RLock()
	var taskID string
	for id := range btm.tasks {
		taskID = id
		break
	}
	btm.mu.RUnlock()

	// Wait a bit for task to start running
	time.Sleep(100 * time.Millisecond)

	// Cancel the task
	cancelResult, _ := btm.Execute(ctx, map[string]interface{}{
		"action":  "cancel",
		"task_id": taskID,
	})

	if !cancelResult.Success {
		t.Errorf("Cancel should be successful, got error: %s", cancelResult.Error)
	}

	// Verify task status is failed
	btm.mu.RLock()
	task := btm.tasks[taskID]
	btm.mu.RUnlock()

	if task.Status != TaskStatusFailed {
		t.Errorf("Task status should be 'failed', got '%s'", task.Status)
	}

	if !strings.Contains(task.Error, "Cancelado") {
		t.Error("Task error should mention cancellation")
	}
}

func TestBackgroundTaskManager_Execute_Result_NonExistentTask(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-btm-*")
	defer os.RemoveAll(tmpDir)

	btm := NewBackgroundTaskManager(tmpDir)
	ctx := context.Background()

	params := map[string]interface{}{
		"action":  "result",
		"task_id": "nonexistent_task_id",
	}

	result, _ := btm.Execute(ctx, params)

	if result.Success {
		t.Error("Result should not be successful for non-existent task")
	}

	if !strings.Contains(result.Error, "não encontrada") {
		t.Error("Error should mention task not found")
	}
}

func TestBackgroundTaskManager_Execute_Result_IncompleteTask(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-btm-*")
	defer os.RemoveAll(tmpDir)

	btm := NewBackgroundTaskManager(tmpDir)
	ctx := context.Background()

	// Start a task
	_, _ = btm.Execute(ctx, map[string]interface{}{
		"action": "start",
		"task":   "long_test",
	})

	// Get task ID
	btm.mu.RLock()
	var taskID string
	for id := range btm.tasks {
		taskID = id
		break
	}
	btm.mu.RUnlock()

	// Try to get result immediately (task not completed yet)
	resultResponse, _ := btm.Execute(ctx, map[string]interface{}{
		"action":  "result",
		"task_id": taskID,
	})

	if resultResponse.Success {
		t.Error("Result should not be successful for incomplete task")
	}

	if !strings.Contains(resultResponse.Error, "não concluída") {
		t.Error("Error should mention task not completed")
	}
}

func TestBackgroundTaskManager_Execute_DefaultAction(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-btm-*")
	defer os.RemoveAll(tmpDir)

	btm := NewBackgroundTaskManager(tmpDir)
	ctx := context.Background()

	// No action specified - should default to "list"
	params := map[string]interface{}{}

	result, _ := btm.Execute(ctx, params)

	if !result.Success {
		t.Errorf("Result should be successful, got error: %s", result.Error)
	}

	// Should show empty list or tasks
	if !strings.Contains(result.Message, "Tarefas") && !strings.Contains(result.Message, "Nenhuma") {
		t.Error("Default action should list tasks")
	}
}

func TestBackgroundTaskManager_TaskStatuses(t *testing.T) {
	// Verify task status constants
	if TaskStatusPending != "pending" {
		t.Error("TaskStatusPending should be 'pending'")
	}
	if TaskStatusRunning != "running" {
		t.Error("TaskStatusRunning should be 'running'")
	}
	if TaskStatusCompleted != "completed" {
		t.Error("TaskStatusCompleted should be 'completed'")
	}
	if TaskStatusFailed != "failed" {
		t.Error("TaskStatusFailed should be 'failed'")
	}
}

func TestBackgroundTaskManager_TaskProgress(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-btm-*")
	defer os.RemoveAll(tmpDir)

	btm := NewBackgroundTaskManager(tmpDir)
	ctx := context.Background()

	// Start a task
	_, _ = btm.Execute(ctx, map[string]interface{}{
		"action": "start",
		"task":   "analysis",
	})

	// Get task ID
	btm.mu.RLock()
	var taskID string
	for id := range btm.tasks {
		taskID = id
		break
	}
	btm.mu.RUnlock()

	// Wait a bit for progress
	time.Sleep(500 * time.Millisecond)

	// Check progress
	btm.mu.RLock()
	task := btm.tasks[taskID]
	btm.mu.RUnlock()

	if task.Progress < 0 || task.Progress > 100 {
		t.Errorf("Task progress should be between 0 and 100, got %.1f", task.Progress)
	}
}

func TestBackgroundTaskManager_DifferentTaskTypes(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-btm-*")
	defer os.RemoveAll(tmpDir)

	btm := NewBackgroundTaskManager(tmpDir)
	ctx := context.Background()

	taskTypes := []string{"long_test", "build", "deploy", "analysis"}

	for _, taskType := range taskTypes {
		result, _ := btm.Execute(ctx, map[string]interface{}{
			"action": "start",
			"task":   taskType,
		})

		if !result.Success {
			t.Errorf("Starting task '%s' should succeed, got error: %s", taskType, result.Error)
		}

		if !strings.Contains(result.Message, taskType) {
			t.Errorf("Result should mention task type '%s'", taskType)
		}
	}

	// Wait a bit to ensure all tasks are registered
	time.Sleep(100 * time.Millisecond)

	// Verify all tasks are in the list
	listResult, _ := btm.Execute(ctx, map[string]interface{}{"action": "list"})

	for _, taskType := range taskTypes {
		if !strings.Contains(listResult.Message, taskType) {
			t.Errorf("List should contain task type '%s'", taskType)
		}
	}
}

func TestBackgroundTaskManager_UnknownTaskType(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test-btm-*")
	defer os.RemoveAll(tmpDir)

	btm := NewBackgroundTaskManager(tmpDir)
	ctx := context.Background()

	// Start unknown task type
	_, _ = btm.Execute(ctx, map[string]interface{}{
		"action": "start",
		"task":   "unknown_task_type",
	})

	// Get task ID
	btm.mu.RLock()
	var taskID string
	for id := range btm.tasks {
		taskID = id
		break
	}
	btm.mu.RUnlock()

	// Wait for task to fail
	time.Sleep(100 * time.Millisecond)

	// Check task status
	btm.mu.RLock()
	task := btm.tasks[taskID]
	btm.mu.RUnlock()

	if task.Status != TaskStatusFailed {
		t.Errorf("Unknown task should fail, got status: %s", task.Status)
	}

	if !strings.Contains(task.Error, "desconhecida") {
		t.Error("Error should mention unknown task")
	}
}
