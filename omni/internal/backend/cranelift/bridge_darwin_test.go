//go:build darwin
// +build darwin

package cranelift

import (
	"os"
	"strings"
	"testing"

	"github.com/omni-lang/omni/internal/mir"
)

func TestCompileMIRJSON(t *testing.T) {
	// Test that CompileMIRJSON returns an error on macOS
	err := CompileMIRJSON("{}")
	if err == nil {
		t.Error("Expected CompileMIRJSON to return an error on macOS")
	}

	expectedError := "cranelift backend not available on macOS"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestCompileToObject(t *testing.T) {
	// Test that CompileToObject returns an error on macOS
	err := CompileToObject("{}", "test.o")
	if err == nil {
		t.Error("Expected CompileToObject to return an error on macOS")
	}

	expectedError := "cranelift backend not available on macOS"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestCompileModuleToObject(t *testing.T) {
	// Test that CompileModuleToObject creates a placeholder file on macOS
	module := &mir.Module{
		Functions: []*mir.Function{},
	}

	outputPath := "test_placeholder.o"
	defer os.Remove(outputPath)

	err := CompileModuleToObject(module, outputPath)
	if err != nil {
		t.Errorf("CompileModuleToObject failed: %v", err)
	}

	// Check that the placeholder file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Expected placeholder file to be created")
	}

	// Check the content of the placeholder file
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Errorf("Failed to read placeholder file: %v", err)
	}

	expectedContent := "# OmniLang Object File Placeholder (macOS)"
	if !strings.Contains(string(content), expectedContent) {
		t.Errorf("Expected placeholder file to contain '%s'", expectedContent)
	}
}
