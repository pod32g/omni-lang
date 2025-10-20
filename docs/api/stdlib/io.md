# std.io - Input/Output Functions

The `std.io` module provides functions for input and output operations.

## Functions

### print(value: string): void

Prints a string to standard output without a newline.

**Parameters:**
- `value` (string): The string to print

**Example:**
```omni
import std.io as io

func main():int {
    io.print("Hello")
    io.print(" ")
    io.print("World")
    // Output: Hello World
    return 0
}
```

### println(value: string): void

Prints a string to standard output with a newline.

**Parameters:**
- `value` (string): The string to print

**Example:**
```omni
import std.io as io

func main():int {
    io.println("Hello, World!")
    io.println("This is a new line")
    return 0
}
```

### print_int(value: int): void

Prints an integer to standard output without a newline.

**Parameters:**
- `value` (int): The integer to print

**Example:**
```omni
import std.io as io

func main():int {
    let x:int = 42
    io.print("The answer is: ")
    io.print_int(x)
    io.println("")
    return 0
}
```

### println_int(value: int): void

Prints an integer to standard output with a newline.

**Parameters:**
- `value` (int): The integer to print

**Example:**
```omni
import std.io as io

func main():int {
    let x:int = 42
    io.println_int(x)  // Output: 42
    return 0
}
```

### print_float(value: float): void

Prints a float to standard output without a newline.

**Parameters:**
- `value` (float): The float to print

**Example:**
```omni
import std.io as io

func main():int {
    let pi:float = 3.14159
    io.print("Pi is approximately: ")
    io.print_float(pi)
    io.println("")
    return 0
}
```

### println_float(value: float): void

Prints a float to standard output with a newline.

**Parameters:**
- `value` (float): The float to print

**Example:**
```omni
import std.io as io

func main():int {
    let pi:float = 3.14159
    io.println_float(pi)  // Output: 3.14159
    return 0
}
```

### print_bool(value: bool): void

Prints a boolean to standard output without a newline.

**Parameters:**
- `value` (bool): The boolean to print

**Example:**
```omni
import std.io as io

func main():int {
    let flag:bool = true
    io.print("Flag is: ")
    io.print_bool(flag)
    io.println("")
    return 0
}
```

### println_bool(value: bool): void

Prints a boolean to standard output with a newline.

**Parameters:**
- `value` (bool): The boolean to print

**Example:**
```omni
import std.io as io

func main():int {
    let flag:bool = true
    io.println_bool(flag)  // Output: true
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
    io.println_int(age)
    io.print("Height: ")
    io.println_float(height)
    io.print("Is student: ")
    io.println_bool(is_student)
    
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
    io.print_int(x)
    io.print(", y = ")
    io.print_int(y)
    io.print(", sum = ")
    io.println_int(sum)
    
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
