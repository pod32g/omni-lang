# std.array — Array list operations

The `std.array` module provides searches, mutations (returned as
fresh copies), and slicing for array values. The functions are
specialized per element type at the C runtime layer; today `array<int>`
and `array<string>` are wired end-to-end.

Returned arrays are freshly allocated — the input is never modified.
The C codegen forwards the output length onto the result so a
downstream `len()` / index works without bookkeeping.

## Length and indexing

`std.array.length`, `std.array.get`, `std.array.set` are kept around
as compatibility shims, but the canonical way to read an array's
length and elements is the language built-ins:

```omni
let arr: array<int> = [10, 20, 30]
let n: int = len(arr)        // 3
let first: int = arr[0]      // 10
arr[0] = 99                  // mutation through index
```

`len(arr)` works on parameters too, thanks to the implicit
`__omni_len_<name>` companion the C ABI passes alongside every
`array<T>` argument.

## Membership

### contains(arr: array<T>, value: T): bool

Returns `true` if `value` appears anywhere in `arr`. String compare
uses byte-wise equality.

```omni
let names: array<string> = ["alice", "bob", "carol"]
if std.array.contains(names, "bob") { ... }
```

### index_of(arr: array<T>, value: T): int

Returns the index of the first occurrence of `value`, or `-1` if not
found.

## Mutators (returning fresh copies)

### append(arr: array<T>, value: T): array<T>

Returns a new array with `value` appended.

```omni
let arr: array<int> = [1, 2, 3]
let bigger: array<int> = std.array.append(arr, 4)
// [1, 2, 3, 4]
```

### prepend(arr: array<T>, value: T): array<T>

Returns a new array with `value` at index 0.

### insert(arr: array<T>, index: int, value: T): array<T>

Returns a new array with `value` inserted at `index`. Out-of-range
indices clamp to `[0, len(arr)]`.

### remove(arr: array<T>, index: int): array<T>

Returns a new array with the element at `index` removed. Out-of-range
indices return a copy of the input unchanged.

### concat(a: array<T>, b: array<T>): array<T>

Returns the concatenation of `a` followed by `b`. Output length is
`len(a) + len(b)`.

```omni
let combined: array<string> = std.array.concat(["a", "b"], ["c", "d"])
// ["a", "b", "c", "d"]
```

### slice(arr: array<T>, start: int, end: int): array<T>

Returns the half-open slice `[start, end)`. `start < 0` clamps to 0;
`end > len(arr)` clamps to `len(arr)`; `end < start` produces an
empty array. Output length is `end - start`.

```omni
let middle: array<int> = std.array.slice([1, 2, 3, 4, 5], 1, 4)
// [2, 3, 4]
```

## Element type support

| Operation     | `array<int>` | `array<string>` | other |
|---------------|--------------|-----------------|-------|
| contains      | ✅           | ✅              | ❌    |
| index_of      | ✅           | ✅              | ❌    |
| append        | ✅           | ✅              | ❌    |
| prepend       | ✅           | ✅              | ❌    |
| insert        | ✅           | ✅              | ❌    |
| remove        | ✅           | ✅              | ❌    |
| concat        | ✅           | ✅              | ❌    |
| slice         | ✅           | ✅              | ❌    |

Other element types fall through to a passthrough that returns the
input unchanged. File an issue if you need a specific element type.

## Limitations

- **`fill` / `copy`**: still stubs. They mutate a destination through
  a parameter, which needs an in-place mutation pattern the C ABI
  doesn't model yet.
- **String arrays alias the input pointers**: `omni_array_str_*`
  functions only deep-copy the pointer table, not the payload
  strings. Mutating the original string contents through aliases
  would be visible — but OmniLang strings are immutable from the
  language layer, so this is invisible in practice.

## See Also

- [`std.algorithms`](algorithms.md) — sorts, searches, aggregates,
  shuffle, unique, distance metrics
- [`std.string.split`](string.md) — produces an `array<string>`
- [`std.collections`](collections.md) — maps and other compound types
