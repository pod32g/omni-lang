# Contributing to OmniLang

Thank you for your interest in contributing to OmniLang! This guide will help you get started with contributing to the project.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Code Style](#code-style)
- [Pull Request Process](#pull-request-process)
- [Issue Reporting](#issue-reporting)

## Getting Started

### Prerequisites

- **Go 1.22+**: For the frontend and midend
- **Rust 1.70+**: For the Cranelift backend
- **Cargo**: Rust package manager
- **Git**: Version control
- **Make**: Build automation

### Installation

```bash
# Clone the repository
git clone https://github.com/omni-lang/omni.git
cd omni

# Install dependencies
go mod download
cd native/clift && cargo build

# Build the project
make build

# Run tests
make test
```

## Development Setup

### 1. Fork and Clone

```bash
# Fork the repository on GitHub, then clone your fork
git clone https://github.com/YOUR_USERNAME/omni.git
cd omni

# Add upstream remote
git remote add upstream https://github.com/omni-lang/omni.git
```

### 2. Create a Branch

```bash
# Create a new branch for your feature
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/your-bug-description
```

### 3. Development Environment

```bash
# Install development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Set up pre-commit hooks (optional)
cp .githooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

## Project Structure

```
omni/
├── cmd/                    # Command-line tools
│   ├── omnic/             # Compiler CLI
│   └── omnir/             # Runner CLI
├── internal/              # Internal packages
│   ├── lexer/             # Tokenization
│   ├── parser/            # Syntax analysis
│   ├── ast/               # Abstract syntax tree
│   ├── types/             # Type system
│   ├── mir/               # SSA intermediate representation
│   ├── passes/            # Optimization passes
│   ├── vm/                # Virtual machine
│   ├── backend/           # Code generation backends
│   └── runtime/           # Runtime library
├── native/                # Native code
│   └── clift/             # Rust Cranelift bridge
├── std/                   # Standard library
├── examples/              # Example programs
├── tests/                 # Test suite
├── docs/                  # Documentation
└── tools/                 # Development tools
```

## Making Changes

### 1. Choose What to Work On

- **Bug fixes**: Look for issues labeled `bug` or `good first issue`
- **Features**: Check the roadmap in `agent.md`
- **Documentation**: Improve existing docs or add new ones
- **Tests**: Add test coverage for existing features

### 2. Understand the Codebase

- Read the [Language Tour](docs/spec/language-tour.md)
- Study the [Grammar Specification](docs/spec/grammar.md)
- Look at existing tests to understand expected behavior
- Check the [Architecture Decision Records](docs/adr/)

### 3. Make Your Changes

- Follow the [Code Style](#code-style) guidelines
- Write tests for new functionality
- Update documentation as needed
- Ensure all tests pass

### 4. Test Your Changes

```bash
# Run all tests
make test

# Run specific test suites
go test ./internal/lexer
go test ./internal/parser
go test ./tests/e2e

# Run with verbose output
go test -v ./internal/lexer

# Run benchmarks
make bench
```

## Testing

### Test Structure

- **Unit tests**: Test individual functions and methods
- **Integration tests**: Test component interactions
- **End-to-end tests**: Test complete programs
- **Golden tests**: Snapshot testing for AST, MIR, etc.

### Writing Tests

```go
// Example unit test
func TestLexer(t *testing.T) {
    tests := []struct {
        input    string
        expected []Token
    }{
        {"42", []Token{{Kind: INT, Lexeme: "42"}}},
        {"hello", []Token{{Kind: IDENT, Lexeme: "hello"}}},
    }
    
    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            lexer := NewLexer(tt.input)
            tokens := lexer.Tokenize()
            assert.Equal(t, tt.expected, tokens)
        })
    }
}
```

### Golden Tests

Golden tests use snapshots to ensure output consistency:

```bash
# Generate golden tests
go run ./tools/gen_ast_goldens
go run ./tools/gen_mir_goldens
go run ./tools/gen_type_goldens

# Update golden tests after changes
go test ./internal/lexer -update
```

### End-to-End Tests

```go
// Example e2e test
func TestHelloWorld(t *testing.T) {
    result, err := runVM("tests/e2e/hello_world.omni")
    require.NoError(t, err)
    assert.Equal(t, "42", result)
}
```

## Code Style

### Go Code

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` and `goimports` for formatting
- Follow the project's naming conventions
- Add comments for exported functions and types

```go
// Package lexer provides tokenization for OmniLang.
package lexer

// Token represents a lexical token.
type Token struct {
    Kind   Kind     // Token kind
    Lexeme string   // Token text
    Span   Span     // Source location
}

// NewLexer creates a new lexer for the given input.
func NewLexer(input string) *Lexer {
    return &Lexer{input: input}
}
```

### Rust Code

- Follow [Rust API Guidelines](https://rust-lang.github.io/api-guidelines/)
- Use `rustfmt` for formatting
- Use `clippy` for linting
- Add documentation comments

```rust
/// Compiles MIR JSON to native object code.
/// 
/// # Arguments
/// * `mir_json` - JSON representation of MIR module
/// * `output_path` - Path to write object file
/// 
/// # Returns
/// * `Ok(())` on success
/// * `Err(String)` on failure
pub fn compile_to_object(mir_json: &str, output_path: &str) -> Result<(), String> {
    // Implementation
}
```

### OmniLang Code

- Use descriptive variable and function names
- Follow the language's naming conventions
- Add comments for complex logic
- Use consistent formatting

```omni
// Calculate the factorial of a number
func factorial(n:int):int {
    if n <= 1 {
        return 1
    }
    return n * factorial(n - 1)
}
```

## Pull Request Process

### 1. Before Submitting

- [ ] All tests pass
- [ ] Code follows style guidelines
- [ ] Documentation is updated
- [ ] Commit messages are clear
- [ ] Branch is up to date with main

### 2. Create Pull Request

- Use a descriptive title
- Reference related issues
- Provide a clear description
- Include test results
- Add screenshots for UI changes

### 3. Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] End-to-end tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No breaking changes (or documented)
```

### 4. Review Process

- Maintainers will review your PR
- Address feedback promptly
- Keep PRs focused and small
- Respond to comments constructively

## Issue Reporting

### Bug Reports

When reporting bugs, include:

- **Description**: Clear description of the issue
- **Steps to reproduce**: Minimal steps to reproduce
- **Expected behavior**: What should happen
- **Actual behavior**: What actually happens
- **Environment**: OS, Go version, etc.
- **Code sample**: Minimal code that reproduces the issue

### Feature Requests

When requesting features, include:

- **Description**: Clear description of the feature
- **Use case**: Why this feature is needed
- **Proposed solution**: How you think it should work
- **Alternatives**: Other approaches considered
- **Additional context**: Any other relevant information

### Issue Labels

- `bug`: Something isn't working
- `enhancement`: New feature or request
- `documentation`: Improvements to documentation
- `good first issue`: Good for newcomers
- `help wanted`: Extra attention is needed
- `question`: Further information is requested

## Development Workflow

### 1. Daily Workflow

```bash
# Start your day
git checkout main
git pull upstream main

# Work on your feature
git checkout -b feature/your-feature
# ... make changes ...
git add .
git commit -m "Add feature X"

# Push and create PR
git push origin feature/your-feature
```

### 2. Keeping Up to Date

```bash
# Fetch latest changes
git fetch upstream

# Rebase your branch
git checkout feature/your-feature
git rebase upstream/main

# Resolve conflicts if any
git add .
git rebase --continue
```

### 3. Code Review

- Review other PRs
- Provide constructive feedback
- Ask questions if something is unclear
- Be respectful and professional

## Getting Help

- **Documentation**: Check the `docs/` directory
- **Issues**: Search existing issues or create new ones
- **Discussions**: Use GitHub Discussions for questions
- **Discord**: Join our Discord server (if available)
- **Email**: Contact maintainers directly

## Recognition

Contributors are recognized in:

- **CONTRIBUTORS.md**: List of all contributors
- **Release notes**: Mentioned in relevant releases
- **Documentation**: Credited in relevant sections
- **GitHub**: Shown in the contributors graph

## Code of Conduct

We are committed to providing a welcoming and inclusive environment. Please:

- Be respectful and inclusive
- Use welcoming and inclusive language
- Be respectful of differing viewpoints
- Accept constructive criticism gracefully
- Focus on what's best for the community
- Show empathy towards other community members

## License

By contributing to OmniLang, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to OmniLang! 🚀
