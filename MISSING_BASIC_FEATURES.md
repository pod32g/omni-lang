# Missing Basic Language Features Analysis

## Overview
This document analyzes what basic language features are missing or incomplete in OmniLang, particularly focusing on the VM backend which has more limitations than the C backend.

## Critical Missing Features (High Priority)

### 1. String Comparisons in VM Backend ⚠️ **CRITICAL**
**Status**: Missing in VM backend
**Impact**: High - Basic language feature
**Details**: 
- VM backend only supports integer comparisons (`==`, `!=`, `<`, `>`, etc.)
- String comparisons fail because `execComparison` tries to convert strings to integers
- C backend works correctly
- **Fix needed**: Modify `execComparison` to handle string types

**Example that fails in VM:**
```omni
let name: string = "hello"
if name == "hello" {  // ❌ Fails in VM
    return 42
}
```

### 2. Float/Double Support in VM Backend ⚠️ **HIGH**
**Status**: Missing in VM backend
**Impact**: High - Basic numeric type
**Details**:
- VM backend only supports integer arithmetic
- Float literals and operations not supported
- C backend works correctly
- **Fix needed**: Add float support to VM arithmetic functions

**Example that fails in VM:**
```omni
let pi: float = 3.14
let result: float = pi * 2.0  // ❌ Fails in VM
```

### 3. Boolean Comparisons in VM Backend ⚠️ **MEDIUM**
**Status**: Missing in VM backend  
**Impact**: Medium - Basic type support
**Details**:
- VM backend doesn't handle boolean comparisons properly
- C backend works correctly
- **Fix needed**: Add boolean support to `execComparison`

**Example that fails in VM:**
```omni
let flag: bool = true
if flag == true {  // ❌ Fails in VM
    return 1
}
```

## Partially Implemented Features

### 4. Maps/Dictionaries ⚠️ **MEDIUM**
**Status**: Partially implemented
**Details**:
- Type checker recognizes `map` type
- MIR builder has `map.init` instruction
- VM has `execMapInit` but it's incomplete
- C backend doesn't support maps
- **Missing**: Map indexing, iteration, common operations

### 5. Structs ⚠️ **MEDIUM**
**Status**: Partially implemented
**Details**:
- Type checker recognizes struct declarations
- MIR builder has `struct.init` instruction  
- VM has `execStructInit` but it's incomplete
- C backend doesn't support structs
- **Missing**: Struct field access, methods, initialization

### 6. Enums ⚠️ **LOW**
**Status**: Partially implemented
**Details**:
- Type checker recognizes enum declarations
- No runtime support in either backend
- **Missing**: Enum value access, pattern matching

## Advanced Missing Features (Lower Priority)

### 7. Generics ⚠️ **LOW**
**Status**: Not implemented
**Details**:
- Type checker has some generic type parameter support
- No runtime support
- **Missing**: Generic function calls, type instantiation

### 8. Function Overloading ⚠️ **LOW**
**Status**: Not implemented
**Details**:
- Type checker doesn't support multiple functions with same name
- **Missing**: Overload resolution, dispatch

### 9. Union Types ⚠️ **LOW**
**Status**: Partially implemented
**Details**:
- Type checker has union type parsing
- No runtime support
- **Missing**: Union value creation, type narrowing

## Backend-Specific Issues

### VM Backend Limitations
1. **String comparisons** - Critical missing feature
2. **Float arithmetic** - High priority missing feature  
3. **Boolean comparisons** - Medium priority
4. **Map operations** - Incomplete implementation
5. **Struct field access** - Incomplete implementation

### C Backend Limitations
1. **Map support** - Not implemented
2. **Struct support** - Not implemented
3. **Enum support** - Not implemented

## Recommended Fix Priority

### Phase 1: Critical Basic Features (Immediate)
1. **Fix string comparisons in VM backend** - Enables basic string operations
2. **Add float support to VM backend** - Enables numeric calculations
3. **Add boolean comparisons to VM backend** - Completes basic type support

### Phase 2: Data Structures (Short term)
4. **Complete map implementation** - Both VM and C backends
5. **Complete struct implementation** - Both VM and C backends

### Phase 3: Advanced Features (Medium term)
6. **Enum runtime support**
7. **Generic function support**
8. **Function overloading**

## Implementation Notes

### String Comparisons Fix
The fix for string comparisons is straightforward:
1. Modify `execComparison` in `vm.go` to check operand types
2. Add string comparison logic similar to `std.string.equals`
3. Handle all comparison operators (`==`, `!=`, `<`, `>`, etc.)

### Float Support Fix
1. Add float support to `execArithmetic` function
2. Add float literal handling in `literalResult`
3. Add float comparison support in `execComparison`

### Boolean Comparisons Fix
1. Add boolean comparison logic to `execComparison`
2. Handle boolean literals properly

## Testing Strategy
For each fix:
1. Create comprehensive test cases
2. Test both VM and C backends
3. Ensure no regressions in existing functionality
4. Update e2e test suite

## Conclusion
The most critical missing feature is **string comparisons in the VM backend**. This is a fundamental language feature that should work in both backends. The other missing features are also important but can be prioritized based on user needs and development resources.
