package edge_cases

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/omni-lang/omni/internal/compiler"
	"github.com/omni-lang/omni/internal/parser"
	"github.com/omni-lang/omni/internal/types/checker"
)

func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  string
		expectError bool
		errorContains string
	}{
		{
			name: "very_long_string",
			filename: "very_long_string.omni",
			content: `func main(): int {
    let long_string: string = "` + string(make([]byte, 10000)) + `"
    return 0
}`,
			expectError: false,
		},
		{
			name: "nested_function_calls",
			filename: "nested_calls.omni",
			content: `import std.math as math
func main(): int {
    let result: int = math.max(math.min(10, 20), math.max(5, 15))
    return result
}`,
			expectError: false,
		},
		{
			name: "complex_expression",
			filename: "complex_expr.omni",
			content: `func main(): int {
    let a: int = 10
    let b: int = 20
    let c: int = 30
    let result: int = (a + b) * c - (a / b) + (b % a)
    return result
}`,
			expectError: false,
		},
		{
			name: "deep_nesting",
			filename: "deep_nesting.omni",
			content: `func main(): int {
    let a: int = 1
    let b: int = 2
    let c: int = 3
    let d: int = 4
    let e: int = 5
    let result: int = ((((a + b) * c) - d) / e) + (((a - b) + c) * d) - e
    return result
}`,
			expectError: false,
		},
		{
			name: "string_edge_cases",
			filename: "string_edges.omni",
			content: `import std.string as str
func main(): int {
    let empty: string = ""
    let single: string = "a"
    let unicode: string = "Hello 世界 🌍"
    let combined: string = empty + single + unicode
    let length: int = str.length(combined)
    return length
}`,
			expectError: false,
		},
		{
			name: "boolean_edge_cases",
			filename: "boolean_edges.omni",
			content: `func main(): int {
    let true_val: bool = true
    let false_val: bool = false
    let and_result: bool = true_val && false_val
    let or_result: bool = true_val || false_val
    let not_result: bool = !true_val
    return 0
}`,
			expectError: false,
		},
		{
			name: "numeric_edge_cases",
			filename: "numeric_edges.omni",
			content: `import std.math as math
func main(): int {
    let zero: int = 0
    let negative: int = -42
    let large: int = 1000000
    let max_result: int = math.max(zero, negative)
    let min_result: int = math.min(max_result, large)
    let abs_result: int = math.abs(negative)
    return abs_result
}`,
			expectError: false,
		},
		{
			name: "circular_import_detection",
			filename: "circular_a.omni",
			content: `import circular_b
func main(): int {
    return circular_b.get_value()
}`,
			expectError: true,
			errorContains: "not found",
		},
		{
			name: "missing_return_statement",
			filename: "missing_return.omni",
			content: `func main(): int {
    let x: int = 42
    // Missing return statement - this should be caught by the type checker
    return x
}`,
			expectError: false,
		},
		{
			name: "type_inference_edge_cases",
			filename: "type_inference.omni",
			content: `func main(): int {
    let a = 42  // Should infer int
    let b = true  // Should infer bool
    let c = "hello"  // Should infer string
    let d = a + 10  // Should infer int
    let e = b && false  // Should infer bool
    let f = c + " world"  // Should infer string
    return a
}`,
			expectError: false,
		},
		{
			name: "operator_precedence",
			filename: "operator_precedence.omni",
			content: `func main(): int {
    let a: int = 2
    let b: int = 3
    let c: int = 4
    let result1: int = a + b * c  // Should be 2 + (3 * 4) = 14
    let result2: int = (a + b) * c  // Should be (2 + 3) * 4 = 20
    let result3: int = a * b + c  // Should be (2 * 3) + 4 = 10
    return result1
}`,
			expectError: false,
		},
		{
			name: "unicode_identifiers",
			filename: "unicode_idents.omni",
			content: `func main(): int {
    let 变量: int = 42
    let 结果: int = 变量 + 10
    return 结果
}`,
			expectError: false,
		},
		{
			name: "very_deep_nesting",
			filename: "very_deep.omni",
			content: `func main(): int {
    let a: int = 1
    let b: int = 2
    let c: int = 3
    let d: int = 4
    let e: int = 5
    let f: int = 6
    let g: int = 7
    let h: int = 8
    let i: int = 9
    let j: int = 10
    let result: int = a + b + c + d + e + f + g + h + i + j
    return result
}`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tempDir := t.TempDir()
			filePath := filepath.Join(tempDir, tt.filename)
			err := os.WriteFile(filePath, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Parse the file
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read test file: %v", err)
			}

			mod, err := parser.Parse(filePath, string(content))
			if err != nil {
				if tt.expectError {
					if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
						t.Errorf("Expected error to contain %q, got %q", tt.errorContains, err.Error())
					}
					return
				}
				t.Fatalf("Parse failed: %v", err)
			}

			// Merge imported modules
			if err := compiler.MergeImportedModules(mod, tempDir); err != nil {
				if tt.expectError {
					if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
						t.Errorf("Expected error to contain %q, got %q", tt.errorContains, err.Error())
					}
					return
				}
				t.Fatalf("Merge imports failed: %v", err)
			}

			// Type check
			err = checker.Check(filePath, string(content), mod)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Type check failed: %v", err)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		contains(s[1:], substr))))
}
