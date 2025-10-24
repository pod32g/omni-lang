# String Interpolation API Reference

This document provides the complete API reference for OmniLang's string interpolation feature.

## Table of Contents

1. [Overview](#overview)
2. [Syntax](#syntax)
3. [Basic Usage](#basic-usage)
4. [Advanced Usage](#advanced-usage)
5. [Type Support](#type-support)
6. [Examples](#examples)
7. [Best Practices](#best-practices)

## Overview

String interpolation allows you to embed expressions within string literals using the `${expression}` syntax. This feature provides a convenient and readable way to create dynamic strings.

## Syntax

```omni
"String with ${expression} interpolation"
```

### Basic Syntax

```omni
let variable: string = "Hello"
let interpolated: string = "Greeting: ${variable}"
```

### Complex Expressions

```omni
let x: int = 10
let y: int = 20
let result: string = "Sum: ${x + y}, Product: ${x * y}"
```

## Basic Usage

### Simple Variable Interpolation

```omni
import std.io as io

func basic_interpolation(): int {
    let name: string = "Alice"
    let age: int = 30
    
    // Basic interpolation
    let greeting: string = "Hello, ${name}!"
    let info: string = "Name: ${name}, Age: ${age}"
    
    io.println(greeting)    // "Hello, Alice!"
    io.println(info)        // "Name: Alice, Age: 30"
    
    return 0
}
```

### Multiple Interpolations

```omni
func multiple_interpolation(): int {
    let product: string = "Laptop"
    let price: float = 999.99
    let quantity: int = 5
    
    let message: string = "Product: ${product}, Price: $${price}, Quantity: ${quantity}"
    io.println(message)  // "Product: Laptop, Price: $999.99, Quantity: 5"
    
    return 0
}
```

## Advanced Usage

### Complex Expressions

```omni
func complex_expressions(): int {
    let x: int = 10
    let y: int = 20
    let z: float = 3.14
    
    // Mathematical expressions
    let math_result: string = "Sum: ${x + y}, Product: ${x * y}, Average: ${(x + y) / 2}"
    
    // Mixed types
    let mixed_result: string = "X: ${x}, Y: ${y}, Z: ${z}, Total: ${x + y + z}"
    
    io.println(math_result)  // "Sum: 30, Product: 200, Average: 15"
    io.println(mixed_result) // "X: 10, Y: 20, Z: 3.14, Total: 33.14"
    
    return 0
}
```

### Nested Interpolation

```omni
func nested_interpolation(): int {
    let user: string = "Alice"
    let count: int = 5
    
    // Nested interpolation
    let message: string = "User ${user} has ${count} items: ${"Item " + count}"
    
    io.println(message)  // "User Alice has 5 items: Item 5"
    
    return 0
}
```

### Function Calls in Interpolation

```omni
func get_user_name(): string {
    return "Alice"
}

func get_user_age(): int {
    return 30
}

func function_calls(): int {
    let message: string = "User: ${get_user_name()}, Age: ${get_user_age()}"
    
    io.println(message)  // "User: Alice, Age: 30"
    
    return 0
}
```

## Type Support

String interpolation supports all OmniLang types with automatic conversion to string:

### Supported Types

- **Strings**: Direct interpolation
- **Integers**: Automatic conversion to string
- **Floats**: Automatic conversion to string
- **Booleans**: Automatic conversion to string
- **Union Types**: Automatic conversion to string
- **Optional Types**: Automatic conversion to string

### Type Conversion Examples

```omni
func type_conversion_example(): int {
    let string_val: string = "Hello"
    let int_val: int = 42
    let float_val: float = 3.14
    let bool_val: bool = true
    
    // All types are automatically converted to strings
    let result: string = "String: ${string_val}, Int: ${int_val}, Float: ${float_val}, Bool: ${bool_val}"
    
    io.println(result)  // "String: Hello, Int: 42, Float: 3.14, Bool: true"
    
    return 0
}
```

### Union Type Support

```omni
func union_type_example(): int {
    type StringOrInt = string | int
    
    let value1: StringOrInt = "Hello"
    let value2: StringOrInt = 42
    
    let result: string = "Value1: ${value1}, Value2: ${value2}"
    
    io.println(result)  // "Value1: Hello, Value2: 42"
    
    return 0
}
```

### Optional Type Support

```omni
func optional_type_example(): int {
    let maybe_int: int? = 42
    let maybe_string: string? = "Hello"
    
    let result: string = "Optional int: ${maybe_int}, Optional string: ${maybe_string}"
    
    io.println(result)  // "Optional int: 42, Optional string: Hello"
    
    return 0
}
```

## Examples

### Complete String Interpolation Example

```omni
import std.io as io

func string_interpolation_demo(): int {
    // Basic variables
    let name: string = "Alice"
    let age: int = 30
    let height: float = 5.6
    let is_student: bool = true
    
    // Basic interpolation
    let greeting: string = "Hello, ${name}!"
    let info: string = "Name: ${name}, Age: ${age}, Height: ${height}"
    
    // Complex expressions
    let calculation: string = "Age + 10: ${age + 10}, Height * 2: ${height * 2}"
    
    // Multiple interpolations
    let summary: string = "User ${name} is ${age} years old, ${height} feet tall, and is a student: ${is_student}"
    
    io.println(greeting)        // "Hello, Alice!"
    io.println(info)            // "Name: Alice, Age: 30, Height: 5.6"
    io.println(calculation)     // "Age + 10: 40, Height * 2: 11.2"
    io.println(summary)         // "User Alice is 30 years old, 5.6 feet tall, and is a student: true"
    
    return 0
}
```

### Real-World Application Example

```omni
import std.io as io

func create_user_report(user_id: int, name: string, age: int, email: string): string {
    return """
    User Report
    ===========
    ID: ${user_id}
    Name: ${name}
    Age: ${age}
    Email: ${email}
    Status: Active
    Created: ${"2025-10-24"}
    """
}

func generate_product_summary(product_id: int, name: string, price: float, quantity: int): string {
    let total: float = price * quantity
    
    return "Product ${product_id}: ${name} - $${price} each, ${quantity} in stock, Total value: $${total}"
}

func real_world_example(): int {
    // User report
    let user_report: string = create_user_report(1, "Alice", 30, "alice@example.com")
    io.println(user_report)
    
    // Product summary
    let product_summary: string = generate_product_summary(101, "Laptop", 999.99, 5)
    io.println(product_summary)
    
    return 0
}
```

### Advanced String Interpolation Patterns

```omni
func advanced_patterns(): int {
    let items: array<string> = ["apple", "banana", "cherry"]
    let count: int = items.len()
    
    // Dynamic string building
    let list_header: string = "Shopping List (${count} items):"
    let list_items: string = "Items: ${items[0]}, ${items[1]}, ${items[2]}"
    
    // Conditional interpolation
    let status: string = count > 0 ? "active" : "empty"
    let status_message: string = "List status: ${status} (${count} items)"
    
    io.println(list_header)      // "Shopping List (3 items):"
    io.println(list_items)       // "Items: apple, banana, cherry"
    io.println(status_message)   // "List status: active (3 items)"
    
    return 0
}
```

## Best Practices

### 1. Keep Expressions Simple

```omni
// ✅ Good - simple expressions
let message: string = "Hello, ${name}!"

// ❌ Avoid - complex expressions
let message: string = "Hello, ${get_user_name() + " " + get_user_surname() + " (" + get_user_id() + ")"}!"
```

### 2. Use for Readability

```omni
// ✅ Good - improves readability
let greeting: string = "Welcome, ${user_name}!"

// ❌ Avoid - when simple concatenation is clearer
let greeting: string = "Welcome, " + user_name + "!"
```

### 3. Consider Performance

```omni
// ✅ Good - efficient interpolation
let message: string = "User: ${name}, Age: ${age}"

// ❌ Avoid - unnecessary function calls in interpolation
let message: string = "User: ${get_user_name()}, Age: ${get_user_age()}"
```

### 4. Use Consistent Patterns

```omni
// ✅ Good - consistent interpolation pattern
let user_info: string = "Name: ${name}, Email: ${email}, Age: ${age}"

// ❌ Avoid - mixing interpolation and concatenation
let user_info: string = "Name: " + name + ", Email: ${email}, Age: " + age
```

### 5. Handle Edge Cases

```omni
func handle_edge_cases(): int {
    let name: string = ""
    let age: int = 0
    
    // Handle empty strings
    let greeting: string = name != "" ? "Hello, ${name}!" : "Hello, Anonymous!"
    
    // Handle zero values
    let age_info: string = age > 0 ? "Age: ${age}" : "Age not specified"
    
    io.println(greeting)    // "Hello, Anonymous!"
    io.println(age_info)    // "Age not specified"
    
    return 0
}
```

## Implementation Details

### Compilation Process

String interpolation is compiled to a series of string concatenation operations:

1. **Lexical Analysis**: The lexer identifies interpolation tokens
2. **Parsing**: The parser creates AST nodes for interpolation expressions
3. **Type Checking**: The type checker validates expression types
4. **Code Generation**: The compiler generates string concatenation instructions

### Performance Considerations

- **Compile-time optimization**: Simple interpolations are optimized at compile time
- **Runtime efficiency**: Interpolation is converted to efficient string concatenation
- **Memory usage**: Minimal memory overhead for interpolation

### Error Handling

The string interpolation system provides clear error messages for:

- **Syntax errors**: Invalid interpolation syntax
- **Type errors**: Incompatible types in interpolation expressions
- **Undefined variables**: Variables not in scope
- **Function call errors**: Invalid function calls in interpolation

## Conclusion

String interpolation in OmniLang provides a powerful and expressive way to create dynamic strings. With support for all types, automatic conversion, and clear syntax, it makes string manipulation both readable and efficient.

The feature integrates seamlessly with OmniLang's type system, providing type safety and clear error messages while maintaining excellent performance.
