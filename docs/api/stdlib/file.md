# std.file - File Handle I/O

The `std.file` module provides low-level file-handle operations. Use `std.os.read_file`, `std.os.write_file`, and `std.os.append_file` for whole-file string workflows; use `std.file` when code needs explicit open, seek, tell, write, close, and metadata checks.

## Handles

`open(filename, mode)` returns an integer handle, or `-1` if the file could not be opened. Pass that handle to the other functions and close it when finished.

Supported modes are the host C runtime modes such as `"r"`, `"w"`, `"a"`, `"r+"`, `"w+"`, and `"a+"`.

```omni
import std

func main():int {
    let path:string = "example.txt"
    let handle:int = std.file.open(path, "w+")
    if handle < 0 {
        return 1
    }

    let n:int = std.string.length("hello")
    if std.file.write(handle, "hello", n) != n {
        std.file.close(handle)
        return 2
    }

    std.file.close(handle)
    return 0
}
```

## Functions

| Function | Description |
|----------|-------------|
| `open(filename:string, mode:string): int` | Opens a file and returns a handle, or `-1` on error. |
| `close(file_handle:int): int` | Closes a handle. Returns `0` on success, `-1` on error. |
| `write(file_handle:int, buffer:string, size:int): int` | Writes up to `size` bytes from `buffer`. Returns bytes written, or `-1` on error. |
| `read(file_handle:int, buffer:string, size:int): int` | Reads up to `size` bytes and returns the byte count, or `-1` on error. The current API does not mutate `buffer`; use this as a count/probe operation until mutable byte buffers exist. |
| `seek(file_handle:int, offset:int, whence:int): int` | Moves the file position. Returns `0` on success, `-1` on error. |
| `tell(file_handle:int): int` | Returns the current file position, or `-1` on error. |
| `exists(filename:string): int` | Returns `1` if the path exists, otherwise `0`. |
| `size(filename:string): int` | Returns file size in bytes, or `-1` on error. |

`seek` uses the C `fseek` constants: `0` for start, `1` for current position, and `2` for end.

## Example

```omni
import std

func main():int {
    let path:string = "data.txt"
    let expected:int = std.string.length("hello")
    let handle:int = std.file.open(path, "w+")
    if handle < 0 {
        return 1
    }

    if std.file.write(handle, "hello", expected) != expected {
        std.file.close(handle)
        std.os.remove(path)
        return 2
    }

    if std.file.seek(handle, 0, 0) != 0 {
        std.file.close(handle)
        std.os.remove(path)
        return 3
    }

    if std.file.read(handle, "ignored", expected) != expected {
        std.file.close(handle)
        std.os.remove(path)
        return 4
    }

    std.file.close(handle)
    std.os.remove(path)
    return 0
}
```

## Backend Status

`std.file` is implemented in both `omnir` and `omnic`. The C backend stores runtime handles in pointer-sized slots internally so native `FILE*` values are not truncated.
