# ADR-000: Go Frontend/Midend, Cranelift Backend

Context:
- We need to iterate rapidly on language semantics and tooling.
- Go provides fast builds, simple concurrency, great testing, and straightforward cross-compilation.
- Cranelift offers a simpler embedding story than LLVM while delivering strong codegen.

Decision:
- Implement frontend and midend (lexer, parser, AST, type checker, MIR and passes) in Go.
- Provide a Rust/Cranelift backend via a small C ABI bridge (cgo) for native codegen.
- Defer LLVM integration to a future phase, once the language stabilizes.

Consequences:
- Faster time-to-first-compiler.
- A clean, backend-agnostic MIR allows future LLVM or WASM backends without redesigning the front/mid.
- Slight FFI complexity between Go and Rust is acceptable for v1.
