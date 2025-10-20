package cbackend

import (
	"fmt"
	"strings"

	"github.com/omni-lang/omni/internal/mir"
)

// CGenerator handles the translation from MIR to C code
type CGenerator struct {
	module     *mir.Module
	output     strings.Builder
	optLevel   string
	debugInfo  bool
	sourceFile string
	// Track variable names for SSA values
	variables map[mir.ValueID]string
	// Track PHI variables that need special handling
	phiVars map[mir.ValueID]bool
}

// NewCGenerator creates a new C code generator
func NewCGenerator(module *mir.Module) *CGenerator {
	return &CGenerator{
		module:     module,
		optLevel:   "2", // Default to standard optimization
		debugInfo:  false,
		sourceFile: "",
		variables:  make(map[mir.ValueID]string),
		phiVars:    make(map[mir.ValueID]bool),
	}
}

// NewCGeneratorWithOptLevel creates a new C code generator with specified optimization level
func NewCGeneratorWithOptLevel(module *mir.Module, optLevel string) *CGenerator {
	return &CGenerator{
		module:     module,
		optLevel:   optLevel,
		debugInfo:  false,
		sourceFile: "",
		variables:  make(map[mir.ValueID]string),
		phiVars:    make(map[mir.ValueID]bool),
	}
}

// NewCGeneratorWithDebug creates a new C code generator with debug information
func NewCGeneratorWithDebug(module *mir.Module, optLevel string, debugInfo bool, sourceFile string) *CGenerator {
	return &CGenerator{
		module:     module,
		optLevel:   optLevel,
		debugInfo:  debugInfo,
		sourceFile: sourceFile,
		variables:  make(map[mir.ValueID]string),
		phiVars:    make(map[mir.ValueID]bool),
	}
}

// GenerateC generates C code from a MIR module
func GenerateC(module *mir.Module) (string, error) {
	gen := NewCGenerator(module)
	return gen.generate()
}

// GenerateCOptimized generates optimized C code from a MIR module
func GenerateCOptimized(module *mir.Module, optLevel string) (string, error) {
	gen := NewCGeneratorWithOptLevel(module, optLevel)
	return gen.generate()
}

// GenerateCWithDebug generates C code with debug information from a MIR module
func GenerateCWithDebug(module *mir.Module, optLevel string, debugInfo bool, sourceFile string) (string, error) {
	gen := NewCGeneratorWithDebug(module, optLevel, debugInfo, sourceFile)
	return gen.generate()
}

// generate produces the complete C code
func (g *CGenerator) generate() (string, error) {
	g.writeHeader()
	g.writeStdLibFunctions()

	for _, fn := range g.module.Functions {
		if err := g.generateFunction(fn); err != nil {
			return "", err
		}
	}

	g.writeMain()

	// Apply optimizations
	code := g.output.String()
	optimizedCode := OptimizeC(code, g.optLevel)

	return optimizedCode, nil
}

// writeHeader writes the C header includes and declarations
func (g *CGenerator) writeHeader() {
	g.output.WriteString(`#include "omni_rt.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

`)
}

// writeStdLibFunctions writes standard library function implementations
func (g *CGenerator) writeStdLibFunctions() {
	// Note: Standard library functions are now provided by the runtime
	// No need to generate them here since they're linked from libomni_rt.so
}

// generateFunction generates C code for a single function
func (g *CGenerator) generateFunction(fn *mir.Function) error {
	// Add debug information if enabled
	if g.debugInfo {
		g.output.WriteString(fmt.Sprintf("// Debug: Function %s (Return: %s, Params: %d)\n",
			fn.Name, fn.ReturnType, len(fn.Params)))
	}

	// Generate function signature
	returnType := g.mapType(fn.ReturnType)
	funcName := fn.Name
	if funcName == "main" {
		funcName = "omni_main"
	}
	g.output.WriteString(fmt.Sprintf("%s %s(", returnType, funcName))

	// Generate parameters
	for i, param := range fn.Params {
		if i > 0 {
			g.output.WriteString(", ")
		}
		paramType := g.mapType(param.Type)
		g.output.WriteString(fmt.Sprintf("%s %s", paramType, param.Name))
	}
	g.output.WriteString(") {\n")

	// Generate function body
	for _, block := range fn.Blocks {
		if err := g.generateBlock(block); err != nil {
			return err
		}
	}

	g.output.WriteString("}\n\n")
	return nil
}

// generateBlock generates C code for a basic block
func (g *CGenerator) generateBlock(block *mir.BasicBlock) error {
	// Generate block label if it's not the entry block
	if block.Name != "entry" {
		g.output.WriteString(fmt.Sprintf("  %s:\n", block.Name))
	}

	// Generate instructions
	for _, inst := range block.Instructions {
		if err := g.generateInstruction(&inst); err != nil {
			return err
		}
	}

	// Generate terminator
	if err := g.generateTerminator(&block.Terminator); err != nil {
		return err
	}

	return nil
}

// generateInstruction generates C code for a single instruction
func (g *CGenerator) generateInstruction(inst *mir.Instruction) error {
	// Add debug information if enabled
	if g.debugInfo {
		g.output.WriteString(fmt.Sprintf("  // Debug: %s instruction (ID: %s, Type: %s)\n",
			inst.Op, inst.ID.String(), inst.Type))
	}

	switch inst.Op {
	case "const":
		// Handle constants based on type
		if len(inst.Operands) > 0 && inst.Operands[0].Kind == mir.OperandLiteral {
			varName := g.getVariableName(inst.ID)
			literalValue := g.getOperandValue(inst.Operands[0])
			switch inst.Type {
			case "int":
				g.output.WriteString(fmt.Sprintf("  int32_t %s = %s;\n",
					varName, literalValue))
			case "string":
				g.output.WriteString(fmt.Sprintf("  const char* %s = %s;\n",
					varName, literalValue))
			case "bool":
				g.output.WriteString(fmt.Sprintf("  int32_t %s = %s;\n",
					varName, literalValue))
			default:
				g.output.WriteString(fmt.Sprintf("  // TODO: Handle const type %s\n", inst.Type))
			}
		}
	case "add":
		// Handle addition
		if len(inst.Operands) >= 2 {
			left := g.getOperandValue(inst.Operands[0])
			right := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)
			
			// Check if the left operand is a PHI variable (loop variable)
			if inst.Operands[0].Kind == mir.OperandValue && g.phiVars[inst.Operands[0].Value] {
				// This is an increment to a PHI variable - update it in place
				g.output.WriteString(fmt.Sprintf("  %s = %s + %s;\n",
					left, left, right))
			} else if inst.Operands[0].Kind == mir.OperandValue {
				// Check if this is an assignment to a variable that should be mutable
				// For now, let's make all variables that are used in additions mutable
				// This is a simplification - a proper implementation would need more analysis
				g.output.WriteString(fmt.Sprintf("  %s = %s + %s;\n",
					left, left, right))
			} else {
				// Regular addition - create new variable
				g.output.WriteString(fmt.Sprintf("  int32_t %s = %s + %s;\n",
					varName, left, right))
			}
		}
	case "sub":
		// Handle subtraction
		if len(inst.Operands) >= 2 {
			left := g.getOperandValue(inst.Operands[0])
			right := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)

			// Check if the left operand is a PHI variable (loop variable)
			if inst.Operands[0].Kind == mir.OperandValue && g.phiVars[inst.Operands[0].Value] {
				// This is a decrement to a PHI variable - update it in place
				g.output.WriteString(fmt.Sprintf("  %s = %s - %s;\n",
					left, left, right))
			} else {
				// Regular subtraction - create new variable
				g.output.WriteString(fmt.Sprintf("  int32_t %s = %s - %s;\n",
					varName, left, right))
			}
		}
	case "mul":
		// Handle multiplication
		if len(inst.Operands) >= 2 {
			left := g.getOperandValue(inst.Operands[0])
			right := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  int32_t %s = %s * %s;\n",
				varName, left, right))
		}
	case "div":
		// Handle division
		if len(inst.Operands) >= 2 {
			left := g.getOperandValue(inst.Operands[0])
			right := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  int32_t %s = %s / %s;\n",
				varName, left, right))
		}
	case "call", "call.int", "call.void", "call.string", "call.bool":
		// Handle function calls
		if len(inst.Operands) > 0 {
			funcName := g.getOperandValue(inst.Operands[0])
			cFuncName := g.mapFunctionName(funcName)

			// Handle void function calls differently
			if inst.Type == "void" {
				g.output.WriteString(fmt.Sprintf("  %s(", cFuncName))
				// Add arguments
				for i, arg := range inst.Operands[1:] {
					if i > 0 {
						g.output.WriteString(", ")
					}
					g.output.WriteString(g.getOperandValue(arg))
				}
				g.output.WriteString(");\n")
			} else {
				varName := g.getVariableName(inst.ID)
				g.output.WriteString(fmt.Sprintf("  %s %s = %s(",
					g.mapType(inst.Type), varName, cFuncName))
				// Add arguments
				for i, arg := range inst.Operands[1:] {
					if i > 0 {
						g.output.WriteString(", ")
					}
					g.output.WriteString(g.getOperandValue(arg))
				}
				g.output.WriteString(");\n")
			}
		}
	case "index":
		// Handle array/map indexing
		if len(inst.Operands) >= 2 {
			target := g.getOperandValue(inst.Operands[0])
			index := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  %s %s = %s[%s];\n",
				g.mapType(inst.Type), varName, target, index))
		}
	case "array.init":
		// Handle array literal initialization
		if len(inst.Operands) > 0 {
			varName := g.getVariableName(inst.ID)
			elementType := g.mapType(inst.Type)
			// For now, create a simple array with the elements
			g.output.WriteString(fmt.Sprintf("  %s %s[] = {", elementType, varName))
			for i, op := range inst.Operands {
				if i > 0 {
					g.output.WriteString(", ")
				}
				g.output.WriteString(g.getOperandValue(op))
			}
			g.output.WriteString("};\n")
		}
	case "cmp.eq", "cmp.neq", "cmp.lt", "cmp.lte", "cmp.gt", "cmp.gte":
		// Handle comparison operations
		if len(inst.Operands) >= 2 {
			left := g.getOperandValue(inst.Operands[0])
			right := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)

			var op string
			switch inst.Op {
			case "cmp.eq":
				op = "=="
			case "cmp.neq":
				op = "!="
			case "cmp.lt":
				op = "<"
			case "cmp.lte":
				op = "<="
			case "cmp.gt":
				op = ">"
			case "cmp.gte":
				op = ">="
			}

			g.output.WriteString(fmt.Sprintf("  int32_t %s = (%s %s %s) ? 1 : 0;\n",
				varName, left, op, right))
		}
	case "phi":
		// Handle PHI nodes - for loops, we need to create mutable variables
		if len(inst.Operands) >= 2 {
			varName := g.getVariableName(inst.ID)

			// For PHI nodes in loops, create a mutable variable
			// The first operand is the initial value (from entry block)
			firstValue := g.getOperandValue(inst.Operands[0])

			// Create a mutable variable that can be updated in the loop
			g.output.WriteString(fmt.Sprintf("  %s %s = %s;\n",
				g.mapType(inst.Type), varName, firstValue))

			// Track this as a PHI variable that needs special handling
			g.phiVars[inst.ID] = true
		}
	default:
		// Handle unknown instructions
		g.output.WriteString(fmt.Sprintf("  // TODO: Implement instruction %s\n", inst.Op))
	}

	return nil
}

// generateTerminator generates C code for a terminator
func (g *CGenerator) generateTerminator(term *mir.Terminator) error {
	switch term.Op {
	case "ret":
		// Handle return statement
		if len(term.Operands) > 0 {
			value := g.getOperandValue(term.Operands[0])
			g.output.WriteString(fmt.Sprintf("  return %s;\n", value))
		} else {
			g.output.WriteString("  return;\n")
		}
	case "jmp":
		// Handle unconditional jump
		if len(term.Operands) > 0 {
			blockName := g.getOperandValue(term.Operands[0])
			g.output.WriteString(fmt.Sprintf("  goto %s;\n", blockName))
		}
	case "br":
		// Handle unconditional branch
		if len(term.Operands) > 0 {
			blockName := g.getOperandValue(term.Operands[0])
			g.output.WriteString(fmt.Sprintf("  goto %s;\n", blockName))
		}
	case "cbr":
		// Handle conditional branch
		if len(term.Operands) >= 3 {
			condition := g.getOperandValue(term.Operands[0])
			trueBlock := g.getOperandValue(term.Operands[1])
			falseBlock := g.getOperandValue(term.Operands[2])
			g.output.WriteString(fmt.Sprintf("  if (%s) {\n", condition))
			g.output.WriteString(fmt.Sprintf("    goto %s;\n", trueBlock))
			g.output.WriteString("  } else {\n")
			g.output.WriteString(fmt.Sprintf("    goto %s;\n", falseBlock))
			g.output.WriteString("  }\n")
		}
	default:
		// Handle unknown terminators
		g.output.WriteString(fmt.Sprintf("  // TODO: Implement terminator %s\n", term.Op))
	}

	return nil
}

// getOperandValue converts an operand to its C representation
func (g *CGenerator) getOperandValue(op mir.Operand) string {
	switch op.Kind {
	case mir.OperandValue:
		return g.getVariableName(op.Value)
	case mir.OperandLiteral:
		// Convert boolean literals to C format
		if op.Literal == "true" {
			return "1"
		} else if op.Literal == "false" {
			return "0"
		}
		return op.Literal
	default:
		return "/* unknown operand */"
	}
}

// mapType converts OmniLang types to C types
func (g *CGenerator) mapType(omniType string) string {
	switch omniType {
	case "int":
		return "int32_t"
	case "string":
		return "const char*"
	case "void":
		return "void"
	case "bool":
		return "int32_t"
	default:
		return "int32_t" // Default fallback
	}
}

// mapFunctionName maps OmniLang function names to C function names
func (g *CGenerator) mapFunctionName(funcName string) string {
	switch funcName {
	// Math functions
	case "std.math.abs":
		return "omni_abs"
	case "std.math.max":
		return "omni_max"
	case "std.math.min":
		return "omni_min"
	case "std.math.toString":
		return "omni_int_to_string"

	// IO functions
	case "std.io.print":
		return "omni_print_string"
	case "std.io.println":
		return "omni_println_string"
	case "std.io.print_int":
		return "omni_print_int"
	case "std.io.println_int":
		return "omni_println_int"
	case "std.io.print_float":
		return "omni_print_float"
	case "std.io.println_float":
		return "omni_println_float"
	case "std.io.print_bool":
		return "omni_print_bool"
	case "std.io.println_bool":
		return "omni_println_bool"

	default:
		return funcName
	}
}

// getVariableName returns a C variable name for an SSA value
func (g *CGenerator) getVariableName(id mir.ValueID) string {
	if name, exists := g.variables[id]; exists {
		return name
	}
	// Generate a new variable name
	varName := fmt.Sprintf("v%d", int(id))
	g.variables[id] = varName
	return varName
}

// writeMain writes the main function that calls the OmniLang main
func (g *CGenerator) writeMain() {
	g.output.WriteString(`int main() {
    int32_t result = omni_main();
    printf("OmniLang program result: %d\n", result);
    return (int)result;
}
`)
}
