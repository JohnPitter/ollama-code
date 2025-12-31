package confirmation

// Question representa uma pergunta com múltiplas opções
type Question struct {
	Question    string   // Pergunta completa
	Header      string   // Label curto (max 12 chars)
	Options     []Option // 2-4 opções disponíveis
	MultiSelect bool     // Permite selecionar múltiplas opções
}

// Option representa uma opção de resposta
type Option struct {
	Label       string // Texto da opção (1-5 palavras)
	Description string // Explicação da opção
}

// Answer representa a resposta do usuário
type Answer struct {
	Question       string   // Pergunta respondida
	SelectedLabel  string   // Label da opção selecionada (single select)
	SelectedLabels []string // Labels das opções selecionadas (multi select)
	CustomInput    string   // Input customizado (quando selecionou "Other")
}

// QuestionSet representa um conjunto de perguntas (1-4)
type QuestionSet struct {
	Questions []Question
}

// Validate valida uma Question
func (q *Question) Validate() error {
	if q.Question == "" {
		return ErrEmptyQuestion
	}

	if len(q.Header) > 12 {
		return ErrHeaderTooLong
	}

	if len(q.Options) < 2 || len(q.Options) > 4 {
		return ErrInvalidOptionCount
	}

	for _, opt := range q.Options {
		if opt.Label == "" {
			return ErrEmptyOptionLabel
		}
	}

	return nil
}

// Validate valida um QuestionSet
func (qs *QuestionSet) Validate() error {
	if len(qs.Questions) < 1 || len(qs.Questions) > 4 {
		return ErrInvalidQuestionCount
	}

	for _, q := range qs.Questions {
		if err := q.Validate(); err != nil {
			return err
		}
	}

	return nil
}
