package confirmation

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
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
		m.green.Println("✓ Confirmado")
		return true, nil
	case "n", "não", "nao", "no":
		m.red.Println("✗ Cancelado")
		return false, nil
	default:
		m.red.Println("✗ Resposta inválida. Cancelando.")
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
		m.green.Println("✓ Confirmado")
		return true, nil
	default:
		m.red.Println("✗ Cancelado")
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
		m.green.Println("✓ Confirmado")
		return true, nil
	}

	m.red.Println("✗ Cancelado por segurança")
	return false, nil
}

// AskQuestion faz uma pergunta com múltiplas opções
func (m *Manager) AskQuestion(question Question) (*Answer, error) {
	// Validar pergunta
	if err := question.Validate(); err != nil {
		return nil, err
	}

	// Mostrar header
	if question.Header != "" {
		m.yellow.Printf("\n[%s]\n", question.Header)
	}

	// Mostrar pergunta
	fmt.Printf("\n%s\n\n", question.Question)

	// Mostrar opções
	for i, opt := range question.Options {
		fmt.Printf("%d. %s\n", i+1, opt.Label)
		if opt.Description != "" {
			fmt.Printf("   %s\n", opt.Description)
		}
	}

	// Adicionar opção "Other"
	otherIndex := len(question.Options) + 1
	fmt.Printf("%d. Other (digite sua resposta customizada)\n\n", otherIndex)

	// Ler resposta
	if question.MultiSelect {
		return m.readMultiSelectAnswer(question, otherIndex)
	}
	return m.readSingleSelectAnswer(question, otherIndex)
}

// readSingleSelectAnswer lê resposta single select
func (m *Manager) readSingleSelectAnswer(question Question, otherIndex int) (*Answer, error) {
	m.yellow.Printf("Selecione uma opção (1-%d): ", otherIndex)

	response, err := m.reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	response = strings.TrimSpace(response)
	selection, err := strconv.Atoi(response)
	if err != nil {
		return nil, ErrInvalidSelection
	}

	// Verificar se selecionou "Other"
	if selection == otherIndex {
		m.yellow.Print("Digite sua resposta: ")
		customInput, err := m.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		customInput = strings.TrimSpace(customInput)
		if customInput == "" {
			return nil, ErrNoSelection
		}

		m.green.Printf("✓ Resposta customizada: %s\n", customInput)
		return &Answer{
			Question:    question.Question,
			CustomInput: customInput,
		}, nil
	}

	// Verificar se seleção é válida
	if selection < 1 || selection > len(question.Options) {
		return nil, ErrInvalidSelection
	}

	selectedOption := question.Options[selection-1]
	m.green.Printf("✓ Selecionado: %s\n", selectedOption.Label)

	return &Answer{
		Question:      question.Question,
		SelectedLabel: selectedOption.Label,
	}, nil
}

// readMultiSelectAnswer lê resposta multi select
func (m *Manager) readMultiSelectAnswer(question Question, otherIndex int) (*Answer, error) {
	m.yellow.Print("Selecione opções separadas por vírgula (ex: 1,3): ")

	response, err := m.reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	response = strings.TrimSpace(response)
	if response == "" {
		return nil, ErrNoSelection
	}

	// Parse selections
	parts := strings.Split(response, ",")
	selectedLabels := make([]string, 0)
	customInput := ""

	for _, part := range parts {
		part = strings.TrimSpace(part)
		selection, err := strconv.Atoi(part)
		if err != nil {
			return nil, ErrInvalidSelection
		}

		// "Other" selecionado
		if selection == otherIndex {
			m.yellow.Print("Digite sua resposta customizada: ")
			custom, err := m.reader.ReadString('\n')
			if err != nil {
				return nil, err
			}
			customInput = strings.TrimSpace(custom)
			continue
		}

		// Validar seleção
		if selection < 1 || selection > len(question.Options) {
			return nil, ErrInvalidSelection
		}

		selectedLabels = append(selectedLabels, question.Options[selection-1].Label)
	}

	if len(selectedLabels) == 0 && customInput == "" {
		return nil, ErrNoSelection
	}

	m.green.Printf("✓ Selecionado: %v\n", selectedLabels)
	if customInput != "" {
		m.green.Printf("✓ Custom: %s\n", customInput)
	}

	return &Answer{
		Question:       question.Question,
		SelectedLabels: selectedLabels,
		CustomInput:    customInput,
	}, nil
}

// AskQuestions faz múltiplas perguntas (1-4)
func (m *Manager) AskQuestions(questionSet QuestionSet) (map[string]*Answer, error) {
	// Validar question set
	if err := questionSet.Validate(); err != nil {
		return nil, err
	}

	answers := make(map[string]*Answer)

	m.yellow.Println("\n📋 Respondendo perguntas...")

	for i, question := range questionSet.Questions {
		m.yellow.Printf("\nPergunta %d/%d\n", i+1, len(questionSet.Questions))

		answer, err := m.AskQuestion(question)
		if err != nil {
			return nil, fmt.Errorf("erro na pergunta %d: %w", i+1, err)
		}

		// Usar header como key, ou fallback para índice
		key := question.Header
		if key == "" {
			key = fmt.Sprintf("question_%d", i+1)
		}

		answers[key] = answer
	}

	m.green.Println("\n✓ Todas as perguntas respondidas!")
	return answers, nil
}
