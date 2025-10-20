# OmniLang Examples

This directory contains comprehensive examples demonstrating various features of the OmniLang programming language.

## Table of Contents

- [Getting Started](getting-started.md) - Basic examples for beginners
- [Language Features](language-features.md) - Core language features
- [Standard Library](stdlib-examples.md) - Standard library usage
- [Advanced Examples](advanced.md) - Complex examples and patterns
- [Error Handling](error-handling.md) - Error handling patterns
- [Performance](performance.md) - Performance optimization examples

## Quick Start

If you're new to OmniLang, start with the [Getting Started](getting-started.md) examples.

## Running Examples

All examples can be run using the OmniLang runner:

```bash
# Run an example directly
./bin/omnir examples/hello_world.omni

# Compile to MIR
./bin/omnic examples/hello_world.omni -backend vm -emit mir

# Compile to native object (Linux only)
./bin/omnic examples/hello_world.omni -backend clift -emit obj -o hello.o
```

## Example Categories

### Basic Examples
- Hello World
- Variables and Types
- Functions
- Control Flow

### Intermediate Examples
- Data Structures
- String Manipulation
- Mathematical Operations
- File I/O

### Advanced Examples
- Complex Algorithms
- Design Patterns
- Performance Optimization
- Error Handling

## Contributing

To contribute new examples:

1. Create a new `.omni` file in the appropriate directory
2. Add comprehensive comments explaining the code
3. Include expected output in comments
4. Update this README with a brief description
5. Test the example to ensure it works correctly

## Requirements

- OmniLang v0.2.0 or later
- Go 1.21 or later (for building from source)
- Rust toolchain (for Cranelift backend)
