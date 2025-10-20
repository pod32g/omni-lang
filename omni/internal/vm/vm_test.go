package vm_test

import (
	"testing"

	"github.com/omni-lang/omni/internal/mir"
	"github.com/omni-lang/omni/internal/vm"
)

func TestExecuteSimpleAddition(t *testing.T) {
	fn := mir.NewFunction("main", "int", nil)
	block := fn.NewBlock("entry")

	v0 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:   v0,
		Op:   "const",
		Type: "int",
		Operands: []mir.Operand{
			{Kind: mir.OperandLiteral, Literal: "40", Type: "int"},
		},
	})

	v1 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:   v1,
		Op:   "const",
		Type: "int",
		Operands: []mir.Operand{
			{Kind: mir.OperandLiteral, Literal: "2", Type: "int"},
		},
	})

	v2 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:   v2,
		Op:   "add",
		Type: "int",
		Operands: []mir.Operand{
			{Kind: mir.OperandValue, Value: v0, Type: "int"},
			{Kind: mir.OperandValue, Value: v1, Type: "int"},
		},
	})

	block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: v2, Type: "int"}}}

	mod := &mir.Module{Functions: []*mir.Function{fn}}

	res, err := vm.Execute(mod, "main")
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if res.Value != 42 {
		t.Fatalf("expected 42, got %v", res.Value)
	}
}

func TestExecuteConditional(t *testing.T) {
	fn := mir.NewFunction("main", "int", nil)
	entry := fn.NewBlock("entry")

	aVal := fn.NextValue()
	entry.Instructions = append(entry.Instructions, mir.Instruction{
		ID:   aVal,
		Op:   "const",
		Type: "int",
		Operands: []mir.Operand{
			{Kind: mir.OperandLiteral, Literal: "7", Type: "int"},
		},
	})
	bVal := fn.NextValue()
	entry.Instructions = append(entry.Instructions, mir.Instruction{
		ID:   bVal,
		Op:   "const",
		Type: "int",
		Operands: []mir.Operand{
			{Kind: mir.OperandLiteral, Literal: "9", Type: "int"},
		},
	})
	cmpVal := fn.NextValue()
	entry.Instructions = append(entry.Instructions, mir.Instruction{
		ID:   cmpVal,
		Op:   "cmp.gt",
		Type: "bool",
		Operands: []mir.Operand{
			{Kind: mir.OperandValue, Value: aVal, Type: "int"},
			{Kind: mir.OperandValue, Value: bVal, Type: "int"},
		},
	})

	thenBlock := fn.NewBlock("then")
	elseBlock := fn.NewBlock("else")
	entry.Terminator = mir.Terminator{Op: "cbr", Operands: []mir.Operand{
		{Kind: mir.OperandValue, Value: cmpVal, Type: "bool"},
		{Kind: mir.OperandLiteral, Literal: thenBlock.Name},
		{Kind: mir.OperandLiteral, Literal: elseBlock.Name},
	}}

	thenBlock.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: aVal, Type: "int"}}}
	elseBlock.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: bVal, Type: "int"}}}

	mod := &mir.Module{Functions: []*mir.Function{fn}}

	res, err := vm.Execute(mod, "main")
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if res.Value != 9 {
		t.Fatalf("expected 9, got %v", res.Value)
	}
}
