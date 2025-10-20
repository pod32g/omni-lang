package passes

import (
	"fmt"

	"github.com/omni-lang/omni/internal/mir"
)

// Pipeline owns an ordered set of MIR passes, including verification.
type Pipeline struct {
	Name string
}

// NewPipeline constructs the pass pipeline descriptor.
func NewPipeline(name string) Pipeline {
	return Pipeline{Name: name}
}

// Run executes the configured passes over the module.
func (p Pipeline) Run(mod mir.Module) (mir.Module, error) {
	if err := Verify(&mod); err != nil {
		return mir.Module{}, err
	}
	ConstFold(&mod)
	if err := Verify(&mod); err != nil {
		return mir.Module{}, err
	}
	return mod, nil
}

// Verify ensures the module satisfies basic structural invariants expected by downstream passes.
func Verify(mod *mir.Module) error {
	if mod == nil {
		return fmt.Errorf("mir verifier: nil module")
	}
	for _, fn := range mod.Functions {
		if err := verifyFunction(fn); err != nil {
			return err
		}
	}
	return nil
}

func verifyFunction(fn *mir.Function) error {
	if fn == nil {
		return fmt.Errorf("mir verifier: nil function")
	}
	if fn.Name == "" {
		return fmt.Errorf("mir verifier: function with empty name")
	}
	if len(fn.Blocks) == 0 {
		return fmt.Errorf("mir verifier: function %s has no basic blocks", fn.Name)
	}
	blockNames := make(map[string]struct{}, len(fn.Blocks))
	for _, block := range fn.Blocks {
		if block == nil {
			return fmt.Errorf("mir verifier: function %s has nil block", fn.Name)
		}
		if block.Name == "" {
			return fmt.Errorf("mir verifier: function %s has block with empty name", fn.Name)
		}
		if _, exists := blockNames[block.Name]; exists {
			return fmt.Errorf("mir verifier: function %s has duplicate block name %q", fn.Name, block.Name)
		}
		blockNames[block.Name] = struct{}{}
	}
	for _, block := range fn.Blocks {
		if block.Terminator.Op == "" {
			return fmt.Errorf("mir verifier: block %s in function %s missing terminator", block.Name, fn.Name)
		}
		if err := verifyTerminator(block.Terminator, blockNames); err != nil {
			return fmt.Errorf("mir verifier: block %s in function %s: %w", block.Name, fn.Name, err)
		}
	}
	return nil
}

func verifyTerminator(term mir.Terminator, blocks map[string]struct{}) error {
	switch term.Op {
	case "ret":
		// return may optionally carry a single operand; nothing further to validate here.
		return nil
	case "br":
		if len(term.Operands) != 1 {
			return fmt.Errorf("branch expects 1 operand, got %d", len(term.Operands))
		}
		return verifyBlockOperand(term.Operands[0], blocks)
	case "cbr":
		if len(term.Operands) != 3 {
			return fmt.Errorf("conditional branch expects 3 operands, got %d", len(term.Operands))
		}
		if term.Operands[0].Kind != mir.OperandValue && term.Operands[0].Kind != mir.OperandLiteral {
			return fmt.Errorf("conditional branch requires value operand for condition")
		}
		if err := verifyBlockOperand(term.Operands[1], blocks); err != nil {
			return err
		}
		return verifyBlockOperand(term.Operands[2], blocks)
	default:
		return fmt.Errorf("unsupported terminator %q", term.Op)
	}
}

func verifyBlockOperand(op mir.Operand, blocks map[string]struct{}) error {
	if op.Kind != mir.OperandLiteral {
		return fmt.Errorf("branch target must be literal block name")
	}
	if op.Literal == "" {
		return fmt.Errorf("branch target literal cannot be empty")
	}
	if _, ok := blocks[op.Literal]; !ok {
		return fmt.Errorf("branch target %q not found", op.Literal)
	}
	return nil
}
