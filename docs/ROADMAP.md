# OmniLang Compiler Roadmap

## Current Status (v0.4.3) ✅

### ✅ **Completed Features**
- **Frontend**: Complete lexer, parser, AST, type checker
- **MIR**: SSA-based intermediate representation with builder and PHI nodes
- **Multiple Backends**: C (default), VM, and Cranelift (experimental)
- **Data Structures**: Complete array, map, and struct support
- **Import System**: Both std and local file imports with aliases
- **Standard Library**: I/O, math, and string intrinsics with runtime support
- **CLI Tools**: `omnic` compiler, `omnir` runner, and `omnipkg` packager
- **Testing**: Comprehensive test suite with 100% passing tests
- **Performance**: Optimized compilation and execution
- **Documentation**: Complete language tour, quick reference, and API docs
- **CI/CD**: Multi-platform builds with automated testing
- **Packaging**: Distribution packages for Linux, macOS, and Windows

### 📊 **Current Metrics**
- **VM Coverage**: 32.3% (focused on core functionality)
- **Lexer Coverage**: 78.2%
- **Parser Coverage**: 71.3%
- **Passes Coverage**: 34.6%
- **All Tests**: 100% passing (including e2e tests)
- **Performance**: Sub-second compilation for most programs
- **Backend Support**: C backend fully functional, VM backend for development

---

## Phase 1: Advanced Language Features (3-4 weeks) 🔄

### 1.1 Memory Management (1.5 weeks)
- [ ] Dynamic memory allocation (`new`/`delete`)
- [ ] Ownership system
- [ ] Memory safety checks
- [ ] RAII (Resource Acquisition Is Initialization)

### 1.2 Advanced Type System (1.5 weeks)
- [ ] Generic type support (`func max<T>(a:T, b:T):T`)
- [ ] Union types (`int | string | bool`)
- [ ] Optional types (`T?`)
- [ ] Type aliases (`type UserID = int`)

### 1.3 Control Flow & Error Handling (1 week)
- [ ] Exception handling (`try`/`catch`/`finally`)
- [ ] Error handling (`Result<T, E>`, `?` operator)
- [ ] Pattern matching (`match` expressions)

**Success Criteria**: Memory management working, advanced types implemented, error handling functional

---

## Phase 2: Backend Completion & Optimization (3-4 weeks) 📋

### 2.1 Cranelift Backend Completion (2 weeks)
- [ ] Complete MIR to Cranelift IR translation
- [ ] Cross-platform support (macOS, Windows)
- [ ] Object file generation (ELF, Mach-O, PE)
- [ ] Debug information (DWARF)

### 2.2 Advanced Optimization Passes (1 week)
- [ ] Dead code elimination
- [ ] SSA optimizations (SCCP, mem2reg)
- [ ] Function inlining and loop unrolling
- [ ] Profile-guided optimization

### 2.3 Binary Generation & Linking (1 week)
- [ ] Static and dynamic linking
- [ ] Shared library generation (.so/.dll)
- [ ] Cross-compilation support
- [ ] Binary analysis tools

**Success Criteria**: All backends functional, optimization improves performance by >20%

---

## Phase 3: Language Features & Ergonomics (3-4 weeks) 📋

### 3.1 Advanced Language Features (2 weeks)
- [ ] Concurrency (goroutines, channels)
- [ ] Metaprogramming (macros, reflection)
- [ ] Advanced control flow (labeled breaks, continues)

### 3.2 Developer Experience (1.5 weeks)
- [ ] Language Server Protocol (LSP)
- [ ] Debugging support
- [ ] Tooling (`omnifmt`, `omnilint`, `omnidoc`)

### 3.3 Standard Library Expansion (0.5 weeks)
- [ ] File I/O operations
- [ ] Networking support
- [ ] Collections and algorithms

**Success Criteria**: All language features documented and tested, LSP working

---

## Phase 4: Performance & Production Readiness (2-3 weeks) 📋

### 4.1 Performance Optimization (1.5 weeks)
- [ ] Parallel compilation
- [ ] JIT compilation
- [ ] Profile-guided optimization
- [ ] Comprehensive benchmarking

### 4.2 Production Features (1 week)
- [ ] Security (secure coding, vulnerability scanning)
- [ ] Reliability (error handling, monitoring)
- [ ] Deployment (Docker, CI/CD, distribution)

**Success Criteria**: Compile time <1s for 10k-line projects, production deployment working

---

## Phase 5: Ecosystem & Community (Ongoing) 📋

### 5.1 Package Ecosystem
- [ ] Package registry
- [ ] Third-party libraries
- [ ] Community packages

### 5.2 Community & Governance
- [ ] Community building
- [ ] RFC process
- [ ] Language evolution

---


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
| Phase 1 | 3-4 weeks | Advanced language features (memory, types, error handling) |
| Phase 2 | 3-4 weeks | Backend completion & optimization |
| Phase 3 | 3-4 weeks | Language features & ergonomics |
| Phase 4 | 2-3 weeks | Performance & production readiness |
| Phase 5 | Ongoing | Ecosystem & community |

**Total Estimated Time**: 11-15 weeks for core features

## Recent Achievements (v0.4.0 - v0.4.3)

### ✅ **Major Features Completed**
- **Data Structures**: Arrays, maps, and structs with full support
- **PHI Nodes**: Proper SSA form for control flow
- **Method Syntax**: `x.len()` instead of `len(x)`
- **Enhanced Comparisons**: String, boolean, and float comparisons
- **Complete C Backend**: All MIR instructions supported
- **CI/CD Pipeline**: Multi-platform builds and testing
- **Documentation**: Comprehensive docs and examples
- **Packaging**: Distribution packages for all platforms

---

## Success Metrics

- **Test Coverage**: >90% for all components (currently 32-78% depending on component)
- **Performance**: Compile time <1s for 10k-line projects (currently sub-second for most programs)
- **Reliability**: <0.1% crash rate in production (currently 100% test pass rate)
- **Adoption**: 100+ GitHub stars, 10+ contributors (growing community)
- **Documentation**: Complete API docs and tutorials (comprehensive docs available)

---

## Next Steps

1. **Begin Phase 1.1**: Start with dynamic memory allocation (`new`/`delete`)
2. **Complete Cranelift Backend**: Finish macOS and Windows support
3. **Add Static Linking**: Remove runtime library dependency
4. **Regular Reviews**: Weekly progress reviews and adjustments

---

*This roadmap is a living document and will be updated as we progress. For the most detailed information, see [ADR-002](docs/adr/ADR-002-detailed-roadmap.md).*
