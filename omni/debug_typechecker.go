package main

import (
	"fmt"

	"github.com/omni-lang/omni/internal/ast"
	"github.com/omni-lang/omni/internal/parser"
	"github.com/omni-lang/omni/internal/types/checker"
)

func main() {
	source := `func identity<T>(x:T):T {
    return x
}`

	mod, err := parser.Parse("debug.omni", source)
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	fmt.Printf("Parsed module: %+v\n", mod)

	// Check if function has type parameters
	if len(mod.Decls) > 0 {
		if fn, ok := mod.Decls[0].(*ast.FuncDecl); ok {
			fmt.Printf("Function: %s\n", fn.Name)
			fmt.Printf("Type params: %+v\n", fn.TypeParams)
			fmt.Printf("Params: %+v\n", fn.Params)
			fmt.Printf("Return: %+v\n", fn.Return)
		}
	}

	// Create a custom checker to debug
	c := &checker.Checker{}

	err = checker.Check("debug.omni", source, mod)
	if err != nil {
		fmt.Printf("Type check error: %v\n", err)
	} else {
		fmt.Println("Type check success")
	}
}
