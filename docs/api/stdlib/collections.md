# std.collections — Maps, Sets, and Friends

The `std.collections` module provides advanced data structures.
Today the **map basics** (`size`, `get`, `set`, `has`, `remove`,
`clear`) are real on both backends. Map composition (`keys`,
`values`, `copy`, `merge`) and the `set` / `queue` / `stack` /
`linked_list` / `binary_tree` / `priority_queue` families are
runtime-wired — the underlying C runtime exists and the OmniLang
signatures route to it, but coverage by tests is patchy. File an
issue if a particular operation misbehaves.

## Map literals

```omni
let scores: map<string, int> = {"alice": 95, "bob": 87}
```

Empty map literals (`{}`) need their type to be inferable from
context — explicitly typing the variable works:

```omni
var seen: map<string, int> = {"__sentinel__": 0}
std.collections.remove(seen, "__sentinel__")
// seen is now an empty, but-typed map
```

There's a known sharp edge here: `var counts: map<string, int> = {}`
fails type-checking because `{}` is treated as `<empty_map>` and
the checker can't unify it with the explicit annotation. The
sentinel-and-remove dance above is the workaround.

## Map operations

### size(m: map<K, V>): int

Returns the number of key-value pairs.

```omni
let scores: map<string, int> = {"a": 1, "b": 2}
std.collections.size(scores)  // 2
```

### get(m: map<string, int>, key: string): int

Returns the value associated with `key`, or `0` if not present.

### set(m: map<string, int>, key: string, value: int)

Stores or replaces the value associated with `key`. Returns nothing.

```omni
var counts: map<string, int> = {"first": 1}
std.collections.set(counts, "second", 2)
```

### has(m: map<string, int>, key: string): bool

Returns `true` if `key` is in the map.

```omni
if std.collections.has(scores, "alice") { ... }
```

### remove(m: map<string, int>, key: string): bool

Removes `key` from the map. Returns `true` if the key was present,
`false` otherwise.

### clear(m: map<string, int>)

Removes every entry from the map. The map struct itself stays
alive, so subsequent `set` calls keep working.

## Map iteration

`keys(m)` returns an `array<string>` of the keys; `values(m)`
returns an `array<int>` of the values:

```omni
let scores: map<string, int> = {"alice": 95, "bob": 87}
let names: array<string> = std.collections.keys(scores)
let totals: array<int> = std.collections.values(scores)
```

The order is not guaranteed.

## Map composition

`copy(m)` returns a freshly allocated copy of `m`. `merge(a, b)`
returns a new map containing every entry from `a`, with `b`'s
entries overriding any colliding keys.

## Worked example: word frequency counter

This shape — read input, split into words, count occurrences in a
map — is the canonical `std.collections` use case:

```omni
import std

func main(): int {
    let line: string = std.io.read_line()
    let words: array<string> = std.string.split(line, " ")

    var counts: map<string, int> = {"__sentinel__": 0}
    var i: int = 0
    while i < len(words) {
        let w: string = words[i]
        if std.collections.has(counts, w) {
            std.collections.set(counts, w, std.collections.get(counts, w) + 1)
        } else {
            std.collections.set(counts, w, 1)
        }
        i = i + 1
    }
    std.collections.remove(counts, "__sentinel__")

    return std.collections.size(counts)
}
```

```sh
$ echo "the quick brown fox the lazy dog the" | omnir wc.omni
6
```

## Queue (FIFO)

`queue<int>` is a first-in-first-out integer queue.

```omni
let q: queue<int> = std.collections.queue_create()
std.collections.queue_enqueue(q, 10)
std.collections.queue_enqueue(q, 20)
std.collections.queue_enqueue(q, 30)
std.collections.queue_size(q)        // 3
std.collections.queue_dequeue(q)     // 10
std.collections.queue_peek(q)        // 20 (next to dequeue)
std.collections.queue_is_empty(q)    // false
std.collections.queue_clear(q)
```

Operations: `queue_create`, `queue_enqueue`, `queue_dequeue`,
`queue_peek`, `queue_is_empty`, `queue_size`, `queue_clear`.

## Stack (LIFO)

`stack<int>` is a last-in-first-out integer stack.

```omni
let s: stack<int> = std.collections.stack_create()
std.collections.stack_push(s, 1)
std.collections.stack_push(s, 2)
std.collections.stack_push(s, 3)
std.collections.stack_pop(s)         // 3
std.collections.stack_peek(s)        // 2
std.collections.stack_size(s)        // 2
```

Operations: `stack_create`, `stack_push`, `stack_pop`, `stack_peek`,
`stack_is_empty`, `stack_size`, `stack_clear`.

## Set (unique members)

`set<int>` is a set of unique integers. `set_add` returns `true` if
the element was newly added (not previously present).

```omni
let s: set<int> = std.collections.set_create()
std.collections.set_add(s, 1)        // true
std.collections.set_add(s, 2)        // true
std.collections.set_add(s, 1)        // false (already present)
std.collections.set_size(s)          // 2
std.collections.set_contains(s, 2)   // true
std.collections.set_remove(s, 2)     // true (was present)
```

Operations: `set_create`, `set_add`, `set_remove`, `set_contains`,
`set_size`, `set_clear`, `set_union(a, b)`, `set_intersection(a, b)`,
`set_difference(a, b)`. The set-algebra operations return a freshly
allocated `set<int>`.

## linked_list, binary_tree, priority_queue

These are wired to the runtime
(`omni_linked_list_*`, `omni_binary_tree_*`, `omni_priority_queue_*`)
but don't yet have first-class VM intrinsics or e2e regression
coverage. Treat as experimental — file an issue if you hit a sharp
edge.

Operations follow the same naming pattern:
- `linked_list_create / add / remove / get / size / is_empty / clear`
- `binary_tree_create / insert / search / size / is_empty / clear`
- `priority_queue_create / enqueue / dequeue / peek / size / is_empty / clear`

## Limitations

- Maps are typed as `map<string, int>` in most signatures. Other
  key/value combinations are present in the runtime
  (`omni_map_get_int_int`, `omni_map_get_string_string`, etc.) but
  the OmniLang layer's signatures don't expose them.
- The empty-map literal type-check sharp edge described above.

## See Also

- [`std.string.split`](string.md) — produces the `array<string>`
  most map-keyed programs feed in
- [`std.array`](array.md) — list operations on the keys/values arrays
