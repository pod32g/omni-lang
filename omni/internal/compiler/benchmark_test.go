package compiler

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/omni-lang/omni/internal/parser"
	"github.com/omni-lang/omni/internal/types/checker"
)

func BenchmarkParser(b *testing.B) {
	source := `
func fibonacci(n:int) : int {
    if n <= 1 {
        return n
    }
    return fibonacci(n - 1) + fibonacci(n - 2)
}

func main() : int {
    let result:int = fibonacci(10)
    return result
}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse("benchmark.omni", source)
		if err != nil {
			b.Fatalf("Parse failed: %v", err)
		}
	}
}

func BenchmarkTypeChecker(b *testing.B) {
	source := `
func factorial(n:int) : int {
    if n <= 1 {
        return 1
    }
    return n * factorial(n - 1)
}

func main() : int {
    let result:int = factorial(5)
    return result
}`

	mod, err := parser.Parse("benchmark.omni", source)
	if err != nil {
		b.Fatalf("Parse failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := checker.Check("benchmark.omni", source, mod)
		if err != nil {
			b.Fatalf("Type check failed: %v", err)
		}
	}
}

func BenchmarkCompilation(b *testing.B) {
	source := `
func add(x:int, y:int) : int {
    return x + y
}

func multiply(x:int, y:int) : int {
    return x * y
}

func main() : int {
    let result1:int = add(5, 3)
    let result2:int = multiply(result1, 2)
    return result2
}`

	tmpFile := filepath.Join(b.TempDir(), "benchmark.omni")
	err := os.WriteFile(tmpFile, []byte(source), 0644)
	if err != nil {
		b.Fatalf("Failed to write test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg := Config{
			InputPath: tmpFile,
			Backend:   "vm",
			Emit:      "mir",
		}
		err := Compile(cfg)
		if err != nil {
			b.Fatalf("Compilation failed: %v", err)
		}
	}
}

func BenchmarkLargeFile(b *testing.B) {
	// Generate a large source file with many functions
	var source strings.Builder

	// Generate 100 functions
	for i := 0; i < 100; i++ {
		source.WriteString("func func" + strconv.Itoa(i) + "(x:int, y:int) : int {\n")
		source.WriteString("    let result:int = x + y\n")
		source.WriteString("    return result * 2\n")
		source.WriteString("}\n\n")
	}

	source.WriteString("func main() : int {\n")
	source.WriteString("    var sum:int = 0\n")
	for i := 0; i < 50; i++ {
		source.WriteString("    sum = sum + func" + strconv.Itoa(i) + "(1, 2)\n")
	}
	source.WriteString("    return sum\n")
	source.WriteString("}\n")

	tmpFile := filepath.Join(b.TempDir(), "large.omni")
	err := os.WriteFile(tmpFile, []byte(source.String()), 0644)
	if err != nil {
		b.Fatalf("Failed to write test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg := Config{
			InputPath: tmpFile,
			Backend:   "vm",
			Emit:      "mir",
		}
		err := Compile(cfg)
		if err != nil {
			b.Fatalf("Compilation failed: %v", err)
		}
	}
}

func BenchmarkErrorRecovery(b *testing.B) {
	// Test performance with many errors
	var source strings.Builder
	source.WriteString("func main() : int {\n")

	// Generate many errors
	for i := 0; i < 50; i++ {
		source.WriteString("    prnt" + strconv.Itoa(i) + "(\"Hello\")\n")        // Typo
		source.WriteString("    let x" + strconv.Itoa(i) + ":int = \"string\"\n") // Type error
	}
	source.WriteString("    return 0\n")
	source.WriteString("}\n")

	tmpFile := filepath.Join(b.TempDir(), "errors.omni")
	err := os.WriteFile(tmpFile, []byte(source.String()), 0644)
	if err != nil {
		b.Fatalf("Failed to write test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mod, err := parser.Parse(tmpFile, source.String())
		if err != nil {
			// Expected to have parse errors
		}
		if mod != nil {
			checker.Check(tmpFile, source.String(), mod)
		}
	}
}
