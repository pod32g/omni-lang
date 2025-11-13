//go:build !darwin && !windows

package cranelift

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/omni-lang/omni/internal/mir/builder"
	"github.com/omni-lang/omni/internal/parser"
	"github.com/omni-lang/omni/internal/passes"
	"github.com/omni-lang/omni/internal/types/checker"
)

func TestCompileSimpleProgram(t *testing.T) {
	libDir := filepath.Join("..", "..", "..", "native", "clift", "target", "release")
	libCandidates := []string{
		"libomni_clift.so",
		"libomni_clift.dylib",
		"omni_clift.dll",
	}
	hasLibrary := false
	for _, candidate := range libCandidates {
		if _, err := os.Stat(filepath.Join(libDir, candidate)); err == nil {
			hasLibrary = true
			break
		}
	}
	if !hasLibrary {
		t.Skip("skipping Cranelift smoke test: native bridge library not present")
	}

	const source = `import std

func main(): int {
    std.io.println("hello from cranelift")
    return 0
}
`
	const modulePath = "cranelift_smoke.omni"

	ast, err := parser.Parse(modulePath, source)
	if err != nil {
		t.Fatalf("parse failure: %v", err)
	}

	if err := checker.Check(modulePath, source, ast); err != nil {
		t.Fatalf("type check failure: %v", err)
	}

	mirModule, err := builder.BuildModule(ast)
	if err != nil {
		t.Fatalf("mir build failure: %v", err)
	}

	pipeline := passes.NewPipeline("cranelift-smoke")
	if _, err := pipeline.Run(*mirModule); err != nil {
		t.Fatalf("pipeline failure: %v", err)
	}

	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "smoke.o")

	if err := CompileModuleToObject(mirModule, outputPath); err != nil {
		t.Fatalf("CompileModuleToObject failed: %v", err)
	}

	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("expected object file at %s: %v", outputPath, err)
	}
}
