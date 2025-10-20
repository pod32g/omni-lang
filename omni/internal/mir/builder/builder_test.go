package builder_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/omni-lang/omni/internal/mir/builder"
	mirprinter "github.com/omni-lang/omni/internal/mir/printer"
	"github.com/omni-lang/omni/internal/parser"
	"github.com/omni-lang/omni/internal/passes"
	"github.com/omni-lang/omni/internal/testutil/snapshots"
	"github.com/omni-lang/omni/internal/types/checker"
)

func TestMIRBuilderGoldens(t *testing.T) {
	goldenDir := filepath.Join("..", "..", "..", "tests", "goldens", "mir")
	inputs, err := filepath.Glob(filepath.Join(goldenDir, "*.omni"))
	if err != nil {
		t.Fatalf("glob goldens: %v", err)
	}
	sort.Strings(inputs)
	if len(inputs) == 0 {
		t.Fatalf("no MIR goldens found in %s", goldenDir)
	}

	for _, inputPath := range inputs {
		base := strings.TrimSuffix(filepath.Base(inputPath), ".omni")
		expectedPath := filepath.Join(goldenDir, base+".mir")
		logicalName := filepath.ToSlash(filepath.Join("tests", "goldens", "mir", base+".omni"))

		t.Run(base, func(t *testing.T) {
			src, err := os.ReadFile(inputPath)
			if err != nil {
				t.Fatalf("read input: %v", err)
			}

			astModule, err := parser.Parse(logicalName, string(src))
			if err != nil {
				t.Fatalf("parse %s: %v", inputPath, err)
			}

			if err := checker.Check(logicalName, string(src), astModule); err != nil {
				t.Fatalf("type check %s: %v", inputPath, err)
			}

			mirModule, err := builder.BuildModule(astModule)
			if err != nil {
				t.Fatalf("build MIR: %v", err)
			}

			pipeline := passes.NewPipeline("test")
			if _, err := pipeline.Run(*mirModule); err != nil {
				t.Fatalf("verify MIR: %v", err)
			}

			actual := mirprinter.Format(mirModule)
			snapshots.CompareText(t, actual, expectedPath)
		})
	}
}
