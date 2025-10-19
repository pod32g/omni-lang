package runner

import (
	"errors"
	"fmt"
	"path/filepath"
)

// ErrNotImplemented signals that the execution engine is pending.
var ErrNotImplemented = errors.New("not implemented")

// Run interprets the provided OmniLang source file. Implementation will arrive
// once the virtual machine backend exists.
func Run(path string) error {
	if filepath.Ext(path) != ".omni" {
		return fmt.Errorf("%s: unsupported input (expected .omni)", path)
	}
	return fmt.Errorf("runner: %w", ErrNotImplemented)
}
