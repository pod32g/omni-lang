package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

const eofRune = -1

// keywords maps lowercase keyword strings to their token kinds.
// Keywords are case-sensitive and must be lowercase. Uppercase versions
// will be lexed as identifiers and should be rejected by the parser.
var keywords = map[string]Kind{
	"let":      TokenLet,
	"var":      TokenVar,
	"func":     TokenFunc,
	"return":   TokenReturn,
	"struct":   TokenStruct,
	"enum":     TokenEnum,
	"import":   TokenImport,
	"as":       TokenAs,
	"if":       TokenIf,
	"else":     TokenElse,
	"for":      TokenFor,
	"in":       TokenIn,
	"while":    TokenWhile,
	"break":    TokenBreak,
	"continue": TokenContinue,
	"true":     TokenTrue,
	"false":    TokenFalse,
	"null":     TokenNullLiteral,
	"new":      TokenNew,
	"delete":   TokenDelete,
	"try":      TokenTry,
	"catch":    TokenCatch,
	"finally":  TokenFinally,
	"throw":    TokenThrow,
	"type":     TokenType,
	"optional": TokenOptional,
	"async":    TokenAsync,
	"await":    TokenAwait,
}

// Lexer transforms a source buffer into a stream of tokens while tracking
// source locations for diagnostics.
type Lexer struct {
	filename string
	input    string
	offset   int
	line     int
	column   int
	lines    []string
	// Cache for peekRuneAhead optimization: maps rune offset to byte offset
	peekCache          map[int]int // runeOffset -> byteOffset
	lastPeekRuneOffset int         // Last rune offset we've cached
}

// New constructs a lexer for the provided file name and contents.
func New(filename, input string) *Lexer {
	// Normalise newlines to simplify position tracking. This mirrors how the Go
	// toolchain handles Windows line endings.
	normalised := strings.ReplaceAll(input, "\r\n", "\n")
	normalised = strings.ReplaceAll(normalised, "\r", "\n")
	// Preserve trailing newline by checking if input ends with \n
	lines := strings.Split(normalised, "\n")
	if normalised != "" && strings.HasSuffix(normalised, "\n") {
		// Add empty line to preserve trailing newline
		lines = append(lines, "")
	}
	return &Lexer{
		filename:           filename,
		input:              normalised,
		line:               1,
		column:             1,
		lines:              lines,
		peekCache:          make(map[int]int),
		lastPeekRuneOffset: 0,
	}
}

// LexAll runs the lexer to completion returning all tokens including EOF.
// If an error occurs, it returns the tokens lexed so far along with the error
// to enable better error reporting with partial context.
func LexAll(filename, input string) ([]Token, error) {
	lx := New(filename, input)
	var tokens []Token
	for {
		tok, err := lx.NextToken()
		if err != nil {
			// Return partial tokens for better error reporting
			return tokens, err
		}
		tokens = append(tokens, tok)
		if tok.Kind == TokenEOF {
			break
		}
	}
	return tokens, nil
}

// NextToken extracts the next token from the source stream.
func (l *Lexer) NextToken() (Token, error) {
	if err := l.skipTrivia(); err != nil {
		return Token{}, err
	}
	startPos, startOffset := l.mark()
	r := l.peek()
	if r == eofRune {
		return Token{Kind: TokenEOF, Lexeme: "", Span: Span{Start: startPos, End: startPos}}, nil
	}

	switch {
	case isIdentifierStart(r):
		return l.scanIdentifier()
	case unicode.IsDigit(r):
		return l.scanNumber()
	}

	switch r {
	case '"':
		return l.scanString()
	case '\'':
		return l.scanChar()
	case '(':
		l.advance()
		return l.emitToken(TokenLParen, startPos, startOffset)
	case ')':
		l.advance()
		return l.emitToken(TokenRParen, startPos, startOffset)
	case '{':
		l.advance()
		return l.emitToken(TokenLBrace, startPos, startOffset)
	case '}':
		l.advance()
		return l.emitToken(TokenRBrace, startPos, startOffset)
	case '[':
		l.advance()
		return l.emitToken(TokenLBracket, startPos, startOffset)
	case ']':
		l.advance()
		return l.emitToken(TokenRBracket, startPos, startOffset)
	case ',':
		l.advance()
		return l.emitToken(TokenComma, startPos, startOffset)
	case '.':
		l.advance()
		// Check for .. or ... operators (not currently in language spec, but reserved for future use)
		// For now, we only support single dot
		return l.emitToken(TokenDot, startPos, startOffset)
	case ':':
		l.advance()
		return l.emitToken(TokenColon, startPos, startOffset)
	case ';':
		l.advance()
		return l.emitToken(TokenSemicolon, startPos, startOffset)
	case '+':
		l.advance()
		if l.match('+') {
			return l.emitToken(TokenPlusPlus, startPos, startOffset)
		}
		return l.emitToken(TokenPlus, startPos, startOffset)
	case '-':
		l.advance()
		if l.match('-') {
			return l.emitToken(TokenMinusMinus, startPos, startOffset)
		}
		if l.match('>') {
			return l.emitToken(TokenArrow, startPos, startOffset)
		}
		return l.emitToken(TokenMinus, startPos, startOffset)
	case '*':
		l.advance()
		return l.emitToken(TokenStar, startPos, startOffset)
	case '/':
		l.advance()
		return l.emitToken(TokenSlash, startPos, startOffset)
	case '%':
		l.advance()
		return l.emitToken(TokenPercent, startPos, startOffset)
	case '=':
		l.advance()
		if l.match('=') {
			return l.emitToken(TokenEqualEqual, startPos, startOffset)
		}
		if l.match('>') {
			return l.emitToken(TokenFatArrow, startPos, startOffset)
		}
		return l.emitToken(TokenAssign, startPos, startOffset)
	case '!':
		l.advance()
		if l.match('=') {
			return l.emitToken(TokenBangEqual, startPos, startOffset)
		}
		return l.emitToken(TokenBang, startPos, startOffset)
	case '<':
		l.advance()
		if l.match('<') {
			return l.emitToken(TokenLShift, startPos, startOffset)
		}
		if l.match('=') {
			return l.emitToken(TokenLessEqual, startPos, startOffset)
		}
		return l.emitToken(TokenLess, startPos, startOffset)
	case '>':
		l.advance()
		if l.match('>') {
			return l.emitToken(TokenRShift, startPos, startOffset)
		}
		if l.match('=') {
			return l.emitToken(TokenGreaterEqual, startPos, startOffset)
		}
		return l.emitToken(TokenGreater, startPos, startOffset)
	case '&':
		l.advance()
		if l.match('&') {
			return l.emitToken(TokenAndAnd, startPos, startOffset)
		}
		return l.emitToken(TokenAmpersand, startPos, startOffset)
	case '^':
		l.advance()
		return l.emitToken(TokenCaret, startPos, startOffset)
	case '~':
		l.advance()
		return l.emitToken(TokenTilde, startPos, startOffset)
	case '?':
		l.advance()
		return l.emitToken(TokenQuestion, startPos, startOffset)
	case '|':
		l.advance()
		if l.match('|') {
			return l.emitToken(TokenOrOr, startPos, startOffset)
		}
		return l.emitToken(TokenPipe, startPos, startOffset)
	}

	// Unrecognised rune.
	l.advance()
	lexeme := l.slice(startOffset)
	if lexeme == "" {
		// Guard against zero-length lexemes (e.g., malformed UTF-8)
		lexeme = string(r)
		if lexeme == "" {
			lexeme = fmt.Sprintf("\\u%04x", r)
		}
	}
	return Token{}, l.errorf(startPos, "unexpected character %q (lexeme: %q)", string(r), lexeme)
}

func (l *Lexer) scanIdentifier() (Token, error) {
	startPos, startOffset := l.mark()
	for {
		r := l.peek()
		if !isIdentifierPart(r) {
			break
		}
		l.advance()
	}
	lexeme := l.slice(startOffset)
	// Normalize to NFC to ensure mixed forms of the same identifier are treated identically
	normalized := norm.NFC.String(lexeme)
	if kind, ok := keywords[normalized]; ok {
		return l.emitTokenWithLexeme(kind, startPos, startOffset, normalized), nil
	}
	return l.emitTokenWithLexeme(TokenIdentifier, startPos, startOffset, normalized), nil
}

func (l *Lexer) scanNumber() (Token, error) {
	startPos, startOffset := l.mark()

	// Check for hex literal (0x... or 0X...)
	if l.peek() == '0' && (l.peekRuneAhead(1) == 'x' || l.peekRuneAhead(1) == 'X') {
		l.advance() // consume '0'
		l.advance() // consume 'x' or 'X'

		// Scan hex digits with underscore validation
		lastWasUnderscore := false
		hasDigits := false
		for {
			r := l.peek()
			if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
				hasDigits = true
				lastWasUnderscore = false
				l.advance()
			} else if r == '_' {
				// Check for adjacent underscores first (peek ahead)
				if lastWasUnderscore || l.peekRuneAhead(1) == '_' {
					return Token{}, l.errorf(startPos, "hex literal cannot have adjacent underscores")
				}
				if !hasDigits {
					return Token{}, l.errorf(startPos, "hex literal cannot start with underscore")
				}
				lastWasUnderscore = true
				l.advance()
			} else {
				break
			}
		}

		lexeme := l.slice(startOffset)
		if lastWasUnderscore {
			return Token{}, l.errorf(startPos, "hex literal cannot end with underscore")
		}
		if !hasDigits {
			return Token{}, l.errorf(startPos, "hex literal must contain at least one digit")
		}
		// Strip underscores from hex literals so they can be parsed
		normalizedLexeme := strings.ReplaceAll(lexeme, "_", "")
		return l.emitTokenWithLexeme(TokenHexLiteral, startPos, startOffset, normalizedLexeme), nil
	}

	// Check for binary literal (0b... or 0B...)
	if l.peek() == '0' && (l.peekRuneAhead(1) == 'b' || l.peekRuneAhead(1) == 'B') {
		l.advance() // consume '0'
		l.advance() // consume 'b' or 'B'

		// Scan binary digits with underscore validation
		lastWasUnderscore := false
		hasDigits := false
		for {
			r := l.peek()
			if r == '0' || r == '1' {
				hasDigits = true
				lastWasUnderscore = false
				l.advance()
			} else if r == '_' {
				// Check for adjacent underscores first (peek ahead)
				if lastWasUnderscore || l.peekRuneAhead(1) == '_' {
					return Token{}, l.errorf(startPos, "binary literal cannot have adjacent underscores")
				}
				if !hasDigits {
					return Token{}, l.errorf(startPos, "binary literal cannot start with underscore")
				}
				lastWasUnderscore = true
				l.advance()
			} else {
				break
			}
		}

		lexeme := l.slice(startOffset)
		if lastWasUnderscore {
			return Token{}, l.errorf(startPos, "binary literal cannot end with underscore")
		}
		if !hasDigits {
			return Token{}, l.errorf(startPos, "binary literal must contain at least one digit")
		}
		// Strip underscores from binary literals so they can be parsed
		normalizedLexeme := strings.ReplaceAll(lexeme, "_", "")
		return l.emitTokenWithLexeme(TokenBinaryLiteral, startPos, startOffset, normalizedLexeme), nil
	}

	// Regular decimal number
	hasDot := false
	hasExponent := false
	lastWasUnderscore := false
	hasDigits := false
	inFractionalPart := false

	for {
		r := l.peek()
		switch {
		case r >= '0' && r <= '9': // Only ASCII digits
			hasDigits = true
			lastWasUnderscore = false
			l.advance()
		case r == '_':
			if lastWasUnderscore {
				return Token{}, l.errorf(startPos, "numeric literal cannot have adjacent underscores")
			}
			if !hasDigits && !inFractionalPart {
				return Token{}, l.errorf(startPos, "numeric literal cannot start with underscore")
			}
			// Check that underscore is followed by a digit
			if !l.peekAheadIsDigit() {
				// This will be caught later as trailing underscore, but we can break here
				lastWasUnderscore = true
				l.advance()
				goto done
			}
			lastWasUnderscore = true
			l.advance()
		case r == '.' && !hasDot && !hasExponent:
			if !hasDigits && !lastWasUnderscore {
				// Leading dot like .5 is not supported
				goto done
			}
			nextRune := l.peekRuneAhead(1)
			if nextRune < '0' || nextRune > '9' {
				// Not a float, might be method call
				goto done
			}
			hasDot = true
			inFractionalPart = true
			lastWasUnderscore = false
			l.advance()
		case (r == 'e' || r == 'E') && !hasExponent:
			hasExponent = true
			lastWasUnderscore = false
			l.advance()
			// Check for optional sign after exponent
			if l.peek() == '+' || l.peek() == '-' {
				l.advance()
			}
			// Require at least one digit after exponent (and optional sign)
			nextRune := l.peek()
			if nextRune < '0' || nextRune > '9' {
				return Token{}, l.errorf(startPos, "exponent must be followed by at least one digit")
			}
		default:
			goto done
		}
	}
done:
	lexeme := l.slice(startOffset)
	if lastWasUnderscore {
		return Token{}, l.errorf(startPos, "numeric literal cannot end with underscore")
	}
	// Strip underscores from numeric literals so they can be parsed by strconv
	normalizedLexeme := strings.ReplaceAll(lexeme, "_", "")
	if hasDot || hasExponent {
		return l.emitTokenWithLexeme(TokenFloatLiteral, startPos, startOffset, normalizedLexeme), nil
	}
	return l.emitTokenWithLexeme(TokenIntLiteral, startPos, startOffset, normalizedLexeme), nil
}

// validateEscapeSequence checks if an escape sequence is valid and returns an error if not.
// Valid escapes: \n, \t, \r, \\, \", \', \0, \xHH, \uHHHH
func (l *Lexer) validateEscapeSequence(esc rune, pos Position) error {
	switch esc {
	case 'n', 't', 'r', '\\', '"', '\'', '0':
		return nil
	case 'x':
		// Hex escape: \xHH
		if !((l.peekRuneAhead(1) >= '0' && l.peekRuneAhead(1) <= '9') ||
			(l.peekRuneAhead(1) >= 'a' && l.peekRuneAhead(1) <= 'f') ||
			(l.peekRuneAhead(1) >= 'A' && l.peekRuneAhead(1) <= 'F')) {
			return l.errorf(pos, "invalid hex escape sequence: \\x%c", esc)
		}
		if !((l.peekRuneAhead(2) >= '0' && l.peekRuneAhead(2) <= '9') ||
			(l.peekRuneAhead(2) >= 'a' && l.peekRuneAhead(2) <= 'f') ||
			(l.peekRuneAhead(2) >= 'A' && l.peekRuneAhead(2) <= 'F')) {
			return l.errorf(pos, "invalid hex escape sequence: \\x%c (requires two hex digits)", esc)
		}
		return nil
	case 'u':
		// Unicode escape: \uHHHH
		for i := 1; i <= 4; i++ {
			next := l.peekRuneAhead(i)
			if !((next >= '0' && next <= '9') ||
				(next >= 'a' && next <= 'f') ||
				(next >= 'A' && next <= 'F')) {
				return l.errorf(pos, "invalid unicode escape sequence: \\u%c (requires four hex digits)", esc)
			}
		}
		return nil
	default:
		return l.errorf(pos, "unknown escape sequence: \\%c (valid escapes: \\n, \\t, \\r, \\\\, \\\", \\', \\0, \\xHH, \\uHHHH)", esc)
	}
}

func (l *Lexer) scanString() (Token, error) {
	startPos, startOffset := l.mark()
	l.advance()               // opening quote
	braceDepth := 0           // Track depth of braces in interpolation expressions
	hasInterpolation := false // Track if we actually saw ${ (not escaped)

	for {
		r := l.peek()
		switch r {
		case eofRune, '\n':
			if braceDepth > 0 {
				return Token{}, l.errorf(startPos, "unterminated string interpolation (missing closing brace)")
			}
			return Token{}, l.errorf(startPos, "unterminated string literal")
		case '\\':
			// Capture position of backslash for error reporting
			backslashPos, _ := l.mark()
			l.advance()
			esc := l.peek()
			if esc == eofRune {
				return Token{}, l.errorf(backslashPos, "unterminated escape sequence")
			}
			// Validate escape sequence (pass backslash position, not startPos)
			if err := l.validateEscapeSequence(esc, backslashPos); err != nil {
				return Token{}, err
			}
			// Consume the escape character
			l.advance()
			// For hex and unicode escapes, consume additional characters
			if esc == 'x' {
				l.advance() // consume first hex digit
				l.advance() // consume second hex digit
			} else if esc == 'u' {
				for i := 0; i < 4; i++ {
					l.advance() // consume hex digits
				}
			}
			// Note: Escaped braces and quotes don't affect brace depth or string termination
		case '$':
			// Check for string interpolation: ${
			if l.peekRuneAhead(1) == '{' {
				// This is a string interpolation (not escaped)
				hasInterpolation = true
				l.advance()  // consume '$'
				l.advance()  // consume '{'
				braceDepth++ // Enter an interpolation expression
				// Now we need to handle nested lexical contexts (strings, comments) inside the expression
				if err := l.skipInterpolationExpression(); err != nil {
					return Token{}, err
				}
				braceDepth-- // Expression closed
			} else {
				l.advance()
			}
		case '"':
			// Only end the string if we're not inside any interpolation (braceDepth == 0)
			if braceDepth == 0 {
				l.advance()
				lexeme := l.slice(startOffset)
				// Only emit TokenStringInterpolation if we actually saw ${ (not escaped)
				if hasInterpolation {
					return l.emitTokenWithLexeme(TokenStringInterpolation, startPos, startOffset, lexeme), nil
				}
				return l.emitTokenWithLexeme(TokenStringLiteral, startPos, startOffset, lexeme), nil
			}
			// Inside interpolation, treat quote as regular character
			l.advance()
		default:
			l.advance()
		}
	}
}

func (l *Lexer) scanChar() (Token, error) {
	startPos, startOffset := l.mark()
	l.advance() // opening quote
	r := l.peek()
	if r == eofRune || r == '\n' {
		return Token{}, l.errorf(startPos, "unterminated char literal")
	}
	if r == '\\' {
		// Capture position of backslash for error reporting
		backslashPos, _ := l.mark()
		l.advance()
		esc := l.peek()
		if esc == eofRune {
			return Token{}, l.errorf(backslashPos, "unterminated escape sequence")
		}
		// Validate escape sequence (pass backslash position, not startPos)
		if err := l.validateEscapeSequence(esc, backslashPos); err != nil {
			return Token{}, err
		}
		// Consume the escape character
		l.advance()
		// For hex and unicode escapes, consume additional characters
		if esc == 'x' {
			l.advance() // consume first hex digit
			l.advance() // consume second hex digit
		} else if esc == 'u' {
			for i := 0; i < 4; i++ {
				l.advance() // consume hex digits
			}
		}
	} else {
		// Regular character (may be multi-byte Unicode)
		l.advance()
	}
	if l.peek() != '\'' {
		return Token{}, l.errorf(startPos, "char literal must contain exactly one character or escape sequence")
	}
	l.advance()
	lexeme := l.slice(startOffset)
	return l.emitTokenWithLexeme(TokenCharLiteral, startPos, startOffset, lexeme), nil
}

func (l *Lexer) skipTrivia() error {
	// Skip UTF-8 BOM if present (U+FEFF)
	if l.offset == 0 && len(l.input) >= 3 && l.input[0] == 0xEF && l.input[1] == 0xBB && l.input[2] == 0xBF {
		l.offset += 3
	}

	for {
		r := l.peek()
		switch {
		case r == '\n':
			l.advance()
		case unicode.IsSpace(r):
			// Handle all Unicode whitespace (space, tab, form-feed, non-breaking space, etc.)
			l.advance()
		case r == '/':
			if l.peekRuneAhead(1) == '/' {
				l.advance()
				l.advance()
				for {
					r2 := l.peek()
					if r2 == '\n' || r2 == eofRune {
						break
					}
					l.advance()
				}
			} else if l.peekRuneAhead(1) == '*' {
				l.advance()
				l.advance()
				if err := l.skipBlockComment(); err != nil {
					return err
				}
			} else {
				return nil
			}
		case r == eofRune:
			return nil
		default:
			return nil
		}
	}
}

// skipInterpolationExpression skips over an interpolation expression ${...}
// It properly handles nested strings, comments, and braces
func (l *Lexer) skipInterpolationExpression() error {
	depth := 1 // We're already inside one {
	for depth > 0 {
		r := l.peek()
		switch r {
		case eofRune, '\n':
			return l.errorf(l.position(), "unterminated string interpolation (missing closing brace)")
		case '"':
			// Skip over string literal inside interpolation
			l.advance() // opening quote
			for {
				r2 := l.peek()
				if r2 == eofRune || r2 == '\n' {
					return l.errorf(l.position(), "unterminated string literal in interpolation")
				}
				if r2 == '\\' {
					l.advance() // backslash
					if l.peek() == eofRune {
						return l.errorf(l.position(), "unterminated escape sequence")
					}
					l.advance() // escaped char
					// Handle hex/unicode escapes
					if l.peek() == 'x' {
						l.advance()
						if l.peek() == eofRune {
							return l.errorf(l.position(), "incomplete hex escape")
						}
						l.advance()
					} else if l.peek() == 'u' {
						for i := 0; i < 4; i++ {
							if l.peek() == eofRune {
								return l.errorf(l.position(), "incomplete unicode escape")
							}
							l.advance()
						}
					}
				} else if r2 == '"' {
					l.advance() // closing quote
					break
				} else {
					l.advance()
				}
			}
		case '\'':
			// Skip over char literal inside interpolation
			l.advance() // opening quote
			if l.peek() == '\\' {
				l.advance() // backslash
				if l.peek() == eofRune {
					return l.errorf(l.position(), "unterminated escape sequence")
				}
				l.advance() // escaped char
				// Handle hex/unicode escapes
				if l.peek() == 'x' {
					l.advance()
					if l.peek() == eofRune {
						return l.errorf(l.position(), "incomplete hex escape")
					}
					l.advance()
				} else if l.peek() == 'u' {
					for i := 0; i < 4; i++ {
						if l.peek() == eofRune {
							return l.errorf(l.position(), "incomplete unicode escape")
						}
						l.advance()
					}
				}
			} else {
				l.advance() // regular char
			}
			if l.peek() != '\'' {
				return l.errorf(l.position(), "unterminated char literal in interpolation")
			}
			l.advance() // closing quote
		case '/':
			// Check for comments
			if l.peekRuneAhead(1) == '/' {
				// Line comment
				l.advance()
				l.advance()
				for {
					r2 := l.peek()
					if r2 == '\n' || r2 == eofRune {
						break
					}
					l.advance()
				}
			} else if l.peekRuneAhead(1) == '*' {
				// Block comment
				l.advance()
				l.advance()
				if err := l.skipBlockComment(); err != nil {
					return err
				}
			} else {
				l.advance()
			}
		case '{':
			depth++
			l.advance()
		case '}':
			depth--
			if depth > 0 {
				l.advance()
			} else {
				// Found matching closing brace
				l.advance()
				return nil
			}
		default:
			l.advance()
		}
	}
	return nil
}

func (l *Lexer) skipBlockComment() error {
	commentStartPos, _ := l.mark() // Record start position for error reporting
	depth := 1
	for depth > 0 {
		r := l.peek()
		if r == eofRune {
			// Use comment start position, not current position
			return l.errorf(commentStartPos, "unterminated block comment")
		}
		if r == '/' && l.peekRuneAhead(1) == '*' {
			l.advance()
			l.advance()
			depth++
			continue
		}
		if r == '*' && l.peekRuneAhead(1) == '/' {
			l.advance()
			l.advance()
			depth--
			continue
		}
		l.advance()
	}
	return nil
}

func (l *Lexer) emitToken(kind Kind, start Position, startOffset int) (Token, error) {
	lexeme := l.slice(startOffset)
	return Token{Kind: kind, Lexeme: lexeme, Span: Span{Start: start, End: l.position()}}, nil
}

func (l *Lexer) emitTokenWithLexeme(kind Kind, start Position, startOffset int, lexeme string) Token {
	// Guard against zero-length lexemes
	if lexeme == "" {
		// Fallback to slice if lexeme is empty
		lexeme = l.slice(startOffset)
		if lexeme == "" {
			// Still empty, use a placeholder
			lexeme = "<empty>"
		}
	}
	return Token{Kind: kind, Lexeme: lexeme, Span: Span{Start: start, End: l.position()}}
}

func (l *Lexer) peek() rune {
	if l.offset >= len(l.input) {
		return eofRune
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.offset:])
	return r
}

func (l *Lexer) peekRuneAhead(lookahead int) rune {
	if lookahead <= 0 {
		return l.peek()
	}

	// Use cache if available
	runesFromOffset := lookahead
	if cachedByteOffset, ok := l.peekCache[runesFromOffset]; ok {
		if cachedByteOffset >= len(l.input) {
			return eofRune
		}
		r, _ := utf8.DecodeRuneInString(l.input[cachedByteOffset:])
		return r
	}

	// Decode runes incrementally, caching intermediate positions
	off := l.offset
	runesDecoded := 0

	// Start from last cached position if available
	if l.lastPeekRuneOffset > 0 && l.lastPeekRuneOffset < runesFromOffset {
		if cachedOff, ok := l.peekCache[l.lastPeekRuneOffset]; ok {
			off = cachedOff
			runesDecoded = l.lastPeekRuneOffset
		}
	}

	// Decode remaining runes
	for runesDecoded < runesFromOffset {
		if off >= len(l.input) {
			// Cache this position
			l.peekCache[runesFromOffset] = off
			l.lastPeekRuneOffset = runesFromOffset
			return eofRune
		}
		_, size := utf8.DecodeRuneInString(l.input[off:])
		runesDecoded++
		// Cache intermediate positions for future lookups
		if runesDecoded <= 4 { // Cache first 4 runes ahead (common case)
			l.peekCache[runesDecoded] = off + size
		}
		off += size
	}

	// Cache the final position
	l.peekCache[runesFromOffset] = off
	if runesFromOffset > l.lastPeekRuneOffset {
		l.lastPeekRuneOffset = runesFromOffset
	}

	if off >= len(l.input) {
		return eofRune
	}
	r, _ := utf8.DecodeRuneInString(l.input[off:])
	return r
}

func (l *Lexer) peekAheadIsDigit() bool {
	r := l.peekRuneAhead(1)
	return unicode.IsDigit(r)
}

func (l *Lexer) advance() rune {
	if l.offset >= len(l.input) {
		return eofRune
	}
	// Clear peek cache by deleting entries instead of reallocating
	// This avoids O(n) allocations during lexing
	for k := range l.peekCache {
		delete(l.peekCache, k)
	}
	l.lastPeekRuneOffset = 0

	r, size := utf8.DecodeRuneInString(l.input[l.offset:])
	l.offset += size
	if r == '\n' {
		l.line++
		l.column = 1
	} else if r == '\t' {
		// Expand tabs to next tab stop (typically 8 columns)
		// Tab moves to the next tab stop, which is at positions 1, 9, 17, 25, etc.
		// Use proper 1-based tab stop math: next = ((col-1)/width+1)*width + 1
		tabWidth := 8
		nextTabStop := ((l.column-1)/tabWidth+1)*tabWidth + 1
		l.column = nextTabStop
	} else {
		l.column++
	}
	return r
}

func (l *Lexer) match(expected rune) bool {
	if l.peek() != expected {
		return false
	}
	l.advance()
	return true
}

func (l *Lexer) mark() (Position, int) {
	return Position{Line: l.line, Column: l.column}, l.offset
}

func (l *Lexer) slice(startOffset int) string {
	end := l.offset
	if end > len(l.input) {
		end = len(l.input)
	}
	return l.input[startOffset:end]
}

func (l *Lexer) position() Position {
	return Position{Line: l.line, Column: l.column}
}

func (l *Lexer) errorf(pos Position, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	// Preserve trailing punctuation (periods, exclamation marks, question marks)
	// Only trim trailing period if it was added by TrimSuffix in the old code
	message = strings.TrimSpace(message)
	lineText := ""
	if pos.Line >= 1 && pos.Line <= len(l.lines) {
		lineText = l.lines[pos.Line-1]
	}
	// Provide more specific hints based on error context
	hint := "check the token around this location"
	if strings.Contains(message, "unterminated") {
		hint = "check for missing closing delimiter"
	} else if strings.Contains(message, "escape") {
		hint = "check escape sequence syntax"
	} else if strings.Contains(message, "underscore") {
		hint = "underscores in numeric literals must be between digits"
	} else if strings.Contains(message, "exponent") {
		hint = "exponent must be followed by at least one digit"
	}
	return Diagnostic{
		File:    l.filename,
		Message: message,
		Hint:    hint,
		Span:    Span{Start: pos, End: pos},
		Line:    lineText,
	}
}

func isIdentifierStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isIdentifierPart(r rune) bool {
	// Allow letters, digits, underscores, and Unicode marks (combining marks)
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsMark(r)
}
