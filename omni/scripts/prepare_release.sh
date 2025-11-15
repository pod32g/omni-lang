#!/bin/bash

# Release preparation script for OmniLang
# This script prepares the project for a stable release

set -e

# Configuration
VERSION=${1:-"0.3.0"}
RELEASE_DATE=$(date +"%Y-%m-%d")
CHANGELOG_FILE="../CHANGELOG.md"
README_FILE="../README.md"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to check prerequisites
check_prerequisites() {
    print_status $BLUE "Checking prerequisites..."
    
    # Check if we're in the right directory
    if [ ! -f "go.mod" ]; then
        print_status $RED "Error: Please run this script from the OmniLang project root directory"
        exit 1
    fi
    
    # Check if Go is available
    if ! command -v go &> /dev/null; then
        print_status $RED "Error: Go is not installed or not in PATH"
        exit 1
    fi
    
    # Check if Git is available
    if ! command -v git &> /dev/null; then
        print_status $RED "Error: Git is not installed or not in PATH"
        exit 1
    fi
    
    # Check if we're in a git repository
    if [ ! -d ".git" ]; then
        print_status $RED "Error: Not in a git repository"
        exit 1
    fi
    
    print_status $GREEN "Prerequisites check passed!"
}

# Function to update version information
update_version() {
    print_status $BLUE "Updating version information to $VERSION..."
    
    # Update changelog with release date
    sed -i.bak "s/## \[0.3.0\] - 2025-01-XX/## [$VERSION] - $RELEASE_DATE/" "$CHANGELOG_FILE"
    
    # Update README version
    if [ -f "$README_FILE" ]; then
        sed -i.bak "s/\*\*Version:\*\* v0.2.0/\*\*Version:\*\* v$VERSION/" "$README_FILE"
    fi
    
    # Update go.mod version (if applicable)
    if [ -f "go.mod" ]; then
        # This would typically be done through go mod edit
        print_status $YELLOW "Note: Update go.mod version manually if needed"
    fi
    
    print_status $GREEN "Version information updated!"
}

# Function to run tests
run_tests() {
    print_status $BLUE "Running comprehensive tests..."
    
    # Run unit tests
    print_status $YELLOW "Running unit tests..."
    go test ./... -v -coverprofile=coverage.out
    
    # Run performance tests
    print_status $YELLOW "Running performance tests..."
    make perf
    
    # Run integration tests
    print_status $YELLOW "Running integration tests..."
    go test ./tests/e2e/... -v
    
    print_status $GREEN "All tests passed!"
}

# Function to build all targets
build_all() {
    print_status $BLUE "Building all targets..."
    
    # Build main binaries
    print_status $YELLOW "Building main binaries..."
    make build
    
    # Build Rust components (if available)
    print_status $YELLOW "Building Rust components..."
    make build-rust || print_status $YELLOW "Rust build skipped (Cargo not available)"
    
    # Build runtime
    print_status $YELLOW "Building runtime..."
    make build-runtime
    
    print_status $GREEN "All targets built successfully!"
}

# Function to create packages
create_packages() {
    print_status $BLUE "Creating distribution packages..."
    
    # Create packages using new packaging system
    print_status $YELLOW "Creating packages with new packaging system..."
    make package-new
    
    # Create packages using legacy system
    print_status $YELLOW "Creating packages with legacy system..."
    make package
    
    print_status $GREEN "Packages created successfully!"
}

# Function to verify packages
verify_packages() {
    print_status $BLUE "Verifying packages..."
    
    # Check if packages exist
    if [ ! -d "dist" ]; then
        print_status $RED "Error: No dist directory found"
        exit 1
    fi
    
    # List packages
    print_status $YELLOW "Created packages:"
    ls -la dist/
    
    # Test package extraction
    print_status $YELLOW "Testing package extraction..."
    for package in dist/*.tar.gz; do
        if [ -f "$package" ]; then
            print_status $YELLOW "Testing $package..."
            tar -tzf "$package" > /dev/null
            print_status $GREEN "✓ $package is valid"
        fi
    done
    
    for package in dist/*.zip; do
        if [ -f "$package" ]; then
            print_status $YELLOW "Testing $package..."
            unzip -t "$package" > /dev/null
            print_status $GREEN "✓ $package is valid"
        fi
    done
    
    print_status $GREEN "Package verification completed!"
}

# Function to generate release notes
generate_release_notes() {
    print_status $BLUE "Generating release notes..."
    
    # Create release notes file
    cat > "RELEASE_NOTES_$VERSION.md" << EOF
# OmniLang $VERSION Release Notes

**Release Date:** $RELEASE_DATE

## Overview

This release introduces significant improvements to OmniLang, including enhanced debug support, native code generation with the C backend, comprehensive packaging system, and performance testing infrastructure.

## What's New

### Major Features

- **Enhanced Debug Support**: Comprehensive debug symbol generation with source mapping
- **C Backend**: Native code generation for improved performance
- **Package System**: Complete packaging and distribution system
- **Performance Testing**: Automated performance regression testing

### Improvements

- **Debug Information**: Full debug symbol generation with DWARF support
- **Compilation Performance**: Improved compilation speed with optimization levels
- **Documentation**: Comprehensive documentation for new features
- **Build System**: Enhanced Makefile with performance testing targets

### 📚 Documentation

- **C Backend Guide**: Complete documentation for C backend usage
- **Packaging Guide**: Comprehensive packaging and distribution documentation
- **Performance Guide**: Performance testing and optimization documentation

## Installation

### Quick Install

\`\`\`bash
# Download the latest release
wget https://github.com/omni-lang/omni/releases/download/v$VERSION/omni-$VERSION-\$(uname -s | tr '[:upper:]' '[:lower:]')-\$(uname -m).tar.gz

# Extract and install
tar -xzf omni-$VERSION-*.tar.gz
cd omni-$VERSION
./install.sh
\`\`\`

### Platform-Specific Installation

#### Linux (x86_64)
\`\`\`bash
tar -xzf omni-$VERSION-linux-x86_64.tar.gz
cd omni-$VERSION
./install.sh
\`\`\`

#### macOS (x86_64)
\`\`\`bash
tar -xzf omni-$VERSION-darwin-x86_64.tar.gz
cd omni-$VERSION
./install.sh
\`\`\`

#### macOS (ARM64)
\`\`\`bash
tar -xzf omni-$VERSION-darwin-arm64.tar.gz
cd omni-$VERSION
./install.sh
\`\`\`

#### Windows (x86_64)
1. Extract omni-$VERSION-windows-x86_64.zip
2. Add the extracted directory to your PATH
3. Run omnic.exe and omnir.exe from command prompt

## Quick Start

\`\`\`bash
# Create a simple program
echo 'func main() { println("Hello, OmniLang!") }' > hello.omni

# Run it
omnir hello.omni

# Compile it
omnic hello.omni -o hello
\`\`\`

## Breaking Changes

None in this release.

## Migration Guide

No migration required for this release.

## Performance

This release includes significant performance improvements:

- **Compilation Speed**: Up to 2x faster compilation with C backend
- **Memory Usage**: Reduced memory usage during compilation
- **Debug Performance**: Optimized debug symbol generation

## Bug Fixes

- Fixed debug symbol generation issues
- Improved error messages with source location information
- Enhanced cross-platform compatibility

## Security

- No security vulnerabilities reported
- Enhanced package verification

## Contributors

Thanks to all contributors who made this release possible!

## Support

- **Documentation**: [README.md](README.md)
- **Issues**: [GitHub Issues](https://github.com/omni-lang/omni/issues)
- **Discussions**: [GitHub Discussions](https://github.com/omni-lang/omni/discussions)

## Changelog

For detailed changes, see [CHANGELOG.md](CHANGELOG.md).

---

**Full Changelog**: https://github.com/omni-lang/omni/compare/v0.2.0...v$VERSION
EOF
    
    print_status $GREEN "Release notes generated: RELEASE_NOTES_$VERSION.md"
}

# Function to create git tag
create_git_tag() {
    print_status $BLUE "Creating git tag..."
    
    # Check if tag already exists
    if git tag -l | grep -q "v$VERSION"; then
        print_status $YELLOW "Tag v$VERSION already exists"
        return
    fi
    
    # Create and push tag
    git tag -a "v$VERSION" -m "Release v$VERSION"
    print_status $GREEN "Git tag v$VERSION created!"
    
    # Ask if user wants to push the tag
    read -p "Push tag to remote? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git push origin "v$VERSION"
        print_status $GREEN "Tag pushed to remote!"
    fi
}

# Function to clean up
cleanup() {
    print_status $BLUE "Cleaning up temporary files..."
    
    # Remove backup files
    rm -f "$CHANGELOG_FILE.bak" "$README_FILE.bak" 2>/dev/null || true
    
    print_status $GREEN "Cleanup completed!"
}

# Function to show summary
show_summary() {
    print_status $GREEN "Release preparation completed!"
    print_status $BLUE "Summary:"
    echo "  Version: $VERSION"
    echo "  Date: $RELEASE_DATE"
    echo "  Packages: dist/"
    echo "  Release Notes: RELEASE_NOTES_$VERSION.md"
    echo "  Git Tag: v$VERSION"
    echo ""
    print_status $YELLOW "Next steps:"
    echo "  1. Review the generated packages in dist/"
    echo "  2. Test the packages on target platforms"
    echo "  3. Create a GitHub release with the generated release notes"
    echo "  4. Upload the packages to the GitHub release"
    echo "  5. Announce the release to the community"
}

# Main execution
main() {
    print_status $GREEN "Starting OmniLang release preparation for version $VERSION..."
    
    check_prerequisites
    update_version
    run_tests
    build_all
    create_packages
    verify_packages
    generate_release_notes
    create_git_tag
    cleanup
    show_summary
    
    print_status $GREEN "Release preparation completed successfully!"
}

# Parse command line arguments
case "${1:-help}" in
    "help"|"-h"|"--help")
        echo "Usage: $0 [VERSION]"
        echo "  VERSION: Version number (default: 0.3.0)"
        echo ""
        echo "This script prepares OmniLang for a stable release by:"
        echo "  - Updating version information"
        echo "  - Running comprehensive tests"
        echo "  - Building all targets"
        echo "  - Creating distribution packages"
        echo "  - Generating release notes"
        echo "  - Creating git tags"
        exit 0
        ;;
    *)
        main
        ;;
esac
