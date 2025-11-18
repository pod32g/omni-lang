package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/omni-lang/omni/internal/parser"
	"github.com/omni-lang/omni/internal/types/checker"
)

func TestIntegrationCompilation(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		wantErr  bool
		errorMsg string
	}{
		{
			name: "simple_hello_world",
			source: `
func main() {
    let message:string = "Hello, World!"
}`,
			wantErr: false,
		},
		{
			name: "arithmetic_operations",
			source: `
func main() {
    let a:int = 10
    let b:int = 5
    let sum:int = a + b
    let diff:int = a - b
    let prod:int = a * b
    let quot:int = a / b
}`,
			wantErr: false,
		},
		{
			name: "string_operations",
			source: `
func main() {
    let greeting:string = "  Hello World  "
    let combined:string = "Original: '" + greeting + "'"
}`,
			wantErr: false,
		},
		{
			name: "conditional_logic",
			source: `
func main() {
    let x:int = 15
    if x > 10 {
        let msg:string = "x is greater than 10"
    } else {
        let msg:string = "x is not greater than 10"
    }
}`,
			wantErr: false,
		},
		{
			name: "function_calls",
			source: `
func add(a:int, b:int) {
    let result:int = a + b
}

func main() {
    add(5, 3)
}`,
			wantErr: false,
		},
		{
			name: "type_error",
			source: `
func main() {
    let x:int = 5
    let y:string = "hello"
    let z:int = x + y  // Type error: string assigned to int
}`,
			wantErr:  true,
			errorMsg: "string concatenation requires a string on the left-hand side",
		},
		{
			name: "undefined_identifier",
			source: `
func main() {
    prnt("Hello")  // Typo in print
}`,
			wantErr:  true,
			errorMsg: "undefined identifier",
		},
		{
			name: "syntax_error",
			source: `
func main() {
    let x:int =   // Missing value
}`,
			wantErr:  true,
			errorMsg: "unexpected token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpFile := filepath.Join(t.TempDir(), "test.omni")
			err := os.WriteFile(tmpFile, []byte(tt.source), 0644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Parse the source
			mod, err := parser.Parse(tmpFile, tt.source)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Unexpected parse error: %v", err)
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain %q, got: %v", tt.errorMsg, err)
				}
				return
			}

			// Type check
			err = checker.Check(tmpFile, tt.source, mod)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Unexpected type check error: %v", err)
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain %q, got: %v", tt.errorMsg, err)
				}
				return
			}

			// If we expected an error but didn't get one
			if tt.wantErr {
				t.Errorf("Expected an error but compilation succeeded")
			}
		})
	}
}

func TestIntegrationEndToEnd(t *testing.T) {
	// Test complete compilation pipeline
	source := `
func factorial(n:int) {
    if n <= 1 {
        let result:int = 1
    } else {
        let result:int = n * 1  // Simplified for testing
    }
}

func main() {
    factorial(5)
}`

	tmpFile := filepath.Join(t.TempDir(), "factorial.omni")
	err := os.WriteFile(tmpFile, []byte(source), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test compilation
	cfg := Config{
		InputPath: tmpFile,
		Backend:   "vm",
		Emit:      "mir",
	}

	err = Compile(cfg)
	if err != nil {
		t.Errorf("Compilation failed: %v", err)
	}
}

func TestIntegrationErrorRecovery(t *testing.T) {
	// Test that the compiler can recover from multiple errors
	source := `
func main() {
    prnt("Hello")  // Typo
    let x:int = 5
    let y:string = "world"
    let result = x + y  // Type error
}`

	tmpFile := filepath.Join(t.TempDir(), "error_recovery.omni")
	err := os.WriteFile(tmpFile, []byte(source), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Parse should succeed even with errors
	mod, err := parser.Parse(tmpFile, source)
	if err != nil {
		t.Errorf("Parser should recover from errors, got: %v", err)
	}

	// Type check should report multiple errors
	err = checker.Check(tmpFile, source, mod)
	if err == nil {
		t.Errorf("Expected type check errors but got none")
	}

	// Should contain error (either undefined identifier or type mismatch)
	errorStr := err.Error()
	if !strings.Contains(errorStr, "undefined identifier") && !strings.Contains(errorStr, "cannot assign") {
		t.Errorf("Expected undefined identifier or type mismatch error, got: %s", errorStr)
	}
}
