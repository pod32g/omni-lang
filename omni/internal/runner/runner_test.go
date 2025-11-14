package runner_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/omni-lang/omni/internal/runner"
)

func TestRunnerExecutesProgram(t *testing.T) {
	src := `func inc(x:int):int => x + 1
func main():int {
  let value:int = inc(41)
  return value
}
`

	dir := t.TempDir()
	path := filepath.Join(dir, "main.omni")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	res, err := runner.Execute(path, nil, false)
	if err != nil {
		t.Fatalf("runner execute failed: %v", err)
	}
	if res.Value != 42 {
		t.Fatalf("expected 42, got %v", res.Value)
	}
}
