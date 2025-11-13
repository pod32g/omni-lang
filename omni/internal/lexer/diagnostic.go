package lexer

import (
	"fmt"
	"strings"
)

// Diagnostic conveys a lexing failure with caret formatted context.
type Diagnostic struct {
	File             string
	Message          string
	Hint             string
	Span             Span
	Line             string
	Context          []string
	ContextStartLine int
	Severity         Severity
	Category         string
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

	// Prefer multi-line context if available
	if len(d.Context) > 0 {
		for idx, line := range d.Context {
			lineNo := d.ContextStartLine + idx
			fmt.Fprintf(&b, "%6d | %s\n", lineNo, line)
			if lineNo == d.Span.Start.Line {
				highlight := buildHighlightLine(line, d.Span)
				fmt.Fprintf(&b, "%6s | %s\n", "", highlight)
			}
		}
	} else if d.Line != "" {
		// Backwards compatibility: single-line context
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

		if d.Span.End.Column > d.Span.Start.Column {
			width := d.Span.End.Column - d.Span.Start.Column
			for i := 0; i < width; i++ {
				b.WriteByte('~')
			}
		} else {
			b.WriteString("^~~")
		}
	}

	if strings.TrimSpace(d.Hint) != "" {
		b.WriteString("  hint: ")
		b.WriteString(d.Hint)
		b.WriteByte('\n')
	}

	return b.String()
}

func buildHighlightLine(line string, span Span) string {
	caretPos := span.Start.Column
	if caretPos < 1 {
		caretPos = 1
	}
	width := span.End.Column - span.Start.Column
	if width <= 0 {
		width = 1
	}
	if caretPos-1 > len(line) {
		caretPos = len(line) + 1
	}
	if caretPos-1+width > len(line) {
		width = len(line) - (caretPos - 1)
		if width < 1 {
			width = 1
		}
	}
	return strings.Repeat(" ", caretPos-1) + strings.Repeat("^", width)
}

// BuildContext extracts a small snippet of lines around the diagnostic span for display.
func BuildContext(lines []string, span Span) ([]string, int) {
	if span.Start.Line <= 0 || len(lines) == 0 {
		return nil, 0
	}
	start := span.Start.Line - 2
	if start < 0 {
		start = 0
	}
	end := span.End.Line + 1
	if end <= span.Start.Line {
		end = span.Start.Line + 1
	}
	if end > len(lines) {
		end = len(lines)
	}
	if start >= len(lines) {
		return nil, 0
	}
	if end <= start {
		end = start + 1
		if end > len(lines) {
			end = len(lines)
		}
	}
	context := append([]string{}, lines[start:end]...)
	return context, start + 1
}
