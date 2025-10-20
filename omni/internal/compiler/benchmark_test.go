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
import std.io as io
import std.math as math

func fibonacci(n:int) {
    if n <= 1 {
        return n
    }
    return fibonacci(n - 1) + fibonacci(n - 2)
}

func main() {
    let result:int = fibonacci(10)
    io.println("Fibonacci(10): " + math.toString(result))
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
import std.io as io
import std.math as math

func factorial(n:int) {
    if n <= 1 {
        return 1
    }
    return n * factorial(n - 1)
}

func main() {
    let result:int = factorial(5)
    io.println("Factorial: " + math.toString(result))
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
import std.io as io
import std.math as math

func quicksort(arr:array, low:int, high:int) {
    if low < high {
        let pivot:int = partition(arr, low, high)
        quicksort(arr, low, pivot - 1)
        quicksort(arr, pivot + 1, high)
    }
}

func partition(arr:array, low:int, high:int) {
    let pivot:int = arr[high]
    let i:int = low - 1
    
    for j:int = low; j < high; j++ {
        if arr[j] <= pivot {
            i = i + 1
            let temp:int = arr[i]
            arr[i] = arr[j]
            arr[j] = temp
        }
    }
    
    let temp:int = arr[i + 1]
    arr[i + 1] = arr[high]
    arr[high] = temp
    return i + 1
}

func main() {
    let arr:array = [5, 2, 8, 1, 9, 3, 7, 4, 6]
    quicksort(arr, 0, 8)
    io.println("Sorted array processed")
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
	source.WriteString("import std.io as io\nimport std.math as math\n\n")

	// Generate 100 functions
	for i := 0; i < 100; i++ {
		source.WriteString("func func" + strconv.Itoa(i) + "(x:int, y:int) {\n")
		source.WriteString("    let result:int = x + y\n")
		source.WriteString("    return result * 2\n")
		source.WriteString("}\n\n")
	}

	source.WriteString("func main() {\n")
	source.WriteString("    let sum:int = 0\n")
	for i := 0; i < 50; i++ {
		source.WriteString("    sum = sum + func" + strconv.Itoa(i) + "(1, 2)\n")
	}
	source.WriteString("    io.println(\"Sum: \" + math.toString(sum))\n")
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
	source.WriteString("import std.io as io\n\n")
	source.WriteString("func main() {\n")

	// Generate many errors
	for i := 0; i < 50; i++ {
		source.WriteString("    prnt" + strconv.Itoa(i) + "(\"Hello\")\n")        // Typo
		source.WriteString("    let x" + strconv.Itoa(i) + ":int = \"string\"\n") // Type error
	}
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
