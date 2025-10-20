package checker_test

import (
	"testing"

	"github.com/omni-lang/omni/internal/parser"
	"github.com/omni-lang/omni/internal/types/checker"
)

func TestTypeCheckerAllowsValidProgram(t *testing.T) {
	src := `struct Point {
  x:float
  y:float
}

var scale:float = 2.0

func length(p:Point):float {
  let sqx:float = p.x * p.x
  let sqy:float = p.y * p.y
  return sqx + sqy
}

func main():int {
  let origin:Point = Point{ x:0.0, y:0.0 }
  var count:int = 0
  for i:int = 0; i < 10; i++ {
    count = count + i
  }
  if count > 5 {
    count = count - 1
  }
  return count
}
`

	mod, err := parser.Parse("positive.omni", src)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if err := checker.Check("positive.omni", src, mod); err != nil {
		t.Fatalf("type checker reported error on valid program:\n%s", err)
	}
}
