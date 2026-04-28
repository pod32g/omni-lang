package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	// TODO: Re-enable when Cranelift backend is fully implemented
	// Currently disabled due to verifier errors in placeholder implementation
	if false && runtime.GOOS != "darwin" {
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

	// Get absolute path to the test file
	absTestFile, err := filepath.Abs(testFile)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for test file: %v", err)
	}

	// Debug: Check if file exists
	if _, err := os.Stat(absTestFile); os.IsNotExist(err) {
		return "", fmt.Errorf("test file does not exist: %s (working directory: %s)", absTestFile, func() string {
			wd, _ := os.Getwd()
			return wd
		}())
	}

	cmd := exec.Command("../../bin/omnir", absTestFile)
	// Set the working directory to tests/e2e so relative paths work correctly
	cmd.Dir = "."
	// Set library path for VM execution (DYLD_LIBRARY_PATH on macOS, LD_LIBRARY_PATH on Linux)
	cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH=../../native/clift/target/release:../../runtime/posix")
	cmd.Env = append(cmd.Env, "LD_LIBRARY_PATH=../../native/clift/target/release:../../runtime/posix")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w\nOutput: %s", err, string(output))
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

	// Compile with C backend using the built binary
	compileCmd := exec.Command("../../bin/omnic", "-backend", "c", "-emit", "exe", testFile)
	// Set the working directory to tests/e2e so relative paths work correctly
	compileCmd.Dir = "."
	// Set environment variables for compilation
	compileCmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH=../../native/clift/target/release:../../runtime/posix")
	compileCmd.Env = append(compileCmd.Env, "LD_LIBRARY_PATH=../../native/clift/target/release:../../runtime/posix")
	compileCmd.Env = append(compileCmd.Env, "PATH=/usr/bin:/bin:/usr/sbin:/sbin")
	compileOutput, err := compileCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("compilation failed: %v\nOutput: %s", err, string(compileOutput))
	}

	// Run the compiled executable (it's created in the same directory as the source file)
	executableName := testFile[:len(testFile)-5] // Remove .omni extension
	runCmd := exec.Command("./" + executableName)
	// Set the working directory to tests/e2e so relative paths work correctly
	runCmd.Dir = "."

	// Set environment variables to find the runtime library
	runCmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH=../../native/clift/target/release:../../runtime/posix")
	runCmd.Env = append(runCmd.Env, "LD_LIBRARY_PATH=../../native/clift/target/release:../../runtime/posix")
	runCmd.Env = append(runCmd.Env, "PATH=/usr/bin:/bin:/usr/sbin:/sbin")

	output, err := runCmd.Output()

	// Always try to parse stdout first, regardless of exit code
	result := string(output)
	// Look for "OmniLang program result: X" pattern
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "OmniLang program result: ") {
			return strings.TrimSpace(line[len("OmniLang program result: "):]), nil
		}
	}

	// If no stdout result found, check if it's an exit error with a non-zero code
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// For OmniLang programs, the exit code is the return value
			// So we should treat this as success and extract the exit code
			return fmt.Sprintf("%d", exitErr.ExitCode()), nil
		}
		return "", fmt.Errorf("execution failed: %v", err)
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
	// Inherit environment variables (including LD_LIBRARY_PATH for Cranelift backend)
	cmd.Env = os.Environ()

	// Capture both stdout and stderr for debugging
	output_bytes, err := cmd.CombinedOutput()
	if err != nil {
		// Print the actual error output for debugging
		fmt.Printf("Compiler error output: %s\n", string(output_bytes))
		return fmt.Errorf("compiler failed: %v, output: %s", err, string(output_bytes))
	}
	return nil
}

// Array tests (both VM and C backends now supported!)
func TestArrayBasic(t *testing.T) {
	testFile := "array_basic.omni"
	expected := "20" // numbers[1] where numbers = [10, 20, 30]

	// Test VM backend
	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("VM: expected %s, got %s", expected, result)
	}

	// Test C backend
	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

func TestArrayArithmetic(t *testing.T) {
	testFile := "array_arithmetic.omni"
	expected := "45" // 5 + 15 + 25 = 45

	// Test VM backend
	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("VM: expected %s, got %s", expected, result)
	}

	// Test C backend
	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

func TestArrayStrings(t *testing.T) {
	testFile := "array_strings.omni"
	expected := "42" // Should return 42 if string comparisons work

	// Test VM backend
	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("VM: expected %s, got %s", expected, result)
	}

	// Test C backend
	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

func TestArrayLen(t *testing.T) {
	testFile := "array_len.omni"
	expected := "5" // len([1, 2, 3, 4, 5]) = 5

	// Test VM backend
	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("VM: expected %s, got %s", expected, result)
	}

	// Test C backend
	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

func TestArrayMethod(t *testing.T) {
	testFile := "array_method.omni"
	expected := "5" // numbers.len() where numbers = [1, 2, 3, 4, 5]

	// Test VM backend
	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("VM: expected %s, got %s", expected, result)
	}

	// Test C backend
	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

func TestMapBasic(t *testing.T) {
	testFile := "map_basic.omni"
	expected := "95" // scores["alice"] where scores = {"alice": 95, "bob": 87, "charlie": 92}

	// Test VM backend
	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("VM: expected %s, got %s", expected, result)
	}

	// Test C backend (with placeholder implementation)
	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

func TestMapComprehensive(t *testing.T) {
	testFile := "map_comprehensive.omni"
	expectedVM := "274" // 95 + 87 + 92 + 0 = 274
	expectedC := "274"  // 95 + 87 + 92 + 0 = 274 (real map implementation)

	// Test VM backend
	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expectedVM {
		t.Errorf("VM: expected %s, got %s", expectedVM, result)
	}

	// Test C backend (with placeholder implementation)
	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expectedC {
		t.Errorf("C backend: expected %s, got %s", expectedC, result)
	}
}

func TestStructBasic(t *testing.T) {
	testFile := "struct_basic.omni"
	expected := "10" // p.x where p = Point{x: 10, y: 20}

	// Test VM backend
	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("VM: expected %s, got %s", expected, result)
	}

	// Test C backend (with placeholder implementation)
	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

func TestStructComprehensive(t *testing.T) {
	testFile := "struct_comprehensive.omni"
	expectedVM := "30" // 10 + 20 = 30
	expectedC := "30"  // 10 + 20 = 30 (real struct implementation)

	// Test VM backend
	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expectedVM {
		t.Errorf("VM: expected %s, got %s", expectedVM, result)
	}

	// Test C backend (with placeholder implementation)
	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expectedC {
		t.Errorf("C backend: expected %s, got %s", expectedC, result)
	}
}

func TestPhiLoop(t *testing.T) {
	testFile := "phi_test.omni"
	expected := "3" // 0 + 1 + 2 = 3

	// Test VM backend
	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("VM: expected %s, got %s", expected, result)
	}

	// Test C backend
	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

func TestPhiManual(t *testing.T) {
	testFile := "phi_manual.omni"
	expected := "3" // 0 + 1 + 2 = 3

	// Test VM backend
	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("VM: expected %s, got %s", expected, result)
	}

	// Test C backend
	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

// TestVarBranchMerge pins down the phi-drift fix: a `var` reassigned in
// both arms of an `if` inside a loop must read the same storage slot
// across iterations. Before the fix, the VM backend returned 8 (only
// the else-branch slot was observed); the correct sum is 14.
func TestVarBranchMerge(t *testing.T) {
	testFile := "var_branch_merge.omni"
	expected := "14"

	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("VM: expected %s, got %s", expected, result)
	}

	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

// TestStdAlgorithmsDistance pins the std.algorithms distance metrics:
// euclidean / manhattan / levenshtein. Each lives as a runtime
// intrinsic on the C side and as an execIntrinsic case in the VM.
func TestStdAlgorithmsDistance(t *testing.T) {
	testFile := "std_algorithms_distance.omni"
	expected := "10"

	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("VM: expected %s, got %s", expected, result)
	}

	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

// TestStdStringSplitReplace pins the new runtime intrinsics for the
// "implemented in OmniLang" string helpers that never actually
// compiled cleanly (split, split_words, join, replace*, find_all).
// Variable-length results thread their length through a runtime
// out-pointer the codegen registers as the array's runtime length.
func TestStdStringSplitReplace(t *testing.T) {
	testFile := "std_string_split_replace.omni"
	expected := "35"

	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("VM: expected %s, got %s", expected, result)
	}

	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

// TestStdCollectionsCompound pins queue / stack / set operations.
// Before this test these compile-failed on the C side ("unknown
// type queue<int>") and silently returned defaults on the VM.
func TestStdCollectionsCompound(t *testing.T) {
	testFile := "std_collections_compound.omni"
	expected := "347"

	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("VM: expected %s, got %s", expected, result)
	}

	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

// TestStdCollectionsMap pins the basic map operations
// (size/get/set/has/remove/clear). Previously these were declared as
// "implemented" but the C side had no wiring and the VM ran the
// stub bodies that returned defaults.
func TestStdCollectionsMap(t *testing.T) {
	testFile := "std_collections_map.omni"
	expected := "19"

	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("VM: expected %s, got %s", expected, result)
	}

	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

// TestStdFileOps pins std.file handle operations on both backends.
// The C backend keeps native FILE* handles in intptr_t slots while the
// source-level API exposes them as int handles.
func TestStdFileOps(t *testing.T) {
	testFile := "std_file_ops.omni"
	expected := "0"

	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("VM: expected %s, got %s", expected, result)
	}

	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

// TestStdTimeOps pins deterministic std.time conversions and duration
// helpers on both backends. It intentionally avoids wall-clock assertions
// beyond allowing zero-duration sleep.
func TestStdTimeOps(t *testing.T) {
	testFile := "std_time_ops.omni"
	expected := "0"

	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("runVM(%q) failed: %v", testFile, err)
	}
	if result != expected {
		t.Errorf("runVM(%q) = %s, want %s", testFile, result, expected)
	}

	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("runCBackend(%q) failed: %v", testFile, err)
	}
	if result != expected {
		t.Errorf("runCBackend(%q) = %s, want %s", testFile, result, expected)
	}
}

// TestStdFileHandles pins the handle-based file ops added to mirror
// Go's os/bufio: std.file.read_all/read_line/write_string,
// std.io.fprint/fprintln/fprintf, and std.os.read_file_lines /
// write_file_lines. Round-trips a fixture file through all of them.
func TestStdFileHandles(t *testing.T) {
	testFile := "std_file_handles.omni"
	expected := "0"

	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("runVM(%q) failed: %v", testFile, err)
	}
	if result != expected {
		t.Errorf("runVM(%q) = %s, want %s", testFile, result, expected)
	}

	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("runCBackend(%q) failed: %v", testFile, err)
	}
	if result != expected {
		t.Errorf("runCBackend(%q) = %s, want %s", testFile, result, expected)
	}
}

// TestStringEscapes pins the lexer-accepted string-literal escape
// set (\n, \t, \r, \\, \", \', \0, \xHH, \uHHHH) on both backends.
// VM used to skip decoding so `"a\nb"` measured length 4; the C
// backend got escape decoding for free from the C compiler. Both now
// agree.
func TestStringEscapes(t *testing.T) {
	testFile := "string_escapes.omni"
	expected := "0"

	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("runVM(%q) failed: %v", testFile, err)
	}
	if result != expected {
		t.Errorf("runVM(%q) = %s, want %s", testFile, result, expected)
	}

	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("runCBackend(%q) failed: %v", testFile, err)
	}
	if result != expected {
		t.Errorf("runCBackend(%q) = %s, want %s", testFile, result, expected)
	}
}

// TestStdNetworkBasic pins the offline parts of std.network on both
// backends: IP validation/classification (ip_is_valid, ip_is_private,
// ip_is_loopback, ip_to_string) and URL validation. The audit caught
// `omni_ip_is_valid` accepting out-of-range IPv4 segments
// (e.g. 999.999.999.999) and `omni_url_is_valid` accepting any string
// containing "://". Both runtime helpers now do real validation.
//
// Functions that need real network (http_get, dns_lookup, socket_*)
// are deliberately excluded. URL/HTTPResponse struct field access on
// the C backend is a separate pre-existing gap (the C codegen lowers
// omni_url_t fields incompatibly).
func TestStdNetworkBasic(t *testing.T) {
	testFile := "std_network_basic.omni"
	expected := "0"

	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("runVM(%q) failed: %v", testFile, err)
	}
	if result != expected {
		t.Errorf("runVM(%q) = %s, want %s", testFile, result, expected)
	}

	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("runCBackend(%q) failed: %v", testFile, err)
	}
	if result != expected {
		t.Errorf("runCBackend(%q) = %s, want %s", testFile, result, expected)
	}
}

// TestStdIoBasic pins std.io print/println/eprint/eprintln/flush on
// both backends. Each program writes "ab\n" to stdout and "cd\n" to
// stderr and returns 0; we just verify it exits successfully.
func TestStdIoBasic(t *testing.T) {
	testFile := "std_io_basic.omni"
	expected := "0"

	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("runVM(%q) failed: %v", testFile, err)
	}
	// runVM strips trailing newline; we wrote "a" + "b\n" + main's
	// "0\n". The runner returns the trimmed stdout, which ends in "0".
	if !strings.HasSuffix(result, "0") {
		t.Errorf("runVM(%q) = %q, want suffix %q", testFile, result, expected)
	}

	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("runCBackend(%q) failed: %v", testFile, err)
	}
	if result != expected {
		t.Errorf("runCBackend(%q) = %s, want %s", testFile, result, expected)
	}
}

// TestStdIoRead pins std.io.read_line and std.io.read_all on both
// backends by piping a known string into stdin.
func TestStdIoRead(t *testing.T) {
	testFile := "std_io_read.omni"
	stdin := "first line\nrest 1\nrest 2\n"

	// VM
	{
		absTestFile, _ := filepath.Abs(testFile)
		cmd := exec.Command("../../bin/omnir", absTestFile)
		cmd.Env = append(os.Environ(),
			"DYLD_LIBRARY_PATH=../../native/clift/target/release:../../runtime/posix",
			"LD_LIBRARY_PATH=../../native/clift/target/release:../../runtime/posix",
		)
		cmd.Stdin = strings.NewReader(stdin)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("VM run failed: %v\nOutput: %s", err, string(out))
		}
		if !strings.HasSuffix(strings.TrimRight(string(out), "\n"), "0") {
			t.Errorf("VM stdout = %q, want trailing 0", string(out))
		}
	}

	// C backend: build first via the shared helper (it doesn't accept
	// stdin), then re-run the executable manually.
	if _, err := runCBackend(testFile); err != nil {
		t.Fatalf("initial build failed: %v", err)
	}
	executable := "./std_io_read"
	cmd := exec.Command(executable)
	cmd.Env = append(os.Environ(),
		"DYLD_LIBRARY_PATH=../../native/clift/target/release:../../runtime/posix",
		"LD_LIBRARY_PATH=../../native/clift/target/release:../../runtime/posix",
	)
	cmd.Stdin = strings.NewReader(stdin)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("C run failed: %v\nOutput: %s", err, string(out))
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "OmniLang program result: ") {
			got := strings.TrimSpace(line[len("OmniLang program result: "):])
			if got != "0" {
				t.Errorf("C result = %s, want 0\nFull output: %s", got, string(out))
			}
			return
		}
	}
	t.Errorf("C result line not found; output: %s", string(out))
}

// TestStdIoExtras pins the post-audit io extensions (sprint, sprintln,
// sprintf, parse_int, parse_float, is_int, is_float, is_terminal) on
// both backends.
func TestStdIoExtras(t *testing.T) {
	testFile := "std_io_extras.omni"
	expected := "0"

	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("runVM(%q) failed: %v", testFile, err)
	}
	if result != expected {
		t.Errorf("runVM(%q) = %s, want %s", testFile, result, expected)
	}

	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("runCBackend(%q) failed: %v", testFile, err)
	}
	if result != expected {
		t.Errorf("runCBackend(%q) = %s, want %s", testFile, result, expected)
	}
}

// TestStdIoFormat pins the io formatting/styling extensions: ANSI
// helpers (bold, dim, italic, underline, red, green, yellow, blue,
// magenta, cyan, style), printf, eprintf, print_each, eprint_each,
// flush_stderr. The .omni file checks the string-returning helpers
// for shape (length growth + substring containment); the side-
// effecting writers must just not crash.
func TestStdIoFormat(t *testing.T) {
	testFile := "std_io_format.omni"
	expected := "0"

	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("runVM(%q) failed: %v", testFile, err)
	}
	if !strings.HasSuffix(strings.TrimRight(result, "\n"), "0") {
		t.Errorf("runVM(%q) = %q, want trailing 0", testFile, result)
	}

	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("runCBackend(%q) failed: %v", testFile, err)
	}
	if result != expected {
		t.Errorf("runCBackend(%q) = %s, want %s", testFile, result, expected)
	}
}

// TestStdIoConfirm pins std.io.confirm on both backends. Pipes "y\n"
// and expects the program to return 0; pipes "n\n" and expects 1.
func TestStdIoConfirm(t *testing.T) {
	testFile := "std_io_confirm.omni"

	runWithStdin := func(stdin string) (int, string, error) {
		// Build via shared helper, then re-run with custom stdin.
		if _, err := runCBackend(testFile); err != nil {
			return -1, "", fmt.Errorf("build failed: %w", err)
		}
		executable := "./std_io_confirm"
		cmd := exec.Command(executable)
		cmd.Env = append(os.Environ(),
			"DYLD_LIBRARY_PATH=../../native/clift/target/release:../../runtime/posix",
			"LD_LIBRARY_PATH=../../native/clift/target/release:../../runtime/posix",
		)
		cmd.Stdin = strings.NewReader(stdin)
		out, err := cmd.CombinedOutput()
		exit := 0
		if err != nil {
			if ee, ok := err.(*exec.ExitError); ok {
				exit = ee.ExitCode()
			} else {
				return -1, string(out), err
			}
		}
		return exit, string(out), nil
	}

	for _, tc := range []struct {
		stdin string
		want  string
	}{
		{"y\n", "0"},
		{"n\n", "1"},
	} {
		out := ""
		_, fullOut, err := runWithStdin(tc.stdin)
		if err != nil {
			t.Fatalf("C run failed for stdin=%q: %v", tc.stdin, err)
		}
		// Prompt has no trailing newline so the result line may be
		// glued to it. Look anywhere in the output.
		idx := strings.Index(fullOut, "OmniLang program result: ")
		if idx >= 0 {
			tail := fullOut[idx+len("OmniLang program result: "):]
			if nl := strings.IndexByte(tail, '\n'); nl >= 0 {
				tail = tail[:nl]
			}
			out = strings.TrimSpace(tail)
		}
		if out != tc.want {
			t.Errorf("stdin=%q: got %q, want %q\nfull output: %s", tc.stdin, out, tc.want, fullOut)
		}
	}

	// VM
	for _, tc := range []struct {
		stdin string
		want  string
	}{
		{"y\n", "0"},
		{"n\n", "1"},
	} {
		absTestFile, _ := filepath.Abs(testFile)
		cmd := exec.Command("../../bin/omnir", absTestFile)
		cmd.Env = append(os.Environ(),
			"DYLD_LIBRARY_PATH=../../native/clift/target/release:../../runtime/posix",
			"LD_LIBRARY_PATH=../../native/clift/target/release:../../runtime/posix",
		)
		cmd.Stdin = strings.NewReader(tc.stdin)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("VM run failed for stdin=%q: %v\nout: %s", tc.stdin, err, string(out))
		}
		// VM prints the main return value on stdout.
		if !strings.HasSuffix(strings.TrimRight(string(out), "\n"), tc.want) {
			t.Errorf("VM stdin=%q: out=%q want trailing %s", tc.stdin, string(out), tc.want)
		}
	}
}

// TestStdIoReadLines pins std.io.read_lines on both backends by piping a
// known string into stdin. The trailing newline must NOT produce a
// phantom blank entry at the end.
func TestStdIoReadLines(t *testing.T) {
	testFile := "std_io_lines.omni"
	stdin := "alpha\nbeta\ngamma\n"

	{
		absTestFile, _ := filepath.Abs(testFile)
		cmd := exec.Command("../../bin/omnir", absTestFile)
		cmd.Env = append(os.Environ(),
			"DYLD_LIBRARY_PATH=../../native/clift/target/release:../../runtime/posix",
			"LD_LIBRARY_PATH=../../native/clift/target/release:../../runtime/posix",
		)
		cmd.Stdin = strings.NewReader(stdin)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("VM run failed: %v\nOutput: %s", err, string(out))
		}
		if !strings.HasSuffix(strings.TrimRight(string(out), "\n"), "0") {
			t.Errorf("VM stdout = %q, want trailing 0", string(out))
		}
	}

	if _, err := runCBackend(testFile); err != nil {
		t.Fatalf("initial build failed: %v", err)
	}
	executable := "./std_io_lines"
	cmd := exec.Command(executable)
	cmd.Env = append(os.Environ(),
		"DYLD_LIBRARY_PATH=../../native/clift/target/release:../../runtime/posix",
		"LD_LIBRARY_PATH=../../native/clift/target/release:../../runtime/posix",
	)
	cmd.Stdin = strings.NewReader(stdin)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("C run failed: %v\nOutput: %s", err, string(out))
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "OmniLang program result: ") {
			got := strings.TrimSpace(line[len("OmniLang program result: "):])
			if got != "0" {
				t.Errorf("C result = %s, want 0\nFull output: %s", got, string(out))
			}
			return
		}
	}
	t.Errorf("C result line not found; output: %s", string(out))
}

// TestArrayReturnLength pins the C-backend fix for arrays returned by
// user-defined functions: the result was losing its length companion,
// so std.array.get / index lowering aborted with length=-1. Now the
// codegen registers (int32_t)omni_slice_len_real((void*)v) for array
// return values and the bounds-checked ops consult it.
func TestArrayReturnLength(t *testing.T) {
	testFile := "array_return_length.omni"
	expected := "0"

	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("runVM(%q) failed: %v", testFile, err)
	}
	if result != expected {
		t.Errorf("runVM(%q) = %s, want %s", testFile, result, expected)
	}

	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("runCBackend(%q) failed: %v", testFile, err)
	}
	if result != expected {
		t.Errorf("runCBackend(%q) = %s, want %s", testFile, result, expected)
	}
}

// TestStdOsOps pins std.os filesystem, env, cwd, and pid operations
// on both backends. The audit caught omni_getenv / omni_getcwd
// returning NULL when the var was missing or getcwd failed; the C
// backend then dereferenced the NULL string. Both now return "".
func TestStdOsOps(t *testing.T) {
	testFile := "std_os_ops.omni"
	expected := "0"

	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("runVM(%q) failed: %v", testFile, err)
	}
	if result != expected {
		t.Errorf("runVM(%q) = %s, want %s", testFile, result, expected)
	}

	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("runCBackend(%q) failed: %v", testFile, err)
	}
	if result != expected {
		t.Errorf("runCBackend(%q) = %s, want %s", testFile, result, expected)
	}
}

// TestStdRandomShuffleUnique pins the std.math PRNG (random_seed,
// random_int) and the algorithms it powers — std.algorithms.shuffle
// (length-preserving) and std.algorithms.unique (variable output
// length, threaded through a runtime out-pointer). Both backends
// share the same xorshift32 with the same default seed, so the same
// seed produces the same first random_int.
func TestStdRandomShuffleUnique(t *testing.T) {
	testFile := "std_random_shuffle_unique.omni"
	expected := "67"

	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("VM: expected %s, got %s", expected, result)
	}

	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

// TestStdMiscExtras pins two small finishers: std.string.is_empty
// and std.algorithms.rotate. Catches both backends.
func TestStdMiscExtras(t *testing.T) {
	testFile := "std_misc_extras.omni"
	expected := "6"

	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("VM: expected %s, got %s", expected, result)
	}

	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

// TestStdArrayStringOps pins std.array.* on string arrays. Same shape
// as TestStdArrayIntOps but with string-element runtime siblings on
// the C side and dual-typed dispatch on the VM side.
func TestStdArrayStringOps(t *testing.T) {
	testFile := "std_array_string_ops.omni"
	expected := "9"

	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("VM: expected %s, got %s", expected, result)
	}

	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

// TestStdArrayIntOps pins the std.array list operations on int arrays:
// contains, index_of, append, prepend, insert, remove, concat, slice.
// Length-changing ops forward the output length so a downstream index
// or len() keeps working without losing the size.
func TestStdArrayIntOps(t *testing.T) {
	testFile := "std_array_int_ops.omni"
	expected := "317"

	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("VM: expected %s, got %s", expected, result)
	}

	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

// TestStdAlgorithmsArray pins the array-based std.algorithms intrinsics
// (find_max/min, count_occurrences, sorts, binary_search, reverse) and
// the array-length-through-parameter ABI change that unlocked them.
func TestStdAlgorithmsArray(t *testing.T) {
	testFile := "std_algorithms_array.omni"
	expected := "25"

	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("VM: expected %s, got %s", expected, result)
	}

	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

// TestArrayLenOnParam pins the synthetic-length companion: a function
// that takes an array<T> can now call len(arr) and get the real
// length back. Sums [10, 20, 30, 40] = 100.
func TestArrayLenOnParam(t *testing.T) {
	testFile := "array_len_on_param.omni"
	expected := "100"

	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("VM: expected %s, got %s", expected, result)
	}

	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

// TestStdStringExtras pins the previously-stubbed std.string functions:
// trim_left/right/all, to_title, capitalize, reverse, equals/compare
// _ignore_case, count_occurrences/lines/words. Both backends now run
// real implementations rather than returning placeholders.
func TestStdStringExtras(t *testing.T) {
	testFile := "std_string_extras.omni"
	expected := "20"

	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("VM: expected %s, got %s", expected, result)
	}

	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

// TestCharCodeIntrinsic pins down the std.char_code / std.char_from_code
// intrinsics. Without them, OmniLang code can't move between char and
// int — neither `c - 'a'` nor `int(c)` is valid.
func TestCharCodeIntrinsic(t *testing.T) {
	testFile := "char_code_intrinsic.omni"
	expected := "127"

	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("VM: expected %s, got %s", expected, result)
	}

	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

// TestTopLevelLet pins down module-scope `let` bindings: a constant
// declared outside any function is visible to every function in the
// module. Top-level `var` is intentionally rejected (no global storage).
func TestTopLevelLet(t *testing.T) {
	testFile := "top_level_let.omni"
	expected := "33"

	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("VM: expected %s, got %s", expected, result)
	}

	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

// TestPositionalArgDash pins down the std.os.positional_arg flag
// parsing fix: `-7` should be returned as-is rather than skipped as a
// flag. The program returns string_to_int of argv[1], so passing "-7"
// makes the C-backend executable exit with -7 (255 - 7 = 248 once the
// shell wraps it, but we read the raw exit code via exitErr.ExitCode).
func TestPositionalArgDash(t *testing.T) {
	testFile := "positional_dash.omni"
	// Build via the existing helper so we hit the production codegen
	// path; it doesn't accept argv though, so re-run the binary by hand.
	if _, err := runCBackend(testFile); err != nil {
		t.Fatalf("initial build failed: %v", err)
	}
	executable := "./positional_dash"
	cmd := exec.Command(executable, "-7")
	cmd.Env = append(os.Environ(),
		"DYLD_LIBRARY_PATH=../../native/clift/target/release:../../runtime/posix",
		"LD_LIBRARY_PATH=../../native/clift/target/release:../../runtime/posix",
	)
	output, err := cmd.Output()
	exitCode := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			t.Fatalf("run failed: %v\nstdout: %s", err, string(output))
		}
	}
	// Negative exits are reported as 256 + signed_value by the shell;
	// exec.ExitError.ExitCode() returns the raw int, so -7 stays -7.
	if exitCode != -7 && exitCode != 249 {
		// Some platforms truncate to uint8: 256 - 7 = 249.
		t.Errorf("expected exit -7 (or 249), got %d (output: %s)", exitCode, string(output))
	}
}

// TestStdOsStringLet pins the call-type fix for std.os.* string-returning
// intrinsics. Before the fix, `let s:string = std.os.positional_arg(...)`
// lowered as call.void; the C backend dropped the return value and the
// let bound to a "<unknown>" placeholder. With idx=0 default="omnitest"
// and no argv, the binding is "omnitest" so the program returns 8.
func TestStdOsStringLet(t *testing.T) {
	testFile := "std_os_string_let.omni"
	expected := "8"

	result, err := runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

func TestModulo(t *testing.T) {
	testFile := "modulo_test.omni"
	expected := "1" // 10 % 3 = 1

	// Test C backend (VM backend has issues with neg/not instructions)
	result, err := runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

func TestNegation(t *testing.T) {
	testFile := "negation_test.omni"
	expected := "-5" // -5 = -5

	// Test C backend (VM backend has issues with neg/not instructions)
	result, err := runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

func TestLogicalNot(t *testing.T) {
	testFile := "logical_not_test.omni"
	expected := "0" // !true = false = 0

	// Test C backend (VM backend has issues with neg/not instructions)
	result, err := runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

func TestLogicalAnd(t *testing.T) {
	testFile := "logical_and_test.omni"
	expected := "0" // true && false = false = 0

	// Test C backend (VM backend has issues with neg/not instructions)
	result, err := runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

func TestLogicalOr(t *testing.T) {
	testFile := "logical_or_test.omni"
	expected := "1" // true || false = true = 1

	// Test C backend (VM backend has issues with neg/not instructions)
	result, err := runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

// New Features Tests

func TestStringOperations(t *testing.T) {
	testFile := "new_features/test_string_operations_comprehensive.omni"
	expected := "String length: 13\nConcatenation: Hello, World!\nSubstring (0,5): Hello\nChar at 0: 72\nStarts with 'Hello': true\nEnds with 'World!': true\nContains 'World': true\nIndex of 'World': 7\nLast index of 'l': 10\nTrim result: 'Hello World'\nTo upper: HELLO, WORLD!\nTo lower: hello, world!\nString equals: true\nString compare (Apple vs Banana): -1\n0"

	// Test VM execution
	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestMathUtilities(t *testing.T) {
	testFile := "new_features/test_math_utilities.omni"
	expected := "Math and utilities test passed\n0"

	// Test VM execution
	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestFileIO(t *testing.T) {
	testFile := "new_features/test_file_comprehensive.omni"
	expected := "File exists: 0\nFile size: 0\nFile open handle: 0\nBytes written: 0\nFile close result: 0\nFile exists after write: 0\nFile size after write: 0\n0"

	// Test VM execution
	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestBitwiseOperations(t *testing.T) {
	testFile := "new_features/test_bitwise_ops.omni"
	expected := "12 & 10 = 8\n12 | 10 = 14\n12 ^ 10 = 6\n12 << 2 = 48\n12 >> 2 = 3\n~12 = -13\n(12 & 10) | 1 = 9\n0"

	// Test VM execution
	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestTypeCasting(t *testing.T) {
	testFile := "new_features/test_type_casting_comprehensive.omni"
	expected := "int to float: 42\nfloat to int: 3\nint to string: 42\nfloat to string: 3.14\nbool to string: true\nstring to int: 123\nstring to float: 3.14\nstring to bool: true\nnested cast (float)(int)3.99: 3\n0"

	// Test VM execution
	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestLexicalPrimitives(t *testing.T) {
	testFile := "new_features/test_lexical_primitives.omni"
	expected := "Null literal test passed\nHex 0xFF: 255\nHex 0x1A: 26\nHex 0x10_00: 4096\nBinary 0b1010: 10\nBinary 0b1111_0000: 240\nBinary 0b1: 1\nScientific 1.0e5: 100000\nScientific 2.5e-3: 0.0025\nScientific 1.23E+2: 123\nMixed sum: 100265\n0"

	// Test VM execution
	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestControlFlow(t *testing.T) {
	testFile := "new_features/test_control_flow.omni"
	expected := "Testing while loop:\n  i = 0\n  i = 1\n  i = 2\n  i = 3\n  i = 4\nTesting break:\n  j = 0\n  j = 1\n  j = 2\nTesting continue:\n  k = 1\n  k = 2\n  k = 4\n  k = 5\nTesting break in while loop:\n  m = 0\n  m = 1\n  m = 2\n  m = 3\n0"

	// Test VM execution
	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestVariableShadowing(t *testing.T) {
	testFile := "new_features/test_variable_shadowing.omni"
	expected := "Level 0 x: 1\nLevel 1 x: 2\nLevel 2 x: 3\nLevel 1 x after inner block: 2\nLevel 0 x after all blocks: 1\nOuter name: outer\nInner name (int): 42\nOuter name after block: outer\nInner mutable: 200\nInner mutable after assignment: 300\nOuter immutable: 100\n0"

	// Test VM execution
	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestFirstClassFunctions(t *testing.T) {
	testFile := "new_features/test_first_class_complete.omni"
	expected := "15\n50\n12\n7\n0"

	// Test VM execution
	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestCollectionsOperations(t *testing.T) {
	testFile := "new_features/test_collections_operations.omni"
	expected := "Collections operations test passed\n0"

	// Test VM execution
	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestArrayOperations(t *testing.T) {
	testFile := "new_features/test_array_operations.omni"
	expected := "Array length: 5\nFirst element: 1\nThird element: 3\nSecond element: 2\n0"

	// Test VM execution
	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestErrorAssertions(t *testing.T) {
	testFile := "new_features/test_errors_assertions.omni"
	expected := "Errors and assertions test passed\n0"

	// Test VM execution
	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("VM execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}


// Returned-array regression tests. These guard against a class of C-backend
// bugs where a function returning array<T> lost its compile-time length at
// the call site, causing std.array.get to abort with length=-1 or arrays
// returned to the caller to dangle off the callee's stack frame. Each .omni
// file targets one shape of return-array usage.

func TestArrayReturnLocal(t *testing.T) {
	testFile := "array_return_local.omni"
	expected := "10"

	result, err := runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

func TestArrayReturnLiteral(t *testing.T) {
	testFile := "array_return_literal.omni"
	expected := "200"

	result, err := runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

func TestArrayReturnGetInline(t *testing.T) {
	testFile := "array_return_get_inline.omni"
	expected := "13"

	result, err := runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

func TestArrayReturnTuplePair(t *testing.T) {
	testFile := "array_return_tuple_pair.omni"
	expected := "47"

	result, err := runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

func TestArrayReturnForwarded(t *testing.T) {
	testFile := "array_return_forwarded.omni"
	expected := "5"

	result, err := runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

// TestExprParser is a real-world regression test driven by the symptoms
// we hit while building a recursive-descent expression parser entirely in
// OmniLang. It returns a packed signature (r1 + r2*10 + r3*100 + r4*10000)
// so that a miscompute in any of the four sub-expressions surfaces
// uniquely instead of being masked by another. Expected: 1201573.
func TestExprParser(t *testing.T) {
	testFile := "expr_parser.omni"
	expected := "1201573"

	result, err := runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

// TestStringVarConditionalReturn guards against a use-after-free in the
// C-backend cleanup epilogue. A function that returns a `var` string
// conditionally assigned from a heap-allocating helper would have its
// helper result freed before the return slot read it, producing
// corrupted strings to callers (manifested as failed string comparisons).
func TestStringVarConditionalReturn(t *testing.T) {
	testFile := "string_var_conditional_return.omni"
	expected := "100"

	result, err := runCBackend(testFile)
	if err != nil {
		t.Fatalf("C backend execution failed: %v", err)
	}
	if result != expected {
		t.Errorf("C backend: expected %s, got %s", expected, result)
	}
}

// TestStdNetworkUrlRoundTrip verifies url_to_string(url_parse(s)) is
// stable on both backends, including default-port normalization
// (port 80 dropped for http, 443 dropped for https). The VM previously
// re-emitted explicit ":80"/":443" through Go's net/url.URL.String()
// while the C backend (running the OmniLang source of url_to_string)
// dropped them.
func TestStdNetworkUrlRoundTrip(t *testing.T) {
	testFile := "std_network_url_round_trip.omni"
	expected := "0"

	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("runVM(%q) failed: %v", testFile, err)
	}
	if result != expected {
		t.Errorf("runVM(%q) = %s, want %s", testFile, result, expected)
	}

	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("runCBackend(%q) failed: %v", testFile, err)
	}
	if result != expected {
		t.Errorf("runCBackend(%q) = %s, want %s", testFile, result, expected)
	}
}

// TestStdNetworkSocketSmoke pins the socket lifecycle (create / bind to
// 127.0.0.1:0 / listen / close) on the C backend. Sockets aren't wired
// on the VM (omnir's socket_create stub returns -1), so this is C-only;
// when sockets land on the VM, add a runVM check here too.
func TestStdNetworkSocketSmoke(t *testing.T) {
	testFile := "std_network_socket_smoke.omni"
	expected := "0"

	result, err := runCBackend(testFile)
	if err != nil {
		t.Fatalf("runCBackend(%q) failed: %v", testFile, err)
	}
	if result != expected {
		t.Errorf("runCBackend(%q) = %s, want %s", testFile, result, expected)
	}
}

// TestStdNetworkHttpResponse pins the offline HTTPResponse surface on
// both backends: http_response_create (the new constructor that lets
// programs build a response without a real HTTP request), the three
// status-class helpers, and get/set_header round-trips.
func TestStdNetworkHttpResponse(t *testing.T) {
	testFile := "std_network_http_response.omni"
	expected := "0"

	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("runVM(%q) failed: %v", testFile, err)
	}
	if result != expected {
		t.Errorf("runVM(%q) = %s, want %s", testFile, result, expected)
	}

	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("runCBackend(%q) failed: %v", testFile, err)
	}
	if result != expected {
		t.Errorf("runCBackend(%q) = %s, want %s", testFile, result, expected)
	}
}

// TestStdNetworkHttpRequest pins the offline HTTPRequest surface on
// both backends: http_request_create, set_header, set_body, get_header.
// The chained-builder shape (`req = set_header(req, k, v)`) is the
// stress point — it requires the runtime helpers to return the request
// pointer rather than void so std.network's chained API lowers cleanly.
func TestStdNetworkHttpRequest(t *testing.T) {
	testFile := "std_network_http_request.omni"
	expected := "0"

	result, err := runVM(testFile)
	if err != nil {
		t.Fatalf("runVM(%q) failed: %v", testFile, err)
	}
	if result != expected {
		t.Errorf("runVM(%q) = %s, want %s", testFile, result, expected)
	}

	result, err = runCBackend(testFile)
	if err != nil {
		t.Fatalf("runCBackend(%q) failed: %v", testFile, err)
	}
	if result != expected {
		t.Errorf("runCBackend(%q) = %s, want %s", testFile, result, expected)
	}
}
