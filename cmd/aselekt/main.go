package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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

type FileItem struct {
	Path    string
	Starred bool
}

func (f FileItem) Title() string       { return filepath.Base(f.Path) }
func (f FileItem) Description() string { return "" }
func (f FileItem) FilterValue() string { return f.Path }

type ItemDelegate struct{ S Styles }

func (ItemDelegate) Height() int                             { return 1 }
func (ItemDelegate) Spacing() int                            { return 0 }
func (ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d ItemDelegate) Render(w io.Writer, m list.Model, i int, it list.Item) {
	fileItem := it.(FileItem)

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
	Files     []string
	Input     textinput.Model
	UIList    list.Model
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
		items = append(items, FileItem{Path: s, Starred: true})
	}
	for _, f := range all {
		if q == "" || strings.Contains(strings.ToLower(f), q) {
			items = append(items, FileItem{Path: f})
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
	all, err := AllFiles()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fd error: %v\n", err)
	}

	st := NewStyles()
	in := textinput.New()
	in.Placeholder = "Type to search…"
	in.Focus()
	in.Width = 40

	delegate := ItemDelegate{S: st}
	uiList := list.New(BuildItems(all, "", nil), delegate, 40, 10)
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

	return App{Files: all, Input: in, UIList: uiList, Styles: st}
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

			if fileItem, ok := a.UIList.SelectedItem().(FileItem); ok {
				name = fileItem.Path
			}

			if name != "" {
				if slices.Contains(a.Selected, name) {
					a.Selected = Remove(a.Selected, name)
				} else {
					a.Selected = append(a.Selected, name)
				}

				a.UIList.SetItems(BuildItems(a.Files, a.LastQuery, a.Selected))
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
		a.LastQuery = a.Input.Value()
		cmds = append(cmds, Search(a.Files, a.LastQuery, a.Selected))
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
