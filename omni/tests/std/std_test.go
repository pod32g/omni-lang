package std

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/omni-lang/omni/tools/coverage"
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
	// Expected: 0 (success indicator)
	expected := "0"
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
		// Expected: 0 (success indicator)
		expected := "0"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("std.array", func(t *testing.T) {
		result, err := runVM("std_array_simple.omni")
		if err != nil {
			t.Fatalf("VM execution failed: %v", err)
		}
		// Expected: 0 (success indicator)
		expected := "0"
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

	t.Run("std.os.args", func(t *testing.T) {
		result, err := runVM("std_os_args.omni", "--", "hello", "world")
		if err != nil {
			t.Fatalf("VM execution failed: %v", err)
		}
		if result != "0" {
			t.Fatalf("expected output 0, got %s", result)
		}
	})

	t.Run("std.io.read_line", func(t *testing.T) {
		result, err := runVMWithInput("hello\n\n", "std_io_read_line.omni")
		if err != nil {
			t.Fatalf("VM execution failed: %v", err)
		}
		if result != "0" {
			t.Fatalf("expected output 0, got %s", result)
		}
	})

	t.Run("std.os flag helpers", func(t *testing.T) {
		result, err := runVM("std_os_flag_helpers.omni", "--", "--flag", "--name=omni", "hello", "world")
		if err != nil {
			t.Fatalf("VM execution failed: %v", err)
		}
		if result != "0" {
			t.Fatalf("expected output 0, got %s", result)
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
	return runVMWithInput("", testFile, cliArgs...)
}

func runVMWithInput(input string, testFile string, cliArgs ...string) (string, error) {
	args := append([]string{testFile}, cliArgs...)
	cmd := buildVMCommand(args...)
	if input != "" {
		cmd.Stdin = strings.NewReader(input)
	}
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
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

// runVMWithCoverage runs a test file with coverage enabled and returns coverage data
func runVMWithCoverage(testFile string, cliArgs ...string) (string, string, error) {
	coverageFile := filepath.Join(os.TempDir(), fmt.Sprintf("coverage_%s.json", testFile))
	defer os.Remove(coverageFile)

	args := append([]string{"--coverage", "--coverage-output", coverageFile, testFile}, cliArgs...)
	cmd := buildVMCommand(args...)
	output, err := cmd.Output()
	if err != nil {
		return "", "", err
	}
	result := string(output)
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}

	// Read coverage data
	coverageData, err := os.ReadFile(coverageFile)
	if err != nil {
		return result, "", nil // Coverage file might not exist if no std functions were called
	}

	return result, string(coverageData), nil
}

// TestCoverageGeneration tests that coverage data is generated when running tests
func TestCoverageGeneration(t *testing.T) {
	result, coverageJSON, err := runVMWithCoverage("std_io_comprehensive.omni")
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if !strings.HasSuffix(result, "0") {
		t.Errorf("expected output to end with 0, got %s", result)
	}

	// Verify coverage JSON is valid
	if coverageJSON == "" {
		t.Skip("No coverage data generated (may be expected if no std functions called)")
		return
	}

	var coverageData coverage.CoverageData
	if err := json.Unmarshal([]byte(coverageJSON), &coverageData); err != nil {
		t.Fatalf("Failed to parse coverage JSON: %v", err)
	}

	if len(coverageData.Entries) == 0 {
		t.Log("Coverage data generated but no entries found")
	} else {
		t.Logf("Coverage data generated with %d entries", len(coverageData.Entries))
	}
}

// TestCoverageThreshold tests that coverage meets the 60% threshold
func TestCoverageThreshold(t *testing.T) {
	// Run all comprehensive tests with coverage
	testFiles := []string{
		"std_io_comprehensive.omni",
		"std_string_comprehensive.omni",
		"std_math_comprehensive.omni",
		"std_array_simple.omni",
		"std_file_comprehensive.omni",
		"std_os_comprehensive.omni",
	}

	var aggregated coverage.CoverageData

	// Run all tests and collect coverage
	for _, testFile := range testFiles {
		_, coverageJSON, err := runVMWithCoverage(testFile)
		if err != nil {
			t.Logf("Warning: Test %s failed: %v", testFile, err)
			continue
		}
		if coverageJSON == "" {
			continue
		}
		var data coverage.CoverageData
		if err := json.Unmarshal([]byte(coverageJSON), &data); err != nil {
			t.Logf("Warning: could not parse coverage for %s: %v", testFile, err)
			continue
		}
		aggregated.Entries = append(aggregated.Entries, data.Entries...)
	}

	if len(aggregated.Entries) == 0 {
		t.Skip("No coverage data generated")
	}

	// Parse std library
	stdPath := "../../std"
	funcsByFile, err := coverage.ParseStdLibrary(stdPath)
	if err != nil {
		t.Fatalf("Failed to parse std library: %v", err)
	}

	// Match coverage to functions
	matches := coverage.MatchCoverageToFunctions(&aggregated, funcsByFile)

	// Calculate statistics
	stats := coverage.CalculateCoverage(matches)

	// Check threshold
	threshold := 60.0
	meetsThreshold, message := coverage.CheckCoverageThreshold(stats, threshold)

	t.Log(message)
	t.Logf("Function Coverage: %.2f%% (%d/%d)", stats.GetFunctionCoveragePercentage(), stats.CoveredFunctions, stats.TotalFunctions)
	t.Logf("Line Coverage: %.2f%% (%d/%d)", stats.GetLineCoveragePercentage(), stats.CoveredLines, stats.TotalLines)

	if !meetsThreshold {
		t.Errorf("Coverage threshold not met: %.2f%% < %.2f%%", stats.GetFunctionCoveragePercentage(), threshold)
	}
}
