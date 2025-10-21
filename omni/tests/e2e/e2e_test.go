package e2e

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

func TestHelloWorld(t *testing.T) {
	testFile := "hello_world.omni"
	expected := "42"

	// Test VM execution
	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// Test MIR generation
	if err := runCompiler(testFile, "vm", "mir", ""); err != nil {
		t.Errorf("MIR generation failed: %v", err)
	}

	// Test object file generation (skip on macOS due to Cranelift build constraints)
	if runtime.GOOS != "darwin" {
		if err := runCompiler(testFile, "clift", "obj", ""); err != nil {
			t.Errorf("Object generation failed: %v", err)
		}
	}
}

func TestArithmetic(t *testing.T) {
	testFile := "arithmetic.omni"
	expected := "72" // 15 + 5 + 50 + 2 = 72: 10+5=15, 10-5=5, 10*5=50, 10/5=2, 15+5+50+2=72

	// Test VM execution
	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestConditional(t *testing.T) {
	testFile := "conditional.omni"
	expected := "1"

	// Test VM execution
	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestLoop(t *testing.T) {
	testFile := "loop.omni"
	expected := "15" // 5 * 3

	// Test VM execution
	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestForClassic(t *testing.T) {
	testFile := "for_classic.omni"
	expected := "10" // 0 + 1 + 2 + 3 + 4 = 10

	// Test C backend execution
	result, err := runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// Test MIR generation
	if err := runCompiler(testFile, "vm", "mir", ""); err != nil {
		t.Errorf("MIR generation failed: %v", err)
	}
}

func TestForRange(t *testing.T) {
	testFile := "for_range.omni"
	expected := "15" // 1 + 2 + 3 + 4 + 5 = 15

	// Test C backend execution
	result, err := runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// Test MIR generation
	if err := runCompiler(testFile, "vm", "mir", ""); err != nil {
		t.Errorf("MIR generation failed: %v", err)
	}
}

func TestForNested(t *testing.T) {
	testFile := "for_nested.omni"
	expected := "9" // (0+0) + (0+1) + (1+0) + (1+1) + (2+0) + (2+1) = 0+1+1+2+2+3 = 9

	// Test C backend execution
	result, err := runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestForEmpty(t *testing.T) {
	testFile := "for_empty.omni"
	expected := "0" // Empty loop should return 0

	// Test C backend execution
	result, err := runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func runVM(testFile string) (string, error) {
	// Get the directory where the test files are located
	// Use the directory containing the test file
	testDir := "."

	cmd := exec.Command("../../bin/omnir", testFile)
	cmd.Dir = testDir
	// Set library path for VM execution (DYLD_LIBRARY_PATH on macOS, LD_LIBRARY_PATH on Linux)
	cmd.Env = append(cmd.Env, "DYLD_LIBRARY_PATH=../../runtime/posix")
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

func runCBackend(testFile string) (string, error) {
	// Get the directory where the test files are located
	// Use the directory containing the test file
	testDir := "."

	// Compile with C backend using the built binary
	compileCmd := exec.Command("../../bin/omnic", "-backend", "c", "-emit", "exe", testFile)
	compileCmd.Dir = testDir
	// Set environment variables for compilation
	compileCmd.Env = append(compileCmd.Env, "DYLD_LIBRARY_PATH=../../runtime/posix")
	compileCmd.Env = append(compileCmd.Env, "LD_LIBRARY_PATH=../../runtime/posix")
	compileCmd.Env = append(compileCmd.Env, "PATH=/usr/bin:/bin:/usr/sbin:/sbin")
	compileOutput, err := compileCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("compilation failed: %v\nOutput: %s", err, string(compileOutput))
	}

	// Run the compiled executable (it's created in the same directory as the source file)
	executableName := testFile[:len(testFile)-5] // Remove .omni extension
	runCmd := exec.Command("./" + executableName)
	runCmd.Dir = testDir

	// Set environment variables to find the runtime library
	runCmd.Env = append(runCmd.Env, "DYLD_LIBRARY_PATH=../../runtime/posix")
	runCmd.Env = append(runCmd.Env, "LD_LIBRARY_PATH=../../runtime/posix")
	runCmd.Env = append(runCmd.Env, "PATH=/usr/bin:/bin:/usr/sbin:/sbin")

	output, err := runCmd.Output()
	if err != nil {
		// Check if it's an exit error with a non-zero code
		if exitErr, ok := err.(*exec.ExitError); ok {
			// For OmniLang programs, the exit code is the return value
			// So we should treat this as success and extract the exit code
			return fmt.Sprintf("%d", exitErr.ExitCode()), nil
		}
		return "", fmt.Errorf("execution failed: %v", err)
	}

	// Parse the output to extract the result
	result := string(output)
	// Look for "OmniLang program result: X" pattern
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "OmniLang program result: ") {
			return strings.TrimSpace(line[len("OmniLang program result: "):]), nil
		}
	}
	return "", fmt.Errorf("could not find program result in output: %s", result)
}

func runCompiler(testFile, backend, emit, output string) error {
	args := []string{"-backend", backend, "-emit", emit, testFile}
	if output != "" {
		args = append(args, "-o", output)
	}
	cmd := exec.Command("../../bin/omnic", args...)
	cmd.Dir = "."
	return cmd.Run()
}
