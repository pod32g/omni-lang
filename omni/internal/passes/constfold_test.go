package passes_test

import (
	"testing"

	"github.com/omni-lang/omni/internal/mir"
	"github.com/omni-lang/omni/internal/passes"
)

func TestConstFoldArithmetic(t *testing.T) {
	fn := mir.NewFunction("main", "int", nil)
	block := fn.NewBlock("entry")

	v0 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:       v0,
		Op:       "const",
		Type:     "int",
		Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "40", Type: "int"}},
	})

	v1 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:       v1,
		Op:       "const",
		Type:     "int",
		Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "2", Type: "int"}},
	})

	sum := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:   sum,
		Op:   "add",
		Type: "int",
		Operands: []mir.Operand{
			{Kind: mir.OperandValue, Value: v0, Type: "int"},
			{Kind: mir.OperandValue, Value: v1, Type: "int"},
		},
	})

	block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: sum, Type: "int"}}}

	mod := &mir.Module{Functions: []*mir.Function{fn}}

	passes.ConstFold(mod)

	folded := block.Instructions[2]
	if folded.Op != "const" {
		t.Fatalf("expected folding to const, got %s", folded.Op)
	}
	if folded.Operands[0].Literal != "42" {
		t.Fatalf("expected folded literal 42, got %s", folded.Operands[0].Literal)
	}
}
