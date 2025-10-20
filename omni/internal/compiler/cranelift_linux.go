//go:build !darwin && !windows
// +build !darwin,!windows

package compiler

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/omni-lang/omni/internal/backend/cranelift"
	"github.com/omni-lang/omni/internal/mir"
)

func compileCranelift(cfg Config, emit string, mod *mir.Module) error {
	output := cfg.OutputPath
	if output == "" {
		output = defaultOutputPath(cfg.InputPath, emit)
	}
	if err := ensureDir(output); err != nil {
		return err
	}

	// Serialize MIR to JSON
	jsonData, err := json.Marshal(mod)
	if err != nil {
		return fmt.Errorf("serialize mir to json: %w", err)
	}

	switch emit {
	case "obj":
		if err := cranelift.CompileToObject(string(jsonData), output); err != nil {
			return fmt.Errorf("cranelift object compilation: %w", err)
		}
	case "asm":
		// For now, generate object file and then disassemble
		objPath := strings.TrimSuffix(output, ".s") + ".o"
		if err := cranelift.CompileToObject(string(jsonData), objPath); err != nil {
			return fmt.Errorf("cranelift object compilation: %w", err)
		}
		// TODO: Add disassembly step
		return fmt.Errorf("assembly output not yet implemented")
	default:
		return fmt.Errorf("unsupported emit format for cranelift backend: %s", emit)
	}

	return nil
}
