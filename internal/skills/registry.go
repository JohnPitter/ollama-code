package skills

import (
	"context"
	"fmt"
	"sync"
)

// Registry gerencia registro e descoberta de skills
type Registry struct {
	skills map[string]Skill
	mu     sync.RWMutex
}

// NewRegistry cria novo registry de skills
func NewRegistry() *Registry {
	return &Registry{
		skills: make(map[string]Skill),
	}
}

// Register registra um novo skill
func (r *Registry) Register(skill Skill) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := skill.Name()
	if _, exists := r.skills[name]; exists {
		return fmt.Errorf("skill %s já está registrado", name)
	}

	r.skills[name] = skill
	return nil
}

// Get obtém um skill pelo nome
func (r *Registry) Get(name string) (Skill, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skill, exists := r.skills[name]
	if !exists {
		return nil, fmt.Errorf("skill %s não encontrado", name)
	}

	return skill, nil
}

// List retorna todos os skills registrados
func (r *Registry) List() []Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skills := make([]Skill, 0, len(r.skills))
	for _, skill := range r.skills {
		skills = append(skills, skill)
	}

	return skills
}

// FindCapable encontra skills capazes de processar uma tarefa
func (r *Registry) FindCapable(ctx context.Context, task Task) []Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	capable := []Skill{}
	for _, skill := range r.skills {
		if skill.CanHandle(ctx, task) {
			capable = append(capable, skill)
		}
	}

	return capable
}

// Execute executa um skill específico
func (r *Registry) Execute(ctx context.Context, skillName string, task Task) (*Result, error) {
	skill, err := r.Get(skillName)
	if err != nil {
		return nil, err
	}

	return skill.Execute(ctx, task)
}

// ExecuteAny executa o primeiro skill capaz de processar a tarefa
func (r *Registry) ExecuteAny(ctx context.Context, task Task) (*Result, error) {
	capable := r.FindCapable(ctx, task)
	if len(capable) == 0 {
		return nil, fmt.Errorf("nenhum skill capaz de processar tarefa do tipo %s", task.Type)
	}

	// Usar o primeiro skill capaz
	return capable[0].Execute(ctx, task)
}

// GetCapabilities retorna todas as capabilities disponíveis
func (r *Registry) GetCapabilities() map[string][]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	capabilities := make(map[string][]string)
	for name, skill := range r.skills {
		capabilities[name] = skill.Capabilities()
	}

	return capabilities
}

// Count retorna número de skills registrados
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.skills)
}
