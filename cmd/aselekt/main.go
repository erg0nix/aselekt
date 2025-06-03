package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"aselekt/internal/search"
	"aselekt/internal/view"

	"golang.design/x/clipboard"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ItemDelegate struct{ S view.Styles }

func (ItemDelegate) Height() int                             { return 1 }
func (ItemDelegate) Spacing() int                            { return 0 }
func (ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d ItemDelegate) Render(w io.Writer, m list.Model, i int, it list.Item) {
	fileItem := it.(search.FileItem)

	var rendered string

	if i == m.Index() {
		if fileItem.Starred {
			rendered = d.S.Starred.Render("> * " + fileItem.Path)
		} else {
			rendered = d.S.Cursor.Render("> " + fileItem.Path)
		}
	} else {
		if fileItem.Starred {
			rendered = d.S.Starred.Render("  * " + fileItem.Path)
		} else {
			rendered = d.S.Normal.Render("  " + fileItem.Path)
		}
	}

	fmt.Fprint(w, rendered)
}

type App struct {
	Search search.FileSearch
	Input  textinput.Model
	UIList list.Model
	Err    error
	Styles view.Styles
}

type ResultsMsg []list.Item
type ErrorMsg struct{ error }

func CopyFilesToClipboard(paths []string) (int, error) {
	if err := clipboard.Init(); err != nil {
		return 0, fmt.Errorf("init clipboard: %w", err)
	}
	var b strings.Builder
	lines := 0
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			return 0, err
		}
		lines += strings.Count(string(data), "\n")
		fmt.Fprintf(&b, "# %s\n\n%s\n\n", p, data)
	}
	clipboard.Write(clipboard.FmtText, []byte(b.String()))
	return lines, nil
}

func NewApp() App {
	fs, err := search.NewFileSearch()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fd error: %v\n", err)
	}

	st := view.NewStyles()
	in := textinput.New()
	in.Placeholder = "Type to search…"
	in.Focus()
	in.Width = 40

	delegate := ItemDelegate{S: st}
	items := make([]list.Item, 0)
	for _, f := range fs.BuildItems() {
		items = append(items, f)
	}
	uiList := list.New(items, delegate, 40, 10)
	uiList.Title = ""
	uiList.Styles = list.DefaultStyles()
	uiList.Styles.Title = lipgloss.NewStyle()
	uiList.Styles.TitleBar = lipgloss.NewStyle()
	uiList.Styles.PaginationStyle = lipgloss.NewStyle()

	uiList.SetShowHelp(false)
	uiList.SetShowPagination(false)
	uiList.SetShowStatusBar(false)
	km := uiList.KeyMap
	km.Quit = key.NewBinding()
	km.Filter = key.NewBinding()
	uiList.KeyMap = km

	return App{Search: fs, Input: in, UIList: uiList, Styles: st}
}

func (a App) Init() tea.Cmd { return nil }

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch v := msg.(type) {
	case tea.KeyMsg:
		switch v.String() {
		case "ctrl+c", "esc":
			return a, tea.Quit
		case "enter":
			if fileItem, ok := a.UIList.SelectedItem().(search.FileItem); ok {
				a.Search.ToggleSelection(fileItem.Path)

				items := make([]list.Item, 0)
				for _, f := range a.Search.BuildItems() {
					items = append(items, f)
				}
				a.UIList.SetItems(items)
			}
		}

	case ResultsMsg:
		a.UIList.SetItems(v)
		return a, nil

	case ErrorMsg:
		a.Err = v.error
		return a, nil
	}

	var c tea.Cmd
	a.Input, c = a.Input.Update(msg)
	cmds = append(cmds, c)

	if _, ok := msg.(tea.KeyMsg); ok {
		a.Search.Query = a.Input.Value()

		items := make([]list.Item, 0)
		for _, f := range a.Search.BuildItems() {
			items = append(items, f)
		}
		a.UIList.SetItems(items)
	}

	a.UIList, c = a.UIList.Update(msg)
	cmds = append(cmds, c)

	return a, tea.Batch(cmds...)
}

func (a App) View() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(a.Styles.Label.Render("Search: "))
	b.WriteString(a.Input.View())
	b.WriteString("\n")
	b.WriteString(a.UIList.View())
	b.WriteString(a.Styles.Help.Render(
		"\nPress Esc or Ctrl+C to quit — selected files will be copied to the clipboard.",
	))
	return b.String()
}

func main() {
	model, err := tea.NewProgram(NewApp()).Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	app, ok := model.(App)
	if !ok {
		return
	}

	if len(app.Search.Selected) == 0 {
		fmt.Println("\nNo files selected – clipboard unchanged.")
		return
	}

	lines, err := CopyFilesToClipboard(app.Search.Selected)
	if err != nil {
		fmt.Fprintf(os.Stderr, "clipboard error: %v\n", err)
		return
	}

	success := lipgloss.NewStyle().Foreground(lipgloss.Color("#4ade80")).Bold(true)
	fileStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#cbd5e1"))

	fmt.Println(success.Render("\n✔ Copied to clipboard:"))
	for _, f := range app.Search.Selected {
		fmt.Printf("%s %s\n", fileStyle.Render("•"), fileStyle.Render(f))
	}
	fmt.Printf("\nTotal lines copied: %d\n", lines)
}
