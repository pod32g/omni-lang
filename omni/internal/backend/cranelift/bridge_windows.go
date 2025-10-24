//go:build windows
// +build windows

package cranelift

import (
	"fmt"

	"github.com/omni-lang/omni/internal/mir"
)

// CompileModuleToObject compiles a MIR module to an object file using Cranelift
func CompileModuleToObject(mod *mir.Module, outputPath string) error {
	return fmt.Errorf("Cranelift backend not yet implemented on Windows")
}

// CompileModuleToObjectWithOpt compiles a MIR module to an object file with optimization
func CompileModuleToObjectWithOpt(mod *mir.Module, outputPath string, optLevel string) error {
	return fmt.Errorf("Cranelift backend not yet implemented on Windows")
}

// CompileToObject compiles a MIR module to an object file
func CompileToObject(mod *mir.Module, outputPath string) error {
	return CompileModuleToObject(mod, outputPath)
}

// CompileToObjectWithOpt compiles a MIR module to an object file with optimization
func CompileToObjectWithOpt(mod *mir.Module, outputPath string, optLevel string) error {
	return CompileModuleToObjectWithOpt(mod, outputPath, optLevel)
}
