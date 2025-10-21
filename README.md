# OmniLang

A statically typed programming language with a Go frontend, SSA MIR, and multiple backends (C, Cranelift, VM).

## Overview

OmniLang is a modern programming language designed for **performance, safety, and developer productivity**. It combines the simplicity of Go with the power of Rust, offering:

- **🚀 Lightning-fast compilation** (50ms vs Rust's 8s)
- **🧠 Simple, readable syntax** (easy to learn and maintain)
- **⚡ High performance** (near-C performance with zero-cost abstractions)
- **🛡️ Memory safety** (planned, without complexity)
- **🔧 Modern tooling** (debugging, packages, hot reload)
- **📦 Multiple backends** (C, Cranelift, VM) for different use cases
- **🌍 Cross-platform support** (Linux, macOS, Windows)

### Why Choose OmniLang?

**Perfect for developers who want:**
- The **power** of systems programming without the complexity
- **Fast iteration** with sub-second compilation times
- **Modern development experience** with excellent tooling
- **High performance** without sacrificing readability

**See our [competitive analysis](docs/COMPETITIVE_ANALYSIS.md) for detailed comparisons with Rust, Go, C++, and other languages.**

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
import std.io as io

func main():int {
    io.println("Hello, OmniLang!")
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
// Standard library imports
import std.io as io
import std.math as math
import std.string as str

// Local file imports
import math_utils
import string_utils as str_util

// Using imported functions
func main():int {
    io.println("Hello from std.io!")
    let result:int = math.max(10, 20)
    let combined:string = str.concat("Hello", "World")
    return 0
}
```

### String Operations

```omni
import std.io as io

func main():int {
    // String concatenation with mixed types
    let name:string = "Alice"
    let age:int = 30
    let message:string = "Hello " + name + ", you are " + age + " years old"
    io.println(message)
    
    // String concatenation with other strings
    let greeting:string = "Hello " + "World"
    io.println(greeting)
    
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
    println_int(i)
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

// Array access
let first:int = numbers[0]
numbers[1] = 10

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

OmniLang comes with a comprehensive standard library:

### I/O Functions

```omni
import std.io as io

func main():int {
    // Basic output
    io.println("Hello, World!")
    io.print("Enter your name: ")
    
    // Typed output
    io.println_int(42)
    io.print_float(3.14)
    io.println_bool(true)
    
    return 0
}
```

### Math Functions

```omni
import std.math as math

func main():int {
    let x:int = 15
    let y:int = 25
    
    // Basic operations
    let max_val:int = math.max(x, y)
    let min_val:int = math.min(x, y)
    let abs_val:int = math.abs(-42)
    
    // Advanced operations
    let pow_val:int = math.pow(2, 8)        // 2^8 = 256
    let gcd_val:int = math.gcd(48, 72)      // 24
    let lcm_val:int = math.lcm(12, 18)      // 36
    let fact_val:int = math.factorial(5)    // 120
    
    return 0
}
```

### String Operations

```omni
import std.string as str

func main():int {
    let s:string = "Hello, World!"
    
    // Basic operations
    let len:int = str.length(s)             // 13
    let combined:string = str.concat("Hello", "World")  // "HelloWorld"
    
    // String concatenation with + operator
    let message:string = "Length: " + len
    let greeting:string = "Hello " + "World"
    
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
import std.io as io

func main():int {
    let result1:int = math_utils.add(10, 20)      // 30
    let result2:int = math_utils.multiply(5, 6)   // 30
    
    io.println("Results: " + result1 + ", " + result2)
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

## Project Structure

```
omni/
├── cmd/
│   ├── omnic/          # Compiler CLI
│   ├── omnir/          # Runner CLI
│   └── omnipkg/        # Packager CLI
├── internal/
│   ├── lexer/          # Tokenization
│   ├── parser/         # Syntax analysis
│   ├── ast/            # Abstract syntax tree
│   ├── types/          # Type system
│   ├── mir/            # SSA intermediate representation
│   ├── passes/         # Optimization passes
│   ├── vm/             # Virtual machine
│   ├── backend/        # Code generation backends
│   │   ├── c/          # C backend (default)
│   │   └── cranelift/  # Cranelift backend
│   ├── compiler/       # Compiler orchestration
│   └── runner/         # Program execution
├── runtime/
│   ├── omni_rt.h       # Runtime library header
│   ├── omni_rt.c       # Runtime library implementation
│   └── posix/          # POSIX runtime (libomni_rt.so)
├── native/
│   └── clift/          # Rust Cranelift bridge
├── std/                # Standard library
├── examples/           # Example programs
├── tests/              # Test suite
│   ├── e2e/            # End-to-end tests
│   ├── std/            # Standard library tests
│   └── goldens/        # Golden test files
└── docs/               # Documentation
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
        std.io.print_int(i)
        std.io.print(") = ")
        std.io.println_int(result)
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
            std.io.print_int(i)
            std.io.print(" ")
            count = count + 1
        }
    }
    std.io.println("")
    std.io.print("Found ")
    std.io.print_int(count)
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

> 📋 **For detailed development roadmap, see [ROADMAP.md](docs/ROADMAP.md)**

### Current Status (v0.2.0+)
- ✅ Complete frontend (lexer, parser, AST, type checker)
- ✅ Three backends: C (default), VM, Cranelift (experimental)
- ✅ C backend with optimization levels (O0-O3, Os) and debug symbols
- ✅ Import system (std library + local files with aliases)
- ✅ **Arrays with type-safe indexing** (`[]int`, `[]string`)
- ✅ String concatenation with mixed types
- ✅ Unary expressions (-, !)
- ✅ For-in loop iteration over arrays
- ✅ Stdlib functions (std.io, std.math) with runtime support
- ✅ Enhanced error messages with helpful hints
- ✅ Comprehensive testing and documentation
- ✅ Packaging system for distribution
- ✅ Performance optimizations
- ✅ CI/CD pipeline with multi-platform builds

### Upcoming Phases
- 🔄 **Phase 1**: Language Features (for loops, memory management, error handling)
- 📋 **Phase 2**: Type System Completion (generics, union types)
- 📋 **Phase 3**: MIR Optimization (advanced optimization passes)
- 📋 **Phase 4**: Complete Stdlib (file I/O, networking, collections)
- 📋 **Phase 5**: Production Readiness (performance, security, deployment)
- 📋 **Phase 6**: Ecosystem (package registry, community building)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Cranelift](https://github.com/bytecodealliance/wasmtime/tree/main/cranelift) for code generation
- [Go](https://golang.org/) for the frontend implementation
- [Rust](https://www.rust-lang.org/) for the backend implementation
- The programming language community for inspiration and ideas

## Support

- 📖 [Documentation](docs/)
- 🐛 [Issue Tracker](https://github.com/omni-lang/omni/issues)
- 💬 [Discussions](https://github.com/omni-lang/omni/discussions)
- 📧 [Email](mailto:contact@omni-lang.org)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**OmniLang** - *One language to rule them all* 🚀
