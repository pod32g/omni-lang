package compiler

import (
	"fmt"

	"github.com/omni-lang/omni/internal/backend/cranelift"
	"github.com/omni-lang/omni/internal/mir"
)

func compileCraneliftBackend(cfg Config, emit string, mod *mir.Module) error {
	output := cfg.OutputPath
	if output == "" {
		output = defaultOutputPath(cfg.InputPath, emit)
	}
	if err := ensureDir(output); err != nil {
		return err
	}

	switch emit {
	case "obj":
		return compileToObject(mod, output)
	case "asm":
		return fmt.Errorf("assembly output not yet implemented")
	default:
		return fmt.Errorf("unsupported emit format: %s", emit)
	}
}

func compileToObject(mod *mir.Module, outputPath string) error {
	// Use the Cranelift backend to compile MIR to object file
	return cranelift.CompileModuleToObject(mod, outputPath)
}
