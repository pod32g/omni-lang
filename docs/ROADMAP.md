# OmniLang Compiler Roadmap

## Current Status (v0.2.0) ✅

### ✅ **Completed Features**
- **Frontend**: Complete lexer, parser, AST, type checker
- **MIR**: SSA-based intermediate representation with builder
- **VM Backend**: Working interpreter with 65.7% test coverage
- **Import System**: Both std and local file imports with aliases
- **Basic Stdlib**: I/O, math, and string intrinsics
- **CLI Tools**: `omnic` compiler and `omnir` runner
- **Testing**: Comprehensive test suite with edge cases
- **Performance**: Optimized string operations, jump table dispatch
- **Documentation**: Complete language tour and quick reference

### 📊 **Current Metrics**
- **VM Coverage**: 65.7% (up from 23.5%)
- **Lexer Coverage**: 83.2%
- **Parser Coverage**: 72.1%
- **Type Checker Coverage**: 38.6%
- **All Tests**: 100% passing
- **Performance**: ~290ns/op string concat, ~23ns/op instruction dispatch

---

## Phase 1: Type System Completion (2-3 weeks) 🔄

### 1.1 Advanced Type Inference (1 week)
- [ ] Generic type support (`func max<T>(a:T, b:T):T`)
- [ ] Union types (`int | string | bool`)
- [ ] Optional types (`T?`)
- [ ] Type aliases (`type UserID = int`)

### 1.2 Control Flow Analysis (1 week)
- [ ] Reachability analysis
- [ ] Variable mutability tracking
- [ ] Pattern matching (`match` expressions)

### 1.3 Memory Management (1 week)
- [ ] Ownership system
- [ ] Memory safety checks
- [ ] RAII (Resource Acquisition Is Initialization)

**Success Criteria**: Type checker coverage >80%, no type-related runtime errors

---

## Phase 2: MIR Optimization & Advanced Features (3-4 weeks) 📋

### 2.1 Advanced MIR Features (1.5 weeks)
- [ ] Control flow graph with phi nodes
- [ ] Exception handling (try/catch/finally)
- [ ] Function calls with proper calling conventions
- [ ] Memory operations (load/store, array indexing)

### 2.2 Optimization Passes (1.5 weeks)
- [ ] Constant folding and propagation
- [ ] Dead code elimination
- [ ] SSA optimizations (SCCP, mem2reg)
- [ ] Function inlining and loop unrolling

### 2.3 MIR Verification & Testing (1 week)
- [ ] MIR verifier for SSA form validation
- [ ] MIR golden tests for regression testing
- [ ] Performance regression testing

**Success Criteria**: MIR coverage >90%, optimization improves performance by >20%

---

## Phase 3: Native Code Generation (4-5 weeks) 📋

### 3.1 Cranelift Integration (2 weeks)
- [ ] Complete MIR to Cranelift IR translation
- [ ] Cross-platform support (macOS, Windows)
- [ ] Object file generation (ELF, Mach-O, PE)
- [ ] Debug information (DWARF)

### 3.2 Runtime Integration (1.5 weeks)
- [ ] Runtime library (memory allocation, GC)
- [ ] FFI (Foreign Function Interface)
- [ ] Complete standard library implementation

### 3.3 Linking & Distribution (1.5 weeks)
- [ ] Static and dynamic linking
- [ ] Package management
- [ ] Cross-compilation support

**Success Criteria**: Native code generation on all platforms, performance within 10% of C

---

## Phase 4: Language Features & Ergonomics (3-4 weeks) 📋

### 4.1 Advanced Language Features (2 weeks)
- [ ] Concurrency (goroutines, channels)
- [ ] Error handling (`Result<T, E>`, `?` operator)
- [ ] Metaprogramming (macros, reflection)

### 4.2 Developer Experience (1.5 weeks)
- [ ] Language Server Protocol (LSP)
- [ ] Debugging support
- [ ] Tooling (`omnifmt`, `omnilint`, `omnidoc`)

### 4.3 Documentation & Examples (0.5 weeks)
- [ ] Complete language specification
- [ ] Example programs (web server, CLI tool, game)

**Success Criteria**: All language features documented and tested, LSP working

---

## Phase 5: Performance & Production Readiness (2-3 weeks) 📋

### 5.1 Performance Optimization (1.5 weeks)
- [ ] Parallel compilation
- [ ] JIT compilation
- [ ] Profile-guided optimization
- [ ] Comprehensive benchmarking

### 5.2 Production Features (1 week)
- [ ] Security (secure coding, vulnerability scanning)
- [ ] Reliability (error handling, monitoring)
- [ ] Deployment (Docker, CI/CD, distribution)

**Success Criteria**: Compile time <1s for 10k-line projects, production deployment working

---

## Phase 6: Ecosystem & Community (Ongoing) 📋

### 6.1 Package Ecosystem
- [ ] Package registry
- [ ] Third-party libraries
- [ ] Community packages

### 6.2 Community & Governance
- [ ] Community building
- [ ] RFC process
- [ ] Language evolution

---

## Quick Start Guide

### For Contributors
1. **Pick a Phase**: Choose a phase that interests you
2. **Read the ADR**: Check `docs/adr/ADR-002-detailed-roadmap.md` for details
3. **Find Issues**: Look for issues tagged with the phase name
4. **Write Tests**: Follow test-driven development
5. **Submit PR**: Ensure all tests pass and CI is green

### For Users
1. **Try OmniLang**: Use the current VM backend
2. **Report Issues**: Help us improve by reporting bugs
3. **Request Features**: Suggest new language features
4. **Join Community**: Participate in discussions

---

## Timeline Summary

| Phase | Duration | Key Deliverables |
|-------|----------|------------------|
| Phase 1 | 2-3 weeks | Complete type system |
| Phase 2 | 3-4 weeks | Advanced MIR & optimizations |
| Phase 3 | 4-5 weeks | Native code generation |
| Phase 4 | 3-4 weeks | Language features & tooling |
| Phase 5 | 2-3 weeks | Performance & production |
| Phase 6 | Ongoing | Ecosystem & community |

**Total Estimated Time**: 14-19 weeks for core features

---

## Success Metrics

- **Test Coverage**: >90% for all components
- **Performance**: Compile time <1s for 10k-line projects
- **Reliability**: <0.1% crash rate in production
- **Adoption**: 100+ GitHub stars, 10+ contributors
- **Documentation**: Complete API docs and tutorials

---

## Next Steps

1. **Review Roadmap**: Review and approve ADR-002
2. **Set Up Project Management**: Create GitHub Projects and milestones
3. **Begin Phase 1.1**: Start with advanced type inference
4. **Regular Reviews**: Weekly progress reviews and adjustments

---

*This roadmap is a living document and will be updated as we progress. For the most detailed information, see [ADR-002](docs/adr/ADR-002-detailed-roadmap.md).*
