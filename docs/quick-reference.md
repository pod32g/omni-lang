# OmniLang Quick Reference

## Basic Syntax

### Comments
```omni
// Single-line comment
/* Multi-line comment */
/// Documentation comment
```

### Variables
```omni
let x:int = 10          // Immutable
var y:int = 20          // Mutable
let z = 30              // Type inferred
```

### Functions
```omni
func add(a:int, b:int):int {
    return a + b
}

func square(x:int):int => x * x
```

### Imports
```omni
import std                    // Import entire standard library (recommended)
import std.io as io           // Standard library with alias
import std.math               // Standard library without alias
import math_utils             // Local file import
import string_utils as str    // Local file with alias
```

### String Operations
```omni
let greeting:string = "Hello " + "World"        // String concatenation
let message:string = "Age: " + 30               // Mixed types
let info:string = 42 + " items"                 // Integer + String
```

### Unary Expressions
```omni
let x:int = 42
let negative:int = -x        // Negation
let positive:int = -(-x)     // Double negation

let flag:bool = true
let not_flag:bool = !flag    // Logical NOT
```

## Types

### Primitive Types
```omni
let byte_val:byte = 255
let int_val:int = 42
let long_val:long = 1000000
let float_val:float = 3.14
let double_val:double = 3.14159
let bool_val:bool = true
let char_val:char = 'A'
let string_val:string = "Hello"
```

### Composite Types
```omni
let numbers:array<int> = [1, 2, 3]
let scores:map<string,int> = {"alice": 95, "bob": 87}
```

### Custom Types
```omni
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

## Control Flow

### If Statements
```omni
if x > 0 {
    println("Positive")
} else if x < 0 {
    println("Negative")
} else {
    println("Zero")
}
```

### For Loops
```omni
// Classic
for i:int = 0; i < 10; i++ {
    println_int(i)
}

// Range
for item in items {
    println(item)
}
```

### While Loops
```omni
while condition {
    // do something
}
```

## Data Structures

### Arrays
```omni
let arr:array<int> = [1, 2, 3]
let first:int = arr[0]
arr[0] = 10
let len:int = arr.length
```

### Maps
```omni
let map:map<string,int> = {"key": 42}
let value:int = map["key"]
map["new_key"] = 100
let exists:bool = map.has("key")
```

### Structs
```omni
struct Person {
    name:string
    age:int
}

let person:Person = Person{
    name: "Alice"
    age: 30
}
```

### Enums
```omni
enum Status {
    PENDING
    RUNNING
    COMPLETED
}

let status:Status = Status.RUNNING
```

## Standard Library

### I/O
```omni
import std.io

std.io.print("Hello")
std.io.println(" World!")
std.io.println_int(42)
std.io.println_float(3.14)
std.io.println_bool(true)
```

### Math
```omni
import std.math

let max_val:int = std.math.max(a, b)
let min_val:int = std.math.min(a, b)
let abs_val:int = std.math.abs(-10)
let gcd_val:int = std.math.gcd(15, 25)
let pow_val:int = std.math.pow(2, 8)
```

### String
```omni
import std.string

let len:int = std.string.length(s)
let upper:string = std.string.to_upper(s)
let lower:string = std.string.to_lower(s)
let concat:string = std.string.concat(a, b)
let substr:string = std.string.substring(s, 0, 5)
```

### Array
```omni
import std.array

let len:int = std.array.length(arr)
let item:T = std.array.get(arr, 0)
std.array.set(arr, 0, value)
let new_arr:array<T> = std.array.append(arr, value)
```

### OS
```omni
import std.os

std.os.exit(0)
let env:string = std.os.getenv("PATH")
let cwd:string = std.os.getcwd()
let exists:bool = std.os.exists("file.txt")
let content:string = std.os.read_file("file.txt")
```

## Compiler Usage

### Basic Commands
```bash
# Run with VM
go run ./cmd/omnir program.omni

# Compile to MIR
go run ./cmd/omnic program.omni -backend vm -emit mir

# Compile to object
go run ./cmd/omnic program.omni -backend clift -emit obj -o program.o
```

### Options
```bash
-backend string    # vm|clift (default: vm)
-O string         # O0-O3 (default: O0)
-emit string      # mir|obj|asm (default: obj)
-dump string      # mir (dump intermediate representation)
-o string         # output file path
```

## Common Patterns

### Error Handling
```omni
func safe_divide(a:int, b:int):Result<int, string> {
    if b == 0 {
        return Error("Division by zero")
    }
    return Ok(a / b)
}

match result {
    Ok(value) => return value
    Error(msg) => {
        println("Error: " + msg)
        return 0
    }
}
```

### Pattern Matching
```omni
match x {
    0 => "Zero"
    1 => "One"
    _ => "Other"
}

match color {
    Color.RED => "Red"
    Color.GREEN => "Green"
    Color.BLUE => "Blue"
}
```

### Array Operations
```omni
// Sum array
let sum:int = 0
for num in numbers {
    sum = sum + num
}

// Find maximum
let max_val:int = numbers[0]
for num in numbers {
    if num > max_val {
        max_val = num
    }
}
```

### Map Operations
```omni
// Iterate over map
for key in map.keys() {
    let value:int = map[key]
    println(key + ": " + value)
}

// Check if key exists
if map.has("key") {
    let value:int = map["key"]
}
```

## Naming Conventions

- **Variables**: `camelCase` (e.g., `userName`, `isAuthenticated`)
- **Functions**: `camelCase` (e.g., `calculateTotal`, `isValid`)
- **Types**: `PascalCase` (e.g., `UserAccount`, `HttpStatus`)
- **Constants**: `UPPER_CASE` (e.g., `MAX_SIZE`, `API_URL`)
- **Files**: `snake_case` (e.g., `user_service.omni`)

## Common Gotchas

1. **Immutable by default**: Use `var` for mutable variables
2. **Type inference**: OmniLang can infer types in many cases
3. **Array bounds**: Check bounds before accessing array elements
4. **Map access**: Use `has()` to check if key exists before access
5. **Function returns**: All functions must have explicit return types
6. **Main function**: Must return `int` and be named `main`

## Performance Tips

1. Use appropriate primitive types (`byte` vs `int` vs `long`)
2. Prefer immutable variables when possible
3. Use efficient algorithms (binary search for sorted arrays)
4. Avoid unnecessary string concatenation in loops
5. Use `const` for values that never change

## Debugging

### Print Debugging
```omni
println("Debug: x = " + x)
println_int(x)
println_float(y)
```

### Assertions
```omni
import std

std.assert(x > 0, "x must be positive")
std.assert_eq(actual, expected, "Values don't match")
```

### Common Errors
- `undefined identifier`: Variable not declared
- `type mismatch`: Wrong type used
- `array index out of bounds`: Invalid array access
- `division by zero`: Attempting to divide by zero
- `unreachable code`: Code after return statement
