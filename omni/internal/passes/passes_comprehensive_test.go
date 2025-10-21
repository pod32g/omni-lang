package passes_test

import (
	"testing"

	"github.com/omni-lang/omni/internal/mir"
	"github.com/omni-lang/omni/internal/passes"
)

// TestConstFoldArithmetic tests constant folding for arithmetic operations
func TestConstFoldArithmetic(t *testing.T) {
	tests := []struct {
		name     string
		op       string
		left     string
		right    string
		expected string
	}{
		{"add", "add", "5", "3", "8"},
		{"sub", "sub", "10", "4", "6"},
		{"mul", "mul", "6", "7", "42"},
		{"div", "div", "20", "4", "5"},
		{"mod", "mod", "17", "5", "2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := mir.NewFunction("main", "int", nil)
			block := fn.NewBlock("entry")

			// Create left constant
			v0 := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   v0,
				Op:   "const",
				Type: "int",
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: tt.left, Type: "int"},
				},
			})

			// Create right constant
			v1 := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   v1,
				Op:   "const",
				Type: "int",
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: tt.right, Type: "int"},
				},
			})

			// Create arithmetic operation
			v2 := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   v2,
				Op:   tt.op,
				Type:  "int",
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: v0, Type: "int"},
					{Kind: mir.OperandValue, Value: v1, Type: "int"},
				},
			})

			block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: v2, Type: "int"}}}

			// Apply constant folding
			changed := passes.ConstFold(fn)
			if !changed {
				t.Error("Expected constant folding to make changes")
			}

			// Check that the arithmetic operation was replaced with a constant
			if len(block.Instructions) != 3 {
				t.Errorf("Expected 3 instructions, got %d", len(block.Instructions))
			}

			// The last instruction should be a constant with the folded value
			lastInst := block.Instructions[len(block.Instructions)-1]
			if lastInst.Op != "const" {
				t.Errorf("Expected last instruction to be const, got %s", lastInst.Op)
			}

			if len(lastInst.Operands) != 1 {
				t.Errorf("Expected 1 operand, got %d", len(lastInst.Operands))
			}

			if lastInst.Operands[0].Literal != tt.expected {
				t.Errorf("Expected folded value %s, got %s", tt.expected, lastInst.Operands[0].Literal)
			}
		})
	}
}

// TestConstFoldComparison tests constant folding for comparison operations
func TestConstFoldComparison(t *testing.T) {
	tests := []struct {
		name     string
		op       string
		left     string
		right    string
		expected string
	}{
		{"eq_true", "cmp.eq", "5", "5", "true"},
		{"eq_false", "cmp.eq", "5", "3", "false"},
		{"neq_true", "cmp.neq", "5", "3", "true"},
		{"neq_false", "cmp.neq", "5", "5", "false"},
		{"lt_true", "cmp.lt", "3", "5", "true"},
		{"lt_false", "cmp.lt", "5", "3", "false"},
		{"lte_true", "cmp.lte", "3", "5", "true"},
		{"lte_equal", "cmp.lte", "5", "5", "true"},
		{"lte_false", "cmp.lte", "5", "3", "false"},
		{"gt_true", "cmp.gt", "5", "3", "true"},
		{"gt_false", "cmp.gt", "3", "5", "false"},
		{"gte_true", "cmp.gte", "5", "3", "true"},
		{"gte_equal", "cmp.gte", "5", "5", "true"},
		{"gte_false", "cmp.gte", "3", "5", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := mir.NewFunction("main", "bool", nil)
			block := fn.NewBlock("entry")

			// Create left constant
			v0 := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   v0,
				Op:   "const",
				Type: "int",
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: tt.left, Type: "int"},
				},
			})

			// Create right constant
			v1 := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   v1,
				Op:   "const",
				Type: "int",
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: tt.right, Type: "int"},
				},
			})

			// Create comparison operation
			v2 := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   v2,
				Op:   tt.op,
				Type:  "bool",
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: v0, Type: "int"},
					{Kind: mir.OperandValue, Value: v1, Type: "int"},
				},
			})

			block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: v2, Type: "bool"}}}

			// Apply constant folding
			changed := passes.ConstFold(fn)
			if !changed {
				t.Error("Expected constant folding to make changes")
			}

			// Check that the comparison operation was replaced with a constant
			if len(block.Instructions) != 3 {
				t.Errorf("Expected 3 instructions, got %d", len(block.Instructions))
			}

			// The last instruction should be a constant with the folded value
			lastInst := block.Instructions[len(block.Instructions)-1]
			if lastInst.Op != "const" {
				t.Errorf("Expected last instruction to be const, got %s", lastInst.Op)
			}

			if len(lastInst.Operands) != 1 {
				t.Errorf("Expected 1 operand, got %d", len(lastInst.Operands))
			}

			if lastInst.Operands[0].Literal != tt.expected {
				t.Errorf("Expected folded value %s, got %s", tt.expected, lastInst.Operands[0].Literal)
			}
		})
	}
}

// TestConstFoldLogical tests constant folding for logical operations
func TestConstFoldLogical(t *testing.T) {
	tests := []struct {
		name     string
		op       string
		left     string
		right    string
		expected string
	}{
		{"and_true_true", "and", "true", "true", "true"},
		{"and_true_false", "and", "true", "false", "false"},
		{"and_false_true", "and", "false", "true", "false"},
		{"and_false_false", "and", "false", "false", "false"},
		{"or_true_true", "or", "true", "true", "true"},
		{"or_true_false", "or", "true", "false", "true"},
		{"or_false_true", "or", "false", "true", "true"},
		{"or_false_false", "or", "false", "false", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := mir.NewFunction("main", "bool", nil)
			block := fn.NewBlock("entry")

			// Create left constant
			v0 := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   v0,
				Op:   "const",
				Type: "bool",
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: tt.left, Type: "bool"},
				},
			})

			// Create right constant
			v1 := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   v1,
				Op:   "const",
				Type: "bool",
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: tt.right, Type: "bool"},
				},
			})

			// Create logical operation
			v2 := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   v2,
				Op:   tt.op,
				Type:  "bool",
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: v0, Type: "bool"},
					{Kind: mir.OperandValue, Value: v1, Type: "bool"},
				},
			})

			block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: v2, Type: "bool"}}}

			// Apply constant folding
			changed := passes.ConstFold(fn)
			if !changed {
				t.Error("Expected constant folding to make changes")
			}

			// Check that the logical operation was replaced with a constant
			if len(block.Instructions) != 3 {
				t.Errorf("Expected 3 instructions, got %d", len(block.Instructions))
			}

			// The last instruction should be a constant with the folded value
			lastInst := block.Instructions[len(block.Instructions)-1]
			if lastInst.Op != "const" {
				t.Errorf("Expected last instruction to be const, got %s", lastInst.Op)
			}

			if len(lastInst.Operands) != 1 {
				t.Errorf("Expected 1 operand, got %d", len(lastInst.Operands))
			}

			if lastInst.Operands[0].Literal != tt.expected {
				t.Errorf("Expected folded value %s, got %s", tt.expected, lastInst.Operands[0].Literal)
			}
		})
	}
}

// TestConstFoldUnary tests constant folding for unary operations
func TestConstFoldUnary(t *testing.T) {
	tests := []struct {
		name     string
		op       string
		operand  string
		expected string
	}{
		{"neg_positive", "neg", "5", "-5"},
		{"neg_negative", "neg", "-3", "3"},
		{"neg_zero", "neg", "0", "0"},
		{"not_true", "not", "true", "false"},
		{"not_false", "not", "false", "true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := mir.NewFunction("main", "int", nil)
			block := fn.NewBlock("entry")

			// Create operand constant
			v0 := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   v0,
				Op:   "const",
				Type: "int",
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: tt.operand, Type: "int"},
				},
			})

			// Create unary operation
			v1 := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   v1,
				Op:   tt.op,
				Type:  "int",
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: v0, Type: "int"},
				},
			})

			block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: v1, Type: "int"}}}

			// Apply constant folding
			changed := passes.ConstFold(fn)
			if !changed {
				t.Error("Expected constant folding to make changes")
			}

			// Check that the unary operation was replaced with a constant
			if len(block.Instructions) != 2 {
				t.Errorf("Expected 2 instructions, got %d", len(block.Instructions))
			}

			// The last instruction should be a constant with the folded value
			lastInst := block.Instructions[len(block.Instructions)-1]
			if lastInst.Op != "const" {
				t.Errorf("Expected last instruction to be const, got %s", lastInst.Op)
			}

			if len(lastInst.Operands) != 1 {
				t.Errorf("Expected 1 operand, got %d", len(lastInst.Operands))
			}

			if lastInst.Operands[0].Literal != tt.expected {
				t.Errorf("Expected folded value %s, got %s", tt.expected, lastInst.Operands[0].Literal)
			}
		})
	}
}

// TestConstFoldNoChange tests that constant folding doesn't change when no constants are involved
func TestConstFoldNoChange(t *testing.T) {
	fn := mir.NewFunction("main", "int", nil)
	block := fn.NewBlock("entry")

	// Create a parameter
	param := mir.Param{ID: fn.NextValue(), Name: "x", Type: "int"}
	fn.Params = append(fn.Params, param)

	// Create a variable reference (not a constant)
	v0 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:   v0,
		Op:   "add",
		Type: "int",
		Operands: []mir.Operand{
			{Kind: mir.OperandValue, Value: param.ID, Type: "int"},
			{Kind: mir.OperandValue, Value: param.ID, Type: "int"},
		},
	})

	block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: v0, Type: "int"}}}

	// Apply constant folding
	changed := passes.ConstFold(fn)
	if changed {
		t.Error("Expected constant folding to make no changes when no constants are involved")
	}

	// Check that the instruction is still there
	if len(block.Instructions) != 1 {
		t.Errorf("Expected 1 instruction, got %d", len(block.Instructions))
	}

	if block.Instructions[0].Op != "add" {
		t.Errorf("Expected add instruction, got %s", block.Instructions[0].Op)
	}
}

// TestVerifyDetectsMissingTerminator tests MIR verification
func TestVerifyDetectsMissingTerminator(t *testing.T) {
	fn := mir.NewFunction("main", "int", nil)
	block := fn.NewBlock("entry")

	// Add an instruction but no terminator
	v0 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:   v0,
		Op:   "const",
		Type: "int",
		Operands: []mir.Operand{
			{Kind: mir.OperandLiteral, Literal: "42", Type: "int"},
		},
	})

	// Don't set a terminator - this should cause verification to fail
	err := passes.Verify(fn)
	if err == nil {
		t.Error("Expected verification to fail for missing terminator")
	}
}

// TestVerifyDetectsInvalidInstruction tests MIR verification for invalid instructions
func TestVerifyDetectsInvalidInstruction(t *testing.T) {
	fn := mir.NewFunction("main", "int", nil)
	block := fn.NewBlock("entry")

	// Add an invalid instruction
	v0 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:   v0,
		Op:   "invalid_op",
		Type: "int",
		Operands: []mir.Operand{
			{Kind: mir.OperandLiteral, Literal: "42", Type: "int"},
		},
	})

	block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: v0, Type: "int"}}}

	err := passes.Verify(fn)
	if err == nil {
		t.Error("Expected verification to fail for invalid instruction")
	}
}

// TestVerifyValidMIR tests that valid MIR passes verification
func TestVerifyValidMIR(t *testing.T) {
	fn := mir.NewFunction("main", "int", nil)
	block := fn.NewBlock("entry")

	// Add a valid instruction
	v0 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:   v0,
		Op:   "const",
		Type: "int",
		Operands: []mir.Operand{
			{Kind: mir.OperandLiteral, Literal: "42", Type: "int"},
		},
	})

	block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: v0, Type: "int"}}}

	err := passes.Verify(fn)
	if err != nil {
		t.Errorf("Expected valid MIR to pass verification, got error: %v", err)
	}
}

// TestOptimizePipeline tests the optimization pipeline
func TestOptimizePipeline(t *testing.T) {
	fn := mir.NewFunction("main", "int", nil)
	block := fn.NewBlock("entry")

	// Create a function that can be optimized
	// (5 + 3) * 2
	v0 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:   v0,
		Op:   "const",
		Type: "int",
		Operands: []mir.Operand{
			{Kind: mir.OperandLiteral, Literal: "5", Type: "int"},
		},
	})

	v1 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:   v1,
		Op:   "const",
		Type: "int",
		Operands: []mir.Operand{
			{Kind: mir.OperandLiteral, Literal: "3", Type: "int"},
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

	v3 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:   v3,
		Op:   "const",
		Type: "int",
		Operands: []mir.Operand{
			{Kind: mir.OperandLiteral, Literal: "2", Type: "int"},
		},
	})

	v4 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:   v4,
		Op:   "mul",
		Type: "int",
		Operands: []mir.Operand{
			{Kind: mir.OperandValue, Value: v2, Type: "int"},
			{Kind: mir.OperandValue, Value: v3, Type: "int"},
		},
	})

	block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: v4, Type: "int"}}}

	// Apply optimization pipeline
	passes.Optimize(fn)

	// After optimization, we should have fewer instructions
	// The (5 + 3) should be folded to 8, and (8 * 2) should be folded to 16
	if len(block.Instructions) > 5 {
		t.Errorf("Expected optimization to reduce instruction count, got %d instructions", len(block.Instructions))
	}
}
