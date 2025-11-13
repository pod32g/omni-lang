package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/omni-lang/omni/internal/compiler"
	"github.com/omni-lang/omni/internal/logging"
	"github.com/omni-lang/omni/internal/runner"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	var (
		version    = flag.Bool("version", false, "print version and exit")
		versionAlt = flag.Bool("v", false, "alias for -version")
		verbose    = flag.Bool("verbose", false, "enable verbose output")
		verboseAlt = flag.Bool("V", false, "alias for -verbose")
		backend    = flag.String("backend", "vm", "execution backend (vm|c)")
		backendAlt = flag.String("b", "", "alias for -backend")
		stats      = flag.Bool("stats", false, "print execution duration summary")
		help       = flag.Bool("help", false, "show help and exit")
		showHelp   = flag.Bool("h", false, "show help and exit")
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
	if len(args) == 0 {
		logger.ErrorString("no input file specified")
		fmt.Fprintln(os.Stderr, "")
		showUsage()
		os.Exit(2)
	}

	if *backendAlt != "" {
		*backend = *backendAlt
	}
	program := args[0]
	programArgs := args[1:]

	if err := runProgram(program, programArgs, *backend, *verbose || *verboseAlt, *stats); err != nil {
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
	fmt.Fprintf(os.Stderr, "  -help, -h\n")
	fmt.Fprintf(os.Stderr, "        show help and exit\n\n")
	fmt.Fprintf(os.Stderr, "EXAMPLES:\n")
	fmt.Fprintf(os.Stderr, "  omnir hello.omni                  # Run with VM backend\n")
	fmt.Fprintf(os.Stderr, "  omnir -backend c hello.omni -- hi # Compile to native exe then run with args\n")
}

func runProgram(program string, args []string, backend string, verbose bool, stats bool) error {
	switch backend {
	case "vm":
		if len(args) > 0 {
			return errors.New("vm backend: argument forwarding is not supported")
		}
		return runner.Run(program, verbose)
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
