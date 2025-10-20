package mir

import "fmt"

// ValueID uniquely identifies an SSA value produced within a function.
type ValueID int

// InvalidValue is used when an instruction does not define a result.
const InvalidValue ValueID = -1

// String renders the SSA value name (e.g. %0).
func (v ValueID) String() string {
	if v == InvalidValue {
		return "<invalid>"
	}
	return fmt.Sprintf("%%%d", int(v))
}

// Module contains the MIR for an OmniLang compilation unit.
type Module struct {
	Functions []*Function
}

// Function represents a lowered function in SSA form.
type Function struct {
	Name       string
	ReturnType string
	Params     []Param
	Blocks     []*BasicBlock

	nextValue ValueID
}

// Param captures a function parameter and its SSA value.
type Param struct {
	Name string
	Type string
	ID   ValueID
}

// BasicBlock groups a sequence of instructions terminated by a control-flow edge.
type BasicBlock struct {
	Name         string
	Instructions []Instruction
	Terminator   Terminator
}

// Instruction captures a single SSA instruction.
type Instruction struct {
	ID       ValueID
	Op       string
	Type     string
	Operands []Operand
}

// Terminator marks the end of a basic block.
type Terminator struct {
	Op       string
	Operands []Operand
}

// OperandKind distinguishes literal and SSA value operands.
type OperandKind int

const (
	OperandValue OperandKind = iota
	OperandLiteral
)

// Operand references either another SSA value or a literal constant.
type Operand struct {
	Kind    OperandKind
	Value   ValueID
	Literal string
	Type    string
}

// NewFunction constructs an empty function with the provided signature.
func NewFunction(name, returnType string, params []Param) *Function {
	fn := &Function{Name: name, ReturnType: returnType, Params: params, Blocks: []*BasicBlock{}}
	var next ValueID
	for i := range params {
		params[i].ID = next
		next++
	}
	fn.nextValue = next
	fn.Params = params
	return fn
}

// NewBlock appends a block to the function.
func (f *Function) NewBlock(name string) *BasicBlock {
	block := &BasicBlock{Name: name}
	f.Blocks = append(f.Blocks, block)
	return block
}

// NextValue allocates a fresh SSA value identifier for the function.
func (f *Function) NextValue() ValueID {
	id := f.nextValue
	f.nextValue++
	return id
}

// HasTerminator reports whether the basic block already has a terminator.
func (b *BasicBlock) HasTerminator() bool {
	return b.Terminator.Op != ""
}
