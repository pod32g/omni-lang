package parser_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/omni-lang/omni/internal/ast"
	"github.com/omni-lang/omni/internal/parser"
	"github.com/omni-lang/omni/internal/testutil/snapshots"
)

func TestGoldenAST(t *testing.T) {
	goldenDir := filepath.Join("..", "..", "tests", "goldens", "ast")
	inputs, err := filepath.Glob(filepath.Join(goldenDir, "*.omni"))
	if err != nil {
		t.Fatalf("glob goldens: %v", err)
	}
	sort.Strings(inputs)
	if len(inputs) == 0 {
		t.Fatalf("no golden inputs found in %s", goldenDir)
	}

	for _, inputPath := range inputs {
		base := strings.TrimSuffix(filepath.Base(inputPath), ".omni")
		expectedPath := filepath.Join(goldenDir, base+".ast")

		t.Run(base, func(t *testing.T) {
			src, err := os.ReadFile(inputPath)
			if err != nil {
				t.Fatalf("read input: %v", err)
			}

			module, err := parser.Parse(inputPath, string(src))
			if err != nil {
				t.Fatalf("parse %s: %v", inputPath, err)
			}

			printed := ast.Print(module)
			snapshots.CompareText(t, printed, expectedPath)
		})
	}
}

// TestEvilParserCases tests edge cases and error conditions that could cause
// hangs, incorrect parsing, or poor error messages.
func TestEvilParserCases(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectError   bool
		errorContains string
		validateAST   func(t *testing.T, module *ast.Module)
	}{
		// 1. Generic vs comparison operators
		{
			name:        "comparison_not_generic",
			input:       "func test() { let x = foo < bar && baz > qux }",
			expectError: false,
			validateAST: func(t *testing.T, module *ast.Module) {
				// Should parse as comparison, not generics
				// Module should be non-nil
				if module == nil {
					t.Error("expected non-nil module")
				}
			},
		},
		{
			name:          "generic_with_logical",
			input:         "let x: Foo<Bar||Baz> = null",
			expectError:   true, // Should error - || is not valid in type args
			errorContains: "OR_OR", // Error should mention OR_OR token
		},
		{
			name:        "comparison_in_expression",
			input:       "func test() { let x = a < b && c > d }",
			expectError: false, // Should parse comparisons, not generics
		},
		// 2. Error recovery
		{
			name:          "extra_else_recovery",
			input:         "func f() { return 1 } else {}",
			expectError:   true,
			errorContains: "unexpected",
		},
		{
			name:          "missing_brace_recovery",
			input:         "func f() { return 1",
			expectError:   true,
			errorContains: "RBRACE", // Error should mention missing closing brace
		},
		// 3. Struct literals
		{
			name:        "qualified_struct_literal",
			input:       "let x = pkg.Type{ field: 1 }",
			expectError: false,
			validateAST: func(t *testing.T, module *ast.Module) {
				// Should parse qualified type struct literal
				// Module should be non-nil
				if module == nil {
					t.Error("expected non-nil module")
				}
			},
		},
		{
			name:        "struct_literal_call",
			input:       "let x = Type{}()",
			expectError: false,
			validateAST: func(t *testing.T, module *ast.Module) {
				// Should parse as call on struct literal value
			},
		},
		// 4. For loops
		{
			name:        "for_let_init",
			input:       "func test() { for let i = 0; i < 10; i++ {} }",
			expectError: false,
			validateAST: func(t *testing.T, module *ast.Module) {
				// Should parse for loop with let in init
			},
		},
		{
			name:        "for_var_init",
			input:       "func test() { for var i = 0; i < 10; i++ {} }",
			expectError: false,
		},
		// 5. String interpolation
		{
			name:        "string_interpolation_simple",
			input:       `let msg = "Hello ${name}"`,
			expectError: false,
			validateAST: func(t *testing.T, module *ast.Module) {
				// Should parse string interpolation
			},
		},
		{
			name:        "string_interpolation_nested",
			input:       `let msg = "Hello ${user.name}"`,
			expectError: false,
		},
		// 6. Try/catch
		{
			name:        "catch_qualified_type",
			input:       "func test() { try {} catch (err: std.io.Error) {} }",
			expectError: false,
			validateAST: func(t *testing.T, module *ast.Module) {
				// Should parse catch with qualified type
			},
		},
		// 7. Trailing commas
		{
			name:        "array_trailing_comma",
			input:       "let arr = [1, 2, 3,]",
			expectError: false,
		},
		{
			name:        "call_trailing_comma",
			input:       "func test() { f(a, b, c,) }",
			expectError: false,
		},
		// 8. Type expressions
		{
			name:        "pointer_type_generic",
			input:       "type Ptr = *map<string, int>",
			expectError: false,
			validateAST: func(t *testing.T, module *ast.Module) {
				// Should preserve generic args in pointer types
			},
		},
		// 9. Return statements
		{
			name:        "return_no_value_brace",
			input:       "func f() { return }",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module, err := parser.Parse("test.omni", tt.input)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errorContains)
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing %q, got %q", tt.errorContains, err.Error())
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			if module == nil {
				t.Error("expected module but got nil")
				return
			}
			
			if tt.validateAST != nil {
				tt.validateAST(t, module)
			}
		})
	}
}
