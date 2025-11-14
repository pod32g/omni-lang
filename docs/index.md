# OmniLang

Welcome to OmniLang’s documentation site. OmniLang is a hobbyist statically typed programming language used to experiment with compiler architecture, SSA-based intermediate representations, and multiple backends.

## Why OmniLang?

- **Go frontend** that implements the lexer, parser, type checker, and MIR builder.
- **Multiple backends** including C codegen, a VM interpreter, and an experimental Cranelift bridge.
- **Standard library** modules for I/O, math, collections, networking, OS helpers, dev tooling, and testing utilities.
- **CLI tools**: `omnic` (compiler), `omnir` (runner/test harness), and `omnipkg` (packager).

## Quick Start

```bash
git clone https://github.com/omni-lang/omni.git
cd omni
make build      # build compiler + tools
make test       # run language + stdlib tests
```

Create `hello.omni`:

```omni
import std

func main():int {
    std.io.println("Hello, OmniLang!")
    return 0
}
```

Compile and run:

```bash
./bin/omnic hello.omni -o hello   # native (C backend)
./hello

./bin/omnir hello.omni            # interpret via VM
./bin/omnic hello.omni -emit mir  # inspect SSA MIR
```

## Language Highlights

- **Types**: primitives (`int`, `float`, `bool`, `string`, `char`), arrays, maps, user-defined `struct` and `enum`.
- **Control flow**: `if`, `switch`, `for` loops (classic and range), `defer`, early `return`.
- **Functions**: first-class, support multiple return values, default immutability (`let`) with explicit `var` for mutation.
- **Standard library**: `std.io`, `std.string`, `std.math`, `std.net`, `std.os`, `std.testing`, `std.dev`.

## Tooling

- **VS Code extension** (in `vscode/omni`) provides syntax highlighting, snippets, simple completions, and inline diagnostics by invoking `omnic`.
- **Testing**: `omnir` runs `.omni` files directly and drives end-to-end suites under `omni/tests/e2e`.
- **Packaging**: `make package` builds distributable archives; the GitLab pipeline publishes release bundles.

## Learn More

- Read the in-repo [README](../README.md) for deeper examples.
- Browse `docs/` for reference material and the standard library API docs.
- Explore `omni/std/` and `omni/tests/` to see real programs written in OmniLang.

Feedback and contributions are welcome! Open issues or merge requests on GitLab, and follow the GitHub mirror for read-only browsing.***
