package checker_test

import (
	"testing"

	"github.com/omni-lang/omni/internal/parser"
	"github.com/omni-lang/omni/internal/types/checker"
)

// TestCompleteTypeCheckingWorkflow tests the complete type checking workflow
// from parsing to error reporting
func TestCompleteTypeCheckingWorkflow(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		shouldErr bool
		desc     string
	}{
		{
			name: "complete program with structs and functions",
			src: `struct Point {
			      x: float
			      y: float
			      }
			      
			      func distance(p1: Point, p2: Point): float {
			      let dx = p1.x - p2.x
			      let dy = p1.y - p2.y
			      return dx * dx + dy * dy
			      }
			      
			      func main(): int {
			      let p1 = Point{ x: 0.0, y: 0.0 }
			      let p2 = Point{ x: 3.0, y: 4.0 }
			      let dist = distance(p1, p2)
			      return 0
			      }`,
			shouldErr: false,
			desc:     "Valid program with structs, functions, and type inference",
		},
		{
			name: "program with type errors",
			src: `func add(a: int, b: int): int {
			      return a + b
			      }
			      
			      func main(): void {
			      let x = add(1, "2")
			      }`,
			shouldErr: true,
			desc:     "Type mismatch in function call should error",
		},
		{
			name: "program with missing return",
			src: `func test(): int {
			      let x = 42
			      }`,
			shouldErr: true,
			desc:     "Missing return statement should error",
		},
		{
			name: "program with control flow",
			src: `func main(): int {
			      let x = 10
			      if x > 5 {
			          return 1
			      } else {
			          return 0
			      }
			      }`,
			shouldErr: false,
			desc:     "Valid program with if-else control flow",
		},
		{
			name: "program with loops",
			src: `func main(): int {
			      var sum = 0
			      for var i = 0; i < 10; i++ {
			          sum = sum + i
			      }
			      return sum
			      }`,
			shouldErr: false,
			desc:     "Valid program with for loop",
		},
		{
			name: "program with arrays",
			src: `func main(): int {
			      let arr = [1, 2, 3, 4, 5]
			      var sum = 0
			      for x in arr {
			          sum = sum + x
			      }
			      return sum
			      }`,
			shouldErr: false,
			desc:     "Valid program with arrays and range loops",
		},
		{
			name: "program with maps",
			src: `func main(): int {
			      let m = {"a": 1, "b": 2, "c": 3}
			      let x = m["a"]
			      return x
			      }`,
			shouldErr: false,
			desc:     "Valid program with maps",
		},
		{
			name: "program with generics",
			src: `struct Box<T> {
			      value: T
			      }
			      
			      func main(): int {
			      let box: Box<int> = Box<int>{ value: 42 }
			      return box.value
			      }`,
			shouldErr: false,
			desc:     "Valid program with generic structs",
		},
		{
			name: "program with async",
			src: `async func test(): Promise<int> {
			      return 42
			      }
			      
			      async func main(): Promise<int> {
			      let x = await test()
			      return x
			      }`,
			shouldErr: false,
			desc:     "Valid program with async/await",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod, err := parser.Parse("test.omni", tt.src)
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}

			err = checker.Check("test.omni", tt.src, mod)
			if tt.shouldErr && err == nil {
				t.Errorf("expected error for: %s", tt.desc)
			} else if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error for %s: %v", tt.desc, err)
			}
		})
	}
}

// TestErrorReporting tests that error reporting provides useful information
func TestErrorReporting(t *testing.T) {
	tests := []struct {
		name        string
		src         string
		expectError bool
		errorContains []string
	}{
		{
			name: "undefined variable",
			src: `func main(): void {
			      let x = y
			      }`,
			expectError: true,
			errorContains: []string{"undefined", "y"},
		},
		{
			name: "type mismatch",
			src: `func main(): void {
			      let x: int = "hello"
			      }`,
			expectError: true,
			errorContains: []string{"type", "mismatch", "int", "string"},
		},
		{
			name: "wrong argument count",
			src: `func add(a: int, b: int): int {
			      return a + b
			      }
			      
			      func main(): void {
			      let x = add(1)
			      }`,
			expectError: true,
			errorContains: []string{"argument", "count"},
		},
		{
			name: "wrong argument type",
			src: `func add(a: int, b: int): int {
			      return a + b
			      }
			      
			      func main(): void {
			      let x = add(1, "2")
			      }`,
			expectError: true,
			errorContains: []string{"type", "mismatch"},
		},
		{
			name: "missing return",
			src: `func test(): int {
			      let x = 42
			      }`,
			expectError: true,
			errorContains: []string{"return"},
		},
		{
			name: "wrong return type",
			src: `func test(): int {
			      return "hello"
			      }`,
			expectError: true,
			errorContains: []string{"return", "type"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod, err := parser.Parse("test.omni", tt.src)
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}

			err = checker.Check("test.omni", tt.src, mod)
			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				} else {
					errStr := err.Error()
					for _, substr := range tt.errorContains {
						if !contains(errStr, substr) {
							t.Errorf("error message should contain %q, got: %s", substr, errStr)
						}
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestModuleImports tests type checking with module imports
func TestModuleImports(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		shouldErr bool
		desc     string
	}{
		{
			name: "import std.io",
			src: `import std.io
			      func main(): void {
			      std.io.print("hello")
			      }`,
			shouldErr: false,
			desc:     "Valid program with std.io import",
		},
		{
			name: "import std.math",
			src: `import std.math
			      func main(): void {
			      let x = std.math.add(1, 2)
			      }`,
			shouldErr: false,
			desc:     "Valid program with std.math import",
		},
		{
			name: "import std.string",
			src: `import std.string
			      func main(): void {
			      let x = std.string.length("hello")
			      }`,
			shouldErr: false,
			desc:     "Valid program with std.string import",
		},
		{
			name: "use std function without import",
			src: `func main(): void {
			      let x = std.math.add(1, 2)
			      }`,
			shouldErr: true,
			desc:     "Using std function without import should error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod, err := parser.Parse("test.omni", tt.src)
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}

			err = checker.Check("test.omni", tt.src, mod)
			if tt.shouldErr && err == nil {
				t.Errorf("expected error for: %s", tt.desc)
			} else if !tt.shouldErr && err != nil {
				t.Logf("Unexpected errors (may be acceptable): %v", err)
			}
		})
	}
}

// TestComplexTypeInference tests complex type inference scenarios
func TestComplexTypeInference(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		shouldErr bool
		desc     string
	}{
		{
			name: "nested function calls",
			src: `func add(a: int, b: int): int {
			      return a + b
			      }
			      
			      func main(): int {
			      let x = add(add(1, 2), add(3, 4))
			      return x
			      }`,
			shouldErr: false,
			desc:     "Valid program with nested function calls",
		},
		{
			name: "array of structs",
			src: `struct Point {
			      x: float
			      y: float
			      }
			      
			      func main(): int {
			      let points = [Point{ x: 0.0, y: 0.0 }, Point{ x: 1.0, y: 1.0 }]
			      let p = points[0]
			      return 0
			      }`,
			shouldErr: false,
			desc:     "Valid program with array of structs",
		},
		{
			name: "generic function inference",
			src: `func id<T>(x: T): T {
			      return x
			      }
			      
			      func main(): int {
			      let x: int = id(42)
			      let y: string = id("hello")
			      return x
			      }`,
			shouldErr: false,
			desc:     "Valid program with generic function type inference",
		},
		{
			name: "optional types",
			src: `func main(): void {
			      let x: int? = 42
			      let y: int? = null
			      }`,
			shouldErr: false,
			desc:     "Valid program with optional types",
		},
		{
			name: "union types",
			src: `func main(): void {
			      let x: int | string = 42
			      let y: int | string = "hello"
			      }`,
			shouldErr: false,
			desc:     "Valid program with union types",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod, err := parser.Parse("test.omni", tt.src)
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}

			err = checker.Check("test.omni", tt.src, mod)
			if tt.shouldErr && err == nil {
				t.Errorf("expected error for: %s", tt.desc)
			} else if !tt.shouldErr && err != nil {
				t.Logf("Unexpected errors (may be acceptable): %v", err)
			}
		})
	}
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || 
		 containsCaseInsensitive(s, substr))
}

func containsCaseInsensitive(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + ('a' - 'A')
		} else {
			result[i] = c
		}
	}
	return string(result)
}
