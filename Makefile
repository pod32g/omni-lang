SHELL := /bin/bash

VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO := go
CARGO := cargo
GO_PROJECT_DIR := omni
BIN_DIR := $(GO_PROJECT_DIR)/bin
DIST_DIR := dist
RELEASES_DIR := releases
DIST_PACKAGE_DIR := $(DIST_DIR)/omni-$(VERSION)
PACKAGING_DIR := packaging
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"
.SHELLFLAGS := -eu -o pipefail -c

.ONESHELL:
.PHONY: all fmt lint test build bench clean gen build-rust package release

all: build

fmt:
	cd $(GO_PROJECT_DIR) && $(GO) fmt ./...
	cd $(GO_PROJECT_DIR) && $(GO) vet ./...

lint:
	cd $(GO_PROJECT_DIR) && $(GO) vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		cd $(GO_PROJECT_DIR) && golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping advanced linting"; \
	fi

test:
	cd $(GO_PROJECT_DIR) && $(GO) test ./... -count=1 -v -coverprofile=coverage.out
	cd $(GO_PROJECT_DIR) && $(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-coverage:
	cd $(GO_PROJECT_DIR) && $(GO) test ./... -count=1 -coverprofile=coverage.out -covermode=count
	cd $(GO_PROJECT_DIR) && $(GO) tool cover -func=coverage.out | tail -1

bench:
	cd $(GO_PROJECT_DIR) && $(GO) test ./tests/bench -bench=. -benchmem -count=3

build: build-rust
	@mkdir -p $(BIN_DIR)
	cd $(GO_PROJECT_DIR) && $(GO) build $(LDFLAGS) -o bin/omnic ./cmd/omnic
	cd $(GO_PROJECT_DIR) && $(GO) build $(LDFLAGS) -o bin/omnir ./cmd/omnir
	cd $(GO_PROJECT_DIR) && $(GO) build $(LDFLAGS) -o bin/omnipkg ./cmd/omnipkg

build-rust:
	@if command -v cargo >/dev/null 2>&1; then \
		cd $(GO_PROJECT_DIR)/native/clift && cargo build --release; \
		echo "Rust components built successfully"; \
	else \
		echo "Cargo not found, skipping Rust build (Cranelift backend will be unavailable)"; \
	fi

build-all: build

clean:
	rm -rf $(BIN_DIR)
	rm -rf $(GO_PROJECT_DIR)/native/clift/target
	rm -rf $(DIST_DIR)
	rm -rf $(RELEASES_DIR)
	rm -f $(GO_PROJECT_DIR)/coverage.out $(GO_PROJECT_DIR)/coverage.html

gen:
	@echo "no codegen yet"

# Package creation
package: build-all
	@echo "Creating distribution packages..."
	@mkdir -p $(DIST_PACKAGE_DIR)
	
	# Copy binaries
	@cp $(BIN_DIR)/* $(DIST_PACKAGE_DIR)/
	
	# Copy standard library
	@cp -r $(GO_PROJECT_DIR)/std $(DIST_PACKAGE_DIR)/
	
	# Copy examples
	@cp -r $(GO_PROJECT_DIR)/examples $(DIST_PACKAGE_DIR)/
	
	# Copy documentation
	@cp README.md $(DIST_PACKAGE_DIR)/
	@cp CHANGELOG.md $(DIST_PACKAGE_DIR)/
	@cp CONTRIBUTING.md $(DIST_PACKAGE_DIR)/
	@cp -r docs $(DIST_PACKAGE_DIR)/
	
	# Copy runtime libraries
	@mkdir -p $(DIST_PACKAGE_DIR)/lib
	@if [ -f $(GO_PROJECT_DIR)/runtime/posix/libomni_rt.so ]; then \
		cp $(GO_PROJECT_DIR)/runtime/posix/libomni_rt.so $(DIST_PACKAGE_DIR)/lib/; \
	fi
	@if [ -f native/clift/target/release/libomni_clift.so ]; then \
		cp native/clift/target/release/libomni_clift.so $(DIST_PACKAGE_DIR)/lib/; \
	fi
	@if [ -f native/clift/target/release/libomni_clift.dylib ]; then \
		cp native/clift/target/release/libomni_clift.dylib $(DIST_PACKAGE_DIR)/lib/; \
	fi
	@if [ -f native/clift/target/release/omni_clift.dll ]; then \
		cp native/clift/target/release/omni_clift.dll $(DIST_PACKAGE_DIR)/lib/; \
	fi
	
	# Create installation script
	@sed 's/{{VERSION}}/$(VERSION)/g' $(PACKAGING_DIR)/install.sh.tpl > $(DIST_PACKAGE_DIR)/install.sh
	@chmod +x $(DIST_PACKAGE_DIR)/install.sh
	
	# Create uninstall script
	@sed 's/{{VERSION}}/$(VERSION)/g' $(PACKAGING_DIR)/uninstall.sh.tpl > $(DIST_PACKAGE_DIR)/uninstall.sh
	@chmod +x $(DIST_PACKAGE_DIR)/uninstall.sh
	
	# Create README for distribution
	@sed 's/{{VERSION}}/$(VERSION)/g' $(PACKAGING_DIR)/README.txt.tpl > $(DIST_PACKAGE_DIR)/README.txt
	
	# Create platform-specific packages
	@cd $(DIST_DIR) && tar -czf omni-$(VERSION)-linux-x86_64.tar.gz omni-$(VERSION)/
	@cd $(DIST_DIR) && zip -r omni-$(VERSION)-windows-x86_64.zip omni-$(VERSION)/
	@cd $(DIST_DIR) && tar -czf omni-$(VERSION)-darwin-x86_64.tar.gz omni-$(VERSION)/
	@cd $(DIST_DIR) && tar -czf omni-$(VERSION)-darwin-arm64.tar.gz omni-$(VERSION)/
	
	@echo "Distribution packages created in $(DIST_DIR)/"
	@ls -la $(DIST_DIR)/

# Release creation
release: package
	@echo "Creating release packages..."
	@mkdir -p $(RELEASES_DIR)
	@cp $(DIST_DIR)/*.tar.gz $(RELEASES_DIR)/
	@cp $(DIST_DIR)/*.zip $(RELEASES_DIR)/
	
	# Create checksums
	@cd $(RELEASES_DIR) && sha256sum *.tar.gz *.zip > checksums.txt
	
	# Create release notes
	@sed 's/{{VERSION}}/$(VERSION)/g' $(PACKAGING_DIR)/RELEASE_NOTES.md.tpl > $(RELEASES_DIR)/RELEASE_NOTES.md
	
	@echo "Release packages created in $(RELEASES_DIR)/"
	@ls -la $(RELEASES_DIR)/

# Development helpers
dev-setup:
	@echo "Setting up development environment..."
	@cd $(GO_PROJECT_DIR) && go mod download
	@cd $(GO_PROJECT_DIR) && go install golang.org/x/tools/cmd/goimports@latest
	@cd $(GO_PROJECT_DIR) && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@if command -v cargo >/dev/null 2>&1; then \
		cd $(GO_PROJECT_DIR)/native/clift && cargo build; \
	else \
		echo "Cargo not found, skipping Rust setup"; \
	fi
	@echo "Development environment ready!"

# CI helpers
ci-test: test test-coverage
ci-build: build build-rust
ci-package: package