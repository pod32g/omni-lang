package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/omni-lang/omni/internal/compiler"
	"github.com/omni-lang/omni/internal/logging"
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
		backend        = flag.String("backend", "c", "code generation backend (vm|clift|c)")
		backendShort   = flag.String("b", "", "alias for -backend")
		optLevel       = flag.String("O", "O0", "optimization level (O0-O3)")
		emitFlag       = newStringFlag("exe")
		emitShort      = newStringFlag("")
		dump           = flag.String("dump", "", "dump intermediate representation (mir)")
		dumpShort      = flag.String("d", "", "alias for -dump")
		output         = flag.String("o", "", "output binary path")
		debug          = flag.Bool("debug", false, "generate debug symbols and debug information")
		debugShort     = flag.Bool("g", false, "alias for -debug")
		debugModules   = flag.Bool("debug-modules", false, "show module loading debug information")
		debugModulesSh = flag.Bool("G", false, "alias for -debug-modules")
		version        = flag.Bool("version", false, "print version and exit")
		versionShort   = flag.Bool("v", false, "alias for -version")
		verbose        = flag.Bool("verbose", false, "enable verbose output")
		verboseShort   = flag.Bool("V", false, "alias for -verbose")
		listBackends   = flag.Bool("list-backends", false, "list supported backends and exit")
		listBackendsSh = flag.Bool("B", false, "alias for -list-backends")
		listEmits      = flag.Bool("list-emits", false, "list supported emit targets and exit")
		listEmitsShort = flag.Bool("E", false, "alias for -list-emits")
		help           = flag.Bool("help", false, "show help and exit")
		showHelp       = flag.Bool("h", false, "show help and exit")
	)
	flag.Var(emitFlag, "emit", "emission format (mir|obj|exe|binary|asm)")
	flag.Var(emitShort, "e", "alias for -emit")

	flag.Parse()

	logger := logging.Logger()
	logging.SetLevel(logging.LevelInfo)
	if *verbose || *verboseShort {
		logging.SetLevel(logging.LevelDebug)
	}

	if *listBackends || *listBackendsSh {
		printBackends()
		os.Exit(0)
	}

	if *listEmits || *listEmitsShort {
		printEmits()
		os.Exit(0)
	}

	if *version || *versionShort {
		fmt.Printf("omnic %s (built %s)\n", Version, BuildTime)
		os.Exit(0)
	}

	if *help || *showHelp {
		showUsage()
		os.Exit(0)
	}

	if flag.NArg() == 0 {
		logger.ErrorString("no input file specified")
		fmt.Fprintln(os.Stderr, "")
		showUsage()
		os.Exit(2)
	}

	input := flag.Arg(0)
	if *backendShort != "" {
		*backend = *backendShort
	}
	if *dumpShort != "" {
		*dump = *dumpShort
	}
	if *debugShort {
		*debug = true
	}
	if *debugModulesSh {
		*debugModules = true
	}
	emit := emitFlag.value
	if emitShort.set {
		emit = emitShort.value
	} else if !emitFlag.set {
		// Set appropriate defaults based on backend
		if *backend == "vm" {
			emit = "mir"
		} else if *backend == "c" {
			emit = "exe"
		} else if *backend == "clift" {
			emit = "obj"
		}
	}

	if err := run(input, *output, *backend, *optLevel, emit, *dump, *verbose || *verboseShort, *debug, *debugModules); err != nil {
		logger.ErrorString(err.Error())
		os.Exit(1)
	}
}

func showUsage() {
	fmt.Fprintf(os.Stderr, "OmniLang Compiler (omnic) %s\n", Version)
	fmt.Fprintf(os.Stderr, "Built: %s\n\n", BuildTime)
	fmt.Fprintf(os.Stderr, "USAGE:\n")
	fmt.Fprintf(os.Stderr, "  omnic [options] <file.omni>\n\n")
	fmt.Fprintf(os.Stderr, "OPTIONS:\n")
	fmt.Fprintf(os.Stderr, "  -backend, -b string\n")
	fmt.Fprintf(os.Stderr, "        code generation backend (vm|clift|c) (default \"c\")\n")
	fmt.Fprintf(os.Stderr, "  -O string\n")
	fmt.Fprintf(os.Stderr, "        optimization level (O0-O3) (default \"O0\")\n")
	fmt.Fprintf(os.Stderr, "  -emit, -e string\n")
	fmt.Fprintf(os.Stderr, "        emission format (mir|obj|exe|binary|asm) (default \"exe\")\n")
	fmt.Fprintf(os.Stderr, "  -dump, -d string\n")
	fmt.Fprintf(os.Stderr, "        dump intermediate representation (mir)\n")
	fmt.Fprintf(os.Stderr, "  -o string\n")
	fmt.Fprintf(os.Stderr, "        output binary path\n")
	fmt.Fprintf(os.Stderr, "  -debug, -g\n")
	fmt.Fprintf(os.Stderr, "        generate debug symbols and debug information\n")
	fmt.Fprintf(os.Stderr, "  -debug-modules, -G\n")
	fmt.Fprintf(os.Stderr, "        show module loading debug information\n")
	fmt.Fprintf(os.Stderr, "  -verbose, -V\n")
	fmt.Fprintf(os.Stderr, "        enable verbose output\n")
	fmt.Fprintf(os.Stderr, "  -version, -v\n")
	fmt.Fprintf(os.Stderr, "        print version and exit\n")
	fmt.Fprintf(os.Stderr, "  -list-backends, -B\n")
	fmt.Fprintf(os.Stderr, "        list available backends and exit\n")
	fmt.Fprintf(os.Stderr, "  -list-emits, -E\n")
	fmt.Fprintf(os.Stderr, "        list available emit targets and exit\n")
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

	if output == "" && emit != "exe" {
		output = deriveOutputPath(input, emit)
	}

	logger := logging.Logger()

	if verbose {
		logger.DebugString("Compiling " + input + "...")
		logger.DebugFields("Compilation settings",
			logging.String("backend", backend),
			logging.String("optimization", optLevel),
			logging.String("emit", emit),
		)
		if dump != "" {
			logger.DebugFields("Dump configured", logging.String("path", dump))
		}
		if output != "" {
			logger.DebugFields("Output configured", logging.String("path", output))
		}
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
		logger.DebugString("Starting compilation...")
	}

	if err := compiler.Compile(cfg); err != nil {
		if errors.Is(err, compiler.ErrNotImplemented) {
			return fmt.Errorf("omnic: feature not implemented: %w", err)
		}
		return err
	}

	if verbose {
		logger.DebugString("Compilation completed successfully!")
	}

	return nil
}

func deriveOutputPath(input, emit string) string {
	dir := filepath.Dir(input)
	base := strings.TrimSuffix(filepath.Base(input), filepath.Ext(input))
	var ext string
	switch emit {
	case "mir":
		ext = ".mir"
	case "obj":
		ext = ".o"
	case "asm":
		ext = ".s"
	case "binary":
		ext = ".bin"
	default:
		ext = "." + emit
	}
	return filepath.Join(dir, base+ext)
}

func printBackends() {
	fmt.Println("Available backends:")
	fmt.Println("  c       - C code-generation backend (default)")
	fmt.Println("  vm      - Execute with virtual machine interpreter")
	fmt.Println("  clift   - Cranelift backend (experimental)")
}

func printEmits() {
	fmt.Println("Available emit targets:")
	fmt.Println("  exe     - native executable (default for c backend)")
	fmt.Println("  mir     - OmniLang MIR (default for vm backend)")
	fmt.Println("  obj     - object file")
	fmt.Println("  binary  - raw binary image")
	fmt.Println("  asm     - assembly listing")
}
