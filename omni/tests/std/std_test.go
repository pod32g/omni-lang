package std

import (
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
