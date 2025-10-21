package lexer_test

import (
	"testing"

	"github.com/omni-lang/omni/internal/lexer"
)

// TestLexerEdgeCases tests edge cases and error conditions
func TestLexerEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []lexer.Token
		hasError bool
	}{
		{
			name:  "empty_input",
			input: "",
			expected: []lexer.Token{
				{Kind: lexer.EOF, Lex: "", Span: lexer.SourceSpan{Line: 1, Col: 1}},
			},
		},
		{
			name:  "whitespace_only",
			input: "   \t\n  ",
			expected: []lexer.Token{
				{Kind: lexer.EOF, Lex: "", Span: lexer.SourceSpan{Line: 2, Col: 3}},
			},
		},
		{
			name:  "single_character",
			input: "a",
			expected: []lexer.Token{
				{Kind: lexer.IDENT, Lex: "a", Span: lexer.SourceSpan{Line: 1, Col: 1}},
				{Kind: lexer.EOF, Lex: "", Span: lexer.SourceSpan{Line: 1, Col: 2}},
			},
		},
		{
			name:  "numbers_edge_cases",
			input: "0 123 999999999",
			expected: []lexer.Token{
				{Kind: lexer.INT, Lex: "0", Span: lexer.SourceSpan{Line: 1, Col: 1}},
				{Kind: lexer.INT, Lex: "123", Span: lexer.SourceSpan{Line: 1, Col: 3}},
				{Kind: lexer.INT, Lex: "999999999", Span: lexer.SourceSpan{Line: 1, Col: 7}},
				{Kind: lexer.EOF, Lex: "", Span: lexer.SourceSpan{Line: 1, Col: 17}},
			},
		},
		{
			name:  "floats_edge_cases",
			input: "0.0 3.14 1.0e10 2.5e-5",
			expected: []lexer.Token{
				{Kind: lexer.FLOAT, Lex: "0.0", Span: lexer.SourceSpan{Line: 1, Col: 1}},
				{Kind: lexer.FLOAT, Lex: "3.14", Span: lexer.SourceSpan{Line: 1, Col: 5}},
				{Kind: lexer.FLOAT, Lex: "1.0e10", Span: lexer.SourceSpan{Line: 1, Col: 10}},
				{Kind: lexer.FLOAT, Lex: "2.5e-5", Span: lexer.SourceSpan{Line: 1, Col: 17}},
				{Kind: lexer.EOF, Lex: "", Span: lexer.SourceSpan{Line: 1, Col: 23}},
			},
		},
		{
			name:  "string_escapes",
			input: `"hello\nworld" "tab\there" "quote\"here"`,
			expected: []lexer.Token{
				{Kind: lexer.STRING, Lex: `"hello\nworld"`, Span: lexer.SourceSpan{Line: 1, Col: 1}},
				{Kind: lexer.STRING, Lex: `"tab\there"`, Span: lexer.SourceSpan{Line: 1, Col: 16}},
				{Kind: lexer.STRING, Lex: `"quote\"here"`, Span: lexer.SourceSpan{Line: 1, Col: 28}},
				{Kind: lexer.EOF, Lex: "", Span: lexer.SourceSpan{Line: 1, Col: 42}},
			},
		},
		{
			name:  "operators_combined",
			input: "++ -- == != <= >= && ||",
			expected: []lexer.Token{
				{Kind: lexer.PLUS, Lex: "+", Span: lexer.SourceSpan{Line: 1, Col: 1}},
				{Kind: lexer.PLUS, Lex: "+", Span: lexer.SourceSpan{Line: 1, Col: 2}},
				{Kind: lexer.MINUS, Lex: "-", Span: lexer.SourceSpan{Line: 1, Col: 4}},
				{Kind: lexer.MINUS, Lex: "-", Span: lexer.SourceSpan{Line: 1, Col: 5}},
				{Kind: lexer.EQEQ, Lex: "==", Span: lexer.SourceSpan{Line: 1, Col: 7}},
				{Kind: lexer.NEQ, Lex: "!=", Span: lexer.SourceSpan{Line: 1, Col: 10}},
				{Kind: lexer.LTE, Lex: "<=", Span: lexer.SourceSpan{Line: 1, Col: 13}},
				{Kind: lexer.GTE, Lex: ">=", Span: lexer.SourceSpan{Line: 1, Col: 16}},
				{Kind: lexer.ANDAND, Lex: "&&", Span: lexer.SourceSpan{Line: 1, Col: 19}},
				{Kind: lexer.OROR, Lex: "||", Span: lexer.SourceSpan{Line: 1, Col: 22}},
				{Kind: lexer.EOF, Lex: "", Span: lexer.SourceSpan{Line: 1, Col: 24}},
			},
		},
		{
			name:  "keywords_mixed_case",
			input: "func Func FUNC let Let LET",
			expected: []lexer.Token{
				{Kind: lexer.K_FUNC, Lex: "func", Span: lexer.SourceSpan{Line: 1, Col: 1}},
				{Kind: lexer.IDENT, Lex: "Func", Span: lexer.SourceSpan{Line: 1, Col: 6}},
				{Kind: lexer.IDENT, Lex: "FUNC", Span: lexer.SourceSpan{Line: 1, Col: 11}},
				{Kind: lexer.K_LET, Lex: "let", Span: lexer.SourceSpan{Line: 1, Col: 16}},
				{Kind: lexer.IDENT, Lex: "Let", Span: lexer.SourceSpan{Line: 1, Col: 20}},
				{Kind: lexer.IDENT, Lex: "LET", Span: lexer.SourceSpan{Line: 1, Col: 24}},
				{Kind: lexer.EOF, Lex: "", Span: lexer.SourceSpan{Line: 1, Col: 27}},
			},
		},
		{
			name:  "comments_mixed",
			input: "// line comment\n/* block comment */ code",
			expected: []lexer.Token{
				{Kind: lexer.IDENT, Lex: "code", Span: lexer.SourceSpan{Line: 2, Col: 22}},
				{Kind: lexer.EOF, Lex: "", Span: lexer.SourceSpan{Line: 2, Col: 26}},
			},
		},
		{
			name:  "nested_comments",
			input: "/* outer /* inner */ comment */",
			expected: []lexer.Token{
				{Kind: lexer.EOF, Lex: "", Span: lexer.SourceSpan{Line: 1, Col: 30}},
			},
		},
		{
			name:     "unterminated_string",
			input:    `"hello world`,
			hasError: true,
		},
		{
			name:     "unterminated_block_comment",
			input:    "/* hello world",
			hasError: true,
		},
		{
			name:     "invalid_character",
			input:    "hello @ world",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input, "test.omni")
			var tokens []lexer.Token
			var err error

			for {
				tok := l.Next()
				tokens = append(tokens, tok)
				if tok.Kind == lexer.EOF || tok.Kind == lexer.ILLEGAL {
					if tok.Kind == lexer.ILLEGAL {
						err = l.Error()
					}
					break
				}
			}

			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(tokens) != len(tt.expected) {
				t.Errorf("Expected %d tokens, got %d", len(tt.expected), len(tokens))
				t.Logf("Expected: %v", tt.expected)
				t.Logf("Got: %v", tokens)
				return
			}

			for i, expected := range tt.expected {
				actual := tokens[i]
				if actual.Kind != expected.Kind {
					t.Errorf("Token %d: expected kind %v, got %v", i, expected.Kind, actual.Kind)
				}
				if actual.Lex != expected.Lex {
					t.Errorf("Token %d: expected lexeme %q, got %q", i, expected.Lex, actual.Lex)
				}
				if actual.Span.Line != expected.Span.Line {
					t.Errorf("Token %d: expected line %d, got %d", i, expected.Span.Line, actual.Span.Line)
				}
				if actual.Span.Col != expected.Span.Col {
					t.Errorf("Token %d: expected col %d, got %d", i, expected.Span.Col, actual.Span.Col)
				}
			}
		})
	}
}

// TestLexerPeek tests the peek functionality
func TestLexerPeek(t *testing.T) {
	input := "func main() { return 42; }"
	l := lexer.New(input, "test.omni")

	// Test peek without consuming
	tok1 := l.Peek()
	tok2 := l.Peek()
	if tok1.Kind != tok2.Kind || tok1.Lex != tok2.Lex {
		t.Error("Peek should return the same token multiple times")
	}

	// Test peek after consuming
	tok3 := l.Next()
	if tok1.Kind != tok3.Kind || tok1.Lex != tok3.Lex {
		t.Error("Peek should return the same token as Next")
	}

	// Test peek after consuming
	tok4 := l.Peek()
	tok5 := l.Next()
	if tok4.Kind != tok5.Kind || tok4.Lex != tok5.Lex {
		t.Error("Peek should return the same token as Next after consuming")
	}
}

// TestLexerPositionTracking tests position tracking across different scenarios
func TestLexerPositionTracking(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []lexer.SourceSpan
	}{
		{
			name:  "single_line",
			input: "func main()",
			expected: []lexer.SourceSpan{
				{Line: 1, Col: 1},  // func
				{Line: 1, Col: 6},  // main
				{Line: 1, Col: 10}, // (
				{Line: 1, Col: 11}, // )
			},
		},
		{
			name:  "multi_line",
			input: "func\nmain()\n{",
			expected: []lexer.SourceSpan{
				{Line: 1, Col: 1}, // func
				{Line: 2, Col: 1}, // main
				{Line: 2, Col: 5}, // (
				{Line: 2, Col: 6}, // )
				{Line: 3, Col: 1}, // {
			},
		},
		{
			name:  "with_comments",
			input: "func // comment\nmain()",
			expected: []lexer.SourceSpan{
				{Line: 1, Col: 1}, // func
				{Line: 2, Col: 1}, // main
				{Line: 2, Col: 5}, // (
				{Line: 2, Col: 6}, // )
			},
		},
		{
			name:  "with_whitespace",
			input: "  func   main  (  )  ",
			expected: []lexer.SourceSpan{
				{Line: 1, Col: 3},  // func
				{Line: 1, Col: 9},  // main
				{Line: 1, Col: 15}, // (
				{Line: 1, Col: 18}, // )
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input, "test.omni")
			var spans []lexer.SourceSpan

			for {
				tok := l.Next()
				if tok.Kind == lexer.EOF {
					break
				}
				spans = append(spans, tok.Span)
			}

			if len(spans) != len(tt.expected) {
				t.Errorf("Expected %d spans, got %d", len(tt.expected), len(spans))
				return
			}

			for i, expected := range tt.expected {
				actual := spans[i]
				if actual.Line != expected.Line {
					t.Errorf("Span %d: expected line %d, got %d", i, expected.Line, actual.Line)
				}
				if actual.Col != expected.Col {
					t.Errorf("Span %d: expected col %d, got %d", i, expected.Col, actual.Col)
				}
			}
		})
	}
}

// TestLexerKeywords tests all keywords
func TestLexerKeywords(t *testing.T) {
	keywords := map[string]lexer.Kind{
		"import": lexer.K_IMPORT,
		"let":    lexer.K_LET,
		"var":    lexer.K_VAR,
		"func":   lexer.K_FUNC,
		"struct": lexer.K_STRUCT,
		"enum":   lexer.K_ENUM,
		"if":     lexer.K_IF,
		"else":   lexer.K_ELSE,
		"for":    lexer.K_FOR,
		"in":     lexer.K_IN,
		"true":   lexer.K_TRUE,
		"false":  lexer.K_FALSE,
	}

	for keyword, expectedKind := range keywords {
		t.Run(keyword, func(t *testing.T) {
			l := lexer.New(keyword, "test.omni")
			tok := l.Next()

			if tok.Kind != expectedKind {
				t.Errorf("Expected keyword %s to have kind %v, got %v", keyword, expectedKind, tok.Kind)
			}

			if tok.Lex != keyword {
				t.Errorf("Expected lexeme %q, got %q", keyword, tok.Lex)
			}
		})
	}
}

// TestLexerSymbols tests all symbols
func TestLexerSymbols(t *testing.T) {
	symbols := map[string]lexer.Kind{
		"(":  lexer.LPAREN,
		")":  lexer.RPAREN,
		"{":  lexer.LBRACE,
		"}":  lexer.RBRACE,
		"[":  lexer.LBRACK,
		"]":  lexer.RBRACK,
		"<":  lexer.LANGLE,
		">":  lexer.RANGLE,
		",":  lexer.COMMA,
		".":  lexer.DOT,
		":":  lexer.COLON,
		";":  lexer.SEMI,
		"=>": lexer.ARROW,
		"+":  lexer.PLUS,
		"-":  lexer.MINUS,
		"*":  lexer.STAR,
		"/":  lexer.SLASH,
		"%":  lexer.PERCENT,
		"!":  lexer.BANG,
		"=":  lexer.ASSIGN,
		"==": lexer.EQEQ,
		"!=": lexer.NEQ,
		"<":  lexer.LT,
		"<=": lexer.LTE,
		">":  lexer.GT,
		">=": lexer.GTE,
		"&&": lexer.ANDAND,
		"||": lexer.OROR,
	}

	for symbol, expectedKind := range symbols {
		t.Run(symbol, func(t *testing.T) {
			l := lexer.New(symbol, "test.omni")
			tok := l.Next()

			if tok.Kind != expectedKind {
				t.Errorf("Expected symbol %s to have kind %v, got %v", symbol, expectedKind, tok.Kind)
			}

			if tok.Lex != symbol {
				t.Errorf("Expected lexeme %q, got %q", symbol, tok.Lex)
			}
		})
	}
}

// TestLexerIdentifiers tests identifier recognition
func TestLexerIdentifiers(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"hello123", "hello123"},
		{"_hello", "_hello"},
		{"hello_world", "hello_world"},
		{"HelloWorld", "HelloWorld"},
		{"h123", "h123"},
		{"_", "_"},
		{"_123", "_123"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input, "test.omni")
			tok := l.Next()

			if tok.Kind != lexer.IDENT {
				t.Errorf("Expected IDENT, got %v", tok.Kind)
			}

			if tok.Lex != tt.expected {
				t.Errorf("Expected lexeme %q, got %q", tt.expected, tok.Lex)
			}
		})
	}
}

// TestLexerNumbers tests number recognition
func TestLexerNumbers(t *testing.T) {
	tests := []struct {
		input    string
		expected lexer.Kind
	}{
		{"0", lexer.INT},
		{"123", lexer.INT},
		{"999999999", lexer.INT},
		{"0.0", lexer.FLOAT},
		{"3.14", lexer.FLOAT},
		{"1.0e10", lexer.FLOAT},
		{"2.5e-5", lexer.FLOAT},
		{"1e5", lexer.FLOAT},
		{"1E5", lexer.FLOAT},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input, "test.omni")
			tok := l.Next()

			if tok.Kind != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, tok.Kind)
			}

			if tok.Lex != tt.input {
				t.Errorf("Expected lexeme %q, got %q", tt.input, tok.Lex)
			}
		})
	}
}

// TestLexerStrings tests string recognition
func TestLexerStrings(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, `"hello"`},
		{`"hello world"`, `"hello world"`},
		{`"hello\nworld"`, `"hello\nworld"`},
		{`"hello\tworld"`, `"hello\tworld"`},
		{`"hello\"world"`, `"hello\"world"`},
		{`"hello\\world"`, `"hello\\world"`},
		{`""`, `""`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input, "test.omni")
			tok := l.Next()

			if tok.Kind != lexer.STRING {
				t.Errorf("Expected STRING, got %v", tok.Kind)
			}

			if tok.Lex != tt.expected {
				t.Errorf("Expected lexeme %q, got %q", tt.expected, tok.Lex)
			}
		})
	}
}
