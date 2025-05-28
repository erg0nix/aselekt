package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"

	"golang.design/x/clipboard"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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

type FileItem string

func (f FileItem) Title() string       { return string(f) }
func (f FileItem) Description() string { return "" }
func (f FileItem) FilterValue() string { return string(f) }

type StarredItem struct{ FileItem }

func (s StarredItem) Title() string       { return s.FileItem.Title() }
func (s StarredItem) Description() string { return "" }
func (s StarredItem) FilterValue() string { return s.FileItem.FilterValue() }

type ItemDelegate struct{ S Styles }

func (ItemDelegate) Height() int                             { return 1 }
func (ItemDelegate) Spacing() int                            { return 0 }
func (ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d ItemDelegate) Render(w io.Writer, m list.Model, i int, it list.Item) {
	name := it.FilterValue()
	starred := false
	if _, ok := it.(StarredItem); ok {
		starred = true
	}

	switch {
	case i == m.Index() && !starred:
		fmt.Fprint(w, d.S.Cursor.Render("> "+name))
	case i == m.Index() && starred:
		fmt.Fprint(w, "> "+d.S.Starred.Render("* "+name))
	case starred:
		fmt.Fprint(w, d.S.Starred.Render("  * "+name))
	default:
		fmt.Fprint(w, d.S.Normal.Render("  "+name))
	}
}

type App struct {
	Input     textinput.Model
	List      list.Model
	Selected  []string
	LastQuery string
	Err       error
	Styles    Styles
}

type ResultsMsg []list.Item
type ErrorMsg struct{ error }

func AllFiles() ([]string, error) {
	out, err := exec.Command("fd", "--type", "f", "--strip-cwd-prefix").Output()
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(string(out)), "\n"), nil
}

func BuildItems(all []string, query string, starred []string) []list.Item {
	q := strings.ToLower(query)
	var items []list.Item
	for _, s := range starred {
		items = append(items, StarredItem{FileItem(s)})
	}
	for _, f := range all {
		if q == "" || strings.Contains(strings.ToLower(f), q) {
			items = append(items, FileItem(f))
		}
	}
	return items
}

func Search(all []string, query string, starred []string) tea.Cmd {
	return func() tea.Msg { return ResultsMsg(BuildItems(all, query, starred)) }
}

func Remove(selectedFiles []string, fileToRemove string) []string {
	if idx := slices.Index(selectedFiles, fileToRemove); idx != -1 {
		return slices.Delete(selectedFiles, idx, idx+1)
	}
	return selectedFiles
}

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
	all, _ := AllFiles()
	st := NewStyles()

	in := textinput.New()
	in.Placeholder = "Type to search…"
	in.Focus()
	in.Width = 40

	delegate := ItemDelegate{S: st}
	l := list.New(BuildItems(all, "", nil), delegate, 40, 10)
	l.Title = ""
	l.Styles = list.DefaultStyles()
	l.Styles.Title = lipgloss.NewStyle()
	l.Styles.TitleBar = lipgloss.NewStyle()
	l.Styles.PaginationStyle = lipgloss.NewStyle()

	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.SetShowStatusBar(false)
	km := l.KeyMap
	km.Quit = key.NewBinding()
	km.Filter = key.NewBinding()
	l.KeyMap = km

	return App{Input: in, List: l, Styles: st}
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
			var name string
			switch itm := a.List.SelectedItem().(type) {
			case FileItem:
				name = string(itm)
			case StarredItem:
				name = string(itm.FileItem)
			}
			if name != "" {
				if slices.Contains(a.Selected, name) {
					a.Selected = Remove(a.Selected, name)
				} else {
					a.Selected = append(a.Selected, name)
				}
				if all, err := AllFiles(); err == nil {
					a.List.SetItems(BuildItems(all, a.LastQuery, a.Selected))
				}
			}
		}

	case ResultsMsg:
		a.List.SetItems(v)
		return a, nil

	case ErrorMsg:
		a.Err = v.error
		return a, nil
	}

	var c tea.Cmd
	a.Input, c = a.Input.Update(msg)
	cmds = append(cmds, c)

	if _, ok := msg.(tea.KeyMsg); ok {
		a.LastQuery = a.Input.Value()
		if all, err := AllFiles(); err == nil {
			cmds = append(cmds, Search(all, a.LastQuery, a.Selected))
		}
	}

	a.List, c = a.List.Update(msg)
	cmds = append(cmds, c)

	return a, tea.Batch(cmds...)
}

func (a App) View() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(a.Styles.Label.Render("Search: "))
	b.WriteString(a.Input.View())
	b.WriteString("\n")
	b.WriteString(a.List.View())
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

	if len(app.Selected) == 0 {
		fmt.Println("\nNo files selected – clipboard unchanged.")
		return
	}

	lines, err := CopyFilesToClipboard(app.Selected)
	if err != nil {
		fmt.Fprintf(os.Stderr, "clipboard error: %v\n", err)
		return
	}

	success := lipgloss.NewStyle().Foreground(lipgloss.Color("#4ade80")).Bold(true)
	fileStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#cbd5e1"))

	fmt.Println(success.Render("\n✔ Copied to clipboard:"))
	for _, f := range app.Selected {
		fmt.Printf("%s %s\n", fileStyle.Render("•"), fileStyle.Render(f))
	}
	fmt.Printf("\nTotal lines copied: %d\n", lines)
}
