package compiler

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/omni-lang/omni/internal/ast"
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

	// Merge locally imported modules' functions into the main module so the VM can resolve them
	if err := MergeImportedModules(mod, filepath.Dir(cfg.InputPath)); err != nil {
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

// MergeImportedModules loads imported local modules and appends their function declarations
// into the root module with namespaced names (aliasOrSegment.funcName) so that calls like
// `math_utils.add` resolve at runtime. std imports are ignored here (handled as intrinsics).
func MergeImportedModules(mod *ast.Module, baseDir string) error {
	loader := NewModuleLoader()
	// Prefer absolute path of the current file directory first to avoid cwd issues
	if baseDir != "" {
		if abs, err := filepath.Abs(baseDir); err == nil {
			loader.AddSearchPath(abs)
		} else if baseDir != "." {
			loader.AddSearchPath(baseDir)
		}
	}

	// Collect imports from both Module.Imports and top-level decls
	imports := make([]*ast.ImportDecl, 0, len(mod.Imports))
	imports = append(imports, mod.Imports...)
	for _, d := range mod.Decls {
		if imp, ok := d.(*ast.ImportDecl); ok {
			imports = append(imports, imp)
		}
	}

	for _, imp := range imports {
		if len(imp.Path) == 0 {
			continue
		}
		// Skip std imports (handled elsewhere)
		if imp.Path[0] == "std" {
			continue
		}
		imported, err := loader.LoadModule(imp.Path)
		if err != nil {
			return fmt.Errorf("load import %s: %w", strings.Join(imp.Path, "."), err)
		}
		local := imp.Alias
		if local == "" {
			local = imp.Path[len(imp.Path)-1]
		}
		// Append cloned function decls with namespaced names
		for _, d := range imported.Decls {
			if fn, ok := d.(*ast.FuncDecl); ok {
				cloned := *fn
				cloned.Name = local + "." + fn.Name
				mod.Decls = append(mod.Decls, &cloned)
			}
		}
	}
	return nil
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
