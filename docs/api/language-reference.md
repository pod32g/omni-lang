# OmniLang Language Reference

This document provides a complete reference for the OmniLang programming language syntax and semantics.

## Table of Contents

1. [Lexical Structure](#lexical-structure)
2. [Types](#types)
3. [Expressions](#expressions)
4. [Statements](#statements)
5. [Declarations](#declarations)
6. [Functions](#functions)
7. [Modules](#modules)
8. [Error Handling](#error-handling)

## Lexical Structure

### Comments

```omni
// Single-line comment

/*
 * Multi-line comment
 * Can span multiple lines
 */

/// Documentation comment
```

### Identifiers

Identifiers must start with a letter or underscore and can contain letters, digits, and underscores.

```omni
// Valid identifiers
let variable_name = 42
let _private = "hidden"
let UPPER_CASE = 100
let camelCase = true

// Invalid identifiers
// let 123invalid = 42      // starts with number
// let if = 42              // reserved keyword
// let my-var = 42          // contains hyphen
```

### Keywords

```omni
// Reserved keywords
let var func struct enum import
if else for while break continue return
true false null
int long byte float double bool char string void
array map
```

### Literals

#### Integer Literals

```omni
let decimal = 42
let hex = 0x2A
let binary = 0b101010
let octal = 0o52
```

#### Float Literals

```omni
let pi = 3.14159
let scientific = 1.23e-4
```

#### String Literals

```omni
let single_line = "Hello, World!"
let multi_line = """
    This is a
    multi-line string
    """
```

#### Character Literals

```omni
let letter = 'A'
let newline = '\n'
let tab = '\t'
```

#### Boolean Literals

```omni
let truth = true
let falsity = false
```

## Types

### Primitive Types

| Type | Description | Size | Range |
|------|-------------|------|-------|
| `byte` | 8-bit unsigned integer | 1 byte | 0 to 255 |
| `int` | 32-bit signed integer | 4 bytes | -2,147,483,648 to 2,147,483,647 |
| `long` | 64-bit signed integer | 8 bytes | -9,223,372,036,854,775,808 to 9,223,372,036,854,775,807 |
| `float` | 32-bit floating point | 4 bytes | ~1.4e-45 to ~3.4e38 |
| `double` | 64-bit floating point | 8 bytes | ~4.9e-324 to ~1.8e308 |
| `bool` | Boolean | 1 byte | true or false |
| `char` | Unicode character | 4 bytes | Unicode code points |
| `string` | String of characters | Variable | UTF-8 encoded |
| `void` | No value | 0 bytes | N/A |

### Composite Types

#### Arrays

```omni
let numbers:array<int> = [1, 2, 3, 4, 5]
let empty_array:array<string> = []
let mixed_array:array<int> = [1, 2, 3]
```

#### Maps

```omni
let scores:map<string,int> = {
    "alice": 95
    "bob": 87
    "charlie": 92
}
let empty_map:map<int,string> = {}
```

### Type Inference

OmniLang can infer types in many cases:

```omni
let x = 42          // inferred as int
let y = 3.14        // inferred as float
let z = "hello"     // inferred as string
let w = true        // inferred as bool
```

## Expressions

### Arithmetic Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `+` | Addition | `a + b` |
| `-` | Subtraction | `a - b` |
| `*` | Multiplication | `a * b` |
| `/` | Division | `a / b` |
| `%` | Modulo | `a % b` |

### Comparison Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `==` | Equal | `a == b` |
| `!=` | Not equal | `a != b` |
| `<` | Less than | `a < b` |
| `<=` | Less than or equal | `a <= b` |
| `>` | Greater than | `a > b` |
| `>=` | Greater than or equal | `a >= b` |

### Logical Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `&&` | Logical AND | `a && b` |
| `||` | Logical OR | `a || b` |
| `!` | Logical NOT | `!a` |

### Unary Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `-` | Negation | `-x` |
| `+` | Unary plus | `+x` |
| `!` | Logical NOT | `!flag` |

### String Concatenation

The `+` operator can be used for string concatenation with automatic type conversion:

```omni
let greeting = "Hello " + "World"           // "Hello World"
let message = "Age: " + 25                  // "Age: 25"
let info = 42 + " items found"              // "42 items found"
```

## Statements

### Variable Declaration

```omni
// Immutable variable (default)
let x:int = 10

// Mutable variable
var y:int = 10
y = 20  // OK: reassignment allowed
```

### Expression Statement

```omni
x + y
print("Hello")
```

### Block Statement

```omni
{
    let x = 10
    let y = 20
    x + y
}
```

### If Statement

```omni
if x > 0 {
    print("Positive")
} else if x < 0 {
    print("Negative")
} else {
    print("Zero")
}
```

### For Loop

```omni
// Classic for loop
for i:int = 0; i < 10; i++ {
    print(i)
}

// Range-based for loop
for num in numbers {
    print(num)
}
```

### While Loop

```omni
while count < 5 {
    count = count + 1
}
```

### Break and Continue

```omni
for i:int = 0; i < 10; i++ {
    if i == 3 {
        continue  // Skip this iteration
    }
    if i == 7 {
        break     // Exit the loop
    }
    print(i)
}
```

### Return Statement

```omni
func add(a:int, b:int):int {
    return a + b
}

func no_return():void {
    print("No return value")
    // Implicit return
}
```

## Declarations

### Function Declaration

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

### Struct Declaration

```omni
struct Point {
    x:float
    y:float
}

// Struct instantiation
let p:Point = Point{
    x: 1.0
    y: 2.0
}
```

### Enum Declaration

```omni
enum Color {
    RED
    GREEN
    BLUE
}

// Enum usage
let color:Color = Color.RED
```

### Import Declaration

```omni
// Import with alias
import std.io as io
import math_utils

// Import without alias
import std.io
import std.math
```

## Functions

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
```

### Function Calls

```omni
let result:int = add(10, 20)
let max_val:int = std.math.max(a, b)
let length:int = str.length("hello")
```

### Method Calls

```omni
let numbers:array<int> = [1, 2, 3]
let len:int = numbers.length
let first:int = numbers[0]
```

## Modules

### Standard Library Imports

```omni
// Recommended: Import entire standard library
import std

// Or import specific modules
import std.io as io
import std.math as math
import std.string as str

func main():int {
    std.io.println("Hello from std.io!")
    let result:int = std.math.max(10, 20)
    let combined:string = std.string.concat("Hello", "World")
    return 0
}
```

### Local File Imports

```omni
// math_utils.omni
func add(a:int, b:int):int {
    return a + b
}

// main.omni
import math_utils
import string_utils as str_util

func main():int {
    let result:int = math_utils.add(10, 20)
    let combined:string = str_util.concat("Hello", "World")
    return result
}
```

## Error Handling

### Compile-time Errors

The compiler provides detailed error messages with suggestions:

```omni
// Error: undefined identifier "prnt"
prnt("Hello")  // hint: did you mean one of: print?

// Error: type mismatch
let x:int = "hello"  // hint: convert the expression to int or change the variable type to string

// Error: function declared void cannot return a value
func no_return():void {
    return 42  // hint: remove the expression or change the return type
}
```

### Runtime Errors

Runtime errors are handled by the VM and provide stack traces:

```omni
// Division by zero
let result:int = 10 / 0  // Runtime error: division by zero

// Array bounds
let arr:array<int> = [1, 2, 3]
let value:int = arr[10]  // Runtime error: index out of bounds
```

## Grammar

The complete grammar is defined in [grammar.md](../spec/grammar.md).

## Examples

For practical examples, see the [Examples](../examples/) directory and the [Language Tour](../spec/language-tour.md).
