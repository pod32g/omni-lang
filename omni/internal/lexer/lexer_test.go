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

// TestEvilLexerCases tests edge cases and error conditions that could cause
// hangs, incorrect tokenization, or poor error messages.
func TestEvilLexerCases(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectError    bool
		errorContains  string
		expectTokens   bool
		validateTokens func(t *testing.T, tokens []lexer.Token)
	}{
		// 1. Unterminated constructs
		{
			name:          "unterminated_nested_block_comment",
			input:         "/* nested /* comment",
			expectError:   true,
			errorContains: "unterminated block comment",
		},
		{
			name:          "unterminated_string_interpolation_no_brace",
			input:         `"hello ${foo`,
			expectError:   true,
			errorContains: "unterminated string interpolation",
		},
		{
			name:         "unterminated_string_interpolation_no_quote",
			input:        `"hello ${foo}"`,
			expectError:  false, // This should be valid
			expectTokens: true,
		},
		{
			name:          "unterminated_char_literal",
			input:         "'",
			expectError:   true,
			errorContains: "unterminated char literal",
		},
		{
			name:          "unterminated_char_escape",
			input:         "'\\",
			expectError:   true,
			errorContains: "unterminated escape sequence",
		},
		// 2. Numeric edge cases
		{
			name:          "float_exponent_no_digits_plus",
			input:         "1e+ foo",
			expectError:   true,
			errorContains: "exponent must be followed by at least one digit",
		},
		{
			name:          "float_exponent_no_digits_minus",
			input:         "2E-",
			expectError:   true,
			errorContains: "exponent must be followed by at least one digit",
		},
		{
			name:          "float_underscore_before_dot",
			input:         "123_.5",
			expectError:   true,
			errorContains: "numeric literal cannot end with underscore",
		},
		{
			name:         "float_underscore_after_dot",
			input:        "123._5",
			expectError:  false, // This might be valid, depends on spec
			expectTokens: true,
		},
		{
			name:          "hex_leading_underscore",
			input:         "0x_dead",
			expectError:   true,
			errorContains: "hex literal cannot start with underscore",
		},
		{
			name:          "hex_adjacent_underscores",
			input:         "0x__dead",
			expectError:   true,
			errorContains: "hex literal cannot have adjacent underscores",
		},
		{
			name:          "binary_adjacent_underscores",
			input:         "0b__01",
			expectError:   true,
			errorContains: "binary literal cannot have adjacent underscores",
		},
		{
			name:         "large_number_performance",
			input:        "999999999999999999999999",
			expectError:  false,
			expectTokens: true,
			validateTokens: func(t *testing.T, tokens []lexer.Token) {
				if len(tokens) < 2 {
					t.Fatalf("expected at least 2 tokens (number + EOF), got %d", len(tokens))
				}
				if tokens[0].Kind != lexer.TokenIntLiteral {
					t.Errorf("expected TokenIntLiteral, got %v", tokens[0].Kind)
				}
			},
		},
		// 3. Operator ambiguities
		{
			name:         "double_dot",
			input:        "a..b",
			expectError:  false,
			expectTokens: true,
			validateTokens: func(t *testing.T, tokens []lexer.Token) {
				// Should tokenize as: a, ., ., b
				if len(tokens) < 5 {
					t.Fatalf("expected at least 5 tokens, got %d", len(tokens))
				}
				// Check for two dots
				dotCount := 0
				for _, tok := range tokens {
					if tok.Kind == lexer.TokenDot {
						dotCount++
					}
				}
				if dotCount < 2 {
					t.Errorf("expected at least 2 dots, got %d", dotCount)
				}
			},
		},
		// 4. Strings/interpolation
		{
			name:         "interpolation_unterminated_quote",
			input:        `"${ "unterminated" }"`,
			expectError:  false, // Should be valid - the inner quote is escaped or part of expression
			expectTokens: true,
		},
		{
			name:         "interpolation_multiple",
			input:        `"${foo}bar${baz}"`,
			expectError:  false,
			expectTokens: true,
			validateTokens: func(t *testing.T, tokens []lexer.Token) {
				// Should have one TokenStringInterpolation
				interpCount := 0
				for _, tok := range tokens {
					if tok.Kind == lexer.TokenStringInterpolation {
						interpCount++
						if !strings.Contains(tok.Lexeme, "${foo}") {
							t.Errorf("expected interpolation to contain ${foo}, got %q", tok.Lexeme)
						}
						if !strings.Contains(tok.Lexeme, "${baz}") {
							t.Errorf("expected interpolation to contain ${baz}, got %q", tok.Lexeme)
						}
					}
				}
				if interpCount != 1 {
					t.Errorf("expected 1 TokenStringInterpolation, got %d", interpCount)
				}
			},
		},
		{
			name:         "interpolation_embedded_quotes",
			input:        `"say \"${user}\""`,
			expectError:  false,
			expectTokens: true,
			validateTokens: func(t *testing.T, tokens []lexer.Token) {
				// Should handle escaped quotes inside interpolation
				interpFound := false
				for _, tok := range tokens {
					if tok.Kind == lexer.TokenStringInterpolation {
						interpFound = true
						if !strings.Contains(tok.Lexeme, "${user}") {
							t.Errorf("expected interpolation to contain ${user}, got %q", tok.Lexeme)
						}
					}
				}
				if !interpFound {
					t.Error("expected TokenStringInterpolation token")
				}
			},
		},
		{
			name:         "interpolation_nested_braces",
			input:        `"${outer{inner}}"`,
			expectError:  false,
			expectTokens: true,
			validateTokens: func(t *testing.T, tokens []lexer.Token) {
				interpFound := false
				for _, tok := range tokens {
					if tok.Kind == lexer.TokenStringInterpolation {
						interpFound = true
						if !strings.Contains(tok.Lexeme, "${outer{inner}}") {
							t.Errorf("expected interpolation to contain nested braces, got %q", tok.Lexeme)
						}
					}
				}
				if !interpFound {
					t.Error("expected TokenStringInterpolation token")
				}
			},
		},
		// 5. Unicode/tabs
		{
			name:         "identifier_combining_marks",
			input:        "a\u0301", // a with combining acute accent
			expectError:  false,
			expectTokens: true,
			validateTokens: func(t *testing.T, tokens []lexer.Token) {
				// Should normalize to NFC
				if len(tokens) < 2 {
					t.Fatalf("expected at least 2 tokens, got %d", len(tokens))
				}
				if tokens[0].Kind != lexer.TokenIdentifier {
					t.Errorf("expected TokenIdentifier, got %v", tokens[0].Kind)
				}
			},
		},
		{
			name:         "tabs_column_reporting",
			input:        "\tlet x = 42",
			expectError:  false,
			expectTokens: true,
			validateTokens: func(t *testing.T, tokens []lexer.Token) {
				// Tab should expand to column 9 (next tab stop after column 1)
				// Find the 'let' token
				for _, tok := range tokens {
					if tok.Kind == lexer.TokenLet {
						// Column should be at a tab stop (9, 17, etc.)
						if tok.Span.Start.Column < 8 {
							t.Errorf("expected column >= 8 after tab, got %d", tok.Span.Start.Column)
						}
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := lexer.LexAll("test.omni", tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errorContains)
					return
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing %q, got %q", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.expectTokens && len(tokens) == 0 {
				t.Error("expected tokens but got none")
				return
			}

			if tt.validateTokens != nil {
				tt.validateTokens(t, tokens)
			}
		})
	}
}
