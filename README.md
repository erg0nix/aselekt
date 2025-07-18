# aselekt

Little TUI helper for yanking files to the clipboard.  
Built with Bubble Tea + Lipgloss + fd. Works anywhere Go works.

# Requirements

aselekt depends on a few tools. Make sure these are installed before running:

## 🛠️ Required

  - Go 1.20+ – to build and run the program
  - fd – for fast file listing
  - ripgrep (rg) – for content-based search (optional but recommended)

## 🧠 Clipboard Support

  - macOS: Native clipboard supported via golang.design/x/clipboard
  - Linux: One of the following must be in your $PATH:
      - wl-copy (Wayland)
      - xclip or xsel (X11)

# 🧪 Quick install on macOS (via Homebrew)
```sh
brew install fd ripgrep
```
Go can be installed via:
```sh
brew install go
```

## Quick install

```bash
go install github.com/erg0nix/aselekt/cmd/aselekt@latest
````

Binary drops into `$(go env GOBIN)` (usually `~/go/bin`).
Add that to your `$PATH` if you haven’t already.

## Running

Inside your chosen directory:

```bash
aselekt
```

### Keys

| Key              | Action                                   |
| ---------------- | ---------------------------------------- |
| **↑ / ↓**        | Move cursor                              |
| **Enter**        | Toggle star / un-star file               |
| **Esc / Ctrl-C** | Quit (copies starred files to clipboard) |

Starred items float to the top of the list.
When you quit, you’ll get a green “✔ Copied to clipboard” summary.

## Clipboard format

```
# path/to/file.go

…file contents…

# another/file.txt

…file contents…
```

Total line count is shown so you know you didn’t nuke your paste buffer with 2 MB of code by accident.

## Dev stuff

### Make targets

```bash
make run      # go run ./cmd/aselekt
make build    # build stripped binary (ldflags -s -w)
make install  # go install (same flags) -> $GOBIN
make clean    # remove local binary
```
