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

### Async Functions
```omni
// Declare an async function - returns Promise<T>
async func fetch_data():int {
    return 42
}

// Use await to get the result from a Promise
func main():int {
    let result = await fetch_data()
    return result
}
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
println(i)
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

## Async/Await

### Async Functions
Async functions return `Promise<T>` and can be executed asynchronously.

```omni
// Declare an async function
async func fetch_number():int {
    return 42
}

async func fetch_string():string {
    return "hello"
}
```

### Await Expressions
Use `await` to wait for a Promise to resolve and get its value. **Note:** `await` can only be used inside `async` functions.

```omni
async func main():int {
    // Await a Promise<int> to get int
    let num = await fetch_number()
    
    // Await a Promise<string> to get string
    let str = await fetch_string()
    
    std.io.println("Number: " + num)
    std.io.println("String: " + str)
    
    return 0
}
```

### Async I/O Operations
The standard library provides async versions of I/O operations:

```omni
import std

async func main():int {
    // Async file operations
    let content = await std.os.read_file_async("file.txt")
    let written = await std.os.write_file_async("output.txt", "Hello")
    let appended = await std.os.append_file_async("output.txt", "\nWorld")
    
    // Async input
    let line = await std.io.read_line_async()
    
    return 0
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
std.io.println(42)
std.io.println(3.14)
std.io.println(true)
```

### Math
```omni
import std.math

let max_val:int = std.math.max(a, b)
let min_val:int = std.math.min(a, b)
let abs_val:int = std.math.abs(-10)
let pow_val:float = std.math.pow(2.0, 8.0)
let sqrt_val:float = std.math.sqrt(16.0)
let floor_val:float = std.math.floor(3.7)
let ceil_val:float = std.math.ceil(3.2)
let round_val:float = std.math.round(3.5)
let gcd_val:int = std.math.gcd(15, 25)
let lcm_val:int = std.math.lcm(12, 18)
let fact_val:int = std.math.factorial(5)
```

### String
```omni
import std.string

let len:int = std.string.length(s)
let upper:string = std.string.to_upper(s)
let lower:string = std.string.to_lower(s)
let concat:string = std.string.concat(a, b)
let substr:string = std.string.substring(s, 0, 5)
let char:char = std.string.char_at(s, 0)
let starts:bool = std.string.starts_with(s, "prefix")
let ends:bool = std.string.ends_with(s, "suffix")
let contains:bool = std.string.contains(s, "substring")
let index:int = std.string.index_of(s, "substring")
let trimmed:string = std.string.trim(s)
let equals:bool = std.string.equals(s1, s2)
let compare:int = std.string.compare(s1, s2)
```

### Array
```omni
import std.array

let len:int = std.array.length(arr)
let item:T = std.array.get(arr, 0)
std.array.set(arr, 0, value)
let new_arr:array<T> = std.array.append(arr, value)
```

### File I/O
```omni
import std.file

let handle:int = std.file.open("file.txt", "r")
let content:int = std.file.read(handle, buffer)
std.file.write(handle, "Hello")
std.file.close(handle)
let exists:bool = std.file.exists("file.txt")
let size:int = std.file.size("file.txt")
let pos:int = std.file.tell(handle)
std.file.seek(handle, 0)
```

### Type Conversion
```omni
import std

let num_str:string = std.int_to_string(42)
let float_str:string = std.float_to_string(3.14)
let bool_str:string = std.bool_to_string(true)
let str_to_int:int = std.string_to_int("123")
let str_to_float:float = std.string_to_float("3.14")
let str_to_bool:bool = std.string_to_bool("true")
```

### Logging
```omni
import std.log

std.log.info("Service starting")
std.log.warn("Retrying request")
if std.log.set_level("debug") {
    std.log.debug("Verbose diagnostics enabled")
}
std.log.error("Shutting down")
```

### Testing
```omni
import std.testing

var suite = std.testing.suite()
suite = std.testing.pass(suite, "initialises")
suite = std.testing.expect(suite, "math", 2 + 2 == 4, "math is hard")
suite = std.testing.equal_string(suite, "greeting", "hi", "hi")

let failed = std.testing.summary(suite)
if failed != 0 {
    std.testing.exit(suite)
}
```

### Developer Utilities
```omni
import std.dev

std.io.println("Waiting for change…")
let snapshot = std.dev.wait_for_change("src/main.omni", 250)
if snapshot.exists {
    std.io.println("New size: " + std.int_to_string(snapshot.size))
} else {
    std.io.println("File removed")
}
```

**Environment variables:** `LOG_LEVEL`, `LOG_FORMAT`, `LOG_OUTPUT`, `LOG_COLORIZE`, `LOG_TIME_FORMAT`, `LOG_ROTATE_*`.

### OS Utilities
```omni
import std

let arg_count = std.os.args_count()
if arg_count > 0 {
    let args = std.os.args()
    std.io.println("first arg: " + args[0])
}

let token = std.os.getenv("TOKEN")
if token == "" {
    std.os.setenv("TOKEN", "demo")
}
```

Flag helpers:
```omni
import std

func main():int {
    let env = std.os.get_flag("env", "dev")
    let verbose = std.os.has_flag("verbose")
    let first = std.os.positional_arg(0, "")
    std.io.println(env + " " + first)
    if verbose {
        std.io.println("running in verbose mode")
    }
    return 0
}
```

### Reading Input
```omni
import std

std.io.print("Enter value: ")
let line = std.io.read_line()
std.io.println("you typed: " + line)
```

### Bitwise Operations
```omni
let a:int = 0b1010  // 10 in binary
let b:int = 0b1100  // 12 in binary

let and_result:int = a & b     // 0b1000 = 8
let or_result:int = a | b      // 0b1110 = 14
let xor_result:int = a ^ b     // 0b0110 = 6
let not_result:int = ~a        // bitwise NOT
let left_shift:int = a << 2    // 0b101000 = 40
let right_shift:int = b >> 1   // 0b0110 = 6
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

### Machine-Readable Output
- `omnic --list-backends --json` exposes backend metadata (default, supported emits, experimental status).
- `omnic --list-emits --json` enumerates emit targets, extensions, and whether further linking is needed.
- `omnic --diagnostics-json` turns failing compilations into structured reports suitable for editors and CI bots.

```json
{
  "status": "ok",
  "backends": [
    {
      "name": "c",
      "description": "C code-generation backend (default)",
      "default": true,
      "emits": ["exe", "obj", "asm", "binary"]
    },
    {
      "name": "clift",
      "description": "Cranelift backend",
      "experimental": true,
      "emits": ["obj"],
      "notes": ["requires Rust toolchain for native bridge"]
    }
  ]
}
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
println(x)
println(y)
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
