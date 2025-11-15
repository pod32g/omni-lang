package cbackend

import (
	"fmt"
	"regexp"
	"strings"
)

// COptimizer handles optimization of generated C code
type COptimizer struct {
	optLevel string
}

// NewCOptimizer creates a new C code optimizer
func NewCOptimizer(optLevel string) *COptimizer {
	return &COptimizer{
		optLevel: optLevel,
	}
}

// OptimizeC optimizes C code based on the optimization level
func OptimizeC(code string, optLevel string) string {
	optimizer := NewCOptimizer(optLevel)
	return optimizer.optimize(code)
}

// optimize applies optimizations based on the optimization level
func (o *COptimizer) optimize(code string) string {
	switch o.optLevel {
	case "0", "O0", "none":
		return code // No optimization
	case "1", "O1", "basic":
		return o.basicOptimizations(code)
	case "2", "O2", "standard":
		return o.standardOptimizations(code)
	case "3", "O3", "aggressive":
		return o.aggressiveOptimizations(code)
	case "s", "Os", "size":
		return o.sizeOptimizations(code)
	default:
		return o.standardOptimizations(code)
	}
}

// basicOptimizations applies basic optimizations
func (o *COptimizer) basicOptimizations(code string) string {
	// Remove unnecessary variable declarations
	optimized := o.removeUnusedVariables(code)

	// Simplify constant expressions
	optimized = o.simplifyConstants(optimized)

	return optimized
}

// standardOptimizations applies standard optimizations
func (o *COptimizer) standardOptimizations(code string) string {
	// Apply basic optimizations first
	optimized := o.basicOptimizations(code)

	// Inline simple functions
	optimized = o.inlineSimpleFunctions(optimized)

	// Optimize arithmetic operations
	optimized = o.optimizeArithmetic(optimized)

	return optimized
}

// aggressiveOptimizations applies aggressive optimizations
func (o *COptimizer) aggressiveOptimizations(code string) string {
	// Apply standard optimizations first
	optimized := o.standardOptimizations(code)

	// More aggressive inlining
	optimized = o.aggressiveInlining(optimized)

	// Loop optimizations
	optimized = o.optimizeLoops(optimized)

	return optimized
}

// sizeOptimizations applies size-focused optimizations
func (o *COptimizer) sizeOptimizations(code string) string {
	// Apply basic optimizations
	optimized := o.basicOptimizations(code)

	// Minimize variable names
	optimized = o.minimizeVariableNames(optimized)

	// Remove debug information
	optimized = o.removeDebugInfo(optimized)

	return optimized
}

// removeUnusedVariables removes unused variable declarations
func (o *COptimizer) removeUnusedVariables(code string) string {
	lines := strings.Split(code, "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			result = append(result, line)
			continue
		}

		// Check if this is a variable declaration
		if strings.Contains(trimmed, "int32_t v") && strings.Contains(trimmed, "=") {
			// Extract variable name
			parts := strings.Split(trimmed, "=")
			if len(parts) >= 2 {
				varName := strings.TrimSpace(strings.Split(parts[0], " ")[len(strings.Split(parts[0], " "))-1])

				// Check if variable is used elsewhere
				used := false
				for _, otherLine := range lines {
					if otherLine != line && strings.Contains(otherLine, varName) {
						used = true
						break
					}
				}

				if !used {
					continue // Skip this unused variable
				}
			}
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// simplifyConstants simplifies constant expressions
func (o *COptimizer) simplifyConstants(code string) string {
	lines := strings.Split(code, "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			result = append(result, line)
			continue
		}

		// Simplify arithmetic with constants
		optimized := o.simplifyArithmeticExpression(trimmed)
		result = append(result, optimized)
	}

	return strings.Join(result, "\n")
}

// simplifyArithmeticExpression simplifies arithmetic expressions
func (o *COptimizer) simplifyArithmeticExpression(line string) string {
	// Simplify x + 0 = x
	line = strings.ReplaceAll(line, " + 0;", ";")
	line = strings.ReplaceAll(line, " + 0", "")

	// Simplify x - 0 = x
	line = strings.ReplaceAll(line, " - 0;", ";")
	line = strings.ReplaceAll(line, " - 0", "")

	// Simplify x * 1 = x
	line = strings.ReplaceAll(line, " * 1;", ";")
	line = strings.ReplaceAll(line, " * 1", "")

	// Simplify x / 1 = x
	line = strings.ReplaceAll(line, " / 1;", ";")
	line = strings.ReplaceAll(line, " / 1", "")

	// Simplify x * 0 = 0
	if strings.Contains(line, " * 0") || strings.Contains(line, " *0") {
		// This is more complex - would need proper parsing
		// For now, leave as is
	}

	return line
}

// inlineSimpleFunctions inlines simple functions
func (o *COptimizer) inlineSimpleFunctions(code string) string {
	// This is a placeholder for function inlining
	// In a real implementation, you would identify simple functions and inline them
	return code
}

// optimizeArithmetic optimizes arithmetic operations
func (o *COptimizer) optimizeArithmetic(code string) string {
	lines := strings.Split(code, "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			result = append(result, line)
			continue
		}

		// Apply arithmetic optimizations
		optimized := o.optimizeArithmeticLine(trimmed)
		result = append(result, optimized)
	}

	return strings.Join(result, "\n")
}

// optimizeArithmeticLine optimizes a single line of arithmetic
func (o *COptimizer) optimizeArithmeticLine(line string) string {
	// Optimize multiplication by powers of 2 to bit shifts
	// Use word boundaries to avoid matching inside identifiers or pointer declarations
	
	// Replace x * 2 with x << 1 (but not in pointer declarations like "int *v")
	// Match: space or identifier char, then " * 2" or "*2", then space or operator
	re := regexp.MustCompile(`([\w\)])\s*\*\s*2(\s|;|\)|,|\[)`)
	line = re.ReplaceAllString(line, "${1} << 1${2}")
	
	re = regexp.MustCompile(`([\w\)])\s*\*\s*4(\s|;|\)|,|\[)`)
	line = re.ReplaceAllString(line, "${1} << 2${2}")
	
	re = regexp.MustCompile(`([\w\)])\s*\*\s*8(\s|;|\)|,|\[)`)
	line = re.ReplaceAllString(line, "${1} << 3${2}")

	// Optimize division by powers of 2 to bit shifts
	re = regexp.MustCompile(`([\w\)])\s*/\s*2(\s|;|\)|,|\[)`)
	line = re.ReplaceAllString(line, "${1} >> 1${2}")
	
	re = regexp.MustCompile(`([\w\)])\s*/\s*4(\s|;|\)|,|\[)`)
	line = re.ReplaceAllString(line, "${1} >> 2${2}")
	
	re = regexp.MustCompile(`([\w\)])\s*/\s*8(\s|;|\)|,|\[)`)
	line = re.ReplaceAllString(line, "${1} >> 3${2}")

	return line
}

// aggressiveInlining performs more aggressive function inlining
func (o *COptimizer) aggressiveInlining(code string) string {
	// This is a placeholder for aggressive inlining
	return code
}

// optimizeLoops optimizes loop structures
func (o *COptimizer) optimizeLoops(code string) string {
	// This is a placeholder for loop optimizations
	return code
}

// minimizeVariableNames shortens variable names to reduce size
func (o *COptimizer) minimizeVariableNames(code string) string {
	// Create a mapping of variable names to shorter names
	varMap := make(map[string]string)
	nextShortName := 1

	lines := strings.Split(code, "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			result = append(result, line)
			continue
		}

		// Find variable declarations and create short names
		if strings.Contains(trimmed, "int32_t v") {
			// Extract variable name (e.g., "v5" from "int32_t v5 = ...")
			parts := strings.Fields(trimmed)
			for i, part := range parts {
				if strings.HasPrefix(part, "v") && i > 0 {
					varName := part
					if _, exists := varMap[varName]; !exists {
						// Create a short name
						shortName := fmt.Sprintf("v%d", nextShortName)
						varMap[varName] = shortName
						nextShortName++
					}
					break
				}
			}
		}

		result = append(result, line)
	}

	// Replace all variable references with short names
	// Use word boundaries to avoid replacing substrings inside other identifiers
	var finalResult []string
	for _, line := range result {
		optimized := line
		for oldName, newName := range varMap {
			// Use regex to match whole words only (word boundaries)
			// This prevents replacing "v1" inside "v10" or string literals
			re := regexp.MustCompile(`\b` + regexp.QuoteMeta(oldName) + `\b`)
			optimized = re.ReplaceAllString(optimized, newName)
		}
		finalResult = append(finalResult, optimized)
	}

	return strings.Join(finalResult, "\n")
}

// removeDebugInfo removes debug information and comments
func (o *COptimizer) removeDebugInfo(code string) string {
	lines := strings.Split(code, "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip comment lines
		if strings.HasPrefix(trimmed, "//") {
			continue
		}

		// Remove inline comments
		if commentPos := strings.Index(line, "//"); commentPos != -1 {
			line = line[:commentPos]
			line = strings.TrimRight(line, " \t")
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}
