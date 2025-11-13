package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
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
		emitDir        = flag.String("emit-dir", "", "directory where derived outputs are written")
		emitDirShort   = flag.String("C", "", "alias for -emit-dir")
		emitPrefix     = flag.String("emit-prefix", "", "prefix applied to derived output names")
		emitPrefixSh   = flag.String("P", "", "alias for -emit-prefix")
		noColor        = flag.Bool("no-color", false, "disable colored log output")
		quiet          = flag.Bool("quiet", false, "suppress non-error output")
		quietShort     = flag.Bool("q", false, "alias for -quiet")
		timeCompile    = flag.Bool("time", false, "print compilation timing summary")
		watchFlag      = flag.Bool("watch", false, "watch input file and recompile on changes")
		watchShort     = flag.Bool("w", false, "alias for -watch")
		jsonOutput     = flag.Bool("json", false, "output machine-readable JSON for listings")
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

	if *emitDirShort != "" {
		*emitDir = *emitDirShort
	}
	if *emitPrefixSh != "" {
		*emitPrefix = *emitPrefixSh
	}
	if *quietShort {
		*quiet = true
	}
	if *watchShort {
		*watchFlag = true
	}
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
	if emitShort.set {
		emitFlag.value = emitShort.value
		emitFlag.set = true
	}
	if *noColor {
		os.Setenv("LOG_COLORIZE", "false")
	}

	logger := logging.Logger()
	logging.SetLevel(logging.LevelInfo)
	if *quiet {
		logging.SetLevel(logging.LevelError)
	} else if *verbose || *verboseShort {
		logging.SetLevel(logging.LevelDebug)
	}

	if *listBackends || *listBackendsSh {
		printBackends(*jsonOutput)
		os.Exit(0)
	}

	if *listEmits || *listEmitsShort {
		printEmits(*jsonOutput)
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

	finalOutput := *output
	if finalOutput == "" {
		finalOutput = deriveOutputPath(input, emit, *emitDir, *emitPrefix)
	}

	compileAndReport := func() (string, error) {
		start := time.Now()
		outputPath, err := run(input, finalOutput, *backend, *optLevel, emit, *dump, *verbose || *verboseShort, *debug, *debugModules)
		duration := time.Since(start)
		if err != nil {
			logger.ErrorString(err.Error())
			if *jsonOutput && !*watchFlag {
				enc := json.NewEncoder(os.Stdout)
				_ = enc.Encode(map[string]any{
					"status": "error",
					"error":  err.Error(),
				})
			}
			return outputPath, err
		}

		target := outputPath
		if target == "" {
			target = "(default)"
		}

		if *timeCompile && !*quiet && !*jsonOutput {
			fmt.Fprintf(os.Stderr, "Compiled %s -> %s in %s (backend=%s emit=%s)\n",
				input, target, duration.Round(time.Millisecond), *backend, emit)
		} else if !*quiet && !*jsonOutput && !*watchFlag {
			fmt.Fprintf(os.Stderr, "Compiled %s -> %s (backend=%s emit=%s)\n",
				input, target, *backend, emit)
		}

		if *jsonOutput && !*watchFlag {
			enc := json.NewEncoder(os.Stdout)
			result := map[string]any{
				"status":  "ok",
				"backend": *backend,
				"emit":    emit,
				"output":  outputPath,
			}
			if *timeCompile {
				result["duration_ms"] = float64(duration) / float64(time.Millisecond)
			}
			_ = enc.Encode(result)
		}

		return outputPath, nil
	}

	compileOnce := func() error {
		_, err := compileAndReport()
		return err
	}

	if *watchFlag {
		if *jsonOutput {
			logger.ErrorString("cannot use --watch together with --json")
			os.Exit(2)
		}
		if err := watchAndCompile(input, compileOnce, *quiet); err != nil {
			logger.ErrorString(err.Error())
			os.Exit(1)
		}
		return
	}

	if _, err := compileAndReport(); err != nil {
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
	fmt.Fprintf(os.Stderr, "  -emit-dir, -C string\n")
	fmt.Fprintf(os.Stderr, "        directory for derived outputs when -o is omitted\n")
	fmt.Fprintf(os.Stderr, "  -emit-prefix, -P string\n")
	fmt.Fprintf(os.Stderr, "        prefix applied to derived output names\n")
	fmt.Fprintf(os.Stderr, "  -debug, -g\n")
	fmt.Fprintf(os.Stderr, "        generate debug symbols and debug information\n")
	fmt.Fprintf(os.Stderr, "  -debug-modules, -G\n")
	fmt.Fprintf(os.Stderr, "        show module loading debug information\n")
	fmt.Fprintf(os.Stderr, "  -verbose, -V\n")
	fmt.Fprintf(os.Stderr, "        enable verbose output\n")
	fmt.Fprintf(os.Stderr, "  -quiet, -q\n")
	fmt.Fprintf(os.Stderr, "        suppress non-error output\n")
	fmt.Fprintf(os.Stderr, "  -no-color\n")
	fmt.Fprintf(os.Stderr, "        disable colored log output\n")
	fmt.Fprintf(os.Stderr, "  -time\n")
	fmt.Fprintf(os.Stderr, "        print compilation timing summary\n")
	fmt.Fprintf(os.Stderr, "  -watch, -w\n")
	fmt.Fprintf(os.Stderr, "        watch input file for changes and recompile\n")
	fmt.Fprintf(os.Stderr, "  -json\n")
	fmt.Fprintf(os.Stderr, "        output machine-readable JSON for listings and one-shot builds\n")
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

func run(input, output, backend, optLevel, emit, dump string, verbose, debug, debugModules bool) (string, error) {
	if filepath.Ext(input) != ".omni" {
		return "", fmt.Errorf("%s: unsupported input (expected .omni)", input)
	}

	if output != "" {
		if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
			return "", fmt.Errorf("create output directory: %w", err)
		}
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
			return "", fmt.Errorf("omnic: feature not implemented: %w", err)
		}
		return "", err
	}

	if verbose {
		logger.DebugString("Compilation completed successfully!")
	}

	return cfg.OutputPath, nil
}

func deriveOutputPath(input, emit, emitDir, emitPrefix string) string {
	if emitDir == "" && emitPrefix == "" && emit == "exe" {
		return ""
	}

	dir := filepath.Dir(input)
	if emitDir != "" {
		dir = emitDir
	}
	base := strings.TrimSuffix(filepath.Base(input), filepath.Ext(input))
	if emitPrefix != "" {
		base = emitPrefix + base
	}

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
	case "exe":
		if runtime.GOOS == "windows" {
			ext = ".exe"
		}
	default:
		ext = "." + emit
	}
	return filepath.Join(dir, base+ext)
}

func printBackends(jsonOutput bool) {
	data := []map[string]string{
		{"name": "c", "description": "C code-generation backend (default)"},
		{"name": "vm", "description": "Virtual machine interpreter backend"},
		{"name": "clift", "description": "Cranelift backend (experimental)"},
	}
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(data)
		return
	}
	fmt.Println("Available backends:")
	for _, entry := range data {
		fmt.Printf("  %-7s- %s\n", entry["name"], entry["description"])
	}
}

func printEmits(jsonOutput bool) {
	data := []map[string]string{
		{"name": "exe", "description": "Native executable (default for C backend)"},
		{"name": "mir", "description": "OmniLang MIR (default for VM backend)"},
		{"name": "obj", "description": "Object file"},
		{"name": "binary", "description": "Raw binary image"},
		{"name": "asm", "description": "Assembly listing"},
	}
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(data)
		return
	}
	fmt.Println("Available emit targets:")
	for _, entry := range data {
		fmt.Printf("  %-7s- %s\n", entry["name"], entry["description"])
	}
}

func watchAndCompile(path string, compile func() error, quiet bool) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}
	dir := filepath.Dir(abs)
	base := filepath.Base(abs)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}
	defer watcher.Close()

	if err := watcher.Add(dir); err != nil {
		return fmt.Errorf("watch directory %s: %w", dir, err)
	}

	if !quiet {
		logging.Logger().InfoFields("Watching for changes",
			logging.String("file", abs))
	}

	if err := compile(); err != nil {
		// Errors are already logged; continue watching.
	}

	debounce := time.NewTimer(time.Hour)
	debounce.Stop()

	for {
		select {
		case event := <-watcher.Events:
			if filepath.Base(event.Name) != base {
				continue
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
				continue
			}
			if !debounce.Stop() {
				select {
				case <-debounce.C:
				default:
				}
			}
			debounce.Reset(200 * time.Millisecond)
		case <-debounce.C:
			if err := compile(); err != nil {
				// already logged
			}
			debounce.Stop()
		case err := <-watcher.Errors:
			logging.Logger().ErrorFields("watch error", logging.Error("error", err))
		}
	}
}
