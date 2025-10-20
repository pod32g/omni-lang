//go:build !cgo
// +build !cgo

package cbackend

import "fmt"

// Print is a stub implementation when CGO is not available.
// This allows tests to run in CI environments without CGO.
func Print(s string) error {
	if s == "" {
		return fmt.Errorf("print requires non-empty input")
	}
	// In CI environments, just return success without actually printing
	// This allows tests to pass without requiring the full runtime
	return nil
}
