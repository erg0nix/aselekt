# aselekt

Little TUI helper for yanking files to the clipboard.  
Built with Bubble Tea + Lipgloss + fd. Works anywhere Go works.

---

## Quick install

```bash
go install github.com/erg0nix/aselekt/cmd/aselekt@latest
````

Binary drops into `$(go env GOBIN)` (usually `~/go/bin`).
Add that to your `$PATH` if you haven’t already.

---

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

---

## Clipboard format

```
# path/to/file.go

…file contents…

# another/file.txt

…file contents…
```

Total line count is shown so you know you didn’t nuke your paste buffer with 2 MB of code by accident.

---

## Dev stuff

### Make targets

```bash
make run      # go run ./cmd/aselekt
make build    # build stripped binary (ldflags -s -w)
make install  # go install (same flags) -> $GOBIN
make clean    # remove local binary
```

## Requirements

* Go 1.20+
* [`fd`](https://github.com/sharkdp/fd) on `$PATH` for fast file listing
* Linux: `xclip` or `xsel` if clipboard isn’t working

