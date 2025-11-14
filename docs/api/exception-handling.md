# Exception Handling API Reference

This document provides the complete API reference for OmniLang's exception handling feature.

## Table of Contents

1. [Overview](#overview)
2. [Syntax](#syntax)
3. [Basic Usage](#basic-usage)
4. [Advanced Usage](#advanced-usage)
5. [Exception Types](#exception-types)
6. [Examples](#examples)
7. [Best Practices](#best-practices)

## Overview

Exception handling in OmniLang provides a robust way to handle errors and exceptional conditions using try-catch-finally blocks. This feature allows for graceful error handling and cleanup operations.

## Syntax

### Try-Catch Block

```omni
try {
    // Code that might throw an exception
} catch (exception_variable) {
    // Handle the exception
}
```

### Try-Catch-Finally Block

```omni
try {
    // Code that might throw an exception
} catch (exception_variable) {
    // Handle the exception
} finally {
    // Cleanup code that always runs
}
```

### Multiple Catch Clauses

```omni
try {
    // Code that might throw an exception
} catch (exception_variable) {
    // Handle general exceptions
} catch (specific_variable: type) {
    // Handle specific type exceptions
} finally {
    // Cleanup code
}
```

### Throw Statement

```omni
throw "Exception message"
throw exception_expression
```

## Basic Usage

### Simple Try-Catch

```omni
import std.io as io

func risky_operation(x: int): int {
    if x < 0 {
        throw "Negative values not allowed"
    }
    return x * 2
}

func basic_exception_example(): int {
    try {
        let result: int = risky_operation(10)
        io.println("Result: " + result)
    } catch (e) {
        io.println("Caught exception: " + e)
    }
    
    return 0
}
```

### Exception Variables

```omni
func exception_variable_example(): int {
    try {
        let result: int = risky_operation(-5)
        io.println("Result: " + result)
    } catch (e: string) {
        io.println("Caught string exception: " + e)
    }
    
    return 0
}
```

### Try-Catch-Finally

```omni
func try_catch_finally_example(): int {
    try {
        let result: int = risky_operation(20)
        io.println("Result: " + result)
    } catch (e) {
        io.println("Caught exception: " + e)
    } finally {
        io.println("Finally block executed")
    }
    
    return 0
}
```

## Advanced Usage

### Multiple Catch Clauses

```omni
func multiple_catch_example(): int {
    try {
        let result: int = risky_operation(150)
        io.println("Result: " + result)
    } catch (e) {
        io.println("Caught general exception: " + e)
    } catch (specific: string) {
        io.println("Caught specific string exception: " + specific)
    } finally {
        io.println("Finally block executed")
    }
    
    return 0
}
```

### Exception Propagation

```omni
func propagate_exception(): int {
    try {
        let result: int = risky_operation(-10)
        return result
    } catch (e) {
        io.println("Handling exception: " + e)
        throw "Re-throwing exception"
    }
}
```

### Nested Exception Handling

```omni
func nested_exception_example(): int {
    try {
        try {
            let result: int = risky_operation(-5)
            io.println("Inner result: " + result)
        } catch (inner_e) {
            io.println("Inner exception: " + inner_e)
            throw "Outer exception"
        }
    } catch (outer_e) {
        io.println("Outer exception: " + outer_e)
    }
    
    return 0
}
```

## Exception Types

### String Exceptions

```omni
func string_exception_example(): int {
    try {
        throw "This is a string exception"
    } catch (e: string) {
        io.println("Caught string exception: " + e)
    }
    
    return 0
}
```

### Expression Exceptions

```omni
func expression_exception_example(): int {
    let error_code: int = 404
    let error_message: string = "Not found"
    
    try {
        throw error_message + " (Code: " + error_code + ")"
    } catch (e) {
        io.println("Caught expression exception: " + e)
    }
    
    return 0
}
```

### Function Call Exceptions

```omni
func get_error_message(): string {
    return "Error from function"
}

func function_exception_example(): int {
    try {
        throw get_error_message()
    } catch (e) {
        io.println("Caught function exception: " + e)
    }
    
    return 0
}
```

## Examples

### Complete Exception Handling Example

```omni
import std.io as io

func validate_input(value: int): int {
    if value < 0 {
        throw "Negative values not allowed"
    }
    if value > 100 {
        throw "Value too large"
    }
    if value == 0 {
        throw "Zero value not allowed"
    }
    return value * 2
}

func exception_handling_demo(): int {
    io.println("=== Exception Handling Demo ===")
    
    // Valid input
    try {
        let result: int = validate_input(10)
        io.println("Valid input result: " + result)
    } catch (e) {
        io.println("Unexpected exception: " + e)
    }
    
    // Negative input
    try {
        let result: int = validate_input(-5)
        io.println("Negative input result: " + result)
    } catch (e) {
        io.println("Caught exception: " + e)
    }
    
    // Large input
    try {
        let result: int = validate_input(150)
        io.println("Large input result: " + result)
    } catch (e) {
        io.println("Caught exception: " + e)
    }
    
    // Zero input
    try {
        let result: int = validate_input(0)
        io.println("Zero input result: " + result)
    } catch (e) {
        io.println("Caught exception: " + e)
    }
    
    io.println("=== Demo Complete ===")
    return 0
}
```

### Real-World Application Example

```omni
import std.io as io

func process_file(filename: string): string {
    if filename == "" {
        throw "Filename cannot be empty"
    }
    if filename.len() < 3 {
        throw "Filename too short"
    }
    if !filename.contains(".") {
        throw "Filename must have extension"
    }
    
    return "Processing file: " + filename
}

func safe_file_operation(filename: string): string {
    try {
        let result: string = process_file(filename)
        return result
    } catch (e) {
        return "Error processing file: " + e
    }
}

func real_world_example(): int {
    // Valid filename
    let result1: string = safe_file_operation("document.txt")
    io.println(result1)
    
    // Empty filename
    let result2: string = safe_file_operation("")
    io.println(result2)
    
    // Short filename
    let result3: string = safe_file_operation("a")
    io.println(result3)
    
    // Filename without extension
    let result4: string = safe_file_operation("document")
    io.println(result4)
    
    return 0
}
```

### Advanced Exception Handling Patterns

```omni
func advanced_patterns(): int {
    // Exception with cleanup
    try {
        io.println("Opening resource")
        let result: int = risky_operation(10)
        io.println("Result: " + result)
    } catch (e) {
        io.println("Exception occurred: " + e)
    } finally {
        io.println("Closing resource")
    }
    
    // Exception with specific handling
    try {
        let result: int = risky_operation(-5)
        io.println("Result: " + result)
    } catch (e: string) {
        if e.contains("Negative") {
            io.println("Handling negative value exception: " + e)
        } else {
            io.println("Handling other string exception: " + e)
        }
    }
    
    return 0
}
```

## Best Practices

### 1. Use Specific Exception Messages

```omni
//  Good - specific exception message
if value < 0 {
    throw "Negative values not allowed for age calculation"
}

//  Avoid - vague exception message
if value < 0 {
    throw "Invalid value"
}
```

### 2. Handle Exceptions Appropriately

```omni
//  Good - meaningful exception handling
try {
    let result: int = risky_operation(value)
    return result
} catch (e) {
    io.println("Operation failed: " + e)
    return -1  // Return error code
}

//  Avoid - ignoring exceptions
try {
    let result: int = risky_operation(value)
    return result
} catch (e) {
    // Silent failure - not recommended
}
```

### 3. Use Finally Blocks for Cleanup

```omni
//  Good - proper cleanup
try {
    io.println("Opening file")
    let result: string = process_file("data.txt")
    return result
} catch (e) {
    io.println("File processing failed: " + e)
    return ""
} finally {
    io.println("Closing file")
}

//  Avoid - missing cleanup
try {
    io.println("Opening file")
    let result: string = process_file("data.txt")
    return result
} catch (e) {
    io.println("File processing failed: " + e)
    return ""
}
// File not closed!
```

### 4. Don't Overuse Exceptions

```omni
//  Good - use exceptions for exceptional conditions
if value < 0 {
    throw "Negative values not allowed"
}

//  Avoid - using exceptions for normal flow control
try {
    if value < 0 {
        throw "Negative value"
    }
} catch (e) {
    // Handle negative value as normal case
}
```

### 5. Provide Context in Exception Messages

```omni
//  Good - provide context
if value < 0 {
    throw "Negative values not allowed (received: " + value + ")"
}

//  Avoid - no context
if value < 0 {
    throw "Invalid value"
}
```

## Implementation Details

### Compilation Process

Exception handling is compiled to a series of conditional jumps and cleanup operations:

1. **Lexical Analysis**: The lexer identifies try-catch-finally keywords
2. **Parsing**: The parser creates AST nodes for exception handling constructs
3. **Type Checking**: The type checker validates exception types and variables
4. **Code Generation**: The compiler generates conditional jump instructions

### Runtime Behavior

- **Exception Throwing**: Sets global exception state and jumps to catch block
- **Exception Catching**: Captures exception value and executes catch block
- **Finally Blocks**: Always executed regardless of exception occurrence
- **Exception Propagation**: Re-thrown exceptions continue up the call stack

### Performance Considerations

- **Minimal Overhead**: Exception handling has minimal performance impact
- **Fast Path**: Normal execution path is not affected by exception handling
- **Cleanup Efficiency**: Finally blocks are optimized for fast execution

## Conclusion

Exception handling in OmniLang provides a robust and efficient way to handle errors and exceptional conditions. With support for try-catch-finally blocks, multiple catch clauses, and exception propagation, it offers comprehensive error handling capabilities while maintaining excellent performance.

The feature integrates seamlessly with OmniLang's type system and provides clear error messages and debugging information.
