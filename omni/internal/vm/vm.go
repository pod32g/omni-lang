package vm

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/omni-lang/omni/internal/mir"
)

// instructionHandler defines the signature for instruction execution functions
type instructionHandler func(map[string]*mir.Function, *frame, mir.Instruction) (Result, error)

// instructionHandlers maps instruction names to their execution functions
var instructionHandlers map[string]instructionHandler

func init() {
	instructionHandlers = map[string]instructionHandler{
		"const":       execConst,
		"add":         execArithmetic,
		"sub":         execArithmetic,
		"mul":         execArithmetic,
		"div":         execArithmetic,
		"mod":         execArithmetic,
		"strcat":      execStringConcat,
		"neg":         execUnary,
		"not":         execUnary,
		"cmp.eq":      execComparison,
		"cmp.neq":     execComparison,
		"cmp.lt":      execComparison,
		"cmp.lte":     execComparison,
		"cmp.gt":      execComparison,
		"cmp.gte":     execComparison,
		"and":         execLogical,
		"or":          execLogical,
		"call":        execCall,
		"struct.init": execStructInit,
		"array.init":  execArrayInit,
		"index":       execIndex,
		"assign":      execAssign,
		"map.init":    execMapInit,
		"member":      execMember,
		"phi":         execPhi,
	}
}

const inferTypePlaceholder = "<infer>"

// Result captures the outcome of executing the entry function.
type Result struct {
	Type  string
	Value interface{}
}

// Execute interprets the MIR module starting from the named entry function.
func Execute(mod *mir.Module, entry string) (Result, error) {
	if mod == nil {
		return Result{}, fmt.Errorf("vm: nil module")
	}
	funcs := map[string]*mir.Function{}
	for _, fn := range mod.Functions {
		funcs[fn.Name] = fn
	}
	fn, ok := funcs[entry]
	if !ok {
		return Result{}, fmt.Errorf("vm: entry function %q not found", entry)
	}
	value, err := execFunction(funcs, fn, nil)
	if err != nil {
		return Result{}, err
	}
	return value, nil
}

type frame struct {
	values map[mir.ValueID]Result
}

func execFunction(funcs map[string]*mir.Function, fn *mir.Function, args []Result) (Result, error) {
	fr := &frame{values: make(map[mir.ValueID]Result)}
	if len(args) != 0 && len(args) != len(fn.Params) {
		return Result{}, fmt.Errorf("vm: function %s expects %d arguments, got %d", fn.Name, len(fn.Params), len(args))
	}
	for i, param := range fn.Params {
		var val Result
		if args != nil {
			val = args[i]
		} else {
			val = Result{Type: param.Type, Value: nil}
		}
		if val.Type == "" {
			val.Type = param.Type
		}
		fr.values[param.ID] = val
	}

	if len(fn.Blocks) == 0 {
		return Result{Type: "void"}, nil
	}
	blockMap := make(map[string]*mir.BasicBlock, len(fn.Blocks))
	for _, b := range fn.Blocks {
		blockMap[b.Name] = b
	}
	current := fn.Blocks[0]
	for {
		for _, inst := range current.Instructions {
			res, err := execInstruction(funcs, fr, inst)
			if err != nil {
				return Result{}, fmt.Errorf("vm: %s: %w", fn.Name, err)
			}
			if inst.ID != mir.InvalidValue {
				fr.values[inst.ID] = res
			}
		}

		term := current.Terminator
		switch term.Op {
		case "ret":
			if len(term.Operands) == 0 {
				return Result{Type: "void"}, nil
			}
			op := term.Operands[0]
			if op.Kind != mir.OperandValue {
				return literalResult(op)
			}
			return fr.values[op.Value], nil
		case "br":
			target, err := blockByOperand(blockMap, term.Operands[0])
			if err != nil {
				return Result{}, fmt.Errorf("vm: %s: %w", fn.Name, err)
			}
			current = target
		case "cbr":
			if len(term.Operands) < 3 {
				return Result{}, fmt.Errorf("vm: %s: conditional branch requires condition and two targets", fn.Name)
			}
			cond := operandValue(fr, term.Operands[0])
			b, err := toBool(cond)
			if err != nil {
				return Result{}, fmt.Errorf("vm: %s: %w", fn.Name, err)
			}
			var targetOp mir.Operand
			if b {
				targetOp = term.Operands[1]
			} else {
				targetOp = term.Operands[2]
			}
			target, err := blockByOperand(blockMap, targetOp)
			if err != nil {
				return Result{}, fmt.Errorf("vm: %s: %w", fn.Name, err)
			}
			current = target
		default:
			return Result{}, fmt.Errorf("unsupported terminator %q", term.Op)
		}
	}
}

func blockByOperand(blocks map[string]*mir.BasicBlock, op mir.Operand) (*mir.BasicBlock, error) {
	if op.Kind != mir.OperandLiteral {
		return nil, fmt.Errorf("branch target must be literal block name")
	}
	block, ok := blocks[op.Literal]
	if !ok {
		return nil, fmt.Errorf("branch target %q not found", op.Literal)
	}
	return block, nil
}

func execInstruction(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	handler, exists := instructionHandlers[inst.Op]
	if !exists {
		return Result{}, fmt.Errorf("unsupported instruction %q", inst.Op)
	}
	return handler(funcs, fr, inst)
}

// execConst handles const instructions
func execConst(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	return literalResult(inst.Operands[0])
}

// execStructInit handles struct initialization
func execStructInit(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	// Create a struct value from the operands
	// Operands format: [typeName, field1Name, field1Value, field2Name, field2Value, ...]
	if len(inst.Operands) < 3 || len(inst.Operands)%2 != 1 {
		return Result{}, fmt.Errorf("struct.init: invalid operand count, expected odd number >= 3")
	}

	structValue := make(map[string]interface{})

	// Process operands in pairs (field name, field value)
	for i := 1; i < len(inst.Operands); i += 2 {
		if i+1 >= len(inst.Operands) {
			return Result{}, fmt.Errorf("struct.init: odd number of field operands")
		}

		fieldNameOp := inst.Operands[i]
		fieldValueOp := inst.Operands[i+1]

		if fieldNameOp.Kind != mir.OperandLiteral {
			return Result{}, fmt.Errorf("struct.init: field name must be literal")
		}

		fieldValue := operandValue(fr, fieldValueOp)
		structValue[fieldNameOp.Literal] = fieldValue.Value
	}

	return Result{Type: inst.Type, Value: structValue}, nil
}

// execMember handles struct field access
func execMember(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 2 {
		return Result{}, fmt.Errorf("member: expected 2 operands, got %d", len(inst.Operands))
	}

	target := operandValue(fr, inst.Operands[0])
	fieldNameOp := inst.Operands[1]

	if fieldNameOp.Kind != mir.OperandLiteral {
		return Result{}, fmt.Errorf("member: field name must be literal")
	}

	fieldName := fieldNameOp.Literal

	// Handle struct field access
	if structValue, ok := target.Value.(map[string]interface{}); ok {
		if fieldValue, exists := structValue[fieldName]; exists {
			// Determine the field type based on the value
			var fieldType string
			switch fieldValue.(type) {
			case int:
				fieldType = "int"
			case string:
				fieldType = "string"
			case float64:
				fieldType = "float"
			case bool:
				fieldType = "bool"
			default:
				fieldType = "int" // Default fallback
			}
			return Result{Type: fieldType, Value: fieldValue}, nil
		}
		return Result{}, fmt.Errorf("member: field %q not found in struct", fieldName)
	}

	return Result{}, fmt.Errorf("member: target is not a struct")
}

// execPhi handles PHI nodes - values that can come from different control flow paths
func execPhi(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	// PHI nodes have operands in pairs: (value, block_name)
	// For now, we'll use a simple approach: return the first value
	// In a more sophisticated implementation, we'd track which block we came from

	if len(inst.Operands) < 2 {
		return Result{}, fmt.Errorf("phi: expected at least 2 operands, got %d", len(inst.Operands))
	}

	if len(inst.Operands)%2 != 0 {
		return Result{}, fmt.Errorf("phi: expected even number of operands (value, block pairs), got %d", len(inst.Operands))
	}

	// For now, we'll return the first value
	// In a real implementation, we'd need to track the incoming control flow
	// and select the appropriate value based on which block we came from
	firstValue := operandValue(fr, inst.Operands[0])
	return firstValue, nil
}

// execArrayInit handles array initialization
func execArrayInit(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	// Extract element type from array type (e.g., "[]<int>" -> "int")
	elementType := inst.Type
	if strings.HasPrefix(elementType, "[]<") && strings.HasSuffix(elementType, ">") {
		elementType = elementType[3 : len(elementType)-1]
	} else if strings.HasPrefix(elementType, "array<") && strings.HasSuffix(elementType, ">") {
		elementType = elementType[6 : len(elementType)-1]
	}

	// Create array from operands
	var arrayValue interface{}
	switch elementType {
	case "int":
		elements := make([]int, len(inst.Operands))
		for i, op := range inst.Operands {
			val := operandValue(fr, op)
			if val.Type != "int" {
				return Result{}, fmt.Errorf("array.init: expected int element, got %s", val.Type)
			}
			elements[i] = val.Value.(int)
		}
		arrayValue = elements
	case "string":
		elements := make([]string, len(inst.Operands))
		for i, op := range inst.Operands {
			val := operandValue(fr, op)
			if val.Type != "string" {
				return Result{}, fmt.Errorf("array.init: expected string element, got %s", val.Type)
			}
			elements[i] = val.Value.(string)
		}
		arrayValue = elements
	case "float", "double":
		elements := make([]float64, len(inst.Operands))
		for i, op := range inst.Operands {
			val := operandValue(fr, op)
			if val.Type != "float" && val.Type != "double" {
				return Result{}, fmt.Errorf("array.init: expected float element, got %s", val.Type)
			}
			elements[i] = val.Value.(float64)
		}
		arrayValue = elements
	case "bool":
		elements := make([]bool, len(inst.Operands))
		for i, op := range inst.Operands {
			val := operandValue(fr, op)
			if val.Type != "bool" {
				return Result{}, fmt.Errorf("array.init: expected bool element, got %s", val.Type)
			}
			elements[i] = val.Value.(bool)
		}
		arrayValue = elements
	default:
		return Result{}, fmt.Errorf("array.init: unsupported element type %s", elementType)
	}

	return Result{Type: inst.Type, Value: arrayValue}, nil
}

// execAssign handles variable assignment
func execAssign(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 2 {
		return Result{}, fmt.Errorf("assign: expected 2 operands, got %d", len(inst.Operands))
	}

	// Get the target variable ID (operand 0) and the source value (operand 1)
	targetOp := inst.Operands[0]
	if targetOp.Kind != mir.OperandValue {
		return Result{}, fmt.Errorf("assign: target must be a value, got %v", targetOp.Kind)
	}

	value := operandValue(fr, inst.Operands[1])

	// Update the target variable
	fr.values[targetOp.Value] = value

	// Also store in inst.ID for the result
	if inst.ID != mir.InvalidValue {
		fr.values[inst.ID] = value
	}

	return value, nil
}

// execIndex handles array/map indexing
func execIndex(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 2 {
		return Result{}, fmt.Errorf("index: expected 2 operands, got %d", len(inst.Operands))
	}

	target := operandValue(fr, inst.Operands[0])
	index := operandValue(fr, inst.Operands[1])

	// Handle map indexing first (no need to convert index to int)
	if strings.HasPrefix(target.Type, "map<") {
		mapValue, ok := target.Value.(map[interface{}]interface{})
		if !ok {
			return Result{}, fmt.Errorf("index: target is not a map")
		}

		// Get the value type from the map type
		_, valueType, err := parseMapTypes(target.Type)
		if err != nil {
			return Result{}, fmt.Errorf("index: invalid map type %s: %v", target.Type, err)
		}

		// Look up the key in the map
		keyValue := index.Value
		if foundValue, exists := mapValue[keyValue]; exists {
			return Result{Type: valueType, Value: foundValue}, nil
		}

		// Key not found - return zero value for the value type
		switch valueType {
		case "int":
			return Result{Type: "int", Value: 0}, nil
		case "string":
			return Result{Type: "string", Value: ""}, nil
		case "float", "double":
			return Result{Type: "float", Value: 0.0}, nil
		case "bool":
			return Result{Type: "bool", Value: false}, nil
		default:
			return Result{}, fmt.Errorf("index: unknown value type %s for map", valueType)
		}
	}

	// For arrays, convert index to int
	indexVal, err := toInt(index)
	if err != nil {
		return Result{}, fmt.Errorf("index: index must be int, got %s", index.Type)
	}

	// Handle array indexing
	if strings.HasPrefix(target.Type, "[]<") || strings.HasPrefix(target.Type, "array<") {
		switch arr := target.Value.(type) {
		case []int:
			if indexVal < 0 || indexVal >= len(arr) {
				return Result{}, fmt.Errorf("index: array index %d out of bounds [0, %d)", indexVal, len(arr))
			}
			return Result{Type: "int", Value: arr[indexVal]}, nil
		case []string:
			if indexVal < 0 || indexVal >= len(arr) {
				return Result{}, fmt.Errorf("index: array index %d out of bounds [0, %d)", indexVal, len(arr))
			}
			return Result{Type: "string", Value: arr[indexVal]}, nil
		case []float64:
			if indexVal < 0 || indexVal >= len(arr) {
				return Result{}, fmt.Errorf("index: array index %d out of bounds [0, %d)", indexVal, len(arr))
			}
			return Result{Type: "float", Value: arr[indexVal]}, nil
		case []bool:
			if indexVal < 0 || indexVal >= len(arr) {
				return Result{}, fmt.Errorf("index: array index %d out of bounds [0, %d)", indexVal, len(arr))
			}
			return Result{Type: "bool", Value: arr[indexVal]}, nil
		default:
			return Result{}, fmt.Errorf("index: unsupported array type %T", target.Value)
		}
	}

	return Result{}, fmt.Errorf("index: target type %s does not support indexing", target.Type)
}

// execMapInit handles map initialization
func execMapInit(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	// Extract key and value types from map type (e.g., "map<string,int>" -> "string", "int")
	keyType, valueType, err := parseMapTypes(inst.Type)
	if err != nil {
		return Result{}, fmt.Errorf("map.init: invalid map type %s: %v", inst.Type, err)
	}

	// Create a map from the operands (key-value pairs)
	mapValue := make(map[interface{}]interface{})

	// Process operands in pairs (key, value, key, value, ...)
	for i := 0; i < len(inst.Operands); i += 2 {
		if i+1 >= len(inst.Operands) {
			return Result{}, fmt.Errorf("map.init: odd number of operands, expected key-value pairs")
		}

		key := operandValue(fr, inst.Operands[i])
		value := operandValue(fr, inst.Operands[i+1])

		// Validate key and value types
		if key.Type != keyType {
			return Result{}, fmt.Errorf("map.init: key type mismatch, expected %s, got %s", keyType, key.Type)
		}
		if value.Type != valueType {
			return Result{}, fmt.Errorf("map.init: value type mismatch, expected %s, got %s", valueType, value.Type)
		}

		// Store in map
		mapValue[key.Value] = value.Value
	}

	return Result{Type: inst.Type, Value: mapValue}, nil
}

// parseMapTypes extracts key and value types from a map type string
func parseMapTypes(mapType string) (string, string, error) {
	if !strings.HasPrefix(mapType, "map<") || !strings.HasSuffix(mapType, ">") {
		return "", "", fmt.Errorf("invalid map type format")
	}

	inner := mapType[4 : len(mapType)-1] // Remove "map<" and ">"
	parts := strings.Split(inner, ",")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("map type must have exactly 2 type parameters")
	}

	keyType := strings.TrimSpace(parts[0])
	valueType := strings.TrimSpace(parts[1])

	return keyType, valueType, nil
}

func execArithmetic(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	left := operandValue(fr, inst.Operands[0])
	right := operandValue(fr, inst.Operands[1])

	// Handle float arithmetic
	if left.Type == "float" || right.Type == "float" || left.Type == "double" || right.Type == "double" {
		lf, err := toFloat(left)
		if err != nil {
			return Result{}, err
		}
		rf, err := toFloat(right)
		if err != nil {
			return Result{}, err
		}

		var res float64
		switch inst.Op {
		case "add":
			res = lf + rf
		case "sub":
			res = lf - rf
		case "mul":
			res = lf * rf
		case "div":
			if rf == 0 {
				return Result{}, fmt.Errorf("division by zero")
			}
			res = lf / rf
		case "mod":
			// For floats, use math.Mod
			if rf == 0 {
				return Result{}, fmt.Errorf("modulo by zero")
			}
			res = float64(int(lf) % int(rf)) // Simple modulo for floats
		default:
			return Result{}, fmt.Errorf("vm: unsupported float arithmetic operator %s", inst.Op)
		}
		return Result{Type: "float", Value: res}, nil
	}

	// Handle integer arithmetic (existing logic)
	li, err := toInt(left)
	if err != nil {
		return Result{}, err
	}
	ri, err := toInt(right)
	if err != nil {
		return Result{}, err
	}
	var res int
	switch inst.Op {
	case "add":
		res = li + ri
	case "sub":
		res = li - ri
	case "mul":
		res = li * ri
	case "div":
		if ri == 0 {
			return Result{}, fmt.Errorf("division by zero")
		}
		res = li / ri
	case "mod":
		if ri == 0 {
			return Result{}, fmt.Errorf("modulo by zero")
		}
		res = li % ri
	}
	return Result{Type: chooseType(inst.Type, left.Type), Value: res}, nil
}

func execComparison(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	left := operandValue(fr, inst.Operands[0])
	right := operandValue(fr, inst.Operands[1])

	// Handle different types of comparisons
	var res bool

	// String comparisons
	if left.Type == "string" && right.Type == "string" {
		leftStr, ok1 := left.Value.(string)
		rightStr, ok2 := right.Value.(string)
		if !ok1 || !ok2 {
			return Result{}, fmt.Errorf("vm: string comparison failed - invalid string values")
		}

		switch inst.Op {
		case "cmp.eq":
			res = leftStr == rightStr
		case "cmp.neq":
			res = leftStr != rightStr
		case "cmp.lt":
			res = leftStr < rightStr
		case "cmp.lte":
			res = leftStr <= rightStr
		case "cmp.gt":
			res = leftStr > rightStr
		case "cmp.gte":
			res = leftStr >= rightStr
		default:
			return Result{}, fmt.Errorf("vm: unsupported string comparison operator %s", inst.Op)
		}
		return Result{Type: "bool", Value: res}, nil
	}

	// Float comparisons
	if left.Type == "float" || right.Type == "float" || left.Type == "double" || right.Type == "double" {
		leftFloat, err := toFloat(left)
		if err != nil {
			return Result{}, err
		}
		rightFloat, err := toFloat(right)
		if err != nil {
			return Result{}, err
		}

		switch inst.Op {
		case "cmp.eq":
			res = leftFloat == rightFloat
		case "cmp.neq":
			res = leftFloat != rightFloat
		case "cmp.lt":
			res = leftFloat < rightFloat
		case "cmp.lte":
			res = leftFloat <= rightFloat
		case "cmp.gt":
			res = leftFloat > rightFloat
		case "cmp.gte":
			res = leftFloat >= rightFloat
		default:
			return Result{}, fmt.Errorf("vm: unsupported float comparison operator %s", inst.Op)
		}
		return Result{Type: "bool", Value: res}, nil
	}

	// Boolean comparisons
	if left.Type == "bool" && right.Type == "bool" {
		leftBool, ok1 := left.Value.(bool)
		rightBool, ok2 := right.Value.(bool)
		if !ok1 || !ok2 {
			return Result{}, fmt.Errorf("vm: boolean comparison failed - invalid boolean values")
		}

		switch inst.Op {
		case "cmp.eq":
			res = leftBool == rightBool
		case "cmp.neq":
			res = leftBool != rightBool
		default:
			return Result{}, fmt.Errorf("vm: unsupported boolean comparison operator %s", inst.Op)
		}
		return Result{Type: "bool", Value: res}, nil
	}

	// Integer comparisons (existing logic)
	li, err := toInt(left)
	if err != nil {
		return Result{}, err
	}
	ri, err := toInt(right)
	if err != nil {
		return Result{}, err
	}

	switch inst.Op {
	case "cmp.eq":
		res = li == ri
	case "cmp.neq":
		res = li != ri
	case "cmp.lt":
		res = li < ri
	case "cmp.lte":
		res = li <= ri
	case "cmp.gt":
		res = li > ri
	case "cmp.gte":
		res = li >= ri
	}
	return Result{Type: "bool", Value: res}, nil
}

func execLogical(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	left := operandValue(fr, inst.Operands[0])
	right := operandValue(fr, inst.Operands[1])
	lb, err := toBool(left)
	if err != nil {
		return Result{}, err
	}
	rb, err := toBool(right)
	if err != nil {
		return Result{}, err
	}
	var res bool
	if inst.Op == "and" {
		res = lb && rb
	} else {
		res = lb || rb
	}
	return Result{Type: "bool", Value: res}, nil
}

func execStringConcat(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	left := operandValue(fr, inst.Operands[0])
	right := operandValue(fr, inst.Operands[1])

	// Convert operands to strings
	leftStr, err := toString(left)
	if err != nil {
		return Result{}, fmt.Errorf("strcat: left operand: %w", err)
	}
	rightStr, err := toString(right)
	if err != nil {
		return Result{}, fmt.Errorf("strcat: right operand: %w", err)
	}

	// Use pooled strings.Builder for efficient concatenation
	builder := GetStringBuilder()
	defer PutStringBuilder(builder)

	builder.Grow(len(leftStr) + len(rightStr)) // Pre-allocate capacity
	builder.WriteString(leftStr)
	builder.WriteString(rightStr)

	return Result{Type: "string", Value: builder.String()}, nil
}

func execUnary(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	operand := operandValue(fr, inst.Operands[0])

	switch inst.Op {
	case "neg":
		if operand.Type == "int" {
			val := operand.Value.(int)
			return Result{Type: "int", Value: -val}, nil
		}
		return Result{}, fmt.Errorf("neg: operand must be int, got %s", operand.Type)
	case "not":
		if operand.Type == "bool" {
			val := operand.Value.(bool)
			return Result{Type: "bool", Value: !val}, nil
		}
		return Result{}, fmt.Errorf("not: operand must be bool, got %s", operand.Type)
	default:
		return Result{}, fmt.Errorf("unsupported unary operation %q", inst.Op)
	}
}

func execCall(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) == 0 {
		return Result{}, fmt.Errorf("call instruction missing callee operand")
	}
	calleeOp := inst.Operands[0]
	if calleeOp.Kind != mir.OperandLiteral {
		return Result{}, fmt.Errorf("call expects literal callee operand")
	}
	callee := calleeOp.Literal

	// Check if it's an intrinsic function
	if result, handled := execIntrinsic(callee, inst.Operands[1:], fr); handled {
		return result, nil
	}

	fn, ok := funcs[callee]
	if !ok {
		// Check if it's an imported module function (contains a dot but not std.*)
		if strings.Contains(callee, ".") && !strings.HasPrefix(callee, "std.") {
			// For now, return a default value for imported functions
			// In a real implementation, this would need to load and execute the imported module
			return Result{Type: "int", Value: 0}, nil
		}
		return Result{}, fmt.Errorf("callee %q not found", callee)
	}
	args := make([]Result, 0, len(fn.Params))
	for _, op := range inst.Operands[1:] {
		args = append(args, operandValue(fr, op))
	}
	return execFunction(funcs, fn, args)
}

// execIntrinsic handles intrinsic function calls like std.io.println.
// Returns (result, handled) where handled indicates if the function was an intrinsic.
func execIntrinsic(callee string, operands []mir.Operand, fr *frame) (Result, bool) {
	switch callee {
	case "len":
		// Builtin len function for arrays
		if len(operands) == 1 {
			arg := operandValue(fr, operands[0])
			// Check if it's an array type
			if strings.HasPrefix(arg.Type, "[]<") || strings.HasPrefix(arg.Type, "array<") {
				switch arr := arg.Value.(type) {
				case []int:
					return Result{Type: "int", Value: len(arr)}, true
				case []string:
					return Result{Type: "int", Value: len(arr)}, true
				case []float64:
					return Result{Type: "int", Value: len(arr)}, true
				case []bool:
					return Result{Type: "int", Value: len(arr)}, true
				default:
					return Result{}, false
				}
			}
		}
		return Result{}, false
	case "std.io.print":
		if len(operands) == 1 {
			arg := operandValue(fr, operands[0])
			fmt.Print(arg.Value)
			return Result{Type: "void", Value: nil}, true
		}
	case "std.io.println":
		if len(operands) == 1 {
			arg := operandValue(fr, operands[0])
			fmt.Println(arg.Value)
			return Result{Type: "void", Value: nil}, true
		} else if len(operands) == 0 {
			fmt.Println()
			return Result{Type: "void", Value: nil}, true
		}
	case "std.io.print_int":
		if len(operands) == 1 {
			arg := operandValue(fr, operands[0])
			fmt.Print(arg.Value)
			return Result{Type: "void", Value: nil}, true
		}
	case "std.io.println_int":
		if len(operands) == 1 {
			arg := operandValue(fr, operands[0])
			fmt.Println(arg.Value)
			return Result{Type: "void", Value: nil}, true
		}
	case "std.io.print_float":
		if len(operands) == 1 {
			arg := operandValue(fr, operands[0])
			fmt.Print(arg.Value)
			return Result{Type: "void", Value: nil}, true
		}
	case "std.io.println_float":
		if len(operands) == 1 {
			arg := operandValue(fr, operands[0])
			fmt.Println(arg.Value)
			return Result{Type: "void", Value: nil}, true
		}
	case "std.io.print_bool":
		if len(operands) == 1 {
			arg := operandValue(fr, operands[0])
			fmt.Print(arg.Value)
			return Result{Type: "void", Value: nil}, true
		}
	case "std.io.println_bool":
		if len(operands) == 1 {
			arg := operandValue(fr, operands[0])
			fmt.Println(arg.Value)
			return Result{Type: "void", Value: nil}, true
		}
	case "std.math.max":
		if len(operands) == 2 {
			left := operandValue(fr, operands[0])
			right := operandValue(fr, operands[1])
			if left.Type == "int" && right.Type == "int" {
				leftVal := left.Value.(int)
				rightVal := right.Value.(int)
				if leftVal > rightVal {
					return Result{Type: "int", Value: leftVal}, true
				}
				return Result{Type: "int", Value: rightVal}, true
			}
		}
	case "std.math.min":
		if len(operands) == 2 {
			left := operandValue(fr, operands[0])
			right := operandValue(fr, operands[1])
			if left.Type == "int" && right.Type == "int" {
				leftVal := left.Value.(int)
				rightVal := right.Value.(int)
				if leftVal < rightVal {
					return Result{Type: "int", Value: leftVal}, true
				}
				return Result{Type: "int", Value: rightVal}, true
			}
		}
	case "std.math.abs":
		if len(operands) == 1 {
			arg := operandValue(fr, operands[0])
			if arg.Type == "int" {
				val := arg.Value.(int)
				if val < 0 {
					val = -val
				}
				return Result{Type: "int", Value: val}, true
			}
		}
	case "std.math.pow":
		if len(operands) == 2 {
			base := operandValue(fr, operands[0])
			exp := operandValue(fr, operands[1])
			if base.Type == "int" && exp.Type == "int" {
				baseVal := base.Value.(int)
				expVal := exp.Value.(int)
				result := 1
				for i := 0; i < expVal; i++ {
					result *= baseVal
				}
				return Result{Type: "int", Value: result}, true
			}
		}
	case "std.math.gcd":
		if len(operands) == 2 {
			a := operandValue(fr, operands[0])
			b := operandValue(fr, operands[1])
			if a.Type == "int" && b.Type == "int" {
				aVal := a.Value.(int)
				bVal := b.Value.(int)
				if aVal < 0 {
					aVal = -aVal
				}
				if bVal < 0 {
					bVal = -bVal
				}
				for bVal != 0 {
					aVal, bVal = bVal, aVal%bVal
				}
				return Result{Type: "int", Value: aVal}, true
			}
		}
	case "std.math.lcm":
		if len(operands) == 2 {
			a := operandValue(fr, operands[0])
			b := operandValue(fr, operands[1])
			if a.Type == "int" && b.Type == "int" {
				aVal := a.Value.(int)
				bVal := b.Value.(int)
				if aVal < 0 {
					aVal = -aVal
				}
				if bVal < 0 {
					bVal = -bVal
				}
				gcd := aVal
				temp := bVal
				for temp != 0 {
					gcd, temp = temp, gcd%temp
				}
				if gcd == 0 {
					return Result{Type: "int", Value: 0}, true
				}
				lcm := (aVal * bVal) / gcd
				return Result{Type: "int", Value: lcm}, true
			}
		}
	case "std.math.factorial":
		if len(operands) == 1 {
			n := operandValue(fr, operands[0])
			if n.Type == "int" {
				nVal := n.Value.(int)
				if nVal < 0 {
					return Result{Type: "int", Value: 0}, true
				}
				result := 1
				for i := 1; i <= nVal; i++ {
					result *= i
				}
				return Result{Type: "int", Value: result}, true
			}
		}
	case "std.math.sqrt":
		if len(operands) == 1 {
			x := operandValue(fr, operands[0])
			if x.Type == "int" {
				xVal := x.Value.(int)
				if xVal < 0 {
					return Result{Type: "int", Value: 0}, true
				}
				if xVal == 0 {
					return Result{Type: "int", Value: 0}, true
				}
				result := 1
				for result*result <= xVal {
					result++
				}
				return Result{Type: "int", Value: result - 1}, true
			}
		}
	case "std.math.is_prime":
		if len(operands) == 1 {
			n := operandValue(fr, operands[0])
			if n.Type == "int" {
				nVal := n.Value.(int)
				if nVal < 2 {
					return Result{Type: "bool", Value: false}, true
				}
				if nVal == 2 {
					return Result{Type: "bool", Value: true}, true
				}
				if nVal%2 == 0 {
					return Result{Type: "bool", Value: false}, true
				}
				for i := 3; i*i <= nVal; i += 2 {
					if nVal%i == 0 {
						return Result{Type: "bool", Value: false}, true
					}
				}
				return Result{Type: "bool", Value: true}, true
			}
		}
	case "std.math.max_float":
		if len(operands) == 2 {
			a := operandValue(fr, operands[0])
			b := operandValue(fr, operands[1])
			if a.Type == "float" && b.Type == "float" {
				aVal := a.Value.(float64)
				bVal := b.Value.(float64)
				if aVal > bVal {
					return Result{Type: "float", Value: aVal}, true
				}
				return Result{Type: "float", Value: bVal}, true
			}
		}
	case "std.math.min_float":
		if len(operands) == 2 {
			a := operandValue(fr, operands[0])
			b := operandValue(fr, operands[1])
			if a.Type == "float" && b.Type == "float" {
				aVal := a.Value.(float64)
				bVal := b.Value.(float64)
				if aVal < bVal {
					return Result{Type: "float", Value: aVal}, true
				}
				return Result{Type: "float", Value: bVal}, true
			}
		}
	case "std.math.abs_float":
		if len(operands) == 1 {
			x := operandValue(fr, operands[0])
			if x.Type == "float" {
				xVal := x.Value.(float64)
				if xVal < 0 {
					return Result{Type: "float", Value: -xVal}, true
				}
				return Result{Type: "float", Value: xVal}, true
			}
		}
	case "std.math.toString":
		if len(operands) == 1 {
			n := operandValue(fr, operands[0])
			if n.Type == "int" {
				nVal := n.Value.(int)
				return Result{Type: "string", Value: fmt.Sprintf("%d", nVal)}, true
			}
		}
	case "std.string.length":
		if len(operands) == 1 {
			arg := operandValue(fr, operands[0])
			if arg.Type == "string" {
				str := arg.Value.(string)
				return Result{Type: "int", Value: len(str)}, true
			}
		}
	case "std.string.concat":
		if len(operands) == 2 {
			left := operandValue(fr, operands[0])
			right := operandValue(fr, operands[1])
			if left.Type == "string" && right.Type == "string" {
				leftStr := left.Value.(string)
				rightStr := right.Value.(string)
				return Result{Type: "string", Value: leftStr + rightStr}, true
			}
		}
	case "std.string.substring":
		if len(operands) == 3 {
			s := operandValue(fr, operands[0])
			start := operandValue(fr, operands[1])
			end := operandValue(fr, operands[2])
			if s.Type == "string" && start.Type == "int" && end.Type == "int" {
				str := s.Value.(string)
				startIdx := start.Value.(int)
				endIdx := end.Value.(int)
				if startIdx < 0 || endIdx > len(str) || startIdx > endIdx {
					return Result{Type: "string", Value: ""}, true
				}
				return Result{Type: "string", Value: str[startIdx:endIdx]}, true
			}
		}
	case "std.string.char_at":
		if len(operands) == 2 {
			s := operandValue(fr, operands[0])
			idx := operandValue(fr, operands[1])
			if s.Type == "string" && idx.Type == "int" {
				str := s.Value.(string)
				index := idx.Value.(int)
				if index < 0 || index >= len(str) {
					return Result{Type: "char", Value: ' '}, true
				}
				return Result{Type: "char", Value: rune(str[index])}, true
			}
		}
	case "std.string.starts_with":
		if len(operands) == 2 {
			s := operandValue(fr, operands[0])
			prefix := operandValue(fr, operands[1])
			if s.Type == "string" && prefix.Type == "string" {
				str := s.Value.(string)
				pre := prefix.Value.(string)
				if len(pre) > len(str) {
					return Result{Type: "bool", Value: false}, true
				}
				return Result{Type: "bool", Value: str[:len(pre)] == pre}, true
			}
		}
	case "std.string.ends_with":
		if len(operands) == 2 {
			s := operandValue(fr, operands[0])
			suffix := operandValue(fr, operands[1])
			if s.Type == "string" && suffix.Type == "string" {
				str := s.Value.(string)
				suf := suffix.Value.(string)
				if len(suf) > len(str) {
					return Result{Type: "bool", Value: false}, true
				}
				return Result{Type: "bool", Value: str[len(str)-len(suf):] == suf}, true
			}
		}
	case "std.string.contains":
		if len(operands) == 2 {
			s := operandValue(fr, operands[0])
			substr := operandValue(fr, operands[1])
			if s.Type == "string" && substr.Type == "string" {
				str := s.Value.(string)
				sub := substr.Value.(string)
				return Result{Type: "bool", Value: strings.Contains(str, sub)}, true
			}
		}
	case "std.string.index_of":
		if len(operands) == 2 {
			s := operandValue(fr, operands[0])
			substr := operandValue(fr, operands[1])
			if s.Type == "string" && substr.Type == "string" {
				str := s.Value.(string)
				sub := substr.Value.(string)
				idx := strings.Index(str, sub)
				return Result{Type: "int", Value: idx}, true
			}
		}
	case "std.string.last_index_of":
		if len(operands) == 2 {
			s := operandValue(fr, operands[0])
			substr := operandValue(fr, operands[1])
			if s.Type == "string" && substr.Type == "string" {
				str := s.Value.(string)
				sub := substr.Value.(string)
				idx := strings.LastIndex(str, sub)
				return Result{Type: "int", Value: idx}, true
			}
		}
	case "std.string.trim":
		if len(operands) == 1 {
			s := operandValue(fr, operands[0])
			if s.Type == "string" {
				str := s.Value.(string)
				return Result{Type: "string", Value: strings.TrimSpace(str)}, true
			}
		}
	case "std.string.to_upper":
		if len(operands) == 1 {
			s := operandValue(fr, operands[0])
			if s.Type == "string" {
				str := s.Value.(string)
				return Result{Type: "string", Value: strings.ToUpper(str)}, true
			}
		}
	case "std.string.to_lower":
		if len(operands) == 1 {
			s := operandValue(fr, operands[0])
			if s.Type == "string" {
				str := s.Value.(string)
				return Result{Type: "string", Value: strings.ToLower(str)}, true
			}
		}
	case "std.string.equals":
		if len(operands) == 2 {
			a := operandValue(fr, operands[0])
			b := operandValue(fr, operands[1])
			if a.Type == "string" && b.Type == "string" {
				aStr := a.Value.(string)
				bStr := b.Value.(string)
				return Result{Type: "bool", Value: aStr == bStr}, true
			}
		}
	case "std.string.compare":
		if len(operands) == 2 {
			a := operandValue(fr, operands[0])
			b := operandValue(fr, operands[1])
			if a.Type == "string" && b.Type == "string" {
				aStr := a.Value.(string)
				bStr := b.Value.(string)
				return Result{Type: "int", Value: strings.Compare(aStr, bStr)}, true
			}
		}
	}

	return Result{}, false
}

func operandValue(fr *frame, op mir.Operand) Result {
	switch op.Kind {
	case mir.OperandValue:
		return fr.values[op.Value]
	case mir.OperandLiteral:
		res, err := literalResult(op)
		if err != nil {
			return Result{Type: op.Type, Value: nil}
		}
		return res
	default:
		return Result{Type: "<unknown>", Value: nil}
	}
}

func literalResult(op mir.Operand) (Result, error) {
	typ := op.Type
	if typ == "" {
		typ = inferLiteralType(op.Literal)
	}
	switch typ {
	case "int", "long", "byte":
		v, err := strconv.Atoi(op.Literal)
		if err != nil {
			return Result{}, fmt.Errorf("invalid int literal %q", op.Literal)
		}
		return Result{Type: typ, Value: v}, nil
	case "float", "double":
		v, err := strconv.ParseFloat(op.Literal, 64)
		if err != nil {
			return Result{}, fmt.Errorf("invalid float literal %q", op.Literal)
		}
		return Result{Type: typ, Value: v}, nil
	case "bool":
		return Result{Type: typ, Value: op.Literal == "true"}, nil
	case "char", "string":
		return Result{Type: typ, Value: strings.Trim(op.Literal, "\"")}, nil
	default:
		return Result{Type: typ, Value: nil}, nil
	}
}

func inferLiteralType(lit string) string {
	if lit == "true" || lit == "false" {
		return "bool"
	}
	if _, err := strconv.Atoi(lit); err == nil {
		return "int"
	}
	if strings.HasPrefix(lit, "\"") && strings.HasSuffix(lit, "\"") {
		return "string"
	}
	return "<unknown>"
}

func toInt(value Result) (int, error) {
	switch v := value.Value.(type) {
	case int:
		return v, nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case nil:
		return 0, fmt.Errorf("vm: expected int, got nil")
	default:
		return 0, fmt.Errorf("vm: expected int value, got %T", v)
	}
}

func toBool(value Result) (bool, error) {
	switch v := value.Value.(type) {
	case bool:
		return v, nil
	case int:
		return v != 0, nil
	default:
		return false, fmt.Errorf("vm: expected bool value, got %T", v)
	}
}

func toFloat(value Result) (float64, error) {
	switch v := value.Value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case bool:
		if v {
			return 1.0, nil
		}
		return 0.0, nil
	case nil:
		return 0, fmt.Errorf("vm: expected float, got nil")
	default:
		return 0, fmt.Errorf("vm: expected float value, got %T", v)
	}
}

func toString(value Result) (string, error) {
	switch v := value.Value.(type) {
	case string:
		return v, nil
	case int:
		return fmt.Sprintf("%d", v), nil
	case bool:
		if v {
			return "true", nil
		}
		return "false", nil
	default:
		return "", fmt.Errorf("vm: cannot convert %T to string", v)
	}
}

func chooseType(preferred, fallback string) string {
	if preferred != "" && preferred != inferTypePlaceholder {
		return preferred
	}
	if fallback != "" && fallback != inferTypePlaceholder {
		return fallback
	}
	return "int"
}
