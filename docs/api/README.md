# OmniLang API Documentation

This directory contains comprehensive API documentation for the OmniLang programming language and its standard library.

## Table of Contents

- [Language Reference](language-reference.md) - Core language features and syntax
- [Advanced Types](advanced-types.md) - Type aliases, union types, and optional types
- [String Interpolation](string-interpolation.md) - String interpolation with expressions
- [Exception Handling](exception-handling.md) - Try-catch-finally exception handling
- [Standard Library](stdlib/) - Complete standard library documentation
  - [std.io](stdlib/io.md) - Input/output helpers
  - [std.math](stdlib/math.md) - Numerical utilities (incl. xorshift32 PRNG)
  - [std.string](stdlib/string.md) - String manipulation, splitting, replace, search, casing
  - [std.array](stdlib/array.md) - List operations on `array<int>` / `array<string>`
  - [std.algorithms](stdlib/algorithms.md) - Sorts, searches, aggregates, transforms, distance metrics
  - [std.collections](stdlib/collections.md) - Maps and other compound types
  - [std.os](stdlib/os.md) - Process arguments, environment, working directory, file I/O
  - [std.file](stdlib/file.md) - Low-level file handles, seek/tell, and metadata
  - [std.log](stdlib/log.md) - Structured logging functions and configuration
- [Compiler API](compiler-api.md) - Compiler internals and extension points
- [VM API](vm-api.md) - Virtual Machine internals
- [C Backend](c-backend.md) - Native code generation with C backend
- [Packaging](packaging.md) - Package creation and distribution
- [Examples](examples/) - Code examples and tutorials

## Quick Start

For a quick introduction to OmniLang, see the [Language Tour](../spec/language-tour.md).

For specific API details, browse the sections above or use the search functionality in your documentation viewer.

## Version Information

This documentation corresponds to OmniLang v0.5.2-dev (April 2026).
The Go-foundations slice (methods, structural interfaces, defer,
slices, `spawn` / channels / `select`, multi-return, TCO) is
complete on both backends. The standard library has been overhauled
to match: every function listed as `[IMPLEMENTED]` in
`omni/std/IMPLEMENTATION_STATUS.md` actually runs end-to-end with
e2e regression coverage. See [CHANGELOG.md](../../CHANGELOG.md) for
the full list.

## Contributing

To contribute to the documentation, please see the [Contributing Guide](../../CONTRIBUTING.md).
