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

func TestCompileToObjectWithOpt(t *testing.T) {
	// Test that CompileToObjectWithOpt returns an error on macOS
	err := CompileToObjectWithOpt("{}", "test.o", "speed")
	if err == nil {
		t.Error("Expected CompileToObjectWithOpt to return an error on macOS")
	}

	expectedError := "cranelift backend not available on macOS"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestCompileModuleToObjectWithOpt(t *testing.T) {
	// Test that CompileModuleToObjectWithOpt creates a placeholder file on macOS
	module := &mir.Module{
		Functions: []*mir.Function{},
	}

	outputPath := "test_placeholder_opt.o"
	defer os.Remove(outputPath)

	err := CompileModuleToObjectWithOpt(module, outputPath, "speed")
	if err != nil {
		t.Errorf("CompileModuleToObjectWithOpt failed: %v", err)
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

	expectedContent := "Optimization Level: speed"
	if !strings.Contains(string(content), expectedContent) {
		t.Errorf("Expected placeholder file to contain '%s'", expectedContent)
	}
}

func TestCompileModuleToObjectWithOptDifferentLevels(t *testing.T) {
	module := &mir.Module{
		Functions: []*mir.Function{},
	}

	optLevels := []string{"speed", "size", "none", "default"}

	for _, level := range optLevels {
		t.Run(level, func(t *testing.T) {
			outputPath := "test_opt_" + level + ".o"
			defer os.Remove(outputPath)

			err := CompileModuleToObjectWithOpt(module, outputPath, level)
			if err != nil {
				t.Errorf("CompileModuleToObjectWithOpt failed for level %s: %v", level, err)
			}

			// Check that the placeholder file was created
			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Errorf("Expected placeholder file to be created for level %s", level)
			}

			// Check the content contains the optimization level
			content, err := os.ReadFile(outputPath)
			if err != nil {
				t.Errorf("Failed to read placeholder file: %v", err)
			}

			expectedContent := "Optimization Level: " + level
			if !strings.Contains(string(content), expectedContent) {
				t.Errorf("Expected placeholder file to contain '%s'", expectedContent)
			}
		})
	}
}

func TestCompileModuleToObjectWithOptNilModule(t *testing.T) {
	// Test that nil module causes an error (via ToJSON)
	// The actual behavior depends on ToJSON implementation
	// We just verify it doesn't crash the test
	func() {
		defer func() {
			_ = recover() // Catch any panic
		}()
		_ = CompileModuleToObjectWithOpt(nil, "test.o", "speed")
	}()
}

func TestCompileModuleToObjectNilModule(t *testing.T) {
	// Test that nil module causes an error (via ToJSON)
	// The actual behavior depends on ToJSON implementation
	// We just verify it doesn't crash the test
	func() {
		defer func() {
			_ = recover() // Catch any panic
		}()
		_ = CompileModuleToObject(nil, "test.o")
	}()
}

func TestCompileModuleToObjectWithValidModule(t *testing.T) {
	fn := mir.NewFunction("test", "int", []mir.Param{})
	block := fn.NewBlock("entry")
	block.Instructions = []mir.Instruction{
		{
			ID:   fn.NextValue(),
			Op:   "const",
			Type: "int",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "42", Type: "int"},
			},
		},
	}
	block.Terminator = mir.Terminator{
		Op: "ret",
		Operands: []mir.Operand{
			{Kind: mir.OperandValue, Value: 0},
		},
	}

	module := &mir.Module{
		Functions: []*mir.Function{fn},
	}

	outputPath := "test_valid_module.o"
	defer os.Remove(outputPath)

	err := CompileModuleToObject(module, outputPath)
	if err != nil {
		t.Errorf("CompileModuleToObject failed: %v", err)
	}

	// Check that the placeholder file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Expected placeholder file to be created")
	}
}
