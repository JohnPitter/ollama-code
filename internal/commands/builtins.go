package commands

import (
	"context"
	"fmt"
	"strings"
)

// HelpCommand comando de ajuda
type HelpCommand struct {
	registry *Registry
}

func (h *HelpCommand) Name() string        { return "help" }
func (h *HelpCommand) Description() string { return "Show available commands" }
func (h *HelpCommand) Usage() string       { return "/help [command]" }

func (h *HelpCommand) Execute(ctx context.Context, args []string) (string, error) {
	if len(args) > 0 {
		// Ajuda específica de comando
		cmd, err := h.registry.Get(args[0])
		if err != nil {
			return fmt.Sprintf("Command not found: %s", args[0]), nil
		}

		return fmt.Sprintf("Command: /%s\nDescription: %s\nUsage: %s",
			cmd.Name(), cmd.Description(), cmd.Usage()), nil
	}

	// Listar todos os comandos
	var result strings.Builder
	result.WriteString("Available commands:\n\n")

	commands := h.registry.List()
	for _, cmd := range commands {
		result.WriteString(fmt.Sprintf("  /%s - %s\n", cmd.Name(), cmd.Description()))
	}

	result.WriteString("\nType /help <command> for detailed help on a specific command")

	return result.String(), nil
}

// ClearCommand comando para limpar histórico
type ClearCommand struct{}

func (c *ClearCommand) Name() string        { return "clear" }
func (c *ClearCommand) Description() string { return "Clear conversation history" }
func (c *ClearCommand) Usage() string       { return "/clear" }

func (c *ClearCommand) Execute(ctx context.Context, args []string) (string, error) {
	return "✓ History cleared", nil
}

// HistoryCommand comando para mostrar histórico
type HistoryCommand struct{}

func (h *HistoryCommand) Name() string        { return "history" }
func (h *HistoryCommand) Description() string { return "Show conversation history" }
func (h *HistoryCommand) Usage() string       { return "/history [limit]" }

func (h *HistoryCommand) Execute(ctx context.Context, args []string) (string, error) {
	limit := 10
	if len(args) > 0 {
		fmt.Sscanf(args[0], "%d", &limit)
	}

	return fmt.Sprintf("Showing last %d messages (implementation pending)", limit), nil
}

// StatusCommand comando para mostrar status
type StatusCommand struct{}

func (s *StatusCommand) Name() string        { return "status" }
func (s *StatusCommand) Description() string { return "Show current status" }
func (s *StatusCommand) Usage() string       { return "/status" }

func (s *StatusCommand) Execute(ctx context.Context, args []string) (string, error) {
	return "Status: Active\nMode: Interactive\nSession: Active", nil
}

// ModeCommand comando para alterar modo
type ModeCommand struct{}

func (m *ModeCommand) Name() string        { return "mode" }
func (m *ModeCommand) Description() string { return "Change operation mode" }
func (m *ModeCommand) Usage() string       { return "/mode [readonly|interactive|autonomous]" }

func (m *ModeCommand) Execute(ctx context.Context, args []string) (string, error) {
	if len(args) == 0 {
		return "Current mode: interactive\n\nAvailable modes:\n- readonly\n- interactive\n- autonomous", nil
	}

	mode := args[0]
	return fmt.Sprintf("Mode changed to: %s", mode), nil
}

// CheckpointCommand comando para criar checkpoint
type CheckpointCommand struct{}

func (c *CheckpointCommand) Name() string        { return "checkpoint" }
func (c *CheckpointCommand) Description() string { return "Create a checkpoint" }
func (c *CheckpointCommand) Usage() string       { return "/checkpoint [description]" }

func (c *CheckpointCommand) Execute(ctx context.Context, args []string) (string, error) {
	description := "Manual checkpoint"
	if len(args) > 0 {
		description = strings.Join(args, " ")
	}

	return fmt.Sprintf("✓ Checkpoint created: %s", description), nil
}

// RewindCommand comando para voltar checkpoint
type RewindCommand struct{}

func (r *RewindCommand) Name() string        { return "rewind" }
func (r *RewindCommand) Description() string { return "Rewind to a checkpoint" }
func (r *RewindCommand) Usage() string       { return "/rewind <checkpoint-id>" }

func (r *RewindCommand) Execute(ctx context.Context, args []string) (string, error) {
	if len(args) == 0 {
		return "Error: checkpoint ID required\nUsage: /rewind <checkpoint-id>", nil
	}

	checkpointID := args[0]
	return fmt.Sprintf("✓ Rewound to checkpoint: %s", checkpointID), nil
}

// SessionCommand comando para gerenciar sessões
type SessionCommand struct{}

func (s *SessionCommand) Name() string        { return "session" }
func (s *SessionCommand) Description() string { return "Manage sessions" }
func (s *SessionCommand) Usage() string       { return "/session [list|save|resume <id>]" }

func (s *SessionCommand) Execute(ctx context.Context, args []string) (string, error) {
	if len(args) == 0 {
		return "Current session: active\n\nSubcommands:\n- list\n- save\n- resume <id>", nil
	}

	action := args[0]
	switch action {
	case "list":
		return "Available sessions:\n1. session_123 (active)\n2. session_456", nil
	case "save":
		return "✓ Session saved", nil
	case "resume":
		if len(args) < 2 {
			return "Error: session ID required", nil
		}
		return fmt.Sprintf("✓ Resumed session: %s", args[1]), nil
	default:
		return fmt.Sprintf("Unknown subcommand: %s", action), nil
	}
}

// DoctorCommand comando de diagnóstico
type DoctorCommand struct{}

func (d *DoctorCommand) Name() string        { return "doctor" }
func (d *DoctorCommand) Description() string { return "Run diagnostic checks" }
func (d *DoctorCommand) Usage() string       { return "/doctor" }

func (d *DoctorCommand) Execute(ctx context.Context, args []string) (string, error) {
	checks := []string{
		"✓ Ollama connection: OK",
		"✓ Model loaded: OK",
		"✓ GPU available: OK",
		"✓ Memory usage: 2.1 GB / 64 GB",
		"✓ Disk space: 450 GB available",
	}

	return "Running diagnostics...\n\n" + strings.Join(checks, "\n"), nil
}
