package main

import (
	"flag"
	"fmt"
	"os"

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

	if flag.NArg() == 0 {
		logger.ErrorString("no input file specified")
		fmt.Fprintln(os.Stderr, "")
		showUsage()
		os.Exit(2)
	}

	program := flag.Arg(0)
	if err := runner.Run(program, *verbose || *verboseAlt); err != nil {
		logger.ErrorString(err.Error())
		os.Exit(1)
	}
}

func showUsage() {
	fmt.Fprintf(os.Stderr, "OmniLang Runner (omnir) %s\n", Version)
	fmt.Fprintf(os.Stderr, "Built: %s\n\n", BuildTime)
	fmt.Fprintf(os.Stderr, "USAGE:\n")
	fmt.Fprintf(os.Stderr, "  omnir [options] <file.omni>\n\n")
	fmt.Fprintf(os.Stderr, "OPTIONS:\n")
	fmt.Fprintf(os.Stderr, "  -verbose, -V\n")
	fmt.Fprintf(os.Stderr, "        enable verbose output\n")
	fmt.Fprintf(os.Stderr, "  -version, -v\n")
	fmt.Fprintf(os.Stderr, "        print version and exit\n")
	fmt.Fprintf(os.Stderr, "  -help, -h\n")
	fmt.Fprintf(os.Stderr, "        show help and exit\n\n")
	fmt.Fprintf(os.Stderr, "EXAMPLES:\n")
	fmt.Fprintf(os.Stderr, "  omnir hello.omni           # Run the program\n")
	fmt.Fprintf(os.Stderr, "  omnir -verbose hello.omni  # Run with verbose output\n")
}
