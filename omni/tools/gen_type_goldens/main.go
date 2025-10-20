package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/omni-lang/omni/internal/parser"
	"github.com/omni-lang/omni/internal/types/checker"
)

type caseSpec struct {
	name   string
	source string
}

func main() {
	cases := buildCases()
	dir := filepath.Join("tests", "goldens", "types")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		panic(err)
	}

	for _, c := range cases {
		omniPath := filepath.Join(dir, c.name+".omni")
		errPath := filepath.Join(dir, c.name+".err")
		if err := os.WriteFile(omniPath, []byte(c.source), 0o644); err != nil {
			panic(fmt.Errorf("write %s: %w", omniPath, err))
		}
		logical := filepath.ToSlash(filepath.Join("tests", "goldens", "types", c.name+".omni"))
		mod, err := parser.Parse(logical, c.source)
		if err != nil {
			panic(fmt.Errorf("parse %s: %w", omniPath, err))
		}
		err = checker.Check(logical, c.source, mod)
		output := ""
		if err != nil {
			output = err.Error()
			if len(output) == 0 || output[len(output)-1] != '\n' {
				output += "\n"
			}
		}
		if err := os.WriteFile(errPath, []byte(output), 0o644); err != nil {
			panic(fmt.Errorf("write %s: %w", errPath, err))
		}
	}
}

func buildCases() []caseSpec {
	cases := make([]caseSpec, 0, 50)

	for i := 1; i <= 25; i++ {
		cases = append(cases, caseSpec{
			name:   fmt.Sprintf("unknown_type_%02d", i),
			source: fmt.Sprintf("let value%d:Mystery%d = 1\n", i, i),
		})
	}

	for i := 1; i <= 25; i++ {
		cases = append(cases, caseSpec{
			name:   fmt.Sprintf("undefined_ident_%02d", i),
			source: fmt.Sprintf("let result%d:int = missing%d\n", i, i),
		})
	}

	if len(cases) != 50 {
		panic(fmt.Sprintf("expected 50 cases, got %d", len(cases)))
	}
	return cases
}
