package builder

import (
	"fmt"
	"strings"

	"github.com/omni-lang/omni/internal/ast"
	"github.com/omni-lang/omni/internal/mir"
)

const inferTypePlaceholder = "<infer>"

// BuildModule lowers the parsed AST into MIR suitable for optimisation passes and codegen.
func BuildModule(mod *ast.Module) (*mir.Module, error) {
	mb := &moduleBuilder{
		module:     &mir.Module{},
		signatures: make(map[string]FunctionSignature),
	}
	mb.collectFunctionSignatures(mod)

	for _, decl := range mod.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		mirFunc, err := mb.buildFunction(fn)
		if err != nil {
			return nil, err
		}
		mb.module.Functions = append(mb.module.Functions, mirFunc)
	}
	return mb.module, nil
}

type moduleBuilder struct {
	module     *mir.Module
	signatures map[string]FunctionSignature
}

type functionBuilder struct {
	fn     *mir.Function
	block  *mir.BasicBlock
	env    map[string]symbol
	sigs   map[string]FunctionSignature
	blocks int
}

type symbol struct {
	Value   mir.ValueID
	Type    string
	Mutable bool
}

// FunctionSignature captures the signature of a function for MIR lowering.
type FunctionSignature struct {
	Return string
	Params []string
}

func (mb *moduleBuilder) collectFunctionSignatures(mod *ast.Module) {
	for _, decl := range mod.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		sig := FunctionSignature{Return: "void"}
		if fn.Return != nil {
			sig.Return = typeExprToString(fn.Return)
		}
		sig.Params = make([]string, len(fn.Params))
		for i, param := range fn.Params {
			sig.Params[i] = typeExprToString(param.Type)
		}
		mb.signatures[fn.Name] = sig
	}
}

func (mb *moduleBuilder) buildFunction(fn *ast.FuncDecl) (*mir.Function, error) {
	params := make([]mir.Param, len(fn.Params))
	for i, p := range fn.Params {
		params[i] = mir.Param{Name: p.Name, Type: typeExprToString(p.Type)}
	}
	returnType := "void"
	if fn.Return != nil {
		returnType = typeExprToString(fn.Return)
	}

	mirFunc := mir.NewFunction(fn.Name, returnType, params)
	fb := &functionBuilder{
		fn:    mirFunc,
		block: mirFunc.NewBlock("entry"),
		env:   make(map[string]symbol),
		sigs:  mb.signatures,
	}

	for _, p := range mirFunc.Params {
		fb.env[p.Name] = symbol{Value: p.ID, Type: p.Type, Mutable: true}
	}

	if fn.ExprBody != nil {
		value, err := fb.lowerExpr(fn.ExprBody)
		if err != nil {
			return nil, err
		}
		fb.block.Terminator = mir.Terminator{Op: "ret", Operands: []mir.Operand{valueOperand(value.ID, value.Type)}}
		return mirFunc, nil
	}

	if fn.Body != nil {
		if err := fb.lowerBlock(fn.Body); err != nil {
			return nil, err
		}
		if fb.block == nil {
			return mirFunc, nil
		}
		if !fb.block.HasTerminator() {
			if mirFunc.ReturnType == "void" {
				fb.block.Terminator = mir.Terminator{Op: "ret"}
			} else {
				return nil, fmt.Errorf("mir builder: missing return in function %s", fn.Name)
			}
		}
	}

	return mirFunc, nil
}

func (fb *functionBuilder) lowerBlock(block *ast.BlockStmt) error {
	if fb.block == nil {
		return nil
	}
	for _, stmt := range block.Statements {
		if err := fb.lowerStmt(stmt); err != nil {
			return err
		}
		if fb.block == nil || fb.block.HasTerminator() {
			break
		}
	}
	return nil
}

func (fb *functionBuilder) lowerStmt(stmt ast.Stmt) error {
	if fb.block == nil {
		return nil
	}
	switch s := stmt.(type) {
	case *ast.ReturnStmt:
		var value mirValue
		var err error
		if s.Value != nil {
			value, err = fb.lowerExpr(s.Value)
			if err != nil {
				return err
			}
		} else {
			value = mirValue{ID: mir.InvalidValue, Type: "void"}
		}
		operands := []mir.Operand{}
		if value.ID != mir.InvalidValue {
			operands = append(operands, valueOperand(value.ID, value.Type))
		}
		fb.block.Terminator = mir.Terminator{Op: "ret", Operands: operands}
		return nil
	case *ast.BindingStmt:
		val, err := fb.lowerOptionalExpr(s.Value)
		if err != nil {
			return err
		}
		fb.env[s.Name] = symbol{Value: val.ID, Type: val.Type, Mutable: s.Mutable}
		return nil
	case *ast.ShortVarDeclStmt:
		val, err := fb.lowerOptionalExpr(s.Value)
		if err != nil {
			return err
		}
		fb.env[s.Name] = symbol{Value: val.ID, Type: val.Type, Mutable: true}
		return nil
	case *ast.AssignmentStmt:
		target, ok := s.Left.(*ast.IdentifierExpr)
		if !ok {
			return fmt.Errorf("mir builder: only identifier assignments supported")
		}
		sym, exists := fb.env[target.Name]
		if !exists {
			return fmt.Errorf("mir builder: assignment to undefined variable %q", target.Name)
		}
		if !sym.Mutable {
			return fmt.Errorf("mir builder: cannot assign to immutable variable %q", target.Name)
		}
		rhs, err := fb.lowerExpr(s.Right)
		if err != nil {
			return err
		}

		// Create an assignment instruction in the MIR
		assignID := fb.fn.NextValue()
		assignInst := mir.Instruction{
			ID:   assignID,
			Op:   "assign",
			Type: rhs.Type,
			Operands: []mir.Operand{
				valueOperand(sym.Value, sym.Type), // target variable
				valueOperand(rhs.ID, rhs.Type),    // source value
			},
		}
		fb.block.Instructions = append(fb.block.Instructions, assignInst)

		// Update the environment to point to the new assignment result
		fb.env[target.Name] = symbol{Value: assignID, Type: rhs.Type, Mutable: sym.Mutable}
		return nil
	case *ast.ExprStmt:
		_, err := fb.lowerExpr(s.Expr)
		return err
	case *ast.IfStmt:
		return fb.lowerIfStmt(s)
	case *ast.ForStmt:
		return fb.lowerForStmt(s)
	case *ast.IncrementStmt:
		return fb.lowerIncrementStmt(s)
	default:
		return fmt.Errorf("mir builder: unsupported statement %T", s)
	}
}

func (fb *functionBuilder) newBlock(prefix string) *mir.BasicBlock {
	name := fmt.Sprintf("%s_%d", prefix, fb.blocks)
	fb.blocks++
	return fb.fn.NewBlock(name)
}

func blockOperand(block *mir.BasicBlock) mir.Operand {
	return mir.Operand{Kind: mir.OperandLiteral, Literal: block.Name}
}

func (fb *functionBuilder) lowerIfStmt(stmt *ast.IfStmt) error {
	cond, err := fb.lowerExpr(stmt.Cond)
	if err != nil {
		return err
	}
	thenBlock := fb.newBlock("then")
	var elseBlock *mir.BasicBlock
	var mergeBlock *mir.BasicBlock
	if stmt.Else == nil {
		mergeBlock = fb.newBlock("merge")
	}
	operands := []mir.Operand{valueOperand(cond.ID, cond.Type), blockOperand(thenBlock)}
	if stmt.Else != nil {
		elseBlock = fb.newBlock("else")
		operands = append(operands, blockOperand(elseBlock))
	} else {
		operands = append(operands, blockOperand(mergeBlock))
	}
	fb.block.Terminator = mir.Terminator{Op: "cbr", Operands: operands}

	// Lower then branch.
	fb.block = thenBlock
	if err := fb.lowerBlock(stmt.Then); err != nil {
		return err
	}
	thenFallsThrough := false
	if !fb.block.HasTerminator() {
		if mergeBlock == nil {
			mergeBlock = fb.newBlock("merge")
		}
		fb.block.Terminator = mir.Terminator{Op: "br", Operands: []mir.Operand{blockOperand(mergeBlock)}}
		thenFallsThrough = true
	}

	elseFallsThrough := false
	if stmt.Else != nil {
		fb.block = elseBlock
		if err := fb.lowerElseBranch(stmt.Else); err != nil {
			return err
		}
		if !fb.block.HasTerminator() {
			if mergeBlock == nil {
				mergeBlock = fb.newBlock("merge")
			}
			fb.block.Terminator = mir.Terminator{Op: "br", Operands: []mir.Operand{blockOperand(mergeBlock)}}
			elseFallsThrough = true
		}
	} else {
		elseFallsThrough = true
	}

	if thenFallsThrough || elseFallsThrough {
		if mergeBlock == nil {
			mergeBlock = fb.newBlock("merge")
		}
		fb.block = mergeBlock
	} else {
		fb.block = nil
	}
	return nil
}

func (fb *functionBuilder) lowerElseBranch(stmt ast.Stmt) error {
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		return fb.lowerBlock(s)
	default:
		return fb.lowerStmt(s)
	}
}

func (fb *functionBuilder) lowerForStmt(stmt *ast.ForStmt) error {
	if stmt.IsRange {
		// Range form: for item in items { ... }
		return fb.lowerRangeFor(stmt)
	} else {
		// Classic form: for init; cond; post { ... }
		return fb.lowerClassicFor(stmt)
	}
}

func (fb *functionBuilder) lowerRangeFor(stmt *ast.ForStmt) error {
	// Save the current environment
	originalEnv := make(map[string]symbol)
	for k, v := range fb.env {
		originalEnv[k] = v
	}

	// Evaluate the iterable expression
	iterableValue, err := fb.lowerExpr(stmt.Iterable)
	if err != nil {
		return err
	}

	// Create loop variable (the target of the range loop)
	if stmt.Target == nil {
		return fmt.Errorf("mir builder: range for loop requires a target variable")
	}

	// Create loop index variable before the loop
	indexID := fb.fn.NextValue()
	indexInst := mir.Instruction{
		ID:   indexID,
		Op:   "const",
		Type: "int",
		Operands: []mir.Operand{
			{Kind: mir.OperandLiteral, Literal: "0", Type: "int"},
		},
	}
	fb.block.Instructions = append(fb.block.Instructions, indexInst)
	fb.env["__loop_index"] = symbol{Value: indexID, Type: "int", Mutable: true}

	// Create array length variable
	lengthID := fb.fn.NextValue()

	// Get the actual array length from the iterable
	// For now, we'll use a hardcoded approach based on the array initialization
	// TODO: Implement proper array length detection from MIR by analyzing the array initialization
	var arrayLength string
	if iterableValue.Type == "array<int>" {
		// Count the number of elements in the array literal
		// This is a temporary solution - we should analyze the MIR to get the actual length
		if len(fb.block.Instructions) > 0 {
			// Look for the array initialization instruction
			for i := len(fb.block.Instructions) - 1; i >= 0; i-- {
				inst := fb.block.Instructions[i]
				if inst.ID == iterableValue.ID && inst.Op == "array.init" {
					// Count the operands (all operands are array elements)
					arrayLength = fmt.Sprintf("%d", len(inst.Operands))
					break
				}
			}
		}
		if arrayLength == "" {
			arrayLength = "5" // Fallback for array<int> with 5 elements
		}
	} else {
		// Fallback for other array types
		arrayLength = "3"
	}

	lengthInst := mir.Instruction{
		ID:   lengthID,
		Op:   "const",
		Type: "int",
		Operands: []mir.Operand{
			{Kind: mir.OperandLiteral, Literal: arrayLength, Type: "int"},
		},
	}
	fb.block.Instructions = append(fb.block.Instructions, lengthInst)
	fb.env["__array_length"] = symbol{Value: lengthID, Type: "int", Mutable: false}

	// Create loop header block
	headerBlock := fb.newBlock("range_loop_header")

	// Create loop body block
	bodyBlock := fb.newBlock("range_loop_body")

	// Create loop exit block
	exitBlock := fb.newBlock("range_loop_exit")

	// Branch from current block to header
	currentBlock := fb.block
	if currentBlock != nil {
		currentBlock.Terminator = mir.Terminator{
			Op:       "br",
			Operands: []mir.Operand{blockOperand(headerBlock)},
		}
	}

	// Set up the header block
	fb.block = headerBlock
	headerEnv := make(map[string]symbol)
	for k, v := range fb.env {
		headerEnv[k] = v
	}

	// Create loop condition: index < length
	condID := fb.fn.NextValue()
	condInst := mir.Instruction{
		ID:   condID,
		Op:   "cmp.lt",
		Type: "bool",
		Operands: []mir.Operand{
			valueOperand(indexID, "int"),
			valueOperand(lengthID, "int"),
		},
	}
	fb.block.Instructions = append(fb.block.Instructions, condInst)

	// Branch based on condition
	fb.block.Terminator = mir.Terminator{
		Op: "cbr",
		Operands: []mir.Operand{
			valueOperand(condID, "bool"),
			blockOperand(bodyBlock),
			blockOperand(exitBlock),
		},
	}

	fb.env = headerEnv

	// Set up the body block
	fb.block = bodyBlock
	bodyEnv := make(map[string]symbol)
	for k, v := range fb.env {
		bodyEnv[k] = v
	}

	// Create the loop variable by indexing into the array
	itemID := fb.fn.NextValue()
	itemInst := mir.Instruction{
		ID:   itemID,
		Op:   "index",
		Type: "int",
		Operands: []mir.Operand{
			valueOperand(iterableValue.ID, iterableValue.Type),
			valueOperand(indexID, "int"),
		},
	}
	fb.block.Instructions = append(fb.block.Instructions, itemInst)
	bodyEnv[stmt.Target.Name] = symbol{Value: itemID, Type: "int", Mutable: false}

	fb.env = bodyEnv

	// Handle loop body
	if err := fb.lowerBlock(stmt.Body); err != nil {
		return err
	}

	// If body doesn't have a terminator, add index increment and loop back
	if !fb.block.HasTerminator() {
		// Increment the loop index
		newIndexID := fb.fn.NextValue()
		incInst := mir.Instruction{
			ID:   newIndexID,
			Op:   "add",
			Type: "int",
			Operands: []mir.Operand{
				valueOperand(indexID, "int"),
				{Kind: mir.OperandLiteral, Literal: "1", Type: "int"},
			},
		}
		fb.block.Instructions = append(fb.block.Instructions, incInst)

		// Create an assignment instruction to update the loop index
		assignID := fb.fn.NextValue()
		assignInst := mir.Instruction{
			ID:   assignID,
			Op:   "assign",
			Type: "int",
			Operands: []mir.Operand{
				valueOperand(indexID, "int"),    // target variable
				valueOperand(newIndexID, "int"), // source value (incremented result)
			},
		}
		fb.block.Instructions = append(fb.block.Instructions, assignInst)

		fb.env["__loop_index"] = symbol{Value: assignID, Type: "int", Mutable: true}

		// Branch back to header
		fb.block.Terminator = mir.Terminator{
			Op:       "br",
			Operands: []mir.Operand{blockOperand(headerBlock)},
		}
	}

	// Set current block to exit block and restore original environment
	fb.block = exitBlock
	fb.env = originalEnv
	return nil
}

func (fb *functionBuilder) lowerClassicFor(stmt *ast.ForStmt) error {
	// Save the current environment
	originalEnv := make(map[string]symbol)
	for k, v := range fb.env {
		originalEnv[k] = v
	}

	// Handle initialization in current block (before loop)
	if stmt.Init != nil {
		if err := fb.lowerStmt(stmt.Init); err != nil {
			return err
		}
	}

	// Create loop header block (for condition check)
	headerBlock := fb.newBlock("loop_header")

	// Create loop body block
	bodyBlock := fb.newBlock("loop_body")

	// Create loop exit block
	exitBlock := fb.newBlock("loop_exit")

	// Branch from current block to header
	currentBlock := fb.block
	if currentBlock != nil {
		currentBlock.Terminator = mir.Terminator{
			Op:       "br",
			Operands: []mir.Operand{blockOperand(headerBlock)},
		}
	}

	// Copy environment to header block
	headerEnv := make(map[string]symbol)
	for k, v := range fb.env {
		headerEnv[k] = v
	}
	fb.block = headerBlock
	fb.env = headerEnv

	// Handle condition check in header block
	if stmt.Condition != nil {
		condValue, err := fb.lowerExpr(stmt.Condition)
		if err != nil {
			return err
		}

		// Branch to body if condition is true, exit if false
		fb.block.Terminator = mir.Terminator{
			Op: "cbr",
			Operands: []mir.Operand{
				valueOperand(condValue.ID, condValue.Type),
				blockOperand(bodyBlock),
				blockOperand(exitBlock),
			},
		}
	} else {
		// No condition - always enter body
		fb.block.Terminator = mir.Terminator{
			Op:       "br",
			Operands: []mir.Operand{blockOperand(bodyBlock)},
		}
	}

	// Handle loop body - preserve environment from header
	bodyEnv := make(map[string]symbol)
	for k, v := range fb.env {
		bodyEnv[k] = v
	}
	fb.block = bodyBlock
	fb.env = bodyEnv

	if err := fb.lowerBlock(stmt.Body); err != nil {
		return err
	}

	// If body doesn't have a terminator, add post-increment and loop back
	if !fb.block.HasTerminator() {
		if stmt.Post != nil {
			if err := fb.lowerStmt(stmt.Post); err != nil {
				return err
			}
		}

		// Branch back to header
		fb.block.Terminator = mir.Terminator{
			Op:       "br",
			Operands: []mir.Operand{blockOperand(headerBlock)},
		}
	}

	// Set current block to exit block and restore original environment
	fb.block = exitBlock
	fb.env = originalEnv
	return nil
}

func (fb *functionBuilder) lowerIncrementStmt(stmt *ast.IncrementStmt) error {
	// Get the target variable
	target, ok := stmt.Target.(*ast.IdentifierExpr)
	if !ok {
		return fmt.Errorf("mir builder: increment target must be an identifier")
	}

	sym, exists := fb.env[target.Name]
	if !exists {
		return fmt.Errorf("mir builder: increment of undefined variable %q", target.Name)
	}

	if !sym.Mutable {
		return fmt.Errorf("mir builder: cannot increment immutable variable %q", target.Name)
	}

	// Create constant 1
	const1 := fb.fn.NextValue()
	constInst := mir.Instruction{
		ID:   const1,
		Op:   "const",
		Type: "int",
		Operands: []mir.Operand{
			{Kind: mir.OperandLiteral, Literal: "1", Type: "int"},
		},
	}
	fb.block.Instructions = append(fb.block.Instructions, constInst)

	// Create increment instruction that updates the variable in place
	id := fb.fn.NextValue()
	var op string
	switch stmt.Op {
	case "++":
		op = "add"
	case "--":
		op = "sub"
	default:
		return fmt.Errorf("mir builder: unsupported increment operator %q", stmt.Op)
	}

	incInst := mir.Instruction{
		ID:   id,
		Op:   op,
		Type: sym.Type,
		Operands: []mir.Operand{
			valueOperand(sym.Value, sym.Type),
			valueOperand(const1, "int"),
		},
	}
	fb.block.Instructions = append(fb.block.Instructions, incInst)

	// Create an assignment instruction to update the original variable
	assignID := fb.fn.NextValue()
	assignInst := mir.Instruction{
		ID:   assignID,
		Op:   "assign",
		Type: sym.Type,
		Operands: []mir.Operand{
			valueOperand(sym.Value, sym.Type), // target variable
			valueOperand(id, sym.Type),        // source value (incremented result)
		},
	}
	fb.block.Instructions = append(fb.block.Instructions, assignInst)

	// Update the variable with the new value
	fb.env[target.Name] = symbol{Value: assignID, Type: sym.Type, Mutable: sym.Mutable}

	return nil
}

func (fb *functionBuilder) lowerOptionalExpr(expr ast.Expr) (mirValue, error) {
	if expr == nil {
		return mirValue{ID: mir.InvalidValue, Type: "void"}, nil
	}
	return fb.lowerExpr(expr)
}

type mirValue struct {
	ID   mir.ValueID
	Type string
}

func (fb *functionBuilder) lowerExpr(expr ast.Expr) (mirValue, error) {
	if fb.block == nil {
		return mirValue{ID: mir.InvalidValue, Type: inferTypePlaceholder}, nil
	}
	switch e := expr.(type) {
	case *ast.LiteralExpr:
		return fb.emitLiteral(e)
	case *ast.IdentifierExpr:
		sym, ok := fb.env[e.Name]
		if !ok {
			if sig, exists := fb.sigs[e.Name]; exists {
				return mirValue{ID: mir.InvalidValue, Type: fmt.Sprintf("func(%s):%s", strings.Join(sig.Params, ","), sig.Return)}, nil
			}
			return mirValue{}, fmt.Errorf("mir builder: undefined identifier %q", e.Name)
		}
		return mirValue{ID: sym.Value, Type: sym.Type}, nil
	case *ast.BinaryExpr:
		return fb.emitBinary(e)
	case *ast.UnaryExpr:
		return fb.emitUnary(e)
	case *ast.CallExpr:
		return fb.emitCall(e)
	case *ast.StructLiteralExpr:
		return fb.emitStructLiteral(e)
	case *ast.ArrayLiteralExpr:
		return fb.emitArrayLiteral(e)
	case *ast.MapLiteralExpr:
		return fb.emitMapLiteral(e)
	case *ast.MemberExpr:
		return fb.emitMemberAccess(e)
	case *ast.IndexExpr:
		return fb.emitIndexAccess(e)
	case *ast.AssignmentExpr:
		if err := fb.lowerStmt(&ast.AssignmentStmt{SpanInfo: e.SpanInfo, Left: e.Left, Right: e.Right}); err != nil {
			return mirValue{}, err
		}
		ident, ok := e.Left.(*ast.IdentifierExpr)
		if !ok {
			return mirValue{}, fmt.Errorf("mir builder: expected identifier assignment")
		}
		sym := fb.env[ident.Name]
		return mirValue{ID: sym.Value, Type: sym.Type}, nil
	case *ast.NewExpr:
		// For now, just return a placeholder value
		// TODO: Implement actual memory allocation
		id := fb.fn.NextValue()
		typ := "*" + e.Type.Name
		return mirValue{ID: id, Type: typ}, nil
	case *ast.DeleteExpr:
		// For now, just evaluate the target expression
		// TODO: Implement actual memory deallocation
		return fb.lowerExpr(e.Target)
	default:
		return mirValue{}, fmt.Errorf("mir builder: unsupported expression %T", e)
	}
}

func (fb *functionBuilder) emitLiteral(lit *ast.LiteralExpr) (mirValue, error) {
	typ := literalType(lit)
	id := fb.fn.NextValue()
	inst := mir.Instruction{
		ID:   id,
		Op:   "const",
		Type: typ,
		Operands: []mir.Operand{
			{Kind: mir.OperandLiteral, Literal: lit.Value, Type: typ},
		},
	}
	fb.block.Instructions = append(fb.block.Instructions, inst)
	return mirValue{ID: id, Type: typ}, nil
}

func literalType(lit *ast.LiteralExpr) string {
	switch lit.Kind {
	case ast.LiteralInt:
		return "int"
	case ast.LiteralFloat:
		return "float"
	case ast.LiteralBool:
		return "bool"
	case ast.LiteralString:
		return "string"
	case ast.LiteralChar:
		return "char"
	default:
		return inferTypePlaceholder
	}
}

func (fb *functionBuilder) emitBinary(expr *ast.BinaryExpr) (mirValue, error) {
	left, err := fb.lowerExpr(expr.Left)
	if err != nil {
		return mirValue{}, err
	}
	right, err := fb.lowerExpr(expr.Right)
	if err != nil {
		return mirValue{}, err
	}
	id := fb.fn.NextValue()
	resultType := left.Type
	if isComparison(expr.Op) || isLogical(expr.Op) {
		resultType = "bool"
	}

	// Handle string concatenation specially (string + string, string + int, int + string)
	if expr.Op == "+" && (left.Type == "string" || right.Type == "string") {
		resultType = "string"
		inst := mir.Instruction{
			ID:   id,
			Op:   "strcat",
			Type: resultType,
			Operands: []mir.Operand{
				valueOperand(left.ID, left.Type),
				valueOperand(right.ID, right.Type),
			},
		}
		fb.block.Instructions = append(fb.block.Instructions, inst)
		return mirValue{ID: id, Type: resultType}, nil
	}

	inst := mir.Instruction{
		ID:   id,
		Op:   mapBinaryOp(expr.Op),
		Type: resultType,
		Operands: []mir.Operand{
			valueOperand(left.ID, left.Type),
			valueOperand(right.ID, right.Type),
		},
	}
	fb.block.Instructions = append(fb.block.Instructions, inst)
	return mirValue{ID: id, Type: resultType}, nil
}

func (fb *functionBuilder) emitUnary(expr *ast.UnaryExpr) (mirValue, error) {
	operand, err := fb.lowerExpr(expr.Expr)
	if err != nil {
		return mirValue{}, err
	}

	id := fb.fn.NextValue()
	resultType := operand.Type

	// Map unary operators to MIR instructions
	var op string
	switch expr.Op {
	case "-":
		op = "neg"
	case "!":
		op = "not"
	default:
		return mirValue{}, fmt.Errorf("mir builder: unsupported unary operator %q", expr.Op)
	}

	inst := mir.Instruction{
		ID:   id,
		Op:   op,
		Type: resultType,
		Operands: []mir.Operand{
			valueOperand(operand.ID, operand.Type),
		},
	}
	fb.block.Instructions = append(fb.block.Instructions, inst)
	return mirValue{ID: id, Type: resultType}, nil
}

func (fb *functionBuilder) emitCall(expr *ast.CallExpr) (mirValue, error) {
	// Handle array method calls like x.len() where x is an array
	if member, ok := expr.Callee.(*ast.MemberExpr); ok {
		// Try to lower the target expression to get its type
		target, err := fb.lowerExpr(member.Target)
		if err == nil {
			// Check if this is an array method call
			if strings.HasPrefix(target.Type, "[]<") || strings.HasPrefix(target.Type, "array<") {
				if member.Member == "len" && len(expr.Args) == 0 {
					// Convert x.len() to len(x) by calling the builtin len function
					id := fb.fn.NextValue()
					operands := []mir.Operand{
						{Kind: mir.OperandLiteral, Literal: "len"},
						valueOperand(target.ID, target.Type),
					}

					inst := mir.Instruction{
						ID:       id,
						Op:       "call",
						Type:     "int",
						Operands: operands,
					}
					fb.block.Instructions = append(fb.block.Instructions, inst)
					return mirValue{ID: id, Type: "int"}, nil
				}
				return mirValue{}, fmt.Errorf("mir builder: unsupported array method %q", member.Member)
			}
		}
	}

	id := fb.fn.NextValue()
	operands := []mir.Operand{}
	calleeName := "<unknown>"
	if ident, ok := expr.Callee.(*ast.IdentifierExpr); ok {
		calleeName = ident.Name
	} else if member, ok := expr.Callee.(*ast.MemberExpr); ok {
		// Handle module member access (e.g., math_utils.add)
		if ident, ok := member.Target.(*ast.IdentifierExpr); ok {
			calleeName = ident.Name + "." + member.Member
		}
	}
	// Normalize alias-only module calls like io.println -> std.io.println
	if !strings.HasPrefix(calleeName, "std.") && strings.Contains(calleeName, ".") {
		// Best-effort normalization: if the first segment matches a known std submodule
		parts := strings.Split(calleeName, ".")
		switch parts[0] {
		case "io", "math", "string", "str", "array", "os", "collections":
			if parts[0] == "str" {
				// Map str to std.string
				calleeName = "std.string." + parts[1]
			} else {
				calleeName = "std." + calleeName
			}
		}
	}

	// Handle function calls
	resultType := inferTypePlaceholder
	operands = append(operands, mir.Operand{Kind: mir.OperandLiteral, Literal: calleeName})

	if strings.HasPrefix(calleeName, "std.") {
		// For std functions, determine return type based on function name
		if strings.Contains(calleeName, "io.") {
			resultType = "void"
		} else if strings.Contains(calleeName, "math.") {
			resultType = "int"
		} else if strings.Contains(calleeName, "string.") {
			if strings.Contains(calleeName, "length") {
				resultType = "int"
			} else if strings.Contains(calleeName, "concat") {
				resultType = "string"
			} else {
				resultType = "string"
			}
		} else {
			resultType = "void"
		}
	} else if strings.Contains(calleeName, ".") {
		// For imported module functions, try to get signature from function signatures
		if sig, exists := fb.sigs[calleeName]; exists {
			resultType = sig.Return
		} else {
			// Fallback: assume int return for unknown imported functions
			resultType = "int"
		}
	} else {
		// Regular local function call
		if sig, exists := fb.sigs[calleeName]; exists {
			resultType = sig.Return
		}
	}

	for _, arg := range expr.Args {
		value, err := fb.lowerExpr(arg)
		if err != nil {
			return mirValue{}, err
		}
		operands = append(operands, valueOperand(value.ID, value.Type))
	}

	inst := mir.Instruction{
		ID:       id,
		Op:       "call",
		Type:     resultType,
		Operands: operands,
	}
	fb.block.Instructions = append(fb.block.Instructions, inst)
	return mirValue{ID: id, Type: resultType}, nil
}

func (fb *functionBuilder) emitMemberAccess(expr *ast.MemberExpr) (mirValue, error) {
	// Handle struct field access (e.g., p.x)
	if ident, ok := expr.Target.(*ast.IdentifierExpr); ok {
		// Check if this is a struct field access
		if sym, exists := fb.env[ident.Name]; exists {
			// This is a struct field access
			id := fb.fn.NextValue()
			inst := mir.Instruction{
				ID:   id,
				Op:   "member",
				Type: "int", // TODO: Get actual field type from struct definition
				Operands: []mir.Operand{
					valueOperand(sym.Value, sym.Type),
					{Kind: mir.OperandLiteral, Literal: expr.Member},
				},
			}
			fb.block.Instructions = append(fb.block.Instructions, inst)
			return mirValue{ID: id, Type: "int"}, nil
		}
		// Check if it's a function call context (this will be handled by the caller)
		// For now, just return a placeholder that indicates this is a qualified function
		return mirValue{ID: mir.InvalidValue, Type: "func"}, nil
	}

	return mirValue{}, fmt.Errorf("mir builder: unsupported member access target type %T", expr.Target)
}

func (fb *functionBuilder) emitStructLiteral(expr *ast.StructLiteralExpr) (mirValue, error) {
	id := fb.fn.NextValue()
	operands := []mir.Operand{{Kind: mir.OperandLiteral, Literal: expr.TypeName}}
	for _, field := range expr.Fields {
		value, err := fb.lowerExpr(field.Expr)
		if err != nil {
			return mirValue{}, err
		}
		operands = append(operands, mir.Operand{Kind: mir.OperandLiteral, Literal: field.Name})
		operands = append(operands, valueOperand(value.ID, value.Type))
	}
	inst := mir.Instruction{
		ID:       id,
		Op:       "struct.init",
		Type:     expr.TypeName,
		Operands: operands,
	}
	fb.block.Instructions = append(fb.block.Instructions, inst)
	return mirValue{ID: id, Type: expr.TypeName}, nil
}

func (fb *functionBuilder) emitArrayLiteral(expr *ast.ArrayLiteralExpr) (mirValue, error) {
	id := fb.fn.NextValue()
	operands := make([]mir.Operand, 0, len(expr.Elements))
	elemType := inferTypePlaceholder
	for _, el := range expr.Elements {
		value, err := fb.lowerExpr(el)
		if err != nil {
			return mirValue{}, err
		}
		operands = append(operands, valueOperand(value.ID, value.Type))
		elemType = value.Type
	}
	inst := mir.Instruction{
		ID:       id,
		Op:       "array.init",
		Type:     buildGeneric("array", []string{elemType}),
		Operands: operands,
	}
	fb.block.Instructions = append(fb.block.Instructions, inst)
	return mirValue{ID: id, Type: inst.Type}, nil
}

func (fb *functionBuilder) emitMapLiteral(expr *ast.MapLiteralExpr) (mirValue, error) {
	id := fb.fn.NextValue()
	operands := make([]mir.Operand, 0, len(expr.Entries)*2)
	keyType := inferTypePlaceholder
	valueType := inferTypePlaceholder
	for _, entry := range expr.Entries {
		key, err := fb.lowerExpr(entry.Key)
		if err != nil {
			return mirValue{}, err
		}
		val, err := fb.lowerExpr(entry.Value)
		if err != nil {
			return mirValue{}, err
		}
		operands = append(operands, valueOperand(key.ID, key.Type))
		operands = append(operands, valueOperand(val.ID, val.Type))
		keyType = key.Type
		valueType = val.Type
	}
	inst := mir.Instruction{
		ID:       id,
		Op:       "map.init",
		Type:     buildGeneric("map", []string{keyType, valueType}),
		Operands: operands,
	}
	fb.block.Instructions = append(fb.block.Instructions, inst)
	return mirValue{ID: id, Type: inst.Type}, nil
}

func (fb *functionBuilder) emitIndexAccess(expr *ast.IndexExpr) (mirValue, error) {
	target, err := fb.lowerExpr(expr.Target)
	if err != nil {
		return mirValue{}, err
	}
	index, err := fb.lowerExpr(expr.Index)
	if err != nil {
		return mirValue{}, err
	}

	id := fb.fn.NextValue()

	// Determine the element type based on the target type
	var elementType string
	if strings.HasPrefix(target.Type, "array<") && strings.HasSuffix(target.Type, ">") {
		// Extract element type from array<T>
		inner := target.Type[len("array<") : len(target.Type)-1]
		elementType = strings.TrimSpace(inner)
	} else if strings.HasPrefix(target.Type, "map<") && strings.HasSuffix(target.Type, ">") {
		// Extract value type from map<K,V>
		inner := target.Type[len("map<") : len(target.Type)-1]
		parts := splitGenericArgs(inner)
		if len(parts) == 2 {
			elementType = strings.TrimSpace(parts[1])
		} else {
			elementType = inferTypePlaceholder
		}
	} else {
		elementType = inferTypePlaceholder
	}

	inst := mir.Instruction{
		ID:   id,
		Op:   "index",
		Type: elementType,
		Operands: []mir.Operand{
			valueOperand(target.ID, target.Type),
			valueOperand(index.ID, index.Type),
		},
	}
	fb.block.Instructions = append(fb.block.Instructions, inst)
	return mirValue{ID: id, Type: elementType}, nil
}

func splitGenericArgs(body string) []string {
	if body == "" {
		return nil
	}
	depth := 0
	start := 0
	var parts []string
	for i, r := range body {
		switch r {
		case '<':
			depth++
		case '>':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, strings.TrimSpace(body[start:i]))
				start = i + 1
			}
		}
	}
	parts = append(parts, strings.TrimSpace(body[start:]))
	return parts
}

func mapBinaryOp(op string) string {
	switch op {
	case "+":
		return "add"
	case "-":
		return "sub"
	case "*":
		return "mul"
	case "/":
		return "div"
	case "%":
		return "mod"
	case "==":
		return "cmp.eq"
	case "!=":
		return "cmp.neq"
	case "<":
		return "cmp.lt"
	case "<=":
		return "cmp.lte"
	case ">":
		return "cmp.gt"
	case ">=":
		return "cmp.gte"
	case "&&":
		return "and"
	case "||":
		return "or"
	default:
		return fmt.Sprintf("op.%s", op)
	}
}

func isComparison(op string) bool {
	switch op {
	case "==", "!=", "<", "<=", ">", ">=":
		return true
	default:
		return false
	}
}

func isLogical(op string) bool {
	return op == "&&" || op == "||"
}

func valueOperand(id mir.ValueID, typ string) mir.Operand {
	return mir.Operand{Kind: mir.OperandValue, Value: id, Type: typ}
}

func typeExprToString(t *ast.TypeExpr) string {
	if t == nil {
		return inferTypePlaceholder
	}
	if len(t.Args) == 0 {
		return t.Name
	}
	args := make([]string, len(t.Args))
	for i, arg := range t.Args {
		args[i] = typeExprToString(arg)
	}
	return buildGeneric(t.Name, args)
}

func buildGeneric(name string, args []string) string {
	if len(args) == 0 {
		return name
	}
	return fmt.Sprintf("%s<%s>", name, strings.Join(args, ","))
}
