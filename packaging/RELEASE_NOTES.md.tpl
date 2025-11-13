# OmniLang {{VERSION}} Release

## Installation

### Linux (x86_64)
```bash
tar -xzf omni-{{VERSION}}-linux-x86_64.tar.gz
cd omni-{{VERSION}}
./install.sh
```

### macOS (x86_64)
```bash
tar -xzf omni-{{VERSION}}-darwin-x86_64.tar.gz
cd omni-{{VERSION}}
./install.sh
```

### macOS (ARM64)
```bash
tar -xzf omni-{{VERSION}}-darwin-arm64.tar.gz
cd omni-{{VERSION}}
./install.sh
```

### Windows (x86_64)
1. Extract omni-{{VERSION}}-windows-x86_64.zip
2. Add the extracted directory to your PATH
3. Run omnic.exe and omnir.exe from command prompt

## Verification

After installation, verify the installation:
```bash
omnic --version
omnir --version
```

## Quick Start

```bash
# Create a simple program
echo 'func main():int { println("Hello, OmniLang!"); return 0 }' > hello.omni

# Run it
omnir hello.omni

# Compile it
omnic hello.omni -o hello
```

## What's New

See CHANGELOG.md for detailed changes.

## Support

- Documentation: README.md
- Issues: https://github.com/omni-lang/omni/issues
- Discussions: https://github.com/omni-lang/omni/discussions

