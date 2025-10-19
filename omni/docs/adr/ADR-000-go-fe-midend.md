# ADR-000: Go Frontend and Midend

## Status
Accepted

## Context

We need a productive, maintainable implementation language for OmniLang's
frontend and midend components. The compiler must deliver a fast edit-compile-run
loop, strong tooling and predictable cross-platform builds.

## Decision

We will implement the lexer, parser, type checker, MIR builder and optimization
passes in Go. Go's standard tooling keeps onboarding friction low and aligns
with the team's experience. The language's simplicity makes it easy to share
core algorithms between contributors and enables efficient concurrency for
future compilation tasks.

## Consequences

- Compiler executables (`omnic`, `omnir`) will be Go binaries.
- The repository standardizes on `go fmt` and `go vet` for formatting and linting.
- Interoperability with the Cranelift backend will happen via cgo.
- Contributors should be comfortable working in Go; training material will be
  prioritized for new team members.
