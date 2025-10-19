package compiler

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/omni-lang/omni/internal/parser"
)

// Config captures the minimal inputs needed to drive the compilation pipeline.
type Config struct {
	InputPath  string
	OutputPath string
	Backend    string
	OptLevel   string
	Emit       string
	Dump       string
}

// ErrNotImplemented indicates that a requested stage has not yet been implemented.
var ErrNotImplemented = errors.New("not implemented")

// Compile wires together the compiler pipeline. It currently serves as a thin
// placeholder until the real frontend, midend and backend are ready.
func Compile(cfg Config) error {
	if cfg.InputPath == "" {
		return fmt.Errorf("input path required")
	}

	if cfg.Backend != "vm" && cfg.Backend != "clift" {
		return fmt.Errorf("unsupported backend: %s", cfg.Backend)
	}

	if cfg.Emit != "mir" && cfg.Emit != "obj" && cfg.Emit != "asm" {
		return fmt.Errorf("unsupported emit option: %s", cfg.Emit)
	}

	if cfg.OutputPath != "" {
		if ext := filepath.Ext(cfg.OutputPath); ext == "" {
			return fmt.Errorf("output path must include file extension")
		}
	}

	src, err := os.ReadFile(cfg.InputPath)
	if err != nil {
		return fmt.Errorf("read input %s: %w", cfg.InputPath, err)
	}

	if _, err := parser.Parse(cfg.InputPath, string(src)); err != nil {
		return err
	}

	return ErrNotImplemented
}
