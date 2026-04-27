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

### read_lines(): array<string>

Slurp stdin to EOF and return one entry per line. Trailing newlines are stripped; a final newline does not produce a phantom blank entry. Mirrors the `bufio.Scanner`-over-stdin pattern in Go.

```omni
import std.io as io

func main():int {
    let lines:array<string> = io.read_lines()
    io.println(len(lines))
    return 0
}
```

### read_int() / read_float()

Read one line from stdin and parse it as an int or float. Returns `0` / `0.0` on parse failure or EOF (compose with `is_int` / `is_float` if you need to distinguish).

```omni
import std.io as io

func main():int {
    io.print("How many? ")
    io.flush()
    return io.read_int()
}
```

### prompt(message:string): string

Writes `message` to stdout, flushes, then reads one line. Saves the print + flush + read_line dance for interactive CLIs.

```omni
import std.io as io

func main():int {
    let name:string = io.prompt("Name: ")
    io.println("Hello, " + name)
    return 0
}
```

### is_terminal(): bool

True when stdout is connected to a terminal (TTY). Use this to gate ANSI colors or interactive output when the program might be redirected to a file or pipe. Mirrors `golang.org/x/term.IsTerminal(os.Stdout.Fd())`.

### sprint(value), sprintln(value)

Like `fmt.Sprint` / `fmt.Sprintln` for one value: convert a primitive (`int` / `float` / `bool` / `string`) to its string form, with `sprintln` adding a trailing newline.

```omni
import std.io as io

func main():int {
    let s:string = io.sprint(42)
    io.println(s)        // "42"
    io.print(io.sprintln(7))  // "7\n"
    return 0
}
```

### sprintf(format:string, args:array<string>): string

`%s`-substitution: each `%s` in `format` is replaced in order by entries from `args`. `%%` emits a literal `%`. Other `%`-directives are left intact (no type-dispatch since OmniLang has no varargs). Loosely mirrors `fmt.Sprintf`, scoped to `%s`.

```omni
import std.io as io

func main():int {
    let args:array<string> = ["world", "42"]
    io.println(io.sprintf("hello %s, num %s", args))
    return 0
}
```

### parse_int(s:string): int / parse_float(s:string): float

Quiet parsers that return the parsed value or `0` / `0.0` on any failure. Pair with `is_int` / `is_float` when you need to distinguish a parsed zero from a parse failure. Roughly equivalent to `strconv.Atoi` / `strconv.ParseFloat` with the error discarded.

### is_int(s:string): bool / is_float(s:string): bool

Predicates that report whether `s` parses cleanly (no leading/trailing junk, in range).

```omni
import std.io as io

func main():int {
    let line:string = io.read_line()
    if !io.is_int(line) {
        io.eprintln("expected an integer")
        return 1
    }
    return io.parse_int(line)
}
```

### printf, eprintf

`printf(format, args)` is `sprintf` + write to stdout; `eprintf` writes to stderr. Same `%s`-only substitution rules as `sprintf`.

```omni
import std.io as io

func main():int {
    let args:array<string> = ["world"]
    io.printf("hello, %s!\n", args)
    return 0
}
```

### print_each, eprint_each

Write each entry of `items` on its own line. `print_each` to stdout, `eprint_each` to stderr.

### eprompt(message): string

Like `prompt` but writes the message to stderr. Use this when stdout is being piped to another tool and the prompt shouldn't end up in the pipeline.

### confirm(message): bool

Display `message`, read one line, return `true` when the answer starts with `y` or `Y`. Anything else (including EOF and empty input) returns `false`.

### flush_stderr(): void

Force buffered stderr to be written. Stderr is usually unbuffered, but Windows can buffer it under some conditions.

### Colors and styles

ANSI SGR helpers wrap `s` in an escape sequence and reset at the end. They always emit codes — call `is_terminal()` and choose at the call site if you want to skip color when piping.

| Function | SGR code |
|----------|----------|
| `bold(s)` | `1` |
| `dim(s)` | `2` |
| `italic(s)` | `3` |
| `underline(s)` | `4` |
| `red(s)` | `31` |
| `green(s)` | `32` |
| `yellow(s)` | `33` |
| `blue(s)` | `34` |
| `magenta(s)` | `35` |
| `cyan(s)` | `36` |
| `style(s, code)` | any (e.g. `"38;5;208"` for orange in 256-color) |

```omni
import std.io as io

func main():int {
    if io.is_terminal() {
        io.println(io.bold(io.green("ok")))
    } else {
        io.println("ok")
    }
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
