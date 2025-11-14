package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const eofRune = -1

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
}

// New constructs a lexer for the provided file name and contents.
func New(filename, input string) *Lexer {
	// Normalise newlines to simplify position tracking. This mirrors how the Go
	// toolchain handles Windows line endings.
	normalised := strings.ReplaceAll(input, "\r\n", "\n")
	normalised = strings.ReplaceAll(normalised, "\r", "\n")
	return &Lexer{
		filename: filename,
		input:    normalised,
		line:     1,
		column:   1,
		lines:    strings.Split(normalised, "\n"),
	}
}

// LexAll runs the lexer to completion returning all tokens including EOF.
func LexAll(filename, input string) ([]Token, error) {
	lx := New(filename, input)
	var tokens []Token
	for {
		tok, err := lx.NextToken()
		if err != nil {
			return nil, err
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
	return Token{}, l.errorf(startPos, "unexpected character %q", string(r))
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
	if kind, ok := keywords[lexeme]; ok {
		return l.emitTokenWithLexeme(kind, startPos, startOffset, lexeme), nil
	}
	return l.emitTokenWithLexeme(TokenIdentifier, startPos, startOffset, lexeme), nil
}

func (l *Lexer) scanNumber() (Token, error) {
	startPos, startOffset := l.mark()

	// Check for hex literal (0x...)
	if l.peek() == '0' && l.peekRuneAhead(1) == 'x' {
		l.advance() // consume '0'
		l.advance() // consume 'x'

		// Scan hex digits
		for {
			r := l.peek()
			if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') || r == '_' {
				l.advance()
			} else {
				break
			}
		}

		lexeme := l.slice(startOffset)
		if strings.HasSuffix(lexeme, "_") {
			return Token{}, l.errorf(startPos, "hex literal cannot end with underscore")
		}
		return l.emitTokenWithLexeme(TokenHexLiteral, startPos, startOffset, lexeme), nil
	}

	// Check for binary literal (0b...)
	if l.peek() == '0' && l.peekRuneAhead(1) == 'b' {
		l.advance() // consume '0'
		l.advance() // consume 'b'

		// Scan binary digits
		for {
			r := l.peek()
			if r == '0' || r == '1' || r == '_' {
				l.advance()
			} else {
				break
			}
		}

		lexeme := l.slice(startOffset)
		if strings.HasSuffix(lexeme, "_") {
			return Token{}, l.errorf(startPos, "binary literal cannot end with underscore")
		}
		return l.emitTokenWithLexeme(TokenBinaryLiteral, startPos, startOffset, lexeme), nil
	}

	// Regular decimal number
	hasDot := false
	hasExponent := false

	for {
		r := l.peek()
		switch {
		case unicode.IsDigit(r):
			l.advance()
		case r == '_' && l.peekAheadIsDigit():
			l.advance()
		case r == '.' && !hasDot && !hasExponent && unicode.IsDigit(l.peekRuneAhead(1)):
			hasDot = true
			l.advance()
		case (r == 'e' || r == 'E') && !hasExponent:
			hasExponent = true
			l.advance()
			// Check for optional sign after exponent
			if l.peek() == '+' || l.peek() == '-' {
				l.advance()
			}
		default:
			goto done
		}
	}
done:
	lexeme := l.slice(startOffset)
	if strings.HasSuffix(lexeme, "_") {
		return Token{}, l.errorf(startPos, "numeric literal cannot end with underscore")
	}
	if hasDot || hasExponent {
		return l.emitTokenWithLexeme(TokenFloatLiteral, startPos, startOffset, lexeme), nil
	}
	return l.emitTokenWithLexeme(TokenIntLiteral, startPos, startOffset, lexeme), nil
}

func (l *Lexer) scanString() (Token, error) {
	startPos, startOffset := l.mark()
	l.advance() // opening quote
	for {
		r := l.peek()
		switch r {
		case eofRune, '\n':
			return Token{}, l.errorf(startPos, "unterminated string literal")
		case '\\':
			l.advance()
			esc := l.peek()
			if esc == eofRune {
				return Token{}, l.errorf(startPos, "unterminated escape sequence")
			}
			l.advance()
		case '$':
			// Check for string interpolation: ${
			if l.peekRuneAhead(1) == '{' {
				// This is a string interpolation, not a regular string literal
				l.advance() // consume '$'
				l.advance() // consume '{'
				// Continue scanning until we find the closing quote
				for {
					r := l.peek()
					switch r {
					case eofRune, '\n':
						return Token{}, l.errorf(startPos, "unterminated string interpolation")
					case '\\':
						l.advance()
						esc := l.peek()
						if esc == eofRune {
							return Token{}, l.errorf(startPos, "unterminated escape sequence")
						}
						l.advance()
					case '"':
						l.advance()
						lexeme := l.slice(startOffset)
						return l.emitTokenWithLexeme(TokenStringInterpolation, startPos, startOffset, lexeme), nil
					default:
						l.advance()
					}
				}
			} else {
				l.advance()
			}
		case '"':
			l.advance()
			lexeme := l.slice(startOffset)
			return l.emitTokenWithLexeme(TokenStringLiteral, startPos, startOffset, lexeme), nil
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
		l.advance()
		esc := l.peek()
		if esc == eofRune {
			return Token{}, l.errorf(startPos, "unterminated escape sequence")
		}
		l.advance()
	} else {
		l.advance()
	}
	if l.peek() != '\'' {
		return Token{}, l.errorf(startPos, "char literal must contain exactly one character")
	}
	l.advance()
	lexeme := l.slice(startOffset)
	return l.emitTokenWithLexeme(TokenCharLiteral, startPos, startOffset, lexeme), nil
}

func (l *Lexer) skipTrivia() error {
	for {
		r := l.peek()
		switch r {
		case ' ', '\t':
			l.advance()
		case '\n':
			l.advance()
		case '/':
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
		case eofRune:
			return nil
		default:
			return nil
		}
	}
}

func (l *Lexer) skipBlockComment() error {
	depth := 1
	for depth > 0 {
		r := l.peek()
		if r == eofRune {
			pos, _ := l.mark()
			return l.errorf(pos, "unterminated block comment")
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
	off := l.offset
	for lookahead > 0 {
		if off >= len(l.input) {
			return eofRune
		}
		_, size := utf8.DecodeRuneInString(l.input[off:])
		off += size
		lookahead--
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
	r, size := utf8.DecodeRuneInString(l.input[l.offset:])
	l.offset += size
	if r == '\n' {
		l.line++
		l.column = 1
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
	message := strings.TrimSpace(strings.TrimSuffix(fmt.Sprintf(format, args...), "."))
	lineText := ""
	if pos.Line >= 1 && pos.Line <= len(l.lines) {
		lineText = l.lines[pos.Line-1]
	}
	return Diagnostic{
		File:    l.filename,
		Message: message,
		Hint:    "check the token around this location",
		Span:    Span{Start: pos, End: pos},
		Line:    lineText,
	}
}

func isIdentifierStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isIdentifierPart(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
