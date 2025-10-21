package passes

import (
	"strconv"

	"github.com/omni-lang/omni/internal/mir"
)

// ConstFold performs simple constant folding over arithmetic and comparison instructions.
func ConstFold(mod *mir.Module) {
	if mod == nil {
		return
	}
	for _, fn := range mod.Functions {
		foldFunction(fn)
	}
}

func foldFunction(fn *mir.Function) {
	constValues := make(map[mir.ValueID]mir.Instruction)
	// Track which variables are modified by assign instructions
	modifiedVars := make(map[mir.ValueID]bool)

	// First pass: identify all variables that are modified by assign instructions
	for _, block := range fn.Blocks {
		for _, inst := range block.Instructions {
			if inst.Op == "assign" && len(inst.Operands) > 0 {
				// The first operand is the target variable being assigned to
				if inst.Operands[0].Kind == mir.OperandValue {
					modifiedVars[inst.Operands[0].Value] = true
				}
			}
		}
	}

	// Second pass: perform constant folding, but avoid folding expressions
	// that involve variables that are modified by assign instructions
	for _, block := range fn.Blocks {
		for i := range block.Instructions {
			inst := &block.Instructions[i]
			if inst.ID != mir.InvalidValue && inst.Op == "const" {
				constValues[inst.ID] = *inst
				continue
			}
			switch inst.Op {
			case "add", "sub", "mul", "div", "mod", "cmp.eq", "cmp.neq", "cmp.lt", "cmp.lte", "cmp.gt", "cmp.gte", "and", "or":
				// Check if any operands are modified variables
				hasModifiedVar := false
				for _, op := range inst.Operands {
					if op.Kind == mir.OperandValue && modifiedVars[op.Value] {
						hasModifiedVar = true
						break
					}
				}
				// Only fold if no modified variables are involved
				if !hasModifiedVar {
					if folded, ok := foldBinary(inst, constValues); ok {
						block.Instructions[i] = folded
						constValues[folded.ID] = folded
					}
				}
			}
		}
	}
}

func foldBinary(inst *mir.Instruction, consts map[mir.ValueID]mir.Instruction) (mir.Instruction, bool) {
	if len(inst.Operands) < 2 {
		return mir.Instruction{}, false
	}
	left, ok := constantOperand(inst.Operands[0], consts)
	if !ok {
		return mir.Instruction{}, false
	}
	right, ok := constantOperand(inst.Operands[1], consts)
	if !ok {
		return mir.Instruction{}, false
	}
	// Only handle integer arithmetic/comparisons for now.
	li, err := strconv.Atoi(left.Literal)
	if err != nil {
		return mir.Instruction{}, false
	}
	ri, err := strconv.Atoi(right.Literal)
	if err != nil {
		return mir.Instruction{}, false
	}
	result := mir.Instruction{ID: inst.ID, Op: "const"}
	switch inst.Op {
	case "add":
		result.Type = inst.Type
		result.Operands = []mir.Operand{{Kind: mir.OperandLiteral, Literal: strconv.Itoa(li + ri), Type: inst.Type}}
	case "sub":
		result.Type = inst.Type
		result.Operands = []mir.Operand{{Kind: mir.OperandLiteral, Literal: strconv.Itoa(li - ri), Type: inst.Type}}
	case "mul":
		result.Type = inst.Type
		result.Operands = []mir.Operand{{Kind: mir.OperandLiteral, Literal: strconv.Itoa(li * ri), Type: inst.Type}}
	case "div":
		if ri == 0 {
			return mir.Instruction{}, false
		}
		result.Type = inst.Type
		result.Operands = []mir.Operand{{Kind: mir.OperandLiteral, Literal: strconv.Itoa(li / ri), Type: inst.Type}}
	case "mod":
		if ri == 0 {
			return mir.Instruction{}, false
		}
		result.Type = inst.Type
		result.Operands = []mir.Operand{{Kind: mir.OperandLiteral, Literal: strconv.Itoa(li % ri), Type: inst.Type}}
	case "cmp.eq":
		result.Type = "bool"
		result.Operands = []mir.Operand{{Kind: mir.OperandLiteral, Literal: strconv.FormatBool(li == ri), Type: "bool"}}
	case "cmp.neq":
		result.Type = "bool"
		result.Operands = []mir.Operand{{Kind: mir.OperandLiteral, Literal: strconv.FormatBool(li != ri), Type: "bool"}}
	case "cmp.lt":
		result.Type = "bool"
		result.Operands = []mir.Operand{{Kind: mir.OperandLiteral, Literal: strconv.FormatBool(li < ri), Type: "bool"}}
	case "cmp.lte":
		result.Type = "bool"
		result.Operands = []mir.Operand{{Kind: mir.OperandLiteral, Literal: strconv.FormatBool(li <= ri), Type: "bool"}}
	case "cmp.gt":
		result.Type = "bool"
		result.Operands = []mir.Operand{{Kind: mir.OperandLiteral, Literal: strconv.FormatBool(li > ri), Type: "bool"}}
	case "cmp.gte":
		result.Type = "bool"
		result.Operands = []mir.Operand{{Kind: mir.OperandLiteral, Literal: strconv.FormatBool(li >= ri), Type: "bool"}}
	case "and":
		result.Type = "bool"
		result.Operands = []mir.Operand{{Kind: mir.OperandLiteral, Literal: strconv.FormatBool(li != 0 && ri != 0), Type: "bool"}}
	case "or":
		result.Type = "bool"
		result.Operands = []mir.Operand{{Kind: mir.OperandLiteral, Literal: strconv.FormatBool(li != 0 || ri != 0), Type: "bool"}}
	default:
		return mir.Instruction{}, false
	}
	return result, true
}

func constantOperand(op mir.Operand, consts map[mir.ValueID]mir.Instruction) (mir.Operand, bool) {
	switch op.Kind {
	case mir.OperandLiteral:
		return op, true
	case mir.OperandValue:
		inst, ok := consts[op.Value]
		if !ok || inst.Op != "const" || len(inst.Operands) == 0 {
			return mir.Operand{}, false
		}
		operand := inst.Operands[0]
		if operand.Kind != mir.OperandLiteral {
			return mir.Operand{}, false
		}
		return operand, true
	default:
		return mir.Operand{}, false
	}
}
