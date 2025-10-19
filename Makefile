SHELL := /bin/bash

GO := go
CARGO := cargo

.PHONY: all fmt lint test build bench clean gen

all: build

fmt:
	$(GO) fmt ./...
	$(GO) vet ./...

lint:
	$(GO) vet ./...

test:
	$(GO) test ./... -count=1

bench:
	$(GO) test ./tests/bench -bench=. -benchmem -count=3

build:
	$(GO) build -o bin/omnic ./cmd/omnic
	$(GO) build -o bin/omnir ./cmd/omnir

clean:
	rm -rf bin

gen:
	@echo "no codegen yet"
