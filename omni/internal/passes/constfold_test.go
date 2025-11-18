package passes_test

import (
	"strconv"
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

// Additional tests for constant folding

func TestConstFoldAllOperations(t *testing.T) {
	operations := []struct {
		op       string
		left     int
		right    int
		expected string
	}{
		{"add", 10, 5, "15"},
		{"sub", 10, 5, "5"},
		{"mul", 10, 5, "50"},
		{"div", 10, 5, "2"},
		{"mod", 10, 3, "1"},
		{"cmp.eq", 5, 5, "true"},
		{"cmp.eq", 5, 3, "false"},
		{"cmp.neq", 5, 5, "false"},
		{"cmp.neq", 5, 3, "true"},
		{"cmp.lt", 3, 5, "true"},
		{"cmp.lt", 5, 3, "false"},
		{"cmp.lte", 5, 5, "true"},
		{"cmp.lte", 3, 5, "true"},
		{"cmp.lte", 5, 3, "false"},
		{"cmp.gt", 5, 3, "true"},
		{"cmp.gt", 3, 5, "false"},
		{"cmp.gte", 5, 5, "true"},
		{"cmp.gte", 5, 3, "true"},
		{"cmp.gte", 3, 5, "false"},
		{"and", 1, 1, "true"},
		{"and", 1, 0, "false"},
		{"or", 0, 0, "false"},
		{"or", 1, 0, "true"},
	}

	for _, op := range operations {
		t.Run(op.op, func(t *testing.T) {
			fn := mir.NewFunction("main", "int", nil)
			block := fn.NewBlock("entry")

			v0 := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:       v0,
				Op:       "const",
				Type:     "int",
				Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: strconv.Itoa(op.left), Type: "int"}},
			})

			v1 := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:       v1,
				Op:       "const",
				Type:     "int",
				Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: strconv.Itoa(op.right), Type: "int"}},
			})

			result := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   result,
				Op:   op.op,
				Type: "int",
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: v0, Type: "int"},
					{Kind: mir.OperandValue, Value: v1, Type: "int"},
				},
			})

			block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: result}}}

			mod := &mir.Module{Functions: []*mir.Function{fn}}
			passes.ConstFold(mod)

			folded := block.Instructions[2]
			if folded.Op != "const" {
				t.Fatalf("expected folding to const, got %s", folded.Op)
			}
			if folded.Operands[0].Literal != op.expected {
				t.Fatalf("expected folded literal %s, got %s", op.expected, folded.Operands[0].Literal)
			}
		})
	}
}

func TestConstFoldDivisionByZero(t *testing.T) {
	fn := mir.NewFunction("main", "int", nil)
	block := fn.NewBlock("entry")

	v0 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:       v0,
		Op:       "const",
		Type:     "int",
		Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "10", Type: "int"}},
	})

	v1 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:       v1,
		Op:       "const",
		Type:     "int",
		Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "0", Type: "int"}},
	})

	result := fn.NextValue()
	div := mir.Instruction{
		ID:   result,
		Op:   "div",
		Type: "int",
		Operands: []mir.Operand{
			{Kind: mir.OperandValue, Value: v0, Type: "int"},
			{Kind: mir.OperandValue, Value: v1, Type: "int"},
		},
	}
	block.Instructions = append(block.Instructions, div)

	mod := &mir.Module{Functions: []*mir.Function{fn}}
	passes.ConstFold(mod)

	// Division by zero should not be folded
	folded := block.Instructions[2]
	if folded.Op != "div" {
		t.Fatalf("expected division to remain, got %s", folded.Op)
	}
}

func TestConstFoldModuloByZero(t *testing.T) {
	fn := mir.NewFunction("main", "int", nil)
	block := fn.NewBlock("entry")

	v0 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:       v0,
		Op:       "const",
		Type:     "int",
		Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "10", Type: "int"}},
	})

	v1 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:       v1,
		Op:       "const",
		Type:     "int",
		Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "0", Type: "int"}},
	})

	result := fn.NextValue()
	mod := mir.Instruction{
		ID:   result,
		Op:   "mod",
		Type: "int",
		Operands: []mir.Operand{
			{Kind: mir.OperandValue, Value: v0, Type: "int"},
			{Kind: mir.OperandValue, Value: v1, Type: "int"},
		},
	}
	block.Instructions = append(block.Instructions, mod)

	module := &mir.Module{Functions: []*mir.Function{fn}}
	passes.ConstFold(module)

	// Modulo by zero should not be folded
	folded := block.Instructions[2]
	if folded.Op != "mod" {
		t.Fatalf("expected modulo to remain, got %s", folded.Op)
	}
}

func TestConstFoldWithModifiedVariable(t *testing.T) {
	fn := mir.NewFunction("main", "int", nil)
	block := fn.NewBlock("entry")

	v0 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:       v0,
		Op:       "const",
		Type:     "int",
		Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "10", Type: "int"}},
	})

	// Assign to v0
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:   mir.InvalidValue,
		Op:   "assign",
		Type: "int",
		Operands: []mir.Operand{
			{Kind: mir.OperandValue, Value: v0, Type: "int"},
			{Kind: mir.OperandLiteral, Literal: "20", Type: "int"},
		},
	})

	v1 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:       v1,
		Op:       "const",
		Type:     "int",
		Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "5", Type: "int"}},
	})

	result := fn.NextValue()
	add := mir.Instruction{
		ID:   result,
		Op:   "add",
		Type: "int",
		Operands: []mir.Operand{
			{Kind: mir.OperandValue, Value: v0, Type: "int"},
			{Kind: mir.OperandValue, Value: v1, Type: "int"},
		},
	}
	block.Instructions = append(block.Instructions, add)

	module := &mir.Module{Functions: []*mir.Function{fn}}
	passes.ConstFold(module)

	// Should not fold because v0 is modified
	folded := block.Instructions[3]
	if folded.Op != "add" {
		t.Fatalf("expected add to remain (v0 is modified), got %s", folded.Op)
	}
}

func TestConstFoldWithLiteralOperands(t *testing.T) {
	fn := mir.NewFunction("main", "int", nil)
	block := fn.NewBlock("entry")

	result := fn.NextValue()
	add := mir.Instruction{
		ID:   result,
		Op:   "add",
		Type: "int",
		Operands: []mir.Operand{
			{Kind: mir.OperandLiteral, Literal: "10", Type: "int"},
			{Kind: mir.OperandLiteral, Literal: "5", Type: "int"},
		},
	}
	block.Instructions = append(block.Instructions, add)

	block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: result}}}

	module := &mir.Module{Functions: []*mir.Function{fn}}
	passes.ConstFold(module)

	folded := block.Instructions[0]
	if folded.Op != "const" {
		t.Fatalf("expected folding to const, got %s", folded.Op)
	}
	if folded.Operands[0].Literal != "15" {
		t.Fatalf("expected folded literal 15, got %s", folded.Operands[0].Literal)
	}
}

func TestConstFoldNonInteger(t *testing.T) {
	fn := mir.NewFunction("main", "string", nil)
	block := fn.NewBlock("entry")

	v0 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:       v0,
		Op:       "const",
		Type:     "string",
		Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "hello", Type: "string"}},
	})

	v1 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:       v1,
		Op:       "const",
		Type:     "string",
		Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "world", Type: "string"}},
	})

	result := fn.NextValue()
	add := mir.Instruction{
		ID:   result,
		Op:   "add",
		Type: "string",
		Operands: []mir.Operand{
			{Kind: mir.OperandValue, Value: v0, Type: "string"},
			{Kind: mir.OperandValue, Value: v1, Type: "string"},
		},
	}
	block.Instructions = append(block.Instructions, add)

	module := &mir.Module{Functions: []*mir.Function{fn}}
	passes.ConstFold(module)

	// Should not fold non-integer operations
	folded := block.Instructions[2]
	if folded.Op != "add" {
		t.Fatalf("expected add to remain (non-integer), got %s", folded.Op)
	}
}
