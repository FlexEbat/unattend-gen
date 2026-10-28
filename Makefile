.PHONY: build install install-user uninstall fmt fmt-check vet lint test gate

BIN      := unattend-gen
PREFIX   ?= /usr/local
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS  := -X github.com/FlexEbat/unattend-gen/internal/cli.version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/$(BIN)
	@echo
	@echo "Built ./$(BIN) ($(VERSION))."
	@echo "It is not in your PATH yet. Pick one:"
	@echo "  make install        # system-wide: $(PREFIX)/bin (uses sudo if needed)"
	@echo "  make install-user   # only for you: ~/.local/bin (no sudo)"
	@echo "  ./$(BIN) tui        # or run it right from this directory"

# System-wide install. Re-run with sudo only if the target is not writable.
install: build
	@if [ -w "$(DESTDIR)$(PREFIX)/bin" ] || [ -w "$(DESTDIR)$(PREFIX)" ]; then \
		install -Dm755 $(BIN) "$(DESTDIR)$(PREFIX)/bin/$(BIN)"; \
	else \
		sudo install -Dm755 $(BIN) "$(DESTDIR)$(PREFIX)/bin/$(BIN)"; \
	fi
	@echo "Installed to $(DESTDIR)$(PREFIX)/bin/$(BIN). Try: $(BIN) tui"

# Per-user install, no root needed.
install-user: build
	install -Dm755 $(BIN) "$$HOME/.local/bin/$(BIN)"
	@echo "Installed to $$HOME/.local/bin/$(BIN)."
	@case ":$$PATH:" in \
		*":$$HOME/.local/bin:"*) echo "Try: $(BIN) tui" ;; \
		*) echo; \
		   echo "$$HOME/.local/bin is not in your PATH. Add it with:"; \
		   echo "  echo 'export PATH=\"\$$HOME/.local/bin:\$$PATH\"' >> ~/.bashrc && source ~/.bashrc" ;; \
	esac

uninstall:
	rm -f "$(DESTDIR)$(PREFIX)/bin/$(BIN)" "$$HOME/.local/bin/$(BIN)"

fmt:
	gofmt -w .

fmt-check:
	test -z "$$(gofmt -l .)"

vet:
	go vet ./...

lint:
	golangci-lint run

test:
	go test ./... -race

gate: fmt-check vet lint test
