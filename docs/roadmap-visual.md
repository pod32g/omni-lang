# OmniLang Roadmap Visual

## Current Snapshot (v0.5.1)
```
OmniLang – compiler toolchain status (experimental)

Frontend:   Lexer ✔  Parser ✔  Type checker ✔ (contextual diagnostics live)
MIR:        Builder ✔  PHI nodes ✔  Optimisation passes ↻ (basic only)
Backends:   C ✔  VM ✔  Cranelift experiment ↻ (Linux smoke tests passing)
Stdlib:     IO ✔  Math ✔  String ✔  OS/File ✔  Network ✔  Testing ✔  Dev ✔
Tooling:    omnic/omnir/omnipkg ✔  JSON diagnostics ✔  VS Code extension (preview)
Automation: Release manifest ✔  Docker artefact ✔  CLI smoke checks planned
```

## Phase 1: Backend Hardening & Release Confidence (2-3 weeks)
```
- Complete MIR→Cranelift lowering for std modules and logging
- Add cross-platform smoke tests for each backend in CI
- Validate release outputs (manifest checksums, Docker image sanity checks)
- Document repeatable release + rollback playbooks
```

## Phase 2: Language Foundations (3-4 weeks)
```
- Optional types (T?) and union types (int | string)
- Early generics prototype plus type aliases
- Structured error handling (Result<T,E>, try helpers)
- Ownership/borrowing rules required by the std library
```

## Phase 3: Tooling & IDE Experience (3 weeks)
```
- Ship the VS Code extension with JSON diagnostics wiring
- Prototype a lightweight language server (hover, completion, go-to-definition)
- Introduce omnifmt / omnilint entry points in the CLI
- Expand docs for watch/test flows and structured diagnostics
```

## Phase 4: Performance & Incremental Builds (2-3 weeks)
```
- Introduce incremental MIR caching for faster rebuilds
- Optimise hot MIR passes (dead-code elimination, SCCP refresh)
- Benchmark VM vs C backend, record baselines
- Optional telemetry for compiler timings
```

## Phase 5: Ecosystem Experiments (ongoing)
```
- Draft omni.toml manifest and dependency resolution sketches
- Prototype omnipkg publish with a local registry
- Build multi-module toy examples (CLI tool, tiny service)
- Gather community feedback before formal governance/support
```

## Timeline Overview
```
Phase 1   ███        (backend hardening)      2-3 weeks
Phase 2   ██████     (language foundations)   3-4 weeks
Phase 3   ████       (tooling & IDE)          3 weeks
Phase 4   ███        (performance)            2-3 weeks
Phase 5   █████████  (ecosystem experiments)  ongoing
```

## Success Metrics (remain aspirational)
```
Test coverage > 80% core modules (stretch 90%)
End-to-end backend smoke tests in CI
Release includes validated manifest + Docker image
Incremental build speedup: target 2x over full rebuild
```

## Legend
- ✔ Complete
- ↻ In progress
- Planned / exploratory

*For detailed context and history, see [ROADMAP.md](docs/ROADMAP.md) and [ADR-002](docs/adr/ADR-002-detailed-roadmap.md).*
