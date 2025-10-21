# Array `len()` Function Implementation

## Summary
Successfully implemented the `len()` builtin function for arrays in both VM and C backends.

## What Works

### ✅ Type Checker
- Registered `len` as a builtin function that accepts any array type
- Special handling in `checkCallExpr` to validate array arguments
- Recognizes `len` as an identifier in `checkExpr`

### ✅ VM Backend
- Added `len` to `execIntrinsic` to calculate array length
- Supports `[]int`, `[]string`, `[]float64`, and `[]bool` arrays
- Returns the length as an `int` value

### ✅ C Backend  
- Special handling in `generateInstruction` for `len()` calls
- Generates `sizeof(array) / sizeof(array[0])` directly in C
- No runtime function call needed (compile-time calculation)

### ✅ Tests
- `TestArrayLen` passes for both VM and C backends
- Returns correct length for `[1, 2, 3, 4, 5]` → `5`

## Code Changes

### 1. Type Checker (`internal/types/checker/checker.go`)
```go
// In initBuiltins()
c.functions["len"] = FunctionSignature{
    Params: []string{"array"}, // Special marker for array types
    Return: "int",
}

// In checkExpr() for IdentifierExpr
if sig, exists := c.functions[e.Name]; exists {
    return sig.Return
}

// In checkCallExpr()
if qualifiedName == "len" && expected == "array" {
    if !strings.HasPrefix(argType, "[]<") && !strings.HasPrefix(argType, "array<") {
        c.report(arg.Span(), fmt.Sprintf("len() expects an array, got %s", argType),
            "pass an array to the len() function")
    }
}
```

### 2. VM Backend (`internal/vm/vm.go`)
```go
func execIntrinsic(callee string, operands []mir.Operand, fr *frame) (Result, bool) {
    switch callee {
    case "len":
        if len(operands) == 1 {
            arg := operandValue(fr, operands[0])
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
                }
            }
        }
    }
}
```

### 3. C Backend (`internal/backend/c/c_generator.go`)
```go
case "call", "call.int", "call.void", "call.string", "call.bool":
    if len(inst.Operands) > 0 {
        funcName := g.getOperandValue(inst.Operands[0])
        
        // Special handling for len() function
        if funcName == "len" && len(inst.Operands) == 2 {
            varName := g.getVariableName(inst.ID)
            arrayVar := g.getOperandValue(inst.Operands[1])
            g.output.WriteString(fmt.Sprintf("  %s %s = sizeof(%s) / sizeof(%s[0]);\n",
                g.mapType(inst.Type), varName, arrayVar, arrayVar))
            return nil
        }
    }
```

## Test Results
```bash
$ go test ./tests/e2e -v -run TestArrayLen
=== RUN   TestArrayLen
--- PASS: TestArrayLen (0.39s)
PASS
```

## Example Usage
```omni
func main():int {
    let numbers: []int = [1, 2, 3, 4, 5]
    let length: int = len(numbers)
    return length  // Returns 5
}
```

## Limitations
- VM backend: Loops with assignments have a MIR builder bug (arrays with for-loops work in C backend only)
- String comparisons not yet supported in VM backend

## Related Work
- Implemented array support (`[]int`, `[]string`)
- Added array indexing for VM backend
- Fixed C backend array type generation
- Added `assign` instruction support for VM backend

