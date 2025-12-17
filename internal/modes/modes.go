package modes

// OperationMode modo de operação do agente
type OperationMode string

const (
	// ModeReadOnly apenas leitura, sem modificações
	ModeReadOnly OperationMode = "readonly"

	// ModeInteractive confirmação antes de ações destrutivas (padrão)
	ModeInteractive OperationMode = "interactive"

	// ModeAutonomous totalmente autônomo, sem confirmações
	ModeAutonomous OperationMode = "autonomous"
)

// String retorna string representation
func (m OperationMode) String() string {
	return string(m)
}

// IsValid verifica se modo é válido
func (m OperationMode) IsValid() bool {
	switch m {
	case ModeReadOnly, ModeInteractive, ModeAutonomous:
		return true
	default:
		return false
	}
}

// AllowsWrites verifica se permite escritas
func (m OperationMode) AllowsWrites() bool {
	return m != ModeReadOnly
}

// RequiresConfirmation verifica se requer confirmação
func (m OperationMode) RequiresConfirmation() bool {
	return m == ModeInteractive
}

// Description retorna descrição do modo
func (m OperationMode) Description() string {
	switch m {
	case ModeReadOnly:
		return "Somente leitura - Nenhuma modificação será feita"
	case ModeInteractive:
		return "Interativo - Pede confirmação antes de modificações"
	case ModeAutonomous:
		return "Autônomo - Executa tudo automaticamente"
	default:
		return "Modo desconhecido"
	}
}

// ParseMode faz parse de string para OperationMode
func ParseMode(s string) OperationMode {
	mode := OperationMode(s)
	if mode.IsValid() {
		return mode
	}
	return ModeInteractive // Default
}
