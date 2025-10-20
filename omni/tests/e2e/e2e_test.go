package e2e

import (
	"os/exec"
	"runtime"
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

func runVM(testFile string) (string, error) {
	cmd := exec.Command("go", "run", "../../cmd/omnir", testFile)
	cmd.Dir = "."
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

func runCompiler(testFile, backend, emit, output string) error {
	args := []string{"run", "../../cmd/omnic", "-backend", backend, "-emit", emit, testFile}
	if output != "" {
		args = append(args, "-o", output)
	}
	cmd := exec.Command("go", args...)
	cmd.Dir = "."
	return cmd.Run()
}
