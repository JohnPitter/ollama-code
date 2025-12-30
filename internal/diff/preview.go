package diff

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// Previewer gera previews de diffs
type Previewer struct {
	green  *color.Color
	red    *color.Color
	yellow *color.Color
	cyan   *color.Color
}

// NewPreviewer cria novo previewer
func NewPreviewer() *Previewer {
	return &Previewer{
		green:  color.New(color.FgGreen),
		red:    color.New(color.FgRed),
		yellow: color.New(color.FgYellow),
		cyan:   color.New(color.FgCyan),
	}
}

// Preview gera preview colorizado de um diff
func (p *Previewer) Preview(diff *FileDiff) string {
	var sb strings.Builder

	// Header
	p.cyan.Fprintf(&sb, "📄 Arquivo: %s\n", diff.FilePath)
	sb.WriteString(strings.Repeat("─", 60) + "\n")

	if len(diff.Changes) == 0 {
		sb.WriteString("✓ Sem mudanças\n")
		return sb.String()
	}

	// Estatísticas
	adds := 0
	deletes := 0
	modifies := 0

	for _, change := range diff.Changes {
		switch change.Type {
		case ChangeAdd:
			adds++
		case ChangeDelete:
			deletes++
		case ChangeModify:
			modifies++
		}
	}

	p.yellow.Fprintf(&sb, "📊 Mudanças: ")
	if adds > 0 {
		p.green.Fprintf(&sb, "+%d ", adds)
	}
	if deletes > 0 {
		p.red.Fprintf(&sb, "-%d ", deletes)
	}
	if modifies > 0 {
		p.yellow.Fprintf(&sb, "~%d ", modifies)
	}
	sb.WriteString("\n\n")

	// Mudanças detalhadas
	for i, change := range diff.Changes {
		p.renderChange(&sb, change, i+1)
	}

	sb.WriteString(strings.Repeat("─", 60) + "\n")

	return sb.String()
}

// renderChange renderiza uma mudança
func (p *Previewer) renderChange(sb *strings.Builder, change Change, index int) {
	switch change.Type {
	case ChangeAdd:
		p.cyan.Fprintf(sb, "[%d] Linha %d: ", index, change.StartLine)
		p.green.Fprintln(sb, "ADICIONADA")
		p.green.Fprintf(sb, "+ %s\n\n", change.NewText)

	case ChangeDelete:
		p.cyan.Fprintf(sb, "[%d] Linha %d: ", index, change.StartLine)
		p.red.Fprintln(sb, "DELETADA")
		p.red.Fprintf(sb, "- %s\n\n", change.OldText)

	case ChangeModify:
		p.cyan.Fprintf(sb, "[%d] Linha %d: ", index, change.StartLine)
		p.yellow.Fprintln(sb, "MODIFICADA")
		p.red.Fprintf(sb, "- %s\n", change.OldText)
		p.green.Fprintf(sb, "+ %s\n\n", change.NewText)
	}
}

// PreviewRange gera preview de edição por range
func (p *Previewer) PreviewRange(filePath, oldContent string, editRange EditRange) string {
	var sb strings.Builder

	lines := strings.Split(oldContent, "\n")

	// Validar range
	if editRange.Start < 1 || editRange.Start > len(lines) {
		return fmt.Sprintf("❌ Erro: linha inicial %d fora do range (1-%d)\n", editRange.Start, len(lines))
	}
	if editRange.End < editRange.Start || editRange.End > len(lines) {
		return fmt.Sprintf("❌ Erro: linha final %d fora do range (%d-%d)\n", editRange.End, editRange.Start, len(lines))
	}

	// Header
	p.cyan.Fprintf(&sb, "📄 Arquivo: %s\n", filePath)
	p.yellow.Fprintf(&sb, "📝 Editando linhas %d-%d\n", editRange.Start, editRange.End)
	sb.WriteString(strings.Repeat("─", 60) + "\n\n")

	// Contexto antes (2 linhas)
	contextStart := editRange.Start - 3
	if contextStart < 1 {
		contextStart = 1
	}
	if contextStart < editRange.Start {
		p.cyan.Fprintln(&sb, "Contexto antes:")
		for i := contextStart; i < editRange.Start; i++ {
			fmt.Fprintf(&sb, "  %3d | %s\n", i, lines[i-1])
		}
		sb.WriteString("\n")
	}

	// Linhas que serão removidas
	p.red.Fprintln(&sb, "Linhas a remover:")
	for i := editRange.Start; i <= editRange.End; i++ {
		p.red.Fprintf(&sb, "- %3d | %s\n", i, lines[i-1])
	}
	sb.WriteString("\n")

	// Novo texto
	if editRange.Text != "" {
		p.green.Fprintln(&sb, "Novo texto:")
		newLines := strings.Split(editRange.Text, "\n")
		for i, line := range newLines {
			p.green.Fprintf(&sb, "+ %3d | %s\n", editRange.Start+i, line)
		}
	} else {
		p.yellow.Fprintln(&sb, "(linhas serão deletadas)")
	}
	sb.WriteString("\n")

	// Contexto depois (2 linhas)
	contextEnd := editRange.End + 3
	if contextEnd > len(lines) {
		contextEnd = len(lines)
	}
	if contextEnd > editRange.End {
		p.cyan.Fprintln(&sb, "Contexto depois:")
		for i := editRange.End + 1; i <= contextEnd; i++ {
			fmt.Fprintf(&sb, "  %3d | %s\n", i, lines[i-1])
		}
	}

	sb.WriteString(strings.Repeat("─", 60) + "\n")

	return sb.String()
}

// CompactPreview gera preview compacto (para logs)
func (p *Previewer) CompactPreview(diff *FileDiff) string {
	adds := 0
	deletes := 0
	modifies := 0

	for _, change := range diff.Changes {
		switch change.Type {
		case ChangeAdd:
			adds++
		case ChangeDelete:
			deletes++
		case ChangeModify:
			modifies++
		}
	}

	return fmt.Sprintf("%s: +%d ~%d -%d changes", diff.FilePath, adds, modifies, deletes)
}
