package todos

import (
	"testing"
)

func TestManager_Add(t *testing.T) {
	m := NewManager()

	id, err := m.Add("Test task", "Testing")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if id == "" {
		t.Error("Expected non-empty ID")
	}

	todo, err := m.Get(id)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if todo.Content != "Test task" {
		t.Errorf("Expected content 'Test task', got '%s'", todo.Content)
	}

	if todo.Status != StatusPending {
		t.Errorf("Expected status pending, got %s", todo.Status)
	}
}

func TestManager_Update(t *testing.T) {
	m := NewManager()

	id, _ := m.Add("Test task", "Testing")

	err := m.Update(id, StatusInProgress)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	todo, _ := m.Get(id)
	if todo.Status != StatusInProgress {
		t.Errorf("Expected status in_progress, got %s", todo.Status)
	}
}

func TestManager_Complete(t *testing.T) {
	m := NewManager()

	id, _ := m.Add("Test task", "Testing")

	err := m.Complete(id)
	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}

	todo, _ := m.Get(id)
	if todo.Status != StatusCompleted {
		t.Errorf("Expected status completed, got %s", todo.Status)
	}
}

func TestManager_List(t *testing.T) {
	m := NewManager()

	m.Add("Task 1", "Testing 1")
	m.Add("Task 2", "Testing 2")
	m.Add("Task 3", "Testing 3")

	todos := m.List()
	if len(todos) != 3 {
		t.Errorf("Expected 3 todos, got %d", len(todos))
	}
}

func TestManager_ListByStatus(t *testing.T) {
	m := NewManager()

	id1, _ := m.Add("Task 1", "Testing 1")
	id2, _ := m.Add("Task 2", "Testing 2")
	m.Add("Task 3", "Testing 3")

	m.Complete(id1)
	m.SetInProgress(id2)

	pending := m.ListByStatus(StatusPending)
	if len(pending) != 1 {
		t.Errorf("Expected 1 pending todo, got %d", len(pending))
	}

	inProgress := m.ListByStatus(StatusInProgress)
	if len(inProgress) != 1 {
		t.Errorf("Expected 1 in_progress todo, got %d", len(inProgress))
	}

	completed := m.ListByStatus(StatusCompleted)
	if len(completed) != 1 {
		t.Errorf("Expected 1 completed todo, got %d", len(completed))
	}
}

func TestManager_Summary(t *testing.T) {
	m := NewManager()

	id1, _ := m.Add("Task 1", "Testing 1")
	id2, _ := m.Add("Task 2", "Testing 2")
	m.Add("Task 3", "Testing 3")
	m.Add("Task 4", "Testing 4")

	m.Complete(id1)
	m.SetInProgress(id2)

	summary := m.Summary()

	if summary[StatusPending] != 2 {
		t.Errorf("Expected 2 pending, got %d", summary[StatusPending])
	}

	if summary[StatusInProgress] != 1 {
		t.Errorf("Expected 1 in_progress, got %d", summary[StatusInProgress])
	}

	if summary[StatusCompleted] != 1 {
		t.Errorf("Expected 1 completed, got %d", summary[StatusCompleted])
	}
}

func TestManager_Clear(t *testing.T) {
	m := NewManager()

	m.Add("Task 1", "Testing 1")
	m.Add("Task 2", "Testing 2")

	err := m.Clear()
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	if m.Count() != 0 {
		t.Errorf("Expected 0 todos after clear, got %d", m.Count())
	}
}

func TestManager_Delete(t *testing.T) {
	m := NewManager()

	id, _ := m.Add("Task 1", "Testing 1")
	m.Add("Task 2", "Testing 2")

	err := m.Delete(id)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if m.Count() != 1 {
		t.Errorf("Expected 1 todo after delete, got %d", m.Count())
	}

	_, err = m.Get(id)
	if err == nil {
		t.Error("Expected error when getting deleted todo")
	}
}

func TestFileStorage(t *testing.T) {
	// Criar storage temporário
	tmpFile := t.TempDir() + "/todos.json"
	storage := NewFileStorage(tmpFile)

	// Criar manager com file storage
	m := NewManagerWithStorage(storage)

	// Adicionar TODOs
	id1, _ := m.Add("Task 1", "Testing 1")
	m.Add("Task 2", "Testing 2")
	m.Complete(id1)

	// Criar novo manager com mesmo storage (simula restart)
	m2 := NewManagerWithStorage(storage)

	// Verificar que TODOs foram carregados
	if m2.Count() != 2 {
		t.Errorf("Expected 2 todos after reload, got %d", m2.Count())
	}

	completed := m2.ListByStatus(StatusCompleted)
	if len(completed) != 1 {
		t.Errorf("Expected 1 completed todo after reload, got %d", len(completed))
	}
}

func TestTodoStatus_IsValid(t *testing.T) {
	tests := []struct {
		status TodoStatus
		valid  bool
	}{
		{StatusPending, true},
		{StatusInProgress, true},
		{StatusCompleted, true},
		{TodoStatus("invalid"), false},
		{TodoStatus(""), false},
	}

	for _, tt := range tests {
		if tt.status.IsValid() != tt.valid {
			t.Errorf("Status %s: expected valid=%v, got %v",
				tt.status, tt.valid, tt.status.IsValid())
		}
	}
}
