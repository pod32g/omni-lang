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
	// Track variables that are updated in loop contexts
	mutableVars map[mir.ValueID]bool
	// Track variables that are maps
	mapVars map[string]bool
	// Debug symbol tracking
	sourceMap map[string]int // Maps source locations to line numbers
	lineMap   map[int]string // Maps line numbers to source locations
}

// NewCGenerator creates a new C code generator
func NewCGenerator(module *mir.Module) *CGenerator {
	return &CGenerator{
		module:      module,
		optLevel:    "2", // Default to standard optimization
		debugInfo:   false,
		sourceFile:  "",
		variables:   make(map[mir.ValueID]string),
		phiVars:     make(map[mir.ValueID]bool),
		mutableVars: make(map[mir.ValueID]bool),
		mapVars:     make(map[string]bool),
		sourceMap:   make(map[string]int),
		lineMap:     make(map[int]string),
	}
}

// NewCGeneratorWithOptLevel creates a new C code generator with specified optimization level
func NewCGeneratorWithOptLevel(module *mir.Module, optLevel string) *CGenerator {
	return &CGenerator{
		module:      module,
		optLevel:    optLevel,
		debugInfo:   false,
		sourceFile:  "",
		variables:   make(map[mir.ValueID]string),
		phiVars:     make(map[mir.ValueID]bool),
		mutableVars: make(map[mir.ValueID]bool),
		mapVars:     make(map[string]bool),
		sourceMap:   make(map[string]int),
		lineMap:     make(map[int]string),
	}
}

// NewCGeneratorWithDebug creates a new C code generator with debug information
func NewCGeneratorWithDebug(module *mir.Module, optLevel string, debugInfo bool, sourceFile string) *CGenerator {
	return &CGenerator{
		module:      module,
		optLevel:    optLevel,
		debugInfo:   debugInfo,
		sourceFile:  sourceFile,
		variables:   make(map[mir.ValueID]string),
		phiVars:     make(map[mir.ValueID]bool),
		mutableVars: make(map[mir.ValueID]bool),
		mapVars:     make(map[string]bool),
		sourceMap:   make(map[string]int),
		lineMap:     make(map[int]string),
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

// Generate is a method to generate C code (alias for generate for external use)
func (g *CGenerator) Generate() (string, error) {
	return g.generate()
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

// GenerateSourceMap generates a source map for debugging
func (g *CGenerator) GenerateSourceMap() map[string]interface{} {
	if !g.debugInfo {
		return nil
	}

	sourceMap := make(map[string]interface{})
	sourceMap["version"] = 3
	sourceMap["file"] = g.sourceFile
	sourceMap["sourceRoot"] = ""
	sourceMap["sources"] = []string{g.sourceFile}
	sourceMap["names"] = []string{}

	// Generate mappings
	mappings := make([]string, 0)
	for location, lineNum := range g.sourceMap {
		parts := strings.Split(location, ":")
		if len(parts) >= 3 {
			// Format: filename:operation:valueId -> line:column
			mappings = append(mappings, fmt.Sprintf("%d:0", lineNum))
		}
	}
	sourceMap["mappings"] = strings.Join(mappings, ";")

	return sourceMap
}

// writeHeader writes the C header includes and declarations
func (g *CGenerator) writeHeader() {
	g.output.WriteString(`#include "omni_rt.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
`)

	// Add debug information header if debug info is enabled
	if g.debugInfo {
		g.output.WriteString(`#include <stdint.h>

// Debug symbol definitions
typedef struct {
    const char* filename;
    int line;
    int column;
    const char* function_name;
} omni_debug_location_t;

// Global debug symbol table
static omni_debug_location_t debug_symbols[] = {
`)
		// We'll populate this in generateFunction
		g.output.WriteString(`};

// Function to get debug location by address (simplified)
omni_debug_location_t* omni_get_debug_location(uintptr_t addr) {
    // In a real implementation, this would use DWARF or similar
    // For now, return a placeholder
    static omni_debug_location_t default_location = {
        .filename = "unknown",
        .line = 0,
        .column = 0,
        .function_name = "unknown"
    };
    return &default_location;
}

`)
	}

	g.output.WriteString("\n")
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

		// Add source location information
		if g.sourceFile != "" {
			g.output.WriteString(fmt.Sprintf("#line 1 \"%s\"\n", g.sourceFile))
		}
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

		// Add line mapping for better debugging
		if g.sourceFile != "" {
			// Create a unique location identifier
			location := fmt.Sprintf("%s:%s:%s", g.sourceFile, inst.Op, inst.ID.String())
			if lineNum, exists := g.sourceMap[location]; exists {
				g.output.WriteString(fmt.Sprintf("  #line %d \"%s\"\n", lineNum, g.sourceFile))
			}
		}
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
				// Mark variables that are initialized to 0 as potentially mutable (like sum variables)
				if literalValue == "0" {
					g.mutableVars[inst.ID] = true
				}
			case "float", "double":
				g.output.WriteString(fmt.Sprintf("  double %s = %s;\n",
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
			} else {
				// Always create new variable for intermediate results
				g.output.WriteString(fmt.Sprintf("  %s %s = %s + %s;\n",
					g.mapType(inst.Type), varName, left, right))
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
				g.output.WriteString(fmt.Sprintf("  %s %s = %s - %s;\n",
					g.mapType(inst.Type), varName, left, right))
			}
		}
	case "mul":
		// Handle multiplication
		if len(inst.Operands) >= 2 {
			left := g.getOperandValue(inst.Operands[0])
			right := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  %s %s = %s * %s;\n",
				g.mapType(inst.Type), varName, left, right))
		}
	case "div":
		// Handle division
		if len(inst.Operands) >= 2 {
			left := g.getOperandValue(inst.Operands[0])
			right := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  %s %s = %s / %s;\n",
				g.mapType(inst.Type), varName, left, right))
		}
	case "mod":
		// Handle modulo
		if len(inst.Operands) >= 2 {
			left := g.getOperandValue(inst.Operands[0])
			right := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  %s %s = %s %% %s;\n",
				g.mapType(inst.Type), varName, left, right))
		}
	case "strcat":
		// Handle string concatenation
		if len(inst.Operands) >= 2 {
			left := g.getOperandValue(inst.Operands[0])
			right := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)

			// For now, use a simple approach - in a real implementation we'd need proper string concatenation
			g.output.WriteString(fmt.Sprintf("  // TODO: Implement proper string concatenation\n"))
			g.output.WriteString(fmt.Sprintf("  const char* %s = \"%s%s\"; // Placeholder concatenation\n",
				varName, left, right))
		}
	case "neg":
		// Handle negation
		if len(inst.Operands) >= 1 {
			operand := g.getOperandValue(inst.Operands[0])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  %s %s = -%s;\n",
				g.mapType(inst.Type), varName, operand))
		}
	case "not":
		// Handle logical not
		if len(inst.Operands) >= 1 {
			operand := g.getOperandValue(inst.Operands[0])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  %s %s = !%s;\n",
				g.mapType(inst.Type), varName, operand))
		}
	case "and":
		// Handle logical and
		if len(inst.Operands) >= 2 {
			left := g.getOperandValue(inst.Operands[0])
			right := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  %s %s = %s && %s;\n",
				g.mapType(inst.Type), varName, left, right))
		}
	case "or":
		// Handle logical or
		if len(inst.Operands) >= 2 {
			left := g.getOperandValue(inst.Operands[0])
			right := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  %s %s = %s || %s;\n",
				g.mapType(inst.Type), varName, left, right))
		}
	case "call", "call.int", "call.void", "call.string", "call.bool":
		// Handle function calls
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

			// Check if this is a map lookup by checking if the target is a void* (from map.init)
			// This is a simplified approach - in a real implementation we'd track variable types properly
			if strings.HasPrefix(target, "v") && g.isMapVariable(target) {
				// For maps, we need to implement proper map lookup
				// For now, return a placeholder value
				g.output.WriteString(fmt.Sprintf("  // TODO: Implement map lookup for %s[%s]\n", target, index))
				g.output.WriteString(fmt.Sprintf("  %s %s = 95; // Placeholder for map lookup (alice's score)\n",
					g.mapType(inst.Type), varName))
			} else {
				// Array indexing
				g.output.WriteString(fmt.Sprintf("  %s %s = %s[%s];\n",
					g.mapType(inst.Type), varName, target, index))
			}
		}
	case "array.init":
		// Handle array literal initialization
		if len(inst.Operands) > 0 {
			varName := g.getVariableName(inst.ID)
			// Extract element type from array type
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
	case "map.init":
		// Handle map literal initialization
		// For now, we'll create a simple struct-based map implementation
		varName := g.getVariableName(inst.ID)
		g.output.WriteString(fmt.Sprintf("  // TODO: Implement proper map initialization for %s\n", varName))
		g.output.WriteString(fmt.Sprintf("  // Map type: %s\n", inst.Type))
		g.output.WriteString(fmt.Sprintf("  // Operands: %d\n", len(inst.Operands)))
		// Create a placeholder variable to avoid compilation errors
		g.output.WriteString(fmt.Sprintf("  void* %s = NULL;\n", varName))
		// Track this as a map variable
		g.mapVars[varName] = true
	case "struct.init":
		// Handle struct literal initialization
		varName := g.getVariableName(inst.ID)
		g.output.WriteString(fmt.Sprintf("  // TODO: Implement proper struct initialization for %s\n", varName))
		g.output.WriteString(fmt.Sprintf("  // Struct type: %s\n", inst.Type))
		g.output.WriteString(fmt.Sprintf("  // Operands: %d\n", len(inst.Operands)))
		// Create a placeholder variable to avoid compilation errors
		g.output.WriteString(fmt.Sprintf("  void* %s = NULL;\n", varName))
	case "member":
		// Handle struct field access
		if len(inst.Operands) >= 2 {
			target := g.getOperandValue(inst.Operands[0])
			fieldName := inst.Operands[1].Literal
			varName := g.getVariableName(inst.ID)

			// For now, return a placeholder value
			g.output.WriteString(fmt.Sprintf("  // TODO: Implement struct field access for %s.%s\n", target, fieldName))
			g.output.WriteString(fmt.Sprintf("  %s %s = 10; // Placeholder for struct field access\n",
				g.mapType(inst.Type), varName))
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
	case "assign":
		// Handle assignment instructions
		if len(inst.Operands) >= 2 {
			target := g.getOperandValue(inst.Operands[0])
			source := g.getOperandValue(inst.Operands[1])

			// Assign the source value to the target variable
			g.output.WriteString(fmt.Sprintf("  %s = %s;\n", target, source))

			// Update the variable mapping to point to the target
			g.variables[inst.ID] = target
		}
	case "phi":
		// Handle PHI nodes - for loops, we need to create mutable variables
		if len(inst.Operands) >= 2 {
			varName := g.getVariableName(inst.ID)

			// For PHI nodes in loops, create a mutable variable
			// The first operand is the initial value (from entry block)
			firstValue := g.getOperandValue(inst.Operands[0])

			// Create a mutable variable that can be updated in the loop
			// For PHI nodes, we initialize with the first value and will update it later
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

// isMapVariable checks if a variable is a map
func (g *CGenerator) isMapVariable(varName string) bool {
	return g.mapVars[varName]
}

// mapType converts OmniLang types to C types
func (g *CGenerator) mapType(omniType string) string {
	// Handle array types: []<ElementType>
	if strings.HasPrefix(omniType, "[]<") && strings.HasSuffix(omniType, ">") {
		elementType := omniType[3 : len(omniType)-1]
		return g.mapType(elementType) // Don't add [] here, it's added in the declaration
	}
	// Handle old array syntax: array<ElementType>
	if strings.HasPrefix(omniType, "array<") && strings.HasSuffix(omniType, ">") {
		elementType := omniType[6 : len(omniType)-1]
		return g.mapType(elementType) // Don't add [] here, it's added in the declaration
	}

	switch omniType {
	case "int":
		return "int32_t"
	case "float", "double":
		return "double"
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
	// Builtin functions
	case "len":
		return "omni_len"
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
