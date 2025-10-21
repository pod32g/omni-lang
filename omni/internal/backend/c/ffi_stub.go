//go:build !cgo
// +build !cgo

package cbackend

import "fmt"

// Print is a stub implementation when CGO is disabled.
// This allows the package to compile and run tests without requiring
// the runtime library to be linked at compile time.
func Print(s string) error {
	if s == "" {
		return fmt.Errorf("print requires non-empty input")
	}
	// In a real implementation, this would call the C runtime
	// For now, we just return an error indicating CGO is required
	return fmt.Errorf("CGO is required for runtime functions. Enable CGO to use Print()")
}
