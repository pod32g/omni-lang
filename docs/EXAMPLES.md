# OmniLang Examples

This document provides comprehensive examples demonstrating all the features of OmniLang v0.5.1.

## Table of Contents

1. [Basic Examples](#basic-examples)
2. [Logging](#logging)
3. [String Operations](#string-operations)
4. [Exception Handling](#exception-handling)
5. [Advanced Type System](#advanced-type-system)
6. [Real-World Examples](#real-world-examples)
7. [Complete Programs](#complete-programs)

## Basic Examples

### Hello World

```omni
import std.io as io

func main(): int {
    io.println("Hello, OmniLang!")
    return 0
}
```

### Variables and Types

```omni
import std.io as io

func main(): int {
    // Basic types
    let name: string = "Alice"
    let age: int = 30
    let height: float = 5.6
    let is_student: bool = true
    
    // Type inference
    let inferred_int = 42
    let inferred_float = 3.14
    let inferred_string = "Hello"
    let inferred_bool = true
    
    // Mutable variables
    var counter: int = 0
    counter = counter + 1
    
    io.println("Name: " + name)
    io.println("Age: " + age)
    io.println("Height: " + height)
    io.println("Is student: " + is_student)
    io.println("Counter: " + counter)
    
    return 0
}
```

### Functions

```omni
import std.io as io

func add(a: int, b: int): int {
    return a + b
}

func multiply(a: int, b: int): int {
    return a * b
}

func greet(name: string): string {
    return "Hello, " + name + "!"
}

func main(): int {
    let sum: int = add(10, 20)
    let product: int = multiply(5, 6)
    let greeting: string = greet("Alice")
    
    io.println("Sum: " + sum)
    io.println("Product: " + product)
    io.println(greeting)
    
    return 0
}
```

### Control Flow

```omni
import std.io as io

func control_flow_example(): int {
    let x: int = 10
    let y: int = 20
    
    // If statements
    if x > y {
        io.println("x is greater than y")
    } else if x < y {
        io.println("x is less than y")
    } else {
        io.println("x equals y")
    }
    
    // For loops
    for i: int = 0; i < 5; i++ {
        io.println("Loop iteration: " + i)
    }
    
    // While loops
    var count: int = 0
    while count < 3 {
        io.println("While loop: " + count)
        count = count + 1
    }
    
    return 0
}
```

## Logging

### Structured Logging

```omni
import std

func main(): int {
    std.log.info("Application starting")

    if std.log.set_level("debug") {
        std.log.debug("Debug mode enabled")
    } else {
        std.log.warn("Log level could not be changed")
    }

    std.log.error("Simulated failure path")
    return 0
}
```

**Environment overrides**

```bash
LOG_LEVEL=debug LOG_FORMAT=json ./bin/omnir logging_example.omni
```

Produces JSON output with debug statements included.

## String Operations

### String Concatenation

```omni
import std.io as io

func string_concatenation_example(): int {
    let name: string = "Alice"
    let age: int = 30
    let city: string = "New York"
    
    // String + String
    let greeting: string = "Hello " + "World"
    
    // String + Integer (automatic conversion)
    let message: string = "Hello " + name + ", you are " + age + " years old"
    
    // Integer + String
    let info: string = age + " years old, living in " + city
    
    io.println(greeting)    // "Hello World"
    io.println(message)     // "Hello Alice, you are 30 years old"
    io.println(info)        // "30 years old, living in New York"
    
    return 0
}
```

### String Interpolation

```omni
import std.io as io

func string_interpolation_example(): int {
    let name: string = "Alice"
    let age: int = 30
    let pi: float = 3.14159
    
    // Basic interpolation
    let greeting: string = "Hello, ${name}!"
    let info: string = "Name: ${name}, Age: ${age}, Pi: ${pi}"
    
    // Complex expressions
    let calculation: string = "Sum: ${age + 10}, Product: ${age * 2}"
    
    // Nested interpolation
    let complex: string = "User ${name} is ${age} years old and likes ${pi}"
    
    io.println(greeting)        // "Hello, Alice!"
    io.println(info)            // "Name: Alice, Age: 30, Pi: 3.14159"
    io.println(calculation)     // "Sum: 40, Product: 60"
    io.println(complex)         // "User Alice is 30 years old and likes 3.14159"
    
    return 0
}
```

### String Operations with Standard Library

```omni
import std.io as io
import std.string as str

func string_operations_example(): int {
    let text: string = "  Hello, World!  "
    
    // String manipulation
    let trimmed: string = str.trim(text)
    let upper: string = str.to_upper(text)
    let lower: string = str.to_lower(text)
    
    // String searching
    let contains_hello: bool = str.contains(text, "Hello")
    let index: int = str.index_of(text, "World")
    
    // String operations
    let concat: string = str.concat("Hello", "World")
    let substr: string = str.substring(text, 2, 7)
    
    io.println("Original: '" + text + "'")
    io.println("Trimmed: '" + trimmed + "'")
    io.println("Upper: '" + upper + "'")
    io.println("Lower: '" + lower + "'")
    io.println("Contains 'Hello': " + contains_hello)
    io.println("Index of 'World': " + index)
    io.println("Concatenated: '" + concat + "'")
    io.println("Substring: '" + substr + "'")
    
    return 0
}
```

## Exception Handling

### Basic Exception Handling

```omni
import std.io as io

func risky_operation(x: int): int {
    if x < 0 {
        throw "Negative values not allowed"
    }
    if x > 100 {
        throw "Value too large"
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

## Advanced Type System

### Type Aliases

```omni
import std.io as io

// Type aliases for better code readability
type UserID = int
type Name = string
type Email = string
type Age = int

func create_user(id: UserID, name: Name, email: Email, age: Age): void {
    io.println("Creating user: " + name + " (ID: " + id + ")")
    io.println("Email: " + email + ", Age: " + age)
}

func type_alias_example(): int {
    let user_id: UserID = 42
    let user_name: Name = "Alice"
    let user_email: Email = "alice@example.com"
    let user_age: Age = 30
    
    create_user(user_id, user_name, user_email, user_age)
    
    return 0
}
```

### Union Types

```omni
import std.io as io

// Union types allow values to be one of several types
type StringOrInt = string | int
type Number = int | float
type Status = string | int | bool

func process_value(value: StringOrInt): string {
    return "Value: " + value
}

func union_type_example(): int {
    let value1: StringOrInt = "Hello"
    let value2: StringOrInt = 42
    let num1: Number = 100
    let num2: Number = 3.14
    let status1: Status = "active"
    let status2: Status = 1
    let status3: Status = true
    
    io.println(process_value(value1))  // "Value: Hello"
    io.println(process_value(value2))  // "Value: 42"
    io.println("Number: " + num1)      // "Number: 100"
    io.println("Number: " + num2)      // "Number: 3.14"
    io.println("Status: " + status1)   // "Status: active"
    io.println("Status: " + status2)   // "Status: 1"
    io.println("Status: " + status3)   // "Status: true"
    
    return 0
}
```

### Optional Types

```omni
import std.io as io

// Optional types represent values that might be null
type OptionalInt = int?
type OptionalString = string?
type OptionalNumber = Number?

func find_user(id: int): OptionalString {
    if id > 0 {
        return "User found"
    }
    return null  // No user found
}

func optional_type_example(): int {
    let maybe_int: OptionalInt = 42
    let maybe_string: OptionalString = "Hello"
    let maybe_number: OptionalNumber = 3.14
    
    // Optional types are compatible with their base types
    let regular_int: int = maybe_int  // OK: int? is compatible with int
    let regular_string: string = maybe_string  // OK: string? is compatible with string
    
    io.println("Optional int: " + maybe_int)
    io.println("Optional string: " + maybe_string)
    io.println("Optional number: " + maybe_number)
    io.println("Regular int: " + regular_int)
    io.println("Regular string: " + regular_string)
    
    return 0
}
```

### Mixed Advanced Types

```omni
import std.io as io

// Complex type combinations
type UserID = int
type UserName = string
type UserStatus = string | int | bool
type OptionalUser = string?

func process_user(id: UserID, name: UserName, status: UserStatus): OptionalUser {
    if id > 0 {
        return "User processed: " + name + " with status: " + status
    }
    return null
}

func mixed_types_example(): int {
    let user_id: UserID = 42
    let user_name: UserName = "Alice"
    let user_status: UserStatus = "active"
    
    let result: OptionalUser = process_user(user_id, user_name, user_status)
    
    if result != null {
        io.println("Result: " + result)
    } else {
        io.println("No result")
    }
    
    return 0
}
```

## Real-World Examples

### User Management System

```omni
import std.io as io

// Type aliases for domain modeling
type UserID = int
type UserName = string
type UserEmail = string
type UserAge = int
type UserStatus = string | int | bool

// Optional types for nullable fields
type OptionalUser = string?
type OptionalAge = UserAge?

func create_user(id: UserID, name: UserName, email: UserEmail, age: OptionalAge): string {
    let age_info: string = "Unknown age"
    if age != null {
        age_info = "Age: " + age
    }
    
    return "User created: " + name + " (ID: " + id + "), Email: " + email + ", " + age_info
}

func update_user_status(id: UserID, status: UserStatus): string {
    return "User " + id + " status updated to: " + status
}

func user_management_example(): int {
    let user_id: UserID = 1
    let user_name: UserName = "Alice"
    let user_email: UserEmail = "alice@example.com"
    let user_age: OptionalAge = 30
    
    let creation_result: string = create_user(user_id, user_name, user_email, user_age)
    io.println(creation_result)
    
    let status_update: string = update_user_status(user_id, "active")
    io.println(status_update)
    
    return 0
}
```

### Calculator with Error Handling

```omni
import std.io as io

func safe_divide(a: int, b: int): int {
    if b == 0 {
        throw "Division by zero not allowed"
    }
    return a / b
}

func safe_multiply(a: int, b: int): int {
    if a > 1000 || b > 1000 {
        throw "Values too large for multiplication"
    }
    return a * b
}

func calculator_example(): int {
    let a: int = 20
    let b: int = 4
    
    // Safe division
    try {
        let result: int = safe_divide(a, b)
        io.println("Division result: " + result)
    } catch (e) {
        io.println("Division error: " + e)
    }
    
    // Safe multiplication
    try {
        let result: int = safe_multiply(a, b)
        io.println("Multiplication result: " + result)
    } catch (e) {
        io.println("Multiplication error: " + e)
    }
    
    // Division by zero
    try {
        let result: int = safe_divide(a, 0)
        io.println("Division result: " + result)
    } catch (e) {
        io.println("Division error: " + e)
    }
    
    return 0
}
```

### Data Processing with Types

```omni
import std.io as io

// Type aliases for data processing
type DataID = int
type DataValue = string | int | float
type DataStatus = string | int | bool
type OptionalData = string?

func process_data(id: DataID, value: DataValue, status: DataStatus): OptionalData {
    if id > 0 {
        return "Data processed: ID " + id + ", Value: " + value + ", Status: " + status
    }
    return null
}

func data_processing_example(): int {
    let data_id: DataID = 1
    let string_value: DataValue = "Hello"
    let int_value: DataValue = 42
    let float_value: DataValue = 3.14
    let status: DataStatus = "processed"
    
    let result1: OptionalData = process_data(data_id, string_value, status)
    let result2: OptionalData = process_data(data_id, int_value, status)
    let result3: OptionalData = process_data(data_id, float_value, status)
    
    if result1 != null {
        io.println("Result 1: " + result1)
    }
    if result2 != null {
        io.println("Result 2: " + result2)
    }
    if result3 != null {
        io.println("Result 3: " + result3)
    }
    
    return 0
}
```

## Complete Programs

### Advanced Features Demo

```omni
import std.io as io

// Type aliases
type UserID = int
type Name = string
type Email = string

// Union types
type StringOrInt = string | int
type Number = int | float

// Optional types
type OptionalUser = string?

func advanced_features_demo(): int {
    io.println("=== Advanced Features Demo ===")
    
    // Type aliases
    let user_id: UserID = 42
    let user_name: Name = "Alice"
    let user_email: Email = "alice@example.com"
    
    // String interpolation
    let greeting: string = "Hello, ${user_name}!"
    let info: string = "User ID: ${user_id}, Email: ${user_email}"
    
    io.println(greeting)
    io.println(info)
    
    // Union types
    let value1: StringOrInt = "Hello"
    let value2: StringOrInt = 42
    let num1: Number = 100
    let num2: Number = 3.14
    
    io.println("String value: " + value1)
    io.println("Int value: " + value2)
    io.println("Int number: " + num1)
    io.println("Float number: " + num2)
    
    // Optional types
    let maybe_int: int? = 42
    let maybe_string: string? = "Hello"
    
    io.println("Optional int: " + maybe_int)
    io.println("Optional string: " + maybe_string)
    
    // Exception handling
    try {
        let result: int = risky_operation(20)
        io.println("Result: " + result)
    } catch (e) {
        io.println("Caught exception: " + e)
    } finally {
        io.println("Finally block executed")
    }
    
    io.println("=== Demo Complete ===")
    return 0
}

func risky_operation(x: int): int {
    if x < 0 {
        throw "Negative values not allowed"
    }
    return x * 2
}
```

### Comprehensive Example

```omni
import std.io as io

// Type aliases for a complete application
type ProductID = int
type ProductName = string
type Price = float
type Quantity = int
type ProductStatus = string | int | bool
type OptionalPrice = Price?
type OptionalQuantity = Quantity?

func create_product(id: ProductID, name: ProductName, price: OptionalPrice, quantity: OptionalQuantity): string {
    let price_info: string = "Price not set"
    if price != null {
        price_info = "Price: $" + price
    }
    
    let quantity_info: string = "Quantity not set"
    if quantity != null {
        quantity_info = "Quantity: " + quantity
    }
    
    return "Product created: " + name + " (ID: " + id + "), " + price_info + ", " + quantity_info
}

func update_product_price(id: ProductID, price: Price): string {
    return "Product " + id + " price updated to $" + price
}

func process_product(id: ProductID, name: ProductName, price: Price, quantity: Quantity, status: ProductStatus): string {
    try {
        let price_info: string = "Price: $" + price
        let quantity_info: string = "Quantity: " + quantity
        let status_info: string = "Status: " + status
        
        return "Product ${name} (ID: ${id}) - ${price_info}, ${quantity_info}, ${status_info}"
    } catch (e) {
        return "Error processing product: " + e
    }
}

func comprehensive_example(): int {
    io.println("=== Comprehensive Example ===")
    
    // Create a product
    let product_id: ProductID = 1
    let product_name: ProductName = "Laptop"
    let product_price: OptionalPrice = 999.99
    let product_quantity: OptionalQuantity = 10
    
    let creation_result: string = create_product(product_id, product_name, product_price, product_quantity)
    io.println(creation_result)
    
    // Update the price
    let new_price: Price = 899.99
    let update_result: string = update_product_price(product_id, new_price)
    io.println(update_result)
    
    // Process the product
    let status: ProductStatus = "active"
    let process_result: string = process_product(product_id, product_name, new_price, product_quantity, status)
    io.println(process_result)
    
    io.println("=== Example Complete ===")
    return 0
}
```

## Running Examples

To run these examples:

1. Save the code to a file with `.omni` extension
2. Compile and run:
   ```bash
   ./bin/omnic example.omni -o example
   ./example
   ```

## Best Practices

1. **Use type aliases** for better code readability and maintainability
2. **Use union types** when a value can be one of several types
3. **Use optional types** for values that might be null
4. **Handle exceptions** appropriately with try-catch-finally blocks
5. **Use string interpolation** for better string formatting
6. **Keep examples simple** and focused on specific features
7. **Test edge cases** and error conditions

## Conclusion

These examples demonstrate the power and flexibility of OmniLang's advanced features. The combination of type aliases, union types, optional types, exception handling, and string interpolation provides a modern, expressive programming language that's both safe and easy to use.
