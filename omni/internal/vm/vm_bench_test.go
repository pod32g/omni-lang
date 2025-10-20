package vm

import (
	"testing"

	"github.com/omni-lang/omni/internal/mir"
)

func BenchmarkStringConcat(b *testing.B) {
	// Create a simple MIR module with string concatenation
	mod := &mir.Module{
		Functions: []*mir.Function{
			{
				Name: "main",
				Blocks: []*mir.BasicBlock{
					{
						Name: "entry",
						Instructions: []mir.Instruction{
							{
								ID:   1,
								Op:   "const",
								Type: "string",
								Operands: []mir.Operand{
									{Kind: mir.OperandLiteral, Literal: "Hello", Type: "string"},
								},
							},
							{
								ID:   2,
								Op:   "const",
								Type: "string",
								Operands: []mir.Operand{
									{Kind: mir.OperandLiteral, Literal: " World", Type: "string"},
								},
							},
							{
								ID:   3,
								Op:   "strcat",
								Type: "string",
								Operands: []mir.Operand{
									{Kind: mir.OperandValue, Value: 1, Type: "string"},
									{Kind: mir.OperandValue, Value: 2, Type: "string"},
								},
							},
						},
						Terminator: mir.Terminator{
							Op: "ret",
							Operands: []mir.Operand{
								{Kind: mir.OperandValue, Value: 3, Type: "string"},
							},
						},
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Execute(mod, "main")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkArithmetic(b *testing.B) {
	// Create a simple MIR module with arithmetic operations
	mod := &mir.Module{
		Functions: []*mir.Function{
			{
				Name: "main",
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
							{
								ID:   2,
								Op:   "const",
								Type: "int",
								Operands: []mir.Operand{
									{Kind: mir.OperandLiteral, Literal: "8", Type: "int"},
								},
							},
							{
								ID:   3,
								Op:   "add",
								Type: "int",
								Operands: []mir.Operand{
									{Kind: mir.OperandValue, Value: 1, Type: "int"},
									{Kind: mir.OperandValue, Value: 2, Type: "int"},
								},
							},
						},
						Terminator: mir.Terminator{
							Op: "ret",
							Operands: []mir.Operand{
								{Kind: mir.OperandValue, Value: 3, Type: "int"},
							},
						},
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Execute(mod, "main")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkInstructionDispatch(b *testing.B) {
	// Test instruction dispatch performance
	inst := mir.Instruction{
		ID:   1,
		Op:   "add",
		Type: "int",
		Operands: []mir.Operand{
			{Kind: mir.OperandLiteral, Literal: "1", Type: "int"},
			{Kind: mir.OperandLiteral, Literal: "2", Type: "int"},
		},
	}

	fr := &frame{
		values: make(map[mir.ValueID]Result),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := execInstruction(nil, fr, inst)
		if err != nil {
			b.Fatal(err)
		}
	}
}
