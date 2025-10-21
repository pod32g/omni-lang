# C Backend Documentation

The OmniLang C backend provides native code generation by translating MIR (Mid-level Intermediate Representation) to C code, which is then compiled to native executables.

## Overview

The C backend offers several advantages:
- **Native Performance**: Compiled executables run at native speed
- **Debug Support**: Full debug symbol generation with source mapping
- **Cross-Platform**: Supports multiple platforms and architectures
- **Optimization**: Multiple optimization levels available
- **Standalone**: Generated executables don't require runtime dependencies

## Usage

### Basic Compilation

```bash
# Compile to native executable
omnic -backend c -emit exe program.omni

# With optimization
omnic -backend c -emit exe -opt-level O2 program.omni

# With debug information
omnic -backend c -emit exe -debug program.omni
```

### Output Options

- `exe`: Generate native executable (default)
- `asm`: Generate assembly code

### Optimization Levels

- `O0` / `0`: No optimization
- `O1` / `1`: Basic optimization
- `O2` / `2`: Standard optimization (default)
- `O3` / `3`: Aggressive optimization
- `Os` / `s`: Size optimization

## Debug Information

When using the `-debug` flag, the C backend generates comprehensive debug information:

### Debug Symbols
- Function names and signatures
- Variable names and types
- Source location mapping
- Stack trace support

### Source Maps
A `.map` file is generated alongside the executable containing:
- Source file references
- Line number mappings
- Function boundaries
- Variable scopes

### Example Debug Output

```c
// Generated C code with debug information
#line 1 "program.omni"
int32_t omni_main() {
    // Debug: const instruction (ID: v1, Type: int)
    int32_t v1 = 42;
    // Debug: add instruction (ID: v2, Type: int)
    int32_t v2 = v1 + 10;
    return v2;
}
```

## Generated Code Structure

### Runtime Integration
All generated C code includes the OmniLang runtime:

```c
#include "omni_rt.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
```

### Function Generation
- Functions are translated with proper C signatures
- Parameter types are mapped to C equivalents
- Return types are preserved
- Variable names are preserved when debug info is enabled

### Type Mapping

| OmniLang Type | C Type | Notes |
|---------------|--------|-------|
| `int` | `int32_t` | 32-bit signed integer |
| `string` | `const char*` | Null-terminated string |
| `bool` | `int32_t` | 0 for false, 1 for true |
| `float` | `double` | Double precision float |

## Platform Support

### Supported Platforms
- **macOS**: Darwin (x86_64, ARM64)
- **Linux**: GNU/Linux (x86_64, ARM64)
- **Windows**: Windows (x86_64, ARM64)

### Architecture Support
- x86_64 (AMD64)
- ARM64 (AArch64)

### Platform-Specific Features
- Automatic platform detection
- Architecture-specific optimizations
- Platform-specific runtime linking

## Error Handling

### Common Issues

1. **Missing Runtime Files**
   ```
   Error: runtime directory not found
   ```
   **Solution**: Ensure runtime files are in the correct location relative to the compiler binary.

2. **GCC Not Found**
   ```
   Error: c compilation failed
   ```
   **Solution**: Install GCC or Clang compiler.

3. **Debug Symbols Not Working**
   ```
   Error: debug symbols not found
   ```
   **Solution**: Use `-debug` flag and ensure source file path is correct.

### Debugging Tips

1. **Use Debug Mode**: Always use `-debug` for development
2. **Check Source Maps**: Verify `.map` files are generated correctly
3. **Test Optimizations**: Try different optimization levels
4. **Platform Testing**: Test on target platforms

## Performance Considerations

### Optimization Impact

| Level | Compile Time | Binary Size | Runtime Speed |
|-------|--------------|-------------|---------------|
| O0 | Fastest | Largest | Slowest |
| O1 | Fast | Large | Moderate |
| O2 | Moderate | Medium | Good |
| O3 | Slow | Medium | Best |
| Os | Moderate | Smallest | Good |

### Best Practices

1. **Development**: Use `-debug -opt-level O0` for fastest compilation
2. **Testing**: Use `-opt-level O1` for reasonable performance
3. **Release**: Use `-opt-level O2` or `O3` for best performance
4. **Embedded**: Use `-opt-level Os` for smallest binary size

## Integration with Build Systems

### Makefile Integration

```makefile
%.exe: %.omni
	omnic -backend c -emit exe -opt-level O2 $< -o $@

debug: %.exe
	omnic -backend c -emit exe -debug $< -o $@
```

### CMake Integration

```cmake
add_custom_command(
    OUTPUT ${CMAKE_CURRENT_BINARY_DIR}/${TARGET_NAME}
    COMMAND omnic -backend c -emit exe -opt-level O2 ${SOURCE_FILE} -o $@
    DEPENDS ${SOURCE_FILE}
)
```

## Advanced Features

### Custom Runtime Linking
The C backend automatically links with the OmniLang runtime, but you can customize this:

```bash
# Generate C code only
omnic -backend c -emit asm program.omni

# Then compile manually with custom flags
gcc -O2 program.c runtime/omni_rt.c -o program
```

### Source Map Integration
Source maps can be used with debuggers and profiling tools:

```bash
# Generate with source map
omnic -backend c -emit exe -debug program.omni

# Use with gdb
gdb program
(gdb) source program.map
```

## Troubleshooting

### Verbose Output
Add debug output to see what the compiler is doing:

```bash
omnic -backend c -emit exe -debug -verbose program.omni
```

### Intermediate Files
Keep intermediate C files for inspection:

```bash
# The compiler generates .c files temporarily
# You can modify the compiler to keep them for debugging
```

### Runtime Issues
If the generated executable doesn't run:

1. Check that all runtime dependencies are available
2. Verify platform compatibility
3. Test with a simple program first
4. Check system libraries

## Future Enhancements

Planned improvements to the C backend:

- **DWARF Debug Info**: Full DWARF debug information generation
- **Link-Time Optimization**: LTO support for better optimization
- **Custom Memory Management**: Optional garbage collection
- **SIMD Support**: Vector instruction generation
- **Profile-Guided Optimization**: PGO support

## Examples

### Simple Program
```omni
func main() {
    let x = 42
    let y = x + 8
    return y
}
```

Generated C:
```c
int32_t omni_main() {
    int32_t x = 42;
    int32_t y = x + 8;
    return y;
}
```

### With Debug Information
```omni
func fibonacci(n: int) -> int {
    if n <= 1 {
        return n
    }
    return fibonacci(n - 1) + fibonacci(n - 2)
}

func main() {
    let result = fibonacci(10)
    return result
}
```

Generated C (with debug):
```c
#line 1 "fibonacci.omni"
// Debug: Function fibonacci (Return: int, Params: 1)
int32_t fibonacci(int32_t n) {
    // Debug: const instruction (ID: v1, Type: int)
    int32_t v1 = 1;
    // Debug: cmp instruction (ID: v2, Type: bool)
    int32_t v2 = n <= v1;
    if (v2) {
        return n;
    }
    // ... more generated code
}
```
