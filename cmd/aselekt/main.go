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
	Search search.FileSearch
	Input  textinput.Model
	UIList list.Model
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
	return view.RenderApp(a.Input, a.UIList)
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

	lines, err := clipboard.CopyFilesToClipboard(app.Search.Selected)
	if err != nil {
		fmt.Fprintf(os.Stderr, "clipboard error: %v\n", err)
		return
	}

	clipboard.PrintClipboard(app.Search.Selected, lines)
}
