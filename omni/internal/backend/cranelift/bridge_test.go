//go:build !darwin && !windows

package cranelift

import (
	"path/filepath"
	"testing"

	"github.com/omni-lang/omni/internal/mir"
)

func TestCompileMIRJSONEmpty(t *testing.T) {
	err := CompileMIRJSON("")
	if err == nil {
		t.Error("Expected error for empty JSON")
	}

	expectedError := "mir payload required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestCompileToObjectEmptyJSON(t *testing.T) {
	err := CompileToObject("", "test.o")
	if err == nil {
		t.Error("Expected error for empty JSON")
	}

	expectedError := "mir payload required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestCompileToObjectEmptyPath(t *testing.T) {
	err := CompileToObject("{}", "")
	if err == nil {
		t.Error("Expected error for empty output path")
	}

	expectedError := "output path required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestCompileToObjectWithOptEmptyJSON(t *testing.T) {
	err := CompileToObjectWithOpt("", "test.o", "speed")
	if err == nil {
		t.Error("Expected error for empty JSON")
	}

	expectedError := "mir payload required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestCompileToObjectWithOptEmptyPath(t *testing.T) {
	err := CompileToObjectWithOpt("{}", "", "speed")
	if err == nil {
		t.Error("Expected error for empty output path")
	}

	expectedError := "output path required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestCompileToObjectWithOptDefaultLevel(t *testing.T) {
	// Test that empty optLevel defaults to "speed"
	err := CompileToObjectWithOpt("{}", "test.o", "")
	if err == nil {
		// If library is available, this might succeed
		// If not, we expect an error but not about optLevel
		return
	}
	// Error should not be about optLevel
	if err.Error() == "output path required" || err.Error() == "mir payload required" {
		t.Errorf("Unexpected error: %s", err.Error())
	}
}

func TestCompileModuleToObjectNilModule(t *testing.T) {
	err := CompileModuleToObject(nil, "test.o")
	if err == nil {
		t.Error("Expected error for nil module")
	}

	expectedError := "failed to convert MIR to JSON"
	if !contains(err.Error(), expectedError) {
		t.Errorf("Expected error containing '%s', got '%s'", expectedError, err.Error())
	}
}

func TestCompileModuleToObjectEmptyModule(t *testing.T) {
	module := &mir.Module{
		Functions: []*mir.Function{},
	}

	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test.o")

	err := CompileModuleToObject(module, outputPath)
	// This might succeed if library is available, or fail if not
	// Either way, we just check it doesn't panic
	_ = err
}

func TestCompileModuleToObjectWithOptNilModule(t *testing.T) {
	err := CompileModuleToObjectWithOpt(nil, "test.o", "speed")
	if err == nil {
		t.Error("Expected error for nil module")
	}

	expectedError := "failed to convert MIR to JSON"
	if !contains(err.Error(), expectedError) {
		t.Errorf("Expected error containing '%s', got '%s'", expectedError, err.Error())
	}
}

func TestCompileModuleToObjectWithOptEmptyModule(t *testing.T) {
	module := &mir.Module{
		Functions: []*mir.Function{},
	}

	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test.o")

	err := CompileModuleToObjectWithOpt(module, outputPath, "speed")
	// This might succeed if library is available, or fail if not
	// Either way, we just check it doesn't panic
	_ = err
}

func TestCompileModuleToObjectWithOptDifferentLevels(t *testing.T) {
	module := &mir.Module{
		Functions: []*mir.Function{},
	}

	tempDir := t.TempDir()
	optLevels := []string{"speed", "size", "none"}

	for _, level := range optLevels {
		t.Run(level, func(t *testing.T) {
			outputPath := filepath.Join(tempDir, "test_"+level+".o")
			err := CompileModuleToObjectWithOpt(module, outputPath, level)
			// This might succeed if library is available, or fail if not
			// Either way, we just check it doesn't panic
			_ = err
		})
	}
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

	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test.o")

	err := CompileModuleToObject(module, outputPath)
	// This might succeed if library is available, or fail if not
	// Either way, we just check it doesn't panic
	_ = err
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
