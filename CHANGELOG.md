# Changelog

All notable changes to OmniLang will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- Advanced type system with generics and union types
- MIR optimization passes and advanced features
- Complete native code generation with Cranelift
- Language features: concurrency, error handling, metaprogramming
- Production readiness: performance optimization, security, deployment
- Ecosystem development: package registry, community building

## [0.4.0] - 2025-01-21

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
