# Array Method Syntax Implementation

## Overview
OmniLang now supports method-style syntax for array operations. Instead of using function calls like `len(array)`, you can now use the more intuitive method syntax: `array.len()`.

## Syntax

### Before (Function Style):
```omni
let numbers: []int = [1, 2, 3, 4, 5]
let length: int = len(numbers)
```

### After (Method Style):
```omni
let numbers: []int = [1, 2, 3, 4, 5]
let length: int = numbers.len()
```

## Supported Methods

Currently, the following array methods are supported:

- **`array.len()`** - Returns the length of the array (type: `int`)

## Implementation Details

### Type Checker Changes
The type checker now recognizes array method access in `MemberExpr`:
- When checking `x.len` where `x` is an array, it returns `func():int` type
- When checking `x.len()` (CallExpr with MemberExpr callee), it returns `int`

### MIR Builder Changes
The MIR builder converts array method calls to builtin function calls:
- `x.len()` is internally converted to `len(x)`
- This ensures backward compatibility with the existing `len()` builtin function

### Code Changes

#### `omni/internal/types/checker/checker.go`:
1. **`checkExpr` for `*ast.MemberExpr`**:
   - Added handling for array method access
   - Returns `func():int` for `array.len`

2. **`checkCallExpr`**:
   - Added handling for function type calls
   - Extracts return type from function type strings like `func():int`

#### `omni/internal/mir/builder/builder.go`:
1. **`emitCall`**:
   - Added early check for array method calls
   - Converts `x.len()` to `call len %array_value`
   - Works for both simple identifiers and complex expressions

## Testing

### Test File: `tests/e2e/array_method.omni`
```omni
func main():int {
    let numbers: []int = [1, 2, 3, 4, 5]
    let length: int = numbers.len()
    return length
}
```

### Test Results:
- **VM Backend**: ✅ PASS (returns 5)
- **C Backend**: ✅ PASS (returns 5)

### Test Function: `TestArrayMethod` in `tests/e2e/e2e_test.go`
Tests both VM and C backends to ensure consistent behavior.

## Backward Compatibility

The functional style `len(array)` still works as before. Both syntaxes are supported:
- `len(numbers)` - Function style (original)
- `numbers.len()` - Method style (new)

## Future Enhancements

Potential future array methods could include:
- `array.first()` - Get first element
- `array.last()` - Get last element
- `array.isEmpty()` - Check if array is empty
- `array.slice(start, end)` - Get a slice of the array
- `array.append(element)` - Append an element
- `array.contains(element)` - Check if element exists

## Benefits

1. **More intuitive**: Method syntax is more natural for object-oriented programmers
2. **Better discoverability**: IDE autocomplete can show available methods
3. **Consistent with other languages**: Similar to `len()` in Python, `length` property in JavaScript
4. **Extensible**: Easy to add more array methods in the future

## Examples

### Basic Usage:
```omni
func main():int {
    let numbers: []int = [10, 20, 30]
    return numbers.len() // Returns 3
}
```

### With String Arrays:
```omni
func main():int {
    let words: []string = ["hello", "world"]
    return words.len() // Returns 2
}
```

### In Expressions:
```omni
func main():int {
    let numbers: []int = [1, 2, 3, 4, 5]
    if numbers.len() > 0 {
        return numbers.len()
    }
    return 0
}
```

## Conclusion

The array method syntax implementation provides a more intuitive and modern way to work with arrays in OmniLang. It maintains backward compatibility while offering a cleaner, more object-oriented syntax for common array operations.
