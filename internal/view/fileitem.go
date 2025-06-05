package view

import (
	"aselekt/internal/search"
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type FileItemView struct{ S Styles }

func (FileItemView) Height() int                             { return 1 }
func (FileItemView) Spacing() int                            { return 0 }
func (FileItemView) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d FileItemView) Render(w io.Writer, m list.Model, i int, it list.Item) {
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
