package view

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

var styles = struct {
	Label   lipgloss.Style
	Cursor  lipgloss.Style
	Starred lipgloss.Style
	Normal  lipgloss.Style
	Help    lipgloss.Style
}{
	Label:   lipgloss.NewStyle().Foreground(lipgloss.Color("#7dd3fc")),
	Cursor:  lipgloss.NewStyle().Foreground(lipgloss.Color("#f472b6")).Bold(true),
	Starred: lipgloss.NewStyle().Foreground(lipgloss.Color("#facc15")).Bold(true),
	Normal:  lipgloss.NewStyle().Foreground(lipgloss.Color("#cbd5e1")),
	Help:    lipgloss.NewStyle().Foreground(lipgloss.Color("#94a3b8")).MarginTop(1),
}

func RenderApp(input textinput.Model, list list.Model) string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(styles.Label.Render("Search: "))
	b.WriteString(input.View())
	b.WriteString("\n")
	b.WriteString(list.View())
	b.WriteString(styles.Help.Render(
		"\nPress Esc or Ctrl+C to quit — selected files will be copied to the clipboard.",
	))
	return b.String()
}
