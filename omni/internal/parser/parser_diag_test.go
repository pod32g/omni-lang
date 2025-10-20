package parser_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/omni-lang/omni/internal/lexer"
	"github.com/omni-lang/omni/internal/parser"
)

func TestParserAccumulatesMultipleErrors(t *testing.T) {
	src := "func bad(x int) {}\nlet first:int =\nlet second:int =\n"

	mod, err := parser.Parse("multi.omni", src)
	if err == nil {
		t.Fatalf("expected aggregated error, got nil")
	}
	if mod == nil {
		t.Fatalf("expected non-nil module even when errors occur")
	}

	// Ensure multiple diagnostics are surfaced.
	count := strings.Count(err.Error(), "error:")
	if count < 2 {
		t.Fatalf("expected multiple diagnostics, got %d\n%s", count, err.Error())
	}

	// Each joined diagnostic should still be accessible via errors.As.
	var diag lexer.Diagnostic
	if !errors.As(err, &diag) {
		t.Fatalf("combined error should expose underlying diagnostics")
	}
}
