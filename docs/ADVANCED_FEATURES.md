# OmniLang Advanced Features

This document covers the advanced features implemented in OmniLang v0.5.1, including string interpolation, exception handling, and the advanced type system.

## Table of Contents

1. [String Interpolation](#string-interpolation)
2. [Exception Handling](#exception-handling)
3. [Advanced Type System](#advanced-type-system)
4. [Type Aliases](#type-aliases)
5. [Union Types](#union-types)
6. [Optional Types](#optional-types)
7. [Examples](#examples)

## String Interpolation

String interpolation allows you to embed expressions within string literals using the `${expression}` syntax.

### Basic Syntax

```omni
import std.io as io

func main(): int {
    let name: string = "Alice"
    let age: int = 30
    let pi: float = 3.14159
    
    // String interpolation
    let greeting: string = "Hello, ${name}!"
    let info: string = "Name: ${name}, Age: ${age}, Pi: ${pi}"
    
    io.println(greeting)    // "Hello, Alice!"
    io.println(info)        // "Name: Alice, Age: 30, Pi: 3.14159"
    
    return 0
}
```

### Complex Expressions

```omni
func complex_interpolation(): string {
    let x: int = 10
    let y: int = 20
    
    // Complex expressions in interpolation
    let result: string = "Sum: ${x + y}, Product: ${x * y}, Average: ${(x + y) / 2}"
    
    return result
}
```

### Nested Interpolation

```omni
func nested_interpolation(): string {
    let user: string = "Alice"
    let count: int = 5
    
    // Nested interpolation
    let message: string = "User ${user} has ${count} items: ${"Item " + count}"
    
    return message
}
```

## Exception Handling

OmniLang supports comprehensive exception handling with try-catch-finally blocks.

### Basic Exception Handling

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

### Multiple Catch Clauses

```omni
func multiple_catch_example(): int {
    try {
        let result: int = risky_operation(30)
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

OmniLang features a sophisticated type system with type aliases, union types, and optional types.

## Type Aliases

Type aliases provide semantic meaning and improve code readability.

### Basic Type Aliases

```omni
// Create type aliases for better code readability
type UserID = int
type Name = string
type Email = string

func create_user(id: UserID, name: Name, email: Email): void {
    io.println("Creating user: " + name + " (ID: " + id + ")")
}

func type_alias_example(): int {
    let user_id: UserID = 42
    let user_name: Name = "Alice"
    let user_email: Email = "alice@example.com"
    
    create_user(user_id, user_name, user_email)
    return 0
}
```

### Generic Type Aliases

```omni
// Generic type aliases (syntax supported, implementation in progress)
type Container<T> = array<T>
type Pair<T, U> = struct { first: T, second: U }

func generic_alias_example(): int {
    let int_container: Container<int> = [1, 2, 3]
    let string_pair: Pair<string, string> = { first: "Hello", second: "World" }
    
    return int_container.len()
}
```

## Union Types

Union types allow values to be one of several types, providing flexibility and type safety.

### Basic Union Types

```omni
// Union types allow values to be one of several types
type StringOrInt = string | int
type Number = int | float
type Status = string | int | bool

func process_value(value: StringOrInt): string {
    // The value can be either a string or an int
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
    
    return 0
}
```

### Complex Union Types

```omni
// Complex union types with multiple members
type ComplexType = string | int | float | bool
type Result<T, E> = Ok(T) | Error(E)

func complex_union_example(): int {
    let complex_value: ComplexType = "Hello"
    let number_value: ComplexType = 42
    let float_value: ComplexType = 3.14
    let bool_value: ComplexType = true
    
    io.println("Complex: " + complex_value)
    io.println("Number: " + number_value)
    io.println("Float: " + float_value)
    io.println("Bool: " + bool_value)
    
    return 0
}
```

## Optional Types

Optional types represent values that might be null, providing null safety.

### Basic Optional Types

```omni
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
    
    return 0
}
```

### Optional Type Compatibility

```omni
func optional_compatibility_example(): int {
    let maybe_value: int? = 42
    let regular_value: int = maybe_value  // Automatic conversion
    
    let maybe_string: string? = "Hello"
    let regular_string: string = maybe_string  // Automatic conversion
    
    // Reverse conversion also works
    let back_to_optional: int? = regular_value
    
    return 0
}
```

## Examples

### Comprehensive Example

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

### Real-World Example

```omni
import std.io as io

// Type aliases for domain modeling
type ProductID = int
type ProductName = string
type Price = float
type Quantity = int

// Union types for flexible data
type ProductStatus = string | int | bool
type ProductData = string | int | float

// Optional types for nullable fields
type OptionalPrice = Price?
type OptionalQuantity = Quantity?

struct Product {
    id: ProductID
    name: ProductName
    price: OptionalPrice
    quantity: OptionalQuantity
    status: ProductStatus
}

func create_product(id: ProductID, name: ProductName): Product {
    return Product{
        id: id
        name: name
        price: null
        quantity: null
        status: "draft"
    }
}

func update_product_price(product: Product, price: Price): Product {
    // Update the product with new price
    return Product{
        id: product.id
        name: product.name
        price: price
        quantity: product.quantity
        status: "active"
    }
}

func process_product(product: Product): string {
    try {
        let price_info: string = "Price: ${product.price}"
        let quantity_info: string = "Quantity: ${product.quantity}"
        let status_info: string = "Status: ${product.status}"
        
        return "Product ${product.name} (ID: ${product.id}) - ${price_info}, ${quantity_info}, ${status_info}"
    } catch (e) {
        return "Error processing product: " + e
    }
}

func real_world_example(): int {
    // Create a product
    let product: Product = create_product(1, "Laptop")
    
    // Update the price
    let updated_product: Product = update_product_price(product, 999.99)
    
    // Process the product
    let result: string = process_product(updated_product)
    io.println(result)
    
    return 0
}
```

## Best Practices

### Type Aliases

1. **Use descriptive names**: Choose names that clearly indicate the purpose of the type
2. **Group related types**: Create aliases for related types to improve consistency
3. **Avoid over-abstraction**: Don't create aliases for simple types unless they add semantic value

### Union Types

1. **Keep unions simple**: Avoid overly complex union types with many members
2. **Use meaningful names**: Choose names that clearly indicate what the union represents
3. **Consider type safety**: Ensure that union types don't compromise type safety

### Optional Types

1. **Use for nullable values**: Only use optional types when a value might legitimately be null
2. **Handle null cases**: Always consider what happens when an optional value is null
3. **Avoid optional chaining**: Keep optional types simple and avoid deep nesting

### Exception Handling

1. **Use specific exceptions**: Throw specific exception messages that help with debugging
2. **Handle exceptions appropriately**: Don't catch exceptions unless you can handle them meaningfully
3. **Use finally blocks**: Use finally blocks for cleanup operations that must always run

### String Interpolation

1. **Keep expressions simple**: Avoid complex expressions in string interpolation
2. **Use for readability**: Use interpolation when it improves code readability
3. **Consider performance**: Be mindful of performance implications for complex interpolations

## Migration Guide

### From v0.5.0 to v0.5.1

The following features are new in v0.5.1:

1. **String Interpolation**: Use `${expression}` syntax for embedding expressions in strings
2. **Exception Handling**: Use try-catch-finally blocks for error handling
3. **Type Aliases**: Use `type Name = ExistingType` syntax for type aliases
4. **Union Types**: Use `Type1 | Type2 | Type3` syntax for union types
5. **Optional Types**: Use `Type?` syntax for optional types

### Backward Compatibility

All existing code continues to work without modification. New features are additive and don't break existing functionality.

## Conclusion

OmniLang v0.5.1 introduces powerful advanced features that make the language more expressive and type-safe. These features provide:

- **Better string handling** with interpolation
- **Robust error handling** with exceptions
- **Flexible type system** with aliases, unions, and optionals
- **Improved code readability** and maintainability

These features work together to create a modern, powerful programming language that's both safe and expressive.
