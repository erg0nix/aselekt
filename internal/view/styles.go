package view

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	Label   lipgloss.Style
	Cursor  lipgloss.Style
	Starred lipgloss.Style
	Normal  lipgloss.Style
	Help    lipgloss.Style
}

func NewStyles() Styles {
	return Styles{
		Label:   lipgloss.NewStyle().Foreground(lipgloss.Color("#7dd3fc")),
		Cursor:  lipgloss.NewStyle().Foreground(lipgloss.Color("#f472b6")).Bold(true),
		Starred: lipgloss.NewStyle().Foreground(lipgloss.Color("#facc15")).Bold(true),
		Normal:  lipgloss.NewStyle().Foreground(lipgloss.Color("#cbd5e1")),
		Help:    lipgloss.NewStyle().Foreground(lipgloss.Color("#94a3b8")).MarginTop(1),
	}
}
