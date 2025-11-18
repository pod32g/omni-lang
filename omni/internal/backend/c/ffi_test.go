package cbackend

import (
	"testing"
)

func TestFFI(t *testing.T) {
	t.Run("Print", func(t *testing.T) {
		// Test that Print function exists and can be called
		// This is a simple test to ensure the function is available
		result := Print("test")

		// When CGO is disabled, Print returns an error (stub implementation)
		// When CGO is enabled, Print returns nil (real implementation)
		// Both cases are valid, we just need to ensure the function exists
		if result != nil {
			// Check if it's the expected stub error
			if result.Error() != "CGO is required for runtime functions. Enable CGO to use Print()" {
				t.Errorf("Unexpected error from Print function: %v", result)
			}
			// This is expected when CGO is disabled, so we don't fail the test
		}
		// If result is nil, that's also valid (when CGO is enabled)
	})
}
