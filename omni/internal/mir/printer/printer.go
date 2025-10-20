package printer

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/omni-lang/omni/internal/mir"
)

// Format renders the MIR module as a deterministic textual representation.
func Format(mod *mir.Module) string {
	var buf bytes.Buffer
	for i, fn := range mod.Functions {
		if i > 0 {
			buf.WriteByte('\n')
		}
		buf.WriteString(formatFunction(fn))
	}
	return buf.String()
}

func formatFunction(fn *mir.Function) string {
	var buf bytes.Buffer
	params := make([]string, len(fn.Params))
	for i, p := range fn.Params {
		params[i] = fmt.Sprintf("%s:%s", p.Name, p.Type)
	}
	buf.WriteString(fmt.Sprintf("func %s(%s):%s\n", fn.Name, strings.Join(params, ","), fn.ReturnType))
	for _, block := range fn.Blocks {
		buf.WriteString(fmt.Sprintf("  block %s:\n", block.Name))
		for _, inst := range block.Instructions {
			buf.WriteString("    ")
			buf.WriteString(formatInstruction(inst))
			buf.WriteByte('\n')
		}
		if block.Terminator.Op != "" {
			buf.WriteString("    ")
			buf.WriteString(formatTerminator(block.Terminator))
			buf.WriteByte('\n')
		}
	}
	return buf.String()
}

func formatInstruction(inst mir.Instruction) string {
	var parts []string
	if inst.ID != mir.InvalidValue {
		parts = append(parts, fmt.Sprintf("%s =", inst.ID.String()))
	}
	op := inst.Op
	if inst.Type != "" {
		op = fmt.Sprintf("%s.%s", op, inst.Type)
	}
	parts = append(parts, op)
	if len(inst.Operands) > 0 {
		ops := make([]string, len(inst.Operands))
		for i, o := range inst.Operands {
			ops[i] = formatOperand(o)
		}
		parts = append(parts, strings.Join(ops, ", "))
	}
	return strings.Join(parts, " ")
}

func formatTerminator(term mir.Terminator) string {
	var parts []string
	op := term.Op
	if op == "" {
		op = "<missing>"
	}
	parts = append(parts, op)
	if len(term.Operands) > 0 {
		ops := make([]string, len(term.Operands))
		for i, o := range term.Operands {
			ops[i] = formatOperand(o)
		}
		parts = append(parts, strings.Join(ops, ", "))
	}
	return strings.Join(parts, " ")
}

func formatOperand(op mir.Operand) string {
	switch op.Kind {
	case mir.OperandValue:
		return op.Value.String()
	case mir.OperandLiteral:
		if op.Type != "" {
			return fmt.Sprintf("%s:%s", op.Literal, op.Type)
		}
		return op.Literal
	default:
		return "<unknown>"
	}
}
