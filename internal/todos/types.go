package todos

import "time"

// TodoStatus status do TODO
type TodoStatus string

const (
	StatusPending    TodoStatus = "pending"
	StatusInProgress TodoStatus = "in_progress"
	StatusCompleted  TodoStatus = "completed"
)

// Todo representa uma tarefa
type Todo struct {
	ID         string     `json:"id"`
	Content    string     `json:"content"`
	Status     TodoStatus `json:"status"`
	ActiveForm string     `json:"active_form"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// IsValid valida se status é válido
func (s TodoStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusInProgress, StatusCompleted:
		return true
	default:
		return false
	}
}

// String retorna string representation
func (s TodoStatus) String() string {
	return string(s)
}
