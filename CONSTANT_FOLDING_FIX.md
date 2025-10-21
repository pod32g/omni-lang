# Constant Folding Fix for VM Backend Infinite Loops

## Problem
The VM backend was getting stuck in infinite loops when executing programs with arithmetic operations in loop bodies. For example:

```omni
func main():int {
    var sum: int = 0
    for i: int = 0; i < 3; i++ {
        sum = sum + i  // This caused infinite loops in VM
    }
    return sum
}
```

## Root Cause
The issue was in the **constant folding optimization pass** (`internal/passes/constfold.go`). The pass was too aggressive and was folding arithmetic expressions that involved variables modified by `assign` instructions.

### What was happening:
1. MIR builder correctly generated: `%4 = add %0, %1` (sum + i)
2. Constant folding pass saw that %0 and %1 were initially constants (0 and 0)
3. It folded `0 + 0 = 0` and converted to: `%4 = const 0`
4. This happened on every iteration, causing the loop to never progress

### The bug:
The constant folding pass didn't understand that variables modified by `assign` instructions could change during execution, so it incorrectly folded expressions involving those variables.

## Solution
Modified the constant folding pass to be **smarter about when to fold constants**:

1. **First pass**: Identify all variables that are modified by `assign` instructions
2. **Second pass**: Only fold arithmetic expressions if they don't involve any modified variables

### Code changes in `internal/passes/constfold.go`:

```go
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
```

## Results
After the fix:

### ✅ Classic for loops work:
```omni
func main():int {
    var sum: int = 0
    for i: int = 0; i < 3; i++ {
        sum = sum + i
    }
    return sum  // Returns 3 (0+1+2)
}
```

### ✅ Range for loops work:
```omni
func main():int {
    let numbers: []int = [1, 2, 3]
    var sum: int = 0
    for num in numbers {
        sum = sum + num
    }
    return sum  // Returns 6 (1+2+3)
}
```

### ✅ All e2e tests pass:
- `TestForClassic`: ✅ Pass
- `TestForRange`: ✅ Pass  
- `TestForNested`: ✅ Pass
- `TestArrayArithmetic`: ✅ Pass (both VM and C backends)
- `TestArrayLen`: ✅ Pass (both VM and C backends)

## Impact
- **VM backend**: Now works correctly for all loop types with arithmetic
- **C backend**: Unaffected (was already working)
- **Performance**: Slightly reduced constant folding, but more correct behavior
- **No regressions**: All existing functionality continues to work

## Technical Details
The fix ensures that constant folding only happens when it's safe to do so - when the operands are truly constant and won't change during execution. This prevents the optimization from breaking programs with mutable variables in loops.

The solution is conservative but correct: it errs on the side of not folding rather than incorrectly folding expressions that involve mutable variables.
