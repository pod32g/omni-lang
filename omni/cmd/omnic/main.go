package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/omni-lang/omni/internal/compiler"
)

func main() {
	var (
		backend  = flag.String("backend", "vm", "code generation backend (vm|clift)")
		optLevel = flag.String("O", "O0", "optimization level (O0-O3)")
		emit     = flag.String("emit", "obj", "emission format (mir|obj|asm)")
		dump     = flag.String("dump", "", "dump intermediate representation (mir)")
		output   = flag.String("o", "", "output binary path")
	)

	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: omnic [options] <file.omni>")
		flag.PrintDefaults()
		os.Exit(2)
	}

	input := flag.Arg(0)
	if err := run(input, *output, *backend, *optLevel, *emit, *dump); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(input, output, backend, optLevel, emit, dump string) error {
	if filepath.Ext(input) != ".omni" {
		return fmt.Errorf("%s: unsupported input (expected .omni)", input)
	}

	cfg := compiler.Config{
		InputPath:  input,
		OutputPath: output,
		Backend:    backend,
		OptLevel:   optLevel,
		Emit:       emit,
		Dump:       dump,
	}

	if err := compiler.Compile(cfg); err != nil {
		if errors.Is(err, compiler.ErrNotImplemented) {
			return fmt.Errorf("omnic: feature not implemented: %w", err)
		}
		return err
	}

	return nil
}
