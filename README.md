<div align="center">
  <img src="docs/assets/logo.png" alt="OmniLang Logo" width="200"/>
  
  # OmniLang
  
  A toy statically typed programming language with a Go frontend, SSA MIR, and multiple backends (C, Cranelift, VM).
</div>

## Overview

OmniLang is built as a hobby language for experimenting with compiler pipelines and tooling. The repository includes:

- A Go-based frontend with lexer, parser, type checker, and SSA MIR builder.
- Backends for C, a VM interpreter, and an experimental Cranelift bridge.
- A standard library covering I/O, math, strings, collections, networking, OS helpers, testing utilities, and developer watch loops.
- Command-line tools: `omnic` (compiler), `omnir` (runner/test harness), and `omnipkg` (packager).
- Documentation, examples, and a VS Code extension used during development.
- **Agentic AI-assisted development**: parts of the pipeline configuration, docs, and tooling were authored with the help of agentic AI copilots. All generated changes are reviewed and tested before landing.

### Editor Support

- **VS Code Extension** (experimental): found in `vscode/omni`. Provides:
  - Syntax highlighting, bracket matching, and snippets for common language constructs.
  - Basic completion suggestions covering keywords, primitives, `std` modules, and detected structs/enums in the current file.
  - Hover hints that identify keywords, primitive types, and standard modules.
  - Inline diagnostics by invoking `omnic -emit mir` (configured via `omniLang.omnicPath`, defaults to `omnic` on `PATH`).
- **Current limitations**:
  - Diagnostics depend on a locally built `omnic`; no sandboxing and no partial / incremental analysis.
  - Completions are file-local only-no cross-file symbol indexing yet.
  - No formatter, code actions, go-to-definition, or debugging integration (planned for future LSP work).
  - Publishing flow is manual (run `npm install && npm run compile` inside `vscode/omni`, then `vsce package`).

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/omni-lang/omni.git
cd omni

# Build the compiler
make build

# Run tests
make test
```

### Hello World

Create a file `hello.omni`:

```omni
import std

func main():int {
    std.io.println("Hello, OmniLang!")
    return 0
}
```

Compile and run:

```bash
# Compile to native executable (C backend, default)
./bin/omnic hello.omni -o hello
./hello

# Run with VM (fastest compilation)
./bin/omnir hello.omni

# Compile to MIR only
./bin/omnic hello.omni -backend vm -emit mir

# Compile with optimization
./bin/omnic hello.omni -O O2 -o hello

# Compile with debug symbols
./bin/omnic hello.omni -debug -o hello
```

## Language Features

### Types

OmniLang supports a rich type system:

```omni
// Primitive types
let x:int = 42
let y:float = 3.14
let z:bool = true
let s:string = "Hello"
let c:char = 'A'

// Composite types
let numbers:array<int> = [1, 2, 3, 4, 5]
let scores:map<string,int> = {"alice": 95, "bob": 87}

// Custom types
struct Point {
    x:float
    y:float
}

enum Color {
    RED
    GREEN
    BLUE
}
```

### Variables

```omni
// Immutable variables (default)
let x:int = 10

// Mutable variables
var y:int = 20
y = 30

// Type inference
let z = 42  // inferred as int
```

### Imports

```omni
// Standard library imports (recommended)
import std

// Or import specific modules
import std.io as io
import std.math as math
import std.string as str

// Local file imports
import math_utils
import string_utils as str_util

// Using imported functions
func main():int {
    std.io.println("Hello from std.io!")
    let result:int = std.math.max(10, 20)
    let combined:string = std.string.concat("Hello", "World")
    return 0
}
```


### Unary Expressions

```omni
func main():int {
    let x:int = 42
    let negative:int = -x        // negation
    let positive:int = -(-x)     // double negation
    
    let flag:bool = true
    let not_flag:bool = !flag    // logical not
    
    return 0
}
```

### Functions

```omni
// Function with explicit return type
func add(a:int, b:int):int {
    return a + b
}

// Function with type inference
func multiply(a:int, b:int) {
    return a * b
}

// Arrow function syntax
func square(x:int):int => x * x

// Function with no return value
func print_hello() {
    println("Hello!")
}
```

### Control Flow

```omni
// If statements
if x > 0 {
    println("Positive")
} else if x < 0 {
    println("Negative")
} else {
    println("Zero")
}

// For loops (classic)
for i:int = 0; i < 10; i++ {
    println(i)
}

// For loops (range)
for item in items {
    println(item)
}

// While loops
while condition {
    // do something
}
```

### Structs and Enums

```omni
// Struct definition
struct Person {
    name:string
    age:int
    email:string
}

// Struct instantiation
let person:Person = Person{
    name: "Alice"
    age: 30
    email: "alice@example.com"
}

// Enum definition
enum Status {
    PENDING
    RUNNING
    COMPLETED
    FAILED
}

// Enum usage
let status:Status = Status.RUNNING
```

### Arrays and Maps

```omni
// Array literals
let numbers:array<int> = [1, 2, 3, 4, 5]
let empty:array<string> = []

// Array access and methods
let first:int = numbers[0]
numbers[1] = 10
let length:int = numbers.len()  // Method-style syntax

// Map literals
let scores:map<string,int> = {
    "alice": 95
    "bob": 87
    "charlie": 92
}

// Map access
let alice_score:int = scores["alice"]
scores["david"] = 88
```

## Standard Library

OmniLang ships with a comprehensive standard library backed by structured logging via [`simple-logger`](https://github.com/pod32g/simple-logger):

### Logging

```omni
import std

func main():int {
    std.log.info("Server starting")

    if std.log.set_level("debug") {
        std.log.debug("Debug logging enabled")
    } else {
        std.log.warn("Unrecognised log level")
    }

    std.log.error("Something went wrong")
    return 0
}
```

By default OmniLang emits human-friendly text logs to `stderr` at the `INFO` level. Customise behaviour before running any Omni tool with environment variables:

- `LOG_LEVEL` (`debug`, `info`, `warn`, `error`)
- `LOG_OUTPUT` (`stdout`, `stderr`, or a file path)
- `LOG_FORMAT` (`text` or `json`)
- `LOG_COLORIZE` (`true`/`false`)
- `LOG_ROTATE` (`true`/`false`) with `LOG_ROTATE_MAX_SIZE`, `LOG_ROTATE_MAX_AGE`, `LOG_ROTATE_MAX_BACKUPS`, `LOG_ROTATE_COMPRESS`
- `LOG_TIME_FORMAT` (Go layout string, e.g. `2006-01-02T15:04:05Z07:00`)

All Omni binaries (`omnic`, `omnir`, `omnipkg`) share this global logger. The `-verbose` flag temporarily raises the active level to `DEBUG`, making it easy to toggle detailed traces without changing environment settings.

### I/O Functions

```omni
import std

func main():int {
    // Basic output
    std.io.println("Hello, OmniLang!")
    std.io.print("Enter your name: ")
    
    // Primitive values print directly
    let count:int = 42
    let ratio:float = 3.14
    let ok:bool = true
    std.io.println(count)
    std.io.println(ratio)
    std.io.println(ok)
    
    return 0
}
```

### Math Functions

```omni
import std

func main():int {
    let x:int = 15
    let y:int = 25
    
    // Basic operations
    let max_val:int = std.math.max(x, y)
    let min_val:int = std.math.min(x, y)
    let abs_val:int = std.math.abs(-42)
    
    // Advanced operations
    let pow_val:float = std.math.pow(2.0, 8.0)        // 2^8 = 256.0
    let sqrt_val:float = std.math.sqrt(16.0)          // 4.0
    let floor_val:float = std.math.floor(3.7)         // 3.0
    let ceil_val:float = std.math.ceil(3.2)           // 4.0
    let round_val:float = std.math.round(3.5)         // 4.0
    let gcd_val:int = std.math.gcd(48, 72)            // 24
    let lcm_val:int = std.math.lcm(12, 18)            // 36
    let fact_val:int = std.math.factorial(5)          // 120
    
    return 0
}
```

### Bitwise Operations

```omni
import std

func main():int {
    let a:int = 0b1010  // 10 in binary
    let b:int = 0b1100  // 12 in binary
    
    // Bitwise operations
    let and_result:int = a & b     // 0b1000 = 8
    let or_result:int = a | b      // 0b1110 = 14
    let xor_result:int = a ^ b     // 0b0110 = 6
    let not_result:int = ~a        // bitwise NOT
    let left_shift:int = a << 2    // 0b101000 = 40
    let right_shift:int = b >> 1   // 0b0110 = 6
    
    std.io.println(and_result)
    std.io.println(or_result)
    std.io.println(xor_result)
    
    return 0
}
```

### Type Conversion

```omni
import std

func main():int {
    let num:int = 42
    let pi:float = 3.14159
    let flag:bool = true
    
    // Explicit type casting
    let int_to_float:float = (float)num
    let float_to_int:int = (int)pi
    
    // String conversion functions
    let num_str:string = std.int_to_string(num)        // "42"
    let float_str:string = std.float_to_string(pi)     // "3.14159"
    let bool_str:string = std.bool_to_string(flag)     // "true"
    
    // String to other types
    let str_to_int:int = std.string_to_int("123")      // 123
    let str_to_float:float = std.string_to_float("3.14") // 3.14
    let str_to_bool:bool = std.string_to_bool("true")  // true
    
    return 0
}
```

### File I/O Operations

```omni
import std

func main():int {
    let filename:string = "test.txt"
    let content:string = "Hello, OmniLang!"
    
    // Check if file exists
    if std.file.exists(filename) {
        std.io.println("File exists")
        let size:int = std.file.size(filename)
        std.io.println("File size: " + size)
    }
    
    // Write to file
    let handle:int = std.file.open(filename, "w")
    if handle >= 0 {
        let written:int = std.file.write(handle, content)
        std.file.close(handle)
        std.io.println("Written " + written + " bytes")
    }
    
    // Read from file
    let read_handle:int = std.file.open(filename, "r")
    if read_handle >= 0 {
        let buffer:string = "                "  // 16 character buffer
        let read_bytes:int = std.file.read(read_handle, buffer)
        std.file.close(read_handle)
        std.io.println("Read: " + buffer)
    }
    
    return 0
}
```

### Advanced Control Flow

```omni
import std

func main():int {
    let count:int = 0
    
    // While loop with break and continue
    while count < 10 {
        count = count + 1
        
        if count == 3 {
            continue  // Skip this iteration
        }
        
        if count == 8 {
            break     // Exit the loop
        }
        
        std.io.println(count)
    }
    
    // Block scope and variable shadowing
    let x:int = 10
    {
        let x:int = 20  // Shadows outer x
        std.io.println(x)  // Prints 20
    }
    std.io.println(x)  // Prints 10
    
    return 0
}
```

### First-Class Functions

```omni
import std

func add(a:int, b:int):int {
    return a + b
}

func main():int {
    // Function types
    let func_var:(int, int) -> int = add
    
    // Lambda expressions
    let square = |x:int| x * x
    let result1:int = square(5)  // 25
    
    // Closures with variable capture
    let multiplier:int = 3
    let multiply = |x:int| x * multiplier
    let result2:int = multiply(4)  // 12
    
    // Function calls through variables
    let result3:int = func_var(10, 20)  // 30
    
    std.io.println(result1)
    std.io.println(result2)
    std.io.println(result3)
    
    return 0
}
```

### Testing Framework

```omni
import std

func add(a:int, b:int):int {
    return a + b
}

func main():int {
    test.start("Math Functions")
    
    // Basic assertions
    assert.true(add(2, 3) == 5, "Addition should work")
    assert.false(add(1, 1) == 3, "Wrong addition should fail")
    
    // Equality assertions
    assert.eq(add(5, 5), 10, "5 + 5 should equal 10")
    assert.eq(std.string.length("hello"), 5, "String length should be 5")
    
    // Type conversion testing
    assert.eq(std.int_to_string(42), "42", "Int to string conversion")
    assert.eq(std.string_to_int("123"), 123, "String to int conversion")
    
    test.end()
    
    return 0
}
```

### String Operations

```omni
import std

func main():int {
    let s:string = "Hello, World!"
    
    // Basic operations
    let len:int = std.string.length(s)             // 13
    let combined:string = std.string.concat("Hello", "World")  // "HelloWorld"
    
    // String concatenation with + operator
    let message:string = "Length: " + len
    let greeting:string = "Hello " + "World"
    
    return 0
}
```

### Advanced Features

OmniLang v0.5.1 introduces powerful advanced features:

#### String Interpolation

```omni
import std

func main():int {
    let name:string = "Alice"
    let age:int = 30
    let pi:float = 3.14159
    
    // String interpolation with ${expression} syntax
    let greeting:string = "Hello, ${name}!"
    let info:string = "Name: ${name}, Age: ${age}, Pi: ${pi}"
    
    std.io.println(greeting)    // "Hello, Alice!"
    std.io.println(info)        // "Name: Alice, Age: 30, Pi: 3.14159"
    
    return 0
}
```

#### Exception Handling

```omni
import std

func risky_operation(x:int):int {
    if x < 0 {
        throw "Negative values not allowed"
    }
    return x * 2
}

func main():int {
    try {
        let result:int = risky_operation(10)
        std.io.println("Result: " + result)
    } catch (e) {
        std.io.println("Caught exception: " + e)
    } finally {
        std.io.println("Finally block executed")
    }
    
    return 0
}
```

#### Advanced Type System

```omni
import std

// Type aliases for better code readability
type UserID = int
type Name = string
type Email = string

// Union types for flexible data
type StringOrInt = string | int
type Number = int | float

// Optional types for nullable values
type OptionalInt = int?
type OptionalString = string?

func main():int {
    // Type aliases
    let user_id:UserID = 42
    let user_name:Name = "Alice"
    let user_email:Email = "alice@example.com"
    
    // Union types
    let value1:StringOrInt = "Hello"
    let value2:StringOrInt = 42
    let num1:Number = 100
    let num2:Number = 3.14
    
    // Optional types
    let maybe_int:OptionalInt = 42
    let maybe_string:OptionalString = "Hello"
    
    std.io.println("User: " + user_name + " (ID: " + user_id + ")")
    std.io.println("Values: " + value1 + ", " + value2)
    std.io.println("Numbers: " + num1 + ", " + num2)
    std.io.println("Optional: " + maybe_int + ", " + maybe_string)
    
    return 0
}
```

### Local File Imports

```omni
// math_utils.omni
func add(a:int, b:int):int {
    return a + b
}

func multiply(a:int, b:int):int {
    return a * b
}

// main.omni
import math_utils
import std

func main():int {
    let result1:int = math_utils.add(10, 20)      // 30
    let result2:int = math_utils.multiply(5, 6)   // 30
    
    std.io.println("Results: " + result1 + ", " + result2)
    return 0
}
```

## Compiler Usage

### Command Line Interface

```bash
# Compiler (omnic)
omnic [options] <file.omni>

Options:
  -backend string
        code generation backend (c|clift|vm) (default "c")
  -O string
        optimization level (O0-O3, Os) (default "O0")
  -emit string
        emission format (exe|mir|obj|asm) (default "exe")
  -dump string
        dump intermediate representation (mir)
  -o string
        output executable path
  -debug
        generate debug symbols
  -verbose
        enable verbose output

# Runner (omnir)
omnir <file.omni>

# Packager (omnipkg)
omnipkg [options]

Options:
  -version string
        version string for the package
  -platform string
        target platform (linux|macos|windows)
  -arch string
        target architecture (amd64|arm64)
```

#### Machine-readable CLI output
- `omnic --list-backends --json` and `omnic --list-emits --json` provide structured metadata (defaults, supported emits, notes) for editor integrations.
- `omnic --diagnostics-json` emits structured diagnostics on failure, including spans, hints, and highlighted context.
- Successful builds with `--json` return input path, backend, emit target, derived output, duration, and timestamp in a single object.

```bash
omnic hello.omni --list-backends --json | jq .
omnic failing.omni --diagnostics-json
```

### Backends

**C Backend** (Default):
- Generates C code and compiles with GCC/Clang
- Excellent portability and compatibility
- Supports all optimization levels and debug symbols
- Best for production use

**VM Backend**:
- Interprets MIR directly
- Fast compilation, slower execution
- Good for development, testing, and quick iteration

**Cranelift Backend** (Experimental):
- Compiles to native machine code via Cranelift
- Fast execution, moderate compilation speed
- Linux-only support currently

### Optimization Levels

- `O0`: No optimization (fastest compilation, largest binaries)
- `O1`: Basic optimization (remove redundant operations)
- `O2`: Standard optimization (balanced performance/size)
- `O3`: Aggressive optimization (maximum performance)
- `Os`: Size optimization (smallest binaries)

### Packaging and Distribution

OmniLang includes a packaging system for creating distributable binaries:

```bash
# Build and package for current platform
cd omni
make package

# Creates packages in omni/dist/:
# - omni-<version>-<platform>-<arch>.tar.gz (Linux/macOS)
# - omni-<version>-<platform>-<arch>.zip (Windows)
#
# Each package includes:
# - Compiled binaries (omnic, omnir, omnipkg)
# - Runtime library (libomni_rt.so)
# - Standard library modules
# - Examples and documentation
```

Running `make release` builds on that by copying artifacts into `releases/`, generating `checksums.txt`, templated release notes, a `release.json` manifest summarising every artifact (name, size, sha256), and a Docker image archive (`omni-<version>-docker.tar`) produced from `docker/Dockerfile`.

## Project Structure

```
omni/
 cmd/
    omnic/          # Compiler CLI
    omnir/          # Runner CLI
    omnipkg/        # Packager CLI
 internal/
    lexer/          # Tokenization
    parser/         # Syntax analysis
    ast/            # Abstract syntax tree
    types/          # Type system
    mir/            # SSA intermediate representation
    passes/         # Optimization passes
    vm/             # Virtual machine
    backend/        # Code generation backends
       c/          # C backend (default)
       cranelift/  # Cranelift backend
    compiler/       # Compiler orchestration
    runner/         # Program execution
 runtime/
    omni_rt.h       # Runtime library header
    omni_rt.c       # Runtime library implementation
    posix/          # POSIX runtime (libomni_rt.so)
 native/
    clift/          # Rust Cranelift bridge
 std/                # Standard library
 examples/           # Example programs
 tests/              # Test suite
    e2e/            # End-to-end tests
    std/            # Standard library tests
    goldens/        # Golden test files
 docs/               # Documentation
```

## Development

### Building

```bash
# Build everything (Go compiler + Rust backend + Runtime)
make build

# Build specific components
make build-go         # Build Go compiler only
make build-rust       # Build Rust Cranelift backend
make build-runtime    # Build C runtime library

# Clean build artifacts
make clean

# Format and lint code
make fmt
make lint
```

### Testing

```bash
# Run all tests
make test

# Run specific test suites
go test ./internal/lexer
go test ./internal/parser
go test ./tests/e2e
go test ./tests/std

# Run benchmarks
make bench
```

### Code Generation

```bash
# Generate AST golden tests
go run ./tools/gen_ast_goldens

# Generate MIR golden tests
go run ./tools/gen_mir_goldens

# Generate type checker golden tests
go run ./tools/gen_type_goldens
```

## Language Specification

### Grammar

The complete grammar is defined in `docs/spec/grammar.md`. Key productions:

```
program        := { declaration }
declaration    := importDecl | varDecl | funcDecl | structDecl | enumDecl
varDecl        := ("let" | "var") ident ":" type "=" expr
funcDecl       := "func" ident "(" params? ")" [ ":" type ] ( block | "=>" expr )
structDecl     := "struct" ident "{" { ident ":" type } "}"
enumDecl       := "enum" ident "{" { ident } "}"
```

### Type System

- **Primitive types**: `int`, `long`, `byte`, `float`, `double`, `bool`, `char`, `string`, `void`
- **Composite types**: `array<T>`, `map<K,V>`
- **User-defined types**: `struct`, `enum`
- **Type inference**: Automatic type deduction in many contexts
- **Generic types**: Planned for future releases

### Memory Management

- **Stack allocation**: Automatic for local variables
- **Heap allocation**: Via explicit `malloc`/`free` (planned)
- **Ownership system**: Planned for memory safety
- **Garbage collection**: Optional, planned for future

## Examples

### Fibonacci Sequence

```omni
func fibonacci(n:int):int {
    if n <= 1 {
        return n
    }
    return fibonacci(n - 1) + fibonacci(n - 2)
}

func main():int {
    for i:int = 0; i < 10; i++ {
        let result:int = fibonacci(i)
        std.io.print("fib(")
        std.io.print(i)
        std.io.print(") = ")
        std.io.println(result)
    }
    return 0
}
```

### Prime Number Sieve

```omni
func is_prime(n:int):bool {
    if n < 2 {
        return false
    }
    if n == 2 {
        return true
    }
    if n % 2 == 0 {
        return false
    }
    for i:int = 3; i * i <= n; i = i + 2 {
        if n % i == 0 {
            return false
        }
    }
    return true
}

func main():int {
    let count:int = 0
    for i:int = 2; i < 100; i++ {
        if is_prime(i) {
            std.io.print(i)
            std.io.print(" ")
            count = count + 1
        }
    }
    std.io.println("")
    std.io.print("Found ")
    std.io.print(count)
    std.io.println(" primes")
    return 0
}
```

### File I/O Example

```omni
import std.os

func main():int {
    let filename:string = "output.txt"
    let content:string = "Hello, OmniLang!"
    
    if std.os.write_file(filename, content) {
        std.io.println("File written successfully")
        
        let read_content:string = std.os.read_file(filename)
        std.io.println("Read: " + read_content)
    } else {
        std.io.println("Failed to write file")
    }
    
    return 0
}
```

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run the test suite
6. Submit a pull request

### Code Style

- Follow Go conventions for Go code
- Follow Rust conventions for Rust code
- Use meaningful variable and function names
- Add comments for complex logic
- Write tests for new features

## Roadmap

- See [docs/ROADMAP.md](docs/ROADMAP.md) for the full plan.
- Current focus: hardening backends, refreshing language foundations, improving tooling, and measuring performance.
- Exploratory items (self-hosting, package experiments, debugging integrations) are tracked as long-term ideas and remain unscheduled.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Cranelift](https://github.com/bytecodealliance/wasmtime/tree/main/cranelift) for code generation
- [Go](https://golang.org/) for the frontend implementation
- [Rust](https://www.rust-lang.org/) for the backend implementation
- The programming language community for inspiration and ideas

---

**OmniLang** - *One language to rule them all* 
