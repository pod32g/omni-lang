# std.io - Input/Output Functions

The `std.io` module provides functions for input and output operations.

## Functions

### print(value: string | int | float | double | bool): void

Prints a primitive value to standard output without a newline. Non-string values are automatically converted.

**Parameters:**
- `value` (string \| int \| float \| double \| bool): The value to print

**Example:**
```omni
import std.io as io

func main():int {
    io.print("Result: ")
    io.print(42)
    io.print(", ratio = ")
    io.print(3.14)
    io.print(", ok = ")
    io.print(true)
    io.println("")
    return 0
}
```

### println(value: string | int | float | double | bool): void

Prints a primitive value to standard output with a newline. Non-string values are automatically converted.

**Parameters:**
- `value` (string \| int \| float \| double \| bool): The value to print

**Example:**
```omni
import std.io as io

func main():int {
    io.println("Hello, World!")
    io.println(42)
    io.println(3.14)
    io.println(false)
    return 0
}
```

### read_line(): string

Reads one line from standard input and returns it without the trailing newline characters. If no data is available (EOF), an empty string is returned.

**Example:**
```omni
import std.io as io

func main():int {
    io.print("Enter value: ")
    let line = io.read_line()
    io.println("You typed: " + line)
    return 0
}
```

## Usage Examples

### Basic Output

```omni
import std.io as io

func main():int {
    io.println("=== Basic Output ===")
    
    let name:string = "Alice"
    let age:int = 30
    let height:float = 5.6
    let is_student:bool = false
    
    io.println("Name: " + name)
    io.print("Age: ")
    io.println(age)
    io.print("Height: ")
    io.println(height)
    io.print("Is student: ")
    io.println(is_student)
    
    return 0
}
```

### Formatted Output

```omni
import std.io as io

func main():int {
    let x:int = 10
    let y:int = 20
    let sum:int = x + y
    
    io.print("x = ")
    io.print(x)
    io.print(", y = ")
    io.print(y)
    io.print(", sum = ")
    io.println(sum)
    
    return 0
}
```

### Mixed Type Output

```omni
import std.io as io

func main():int {
    let count:int = 5
    let message:string = "items"
    let price:float = 9.99
    
    // String concatenation with automatic type conversion
    io.println("Found " + count + " " + message)
    io.println("Price: $" + price)
    
    return 0
}
```

## Notes

- All output functions write to standard output (stdout)
- The `print` functions do not add a newline, while `println` functions do
- String concatenation with the `+` operator automatically converts numbers to strings
- The module is automatically available when imported as `std.io` or with an alias
