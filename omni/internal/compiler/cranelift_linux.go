package compiler

import (
	"github.com/omni-lang/omni/internal/backend/cranelift"
	"github.com/omni-lang/omni/internal/mir"
)

// compileToObjectLinux compiles MIR to object file using Cranelift on Linux
func compileToObjectLinux(mod *mir.Module, outputPath string) error {
	// Use the Cranelift backend to compile MIR to object file
	return cranelift.CompileModuleToObject(mod, outputPath)
}
