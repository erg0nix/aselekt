package view

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var success = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ade80")).Bold(true)
var fileStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#cbd5e1"))

func PrintClipboard(selected []string, lines int) {
	fmt.Println(success.Render("\n✔ Copied to clipboard:"))
	for _, f := range selected {
		fmt.Printf("%s %s\n", fileStyle.Render("•"), fileStyle.Render(f))
	}
	fmt.Printf("\nTotal lines copied: %d\n", lines)
}
