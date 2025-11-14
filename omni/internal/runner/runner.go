package runner

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/omni-lang/omni/internal/compiler"
	"github.com/omni-lang/omni/internal/logging"
	"github.com/omni-lang/omni/internal/mir/builder"
	"github.com/omni-lang/omni/internal/parser"
	"github.com/omni-lang/omni/internal/passes"
	"github.com/omni-lang/omni/internal/types/checker"
	"github.com/omni-lang/omni/internal/vm"
)

// Execute compiles and executes the provided OmniLang source via the VM backend.
func Execute(path string, args []string, verbose bool) (vm.Result, error) {
	if filepath.Ext(path) != ".omni" {
		return vm.Result{}, fmt.Errorf("%s: unsupported input (expected .omni)", path)
	}

	vm.SetCLIArgs(args)
	defer vm.SetCLIArgs(nil)

	vm.SetCLIArgs(args)
	defer vm.SetCLIArgs(nil)

	logger := logging.Logger()

	if verbose {
		logger.DebugFields("Running program", logging.String("path", path))
	}

	src, err := os.ReadFile(path)
	if err != nil {
		return vm.Result{}, fmt.Errorf("read source: %w", err)
	}

	if verbose {
		logger.DebugString("Parsing source...")
	}
	mod, err := parser.Parse(path, string(src))
	if err != nil {
		return vm.Result{}, err
	}

	if verbose {
		logger.DebugString("Merging imported modules...")
	}
	// Merge locally imported modules' functions into the main module
	if err := compiler.MergeImportedModules(mod, filepath.Dir(path), false, "vm"); err != nil {
		return vm.Result{}, err
	}

	if verbose {
		logger.DebugString("Type checking...")
	}
	if err := checker.Check(path, string(src), mod); err != nil {
		return vm.Result{}, err
	}

	if verbose {
		logger.DebugString("Building MIR...")
	}
	mirModule, err := builder.BuildModule(mod)
	if err != nil {
		return vm.Result{}, err
	}

	if verbose {
		logger.DebugString("Running optimization passes...")
	}
	pipeline := passes.NewPipeline("runner")
	if _, err := pipeline.Run(*mirModule); err != nil {
		return vm.Result{}, err
	}

	if verbose {
		logger.DebugString("Executing program...")
	}
	result, err := vm.Execute(mirModule, "main")
	if err != nil {
		return vm.Result{}, err
	}

	if verbose {
		logger.DebugString("Execution completed!")
	}
	return result, nil
}

// Run wraps Execute and prints the result to stdout for CLI usage.
func Run(path string, args []string, verbose bool) error {
	result, err := Execute(path, args, verbose)
	if err != nil {
		var exitErr vm.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.Code)
		}
		return err
	}
	switch v := result.Value.(type) {
	case int:
		fmt.Println(v)
	case bool:
		if v {
			fmt.Println(1)
		} else {
			fmt.Println(0)
		}
	case string:
		fmt.Println(v)
	}
	return nil
}
