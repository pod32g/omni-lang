# ADR-002: Detailed OmniLang Compiler Roadmap

## Status
Proposed

## Context

The original ADR-001 provided a high-level roadmap, but we now have a much clearer picture of what's implemented and what needs to be done. The compiler has made significant progress with:

- ✅ **Complete Frontend**: Lexer, Parser, AST, Type Checker
- ✅ **Basic MIR**: SSA-based intermediate representation with builder
- ✅ **VM Backend**: Working interpreter with 65.7% test coverage
- ✅ **Import System**: Both std and local file imports with aliases
- ✅ **Basic Stdlib**: I/O, math, and string intrinsics
- ✅ **CLI Tools**: `omnic` compiler and `omnir` runner
- ✅ **Testing**: Comprehensive test suite with edge cases

However, several critical areas need completion to achieve a production-ready compiler.

## Decision

We will implement the remaining features in the following detailed phases, with clear deliverables and success criteria:

---

## Phase 1: Type System Completion (2-3 weeks)
**Goal**: Complete the type checker with full semantic analysis

### 1.1 Advanced Type Inference (1 week)
- [ ] **Generic Type Support**
  - Implement generic function declarations: `func max<T>(a:T, b:T):T`
  - Add generic struct/enum support: `struct Point<T> { x:T, y:T }`
  - Type parameter constraints and bounds
  - Generic instantiation and specialization

- [ ] **Advanced Type Inference**
  - Bidirectional type inference for complex expressions
  - Type inference across function boundaries
  - Constraint solving for generic types
  - Type inference for closures and anonymous functions

- [ ] **Type System Enhancements**
  - Union types: `int | string | bool`
  - Optional types: `T?` (nullable types)
  - Type aliases: `type UserID = int`
  - Type assertions and type guards

### 1.2 Control Flow Analysis (1 week)
- [ ] **Reachability Analysis**
  - Detect unreachable code after returns/throws
  - Warn about missing return statements
  - Dead code elimination hints

- [ ] **Variable Mutability Tracking**
  - Proper `let` vs `var` enforcement
  - Mutable reference tracking
  - Borrow checker basics (immutable references)

- [ ] **Pattern Matching**
  - `match` expressions with destructuring
  - Enum pattern matching
  - Guard clauses and complex patterns

### 1.3 Memory Management (1 week)
- [ ] **Ownership System**
  - Move semantics for value types
  - Borrowing rules (immutable/mutable)
  - Lifetime tracking for references
  - RAII (Resource Acquisition Is Initialization)

- [ ] **Memory Safety**
  - Null pointer prevention
  - Buffer overflow protection
  - Use-after-free detection
  - Double-free prevention

**Success Criteria**:
- All type system features have comprehensive tests
- Type checker coverage > 80%
- No type-related runtime errors in test suite
- Performance: type checking < 100ms for 1000-line files

---

## Phase 2: MIR Optimization & Advanced Features (3-4 weeks)
**Goal**: Complete the MIR infrastructure with advanced optimizations

### 2.1 Advanced MIR Features (1.5 weeks)
- [ ] **Control Flow Graph**
  - Proper basic block management
  - Phi nodes for SSA form
  - Loop detection and analysis
  - Exception handling (try/catch/finally)

- [ ] **Advanced Instructions**
  - Function calls with proper calling conventions
  - Indirect calls and function pointers
  - Tail call optimization
  - Inline assembly support

- [ ] **Memory Operations**
  - Load/store instructions for memory access
  - Array indexing and bounds checking
  - Struct field access
  - Pointer arithmetic (safe)

### 2.2 Optimization Passes (1.5 weeks)
- [ ] **High-Level Optimizations**
  - Constant folding and propagation
  - Dead code elimination
  - Common subexpression elimination
  - Loop invariant code motion

- [ ] **SSA Optimizations**
  - Sparse conditional constant propagation (SCCP)
  - Mem2reg (promote memory to registers)
  - Global value numbering
  - Partial redundancy elimination

- [ ] **Advanced Optimizations**
  - Function inlining
  - Loop unrolling
  - Vectorization hints
  - Profile-guided optimization

### 2.3 MIR Verification & Testing (1 week)
- [ ] **MIR Verifier**
  - SSA form validation
  - Type consistency checking
  - Control flow validation
  - Memory safety verification

- [ ] **MIR Golden Tests**
  - Snapshot testing for MIR output
  - Optimization pass testing
  - Regression testing for MIR changes
  - Performance regression testing

**Success Criteria**:
- MIR coverage > 90%
- All optimization passes have tests
- MIR generation < 50ms for 1000-line files
- Optimization improves performance by >20% on benchmarks

---

## Phase 3: Native Code Generation (4-5 weeks)
**Goal**: Complete the Cranelift backend for native code generation

### 3.1 Cranelift Integration (2 weeks)
- [ ] **MIR to Cranelift IR Translation**
  - Complete instruction mapping
  - Type system integration
  - Calling convention implementation
  - Exception handling support

- [ ] **Cross-Platform Support**
  - macOS support (currently Linux only)
  - Windows support
  - ARM64 and x86_64 architectures
  - Cross-compilation support

- [ ] **Object File Generation**
  - ELF, Mach-O, and PE object files
  - Debug information (DWARF)
  - Symbol table generation
  - Relocation handling

### 3.2 Runtime Integration (1.5 weeks)
- [ ] **Runtime Library**
  - Memory allocation and deallocation
  - Garbage collection (basic)
  - Exception handling runtime
  - System call interface

- [ ] **FFI (Foreign Function Interface)**
  - C function calling
  - C library integration
  - Go function calling
  - Rust function calling

- [ ] **Standard Library Implementation**
  - Complete std.io implementation
  - Complete std.math implementation
  - Complete std.string implementation
  - std.collections implementation

### 3.3 Linking & Distribution (1.5 weeks)
- [ ] **Linker Integration**
  - Static linking
  - Dynamic linking
  - Library dependency resolution
  - Symbol resolution

- [ ] **Package Management**
  - Dependency resolution
  - Version management
  - Package building and distribution
  - Local package development

**Success Criteria**:
- Native code generation works on all platforms
- Performance within 10% of equivalent C code
- All standard library functions implemented
- Cross-compilation working

---

## Phase 4: Language Features & Ergonomics (3-4 weeks)
**Goal**: Complete the language feature set and improve developer experience

### 4.1 Advanced Language Features (2 weeks)
- [ ] **Concurrency**
  - Goroutines (lightweight threads)
  - Channels for communication
  - Select statements
  - Synchronization primitives

- [ ] **Error Handling**
  - Result types: `Result<T, E>`
  - Error propagation with `?` operator
  - Panic and recovery
  - Custom error types

- [ ] **Metaprogramming**
  - Compile-time code generation
  - Macros and code templates
  - Reflection and introspection
  - Attribute system

### 4.2 Developer Experience (1.5 weeks)
- [ ] **Language Server Protocol (LSP)**
  - Syntax highlighting
  - Auto-completion
  - Go to definition
  - Error highlighting and diagnostics

- [ ] **Debugging Support**
  - Source-level debugging
  - Breakpoint support
  - Variable inspection
  - Call stack navigation

- [ ] **Tooling**
  - Code formatter (`omnifmt`)
  - Linter (`omnilint`)
  - Documentation generator (`omnidoc`)
  - Package manager (`omnipkg`)

### 4.3 Documentation & Examples (0.5 weeks)
- [ ] **Comprehensive Documentation**
  - Complete language specification
  - API documentation
  - Tutorial series
  - Best practices guide

- [ ] **Example Programs**
  - Web server example
  - CLI tool example
  - Game development example
  - System programming example

**Success Criteria**:
- All language features documented and tested
- LSP working with major editors
- Debugging support functional
- Example programs compile and run

---

## Phase 5: Performance & Production Readiness (2-3 weeks)
**Goal**: Optimize performance and prepare for production use

### 5.1 Performance Optimization (1.5 weeks)
- [ ] **Compiler Performance**
  - Parallel compilation
  - Incremental compilation
  - Caching system
  - Memory usage optimization

- [ ] **Runtime Performance**
  - JIT compilation
  - Profile-guided optimization
  - Vectorization
  - SIMD support

- [ ] **Benchmarking**
  - Comprehensive benchmark suite
  - Performance regression testing
  - Comparison with other languages
  - Continuous performance monitoring

### 5.2 Production Features (1 week)
- [ ] **Security**
  - Secure coding guidelines
  - Security audit tools
  - Vulnerability scanning
  - Safe defaults

- [ ] **Reliability**
  - Comprehensive error handling
  - Graceful degradation
  - Recovery mechanisms
  - Monitoring and logging

- [ ] **Deployment**
  - Docker support
  - CI/CD integration
  - Release automation
  - Distribution packages

**Success Criteria**:
- Compiler performance < 1s for 10k-line projects
- Runtime performance competitive with Go/Rust
- All security checks passing
- Production deployment working

---

## Phase 6: Ecosystem & Community (Ongoing)
**Goal**: Build a thriving ecosystem around OmniLang

### 6.1 Package Ecosystem
- [ ] **Package Registry**
  - Central package repository
  - Package discovery and search
  - Version management
  - Security scanning

- [ ] **Third-Party Libraries**
  - Database drivers
  - Web frameworks
  - GUI libraries
  - Scientific computing

### 6.2 Community & Governance
- [ ] **Community Building**
  - Discord/Slack community
  - Regular meetups
  - Conference talks
  - Blog posts and tutorials

- [ ] **Governance**
  - RFC process
  - Language evolution
  - Community guidelines
  - Contributor recognition

---

## Implementation Strategy

### Development Approach
1. **Test-Driven Development**: Every feature must have tests before implementation
2. **Incremental Delivery**: Each phase delivers working functionality
3. **Continuous Integration**: All changes must pass CI/CD pipeline
4. **Performance Monitoring**: Track performance metrics continuously
5. **User Feedback**: Regular feedback from early adopters

### Success Metrics
- **Test Coverage**: >90% for all components
- **Performance**: Compile time <1s for 10k-line projects
- **Reliability**: <0.1% crash rate in production
- **Adoption**: 100+ GitHub stars, 10+ contributors
- **Documentation**: Complete API docs and tutorials

### Risk Mitigation
- **Technical Risks**: Regular code reviews and architecture discussions
- **Timeline Risks**: Buffer time built into each phase
- **Resource Risks**: Clear task breakdown for contributors
- **Quality Risks**: Automated testing and continuous integration

---

## Conclusion

This roadmap provides a clear path from the current state to a production-ready OmniLang compiler. Each phase builds upon the previous one, ensuring continuous progress while maintaining code quality and user experience.

The estimated timeline is 14-19 weeks total, with the first three phases being critical for basic functionality and the later phases focusing on polish and ecosystem development.

**Next Steps**:
1. Review and approve this roadmap
2. Set up project management tools (GitHub Projects, milestones)
3. Begin Phase 1.1: Advanced Type Inference
4. Establish regular progress reviews and adjustments

### Developer Experience
- Improve error messages and diagnostics across the compiler pipeline.
- Add tooling support (LSP, editor plugins) to streamline OmniLang development.
- Surface contextual error snippets and suggestion engine (e.g. import typo hints, highlighted spans).
- Expand CLI automation: `omnic --watch`, `omnir --watch`, JSON outputs for scripting.

### Standard Library
- Expand `std` modules to cover networking, filesystems, and concurrency.
- Ensure parity between the VM and C backends for all `std` functions.
- Introduce developer-oriented utilities (`std.testing`, watch helpers) to simplify rapid iteration.

### Tooling & Ecosystem
- Release official OmniLang VS Code extension with syntax highlighting and snippets.
- Integrate OmniLang tooling into popular CI/CD platforms.
- Update VS Code extension to expose new CLI features (watch commands, backend listings).
- Add lightweight JSON API endpoints in CLI tools for editor integrations.

### Releases
- Automate packaging and checksum generation for OmniLang releases.
- Provide pre-built binaries for major operating systems.
- Publish release manifests (`release.json`) with artifact metadata for downstream consumption.
- Ship official Docker images with Omni toolchain ready-to-use.
