package lexer

import (
	"fmt"
	"strings"
)

// Diagnostic conveys a lexing failure with caret formatted context.
type Diagnostic struct {
	File    string
	Message string
	Hint    string
	Span    Span
	Line    string
}

// Error satisfies the error interface using the canonical OmniLang diagnostic
// format.
func (d Diagnostic) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s:%d:%d: error: %s\n", d.File, d.Span.Start.Line, d.Span.Start.Column, d.Message)
	b.WriteString("  ")
	b.WriteString(d.Line)
	b.WriteByte('\n')
	b.WriteString("  ")
	caretPos := d.Span.Start.Column
	if caretPos < 1 {
		caretPos = 1
	}
	for i := 1; i < caretPos; i++ {
		b.WriteByte(' ')
	}
	b.WriteString("^~~")
	if strings.TrimSpace(d.Hint) != "" {
		b.WriteString(" hint: ")
		b.WriteString(d.Hint)
	}
	return b.String()
}
