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

### eprint(value), eprintln(value)

Same as `print` / `println` but write to standard error instead of standard output. Use these for diagnostics, prompts, and progress messages so the program's "real" output stays on stdout for piping.

```omni
import std.io as io

func main():int {
    io.eprintln("warning: input file is empty")
    return 1
}
```

### flush(): void

Forces buffered standard output to be written immediately. Useful after a `print()` that ends without a newline (e.g. an interactive prompt) when the next operation reads input or sleeps.

```omni
import std.io as io

func main():int {
    io.print("Name? ")
    io.flush()
    let name:string = io.read_line()
    io.println("hello, " + name)
    return 0
}
```

### read_line(): string

Reads one line from standard input and returns it without the trailing newline characters. If no data is available (EOF), an empty string is returned.

```omni
import std.io as io

func main():int {
    io.print("Enter value: ")
    let line = io.read_line()
    io.println("You typed: " + line)
    return 0
}
```

### read_all(): string

Reads stdin until EOF and returns the contents as a single string. Useful for pipe-style scripts that consume their entire input.

```omni
import std.io as io
import std.string as str

func main():int {
    let body:string = io.read_all()
    io.println(str.length(body))
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

- `print` / `println` write to stdout; `eprint` / `eprintln` write to stderr
- `print` / `eprint` do not add a newline, the `*ln` variants do
- Use `flush()` when you write a prompt without a newline and want it visible before the next read or sleep
- String concatenation with the `+` operator automatically converts numbers to strings
- The module is available when imported as `std.io` or with an alias

## Backend status

`std.io` is fully implemented on both `omnir` (VM) and `omnic` (C). Pinned by `TestStdIoBasic` and `TestStdIoRead` in `omni/tests/e2e/`.
