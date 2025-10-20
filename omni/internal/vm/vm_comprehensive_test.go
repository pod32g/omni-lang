package vm_test

import (
	"fmt"
	"testing"

	"github.com/omni-lang/omni/internal/mir"
	"github.com/omni-lang/omni/internal/vm"
)

// TestArithmeticOperations tests all arithmetic operations
func TestArithmeticOperations(t *testing.T) {
	tests := []struct {
		name     string
		op       string
		left     int
		right    int
		expected int
	}{
		{"add", "add", 10, 5, 15},
		{"sub", "sub", 10, 5, 5},
		{"mul", "mul", 10, 5, 50},
		{"div", "div", 10, 5, 2},
		{"mod", "mod", 10, 3, 1},
		{"add_zero", "add", 0, 0, 0},
		{"sub_zero", "sub", 0, 0, 0},
		{"mul_zero", "mul", 10, 0, 0},
		{"div_by_one", "div", 10, 1, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := mir.NewFunction("main", "int", nil)
			block := fn.NewBlock("entry")

			// Create left operand
			leftVal := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   leftVal,
				Op:   "const",
				Type: "int",
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: fmt.Sprintf("%d", tt.left), Type: "int"},
				},
			})

			// Create right operand
			rightVal := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   rightVal,
				Op:   "const",
				Type: "int",
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: fmt.Sprintf("%d", tt.right), Type: "int"},
				},
			})

			// Create operation
			resultVal := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   resultVal,
				Op:   tt.op,
				Type: "int",
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: leftVal, Type: "int"},
					{Kind: mir.OperandValue, Value: rightVal, Type: "int"},
				},
			})

			block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: resultVal, Type: "int"}}}

			mod := &mir.Module{Functions: []*mir.Function{fn}}
			res, err := vm.Execute(mod, "main")
			if err != nil {
				t.Fatalf("execute failed: %v", err)
			}
			if res.Value != tt.expected {
				t.Fatalf("expected %d, got %v", tt.expected, res.Value)
			}
		})
	}
}

// TestComparisonOperations tests all comparison operations
func TestComparisonOperations(t *testing.T) {
	tests := []struct {
		name     string
		op       string
		left     int
		right    int
		expected bool
	}{
		{"eq_true", "cmp.eq", 5, 5, true},
		{"eq_false", "cmp.eq", 5, 3, false},
		{"ne_true", "cmp.neq", 5, 3, true},
		{"ne_false", "cmp.neq", 5, 5, false},
		{"lt_true", "cmp.lt", 3, 5, true},
		{"lt_false", "cmp.lt", 5, 3, false},
		{"le_true", "cmp.lte", 3, 5, true},
		{"le_equal", "cmp.lte", 5, 5, true},
		{"le_false", "cmp.lte", 5, 3, false},
		{"gt_true", "cmp.gt", 5, 3, true},
		{"gt_false", "cmp.gt", 3, 5, false},
		{"ge_true", "cmp.gte", 5, 3, true},
		{"ge_equal", "cmp.gte", 5, 5, true},
		{"ge_false", "cmp.gte", 3, 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := mir.NewFunction("main", "bool", nil)
			block := fn.NewBlock("entry")

			// Create left operand
			leftVal := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   leftVal,
				Op:   "const",
				Type: "int",
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: fmt.Sprintf("%d", tt.left), Type: "int"},
				},
			})

			// Create right operand
			rightVal := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   rightVal,
				Op:   "const",
				Type: "int",
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: fmt.Sprintf("%d", tt.right), Type: "int"},
				},
			})

			// Create comparison
			resultVal := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   resultVal,
				Op:   tt.op,
				Type: "bool",
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: leftVal, Type: "int"},
					{Kind: mir.OperandValue, Value: rightVal, Type: "int"},
				},
			})

			block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: resultVal, Type: "bool"}}}

			mod := &mir.Module{Functions: []*mir.Function{fn}}
			res, err := vm.Execute(mod, "main")
			if err != nil {
				t.Fatalf("execute failed: %v", err)
			}
			if res.Value != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, res.Value)
			}
		})
	}
}

// TestLogicalOperations tests all logical operations
func TestLogicalOperations(t *testing.T) {
	tests := []struct {
		name     string
		op       string
		left     bool
		right    bool
		expected bool
	}{
		{"and_true_true", "and", true, true, true},
		{"and_true_false", "and", true, false, false},
		{"and_false_true", "and", false, true, false},
		{"and_false_false", "and", false, false, false},
		{"or_true_true", "or", true, true, true},
		{"or_true_false", "or", true, false, true},
		{"or_false_true", "or", false, true, true},
		{"or_false_false", "or", false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := mir.NewFunction("main", "bool", nil)
			block := fn.NewBlock("entry")

			// Create left operand
			leftVal := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   leftVal,
				Op:   "const",
				Type: "bool",
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: "true", Type: "bool"},
				},
			})
			if !tt.left {
				leftVal = fn.NextValue()
				block.Instructions = append(block.Instructions, mir.Instruction{
					ID:   leftVal,
					Op:   "const",
					Type: "bool",
					Operands: []mir.Operand{
						{Kind: mir.OperandLiteral, Literal: "false", Type: "bool"},
					},
				})
			}

			// Create right operand
			rightVal := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   rightVal,
				Op:   "const",
				Type: "bool",
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: "true", Type: "bool"},
				},
			})
			if !tt.right {
				rightVal = fn.NextValue()
				block.Instructions = append(block.Instructions, mir.Instruction{
					ID:   rightVal,
					Op:   "const",
					Type: "bool",
					Operands: []mir.Operand{
						{Kind: mir.OperandLiteral, Literal: "false", Type: "bool"},
					},
				})
			}

			// Create logical operation
			resultVal := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   resultVal,
				Op:   tt.op,
				Type: "bool",
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: leftVal, Type: "bool"},
					{Kind: mir.OperandValue, Value: rightVal, Type: "bool"},
				},
			})

			block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: resultVal, Type: "bool"}}}

			mod := &mir.Module{Functions: []*mir.Function{fn}}
			res, err := vm.Execute(mod, "main")
			if err != nil {
				t.Fatalf("execute failed: %v", err)
			}
			if res.Value != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, res.Value)
			}
		})
	}
}

// TestUnaryOperations tests unary operations
func TestUnaryOperations(t *testing.T) {
	tests := []struct {
		name     string
		op       string
		operand  int
		expected int
	}{
		{"neg_positive", "neg", 5, -5},
		{"neg_negative", "neg", -5, 5},
		{"neg_zero", "neg", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := mir.NewFunction("main", "int", nil)
			block := fn.NewBlock("entry")

			// Create operand
			operandVal := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   operandVal,
				Op:   "const",
				Type: "int",
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: fmt.Sprintf("%d", tt.operand), Type: "int"},
				},
			})

			// Create unary operation
			resultVal := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   resultVal,
				Op:   tt.op,
				Type: "int",
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: operandVal, Type: "int"},
				},
			})

			block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: resultVal, Type: "int"}}}

			mod := &mir.Module{Functions: []*mir.Function{fn}}
			res, err := vm.Execute(mod, "main")
			if err != nil {
				t.Fatalf("execute failed: %v", err)
			}
			if res.Value != tt.expected {
				t.Fatalf("expected %d, got %v", tt.expected, res.Value)
			}
		})
	}
}

// TestUnaryNot tests logical not operation
func TestUnaryNot(t *testing.T) {
	tests := []struct {
		name     string
		operand  bool
		expected bool
	}{
		{"not_true", true, false},
		{"not_false", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := mir.NewFunction("main", "bool", nil)
			block := fn.NewBlock("entry")

			// Create operand
			operandVal := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   operandVal,
				Op:   "const",
				Type: "bool",
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: "true", Type: "bool"},
				},
			})
			if !tt.operand {
				operandVal = fn.NextValue()
				block.Instructions = append(block.Instructions, mir.Instruction{
					ID:   operandVal,
					Op:   "const",
					Type: "bool",
					Operands: []mir.Operand{
						{Kind: mir.OperandLiteral, Literal: "false", Type: "bool"},
					},
				})
			}

			// Create not operation
			resultVal := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   resultVal,
				Op:   "not",
				Type: "bool",
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: operandVal, Type: "bool"},
				},
			})

			block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: resultVal, Type: "bool"}}}

			mod := &mir.Module{Functions: []*mir.Function{fn}}
			res, err := vm.Execute(mod, "main")
			if err != nil {
				t.Fatalf("execute failed: %v", err)
			}
			if res.Value != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, res.Value)
			}
		})
	}
}

// TestStringConcat tests string concatenation
func TestStringConcat(t *testing.T) {
	tests := []struct {
		name     string
		left     string
		right    string
		expected string
	}{
		{"concat_strings", "hello", "world", "helloworld"},
		{"concat_empty", "", "world", "world"},
		{"concat_both_empty", "", "", ""},
		{"concat_with_spaces", "hello ", "world", "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := mir.NewFunction("main", "string", nil)
			block := fn.NewBlock("entry")

			// Create left operand
			leftVal := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   leftVal,
				Op:   "const",
				Type: "string",
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: tt.left, Type: "string"},
				},
			})

			// Create right operand
			rightVal := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   rightVal,
				Op:   "const",
				Type: "string",
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: tt.right, Type: "string"},
				},
			})

			// Create string concat
			resultVal := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   resultVal,
				Op:   "strcat",
				Type: "string",
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: leftVal, Type: "string"},
					{Kind: mir.OperandValue, Value: rightVal, Type: "string"},
				},
			})

			block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: resultVal, Type: "string"}}}

			mod := &mir.Module{Functions: []*mir.Function{fn}}
			res, err := vm.Execute(mod, "main")
			if err != nil {
				t.Fatalf("execute failed: %v", err)
			}
			if res.Value != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, res.Value)
			}
		})
	}
}

// TestStringConcatMixedTypes tests string concatenation with mixed types
func TestStringConcatMixedTypes(t *testing.T) {
	tests := []struct {
		name      string
		left      interface{}
		right     interface{}
		leftType  string
		rightType string
		expected  string
	}{
		{"int_string", 42, "hello", "int", "string", "42hello"},
		{"string_int", "hello", 42, "string", "int", "hello42"},
		{"bool_string", true, "world", "bool", "string", "trueworld"},
		{"string_bool", "world", false, "string", "bool", "worldfalse"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := mir.NewFunction("main", "string", nil)
			block := fn.NewBlock("entry")

			// Create left operand
			leftVal := fn.NextValue()
			leftLiteral := ""
			switch v := tt.left.(type) {
			case int:
				leftLiteral = fmt.Sprintf("%d", v)
			case string:
				leftLiteral = v
			case bool:
				if v {
					leftLiteral = "true"
				} else {
					leftLiteral = "false"
				}
			}
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   leftVal,
				Op:   "const",
				Type: tt.leftType,
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: leftLiteral, Type: tt.leftType},
				},
			})

			// Create right operand
			rightVal := fn.NextValue()
			rightLiteral := ""
			switch v := tt.right.(type) {
			case int:
				rightLiteral = fmt.Sprintf("%d", v)
			case string:
				rightLiteral = v
			case bool:
				if v {
					rightLiteral = "true"
				} else {
					rightLiteral = "false"
				}
			}
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   rightVal,
				Op:   "const",
				Type: tt.rightType,
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: rightLiteral, Type: tt.rightType},
				},
			})

			// Create string concat
			resultVal := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   resultVal,
				Op:   "strcat",
				Type: "string",
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: leftVal, Type: tt.leftType},
					{Kind: mir.OperandValue, Value: rightVal, Type: tt.rightType},
				},
			})

			block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: resultVal, Type: "string"}}}

			mod := &mir.Module{Functions: []*mir.Function{fn}}
			res, err := vm.Execute(mod, "main")
			if err != nil {
				t.Fatalf("execute failed: %v", err)
			}
			if res.Value != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, res.Value)
			}
		})
	}
}

// TestMathIntrinsics tests all math intrinsic functions
func TestMathIntrinsics(t *testing.T) {
	tests := []struct {
		name     string
		callee   string
		args     []int
		expected int
	}{
		{"max_positive", "std.math.max", []int{10, 5}, 10},
		{"max_negative", "std.math.max", []int{-10, -5}, -5},
		{"max_equal", "std.math.max", []int{5, 5}, 5},
		{"min_positive", "std.math.min", []int{10, 5}, 5},
		{"min_negative", "std.math.min", []int{-10, -5}, -10},
		{"min_equal", "std.math.min", []int{5, 5}, 5},
		{"abs_positive", "std.math.abs", []int{5}, 5},
		{"abs_negative", "std.math.abs", []int{-5}, 5},
		{"abs_zero", "std.math.abs", []int{0}, 0},
		{"pow_small", "std.math.pow", []int{2, 3}, 8},
		{"pow_zero", "std.math.pow", []int{5, 0}, 1},
		{"pow_one", "std.math.pow", []int{5, 1}, 5},
		{"gcd_simple", "std.math.gcd", []int{12, 18}, 6},
		{"gcd_prime", "std.math.gcd", []int{17, 13}, 1},
		{"gcd_zero", "std.math.gcd", []int{0, 5}, 5},
		{"lcm_simple", "std.math.lcm", []int{12, 18}, 36},
		{"lcm_prime", "std.math.lcm", []int{17, 13}, 221},
		{"lcm_zero", "std.math.lcm", []int{0, 5}, 0},
		{"factorial_small", "std.math.factorial", []int{5}, 120},
		{"factorial_zero", "std.math.factorial", []int{0}, 1},
		{"factorial_one", "std.math.factorial", []int{1}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := mir.NewFunction("main", "int", nil)
			block := fn.NewBlock("entry")

			// Create argument values
			argVals := make([]mir.ValueID, len(tt.args))
			for i, arg := range tt.args {
				argVals[i] = fn.NextValue()
				block.Instructions = append(block.Instructions, mir.Instruction{
					ID:   argVals[i],
					Op:   "const",
					Type: "int",
					Operands: []mir.Operand{
						{Kind: mir.OperandLiteral, Literal: fmt.Sprintf("%d", arg), Type: "int"},
					},
				})
			}

			// Create call instruction
			callVal := fn.NextValue()
			operands := []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: tt.callee},
			}
			for _, argVal := range argVals {
				operands = append(operands, mir.Operand{Kind: mir.OperandValue, Value: argVal, Type: "int"})
			}
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:       callVal,
				Op:       "call",
				Type:     "int",
				Operands: operands,
			})

			block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: callVal, Type: "int"}}}

			mod := &mir.Module{Functions: []*mir.Function{fn}}
			res, err := vm.Execute(mod, "main")
			if err != nil {
				t.Fatalf("execute failed: %v", err)
			}
			if res.Value != tt.expected {
				t.Fatalf("expected %d, got %v", tt.expected, res.Value)
			}
		})
	}
}

// TestStringIntrinsics tests all string intrinsic functions
func TestStringIntrinsics(t *testing.T) {
	tests := []struct {
		name         string
		callee       string
		args         []string
		expected     interface{}
		expectedType string
	}{
		{"length_short", "std.string.length", []string{"hello"}, 5, "int"},
		{"length_empty", "std.string.length", []string{""}, 0, "int"},
		{"length_long", "std.string.length", []string{"hello world"}, 11, "int"},
		{"concat_basic", "std.string.concat", []string{"hello", "world"}, "helloworld", "string"},
		{"concat_empty", "std.string.concat", []string{"", "world"}, "world", "string"},
		{"concat_both_empty", "std.string.concat", []string{"", ""}, "", "string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := mir.NewFunction("main", tt.expectedType, nil)
			block := fn.NewBlock("entry")

			// Create argument values
			argVals := make([]mir.ValueID, len(tt.args))
			for i, arg := range tt.args {
				argVals[i] = fn.NextValue()
				block.Instructions = append(block.Instructions, mir.Instruction{
					ID:   argVals[i],
					Op:   "const",
					Type: "string",
					Operands: []mir.Operand{
						{Kind: mir.OperandLiteral, Literal: arg, Type: "string"},
					},
				})
			}

			// Create call instruction
			callVal := fn.NextValue()
			operands := []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: tt.callee},
			}
			for _, argVal := range argVals {
				operands = append(operands, mir.Operand{Kind: mir.OperandValue, Value: argVal, Type: "string"})
			}
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:       callVal,
				Op:       "call",
				Type:     tt.expectedType,
				Operands: operands,
			})

			block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: callVal, Type: tt.expectedType}}}

			mod := &mir.Module{Functions: []*mir.Function{fn}}
			res, err := vm.Execute(mod, "main")
			if err != nil {
				t.Fatalf("execute failed: %v", err)
			}
			if res.Value != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, res.Value)
			}
		})
	}
}

// TestErrorConditions tests various error conditions
func TestErrorConditions(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *mir.Module
		expectError bool
		errorMsg    string
	}{
		{
			"invalid_instruction",
			func() *mir.Module {
				fn := mir.NewFunction("main", "int", nil)
				block := fn.NewBlock("entry")
				block.Instructions = append(block.Instructions, mir.Instruction{
					ID:       fn.NextValue(),
					Op:       "invalid_op",
					Type:     "int",
					Operands: []mir.Operand{},
				})
				block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "0", Type: "int"}}}
				return &mir.Module{Functions: []*mir.Function{fn}}
			},
			true,
			"unsupported instruction",
		},
		{
			"missing_function",
			func() *mir.Module {
				return &mir.Module{Functions: []*mir.Function{}}
			},
			true,
			"entry function \"main\" not found",
		},
		{
			"invalid_operand_type",
			func() *mir.Module {
				fn := mir.NewFunction("main", "int", nil)
				block := fn.NewBlock("entry")
				block.Instructions = append(block.Instructions, mir.Instruction{
					ID:   fn.NextValue(),
					Op:   "add",
					Type: "int",
					Operands: []mir.Operand{
						{Kind: mir.OperandValue, Value: 999, Type: "int"}, // Invalid value ID
						{Kind: mir.OperandValue, Value: 998, Type: "int"}, // Invalid value ID
					},
				})
				block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "0", Type: "int"}}}
				return &mir.Module{Functions: []*mir.Function{fn}}
			},
			true,
			"expected int, got nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod := tt.setup()
			_, err := vm.Execute(mod, "main")
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Fatalf("expected error to contain %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			contains(s[1:], substr))))
}

// TestComplexExpressions tests complex nested expressions
func TestComplexExpressions(t *testing.T) {
	fn := mir.NewFunction("main", "int", nil)
	block := fn.NewBlock("entry")

	// Create: (5 + 3) * (10 - 2) / 4
	// Step 1: 5 + 3 = 8
	v1 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:       v1,
		Op:       "const",
		Type:     "int",
		Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "5", Type: "int"}},
	})
	v2 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:       v2,
		Op:       "const",
		Type:     "int",
		Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "3", Type: "int"}},
	})
	v3 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:   v3,
		Op:   "add",
		Type: "int",
		Operands: []mir.Operand{
			{Kind: mir.OperandValue, Value: v1, Type: "int"},
			{Kind: mir.OperandValue, Value: v2, Type: "int"},
		},
	})

	// Step 2: 10 - 2 = 8
	v4 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:       v4,
		Op:       "const",
		Type:     "int",
		Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "10", Type: "int"}},
	})
	v5 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:       v5,
		Op:       "const",
		Type:     "int",
		Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "2", Type: "int"}},
	})
	v6 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:   v6,
		Op:   "sub",
		Type: "int",
		Operands: []mir.Operand{
			{Kind: mir.OperandValue, Value: v4, Type: "int"},
			{Kind: mir.OperandValue, Value: v5, Type: "int"},
		},
	})

	// Step 3: 8 * 8 = 64
	v7 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:   v7,
		Op:   "mul",
		Type: "int",
		Operands: []mir.Operand{
			{Kind: mir.OperandValue, Value: v3, Type: "int"},
			{Kind: mir.OperandValue, Value: v6, Type: "int"},
		},
	})

	// Step 4: 64 / 4 = 16
	v8 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:       v8,
		Op:       "const",
		Type:     "int",
		Operands: []mir.Operand{{Kind: mir.OperandLiteral, Literal: "4", Type: "int"}},
	})
	v9 := fn.NextValue()
	block.Instructions = append(block.Instructions, mir.Instruction{
		ID:   v9,
		Op:   "div",
		Type: "int",
		Operands: []mir.Operand{
			{Kind: mir.OperandValue, Value: v7, Type: "int"},
			{Kind: mir.OperandValue, Value: v8, Type: "int"},
		},
	})

	block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: v9, Type: "int"}}}

	mod := &mir.Module{Functions: []*mir.Function{fn}}
	res, err := vm.Execute(mod, "main")
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if res.Value != 16 {
		t.Fatalf("expected 16, got %v", res.Value)
	}
}
