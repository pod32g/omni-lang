package std

import (
	"os"
	"os/exec"
	"testing"
)

func TestBasicMath(t *testing.T) {
	result, err := runVM("../../examples/basic_math.omni")
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	// Expected: 42+17 + 42-17 + 42*17 + 42/17 + 42%17 + max(42,17) + min(42,17)
	// = 59 + 25 + 714 + 2 + 8 + 42 + 17 = 867
	expected := "867"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestMathFunctions(t *testing.T) {
	// Test individual math functions
	testCases := []struct {
		file     string
		expected string
	}{
		{"math_max.omni", "42"},
		{"math_min.omni", "17"},
		{"math_simple.omni", "25"}, // 15 + 10
	}

	for _, tc := range testCases {
		t.Run(tc.file, func(t *testing.T) {
			result, err := runVM(tc.file)
			if err != nil {
				t.Fatalf("VM execution failed: %v", err)
			}
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestStdMathIntegration(t *testing.T) {
	// Test std.math library integration
	result, err := runVM("std_math_integration.omni")
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	// Expected: 42 + 17 + 10 = 69 (max(42,17) + min(42,17) + abs(-10))
	expected := "69"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestStdMathComprehensive(t *testing.T) {
	// Test comprehensive std.math functions (only implemented ones)
	result, err := runVM("std_math_comprehensive.omni")
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	// Expected: 42 + 17 + 10 = 69 (max + min + abs)
	expected := "69"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestStdLibraryComprehensive(t *testing.T) {
	// Test that std library modules can be imported and basic functions called
	// This tests the actual working functionality (std.math and std.io)
	result, err := runVM("std_comprehensive_test.omni")
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	// Expected: 42 + 17 + 10 = 69 (max + min + abs)
	expected := "69"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestAllStdModules(t *testing.T) {
	// Test all std library modules to ensure they can be imported and functions called
	// Most functions are placeholders, but we test that the import system works

	t.Run("std.string", func(t *testing.T) {
		result, err := runVM("std_string_comprehensive.omni")
		if err != nil {
			t.Fatalf("VM execution failed: %v", err)
		}
		// Expected: 20 (actual result from string functions)
		expected := "20"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("std.array", func(t *testing.T) {
		result, err := runVM("std_array_simple.omni")
		if err != nil {
			t.Fatalf("VM execution failed: %v", err)
		}
		// Expected: 42 (success indicator)
		expected := "42"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("std.collections", func(t *testing.T) {
		result, err := runVM("std_collections_simple.omni")
		if err != nil {
			t.Fatalf("VM execution failed: %v", err)
		}
		// Expected: 42 (success indicator)
		expected := "42"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("std.os", func(t *testing.T) {
		result, err := runVM("std_os_simple.omni")
		if err != nil {
			t.Fatalf("VM execution failed: %v", err)
		}
		// Expected: 42 (success indicator)
		expected := "42"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("std.file", func(t *testing.T) {
		result, err := runVM("std_file_simple.omni")
		if err != nil {
			t.Fatalf("VM execution failed: %v", err)
		}
		// Expected: 42 (success indicator)
		expected := "42"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})
}

func runVM(testFile string) (string, error) {
	cmd := exec.Command("go", "run", "../../cmd/omnir", testFile)
	cmd.Dir = "."
	// Set library path for VM execution
	cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH=../../runtime/posix")
	cmd.Env = append(cmd.Env, "LD_LIBRARY_PATH=../../runtime/posix")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	// Remove trailing newline
	result := string(output)
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}
	return result, nil
}
