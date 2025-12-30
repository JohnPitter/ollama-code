package diff

import (
	"fmt"
	"strings"
	"time"
)

// Differ engine de diff
type Differ struct {
	history []EditHistory // Histórico de edições
}

// NewDiffer cria novo differ
func NewDiffer() *Differ {
	return &Differ{
		history: make([]EditHistory, 0),
	}
}

// ComputeDiff calcula diff entre dois conteúdos
func (d *Differ) ComputeDiff(filePath, oldContent, newContent string) *FileDiff {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	changes := d.computeChanges(oldLines, newLines)

	return &FileDiff{
		FilePath:   filePath,
		Changes:    changes,
		OldContent: oldContent,
		NewContent: newContent,
		Timestamp:  time.Now(),
	}
}

// computeChanges calcula mudanças linha por linha
func (d *Differ) computeChanges(oldLines, newLines []string) []Change {
	changes := make([]Change, 0)

	maxLen := len(oldLines)
	if len(newLines) > maxLen {
		maxLen = len(newLines)
	}

	i := 0
	for i < maxLen {
		// Ambas as linhas existem
		if i < len(oldLines) && i < len(newLines) {
			if oldLines[i] != newLines[i] {
				// Linha modificada
				changes = append(changes, Change{
					Type:      ChangeModify,
					StartLine: i + 1,
					EndLine:   i + 1,
					OldText:   oldLines[i],
					NewText:   newLines[i],
				})
			}
		} else if i < len(newLines) {
			// Linha adicionada
			changes = append(changes, Change{
				Type:      ChangeAdd,
				StartLine: i + 1,
				EndLine:   i + 1,
				OldText:   "",
				NewText:   newLines[i],
			})
		} else if i < len(oldLines) {
			// Linha deletada
			changes = append(changes, Change{
				Type:      ChangeDelete,
				StartLine: i + 1,
				EndLine:   i + 1,
				OldText:   oldLines[i],
				NewText:   "",
			})
		}
		i++
	}

	return changes
}

// ApplyEdit aplica edição em um range de linhas
func (d *Differ) ApplyEdit(filePath, content string, editRange EditRange) (string, *FileDiff, error) {
	lines := strings.Split(content, "\n")

	// Validar range
	if editRange.Start < 1 || editRange.Start > len(lines) {
		return "", nil, fmt.Errorf("start line %d out of range (1-%d)", editRange.Start, len(lines))
	}
	if editRange.End < editRange.Start || editRange.End > len(lines) {
		return "", nil, fmt.Errorf("end line %d out of range (%d-%d)", editRange.End, editRange.Start, len(lines))
	}

	// Salvar conteúdo antigo
	oldContent := content

	// Aplicar edição
	newLines := make([]string, 0)

	// Linhas antes do range
	newLines = append(newLines, lines[:editRange.Start-1]...)

	// Substituir range pelo novo texto
	if editRange.Text != "" {
		newTextLines := strings.Split(editRange.Text, "\n")
		newLines = append(newLines, newTextLines...)
	}

	// Linhas depois do range
	if editRange.End < len(lines) {
		newLines = append(newLines, lines[editRange.End:]...)
	}

	newContent := strings.Join(newLines, "\n")

	// Calcular diff
	diff := d.ComputeDiff(filePath, oldContent, newContent)

	// Adicionar ao histórico
	d.addToHistory(filePath, diff)

	return newContent, diff, nil
}

// Rollback desfaz a última edição de um arquivo
func (d *Differ) Rollback(filePath string) (string, error) {
	// Buscar última edição do arquivo
	for i := len(d.history) - 1; i >= 0; i-- {
		if d.history[i].FilePath == filePath {
			// Retornar conteúdo antigo
			oldContent := d.history[i].Diff.OldContent

			// Remover do histórico
			d.history = append(d.history[:i], d.history[i+1:]...)

			return oldContent, nil
		}
	}

	return "", fmt.Errorf("no edit history found for %s", filePath)
}

// GetHistory retorna histórico de edições
func (d *Differ) GetHistory(filePath string) []EditHistory {
	if filePath == "" {
		return d.history
	}

	filtered := make([]EditHistory, 0)
	for _, h := range d.history {
		if h.FilePath == filePath {
			filtered = append(filtered, h)
		}
	}

	return filtered
}

// ClearHistory limpa histórico
func (d *Differ) ClearHistory() {
	d.history = make([]EditHistory, 0)
}

// addToHistory adiciona ao histórico
func (d *Differ) addToHistory(filePath string, diff *FileDiff) {
	d.history = append(d.history, EditHistory{
		FilePath:  filePath,
		Diff:      diff,
		Timestamp: time.Now(),
	})
}
