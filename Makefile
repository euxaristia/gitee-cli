BIN := gitee
PKG := ./cmd/gitee
PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin
GOCACHE ?= /tmp/go-build

.PHONY: all build test install clean

all: build

build:
	GOCACHE=$(GOCACHE) go build -o $(BIN) $(PKG)

test:
	GOCACHE=$(GOCACHE) go test ./...

install: build
	mkdir -p "$(BINDIR)"
	install -m 0755 $(BIN) "$(BINDIR)/$(BIN)"
	@echo "Installed $(BIN) to $(BINDIR)/$(BIN)"

clean:
	rm -f $(BIN)
