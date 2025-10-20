# Getting Started Examples

This document provides simple examples to help you get started with OmniLang.

## Hello World

The traditional first program:

```omni
// hello_world.omni
import std.io as io

func main():int {
    io.println("Hello, World!")
    return 0
}
```

**Run with:**
```bash
./bin/omnir hello_world.omni
```

**Expected output:**
```
Hello, World!
```

## Variables and Types

Demonstrates variable declaration and type inference:

```omni
// variables.omni
import std.io as io

func main():int {
    // Explicit type declaration
    let name:string = "Alice"
    let age:int = 30
    let height:float = 5.6
    let is_student:bool = false
    
    // Type inference
    let city = "New York"        // inferred as string
    let count = 42               // inferred as int
    let pi = 3.14159            // inferred as float
    let active = true            // inferred as bool
    
    io.println("=== Variable Examples ===")
    io.println("Name: " + name)
    io.println("Age: " + age)
    io.println("Height: " + height)
    io.println("Is student: " + is_student)
    io.println("City: " + city)
    io.println("Count: " + count)
    io.println("Pi: " + pi)
    io.println("Active: " + active)
    
    return 0
}
```

**Expected output:**
```
=== Variable Examples ===
Name: Alice
Age: 30
Height: 5.6
Is student: false
City: New York
Count: 42
Pi: 3.14159
Active: true
```

## Basic Functions

Shows function declaration and calling:

```omni
// functions.omni
import std.io as io

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
    io.println("Hello from function!")
}

func main():int {
    io.println("=== Function Examples ===")
    
    let x:int = 10
    let y:int = 20
    
    let sum:int = add(x, y)
    let product:int = multiply(x, y)
    let answer:int = get_answer()
    
    io.println("x = " + x)
    io.println("y = " + y)
    io.println("x + y = " + sum)
    io.println("x * y = " + product)
    io.println("Answer = " + answer)
    
    print_hello()
    
    return 0
}
```

**Expected output:**
```
=== Function Examples ===
x = 10
y = 20
x + y = 30
x * y = 200
Answer = 42
Hello from function!
```

## Control Flow

Demonstrates if statements and loops:

```omni
// control_flow.omni
import std.io as io

func main():int {
    io.println("=== Control Flow Examples ===")
    
    // If statement
    let x:int = 15
    if x > 10 {
        io.println("x is greater than 10")
    } else if x < 5 {
        io.println("x is less than 5")
    } else {
        io.println("x is between 5 and 10")
    }
    
    // For loop
    io.println("Counting from 1 to 5:")
    for i:int = 1; i <= 5; i++ {
        io.println("Count: " + i)
    }
    
    // While loop
    io.println("Counting down from 3:")
    let count:int = 3
    while count > 0 {
        io.println("Countdown: " + count)
        count = count - 1
    }
    
    return 0
}
```

**Expected output:**
```
=== Control Flow Examples ===
x is greater than 10
Counting from 1 to 5:
Count: 1
Count: 2
Count: 3
Count: 4
Count: 5
Counting down from 3:
Countdown: 3
Countdown: 2
Countdown: 1
```

## String Operations

Shows string manipulation and concatenation:

```omni
// strings.omni
import std.io as io
import std.string as str

func main():int {
    io.println("=== String Examples ===")
    
    let greeting:string = "Hello"
    let name:string = "World"
    let age:int = 42
    
    // String concatenation
    let message:string = greeting + ", " + name + "!"
    io.println("Message: " + message)
    
    // Mixed type concatenation
    let info:string = "Age: " + age + " years"
    io.println("Info: " + info)
    
    // String functions
    let text:string = "  OmniLang  "
    io.println("Original: '" + text + "'")
    io.println("Length: " + str.length(text))
    io.println("Trimmed: '" + str.trim(text) + "'")
    io.println("Uppercase: '" + str.to_upper(text) + "'")
    io.println("Lowercase: '" + str.to_lower(text) + "'")
    
    // String searching
    let search_text:string = "Hello, World!"
    io.println("Text: " + search_text)
    io.println("Contains 'World': " + str.contains(search_text, "World"))
    io.println("Starts with 'Hello': " + str.starts_with(search_text, "Hello"))
    io.println("Index of 'World': " + str.index_of(search_text, "World"))
    
    return 0
}
```

**Expected output:**
```
=== String Examples ===
Message: Hello, World!
Info: Age: 42 years
Original: '  OmniLang  '
Length: 12
Trimmed: 'OmniLang'
Uppercase: '  OMNILANG  '
Lowercase: '  omnilang  '
Text: Hello, World!
Contains 'World': true
Starts with 'Hello': true
Index of 'World': 7
```

## Math Operations

Demonstrates mathematical functions:

```omni
// math.omni
import std.io as io
import std.math as math

func main():int {
    io.println("=== Math Examples ===")
    
    let a:int = 15
    let b:int = 25
    
    // Basic operations
    io.println("a = " + math.toString(a))
    io.println("b = " + math.toString(b))
    io.println("a + b = " + math.toString(a + b))
    io.println("a - b = " + math.toString(a - b))
    io.println("a * b = " + math.toString(a * b))
    io.println("a / b = " + math.toString(a / b))
    
    // Math functions
    io.println("max(a, b) = " + math.toString(math.max(a, b)))
    io.println("min(a, b) = " + math.toString(math.min(a, b)))
    io.println("abs(-a) = " + math.toString(math.abs(-a)))
    io.println("sqrt(64) = " + math.toString(math.sqrt(64)))
    
    // Prime checking
    let numbers:array<int> = [2, 3, 4, 5, 6, 7, 8, 9, 10, 11]
    io.println("Prime numbers:")
    for num in numbers {
        if math.is_prime(num) {
            io.println(math.toString(num) + " is prime")
        }
    }
    
    return 0
}
```

**Expected output:**
```
=== Math Examples ===
a = 15
b = 25
a + b = 40
a - b = -10
a * b = 375
a / b = 0
max(a, b) = 25
min(a, b) = 15
abs(-a) = 15
sqrt(64) = 8
Prime numbers:
2 is prime
3 is prime
5 is prime
7 is prime
11 is prime
```

## Next Steps

Now that you understand the basics, you can explore:

- [Language Features](language-features.md) - More advanced language features
- [Standard Library](stdlib-examples.md) - Comprehensive standard library usage
- [Advanced Examples](advanced.md) - Complex examples and patterns

## Common Issues

### Type Errors
Make sure variable types match their usage:
```omni
let x:int = 42
let y:string = "hello"
// let result = x + y  // Error: cannot add int and string
let result = math.toString(x) + y  // OK: convert int to string first
```

### Function Return Types
Functions without explicit return types are inferred as `void`:
```omni
func add(a:int, b:int) {  // Inferred as void
    return a + b  // Error: void function cannot return value
}

func add(a:int, b:int):int {  // Explicit return type
    return a + b  // OK
}
```

### String Concatenation
Use the `+` operator for string concatenation:
```omni
let name:string = "Alice"
let age:int = 30
let message:string = "Hello " + name + ", you are " + age + " years old"
```
