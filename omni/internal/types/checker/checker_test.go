package checker_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/omni-lang/omni/internal/parser"
	"github.com/omni-lang/omni/internal/testutil/snapshots"
	"github.com/omni-lang/omni/internal/types/checker"
)

func TestTypeCheckerGoldens(t *testing.T) {
	goldenDir := filepath.Join("..", "..", "..", "tests", "goldens", "types")
	inputs, err := filepath.Glob(filepath.Join(goldenDir, "*.omni"))
	if err != nil {
		t.Fatalf("glob goldens: %v", err)
	}
	sort.Strings(inputs)
	if len(inputs) == 0 {
		t.Fatalf("no type checker goldens found in %s", goldenDir)
	}

	for _, inputPath := range inputs {
		base := strings.TrimSuffix(filepath.Base(inputPath), ".omni")
		expectedPath := filepath.Join(goldenDir, base+".err")
		logicalName := filepath.ToSlash(filepath.Join("tests", "goldens", "types", base+".omni"))

		t.Run(base, func(t *testing.T) {
			src, err := os.ReadFile(inputPath)
			if err != nil {
				t.Fatalf("read input: %v", err)
			}

			mod, err := parser.Parse(logicalName, string(src))
			if err != nil {
				t.Fatalf("parse %s: %v", inputPath, err)
			}

			err = checker.Check(logicalName, string(src), mod)
			actual := ""
			if err != nil {
				actual = err.Error()
				if !strings.HasSuffix(actual, "\n") {
					actual += "\n"
				}
			}

			snapshots.CompareText(t, actual, expectedPath)
		})
	}
}
