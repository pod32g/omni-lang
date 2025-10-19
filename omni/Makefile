SHELL := /bin/bash

GO := go
CARGO := cargo
HAVE_CARGO := $(shell command -v $(CARGO) 2>/dev/null)

.PHONY: all fmt lint test build bench clean tools

all: build

fmt:
	$(GO) fmt ./...
	$(GO) vet ./...

lint:
	$(GO) vet ./...

# `go test` honors GOFLAGS from the environment for reproducibility.
test:
	$(GO) test ./... -count=1

bench:
	$(GO) test ./... -bench=. -run=^$

build:
	$(GO) build ./...
	@if [ -n "$(HAVE_CARGO)" ]; then \
		cd native/clift && $(CARGO) build; \
	else \
		echo "cargo not found; skipping native/clift build"; \
	fi

clean:
	$(GO) clean ./...
	@if [ -n "$(HAVE_CARGO)" ]; then \
		cd native/clift && $(CARGO) clean; \
	else \
		echo "cargo not found; skipping native/clift clean"; \
	fi
