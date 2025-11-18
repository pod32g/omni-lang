package mir

import (
	"testing"
)

func TestModuleString(t *testing.T) {
	module := &Module{
		Functions: []*Function{},
	}

	// Test that module can be created
	if len(module.Functions) != 0 {
		t.Errorf("Expected 0 functions, got %d", len(module.Functions))
	}
}

func TestNewFunction(t *testing.T) {
	params := []Param{
		{Name: "x", Type: "int"},
		{Name: "y", Type: "int"},
	}

	function := NewFunction("add", "int", params)
	if function == nil {
		t.Fatal("NewFunction returned nil")
	}

	if function.Name != "add" {
		t.Errorf("Expected name 'add', got '%s'", function.Name)
	}

	if function.ReturnType != "int" {
		t.Errorf("Expected return type 'int', got '%s'", function.ReturnType)
	}

	if len(function.Params) != 2 {
		t.Errorf("Expected 2 params, got %d", len(function.Params))
	}

	if len(function.Blocks) != 0 {
		t.Errorf("Expected 0 blocks initially, got %d", len(function.Blocks))
	}
}

func TestNewBlock(t *testing.T) {
	function := &Function{
		Name:       "test",
		ReturnType: "int",
		Params:     []Param{},
		Blocks:     []*BasicBlock{},
	}

	block := function.NewBlock("entry")
	if block == nil {
		t.Fatal("NewBlock returned nil")
	}

	if block.Name != "entry" {
		t.Errorf("Expected block name 'entry', got '%s'", block.Name)
	}

	if len(block.Instructions) != 0 {
		t.Errorf("Expected 0 instructions initially, got %d", len(block.Instructions))
	}

	if block.Terminator.Op != "" {
		t.Error("Expected empty terminator initially")
	}
}

func TestNextValue(t *testing.T) {
	function := &Function{
		Name:       "test",
		ReturnType: "int",
		Params:     []Param{},
		Blocks:     []*BasicBlock{},
	}

	// Test getting next value ID
	value1 := function.NextValue()
	if value1 != 0 {
		t.Errorf("Expected first value ID to be 0, got %d", value1)
	}

	value2 := function.NextValue()
	if value2 != 1 {
		t.Errorf("Expected second value ID to be 1, got %d", value2)
	}
}

func TestHasTerminator(t *testing.T) {
	block := &BasicBlock{
		Name:         "test",
		Instructions: []Instruction{},
		Terminator:   Terminator{},
	}

	if block.HasTerminator() {
		t.Error("Expected block without terminator to return false")
	}

	block.Terminator = Terminator{
		Op:       "ret",
		Operands: []Operand{},
	}

	if !block.HasTerminator() {
		t.Error("Expected block with terminator to return true")
	}
}

func TestBasicBlock(t *testing.T) {
	block := &BasicBlock{
		Name: "test_block",
		Instructions: []Instruction{
			{
				ID:   1,
				Op:   "const",
				Type: "int",
				Operands: []Operand{
					{Kind: OperandLiteral, Literal: "42", Type: "int"},
				},
			},
		},
		Terminator: Terminator{
			Op: "ret",
			Operands: []Operand{
				{Kind: OperandValue, Value: 1},
			},
		},
	}

	if block.Name != "test_block" {
		t.Errorf("Expected block name 'test_block', got '%s'", block.Name)
	}

	if len(block.Instructions) != 1 {
		t.Errorf("Expected 1 instruction, got %d", len(block.Instructions))
	}

	if block.Terminator.Op != "ret" {
		t.Errorf("Expected terminator op 'ret', got '%s'", block.Terminator.Op)
	}
}

func TestInstruction(t *testing.T) {
	instruction := Instruction{
		ID:   1,
		Op:   "add",
		Type: "int",
		Operands: []Operand{
			{Kind: OperandValue, Value: 2, Type: "int"},
			{Kind: OperandValue, Value: 3, Type: "int"},
		},
	}

	if instruction.ID != 1 {
		t.Errorf("Expected instruction ID 1, got %d", instruction.ID)
	}

	if instruction.Op != "add" {
		t.Errorf("Expected operation 'add', got '%s'", instruction.Op)
	}

	if instruction.Type != "int" {
		t.Errorf("Expected type 'int', got '%s'", instruction.Type)
	}

	if len(instruction.Operands) != 2 {
		t.Errorf("Expected 2 operands, got %d", len(instruction.Operands))
	}
}

func TestOperand(t *testing.T) {
	operand := Operand{
		Kind:    OperandLiteral,
		Literal: "42",
		Type:    "int",
	}

	if operand.Kind != OperandLiteral {
		t.Errorf("Expected operand kind OperandLiteral, got %d", operand.Kind)
	}

	if operand.Literal != "42" {
		t.Errorf("Expected literal '42', got '%s'", operand.Literal)
	}

	if operand.Type != "int" {
		t.Errorf("Expected type 'int', got '%s'", operand.Type)
	}
}

func TestTerminator(t *testing.T) {
	terminator := Terminator{
		Op: "ret",
		Operands: []Operand{
			{Kind: OperandLiteral, Literal: "0", Type: "int"},
		},
	}

	if terminator.Op != "ret" {
		t.Errorf("Expected terminator op 'ret', got '%s'", terminator.Op)
	}

	if len(terminator.Operands) != 1 {
		t.Errorf("Expected 1 operand, got %d", len(terminator.Operands))
	}
}

func TestParam(t *testing.T) {
	param := Param{
		Name: "x",
		Type: "int",
	}

	if param.Name != "x" {
		t.Errorf("Expected param name 'x', got '%s'", param.Name)
	}

	if param.Type != "int" {
		t.Errorf("Expected param type 'int', got '%s'", param.Type)
	}
}

// Additional comprehensive tests for MIR module

func TestValueIDString(t *testing.T) {
	tests := []struct {
		name     string
		valueID  ValueID
		expected string
	}{
		{
			name:     "InvalidValue",
			valueID:  InvalidValue,
			expected: "<invalid>",
		},
		{
			name:     "zero value",
			valueID:  0,
			expected: "%0",
		},
		{
			name:     "positive value",
			valueID:  42,
			expected: "%42",
		},
		{
			name:     "large value",
			valueID:  1000,
			expected: "%1000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.valueID.String()
			if result != tt.expected {
				t.Errorf("ValueID(%d).String() = %q, expected %q", tt.valueID, result, tt.expected)
			}
		})
	}
}

func TestNewFunctionWithParams(t *testing.T) {
	tests := []struct {
		name       string
		params     []Param
		expectedID ValueID
	}{
		{
			name:       "no params",
			params:     []Param{},
			expectedID: 0,
		},
		{
			name: "single param",
			params: []Param{
				{Name: "x", Type: "int"},
			},
			expectedID: 1,
		},
		{
			name: "multiple params",
			params: []Param{
				{Name: "a", Type: "int"},
				{Name: "b", Type: "int"},
				{Name: "c", Type: "string"},
			},
			expectedID: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := NewFunction("test", "int", tt.params)
			if fn == nil {
				t.Fatal("NewFunction returned nil")
			}

			// Check that params have correct IDs
			for i, param := range fn.Params {
				if param.ID != ValueID(i) {
					t.Errorf("Param %d has ID %d, expected %d", i, param.ID, i)
				}
			}

			// Check that nextValue is set correctly
			if fn.nextValue != tt.expectedID {
				t.Errorf("nextValue = %d, expected %d", fn.nextValue, tt.expectedID)
			}

			// Check that NextValue() returns the expected value
			nextID := fn.NextValue()
			if nextID != tt.expectedID {
				t.Errorf("NextValue() = %d, expected %d", nextID, tt.expectedID)
			}
		})
	}
}

func TestNextValueSequence(t *testing.T) {
	fn := NewFunction("test", "int", []Param{})

	// Test that NextValue() returns sequential IDs
	expected := ValueID(0)
	for i := 0; i < 10; i++ {
		id := fn.NextValue()
		if id != expected {
			t.Errorf("NextValue() iteration %d = %d, expected %d", i, id, expected)
		}
		expected++
	}
}

func TestNewBlockMultiple(t *testing.T) {
	fn := NewFunction("test", "int", []Param{})

	// Create multiple blocks
	block1 := fn.NewBlock("entry")
	block2 := fn.NewBlock("loop")
	block3 := fn.NewBlock("exit")

	if len(fn.Blocks) != 3 {
		t.Errorf("Expected 3 blocks, got %d", len(fn.Blocks))
	}

	if fn.Blocks[0] != block1 || fn.Blocks[1] != block2 || fn.Blocks[2] != block3 {
		t.Error("Blocks not stored in correct order")
	}

	if block1.Name != "entry" || block2.Name != "loop" || block3.Name != "exit" {
		t.Error("Block names not set correctly")
	}
}

func TestHasTerminatorVariations(t *testing.T) {
	tests := []struct {
		name       string
		terminator Terminator
		expected   bool
	}{
		{
			name:       "empty terminator",
			terminator: Terminator{},
			expected:   false,
		},
		{
			name:       "ret terminator",
			terminator: Terminator{Op: "ret"},
			expected:   true,
		},
		{
			name:       "jmp terminator",
			terminator: Terminator{Op: "jmp"},
			expected:   true,
		},
		{
			name:       "cbr terminator",
			terminator: Terminator{Op: "cbr"},
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block := &BasicBlock{
				Name:       "test",
				Terminator: tt.terminator,
			}
			result := block.HasTerminator()
			if result != tt.expected {
				t.Errorf("HasTerminator() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestModuleWithMultipleFunctions(t *testing.T) {
	module := &Module{
		Functions: []*Function{
			NewFunction("func1", "int", []Param{}),
			NewFunction("func2", "string", []Param{{Name: "x", Type: "int"}}),
			NewFunction("func3", "void", []Param{
				{Name: "a", Type: "int"},
				{Name: "b", Type: "string"},
			}),
		},
	}

	if len(module.Functions) != 3 {
		t.Errorf("Expected 3 functions, got %d", len(module.Functions))
	}

	if module.Functions[0].Name != "func1" {
		t.Errorf("Function 0 name = %q, expected %q", module.Functions[0].Name, "func1")
	}

	if module.Functions[1].ReturnType != "string" {
		t.Errorf("Function 1 return type = %q, expected %q", module.Functions[1].ReturnType, "string")
	}

	if len(module.Functions[2].Params) != 2 {
		t.Errorf("Function 2 params = %d, expected 2", len(module.Functions[2].Params))
	}
}

func TestInstructionWithInvalidValue(t *testing.T) {
	inst := Instruction{
		ID:   InvalidValue,
		Op:   "call.void",
		Type: "void",
	}

	if inst.ID != InvalidValue {
		t.Error("Instruction ID should be InvalidValue")
	}

	// Instructions with InvalidValue should still be valid
	if inst.Op != "call.void" {
		t.Error("Instruction op not set correctly")
	}
}

func TestOperandKinds(t *testing.T) {
	tests := []struct {
		name       string
		operand    Operand
		kind       OperandKind
		hasValue   bool
		hasLiteral bool
	}{
		{
			name: "value operand",
			operand: Operand{
				Kind:  OperandValue,
				Value: 42,
				Type:  "int",
			},
			kind:       OperandValue,
			hasValue:   true,
			hasLiteral: false,
		},
		{
			name: "literal operand",
			operand: Operand{
				Kind:    OperandLiteral,
				Literal: "42",
				Type:    "int",
			},
			kind:       OperandLiteral,
			hasValue:   false,
			hasLiteral: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.operand.Kind != tt.kind {
				t.Errorf("Operand kind = %d, expected %d", tt.operand.Kind, tt.kind)
			}
			if (tt.operand.Value != 0) != tt.hasValue {
				t.Errorf("Operand has value = %v, expected %v", tt.operand.Value != 0, tt.hasValue)
			}
			if (tt.operand.Literal != "") != tt.hasLiteral {
				t.Errorf("Operand has literal = %v, expected %v", tt.operand.Literal != "", tt.hasLiteral)
			}
		})
	}
}

func TestTerminatorWithOperands(t *testing.T) {
	tests := []struct {
		name       string
		terminator Terminator
		opCount    int
	}{
		{
			name: "ret with value",
			terminator: Terminator{
				Op: "ret",
				Operands: []Operand{
					{Kind: OperandValue, Value: 1},
				},
			},
			opCount: 1,
		},
		{
			name: "cbr with condition and targets",
			terminator: Terminator{
				Op: "cbr",
				Operands: []Operand{
					{Kind: OperandValue, Value: 0},
					{Kind: OperandLiteral, Literal: "block1"},
					{Kind: OperandLiteral, Literal: "block2"},
				},
			},
			opCount: 3,
		},
		{
			name: "jmp with target",
			terminator: Terminator{
				Op: "jmp",
				Operands: []Operand{
					{Kind: OperandLiteral, Literal: "block1"},
				},
			},
			opCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.terminator.Operands) != tt.opCount {
				t.Errorf("Terminator operands = %d, expected %d", len(tt.terminator.Operands), tt.opCount)
			}
			if tt.terminator.Op == "" {
				t.Error("Terminator op should not be empty")
			}
		})
	}
}

func TestParamWithID(t *testing.T) {
	param := Param{
		Name: "x",
		Type: "int",
		ID:   5,
	}

	if param.Name != "x" {
		t.Errorf("Param name = %q, expected %q", param.Name, "x")
	}

	if param.Type != "int" {
		t.Errorf("Param type = %q, expected %q", param.Type, "int")
	}

	if param.ID != 5 {
		t.Errorf("Param ID = %d, expected 5", param.ID)
	}
}

func TestComplexFunctionStructure(t *testing.T) {
	fn := NewFunction("complex", "int", []Param{
		{Name: "x", Type: "int"},
		{Name: "y", Type: "int"},
	})

	// Create entry block
	entry := fn.NewBlock("entry")
	entry.Instructions = []Instruction{
		{
			ID:   fn.NextValue(),
			Op:   "add",
			Type: "int",
			Operands: []Operand{
				{Kind: OperandValue, Value: 0, Type: "int"}, // x
				{Kind: OperandValue, Value: 1, Type: "int"}, // y
			},
		},
	}
	entry.Terminator = Terminator{
		Op: "ret",
		Operands: []Operand{
			{Kind: OperandValue, Value: 2}, // result of add
		},
	}

	// Verify structure
	if len(fn.Blocks) != 1 {
		t.Errorf("Expected 1 block, got %d", len(fn.Blocks))
	}

	if len(entry.Instructions) != 1 {
		t.Errorf("Expected 1 instruction, got %d", len(entry.Instructions))
	}

	if !entry.HasTerminator() {
		t.Error("Entry block should have terminator")
	}

	if entry.Instructions[0].Op != "add" {
		t.Errorf("Instruction op = %q, expected %q", entry.Instructions[0].Op, "add")
	}
}
