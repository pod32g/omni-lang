# std.os - Operating System Helpers

The `std.os` module surfaces process-level functionality such as exiting, reading environment variables, working with files, and (newly) reading command-line arguments.

## Process & CLI helpers

| Function | Description |
|----------|-------------|
| `args(): array<string>` | Returns the arguments passed to the program (excluding the executable name). |
| `args_count(): int` | Returns the number of arguments available. |
| `has_flag(name:string): bool` | Returns true if `--name` is present in the argument list. |
| `get_flag(name:string, default:string): string` | Retrieves `--name=value` or the following argument. |
| `positional_arg(index:int, default:string): string` | Fetches positional args by index. |
| `exit(code:int)` | Terminates the program with the provided exit code. |
| `getpid(): int` / `getppid(): int` | Process identifiers. |

```omni
import std

func main():int {
    if std.os.args_count() == 0 {
        std.io.println("usage: greet <name>")
        std.os.exit(1)
    }
    let args = std.os.args()
    std.io.println("hello " + args[0])
    return 0
}
```

## Environment variables

| Function | Description |
|----------|-------------|
| `getenv(name:string): string` | Returns the value of the named variable (empty string if unset). |
| `setenv(name:string, value:string): bool` | Sets/overrides an environment variable. |
| `unsetenv(name:string): bool` | Removes a variable from the environment. |

```omni
import std

func main():int {
    let token = std.os.getenv("TOKEN")
    if token == "" {
        std.os.setenv("TOKEN", "dev-token")
        token = std.os.getenv("TOKEN")
    }
    std.io.println("TOKEN=" + token)
    return 0
}
```

## Files, directories, and metadata

The remaining helpers proxy through to the VM/runtime for common file operations:

- `getcwd()`, `chdir(path)`
- `mkdir`, `rmdir`, `exists`, `is_file`, `is_dir`
- `remove`, `rename`, `copy`
- `read_file`, `write_file`, `append_file`

These are already exercised by `omni/tests/std/std_os_comprehensive.omni` and reflect the VM parity work. For more examples, see the test case or the quick reference OS section.


Use the helpers to mix flag parsing with positional fallbacks:

```omni
import std

func main():int {
    let name = std.os.get_flag("name", "world")
    let loud = std.os.has_flag("loud")
    let firstArg = std.os.positional_arg(0, "")
    if firstArg != "" {
        name = firstArg
    }
    if loud {
        std.io.println("HELLO, " + name + "!")
    } else {
        std.io.println("Hello, " + name)
    }
    return 0
}
```

`get_flag` accepts both `--name=value` and `--name value` forms. Bare flags without values (e.g. `--loud`) evaluate to `true`. `positional_arg` skips entries that begin with `--` (and the value immediately following a bare `--flag`).
