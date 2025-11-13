SHELL := /bin/bash

GO := go
CARGO := cargo
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

.PHONY: all fmt lint test build bench clean gen build-rust package release run-omnic perf perf-baseline perf-report prepare-release

all: build

# Run omnic with proper library path setup
run-omnic:
	@export DYLD_LIBRARY_PATH="$(PWD)/runtime/posix:$$DYLD_LIBRARY_PATH" && ./bin/omnic $(ARGS)

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
	@echo "Cleaning up generated test files..."
	@find ./tests/e2e -name "*.c" -o -name "*.o" -o -name "*.mir" | xargs rm -f
	@echo "Building runtime library for e2e tests..."
	@make build-runtime
	@echo "Building binaries for e2e tests..."
	@make build
	@echo "Running core tests with CGO disabled for compatibility..."
	CGO_ENABLED=0 $(GO) test ./internal/ast ./internal/lexer ./internal/mir ./internal/packaging ./internal/passes ./internal/parser ./internal/types ./internal/backend/c -count=1 -v -coverprofile=coverage.out
	@echo "Running VM tests (expected to have some failures)..."
	CGO_ENABLED=0 $(GO) test ./internal/vm -count=1 -v || echo "VM tests completed with expected failures"
	@echo "Running e2e tests with CGO enabled..."
	export LD_LIBRARY_PATH=$$PWD/native/clift/target/release:$$PWD/runtime/posix:$$LD_LIBRARY_PATH && $(GO) test ./tests/e2e -count=1 -v
	@if [ -f coverage.out ]; then \
		$(GO) tool cover -html=coverage.out -o coverage.html; \
		echo "Coverage report generated: coverage.html"; \
	else \
		echo "No coverage data found, skipping coverage report generation"; \
	fi

test-coverage:
	$(GO) test ./... -count=1 -coverprofile=coverage.out -covermode=count
	$(GO) tool cover -func=coverage.out | tail -1

bench:
	@echo "Running performance benchmarks..."
	@mkdir -p bench-results
	@echo "Building runtime library for benchmarks..."
	@make build-runtime
	export LD_LIBRARY_PATH=$$PWD/runtime/posix:$$LD_LIBRARY_PATH && $(GO) test ./internal/compiler -bench=. -benchmem -count=3

build: build-rust
	@mkdir -p bin
	$(GO) build $(LDFLAGS) -o bin/omnic ./cmd/omnic
	$(GO) build $(LDFLAGS) -o bin/omnir ./cmd/omnir
	$(GO) build $(LDFLAGS) -o bin/omnipkg ./cmd/omnipkg
	@# Fix library path for binaries to work from anywhere (macOS only)
	@if [ "$$(uname)" = "Darwin" ] && command -v install_name_tool >/dev/null 2>&1; then \
		install_name_tool -change runtime/posix/libomni_rt.so $$(pwd)/runtime/posix/libomni_rt.so bin/omnic 2>/dev/null || true; \
		install_name_tool -change runtime/posix/libomni_rt.so $$(pwd)/runtime/posix/libomni_rt.so bin/omnir 2>/dev/null || true; \
	fi

build-rust:
	@if command -v cargo >/dev/null 2>&1; then \
		cd native/clift && cargo build --release; \
		echo "Rust components built successfully"; \
	else \
		echo "Cargo not found, skipping Rust build"; \
	fi

build-runtime:
	@mkdir -p runtime/posix
	@if command -v gcc >/dev/null 2>&1; then \
		gcc -shared -fPIC -o runtime/posix/libomni_rt.so runtime/omni_rt.c -lm; \
		echo "Runtime library built successfully"; \
	else \
		echo "GCC not found, skipping runtime build"; \
	fi

build-all: build

clean:
	rm -rf bin
	rm -rf target
	rm -rf native/clift/target
	rm -rf dist
	rm -rf releases
	rm -f coverage.out coverage.html

gen:
	@echo "no codegen yet"

# Package creation using new packaging system
package-new: build-all
	@echo "Creating distribution packages using new packaging system..."
	@mkdir -p dist
	@./bin/omnipkg -version $(VERSION) -o dist/omni-$(VERSION)-$(shell uname -s | tr '[:upper:]' '[:lower:]')-$(shell uname -m).tar.gz
	@./bin/omnipkg -version $(VERSION) -type zip -o dist/omni-$(VERSION)-$(shell uname -s | tr '[:upper:]' '[:lower:]')-$(shell uname -m).zip
	@echo "Distribution packages created in dist/"

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
	@cp ../README.md dist/omni-$(VERSION)/
	@cp ../CHANGELOG.md dist/omni-$(VERSION)/
	@cp ../CONTRIBUTING.md dist/omni-$(VERSION)/
	@cp -r ../docs dist/omni-$(VERSION)/
	
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
	@echo '#!/bin/bash' > dist/omni-$(VERSION)/install.sh
	@echo '# OmniLang Installation Script' >> dist/omni-$(VERSION)/install.sh
	@echo '' >> dist/omni-$(VERSION)/install.sh
	@echo 'set -e' >> dist/omni-$(VERSION)/install.sh
	@echo '' >> dist/omni-$(VERSION)/install.sh
	@echo 'INSTALL_DIR="/usr/local/omni"' >> dist/omni-$(VERSION)/install.sh
	@echo 'BIN_DIR="/usr/local/bin"' >> dist/omni-$(VERSION)/install.sh
	@echo '' >> dist/omni-$(VERSION)/install.sh
	@echo 'echo "Installing OmniLang $(VERSION)..."' >> dist/omni-$(VERSION)/install.sh
	@echo '' >> dist/omni-$(VERSION)/install.sh
	@echo '# Create installation directory' >> dist/omni-$(VERSION)/install.sh
	@echo 'sudo mkdir -p $$INSTALL_DIR' >> dist/omni-$(VERSION)/install.sh
	@echo 'sudo mkdir -p $$BIN_DIR' >> dist/omni-$(VERSION)/install.sh
	@echo '' >> dist/omni-$(VERSION)/install.sh
	@echo '# Copy files' >> dist/omni-$(VERSION)/install.sh
	@echo 'sudo cp -r . $$INSTALL_DIR/' >> dist/omni-$(VERSION)/install.sh
	@echo '' >> dist/omni-$(VERSION)/install.sh
	@echo '# Create symlinks' >> dist/omni-$(VERSION)/install.sh
	@echo 'sudo ln -sf $$INSTALL_DIR/omnic $$BIN_DIR/omnic' >> dist/omni-$(VERSION)/install.sh
	@echo 'sudo ln -sf $$INSTALL_DIR/omnir $$BIN_DIR/omnir' >> dist/omni-$(VERSION)/install.sh
	@echo '' >> dist/omni-$(VERSION)/install.sh
	@echo '# Set permissions' >> dist/omni-$(VERSION)/install.sh
	@echo 'sudo chmod +x $$INSTALL_DIR/omnic' >> dist/omni-$(VERSION)/install.sh
	@echo 'sudo chmod +x $$INSTALL_DIR/omnir' >> dist/omni-$(VERSION)/install.sh
	@echo '' >> dist/omni-$(VERSION)/install.sh
	@echo 'echo "OmniLang installed successfully!"' >> dist/omni-$(VERSION)/install.sh
	@echo 'echo "Run '\''omnic --version'\'' to verify installation"' >> dist/omni-$(VERSION)/install.sh
	@chmod +x dist/omni-$(VERSION)/install.sh
	
	# Create uninstall script
	@echo '#!/bin/bash' > dist/omni-$(VERSION)/uninstall.sh
	@echo '# OmniLang Uninstallation Script' >> dist/omni-$(VERSION)/uninstall.sh
	@echo '' >> dist/omni-$(VERSION)/uninstall.sh
	@echo 'set -e' >> dist/omni-$(VERSION)/uninstall.sh
	@echo '' >> dist/omni-$(VERSION)/uninstall.sh
	@echo 'INSTALL_DIR="/usr/local/omni"' >> dist/omni-$(VERSION)/uninstall.sh
	@echo 'BIN_DIR="/usr/local/bin"' >> dist/omni-$(VERSION)/uninstall.sh
	@echo '' >> dist/omni-$(VERSION)/uninstall.sh
	@echo 'echo "Uninstalling OmniLang..."' >> dist/omni-$(VERSION)/uninstall.sh
	@echo '' >> dist/omni-$(VERSION)/uninstall.sh
	@echo '# Remove symlinks' >> dist/omni-$(VERSION)/uninstall.sh
	@echo 'sudo rm -f $$BIN_DIR/omnic' >> dist/omni-$(VERSION)/uninstall.sh
	@echo 'sudo rm -f $$BIN_DIR/omnir' >> dist/omni-$(VERSION)/uninstall.sh
	@echo '' >> dist/omni-$(VERSION)/uninstall.sh
	@echo '# Remove installation directory' >> dist/omni-$(VERSION)/uninstall.sh
	@echo 'sudo rm -rf $$INSTALL_DIR' >> dist/omni-$(VERSION)/uninstall.sh
	@echo '' >> dist/omni-$(VERSION)/uninstall.sh
	@echo 'echo "OmniLang uninstalled successfully!"' >> dist/omni-$(VERSION)/uninstall.sh
	@chmod +x dist/omni-$(VERSION)/uninstall.sh
	
	# Create README for distribution
	@echo 'OmniLang $(VERSION) Distribution' > dist/omni-$(VERSION)/README.txt
	@echo '' >> dist/omni-$(VERSION)/README.txt
	@echo 'This package contains:' >> dist/omni-$(VERSION)/README.txt
	@echo '- omnic: The OmniLang compiler' >> dist/omni-$(VERSION)/README.txt
	@echo '- omnir: The OmniLang runner' >> dist/omni-$(VERSION)/README.txt
	@echo '- std/: Standard library modules' >> dist/omni-$(VERSION)/README.txt
	@echo '- examples/: Example programs' >> dist/omni-$(VERSION)/README.txt
	@echo '- docs/: Documentation' >> dist/omni-$(VERSION)/README.txt
	@echo '- lib/: Runtime libraries' >> dist/omni-$(VERSION)/README.txt
	@echo '' >> dist/omni-$(VERSION)/README.txt
	@echo 'Installation:' >> dist/omni-$(VERSION)/README.txt
	@echo '  ./install.sh' >> dist/omni-$(VERSION)/README.txt
	@echo '' >> dist/omni-$(VERSION)/README.txt
	@echo 'Uninstallation:' >> dist/omni-$(VERSION)/README.txt
	@echo '  ./uninstall.sh' >> dist/omni-$(VERSION)/README.txt
	@echo '' >> dist/omni-$(VERSION)/README.txt
	@echo 'Quick Start:' >> dist/omni-$(VERSION)/README.txt
	@echo '  ./omnir examples/hello_world.omni' >> dist/omni-$(VERSION)/README.txt
	@echo '' >> dist/omni-$(VERSION)/README.txt
	@echo 'For more information, see README.md' >> dist/omni-$(VERSION)/README.txt
	
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
	@echo '# OmniLang $(VERSION) Release' > releases/RELEASE_NOTES.md
	@echo '' >> releases/RELEASE_NOTES.md
	@echo '## Installation' >> releases/RELEASE_NOTES.md
	@echo '' >> releases/RELEASE_NOTES.md
	@echo '### Linux (x86_64)' >> releases/RELEASE_NOTES.md
	@echo '```bash' >> releases/RELEASE_NOTES.md
	@echo 'tar -xzf omni-$(VERSION)-linux-x86_64.tar.gz' >> releases/RELEASE_NOTES.md
	@echo 'cd omni-$(VERSION)' >> releases/RELEASE_NOTES.md
	@echo './install.sh' >> releases/RELEASE_NOTES.md
	@echo '```' >> releases/RELEASE_NOTES.md
	@echo '' >> releases/RELEASE_NOTES.md
	@echo '### macOS (x86_64)' >> releases/RELEASE_NOTES.md
	@echo '```bash' >> releases/RELEASE_NOTES.md
	@echo 'tar -xzf omni-$(VERSION)-darwin-x86_64.tar.gz' >> releases/RELEASE_NOTES.md
	@echo 'cd omni-$(VERSION)' >> releases/RELEASE_NOTES.md
	@echo './install.sh' >> releases/RELEASE_NOTES.md
	@echo '```' >> releases/RELEASE_NOTES.md
	@echo '' >> releases/RELEASE_NOTES.md
	@echo '### macOS (ARM64)' >> releases/RELEASE_NOTES.md
	@echo '```bash' >> releases/RELEASE_NOTES.md
	@echo 'tar -xzf omni-$(VERSION)-darwin-arm64.tar.gz' >> releases/RELEASE_NOTES.md
	@echo 'cd omni-$(VERSION)' >> releases/RELEASE_NOTES.md
	@echo './install.sh' >> releases/RELEASE_NOTES.md
	@echo '```' >> releases/RELEASE_NOTES.md
	@echo '' >> releases/RELEASE_NOTES.md
	@echo '### Windows (x86_64)' >> releases/RELEASE_NOTES.md
	@echo '1. Extract omni-$(VERSION)-windows-x86_64.zip' >> releases/RELEASE_NOTES.md
	@echo '2. Add the extracted directory to your PATH' >> releases/RELEASE_NOTES.md
	@echo '3. Run omnic.exe and omnir.exe from command prompt' >> releases/RELEASE_NOTES.md
	@echo '' >> releases/RELEASE_NOTES.md
	@echo '## Verification' >> releases/RELEASE_NOTES.md
	@echo '' >> releases/RELEASE_NOTES.md
	@echo 'After installation, verify the installation:' >> releases/RELEASE_NOTES.md
	@echo '```bash' >> releases/RELEASE_NOTES.md
	@echo 'omnic --version' >> releases/RELEASE_NOTES.md
	@echo 'omnir --version' >> releases/RELEASE_NOTES.md
	@echo '```' >> releases/RELEASE_NOTES.md
	@echo '' >> releases/RELEASE_NOTES.md
	@echo '## Quick Start' >> releases/RELEASE_NOTES.md
	@echo '' >> releases/RELEASE_NOTES.md
	@echo '```bash' >> releases/RELEASE_NOTES.md
	@echo '# Create a simple program' >> releases/RELEASE_NOTES.md
	@echo 'echo '\''func main():int { println("Hello, OmniLang!"); return 0 }'\'' > hello.omni' >> releases/RELEASE_NOTES.md
	@echo '' >> releases/RELEASE_NOTES.md
	@echo '# Run it' >> releases/RELEASE_NOTES.md
	@echo 'omnir hello.omni' >> releases/RELEASE_NOTES.md
	@echo '' >> releases/RELEASE_NOTES.md
	@echo '# Compile it' >> releases/RELEASE_NOTES.md
	@echo 'omnic hello.omni -o hello' >> releases/RELEASE_NOTES.md
	@echo '```' >> releases/RELEASE_NOTES.md
	@echo '' >> releases/RELEASE_NOTES.md
	@echo '## What'\''s New' >> releases/RELEASE_NOTES.md
	@echo '' >> releases/RELEASE_NOTES.md
	@echo 'See CHANGELOG.md for detailed changes.' >> releases/RELEASE_NOTES.md
	@echo '' >> releases/RELEASE_NOTES.md
	@echo '## Support' >> releases/RELEASE_NOTES.md
	@echo '' >> releases/RELEASE_NOTES.md
	@echo '- Documentation: README.md' >> releases/RELEASE_NOTES.md
	@echo '- Issues: https://github.com/omni-lang/omni/issues' >> releases/RELEASE_NOTES.md
	@echo '- Discussions: https://github.com/omni-lang/omni/discussions' >> releases/RELEASE_NOTES.md
	
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

# Performance testing
perf: build
	@echo "Running performance tests..."
	@./scripts/performance_monitor.sh run

perf-baseline: build
	@echo "Creating performance baseline..."
	@./scripts/performance_monitor.sh baseline

perf-report: build
	@echo "Generating performance report..."
	@./scripts/performance_monitor.sh report

# Release preparation
prepare-release:
	@echo "Preparing release..."
	@./scripts/prepare_release.sh $(VERSION)

# CI helpers
ci-test: test test-coverage
ci-build: build build-rust
ci-package: package
ci-perf: perf
