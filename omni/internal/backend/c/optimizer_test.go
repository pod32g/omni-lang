package cbackend

import (
	"strings"
	"testing"
)

func TestCOptimizer(t *testing.T) {
	t.Run("NewCOptimizer", func(t *testing.T) {
		optimizer := NewCOptimizer("O1")
		if optimizer == nil {
			t.Fatal("NewCOptimizer returned nil")
		}
	})

	t.Run("OptimizeC", func(t *testing.T) {
		// Test with basic C code
		input := `
#include <stdio.h>
int main() {
    int x = 42;
    int y = x + 5;
    return y;
}
`

		result := OptimizeC(input, "O1")
		if result == "" {
			t.Error("Expected non-empty optimized C code")
		}

		// Should still contain the main function
		if !strings.Contains(result, "int main()") {
			t.Error("Expected main function in optimized code")
		}
	})

	t.Run("BasicOptimizations", func(t *testing.T) {
		optimizer := NewCOptimizer("O1")

		// Test with code that can be optimized
		input := `
int main() {
    int x = 42;
    int y = x + 0;
    return y;
}
`

		result := optimizer.optimize(input)
		if result == "" {
			t.Error("Expected non-empty optimized C code")
		}
	})

	t.Run("StandardOptimizations", func(t *testing.T) {
		optimizer := NewCOptimizer("O1")

		// Test with code that can be optimized
		input := `
int main() {
    int x = 42;
    int y = x * 1;
    return y;
}
`

		result := optimizer.optimize(input)
		if result == "" {
			t.Error("Expected non-empty optimized C code")
		}
	})

	t.Run("AggressiveOptimizations", func(t *testing.T) {
		optimizer := NewCOptimizer("O1")

		// Test with code that can be optimized
		input := `
int main() {
    int x = 42;
    int y = x + x;
    return y;
}
`

		result := optimizer.optimize(input)
		if result == "" {
			t.Error("Expected non-empty optimized C code")
		}
	})

	t.Run("SizeOptimizations", func(t *testing.T) {
		optimizer := NewCOptimizer("O1")

		// Test with code that can be optimized for size
		input := `
int main() {
    int x = 42;
    int y = x + 5;
    return y;
}
`

		result := optimizer.optimize(input)
		if result == "" {
			t.Error("Expected non-empty size-optimized C code")
		}
	})

	t.Run("RemoveUnusedVariables", func(t *testing.T) {
		optimizer := NewCOptimizer("O1")

		// Test with unused variables
		input := `
int main() {
    int x = 42;
    int y = 5;
    return x;
}
`

		result := optimizer.removeUnusedVariables(input)
		if result == "" {
			t.Error("Expected non-empty result after removing unused variables")
		}

		// Should still contain the used variable
		if !strings.Contains(result, "int x = 42") {
			t.Error("Expected used variable to remain")
		}
	})

	t.Run("SimplifyConstants", func(t *testing.T) {
		optimizer := NewCOptimizer("O1")

		// Test with constant expressions
		input := `
int main() {
    int x = 42 + 5;
    return x;
}
`

		result := optimizer.simplifyConstants(input)
		if result == "" {
			t.Error("Expected non-empty result after simplifying constants")
		}
	})

	t.Run("SimplifyArithmeticExpression", func(t *testing.T) {
		optimizer := NewCOptimizer("O1")

		// Test with arithmetic expressions
		input := `
int main() {
    int x = 42 + 0;
    return x;
}
`

		result := optimizer.simplifyArithmeticExpression(input)
		if result == "" {
			t.Error("Expected non-empty result after simplifying arithmetic")
		}
	})

	t.Run("InlineSimpleFunctions", func(t *testing.T) {
		optimizer := NewCOptimizer("O1")

		// Test with simple functions
		input := `
int add(int a, int b) {
    return a + b;
}

int main() {
    int x = add(42, 5);
    return x;
}
`

		result := optimizer.inlineSimpleFunctions(input)
		if result == "" {
			t.Error("Expected non-empty result after inlining functions")
		}
	})

	t.Run("OptimizeArithmetic", func(t *testing.T) {
		optimizer := NewCOptimizer("O1")

		// Test with arithmetic operations
		input := `
int main() {
    int x = 42 * 2;
    return x;
}
`

		result := optimizer.optimizeArithmetic(input)
		if result == "" {
			t.Error("Expected non-empty result after optimizing arithmetic")
		}
	})

	t.Run("OptimizeArithmeticLine", func(t *testing.T) {
		optimizer := NewCOptimizer("O1")

		// Test with arithmetic line
		line := "int x = 42 * 2;"

		result := optimizer.optimizeArithmeticLine(line)
		if result == "" {
			t.Error("Expected non-empty result after optimizing arithmetic line")
		}
	})

	t.Run("AggressiveInlining", func(t *testing.T) {
		optimizer := NewCOptimizer("O1")

		// Test with functions that can be inlined
		input := `
int add(int a, int b) {
    return a + b;
}

int main() {
    int x = add(42, 5);
    return x;
}
`

		result := optimizer.aggressiveInlining(input)
		if result == "" {
			t.Error("Expected non-empty result after aggressive inlining")
		}
	})

	t.Run("OptimizeLoops", func(t *testing.T) {
		optimizer := NewCOptimizer("O1")

		// Test with loops
		input := `
int main() {
    int sum = 0;
    for (int i = 0; i < 10; i++) {
        sum += i;
    }
    return sum;
}
`

		result := optimizer.optimizeLoops(input)
		if result == "" {
			t.Error("Expected non-empty result after optimizing loops")
		}
	})

	t.Run("MinimizeVariableNames", func(t *testing.T) {
		optimizer := NewCOptimizer("O1")

		// Test with variable names
		input := `
int main() {
    int very_long_variable_name = 42;
    int another_long_name = very_long_variable_name + 5;
    return another_long_name;
}
`

		result := optimizer.minimizeVariableNames(input)
		if result == "" {
			t.Error("Expected non-empty result after minimizing variable names")
		}
	})

	t.Run("RemoveDebugInfo", func(t *testing.T) {
		optimizer := NewCOptimizer("O1")

		// Test with debug info
		input := `
int main() {
    int x = 42;
    // Debug comment
    return x;
}
`

		result := optimizer.removeDebugInfo(input)
		if result == "" {
			t.Error("Expected non-empty result after removing debug info")
		}
	})
}
