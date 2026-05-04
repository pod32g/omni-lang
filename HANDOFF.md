# OmniLang dev session handoff

You're picking up an OmniLang work stream. This file is the entire context dump — the project memory files (`~/.claude/projects/-Users-pod32g-Documents-code-omni-lang/memory/*`) hold the same information cross-linked, but read this end-to-end first.

## Branch + identity

- **`dev` is the primary branch.** The session-start hint may say "main" — it's stale for this repo. Target dev for new branches, PRs, and "commit and push" requests.
- The user's commits are signed `Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>`.
- Concurrency keyword in OmniLang is `spawn`, not Go's `go`. Channel keywords stay as planned.

## Build + test cycle

The repo has a Go compiler frontend + Cranelift native backend in Rust:

```bash
make build                    # builds omni/bin/omnic, omnir, omnipkg + cargo release
cd omni && go test ./... -count=1 -timeout 600s   # full suite
```

Single-test runs that come up often:

```bash
go test ./tests/e2e/ -run "TestName" -count=1 -v -timeout 60s
go test ./internal/backend/c/ -run "TestStdNetworkIntrinsicWiring" -count=1
OMNI_NETWORK_TESTS=1 go test ./tests/e2e/ -run "TestStdNetworkHttpGetFixture" -count=1 -v
```

Network-touching tests (real HTTP via httptest) are gated behind `OMNI_NETWORK_TESTS=1`. Default CI skips them.

### Compiling a single .omni file by hand

Useful for ad-hoc probes. The library paths are fiddly:

```bash
DYLD_LIBRARY_PATH=omni/native/clift/target/release:omni/runtime/posix \
  omni/bin/omnic -backend c -emit exe path/to/file.omni
DYLD_LIBRARY_PATH=omni/native/clift/target/release:omni/runtime/posix \
  ./path/to/file
```

Same env for `omnir` (the VM):

```bash
DYLD_LIBRARY_PATH=omni/native/clift/target/release:omni/runtime/posix \
  omni/bin/omnir path/to/file.omni
```

If you need to inspect generated C, temporarily comment out `os.Remove(cPath)` in `omni/internal/compiler/compiler.go` (~7 occurrences). Always restore before committing.

If you need to inspect MIR:

```bash
omni/bin/omnic -backend vm -emit mir path/to/file.omni -o /tmp/file.mir
awk '/^func main/,/^func [^m]/' /tmp/file.mir
```

## Worktree pattern for dev work

The user prefers committing to dev via a separate worktree off `origin/dev`, not from the session worktree (which started on a stale branch). Standard sequence:

```bash
git fetch origin dev
git worktree add -b feat/your-branch /tmp/omni-feat origin/dev
cd /tmp/omni-feat/omni
go build -o bin/omnic ./cmd/omnic && go build -o bin/omnir ./cmd/omnir
(cd native/clift && cargo build --release)
# ... edit, test ...
git add <files>
git commit -m "..."
git push origin HEAD:dev
cd -
git worktree remove /tmp/omni-feat --force
git branch -D feat/your-branch
git fetch origin dev && git reset --hard origin/dev    # sync local to upstream
```

## Recent landed work (most recent first, all on `dev`)

| Commit | Summary |
|--------|---------|
| `6a1df67` | std.network API design review: every function categorized `[INTRINSIC]` / `[INTRINSIC-C]` / `[PURE]` in doc comments; intrinsic stub bodies fail loud via `_intrinsic_not_wired` (exit 99); new `TestStdNetworkIntrinsicWiring` static validator; ABI fixes for user-struct map fields and user fns returning std.network struct types; `TestStdNetworkABIReview` pins 22 cases. |
| `6450c2f` | std.json: tagged-union `omni_json_t` runtime + parser/stringifier replacing the half-done shim; std/json/json.omni; full kind/accessor/constructor surface; `TestStdJson` covers 28 cases. C-only — VM dispatch is a follow-up. |
| `785e40c` | Gated httptest fixture coverage for `http_get` and `http_post` (server returns 201 on byte-for-byte body match, 422 otherwise, so dropped bodies surface). Also fixed a third HTTP-call dispatch site that emitted unconditional fresh declarations. |
| `51e0792` | Pure-data HTTP request/response surface on C: `http_response_create` runtime intrinsic, status-class helpers, header chains. `mapType` learned the std.network concrete types. New `isHTTPRequest` field-access path. |
| `760add4` | Socket lifecycle smoke test (create/bind 127.0.0.1:0/listen/close), C-only. |
| `bb4839e` | VM `url_to_string` matches the OmniLang body's default-port-skip logic so the round-trip is stable across backends. |
| `d0d038d` | C-backend cleanup epilogue: assigns now transfer string ownership; cleanup deduplicates by C variable name so SSA aliases of returned heap values are protected. |
| `0e58284` | Returned-array and expression-parser regression tests. |

## Open work, ranked

### High-impact ABI / correctness

1. **`omni_dns_lookup` ABI mismatch.** Runtime is `omni_ip_address_t** omni_dns_lookup(const char* hostname, int32_t* count)` — takes a count out-pointer. C-backend generic call dispatch doesn't synthesize the count arg, so the call links short. Fix shape: dedicated call-site dispatch like the existing `omni_string_split` family — allocate a stack int, pass `&len`, register the slot as the array's runtime length via `arrayLengthExprs`. Documented in `std_network_abi_review.omni` as KNOWN BROKEN.
2. **VM map indexing on user struct fields.** `cfg.headers["k"]` errors `index: index must be int, got string` on omnir even when typed correctly. C side fixed via `omni_struct_get_map_field` (in 6a1df67); VM needs the analogous fix in `internal/vm/vm.go` index dispatch.
3. **std.network struct name qualification on cloning.** Cloned function decls keep return types like `array<IPAddress>` instead of `array<std.network.IPAddress>`. std.web's path qualifies; std.network's doesn't. Fix is mechanical — mirror the std.web qualifier loop in `compiler.go`'s `MergeImportedModules`.
4. **`std.collections.set/get` is map<string,int>-only.** No string-string overload exists. std.network bodies that call it (e.g. `http_response_set_header`'s `std.collections.set(resp.headers, name, value)` where headers is `map<string,string>`) only type-check today because the type checker skips body-check on namespaced functions (`!strings.Contains(d.Name, ".")` gate at `checker.go:471-475`). User code calling `std.collections.set` on a `map<string,string>` errors. Adding the overload is the clean fix; tightening the body-check skip surfaces the gap loudly but breaks the std.network hybrid pattern until then.
5. **Returned-array length tracking is via slice headers (`omni_slice_make` / `omni_slice_len_real`)** — works for runtime-allocated arrays but the broader story is in `arrayLengthExprs` / `userFuncReturnArrayLength` / `escapingArrays`. The current implementation handles the common cases; nested/chained returns may still drop length info.

### VM gaps

- VM has no `std.json` dispatch — every call errors `undefined identifier`. Cleanest impl: use Go's `encoding/json`, wrap the `interface{}` result in `Result`, dispatch kind/accessor helpers off a Go type-switch.
- VM has no socket implementation — `omni_socket_create` etc. all return `-1`/`false` stubs. Wire to `net.Dial` / `net.Listen` to make the socket smoke test work on both backends.
- VM has no real HTTP intrinsics — `http_get`/`http_post`/`http_put`/`http_delete` return a placeholder `HTTPResponse`. Real impls would let the gated networked tests run on VM too.

### More gated networked tests

The fixture-server pattern lives at `TestStdNetworkHttpGetFixture` / `TestStdNetworkHttpPostFixture`. Cloning is mechanical for: `http_put`, `http_delete`, `http_request` (chained builder), `dns_lookup`, `dns_reverse_lookup`, `network_ping`, `network_get_local_ip`. All are wired on C; once `omni_dns_lookup` ABI is fixed, these become straightforward.

### Stdlib audit backlog (from `project_stdlib_audit_progress.md`, ~1 week old, claims may need re-verification)

Pinned: `std.collections` queue/stack/set, `std.file` handles, `std.time`, `std.os` fs/env/cwd/pid, `std.io`, `std.network` (now), `std.json` (now C-only).

Next priority list:
1. `std.web` — basic routing/middleware/context audit (gzip/websocket/auth/session/rate_limit blocked on real libs).
2. `linked_list` / `binary_tree` / `priority_queue` — same audit shape as queue/stack/set.
3. `std.array.fill` / `copy` — flagged "needs in-place mutation ABI design; talk to user first."
4. Float arrays for `std.array` — mechanical extension of `emitStdArrayIntOp`.

Out of scope per memory: `std.dev` watch loops (real fsnotify), `std.web` partials needing real libs, `std.algorithms.is_connected` (no graph type), array+array concatenation in the type checker.

## Codebase gotchas (things I rediscovered the hard way)

### Three parallel registries for runtime intrinsics in the C backend

Every "intrinsic" std function lives in three places at once:

- `mapFunctionName(funcName)` switch in `omni/internal/backend/c/c_generator.go` (~line 7170+). Maps OmniLang qualified name → C symbol.
- `nameMapping` map in the same file (~line 8520+). Same mapping in different form, used by other dispatch paths.
- `runtimeProvided` map in `isRuntimeProvidedFunction` (~line 8870+). Gates whether the function definition is emitted at all (true → skip, the runtime symbol replaces it).

Plus `isStdFunctionRuntimeProvided` in `compiler.go` for the cloning-skip at body-load time.

Drop one and you get either silent stub-body fallthrough or duplicate-symbol link errors. **`TestStdNetworkIntrinsicWiring`** in `omni/internal/backend/c/intrinsic_wiring_test.go` is the static cross-check — extend it when you add new stdlib modules.

### Type checker skips body-check on namespaced functions

`internal/types/checker/checker.go:471-475`:

```go
case *ast.FuncDecl:
    if !strings.Contains(d.Name, ".") {
        c.checkFunc(d)
    }
```

This is what lets std.network's `http_response_set_header` call `std.collections.set` on `map<string,string>` even though `std.collections.set` only declares the `map<string,int>` overload. **Don't tighten this without first adding the missing overloads** — you'll break the entire std.network hybrid pattern.

### `MergeImportedModules` body-load list

`internal/compiler/compiler.go:256-265`:

```go
needsBodyLoad := qualified == "std.web" ||
    qualified == "std" ||
    qualified == "std.testing" ||
    qualified == "std.math" ||
    qualified == "std.collections" ||
    qualified == "std.network" ||
    qualified == "std.json"
```

If you add a new std module that has user-callable functions (not pure-intrinsic), add it here. Otherwise the type checker won't have return-type info for cross-module calls and they'll lower as `void`.

The `import std` umbrella also has its own list at line ~285 (`testing`, `math`, `collections`, `network`, `json`) — keep both in sync.

### C-backend dispatch tree is huge and has duplicates

There are at least three places that emit `omni_http_response_t* %s = omni_http_get(%s);` for HTTP calls (around lines 3820, 4046, 4525). When you change one, check whether the others need the same change. The recent fix in 785e40c only patched two of them; the third may still emit a redundant declaration in some edge cases.

### Pre-pass variable typing vs call-site emission

The C backend has a pre-pass at `generateFunction` start that walks all instructions and decides each SSA value's C variable declaration type. The call-site emission later writes the value. **They have to agree** — if pre-pass declares `omni_http_response_t* v0;` and call-site emits `omni_http_response_t* v0 = omni_http_get(...);`, you get a redefinition error. Guard call-site emissions with `g.declaredVariables[inst.ID]`:

```go
if g.declaredVariables[inst.ID] {
    g.output.WriteString(fmt.Sprintf("  %s = %s;\n", varName, rhs))
} else {
    g.output.WriteString(fmt.Sprintf("  TYPE %s = %s;\n", varName, rhs))
    g.declaredVariables[inst.ID] = true
}
```

### `assign` instructions create C-variable aliases

`var x = a; x = b` lowers as a chain of `assign` MIR instructions sharing storage with their target SSA value. The pre-pass at `generateFunction` resolves these aliases (see commit d0d038d's pre-resolve `assign` aliases block) so block traversal in MIR order doesn't read from an SSA id whose definition hasn't been emitted yet. If you see `error: use of undeclared identifier 'vN'` in generated C, this is usually the cause.

### Cleanup epilogue dedupes by C variable name

`stringsToFree` is keyed by SSA id, but multiple ids can map to the same C variable through `assign`. The epilogue resolves to variable names and dedupes — see commit d0d038d. If you add a new heap-tracking map (e.g. for some new resource type), follow the same pattern.

## Style conventions

- Type formatting: `map<string,string>` (no internal space, no trailing space). The trailing-space variant `map<string, string >` was a real source of bugs.
- Stdlib functions documented in network.omni-style markers: `[INTRINSIC]` / `[INTRINSIC-C]` / `[PURE]` doc comment before `func`. The static validator only checks std.network today; extend it as you add new modules.
- Stub bodies for unimplemented intrinsics should fail loud, not return defaults. Use `_intrinsic_not_wired(name)` (or your module's equivalent) which prints to stderr and exits with code 99.
- New regression tests go in `omni/tests/e2e/<feature>.omni` + a `Test*` entry in `e2e_test.go`. Run on both backends (`runVM` + `runCBackend`) when possible. C-only is acceptable when VM lags — note it in the test's doc comment.
- For tasks crossing module boundaries, update `omni/std/IMPLEMENTATION_STATUS.md` in the same commit.

## Memory files to read

- `~/.claude/projects/-Users-pod32g-Documents-code-omni-lang/memory/MEMORY.md` — index.
- `project_dev_is_primary_branch.md` — the dev-not-main rule.
- `project_stdlib_audit_progress.md` — what's pinned, what's queued (~1 week old, verify before quoting).
- `project_network_todo.md` — std.network state (current as of commit 6a1df67).
- `feedback_concurrency_keyword.md` — `spawn`, not `go`.

## Workflow expectation

The user prefers one focused commit per logical change. When asked to fix multiple issues, push them as separate commits in sequence rather than one big one. Memory files get updated alongside the relevant commit (or right after, if the work spawns new gaps).
