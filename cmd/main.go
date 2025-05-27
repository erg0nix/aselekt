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

type FileItem string

func (f FileItem) Title() string       { return string(f) }
func (f FileItem) Description() string { return "" }
func (f FileItem) FilterValue() string { return string(f) }

type StarredItem struct{ FileItem }

func (s StarredItem) Title() string       { return s.FileItem.Title() }
func (s StarredItem) Description() string { return "" }
func (s StarredItem) FilterValue() string { return s.FileItem.FilterValue() }

type ItemDelegate struct{}

func (ItemDelegate) Height() int                             { return 1 }
func (ItemDelegate) Spacing() int                            { return 0 }
func (ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (ItemDelegate) Render(w io.Writer, m list.Model, i int, it list.Item) {
	name := it.FilterValue()
	prefix := "  "
	if i == m.Index() {
		prefix = "> "
	}
	if _, ok := it.(StarredItem); ok {
		fmt.Fprintf(w, "%s* %s", prefix, name)
	} else {
		fmt.Fprintf(w, "%s  %s", prefix, name)
	}
}

// We need a separate list of starred files to be read by delegate
var Starred []string

type App struct {
	Input     textinput.Model
	List      list.Model
	Selected  []string
	LastQuery string
	Err       error
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
	query = strings.ToLower(query)

	items := make([]list.Item, 0, len(starred)+len(all))
	for _, s := range starred {
		items = append(items, StarredItem{FileItem(s)})
	}
	for _, f := range all {
		if query == "" || strings.Contains(strings.ToLower(f), query) {
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

func CopyFilesToClipboard(paths []string) (string, int, error) {
	if err := clipboard.Init(); err != nil {
		return "", 0, fmt.Errorf("init clipboard: %w", err)
	}

	var b strings.Builder
	totalLines := 0

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", 0, fmt.Errorf("read %s: %w", path, err)
		}

		lines := strings.Count(string(data), "\n")
		totalLines += lines

		fmt.Fprintf(&b, "# %s\n\n%s\n\n", path, data)
	}

	text := b.String()
	clipboard.Write(clipboard.FmtText, []byte(text))

	return text, totalLines, nil
}

func NewApp() App {
	all, _ := AllFiles()

	in := textinput.New()
	in.Placeholder = "Type to search…"
	in.Focus()
	in.Width = 40

	l := list.New(BuildItems(all, "", nil), ItemDelegate{}, 40, 10)
	l.Title = ""
	l.Styles.Title, l.Styles.TitleBar, l.Styles.PaginationStyle =
		lipgloss.NewStyle(), lipgloss.NewStyle(), lipgloss.NewStyle()
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.SetShowStatusBar(false)
	km := l.KeyMap
	km.Quit = key.NewBinding()
	l.KeyMap = km

	return App{Input: in, List: l}
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

				Starred = a.Selected
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

	// propagate to input
	var c tea.Cmd
	a.Input, c = a.Input.Update(msg)
	cmds = append(cmds, c)

	// new query
	if _, ok := msg.(tea.KeyMsg); ok {
		a.LastQuery = a.Input.Value()
		if all, err := AllFiles(); err == nil {
			cmds = append(cmds, Search(all, a.LastQuery, a.Selected))
		}
	}

	// propagate to list
	a.List, c = a.List.Update(msg)
	cmds = append(cmds, c)

	return a, tea.Batch(cmds...)
}

func (a App) View() string {
	var b strings.Builder
	b.WriteString("\n───────── Search ─────────\n\n")
	b.WriteString(a.Input.View())
	b.WriteString("\n\n───────── Results ────────")
	b.WriteString(a.List.View())
	return b.String()
}

func main() {
	model, err := tea.NewProgram(NewApp()).Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	app, ok := model.(App)
	if !ok || len(app.Selected) == 0 {
		return
	}

	_, totalLines, err := CopyFilesToClipboard(app.Selected)
	if err != nil {
		fmt.Fprintf(os.Stderr, "clipboard error: %v\n", err)
		return
	}

	fmt.Println("\n✔ Copied to clipboard:")
	for _, f := range app.Selected {
		fmt.Printf("  • %s\n", f)
	}
	fmt.Printf("\n📄 Total lines copied: %d\n", totalLines)
}
