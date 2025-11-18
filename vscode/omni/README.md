# OmniLang VS Code Extension

<div align="center">
  <img src="icon.png" alt="OmniLang Logo" width="128"/>
  
  **VS Code support for the Omni programming language**
</div>

## Features

### Language Support
- **Syntax Highlighting** - Full syntax highlighting for OmniLang code
- **Bracket Matching** - Automatic bracket and parenthesis matching
- **Code Snippets** - Quick snippets for common language constructs

### IntelliSense (Language Server Protocol)
- **Autocomplete** - Smart code completion with:
  - Standard library function completions (all `std.*` modules)
  - Context-aware completions (after `std.io.`, `std.math.`, etc.)
  - Local symbol completions (functions, variables, structs, enums)
  - Type-aware completions
- **Signature Help** - Parameter hints when typing function calls
- **Hover Information** - Detailed information about functions, types, and symbols
- **Go to Definition** - Jump to symbol definitions
- **Find References** - Find all usages of symbols
- **Document Symbols** - Navigate symbols in your file (Cmd/Ctrl+Shift+O)
- **Workspace Symbols** - Search symbols across your workspace

### Diagnostics
- **Real-time Error Checking** - Automatic compilation diagnostics
- **Error Highlighting** - Visual indicators for errors and warnings
- **JSON Diagnostics** - Structured error information with hints

### Code Formatting
- **Document Formatting** - Format entire documents (Shift+Alt+F / Shift+Option+F)

## Requirements

- **VS Code** version 1.85.0 or higher
- **OmniLang Compiler** (`omnic`) - Should be available in your PATH or configured via settings

## Installation

### From VSIX
1. Download the latest `.vsix` file from the [releases](https://github.com/omni-lang/omni/releases)
2. Open VS Code
3. Go to Extensions view (Cmd/Ctrl+Shift+X)
4. Click the "..." menu → "Install from VSIX..."
5. Select the downloaded `.vsix` file

### From Source
```bash
# Clone the repository
git clone https://github.com/omni-lang/omni.git
cd omni/vscode/omni

# Install dependencies
npm install

# Compile the extension
npm run compile

# Package the extension
npx vsce package

# Install the generated .vsix file in VS Code
```

## Configuration

### OmniLang Compiler Path
Configure the path to the `omnic` compiler:

1. Open VS Code Settings (Cmd/Ctrl+,)
2. Search for "OmniLang"
3. Set `omniLang.omnicPath` to the path of your `omnic` executable
   - Default: `omnic` (assumes it's in your PATH)
   - Example: `/usr/local/bin/omnic` or `C:\omni\bin\omnic.exe`

## Usage

### Opening OmniLang Files
Simply open any `.omni` file in VS Code. The extension will automatically activate and provide:
- Syntax highlighting
- IntelliSense features
- Real-time diagnostics

### Standard Library Autocomplete
The extension automatically parses the OmniLang standard library and provides completions:

```omni
import std

func main():int {
    // Type "std.io." to see all I/O functions
    std.io.println("Hello, World!")
    
    // Type "std.math." to see all math functions
    let result = std.math.max(10, 20)
    
    return 0
}
```

### Compiling Files
Use the command palette (Cmd/Ctrl+Shift+P) and search for:
- **"OmniLang: Compile Current File"** - Compile the currently open file

Or use the terminal:
```bash
omnic yourfile.omni
```

## Standard Library Support

The extension provides full IntelliSense support for all standard library modules:

- `std.io` - Input/Output operations
- `std.math` - Mathematical functions
- `std.string` - String manipulation
- `std.array` - Array operations
- `std.collections` - Data structures (maps, sets, queues, etc.)
- `std.file` - File operations
- `std.os` - Operating system functions
- `std.log` - Logging utilities
- `std.time` - Time and date functions
- `std.network` - Network operations (DNS, HTTP, etc.)
- `std.dev` - Developer utilities
- `std.test` - Testing framework

## Language Server Protocol

This extension uses the Language Server Protocol (LSP) for advanced language features:

- **Server Process** - Runs in a separate Node.js process
- **Incremental Updates** - Efficient document synchronization
- **Workspace Awareness** - Understands your entire workspace
- **Standard Library Parsing** - Automatically parses and indexes the standard library

### Troubleshooting LSP

If IntelliSense isn't working:

1. Check the Output panel (View → Output)
2. Select "OmniLang Language Server" from the dropdown
3. Look for error messages about standard library loading
4. Ensure the workspace contains the `omni/std` directory or configure the path

## Development

### Building from Source

```bash
# Install dependencies
npm install

# Compile TypeScript
npm run compile

# Watch mode for development
npm run watch

# Package extension
npx vsce package
```

### Project Structure

```
vscode/omni/
├── src/
│   ├── extension.ts      # Main extension entry point
│   ├── client.ts         # LSP client
│   └── server/           # LSP server implementation
│       ├── index.ts      # Server entry point
│       ├── completion.ts # Autocomplete handler
│       ├── hover.ts      # Hover information
│       ├── stdlib.ts     # Standard library parser
│       └── ...
├── dist/                 # Compiled JavaScript (generated)
├── syntaxes/             # TextMate grammar
├── snippets/             # Code snippets
└── package.json          # Extension manifest
```

## Contributing

Contributions are welcome! Please see the main [OmniLang repository](https://github.com/omni-lang/omni) for contribution guidelines.

## License

This extension is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Links

- **Repository**: [https://github.com/omni-lang/omni](https://github.com/omni-lang/omni)
- **Documentation**: [https://omni-lang.github.io/omni/](https://omni-lang.github.io/omni/)
- **Issues**: [GitHub Issues](https://github.com/omni-lang/omni/issues)

## Acknowledgments

Built with:
- [VS Code Language Server Protocol](https://microsoft.github.io/language-server-protocol/)
- [TypeScript](https://www.typescriptlang.org/)
- [vscode-languageclient](https://github.com/Microsoft/vscode-languageserver-node)

---

**OmniLang** - *One language to rule them all*

