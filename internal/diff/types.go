package diff

import (
	"fmt"
	"time"
)

// ChangeType tipo de mudança
type ChangeType string

const (
	ChangeAdd    ChangeType = "add"
	ChangeDelete ChangeType = "delete"
	ChangeModify ChangeType = "modify"
)

// Change representa uma mudança em um arquivo
type Change struct {
	Type      ChangeType // Tipo de mudança
	StartLine int        // Linha inicial (1-indexed)
	EndLine   int        // Linha final (1-indexed)
	OldText   string     // Texto antigo
	NewText   string     // Texto novo
}

// FileDiff representa diff de um arquivo
type FileDiff struct {
	FilePath   string    // Caminho do arquivo
	Changes    []Change  // Lista de mudanças
	OldContent string    // Conteúdo antigo completo
	NewContent string    // Conteúdo novo completo
	Timestamp  time.Time // Quando foi criado
}

// EditRange representa um range de linhas para edição
type EditRange struct {
	Start int    // Linha inicial (1-indexed, inclusivo)
	End   int    // Linha final (1-indexed, inclusivo)
	Text  string // Texto a ser inserido no range
}

// EditHistory item do histórico de edições
type EditHistory struct {
	FilePath  string    // Arquivo editado
	Diff      *FileDiff // Diff da edição
	Timestamp time.Time // Quando foi feito
}

// ParseRange parseia string "start:end" para EditRange
// Exemplos: "10:20", "5:5", "1:10"
func ParseRange(rangeStr string) (*EditRange, error) {
	var start, end int
	_, err := fmt.Sscanf(rangeStr, "%d:%d", &start, &end)
	if err != nil {
		return nil, fmt.Errorf("invalid range format, expected 'start:end': %w", err)
	}

	if start < 1 || end < 1 {
		return nil, fmt.Errorf("line numbers must be >= 1")
	}

	if start > end {
		return nil, fmt.Errorf("start line must be <= end line")
	}

	return &EditRange{
		Start: start,
		End:   end,
	}, nil
}
