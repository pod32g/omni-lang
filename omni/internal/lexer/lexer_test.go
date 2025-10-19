package lexer_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/omni-lang/omni/internal/lexer"
	"github.com/omni-lang/omni/internal/testutil/snapshots"
)

func TestGoldenTokens(t *testing.T) {
	goldenDir := filepath.Join("..", "..", "tests", "goldens", "tokens")
	paths, err := filepath.Glob(filepath.Join(goldenDir, "*.omni"))
	if err != nil {
		t.Fatalf("glob goldens: %v", err)
	}
	sort.Strings(paths)
	if len(paths) == 0 {
		t.Fatalf("no golden inputs found in %s", goldenDir)
	}

	for _, inputPath := range paths {
		base := strings.TrimSuffix(filepath.Base(inputPath), ".omni")
		expectedPath := filepath.Join(goldenDir, base+".tok")

		t.Run(base, func(t *testing.T) {
			src, err := os.ReadFile(inputPath)
			if err != nil {
				t.Fatalf("read input: %v", err)
			}

			tokens, err := lexer.LexAll(inputPath, string(src))
			if err != nil {
				t.Fatalf("lex %s: %v", inputPath, err)
			}

			lines := make([]string, 0, len(tokens))
			for _, tok := range tokens {
				lines = append(lines, tok.Format())
			}
			actual := strings.Join(lines, "\n") + "\n"

			snapshots.CompareText(t, actual, expectedPath)
		})
	}
}
