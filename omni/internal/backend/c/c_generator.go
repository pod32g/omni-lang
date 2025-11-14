package cbackend

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/omni-lang/omni/internal/mir"
)

const inferTypePlaceholder = "<infer>"

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
	// Track discovered value types to help with conversions
	valueTypes map[mir.ValueID]string
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
		valueTypes:  make(map[mir.ValueID]string),
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
		valueTypes:  make(map[mir.ValueID]string),
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
		valueTypes:  make(map[mir.ValueID]string),
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
			returnType = "int32_t" // Always use int32_t for omni_main to match runtime
		}
		
		// For async functions (Promise<T>), the function should return the inner type, not Promise
		// The Promise wrapper is added when the function is called
		if strings.HasPrefix(fn.ReturnType, "Promise<") {
			innerType := fn.ReturnType[8 : len(fn.ReturnType)-1] // Remove "Promise<" and ">"
			returnType = g.mapType(innerType)
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
		returnType = "int32_t" // Always use int32_t for omni_main to match runtime
	}
	
	// For async functions (Promise<T>), the function should return the inner type, not Promise
	// The Promise wrapper is added when the function is called
	if strings.HasPrefix(fn.ReturnType, "Promise<") {
		innerType := fn.ReturnType[8 : len(fn.ReturnType)-1] // Remove "Promise<" and ">"
		returnType = g.mapType(innerType)
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

	// Reset maps for this function to avoid conflicts
	g.variables = make(map[mir.ValueID]string)
	g.phiVars = make(map[mir.ValueID]bool)
	g.mutableVars = make(map[mir.ValueID]bool)

	// Map parameter SSA values to their names
	for _, param := range fn.Params {
		g.variables[param.ID] = param.Name
	}

	// Collect all variables that need to be declared
	allVariables := make(map[mir.ValueID]string)
	for _, block := range fn.Blocks {
		for _, inst := range block.Instructions {
			if inst.ID != mir.InvalidValue {
				varName := fmt.Sprintf("v%d", int(inst.ID))
				allVariables[inst.ID] = varName
			}
		}
		// Terminators don't produce values, so we don't need to track them
	}

	// Declare all variables at the beginning of the function
	for id, varName := range allVariables {
		// Skip parameters (they're already declared)
		if _, isParam := g.variables[id]; !isParam {
			// Determine the type based on the instruction that produces this value
			var varType string
			var isArrayInit bool
			// Find the instruction that produces this value
			for _, block := range fn.Blocks {
				for _, inst := range block.Instructions {
					if inst.ID == id {
						// Special case for read_line() - always returns string
						if inst.Op == "call" || inst.Op == "call.string" {
							if len(inst.Operands) > 0 && inst.Operands[0].Kind == mir.OperandLiteral {
								funcName := inst.Operands[0].Literal
								if funcName == "std.io.read_line" || funcName == "io.read_line" {
									varType = "const char*"
									break
								}
							}
						}
						// Special case for await - always use string for async I/O operations
						if inst.Op == "await" {
							// For async I/O operations, the result is typically a string
							// Check the actual type from inst.Type, but default to string
							if inst.Type == "string" {
								varType = "const char*"
							} else if inst.Type == "int" {
								varType = "int32_t"
							} else if inst.Type == "bool" {
								varType = "int32_t"
							} else if inst.Type == "float" || inst.Type == "double" {
								varType = "double"
							} else {
								// Default to string for await (most common case for async I/O)
								varType = "const char*"
							}
							break
						}
						// Special case for index - check if indexing into struct array
						if inst.Op == "index" && len(inst.Operands) >= 2 {
							// Check the target array type
							targetOp := inst.Operands[0]
							if targetOp.Kind == mir.OperandValue {
								// Try to find the array type from the target variable
								// Look for array.init instruction that created the target
								for _, block := range fn.Blocks {
									for _, prevInst := range block.Instructions {
										if prevInst.ID == targetOp.Value {
											// Found the instruction that created the array
											if prevInst.Op == "array.init" {
												// Extract element type
												var elementTypeStr string
												if strings.HasPrefix(prevInst.Type, "array<") && strings.HasSuffix(prevInst.Type, ">") {
													elementTypeStr = prevInst.Type[6 : len(prevInst.Type)-1]
												} else if strings.HasPrefix(prevInst.Type, "[]<") && strings.HasSuffix(prevInst.Type, ">") {
													elementTypeStr = prevInst.Type[3 : len(prevInst.Type)-1]
												}
												// Check if element type is a struct
												if elementTypeStr != "" && !g.isPrimitiveType(elementTypeStr) && !strings.Contains(elementTypeStr, "<") && !strings.Contains(elementTypeStr, "(") {
													varType = "omni_struct_t*"
													break
												}
											}
										}
									}
								}
							}
							// If we didn't find it, use the instruction type
							if varType == "" {
								resultType := inst.Type
								isStruct := !g.isPrimitiveType(resultType) && !strings.Contains(resultType, "<") && !strings.Contains(resultType, "(")
								if isStruct {
									varType = "omni_struct_t*"
								} else {
									varType = g.mapType(inst.Type)
								}
							}
							break
						}
						// Special case for phi - check if it's a struct type (for struct array iteration)
						if inst.Op == "phi" {
							resultType := inst.Type
							isStruct := !g.isPrimitiveType(resultType) && !strings.Contains(resultType, "<") && !strings.Contains(resultType, "(")
							if isStruct {
								varType = "omni_struct_t*"
								break
							}
						}
						varType = g.mapType(inst.Type)
						// Check if this is an array.init instruction
						if inst.Op == "array.init" {
							isArrayInit = true
							// For struct arrays, the type should be omni_struct_t*
							var elementTypeStr string
							if strings.HasPrefix(inst.Type, "array<") && strings.HasSuffix(inst.Type, ">") {
								elementTypeStr = inst.Type[6 : len(inst.Type)-1]
							} else if strings.HasPrefix(inst.Type, "[]<") && strings.HasSuffix(inst.Type, ">") {
								elementTypeStr = inst.Type[3 : len(inst.Type)-1]
							} else {
								elementTypeStr = inst.Type
							}
							// Check if element type is a struct
							if !g.isPrimitiveType(elementTypeStr) && !strings.Contains(elementTypeStr, "<") && !strings.Contains(elementTypeStr, "(") {
								varType = "omni_struct_t*"
							}
						}
						// Check if this is a struct.init instruction
						if inst.Op == "struct.init" {
							varType = "omni_struct_t*"
						}
						break
					}
				}
				if varType != "" {
					break
				}
			}
			if varType == "" {
				varType = "int32_t" // default type
			}
			// Skip declaring void variables (they don't produce values)
			// Skip declaring array variables (they're declared in array.init)
			// For string constants, we'll initialize them in the const instruction
			if varType != "void" && !isArrayInit {
				// Check if this is a const instruction with a string literal
				isStringConst := false
				for _, block := range fn.Blocks {
					for _, inst := range block.Instructions {
						if inst.ID == id && inst.Op == "const" && len(inst.Operands) > 0 {
							// Check if it's a string literal (either by type or by literal value)
							isString := inst.Type == "string" || 
								(inst.Operands[0].Kind == mir.OperandLiteral && 
								 strings.HasPrefix(inst.Operands[0].Literal, "\"") && 
								 strings.HasSuffix(inst.Operands[0].Literal, "\""))
							if isString && inst.Operands[0].Kind == mir.OperandLiteral {
								isStringConst = true
								// Initialize string constant at declaration
								literalValue := g.getOperandValue(inst.Operands[0])
								g.output.WriteString(fmt.Sprintf("  %s %s = %s;\n", varType, varName, literalValue))
								break
							}
						}
					}
					if isStringConst {
						break
					}
				}
				if !isStringConst {
					g.output.WriteString(fmt.Sprintf("  %s %s;\n", varType, varName))
				}
			}
		}
	}

	// Pre-populate valueTypes for ALL blocks in this function
	// This ensures type information is available when processing struct.init
	// Also handle const instructions specially to infer types from literals
	for _, block := range fn.Blocks {
		for _, inst := range block.Instructions {
			if inst.ID != mir.InvalidValue {
				// Always store the type if it's set, even for const instructions
				if inst.Type != "" && inst.Type != inferTypePlaceholder {
					g.valueTypes[inst.ID] = inst.Type
				} else if inst.Op == "const" && len(inst.Operands) > 0 {
					// For const instructions, infer type from the literal if Type is not set
					// This is important because const instructions should have their type set
					if inst.Operands[0].Type != "" && inst.Operands[0].Type != inferTypePlaceholder {
						g.valueTypes[inst.ID] = inst.Operands[0].Type
					} else if inst.Operands[0].Kind == mir.OperandLiteral {
						lit := inst.Operands[0].Literal
						if strings.HasPrefix(lit, "\"") && strings.HasSuffix(lit, "\"") {
							g.valueTypes[inst.ID] = "string"
						} else if lit == "true" || lit == "false" {
							g.valueTypes[inst.ID] = "bool"
						} else if strings.Contains(lit, ".") {
							g.valueTypes[inst.ID] = "float"
						} else {
							g.valueTypes[inst.ID] = "int"
						}
					}
				}
			}
		}
	}

	// Generate function body
	for _, block := range fn.Blocks {
		if err := g.generateBlock(block, fn); err != nil {
			return err
		}
	}

	g.output.WriteString("}\n\n")
	return nil
}

// generateBlock generates C code for a basic block
func (g *CGenerator) generateBlock(block *mir.BasicBlock, fn *mir.Function) error {
	funcName := fn.Name
	if funcName == "main" {
		funcName = "omni_main"
	}
	// Generate block label if it's not the entry block
	if block.Name != "entry" {
		g.output.WriteString(fmt.Sprintf("  %s:\n", block.Name))
	}

	// Pre-populate valueTypes for this block's instructions
	// This ensures type information is available when processing struct.init
	// Also handle const instructions specially to infer types from literals
	for _, inst := range block.Instructions {
		if inst.ID != mir.InvalidValue {
			if inst.Type != "" && inst.Type != inferTypePlaceholder {
				g.valueTypes[inst.ID] = inst.Type
			} else if inst.Op == "const" && len(inst.Operands) > 0 {
				// For const instructions, infer type from the literal if Type is not set
				if inst.Operands[0].Type != "" && inst.Operands[0].Type != inferTypePlaceholder {
					g.valueTypes[inst.ID] = inst.Operands[0].Type
				} else if inst.Operands[0].Kind == mir.OperandLiteral {
					lit := inst.Operands[0].Literal
					if strings.HasPrefix(lit, "\"") && strings.HasSuffix(lit, "\"") {
						g.valueTypes[inst.ID] = "string"
					} else if lit == "true" || lit == "false" {
						g.valueTypes[inst.ID] = "bool"
					} else if strings.Contains(lit, ".") {
						g.valueTypes[inst.ID] = "float"
					} else {
						g.valueTypes[inst.ID] = "int"
					}
				}
			}
		}
	}

	// Generate instructions
	for _, inst := range block.Instructions {
		if err := g.generateInstruction(&inst); err != nil {
			return err
		}
	}

	// Generate terminator
	if err := g.generateTerminator(&block.Terminator, funcName); err != nil {
		return err
	}

	return nil
}

// generateInstruction generates C code for a single instruction
func (g *CGenerator) generateInstruction(inst *mir.Instruction) error {
	if inst.ID != mir.InvalidValue && inst.Type != "" && inst.Type != inferTypePlaceholder {
		g.valueTypes[inst.ID] = inst.Type
	}

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
				// Assign to already declared variable
				g.output.WriteString(fmt.Sprintf("  %s = %s;\n",
					varName, convertedValue))
				// Mark variables that are initialized to 0 as potentially mutable (like sum variables)
				if convertedValue == "0" {
					g.mutableVars[inst.ID] = true
				}
			case "float", "double":
				// Assign to already declared variable
				g.output.WriteString(fmt.Sprintf("  %s = %s;\n",
					varName, literalValue))
			case "string":
				// String constants are initialized at declaration, so skip assignment here
				// (This avoids duplicate initialization)
			case "bool":
				// Assign to already declared variable
				g.output.WriteString(fmt.Sprintf("  %s = %s;\n",
					varName, literalValue))
			case "null":
				// Assign to already declared variable
				g.output.WriteString(fmt.Sprintf("  %s = NULL;\n",
					varName))
			default:
				// Check if this is a function type
				if strings.Contains(inst.Type, ") -> ") {
					// Function type - assign function pointer to already declared variable
					g.output.WriteString(fmt.Sprintf("  %s = %s;\n",
						varName, literalValue))
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
				// Assign to already declared variable
				g.output.WriteString(fmt.Sprintf("  %s = %s + %s;\n",
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
				// Regular subtraction - assign to already declared variable
				g.output.WriteString(fmt.Sprintf("  %s = %s - %s;\n",
					varName, left, right))
			}
		}
	case "mul":
		// Handle multiplication
		if len(inst.Operands) >= 2 {
			left := g.getOperandValue(inst.Operands[0])
			right := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)
			g.output.WriteString(fmt.Sprintf("  %s = %s * %s;\n",
				varName, left, right))
		}
	case "div":
		// Handle division
		if len(inst.Operands) >= 2 {
			left := g.getOperandValue(inst.Operands[0])
			right := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)
			// Division - assign to already declared variable
			g.output.WriteString(fmt.Sprintf("  %s = %s / %s;\n",
				varName, left, right))
		}
	case "mod":
		// Handle modulo
		if len(inst.Operands) >= 2 {
			left := g.getOperandValue(inst.Operands[0])
			right := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)
			// Modulo - assign to already declared variable
			g.output.WriteString(fmt.Sprintf("  %s = %s %% %s;\n",
				varName, left, right))
		}
	case "bitand":
		// Handle bitwise AND
		if len(inst.Operands) >= 2 {
			left := g.getOperandValue(inst.Operands[0])
			right := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)
			// Bitwise AND - assign to already declared variable
			g.output.WriteString(fmt.Sprintf("  %s = %s & %s;\n",
				varName, left, right))
		}
	case "bitor":
		// Handle bitwise OR
		if len(inst.Operands) >= 2 {
			left := g.getOperandValue(inst.Operands[0])
			right := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)
			// Bitwise OR - assign to already declared variable
			g.output.WriteString(fmt.Sprintf("  %s = %s | %s;\n",
				varName, left, right))
		}
	case "bitxor":
		// Handle bitwise XOR
		if len(inst.Operands) >= 2 {
			left := g.getOperandValue(inst.Operands[0])
			right := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)
			// Bitwise XOR - assign to already declared variable
			g.output.WriteString(fmt.Sprintf("  %s = %s ^ %s;\n",
				varName, left, right))
		}
	case "lshift":
		// Handle left shift
		if len(inst.Operands) >= 2 {
			left := g.getOperandValue(inst.Operands[0])
			right := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)
			// Left shift - assign to already declared variable
			g.output.WriteString(fmt.Sprintf("  %s = %s << %s;\n",
				varName, left, right))
		}
	case "rshift":
		// Handle right shift
		if len(inst.Operands) >= 2 {
			left := g.getOperandValue(inst.Operands[0])
			right := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)
			// Right shift - assign to already declared variable
			g.output.WriteString(fmt.Sprintf("  %s = %s >> %s;\n",
				varName, left, right))
		}
	case "strcat":
		// Handle string concatenation
		if len(inst.Operands) >= 2 {
			varName := g.getVariableName(inst.ID)

			// Convert operands to strings if needed
			leftStr := g.convertOperandToString(inst.Operands[0])
			rightStr := g.convertOperandToString(inst.Operands[1])

			// Generate proper string concatenation using runtime function
			g.output.WriteString(fmt.Sprintf("  %s = omni_strcat(%s, %s);\n",
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
			// Negation - assign to already declared variable
			g.output.WriteString(fmt.Sprintf("  %s = -%s;\n",
				varName, operand))
		}
	case "not":
		// Handle logical not
		if len(inst.Operands) >= 1 {
			operand := g.getOperandValue(inst.Operands[0])
			varName := g.getVariableName(inst.ID)
			// Logical NOT - assign to already declared variable
			g.output.WriteString(fmt.Sprintf("  %s = !%s;\n",
				varName, operand))
		}
	case "bitnot":
		// Handle bitwise not
		if len(inst.Operands) >= 1 {
			operand := g.getOperandValue(inst.Operands[0])
			varName := g.getVariableName(inst.ID)
			// Bitwise NOT - assign to already declared variable
			g.output.WriteString(fmt.Sprintf("  %s = ~%s;\n",
				varName, operand))
		}
	case "cast":
		// Handle type cast
		if len(inst.Operands) >= 1 {
			operand := g.getOperandValue(inst.Operands[0])
			varName := g.getVariableName(inst.ID)
			targetType := g.mapType(inst.Type)
			// Type cast - assign to already declared variable
			g.output.WriteString(fmt.Sprintf("  %s = (%s)%s;\n",
				varName, targetType, operand))
		}
	case "and":
		// Handle logical and
		if len(inst.Operands) >= 2 {
			left := g.getOperandValue(inst.Operands[0])
			right := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)
			// Logical AND - assign to already declared variable
			g.output.WriteString(fmt.Sprintf("  %s = %s && %s;\n",
				varName, left, right))
		}
	case "or":
		// Handle logical or
		if len(inst.Operands) >= 2 {
			left := g.getOperandValue(inst.Operands[0])
			right := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)
			// Logical OR - assign to already declared variable
			g.output.WriteString(fmt.Sprintf("  %s = %s || %s;\n",
				varName, left, right))
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
				// Array length - assign to already declared variable
				g.output.WriteString(fmt.Sprintf("  %s = sizeof(%s) / sizeof(%s[0]);\n",
					varName, arrayVar, arrayVar))
				return nil
			}

			// Special-case std.io print helpers so we can perform type conversion.
			if (funcName == "std.io.print" || funcName == "io.print") && len(inst.Operands) >= 2 {
				g.emitPrint(inst.Operands[1], false)
				return nil
			}
			if funcName == "std.io.println" || funcName == "io.println" {
				if len(inst.Operands) >= 2 {
					g.emitPrint(inst.Operands[1], true)
				} else {
					g.output.WriteString("  omni_println_string(\"\");\n")
				}
				return nil
			}

			// Special-case std.io.read_line to ensure result is assigned
			if funcName == "std.io.read_line" || funcName == "io.read_line" {
				// read_line() always returns a string, so assign it to a variable
				if inst.ID != mir.InvalidValue {
					varName := g.getVariableName(inst.ID)
					g.output.WriteString(fmt.Sprintf("  %s = omni_read_line();\n", varName))
					// Record the type for this variable
					g.valueTypes[inst.ID] = "string"
				} else {
					// If no ID, just call it (shouldn't happen, but handle gracefully)
					g.output.WriteString("  omni_read_line();\n")
				}
				return nil
			}
			
			// Special-case async I/O functions - they return Promise<T>
			if funcName == "std.io.read_line_async" || funcName == "io.read_line_async" {
				if inst.ID != mir.InvalidValue {
					varName := g.getVariableName(inst.ID)
					tempVar := fmt.Sprintf("_temp_%d", inst.ID)
					g.output.WriteString(fmt.Sprintf("  const char* %s = omni_read_line();\n", tempVar))
					g.output.WriteString(fmt.Sprintf("  omni_promise_t* %s = omni_promise_create_string(%s);\n", varName, tempVar))
					g.valueTypes[inst.ID] = "Promise<string>"
				}
				return nil
			}
			
			if funcName == "std.os.read_file_async" || funcName == "os.read_file_async" {
				if inst.ID != mir.InvalidValue && len(inst.Operands) >= 2 {
					varName := g.getVariableName(inst.ID)
					pathVar := g.getOperandValue(inst.Operands[1])
					tempVar := fmt.Sprintf("_temp_%d", inst.ID)
					// Declare temp variable first
					g.output.WriteString(fmt.Sprintf("  const char* %s = omni_read_file(%s);\n", tempVar, pathVar))
					g.output.WriteString(fmt.Sprintf("  omni_promise_t* %s = omni_promise_create_string(%s);\n", varName, tempVar))
					g.valueTypes[inst.ID] = "Promise<string>"
				}
				return nil
			}
			
			if funcName == "std.os.write_file_async" || funcName == "os.write_file_async" {
				if inst.ID != mir.InvalidValue && len(inst.Operands) >= 3 {
					varName := g.getVariableName(inst.ID)
					pathVar := g.getOperandValue(inst.Operands[1])
					contentVar := g.getOperandValue(inst.Operands[2])
					tempVar := fmt.Sprintf("_temp_%d", inst.ID)
					g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_write_file(%s, %s);\n", tempVar, pathVar, contentVar))
					g.output.WriteString(fmt.Sprintf("  omni_promise_t* %s = omni_promise_create_bool(%s != 0);\n", varName, tempVar))
					g.valueTypes[inst.ID] = "Promise<bool>"
				}
				return nil
			}
			
			if funcName == "std.os.append_file_async" || funcName == "os.append_file_async" {
				if inst.ID != mir.InvalidValue && len(inst.Operands) >= 3 {
					varName := g.getVariableName(inst.ID)
					pathVar := g.getOperandValue(inst.Operands[1])
					contentVar := g.getOperandValue(inst.Operands[2])
					tempVar := fmt.Sprintf("_temp_%d", inst.ID)
					g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_append_file(%s, %s);\n", tempVar, pathVar, contentVar))
					g.output.WriteString(fmt.Sprintf("  omni_promise_t* %s = omni_promise_create_bool(%s != 0);\n", varName, tempVar))
					g.valueTypes[inst.ID] = "Promise<bool>"
				}
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
				// Check if the result type is a Promise (async function)
				if strings.HasPrefix(inst.Type, "Promise<") {
					// Extract inner type from Promise<T>
					innerType := inst.Type[8 : len(inst.Type)-1] // Remove "Promise<" and ">"
					// Call the function and wrap result in Promise
					tempVar := fmt.Sprintf("_temp_%d", inst.ID)
					g.output.WriteString(fmt.Sprintf("  %s %s = %s(",
						g.mapType(innerType), tempVar, cFuncName))
					// Add arguments
					for i, arg := range inst.Operands[1:] {
						if i > 0 {
							g.output.WriteString(", ")
						}
						g.output.WriteString(g.getOperandValue(arg))
					}
					g.output.WriteString(");\n")
					// Wrap the result in a Promise based on inner type
					switch innerType {
					case "int":
						g.output.WriteString(fmt.Sprintf("  %s = omni_promise_create_int(%s);\n", varName, tempVar))
					case "string":
						g.output.WriteString(fmt.Sprintf("  %s = omni_promise_create_string(%s);\n", varName, tempVar))
					case "float", "double":
						g.output.WriteString(fmt.Sprintf("  %s = omni_promise_create_float(%s);\n", varName, tempVar))
					case "bool":
						g.output.WriteString(fmt.Sprintf("  %s = omni_promise_create_bool(%s);\n", varName, tempVar))
					default:
						// Default to int
						g.output.WriteString(fmt.Sprintf("  %s = omni_promise_create_int(%s);\n", varName, tempVar))
					}
				} else if strings.Contains(inst.Type, ") -> ") {
					// Generate function pointer variable declaration
					g.output.WriteString(fmt.Sprintf("  %s = %s(",
						g.mapFunctionTypeWithName(inst.Type, varName), cFuncName))
					// Add arguments
					for i, arg := range inst.Operands[1:] {
						if i > 0 {
							g.output.WriteString(", ")
						}
						g.output.WriteString(g.getOperandValue(arg))
					}
					g.output.WriteString(");\n")
				} else {
					// Regular function call - assign to already declared variable
					g.output.WriteString(fmt.Sprintf("  %s = %s(",
						varName, cFuncName))
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
		}
	case "index":
		// Handle array/map indexing
		if len(inst.Operands) >= 2 {
			target := g.getOperandValue(inst.Operands[0])
			index := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)

			// Check if target is a map variable
			if g.isMapVariable(target) {
				// Map indexing - assign to already declared variable
				g.output.WriteString(fmt.Sprintf("  %s = omni_map_get_string_int(%s, %s);\n", varName, target, index))
			} else {
				// Array indexing - assign to already declared variable
				// Check if result type is a struct (array of structs)
				resultType := inst.Type
				if storedType, ok := g.valueTypes[inst.ID]; ok && storedType != "" {
					resultType = storedType
				}
				isStruct := !g.isPrimitiveType(resultType) && !strings.Contains(resultType, "<") && !strings.Contains(resultType, "(")
				
				if isStruct {
					// For struct arrays, the element is already a pointer, so just index
					g.output.WriteString(fmt.Sprintf("  %s = %s[%s];\n", varName, target, index))
					// Store the struct type
					g.valueTypes[inst.ID] = resultType
				} else {
					// For primitive arrays, simple indexing
					g.output.WriteString(fmt.Sprintf("  %s = %s[%s];\n", varName, target, index))
				}
			}
		}
	case "array.init":
		// Handle array literal initialization
		if len(inst.Operands) > 0 {
			varName := g.getVariableName(inst.ID)
			// Extract element type from array type
			var elementTypeStr string
			if strings.HasPrefix(inst.Type, "array<") && strings.HasSuffix(inst.Type, ">") {
				elementTypeStr = inst.Type[6 : len(inst.Type)-1] // Extract "int" from "array<int>"
			} else if strings.HasPrefix(inst.Type, "[]<") && strings.HasSuffix(inst.Type, ">") {
				elementTypeStr = inst.Type[3 : len(inst.Type)-1] // Extract "int" from "[]<int>"
			} else {
				elementTypeStr = inst.Type // fallback
			}
			
			// Check if element type is a struct (not a primitive)
			isStruct := !g.isPrimitiveType(elementTypeStr) && !strings.Contains(elementTypeStr, "<") && !strings.Contains(elementTypeStr, "(")
			
			if isStruct {
				// For struct arrays, create array of pointers
				elementType := "omni_struct_t*"
				g.output.WriteString(fmt.Sprintf("  %s %s[%d];\n", elementType, varName, len(inst.Operands)))
				// Initialize each struct element
				for i, op := range inst.Operands {
					// Each operand should be a struct.init instruction result
					// Get the variable name for this struct
					structVar := g.getOperandValue(op)
					g.output.WriteString(fmt.Sprintf("  %s[%d] = %s;\n", varName, i, structVar))
				}
			} else {
				// For primitive types, create simple array
				elementType := g.mapType(elementTypeStr) // Map "int" to "int32_t"
				g.output.WriteString(fmt.Sprintf("  %s %s[] = {", elementType, varName))
				for i, op := range inst.Operands {
					if i > 0 {
						g.output.WriteString(", ")
					}
					g.output.WriteString(g.getOperandValue(op))
				}
				g.output.WriteString("};\n")
			}
		}
	case "map.init":
		// Handle map initialization
		varName := g.getVariableName(inst.ID)
		// Assign to already declared variable
		g.output.WriteString(fmt.Sprintf("  %s = omni_map_create();\n", varName))

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
		// Assign to already declared variable
		g.output.WriteString(fmt.Sprintf("  %s = omni_struct_create();\n", varName))

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
						fieldValueOp := inst.Operands[i+1]
						fieldValue := g.getOperandValue(fieldValueOp)

						// Determine field type from the value operand type
						// Priority: 1) operand Type field (set by MIR builder - most reliable), 2) stored valueTypes, 3) lookup from module
						fieldType := ""
						
						// First, check the operand's Type field (set by MIR builder via valueOperand)
						// This is the most reliable source since it's set directly by the MIR builder
						// The MIR builder sets this in valueOperand(value.ID, value.Type) where value.Type should be "string"
						if fieldValueOp.Type != "" && fieldValueOp.Type != inferTypePlaceholder && fieldValueOp.Type != "<infer>" {
							fieldType = fieldValueOp.Type
							// Store it for future use
							if fieldValueOp.Kind == mir.OperandValue {
								g.valueTypes[fieldValueOp.Value] = fieldType
							}
						}
						
						// Also check stored types (should have been set when we pre-populated)
						if fieldType == "" && fieldValueOp.Kind == mir.OperandValue {
							if storedType, ok := g.valueTypes[fieldValueOp.Value]; ok && storedType != "" && storedType != inferTypePlaceholder {
								fieldType = storedType
							}
						}
						
						// If still not found, try to look it up from the module's functions
						// Search in reverse order (most recent functions first) to find the instruction
						if fieldType == "" && fieldValueOp.Kind == mir.OperandValue && g.module != nil {
							// Search all functions and blocks to find the instruction that produces this value
							for i := len(g.module.Functions) - 1; i >= 0; i-- {
								fn := g.module.Functions[i]
								for _, block := range fn.Blocks {
									for _, inst := range block.Instructions {
										if inst.ID == fieldValueOp.Value {
											// Found the instruction - use its type
											// First check the instruction's Type field (most reliable)
											if inst.Type != "" && inst.Type != inferTypePlaceholder {
												fieldType = inst.Type
												g.valueTypes[fieldValueOp.Value] = inst.Type
												goto foundTypeInLookup
											} else if inst.Op == "const" && len(inst.Operands) > 0 {
												// For const instructions, check the operand type or infer from literal
												if inst.Operands[0].Type != "" && inst.Operands[0].Type != inferTypePlaceholder {
													fieldType = inst.Operands[0].Type
													g.valueTypes[fieldValueOp.Value] = fieldType
													goto foundTypeInLookup
												} else if inst.Operands[0].Kind == mir.OperandLiteral {
													// Infer type from literal value
													lit := inst.Operands[0].Literal
													if strings.HasPrefix(lit, "\"") && strings.HasSuffix(lit, "\"") {
														fieldType = "string"
														g.valueTypes[fieldValueOp.Value] = "string"
													} else if lit == "true" || lit == "false" {
														fieldType = "bool"
														g.valueTypes[fieldValueOp.Value] = "bool"
													} else if strings.Contains(lit, ".") {
														fieldType = "float"
														g.valueTypes[fieldValueOp.Value] = "float"
													} else {
														fieldType = "int"
														g.valueTypes[fieldValueOp.Value] = "int"
													}
													if fieldType != "" {
														goto foundTypeInLookup
													}
												}
											}
										}
									}
								}
							}
						foundTypeInLookup:
						}
						
						// Check if it's a literal operand (string literals)
						if fieldValueOp.Kind == mir.OperandLiteral {
							// Check if it's a string literal
							if strings.HasPrefix(fieldValueOp.Literal, "\"") && strings.HasSuffix(fieldValueOp.Literal, "\"") {
								fieldType = "string"
							}
						}
						
						// Last resort: if we still don't have a type and it's a value operand,
						// try to find the const instruction and check its literal or type
						if (fieldType == "" || fieldType == "<inferred>") && fieldValueOp.Kind == mir.OperandValue && g.module != nil {
							for _, fn := range g.module.Functions {
								for _, block := range fn.Blocks {
									for _, inst := range block.Instructions {
										if inst.ID == fieldValueOp.Value {
											// Found the instruction that produces this value
											// First, always check the instruction's Type field (most reliable)
											// For const instructions, this should be "string", "int", "bool", etc.
											// The MIR builder sets inst.Type to "string" for string literals
											if inst.Type != "" && inst.Type != inferTypePlaceholder {
												fieldType = inst.Type
												g.valueTypes[fieldValueOp.Value] = fieldType
												goto foundType
											}
											// If it's a const instruction and Type is not set, check the operand
											if inst.Op == "const" && len(inst.Operands) > 0 {
												// Check the operand's type
												if inst.Operands[0].Type != "" && inst.Operands[0].Type != inferTypePlaceholder {
													fieldType = inst.Operands[0].Type
													g.valueTypes[fieldValueOp.Value] = fieldType
													goto foundType
												} else if inst.Operands[0].Kind == mir.OperandLiteral {
													// Infer from literal value - this is the fallback
													lit := inst.Operands[0].Literal
													if strings.HasPrefix(lit, "\"") && strings.HasSuffix(lit, "\"") {
														fieldType = "string"
														g.valueTypes[fieldValueOp.Value] = "string"
													} else if lit == "true" || lit == "false" {
														fieldType = "bool"
														g.valueTypes[fieldValueOp.Value] = "bool"
													} else if strings.Contains(lit, ".") {
														fieldType = "float"
														g.valueTypes[fieldValueOp.Value] = "float"
													} else {
														fieldType = "int"
														g.valueTypes[fieldValueOp.Value] = "int"
													}
													goto foundType
												}
											} else if inst.Type != "" && inst.Type != inferTypePlaceholder {
												// For non-const instructions, use the instruction's type
												fieldType = inst.Type
												g.valueTypes[fieldValueOp.Value] = fieldType
												goto foundType
											}
										}
									}
								}
							}
						foundType:
						}
						
						// Before defaulting to int, do one final check - maybe the type wasn't found
						// but we can still infer it from the const instruction's literal
						if (fieldType == "" || fieldType == "<inferred>") && fieldValueOp.Kind == mir.OperandValue && g.module != nil {
							// Try one more time to find the const instruction and check its literal
							for _, fn := range g.module.Functions {
								for _, block := range fn.Blocks {
									for _, inst := range block.Instructions {
										if inst.ID == fieldValueOp.Value && inst.Op == "const" && len(inst.Operands) > 0 {
											if inst.Operands[0].Kind == mir.OperandLiteral {
												lit := inst.Operands[0].Literal
												if strings.HasPrefix(lit, "\"") && strings.HasSuffix(lit, "\"") {
													fieldType = "string"
													g.valueTypes[fieldValueOp.Value] = "string"
													goto finalTypeFound
												} else if lit == "true" || lit == "false" {
													fieldType = "bool"
													g.valueTypes[fieldValueOp.Value] = "bool"
													goto finalTypeFound
												} else if strings.Contains(lit, ".") {
													fieldType = "float"
													g.valueTypes[fieldValueOp.Value] = "float"
													goto finalTypeFound
												} else {
													// For numeric literals, default to int
													fieldType = "int"
													g.valueTypes[fieldValueOp.Value] = "int"
													goto finalTypeFound
												}
											}
										}
									}
								}
							}
						finalTypeFound:
						}
						
						// Default to int if still no type
						if fieldType == "" || fieldType == "<inferred>" {
							fieldType = "int"
						}
						
						// Use appropriate setter based on field type
						switch fieldType {
						case "string":
							// If the field value is a literal string, use it directly
							// Otherwise, use the variable name
							actualValue := fieldValue
							if fieldValueOp.Kind == mir.OperandLiteral && strings.HasPrefix(fieldValueOp.Literal, "\"") && strings.HasSuffix(fieldValueOp.Literal, "\"") {
								actualValue = fieldValueOp.Literal
							}
							g.output.WriteString(fmt.Sprintf("  omni_struct_set_string_field(%s, \"%s\", %s);\n", varName, fieldName, actualValue))
						case "float", "double":
							g.output.WriteString(fmt.Sprintf("  omni_struct_set_float_field(%s, \"%s\", %s);\n", varName, fieldName, fieldValue))
						case "bool":
							g.output.WriteString(fmt.Sprintf("  omni_struct_set_bool_field(%s, \"%s\", %s);\n", varName, fieldName, fieldValue))
						default:
							// Default to int
							g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field(%s, \"%s\", %s);\n", varName, fieldName, fieldValue))
						}
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

			// Determine field type from instruction type
			fieldType := inst.Type
			if fieldType == "" || fieldType == "<inferred>" {
				// Try to look up the type from stored valueTypes
				// This should have been set when the struct field was accessed
				if len(inst.Operands) > 0 && inst.Operands[0].Kind == mir.OperandValue {
					// Check if we have type information for this struct variable
					if storedType, ok := g.valueTypes[inst.Operands[0].Value]; ok && storedType != "" && storedType != inferTypePlaceholder {
						// This is the struct type, not the field type
						// We need to look up the field type from the struct definition
						// For now, try to infer from the instruction's type or default
					}
				}
				// If still not found, try to look up from the module to find struct field definitions
				// For now, we'll try to use the instruction type if it was set by the type checker
				// The type checker should set inst.Type to the field's type
				if fieldType == "" || fieldType == "<inferred>" {
					// Last resort: check if we can infer from context
					// But we really should have the type from the type checker
					fieldType = "int" // Default fallback
				}
			}
			
			// Use appropriate getter based on field type
			switch fieldType {
			case "string":
				g.output.WriteString(fmt.Sprintf("  %s = omni_struct_get_string_field(%s, \"%s\");\n", varName, structVar, fieldName))
				g.valueTypes[inst.ID] = "string"
			case "float", "double":
				g.output.WriteString(fmt.Sprintf("  %s = omni_struct_get_float_field(%s, \"%s\");\n", varName, structVar, fieldName))
				g.valueTypes[inst.ID] = fieldType
			case "bool":
				g.output.WriteString(fmt.Sprintf("  %s = omni_struct_get_bool_field(%s, \"%s\");\n", varName, structVar, fieldName))
				g.valueTypes[inst.ID] = "bool"
			default:
				// Default to int
				g.output.WriteString(fmt.Sprintf("  %s = omni_struct_get_int_field(%s, \"%s\");\n", varName, structVar, fieldName))
				g.valueTypes[inst.ID] = "int"
			}
		}
	case "cmp.eq", "cmp.neq", "cmp.lt", "cmp.lte", "cmp.gt", "cmp.gte":
		// Handle comparison operations
		if len(inst.Operands) >= 2 {
			left := g.getOperandValue(inst.Operands[0])
			right := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)

			// Check if we're comparing strings
			leftType := inst.Operands[0].Type
			rightType := inst.Operands[1].Type
			// Also check valueTypes for more accurate type information
			if inst.Operands[0].Kind == mir.OperandValue {
				if storedType, ok := g.valueTypes[inst.Operands[0].Value]; ok && storedType != "" {
					leftType = storedType
				}
			}
			if inst.Operands[1].Kind == mir.OperandValue {
				if storedType, ok := g.valueTypes[inst.Operands[1].Value]; ok && storedType != "" {
					rightType = storedType
				}
			}

			// If either operand is a string, use string comparison function
			if leftType == "string" || rightType == "string" {
				switch inst.Op {
				case "cmp.eq":
					g.output.WriteString(fmt.Sprintf("  %s = omni_string_equals(%s, %s) ? 1 : 0;\n",
						varName, left, right))
				case "cmp.neq":
					g.output.WriteString(fmt.Sprintf("  %s = omni_string_equals(%s, %s) ? 0 : 1;\n",
						varName, left, right))
				case "cmp.lt":
					g.output.WriteString(fmt.Sprintf("  %s = (omni_string_compare(%s, %s) < 0) ? 1 : 0;\n",
						varName, left, right))
				case "cmp.lte":
					g.output.WriteString(fmt.Sprintf("  %s = (omni_string_compare(%s, %s) <= 0) ? 1 : 0;\n",
						varName, left, right))
				case "cmp.gt":
					g.output.WriteString(fmt.Sprintf("  %s = (omni_string_compare(%s, %s) > 0) ? 1 : 0;\n",
						varName, left, right))
				case "cmp.gte":
					g.output.WriteString(fmt.Sprintf("  %s = (omni_string_compare(%s, %s) >= 0) ? 1 : 0;\n",
						varName, left, right))
				}
			} else {
				// For non-string types, use direct comparison
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

				// Comparison - assign to already declared variable
				g.output.WriteString(fmt.Sprintf("  %s = (%s %s %s) ? 1 : 0;\n",
					varName, left, op, right))
			}
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

			// Check if this is a struct type (for struct array iteration)
			resultType := inst.Type
			if storedType, ok := g.valueTypes[inst.ID]; ok && storedType != "" {
				resultType = storedType
			}
			isStruct := !g.isPrimitiveType(resultType) && !strings.Contains(resultType, "<") && !strings.Contains(resultType, "(")
			
			// Store the type for later use
			if isStruct {
				g.valueTypes[inst.ID] = resultType
			}

			// Create a mutable variable that can be updated in the loop
			// For PHI nodes, we initialize with the first value and will update it later
			// Note: This is a special case where we need to initialize the variable
			g.output.WriteString(fmt.Sprintf("  %s = %s;\n",
				varName, firstValue))

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
				// Regular variable assignment - assign to already declared variable
				g.output.WriteString(fmt.Sprintf("  %s = %s;\n",
					varName, funcName))
			}
		}
	case "func.call":
		// Handle function call through function pointer
		if len(inst.Operands) >= 1 {
			funcPtr := g.getOperandValue(inst.Operands[0])
			varName := g.getVariableName(inst.ID)

			// Build function call with arguments - assign to already declared variable
			g.output.WriteString(fmt.Sprintf("  %s = %s(",
				varName, funcPtr))

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
	case "std.io.print":
		if len(inst.Operands) >= 1 {
			g.emitPrint(inst.Operands[0], false)
		}
	case "std.io.println":
		if len(inst.Operands) >= 1 {
			g.emitPrint(inst.Operands[0], true)
		} else {
			g.output.WriteString("  omni_println_string(\"\");\n")
		}
	case "std.log.debug":
		// Handle debug log statement
		if len(inst.Operands) >= 1 {
			message := g.getOperandValue(inst.Operands[0])
			g.output.WriteString(fmt.Sprintf("  omni_log_debug(%s);\n", message))
		}
	case "std.log.info":
		// Handle info log statement
		if len(inst.Operands) >= 1 {
			message := g.getOperandValue(inst.Operands[0])
			g.output.WriteString(fmt.Sprintf("  omni_log_info(%s);\n", message))
		}
	case "std.log.warn":
		// Handle warn log statement
		if len(inst.Operands) >= 1 {
			message := g.getOperandValue(inst.Operands[0])
			g.output.WriteString(fmt.Sprintf("  omni_log_warn(%s);\n", message))
		}
	case "std.log.error":
		// Handle error log statement
		if len(inst.Operands) >= 1 {
			message := g.getOperandValue(inst.Operands[0])
			g.output.WriteString(fmt.Sprintf("  omni_log_error(%s);\n", message))
		}
	case "std.log.set_level":
		// Handle log level setting
		if len(inst.Operands) >= 1 {
			level := g.getOperandValue(inst.Operands[0])
			g.output.WriteString(fmt.Sprintf("  omni_log_set_level(%s);\n", level))
		}
	case "await":
		// Handle await instruction - unwrap Promise<T> to T
		if len(inst.Operands) >= 1 {
			promiseVar := g.getOperandValue(inst.Operands[0])
			varName := g.getVariableName(inst.ID)
			
			// Determine the await function based on result type
			// The inst.Type should already be the unwrapped type (e.g., "string" not "Promise<string>")
			resultType := inst.Type
			
			// If type is empty or unknown, try to infer from the instruction type
			if resultType == "" || resultType == "<inferred>" {
				// Try to get type from operand's promise type
				if len(inst.Operands) > 0 && inst.Operands[0].Kind == mir.OperandValue {
					operandID := inst.Operands[0].Value
					if storedType, ok := g.valueTypes[operandID]; ok {
						// Extract inner type from Promise<T>
						if strings.HasPrefix(storedType, "Promise<") && strings.HasSuffix(storedType, ">") {
							resultType = storedType[len("Promise<") : len(storedType)-1]
						}
					}
				}
			}
			
			switch resultType {
			case "int":
				g.output.WriteString(fmt.Sprintf("  %s = omni_await_int(%s);\n", varName, promiseVar))
			case "string":
				g.output.WriteString(fmt.Sprintf("  %s = omni_await_string(%s);\n", varName, promiseVar))
			case "float", "double":
				g.output.WriteString(fmt.Sprintf("  %s = omni_await_float(%s);\n", varName, promiseVar))
			case "bool":
				g.output.WriteString(fmt.Sprintf("  %s = omni_await_bool(%s);\n", varName, promiseVar))
			default:
				// For unknown types, default to string (most common for I/O)
				g.output.WriteString(fmt.Sprintf("  %s = omni_await_string(%s);\n", varName, promiseVar))
				resultType = "string"
			}
			// Store the result type
			g.valueTypes[inst.ID] = resultType
		}
	default:
		// Handle unknown instructions
		g.output.WriteString(fmt.Sprintf("  // TODO: Implement instruction %s\n", inst.Op))
	}

	return nil
}

// generateTerminator generates C code for a terminator
func (g *CGenerator) generateTerminator(term *mir.Terminator, funcName string) error {
	switch term.Op {
	case "ret":
		// Handle return statement
		if len(term.Operands) > 0 {
			value := g.getOperandValue(term.Operands[0])
			g.output.WriteString(fmt.Sprintf("  return %s;\n", value))
		} else {
			// For main function, return 0 instead of void return
			if funcName == "main" {
				g.output.WriteString("  return 0;\n")
			} else {
				g.output.WriteString("  return;\n")
			}
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

		operandType := op.Type
		// Always check g.valueTypes as it may have more accurate type information
		if recorded, exists := g.valueTypes[op.Value]; exists && recorded != "" {
			operandType = recorded
		}
		// Fallback to op.Type if g.valueTypes doesn't have it
		if operandType == "" || operandType == inferTypePlaceholder {
			operandType = op.Type
		}

		// Check the operand type and convert if necessary
		if operandType == "string" {
			// Already a string, return as is
			return varName
		} else if operandType == "int" {
			// Convert int to string - use unique counter to avoid conflicts
			tempVar := fmt.Sprintf("temp_str_%d_%d", op.Value, g.output.Len())
			g.output.WriteString(fmt.Sprintf("  const char* %s = omni_int_to_string(%s);\n", tempVar, varName))
			return tempVar
		} else if operandType == "float" || operandType == "double" {
			// Convert float to string - use unique counter to avoid conflicts
			tempVar := fmt.Sprintf("temp_str_%d_%d", op.Value, g.output.Len())
			g.output.WriteString(fmt.Sprintf("  const char* %s = omni_float_to_string(%s);\n", tempVar, varName))
			return tempVar
		} else if operandType == "bool" {
			// Convert bool to string - use unique counter to avoid conflicts
			tempVar := fmt.Sprintf("temp_str_%d_%d", op.Value, g.output.Len())
			g.output.WriteString(fmt.Sprintf("  const char* %s = omni_bool_to_string(%s);\n", tempVar, varName))
			return tempVar
		} else {
			// Default: assume int - use unique counter to avoid conflicts
			tempVar := fmt.Sprintf("temp_str_%d_%d", op.Value, g.output.Len())
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

// emitPrint handles std.io.print/println for primitive and convertible types.
func (g *CGenerator) emitPrint(op mir.Operand, newline bool) {
	funcName := "omni_print_string"
	if newline {
		funcName = "omni_println_string"
	}

	var arg string
	if op.Type == "string" {
		arg = g.getOperandValue(op)
	} else {
		arg = g.convertOperandToString(op)
	}

	g.output.WriteString(fmt.Sprintf("  %s(%s);\n", funcName, arg))
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

	// Handle Promise types: Promise<T>
	if strings.HasPrefix(omniType, "Promise<") && strings.HasSuffix(omniType, ">") {
		return "omni_promise_t*"
	}

	// Handle array types: []<ElementType>
	if strings.HasPrefix(omniType, "[]<") && strings.HasSuffix(omniType, ">") {
		elementType := omniType[3 : len(omniType)-1]
		// Arrays in C are represented as pointers to the element type
		return g.mapType(elementType) + "*"
	}
	// Handle old array syntax: array<ElementType>
	if strings.HasPrefix(omniType, "array<") && strings.HasSuffix(omniType, ">") {
		elementType := omniType[6 : len(omniType)-1]
		// Arrays in C are represented as pointers to the element type
		return g.mapType(elementType) + "*"
	}

	// Handle pointer types: *Type
	if strings.HasPrefix(omniType, "*") {
		baseType := omniType[1:] // Remove the *
		baseCType := g.mapType(baseType)
		return baseCType + "*"
	}

	// Handle map types: map<KeyType,ValueType>
	if strings.HasPrefix(omniType, "map<") && strings.HasSuffix(omniType, ">") {
		return "omni_map_t*"
	}

	// Handle struct types: struct<Field1Type,Field2Type,...>
	if strings.HasPrefix(omniType, "struct<") && strings.HasSuffix(omniType, ">") {
		return "omni_struct_t*"
	}

	// Handle named struct types (like Point, User, etc.)
	// For now, assume any unknown type that's not a primitive is a struct
	if !g.isPrimitiveType(omniType) && !strings.Contains(omniType, "(") && !strings.Contains(omniType, "<") {
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
	case "void*":
		return "void*"
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
	case "std.io.read_line":
		return "omni_read_line"

	// Logging functions
	case "std.log.debug":
		return "omni_log_debug"
	case "std.log.info":
		return "omni_log_info"
	case "std.log.warn":
		return "omni_log_warn"
	case "std.log.error":
		return "omni_log_error"
	case "std.log.set_level":
		return "omni_log_set_level"

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
		"io.print":                 true,
		"io.println":               true,
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
		"std.array.length":  true,
		"std.array.get":     true,
		"std.array.set":     true,
		"std.log.debug":     true,
		"std.log.info":      true,
		"std.log.warn":      true,
		"std.log.error":     true,
		"std.log.set_level": true,
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

// isVariableDeclared checks if a variable has already been declared in the current function
func (g *CGenerator) isVariableDeclared(varName string) bool {
	// Check if the variable name exists in the variables map
	for _, name := range g.variables {
		if name == varName {
			return true
		}
	}
	return false
}

// declareVariable marks a variable as declared
func (g *CGenerator) declareVariable(varName string) {
	// Find the ValueID that corresponds to this variable name
	for _, name := range g.variables {
		if name == varName {
			// Variable is already tracked, no need to do anything
			return
		}
	}
	// If we get here, the variable is not in the map, which means it's a new variable
	// We don't need to add it to the map since it will be added when getVariableName is called
}

// isPrimitiveType checks if a type is a primitive type
func (g *CGenerator) isPrimitiveType(omniType string) bool {
	switch omniType {
	case "int", "float", "double", "string", "void", "void*", "bool", "ptr":
		return true
	default:
		return false
	}
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
