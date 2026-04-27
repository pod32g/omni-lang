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

## Trim variants

### trim_left(s: string): string

Returns `s` with leading ASCII whitespace removed (space, tab,
newline, carriage return, vertical tab, form feed).

### trim_right(s: string): string

Returns `s` with trailing whitespace removed.

### trim_all(s: string): string

Returns `s` with **every** whitespace character removed, including
internal ones. `"a b c"` becomes `"abc"`.

`trim` (without a suffix) removes leading and trailing whitespace
only, leaving internal spacing alone.

## Case manipulation

### to_title(s: string): string

Returns `s` with the first letter of each whitespace-separated word
uppercased and the rest lowercased.

```omni
std.string.to_title("hello world from omni")
// "Hello World From Omni"
```

### capitalize(s: string): string

Uppercases the first character; the rest of the string is left
untouched. (Compare with `to_title`, which lowercases the tail.)

### reverse(s: string): string

Returns `s` with its characters in reverse order.

## Case-insensitive comparison

### equals_ignore_case(a: string, b: string): bool

Returns `true` if the strings are equal under case-insensitive ASCII
comparison.

### compare_ignore_case(a: string, b: string): int

Returns -1, 0, or 1 — like `compare`, but case-insensitive.

## Splitting and joining

### split(s: string, delimiter: string): array<string>

Splits `s` on every occurrence of `delimiter`. An empty delimiter
splits into individual characters.

```omni
let parts: array<string> = std.string.split("a,b,c,d", ",")
// ["a", "b", "c", "d"]
```

### split_lines(s: string): array<string>

Equivalent to `split(s, "\n")`.

### split_words(s: string): array<string>

Splits on runs of whitespace, collapsing consecutive whitespace and
ignoring leading/trailing whitespace. Different from
`split(s, " ")`, which would produce empty strings for runs.

```omni
std.string.split_words("  one   two three ")
// ["one", "two", "three"]
```

### join(parts: array<string>, separator: string): string

Concatenates `parts` with `separator` between each element.

```omni
std.string.join(["x", "y", "z"], "-")
// "x-y-z"
```

### join_lines(parts: array<string>): string

Equivalent to `join(parts, "\n")`.

## Replacement

### replace_all(s: string, old: string, new: string): string

Replaces every occurrence of `old` in `s` with `new`. An empty `old`
returns `s` unchanged.

### replace(s: string, old: string, new: string): string

Alias for `replace_all`.

### replace_first(s: string, old: string, new: string): string

Replaces only the first occurrence.

### replace_last(s: string, old: string, new: string): string

Replaces only the last occurrence.

## Searching

### find_all(s: string, sub: string): array<int>

Returns the byte offset of every non-overlapping occurrence of `sub`
in `s`. Returns an empty array if `sub` is empty or not found.

```omni
let offsets: array<int> = std.string.find_all("ababcabcabc", "abc")
// [2, 5, 8]
```

## Counting

### is_empty(s: string): bool

Returns `true` if `s` has zero characters.

### count_occurrences(s: string, sub: string): int

Returns the number of non-overlapping occurrences of `sub` in `s`.

### count_lines(s: string): int

Returns the number of lines in `s`. A trailing newline is **not**
counted as an extra empty line: `"a\nb"` and `"a\nb\n"` both return 2.
An empty string returns 0.

### count_words(s: string): int

Returns the number of whitespace-separated word runs.

### count_chars(s: string): int

Equivalent to `length(s)`.

## Char ↔ int

These live at the top of `std` (not `std.string`) but are essential
for any character-level work:

### std.char_code(c: char): int

Returns the Unicode code point of `c`.

### std.char_from_code(code: int): char

Returns the char with the given code point.

### std.char_to_string(c: char): string

Returns a single-character string holding `c`.

```omni
let c: char = 'A'
let code: int = std.char_code(c)               // 65
let next: char = std.char_from_code(code + 1)  // 'B'
let s: string = std.char_to_string(next)       // "B"
```

## Notes

- All string functions work with UTF-8 encoded strings
- String indices are 0-based
- The `length` function returns the number of characters, not bytes
- String comparison is case-sensitive (use `equals_ignore_case` /
  `compare_ignore_case` for the alternative)
- All functions return new strings rather than modifying the original
