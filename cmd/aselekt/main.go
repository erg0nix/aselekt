package main

import (
	"fmt"
	"os"

	"aselekt/internal/clipboard"
	"aselekt/internal/search"
	"aselekt/internal/view"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type App struct {
	Search    search.FileSearch
	Input     textinput.Model
	UIList    list.Model
	StatusMsg string
}

func NewApp() App {
	fs, err := search.NewFileSearch()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fd error: %v\n", err)
	}

	return App{
		Search: fs,
		Input:  view.InitTextInput(),
		UIList: view.InitFileList(fs),
	}
}

func (a App) Init() tea.Cmd { return nil }

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch v := msg.(type) {
	case tea.KeyMsg:
		switch v.String() {
		case "ctrl+c", "esc":
			return a, tea.Quit
		case "ctrl+y":
			if len(a.Search.Selected) == 0 {
				a.StatusMsg = view.StylesInstance.Help.Render("No files selected!")
			} else {
				lines, err := clipboard.CopyFilesToClipboard(a.Search.Selected)
				if err != nil {
					a.StatusMsg = view.StylesInstance.Help.Render(fmt.Sprintf("Clipboard error: %v", err))
				} else {
					a.StatusMsg = clipboard.ClipboardOutputStatus(a.Search.Selected, lines)
				}
			}
		case "enter":
			a.StatusMsg = ""
			if fileItem, ok := a.UIList.SelectedItem().(search.FileItem); ok {
				a.Search.ToggleSelection(fileItem.Path)

				items := make([]list.Item, 0)
				for _, f := range a.Search.BuildItems() {
					items = append(items, f)
				}
				a.UIList.SetItems(items)
			}
		default:
			a.StatusMsg = ""
		}

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
	return view.RenderApp(a.Input, a.UIList, a.StatusMsg)
}

func main() {
	_, err := tea.NewProgram(NewApp()).Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
