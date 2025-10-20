package compiler

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/omni-lang/omni/internal/backend/cranelift"
	"github.com/omni-lang/omni/internal/mir"
	"github.com/omni-lang/omni/internal/mir/builder"
	"github.com/omni-lang/omni/internal/mir/printer"
	"github.com/omni-lang/omni/internal/parser"
	"github.com/omni-lang/omni/internal/passes"
	"github.com/omni-lang/omni/internal/types/checker"
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

	backend := cfg.Backend
	if backend == "" {
		backend = "vm"
	}
	if backend != "vm" && backend != "clift" {
		return fmt.Errorf("unsupported backend: %s", backend)
	}

	emit := cfg.Emit
	if emit == "" {
		if backend == "vm" {
			emit = "mir"
		} else {
			emit = "obj"
		}
	}

	switch backend {
	case "vm":
		if emit != "mir" {
			return fmt.Errorf("vm backend: emit option %q not supported", emit)
		}
	case "clift":
		if emit != "obj" && emit != "asm" {
			return fmt.Errorf("clift backend: emit option %q not supported", emit)
		}
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

	mod, err := parser.Parse(cfg.InputPath, string(src))
	if err != nil {
		return err
	}

	if err := checker.Check(cfg.InputPath, string(src), mod); err != nil {
		return err
	}

	mirMod, err := builder.BuildModule(mod)
	if err != nil {
		return err
	}

	pipeline := passes.NewPipeline("default")
	if _, err := pipeline.Run(*mirMod); err != nil {
		return err
	}

	if cfg.Dump == "mir" {
		fmt.Println(printer.Format(mirMod))
	}

	switch backend {
	case "vm":
		return compileVM(cfg, emit, mirMod)
	case "clift":
		return compileCranelift(cfg, emit, mirMod)
	default:
		return fmt.Errorf("unsupported backend: %s", backend)
	}
}

func compileVM(cfg Config, emit string, mod *mir.Module) error {
	output := cfg.OutputPath
	if output == "" {
		output = defaultOutputPath(cfg.InputPath, emit)
	}
	if err := ensureDir(output); err != nil {
		return err
	}

	rendered := printer.Format(mod)
	if !strings.HasSuffix(rendered, "\n") {
		rendered += "\n"
	}
	if err := os.WriteFile(output, []byte(rendered), 0o644); err != nil {
		return fmt.Errorf("write mir output: %w", err)
	}
	return nil
}

func ensureDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	return nil
}

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
		// For now, generate object file and let user disassemble
		objPath := strings.TrimSuffix(output, ".s") + ".o"
		if err := cranelift.CompileToObject(string(jsonData), objPath); err != nil {
			return fmt.Errorf("cranelift object compilation: %w", err)
		}
		return fmt.Errorf("assembly output not yet implemented")
	default:
		return fmt.Errorf("unsupported emit format for cranelift backend: %s", emit)
	}

	return nil
}

func defaultOutputPath(input, emit string) string {
	base := strings.TrimSuffix(input, filepath.Ext(input))
	switch emit {
	case "mir":
		return base + ".mir"
	case "asm":
		return base + ".s"
	case "obj":
		return base + ".o"
	default:
		return base + ".out"
	}
}
