package moduleloader

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/omni-lang/omni/internal/ast"
	"github.com/omni-lang/omni/internal/parser"
)

// ModuleLoader handles loading and caching of imported modules.
type ModuleLoader struct {
	// Cache of loaded modules by import path
	cache map[string]*ast.Module
	// Search paths for finding modules
	searchPaths []string
}

// NewModuleLoader creates a new module loader with improved search paths.
// The search paths are now relative to the binary location and support environment variables.
func NewModuleLoader() *ModuleLoader {
	searchPaths := buildSearchPaths()
	return &ModuleLoader{
		cache:       make(map[string]*ast.Module),
		searchPaths: searchPaths,
	}
}

// buildSearchPaths constructs the module search paths in order of precedence:
// 1. Environment variable OMNI_STD_PATH (if set)
// 2. Binary-relative paths (std directory relative to omnic binary)
// 3. Current working directory (for local modules)
func buildSearchPaths() []string {
	var paths []string

	// 1. Check for OMNI_STD_PATH environment variable
	if stdPath := os.Getenv("OMNI_STD_PATH"); stdPath != "" {
		paths = append(paths, stdPath)
	}

	// 2. Add binary-relative paths
	binaryPaths := findBinaryRelativePaths()
	paths = append(paths, binaryPaths...)

	// 3. Add current working directory for local modules
	paths = append(paths, ".")

	return paths
}

// findBinaryRelativePaths finds paths relative to the omnic binary location
func findBinaryRelativePaths() []string {
	var paths []string

	// Get the directory where the omnic binary is located
	execPath, err := os.Executable()
	if err != nil {
		// Fallback to current working directory if we can't determine executable path
		return []string{"."}
	}
	execDir := filepath.Dir(execPath)

	// Check if we're running in a test environment (temporary directory)
	isTestEnv := strings.Contains(execDir, "/tmp/") || strings.Contains(execDir, "go-build")

	if isTestEnv {
		// In test environment, try to find the omni root directory
		// by walking up from the current working directory
		if cwd, err := os.Getwd(); err == nil {
			omniRoot := findOmniRootFromCWD(cwd)
			if omniRoot != "" {
				paths = append(paths, omniRoot)
			}
		}
	} else {
		// In normal operation, use paths relative to the binary
		paths = append(paths, execDir)

		// Also try parent directory (in case binary is in bin/ subdirectory)
		parentDir := filepath.Dir(execDir)
		paths = append(paths, parentDir)
	}

	return paths
}

// findOmniRootFromCWD finds the omni root directory by walking up from current working directory
func findOmniRootFromCWD(cwd string) string {
	current := cwd
	for {
		// Check if this directory contains the std directory with std.omni
		stdPath := filepath.Join(current, "std")
		mainStdPath := filepath.Join(stdPath, "std.omni")
		if _, err := os.Stat(mainStdPath); err == nil {
			return current
		}

		parent := filepath.Dir(current)
		if parent == current {
			break // Reached root
		}
		current = parent
	}
	return ""
}

// LoadModule loads a module by its import path.
func (ml *ModuleLoader) LoadModule(importPath []string) (*ast.Module, error) {
	pathKey := strings.Join(importPath, ".")

	// Check cache first
	if module, exists := ml.cache[pathKey]; exists {
		return module, nil
	}

	// Try to find the module file
	modulePath, err := ml.findModuleFile(importPath)
	if err != nil {
		return nil, err
	}

	// Read and parse the file
	content, err := os.ReadFile(modulePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read module %s: %w", pathKey, err)
	}

	module, err := parser.Parse(modulePath, string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse module %s: %w", pathKey, err)
	}

	// Cache the module
	ml.cache[pathKey] = module
	return module, nil
}

// findModuleFile searches for a module file in the search paths.
func (ml *ModuleLoader) findModuleFile(importPath []string) (string, error) {
	// Convert import path to file path
	// For std.math, we want std/math/math.omni
	// For std.io, we want std/io/print.omni
	// For local modules, we want module.omni

	var fileName string
	if len(importPath) > 0 && importPath[0] == "std" {
		if len(importPath) == 1 {
			// For "import std", look for std/std.omni
			fileName = filepath.Join("std", "std") + ".omni"
		} else {
			// For std modules, use the subdirectory structure
			// std.io -> std/io/print.omni
			// std.math -> std/math/math.omni
			// std.string -> std/string/string.omni
			moduleName := importPath[1] // "io", "math", "string", etc.

			// Special case for std.io which uses print.omni instead of io.omni
			if moduleName == "io" {
				fileName = filepath.Join("std", moduleName, "print") + ".omni"
			} else {
				fileName = filepath.Join("std", moduleName, moduleName) + ".omni"
			}
		}
	} else {
		// For local modules, use the module name directly
		moduleName := importPath[len(importPath)-1]
		fileName = moduleName + ".omni"
	}

	// Search in each search path
	for _, searchPath := range ml.searchPaths {
		fullPath := filepath.Join(searchPath, fileName)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, nil
		}
	}

	// Create a more helpful error message
	searchPathsStr := strings.Join(ml.searchPaths, ", ")
	suggestion := ml.suggestImport(importPath)
	message := fmt.Sprintf("module %s not found in search paths: [%s]\n  Hint: Make sure the file %s exists in one of these directories",
		strings.Join(importPath, "."), searchPathsStr, fileName)
	if suggestion != "" {
		message += "\n  Suggestion: " + suggestion
	}
	return "", fmt.Errorf("%s", message)
}

// AddSearchPath adds a directory to the module search paths.
func (ml *ModuleLoader) AddSearchPath(path string) {
	ml.searchPaths = append(ml.searchPaths, path)
}

// GetSearchPaths returns the current search paths for debugging
func (ml *ModuleLoader) GetSearchPaths() []string {
	return ml.searchPaths
}

// DebugInfo returns debug information about the module loader
func (ml *ModuleLoader) DebugInfo() string {
	var info []string
	info = append(info, "ModuleLoader Debug Info:")
	info = append(info, fmt.Sprintf("  Search paths: %v", ml.searchPaths))

	// Show environment variable status
	if stdPath := os.Getenv("OMNI_STD_PATH"); stdPath != "" {
		info = append(info, fmt.Sprintf("  OMNI_STD_PATH: %s", stdPath))
	} else {
		info = append(info, "  OMNI_STD_PATH: not set")
	}

	// Show executable path
	if execPath, err := os.Executable(); err == nil {
		info = append(info, fmt.Sprintf("  Executable path: %s", execPath))
		info = append(info, fmt.Sprintf("  Executable dir: %s", filepath.Dir(execPath)))
	} else {
		info = append(info, fmt.Sprintf("  Executable path: error getting path: %v", err))
	}

	// Show current working directory
	if cwd, err := os.Getwd(); err == nil {
		info = append(info, fmt.Sprintf("  Current working directory: %s", cwd))
	} else {
		info = append(info, fmt.Sprintf("  Current working directory: error getting cwd: %v", err))
	}

	return strings.Join(info, "\n")
}

func (ml *ModuleLoader) suggestImport(importPath []string) string {
	if len(importPath) == 0 {
		return ""
	}
	if importPath[0] == "std" {
		return suggestStdImport(importPath)
	}
	return suggestLocalImport(importPath, ml.searchPaths)
}

func suggestStdImport(importPath []string) string {
	if len(importPath) < 2 {
		return ""
	}
	moduleName := importPath[1]
	stdModules := []string{"io", "math", "string", "array", "os", "collections", "file", "algorithms", "time", "network", "log"}
	best := ""
	bestDist := math.MaxInt
	for _, candidate := range stdModules {
		dist := levenshtein(moduleName, candidate)
		if dist < bestDist {
			bestDist = dist
			best = candidate
		}
	}
	if best != "" && bestDist <= 2 {
		return fmt.Sprintf("did you mean 'std.%s'?", best)
	}
	return ""
}

func suggestLocalImport(importPath []string, searchPaths []string) string {
	moduleName := importPath[len(importPath)-1]
	best := ""
	bestDist := math.MaxInt
	for _, root := range searchPaths {
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() {
				continue
			}
			if filepath.Ext(name) != ".omni" {
				continue
			}
			candidate := strings.TrimSuffix(name, ".omni")
			dist := levenshtein(moduleName, candidate)
			if dist < bestDist {
				bestDist = dist
				best = candidate
			}
		}
	}
	if best != "" && bestDist <= 2 {
		return fmt.Sprintf("did you mean '%s'?", best)
	}
	return ""
}

func levenshtein(a, b string) int {
	if a == b {
		return 0
	}
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	prev := make([]int, len(b)+1)
	for j := 0; j <= len(b); j++ {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		curr := make([]int, len(b)+1)
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			curr[j] = min3(
				prev[j]+1,
				curr[j-1]+1,
				prev[j-1]+cost,
			)
		}
		prev = curr
	}
	return prev[len(b)]
}

func min3(a, b, c int) int {
	if a <= b && a <= c {
		return a
	}
	if b <= a && b <= c {
		return b
	}
	return c
}
