package confirmation

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

// Manager gerenciador de confirmações
type Manager struct {
	reader *bufio.Reader
	yellow *color.Color
	red    *color.Color
	green  *color.Color
}

// NewManager cria novo gerenciador
func NewManager() *Manager {
	return &Manager{
		reader: bufio.NewReader(os.Stdin),
		yellow: color.New(color.FgYellow, color.Bold),
		red:    color.New(color.FgRed, color.Bold),
		green:  color.New(color.FgGreen, color.Bold),
	}
}

// Confirm pede confirmação ao usuário
func (m *Manager) Confirm(action, details string) (bool, error) {
	m.yellow.Println("\n⚠️  CONFIRMAÇÃO NECESSÁRIA")
	fmt.Printf("\nAção: %s\n", action)

	if details != "" {
		fmt.Printf("Detalhes:\n%s\n", details)
	}

	m.yellow.Print("\nDeseja continuar? (s/n): ")

	response, err := m.reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	response = strings.ToLower(strings.TrimSpace(response))

	switch response {
	case "s", "sim", "y", "yes":
		m.green.Println("✓ Confirmado\n")
		return true, nil
	case "n", "não", "nao", "no":
		m.red.Println("✗ Cancelado\n")
		return false, nil
	default:
		m.red.Println("✗ Resposta inválida. Cancelando.\n")
		return false, nil
	}
}

// ConfirmWithPreview pede confirmação mostrando preview
func (m *Manager) ConfirmWithPreview(action, preview string) (bool, error) {
	m.yellow.Println("\n⚠️  CONFIRMAÇÃO NECESSÁRIA")
	fmt.Printf("\nAção: %s\n", action)

	if preview != "" {
		fmt.Println("\nPreview:")
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println(preview)
		fmt.Println(strings.Repeat("─", 60))
	}

	m.yellow.Print("\nDeseja continuar? (s/n): ")

	response, err := m.reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	response = strings.ToLower(strings.TrimSpace(response))

	switch response {
	case "s", "sim", "y", "yes":
		m.green.Println("✓ Confirmado\n")
		return true, nil
	default:
		m.red.Println("✗ Cancelado\n")
		return false, nil
	}
}

// ConfirmDangerousAction pede confirmação para ação perigosa
func (m *Manager) ConfirmDangerousAction(action, warning string) (bool, error) {
	m.red.Println("\n⚠️  ATENÇÃO: AÇÃO POTENCIALMENTE PERIGOSA ⚠️")
	fmt.Printf("\nAção: %s\n", action)

	if warning != "" {
		m.red.Printf("\nAviso:\n%s\n", warning)
	}

	m.red.Print("\nTem CERTEZA que deseja continuar? Digite 'CONFIRMO' para prosseguir: ")

	response, err := m.reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	response = strings.TrimSpace(response)

	if response == "CONFIRMO" {
		m.green.Println("✓ Confirmado\n")
		return true, nil
	}

	m.red.Println("✗ Cancelado por segurança\n")
	return false, nil
}
