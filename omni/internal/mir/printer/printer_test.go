package printer

import (
	"strings"
	"testing"

	"github.com/omni-lang/omni/internal/mir"
)

func TestFormat(t *testing.T) {
	module := &mir.Module{
		Functions: []*mir.Function{
			{
				Name:       "test_func",
				ReturnType: "int",
				Params: []mir.Param{
					{Name: "x", Type: "int"},
				},
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

	result := Format(module)
	if result == "" {
		t.Error("Expected non-empty formatted output")
	}

	// Check for basic structure
	if !strings.Contains(result, "test_func") {
		t.Error("Expected function name in output")
	}

	if !strings.Contains(result, "entry") {
		t.Error("Expected block name in output")
	}
}

func TestFormatFunction(t *testing.T) {
	function := &mir.Function{
		Name:       "add",
		ReturnType: "int",
		Params: []mir.Param{
			{Name: "a", Type: "int"},
			{Name: "b", Type: "int"},
		},
		Blocks: []*mir.BasicBlock{
			{
				Name: "entry",
				Instructions: []mir.Instruction{
					{
						ID:   1,
						Op:   "add",
						Type: "int",
						Operands: []mir.Operand{
							{Kind: mir.OperandValue, Value: 0, Type: "int"},
							{Kind: mir.OperandValue, Value: 1, Type: "int"},
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
	}

	result := formatFunction(function)
	if result == "" {
		t.Error("Expected non-empty function format")
	}

	if !strings.Contains(result, "add") {
		t.Error("Expected function name in output")
	}
}

func TestFormatInstruction(t *testing.T) {
	instruction := mir.Instruction{
		ID:   1,
		Op:   "const",
		Type: "int",
		Operands: []mir.Operand{
			{Kind: mir.OperandLiteral, Literal: "42", Type: "int"},
		},
	}

	result := formatInstruction(instruction)
	if result == "" {
		t.Error("Expected non-empty instruction format")
	}

	if !strings.Contains(result, "const") {
		t.Error("Expected operation name in output")
	}
}

func TestFormatTerminator(t *testing.T) {
	terminator := mir.Terminator{
		Op: "ret",
		Operands: []mir.Operand{
			{Kind: mir.OperandLiteral, Literal: "0", Type: "int"},
		},
	}

	result := formatTerminator(terminator)
	if result == "" {
		t.Error("Expected non-empty terminator format")
	}

	if !strings.Contains(result, "ret") {
		t.Error("Expected terminator kind in output")
	}
}

func TestFormatOperand(t *testing.T) {
	tests := []struct {
		name     string
		operand  mir.Operand
		expected string
	}{
		{
			name: "literal operand",
			operand: mir.Operand{
				Kind:    mir.OperandLiteral,
				Literal: "42",
				Type:    "int",
			},
			expected: "42",
		},
		{
			name: "value operand",
			operand: mir.Operand{
				Kind:  mir.OperandValue,
				Value: 1,
				Type:  "int",
			},
			expected: "%1",
		},
		{
			name: "value operand",
			operand: mir.Operand{
				Kind:  mir.OperandValue,
				Value: 0,
				Type:  "int",
			},
			expected: "%0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatOperand(tt.operand)
			if result == "" {
				t.Error("Expected non-empty operand format")
			}
		})
	}
}
