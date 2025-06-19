package view

import (
	"aselekt/internal/search"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"strings"
)

func RenderStatusMessage(err error) string {
	return StylesInstance.Help.Render(fmt.Sprintf("Search error: %v", err))
}

func RenderNoSelectionMessage() string {
	return StylesInstance.Help.Render("No files selected!")
}

func RenderClipboardError(err error) string {
	return StylesInstance.Help.Render(fmt.Sprintf("Clipboard error: %v", err))
}

func RenderClipboardSuccess(selected []string, lines int) string {
	var b strings.Builder

	b.WriteString(StylesInstance.Success.Render("\n✔ Copied to clipboard:"))
	b.WriteString("\n")

	for _, f := range selected {
		b.WriteString(fmt.Sprintf(
			"%s %s\n",
			StylesInstance.FileBullet.Render("•"),
			StylesInstance.FileBullet.Render(f),
		))
	}

	b.WriteString(fmt.Sprintf("\nTotal lines copied: %d\n", lines))
	return b.String()
}

func RenderSearchModeSwitched(mode search.SearchMode) string {
	return StylesInstance.Label.Render(fmt.Sprintf("\nSwitched to %s mode", mode))
}

func ToListItems(items []search.FileItem) []list.Item {
	uiItems := make([]list.Item, len(items))
	for i, f := range items {
		uiItems[i] = f
	}
	return uiItems
}
