# OmniLang

A statically typed programming language with a Go frontend, SSA MIR, and Cranelift backend.

## Overview

OmniLang is a modern programming language designed for performance, safety, and developer productivity. It features:

- **Static typing** with type inference
- **Memory safety** through ownership and borrowing
- **High performance** via Cranelift code generation
- **Modern syntax** inspired by Rust and Go
- **Comprehensive standard library**
- **Cross-platform support** (Linux, macOS, Windows)

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
func main():int {
    println("Hello, OmniLang!")
    return 0
}
```

Compile and run:

```bash
# Run with VM
go run ./cmd/omnir hello.omni

# Compile to MIR
go run ./cmd/omnic hello.omni -backend vm -emit mir

# Compile to native object
go run ./cmd/omnic hello.omni -backend clift -emit obj -o hello.o
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
import std.io

func main():int {
    std.io.println("Hello, World!")
    std.io.print("Enter your name: ")
    // std.io.read_line() // TODO: implement
    return 0
}
```

### Math Functions

```omni
import std.math

func main():int {
    let x:int = 15
    let y:int = 25
    
    let max_val:int = std.math.max(x, y)
    let min_val:int = std.math.min(x, y)
    let gcd_val:int = std.math.gcd(x, y)
    
    std.io.println_int(max_val)
    return 0
}
```

### String Operations

```omni
import std.string

func main():int {
    let s:string = "Hello, World!"
    let len:int = std.string.length(s)
    let upper:string = std.string.to_upper(s)
    
    std.io.println_int(len)
    std.io.println(upper)
    return 0
}
```

### Array Operations

```omni
import std.array

func main():int {
    let numbers:array<int> = [1, 2, 3, 4, 5]
    let len:int = std.array.length(numbers)
    let first:int = std.array.get(numbers, 0)
    
    std.io.println_int(len)
    std.io.println_int(first)
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
        code generation backend (vm|clift) (default "vm")
  -O string
        optimization level (O0-O3) (default "O0")
  -emit string
        emission format (mir|obj|asm) (default obj)
  -dump string
        dump intermediate representation (mir)
  -o string
        output binary path

# Runner (omnir)
omnir <file.omni>
```

### Backends

**VM Backend** (Default):
- Interprets MIR directly
- Fast compilation, slower execution
- Good for development and testing

**Cranelift Backend**:
- Compiles to native machine code
- Slower compilation, faster execution
- Production-ready performance

### Optimization Levels

- `O0`: No optimization (fastest compilation)
- `O1`: Basic optimization
- `O2`: Standard optimization
- `O3`: Aggressive optimization

## Project Structure

```
omni/
├── cmd/
│   ├── omnic/          # Compiler CLI
│   └── omnir/          # Runner CLI
├── internal/
│   ├── lexer/          # Tokenization
│   ├── parser/         # Syntax analysis
│   ├── ast/            # Abstract syntax tree
│   ├── types/          # Type system
│   ├── mir/            # SSA intermediate representation
│   ├── passes/         # Optimization passes
│   ├── vm/             # Virtual machine
│   ├── backend/        # Code generation backends
│   └── runtime/        # Runtime library
├── native/
│   └── clift/          # Rust Cranelift bridge
├── std/                # Standard library
├── examples/           # Example programs
├── tests/              # Test suite
└── docs/               # Documentation
```

## Development

### Building

```bash
# Build everything
make build

# Build specific components
make build-go
make build-rust

# Clean build artifacts
make clean
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

### Version 1.0 (Current)
- ✅ Basic language features
- ✅ Compiler pipeline
- ✅ VM backend
- ✅ Cranelift backend (stub)
- ✅ Standard library (declarations)

### Version 1.1 (Planned)
- 🔄 Full Cranelift integration
- 🔄 Import system
- 🔄 Generic types
- 🔄 Error handling improvements

### Version 1.2 (Planned)
- 📋 Memory management
- 📋 Ownership system
- 📋 Pattern matching
- 📋 Concurrency primitives

### Version 2.0 (Future)
- 📋 Garbage collection
- 📋 Advanced optimizations
- 📋 Package manager
- 📋 IDE support

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

---

**OmniLang** - *One language to rule them all* 🚀
