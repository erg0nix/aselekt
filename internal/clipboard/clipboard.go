package clipboard

import (
	"fmt"
	"os"
	"strings"

	"golang.design/x/clipboard"
)

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
