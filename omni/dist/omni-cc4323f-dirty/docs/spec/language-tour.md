# OmniLang Language Tour

This document provides a comprehensive tour of the OmniLang programming language, covering all major features and syntax.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Basic Syntax](#basic-syntax)
3. [Types](#types)
4. [Variables](#variables)
5. [Functions](#functions)
6. [Control Flow](#control-flow)
7. [Data Structures](#data-structures)
8. [Standard Library](#standard-library)
9. [Advanced Features](#advanced-features)
10. [Examples](#examples)

## Getting Started

### Your First Program

```omni
func main():int {
    println("Hello, OmniLang!")
    return 0
}
```

This simple program demonstrates:
- Function declaration with explicit return type
- String literal
- Return statement
- Entry point convention (`main` function)

### Compiling and Running

```bash
# Run directly with VM
go run ./cmd/omnir hello.omni

# Compile to MIR
go run ./cmd/omnic hello.omni -backend vm -emit mir

# Compile to native object
go run ./cmd/omnic hello.omni -backend clift -emit obj -o hello.o
```

## Basic Syntax

### Comments

```omni
// Single-line comment

/*
 * Multi-line comment
 * Can span multiple lines
 */

/// Documentation comment
func documented_function():int {
    return 42
}
```

### Identifiers

```omni
// Valid identifiers
let variable_name = 42
let _private = "hidden"
let UPPER_CASE = 100
let camelCase = true

// Invalid identifiers (examples)
// let 123invalid = 42      // starts with number
// let if = 42              // reserved keyword
// let my-var = 42          // contains hyphen
```

### Literals

```omni
// Integer literals
let decimal = 42
let hex = 0x2A
let binary = 0b101010
let octal = 0o52

// Float literals
let pi = 3.14159
let scientific = 1.23e-4

// String literals
let single_line = "Hello, World!"
let multi_line = """
    This is a
    multi-line string
    """

// Character literals
let letter = 'A'
let newline = '\n'
let tab = '\t'

// Boolean literals
let truth = true
let falsity = false
```

## Types

### Primitive Types

```omni
// Integer types
let byte_val:byte = 255
let int_val:int = 2147483647
let long_val:long = 9223372036854775807

// Floating-point types
let float_val:float = 3.14159
let double_val:double = 3.141592653589793

// Boolean type
let bool_val:bool = true

// Character type
let char_val:char = 'A'

// String type
let string_val:string = "Hello, World!"

// Void type (for functions that don't return)
func no_return():void {
    println("No return value")
}
```

### Type Inference

```omni
// OmniLang can infer types in many cases
let x = 42          // inferred as int
let y = 3.14        // inferred as float
let z = "hello"     // inferred as string
let w = true        // inferred as bool

// Explicit types when needed
let explicit_int:int = 42
let explicit_float:float = 3.14
```

### Composite Types

```omni
// Arrays
let numbers:array<int> = [1, 2, 3, 4, 5]
let empty_array:array<string> = []
let mixed_array:array<int> = [1, 2, 3]

// Maps
let scores:map<string,int> = {
    "alice": 95
    "bob": 87
    "charlie": 92
}
let empty_map:map<int,string> = {}
```

## Variables

### Immutable Variables (Default)

```omni
// Immutable variables cannot be reassigned
let x:int = 10
// x = 20  // Error: cannot assign to immutable variable

// But their contents can be modified if they're composite types
let numbers:array<int> = [1, 2, 3]
numbers[0] = 10  // OK: modifying array contents
```

### Mutable Variables

```omni
// Mutable variables can be reassigned
var y:int = 10
y = 20  // OK: reassignment allowed

var counter:int = 0
counter = counter + 1  // OK: incrementing
```

### Variable Scope

```omni
func scope_example():int {
    let outer:int = 10
    
    if true {
        let inner:int = 20
        // inner is only accessible within this block
        return outer + inner
    }
    
    // inner is not accessible here
    return outer
}
```

## Functions

### Basic Function Declaration

```omni
// Function with explicit return type
func add(a:int, b:int):int {
    return a + b
}

// Function with type inference
func multiply(a:int, b:int) {
    return a * b
}

// Function with no parameters
func get_answer():int {
    return 42
}

// Function with no return value
func print_hello():void {
    println("Hello!")
}
```

### Arrow Function Syntax

```omni
// Short function syntax
func square(x:int):int => x * x

// Equivalent to:
func square_long(x:int):int {
    return x * x
}
```

### Function Parameters

```omni
// Multiple parameters
func calculate(a:int, b:int, c:int):int {
    return a * b + c
}

// No parameters
func get_random():int {
    return 42  // placeholder
}

// Complex parameter types
func process_data(
    numbers:array<int>,
    config:map<string,string>
):array<int> {
    // Process the data
    return numbers
}
```

### Function Overloading

```omni
// OmniLang supports function overloading based on parameter types
func add(a:int, b:int):int {
    return a + b
}

func add(a:float, b:float):float {
    return a + b
}

func add(a:string, b:string):string {
    return a + b
}
```

## Control Flow

### If Statements

```omni
func if_example(x:int):string {
    if x > 0 {
        return "Positive"
    } else if x < 0 {
        return "Negative"
    } else {
        return "Zero"
    }
}

// Nested if statements
func nested_if(x:int, y:int):string {
    if x > 0 {
        if y > 0 {
            return "Both positive"
        } else {
            return "X positive, Y not"
        }
    } else {
        return "X not positive"
    }
}
```

### For Loops

```omni
// Classic for loop
func classic_for():int {
    let sum:int = 0
    for i:int = 0; i < 10; i++ {
        sum = sum + i
    }
    return sum
}

// Range-based for loop
func range_for():int {
    let numbers:array<int> = [1, 2, 3, 4, 5]
    let sum:int = 0
    for num in numbers {
        sum = sum + num
    }
    return sum
}

// For loop with step
func step_for():int {
    let sum:int = 0
    for i:int = 0; i < 10; i = i + 2 {
        sum = sum + i
    }
    return sum
}
```

### While Loops

```omni
func while_example():int {
    let count:int = 0
    while count < 5 {
        count = count + 1
    }
    return count
}

// Do-while equivalent
func do_while_example():int {
    let count:int = 0
    while true {
        count = count + 1
        if count >= 5 {
            break
        }
    }
    return count
}
```

### Break and Continue

```omni
func break_continue_example():int {
    let sum:int = 0
    for i:int = 0; i < 10; i++ {
        if i == 3 {
            continue  // Skip this iteration
        }
        if i == 7 {
            break     // Exit the loop
        }
        sum = sum + i
    }
    return sum
}
```

## Data Structures

### Structs

```omni
// Basic struct
struct Point {
    x:float
    y:float
}

// Struct with methods
struct Rectangle {
    width:float
    height:float
}

func area(r:Rectangle):float {
    return r.width * r.height
}

// Struct instantiation
func struct_example():float {
    let p:Point = Point{
        x: 1.0
        y: 2.0
    }
    
    let rect:Rectangle = Rectangle{
        width: 10.0
        height: 5.0
    }
    
    return area(rect)
}
```

### Enums

```omni
// Simple enum
enum Color {
    RED
    GREEN
    BLUE
}

// Enum with values
enum HttpStatus {
    OK = 200
    NOT_FOUND = 404
    SERVER_ERROR = 500
}

// Enum usage
func enum_example():string {
    let color:Color = Color.RED
    let status:HttpStatus = HttpStatus.OK
    
    if color == Color.RED {
        return "Red color"
    }
    
    return "Other color"
}
```

### Arrays

```omni
func array_example():int {
    // Array declaration and initialization
    let numbers:array<int> = [1, 2, 3, 4, 5]
    
    // Array access
    let first:int = numbers[0]
    let last:int = numbers[4]
    
    // Array modification
    numbers[0] = 10
    
    // Array length
    let len:int = numbers.length
    
    return first + last + len
}
```

### Maps

```omni
func map_example():int {
    // Map declaration and initialization
    let scores:map<string,int> = {
        "alice": 95
        "bob": 87
        "charlie": 92
    }
    
    // Map access
    let alice_score:int = scores["alice"]
    
    // Map modification
    scores["david"] = 88
    
    // Check if key exists
    if scores.has("eve") {
        return scores["eve"]
    }
    
    return alice_score
}
```

## Standard Library

### I/O Functions

```omni
import std.io

func io_example():void {
    // Print functions
    std.io.print("Hello")
    std.io.println(" World!")
    
    // Print with types
    std.io.println_int(42)
    std.io.println_float(3.14)
    std.io.println_bool(true)
}
```

### Math Functions

```omni
import std.math

func math_example():int {
    let a:int = 15
    let b:int = 25
    
    // Basic math functions
    let max_val:int = std.math.max(a, b)
    let min_val:int = std.math.min(a, b)
    let abs_val:int = std.math.abs(-10)
    
    // Advanced math functions
    let gcd_val:int = std.math.gcd(a, b)
    let lcm_val:int = std.math.lcm(a, b)
    let pow_val:int = std.math.pow(2, 8)
    let sqrt_val:int = std.math.sqrt(64)
    
    return max_val + min_val + abs_val
}
```

### String Functions

```omni
import std.string

func string_example():string {
    let s:string = "Hello, World!"
    
    // String manipulation
    let len:int = std.string.length(s)
    let upper:string = std.string.to_upper(s)
    let lower:string = std.string.to_lower(s)
    let trimmed:string = std.string.trim("  hello  ")
    
    // String searching
    let contains_hello:bool = std.string.contains(s, "Hello")
    let index:int = std.string.index_of(s, "World")
    
    // String operations
    let concat:string = std.string.concat("Hello", "World")
    let substr:string = std.string.substring(s, 0, 5)
    
    return concat
}
```

### Array Functions

```omni
import std.array

func array_example():array<int> {
    let numbers:array<int> = [1, 2, 3, 4, 5]
    
    // Array operations
    let len:int = std.array.length(numbers)
    let first:int = std.array.get(numbers, 0)
    
    // Array modification
    std.array.set(numbers, 0, 10)
    let new_array:array<int> = std.array.append(numbers, 6)
    let reversed:array<int> = std.array.reverse(new_array)
    
    return reversed
}
```

## Advanced Features

### Error Handling

```omni
// OmniLang uses Result types for error handling
func divide(a:int, b:int):Result<int, string> {
    if b == 0 {
        return Error("Division by zero")
    }
    return Ok(a / b)
}

func error_example():int {
    let result:Result<int, string> = divide(10, 2)
    
    match result {
        Ok(value) => return value
        Error(msg) => {
            println("Error: " + msg)
            return 0
        }
    }
}
```

### Pattern Matching

```omni
func pattern_example(x:int):string {
    match x {
        0 => "Zero"
        1 => "One"
        2 => "Two"
        _ => "Other"
    }
}

func enum_pattern(color:Color):string {
    match color {
        Color.RED => "Red"
        Color.GREEN => "Green"
        Color.BLUE => "Blue"
    }
}
```

### Generics (Planned)

```omni
// Generic functions (planned feature)
func identity<T>(x:T):T {
    return x
}

func max<T>(a:T, b:T):T where T: Comparable {
    if a > b {
        return a
    }
    return b
}

// Generic structs (planned feature)
struct Container<T> {
    value:T
}

func generic_example():int {
    let container:Container<int> = Container{value: 42}
    return container.value
}
```

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

### File Processing

```omni
import std.os

func process_file(filename:string):int {
    if !std.os.exists(filename) {
        std.io.println("File does not exist: " + filename)
        return 1
    }
    
    let content:string = std.os.read_file(filename)
    let lines:array<string> = std.string.split(content, "\n")
    
    std.io.print("File has ")
    std.io.print_int(std.array.length(lines))
    std.io.println(" lines")
    
    return 0
}

func main():int {
    return process_file("input.txt")
}
```

### Data Structures Example

```omni
struct Person {
    name:string
    age:int
    email:string
}

func create_person(name:string, age:int, email:string):Person {
    return Person{
        name: name
        age: age
        email: email
    }
}

func main():int {
    let people:array<Person> = [
        create_person("Alice", 30, "alice@example.com")
        create_person("Bob", 25, "bob@example.com")
        create_person("Charlie", 35, "charlie@example.com")
    ]
    
    for person in people {
        std.io.print("Name: ")
        std.io.println(person.name)
        std.io.print("Age: ")
        std.io.println_int(person.age)
        std.io.print("Email: ")
        std.io.println(person.email)
        std.io.println("---")
    }
    
    return 0
}
```

## Best Practices

### Naming Conventions

```omni
// Use descriptive names
let user_count:int = 0
let is_authenticated:bool = false

// Use camelCase for variables and functions
func calculate_total_price():float {
    return 0.0
}

// Use PascalCase for types
struct UserAccount {
    username:string
    balance:float
}

// Use UPPER_CASE for constants
let MAX_RETRY_COUNT:int = 3
let API_BASE_URL:string = "https://api.example.com"
```

### Error Handling

```omni
// Always handle errors explicitly
func safe_divide(a:int, b:int):Result<int, string> {
    if b == 0 {
        return Error("Cannot divide by zero")
    }
    return Ok(a / b)
}

// Use match for pattern matching
func handle_result(result:Result<int, string>):int {
    match result {
        Ok(value) => return value
        Error(msg) => {
            std.io.println("Error: " + msg)
            return 0
        }
    }
}
```

### Performance Tips

```omni
// Use appropriate data types
let small_number:byte = 255        // Use byte for small numbers
let large_number:long = 1000000   // Use long for large numbers

// Prefer immutable variables when possible
let config:map<string,string> = {
    "debug": "true"
    "timeout": "30"
}

// Use efficient algorithms
func efficient_search(arr:array<int>, target:int):int {
    // Binary search for sorted arrays
    // Linear search for unsorted arrays
    return -1
}
```

This concludes the OmniLang language tour. For more detailed information, see the [Grammar Specification](grammar.md) and [API Documentation](../api/).