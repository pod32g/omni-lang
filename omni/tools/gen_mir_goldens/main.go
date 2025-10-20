package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/omni-lang/omni/internal/mir/builder"
	mirprinter "github.com/omni-lang/omni/internal/mir/printer"
	"github.com/omni-lang/omni/internal/parser"
	"github.com/omni-lang/omni/internal/passes"
	"github.com/omni-lang/omni/internal/types/checker"
)

type caseSpec struct {
	name   string
	source string
}

func main() {
	cases := buildCases()
	dir := filepath.Join("tests", "goldens", "mir")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		panic(err)
	}

	for _, c := range cases {
		omniPath := filepath.Join(dir, c.name+".omni")
		mirPath := filepath.Join(dir, c.name+".mir")
		if err := os.WriteFile(omniPath, []byte(c.source), 0o644); err != nil {
			panic(fmt.Errorf("write %s: %w", omniPath, err))
		}

		logical := filepath.ToSlash(filepath.Join("tests", "goldens", "mir", c.name+".omni"))
		astModule, err := parser.Parse(logical, c.source)
		if err != nil {
			panic(fmt.Errorf("parse %s: %w", c.name, err))
		}

		if err := checker.Check(logical, c.source, astModule); err != nil {
			panic(fmt.Errorf("type check %s: %w", c.name, err))
		}

		mirModule, err := builder.BuildModule(astModule)
		if err != nil {
			panic(fmt.Errorf("build MIR %s: %w", c.name, err))
		}

		pipeline := passes.NewPipeline("mir-golden")
		if _, err := pipeline.Run(*mirModule); err != nil {
			panic(fmt.Errorf("verify MIR %s: %w", c.name, err))
		}

		output := mirprinter.Format(mirModule)
		if !strings.HasSuffix(output, "\n") {
			output += "\n"
		}
		if err := os.WriteFile(mirPath, []byte(output), 0o644); err != nil {
			panic(fmt.Errorf("write %s: %w", mirPath, err))
		}
	}
}

func buildCases() []caseSpec {
	return []caseSpec{
		{
			name:   "simple_return",
			source: "func fortyTwo():int => 42\n",
		},
		{
			name:   "add_params",
			source: "func add(a:int, b:int):int { return a + b }\n",
		},
		{
			name:   "let_binding",
			source: "func sumTo(n:int):int { let total:int = n + 1\n  return total\n}\n",
		},
		{
			name:   "call_function",
			source: "func inc(x:int):int => x + 1\nfunc main():int { let value:int = inc(41)\n  return value\n}\n",
		},
		{
			name:   "if_else",
			source: "func max(a:int, b:int):int { if a > b { return a } else { return b } }\n",
		},
	}
}
