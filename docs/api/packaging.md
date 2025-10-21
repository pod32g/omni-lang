# Packaging and Distribution Documentation

The OmniLang packaging system provides tools for creating distribution packages, managing dependencies, and deploying applications.

## Overview

The packaging system supports:
- **Distribution Packages**: Create tar.gz and zip packages
- **Cross-Platform**: Support for multiple platforms and architectures
- **Debug Symbols**: Optional inclusion of debug information
- **Source Code**: Optional inclusion of source code
- **Standard Library**: Automatic inclusion of standard library
- **Runtime**: Automatic inclusion of runtime libraries

## Package Types

### Supported Formats
- **tar.gz**: Unix/Linux standard archive format
- **zip**: Cross-platform archive format

### Package Structure
```
omni-lang-{version}-{platform}-{arch}.{format}
├── bin/
│   ├── omnic          # Compiler binary
│   └── omnir          # Runtime binary
├── runtime/
│   ├── omni_rt.c      # Runtime source
│   └── omni_rt.h      # Runtime headers
├── std/
│   ├── io/            # I/O standard library
│   ├── math/          # Math standard library
│   └── string/        # String standard library
├── examples/
│   ├── hello_world.omni
│   └── ...
├── docs/
│   ├── README.md
│   └── ...
└── [debug/]           # Optional debug symbols
    └── ...
```

## Usage

### Creating Packages

```bash
# Create a basic package
omnipkg create --version 1.0.0 --platform darwin --arch x86_64

# Create with debug symbols
omnipkg create --version 1.0.0 --platform linux --arch x86_64 --include-debug

# Create with source code
omnipkg create --version 1.0.0 --platform windows --arch x86_64 --include-src

# Create zip package
omnipkg create --version 1.0.0 --platform darwin --arch arm64 --format zip
```

### Package Configuration

```go
// Package configuration options
type PackageConfig struct {
    OutputPath   string      // Output file path
    PackageType  PackageType // tar.gz or zip
    IncludeDebug bool        // Include debug symbols
    IncludeSrc   bool        // Include source code
    Version      string      // Package version
    Platform     string      // Target platform
    Architecture string      // Target architecture
}
```

## Command Line Interface

### omnipkg Command

```bash
omnipkg [command] [options]
```

#### Commands

- `create`: Create a new package
- `info`: Show package information
- `extract`: Extract package contents
- `list`: List package contents

#### Options

- `--version`: Package version (required)
- `--platform`: Target platform (darwin, linux, windows)
- `--arch`: Target architecture (x86_64, arm64)
- `--format`: Package format (tar.gz, zip)
- `--include-debug`: Include debug symbols
- `--include-src`: Include source code
- `--output`: Output file path

### Examples

```bash
# Create a release package
omnipkg create \
  --version 1.0.0 \
  --platform darwin \
  --arch x86_64 \
  --format tar.gz \
  --output omni-lang-1.0.0-darwin-x86_64.tar.gz

# Create a development package with debug info
omnipkg create \
  --version 1.0.0-dev \
  --platform linux \
  --arch x86_64 \
  --include-debug \
  --include-src

# Show package information
omnipkg info omni-lang-1.0.0-darwin-x86_64.tar.gz

# Extract package
omnipkg extract omni-lang-1.0.0-darwin-x86_64.tar.gz
```

## Package Contents

### Required Files

#### Binaries
- `bin/omnic`: Compiler executable
- `bin/omnir`: Runtime executable (if available)

#### Runtime
- `runtime/omni_rt.c`: Runtime implementation
- `runtime/omni_rt.h`: Runtime headers

#### Standard Library
- `std/`: Complete standard library
  - `io/`: Input/output functions
  - `math/`: Mathematical functions
  - `string/`: String manipulation
  - `array/`: Array operations
  - `collections/`: Collection types

#### Documentation
- `README.md`: Main documentation
- `LICENSE`: License information
- `docs/`: Additional documentation
  - `quick-reference.md`: Quick reference guide
  - `api/`: API documentation

#### Examples
- `examples/`: Example programs
  - `hello_world.omni`: Basic hello world
  - `math_operations.omni`: Math examples
  - `string_operations.omni`: String examples

### Optional Files

#### Debug Symbols
- `debug/`: Debug symbol files
- `*.map`: Source map files
- `*.dSYM`: Debug symbol bundles (macOS)

#### Source Code
- `src/`: Source code
  - `internal/`: Internal compiler code
  - `cmd/`: Command line tools
  - `go.mod`: Go module file
  - `Makefile`: Build configuration

## Platform Support

### Supported Platforms

| Platform | Identifier | Notes |
|----------|------------|-------|
| macOS | `darwin` | Intel and Apple Silicon |
| Linux | `linux` | Most distributions |
| Windows | `windows` | Windows 10/11 |

### Supported Architectures

| Architecture | Identifier | Notes |
|--------------|------------|-------|
| x86_64 | `x86_64` | Intel/AMD 64-bit |
| ARM64 | `arm64` | ARM 64-bit |
| x86 | `x86` | Intel/AMD 32-bit |

### Platform-Specific Features

#### macOS
- Universal binaries support
- Code signing integration
- dSYM debug symbol bundles

#### Linux
- ELF binary format
- Shared library support
- Package manager integration

#### Windows
- PE binary format
- DLL support
- Windows Defender compatibility

## Package Management

### Versioning

Packages use semantic versioning:
- **Major.Minor.Patch** (e.g., 1.0.0)
- **Pre-release** (e.g., 1.0.0-alpha, 1.0.0-beta)
- **Build metadata** (e.g., 1.0.0+20190901)

### Dependencies

Packages are self-contained and include:
- Runtime libraries
- Standard library
- Compiler binaries
- Documentation

### Installation

#### Manual Installation
```bash
# Extract package
tar -xzf omni-lang-1.0.0-darwin-x86_64.tar.gz

# Add to PATH
export PATH=$PATH:/path/to/omni-lang-1.0.0/bin

# Verify installation
omnic --version
```

#### Package Manager Integration

##### Homebrew (macOS)
```bash
# Install via Homebrew
brew install omni-lang

# Or install from local package
brew install --build-from-source omni-lang
```

##### apt (Linux)
```bash
# Install via apt
sudo apt install omni-lang

# Or install from local package
sudo dpkg -i omni-lang_1.0.0_amd64.deb
```

## Development Workflow

### Creating Packages

1. **Build Binaries**: Ensure all binaries are built
2. **Prepare Runtime**: Verify runtime files are available
3. **Create Package**: Use omnipkg to create package
4. **Test Package**: Verify package contents and functionality
5. **Distribute**: Upload to distribution channels

### Testing Packages

```bash
# Test package creation
omnipkg create --version test --platform darwin --arch x86_64

# Test package extraction
omnipkg extract omni-lang-test-darwin-x86_64.tar.gz

# Test package functionality
cd omni-lang-test/
./bin/omnic --version
./bin/omnic examples/hello_world.omni
```

### Release Process

1. **Version Bump**: Update version numbers
2. **Build All Platforms**: Create packages for all supported platforms
3. **Test Packages**: Verify functionality on target platforms
4. **Create Release**: Tag and upload packages
5. **Update Documentation**: Update download links and documentation

## Integration Examples

### CI/CD Integration

#### GitHub Actions
```yaml
name: Create Packages
on:
  release:
    types: [published]

jobs:
  create-packages:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [macos-latest, ubuntu-latest, windows-latest]
    
    steps:
      - uses: actions/checkout@v3
      - name: Build
        run: make build
      - name: Create Package
        run: |
          omnipkg create \
            --version ${{ github.ref_name }} \
            --platform ${{ matrix.platform }} \
            --arch ${{ matrix.arch }}
```

#### GitLab CI
```yaml
create-packages:
  stage: package
  script:
    - make build
    - omnipkg create --version $CI_COMMIT_TAG --platform $PLATFORM --arch $ARCH
  artifacts:
    paths:
      - "*.tar.gz"
      - "*.zip"
```

### Docker Integration

```dockerfile
FROM ubuntu:20.04

# Install dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    make \
    && rm -rf /var/lib/apt/lists/*

# Copy and extract OmniLang package
COPY omni-lang-1.0.0-linux-x86_64.tar.gz /tmp/
RUN tar -xzf /tmp/omni-lang-1.0.0-linux-x86_64.tar.gz -C /opt/

# Add to PATH
ENV PATH="/opt/omni-lang-1.0.0/bin:${PATH}"

# Verify installation
RUN omnic --version
```

## Troubleshooting

### Common Issues

1. **Missing Runtime Files**
   ```
   Error: runtime directory not found
   ```
   **Solution**: Ensure runtime files are in the correct location.

2. **Platform Mismatch**
   ```
   Error: unsupported platform
   ```
   **Solution**: Use supported platform identifiers.

3. **Architecture Mismatch**
   ```
   Error: unsupported architecture
   ```
   **Solution**: Use supported architecture identifiers.

### Debug Information

```bash
# Verbose package creation
omnipkg create --version 1.0.0 --platform darwin --arch x86_64 --verbose

# List package contents
omnipkg list omni-lang-1.0.0-darwin-x86_64.tar.gz

# Show package information
omnipkg info omni-lang-1.0.0-darwin-x86_64.tar.gz
```

## Future Enhancements

Planned improvements to the packaging system:

- **Package Signing**: Digital signature support
- **Dependency Management**: Automatic dependency resolution
- **Update System**: Automatic update checking and installation
- **Package Registry**: Central package repository
- **Plugin System**: Extensible packaging system
- **Compression**: Better compression algorithms
- **Incremental Updates**: Delta package updates
