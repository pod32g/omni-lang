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

**Version:** v0.5.1 (October 2025)  
**Status:** Active Development - Advanced Type System & Enhanced Language Features

### ✅ Implemented Features

- **Core Language:**
  - Variables (`let`, `var`)
  - Functions with type inference
  - Control flow (`if`, `for`, `while`)
  - Basic types (`int`, `float`, `string`, `bool`)
  - **Arrays** (`[]int`, `[]string`) with indexing and iteration
  - String concatenation with automatic type conversion
  - **String interpolation** with `${expression}` syntax
  - Unary operators (`-`, `!`)

- **Advanced Features:**
  - **Exception handling** with try-catch-finally blocks
  - **Type aliases** (`type UserID = int`)
  - **Union types** (`string | int | bool`)
  - **Optional types** (`int?`, `string?`)
  - **Advanced type system** with full type checking

- **Import System:**
  - Standard library imports (`import std.io as io`)
  - Local file imports (`import math_utils`)
  - Alias support

- **Standard Library:**
  - `std.io` - Input/output functions
  - `std.math` - Mathematical utilities
  - `std.string` - String manipulation helpers
  - `std.log` - Structured logging backed by `simple-logger`

- **Compiler:**
  - Lexer with detailed error reporting
  - Parser with error recovery
  - Type checker with helpful error messages
  - MIR (Mid-level IR) generation with assignment instructions
  - VM backend for execution
  - C backend for native compilation
  - Debug symbol generation and source mapping
  - Package creation and distribution tools

### 🐛 Recent Bug Fixes (v0.3.1)

- **Critical Infinite Loop Fixes**: Resolved infinite loops in nested for loops and range loops
- **Assignment Instruction Generation**: Fixed MIR builder to generate proper assignment instructions
- **Variable Mutability**: Corrected variable assignment handling for mutable variables
- **Loop Variable Updates**: Fixed increment statements (i++, j++) to generate proper assignments
- **Test Framework**: Resolved e2e test framework timeout issues and environment setup
- **Performance Tests**: Fixed performance benchmark code generation

### 🚧 Current Limitations

- **Parser:**
  - No semicolons required (but not supported)
  - Limited control flow constructs

- **Type System:**
  - No generics (planned)
  - No structs or enums
  - No maps
  - No function overloading
  - Arrays work with C backend only (VM support pending)

- **Standard Library:**
  - ~~Limited function set~~ ✅ **RESOLVED** - Comprehensive standard library with 100+ functions
  - ~~No file I/O~~ ✅ **RESOLVED** - Complete file I/O operations
  - ~~No OS interface~~ ✅ **RESOLVED** - Full OS interface with process, file, and system operations

### 🔄 In Development

- Enhanced error messages and diagnostics
- Improved testing infrastructure
- Performance optimizations
- Extended standard library

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
// ❌ Wrong - undefined identifier
prnt("Hello")

// ✅ Correct - use std.io
import std.io as io
io.println("Hello")
```

```omni
// ❌ Wrong - type error
let message = "Age: " + 25

// ✅ Correct - convert to string
let message = "Age: " + math.toString(25)
```

## Contributing

To contribute to the documentation:

1. **Report Issues:** Found an error or missing information? Open an issue!
2. **Improve Examples:** Better examples help everyone learn faster
3. **Add Missing Docs:** Help complete the API documentation
4. **Fix Typos:** Even small improvements matter

## Version History

- **v0.2.0** (Current) - Enhanced error messages, improved testing, expanded stdlib
- **v0.1.0** - Initial release with basic language features

## License

This documentation is part of the OmniLang project and is licensed under the MIT License.

---

**Happy coding with OmniLang!** 🚀
