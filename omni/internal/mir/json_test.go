package mir

import (
	"encoding/json"
	"testing"
)

func TestToJSONEmptyModule(t *testing.T) {
	module := &Module{
		Functions: []*Function{},
	}

	jsonData, err := module.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("Expected non-empty JSON output")
	}

	// Verify it's valid JSON
	var decoded MirModule
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(decoded.Functions) != 0 {
		t.Errorf("Expected 0 functions, got %d", len(decoded.Functions))
	}
}

func TestToJSONSimpleFunction(t *testing.T) {
	fn := NewFunction("test", "int", []Param{
		{Name: "x", Type: "int"},
	})

	entry := fn.NewBlock("entry")
	entry.Instructions = []Instruction{
		{
			ID:   fn.NextValue(),
			Op:   "const",
			Type: "int",
			Operands: []Operand{
				{Kind: OperandLiteral, Literal: "42", Type: "int"},
			},
		},
	}
	entry.Terminator = Terminator{
		Op: "ret",
		Operands: []Operand{
			{Kind: OperandValue, Value: 0},
		},
	}

	module := &Module{
		Functions: []*Function{fn},
	}

	jsonData, err := module.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var decoded MirModule
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(decoded.Functions) != 1 {
		t.Fatalf("Expected 1 function, got %d", len(decoded.Functions))
	}

	mirFunc := decoded.Functions[0]
	if mirFunc.Name != "test" {
		t.Errorf("Function name = %q, expected %q", mirFunc.Name, "test")
	}

	if mirFunc.ReturnType != "int" {
		t.Errorf("Return type = %q, expected %q", mirFunc.ReturnType, "int")
	}

	if len(mirFunc.Params) != 1 {
		t.Fatalf("Expected 1 param, got %d", len(mirFunc.Params))
	}

	if mirFunc.Params[0].Name != "x" {
		t.Errorf("Param name = %q, expected %q", mirFunc.Params[0].Name, "x")
	}

	if len(mirFunc.Blocks) != 1 {
		t.Fatalf("Expected 1 block, got %d", len(mirFunc.Blocks))
	}

	if mirFunc.Blocks[0].Name != "entry" {
		t.Errorf("Block name = %q, expected %q", mirFunc.Blocks[0].Name, "entry")
	}

	if len(mirFunc.Blocks[0].Instructions) != 1 {
		t.Fatalf("Expected 1 instruction, got %d", len(mirFunc.Blocks[0].Instructions))
	}

	inst := mirFunc.Blocks[0].Instructions[0]
	if inst.Op != "const" {
		t.Errorf("Instruction op = %q, expected %q", inst.Op, "const")
	}

	if inst.InstType != "int" {
		t.Errorf("Instruction type = %q, expected %q", inst.InstType, "int")
	}

	if len(inst.Operands) != 1 {
		t.Fatalf("Expected 1 operand, got %d", len(inst.Operands))
	}

	op := inst.Operands[0]
	if op.Kind != "literal" {
		t.Errorf("Operand kind = %q, expected %q", op.Kind, "literal")
	}

	if op.Literal == nil || *op.Literal != "42" {
		t.Errorf("Operand literal = %v, expected %q", op.Literal, "42")
	}
}

func TestToJSONWithValueOperands(t *testing.T) {
	fn := NewFunction("add", "int", []Param{
		{Name: "a", Type: "int"},
		{Name: "b", Type: "int"},
	})

	entry := fn.NewBlock("entry")
	entry.Instructions = []Instruction{
		{
			ID:   fn.NextValue(),
			Op:   "add",
			Type: "int",
			Operands: []Operand{
				{Kind: OperandValue, Value: 0, Type: "int"}, // a
				{Kind: OperandValue, Value: 1, Type: "int"}, // b
			},
		},
	}
	entry.Terminator = Terminator{
		Op: "ret",
		Operands: []Operand{
			{Kind: OperandValue, Value: 2},
		},
	}

	module := &Module{
		Functions: []*Function{fn},
	}

	jsonData, err := module.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var decoded MirModule
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	inst := decoded.Functions[0].Blocks[0].Instructions[0]
	if len(inst.Operands) != 2 {
		t.Fatalf("Expected 2 operands, got %d", len(inst.Operands))
	}

	op1 := inst.Operands[0]
	if op1.Kind != "value" {
		t.Errorf("Operand 0 kind = %q, expected %q", op1.Kind, "value")
	}

	if op1.Value == nil || *op1.Value != 0 {
		t.Errorf("Operand 0 value = %v, expected 0", op1.Value)
	}

	op2 := inst.Operands[1]
	if op2.Kind != "value" {
		t.Errorf("Operand 1 kind = %q, expected %q", op2.Kind, "value")
	}

	if op2.Value == nil || *op2.Value != 1 {
		t.Errorf("Operand 1 value = %v, expected 1", op2.Value)
	}
}

func TestToJSONMultipleFunctions(t *testing.T) {
	module := &Module{
		Functions: []*Function{
			NewFunction("func1", "int", []Param{}),
			NewFunction("func2", "string", []Param{{Name: "x", Type: "int"}}),
		},
	}

	jsonData, err := module.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var decoded MirModule
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(decoded.Functions) != 2 {
		t.Fatalf("Expected 2 functions, got %d", len(decoded.Functions))
	}

	if decoded.Functions[0].Name != "func1" {
		t.Errorf("Function 0 name = %q, expected %q", decoded.Functions[0].Name, "func1")
	}

	if decoded.Functions[1].Name != "func2" {
		t.Errorf("Function 1 name = %q, expected %q", decoded.Functions[1].Name, "func2")
	}
}

func TestToJSONWithMultipleBlocks(t *testing.T) {
	fn := NewFunction("test", "int", []Param{})

	entry := fn.NewBlock("entry")
	entry.Instructions = []Instruction{
		{
			ID:   fn.NextValue(),
			Op:   "const",
			Type: "int",
			Operands: []Operand{
				{Kind: OperandLiteral, Literal: "0", Type: "int"},
			},
		},
	}
	entry.Terminator = Terminator{
		Op: "cbr",
		Operands: []Operand{
			{Kind: OperandValue, Value: 0},
			{Kind: OperandLiteral, Literal: "then", Type: "string"},
			{Kind: OperandLiteral, Literal: "else", Type: "string"},
		},
	}

	thenBlock := fn.NewBlock("then")
	thenBlock.Instructions = []Instruction{
		{
			ID:   fn.NextValue(),
			Op:   "const",
			Type: "int",
			Operands: []Operand{
				{Kind: OperandLiteral, Literal: "1", Type: "int"},
			},
		},
	}
	thenBlock.Terminator = Terminator{
		Op: "ret",
		Operands: []Operand{
			{Kind: OperandValue, Value: 1},
		},
	}

	module := &Module{
		Functions: []*Function{fn},
	}

	jsonData, err := module.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var decoded MirModule
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(decoded.Functions[0].Blocks) != 2 {
		t.Fatalf("Expected 2 blocks, got %d", len(decoded.Functions[0].Blocks))
	}

	if decoded.Functions[0].Blocks[0].Name != "entry" {
		t.Errorf("Block 0 name = %q, expected %q", decoded.Functions[0].Blocks[0].Name, "entry")
	}

	if decoded.Functions[0].Blocks[1].Name != "then" {
		t.Errorf("Block 1 name = %q, expected %q", decoded.Functions[0].Blocks[1].Name, "then")
	}
}

func TestToJSONTerminatorOperands(t *testing.T) {
	fn := NewFunction("test", "void", []Param{})

	entry := fn.NewBlock("entry")
	entry.Terminator = Terminator{
		Op: "ret",
		Operands: []Operand{
			{Kind: OperandLiteral, Literal: "0", Type: "int"},
		},
	}

	module := &Module{
		Functions: []*Function{fn},
	}

	jsonData, err := module.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var decoded MirModule
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	term := decoded.Functions[0].Blocks[0].Terminator
	if term.Op != "ret" {
		t.Errorf("Terminator op = %q, expected %q", term.Op, "ret")
	}

	if len(term.Operands) != 1 {
		t.Fatalf("Expected 1 terminator operand, got %d", len(term.Operands))
	}

	op := term.Operands[0]
	if op.Kind != "literal" {
		t.Errorf("Terminator operand kind = %q, expected %q", op.Kind, "literal")
	}

	if op.Literal == nil || *op.Literal != "0" {
		t.Errorf("Terminator operand literal = %v, expected %q", op.Literal, "0")
	}
}
