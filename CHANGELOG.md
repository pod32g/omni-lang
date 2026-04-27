# Changelog

All notable changes to OmniLang will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Go-foundations slice complete**: methods on user types, structural
  interfaces with method dispatch, `defer` statements (including
  lambda- and interface-typed deferred calls), `append` builtin and
  `s[lo:hi]` slicing on heap-allocated arrays, `spawn` for goroutines
  + buffered channels with `select`, multi-return functions, channel
  ok-form receive (`v, ok := <-ch`), `close(ch)`, and tail-call
  optimization (self-recursion turns into `goto entry`, cross-function
  tail calls become `return f(...)` for clang's sibling-call
  optimization). Both `omnir` (VM) and `omnic` (C backend) are kept in
  lockstep across all of these.
- **Char ↔ int interop**: `std.char_code(c) -> int`,
  `std.char_from_code(i) -> char`, and `std.char_to_string(c) -> string`
  expose a char's code point. Char now maps to `int32_t` in the C
  backend (was `omni_struct_t*`), so chars can be compared, arithmetic
  results survive across function calls, and char literals emit
  correctly.
- **Top-level `let` declarations**: module-scope constants
  (`let LOWER:string = "abc..."`) are visible to every function in
  the module. Implemented via lazy materialization into each
  referencing function's entry block — keeps unrelated functions
  (notably std.* stub bodies) free of unused materialized constants.
  Top-level `var` is rejected with a clear error (no global storage
  yet).
- **Array length through function parameters**: every `array<T>`
  parameter now gets a synthetic `int32_t __omni_len_<name>`
  companion at the C ABI level. `len(arr)` works on parameters, sorts
  and reverses forward the input length onto the result, and
  std.algorithms / std.array operations can compute over the
  caller's array directly.
- **std.string completions**:
  - `trim_left`, `trim_right`, `trim_all`
  - `to_title`, `capitalize`, `reverse`
  - `equals_ignore_case`, `compare_ignore_case`
  - `count_occurrences`, `count_lines`, `count_words`
  - `is_empty`
  - Real implementations for the previously-aspirational `split`,
    `split_lines`, `split_words`, `join`, `replace` / `replace_all`
    / `replace_first` / `replace_last`, and `find_all` (the OmniLang
    bodies referenced `array<T> + []<T>` which the type checker
    rejects, so they never actually compiled).
- **std.algorithms completions**:
  - Distance metrics: `euclidean_distance`, `manhattan_distance`,
    `levenshtein_distance` (two-row DP, O(min(m,n)) memory)
  - Sorts: `bubble_sort`, `selection_sort`, `insertion_sort`
    (return a freshly allocated sorted copy)
  - Searches: `linear_search`, `binary_search`
  - Aggregates: `find_max`, `find_min`, `count_occurrences`
  - Transforms: `reverse`, `rotate(arr, k)`, `shuffle(arr)`,
    `unique(arr)` (the last via a runtime out-pointer for the
    runtime-determined result length)
- **std.array list operations** for `array<int>` and `array<string>`:
  `contains`, `index_of`, `append`, `prepend`, `insert`, `remove`,
  `concat`, `slice`. Each returns a freshly allocated array; the
  C codegen forwards the output length so a downstream `len()` /
  index keeps working.
- **std.collections map basics**: `size`, `get`, `set`, `has`,
  `remove`, `clear` are now real on both backends (previously the
  C side had no wiring and the VM ran stub bodies that returned
  defaults). `omni_map_has_string`, `omni_map_remove_string`
  (bool-returning wrapper around the void delete), and
  `omni_map_clear` are new runtime helpers.
- **std.os.* string-returning calls** (`positional_arg`, `getenv`,
  `read_file`, `getcwd`, ...) now return `string` correctly when
  bound with `let s:string = ...`. Previously the MIR builder typed
  the call as `void` and the C backend dropped the return value into
  a `<unknown>` placeholder.
- **std.os.positional_arg** respects POSIX `--` end-of-flags and
  treats `-N` (digit-prefixed) and bare `-` (stdin sentinel) as
  positional, so `caesar_cipher -3 ...` and similar invocations work.
- **std.file handle operations** are now covered end-to-end on both
  backends. The C backend stores `open` results in pointer-sized
  slots so native `FILE*` handles are not truncated on 64-bit
  platforms.
- **std.time runtime core** now has VM and C backend parity coverage
  for UTC Unix/RFC3339 conversions, duration formatting, sleep, and
  timezone helpers. The audit also records the remaining C-backend gap
  for pure Omni helper bodies such as duration arithmetic.
- **File-handle wrappers (Go os/bufio analogs)**: `std.file` gains
  `read_all(handle)`, `read_line(handle)`, `write_string(handle, s)`
  — fills the existing hole where `read(handle, ...)` couldn't
  return content because OmniLang strings are immutable. `std.io`
  gains `fprint(handle, value)`, `fprintln(handle, value)`,
  `fprintf(handle, format, args)` mirroring Go's `Fprint(w, ...)`.
  `std.os` gains `read_file_lines(path)` and
  `write_file_lines(path, lines)` for the very common whole-file
  line workflow. All wired through both backends. Pinned by
  `TestStdFileHandles` (round-trips a fixture file through every
  new helper).
- **VM string-literal escape decoding**: `"a\nb"` used to measure
  length 4 on the VM and 3 on the C backend, because the VM's
  `literalResult` stripped the wrapping quotes but never decoded
  escape sequences — the lexer only validated them. The C backend
  got correct decoding for free by forwarding the raw lexeme to the
  C compiler. The VM now decodes the same set the lexer accepts
  (`\n`, `\t`, `\r`, `\\`, `\"`, `\'`, `\0`, `\xHH`, `\uHHHH`), so
  string semantics match across backends. Pinned by
  `TestStringEscapes`. NUL-byte-in-string (`\0` mid-string) still
  diverges — `omni_strlen` truncates at NUL on the C side; that's
  a C-string-representation choice, not in scope here.
- **std.io: printing, prompting, ANSI styling, formatted writes**.
  Adds `printf` / `eprintf` (sprintf + stdout/stderr), `print_each` /
  `eprint_each` (one line per array entry), `eprompt` (prompt to
  stderr — keeps stdout pipes clean), `confirm` (y/n prompt),
  `flush_stderr`, plus ANSI styling: a generic `style(s, code)` and
  named shortcuts `bold`, `dim`, `italic`, `underline`, `red`,
  `green`, `yellow`, `blue`, `magenta`, `cyan`. Helpers always emit
  the SGR escape codes — call sites should gate on `is_terminal()`
  when output may be piped. Pinned by `TestStdIoFormat` and
  `TestStdIoConfirm` (the latter pipes both `y\n` and `n\n` to
  verify the predicate's truthy/falsy paths on both backends).
- **std.io expansion (Go fmt/bufio analogs)**: `is_terminal`,
  `read_lines`, `read_int`, `read_float`, `prompt`, `sprint`,
  `sprintln`, `sprintf` (`%s` only), `parse_int`, `parse_float`,
  `is_int`, `is_float`. Surface mirrors Go's `fmt` + `bufio` + `io`
  as far as makes sense without varargs, byte arrays, or Reader/
  Writer interfaces. Wired through both backends — runtime adds
  `omni_io_is_terminal`, `omni_io_sprintf`, `omni_io_sprintln`,
  `omni_io_prompt`, `omni_io_read_lines`, `omni_io_parse_int`,
  `omni_io_parse_float`, `omni_io_is_int`, `omni_io_is_float`.
  Pinned by `TestStdIoExtras` and `TestStdIoReadLines`.
- **std.network: URL parsing + struct field access on C backend**.
  Several issues stacked together prevented `url_parse(...)`'s
  fields from being read on the C backend:
  - `omni_url_parse` dumped path+query+fragment into `path`. Now
    splits them on `?` and `#` correctly, so `u.path`, `u.query`,
    `u.fragment` carry their proper sections.
  - The MIR builder had no return-type heuristic for `std.network.*`,
    so calls came back as `inst.Type = "void"` and the C-backend
    member-access path fell through to the generic
    `omni_struct_get_int_field((omni_struct_t*)v0, "scheme")`
    accessor, which is wrong for the concrete `omni_url_t*` /
    `omni_ip_address_t*` / `omni_http_response_t*` runtime structs.
    Added a network branch to the builder's name-based heuristic and
    extended the C-backend hoisting + member-access codegen to emit
    direct field reads (`v0->scheme`, `v0->host`, `v0->port`,
    `v0->path`, `v0->query`, `v0->fragment`, plus the IPAddress
    fields). Pinned by an expanded `TestStdNetworkBasic`.
- **std.network audit** pinned by `TestStdNetworkBasic`. Two real
  validation bugs in the C runtime:
  - `omni_ip_is_valid` previously accepted IPv4 strings whose
    segments were out of range (e.g. `999.999.999.999`, `256.0.0.0`).
    Replaced with strict per-segment validation. The IPv6 check was
    "contains a colon → valid"; replaced with a real grammar
    (≤8 hex groups of 1–4 chars, at most one `::` elision, optional
    embedded IPv4 tail) so junk like `hello:world` and `1::2::3` is
    rejected.
  - `omni_url_is_valid` previously returned true for any string
    containing `://`. Now requires a real scheme
    (`[A-Za-z][A-Za-z0-9+.-]*`) followed by `://` and a non-empty
    host before any of `/?#`; whitespace is rejected.
  Network-touching parts (HTTP, DNS, sockets) wired but not
  exercised in CI; the C-backend struct-access gap for `URL`/
  `HTTPResponse` is a known pre-existing issue tracked separately.
- **Empty-map literal in typed `let`/`var`**: writing
  `let h: map<string, string> = {}` used to fail with
  `cannot assign <empty_map> to map<string, string>`. The empty-map
  placeholder type is now coerced to the declared `map<...>`
  annotation in the binding-statement type checker (it already worked
  in function arguments).
- **std.io expansion + audit** pinned by `TestStdIoBasic` and
  `TestStdIoRead`. Added `eprint(value)`, `eprintln(value)` (stderr
  variants of print/println), `flush()` (force stdout flush), and
  `read_all()` (slurp stdin to EOF). All wired through both backends:
  C runtime gains `omni_eprint_string`, `omni_eprintln_string`,
  `omni_io_flush`, `omni_read_all`; the VM has matching intrinsic
  cases. The pre-existing `std.io.read_line` already returned a
  heap string; `read_all` joins it in `stringsToFree` and gets the
  same `const char*` declaration treatment.
- **Returning `array<T>` from a user-defined function** now works on
  both backends. On the C backend the call result lost its length
  companion (the parameter ABI's `__omni_len_<name>` doesn't apply to
  return values), so `std.array.get` and indexed reads aborted with
  `length=-1`. The codegen now registers
  `(int32_t)omni_slice_len_real((void*)v)` as a runtime length
  expression for array-typed call results, and the bounds-checked
  ops (`omni_array_get_int`/`omni_array_set_int`, plus the int-array
  index lowering) consult `arrayLengthExprs` in addition to the
  compile-time `arrayLengths` map. On the VM, `std.array.get`'s stub
  body unconditionally returned `arr[0]` regardless of the requested
  index — replaced with `arr[index]`. Pinned by `TestArrayReturnLength`.
- **std.os audit** pinned by `TestStdOsOps`: mkdir/rmdir, write/read/
  append, copy/rename/remove, exists/is_file/is_dir, set/get/unsetenv,
  getcwd, and getpid all verified on both backends. `omni_getenv` and
  `omni_getcwd` previously returned `NULL` on missing-var / failed
  syscall, which segfaulted the C backend when callers chained the
  result into `std.string.*`. Both now return `""` to match the
  OmniLang `string` return type.
- **std.math.random_seed / random_int**: shared xorshift32 PRNG
  across both backends so the same seed produces the same first
  output. Used by `std.algorithms.shuffle`.
- **MIR builder: signature-first call typing**. When `fb.sigs` knows
  a function's return type the builder prefers it over the
  name-based heuristic, eliminating an entire class of "stdlib call
  silently does nothing" bugs.
- **VM `var`-merge fix**: a `var` reassigned in both arms of an `if`
  inside a loop now reads the same storage slot across iterations.
  The MIR builder used to drift `env[var]` to each assign's result
  id; now it stays pinned to the original SSA slot, which both
  backends already mutate in place.
- **VM tail-call detector** skips rebinding into single-const stub
  bodies and into known intrinsic overrides (`std.algorithms.*`,
  `std.array.*`, `std.string.split` family, `std.collections.*`).
  Calls to those go through `execIntrinsic` instead of falling back
  to the OmniLang stub body.
- **Bundled examples**: `examples/caesar_cipher.omni` is a worked
  example covering the new char-int intrinsics, top-level `let`,
  `std.os.positional_arg`, and `std.io.read_line`.

### Documentation
- Updated `docs/api/stdlib/string.md` to cover the trim / case / count
  / split / join / replace / find_all / is_empty additions.
- Added `docs/api/stdlib/algorithms.md`, `docs/api/stdlib/array.md`,
  `docs/api/stdlib/collections.md`, `docs/api/stdlib/file.md`, and
  `docs/api/stdlib/time.md` covering the newly-real modules.
- Added a "What's new" section to `docs/spec/language-tour.md`
  pointing at the recently-landed language features.
- `omni/std/IMPLEMENTATION_STATUS.md` now matches reality: every
  function listed as `[IMPLEMENTED]` actually runs end-to-end on
  both backends and is exercised by an e2e regression.
- `omni/docs/PENDING.md` continues as the running ledger of
  filed bugs, known quirks, and out-of-scope phases.

### Changed
- The C backend's array-with-length call ABI changed: every
  function whose signature includes one or more `array<T>`
  parameters now also takes synthetic `int32_t` length companions
  after each. Recompile dependents — old-ABI calls from older
  builds will fail to link.
- The C backend now treats `char` as a primitive (maps to
  `int32_t`). Any user code that relied on `char` being a struct
  type will need to update.

### Documented
- Documented `std.log` structured logging module, including examples and configuration guidance.
- Added logging quick-start examples and references across the documentation set.

### Planned
- Advanced type system with generics and union types
- MIR optimization passes and advanced features
- Complete native code generation with Cranelift
- Language features: concurrency, error handling, metaprogramming
- Production readiness: performance optimization, security, deployment
- Ecosystem development: package registry, community building

## [0.5.1] - 2025-10-24

### Added
- **String Interpolation**: `${expression}` syntax for dynamic string creation with automatic type conversion
- **Exception Handling**: Complete try-catch-finally blocks with exception propagation and cleanup
- **Advanced Type System**: Type aliases, union types, and optional types with full type checking
- **Type Aliases**: `type UserID = int` syntax for better code readability and semantic meaning
- **Union Types**: `string | int | bool` syntax for flexible data handling with multiple types
- **Optional Types**: `int?`, `string?` syntax for nullable values with type safety
- **Improved Module Loading System**: Binary-relative path resolution for std library
- **Environment Variable Support**: `OMNI_STD_PATH` environment variable for custom std library locations
- **Debug Support**: New `-debug-modules` flag for troubleshooting module loading issues
- **Comprehensive Documentation**: Complete API documentation for all advanced features
- **Logo Integration**: Professional logo branding throughout documentation
- **Consistent Behavior**: Module search behavior is now consistent regardless of working directory

### Fixed
- **String Concatenation**: Fixed C backend string concatenation with mixed types (int/float to string)
- **Type Conversion**: Fixed automatic type conversion in string operations
- **Exception Handling**: Fixed runtime exception throwing and catching mechanism
- **Type System**: Fixed type alias resolution and union type compatibility checking
- **Critical Module Loading Bug**: Fixed issue where compiler was looking for std library relative to current working directory instead of binary location
- **Return Type Issues**: Resolved `std.math.max/min/abs` returning `void` instead of `int`
- **CI/CD Pipeline**: Fixed failing code-quality, fmt, and lint jobs
- **Circular Dependencies**: Resolved import cycle between compiler and type checker packages

### Changed
- **Architecture**: Created separate `moduleloader` package to avoid circular dependencies
- **Error Messages**: Improved debugging capabilities and error reporting
- **Module Resolution**: Prioritizes `OMNI_STD_PATH`, binary-relative paths, then current working directory
- **Documentation**: Updated all documentation with advanced features and logo branding
- **Type System**: Enhanced type checking with support for advanced type features

## [0.5.0] - 2025-10-23

### Added
- **Comprehensive Standard Library Testing**: Complete test suite for all std modules (math, io, string, array, collections, os, file)
- **Integration Testing**: Added integration tests for std library usage patterns
- **Enhanced CI/CD Pipeline**: Improved error handling, debugging, and robustness
- **Standard Library Documentation**: Comprehensive documentation for all std modules

### Fixed
- **std.math Return Types**: Fixed missing return type annotations for max, min, abs functions
- **CI Pipeline Issues**: Resolved path resolution and test execution problems
- **VM e2e Tests**: Fixed path resolution issues in VM test execution
- **Type Checker**: Fixed type checker and VM runtime issues
- **C Generator**: Fixed placeholder function generation issues

### Improved
- **Test Coverage**: Comprehensive test coverage for all standard library modules
- **Error Reporting**: Enhanced debugging capabilities and error reporting
- **Development Workflow**: Improved CI/CD pipeline with better error handling
- **Code Quality**: Enhanced code organization and testing framework

## [0.4.3] - 2025-10-21

### Fixed
- **Library Path**: Fixed Makefile to use absolute paths with install_name_tool
- **macOS Compatibility**: Resolved dyld library loading error when running binaries from different directories
- **Binary Usability**: Binaries now work correctly when called from any directory in PATH
- **Build Process**: Changed from $(PWD) to $$(pwd) for proper shell variable expansion

## [0.4.2] - 2025-10-21

### Fixed
- **Negation Test**: Fixed TestNegation to expect -5 instead of 251
- **Test Expectations**: Corrected test expectation to match improved test runner behavior
- **CI/CD Tests**: Resolved remaining CI/CD test failure in TestNegation

## [0.4.1] - 2025-10-21

### Fixed
- **Test Runner**: Fixed test runner to prioritize stdout parsing over exit code interpretation
- **CI/CD Tests**: Resolved issue where large positive return values (≥256) were being interpreted as exit codes instead of actual program results
- **Map Tests**: Fixed CI/CD test failure in TestMapComprehensive

## [0.4.0] - 2025-10-21

### Added
- **Array Support**: Complete array implementation with literals, indexing, and length function
- **Array Method Syntax**: Support for `x.len()` method-style calls instead of `len(x)`
- **Map/Dictionary Support**: Basic map implementation with key-value operations
- **Struct Support**: Struct declarations, literals, and field access
- **PHI Node Support**: Proper SSA form with PHI nodes for control flow
- **Enhanced Comparisons**: String, boolean, and float comparisons in VM backend
- **Complete C Backend**: Support for all core MIR instructions (mod, neg, not, and, or, strcat)
- **Comprehensive Testing**: Extensive e2e tests for all new language features

### Fixed
- **Infinite Loop Bug**: Fixed critical bug in constant folding optimization that caused infinite loops
- **MIR Verifier**: Enhanced with support for new instruction types
- **Type Checking**: Improved validation for arrays and builtin functions
- **Control Flow**: Better handling of complex loop and conditional scenarios

### Changed
- **Backend Parity**: VM and C backends now support the same core instruction set
- **Code Generation**: Improved quality and reliability of generated code
- **Documentation**: Updated with new features, examples, and implementation guides

## [0.3.2] - 2025-01-21

### Fixed
- **CI/CD Pipeline Issues**: Resolved shared library loading errors in CI environment
- **Library Path Configuration**: Fixed `LD_LIBRARY_PATH` in e2e tests to include Cranelift library path
- **Test Framework**: Updated `runVM` and `runCBackend` functions to properly locate `libomni_clift.so`
- **Cross-Platform Compatibility**: Ensured both `DYLD_LIBRARY_PATH` (macOS) and `LD_LIBRARY_PATH` (Linux) are correctly set

### Technical Improvements
- **Build System**: Enhanced Makefile to include Cranelift library path in test environment
- **Test Infrastructure**: Improved e2e test reliability across different environments
- **CI/CD Reliability**: Fixed pipeline failures that were preventing automated testing

## [0.3.1] - 2025-01-21

### Fixed
- **Critical Infinite Loop Bugs**: Fixed infinite loops in nested for loops and range loops
- **Assignment Instruction Generation**: Added proper assignment instruction generation in MIR builder
- **Variable Mutability**: Fixed variable assignment handling for mutable variables
- **Loop Variable Updates**: Fixed increment statements (i++, j++) to generate proper assignments
- **Range Loop Index**: Fixed range loop index increment assignments
- **Test Framework**: Fixed e2e test framework timeout issues and environment setup
- **Performance Tests**: Fixed performance benchmark code generation to match current capabilities

### Enhanced
- **MIR Builder**: Enhanced MIR builder to generate explicit assignment instructions
- **C Backend**: Improved C backend to handle assignment instructions correctly
- **Test Framework**: Enhanced test framework with better error handling and environment setup
- **Safe Testing**: Added safe_run.sh script for testing with timeout protection
- **Performance Validation**: All performance benchmarks now pass consistently

### Technical Improvements
- **Assignment Instructions**: Added new "assign" MIR instruction for explicit variable assignments
- **MIR Verifier**: Updated MIR verifier to recognize assignment instructions
- **Runtime Environment**: Improved runtime library loading in test environments
- **CGO Handling**: Better CGO handling in test environments with stub implementations
- **Error Recovery**: Enhanced error recovery and debugging in test framework

### Infrastructure
- **Safe Runner**: Added safe_run.sh script for preventing infinite loops during testing
- **Test Environment**: Improved test environment setup with proper PATH and LD_LIBRARY_PATH
- **Performance Monitoring**: Fixed performance test suite to provide accurate benchmarks
- **Build System**: Enhanced build system with better test isolation

## [0.3.0] - 2025-01-20

### Added
- **Enhanced Debug Support**: Comprehensive debug symbol generation with source mapping
- **C Backend**: Native code generation with C backend for improved performance
- **Package System**: Complete packaging and distribution system with tar.gz and zip support
- **Performance Testing**: Comprehensive performance regression testing and benchmarking
- **Source Maps**: Debug source mapping for better debugging experience
- **Cross-Platform Support**: Enhanced platform and architecture support
- **Performance Monitoring**: Automated performance monitoring and regression detection

### Enhanced
- **Debug Information**: Full debug symbol generation with DWARF support
- **Compilation Performance**: Improved compilation speed with optimization levels
- **Documentation**: Comprehensive documentation for C backend and packaging features
- **Build System**: Enhanced Makefile with performance testing targets
- **Error Reporting**: Better error messages with source location information

### Technical Improvements
- **Debug Symbols**: Enhanced debug symbol generation in C backend
- **Source Mapping**: Source map generation for debugging tools
- **Performance Metrics**: Detailed performance tracking and regression detection
- **Package Management**: Robust packaging system with multiple formats
- **Cross-Platform**: Improved cross-platform compilation support
- **Memory Management**: Better memory usage tracking and optimization

### Documentation
- **C Backend Guide**: Complete documentation for C backend usage
- **Packaging Guide**: Comprehensive packaging and distribution documentation
- **Performance Guide**: Performance testing and optimization documentation
- **Debug Guide**: Debug symbol usage and troubleshooting guide

### Infrastructure
- **Performance Testing**: Automated performance regression testing
- **Benchmarking**: Comprehensive benchmarking suite
- **CI/CD**: Enhanced CI/CD with performance monitoring
- **Release Process**: Streamlined release process with proper versioning

## [0.2.0] - 2025-10-20

### Added
- Comprehensive import system with alias support
- String concatenation with mixed types using `+` operator
- Unary expressions: negation (`-`) and logical NOT (`!`)
- Enhanced error messages with helpful hints and suggestions
- Comprehensive VM test suite with 65.7% coverage
- Edge case testing for various scenarios
- Local file import support with module loading
- Standard library intrinsics for math, string, and I/O operations
- Type inference improvements for mixed-type expressions
- Enhanced diagnostic messages for better developer experience
- Comprehensive roadmap with 6 development phases
- Visual roadmap with ASCII diagrams and progress indicators

### Changed
- Updated all examples to use new import syntax
- Improved error messages throughout the compiler pipeline
- Enhanced type checker with better error reporting
- Updated documentation to reflect new features
- Regenerated golden test files with improved error messages
- Optimized string operations with object pooling
- Replaced VM switch statement with efficient jump table dispatch

### Fixed
- Fixed e2e tests on macOS by skipping Cranelift-specific tests
- Fixed string concatenation type inference
- Fixed unary expression handling in MIR generation
- Fixed import resolution for both std and local modules
- Fixed VM intrinsic function dispatch
- Removed unused code and fixed linting issues

### Technical Details
- **VM Coverage**: Improved from 23.5% to 65.7%
- **Test Coverage**: Added 8 comprehensive test functions with 50+ test cases
- **Error Messages**: Enhanced with specific hints and conversion suggestions
- **Import System**: Supports both `import std.io as io` and `import math_utils` syntax
- **String Operations**: Automatic type conversion for mixed string/integer concatenation
- **Unary Operations**: Full support for `-` (negation) and `!` (logical NOT)
- **Performance**: ~290ns/op string concat, ~23ns/op instruction dispatch

## [0.1.0] - 2025-10-20

### Added
- Initial language implementation
- Basic compiler pipeline (lexer, parser, AST, type checker, MIR)
- VM backend for interpretation
- Cranelift backend stub (Linux only)
- Basic standard library declarations
- Golden test system for regression testing
- CLI tools (`omnic` compiler, `omnir` runner)
- Cross-platform build system with Makefile
- GitLab CI/CD pipeline
- Comprehensive documentation

### Language Features
- Static typing with type inference
- Variables (`let` immutable, `var` mutable)
- Functions with explicit and inferred return types
- Control flow (if/else, for loops, while loops)
- Data structures (arrays, maps, structs, enums)
- Basic operators (arithmetic, comparison, logical)
- Comments (single-line and multi-line)

### Compiler Features
- Hand-rolled lexer with comprehensive token support
- Recursive descent parser with Pratt parsing for expressions
- Abstract Syntax Tree (AST) representation
- Type checker with scope management
- SSA-based MIR (Middle Intermediate Representation)
- Virtual Machine for execution
- Cranelift integration (stub implementation)
- Runtime library for system calls

### Standard Library
- I/O functions (print, println variants)
- Math functions (max, min, abs, pow, gcd, lcm, factorial)
- String functions (length, concat)
- Array operations (planned)
- OS interface (planned)
- Collections (planned)

### Development Tools
- Golden test generation tools
- Comprehensive test suite
- Code coverage reporting
- Linting and formatting
- Build automation
- Package and release management

### Documentation
- Language specification and grammar
- Comprehensive language tour
- Quick reference guide
- API documentation
- Contributing guidelines
- Examples and tutorials

---

## Version History

- **v0.1.0**: Initial release with basic language features
- **v0.2.0**: Enhanced with import system, string operations, comprehensive testing, and roadmap

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on how to contribute to this project.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
