package main

import (
	"fmt"

	"github.com/omni-lang/omni/internal/lexer"
	"github.com/omni-lang/omni/internal/parser"
)

func main() {
	source := `func identity<T>(x:T) {
    return x
}`

	lex := lexer.New("debug.omni", source)
	for {
		tok, err := lex.NextToken()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			break
		}
		fmt.Printf("Token: %s (%s)\n", tok.Kind, tok.Lexeme)
		if tok.Kind == lexer.TokenEOF {
			break
		}
	}

	fmt.Println("\n--- Parsing ---")
	mod, err := parser.Parse("debug.omni", source)
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
	} else {
		fmt.Printf("Parse success: %+v\n", mod)
	}
}
