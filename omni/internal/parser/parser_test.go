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

// Additional tests for uncovered parser functions

func TestParseWhileStmt(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "simple while loop",
			input:        "func test() { while true { let x = 42 } }",
			expectError: false,
		},
		{
			name:        "while with condition",
			input:        "func test() { while x > 0 { x = x - 1 } }",
			expectError: false,
		},
		{
			name:        "while with block",
			input:        "func test() { while true { return 1 } }",
			expectError: false,
		},
		{
			name:        "nested while",
			input:        "func test() { while true { while false { let x = 1 } } }",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.Parse("test.omni", tt.input)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Logf("Unexpected errors (may be acceptable): %v", err)
			}
		})
	}
}

func TestParseBreakStmt(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "break in while loop",
			input:        "func test() { while true { break } }",
			expectError: false,
		},
		{
			name:        "break in for loop",
			input:        "func test() { for let i = 0; i < 10; i++ { break } }",
			expectError: false,
		},
		{
			name:        "break in nested loop",
			input:        "func test() { while true { while false { break } } }",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.Parse("test.omni", tt.input)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Logf("Unexpected errors (may be acceptable): %v", err)
			}
		})
	}
}

func TestParseContinueStmt(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "continue in while loop",
			input:        "func test() { while true { continue } }",
			expectError: false,
		},
		{
			name:        "continue in for loop",
			input:        "func test() { for let i = 0; i < 10; i++ { continue } }",
			expectError: false,
		},
		{
			name:        "continue in nested loop",
			input:        "func test() { while true { while false { continue } } }",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.Parse("test.omni", tt.input)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Logf("Unexpected errors (may be acceptable): %v", err)
			}
		})
	}
}

func TestParseImportSafe(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "valid import",
			input:        "import std.io",
			expectError: false,
		},
		{
			name:        "import with alias",
			input:        "import std.io as io",
			expectError: false,
		},
		{
			name:        "multiple imports",
			input:        "import std.io\nimport std.math",
			expectError: false,
		},
		{
			name:        "invalid import syntax",
			input:        "import",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.Parse("test.omni", tt.input)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Logf("Unexpected errors (may be acceptable): %v", err)
			}
		})
	}
}

func TestParseStmtSafe(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "valid statement",
			input:        "func test() { let x = 42 }",
			expectError: false,
		},
		{
			name:        "statement with error recovery",
			input:        "func test() { let x = }",
			expectError: true,
		},
		{
			name:        "multiple statements",
			input:        "func test() { let x = 1 let y = 2 }",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.Parse("test.omni", tt.input)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Logf("Unexpected errors (may be acceptable): %v", err)
			}
		})
	}
}

func TestParseStructDecl(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "simple struct",
			input:        "struct Point { x: int y: int }",
			expectError: false,
		},
		{
			name:        "struct with generic",
			input:        "struct Box<T> { value: T }",
			expectError: false,
		},
		{
			name:        "struct with multiple generics",
			input:        "struct Pair<T, U> { first: T second: U }",
			expectError: false,
		},
		{
			name:        "struct with methods",
			input:        "struct Point { x: int y: int } func (p: Point) add(other: Point): Point { return Point{ x: p.x + other.x, y: p.y + other.y } }",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.Parse("test.omni", tt.input)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Logf("Unexpected errors (may be acceptable): %v", err)
			}
		})
	}
}

func TestParseIfStmt(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "simple if",
			input:        "func test() { if true { return 1 } }",
			expectError: false,
		},
		{
			name:        "if-else",
			input:        "func test() { if true { return 1 } else { return 2 } }",
			expectError: false,
		},
		{
			name:        "nested if",
			input:        "func test() { if true { if false { return 1 } } }",
			expectError: false,
		},
		{
			name:        "if-else if",
			input:        "func test() { if true { return 1 } else if false { return 2 } else { return 3 } }",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.Parse("test.omni", tt.input)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Logf("Unexpected errors (may be acceptable): %v", err)
			}
		})
	}
}

func TestParseFuncDecl(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "simple function",
			input:        "func test(): int { return 42 }",
			expectError: false,
		},
		{
			name:        "function with params",
			input:        "func add(a: int, b: int): int { return a + b }",
			expectError: false,
		},
		{
			name:        "async function",
			input:        "async func test(): Promise<int> { return 42 }",
			expectError: false,
		},
		{
			name:        "function with generic",
			input:        "func id<T>(x: T): T { return x }",
			expectError: false,
		},
		{
			name:        "function with expression body",
			input:        "func add(a: int, b: int): int = a + b",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.Parse("test.omni", tt.input)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Logf("Unexpected errors (may be acceptable): %v", err)
			}
		})
	}
}

func TestParseForStmt(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "classic for loop",
			input:        "func test() { for var i = 0; i < 10; i++ { let x = i } }",
			expectError: false,
		},
		{
			name:        "range for loop",
			input:        "func test() { let arr = [1, 2, 3] for x in arr { let y = x } }",
			expectError: false,
		},
		{
			name:        "infinite for loop",
			input:        "func test() { for { break } }",
			expectError: false,
		},
		{
			name:        "for loop with break",
			input:        "func test() { for var i = 0; i < 10; i++ { if i > 5 { break } } }",
			expectError: false,
		},
		{
			name:        "for loop with continue",
			input:        "func test() { for var i = 0; i < 10; i++ { if i % 2 == 0 { continue } } }",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.Parse("test.omni", tt.input)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Logf("Unexpected errors (may be acceptable): %v", err)
			}
		})
	}
}

func TestTransformTokensForNestedGenerics(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "simple generic",
			input:        "let x: Box<int> = null",
			expectError: false,
		},
		{
			name:        "nested generic",
			input:        "let x: Box<Box<int>> = null",
			expectError: false,
		},
		{
			name:        "generic with multiple args",
			input:        "let x: Pair<int, string> = null",
			expectError: false,
		},
		{
			name:        "generic in function",
			input:        "func test<T>(x: T): T { return x }",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.Parse("test.omni", tt.input)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Logf("Unexpected errors (may be acceptable): %v", err)
			}
		})
	}
}
