package compiler

import (
	"github.com/omni-lang/omni/internal/moduleloader"
)

// Re-export the ModuleLoader from the moduleloader package
type ModuleLoader = moduleloader.ModuleLoader

// NewModuleLoader creates a new module loader with improved search paths.
func NewModuleLoader() *ModuleLoader {
	return moduleloader.NewModuleLoader()
}
