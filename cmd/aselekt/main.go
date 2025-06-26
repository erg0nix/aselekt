package main

import (
	"fmt"
	"os"

	"github.com/erg0nix/aselekt/internal/clipboard"
	"github.com/erg0nix/aselekt/internal/search"
	"github.com/erg0nix/aselekt/internal/view"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type App struct {
	Search     search.FileSearch
	Input      textinput.Model
	UIList     list.Model
	SearchMode search.SearchMode
	StatusMsg  string
}

func NewApp() (App, error) {
	fs, err := search.NewFileSearch()
	if err != nil {
		return App{}, fmt.Errorf("fd error: %w", err)
	}

	app := App{
		Search:     fs,
		Input:      view.InitTextInput(),
		UIList:     view.InitFileList(fs),
		SearchMode: search.Filename,
	}

	return app, nil
}

func (a *App) RefreshList() {
	items, err := a.Search.BuildItems(a.SearchMode)

	if err != nil {
		a.StatusMsg = view.RenderStatusMessage(err)
		a.UIList.SetItems(nil)
		return
	}

	a.UIList.SetItems(view.ToListItems(items))
}

func (a *App) HandleYank() {
	if len(a.Search.Selected) == 0 {
		a.StatusMsg = view.RenderNoSelectionMessage()
		return
	}

	lines, err := clipboard.CopyFilesToClipboard(a.Search.Selected)
	if err != nil {
		a.StatusMsg = view.RenderClipboardError(err)
		return
	}

	a.StatusMsg = view.RenderClipboardSuccess(a.Search.Selected, lines)
}

func (a *App) ToggleSearchMode() {
	a.SearchMode = a.SearchMode.Toggle()
	a.StatusMsg = view.RenderSearchModeSwitched(a.SearchMode)
	a.RefreshList()
}

func (a *App) ToggleSelection() {
	a.StatusMsg = ""
	if fileitem, ok := a.UIList.SelectedItem().(search.FileItem); ok {
		a.Search.ToggleSelection(fileitem.Path)
		a.RefreshList()
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
			a.HandleYank()
		case "ctrl+o":
			a.ToggleSearchMode()
		case "enter":
			a.ToggleSelection()
		default:
			a.StatusMsg = ""
		}

	}

	var c tea.Cmd
	a.Input, c = a.Input.Update(msg)
	cmds = append(cmds, c)

	if _, ok := msg.(tea.KeyMsg); ok {
		a.Search.Query = a.Input.Value()
		a.RefreshList()
	}

	a.UIList, c = a.UIList.Update(msg)
	cmds = append(cmds, c)

	return a, tea.Batch(cmds...)
}

func (a App) View() string {
	return view.RenderApp(a.Input, a.UIList, a.StatusMsg, a.SearchMode)
}

func main() {
	app, err := NewApp()
	if err != nil {
		fmt.Fprintln(os.Stderr, "startup error:", err)
		os.Exit(1)
	}
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "runtime error:", err)
		os.Exit(1)
	}
}
