package compiler

import (
	"testing"
)

func TestCompile(t *testing.T) {
	// Test compiling with empty input path
	config := Config{
		InputPath: "",
	}

	err := Compile(config)
	if err == nil {
		t.Error("Expected error for empty input path")
	}

	expectedError := "input path required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestCompileWithUnsupportedBackend(t *testing.T) {
	// Test compiling with unsupported backend
	config := Config{
		InputPath: "test.omni",
		Backend:   "unsupported",
	}

	err := Compile(config)
	if err == nil {
		t.Error("Expected error for unsupported backend")
	}

	expectedError := "unsupported backend: unsupported"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestCompileWithValidBackend(t *testing.T) {
	// Test compiling with valid backend
	config := Config{
		InputPath: "test.omni",
		Backend:   "c",
	}

	err := Compile(config)
	// We expect an error because the input file doesn't exist
	if err == nil {
		t.Error("Expected error for non-existent input file")
	}
}

func TestConfig(t *testing.T) {
	// Test config creation
	config := Config{
		InputPath:    "test.omni",
		OutputPath:   "test.out",
		Backend:      "c",
		OptLevel:     "2",
		Emit:         "exe",
		Dump:         "mir",
		DebugInfo:    true,
		DebugModules: false,
	}

	if config.InputPath != "test.omni" {
		t.Errorf("Expected InputPath 'test.omni', got '%s'", config.InputPath)
	}

	if config.OutputPath != "test.out" {
		t.Errorf("Expected OutputPath 'test.out', got '%s'", config.OutputPath)
	}

	if config.Backend != "c" {
		t.Errorf("Expected Backend 'c', got '%s'", config.Backend)
	}

	if config.OptLevel != "2" {
		t.Errorf("Expected OptLevel '2', got '%s'", config.OptLevel)
	}

	if config.Emit != "exe" {
		t.Errorf("Expected Emit 'exe', got '%s'", config.Emit)
	}

	if config.Dump != "mir" {
		t.Errorf("Expected Dump 'mir', got '%s'", config.Dump)
	}

	if !config.DebugInfo {
		t.Error("Expected DebugInfo to be true")
	}

	if config.DebugModules {
		t.Error("Expected DebugModules to be false")
	}
}

func TestErrNotImplemented(t *testing.T) {
	// Test that ErrNotImplemented is defined
	if ErrNotImplemented == nil {
		t.Error("Expected ErrNotImplemented to be defined")
	}

	if ErrNotImplemented.Error() != "not implemented" {
		t.Errorf("Expected ErrNotImplemented to be 'not implemented', got '%s'", ErrNotImplemented.Error())
	}
}
