package lexer

import "fmt"

// Kind identifies the specific category of lexical token recognised by the
// OmniLang lexer. The explicit enumeration keeps golden fixtures stable as the
// language evolves.
type Kind int

const (
	TokenIllegal Kind = iota
	TokenEOF
	TokenIdentifier
	TokenIntLiteral
	TokenFloatLiteral
	TokenStringLiteral
	TokenStringInterpolation
	TokenCharLiteral
	TokenNullLiteral
	TokenHexLiteral
	TokenBinaryLiteral

	// Keywords
	TokenLet
	TokenVar
	TokenFunc
	TokenReturn
	TokenStruct
	TokenEnum
	TokenImport
	TokenAs
	TokenIf
	TokenElse
	TokenFor
	TokenIn
	TokenWhile
	TokenBreak
	TokenContinue
	TokenTrue
	TokenFalse
	TokenNew
	TokenDelete

	// Delimiters
	TokenLParen
	TokenRParen
	TokenLBrace
	TokenRBrace
	TokenLBracket
	TokenRBracket
	TokenComma
	TokenDot
	TokenColon
	TokenSemicolon

	// Operators
	TokenAssign
	TokenPlus
	TokenMinus
	TokenStar
	TokenSlash
	TokenPercent
	TokenBang
	TokenBangEqual
	TokenEqualEqual
	TokenLess
	TokenLessEqual
	TokenGreater
	TokenGreaterEqual
	TokenAndAnd
	TokenOrOr
	TokenPipe      // |
	TokenAmpersand // &
	TokenCaret     // ^
	TokenTilde     // ~
	TokenLShift    // <<
	TokenRShift    // >>
	TokenQuestion  // ?
	TokenPlusPlus
	TokenMinusMinus
	TokenArrow    // ->
	TokenFatArrow // =>
)

var kindNames = map[Kind]string{
	TokenIllegal:       "ILLEGAL",
	TokenEOF:           "EOF",
	TokenIdentifier:    "IDENT",
	TokenIntLiteral:         "INT",
	TokenFloatLiteral:       "FLOAT",
	TokenStringLiteral:      "STRING",
	TokenStringInterpolation: "STRING_INTERP",
	TokenCharLiteral:        "CHAR",
	TokenNullLiteral:   "NULL",
	TokenHexLiteral:    "HEX",
	TokenBinaryLiteral: "BINARY",
	TokenLet:           "LET",
	TokenVar:           "VAR",
	TokenFunc:          "FUNC",
	TokenReturn:        "RETURN",
	TokenStruct:        "STRUCT",
	TokenEnum:          "ENUM",
	TokenImport:        "IMPORT",
	TokenAs:            "AS",
	TokenIf:            "IF",
	TokenElse:          "ELSE",
	TokenFor:           "FOR",
	TokenIn:            "IN",
	TokenWhile:         "WHILE",
	TokenBreak:         "BREAK",
	TokenContinue:      "CONTINUE",
	TokenTrue:          "TRUE",
	TokenFalse:         "FALSE",
	TokenNew:           "NEW",
	TokenDelete:        "DELETE",
	TokenLParen:        "LPAREN",
	TokenRParen:        "RPAREN",
	TokenLBrace:        "LBRACE",
	TokenRBrace:        "RBRACE",
	TokenLBracket:      "LBRACKET",
	TokenRBracket:      "RBRACKET",
	TokenComma:         "COMMA",
	TokenDot:           "DOT",
	TokenColon:         "COLON",
	TokenSemicolon:     "SEMICOLON",
	TokenAssign:        "ASSIGN",
	TokenPlus:          "PLUS",
	TokenMinus:         "MINUS",
	TokenStar:          "STAR",
	TokenSlash:         "SLASH",
	TokenPercent:       "PERCENT",
	TokenBang:          "BANG",
	TokenBangEqual:     "BANG_EQUAL",
	TokenEqualEqual:    "EQUAL_EQUAL",
	TokenLess:          "LESS",
	TokenLessEqual:     "LESS_EQUAL",
	TokenGreater:       "GREATER",
	TokenGreaterEqual:  "GREATER_EQUAL",
	TokenAndAnd:        "AND_AND",
	TokenOrOr:          "OR_OR",
	TokenPipe:          "PIPE",
	TokenAmpersand:     "AMPERSAND",
	TokenCaret:         "CARET",
	TokenTilde:         "TILDE",
	TokenLShift:        "L_SHIFT",
	TokenRShift:        "R_SHIFT",
	TokenQuestion:      "QUESTION",
	TokenPlusPlus:      "PLUS_PLUS",
	TokenMinusMinus:    "MINUS_MINUS",
	TokenArrow:         "ARROW",
	TokenFatArrow:      "FAT_ARROW",
}

// String returns the stable textual representation for the token kind.
func (k Kind) String() string {
	if name, ok := kindNames[k]; ok {
		return name
	}
	return fmt.Sprintf("Kind(%d)", int(k))
}

// Position represents a 1-based line/column location in the source file.
type Position struct {
	Line   int
	Column int
}

// Span captures the half-open interval [Start, End) for a token.
type Span struct {
	Start Position
	End   Position
}

// Token carries the lexical classification and source metadata for a lexeme.
type Token struct {
	Kind   Kind
	Lexeme string
	Span   Span
}

// Format renders the token in a deterministic string form suited for golden
// tests.
func (t Token) Format() string {
	return fmt.Sprintf("%d:%d\t%s\t%q", t.Span.Start.Line, t.Span.Start.Column, t.Kind, t.Lexeme)
}
