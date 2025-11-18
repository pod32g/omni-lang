package passes

import (
	"testing"

	"github.com/omni-lang/omni/internal/mir"
)

func TestNewPipeline(t *testing.T) {
	// Test creating a new pipeline
	name := "test_pipeline"
	pipeline := NewPipeline(name)

	if pipeline.Name != name {
		t.Errorf("Expected pipeline name '%s', got '%s'", name, pipeline.Name)
	}
}

func TestPipelineRun(t *testing.T) {
	// Test running a pipeline with a valid module
	module := mir.Module{
		Functions: []*mir.Function{
			{
				Name:       "test_func",
				ReturnType: "void",
				Params:     []mir.Param{},
				Blocks: []*mir.BasicBlock{
					{
						Name: "entry",
						Instructions: []mir.Instruction{
							{
								ID:   1,
								Op:   "const",
								Type: "int",
								Operands: []mir.Operand{
									{Kind: mir.OperandLiteral, Literal: "42", Type: "int"},
								},
							},
						},
						Terminator: mir.Terminator{
							Op: "ret",
							Operands: []mir.Operand{
								{Kind: mir.OperandValue, Value: 1},
							},
						},
					},
				},
			},
		},
	}

	pipeline := NewPipeline("test")
	result, err := pipeline.Run(module)
	if err != nil {
		t.Errorf("Pipeline.Run failed: %v", err)
	}

	if len(result.Functions) != 1 {
		t.Errorf("Expected 1 function, got %d", len(result.Functions))
	}

	if result.Functions[0].Name != "test_func" {
		t.Errorf("Expected function name 'test_func', got '%s'", result.Functions[0].Name)
	}
}

func TestPipelineRunEmptyModule(t *testing.T) {
	// Test running a pipeline with an empty module
	module := mir.Module{
		Functions: []*mir.Function{},
	}

	pipeline := NewPipeline("test")
	result, err := pipeline.Run(module)
	if err != nil {
		t.Errorf("Pipeline.Run failed: %v", err)
	}

	if len(result.Functions) != 0 {
		t.Errorf("Expected 0 functions, got %d", len(result.Functions))
	}
}

func TestVerify(t *testing.T) {
	// Test verifying a valid module
	module := &mir.Module{
		Functions: []*mir.Function{
			{
				Name:       "test_func",
				ReturnType: "void",
				Params:     []mir.Param{},
				Blocks: []*mir.BasicBlock{
					{
						Name:         "entry",
						Instructions: []mir.Instruction{},
						Terminator: mir.Terminator{
							Op:       "ret",
							Operands: []mir.Operand{},
						},
					},
				},
			},
		},
	}

	err := Verify(module)
	if err != nil {
		t.Errorf("Verify failed: %v", err)
	}
}

func TestVerifyNilModule(t *testing.T) {
	// Test verifying a nil module
	err := Verify(nil)
	if err == nil {
		t.Error("Expected error for nil module")
	}

	expectedError := "mir verifier: nil module"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestVerifyEmptyModule(t *testing.T) {
	// Test verifying an empty module
	module := &mir.Module{
		Functions: []*mir.Function{},
	}

	err := Verify(module)
	if err != nil {
		t.Errorf("Verify failed: %v", err)
	}
}

func TestVerifyFunction(t *testing.T) {
	// Test verifying a valid function
	function := &mir.Function{
		Name:       "test_func",
		ReturnType: "void",
		Params:     []mir.Param{},
		Blocks: []*mir.BasicBlock{
			{
				Name:         "entry",
				Instructions: []mir.Instruction{},
				Terminator: mir.Terminator{
					Op:       "ret",
					Operands: []mir.Operand{},
				},
			},
		},
	}

	err := verifyFunction(function)
	if err != nil {
		t.Errorf("verifyFunction failed: %v", err)
	}
}

func TestVerifyFunctionNil(t *testing.T) {
	// Test verifying a nil function
	err := verifyFunction(nil)
	if err == nil {
		t.Error("Expected error for nil function")
	}

	expectedError := "mir verifier: nil function"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestVerifyFunctionEmptyName(t *testing.T) {
	// Test verifying a function with empty name
	function := &mir.Function{
		Name:       "",
		ReturnType: "void",
		Params:     []mir.Param{},
		Blocks: []*mir.BasicBlock{
			{
				Name:         "entry",
				Instructions: []mir.Instruction{},
				Terminator: mir.Terminator{
					Op:       "ret",
					Operands: []mir.Operand{},
				},
			},
		},
	}

	err := verifyFunction(function)
	if err == nil {
		t.Error("Expected error for function with empty name")
	}

	expectedError := "mir verifier: function with empty name"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestVerifyFunctionEmptyBlocks(t *testing.T) {
	// Test verifying a function with empty blocks
	function := &mir.Function{
		Name:       "test_func",
		ReturnType: "void",
		Params:     []mir.Param{},
		Blocks:     []*mir.BasicBlock{},
	}

	err := verifyFunction(function)
	if err == nil {
		t.Error("Expected error for function with empty blocks")
	}

	expectedError := "mir verifier: function test_func has no basic blocks"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestVerifyFunctionNoEntryBlock(t *testing.T) {
	// Test verifying a function without an entry block
	// Note: The current implementation doesn't check for entry block specifically
	// So this test should pass without error
	function := &mir.Function{
		Name:       "test_func",
		ReturnType: "void",
		Params:     []mir.Param{},
		Blocks: []*mir.BasicBlock{
			{
				Name:         "not_entry",
				Instructions: []mir.Instruction{},
				Terminator: mir.Terminator{
					Op:       "ret",
					Operands: []mir.Operand{},
				},
			},
		},
	}

	err := verifyFunction(function)
	if err != nil {
		t.Errorf("verifyFunction failed: %v", err)
	}
}

func TestVerifyFunctionNilBlock(t *testing.T) {
	// Test verifying a function with nil block
	function := &mir.Function{
		Name:       "test_func",
		ReturnType: "void",
		Params:     []mir.Param{},
		Blocks: []*mir.BasicBlock{
			nil,
		},
	}

	err := verifyFunction(function)
	if err == nil {
		t.Error("Expected error for function with nil block")
	}

	expectedError := "mir verifier: function test_func has nil block"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestVerifyFunctionEmptyBlockName(t *testing.T) {
	// Test verifying a function with empty block name
	function := &mir.Function{
		Name:       "test_func",
		ReturnType: "void",
		Params:     []mir.Param{},
		Blocks: []*mir.BasicBlock{
			{
				Name:         "",
				Instructions: []mir.Instruction{},
				Terminator: mir.Terminator{
					Op:       "ret",
					Operands: []mir.Operand{},
				},
			},
		},
	}

	err := verifyFunction(function)
	if err == nil {
		t.Error("Expected error for function with empty block name")
	}

	expectedError := "mir verifier: function test_func has block with empty name"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestVerifyFunctionBlockWithoutTerminator(t *testing.T) {
	// Test verifying a function with block without terminator
	function := &mir.Function{
		Name:       "test_func",
		ReturnType: "void",
		Params:     []mir.Param{},
		Blocks: []*mir.BasicBlock{
			{
				Name:         "entry",
				Instructions: []mir.Instruction{},
				Terminator:   mir.Terminator{},
			},
		},
	}

	err := verifyFunction(function)
	if err == nil {
		t.Error("Expected error for function with block without terminator")
	}

	expectedError := "mir verifier: block entry in function test_func missing terminator"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestVerifyFunctionBlockWithInvalidTerminator(t *testing.T) {
	// Test verifying a function with block with invalid terminator
	function := &mir.Function{
		Name:       "test_func",
		ReturnType: "void",
		Params:     []mir.Param{},
		Blocks: []*mir.BasicBlock{
			{
				Name:         "entry",
				Instructions: []mir.Instruction{},
				Terminator: mir.Terminator{
					Op:       "invalid",
					Operands: []mir.Operand{},
				},
			},
		},
	}

	err := verifyFunction(function)
	if err == nil {
		t.Error("Expected error for function with block with invalid terminator")
	}

	expectedError := "mir verifier: block entry in function test_func: unsupported terminator \"invalid\""
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestVerifyFunctionBlockWithValidReturn(t *testing.T) {
	// Test verifying a function with block with valid return
	function := &mir.Function{
		Name:       "test_func",
		ReturnType: "void",
		Params:     []mir.Param{},
		Blocks: []*mir.BasicBlock{
			{
				Name:         "entry",
				Instructions: []mir.Instruction{},
				Terminator: mir.Terminator{
					Op:       "ret",
					Operands: []mir.Operand{},
				},
			},
		},
	}

	err := verifyFunction(function)
	if err != nil {
		t.Errorf("verifyFunction failed: %v", err)
	}
}

// Additional tests for verification functions

func TestVerifyInstructionComparison(t *testing.T) {
	tests := []struct {
		name        string
		instruction mir.Instruction
		expectError bool
	}{
		{
			name: "valid comparison with 2 operands",
			instruction: mir.Instruction{
				Op:   "cmp.eq",
				Type: "bool",
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: 0},
					{Kind: mir.OperandValue, Value: 1},
				},
			},
			expectError: false,
		},
		{
			name: "comparison with insufficient operands",
			instruction: mir.Instruction{
				Op:   "cmp.eq",
				Type: "bool",
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: 0},
				},
			},
			expectError: true,
		},
		{
			name: "logical and with 2 operands",
			instruction: mir.Instruction{
				Op:   "and",
				Type: "bool",
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: 0},
					{Kind: mir.OperandValue, Value: 1},
				},
			},
			expectError: false,
		},
		{
			name: "logical or with 2 operands",
			instruction: mir.Instruction{
				Op:   "or",
				Type: "bool",
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: 0},
					{Kind: mir.OperandValue, Value: 1},
				},
			},
			expectError: false,
		},
		{
			name: "unsupported instruction",
			instruction: mir.Instruction{
				Op:   "invalid_op",
				Type: "int",
				Operands: []mir.Operand{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifyInstruction(tt.instruction)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestVerifyInstructionPHI(t *testing.T) {
	tests := []struct {
		name        string
		instruction mir.Instruction
		expectError bool
	}{
		{
			name: "valid PHI with 2 operands",
			instruction: mir.Instruction{
				Op:   "phi",
				Type: "int",
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: 0},
					{Kind: mir.OperandLiteral, Literal: "block1"},
				},
			},
			expectError: false,
		},
		{
			name: "PHI with odd number of operands",
			instruction: mir.Instruction{
				Op:   "phi",
				Type: "int",
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: 0},
				},
			},
			expectError: true,
		},
		{
			name: "PHI with insufficient operands",
			instruction: mir.Instruction{
				Op:   "phi",
				Type: "int",
				Operands: []mir.Operand{},
			},
			expectError: true,
		},
		{
			name: "PHI with 4 operands (2 pairs)",
			instruction: mir.Instruction{
				Op:   "phi",
				Type: "int",
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: 0},
					{Kind: mir.OperandLiteral, Literal: "block1"},
					{Kind: mir.OperandValue, Value: 1},
					{Kind: mir.OperandLiteral, Literal: "block2"},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifyInstruction(tt.instruction)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestVerifyTerminatorBR(t *testing.T) {
	blocks := map[string]struct{}{
		"block1": {},
		"block2": {},
	}

	tests := []struct {
		name        string
		terminator  mir.Terminator
		expectError bool
	}{
		{
			name: "valid branch with 1 operand",
			terminator: mir.Terminator{
				Op: "br",
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: "block1"},
				},
			},
			expectError: false,
		},
		{
			name: "branch with wrong operand count",
			terminator: mir.Terminator{
				Op:       "br",
				Operands: []mir.Operand{},
			},
			expectError: true,
		},
		{
			name: "branch to non-existent block",
			terminator: mir.Terminator{
				Op: "br",
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: "nonexistent"},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifyTerminator(tt.terminator, blocks)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestVerifyTerminatorCBR(t *testing.T) {
	blocks := map[string]struct{}{
		"block1": {},
		"block2": {},
	}

	tests := []struct {
		name        string
		terminator  mir.Terminator
		expectError bool
	}{
		{
			name: "valid conditional branch",
			terminator: mir.Terminator{
				Op: "cbr",
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: 0},
					{Kind: mir.OperandLiteral, Literal: "block1"},
					{Kind: mir.OperandLiteral, Literal: "block2"},
				},
			},
			expectError: false,
		},
		{
			name: "conditional branch with wrong operand count",
			terminator: mir.Terminator{
				Op: "cbr",
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: 0},
				},
			},
			expectError: true,
		},
		{
			name: "conditional branch with invalid condition",
			terminator: mir.Terminator{
				Op: "cbr",
				Operands: []mir.Operand{
					{Kind: mir.OperandLiteral, Literal: "true"},
					{Kind: mir.OperandLiteral, Literal: "block1"},
					{Kind: mir.OperandLiteral, Literal: "block2"},
				},
			},
			expectError: false, // Literal is also acceptable
		},
		{
			name: "conditional branch to non-existent block",
			terminator: mir.Terminator{
				Op: "cbr",
				Operands: []mir.Operand{
					{Kind: mir.OperandValue, Value: 0},
					{Kind: mir.OperandLiteral, Literal: "nonexistent"},
					{Kind: mir.OperandLiteral, Literal: "block2"},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifyTerminator(tt.terminator, blocks)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestVerifyBlockOperand(t *testing.T) {
	blocks := map[string]struct{}{
		"block1": {},
		"block2": {},
	}

	tests := []struct {
		name        string
		operand     mir.Operand
		expectError bool
	}{
		{
			name: "valid block operand",
			operand: mir.Operand{
				Kind:    mir.OperandLiteral,
				Literal: "block1",
			},
			expectError: false,
		},
		{
			name: "block operand with value kind",
			operand: mir.Operand{
				Kind:  mir.OperandValue,
				Value: 0,
			},
			expectError: true,
		},
		{
			name: "block operand with empty literal",
			operand: mir.Operand{
				Kind:    mir.OperandLiteral,
				Literal: "",
			},
			expectError: true,
		},
		{
			name: "block operand to non-existent block",
			operand: mir.Operand{
				Kind:    mir.OperandLiteral,
				Literal: "nonexistent",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifyBlockOperand(tt.operand, blocks)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestVerifyFunctionDuplicateBlockNames(t *testing.T) {
	function := &mir.Function{
		Name:       "test_func",
		ReturnType: "void",
		Params:     []mir.Param{},
		Blocks: []*mir.BasicBlock{
			{
				Name:         "entry",
				Instructions: []mir.Instruction{},
				Terminator: mir.Terminator{
					Op:       "ret",
					Operands: []mir.Operand{},
				},
			},
			{
				Name:         "entry",
				Instructions: []mir.Instruction{},
				Terminator: mir.Terminator{
					Op:       "ret",
					Operands: []mir.Operand{},
				},
			},
		},
	}

	err := verifyFunction(function)
	if err == nil {
		t.Error("expected error for duplicate block names")
	}

	expectedError := "mir verifier: function test_func has duplicate block name \"entry\""
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}
