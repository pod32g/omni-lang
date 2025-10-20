# OmniLang Standard Library

This directory contains the standard library for OmniLang, providing essential functions and utilities for common programming tasks.

## Modules

### std.io
Input/Output functions for console operations.

**Functions:**
- `print(s:string)` - Print string without newline
- `println(s:string)` - Print string with newline
- `print_int(i:int)` - Print integer without newline
- `println_int(i:int)` - Print integer with newline
- `print_float(f:float)` - Print float without newline
- `println_float(f:float)` - Print float with newline
- `print_bool(b:bool)` - Print boolean without newline
- `println_bool(b:bool)` - Print boolean with newline

### std.math
Mathematical functions and utilities.

**Functions:**
- `abs(x:int):int` - Absolute value of integer
- `abs_float(x:float):float` - Absolute value of float
- `max(a:int, b:int):int` - Maximum of two integers
- `min(a:int, b:int):int` - Minimum of two integers
- `max_float(a:float, b:float):float` - Maximum of two floats
- `min_float(a:float, b:float):float` - Minimum of two floats
- `pow(base:int, exp:int):int` - Power function
- `sqrt(x:int):int` - Integer square root
- `gcd(a:int, b:int):int` - Greatest common divisor
- `lcm(a:int, b:int):int` - Least common multiple
- `is_prime(n:int):bool` - Check if number is prime
- `factorial(n:int):int` - Factorial function

### std.string
String manipulation functions.

**Functions:**
- `length(s:string):int` - String length
- `concat(a:string, b:string):string` - Concatenate strings
- `substring(s:string, start:int, end:int):string` - Extract substring
- `char_at(s:string, index:int):char` - Get character at index
- `starts_with(s:string, prefix:string):bool` - Check prefix
- `ends_with(s:string, suffix:string):bool` - Check suffix
- `contains(s:string, substr:string):bool` - Check substring
- `index_of(s:string, substr:string):int` - Find substring index
- `last_index_of(s:string, substr:string):int` - Find last substring index
- `trim(s:string):string` - Remove whitespace
- `to_upper(s:string):string` - Convert to uppercase
- `to_lower(s:string):string` - Convert to lowercase
- `equals(a:string, b:string):bool` - String equality
- `compare(a:string, b:string):int` - String comparison

### std.array
Array manipulation functions.

**Functions:**
- `length<T>(arr:array<T>):int` - Array length
- `get<T>(arr:array<T>, index:int):T` - Get element at index
- `set<T>(arr:array<T>, index:int, value:T)` - Set element at index
- `append<T>(arr:array<T>, value:T):array<T>` - Append element
- `prepend<T>(arr:array<T>, value:T):array<T>` - Prepend element
- `insert<T>(arr:array<T>, index:int, value:T):array<T>` - Insert element
- `remove<T>(arr:array<T>, index:int):array<T>` - Remove element
- `contains<T>(arr:array<T>, value:T):bool` - Check if contains value
- `index_of<T>(arr:array<T>, value:T):int` - Find value index
- `reverse<T>(arr:array<T>):array<T>` - Reverse array
- `slice<T>(arr:array<T>, start:int, end:int):array<T>` - Extract slice
- `concat<T>(a:array<T>, b:array<T>):array<T>` - Concatenate arrays
- `fill<T>(arr:array<T>, value:T)` - Fill array with value
- `copy<T>(src:array<T>, dest:array<T>, count:int)` - Copy elements

### std.os
Operating system interface functions.

**Functions:**
- `exit(code:int)` - Terminate program
- `getenv(name:string):string` - Get environment variable
- `setenv(name:string, value:string):bool` - Set environment variable
- `unsetenv(name:string):bool` - Remove environment variable
- `getpid():int` - Get process ID
- `getppid():int` - Get parent process ID
- `getcwd():string` - Get current working directory
- `chdir(path:string):bool` - Change directory
- `mkdir(path:string):bool` - Create directory
- `rmdir(path:string):bool` - Remove directory
- `exists(path:string):bool` - Check if path exists
- `is_file(path:string):bool` - Check if path is file
- `is_dir(path:string):bool` - Check if path is directory
- `remove(path:string):bool` - Remove file/directory
- `rename(old_path:string, new_path:string):bool` - Rename file/directory
- `copy(src:string, dest:string):bool` - Copy file
- `read_file(path:string):string` - Read file contents
- `write_file(path:string, contents:string):bool` - Write file contents
- `append_file(path:string, contents:string):bool` - Append to file

### std.collections
Collection data structures.

**Map Functions:**
- `size<K,V>(m:map<K,V>):int` - Map size
- `get<K,V>(m:map<K,V>, key:K):V` - Get value by key
- `set<K,V>(m:map<K,V>, key:K, value:V)` - Set value by key
- `has<K,V>(m:map<K,V>, key:K):bool` - Check if key exists
- `remove<K,V>(m:map<K,V>, key:K):bool` - Remove key-value pair
- `clear<K,V>(m:map<K,V>)` - Clear all entries
- `keys<K,V>(m:map<K,V>):array<K>` - Get all keys
- `values<K,V>(m:map<K,V>):array<V>` - Get all values
- `copy<K,V>(m:map<K,V>):map<K,V>` - Copy map
- `merge<K,V>(a:map<K,V>, b:map<K,V>):map<K,V>` - Merge maps

## Usage

Import the standard library in your OmniLang programs:

```omni
import std

func main():int {
    std.io.println("Hello, World!")
    let result:int = std.math.max(10, 20)
    std.io.println_int(result)
    return 0
}
```

Or import specific modules:

```omni
import std.io
import std.math

func main():int {
    io.println("Hello, World!")
    let result:int = math.max(10, 20)
    io.println_int(result)
    return 0
}
```

## Implementation Status

⚠️ **Note**: Most standard library functions are currently declared as intrinsic functions that need to be implemented in the runtime backends. The current implementation provides:

- ✅ Function declarations and type signatures
- ✅ Basic math functions (implemented in Go)
- ❌ String manipulation (needs runtime implementation)
- ❌ Array operations (needs runtime implementation)
- ❌ OS operations (needs runtime implementation)
- ❌ Collection operations (needs runtime implementation)

## Examples

See the `examples/` directory for demonstration programs using the standard library.

## Contributing

To add new standard library functions:

1. Add the function declaration to the appropriate module
2. Implement the function in the runtime backends (VM and Cranelift)
3. Add tests for the new functionality
4. Update this documentation
