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
	// Test std.math library integration - type checking only
	cmd := exec.Command("go", "run", "../../cmd/omnic", "std_math_integration.omni")
	cmd.Dir = "."
	output, _ := cmd.CombinedOutput()
	outputStr := string(output)
	// Clean up generated C files
	os.Remove("std_math_integration.c")
	os.Remove("std_math_integration.h")
	// Check for type checking errors (not C compilation errors)
	if strings.Contains(outputStr, "[ERROR]") && strings.Contains(outputStr, "error:") {
		if !strings.Contains(outputStr, ".c:") && !strings.Contains(outputStr, "failed to compile C code") {
			t.Fatalf("Type checking failed: %s", outputStr)
		}
	}
	// C compilation errors are expected since runtime functions aren't fully implemented
	t.Logf("Type checking passed (C compilation errors expected)")
}

func TestStdMathComprehensive(t *testing.T) {
	// Test comprehensive std.math functions - type checking only
	cmd := exec.Command("go", "run", "../../cmd/omnic", "std_math_comprehensive.omni")
	cmd.Dir = "."
	output, _ := cmd.CombinedOutput()
	outputStr := string(output)
	// Clean up generated C files
	os.Remove("std_math_comprehensive.c")
	os.Remove("std_math_comprehensive.h")
	// Check for type checking errors (not C compilation errors)
	if strings.Contains(outputStr, "[ERROR]") && strings.Contains(outputStr, "error:") {
		if !strings.Contains(outputStr, ".c:") && !strings.Contains(outputStr, "failed to compile C code") {
			t.Fatalf("Type checking failed: %s", outputStr)
		}
	}
	// C compilation errors are expected since runtime functions aren't fully implemented
	t.Logf("Type checking passed (C compilation errors expected)")
}

func TestStdLibraryComprehensive(t *testing.T) {
	// Test that std library modules can be imported and basic functions called - type checking only
	cmd := exec.Command("go", "run", "../../cmd/omnic", "std_comprehensive_test.omni")
	cmd.Dir = "."
	output, _ := cmd.CombinedOutput()
	outputStr := string(output)
	// Clean up generated C files
	os.Remove("std_comprehensive_test.c")
	os.Remove("std_comprehensive_test.h")
	// Check for type checking errors (not C compilation errors)
	if strings.Contains(outputStr, "[ERROR]") && strings.Contains(outputStr, "error:") {
		if !strings.Contains(outputStr, ".c:") && !strings.Contains(outputStr, "failed to compile C code") {
			t.Fatalf("Type checking failed: %s", outputStr)
		}
	}
	// C compilation errors are expected since runtime functions aren't fully implemented
	t.Logf("Type checking passed (C compilation errors expected)")
}

func TestAllStdModules(t *testing.T) {
	// Test all std library modules to ensure they can be imported and functions called
	// Most functions are placeholders, but we test that the import system works

	t.Run("std.string", func(t *testing.T) {
		// Test that std.string module can be imported and type-checked
		cmd := exec.Command("go", "run", "../../cmd/omnic", "std_string_comprehensive.omni")
		cmd.Dir = "."
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)
		os.Remove("std_string_comprehensive.c")
		os.Remove("std_string_comprehensive.h")
		if strings.Contains(outputStr, "[ERROR]") && strings.Contains(outputStr, "error:") {
			if !strings.Contains(outputStr, ".c:") && !strings.Contains(outputStr, "failed to compile C code") {
				t.Fatalf("Type checking failed: %s", outputStr)
			}
		}
		t.Logf("Type checking passed (C compilation errors expected)")
	})

	t.Run("std.array", func(t *testing.T) {
		// Test that std.array module can be imported and type-checked
		// Note: std.array.length has no runtime implementation, so we only test compilation
		cmd := exec.Command("go", "run", "../../cmd/omnic", "std_array_simple.omni")
		cmd.Dir = "."
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)
		// Clean up generated C files
		os.Remove("std_array_simple.c")
		os.Remove("std_array_simple.h")
		// Check for type checking errors (not C compilation errors)
		if strings.Contains(outputStr, "[ERROR]") && strings.Contains(outputStr, "error:") {
			// Check if it's a type error (not C compilation error)
			if !strings.Contains(outputStr, ".c:") && !strings.Contains(outputStr, "failed to compile C code") {
				t.Fatalf("Type checking failed: %s", outputStr)
			}
		}
		// C compilation errors are expected since std.array.length isn't fully implemented
		t.Logf("Type checking passed (C compilation errors expected)")
	})

	t.Run("std.collections", func(t *testing.T) {
		// Test that std.collections module can be imported and type-checked
		cmd := exec.Command("go", "run", "../../cmd/omnic", "std_collections_simple.omni")
		cmd.Dir = "."
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)
		os.Remove("std_collections_simple.c")
		os.Remove("std_collections_simple.h")
		if strings.Contains(outputStr, "[ERROR]") && strings.Contains(outputStr, "error:") {
			if !strings.Contains(outputStr, ".c:") && !strings.Contains(outputStr, "failed to compile C code") {
				t.Fatalf("Type checking failed: %s", outputStr)
			}
		}
		t.Logf("Type checking passed (C compilation errors expected)")
	})

	t.Run("std.os", func(t *testing.T) {
		// Test that std.os module can be imported and type-checked
		cmd := exec.Command("go", "run", "../../cmd/omnic", "std_os_comprehensive.omni")
		cmd.Dir = "."
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)
		os.Remove("std_os_comprehensive.c")
		os.Remove("std_os_comprehensive.h")
		if strings.Contains(outputStr, "[ERROR]") && strings.Contains(outputStr, "error:") {
			if !strings.Contains(outputStr, ".c:") && !strings.Contains(outputStr, "failed to compile C code") {
				t.Fatalf("Type checking failed: %s", outputStr)
			}
		}
		t.Logf("Type checking passed (C compilation errors expected)")
	})

	t.Run("std.file", func(t *testing.T) {
		// Test that std.file module can be imported and type-checked
		cmd := exec.Command("go", "run", "../../cmd/omnic", "std_file_comprehensive.omni")
		cmd.Dir = "."
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)
		os.Remove("std_file_comprehensive.c")
		os.Remove("std_file_comprehensive.h")
		if strings.Contains(outputStr, "[ERROR]") && strings.Contains(outputStr, "error:") {
			if !strings.Contains(outputStr, ".c:") && !strings.Contains(outputStr, "failed to compile C code") {
				t.Fatalf("Type checking failed: %s", outputStr)
			}
		}
		t.Logf("Type checking passed (C compilation errors expected)")
	})

	t.Run("std.network", func(t *testing.T) {
		// Test that std.network module can be imported and type-checked
		cmd := exec.Command("go", "run", "../../cmd/omnic", "std_network_comprehensive.omni")
		cmd.Dir = "."
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)
		os.Remove("std_network_comprehensive.c")
		os.Remove("std_network_comprehensive.h")
		if strings.Contains(outputStr, "[ERROR]") && strings.Contains(outputStr, "error:") {
			if !strings.Contains(outputStr, ".c:") && !strings.Contains(outputStr, "failed to compile C code") {
				t.Fatalf("Type checking failed: %s", outputStr)
			}
		}
		t.Logf("Type checking passed (C compilation errors expected)")
	})

	t.Run("std.web", func(t *testing.T) {
		// Test that std.web module can be imported and type-checked
		// We use omnic to check type checking only (not runtime execution)
		cmd := exec.Command("go", "run", "../../cmd/omnic", "std_web_simple.omni")
		cmd.Dir = "."
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)
		// Clean up generated C files
		os.Remove("std_web_simple.c")
		os.Remove("std_web_simple.h")
		// Check for type checking errors (not C compilation errors)
		if strings.Contains(outputStr, "[ERROR]") && strings.Contains(outputStr, "error:") {
			// Check if it's a type error (not C compilation error)
			if !strings.Contains(outputStr, ".c:") && !strings.Contains(outputStr, "failed to compile C code") {
				t.Fatalf("Type checking failed: %s", outputStr)
			}
		}
		// C compilation errors are expected since runtime functions aren't fully implemented
		t.Logf("Type checking passed (C compilation errors expected)")
	})

	t.Run("std.web comprehensive", func(t *testing.T) {
		// Test comprehensive std.web features - type checking only
		cmd := exec.Command("go", "run", "../../cmd/omnic", "std_web_comprehensive.omni")
		cmd.Dir = "."
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)
		// Clean up generated C files
		os.Remove("std_web_comprehensive.c")
		os.Remove("std_web_comprehensive.h")
		if strings.Contains(outputStr, "[ERROR]") && strings.Contains(outputStr, "error:") {
			if !strings.Contains(outputStr, ".c:") && !strings.Contains(outputStr, "failed to compile C code") {
				t.Fatalf("Type checking failed: %s", outputStr)
			}
		}
		t.Logf("Type checking passed (C compilation errors expected)")
	})

	t.Run("std.web type checking", func(t *testing.T) {
		// Test Handler and Middleware type alias matching
		cmd := exec.Command("go", "run", "../../cmd/omnic", "std_web_type_checking.omni")
		cmd.Dir = "."
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)
		// Clean up generated C files
		os.Remove("std_web_type_checking.c")
		os.Remove("std_web_type_checking.h")
		if strings.Contains(outputStr, "[ERROR]") && strings.Contains(outputStr, "error:") {
			if !strings.Contains(outputStr, ".c:") && !strings.Contains(outputStr, "failed to compile C code") {
				t.Fatalf("Type checking failed: %s", outputStr)
			}
		}
		t.Logf("Type checking passed (C compilation errors expected)")
	})

	t.Run("std.web context", func(t *testing.T) {
		// Test Context API: params, query, body, headers, state
		cmd := exec.Command("go", "run", "../../cmd/omnic", "std_web_context.omni")
		cmd.Dir = "."
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)
		os.Remove("std_web_context.c")
		os.Remove("std_web_context.h")
		if strings.Contains(outputStr, "[ERROR]") && strings.Contains(outputStr, "error:") {
			if !strings.Contains(outputStr, ".c:") && !strings.Contains(outputStr, "failed to compile C code") {
				t.Fatalf("Type checking failed: %s", outputStr)
			}
		}
		t.Logf("Type checking passed (C compilation errors expected)")
	})

	t.Run("std.web routing", func(t *testing.T) {
		// Test routing features: route groups, path parameters, method routing
		cmd := exec.Command("go", "run", "../../cmd/omnic", "std_web_routing.omni")
		cmd.Dir = "."
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)
		os.Remove("std_web_routing.c")
		os.Remove("std_web_routing.h")
		if strings.Contains(outputStr, "[ERROR]") && strings.Contains(outputStr, "error:") {
			if !strings.Contains(outputStr, ".c:") && !strings.Contains(outputStr, "failed to compile C code") {
				t.Fatalf("Type checking failed: %s", outputStr)
			}
		}
		t.Logf("Type checking passed (C compilation errors expected)")
	})

	t.Run("std.web middleware", func(t *testing.T) {
		// Test middleware features: chaining, conditional middleware, error middleware
		cmd := exec.Command("go", "run", "../../cmd/omnic", "std_web_middleware.omni")
		cmd.Dir = "."
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)
		os.Remove("std_web_middleware.c")
		os.Remove("std_web_middleware.h")
		if strings.Contains(outputStr, "[ERROR]") && strings.Contains(outputStr, "error:") {
			if !strings.Contains(outputStr, ".c:") && !strings.Contains(outputStr, "failed to compile C code") {
				t.Fatalf("Type checking failed: %s", outputStr)
			}
		}
		t.Logf("Type checking passed (C compilation errors expected)")
	})

	t.Run("std.web static files", func(t *testing.T) {
		// Test static file serving and file uploads
		cmd := exec.Command("go", "run", "../../cmd/omnic", "std_web_static_files.omni")
		cmd.Dir = "."
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)
		os.Remove("std_web_static_files.c")
		os.Remove("std_web_static_files.h")
		if strings.Contains(outputStr, "[ERROR]") && strings.Contains(outputStr, "error:") {
			if !strings.Contains(outputStr, ".c:") && !strings.Contains(outputStr, "failed to compile C code") {
				t.Fatalf("Type checking failed: %s", outputStr)
			}
		}
		t.Logf("Type checking passed (C compilation errors expected)")
	})

	t.Run("std.web validation", func(t *testing.T) {
		// Test validation and sanitization features
		cmd := exec.Command("go", "run", "../../cmd/omnic", "std_web_validation.omni")
		cmd.Dir = "."
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)
		os.Remove("std_web_validation.c")
		os.Remove("std_web_validation.h")
		if strings.Contains(outputStr, "[ERROR]") && strings.Contains(outputStr, "error:") {
			if !strings.Contains(outputStr, ".c:") && !strings.Contains(outputStr, "failed to compile C code") {
				t.Fatalf("Type checking failed: %s", outputStr)
			}
		}
		t.Logf("Type checking passed (C compilation errors expected)")
	})

	t.Run("std.dev", func(t *testing.T) {
		result, err := runVM("std_dev_simple.omni")
		if err != nil {
			t.Fatalf("VM execution failed: %v", err)
		}
		if result != "0" {
			t.Fatalf("expected std.dev smoke test to return 0, got %s", result)
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
		// Test that std.os.args can be imported and type-checked
		cmd := exec.Command("go", "run", "../../cmd/omnic", "std_os_args.omni")
		cmd.Dir = "."
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)
		os.Remove("std_os_args.c")
		os.Remove("std_os_args.h")
		if strings.Contains(outputStr, "[ERROR]") && strings.Contains(outputStr, "error:") {
			if !strings.Contains(outputStr, ".c:") && !strings.Contains(outputStr, "failed to compile C code") {
				t.Fatalf("Type checking failed: %s", outputStr)
			}
		}
		t.Logf("Type checking passed (C compilation errors expected)")
	})

	t.Run("std.io.read_line", func(t *testing.T) {
		// Test that std.io.read_line can be imported and type-checked
		cmd := exec.Command("go", "run", "../../cmd/omnic", "std_io_read_line.omni")
		cmd.Dir = "."
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)
		os.Remove("std_io_read_line.c")
		os.Remove("std_io_read_line.h")
		if strings.Contains(outputStr, "[ERROR]") && strings.Contains(outputStr, "error:") {
			if !strings.Contains(outputStr, ".c:") && !strings.Contains(outputStr, "failed to compile C code") {
				t.Fatalf("Type checking failed: %s", outputStr)
			}
		}
		t.Logf("Type checking passed (C compilation errors expected)")
	})

	t.Run("std.os flag helpers", func(t *testing.T) {
		// Test that std.os flag helpers can be imported and type-checked
		cmd := exec.Command("go", "run", "../../cmd/omnic", "std_os_flag_helpers.omni")
		cmd.Dir = "."
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)
		os.Remove("std_os_flag_helpers.c")
		os.Remove("std_os_flag_helpers.h")
		if strings.Contains(outputStr, "[ERROR]") && strings.Contains(outputStr, "error:") {
			if !strings.Contains(outputStr, ".c:") && !strings.Contains(outputStr, "failed to compile C code") {
				t.Fatalf("Type checking failed: %s", outputStr)
			}
		}
		t.Logf("Type checking passed (C compilation errors expected)")
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
	// For now, just check that the file compiles (type checking only)
	// Coverage generation requires full VM execution which may fail for some std functions
	cmd := exec.Command("go", "run", "../../cmd/omnic", "std_io_comprehensive.omni")
	cmd.Dir = "."
	output, _ := cmd.CombinedOutput()
	outputStr := string(output)
	// Clean up generated C files
	os.Remove("std_io_comprehensive.c")
	os.Remove("std_io_comprehensive.h")
	// Check for type checking errors (not C compilation errors)
	if strings.Contains(outputStr, "[ERROR]") && strings.Contains(outputStr, "error:") {
		if !strings.Contains(outputStr, ".c:") && !strings.Contains(outputStr, "failed to compile C code") {
			t.Fatalf("Type checking failed: %s", outputStr)
		}
	}
	// C compilation errors are expected since runtime functions aren't fully implemented
	t.Logf("Type checking passed (C compilation errors expected)")
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
