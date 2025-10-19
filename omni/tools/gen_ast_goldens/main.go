package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/omni-lang/omni/internal/ast"
	"github.com/omni-lang/omni/internal/parser"
)

type caseSpec struct {
	name   string
	source string
}

func main() {
	cases := buildCases()
	dir := filepath.Join("tests", "goldens", "ast")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		panic(err)
	}

	for _, c := range cases {
		omniPath := filepath.Join(dir, c.name+".omni")
		astPath := filepath.Join(dir, c.name+".ast")
		src := normalizeSource(c.source)
		if err := os.WriteFile(omniPath, []byte(src), 0o644); err != nil {
			panic(fmt.Errorf("write %s: %w", omniPath, err))
		}
		mod, err := parser.Parse(omniPath, src)
		if err != nil {
			panic(fmt.Errorf("parse %s: %w", omniPath, err))
		}
		output := ast.Print(mod)
		if err := os.WriteFile(astPath, []byte(output), 0o644); err != nil {
			panic(fmt.Errorf("write %s: %w", astPath, err))
		}
	}
}

func normalizeSource(src string) string {
	trimmed := strings.TrimSpace(src)
	if !strings.HasSuffix(trimmed, "\n") {
		trimmed += "\n"
	}
	return trimmed
}

func buildCases() []caseSpec {
	cases := make([]caseSpec, 0, 100)

	// Category A: arrow functions
	for i := 1; i <= 10; i++ {
		cases = append(cases, caseSpec{
			name:   fmt.Sprintf("func_arrow_%02d", i),
			source: fmt.Sprintf("func f%d(x:int):int => x + %d", i, i),
		})
	}

	// Category B: block functions with returns
	for i := 1; i <= 10; i++ {
		cases = append(cases, caseSpec{
			name: fmt.Sprintf("func_block_%02d", i),
			source: fmt.Sprintf(`func sum%d(a:int, b:int):int {
  return a + b + %d
}`, i, i),
		})
	}

	// Category C: top-level let/var declarations
	for i := 1; i <= 5; i++ {
		cases = append(cases, caseSpec{
			name:   fmt.Sprintf("decl_bind_%02d", i),
			source: fmt.Sprintf("let value%d:int = %d\nvar counter%d:int = value%d + 1", i, i, i, i),
		})
	}

	// Category D: struct declarations
	for i := 1; i <= 10; i++ {
		cases = append(cases, caseSpec{
			name: fmt.Sprintf("decl_struct_%02d", i),
			source: fmt.Sprintf(`struct Pair%d {
  left:int
  right:int
}

func makePair%d():Pair%d {
  return Pair%d{ left:%d, right:%d }
}`, i, i, i, i, i, i+1),
		})
	}

	// Category E: enum declarations
	colors := [][]string{
		{"ALPHA", "BETA", "GAMMA"},
		{"SPRING", "SUMMER", "FALL", "WINTER"},
		{"RED", "GREEN", "BLUE"},
		{"ONE", "TWO"},
		{"NORTH", "SOUTH", "EAST", "WEST"},
		{"LOW", "MEDIUM", "HIGH"},
		{"SMALL", "MEDIUM", "LARGE"},
		{"COPPER", "SILVER", "GOLD"},
		{"EARTH", "WIND", "FIRE"},
		{"TRUE", "FALSE"},
	}
	for i, variants := range colors {
		cases = append(cases, caseSpec{
			name: fmt.Sprintf("decl_enum_%02d", i+1),
			source: fmt.Sprintf(`enum Enum%d { %s }

func chooseEnum%d():Enum%d {
  return Enum%d.%s
}`, i+1, strings.Join(variants, " "), i+1, i+1, i+1, variants[0]),
		})
	}

	// Category F: range for loops
	for i := 1; i <= 10; i++ {
		cases = append(cases, caseSpec{
			name: fmt.Sprintf("loop_range_%02d", i),
			source: fmt.Sprintf(`func iterate%d(items:array<int>):int {
  for item in items {
    return item
  }
  return %d
}`, i, i),
		})
	}

	// Category G: classic for loops
	for i := 1; i <= 10; i++ {
		cases = append(cases, caseSpec{
			name: fmt.Sprintf("loop_classic_%02d", i),
			source: fmt.Sprintf(`func loop%d():int {
  for i:int = 0; i < %d; i++ {
    return i
  }
  return %d
}`, i, i, i),
		})
	}

	// Category H: if/else blocks
	for i := 1; i <= 10; i++ {
		cases = append(cases, caseSpec{
			name: fmt.Sprintf("cond_if_%02d", i),
			source: fmt.Sprintf(`func check%d(v:int):int {
  if v > %d {
    return v
  } else {
    return %d
  }
}`, i, i, i-1),
		})
	}

	// Category I: map literals
	for i := 1; i <= 10; i++ {
		cases = append(cases, caseSpec{
			name: fmt.Sprintf("literal_map_%02d", i),
			source: fmt.Sprintf(`func maps%d():map<string,int> {
  return { "a":%d, "b":%d }
}`, i, i, i+1),
		})
	}

	// Category J: array and struct literals combined
	for i := 1; i <= 10; i++ {
		cases = append(cases, caseSpec{
			name: fmt.Sprintf("literal_struct_%02d", i),
			source: fmt.Sprintf(`struct Node%d {
  value:int
}

func makeNode%d():array<Node%d> {
  return [Node%d{ value:%d }, Node%d{ value:%d }]
}`, i, i, i, i, i, i, i+1),
		})
	}

	// Additional mixed cases with imports and control flow
	for i := 1; i <= 5; i++ {
		cases = append(cases, caseSpec{
			name: fmt.Sprintf("mixed_case_%02d", i),
			source: fmt.Sprintf(`import std.io

let threshold%d:int = %d

func process%d(values:array<int>):int {
  var sum:int = 0
  for value in values {
    if value > threshold%d {
      sum = sum + value
    }
  }
  return sum
}`, i, i*10, i, i),
		})
	}

	if len(cases) != 100 {
		panic(fmt.Sprintf("expected 100 cases, got %d", len(cases)))
	}

	return cases
}
