# Advanced Types API Reference

This document provides the complete API reference for OmniLang's advanced type system features.

## Table of Contents

1. [Type Aliases](#type-aliases)
2. [Union Types](#union-types)
3. [Optional Types](#optional-types)
4. [Type System Functions](#type-system-functions)
5. [Examples](#examples)

## Type Aliases

Type aliases provide semantic meaning and improve code readability by creating alternative names for existing types.

### Syntax

```omni
type AliasName = ExistingType
```

### Examples

```omni
// Basic type aliases
type UserID = int
type Name = string
type Email = string
type Age = int

// Usage
let user_id: UserID = 42
let user_name: Name = "Alice"
let user_email: Email = "alice@example.com"
let user_age: Age = 30
```

### Generic Type Aliases (Planned)

```omni
// Generic type aliases (syntax supported, implementation in progress)
type Container<T> = array<T>
type Pair<T, U> = struct { first: T, second: U }

// Usage
let int_container: Container<int> = [1, 2, 3]
let string_pair: Pair<string, string> = { first: "Hello", second: "World" }
```

## Union Types

Union types allow values to be one of several types, providing flexibility and type safety.

### Syntax

```omni
type UnionName = Type1 | Type2 | Type3
```

### Examples

```omni
// Basic union types
type StringOrInt = string | int
type Number = int | float
type Status = string | int | bool

// Usage
let value1: StringOrInt = "Hello"
let value2: StringOrInt = 42
let num1: Number = 100
let num2: Number = 3.14
let status1: Status = "active"
let status2: Status = 1
let status3: Status = true
```

### Complex Union Types

```omni
// Complex union types with multiple members
type ComplexType = string | int | float | bool
type Result<T, E> = Ok(T) | Error(E)

// Usage
let complex_value: ComplexType = "Hello"
let number_value: ComplexType = 42
let float_value: ComplexType = 3.14
let bool_value: ComplexType = true
```

## Optional Types

Optional types represent values that might be null, providing null safety.

### Syntax

```omni
type OptionalName = BaseType?
```

### Examples

```omni
// Basic optional types
type OptionalInt = int?
type OptionalString = string?
type OptionalNumber = Number?

// Usage
let maybe_int: OptionalInt = 42
let maybe_string: OptionalString = "Hello"
let maybe_number: OptionalNumber = 3.14
```

### Optional Type Compatibility

```omni
// Optional types are compatible with their base types
let maybe_value: int? = 42
let regular_value: int = maybe_value  // Automatic conversion

let maybe_string: string? = "Hello"
let regular_string: string = maybe_string  // Automatic conversion

// Reverse conversion also works
let back_to_optional: int? = regular_value
```

## Type System Functions

### Type Checking Functions

```omni
// Type checking functions (internal to the compiler)
func isTypeAlias(name: string): bool
func isUnionType(type: string): bool
func isOptionalType(type: string): bool
func resolveTypeAlias(alias: string): string
func getUnionMembers(union: string): array<string>
func getOptionalBaseType(optional: string): string
```

### Type Conversion Functions

```omni
// Type conversion functions (internal to the compiler)
func convertToUnion(value: any, union: string): any
func convertToOptional(value: any, optional: string): any
func resolveTypeExpression(expr: TypeExpr): string
```

## Examples

### Complete Type System Example

```omni
import std.io as io

// Type aliases
type UserID = int
type Name = string
type Email = string
type Age = int

// Union types
type StringOrInt = string | int
type Number = int | float
type Status = string | int | bool

// Optional types
type OptionalUser = string?
type OptionalAge = Age?

func create_user(id: UserID, name: Name, email: Email, age: OptionalAge): string {
    let age_info: string = "Unknown age"
    if age != null {
        age_info = "Age: " + age
    }
    
    return "User created: " + name + " (ID: " + id + "), Email: " + email + ", " + age_info
}

func process_user_data(id: UserID, data: StringOrInt, status: Status): string {
    return "User " + id + " data: " + data + ", status: " + status
}

func type_system_example(): int {
    // Type aliases
    let user_id: UserID = 1
    let user_name: Name = "Alice"
    let user_email: Email = "alice@example.com"
    let user_age: OptionalAge = 30
    
    // Union types
    let string_data: StringOrInt = "Hello"
    let int_data: StringOrInt = 42
    let status: Status = "active"
    
    // Create user
    let creation_result: string = create_user(user_id, user_name, user_email, user_age)
    io.println(creation_result)
    
    // Process user data
    let process_result1: string = process_user_data(user_id, string_data, status)
    let process_result2: string = process_user_data(user_id, int_data, status)
    
    io.println(process_result1)
    io.println(process_result2)
    
    return 0
}
```

### Real-World Application Example

```omni
import std.io as io

// Domain-specific type aliases
type ProductID = int
type ProductName = string
type Price = float
type Quantity = int
type Category = string

// Union types for flexible data
type ProductStatus = string | int | bool
type ProductData = string | int | float

// Optional types for nullable fields
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

func update_product_status(id: ProductID, status: ProductStatus): string {
    return "Product " + id + " status updated to: " + status
}

func process_product_data(id: ProductID, data: ProductData, status: ProductStatus): string {
    return "Product " + id + " data: " + data + ", status: " + status
}

func real_world_example(): int {
    // Create products with different data types
    let product_id: ProductID = 1
    let product_name: ProductName = "Laptop"
    let product_price: OptionalPrice = 999.99
    let product_quantity: OptionalQuantity = 10
    
    let creation_result: string = create_product(product_id, product_name, product_price, product_quantity)
    io.println(creation_result)
    
    // Update product status
    let status: ProductStatus = "active"
    let status_update: string = update_product_status(product_id, status)
    io.println(status_update)
    
    // Process product data
    let string_data: ProductData = "High-end laptop"
    let int_data: ProductData = 100
    let float_data: ProductData = 4.5
    
    let process_result1: string = process_product_data(product_id, string_data, status)
    let process_result2: string = process_product_data(product_id, int_data, status)
    let process_result3: string = process_product_data(product_id, float_data, status)
    
    io.println(process_result1)
    io.println(process_result2)
    io.println(process_result3)
    
    return 0
}
```

## Best Practices

### Type Aliases

1. **Use descriptive names**: Choose names that clearly indicate the purpose of the type
2. **Group related types**: Create aliases for related types to improve consistency
3. **Avoid over-abstraction**: Don't create aliases for simple types unless they add semantic value
4. **Use consistent naming**: Follow a consistent naming convention for type aliases

### Union Types

1. **Keep unions simple**: Avoid overly complex union types with many members
2. **Use meaningful names**: Choose names that clearly indicate what the union represents
3. **Consider type safety**: Ensure that union types don't compromise type safety
4. **Document union members**: Clearly document what each union member represents

### Optional Types

1. **Use for nullable values**: Only use optional types when a value might legitimately be null
2. **Handle null cases**: Always consider what happens when an optional value is null
3. **Avoid optional chaining**: Keep optional types simple and avoid deep nesting
4. **Use consistent patterns**: Follow consistent patterns for handling optional values

## Implementation Details

### Type Resolution

The type system resolves types in the following order:

1. **Type aliases**: Resolved to their underlying types
2. **Union types**: Checked for member compatibility
3. **Optional types**: Checked for base type compatibility
4. **Base types**: Standard type checking

### Type Compatibility

- **Type aliases**: Fully compatible with their underlying types
- **Union types**: Compatible with any of their member types
- **Optional types**: Compatible with their base types and null
- **Base types**: Standard type compatibility rules apply

### Error Handling

The type system provides detailed error messages for:

- **Unknown type aliases**: Clear error messages for undefined type aliases
- **Invalid union types**: Error messages for invalid union type usage
- **Type mismatches**: Detailed error messages for type compatibility issues
- **Optional type errors**: Clear error messages for optional type misuse

## Conclusion

OmniLang's advanced type system provides powerful features for creating expressive and type-safe code. The combination of type aliases, union types, and optional types allows developers to create sophisticated type systems that improve code readability, maintainability, and safety.
