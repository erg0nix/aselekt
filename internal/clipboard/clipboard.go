package clipboard

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"aselekt/internal/view"

	"golang.design/x/clipboard"
)

func CopyFilesToClipboard(paths []string) (int, error) {
	var b strings.Builder
	lines := 0

	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			return 0, fmt.Errorf("reading file %s: %w", p, err)
		}
		lines += strings.Count(string(data), "\n")
		fmt.Fprintf(&b, "# %s\n\n%s\n\n", p, data)
	}

	clipboardData := b.String()

	if runtime.GOOS == "linux" {
		if path, err := exec.LookPath("wl-copy"); err == nil {
			cmd := exec.Command(path)
			cmd.Stdin = strings.NewReader(clipboardData)
			if err := cmd.Run(); err == nil {
				return lines, nil
			}
		}

		if path, err := exec.LookPath("xclip"); err == nil {
			cmd := exec.Command(path, "-selection", "clipboard")
			cmd.Stdin = strings.NewReader(clipboardData)
			if err := cmd.Run(); err == nil {
				return lines, nil
			}
		}

		if path, err := exec.LookPath("xsel"); err == nil {
			cmd := exec.Command(path, "--clipboard", "--input")
			cmd.Stdin = strings.NewReader(clipboardData)
			if err := cmd.Run(); err == nil {
				return lines, nil
			}
		}
	}

	if err := clipboard.Init(); err != nil {
		return 0, fmt.Errorf("clipboard init: %w", err)
	}

	clipboard.Write(clipboard.FmtText, []byte(clipboardData))
	return lines, nil
}

func ClipboardOutputStatus(selected []string, lines int) string {
	var b strings.Builder

	b.WriteString(view.StylesInstance.Success.Render("\n✔ Copied to clipboard:"))
	b.WriteString("\n")

	for _, f := range selected {
		b.WriteString(fmt.Sprintf(
			"%s %s\n",
			view.StylesInstance.FileBullet.Render("•"),
			view.StylesInstance.FileBullet.Render(f),
		))
	}

	b.WriteString(fmt.Sprintf("\nTotal lines copied: %d\n", lines))
	return b.String()
}
