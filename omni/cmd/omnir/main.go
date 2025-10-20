package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/omni-lang/omni/internal/runner"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	var (
		version = flag.Bool("version", false, "print version and exit")
	)
	flag.Parse()

	if *version {
		fmt.Printf("omnir %s (built %s)\n", Version, BuildTime)
		os.Exit(0)
	}

	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: omnir [options] <file.omni>")
		flag.PrintDefaults()
		os.Exit(2)
	}

	program := flag.Arg(0)
	if err := runner.Run(program); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
