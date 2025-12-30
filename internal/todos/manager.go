package todos

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Manager gerencia TODOs
type Manager struct {
	todos   map[string]*Todo
	mu      sync.RWMutex
	storage Storage
}

// NewManager cria novo manager
func NewManager() *Manager {
	return &Manager{
		todos:   make(map[string]*Todo),
		storage: NewMemoryStorage(),
	}
}

// NewManagerWithStorage cria manager com storage customizado
func NewManagerWithStorage(storage Storage) *Manager {
	m := &Manager{
		todos:   make(map[string]*Todo),
		storage: storage,
	}

	// Carregar TODOs do storage
	if todos, err := storage.Load(); err == nil {
		for _, todo := range todos {
			m.todos[todo.ID] = todo
		}
	}

	return m
}

// Add adiciona novo TODO
func (m *Manager) Add(content, activeForm string) (string, error) {
	if content == "" {
		return "", fmt.Errorf("content cannot be empty")
	}
	if activeForm == "" {
		activeForm = content // fallback
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	id := uuid.New().String()
	now := time.Now()

	todo := &Todo{
		ID:         id,
		Content:    content,
		Status:     StatusPending,
		ActiveForm: activeForm,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	m.todos[id] = todo

	// Persistir
	if err := m.save(); err != nil {
		return "", fmt.Errorf("failed to save: %w", err)
	}

	return id, nil
}

// Update atualiza status de TODO
func (m *Manager) Update(id string, status TodoStatus) error {
	if !status.IsValid() {
		return fmt.Errorf("invalid status: %s", status)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	todo, exists := m.todos[id]
	if !exists {
		return fmt.Errorf("todo not found: %s", id)
	}

	todo.Status = status
	todo.UpdatedAt = time.Now()

	// Persistir
	if err := m.save(); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	return nil
}

// Complete marca TODO como completo
func (m *Manager) Complete(id string) error {
	return m.Update(id, StatusCompleted)
}

// SetInProgress marca TODO como in_progress
func (m *Manager) SetInProgress(id string) error {
	return m.Update(id, StatusInProgress)
}

// Get obtém TODO por ID
func (m *Manager) Get(id string) (*Todo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	todo, exists := m.todos[id]
	if !exists {
		return nil, fmt.Errorf("todo not found: %s", id)
	}

	return todo, nil
}

// List lista todos os TODOs
func (m *Manager) List() []*Todo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	todos := make([]*Todo, 0, len(m.todos))
	for _, todo := range m.todos {
		todos = append(todos, todo)
	}

	return todos
}

// ListByStatus lista TODOs por status
func (m *Manager) ListByStatus(status TodoStatus) []*Todo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	todos := make([]*Todo, 0)
	for _, todo := range m.todos {
		if todo.Status == status {
			todos = append(todos, todo)
		}
	}

	return todos
}

// Clear limpa todos os TODOs
func (m *Manager) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.todos = make(map[string]*Todo)

	// Persistir
	if err := m.save(); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	return nil
}

// Delete remove TODO por ID
func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.todos[id]; !exists {
		return fmt.Errorf("todo not found: %s", id)
	}

	delete(m.todos, id)

	// Persistir
	if err := m.save(); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	return nil
}

// Count retorna número de TODOs
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.todos)
}

// CountByStatus retorna número de TODOs por status
func (m *Manager) CountByStatus(status TodoStatus) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, todo := range m.todos {
		if todo.Status == status {
			count++
		}
	}

	return count
}

// Summary retorna sumário de TODOs
func (m *Manager) Summary() map[TodoStatus]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summary := map[TodoStatus]int{
		StatusPending:    0,
		StatusInProgress: 0,
		StatusCompleted:  0,
	}

	for _, todo := range m.todos {
		summary[todo.Status]++
	}

	return summary
}

// save persiste TODOs (sem lock - deve ser chamado dentro de lock)
func (m *Manager) save() error {
	todos := make([]*Todo, 0, len(m.todos))
	for _, todo := range m.todos {
		todos = append(todos, todo)
	}

	return m.storage.Save(todos)
}
