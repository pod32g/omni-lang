<div align="center">
  <img src="assets/logo.png" alt="OmniLang Logo" width="150"/>

  # OmniLang Documentation

  Welcome to the OmniLang documentation! This directory contains comprehensive documentation for the OmniLang programming language and compiler.
</div>

## Table of Contents

- [Language Tour](spec/language-tour.md) - Complete language overview and features
- [Advanced Features](ADVANCED_FEATURES.md) - String interpolation, exception handling, and advanced types
- [Examples](EXAMPLES.md) - Comprehensive code examples and tutorials
- [Grammar Specification](spec/grammar.md) - Formal grammar definition
- [API Documentation](api/) - Complete API reference
- [Roadmap](ROADMAP.md) - Development roadmap and future plans
- [Architecture Decision Records](adr/) - Technical decisions and rationale

## Quick Start

If you're new to OmniLang, start here:

1. **Read the [Language Tour](spec/language-tour.md)** - Comprehensive overview of all language features
2. **Check out [Getting Started Examples](api/examples/getting-started.md)** - Simple examples to get you coding
3. **Browse the [API Documentation](api/)** - Detailed reference for all functions and modules

## Current Status

**Version:** v0.5.2-dev (April 2026)
**Status:** Active development - Go-foundations complete, backend/tooling hardening in progress

For the short actionable backlog, use [`../omni/docs/PENDING.md`](../omni/docs/PENDING.md). The roadmap documents are broader planning material and may include aspirational phases.

###  Implemented Features

- **Core Language:**
  - Variables (`let`, `var`)
  - Functions with type inference
  - Control flow (`if`, `for`, `while`, `defer`, `select`)
  - Basic types (`int`, `float`, `string`, `bool`)
  - Arrays, slices, maps, structs, and C-style enums
  - String concatenation with automatic type conversion
  - String interpolation with `${expression}` syntax
  - Unary operators (`-`, `!`)
  - Methods and structural interfaces
  - Channels and spawn-style concurrency

- **Advanced Features:**
  - Exception syntax with try-catch-finally plumbing
  - Type aliases (`type UserID = int`)
  - Union and optional type syntax
  - Early generic function/type-parameter infrastructure
  - Tail-call optimization coverage for tested VM and C backend paths

- **Import System:**
  - Standard library imports (`import std`, `import std.io as io`)
  - Local file imports (`import math_utils`)
  - Alias support

- **Standard Library:**
  - `std.io` - Input/output functions
  - `std.math` - Mathematical utilities
  - `std.string` - String manipulation, regex, encoding, and escaping helpers
  - `std.array`, `std.collections`, and `std.algorithms` - Tested collection helpers for supported element shapes
  - `std.os` and `std.file` - CLI args, environment, process IDs, and file helpers
  - `std.network` and `std.web` - Experimental network/web helpers with documented partial areas
  - `std.testing` / `std.test` - Assertions and test harness helpers

- **Compiler:**
  - Lexer with detailed error reporting
  - Parser with error recovery
  - Type checker with helpful error messages
  - SSA MIR generation with PHI nodes
  - VM backend and C backend for the main development workflow
  - Experimental Cranelift bridge
  - Debug symbol generation and source mapping
  - Package creation and distribution tools
  - JSON diagnostics endpoints for editor/tooling integration

###  Recent Focus

- Go-like language foundations across VM and C backend: methods, interfaces, defer, slices, concurrency, TCO, and select.
- Standard library runtime wiring for strings, arrays, collections, algorithms, network, web, OS, file, and testing helpers.
- More release automation: manifests, Docker packaging, diagnostics JSON, and CLI smoke coverage.

###  Current Limitations

- **Language/runtime:**
  - Real global mutable storage is not implemented; top-level `var` is rejected.
  - Panic/recover semantics and stack traces are not wired.
  - Proper generic monomorphization is incomplete.
  - Pattern matching, tagged enums, visibility rules, `const`, and `iota` remain future work.

- **Backends and tooling:**
  - Cranelift is still experimental; macOS object output is placeholder-only.
  - Cross-compilation and static/shared-library linking are not production-ready.
  - VS Code support is preview quality; formatter, full LSP, debugger, and workspace-wide references are pending.

- **Standard library:**
  - Some helpers are intentionally partial or stubbed; see [`../omni/std/IMPLEMENTATION_STATUS.md`](../omni/std/IMPLEMENTATION_STATUS.md).
  - No function overloading

###  In Development

- Cranelift backend completion and cross-platform object generation
- Documentation/status consolidation around v0.5.2-dev
- VS Code diagnostics and command integration
- More compiler optimization passes and benchmark baselines

## Documentation Structure

### Language Specification
- **[Language Tour](spec/language-tour.md)** - Complete language overview
- **[Grammar](spec/grammar.md)** - Formal grammar definition

### API Reference
- **[Language Reference](api/language-reference.md)** - Core language features
- **[Standard Library](api/stdlib/)** - Complete stdlib documentation
  - [std.io](api/stdlib/io.md) - Input/output utilities
  - [std.math](api/stdlib/math.md) - Numerical helpers
  - [std.string](api/stdlib/string.md) - String manipulation
  - [std.log](api/stdlib/log.md) - Structured logging functions and configuration
  - [std.os](api/stdlib/os.md) - Process, CLI argument, and environment helpers
- **[Examples](api/examples/)** - Code examples and tutorials

### Project Information
- **[Roadmap](ROADMAP.md)** - Development plans and timeline
- **[Architecture Decisions](adr/)** - Technical decisions and rationale

## Getting Help

### Common Issues

1. **"undefined identifier" errors:**
   - Check spelling and case
   - Ensure the identifier is declared before use
   - Use `import` statements for standard library functions

2. **Type errors:**
   - Use `math.toString()` to convert numbers to strings
   - Check that function parameters match expected types
   - Use explicit type annotations when needed

3. **Import errors:**
   - Use `import std.io as io` for standard library
   - Ensure local files exist and are in the same directory
   - Check file extensions (`.omni`)

### Example Fixes

```omni
//  Wrong - undefined identifier
prnt("Hello")

//  Correct - use std.io
import std.io as io
io.println("Hello")
```

```omni
//  Wrong - type error
let message = "Age: " + 25

//  Correct - convert to string
let message = "Age: " + math.toString(25)
```

## Contributing

To contribute to the documentation:

1. **Report Issues:** Found an error or missing information? Open an issue!
2. **Improve Examples:** Better examples help everyone learn faster
3. **Add Missing Docs:** Help complete the API documentation
4. **Fix Typos:** Even small improvements matter

## Version History

- **v0.5.2-dev** (Current) - Go-foundations complete; backend, tooling, and docs hardening
- **v0.5.1** - Advanced type-system and language feature pass
- **v0.5.0** - Standard library and integration-test expansion
- **v0.2.0** - Enhanced error messages, improved testing, expanded stdlib
- **v0.1.0** - Initial release with basic language features

## License

This documentation is part of the OmniLang project and is licensed under the MIT License.

---

**Happy coding with OmniLang!**
