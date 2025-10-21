package vm_test

import (
	"testing"

	"github.com/omni-lang/omni/internal/mir"
	"github.com/omni-lang/omni/internal/vm"
)

// TestArithmeticOperations tests all arithmetic operations
func TestArithmeticOperations(t *testing.T) {
	tests := []struct {
		name     string
		op       string
		left     string
		right    string
		expected int
	}{
		{"add", "add", "10", "5", 15},
		{"sub", "sub", "10", "3", 7},
		{"mul", "mul", "4", "5", 20},
		{"div", "div", "20", "4", 5},
		{"mod", "mod", "17", "5", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := mir.NewFunction("main", "int", nil)
			block := fn.NewBlock("entry")

			// Create left operand
			v0 := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   v0,
				Op:   "const",
				Type: "int",
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: tt.left, Type: "int"},
				},
			})

			// Create right operand
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
				t.Fatalf("Execution failed: %v", err)
			}

			if res.Value != tt.expected {
				t.Errorf("Expected %d, got %v", tt.expected, res.Value)
			}
		})
	}
}

// TestComparisonOperations tests all comparison operations
func TestComparisonOperations(t *testing.T) {
	tests := []struct {
		name     string
		op       string
		left     string
		right    string
		expected bool
	}{
		{"eq_true", "cmp.eq", "5", "5", true},
		{"eq_false", "cmp.eq", "5", "3", false},
		{"neq_true", "cmp.neq", "5", "3", true},
		{"neq_false", "cmp.neq", "5", "5", false},
		{"lt_true", "cmp.lt", "3", "5", true},
		{"lt_false", "cmp.lt", "5", "3", false},
		{"lte_true", "cmp.lte", "3", "5", true},
		{"lte_equal", "cmp.lte", "5", "5", true},
		{"lte_false", "cmp.lte", "5", "3", false},
		{"gt_true", "cmp.gt", "5", "3", true},
		{"gt_false", "cmp.gt", "3", "5", false},
		{"gte_true", "cmp.gte", "5", "3", true},
		{"gte_equal", "cmp.gte", "5", "5", true},
		{"gte_false", "cmp.gte", "3", "5", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := mir.NewFunction("main", "bool", nil)
			block := fn.NewBlock("entry")

			// Create left operand
			v0 := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   v0,
				Op:   "const",
				Type: "int",
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: tt.left, Type: "int"},
				},
			})

			// Create right operand
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
				Type: "bool",
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: v0, Type: "int"},
					{Kind: mir.OperandValue, Value: v1, Type: "int"},
				},
			})

			block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: v2, Type: "bool"}}}

			mod := &mir.Module{Functions: []*mir.Function{fn}}

			res, err := vm.Execute(mod, "main")
			if err != nil {
				t.Fatalf("Execution failed: %v", err)
			}

			if res.Value != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, res.Value)
			}
		})
	}
}

// TestLogicalOperations tests logical operations
func TestLogicalOperations(t *testing.T) {
	tests := []struct {
		name     string
		op       string
		left     string
		right    string
		expected bool
	}{
		{"and_true_true", "and", "true", "true", true},
		{"and_true_false", "and", "true", "false", false},
		{"and_false_true", "and", "false", "true", false},
		{"and_false_false", "and", "false", "false", false},
		{"or_true_true", "or", "true", "true", true},
		{"or_true_false", "or", "true", "false", true},
		{"or_false_true", "or", "false", "true", true},
		{"or_false_false", "or", "false", "false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := mir.NewFunction("main", "bool", nil)
			block := fn.NewBlock("entry")

			// Create left operand
			v0 := fn.NextValue()
			block.Instructions = append(block.Instructions, mir.Instruction{
				ID:   v0,
				Op:   "const",
				Type: "bool",
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: tt.left, Type: "bool"},
				},
			})

			// Create right operand
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
				Type: "bool",
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: v0, Type: "bool"},
					{Kind: mir.OperandValue, Value: v1, Type: "bool"},
				},
			})

			block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: v2, Type: "bool"}}}

			mod := &mir.Module{Functions: []*mir.Function{fn}}

			res, err := vm.Execute(mod, "main")
			if err != nil {
				t.Fatalf("Execution failed: %v", err)
			}

			if res.Value != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, res.Value)
			}
		})
	}
}

// TestUnaryOperations tests unary operations
func TestUnaryOperations(t *testing.T) {
	tests := []struct {
		name     string
		op       string
		operand  string
		expected interface{}
	}{
		{"neg_positive", "neg", "5", -5},
		{"neg_negative", "neg", "-3", 3},
		{"neg_zero", "neg", "0", 0},
		{"not_true", "not", "true", false},
		{"not_false", "not", "false", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := mir.NewFunction("main", "int", nil)
			block := fn.NewBlock("entry")

			// Create operand
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
				Type: "int",
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: v0, Type: "int"},
				},
			})

			block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: v1, Type: "int"}}}

			mod := &mir.Module{Functions: []*mir.Function{fn}}

			res, err := vm.Execute(mod, "main")
			if err != nil {
				t.Fatalf("Execution failed: %v", err)
			}

			if res.Value != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, res.Value)
			}
		})
	}
}

// TestStringOperations tests string operations
func TestStringOperations(t *testing.T) {
	t.Run("string_concat", func(t *testing.T) {
		fn := mir.NewFunction("main", "string", nil)
		block := fn.NewBlock("entry")

		// Create first string
		v0 := fn.NextValue()
		block.Instructions = append(block.Instructions, mir.Instruction{
			ID:   v0,
			Op:   "const",
			Type: "string",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "Hello", Type: "string"},
			},
		})

		// Create second string
		v1 := fn.NextValue()
		block.Instructions = append(block.Instructions, mir.Instruction{
			ID:   v1,
			Op:   "const",
			Type: "string",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "World", Type: "string"},
			},
		})

		// Create string concatenation
		v2 := fn.NextValue()
		block.Instructions = append(block.Instructions, mir.Instruction{
			ID:   v2,
			Op:   "strcat",
			Type: "string",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: v0, Type: "string"},
				{Kind: mir.OperandValue, Value: v1, Type: "string"},
			},
		})

		block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: v2, Type: "string"}}}

		mod := &mir.Module{Functions: []*mir.Function{fn}}

		res, err := vm.Execute(mod, "main")
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		expected := "HelloWorld"
		if res.Value != expected {
			t.Errorf("Expected %q, got %q", expected, res.Value)
		}
	})
}

// TestArrayOperations tests array operations
func TestArrayOperations(t *testing.T) {
	t.Run("array_init", func(t *testing.T) {
		fn := mir.NewFunction("main", "[]int", nil)
		block := fn.NewBlock("entry")

		// Create array initialization
		v0 := fn.NextValue()
		block.Instructions = append(block.Instructions, mir.Instruction{
			ID:   v0,
			Op:   "array.init",
			Type: "[]int",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "1", Type: "int"},
				{Kind: mir.OperandLiteral, Literal: "2", Type: "int"},
				{Kind: mir.OperandLiteral, Literal: "3", Type: "int"},
			},
		})

		block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: v0, Type: "[]int"}}}

		mod := &mir.Module{Functions: []*mir.Function{fn}}

		res, err := vm.Execute(mod, "main")
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		// Check that we got an array
		if res.Type != "[]int" {
			t.Errorf("Expected type []int, got %s", res.Type)
		}

		arr, ok := res.Value.([]interface{})
		if !ok {
			t.Fatalf("Expected array, got %T", res.Value)
		}

		if len(arr) != 3 {
			t.Errorf("Expected array length 3, got %d", len(arr))
		}

		expected := []int{1, 2, 3}
		for i, val := range arr {
			if val != expected[i] {
				t.Errorf("Expected arr[%d] = %d, got %v", i, expected[i], val)
			}
		}
	})

	t.Run("array_index", func(t *testing.T) {
		fn := mir.NewFunction("main", "int", nil)
		block := fn.NewBlock("entry")

		// Create array
		v0 := fn.NextValue()
		block.Instructions = append(block.Instructions, mir.Instruction{
			ID:   v0,
			Op:   "array.init",
			Type: "[]int",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "10", Type: "int"},
				{Kind: mir.OperandLiteral, Literal: "20", Type: "int"},
				{Kind: mir.OperandLiteral, Literal: "30", Type: "int"},
			},
		})

		// Create index
		v1 := fn.NextValue()
		block.Instructions = append(block.Instructions, mir.Instruction{
			ID:   v1,
			Op:   "const",
			Type: "int",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "1", Type: "int"},
			},
		})

		// Create array indexing
		v2 := fn.NextValue()
		block.Instructions = append(block.Instructions, mir.Instruction{
			ID:   v2,
			Op:   "index",
			Type: "int",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: v0, Type: "[]int"},
				{Kind: mir.OperandValue, Value: v1, Type: "int"},
			},
		})

		block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: v2, Type: "int"}}}

		mod := &mir.Module{Functions: []*mir.Function{fn}}

		res, err := vm.Execute(mod, "main")
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		expected := 20
		if res.Value != expected {
			t.Errorf("Expected %d, got %v", expected, res.Value)
		}
	})
}

// TestMapOperations tests map operations
func TestMapOperations(t *testing.T) {
	t.Run("map_init", func(t *testing.T) {
		fn := mir.NewFunction("main", "map<string,int>", nil)
		block := fn.NewBlock("entry")

		// Create map initialization
		v0 := fn.NextValue()
		block.Instructions = append(block.Instructions, mir.Instruction{
			ID:   v0,
			Op:   "map.init",
			Type: "map<string,int>",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "key1", Type: "string"},
				{Kind: mir.OperandLiteral, Literal: "100", Type: "int"},
				{Kind: mir.OperandLiteral, Literal: "key2", Type: "string"},
				{Kind: mir.OperandLiteral, Literal: "200", Type: "int"},
			},
		})

		block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: v0, Type: "map<string,int>"}}}

		mod := &mir.Module{Functions: []*mir.Function{fn}}

		res, err := vm.Execute(mod, "main")
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		// Check that we got a map
		if res.Type != "map<string,int>" {
			t.Errorf("Expected type map<string,int>, got %s", res.Type)
		}

		m, ok := res.Value.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map, got %T", res.Value)
		}

		if len(m) != 2 {
			t.Errorf("Expected map size 2, got %d", len(m))
		}

		if m["key1"] != 100 {
			t.Errorf("Expected m[\"key1\"] = 100, got %v", m["key1"])
		}

		if m["key2"] != 200 {
			t.Errorf("Expected m[\"key2\"] = 200, got %v", m["key2"])
		}
	})
}

// TestStructOperations tests struct operations
func TestStructOperations(t *testing.T) {
	t.Run("struct_init", func(t *testing.T) {
		fn := mir.NewFunction("main", "struct{name:string,age:int}", nil)
		block := fn.NewBlock("entry")

		// Create struct initialization
		v0 := fn.NextValue()
		block.Instructions = append(block.Instructions, mir.Instruction{
			ID:   v0,
			Op:   "struct.init",
			Type: "struct{name:string,age:int}",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "name", Type: "string"},
				{Kind: mir.OperandLiteral, Literal: "Alice", Type: "string"},
				{Kind: mir.OperandLiteral, Literal: "age", Type: "string"},
				{Kind: mir.OperandLiteral, Literal: "30", Type: "int"},
			},
		})

		block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: v0, Type: "struct{name:string,age:int}"}}}

		mod := &mir.Module{Functions: []*mir.Function{fn}}

		res, err := vm.Execute(mod, "main")
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		// Check that we got a struct
		if res.Type != "struct{name:string,age:int}" {
			t.Errorf("Expected type struct{name:string,age:int}, got %s", res.Type)
		}

		s, ok := res.Value.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected struct, got %T", res.Value)
		}

		if s["name"] != "Alice" {
			t.Errorf("Expected s.name = \"Alice\", got %v", s["name"])
		}

		if s["age"] != 30 {
			t.Errorf("Expected s.age = 30, got %v", s["age"])
		}
	})
}

// TestErrorConditions tests error handling
func TestErrorConditions(t *testing.T) {
	t.Run("invalid_instruction", func(t *testing.T) {
		fn := mir.NewFunction("main", "int", nil)
		block := fn.NewBlock("entry")

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

		mod := &mir.Module{Functions: []*mir.Function{fn}}

		_, err := vm.Execute(mod, "main")
		if err == nil {
			t.Fatal("Expected error for invalid instruction")
		}
	})

	t.Run("missing_function", func(t *testing.T) {
		mod := &mir.Module{Functions: []*mir.Function{}}

		_, err := vm.Execute(mod, "nonexistent")
		if err == nil {
			t.Fatal("Expected error for missing function")
		}
	})

	t.Run("invalid_operand_type", func(t *testing.T) {
		fn := mir.NewFunction("main", "int", nil)
		block := fn.NewBlock("entry")

		// Create string operand
		v0 := fn.NextValue()
		block.Instructions = append(block.Instructions, mir.Instruction{
			ID:   v0,
			Op:   "const",
			Type: "string",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "hello", Type: "string"},
			},
		})

		// Try to add string to int (should fail)
		v1 := fn.NextValue()
		block.Instructions = append(block.Instructions, mir.Instruction{
			ID:   v1,
			Op:   "add",
			Type: "int",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: v0, Type: "string"},
				{Kind: mir.OperandValue, Value: v0, Type: "string"},
			},
		})

		block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: v1, Type: "int"}}}

		mod := &mir.Module{Functions: []*mir.Function{fn}}

		_, err := vm.Execute(mod, "main")
		if err == nil {
			t.Fatal("Expected error for invalid operand type")
		}
	})
}

// TestComplexExpressions tests complex nested expressions
func TestComplexExpressions(t *testing.T) {
	t.Run("nested_arithmetic", func(t *testing.T) {
		fn := mir.NewFunction("main", "int", nil)
		block := fn.NewBlock("entry")

		// Create constants
		v0 := fn.NextValue()
		block.Instructions = append(block.Instructions, mir.Instruction{
			ID:   v0,
			Op:   "const",
			Type: "int",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "10", Type: "int"},
			},
		})

		v1 := fn.NextValue()
		block.Instructions = append(block.Instructions, mir.Instruction{
			ID:   v1,
			Op:   "const",
			Type: "int",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "5", Type: "int"},
			},
		})

		v2 := fn.NextValue()
		block.Instructions = append(block.Instructions, mir.Instruction{
			ID:   v2,
			Op:   "const",
			Type: "int",
			Operands: []mir.Operand{
				{Kind: mir.OperandLiteral, Literal: "3", Type: "int"},
			},
		})

		// Create nested expression: (10 + 5) * 3
		v3 := fn.NextValue()
		block.Instructions = append(block.Instructions, mir.Instruction{
			ID:   v3,
			Op:   "add",
			Type: "int",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: v0, Type: "int"},
				{Kind: mir.OperandValue, Value: v1, Type: "int"},
			},
		})

		v4 := fn.NextValue()
		block.Instructions = append(block.Instructions, mir.Instruction{
			ID:   v4,
			Op:   "mul",
			Type: "int",
			Operands: []mir.Operand{
				{Kind: mir.OperandValue, Value: v3, Type: "int"},
				{Kind: mir.OperandValue, Value: v2, Type: "int"},
			},
		})

		block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{{Kind: mir.OperandValue, Value: v4, Type: "int"}}}

		mod := &mir.Module{Functions: []*mir.Function{fn}}

		res, err := vm.Execute(mod, "main")
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		expected := 45 // (10 + 5) * 3
		if res.Value != expected {
			t.Errorf("Expected %d, got %v", expected, res.Value)
		}
	})
}
