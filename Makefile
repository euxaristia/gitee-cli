BIN := gitee
PKG := ./cmd/gitee
PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin
GOCACHE ?= /tmp/go-build
LDFLAGS ?= -s -w -X github.com/euxaristia/gitee-cli/internal/cmd.commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo dev)

.PHONY: all build test install clean

all: build

build:
	GOCACHE=$(GOCACHE) go build -ldflags "$(LDFLAGS)" -o $(BIN) $(PKG)

test:
	GOCACHE=$(GOCACHE) go test ./...

install: build
	mkdir -p "$(BINDIR)"
	install -m 0755 $(BIN) "$(BINDIR)/$(BIN)"
	@echo "Installed $(BIN) to $(BINDIR)/$(BIN)"

clean:
	rm -f $(BIN)
