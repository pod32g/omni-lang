# Nested Generic Types - Current Workaround

## Issue

Currently, OmniLang's parser has a limitation with nested generic types due to the lexer treating `>>` as a right-shift operator instead of two closing angle brackets.

```omni
// This does NOT work currently:
let matrix:array<array<int>> = []
```

## Workaround

Until nested generic support is fully implemented, use a space before the closing `>`:

```omni
// This WORKS:
let matrix:array<array<int> > = []
let nested_map:map<string, map<string, int> > = {}
let complex:array<map<string, array<int> > > = []
```

## Implementation Status

The fix requires:
1. ✅ Parser enhancement to handle `>>` in generic contexts
2. ⏳ Type system updates for nested type resolution  
3. ⏳ MIR support for nested data structure operations
4. ⏳ Runtime support for nested collections
5. ⏳ Comprehensive testing

## Alternative: Use Type Aliases

For complex nested types, use type aliases to improve readability:

```omni
type IntArray = array<int>
type Matrix = array<IntArray>

func main() {
    let matrix:Matrix = []
}
```

## Tracking

See GitHub issue #XXX for implementation progress.

