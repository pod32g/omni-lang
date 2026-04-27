# std.algorithms — Common Algorithms

The `std.algorithms` module provides sorts, searches, aggregates, and
distance metrics. Array-based functions take `array<int>` (other
element types are not yet specialized — file an issue if you hit
this). All functions that return arrays return a freshly allocated
copy; the input is never mutated.

Sorts and `reverse` / `rotate` / `shuffle` preserve the input length;
`unique` returns a runtime-determined length that the codegen tracks
through to downstream `len()` / index calls.

## Sorts

### bubble_sort(arr: array<int>): array<int>

Returns a freshly allocated copy of `arr` sorted in ascending order
using bubble sort with an early-exit when a pass produces no swaps.

```omni
import std

func main(): int {
    let arr: array<int> = [5, 2, 8, 1, 9, 3]
    let sorted: array<int> = std.algorithms.bubble_sort(arr)
    return sorted[0]  // 1
}
```

### selection_sort(arr: array<int>): array<int>

In-place selection sort applied to a freshly allocated copy.

### insertion_sort(arr: array<int>): array<int>

Insertion sort — fastest of the three on nearly-sorted input.

## Searches

### linear_search(arr: array<int>, target: int): int

Returns the first index where `target` appears, or `-1` if not found.
O(n).

### binary_search(arr: array<int>, target: int): int

Returns the index of `target`, or `-1` if not found. **Assumes `arr`
is sorted in ascending order**; behavior is undefined otherwise.
O(log n).

```omni
let sorted: array<int> = std.algorithms.bubble_sort(arr)
let idx: int = std.algorithms.binary_search(sorted, 8)
```

## Aggregates

### find_max(arr: array<int>): int

Returns the maximum element. Returns `0` if the array is empty (rather
than panicking — bounds-handling stays consistent with the rest of
the algorithm helpers).

### find_min(arr: array<int>): int

Returns the minimum element. Returns `0` if empty.

### count_occurrences(arr: array<int>, value: int): int

Returns the number of elements equal to `value`.

## Transforms

Each returns a freshly allocated array; the input is never modified.

### reverse(arr: array<int>): array<int>

Returns the elements in reverse order.

### rotate(arr: array<int>, k: int): array<int>

Rotates `arr` to the right by `k` positions. Negative `k` rotates left;
out-of-range `k` is reduced mod `n`.

```omni
let arr: array<int> = [1, 2, 3, 4, 5]
let rotated: array<int> = std.algorithms.rotate(arr, 2)
// [4, 5, 1, 2, 3]
```

### shuffle(arr: array<int>): array<int>

Fisher–Yates shuffle backed by the `std.math` PRNG. Seed it with
`std.math.random_seed(...)` first if you want reproducible output.

```omni
std.math.random_seed(42)
let shuffled: array<int> = std.algorithms.shuffle([1, 2, 3, 4, 5])
```

### unique(arr: array<int>): array<int>

Returns a freshly allocated array containing the first occurrence of
each distinct value, in input order. Output length is determined at
runtime; `len(unique(arr))` returns the actual count.

```omni
let dups: array<int> = [1, 2, 2, 3, 1, 4, 3, 5]
let u: array<int> = std.algorithms.unique(dups)
// [1, 2, 3, 4, 5], len = 5
```

## Distance Metrics

### euclidean_distance(x1: float, y1: float, x2: float, y2: float): float

Returns √((x2 - x1)² + (y2 - y1)²).

### manhattan_distance(x1: float, y1: float, x2: float, y2: float): float

Returns |x2 - x1| + |y2 - y1|.

### levenshtein_distance(s1: string, s2: string): int

Returns the edit distance (insertions + deletions + substitutions)
between two strings. Uses two-row dynamic programming, O(min(m, n))
memory.

```omni
let d: int = std.algorithms.levenshtein_distance("kitten", "sitting")
// 3
```

## Limitations

- **Element types**: array-based functions are int-typed. String-array
  variants (`contains`, `index_of`, `append`, etc.) live in
  [`std.array`](array.md). Float arrays are not yet specialized.
- **`is_connected`**: stub — needs a graph type the language doesn't
  have yet.

## See Also

- [`std.array`](array.md) — list operations on int and string arrays
- [`std.math`](math.md) — `random_seed` / `random_int` for shuffle
