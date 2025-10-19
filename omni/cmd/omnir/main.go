package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/omni-lang/omni/internal/runner"
)

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: omnir <file.omni>")
		os.Exit(2)
	}

	program := flag.Arg(0)
	if err := runner.Run(program); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
