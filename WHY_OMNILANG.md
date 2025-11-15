# Why Choose OmniLang?

## The Problem with Current Languages

**Rust** is powerful but complex - the learning curve is steep and compilation is slow.  
**Go** is simple but limited - garbage collection and lack of zero-cost abstractions.  
**C++** is fast but archaic - complex syntax and slow compilation.  
**C** is simple but unsafe - manual memory management and no modern features.

## OmniLang's Solution

**OmniLang combines the best of all worlds:**
- **Lightning-fast compilation** (sub-second vs Rust's 8s)
- **Simple syntax** (easy to learn and read)
- **High performance** (near-C performance with multiple backends)
- **Memory safety** (planned, without complexity)
- **Modern tooling** (debugging, packages, hot reload)
- **Multiple backends** (C, VM, Cranelift) for different use cases

## Key Differentiators

### 1. **Lightning-Fast Development Cycle**
```omni
// Write code
func main() : int {
    print("Hello, World!")
    return 0
}

// Compile in 50ms
omnic main.omni

// Run immediately
./main
```

### 2. **Simplicity Without Sacrifice**
```omni
// Clean, readable syntax
func fibonacci(n: int) : int {
    if n <= 1 {
        return n
    }
    return fibonacci(n - 1) + fibonacci(n - 2)
}
```

### 3. **Advanced Debugging**
- Source maps linking generated code to original
- Variable tracking across compilation stages
- Detailed error messages with suggestions
- Interactive debugging support

### 4. **Modern Package System**
```omni
import std.io as io
import std.math as math
import my_package.utils as utils
```

## Perfect For

- **Systems Programming**: OS kernels, drivers, embedded systems
- **High-Performance Apps**: Game engines, real-time systems
- **Web Backends**: APIs, microservices, data processing
- **CLI Tools**: Developer utilities, automation
- **Learning**: Systems programming concepts

## Performance Comparison

| Language | Compile Time | Runtime | Memory | Learning Curve |
|----------|-------------|---------|---------|----------------|
| **OmniLang** | 50ms | Fast | Low | Easy |
| Rust | 8s | Fast | Low | Hard |
| Go | 200ms | Medium | High | Easy |
| C++ | 15s | Fast | Low | Hard |

## Get Started in 30 Seconds

```bash
# Install
curl -sSL https://install.omni-lang.dev | sh

# Create project
omni new my-app
cd my-app

# Write code
echo 'func main() : int { print("Hello!"); return 0 }' > main.omni

# Run
omnic main.omni && ./main
```

## The Bottom Line

**OmniLang is the language for developers who want:**
- The **power** of systems programming
- The **simplicity** of modern languages  
- The **speed** of fast compilation
- The **safety** of memory management
- The **productivity** of great tooling

**Stop choosing between performance and simplicity. Choose OmniLang.**
