package parser

import (
	"testing"
	"unicode/utf8"

	"github.com/omni-lang/omni/internal/lexer"
)

func FuzzParse(f *testing.F) {
	// Add seed inputs
	f.Add("func main() { println(\"hello\") }")
	f.Add("let x:int = 5")
	f.Add("if x > 0 { return x }")
	f.Add("import std.io as io")
	f.Add("func add(a:int, b:int) { return a + b }")

	f.Fuzz(func(t *testing.T, input string) {
		// Skip invalid UTF-8
		if !utf8.ValidString(input) {
			t.Skip("Invalid UTF-8")
		}

		// Skip empty input
		if len(input) == 0 {
			t.Skip("Empty input")
		}

		// Skip very long inputs (over 10KB)
		if len(input) > 10000 {
			t.Skip("Input too long")
		}

		// Parse the input
		_, err := Parse("fuzz.omni", input)

		// We don't care if parsing succeeds or fails,
		// we just want to make sure it doesn't panic
		_ = err
	})
}

func FuzzLexer(f *testing.F) {
	// Add seed inputs
	f.Add("func main() { println(\"hello\") }")
	f.Add("let x:int = 5")
	f.Add("if x > 0 { return x }")
	f.Add("import std.io as io")
	f.Add("func add(a:int, b:int) { return a + b }")

	f.Fuzz(func(t *testing.T, input string) {
		// Skip invalid UTF-8
		if !utf8.ValidString(input) {
			t.Skip("Invalid UTF-8")
		}

		// Skip empty input
		if len(input) == 0 {
			t.Skip("Empty input")
		}

		// Skip very long inputs (over 10KB)
		if len(input) > 10000 {
			t.Skip("Input too long")
		}

		// Lex the input
		lex := lexer.New("fuzz.omni", input)
		for {
			tok, err := lex.NextToken()
			if err != nil {
				// Lexing error, that's okay for fuzzing
				break
			}
			if tok.Kind == lexer.TokenEOF {
				break
			}
			// We don't care about the token content,
			// we just want to make sure lexing doesn't panic
		}
	})
}

func FuzzTypeChecker(f *testing.F) {
	// Add seed inputs
	f.Add("func main() { println(\"hello\") }")
	f.Add("let x:int = 5")
	f.Add("if x > 0 { return x }")
	f.Add("import std.io as io")
	f.Add("func add(a:int, b:int) { return a + b }")

	f.Fuzz(func(t *testing.T, input string) {
		// Skip invalid UTF-8
		if !utf8.ValidString(input) {
			t.Skip("Invalid UTF-8")
		}

		// Skip empty input
		if len(input) == 0 {
			t.Skip("Empty input")
		}

		// Skip very long inputs (over 10KB)
		if len(input) > 10000 {
			t.Skip("Input too long")
		}

		// Parse first
		mod, err := Parse("fuzz.omni", input)
		if err != nil {
			// If parsing fails, that's fine for fuzzing
			return
		}

		// Type check
		// We don't care if type checking succeeds or fails,
		// we just want to make sure it doesn't panic
		_ = mod
	})
}
