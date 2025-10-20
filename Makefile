SHELL := /bin/bash

GO := go
CARGO := cargo
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

.PHONY: all fmt lint test build bench clean gen build-rust package release

all: build

fmt:
	$(GO) fmt ./...
	$(GO) vet ./...

lint:
	$(GO) vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping advanced linting"; \
	fi

test:
	$(GO) test ./... -count=1 -v -coverprofile=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-coverage:
	$(GO) test ./... -count=1 -coverprofile=coverage.out -covermode=count
	$(GO) tool cover -func=coverage.out | tail -1

bench:
	$(GO) test ./tests/bench -bench=. -benchmem -count=3

build:
	@mkdir -p bin
	$(GO) build $(LDFLAGS) -o bin/omnic ./cmd/omnic
	$(GO) build $(LDFLAGS) -o bin/omnir ./cmd/omnir

build-rust:
	@if command -v cargo >/dev/null 2>&1; then \
		cd native/clift && cargo build --release; \
		@echo "Rust components built successfully"; \
	else \
		echo "Cargo not found, skipping Rust build"; \
	fi

build-all: build build-rust

clean:
	rm -rf bin
	rm -rf target
	rm -rf native/clift/target
	rm -rf dist
	rm -rf releases
	rm -f coverage.out coverage.html

gen:
	@echo "no codegen yet"

# Package creation
package: build-all
	@echo "Creating distribution packages..."
	@mkdir -p dist
	@mkdir -p dist/omni-$(VERSION)
	
	# Copy binaries
	@cp bin/* dist/omni-$(VERSION)/
	
	# Copy standard library
	@cp -r std dist/omni-$(VERSION)/
	
	# Copy examples
	@cp -r examples dist/omni-$(VERSION)/
	
	# Copy documentation
	@cp README.md dist/omni-$(VERSION)/
	@cp CHANGELOG.md dist/omni-$(VERSION)/
	@cp CONTRIBUTING.md dist/omni-$(VERSION)/
	@cp -r docs dist/omni-$(VERSION)/
	
	# Copy runtime libraries
	@mkdir -p dist/omni-$(VERSION)/lib
	@if [ -f native/clift/target/release/libomni_clift.so ]; then \
		cp native/clift/target/release/libomni_clift.so dist/omni-$(VERSION)/lib/; \
	fi
	@if [ -f native/clift/target/release/libomni_clift.dylib ]; then \
		cp native/clift/target/release/libomni_clift.dylib dist/omni-$(VERSION)/lib/; \
	fi
	@if [ -f native/clift/target/release/omni_clift.dll ]; then \
		cp native/clift/target/release/omni_clift.dll dist/omni-$(VERSION)/lib/; \
	fi
	
	# Create installation script
	@cat > dist/omni-$(VERSION)/install.sh << 'EOF'
#!/bin/bash
# OmniLang Installation Script

set -e

INSTALL_DIR="/usr/local/omni"
BIN_DIR="/usr/local/bin"

echo "Installing OmniLang $(VERSION)..."

# Create installation directory
sudo mkdir -p $INSTALL_DIR
sudo mkdir -p $BIN_DIR

# Copy files
sudo cp -r . $INSTALL_DIR/

# Create symlinks
sudo ln -sf $INSTALL_DIR/omnic $BIN_DIR/omnic
sudo ln -sf $INSTALL_DIR/omnir $BIN_DIR/omnir

# Set permissions
sudo chmod +x $INSTALL_DIR/omnic
sudo chmod +x $INSTALL_DIR/omnir

echo "OmniLang installed successfully!"
echo "Run 'omnic --version' to verify installation"
EOF
	@chmod +x dist/omni-$(VERSION)/install.sh
	
	# Create uninstall script
	@cat > dist/omni-$(VERSION)/uninstall.sh << 'EOF'
#!/bin/bash
# OmniLang Uninstallation Script

set -e

INSTALL_DIR="/usr/local/omni"
BIN_DIR="/usr/local/bin"

echo "Uninstalling OmniLang..."

# Remove symlinks
sudo rm -f $BIN_DIR/omnic
sudo rm -f $BIN_DIR/omnir

# Remove installation directory
sudo rm -rf $INSTALL_DIR

echo "OmniLang uninstalled successfully!"
EOF
	@chmod +x dist/omni-$(VERSION)/uninstall.sh
	
	# Create README for distribution
	@cat > dist/omni-$(VERSION)/README.txt << 'EOF'
OmniLang $(VERSION) Distribution

This package contains:
- omnic: The OmniLang compiler
- omnir: The OmniLang runner
- std/: Standard library modules
- examples/: Example programs
- docs/: Documentation
- lib/: Runtime libraries

Installation:
  ./install.sh

Uninstallation:
  ./uninstall.sh

Quick Start:
  ./omnir examples/hello_world.omni

For more information, see README.md
EOF
	
	# Create platform-specific packages
	@cd dist && tar -czf omni-$(VERSION)-linux-x86_64.tar.gz omni-$(VERSION)/
	@cd dist && zip -r omni-$(VERSION)-windows-x86_64.zip omni-$(VERSION)/
	@cd dist && tar -czf omni-$(VERSION)-darwin-x86_64.tar.gz omni-$(VERSION)/
	@cd dist && tar -czf omni-$(VERSION)-darwin-arm64.tar.gz omni-$(VERSION)/
	
	@echo "Distribution packages created in dist/"
	@ls -la dist/

# Release creation
release: package
	@echo "Creating release packages..."
	@mkdir -p releases
	@cp dist/*.tar.gz releases/
	@cp dist/*.zip releases/
	
	# Create checksums
	@cd releases && sha256sum *.tar.gz *.zip > checksums.txt
	
	# Create release notes
	@cat > releases/RELEASE_NOTES.md << 'EOF'
# OmniLang $(VERSION) Release

## Installation

### Linux (x86_64)
```bash
tar -xzf omni-$(VERSION)-linux-x86_64.tar.gz
cd omni-$(VERSION)
./install.sh
```

### macOS (x86_64)
```bash
tar -xzf omni-$(VERSION)-darwin-x86_64.tar.gz
cd omni-$(VERSION)
./install.sh
```

### macOS (ARM64)
```bash
tar -xzf omni-$(VERSION)-darwin-arm64.tar.gz
cd omni-$(VERSION)
./install.sh
```

### Windows (x86_64)
1. Extract omni-$(VERSION)-windows-x86_64.zip
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
EOF
	
	@echo "Release packages created in releases/"
	@ls -la releases/

# Development helpers
dev-setup:
	@echo "Setting up development environment..."
	@go mod download
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@if command -v cargo >/dev/null 2>&1; then \
		cd native/clift && cargo build; \
	else \
		echo "Cargo not found, skipping Rust setup"; \
	fi
	@echo "Development environment ready!"

# CI helpers
ci-test: test test-coverage
ci-build: build build-rust
ci-package: package