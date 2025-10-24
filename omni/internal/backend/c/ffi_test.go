package cbackend

import (
	"testing"
)

func TestFFI(t *testing.T) {
	t.Run("Print", func(t *testing.T) {
		// Test that Print function exists and can be called
		// This is a simple test to ensure the function is available
		result := Print("test")
		if result != nil {
			t.Error("Expected nil result from Print function")
		}
	})
}
