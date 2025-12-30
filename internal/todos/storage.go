package todos

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Storage interface para persistência de TODOs
type Storage interface {
	Save(todos []*Todo) error
	Load() ([]*Todo, error)
}

// MemoryStorage storage em memória (não persiste)
type MemoryStorage struct {
	data []*Todo
}

// NewMemoryStorage cria memory storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make([]*Todo, 0),
	}
}

// Save salva TODOs em memória
func (s *MemoryStorage) Save(todos []*Todo) error {
	s.data = todos
	return nil
}

// Load carrega TODOs da memória
func (s *MemoryStorage) Load() ([]*Todo, error) {
	return s.data, nil
}

// FileStorage storage em arquivo JSON
type FileStorage struct {
	filePath string
}

// NewFileStorage cria file storage
func NewFileStorage(filePath string) *FileStorage {
	return &FileStorage{
		filePath: filePath,
	}
}

// Save salva TODOs em arquivo JSON
func (s *FileStorage) Save(todos []*Todo) error {
	// Criar diretório se não existir
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Serializar para JSON
	data, err := json.MarshalIndent(todos, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal todos: %w", err)
	}

	// Escrever arquivo
	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Load carrega TODOs do arquivo JSON
func (s *FileStorage) Load() ([]*Todo, error) {
	// Ler arquivo
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Arquivo não existe, retornar lista vazia
			return make([]*Todo, 0), nil
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Deserializar JSON
	var todos []*Todo
	if err := json.Unmarshal(data, &todos); err != nil {
		return nil, fmt.Errorf("failed to unmarshal todos: %w", err)
	}

	return todos, nil
}

// DefaultFileStorage retorna storage padrão em ~/.ollama-code/todos.json
func DefaultFileStorage() (*FileStorage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	filePath := filepath.Join(homeDir, ".ollama-code", "todos.json")
	return NewFileStorage(filePath), nil
}
