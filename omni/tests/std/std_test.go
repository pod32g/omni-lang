package std

import (
	"os"
	"os/exec"
	"strings"
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
		result, err := runVM("std_os_comprehensive.omni")
		if err != nil {
			t.Fatalf("VM execution failed: %v", err)
		}
		expected := "0"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("std.file", func(t *testing.T) {
		result, err := runVM("std_file_comprehensive.omni")
		if err != nil {
			t.Fatalf("VM execution failed: %v", err)
		}
		expected := "0"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("std.network", func(t *testing.T) {
		result, err := runVM("std_network_comprehensive.omni")
		if err != nil {
			t.Fatalf("VM execution failed: %v", err)
		}
		expected := "0"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("std.testing", func(t *testing.T) {
		exitCode, err := runVMTestHarness("std_testing_suite.omni")
		if err != nil {
			t.Fatalf("VM execution failed: %v", err)
		}
		if exitCode != 0 {
			t.Fatalf("Expected exit code 0, got %d", exitCode)
		}
	})

	t.Run("std.testing summary reporting", func(t *testing.T) {
		exitCode, output, err := runVMTestHarnessWithOutput("std_testing_failures.omni")
		if err != nil {
			t.Fatalf("VM execution failed: %v", err)
		}
		if exitCode == 0 {
			t.Fatalf("expected non-zero exit code when tests fail, got %d", exitCode)
		}
		if !strings.Contains(output, "Test Summary: 5 total, 3 passed, 2 failed") {
			t.Fatalf("expected summary line in output, got:\n%s", output)
		}
		if !strings.Contains(output, "2 test(s) failed") {
			t.Fatalf("expected failure count log in output, got:\n%s", output)
		}
	})
}

func buildVMCommand(args ...string) *exec.Cmd {
	cmd := exec.Command("go", append([]string{"run", "../../cmd/omnir"}, args...)...)
	cmd.Dir = "."
	// Set library path for VM execution
	cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH=../../runtime/posix")
	cmd.Env = append(cmd.Env, "LD_LIBRARY_PATH=../../runtime/posix")
	return cmd
}

func runVM(testFile string, cliArgs ...string) (string, error) {
	args := append(cliArgs, testFile)
	cmd := buildVMCommand(args...)
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

func runVMTestHarness(testFile string) (int, error) {
	exitCode, _, err := runVMTestHarnessWithOutput(testFile)
	return exitCode, err
}

func runVMTestHarnessWithOutput(testFile string) (int, string, error) {
	cmd := buildVMCommand("--test", testFile)
	output, err := cmd.CombinedOutput()
	normalized := normalizeOutput(output)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), normalized, nil
		}
		return -1, normalized, err
	}
	exitCode := 0
	if ps := cmd.ProcessState; ps != nil {
		exitCode = ps.ExitCode()
	}
	return exitCode, normalized, nil
}

func normalizeOutput(output []byte) string {
	out := string(output)
	return strings.TrimRight(out, "\n")
}
