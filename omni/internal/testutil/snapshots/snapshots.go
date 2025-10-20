package snapshots

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// CompareText verifies that the actual output matches the golden file when
// trailing whitespace differences are ignored.
func CompareText(t *testing.T, actual string, goldenPath string) {
	t.Helper()

	// Check if we should update golden files
	if os.Getenv("UPDATE_GOLDENS") == "1" {
		err := os.WriteFile(goldenPath, []byte(actual), 0644)
		if err != nil {
			t.Fatalf("write golden %s: %v", goldenPath, err)
		}
		t.Logf("Updated golden file: %s", goldenPath)
		return
	}

	data, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden %s: %v", goldenPath, err)
	}

	expected := string(data)
	if normalize(expected) == normalize(actual) {
		return
	}

	t.Fatalf("golden mismatch for %s\n--- expected ---\n%s\n--- actual ---\n%s\n", filepath.Base(goldenPath), expected, actual)
}

func normalize(s string) string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \t\r")
	}
	return strings.Join(lines, "\n")
}
