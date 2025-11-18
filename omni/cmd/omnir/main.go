package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/omni-lang/omni/internal/compiler"
	"github.com/omni-lang/omni/internal/logging"
	"github.com/omni-lang/omni/internal/runner"
	"github.com/omni-lang/omni/internal/vm"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	var (
		version        = flag.Bool("version", false, "print version and exit")
		versionAlt     = flag.Bool("v", false, "alias for -version")
		verbose        = flag.Bool("verbose", false, "enable verbose output")
		verboseAlt     = flag.Bool("V", false, "alias for -verbose")
		backend        = flag.String("backend", "vm", "execution backend (vm|c)")
		backendAlt     = flag.String("b", "", "alias for -backend")
		stats          = flag.Bool("stats", false, "print execution duration summary")
		stdinSrc       = flag.Bool("stdin", false, "read OmniLang source from stdin")
		watch          = flag.Bool("watch", false, "watch program file and rerun on changes")
		watchShort     = flag.Bool("w", false, "alias for -watch")
		testMode       = flag.Bool("test", false, "run using the built-in testing harness (vm backend only)")
		coverage       = flag.Bool("coverage", false, "enable coverage tracking for standard library functions")
		coverageOutput = flag.String("coverage-output", "", "file path to write coverage data (JSON format)")
		help           = flag.Bool("help", false, "show help and exit")
		showHelp       = flag.Bool("h", false, "show help and exit")
	)
	flag.Parse()

	logger := logging.Logger()
	logging.SetLevel(logging.LevelInfo)
	if *verbose || *verboseAlt {
		logging.SetLevel(logging.LevelDebug)
	}

	if *version || *versionAlt {
		fmt.Printf("omnir %s (built %s)\n", Version, BuildTime)
		os.Exit(0)
	}

	if *help || *showHelp {
		showUsage()
		os.Exit(0)
	}

	args := flag.Args()
	if !*stdinSrc && len(args) == 0 {
		logger.ErrorString("no input file specified")
		fmt.Fprintln(os.Stderr, "")
		showUsage()
		os.Exit(2)
	}

	if *backendAlt != "" {
		*backend = *backendAlt
	}
	if *watchShort {
		*watch = true
	}

	var (
		program     string
		programArgs []string
		cleanup     func()
	)

	if *stdinSrc {
		if len(args) > 0 && args[0] != "--" {
			logger.ErrorString("cannot specify program path when using --stdin")
			fmt.Fprintln(os.Stderr, "")
			showUsage()
			os.Exit(2)
		}
		if len(args) > 0 && args[0] == "--" {
			args = args[1:]
		}
		path, clean, err := writeStdinProgram()
		if err != nil {
			logger.ErrorString(fmt.Sprintf("failed to read stdin: %v", err))
			os.Exit(1)
		}
		program = path
		programArgs = args
		cleanup = clean
	} else {
		if len(args) == 0 {
			logger.ErrorString("no input file specified")
			fmt.Fprintln(os.Stderr, "")
			showUsage()
			os.Exit(2)
		}
		program = args[0]
		programArgs = args[1:]
		if len(programArgs) > 0 && programArgs[0] == "--" {
			programArgs = programArgs[1:]
		}
	}
	if cleanup != nil {
		defer cleanup()
	}

	// Enable coverage tracking if requested
	if *coverage {
		vm.SetCoverageEnabled(true)
		vm.ResetCoverage()
	}

	if *testMode {
		if *watch || *watchShort {
			logger.ErrorString("--test cannot be combined with --watch")
			os.Exit(2)
		}
		if *backend != "vm" {
			logger.ErrorString("--test mode currently supports only the vm backend")
			os.Exit(2)
		}
		if len(programArgs) > 0 {
			logger.ErrorString("--test mode does not support forwarding program arguments")
			os.Exit(2)
		}
		code := runTests(program, *verbose || *verboseAlt, *stats, *coverage, *coverageOutput)
		if code != 0 {
			logger.ErrorString(fmt.Sprintf("%d test(s) failed", code))
		}
		os.Exit(code)
	}

	if *watch || *watchShort {
		if *stdinSrc {
			logger.ErrorString("watch mode is not supported with --stdin")
			os.Exit(2)
		}
		if err := watchAndRun(program, programArgs, *backend, *verbose || *verboseAlt, *stats, *coverage, *coverageOutput); err != nil {
			logger.ErrorString(err.Error())
			os.Exit(1)
		}
		return
	}

	if err := runProgram(program, programArgs, *backend, *verbose || *verboseAlt, *stats, *coverage, *coverageOutput); err != nil {
		logger.ErrorString(err.Error())
		os.Exit(1)
	}
}

func showUsage() {
	fmt.Fprintf(os.Stderr, "OmniLang Runner (omnir) %s\n", Version)
	fmt.Fprintf(os.Stderr, "Built: %s\n\n", BuildTime)
	fmt.Fprintf(os.Stderr, "USAGE:\n")
	fmt.Fprintf(os.Stderr, "  omnir [options] <program> [-- <args>]\n\n")
	fmt.Fprintf(os.Stderr, "OPTIONS:\n")
	fmt.Fprintf(os.Stderr, "  -verbose, -V\n")
	fmt.Fprintf(os.Stderr, "        enable verbose output\n")
	fmt.Fprintf(os.Stderr, "  -version, -v\n")
	fmt.Fprintf(os.Stderr, "        print version and exit\n")
	fmt.Fprintf(os.Stderr, "  -backend, -b string\n")
	fmt.Fprintf(os.Stderr, "        execution backend (vm|c) (default \"vm\")\n")
	fmt.Fprintf(os.Stderr, "  -stats\n")
	fmt.Fprintf(os.Stderr, "        print execution duration summary\n")
	fmt.Fprintf(os.Stderr, "  -test\n")
	fmt.Fprintf(os.Stderr, "        execute using the OmniLang test harness (vm backend only)\n")
	fmt.Fprintf(os.Stderr, "  -coverage\n")
	fmt.Fprintf(os.Stderr, "        enable coverage tracking for standard library functions\n")
	fmt.Fprintf(os.Stderr, "  -coverage-output string\n")
	fmt.Fprintf(os.Stderr, "        file path to write coverage data (JSON format)\n")
	fmt.Fprintf(os.Stderr, "  -stdin\n")
	fmt.Fprintf(os.Stderr, "        read source code from standard input\n")
	fmt.Fprintf(os.Stderr, "  -watch, -w\n")
	fmt.Fprintf(os.Stderr, "        watch program for changes and rerun automatically\n")
	fmt.Fprintf(os.Stderr, "  -help, -h\n")
	fmt.Fprintf(os.Stderr, "        show help and exit\n\n")
	fmt.Fprintf(os.Stderr, "EXAMPLES:\n")
	fmt.Fprintf(os.Stderr, "  omnir hello.omni                  # Run with VM backend\n")
	fmt.Fprintf(os.Stderr, "  omnir -backend c hello.omni -- hi # Compile to native exe then run with args\n")
	fmt.Fprintf(os.Stderr, "  cat hello.omni | omnir --stdin    # Run source from stdin\n")
	fmt.Fprintf(os.Stderr, "  omnir --watch hello.omni          # Automatically rerun on file changes\n")
}

func runTests(program string, verbose bool, stats bool, coverageEnabled bool, coverageOutput string) int {
	start := time.Now()
	result, err := runner.Execute(program, nil, verbose)
	code := 0
	if err != nil {
		var exitErr vm.ExitError
		if errors.As(err, &exitErr) {
			code = exitErr.Code
		} else {
			fmt.Fprintf(os.Stderr, "failed to execute tests: %v\n", err)
			return 1
		}
	} else if result.Type == "int" {
		if value, ok := result.Value.(int); ok {
			code = value
		}
	}

	// Export coverage data if enabled
	if coverageEnabled {
		coverageData, err := vm.ExportCoverage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to export coverage: %v\n", err)
		} else if coverageOutput != "" {
			if err := os.WriteFile(coverageOutput, coverageData, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "failed to write coverage file: %v\n", err)
			} else if verbose {
				fmt.Fprintf(os.Stderr, "Coverage data written to %s\n", coverageOutput)
			}
		} else {
			// Write to default location
			defaultPath := "coverage.json"
			if err := os.WriteFile(defaultPath, coverageData, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "failed to write coverage file: %v\n", err)
			} else if verbose {
				fmt.Fprintf(os.Stderr, "Coverage data written to %s\n", defaultPath)
			}
		}
	}

	if stats {
		elapsed := time.Since(start).Round(time.Millisecond)
		fmt.Fprintf(os.Stderr, "Test run completed in %s (failures=%d)\n", elapsed, code)
	}
	return code
}

func runProgram(program string, args []string, backend string, verbose bool, stats bool, coverageEnabled bool, coverageOutput string) error {
	switch backend {
	case "vm":
		err := runner.Run(program, args, verbose)
		// Export coverage data if enabled
		if coverageEnabled {
			coverageData, exportErr := vm.ExportCoverage()
			if exportErr != nil {
				fmt.Fprintf(os.Stderr, "failed to export coverage: %v\n", exportErr)
			} else if coverageOutput != "" {
				if writeErr := os.WriteFile(coverageOutput, coverageData, 0644); writeErr != nil {
					fmt.Fprintf(os.Stderr, "failed to write coverage file: %v\n", writeErr)
				} else if verbose {
					fmt.Fprintf(os.Stderr, "Coverage data written to %s\n", coverageOutput)
				}
			} else {
				// Write to default location
				defaultPath := "coverage.json"
				if writeErr := os.WriteFile(defaultPath, coverageData, 0644); writeErr != nil {
					fmt.Fprintf(os.Stderr, "failed to write coverage file: %v\n", writeErr)
				} else if verbose {
					fmt.Fprintf(os.Stderr, "Coverage data written to %s\n", defaultPath)
				}
			}
		}
		return err
	case "c":
		return runNative(program, args, verbose, stats)
	default:
		return fmt.Errorf("unsupported backend: %s", backend)
	}
}

func runNative(program string, args []string, verbose bool, stats bool) error {
	if filepath.Ext(program) != ".omni" {
		// Assume it's an already-built binary; execute directly.
		return execWithStats(program, args, stats)
	}

	tmpDir, err := os.MkdirTemp("", "omnir-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	base := strings.TrimSuffix(filepath.Base(program), filepath.Ext(program))
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	output := filepath.Join(tmpDir, base+ext)

	cfg := compiler.Config{
		InputPath:    program,
		OutputPath:   output,
		Backend:      "c",
		OptLevel:     "O2",
		Emit:         "exe",
		DebugInfo:    false,
		DebugModules: false,
	}

	if verbose {
		logging.Logger().DebugFields("Compiling program",
			logging.String("input", program),
			logging.String("output", output),
		)
	}

	if err := compiler.Compile(cfg); err != nil {
		return fmt.Errorf("compile program: %w", err)
	}

	return execWithStats(output, args, stats)
}

func execWithStats(binary string, args []string, stats bool) error {
	cmd := exec.Command(binary, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	var start time.Time
	if stats {
		start = time.Now()
	}

	if err := cmd.Run(); err != nil {
		return err
	}

	if stats {
		fmt.Fprintf(os.Stderr, "Program finished in %s\n", time.Since(start).Round(time.Millisecond))
	}
	return nil
}

func writeStdinProgram() (string, func(), error) {
	tempFile, err := os.CreateTemp("", "omnir-stdin-*.omni")
	if err != nil {
		return "", nil, err
	}
	defer tempFile.Close()

	if _, err := io.Copy(tempFile, os.Stdin); err != nil {
		path := tempFile.Name()
		_ = os.Remove(path)
		return "", nil, err
	}

	path := tempFile.Name()
	cleanup := func() {
		_ = os.Remove(path)
	}
	return path, cleanup, nil
}

func watchAndRun(program string, args []string, backend string, verbose bool, stats bool, coverageEnabled bool, coverageOutput string) error {
	abs, err := filepath.Abs(program)
	if err != nil {
		return fmt.Errorf("resolve program path: %w", err)
	}
	if filepath.Ext(abs) != ".omni" {
		return errors.New("watch mode requires an OmniLang source file")
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

	logger := logging.Logger()
	logger.InfoFields("Watching file for changes", logging.String("file", abs))

	runOnce := func() {
		if err := runProgram(program, args, backend, verbose, stats, coverageEnabled, coverageOutput); err != nil {
			logger.ErrorString(err.Error())
		}
	}

	runOnce()

	debounce := time.NewTimer(time.Hour)
	debounce.Stop()

	for {
		select {
		case evt := <-watcher.Events:
			if filepath.Base(evt.Name) != base {
				continue
			}
			if evt.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
				continue
			}
			if !debounce.Stop() {
				select {
				case <-debounce.C:
				default:
				}
			}
			debounce.Reset(250 * time.Millisecond)
		case <-debounce.C:
			runOnce()
			debounce.Stop()
		case err := <-watcher.Errors:
			logger.ErrorFields("watch error", logging.Error("error", err))
		}
	}
}
