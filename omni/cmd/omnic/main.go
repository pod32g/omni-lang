package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/omni-lang/omni/internal/compiler"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

type stringFlag struct {
	value string
	set   bool
}

func newStringFlag(defaultValue string) *stringFlag {
	return &stringFlag{value: defaultValue}
}

func (f *stringFlag) String() string {
	return f.value
}

func (f *stringFlag) Set(v string) error {
	f.value = v
	f.set = true
	return nil
}

func main() {
	var (
		backend  = flag.String("backend", "vm", "code generation backend (vm|clift)")
		optLevel = flag.String("O", "O0", "optimization level (O0-O3)")
		emitFlag = newStringFlag("obj")
		dump     = flag.String("dump", "", "dump intermediate representation (mir)")
		output   = flag.String("o", "", "output binary path")
		version  = flag.Bool("version", false, "print version and exit")
	)
	flag.Var(emitFlag, "emit", "emission format (mir|obj|asm)")

	flag.Parse()

	if *version {
		fmt.Printf("omnic %s (built %s)\n", Version, BuildTime)
		os.Exit(0)
	}

	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: omnic [options] <file.omni>")
		flag.PrintDefaults()
		os.Exit(2)
	}

	input := flag.Arg(0)
	emit := emitFlag.value
	if !emitFlag.set && *backend == "vm" {
		emit = "mir"
	}

	if err := run(input, *output, *backend, *optLevel, emit, *dump); err != nil {
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
