package compiler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/omni-lang/omni/internal/mir/builder"
	"github.com/omni-lang/omni/internal/parser"
	"github.com/omni-lang/omni/internal/types/checker"
)

// PerformanceMetrics tracks performance data for regression testing
type PerformanceMetrics struct {
	Timestamp    time.Time              `json:"timestamp"`
	Platform     string                 `json:"platform"`
	Architecture string                 `json:"architecture"`
	GoVersion    string                 `json:"go_version"`
	Tests        map[string]TestMetrics `json:"tests"`
}

// TestMetrics tracks individual test performance
type TestMetrics struct {
	Name        string        `json:"name"`
	Duration    time.Duration `json:"duration"`
	MemoryUsage uint64        `json:"memory_usage"`
	Iterations  int           `json:"iterations"`
	Throughput  float64       `json:"throughput"` // operations per second
}

// PerformanceTestSuite manages performance testing and regression detection
type PerformanceTestSuite struct {
	metrics    PerformanceMetrics
	baseline   *PerformanceMetrics
	thresholds map[string]float64 // Performance regression thresholds
}

// NewPerformanceTestSuite creates a new performance test suite
func NewPerformanceTestSuite() *PerformanceTestSuite {
	return &PerformanceTestSuite{
		metrics: PerformanceMetrics{
			Timestamp:    time.Now(),
			Platform:     runtime.GOOS,
			Architecture: runtime.GOARCH,
			GoVersion:    runtime.Version(),
			Tests:        make(map[string]TestMetrics),
		},
		thresholds: map[string]float64{
			"parser":       0.10, // 10% regression threshold
			"typechecker":  0.15, // 15% regression threshold
			"compilation":  0.20, // 20% regression threshold
			"optimization": 0.25, // 25% regression threshold
		},
	}
}

// BenchmarkParserPerformance benchmarks parser performance
func BenchmarkParserPerformance(b *testing.B) {
	suite := NewPerformanceTestSuite()

	// Test with different source sizes
	sizes := []int{100, 500, 1000, 2000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			source := generateLargeSource(size)

			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			startMem := memStats.Alloc

			b.ResetTimer()
			start := time.Now()

			for i := 0; i < b.N; i++ {
				_, err := parser.Parse("benchmark.omni", source)
				if err != nil {
					b.Fatalf("Parse failed: %v", err)
				}
			}

			duration := time.Since(start)
			runtime.ReadMemStats(&memStats)
			endMem := memStats.Alloc

			// Record metrics
			testName := fmt.Sprintf("parser_size_%d", size)
			suite.metrics.Tests[testName] = TestMetrics{
				Name:        testName,
				Duration:    duration,
				MemoryUsage: endMem - startMem,
				Iterations:  b.N,
				Throughput:  float64(b.N) / duration.Seconds(),
			}
		})
	}

	// Save metrics
	suite.saveMetrics("parser_performance.json")
}

// BenchmarkTypeCheckerPerformance benchmarks type checker performance
func BenchmarkTypeCheckerPerformance(b *testing.B) {
	suite := NewPerformanceTestSuite()

	// Test with different complexity levels
	complexities := []int{10, 50, 100, 200}

	for _, complexity := range complexities {
		b.Run(fmt.Sprintf("complexity_%d", complexity), func(b *testing.B) {
			source := generateComplexSource(complexity)

			mod, err := parser.Parse("benchmark.omni", source)
			if err != nil {
				b.Fatalf("Parse failed: %v", err)
			}

			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			startMem := memStats.Alloc

			b.ResetTimer()
			start := time.Now()

			for i := 0; i < b.N; i++ {
				err := checker.Check("benchmark.omni", source, mod)
				if err != nil {
					b.Fatalf("Type check failed: %v", err)
				}
			}

			duration := time.Since(start)
			runtime.ReadMemStats(&memStats)
			endMem := memStats.Alloc

			// Record metrics
			testName := fmt.Sprintf("typechecker_complexity_%d", complexity)
			suite.metrics.Tests[testName] = TestMetrics{
				Name:        testName,
				Duration:    duration,
				MemoryUsage: endMem - startMem,
				Iterations:  b.N,
				Throughput:  float64(b.N) / duration.Seconds(),
			}
		})
	}

	// Save metrics
	suite.saveMetrics("typechecker_performance.json")
}

// BenchmarkCompilationPerformance benchmarks full compilation performance
func BenchmarkCompilationPerformance(b *testing.B) {
	suite := NewPerformanceTestSuite()

	// Test different backends (only VM for now due to C backend issues)
	backends := []string{"vm"}
	optimizationLevels := []string{"O0", "O1", "O2", "O3"}

	for _, backend := range backends {
		for _, optLevel := range optimizationLevels {
			b.Run(fmt.Sprintf("%s_%s", backend, optLevel), func(b *testing.B) {
				source := generateMediumComplexitySource()

				tmpFile := filepath.Join(b.TempDir(), "benchmark.omni")
				err := os.WriteFile(tmpFile, []byte(source), 0644)
				if err != nil {
					b.Fatalf("Failed to write test file: %v", err)
				}

				var memStats runtime.MemStats
				runtime.ReadMemStats(&memStats)
				startMem := memStats.Alloc

				b.ResetTimer()
				start := time.Now()

				for i := 0; i < b.N; i++ {
					emitType := "mir"
					if backend == "c" {
						emitType = "exe"
					}
					cfg := Config{
						InputPath: tmpFile,
						Backend:   backend,
						OptLevel:  optLevel,
						Emit:      emitType,
					}
					err := Compile(cfg)
					if err != nil {
						b.Fatalf("Compilation failed: %v", err)
					}
				}

				duration := time.Since(start)
				runtime.ReadMemStats(&memStats)
				endMem := memStats.Alloc

				// Record metrics
				testName := fmt.Sprintf("compilation_%s_%s", backend, optLevel)
				suite.metrics.Tests[testName] = TestMetrics{
					Name:        testName,
					Duration:    duration,
					MemoryUsage: endMem - startMem,
					Iterations:  b.N,
					Throughput:  float64(b.N) / duration.Seconds(),
				}
			})
		}
	}

	// Save metrics
	suite.saveMetrics("compilation_performance.json")
}

// BenchmarkMIRGeneration benchmarks MIR generation performance
func BenchmarkMIRGeneration(b *testing.B) {
	suite := NewPerformanceTestSuite()

	// Test with different function counts
	functionCounts := []int{10, 50, 100, 200}

	for _, count := range functionCounts {
		b.Run(fmt.Sprintf("functions_%d", count), func(b *testing.B) {
			source := generateFunctionHeavySource(count)

			mod, err := parser.Parse("benchmark.omni", source)
			if err != nil {
				b.Fatalf("Parse failed: %v", err)
			}

			err = checker.Check("benchmark.omni", source, mod)
			if err != nil {
				b.Fatalf("Type check failed: %v", err)
			}

			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			startMem := memStats.Alloc

			b.ResetTimer()
			start := time.Now()

			for i := 0; i < b.N; i++ {
				_, err := builder.BuildModule(mod)
				if err != nil {
					b.Fatalf("MIR generation failed: %v", err)
				}
			}

			duration := time.Since(start)
			runtime.ReadMemStats(&memStats)
			endMem := memStats.Alloc

			// Record metrics
			testName := fmt.Sprintf("mir_generation_functions_%d", count)
			suite.metrics.Tests[testName] = TestMetrics{
				Name:        testName,
				Duration:    duration,
				MemoryUsage: endMem - startMem,
				Iterations:  b.N,
				Throughput:  float64(b.N) / duration.Seconds(),
			}
		})
	}

	// Save metrics
	suite.saveMetrics("mir_generation_performance.json")
}

// BenchmarkMemoryUsage benchmarks memory usage patterns
func BenchmarkMemoryUsage(b *testing.B) {
	suite := NewPerformanceTestSuite()

	// Test memory usage with different source sizes
	sizes := []int{1000, 5000, 10000, 20000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("source_size_%d", size), func(b *testing.B) {
			source := generateLargeSource(size)

			var memStats runtime.MemStats
			runtime.GC() // Force garbage collection
			runtime.ReadMemStats(&memStats)
			startMem := memStats.Alloc

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				mod, err := parser.Parse("benchmark.omni", source)
				if err != nil {
					b.Fatalf("Parse failed: %v", err)
				}

				err = checker.Check("benchmark.omni", source, mod)
				if err != nil {
					b.Fatalf("Type check failed: %v", err)
				}

				mirMod, err := builder.BuildModule(mod)
				if err != nil {
					b.Fatalf("MIR generation failed: %v", err)
				}

				_ = mirMod // Use the result to prevent optimization
			}

			runtime.GC() // Force garbage collection
			runtime.ReadMemStats(&memStats)
			endMem := memStats.Alloc

			// Record metrics
			testName := fmt.Sprintf("memory_usage_size_%d", size)
			suite.metrics.Tests[testName] = TestMetrics{
				Name:        testName,
				Duration:    time.Duration(0), // Not timing-based
				MemoryUsage: endMem - startMem,
				Iterations:  b.N,
				Throughput:  0, // Not applicable for memory tests
			}
		})
	}

	// Save metrics
	suite.saveMetrics("memory_usage_performance.json")
}

// Helper functions for generating test sources

func generateLargeSource(size int) string {
	var source strings.Builder

	// Generate simple functions that just return values
	for i := 0; i < size/10; i++ {
		source.WriteString(fmt.Sprintf("func func%d(x:int, y:int) : int {\n", i))
		source.WriteString("    return x + y\n")
		source.WriteString("}\n\n")
	}

	source.WriteString("func main() : int {\n")
	source.WriteString("    var sum:int = 0\n")
	for i := 0; i < size/20; i++ {
		source.WriteString(fmt.Sprintf("    sum = sum + func%d(1, 2)\n", i))
	}
	source.WriteString("    return sum\n")
	source.WriteString("}\n")

	return source.String()
}

func generateComplexSource(complexity int) string {
	var source strings.Builder

	// Generate simple functions with basic arithmetic
	for i := 0; i < complexity; i++ {
		source.WriteString(fmt.Sprintf("func complex%d(x:int, y:int, z:int) : int {\n", i))
		source.WriteString("    return x + y + z\n")
		source.WriteString("}\n\n")
	}

	source.WriteString("func main() : int {\n")
	source.WriteString("    var result:int = 0\n")
	for i := 0; i < complexity/2; i++ {
		source.WriteString(fmt.Sprintf("    result = result + complex%d(%d, %d, %d)\n", i, i, i+1, i+2))
	}
	source.WriteString("    return result\n")
	source.WriteString("}\n")

	return source.String()
}

func generateMediumComplexitySource() string {
	return `
func fibonacci(n:int) : int {
    if n <= 1 {
        return n
    }
    return fibonacci(n - 1) + fibonacci(n - 2)
}

func main() : int {
    let result:int = fibonacci(10)
    return result
}
`
}

func generateFunctionHeavySource(count int) string {
	var source strings.Builder

	// Generate many small functions
	for i := 0; i < count; i++ {
		source.WriteString(fmt.Sprintf("func func%d(x:int, y:int) : int {\n", i))
		source.WriteString("    let result:int = x + y\n")
		source.WriteString("    return result * 2\n")
		source.WriteString("}\n\n")
	}

	source.WriteString("func main() : int {\n")
	source.WriteString("    var sum:int = 0\n")
	for i := 0; i < count/2; i++ {
		source.WriteString(fmt.Sprintf("    sum = sum + func%d(1, 2)\n", i))
	}
	source.WriteString("    return sum\n")
	source.WriteString("}\n")

	return source.String()
}

// Performance regression testing methods

func (suite *PerformanceTestSuite) saveMetrics(filename string) error {
	data, err := json.MarshalIndent(suite.metrics, "", "  ")
	if err != nil {
		return err
	}

	// Create performance directory if it doesn't exist
	perfDir := "performance"
	if err := os.MkdirAll(perfDir, 0755); err != nil {
		return err
	}

	filepath := filepath.Join(perfDir, filename)
	return os.WriteFile(filepath, data, 0644)
}

func (suite *PerformanceTestSuite) loadBaseline(filename string) error {
	filepath := filepath.Join("performance", filename)
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &suite.baseline)
}

func (suite *PerformanceTestSuite) detectRegressions() []string {
	var regressions []string

	if suite.baseline == nil {
		return regressions
	}

	for testName, current := range suite.metrics.Tests {
		baseline, exists := suite.baseline.Tests[testName]
		if !exists {
			continue
		}

		// Calculate performance regression
		regression := (current.Duration.Seconds() - baseline.Duration.Seconds()) / baseline.Duration.Seconds()

		// Get threshold for this test type
		testType := strings.Split(testName, "_")[0]
		threshold, exists := suite.thresholds[testType]
		if !exists {
			threshold = 0.20 // Default 20% threshold
		}

		if regression > threshold {
			regressions = append(regressions, fmt.Sprintf(
				"Performance regression detected in %s: %.2f%% slower (threshold: %.2f%%)",
				testName, regression*100, threshold*100,
			))
		}
	}

	return regressions
}

// TestPerformanceRegression tests for performance regressions
func TestPerformanceRegression(t *testing.T) {
	suite := NewPerformanceTestSuite()

	// Load baseline metrics
	if err := suite.loadBaseline("baseline_performance.json"); err != nil {
		t.Skipf("No baseline metrics found: %v", err)
	}

	// Run a subset of performance tests
	t.Run("parser_regression", func(t *testing.T) {
		source := generateMediumComplexitySource()

		start := time.Now()
		for i := 0; i < 100; i++ {
			_, err := parser.Parse("benchmark.omni", source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
		}
		duration := time.Since(start)

		suite.metrics.Tests["parser_regression"] = TestMetrics{
			Name:        "parser_regression",
			Duration:    duration,
			MemoryUsage: 0,
			Iterations:  100,
			Throughput:  100 / duration.Seconds(),
		}
	})

	// Detect regressions
	regressions := suite.detectRegressions()
	if len(regressions) > 0 {
		t.Errorf("Performance regressions detected:\n%s", strings.Join(regressions, "\n"))
	}

	// Save current metrics
	suite.saveMetrics("current_performance.json")
}
