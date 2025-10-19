package main

import (
	"fmt"
	"os"

	"github.com/omni-lang/omni/internal/lexer"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: debug_tokens <path>")
		os.Exit(2)
	}
	path := os.Args[1]
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	tokens, err := lexer.LexAll(path, string(data))
	if err != nil {
		panic(err)
	}
	for _, tok := range tokens {
		fmt.Printf("%s %q (%d:%d)\n", tok.Kind, tok.Lexeme, tok.Span.Start.Line, tok.Span.Start.Column)
	}
}
