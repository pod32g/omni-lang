package mir

import (
	"encoding/json"
)

// MirModule represents the JSON structure for MIR modules
type MirModule struct {
	Functions []MirFunction `json:"functions"`
}

// MirFunction represents the JSON structure for MIR functions
type MirFunction struct {
	Name       string     `json:"name"`
	ReturnType string     `json:"return_type"`
	Params     []MirParam `json:"params"`
	Blocks     []MirBlock `json:"blocks"`
}

// MirParam represents the JSON structure for MIR parameters
type MirParam struct {
	Name      string `json:"name"`
	ParamType string `json:"param_type"`
	ID        uint32 `json:"id"`
}

// MirBlock represents the JSON structure for MIR blocks
type MirBlock struct {
	Name         string           `json:"name"`
	Instructions []MirInstruction `json:"instructions"`
	Terminator   MirTerminator    `json:"terminator"`
}

// MirInstruction represents the JSON structure for MIR instructions
type MirInstruction struct {
	ID       uint32       `json:"id"`
	Op       string       `json:"op"`
	InstType string       `json:"inst_type"`
	Operands []MirOperand `json:"operands"`
}

// MirTerminator represents the JSON structure for MIR terminators
type MirTerminator struct {
	Op       string       `json:"op"`
	Operands []MirOperand `json:"operands"`
}

// MirOperand represents the JSON structure for MIR operands
type MirOperand struct {
	Kind        string  `json:"kind"` // "value" or "literal"
	Value       *uint32 `json:"value,omitempty"`
	Literal     *string `json:"literal,omitempty"`
	OperandType string  `json:"operand_type"`
}

// ToJSON converts a MIR module to JSON format
func (m *Module) ToJSON() ([]byte, error) {
	mirModule := MirModule{
		Functions: make([]MirFunction, len(m.Functions)),
	}

	for i, fn := range m.Functions {
		mirFunc := MirFunction{
			Name:       fn.Name,
			ReturnType: fn.ReturnType,
			Params:     make([]MirParam, len(fn.Params)),
			Blocks:     make([]MirBlock, len(fn.Blocks)),
		}

		// Convert parameters
		for j, param := range fn.Params {
			mirFunc.Params[j] = MirParam{
				Name:      param.Name,
				ParamType: param.Type,
				ID:        uint32(param.ID),
			}
		}

		// Convert blocks
		for j, block := range fn.Blocks {
			mirBlock := MirBlock{
				Name:         block.Name,
				Instructions: make([]MirInstruction, len(block.Instructions)),
				Terminator:   MirTerminator{Op: block.Terminator.Op},
			}

			// Convert instructions
			for k, inst := range block.Instructions {
				mirInst := MirInstruction{
					ID:       uint32(inst.ID),
					Op:       inst.Op,
					InstType: inst.Type,
					Operands: make([]MirOperand, len(inst.Operands)),
				}

				// Convert operands
				for l, op := range inst.Operands {
					mirOp := MirOperand{
						Kind:        "literal", // default
						OperandType: op.Type,
					}

					if op.Kind == OperandValue {
						mirOp.Kind = "value"
						val := uint32(op.Value)
						mirOp.Value = &val
					} else if op.Kind == OperandLiteral {
						mirOp.Kind = "literal"
						mirOp.Literal = &op.Literal
					}

					mirInst.Operands[l] = mirOp
				}

				mirBlock.Instructions[k] = mirInst
			}

			// Convert terminator operands
			mirBlock.Terminator.Operands = make([]MirOperand, len(block.Terminator.Operands))
			for k, op := range block.Terminator.Operands {
				mirOp := MirOperand{
					Kind:        "literal", // default
					OperandType: op.Type,
				}

				if op.Kind == OperandValue {
					mirOp.Kind = "value"
					val := uint32(op.Value)
					mirOp.Value = &val
				} else if op.Kind == OperandLiteral {
					mirOp.Kind = "literal"
					mirOp.Literal = &op.Literal
				}

				mirBlock.Terminator.Operands[k] = mirOp
			}

			mirFunc.Blocks[j] = mirBlock
		}

		mirModule.Functions[i] = mirFunc
	}

	return json.Marshal(mirModule)
}
