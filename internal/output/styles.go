package output

import "github.com/fatih/color"

// Style estilo de output
type Style string

const (
	StyleDefault     Style = "default"
	StyleExplanatory Style = "explanatory"
	StyleLearning    Style = "learning"
	StyleCorporate   Style = "corporate"
)

// Formatter formatador de output
type Formatter struct {
	style  Style
	colors bool
}

// NewFormatter cria novo formatador
func NewFormatter(style Style, useColors bool) *Formatter {
	return &Formatter{
		style:  style,
		colors: useColors,
	}
}

// Format formata mensagem de acordo com estilo
func (f *Formatter) Format(message string) string {
	switch f.style {
	case StyleExplanatory:
		return f.formatExplanatory(message)
	case StyleLearning:
		return f.formatLearning(message)
	case StyleCorporate:
		return f.formatCorporate(message)
	default:
		return message
	}
}

func (f *Formatter) formatExplanatory(message string) string {
	if f.colors {
		blue := color.New(color.FgBlue, color.Bold)
		return blue.Sprint("💡 ") + message
	}
	return "💡 " + message
}

func (f *Formatter) formatLearning(message string) string {
	if f.colors {
		green := color.New(color.FgGreen)
		return green.Sprint("📚 ") + message
	}
	return "📚 " + message
}

func (f *Formatter) formatCorporate(message string) string {
	// Estilo mais formal, sem emojis
	return message
}
