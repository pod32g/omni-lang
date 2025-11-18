package coverage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// CoverageEntry represents a single coverage entry from the coverage JSON
type CoverageEntry struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Count    int    `json:"count"`
}

// CoverageData represents the full coverage data structure
type CoverageData struct {
	Entries []CoverageEntry `json:"entries"`
}

// FunctionInfo represents information about a function in a std library file
type FunctionInfo struct {
	Name       string
	File       string
	LineNumber int
	IsWired    bool
	RuntimeFn  *RuntimeFunction
}

// ParseCoverageFile parses a coverage JSON file
func ParseCoverageFile(path string) (*CoverageData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read coverage file: %w", err)
	}

	var coverage CoverageData
	if err := json.Unmarshal(data, &coverage); err != nil {
		return nil, fmt.Errorf("parse coverage JSON: %w", err)
	}

	return &coverage, nil
}

// ParseStdLibrary parses std library .omni files to extract function information
func ParseStdLibrary(stdPath string) (map[string][]FunctionInfo, error) {
	funcsByFile := make(map[string][]FunctionInfo)
	runtimeFuncs := GetRuntimeWiredFunctions()

	err := filepath.Walk(stdPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".omni") {
			return nil
		}

		funcs, err := parseOmniFile(path, runtimeFuncs)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}

		if len(funcs) > 0 {
			funcsByFile[path] = funcs
		}

		return nil
	})

	return funcsByFile, err
}

// parseOmniFile parses a single .omni file to extract function definitions
func parseOmniFile(path string, runtimeFuncs map[string]*RuntimeFunction) ([]FunctionInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	var funcs []FunctionInfo
	funcRegex := regexp.MustCompile(`^func\s+(\w+)\s*\(`)

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Skip comments and empty lines
		if strings.HasPrefix(trimmed, "//") || trimmed == "" {
			continue
		}

		// Check for function definition
		if matches := funcRegex.FindStringSubmatch(trimmed); matches != nil {
			funcName := matches[1]

			// Try to determine full qualified name
			// This is a simplified approach - in reality we'd need to parse imports
			qualifiedName := determineQualifiedName(path, funcName)

			// Check if this function is runtime-wired
			runtimeFn := runtimeFuncs[qualifiedName]
			isWired := runtimeFn != nil

			funcs = append(funcs, FunctionInfo{
				Name:       qualifiedName,
				File:       path,
				LineNumber: lineNum,
				IsWired:    isWired,
				RuntimeFn:  runtimeFn,
			})
		}
	}

	return funcs, nil
}

// determineQualifiedName attempts to determine the fully qualified function name
// This is a simplified implementation - a full parser would be more accurate
func determineQualifiedName(filePath, funcName string) string {
	// Extract module name from file path
	// e.g., std/io/print.omni -> std.io.print
	relPath := filePath
	if strings.Contains(relPath, "std/") {
		parts := strings.Split(relPath, "std/")
		if len(parts) > 1 {
			modulePath := parts[1]
			modulePath = strings.TrimSuffix(modulePath, ".omni")
			modulePath = strings.ReplaceAll(modulePath, "/", ".")

			// Handle special cases
			if modulePath == "io.print" || modulePath == "print" {
				return fmt.Sprintf("std.io.%s", funcName)
			}
			if modulePath == "string.string" || modulePath == "string" {
				return fmt.Sprintf("std.string.%s", funcName)
			}
			if modulePath == "math.math" || modulePath == "math" {
				return fmt.Sprintf("std.math.%s", funcName)
			}

			return fmt.Sprintf("std.%s.%s", modulePath, funcName)
		}
	}

	// Fallback: try to infer from common patterns
	if strings.Contains(filePath, "io/") {
		return fmt.Sprintf("std.io.%s", funcName)
	}
	if strings.Contains(filePath, "string/") {
		return fmt.Sprintf("std.string.%s", funcName)
	}
	if strings.Contains(filePath, "math/") {
		return fmt.Sprintf("std.math.%s", funcName)
	}
	if strings.Contains(filePath, "array/") {
		return fmt.Sprintf("std.array.%s", funcName)
	}
	if strings.Contains(filePath, "file/") {
		return fmt.Sprintf("std.file.%s", funcName)
	}
	if strings.Contains(filePath, "os/") {
		return fmt.Sprintf("std.os.%s", funcName)
	}
	if strings.Contains(filePath, "collections/") {
		return fmt.Sprintf("std.collections.%s", funcName)
	}

	return funcName
}

// MatchCoverageToFunctions matches coverage data to function definitions
func MatchCoverageToFunctions(coverage *CoverageData, funcsByFile map[string][]FunctionInfo) map[string]*CoverageMatch {
	matches := make(map[string]*CoverageMatch)

	// Create a map of covered functions
	covered := make(map[string]bool)
	for _, entry := range coverage.Entries {
		key := entry.Function
		covered[key] = true
	}

	// Match functions to coverage
	for _, funcs := range funcsByFile {
		for _, fn := range funcs {
			if !fn.IsWired {
				continue // Only track runtime-wired functions
			}

			match := &CoverageMatch{
				Function: fn,
				Covered:  covered[fn.Name],
			}

			if match.Covered {
				// Find the coverage entry
				for _, entry := range coverage.Entries {
					if entry.Function == fn.Name {
						match.CallCount = entry.Count
						break
					}
				}
			}

			matches[fn.Name] = match
		}
	}

	return matches
}

// CoverageMatch represents a match between a function definition and coverage data
type CoverageMatch struct {
	Function  FunctionInfo
	Covered   bool
	CallCount int
}

// CalculateCoverage calculates coverage statistics
func CalculateCoverage(matches map[string]*CoverageMatch) CoverageStats {
	stats := CoverageStats{
		TotalFunctions:   len(matches),
		CoveredFunctions: 0,
		TotalLines:       0,
		CoveredLines:     0,
		FunctionDetails:  make(map[string]FunctionCoverage),
	}

	for name, match := range matches {
		stats.TotalLines++
		if match.Covered {
			stats.CoveredFunctions++
			stats.CoveredLines++
		}

		stats.FunctionDetails[name] = FunctionCoverage{
			Function:  match.Function.Name,
			File:      match.Function.File,
			Line:      match.Function.LineNumber,
			Covered:   match.Covered,
			CallCount: match.CallCount,
		}
	}

	return stats
}

// CoverageStats represents overall coverage statistics
type CoverageStats struct {
	TotalFunctions   int
	CoveredFunctions int
	TotalLines       int
	CoveredLines     int
	FunctionDetails  map[string]FunctionCoverage
}

// FunctionCoverage represents coverage for a single function
type FunctionCoverage struct {
	Function  string
	File      string
	Line      int
	Covered   bool
	CallCount int
}

// GetFunctionCoveragePercentage returns the percentage of functions covered
func (s CoverageStats) GetFunctionCoveragePercentage() float64 {
	if s.TotalFunctions == 0 {
		return 0.0
	}
	return float64(s.CoveredFunctions) / float64(s.TotalFunctions) * 100.0
}

// GetLineCoveragePercentage returns the percentage of lines covered
func (s CoverageStats) GetLineCoveragePercentage() float64 {
	if s.TotalLines == 0 {
		return 0.0
	}
	return float64(s.CoveredLines) / float64(s.TotalLines) * 100.0
}
