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
		backend      = flag.String("backend", "c", "code generation backend (vm|clift|c)")
		optLevel     = flag.String("O", "O0", "optimization level (O0-O3)")
		emitFlag     = newStringFlag("exe")
		dump         = flag.String("dump", "", "dump intermediate representation (mir)")
		output       = flag.String("o", "", "output binary path")
		debug        = flag.Bool("debug", false, "generate debug symbols and debug information")
		debugModules = flag.Bool("debug-modules", false, "show module loading debug information")
		version      = flag.Bool("version", false, "print version and exit")
		verbose      = flag.Bool("verbose", false, "enable verbose output")
		help         = flag.Bool("help", false, "show help and exit")
		showHelp     = flag.Bool("h", false, "show help and exit")
	)
	flag.Var(emitFlag, "emit", "emission format (mir|obj|exe|binary|asm)")

	flag.Parse()

	if *version {
		fmt.Printf("omnic %s (built %s)\n", Version, BuildTime)
		os.Exit(0)
	}

	if *help || *showHelp {
		showUsage()
		os.Exit(0)
	}

	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "error: no input file specified")
		fmt.Fprintln(os.Stderr, "")
		showUsage()
		os.Exit(2)
	}

	input := flag.Arg(0)
	emit := emitFlag.value
	if !emitFlag.set {
		// Set appropriate defaults based on backend
		if *backend == "vm" {
			emit = "mir"
		} else if *backend == "c" {
			emit = "exe"
		} else if *backend == "clift" {
			emit = "obj"
		}
	}

	if err := run(input, *output, *backend, *optLevel, emit, *dump, *verbose, *debug, *debugModules); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func showUsage() {
	fmt.Fprintf(os.Stderr, "OmniLang Compiler (omnic) %s\n", Version)
	fmt.Fprintf(os.Stderr, "Built: %s\n\n", BuildTime)
	fmt.Fprintf(os.Stderr, "USAGE:\n")
	fmt.Fprintf(os.Stderr, "  omnic [options] <file.omni>\n\n")
	fmt.Fprintf(os.Stderr, "OPTIONS:\n")
	fmt.Fprintf(os.Stderr, "  -backend string\n")
	fmt.Fprintf(os.Stderr, "        code generation backend (vm|clift|c) (default \"c\")\n")
	fmt.Fprintf(os.Stderr, "  -O string\n")
	fmt.Fprintf(os.Stderr, "        optimization level (O0-O3) (default \"O0\")\n")
	fmt.Fprintf(os.Stderr, "  -emit string\n")
	fmt.Fprintf(os.Stderr, "        emission format (mir|obj|exe|binary|asm) (default \"exe\")\n")
	fmt.Fprintf(os.Stderr, "  -dump string\n")
	fmt.Fprintf(os.Stderr, "        dump intermediate representation (mir)\n")
	fmt.Fprintf(os.Stderr, "  -o string\n")
	fmt.Fprintf(os.Stderr, "        output binary path\n")
	fmt.Fprintf(os.Stderr, "  -verbose\n")
	fmt.Fprintf(os.Stderr, "        enable verbose output\n")
	fmt.Fprintf(os.Stderr, "  -version\n")
	fmt.Fprintf(os.Stderr, "        print version and exit\n")
	fmt.Fprintf(os.Stderr, "  -help, -h\n")
	fmt.Fprintf(os.Stderr, "        show help and exit\n\n")
	fmt.Fprintf(os.Stderr, "EXAMPLES:\n")
	fmt.Fprintf(os.Stderr, "  omnic hello.omni                    # Compile with C backend to executable\n")
	fmt.Fprintf(os.Stderr, "  omnic -backend vm hello.omni        # Compile with VM backend to MIR\n")
	fmt.Fprintf(os.Stderr, "  omnic -backend clift hello.omni     # Compile with Cranelift backend\n")
	fmt.Fprintf(os.Stderr, "  omnic -emit mir hello.omni          # Emit MIR instead of binary\n")
	fmt.Fprintf(os.Stderr, "  omnic -verbose hello.omni           # Show compilation steps\n")
	fmt.Fprintf(os.Stderr, "  omnic -dump mir hello.omni          # Dump MIR to file\n")
}

func run(input, output, backend, optLevel, emit, dump string, verbose, debug, debugModules bool) error {
	if filepath.Ext(input) != ".omni" {
		return fmt.Errorf("%s: unsupported input (expected .omni)", input)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Compiling %s...\n", input)
		fmt.Fprintf(os.Stderr, "  Backend: %s\n", backend)
		fmt.Fprintf(os.Stderr, "  Optimization: %s\n", optLevel)
		fmt.Fprintf(os.Stderr, "  Emit: %s\n", emit)
		if dump != "" {
			fmt.Fprintf(os.Stderr, "  Dump: %s\n", dump)
		}
		if output != "" {
			fmt.Fprintf(os.Stderr, "  Output: %s\n", output)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	cfg := compiler.Config{
		InputPath:    input,
		OutputPath:   output,
		Backend:      backend,
		OptLevel:     optLevel,
		Emit:         emit,
		Dump:         dump,
		DebugInfo:    debug,
		DebugModules: debugModules,
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Starting compilation...\n")
	}

	if err := compiler.Compile(cfg); err != nil {
		if errors.Is(err, compiler.ErrNotImplemented) {
			return fmt.Errorf("omnic: feature not implemented: %w", err)
		}
		return err
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Compilation completed successfully!\n")
	}

	return nil
}
