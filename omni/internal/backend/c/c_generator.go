package cbackend

import (
	"fmt"
	"strconv"
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

	// Generate function declarations first
	g.writeFunctionDeclarations()

	// Then generate function definitions
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

// writeFunctionDeclarations writes function declarations for all functions
func (g *CGenerator) writeFunctionDeclarations() {
	for _, fn := range g.module.Functions {
		// Skip functions that are provided by the runtime
		if g.isRuntimeProvidedFunction(fn.Name) {
			continue
		}

		// Generate function declaration
		returnType := g.mapType(fn.ReturnType)
		funcName := g.mapFunctionName(fn.Name)
		if fn.Name == "main" {
			funcName = "omni_main"
		}

		// Handle function pointer return types
		if strings.Contains(fn.ReturnType, ") -> ") {
			// This is a function pointer return type - need special handling
			g.output.WriteString(g.generateCompleteFunctionSignature(fn.ReturnType, funcName, fn.Params))
			g.output.WriteString(";\n")
		} else {
			g.output.WriteString(fmt.Sprintf("%s %s(", returnType, funcName))

			// Generate parameters
			for i, param := range fn.Params {
				if i > 0 {
					g.output.WriteString(", ")
				}
				// Check if this is a function pointer type
				if strings.Contains(param.Type, ") -> ") {
					// Generate function pointer parameter with correct C syntax
					g.output.WriteString(g.mapFunctionTypeWithName(param.Type, param.Name))
				} else {
					paramType := g.mapType(param.Type)

					// Special handling for file.read buffer parameter
					if fn.Name == "file.read" && param.Name == "buffer" && param.Type == "string" {
						paramType = "char*" // Buffer parameter should be writable
					}

					g.output.WriteString(fmt.Sprintf("%s %s", paramType, param.Name))
				}
			}
			g.output.WriteString(");\n")
		}
	}
	g.output.WriteString("\n")
}

// generateFunction generates C code for a single function
func (g *CGenerator) generateFunction(fn *mir.Function) error {
	// Skip functions that are provided by the runtime
	if g.isRuntimeProvidedFunction(fn.Name) {
		return nil
	}

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
	funcName := g.mapFunctionName(fn.Name)
	if fn.Name == "main" {
		funcName = "omni_main"
	}

	// Handle function pointer return types
	if strings.Contains(fn.ReturnType, ") -> ") {
		// This is a function pointer return type - need special handling
		g.output.WriteString(g.generateCompleteFunctionSignature(fn.ReturnType, funcName, fn.Params))
		g.output.WriteString(" {\n")
	} else {
		g.output.WriteString(fmt.Sprintf("%s %s(", returnType, funcName))

		// Generate parameters
		for i, param := range fn.Params {
			if i > 0 {
				g.output.WriteString(", ")
			}
			// Check if this is a function pointer type
			if strings.Contains(param.Type, ") -> ") {
				// Generate function pointer parameter with correct C syntax
				g.output.WriteString(g.mapFunctionTypeWithName(param.Type, param.Name))
			} else {
				paramType := g.mapType(param.Type)

				// Special handling for file.read buffer parameter
				if fn.Name == "file.read" && param.Name == "buffer" && param.Type == "string" {
					paramType = "char*" // Buffer parameter should be writable
				}

				g.output.WriteString(fmt.Sprintf("%s %s", paramType, param.Name))
			}
		}
		g.output.WriteString(") {\n")
	}

	// Map parameter SSA values to their names
	for _, param := range fn.Params {
		g.variables[param.ID] = param.Name
	}

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
				// Convert hex and binary literals to decimal
				convertedValue := g.convertLiteralToDecimal(literalValue)
				g.output.WriteString(fmt.Sprintf("  int32_t %s = %s;\n",
					varName, convertedValue))
				// Mark variables that are initialized to 0 as potentially mutable (like sum variables)
				if convertedValue == "0" {
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
			case "null":
				g.output.WriteString(fmt.Sprintf("  void* %s = NULL;\n",
					varName))
			default:
				// Check if this is a function type
				if strings.Contains(inst.Type, ") -> ") {
					// Function type - assign function pointer
					g.output.WriteString(fmt.Sprintf("  %s %s = %s;\n",
						g.mapType(inst.Type), varName, literalValue))
				} else {
					g.output.WriteString(fmt.Sprintf("  // TODO: Handle const type %s\n", inst.Type))
				}
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
	case "bitand":
		// Handle bitwise AND
		if len(inst.Operands) >= 2 {
			left := g.getOperandValue(inst.Operands[0])
			right := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  %s %s = %s & %s;\n",
				g.mapType(inst.Type), varName, left, right))
		}
	case "bitor":
		// Handle bitwise OR
		if len(inst.Operands) >= 2 {
			left := g.getOperandValue(inst.Operands[0])
			right := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  %s %s = %s | %s;\n",
				g.mapType(inst.Type), varName, left, right))
		}
	case "bitxor":
		// Handle bitwise XOR
		if len(inst.Operands) >= 2 {
			left := g.getOperandValue(inst.Operands[0])
			right := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  %s %s = %s ^ %s;\n",
				g.mapType(inst.Type), varName, left, right))
		}
	case "lshift":
		// Handle left shift
		if len(inst.Operands) >= 2 {
			left := g.getOperandValue(inst.Operands[0])
			right := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  %s %s = %s << %s;\n",
				g.mapType(inst.Type), varName, left, right))
		}
	case "rshift":
		// Handle right shift
		if len(inst.Operands) >= 2 {
			left := g.getOperandValue(inst.Operands[0])
			right := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  %s %s = %s >> %s;\n",
				g.mapType(inst.Type), varName, left, right))
		}
	case "strcat":
		// Handle string concatenation
		if len(inst.Operands) >= 2 {
			varName := g.getVariableName(inst.ID)

			// Convert operands to strings if needed
			leftStr := g.convertOperandToString(inst.Operands[0])
			rightStr := g.convertOperandToString(inst.Operands[1])

			// Generate proper string concatenation using runtime function
			g.output.WriteString(fmt.Sprintf("  const char* %s = omni_strcat(%s, %s);\n",
				varName, leftStr, rightStr))
		}
	case "throw":
		// Handle throw statement - for now, just print the exception and continue
		// In a full implementation, this would set a global exception state
		if len(inst.Operands) >= 1 {
			exceptionValue := g.getOperandValue(inst.Operands[0])
			g.output.WriteString(fmt.Sprintf("  // Throwing exception: %s\n", exceptionValue))
			g.output.WriteString(fmt.Sprintf("  printf(\"Exception: %%s\\n\", %s);\n", exceptionValue))
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
	case "bitnot":
		// Handle bitwise not
		if len(inst.Operands) >= 1 {
			operand := g.getOperandValue(inst.Operands[0])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  %s %s = ~%s;\n",
				g.mapType(inst.Type), varName, operand))
		}
	case "cast":
		// Handle type cast
		if len(inst.Operands) >= 1 {
			operand := g.getOperandValue(inst.Operands[0])
			varName := g.getVariableName(inst.ID)
			targetType := g.mapType(inst.Type)
			g.output.WriteString(fmt.Sprintf("  %s %s = (%s)%s;\n",
				targetType, varName, targetType, operand))
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
			// Get the function name from the first operand
			// For literal function names, use the Literal field directly
			var funcName string
			if inst.Operands[0].Kind == mir.OperandLiteral {
				funcName = inst.Operands[0].Literal
			} else {
				funcName = g.getOperandValue(inst.Operands[0])
			}

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
				// Check if the result type is a function type
				if strings.Contains(inst.Type, ") -> ") {
					// Generate function pointer variable declaration
					g.output.WriteString(fmt.Sprintf("  %s = %s(",
						g.mapFunctionTypeWithName(inst.Type, varName), cFuncName))
				} else {
					// Regular variable declaration
					g.output.WriteString(fmt.Sprintf("  %s %s = %s(",
						g.mapType(inst.Type), varName, cFuncName))
				}
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

			// Check if target is a map variable
			if g.isMapVariable(target) {
				// Map indexing - assume string key for now (can be enhanced later)
				g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_map_get_string_int(%s, %s);\n", varName, target, index))
			} else {
				// Array indexing
				g.output.WriteString(fmt.Sprintf("  %s %s = %s[%s];\n", g.mapType(inst.Type), varName, target, index))
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
		// Handle map initialization
		varName := g.getVariableName(inst.ID)
		g.output.WriteString(fmt.Sprintf("  omni_map_t* %s = omni_map_create();\n", varName))

		// Process key-value pairs from operands
		for i := 0; i < len(inst.Operands); i += 2 {
			if i+1 < len(inst.Operands) {
				key := g.getOperandValue(inst.Operands[i])
				value := g.getOperandValue(inst.Operands[i+1])

				// Determine key and value types from the map type
				mapType := inst.Type
				if strings.HasPrefix(mapType, "map<") && strings.HasSuffix(mapType, ">") {
					inner := mapType[4 : len(mapType)-1] // Remove "map<" and ">"
					parts := strings.Split(inner, ",")
					if len(parts) == 2 {
						keyType := strings.TrimSpace(parts[0])
						valueType := strings.TrimSpace(parts[1])

						// Generate appropriate put function call based on types
						if keyType == "string" && valueType == "int" {
							g.output.WriteString(fmt.Sprintf("  omni_map_put_string_int(%s, %s, %s);\n", varName, key, value))
						} else if keyType == "int" && valueType == "int" {
							g.output.WriteString(fmt.Sprintf("  omni_map_put_int_int(%s, %s, %s);\n", varName, key, value))
						} else {
							// Fallback for other type combinations
							g.output.WriteString(fmt.Sprintf("  // TODO: Handle map type %s\n", mapType))
						}
					}
				}
			}
		}

		// Mark this variable as a map
		g.mapVars[varName] = true
	case "struct.init":
		// Handle struct initialization
		varName := g.getVariableName(inst.ID)
		g.output.WriteString(fmt.Sprintf("  omni_struct_t* %s = omni_struct_create();\n", varName))

		// Process field-value pairs from operands
		// Handle different operand formats: [field1Value, field2Value, ...] or [field1Name, field1Value, field2Name, field2Value, ...]
		if len(inst.Operands) > 0 {
			// Check if first operand is a field name (literal) or field value
			if len(inst.Operands) >= 2 && inst.Operands[0].Kind == mir.OperandLiteral {
				// Named field format: [field1Name, field1Value, field2Name, field2Value, ...]
				// Skip first operand if it's a struct type (odd number of operands)
				startIndex := 0
				if len(inst.Operands)%2 == 1 {
					startIndex = 1 // First operand is struct type
				}

				// Process field-value pairs
				for i := startIndex; i < len(inst.Operands); i += 2 {
					if i+1 < len(inst.Operands) {
						fieldName := inst.Operands[i].Literal
						fieldValue := g.getOperandValue(inst.Operands[i+1])

						// Determine field type from the value operand
						// For now, assume all values are int (can be enhanced later)
						g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field(%s, \"%s\", %s);\n", varName, fieldName, fieldValue))
					}
				}
			} else {
				// Positional field format: [field1Value, field2Value, ...]
				// For now, just create empty struct (can be enhanced later with field names)
				g.output.WriteString(fmt.Sprintf("  // TODO: Handle positional struct initialization\n"))
			}
		}
	case "member":
		// Handle struct member access
		if len(inst.Operands) >= 2 {
			structVar := g.getOperandValue(inst.Operands[0])
			fieldName := inst.Operands[1].Literal
			varName := g.getVariableName(inst.ID)

			// For now, assume all fields are int (can be enhanced later)
			g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_int_field(%s, \"%s\");\n", varName, structVar, fieldName))
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
	case "malloc":
		// Handle dynamic memory allocation
		if len(inst.Operands) >= 1 {
			size := g.getOperandValue(inst.Operands[0])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  void* %s = malloc(%s);\n",
				varName, size))
		}
	case "free":
		// Handle dynamic memory deallocation
		if len(inst.Operands) >= 1 {
			ptr := g.getOperandValue(inst.Operands[0])
			g.output.WriteString(fmt.Sprintf("  free(%s);\n", ptr))
		}
	case "realloc":
		// Handle dynamic memory reallocation
		if len(inst.Operands) >= 2 {
			ptr := g.getOperandValue(inst.Operands[0])
			newSize := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  void* %s = realloc(%s, %s);\n",
				varName, ptr, newSize))
		}
	case "file.open":
		// Handle file opening
		if len(inst.Operands) >= 2 {
			filename := g.getOperandValue(inst.Operands[0])
			mode := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_file_open(%s, %s);\n",
				varName, filename, mode))
		}
	case "file.close":
		// Handle file closing
		if len(inst.Operands) >= 1 {
			fileHandle := g.getOperandValue(inst.Operands[0])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_file_close(%s);\n",
				varName, fileHandle))
		}
	case "file.read":
		// Handle file reading
		if len(inst.Operands) >= 3 {
			fileHandle := g.getOperandValue(inst.Operands[0])
			buffer := g.getOperandValue(inst.Operands[1])
			size := g.getOperandValue(inst.Operands[2])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_file_read(%s, %s, %s);\n",
				varName, fileHandle, buffer, size))
		}
	case "file.write":
		// Handle file writing
		if len(inst.Operands) >= 3 {
			fileHandle := g.getOperandValue(inst.Operands[0])
			buffer := g.getOperandValue(inst.Operands[1])
			size := g.getOperandValue(inst.Operands[2])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_file_write(%s, %s, %s);\n",
				varName, fileHandle, buffer, size))
		}
	case "file.seek":
		// Handle file seeking
		if len(inst.Operands) >= 3 {
			fileHandle := g.getOperandValue(inst.Operands[0])
			offset := g.getOperandValue(inst.Operands[1])
			whence := g.getOperandValue(inst.Operands[2])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_file_seek(%s, %s, %s);\n",
				varName, fileHandle, offset, whence))
		}
	case "file.tell":
		// Handle file position querying
		if len(inst.Operands) >= 1 {
			fileHandle := g.getOperandValue(inst.Operands[0])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_file_tell(%s);\n",
				varName, fileHandle))
		}
	case "file.exists":
		// Handle file existence checking
		if len(inst.Operands) >= 1 {
			filename := g.getOperandValue(inst.Operands[0])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_file_exists(%s);\n",
				varName, filename))
		}
	case "file.size":
		// Handle file size querying
		if len(inst.Operands) >= 1 {
			filename := g.getOperandValue(inst.Operands[0])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_file_size(%s);\n",
				varName, filename))
		}
	case "test.start":
		// Handle test start
		if len(inst.Operands) >= 1 {
			testName := g.getOperandValue(inst.Operands[0])
			g.output.WriteString(fmt.Sprintf("  omni_test_start(%s);\n", testName))
		}
	case "test.end":
		// Handle test end
		if len(inst.Operands) >= 2 {
			testName := g.getOperandValue(inst.Operands[0])
			passed := g.getOperandValue(inst.Operands[1])
			g.output.WriteString(fmt.Sprintf("  omni_test_end(%s, %s);\n", testName, passed))
		}
	case "assert":
		// Handle basic assertion
		if len(inst.Operands) >= 2 {
			condition := g.getOperandValue(inst.Operands[0])
			message := g.getOperandValue(inst.Operands[1])
			g.output.WriteString(fmt.Sprintf("  omni_assert(%s, %s);\n", condition, message))
		}
	case "assert.eq":
		// Handle equality assertion
		if len(inst.Operands) >= 3 {
			expected := g.getOperandValue(inst.Operands[0])
			actual := g.getOperandValue(inst.Operands[1])
			message := g.getOperandValue(inst.Operands[2])
			// Use appropriate assertion function based on instruction type
			switch inst.Type {
			case "int":
				g.output.WriteString(fmt.Sprintf("  omni_assert_eq_int(%s, %s, %s);\n", expected, actual, message))
			case "string":
				g.output.WriteString(fmt.Sprintf("  omni_assert_eq_string(%s, %s, %s);\n", expected, actual, message))
			case "float", "double":
				g.output.WriteString(fmt.Sprintf("  omni_assert_eq_float(%s, %s, %s);\n", expected, actual, message))
			default:
				// Default to int comparison
				g.output.WriteString(fmt.Sprintf("  omni_assert_eq_int(%s, %s, %s);\n", expected, actual, message))
			}
		}
	case "assert.true":
		// Handle true assertion
		if len(inst.Operands) >= 2 {
			condition := g.getOperandValue(inst.Operands[0])
			message := g.getOperandValue(inst.Operands[1])
			g.output.WriteString(fmt.Sprintf("  omni_assert_true(%s, %s);\n", condition, message))
		}
	case "assert.false":
		// Handle false assertion
		if len(inst.Operands) >= 2 {
			condition := g.getOperandValue(inst.Operands[0])
			message := g.getOperandValue(inst.Operands[1])
			g.output.WriteString(fmt.Sprintf("  omni_assert_false(%s, %s);\n", condition, message))
		}
	case "func.ref":
		// Handle function reference: getting a pointer to a function
		if len(inst.Operands) >= 1 {
			funcName := g.getOperandValue(inst.Operands[0])
			varName := g.getVariableName(inst.ID)
			// Map the function name (e.g., std.io.println -> omni_println)
			mappedName := g.mapFunctionName(funcName)
			// Generate C code to reference the function
			// The syntax is: returnType (*varName)(params) = funcName;
			g.output.WriteString(fmt.Sprintf("  %s = %s;\n",
				g.mapFunctionTypeWithName(inst.Type, varName), mappedName))
		}
	case "func.assign":
		// Handle function assignment: func_var = function_name
		if len(inst.Operands) >= 1 {
			funcName := g.getOperandValue(inst.Operands[0])
			varName := g.getVariableName(inst.ID)
			// Check if this is a function type
			if strings.Contains(inst.Type, ") -> ") {
				// Generate function pointer variable declaration
				g.output.WriteString(fmt.Sprintf("  %s = %s;\n",
					g.mapFunctionTypeWithName(inst.Type, varName), funcName))
			} else {
				// Regular variable assignment
				g.output.WriteString(fmt.Sprintf("  %s %s = %s;\n",
					g.mapType(inst.Type), varName, funcName))
			}
		}
	case "func.call":
		// Handle function call through function pointer
		if len(inst.Operands) >= 1 {
			funcPtr := g.getOperandValue(inst.Operands[0])
			varName := g.getVariableName(inst.ID)

			// Build function call with arguments
			g.output.WriteString(fmt.Sprintf("  %s %s = %s(",
				g.mapType(inst.Type), varName, funcPtr))

			// Add arguments
			for i, arg := range inst.Operands[1:] {
				if i > 0 {
					g.output.WriteString(", ")
				}
				g.output.WriteString(g.getOperandValue(arg))
			}
			g.output.WriteString(");\n")
		}
	case "closure.create":
		// Handle closure creation
		if len(inst.Operands) >= 1 {
			funcName := g.getOperandValue(inst.Operands[0])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  // TODO: Implement closure creation for %s\n", varName))
			g.output.WriteString(fmt.Sprintf("  // Function: %s\n", funcName))
			g.output.WriteString(fmt.Sprintf("  void* %s = NULL; // Placeholder for closure\n", varName))
		}
	case "closure.capture":
		// Handle closure variable capture
		if len(inst.Operands) >= 3 {
			closure := g.getOperandValue(inst.Operands[0])
			varName := inst.Operands[1].Literal
			varValue := g.getOperandValue(inst.Operands[2])
			g.output.WriteString(fmt.Sprintf("  // TODO: Implement closure capture for %s.%s = %s\n",
				closure, varName, varValue))
		}
	case "closure.bind":
		// Handle closure binding
		if len(inst.Operands) >= 1 {
			closure := g.getOperandValue(inst.Operands[0])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  // TODO: Implement closure binding for %s\n", varName))
			g.output.WriteString(fmt.Sprintf("  void* %s = %s; // Placeholder for bound closure\n", varName, closure))
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

// convertOperandToString converts an operand to a string representation for string concatenation
// It generates the necessary conversion code and returns the variable name to use
func (g *CGenerator) convertOperandToString(op mir.Operand) string {
	switch op.Kind {
	case mir.OperandValue:
		// For SSA values, use the type information from the operand
		varName := g.getVariableName(op.Value)

		// Check the operand type and convert if necessary
		if op.Type == "string" {
			// Already a string, return as is
			return varName
		} else if op.Type == "int" {
			// Convert int to string
			tempVar := fmt.Sprintf("temp_str_%d", op.Value)
			g.output.WriteString(fmt.Sprintf("  const char* %s = omni_int_to_string(%s);\n", tempVar, varName))
			return tempVar
		} else if op.Type == "float" || op.Type == "double" {
			// Convert float to string
			tempVar := fmt.Sprintf("temp_str_%d", op.Value)
			g.output.WriteString(fmt.Sprintf("  const char* %s = omni_float_to_string(%s);\n", tempVar, varName))
			return tempVar
		} else if op.Type == "bool" {
			// Convert bool to string
			tempVar := fmt.Sprintf("temp_str_%d", op.Value)
			g.output.WriteString(fmt.Sprintf("  const char* %s = omni_bool_to_string(%s);\n", tempVar, varName))
			return tempVar
		} else {
			// Default: assume int
			tempVar := fmt.Sprintf("temp_str_%d", op.Value)
			g.output.WriteString(fmt.Sprintf("  const char* %s = omni_int_to_string(%s);\n", tempVar, varName))
			return tempVar
		}

	case mir.OperandLiteral:
		// For literals, we need to convert based on the literal type
		if op.Literal == "true" {
			return "omni_bool_to_string(1)"
		} else if op.Literal == "false" {
			return "omni_bool_to_string(0)"
		} else if strings.HasPrefix(op.Literal, "\"") && strings.HasSuffix(op.Literal, "\"") {
			// String literal - return as is
			return op.Literal
		} else if strings.Contains(op.Literal, ".") {
			// Float literal - convert to string
			return fmt.Sprintf("omni_float_to_string(%s)", op.Literal)
		} else {
			// Integer literal - convert to string
			return fmt.Sprintf("omni_int_to_string(%s)", op.Literal)
		}
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
	// Handle function types: (param1, param2) -> returnType
	if strings.Contains(omniType, ") -> ") {
		return g.mapFunctionType(omniType)
	}

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

	// Handle map types: map<KeyType,ValueType>
	if strings.HasPrefix(omniType, "map<") && strings.HasSuffix(omniType, ">") {
		return "omni_map_t*"
	}

	// Handle struct types: struct<Field1Type,Field2Type,...>
	if strings.HasPrefix(omniType, "struct<") && strings.HasSuffix(omniType, ">") {
		return "omni_struct_t*"
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
	case "ptr":
		return "void*"
	default:
		return "int32_t" // Default fallback
	}
}

// mapFunctionType converts OmniLang function types to C function pointer types
func (g *CGenerator) mapFunctionType(omniType string) string {
	// Parse function type: (param1, param2) -> returnType
	arrowIndex := strings.Index(omniType, ") -> ")
	if arrowIndex == -1 {
		return "void*" // Fallback for malformed function types
	}

	paramPart := omniType[1:arrowIndex]                      // Remove opening (
	returnType := strings.TrimSpace(omniType[arrowIndex+5:]) // After " -> "

	// Parse parameter types
	var paramTypes []string
	if paramPart != "" {
		// Split by comma and trim spaces
		paramStrs := strings.Split(paramPart, ",")
		for _, paramStr := range paramStrs {
			paramType := strings.TrimSpace(paramStr)
			paramTypes = append(paramTypes, g.mapType(paramType))
		}
	}

	// Build C function pointer type: returnType (*)(param1, param2, ...)
	var cType strings.Builder
	cType.WriteString(g.mapType(returnType))
	cType.WriteString(" (*)(")
	for i, paramType := range paramTypes {
		if i > 0 {
			cType.WriteString(", ")
		}
		cType.WriteString(paramType)
	}
	cType.WriteString(")")

	return cType.String()
}

// mapFunctionTypeWithName converts OmniLang function types to C function pointer types with parameter name
func (g *CGenerator) mapFunctionTypeWithName(omniType string, paramName string) string {
	// Parse function type: (param1, param2) -> returnType
	arrowIndex := strings.Index(omniType, ") -> ")
	if arrowIndex == -1 {
		return "void* " + paramName // Fallback for malformed function types
	}

	paramPart := omniType[1:arrowIndex]                      // Remove opening (
	returnType := strings.TrimSpace(omniType[arrowIndex+5:]) // After " -> "

	// Parse parameter types
	var paramTypes []string
	if paramPart != "" {
		// Split by comma and trim spaces
		paramStrs := strings.Split(paramPart, ",")
		for _, paramStr := range paramStrs {
			paramType := strings.TrimSpace(paramStr)
			paramTypes = append(paramTypes, g.mapType(paramType))
		}
	}

	// Build C function pointer type with name: returnType (*paramName)(param1, param2, ...)
	var cType strings.Builder
	cType.WriteString(g.mapType(returnType))
	cType.WriteString(" (*")
	cType.WriteString(paramName)
	cType.WriteString(")(")
	for i, paramType := range paramTypes {
		if i > 0 {
			cType.WriteString(", ")
		}
		cType.WriteString(paramType)
	}
	cType.WriteString(")")

	return cType.String()
}

// mapFunctionReturnType converts OmniLang function types to C function pointer return types
func (g *CGenerator) mapFunctionReturnType(omniType string) string {
	// Parse function type: (param1, param2) -> returnType
	arrowIndex := strings.Index(omniType, ") -> ")
	if arrowIndex == -1 {
		return "void*" // Fallback for malformed function types
	}

	paramPart := omniType[1:arrowIndex]                      // Remove opening (
	returnType := strings.TrimSpace(omniType[arrowIndex+5:]) // After " -> "

	// Parse parameter types
	var paramTypes []string
	if paramPart != "" {
		// Split by comma and trim spaces
		paramStrs := strings.Split(paramPart, ",")
		for _, paramStr := range paramStrs {
			paramType := strings.TrimSpace(paramStr)
			paramTypes = append(paramTypes, g.mapType(paramType))
		}
	}

	// Build C function pointer return type: returnType (*)(param1, param2, ...)
	var cType strings.Builder
	cType.WriteString(g.mapType(returnType))
	cType.WriteString(" (*)(")
	for i, paramType := range paramTypes {
		if i > 0 {
			cType.WriteString(", ")
		}
		cType.WriteString(paramType)
	}
	cType.WriteString(")")

	return cType.String()
}

// generateFunctionPointerReturnType generates C function pointer return type with function name
func (g *CGenerator) generateFunctionPointerReturnType(omniType string, funcName string) string {
	// Parse function type: (param1, param2) -> returnType
	arrowIndex := strings.Index(omniType, ") -> ")
	if arrowIndex == -1 {
		return "void* " + funcName + "(" // Fallback for malformed function types
	}

	paramPart := omniType[1:arrowIndex]                      // Remove opening (
	returnType := strings.TrimSpace(omniType[arrowIndex+5:]) // After " -> "

	// Parse parameter types
	var paramTypes []string
	if paramPart != "" {
		// Split by comma and trim spaces
		paramStrs := strings.Split(paramPart, ",")
		for _, paramStr := range paramStrs {
			paramType := strings.TrimSpace(paramStr)
			paramTypes = append(paramTypes, g.mapType(paramType))
		}
	}

	// Build C function pointer return type: returnType (*funcName)(param1, param2, ...)
	var cType strings.Builder
	cType.WriteString(g.mapType(returnType))
	cType.WriteString(" (*")
	cType.WriteString(funcName)
	cType.WriteString(")(")
	for i, paramType := range paramTypes {
		if i > 0 {
			cType.WriteString(", ")
		}
		cType.WriteString(paramType)
	}
	cType.WriteString(")")

	return cType.String()
}

// generateCompleteFunctionSignature generates complete C function signature with function pointer return type
func (g *CGenerator) generateCompleteFunctionSignature(returnType string, funcName string, params []mir.Param) string {
	// Parse function type: (param1, param2) -> returnType
	arrowIndex := strings.Index(returnType, ") -> ")
	if arrowIndex == -1 {
		return "void* " + funcName + "(" // Fallback for malformed function types
	}

	paramPart := returnType[1:arrowIndex]                          // Remove opening (
	funcReturnType := strings.TrimSpace(returnType[arrowIndex+5:]) // After " -> "

	// Parse function pointer parameter types
	var funcParamTypes []string
	if paramPart != "" {
		// Split by comma and trim spaces
		paramStrs := strings.Split(paramPart, ",")
		for _, paramStr := range paramStrs {
			paramType := strings.TrimSpace(paramStr)
			funcParamTypes = append(funcParamTypes, g.mapType(paramType))
		}
	}

	// Build C function signature: returnType (*funcName(function_params))(function_pointer_params)
	var cType strings.Builder
	cType.WriteString(g.mapType(funcReturnType))
	cType.WriteString(" (*")
	cType.WriteString(funcName)
	cType.WriteString("(")

	// Add function parameters
	for i, param := range params {
		if i > 0 {
			cType.WriteString(", ")
		}
		// Check if this is a function pointer type
		if strings.Contains(param.Type, ") -> ") {
			// Generate function pointer parameter with correct C syntax
			cType.WriteString(g.mapFunctionTypeWithName(param.Type, param.Name))
		} else {
			paramType := g.mapType(param.Type)
			cType.WriteString(fmt.Sprintf("%s %s", paramType, param.Name))
		}
	}

	cType.WriteString("))(")

	// Add function pointer parameter types
	for i, paramType := range funcParamTypes {
		if i > 0 {
			cType.WriteString(", ")
		}
		cType.WriteString(paramType)
	}
	cType.WriteString(")")

	return cType.String()
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
	case "std.math.pow":
		return "omni_pow"
	case "std.math.sqrt":
		return "omni_sqrt"
	case "std.math.floor":
		return "omni_floor"
	case "std.math.ceil":
		return "omni_ceil"
	case "std.math.round":
		return "omni_round"
	case "std.math.gcd":
		return "omni_gcd"
	case "std.math.lcm":
		return "omni_lcm"
	case "std.math.factorial":
		return "omni_factorial"

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

	// String functions
	case "std.string.length":
		return "omni_strlen"
	case "std.string.concat":
		return "omni_strcat"
	case "std.string.substring":
		return "omni_substring"
	case "std.string.char_at":
		return "omni_char_at"
	case "std.string.starts_with":
		return "omni_starts_with"
	case "std.string.ends_with":
		return "omni_ends_with"
	case "std.string.contains":
		return "omni_contains"
	case "std.string.index_of":
		return "omni_index_of"
	case "std.string.last_index_of":
		return "omni_last_index_of"
	case "std.string.trim":
		return "omni_trim"
	case "std.string.to_upper":
		return "omni_to_upper"
	case "std.string.to_lower":
		return "omni_to_lower"
	case "std.string.equals":
		return "omni_string_equals"
	case "std.string.compare":
		return "omni_string_compare"

	// OS functions
	case "std.os.exit":
		return "omni_exit"

	// Utility functions
	case "std.assert":
		return "omni_assert"
	case "std.panic":
		return "omni_panic"
	case "std.int_to_string":
		return "omni_int_to_string"
	case "std.float_to_string":
		return "omni_float_to_string"
	case "std.bool_to_string":
		return "omni_bool_to_string"
	case "std.string_to_int":
		return "omni_string_to_int"
	case "std.string_to_float":
		return "omni_string_to_float"
	case "std.string_to_bool":
		return "omni_string_to_bool"
	// Array operations
	case "std.array.length":
		return "omni_array_length"
	case "std.array.get":
		return "omni_array_get_int"
	case "std.array.set":
		return "omni_array_set_int"
	case "std.test.start":
		return "omni_test_start"
	case "std.test.end":
		return "omni_test_end"
	case "std.assert.eq":
		return "omni_assert_eq"
	case "std.assert.true":
		return "omni_assert_true"
	case "std.assert.false":
		return "omni_assert_false"
	case "std.test.summary":
		return "omni_test_summary"
	case "std.malloc":
		return "omni_malloc"
	case "std.free":
		return "omni_free"
	case "std.realloc":
		return "omni_realloc"
	case "malloc":
		return "malloc"
	case "free":
		return "free"
	case "realloc":
		return "realloc"
	// File operations
	case "std.file.open":
		return "omni_file_open"
	case "std.file.close":
		return "omni_file_close"
	case "std.file.read":
		return "omni_file_read"
	case "std.file.write":
		return "omni_file_write"
	case "std.file.seek":
		return "omni_file_seek"
	case "std.file.tell":
		return "omni_file_tell"
	case "std.file.exists":
		return "omni_file_exists"
	case "std.file.size":
		return "omni_file_size"
	case "file.open":
		return "omni_file_open"
	case "file.close":
		return "omni_file_close"
	case "file.read":
		return "omni_file_read"
	case "file.write":
		return "omni_file_write"
	case "file.seek":
		return "omni_file_seek"
	case "file.tell":
		return "omni_file_tell"
	case "file.exists":
		return "omni_file_exists"
	case "file.size":
		return "omni_file_size"
	case "test.start":
		return "omni_test_start"
	case "test.end":
		return "omni_test_end"
	case "assert":
		return "omni_assert"
	case "assert.eq":
		return "omni_assert_eq"
	case "assert.true":
		return "omni_assert_true"
	case "assert.false":
		return "omni_assert_false"

	default:
		// For any other function names with dots, replace dots with underscores
		return strings.ReplaceAll(funcName, ".", "_")
	}
}

// isRuntimeProvidedFunction checks if a function is provided by the runtime
func (g *CGenerator) isRuntimeProvidedFunction(funcName string) bool {
	// List of functions that are provided by the runtime
	runtimeFunctions := map[string]bool{
		"std.io.print":             true,
		"std.io.println":           true,
		"std.io.print_int":         true,
		"std.io.println_int":       true,
		"std.io.print_float":       true,
		"std.io.println_float":     true,
		"std.io.print_bool":        true,
		"std.io.println_bool":      true,
		"io.print":                 true,
		"io.println":               true,
		"io.print_int":             true,
		"io.println_int":           true,
		"io.print_float":           true,
		"io.println_float":         true,
		"io.print_bool":            true,
		"io.println_bool":          true,
		"std.string.length":        true,
		"std.string.concat":        true,
		"std.string.substring":     true,
		"std.string.char_at":       true,
		"std.string.starts_with":   true,
		"std.string.ends_with":     true,
		"std.string.contains":      true,
		"std.string.index_of":      true,
		"std.string.last_index_of": true,
		"std.string.trim":          true,
		"std.string.to_upper":      true,
		"std.string.to_lower":      true,
		"std.string.equals":        true,
		"std.string.compare":       true,
		"string.length":            true,
		"string.concat":            true,
		"string.substring":         true,
		"string.char_at":           true,
		"string.starts_with":       true,
		"string.ends_with":         true,
		"string.contains":          true,
		"string.index_of":          true,
		"string.last_index_of":     true,
		"string.trim":              true,
		"string.to_upper":          true,
		"string.to_lower":          true,
		"string.equals":            true,
		"string.compare":           true,
		"std.math.abs":             true,
		"std.math.max":             true,
		"std.math.min":             true,
		"std.math.pow":             true,
		"std.math.sqrt":            true,
		"std.math.floor":           true,
		"std.math.ceil":            true,
		"std.math.round":           true,
		"std.math.gcd":             true,
		"std.math.lcm":             true,
		"std.math.factorial":       true,
		"math.abs":                 true,
		"math.max":                 true,
		"math.min":                 true,
		"math.pow":                 true,
		"math.sqrt":                true,
		"math.floor":               true,
		"math.ceil":                true,
		"math.round":               true,
		"math.gcd":                 true,
		"math.lcm":                 true,
		"math.factorial":           true,
		"math.toString":            true,
		"std.os.exit":              true,
		"os.exit":                  true,
		"os.getenv":                true,
		"os.setenv":                true,
		"os.unsetenv":              true,
		"os.getcwd":                true,
		"os.chdir":                 true,
		"os.mkdir":                 true,
		"os.rmdir":                 true,
		"os.exists":                true,
		"os.is_file":               true,
		"os.is_dir":                true,
		"os.remove":                true,
		"os.rename":                true,
		"os.copy":                  true,
		"os.read_file":             true,
		"os.write_file":            true,
		"os.append_file":           true,
		"array.length":             true,
		"array.get":                true,
		"array.set":                true,
		"array.append":             true,
		"array.prepend":            true,
		"array.insert":             true,
		"array.remove":             true,
		"array.contains":           true,
		"array.index_of":           true,
		"array.reverse":            true,
		"array.sort":               true,
		"array.slice":              true,
		"array.concat":             true,
		"array.fill":               true,
		"array.copy":               true,
		"collections.size":         true,
		"collections.get":          true,
		"collections.set":          true,
		"collections.has":          true,
		"collections.remove":       true,
		"collections.clear":        true,
		"collections.keys":         true,
		"collections.values":       true,
		"collections.entries":      true,
		"collections.merge":        true,
		// File operations
		"file.open":         true,
		"file.close":        true,
		"file.read":         true,
		"file.write":        true,
		"file.seek":         true,
		"file.tell":         true,
		"file.exists":       true,
		"file.size":         true,
		"std.int_to_string": true,
		"std.free":          true,
		// Skip std module utility functions that are not implemented in runtime
		"std.assert":          true,
		"std.assert_eq":       true,
		"std.panic":           true,
		"std.float_to_string": true,
		"std.bool_to_string":  true,
		"std.string_to_int":   true,
		"std.string_to_float": true,
		"std.string_to_bool":  true,
		"std.malloc":          true,
		"std.realloc":         true,
		// Array operations
		"std.array.length": true,
		"std.array.get":    true,
		"std.array.set":    true,
	}

	return runtimeFunctions[funcName]
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

// convertLiteralToDecimal converts hex and binary literals to decimal
func (g *CGenerator) convertLiteralToDecimal(literal string) string {
	if strings.HasPrefix(literal, "0x") || strings.HasPrefix(literal, "0X") {
		// Hex literal - convert to decimal
		hexStr := literal[2:]
		// Remove underscores
		hexStr = strings.ReplaceAll(hexStr, "_", "")
		// Convert to int64 and back to string
		if val, err := strconv.ParseInt(hexStr, 16, 64); err == nil {
			return strconv.FormatInt(val, 10)
		}
	} else if strings.HasPrefix(literal, "0b") || strings.HasPrefix(literal, "0B") {
		// Binary literal - convert to decimal
		binaryStr := literal[2:]
		// Remove underscores
		binaryStr = strings.ReplaceAll(binaryStr, "_", "")
		// Convert to int64 and back to string
		if val, err := strconv.ParseInt(binaryStr, 2, 64); err == nil {
			return strconv.FormatInt(val, 10)
		}
	}
	// Return as-is for regular decimal literals
	return literal
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
