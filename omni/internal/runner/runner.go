package runner

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/omni-lang/omni/internal/compiler"
	"github.com/omni-lang/omni/internal/mir/builder"
	"github.com/omni-lang/omni/internal/parser"
	"github.com/omni-lang/omni/internal/passes"
	"github.com/omni-lang/omni/internal/types/checker"
	"github.com/omni-lang/omni/internal/vm"
)

// Execute compiles and executes the provided OmniLang source via the VM backend.
func Execute(path string, verbose bool) (vm.Result, error) {
	if filepath.Ext(path) != ".omni" {
		return vm.Result{}, fmt.Errorf("%s: unsupported input (expected .omni)", path)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Running %s...\n", path)
	}

	src, err := os.ReadFile(path)
	if err != nil {
		return vm.Result{}, fmt.Errorf("read source: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "  Parsing source...\n")
	}
	mod, err := parser.Parse(path, string(src))
	if err != nil {
		return vm.Result{}, err
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "  Merging imported modules...\n")
	}
	// Merge locally imported modules' functions into the main module
	if err := compiler.MergeImportedModules(mod, filepath.Dir(path)); err != nil {
		return vm.Result{}, err
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "  Type checking...\n")
	}
	if err := checker.Check(path, string(src), mod); err != nil {
		return vm.Result{}, err
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "  Building MIR...\n")
	}
	mirModule, err := builder.BuildModule(mod)
	if err != nil {
		return vm.Result{}, err
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "  Running optimization passes...\n")
	}
	pipeline := passes.NewPipeline("runner")
	if _, err := pipeline.Run(*mirModule); err != nil {
		return vm.Result{}, err
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "  Executing program...\n")
	}
	result, err := vm.Execute(mirModule, "main")
	if err != nil {
		return vm.Result{}, err
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "  Execution completed!\n")
	}
	return result, nil
}

// Run wraps Execute and prints the result to stdout for CLI usage.
func Run(path string, verbose bool) error {
	result, err := Execute(path, verbose)
	if err != nil {
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
