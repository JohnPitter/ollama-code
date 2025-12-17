package commands

import (
	"context"
	"fmt"
	"sync"
)

// Command interface para comandos
type Command interface {
	Name() string
	Description() string
	Usage() string
	Execute(ctx context.Context, args []string) (string, error)
}

// Registry registro de comandos
type Registry struct {
	commands map[string]Command
	mu       sync.RWMutex
}

// NewRegistry cria novo registro
func NewRegistry() *Registry {
	r := &Registry{
		commands: make(map[string]Command),
	}

	// Registrar comandos built-in
	r.registerBuiltins()

	return r
}

// Register registra um comando
func (r *Registry) Register(cmd Command) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := cmd.Name()
	if _, exists := r.commands[name]; exists {
		return fmt.Errorf("command %s already registered", name)
	}

	r.commands[name] = cmd
	return nil
}

// Get obtém comando por nome
func (r *Registry) Get(name string) (Command, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cmd, exists := r.commands[name]
	if !exists {
		return nil, fmt.Errorf("command %s not found", name)
	}

	return cmd, nil
}

// List lista todos os comandos
func (r *Registry) List() []Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	commands := make([]Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		commands = append(commands, cmd)
	}

	return commands
}

// Execute executa um comando
func (r *Registry) Execute(ctx context.Context, name string, args []string) (string, error) {
	cmd, err := r.Get(name)
	if err != nil {
		return "", err
	}

	return cmd.Execute(ctx, args)
}

// IsCommand verifica se string é comando
func (r *Registry) IsCommand(input string) bool {
	if len(input) == 0 || input[0] != '/' {
		return false
	}

	// Extrair nome do comando
	parts := splitCommand(input)
	if len(parts) == 0 {
		return false
	}

	name := parts[0][1:] // Remove '/'
	_, err := r.Get(name)
	return err == nil
}

// ParseAndExecute faz parse e executa comando
func (r *Registry) ParseAndExecute(ctx context.Context, input string) (string, error) {
	if !r.IsCommand(input) {
		return "", fmt.Errorf("not a command")
	}

	parts := splitCommand(input)
	name := parts[0][1:] // Remove '/'
	args := parts[1:]

	return r.Execute(ctx, name, args)
}

// splitCommand divide comando em partes
func splitCommand(input string) []string {
	// Implementação simples - na produção usar parser melhor
	parts := []string{}
	current := ""
	inQuotes := false

	for _, c := range input {
		switch c {
		case ' ':
			if inQuotes {
				current += string(c)
			} else if current != "" {
				parts = append(parts, current)
				current = ""
			}
		case '"':
			inQuotes = !inQuotes
		default:
			current += string(c)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

// registerBuiltins registra comandos built-in
func (r *Registry) registerBuiltins() {
	r.Register(&HelpCommand{registry: r})
	r.Register(&ClearCommand{})
	r.Register(&HistoryCommand{})
	r.Register(&StatusCommand{})
	r.Register(&ModeCommand{})
}
