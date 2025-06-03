package view

import (
	"strings"

	"aselekt/internal/search"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	Label   lipgloss.Style
	Cursor  lipgloss.Style
	Starred lipgloss.Style
	Normal  lipgloss.Style
	Help    lipgloss.Style
}

var StylesInstance = Styles{
	Label:   lipgloss.NewStyle().Foreground(lipgloss.Color("#7dd3fc")),
	Cursor:  lipgloss.NewStyle().Foreground(lipgloss.Color("#f472b6")).Bold(true),
	Starred: lipgloss.NewStyle().Foreground(lipgloss.Color("#facc15")).Bold(true),
	Normal:  lipgloss.NewStyle().Foreground(lipgloss.Color("#cbd5e1")),
	Help:    lipgloss.NewStyle().Foreground(lipgloss.Color("#94a3b8")).MarginTop(1),
}

func InitTextInput() textinput.Model {
	in := textinput.New()
	in.Placeholder = "Type to search…"
	in.Focus()
	in.Width = 40
	return in
}

func InitFileList(fs search.FileSearch) list.Model {
	delegate := FileItemView{S: StylesInstance}

	items := make([]list.Item, 0, len(fs.Files))
	for _, f := range fs.BuildItems() {
		items = append(items, f)
	}

	l := list.New(items, delegate, 40, 10)
	l.Title = ""
	l.Styles = list.DefaultStyles()
	l.Styles.Title = lipgloss.NewStyle()
	l.Styles.TitleBar = lipgloss.NewStyle()
	l.Styles.PaginationStyle = lipgloss.NewStyle()

	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.SetShowStatusBar(false)

	l.KeyMap.Quit = key.NewBinding()
	l.KeyMap.Filter = key.NewBinding()

	return l
}

func RenderApp(input textinput.Model, list list.Model) string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(StylesInstance.Label.Render("Search: "))
	b.WriteString(input.View())
	b.WriteString("\n")
	b.WriteString(list.View())
	b.WriteString(StylesInstance.Help.Render(
		"\nPress Esc or Ctrl+C to quit — selected files will be copied to the clipboard.",
	))
	return b.String()
}
