# std.string - String Manipulation Functions

The `std.string` module provides functions for string manipulation and operations.

## Functions

### length(s: string): int

Returns the length of a string.

**Parameters:**
- `s` (string): The string to measure

**Returns:**
- `int`: The length of the string

**Example:**
```omni
import std.string as str

func main():int {
    let text:string = "Hello, World!"
    let len:int = str.length(text)  // 13
    return len
}
```

### substring(s: string, start: int, end: int): string

Returns a substring of the given string.

**Parameters:**
- `s` (string): The source string
- `start` (int): Starting index (inclusive)
- `end` (int): Ending index (exclusive)

**Returns:**
- `string`: The substring

**Example:**
```omni
import std.string as str

func main():int {
    let text:string = "Hello, World!"
    let sub:string = str.substring(text, 0, 5)  // "Hello"
    return 0
}
```

### char_at(s: string, index: int): char

Returns the character at the specified index.

**Parameters:**
- `s` (string): The source string
- `index` (int): The character index

**Returns:**
- `char`: The character at the index

**Example:**
```omni
import std.string as str

func main():int {
    let text:string = "Hello"
    let first_char:char = str.char_at(text, 0)  // 'H'
    return 0
}
```

### starts_with(s: string, prefix: string): bool

Checks if a string starts with the given prefix.

**Parameters:**
- `s` (string): The string to check
- `prefix` (string): The prefix to look for

**Returns:**
- `bool`: true if the string starts with the prefix

**Example:**
```omni
import std.string as str

func main():int {
    let text:string = "Hello, World!"
    let result1:bool = str.starts_with(text, "Hello")  // true
    let result2:bool = str.starts_with(text, "World")  // false
    return 0
}
```

### ends_with(s: string, suffix: string): bool

Checks if a string ends with the given suffix.

**Parameters:**
- `s` (string): The string to check
- `suffix` (string): The suffix to look for

**Returns:**
- `bool`: true if the string ends with the suffix

**Example:**
```omni
import std.string as str

func main():int {
    let text:string = "Hello, World!"
    let result1:bool = str.ends_with(text, "World!")  // true
    let result2:bool = str.ends_with(text, "Hello")   // false
    return 0
}
```

### contains(s: string, substr: string): bool

Checks if a string contains the given substring.

**Parameters:**
- `s` (string): The string to search in
- `substr` (string): The substring to look for

**Returns:**
- `bool`: true if the string contains the substring

**Example:**
```omni
import std.string as str

func main():int {
    let text:string = "Hello, World!"
    let result1:bool = str.contains(text, "World")  // true
    let result2:bool = str.contains(text, "sample") // false
    return 0
}
```

### index_of(s: string, substr: string): int

Returns the index of the first occurrence of a substring.

**Parameters:**
- `s` (string): The string to search in
- `substr` (string): The substring to look for

**Returns:**
- `int`: The index of the first occurrence, or -1 if not found

**Example:**
```omni
import std.string as str

func main():int {
    let text:string = "Hello, World!"
    let index1:int = str.index_of(text, "World")  // 7
    let index2:int = str.index_of(text, "sample") // -1
    return 0
}
```

### last_index_of(s: string, substr: string): int

Returns the index of the last occurrence of a substring.

**Parameters:**
- `s` (string): The string to search in
- `substr` (string): The substring to look for

**Returns:**
- `int`: The index of the last occurrence, or -1 if not found

**Example:**
```omni
import std.string as str

func main():int {
    let text:string = "Hello, Hello, World!"
    let index:int = str.last_index_of(text, "Hello")  // 7
    return 0
}
```

### trim(s: string): string

Removes leading and trailing whitespace from a string.

**Parameters:**
- `s` (string): The string to trim

**Returns:**
- `string`: The trimmed string

**Example:**
```omni
import std.string as str

func main():int {
    let text:string = "  Hello, World!  "
    let trimmed:string = str.trim(text)  // "Hello, World!"
    return 0
}
```

### to_upper(s: string): string

Converts a string to uppercase.

**Parameters:**
- `s` (string): The string to convert

**Returns:**
- `string`: The uppercase string

**Example:**
```omni
import std.string as str

func main():int {
    let text:string = "Hello, World!"
    let upper:string = str.to_upper(text)  // "HELLO, WORLD!"
    return 0
}
```

### to_lower(s: string): string

Converts a string to lowercase.

**Parameters:**
- `s` (string): The string to convert

**Returns:**
- `string`: The lowercase string

**Example:**
```omni
import std.string as str

func main():int {
    let text:string = "Hello, World!"
    let lower:string = str.to_lower(text)  // "hello, world!"
    return 0
}
```

### equals(a: string, b: string): bool

Checks if two strings are equal.

**Parameters:**
- `a` (string): First string
- `b` (string): Second string

**Returns:**
- `bool`: true if the strings are equal

**Example:**
```omni
import std.string as str

func main():int {
    let text1:string = "Hello"
    let text2:string = "Hello"
    let text3:string = "World"
    
    let result1:bool = str.equals(text1, text2)  // true
    let result2:bool = str.equals(text1, text3)  // false
    return 0
}
```

### compare(a: string, b: string): int

Compares two strings lexicographically.

**Parameters:**
- `a` (string): First string
- `b` (string): Second string

**Returns:**
- `int`: Negative if a < b, 0 if a == b, positive if a > b

**Example:**
```omni
import std.string as str

func main():int {
    let result1:int = str.compare("apple", "banana")  // negative
    let result2:int = str.compare("banana", "apple")  // positive
    let result3:int = str.compare("apple", "apple")   // 0
    return 0
}
```

## Usage Examples

### Basic String Operations

```omni
import std.string as str
import std.io as io

func main():int {
    let text:string = "  Hello, World!  "
    
    io.println("=== Basic String Operations ===")
    io.println("Original: '" + text + "'")
    io.println("Length: " + str.length(text))
    io.println("Trimmed: '" + str.trim(text) + "'")
    io.println("Uppercase: '" + str.to_upper(text) + "'")
    io.println("Lowercase: '" + str.to_lower(text) + "'")
    
    return 0
}
```

### String Searching

```omni
import std.string as str
import std.io as io

func main():int {
    let text:string = "Hello, World! Hello, OmniLang!"
    
    io.println("=== String Searching ===")
    io.println("Text: " + text)
    io.println("Starts with 'Hello': " + str.starts_with(text, "Hello"))
    io.println("Ends with 'OmniLang!': " + str.ends_with(text, "OmniLang!"))
    io.println("Contains 'World': " + str.contains(text, "World"))
    io.println("Index of 'World': " + str.index_of(text, "World"))
    io.println("Last index of 'Hello': " + str.last_index_of(text, "Hello"))
    
    return 0
}
```

### String Comparison

```omni
import std.string as str
import std.io as io

func main():int {
    let str1:string = "apple"
    let str2:string = "banana"
    let str3:string = "apple"
    
    io.println("=== String Comparison ===")
    io.println("str1: " + str1)
    io.println("str2: " + str2)
    io.println("str3: " + str3)
    io.println("str1 equals str2: " + str.equals(str1, str2))
    io.println("str1 equals str3: " + str.equals(str1, str3))
    io.println("str1 compare str2: " + str.compare(str1, str2))
    io.println("str1 compare str3: " + str.compare(str1, str3))
    
    return 0
}
```

### Character and Substring Operations

```omni
import std.string as str
import std.io as io

func main():int {
    let text:string = "OmniLang"
    
    io.println("=== Character and Substring Operations ===")
    io.println("Text: " + text)
    io.println("Length: " + str.length(text))
    io.println("First character: " + str.char_at(text, 0))
    io.println("Last character: " + str.char_at(text, str.length(text) - 1))
    io.println("Substring(0, 4): " + str.substring(text, 0, 4))
    io.println("Substring(4, 8): " + str.substring(text, 4, 8))
    
    return 0
}
```

## Notes

- All string functions work with UTF-8 encoded strings
- String indices are 0-based
- The `length` function returns the number of characters, not bytes
- String comparison is case-sensitive
- All functions return new strings rather than modifying the original
