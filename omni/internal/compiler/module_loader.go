package compiler

import (
	"fmt"
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

// NewModuleLoader creates a new module loader with default search paths.
func NewModuleLoader() *ModuleLoader {
	return &ModuleLoader{
		cache:       make(map[string]*ast.Module),
		searchPaths: []string{".", "examples", "std"},
	}
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
	return "", fmt.Errorf("module %s not found in search paths: [%s]\n  Hint: Make sure the file %s exists in one of these directories",
		strings.Join(importPath, "."), searchPathsStr, fileName)
}

// AddSearchPath adds a directory to the module search paths.
func (ml *ModuleLoader) AddSearchPath(path string) {
	ml.searchPaths = append(ml.searchPaths, path)
}
