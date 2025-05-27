# --- config ----------------------------------------------------
BINARY      := aselekt           # final binary name
PKG         := ./cmd/aselekt     # main package path
BUILD_FLAGS := -ldflags="-s -w"  # strip symbols & debug info

# --- targets ---------------------------------------------------
.PHONY: run build install clean

run:
	go run $(PKG)

build:
	go build $(BUILD_FLAGS) -o $(BINARY) $(PKG)

install:
	go install $(BUILD_FLAGS) $(PKG)

clean:
	@rm -f $(BINARY)

