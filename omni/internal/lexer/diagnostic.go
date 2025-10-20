package lexer

import (
	"fmt"
	"strings"
)

// Diagnostic conveys a lexing failure with caret formatted context.
type Diagnostic struct {
	File     string
	Message  string
	Hint     string
	Span     Span
	Line     string
	Severity Severity
	Category string
}

// Severity represents the severity level of a diagnostic.
type Severity int

const (
	Error Severity = iota
	Warning
	Info
)

// Error satisfies the error interface using the canonical OmniLang diagnostic
// format.
func (d Diagnostic) Error() string {
	var b strings.Builder

	// Set default severity if not specified
	severity := d.Severity
	if severity == 0 {
		severity = Error
	}

	// Format severity string
	severityStr := "error"
	switch severity {
	case Warning:
		severityStr = "warning"
	case Info:
		severityStr = "info"
	}

	// Add category if specified
	categoryStr := ""
	if d.Category != "" {
		categoryStr = " [" + d.Category + "]"
	}

	fmt.Fprintf(&b, "%s:%d:%d: %s%s: %s\n", d.File, d.Span.Start.Line, d.Span.Start.Column, severityStr, categoryStr, d.Message)

	// Add source line with context
	if d.Line != "" {
		b.WriteString("  ")
		b.WriteString(d.Line)
		b.WriteByte('\n')

		// Add caret pointing to error location
		b.WriteString("  ")
		caretPos := d.Span.Start.Column
		if caretPos < 1 {
			caretPos = 1
		}
		for i := 1; i < caretPos; i++ {
			b.WriteByte(' ')
		}

		// Show range if span has width
		if d.Span.End.Column > d.Span.Start.Column {
			width := d.Span.End.Column - d.Span.Start.Column
			for i := 0; i < width; i++ {
				b.WriteByte('~')
			}
		} else {
			b.WriteString("^~~")
		}

		// Add hint if provided
		if strings.TrimSpace(d.Hint) != "" {
			b.WriteString(" hint: ")
			b.WriteString(d.Hint)
		}
	}

	return b.String()
}
