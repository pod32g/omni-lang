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
