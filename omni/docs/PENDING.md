# Pending Work

This file is the running ledger of everything *not* shipped yet — open
bugs, known-but-unfiled quirks, and phases the original plan flagged as
out-of-scope. The Go-foundations slice (methods, structural interfaces,
defer, slices, concurrency including TCO and `select`) is complete
end-to-end through both the VM (`omnir`) and the C backend (`omnic`);
this document catalogs what comes after.

Last updated: 2026-04-23. The most recent commit on `dev` is the
authoritative source — when in doubt, check `git log` first.

---

## Filed follow-ups

### `std.math.pow_int` codegen miss
The C backend emits a `return v18;` where `v18` is uninitialized for
the post-loop tail of `std.math.pow_int(x, y)` (defined in
`omni/std/math/math.omni`). Surfaces as a single
`-Wuninitialized` warning when compiling any program that imports
`std`, even FizzBuzz which never calls the function.

The Omni source is correct (`return result` after the `while exp > 0`
loop). The miss is in the SSA → C lowering — likely the same
phi-merge issue the rest of this document touches on, but isolated to
this one function. Worth fixing on its own before it grows roots.

A self-contained repro plus prior analysis is queued as a UI chip
spawned on 2026-04-23.

---

## Known quirks (not blocking, not yet filed)

### `std.collections.*` stub bodies use stack returns
A handful of stub bodies in `std/collections/` (`keys`, `values`,
`collection_to_array`) used to allocate an array on the stack and
return its address. Phase 4's heap-slice rework fixed the ones that
fire today, but the stub bodies still exist. If a future call site
exercises them in a shape the rework didn't cover, they'll surface
as `-Wreturn-stack-address` again. Keep `-Wreturn-stack-address` on
in the C compile flags; revisit when a real stdlib `collections`
implementation lands.

---

## Out of scope for the Go-foundations slice

These are explicit follow-up phases the original plan flagged. None
are blockers; they're the "what comes after" wishlist.

### Tooling
- **`omnifmt`** — canonical formatter.
- **`omnivet`** — lints (unused vars, shadowing, unreachable, etc.).
- **`omnils`** — language server / LSP.
- **`omni.mod`** — proper module manifest + resolver. Replaces the
  current `omnipkg` archiver shape.

### Runtime
- **Tracing GC** — first pass via Boehm to remove the manual `delete` /
  leak surface. Then a custom collector if needed.
- **`panic` / `recover`** with stack traces. The current `throw` /
  `try` / `catch` / `finally` plumbing exists, but real panic
  semantics + unwinding aren't wired.
- **Cross-compilation via Cranelift** — `-target` flag, ELF/Mach-O/PE
  emission. The Cranelift backend exists but is minimal.

### Stdlib breadth
- `bufio` (buffered reader/writer over the new `io.Reader` / `io.Writer`
  interfaces).
- Full `net/http` server built on a `Handler` interface — currently
  the HTTP surface is a flat builtin set.
- `encoding/json` with struct tags and reflection.
- `regexp` package wrapping the existing POSIX regex runtime.
- `context.Context` — cancellation, deadlines, value-passing.
- `strconv`, `reflect`, `crypto`, `os/exec`, `flag`.

### Language
- **Tagged enums + `match`** — sum types with exhaustiveness checking.
  The current `enum` is C-style (just variant names).
- **Visibility rules** — public vs private at the package boundary.
- **`const` + `iota`** — compile-time constants and enumerated
  generators.
- **Proper generic monomorphization** — the type-param infrastructure
  lands a function-per-instantiation pass; today generics flow
  through but don't always specialize.
- **`select` C backend** — DONE (2026-04-23).

---

## How to contribute to this list

When you spot something that should land here:

1. If it's a real bug with a repro, file it in this doc under
   "Filed follow-ups" with a one-paragraph summary, repro steps, and
   a hypothesis for where in the codegen / runtime it likely lives.
2. If it's a "this works but feels wrong" observation, put it under
   "Known quirks" with the workaround you used.
3. If it's a feature the original plan listed but we never built, add
   it to "Out of scope" — it's not a bug, it's the roadmap.

Closing a follow-up: delete its section from this file in the same
commit that fixes it. The git history keeps the receipt.
