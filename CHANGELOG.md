# Changelog

All notable changes to OmniLang will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

### Changed
- Updated all examples to use new import syntax
- Improved error messages throughout the compiler pipeline
- Enhanced type checker with better error reporting
- Updated documentation to reflect new features
- Regenerated golden test files with improved error messages

### Fixed
- Fixed e2e tests on macOS by skipping Cranelift-specific tests
- Fixed string concatenation type inference
- Fixed unary expression handling in MIR generation
- Fixed import resolution for both std and local modules
- Fixed VM intrinsic function dispatch

### Technical Details
- **VM Coverage**: Improved from 23.5% to 65.7%
- **Test Coverage**: Added 8 comprehensive test functions with 50+ test cases
- **Error Messages**: Enhanced with specific hints and conversion suggestions
- **Import System**: Supports both `import std.io as io` and `import math_utils` syntax
- **String Operations**: Automatic type conversion for mixed string/integer concatenation
- **Unary Operations**: Full support for `-` (negation) and `!` (logical NOT)

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
- **v0.2.0** (Unreleased): Enhanced with import system, string operations, and comprehensive testing

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on how to contribute to this project.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
