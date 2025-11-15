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
	// Track variables that are maps (by variable name - legacy)
	mapVars map[string]bool
	// Track map types by value ID (more reliable)
	mapTypes map[mir.ValueID]string
	// Track array lengths by value ID (for runtime bounds checking and len())
	arrayLengths map[mir.ValueID]int
	// Debug symbol tracking
	sourceMap map[string]int // Maps source locations to line numbers
	lineMap   map[int]string // Maps line numbers to source locations
	// Track discovered value types to help with conversions
	valueTypes map[mir.ValueID]string
	// Track errors during code generation
	errors []string
	// Track variables that hold heap-allocated strings (need to be freed)
	stringsToFree map[mir.ValueID]bool
	// Track promises that need to be freed
	promisesToFree map[mir.ValueID]bool
	// Track temporary string variables created in convertOperandToString
	tempStringsToFree []string
	// Track the value ID that is being returned (to exclude from cleanup)
	returnedValueID mir.ValueID
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
		mapTypes:    make(map[mir.ValueID]string),
		arrayLengths: make(map[mir.ValueID]int),
		sourceMap:   make(map[string]int),
		lineMap:     make(map[int]string),
		valueTypes:    make(map[mir.ValueID]string),
		errors:        []string{},
		stringsToFree:    make(map[mir.ValueID]bool),
		promisesToFree:   make(map[mir.ValueID]bool),
		tempStringsToFree: []string{},
		returnedValueID:  mir.InvalidValue,
	}
}

// NewCGeneratorWithOptLevel creates a new C code generator with specified optimization level
func NewCGeneratorWithOptLevel(module *mir.Module, optLevel string) *CGenerator {
	return &CGenerator{
		module:        module,
		optLevel:      optLevel,
		debugInfo:     false,
		sourceFile:    "",
		variables:     make(map[mir.ValueID]string),
		phiVars:       make(map[mir.ValueID]bool),
		mutableVars:   make(map[mir.ValueID]bool),
		mapVars:       make(map[string]bool),
		mapTypes:      make(map[mir.ValueID]string),
		arrayLengths:  make(map[mir.ValueID]int),
		sourceMap:     make(map[string]int),
		lineMap:       make(map[int]string),
		valueTypes:    make(map[mir.ValueID]string),
		errors:        []string{},
		stringsToFree:    make(map[mir.ValueID]bool),
		promisesToFree:   make(map[mir.ValueID]bool),
		tempStringsToFree: []string{},
		returnedValueID:  mir.InvalidValue,
	}
}

// NewCGeneratorWithDebug creates a new C code generator with debug information
func NewCGeneratorWithDebug(module *mir.Module, optLevel string, debugInfo bool, sourceFile string) *CGenerator {
	return &CGenerator{
		module:        module,
		optLevel:      optLevel,
		debugInfo:     debugInfo,
		sourceFile:    sourceFile,
		variables:     make(map[mir.ValueID]string),
		phiVars:       make(map[mir.ValueID]bool),
		mutableVars:   make(map[mir.ValueID]bool),
		mapVars:       make(map[string]bool),
		mapTypes:      make(map[mir.ValueID]string),
		arrayLengths:  make(map[mir.ValueID]int),
		sourceMap:     make(map[string]int),
		lineMap:       make(map[int]string),
		valueTypes:    make(map[mir.ValueID]string),
		errors:        []string{},
		stringsToFree:    make(map[mir.ValueID]bool),
		promisesToFree:   make(map[mir.ValueID]bool),
		tempStringsToFree: []string{},
		returnedValueID:  mir.InvalidValue,
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

	// Check for errors collected during code generation
	if len(g.errors) > 0 {
		return "", fmt.Errorf("code generation errors:\n%s", strings.Join(g.errors, "\n"))
	}

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
		
		// For async functions (Promise<T>), the function should return omni_promise_t*
		// The Promise is created at the call site, but the function itself returns the promise
		if strings.HasPrefix(fn.ReturnType, "Promise<") {
			returnType = "omni_promise_t*"
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
		// Verify that the function actually has a runtime implementation
		if !g.hasRuntimeImplementation(fn.Name) {
			g.errors = append(g.errors, fmt.Sprintf("ERROR: Function '%s' is marked as runtime-provided but has no runtime implementation. Remove it from isRuntimeProvidedFunction or implement it in the runtime.", fn.Name))
		}
		return nil
	}
	
	// Warn if this is a stdlib function that should be an intrinsic but isn't implemented
	if g.isStdFunction(fn.Name) && !g.hasRuntimeImplementation(fn.Name) {
		// This is a stdlib function without a runtime implementation
		// It will use its stub body, which is likely wrong
		g.errors = append(g.errors, fmt.Sprintf("WARNING: stdlib function '%s' is not implemented in the runtime. It will use a stub body that returns a default value. Consider implementing it or removing it from the stdlib.", fn.Name))
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
	
	// For async functions (Promise<T>), the function should return omni_promise_t*
	// The function body will create a promise and return it
	if strings.HasPrefix(fn.ReturnType, "Promise<") {
		returnType = "omni_promise_t*"
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
	g.stringsToFree = make(map[mir.ValueID]bool)
	g.promisesToFree = make(map[mir.ValueID]bool)
	g.tempStringsToFree = []string{}
	g.returnedValueID = mir.InvalidValue

	// Map parameter SSA values to their names
	for _, param := range fn.Params {
		g.variables[param.ID] = param.Name
	}

	// Collect all variables that need to be declared
	allVariables := make(map[mir.ValueID]string)
	// Build a map from ValueID to Instruction for O(1) lookups (optimization: O(N) instead of O(N²))
	instructionMap := make(map[mir.ValueID]*mir.Instruction)
	for _, block := range fn.Blocks {
		for i := range block.Instructions {
			inst := &block.Instructions[i]
			if inst.ID != mir.InvalidValue {
				varName := fmt.Sprintf("v%d", int(inst.ID))
				allVariables[inst.ID] = varName
				instructionMap[inst.ID] = inst
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
			// Use the instruction map for O(1) lookup instead of scanning all instructions
			if inst, found := instructionMap[id]; found {
				// Special case for read_line() - always returns string
				if inst.Op == "call" || inst.Op == "call.string" {
					if len(inst.Operands) > 0 && inst.Operands[0].Kind == mir.OperandLiteral {
						funcName := inst.Operands[0].Literal
						if funcName == "std.io.read_line" || funcName == "io.read_line" {
							varType = "const char*"
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
				}
				// Special case for index - check if indexing into struct array
				if inst.Op == "index" && len(inst.Operands) >= 2 {
					// Check the target array type
					targetOp := inst.Operands[0]
					if targetOp.Kind == mir.OperandValue {
						// Use instruction map for O(1) lookup instead of scanning
						if targetInst, targetFound := instructionMap[targetOp.Value]; targetFound {
							if targetInst.Op == "array.init" {
								// Extract element type
								var elementTypeStr string
								if strings.HasPrefix(targetInst.Type, "array<") && strings.HasSuffix(targetInst.Type, ">") {
									elementTypeStr = targetInst.Type[6 : len(targetInst.Type)-1]
								} else if strings.HasPrefix(targetInst.Type, "[]<") && strings.HasSuffix(targetInst.Type, ">") {
									elementTypeStr = targetInst.Type[3 : len(targetInst.Type)-1]
								}
								// Check if element type is a struct
								if elementTypeStr != "" && !g.isPrimitiveType(elementTypeStr) && !strings.Contains(elementTypeStr, "<") && !strings.Contains(elementTypeStr, "(") {
									varType = "omni_struct_t*"
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
				}
				// Special case for phi - check if it's a struct type (for struct array iteration)
				if inst.Op == "phi" {
					resultType := inst.Type
					isStruct := !g.isPrimitiveType(resultType) && !strings.Contains(resultType, "<") && !strings.Contains(resultType, "(")
					if isStruct {
						varType = "omni_struct_t*"
					}
				}
				// Default: use the instruction type
				if varType == "" {
					varType = g.mapType(inst.Type)
				}
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
			}
			if varType == "" {
				varType = "int32_t" // default type
			}
			// Skip declaring void variables (they don't produce values)
			// Skip declaring array variables (they're declared in array.init)
			// For string constants, we'll initialize them in the const instruction
			if varType != "void" && !isArrayInit {
				// Check if this is a const instruction with a string literal
				// Use instruction map for O(1) lookup
				isStringConst := false
				if inst, found := instructionMap[id]; found {
					if inst.Op == "const" && len(inst.Operands) > 0 {
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
						}
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

	// Generate cleanup code for heap-allocated strings and promises
	// Free all tracked strings (in reverse order to handle dependencies)
	// Note: We iterate in reverse to free strings that might depend on others
	// Exclude the returned value from cleanup (caller owns it)
	if len(g.stringsToFree) > 0 {
		g.output.WriteString("  // Cleanup: free heap-allocated strings\n")
		// Collect all string IDs and sort them in reverse order
		var stringIDs []mir.ValueID
		for id := range g.stringsToFree {
			// Skip the returned value - caller owns it
			if id == g.returnedValueID {
				continue
			}
			stringIDs = append(stringIDs, id)
		}
		// Sort in reverse order (free later variables first)
		for i := len(stringIDs) - 1; i >= 0; i-- {
			id := stringIDs[i]
			varName := g.getVariableName(id)
			g.output.WriteString(fmt.Sprintf("  if (%s != NULL) { free((void*)%s); %s = NULL; }\n", varName, varName, varName))
		}
	}

	// Free temporary string variables created in convertOperandToString
	if len(g.tempStringsToFree) > 0 {
		g.output.WriteString("  // Cleanup: free temporary string conversion variables\n")
		// Free in reverse order (last created first)
		for i := len(g.tempStringsToFree) - 1; i >= 0; i-- {
			tempVar := g.tempStringsToFree[i]
			g.output.WriteString(fmt.Sprintf("  if (%s != NULL) { free((void*)%s); %s = NULL; }\n", tempVar, tempVar, tempVar))
		}
	}

	// Free all tracked promises
	if len(g.promisesToFree) > 0 {
		g.output.WriteString("  // Cleanup: free promises\n")
		for id := range g.promisesToFree {
			varName := g.getVariableName(id)
			g.output.WriteString(fmt.Sprintf("  if (%s != NULL) { omni_promise_free(%s); %s = NULL; }\n", varName, varName, varName))
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
	if err := g.generateTerminator(&block.Terminator, funcName, fn.ReturnType); err != nil {
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
			// Track this string for cleanup (omni_strcat returns heap-allocated string)
			if inst.ID != mir.InvalidValue {
				g.stringsToFree[inst.ID] = true
			}
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
				// Get array length from tracked array lengths
				arrayLength := -1 // Use -1 as sentinel for "unknown length"
				if inst.Operands[1].Kind == mir.OperandValue {
					arrayOperandID := inst.Operands[1].Value
					if length, ok := g.arrayLengths[arrayOperandID]; ok {
						arrayLength = length
					} else {
						// Array length not found - this might be a parameter or passed array
						// For function parameters, we need to track array lengths separately
						// For now, fail loudly to prevent silent bugs
						g.errors = append(g.errors, fmt.Sprintf("array length not known for variable %s (ID: %d) - len() requires compile-time known array length or explicit length parameter", arrayVar, arrayOperandID))
						// Still emit code with -1 so the program fails at runtime rather than silently returning 0
						arrayLength = -1
					}
				}
				// Get array type to determine element size
				arrayType := "int32_t" // Default element type
				if inst.Operands[1].Kind == mir.OperandValue {
					arrayOperandID := inst.Operands[1].Value
					if arrType, ok := g.valueTypes[arrayOperandID]; ok {
						// Extract element type from array type (e.g., "[]<int>" -> "int")
						if strings.HasPrefix(arrType, "[]<") && strings.HasSuffix(arrType, ">") {
							arrayType = arrType[3 : len(arrType)-1]
						} else if strings.HasPrefix(arrType, "array<") && strings.HasSuffix(arrType, ">") {
							arrayType = arrType[6 : len(arrType)-1]
						}
					}
				}
				// Map element type to C type to get element size
				elementCType := g.mapType(arrayType)
				// Calculate element size
				elementSize := "sizeof(" + elementCType + ")"
				// Use runtime function omni_len with explicit length
				// If length is -1 (unknown), the runtime should handle it (currently returns -1, which is wrong)
				// TODO: Implement proper array length tracking through function parameters
				if arrayLength < 0 {
					g.output.WriteString(fmt.Sprintf("  // WARNING: Array length unknown for %s, len() may return incorrect value\n", arrayVar))
				}
				g.output.WriteString(fmt.Sprintf("  %s = omni_len((void*)%s, %s, %d);\n",
					varName, arrayVar, elementSize, arrayLength))
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
					// Track this string for cleanup (omni_read_line returns heap-allocated string)
					g.stringsToFree[inst.ID] = true
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
					// Track the promise for cleanup
					g.promisesToFree[inst.ID] = true
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
					// Track the promise for cleanup
					g.promisesToFree[inst.ID] = true
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
					// Track the promise for cleanup
					g.promisesToFree[inst.ID] = true
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
						// Track the promise for cleanup
						g.promisesToFree[inst.ID] = true
					case "float", "double":
						g.output.WriteString(fmt.Sprintf("  %s = omni_promise_create_float(%s);\n", varName, tempVar))
					case "bool":
						g.output.WriteString(fmt.Sprintf("  %s = omni_promise_create_bool(%s);\n", varName, tempVar))
					default:
						// For user-defined types, we cannot create promises yet
						g.errors = append(g.errors, fmt.Sprintf("cannot create promise for user-defined type: %s in async call to %s", innerType, funcName))
						// Still emit code to prevent compilation errors, but it will be wrong
						g.output.WriteString(fmt.Sprintf("  // ERROR: Cannot create promise for type %s, using int (WRONG)\n", innerType))
						g.output.WriteString(fmt.Sprintf("  %s = omni_promise_create_int(%s); // WRONG TYPE\n", varName, tempVar))
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
					// Special handling for functions that return structs (IPAddress, URL, HTTPResponse, etc.)
					// Network functions returning structs
					if (funcName == "omni_ip_parse" || funcName == "omni_network_get_local_ip") && inst.Type != "" && strings.Contains(inst.Type, "IPAddress") {
						// IP functions return omni_ip_address_t*
						varName := g.getVariableName(inst.ID)
						if len(inst.Operands) >= 2 {
							ipStr := g.getOperandValue(inst.Operands[1])
							g.output.WriteString(fmt.Sprintf("  omni_ip_address_t* %s = omni_ip_parse(%s);\n", varName, ipStr))
						} else if funcName == "omni_network_get_local_ip" {
							g.output.WriteString(fmt.Sprintf("  omni_ip_address_t* %s = omni_network_get_local_ip();\n", varName))
						}
					} else if funcName == "omni_url_parse" && inst.Type != "" && strings.Contains(inst.Type, "URL") {
						// URL functions return omni_url_t*
						varName := g.getVariableName(inst.ID)
						if len(inst.Operands) >= 2 {
							urlStr := g.getOperandValue(inst.Operands[1])
							g.output.WriteString(fmt.Sprintf("  omni_url_t* %s = omni_url_parse(%s);\n", varName, urlStr))
						}
					} else if (funcName == "omni_http_get" || funcName == "omni_http_post" || funcName == "omni_http_put" || funcName == "omni_http_delete" || funcName == "omni_http_request") && inst.Type != "" && strings.Contains(inst.Type, "HTTPResponse") {
						// HTTP functions return omni_http_response_t*
						varName := g.getVariableName(inst.ID)
						if funcName == "omni_http_get" && len(inst.Operands) >= 2 {
							url := g.getOperandValue(inst.Operands[1])
							g.output.WriteString(fmt.Sprintf("  omni_http_response_t* %s = omni_http_get(%s);\n", varName, url))
						} else if funcName == "omni_http_post" && len(inst.Operands) >= 3 {
							url := g.getOperandValue(inst.Operands[1])
							body := g.getOperandValue(inst.Operands[2])
							g.output.WriteString(fmt.Sprintf("  omni_http_response_t* %s = omni_http_post(%s, %s);\n", varName, url, body))
						} else if funcName == "omni_http_put" && len(inst.Operands) >= 3 {
							url := g.getOperandValue(inst.Operands[1])
							body := g.getOperandValue(inst.Operands[2])
							g.output.WriteString(fmt.Sprintf("  omni_http_response_t* %s = omni_http_put(%s, %s);\n", varName, url, body))
						} else if funcName == "omni_http_delete" && len(inst.Operands) >= 2 {
							url := g.getOperandValue(inst.Operands[1])
							g.output.WriteString(fmt.Sprintf("  omni_http_response_t* %s = omni_http_delete(%s);\n", varName, url))
						} else if funcName == "omni_http_request" && len(inst.Operands) >= 2 {
							req := g.getOperandValue(inst.Operands[1])
							g.output.WriteString(fmt.Sprintf("  omni_http_response_t* %s = omni_http_request(%s);\n", varName, req))
						}
					} else if funcName == "omni_ip_to_string" && inst.Type == "string" {
						// IP to string conversion
						varName := g.getVariableName(inst.ID)
						if len(inst.Operands) >= 2 {
							ipVar := g.getOperandValue(inst.Operands[1])
							g.output.WriteString(fmt.Sprintf("  const char* %s = omni_ip_to_string(%s);\n", varName, ipVar))
							g.stringsToFree[inst.ID] = true
						}
					} else if funcName == "omni_url_to_string" && inst.Type == "string" {
						// URL to string conversion
						varName := g.getVariableName(inst.ID)
						if len(inst.Operands) >= 2 {
							urlVar := g.getOperandValue(inst.Operands[1])
							g.output.WriteString(fmt.Sprintf("  const char* %s = omni_url_to_string(%s);\n", varName, urlVar))
							g.stringsToFree[inst.ID] = true
						}
					} else if funcName == "omni_dns_reverse_lookup" && inst.Type == "string" {
						// DNS reverse lookup returns string
						varName := g.getVariableName(inst.ID)
						if len(inst.Operands) >= 2 {
							ipVar := g.getOperandValue(inst.Operands[1])
							g.output.WriteString(fmt.Sprintf("  const char* %s = omni_dns_reverse_lookup(%s);\n", varName, ipVar))
							g.stringsToFree[inst.ID] = true
						}
					} else if funcName == "omni_http_response_get_header" && inst.Type == "string" {
						// HTTP response get header returns string
						varName := g.getVariableName(inst.ID)
						if len(inst.Operands) >= 3 {
							respVar := g.getOperandValue(inst.Operands[1])
							headerName := g.getOperandValue(inst.Operands[2])
							g.output.WriteString(fmt.Sprintf("  const char* %s = omni_http_response_get_header(%s, %s);\n", varName, respVar, headerName))
							g.stringsToFree[inst.ID] = true
						}
					} else if funcName == "omni_http_request_get_header" && inst.Type == "string" {
						// HTTP request get header returns string
						varName := g.getVariableName(inst.ID)
						if len(inst.Operands) >= 3 {
							reqVar := g.getOperandValue(inst.Operands[1])
							headerName := g.getOperandValue(inst.Operands[2])
							g.output.WriteString(fmt.Sprintf("  const char* %s = omni_http_request_get_header(%s, %s);\n", varName, reqVar, headerName))
							g.stringsToFree[inst.ID] = true
						}
					} else if funcName == "omni_socket_receive" && inst.Type == "string" {
						// Socket receive returns string
						varName := g.getVariableName(inst.ID)
						if len(inst.Operands) >= 3 {
							socketVar := g.getOperandValue(inst.Operands[1])
							bufferSizeVar := g.getOperandValue(inst.Operands[2])
							// Allocate buffer
							bufferVar := fmt.Sprintf("_buffer_%d", inst.ID)
							g.output.WriteString(fmt.Sprintf("  char %s[%s + 1];\n", bufferVar, bufferSizeVar))
							g.output.WriteString(fmt.Sprintf("  int32_t %s_len = omni_socket_receive(%s, %s, %s);\n", varName, socketVar, bufferVar, bufferSizeVar))
							g.output.WriteString(fmt.Sprintf("  %s[%s_len] = '\\0';\n", bufferVar, varName))
							g.output.WriteString(fmt.Sprintf("  const char* %s = strdup(%s);\n", varName, bufferVar))
							g.stringsToFree[inst.ID] = true
						}
					} else if funcName == "omni_dns_lookup" && inst.Type != "" && strings.HasPrefix(inst.Type, "array<") {
						// DNS lookup returns array of IPAddress
						// For now, return empty array (stub implementation)
						varName := g.getVariableName(inst.ID)
						if len(inst.Operands) >= 2 {
							hostname := g.getOperandValue(inst.Operands[1])
							g.output.WriteString(fmt.Sprintf("  // DNS lookup stub - returns empty array\n"))
							g.output.WriteString(fmt.Sprintf("  int32_t %s_count = 0;\n", varName))
							g.output.WriteString(fmt.Sprintf("  omni_ip_address_t** %s = omni_dns_lookup(%s, &%s_count);\n", varName, hostname, varName))
						}
					} else if funcName == "omni_map_keys_string_int" && inst.Type != "" && strings.HasPrefix(inst.Type, "array<") {
						// Map keys returns array of strings
						// For now, use a fixed-size buffer (limitation)
						varName := g.getVariableName(inst.ID)
						if len(inst.Operands) >= 2 {
							mapVar := g.getOperandValue(inst.Operands[1])
							bufferVar := fmt.Sprintf("_keys_buffer_%d", inst.ID)
							g.output.WriteString(fmt.Sprintf("  char* %s[256];\n", bufferVar))
							g.output.WriteString(fmt.Sprintf("  int32_t %s_count = omni_map_keys_string_int(%s, %s, 256);\n", varName, mapVar, bufferVar))
							// Note: The array would need to be constructed from the buffer
							// For now, just track the count
						}
					} else if funcName == "omni_map_values_string_int" && inst.Type != "" && strings.HasPrefix(inst.Type, "array<") {
						// Map values returns array of ints
						// For now, use a fixed-size buffer (limitation)
						varName := g.getVariableName(inst.ID)
						if len(inst.Operands) >= 2 {
							mapVar := g.getOperandValue(inst.Operands[1])
							bufferVar := fmt.Sprintf("_values_buffer_%d", inst.ID)
							g.output.WriteString(fmt.Sprintf("  int32_t %s[256];\n", bufferVar))
							g.output.WriteString(fmt.Sprintf("  int32_t %s_count = omni_map_values_string_int(%s, %s, 256);\n", varName, mapVar, bufferVar))
							// Note: The array would need to be constructed from the buffer
							// For now, just track the count
						}
					} else if funcName == "omni_map_copy_string_int" && inst.Type != "" && strings.HasPrefix(inst.Type, "map<") {
						// Map copy returns a new map
						varName := g.getVariableName(inst.ID)
						if len(inst.Operands) >= 2 {
							mapVar := g.getOperandValue(inst.Operands[1])
							g.output.WriteString(fmt.Sprintf("  omni_map_t* %s = omni_map_copy_string_int(%s);\n", varName, mapVar))
						}
					} else if funcName == "omni_map_merge_string_int" && inst.Type != "" && strings.HasPrefix(inst.Type, "map<") {
						// Map merge returns a new map
						varName := g.getVariableName(inst.ID)
						if len(inst.Operands) >= 3 {
							mapA := g.getOperandValue(inst.Operands[1])
							mapB := g.getOperandValue(inst.Operands[2])
							g.output.WriteString(fmt.Sprintf("  omni_map_t* %s = omni_map_merge_string_int(%s, %s);\n", varName, mapA, mapB))
						}
					} else if funcName == "omni_set_union" && inst.Type != "" && strings.HasPrefix(inst.Type, "set<") {
						// Set union returns a new set
						varName := g.getVariableName(inst.ID)
						if len(inst.Operands) >= 3 {
							setA := g.getOperandValue(inst.Operands[1])
							setB := g.getOperandValue(inst.Operands[2])
							g.output.WriteString(fmt.Sprintf("  omni_set_t* %s = omni_set_union(%s, %s);\n", varName, setA, setB))
						}
					} else if funcName == "omni_set_intersection" && inst.Type != "" && strings.HasPrefix(inst.Type, "set<") {
						// Set intersection returns a new set
						varName := g.getVariableName(inst.ID)
						if len(inst.Operands) >= 3 {
							setA := g.getOperandValue(inst.Operands[1])
							setB := g.getOperandValue(inst.Operands[2])
							g.output.WriteString(fmt.Sprintf("  omni_set_t* %s = omni_set_intersection(%s, %s);\n", varName, setA, setB))
						}
					} else if funcName == "omni_set_difference" && inst.Type != "" && strings.HasPrefix(inst.Type, "set<") {
						// Set difference returns a new set
						varName := g.getVariableName(inst.ID)
						if len(inst.Operands) >= 3 {
							setA := g.getOperandValue(inst.Operands[1])
							setB := g.getOperandValue(inst.Operands[2])
							g.output.WriteString(fmt.Sprintf("  omni_set_t* %s = omni_set_difference(%s, %s);\n", varName, setA, setB))
						}
					} else if funcName == "omni_set_create" && inst.Type != "" && strings.HasPrefix(inst.Type, "set<") {
						// Set create returns a new set
						varName := g.getVariableName(inst.ID)
						g.output.WriteString(fmt.Sprintf("  omni_set_t* %s = omni_set_create();\n", varName))
					} else if funcName == "omni_queue_create" && inst.Type != "" && strings.HasPrefix(inst.Type, "queue<") {
						// Queue create returns a new queue
						varName := g.getVariableName(inst.ID)
						g.output.WriteString(fmt.Sprintf("  omni_queue_t* %s = omni_queue_create();\n", varName))
					} else if funcName == "omni_stack_create" && inst.Type != "" && strings.HasPrefix(inst.Type, "stack<") {
						// Stack create returns a new stack
						varName := g.getVariableName(inst.ID)
						g.output.WriteString(fmt.Sprintf("  omni_stack_t* %s = omni_stack_create();\n", varName))
					} else if funcName == "omni_priority_queue_create" && inst.Type != "" && strings.HasPrefix(inst.Type, "priority_queue<") {
						// Priority queue create returns a new priority queue
						varName := g.getVariableName(inst.ID)
						g.output.WriteString(fmt.Sprintf("  omni_priority_queue_t* %s = omni_priority_queue_create();\n", varName))
					} else if funcName == "omni_linked_list_create" && inst.Type != "" && strings.HasPrefix(inst.Type, "linked_list<") {
						// Linked list create returns a new linked list
						varName := g.getVariableName(inst.ID)
						g.output.WriteString(fmt.Sprintf("  omni_linked_list_t* %s = omni_linked_list_create();\n", varName))
					} else if funcName == "omni_binary_tree_create" && inst.Type != "" && strings.HasPrefix(inst.Type, "binary_tree<") {
						// Binary tree create returns a new binary tree
						varName := g.getVariableName(inst.ID)
						g.output.WriteString(fmt.Sprintf("  omni_binary_tree_t* %s = omni_binary_tree_create();\n", varName))
					} else if funcName == "omni_http_request_create" && inst.Type != "" && strings.Contains(inst.Type, "HTTPRequest") {
						// HTTP request create returns omni_http_request_t*
						varName := g.getVariableName(inst.ID)
						if len(inst.Operands) >= 3 {
							method := g.getOperandValue(inst.Operands[1])
							url := g.getOperandValue(inst.Operands[2])
							g.output.WriteString(fmt.Sprintf("  omni_http_request_t* %s = omni_http_request_create(%s, %s);\n", varName, method, url))
						}
					} else {
						// Special handling for Time struct conversion functions
						if funcName == "omni_time_from_unix" && inst.Type != "" && strings.Contains(inst.Type, "Time") {
							// time_from_unix(timestamp) -> Time
							// Need to extract fields from Time struct and pass as output parameters
							if len(inst.Operands) >= 2 {
								timestamp := g.getOperandValue(inst.Operands[1])
								// Create temporary variables for output parameters
								yearVar := fmt.Sprintf("_year_%d", inst.ID)
								monthVar := fmt.Sprintf("_month_%d", inst.ID)
								dayVar := fmt.Sprintf("_day_%d", inst.ID)
								hourVar := fmt.Sprintf("_hour_%d", inst.ID)
								minuteVar := fmt.Sprintf("_minute_%d", inst.ID)
								secondVar := fmt.Sprintf("_second_%d", inst.ID)
								nanosecondVar := fmt.Sprintf("_nanosecond_%d", inst.ID)
								
								g.output.WriteString(fmt.Sprintf("  int32_t %s, %s, %s, %s, %s, %s, %s;\n",
									yearVar, monthVar, dayVar, hourVar, minuteVar, secondVar, nanosecondVar))
								g.output.WriteString(fmt.Sprintf("  omni_time_from_unix(%s, &%s, &%s, &%s, &%s, &%s, &%s, &%s);\n",
									timestamp, yearVar, monthVar, dayVar, hourVar, minuteVar, secondVar, nanosecondVar))
								
								// Create Time struct from the extracted fields
								g.output.WriteString(fmt.Sprintf("  %s = omni_struct_create();\n", varName))
								g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field(%s, \"year\", %s);\n", varName, yearVar))
								g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field(%s, \"month\", %s);\n", varName, monthVar))
								g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field(%s, \"day\", %s);\n", varName, dayVar))
								g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field(%s, \"hour\", %s);\n", varName, hourVar))
								g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field(%s, \"minute\", %s);\n", varName, minuteVar))
								g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field(%s, \"second\", %s);\n", varName, secondVar))
								g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field(%s, \"nanosecond\", %s);\n", varName, nanosecondVar))
							}
						} else if funcName == "omni_time_from_string" && inst.Type != "" && strings.Contains(inst.Type, "Time") {
							// time_from_string(time_str) -> Time
							if len(inst.Operands) >= 2 {
								timeStr := g.getOperandValue(inst.Operands[1])
								yearVar := fmt.Sprintf("_year_%d", inst.ID)
								monthVar := fmt.Sprintf("_month_%d", inst.ID)
								dayVar := fmt.Sprintf("_day_%d", inst.ID)
								hourVar := fmt.Sprintf("_hour_%d", inst.ID)
								minuteVar := fmt.Sprintf("_minute_%d", inst.ID)
								secondVar := fmt.Sprintf("_second_%d", inst.ID)
								nanosecondVar := fmt.Sprintf("_nanosecond_%d", inst.ID)
								
								g.output.WriteString(fmt.Sprintf("  int32_t %s, %s, %s, %s, %s, %s, %s;\n",
									yearVar, monthVar, dayVar, hourVar, minuteVar, secondVar, nanosecondVar))
								g.output.WriteString(fmt.Sprintf("  omni_time_from_string(%s, &%s, &%s, &%s, &%s, &%s, &%s, &%s);\n",
									timeStr, yearVar, monthVar, dayVar, hourVar, minuteVar, secondVar, nanosecondVar))
								
								g.output.WriteString(fmt.Sprintf("  %s = omni_struct_create();\n", varName))
								g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field(%s, \"year\", %s);\n", varName, yearVar))
								g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field(%s, \"month\", %s);\n", varName, monthVar))
								g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field(%s, \"day\", %s);\n", varName, dayVar))
								g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field(%s, \"hour\", %s);\n", varName, hourVar))
								g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field(%s, \"minute\", %s);\n", varName, minuteVar))
								g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field(%s, \"second\", %s);\n", varName, secondVar))
								g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field(%s, \"nanosecond\", %s);\n", varName, nanosecondVar))
							}
						} else if funcName == "omni_time_to_unix" && len(inst.Operands) >= 2 {
							// time_to_unix(t:Time) -> int
							// Need to extract fields from Time struct parameter
							timeStruct := g.getOperandValue(inst.Operands[1])
							yearVar := fmt.Sprintf("_year_%d", inst.ID)
							monthVar := fmt.Sprintf("_month_%d", inst.ID)
							dayVar := fmt.Sprintf("_day_%d", inst.ID)
							hourVar := fmt.Sprintf("_hour_%d", inst.ID)
							minuteVar := fmt.Sprintf("_minute_%d", inst.ID)
							secondVar := fmt.Sprintf("_second_%d", inst.ID)
							nanosecondVar := fmt.Sprintf("_nanosecond_%d", inst.ID)
							
							g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_int_field(%s, \"year\");\n", yearVar, timeStruct))
							g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_int_field(%s, \"month\");\n", monthVar, timeStruct))
							g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_int_field(%s, \"day\");\n", dayVar, timeStruct))
							g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_int_field(%s, \"hour\");\n", hourVar, timeStruct))
							g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_int_field(%s, \"minute\");\n", minuteVar, timeStruct))
							g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_int_field(%s, \"second\");\n", secondVar, timeStruct))
							g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_int_field(%s, \"nanosecond\");\n", nanosecondVar, timeStruct))
							g.output.WriteString(fmt.Sprintf("  %s = omni_time_to_unix(%s, %s, %s, %s, %s, %s, %s);\n",
								varName, yearVar, monthVar, dayVar, hourVar, minuteVar, secondVar, nanosecondVar))
						} else if funcName == "omni_time_to_string" && len(inst.Operands) >= 2 {
							// time_to_string(t:Time) -> string
							timeStruct := g.getOperandValue(inst.Operands[1])
							yearVar := fmt.Sprintf("_year_%d", inst.ID)
							monthVar := fmt.Sprintf("_month_%d", inst.ID)
							dayVar := fmt.Sprintf("_day_%d", inst.ID)
							hourVar := fmt.Sprintf("_hour_%d", inst.ID)
							minuteVar := fmt.Sprintf("_minute_%d", inst.ID)
							secondVar := fmt.Sprintf("_second_%d", inst.ID)
							nanosecondVar := fmt.Sprintf("_nanosecond_%d", inst.ID)
							
							g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_int_field(%s, \"year\");\n", yearVar, timeStruct))
							g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_int_field(%s, \"month\");\n", monthVar, timeStruct))
							g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_int_field(%s, \"day\");\n", dayVar, timeStruct))
							g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_int_field(%s, \"hour\");\n", hourVar, timeStruct))
							g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_int_field(%s, \"minute\");\n", minuteVar, timeStruct))
							g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_int_field(%s, \"second\");\n", secondVar, timeStruct))
							g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_int_field(%s, \"nanosecond\");\n", nanosecondVar, timeStruct))
							g.output.WriteString(fmt.Sprintf("  %s = omni_time_to_string(%s, %s, %s, %s, %s, %s, %s);\n",
								varName, yearVar, monthVar, dayVar, hourVar, minuteVar, secondVar, nanosecondVar))
							// Track string for cleanup
							if inst.Type == "string" {
								g.stringsToFree[inst.ID] = true
							}
						} else if funcName == "omni_time_to_unix_nano" && len(inst.Operands) >= 2 {
							// time_to_unix_nano(t:Time) -> int
							timeStruct := g.getOperandValue(inst.Operands[1])
							yearVar := fmt.Sprintf("_year_%d", inst.ID)
							monthVar := fmt.Sprintf("_month_%d", inst.ID)
							dayVar := fmt.Sprintf("_day_%d", inst.ID)
							hourVar := fmt.Sprintf("_hour_%d", inst.ID)
							minuteVar := fmt.Sprintf("_minute_%d", inst.ID)
							secondVar := fmt.Sprintf("_second_%d", inst.ID)
							nanosecondVar := fmt.Sprintf("_nanosecond_%d", inst.ID)
							
							g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_int_field(%s, \"year\");\n", yearVar, timeStruct))
							g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_int_field(%s, \"month\");\n", monthVar, timeStruct))
							g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_int_field(%s, \"day\");\n", dayVar, timeStruct))
							g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_int_field(%s, \"hour\");\n", hourVar, timeStruct))
							g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_int_field(%s, \"minute\");\n", minuteVar, timeStruct))
							g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_int_field(%s, \"second\");\n", secondVar, timeStruct))
							g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_int_field(%s, \"nanosecond\");\n", nanosecondVar, timeStruct))
							g.output.WriteString(fmt.Sprintf("  %s = omni_time_to_unix_nano(%s, %s, %s, %s, %s, %s, %s);\n",
								varName, yearVar, monthVar, dayVar, hourVar, minuteVar, secondVar, nanosecondVar))
						} else if funcName == "omni_duration_to_string" && len(inst.Operands) >= 2 {
							// duration_to_string(d:Duration) -> string
							durationStruct := g.getOperandValue(inst.Operands[1])
							secondsVar := fmt.Sprintf("_seconds_%d", inst.ID)
							nanosecondsVar := fmt.Sprintf("_nanoseconds_%d", inst.ID)
							
							g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_int_field(%s, \"seconds\");\n", secondsVar, durationStruct))
							g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_int_field(%s, \"nanoseconds\");\n", nanosecondsVar, durationStruct))
							g.output.WriteString(fmt.Sprintf("  %s = omni_duration_to_string(%s, %s);\n",
								varName, secondsVar, nanosecondsVar))
							// Track string for cleanup
							if inst.Type == "string" {
								g.stringsToFree[inst.ID] = true
							}
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
							// Track strings that need freeing if this function returns a heap-allocated string
							if g.isStringReturningFunction(funcName) && inst.Type == "string" {
								g.stringsToFree[inst.ID] = true
							}
							// Warn if function is called but doesn't have a runtime implementation
							if !g.hasRuntimeImplementation(funcName) && g.isStdFunction(funcName) {
								g.errors = append(g.errors, fmt.Sprintf("WARNING: stdlib function '%s' is called but has no runtime implementation. It will return a default value or do nothing.", funcName))
							}
						}
					}
				}
			}
		}
	case "index":
		// Handle array/map indexing
		if len(inst.Operands) >= 2 {
			target := g.getOperandValue(inst.Operands[0])
			index := g.getOperandValue(inst.Operands[1])
			varName := g.getVariableName(inst.ID)

			// Check if target is a map - use type-based detection (more reliable)
			isMap := false
			if inst.Operands[0].Kind == mir.OperandValue {
				if mapType, ok := g.mapTypes[inst.Operands[0].Value]; ok {
					isMap = strings.HasPrefix(mapType, "map<")
				} else if storedType, ok := g.valueTypes[inst.Operands[0].Value]; ok {
					isMap = strings.HasPrefix(storedType, "map<")
				}
			}
			// Fallback to legacy name-based detection
			if !isMap {
				isMap = g.isMapVariable(target)
			}
			
			if isMap {
				// Map indexing - need to determine key and value types from map type
				mapType := "map<string,int>" // Default
				if inst.Operands[0].Kind == mir.OperandValue {
					if storedType, ok := g.valueTypes[inst.Operands[0].Value]; ok && strings.HasPrefix(storedType, "map<") {
						mapType = storedType
					}
				}
				// Extract key and value types
				keyType, valueType := g.extractMapTypes(mapType)
				getFunc := g.getMapGetFunction(keyType, valueType)
				if getFunc != "" {
					g.output.WriteString(fmt.Sprintf("  %s = %s(%s, %s);\n", varName, getFunc, target, index))
				} else {
					g.errors = append(g.errors, fmt.Sprintf("unsupported map get operation for type: %s", mapType))
					g.output.WriteString(fmt.Sprintf("  // ERROR: Unsupported map get for %s\n", mapType))
				}
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
					// For primitive arrays, use runtime function with bounds checking if length is known
					arrayLength := -1
					if inst.Operands[0].Kind == mir.OperandValue {
						arrayOperandID := inst.Operands[0].Value
						if length, ok := g.arrayLengths[arrayOperandID]; ok {
							arrayLength = length
						}
					}
					if arrayLength >= 0 {
						// Use runtime function with bounds checking
						g.output.WriteString(fmt.Sprintf("  %s = omni_array_get_int(%s, %s, %d);\n", 
							varName, target, index, arrayLength))
					} else {
						// Length unknown (might be parameter) - still use runtime function but with -1
						// This will cause a runtime error rather than silent memory corruption
						g.errors = append(g.errors, fmt.Sprintf("array length not known for indexing %s (ID: %d) - bounds checking disabled, may cause memory corruption", target, inst.Operands[0].Value))
						g.output.WriteString(fmt.Sprintf("  // WARNING: Array length unknown, bounds checking disabled\n"))
						g.output.WriteString(fmt.Sprintf("  %s = %s[%s]; // UNSAFE: No bounds check\n", varName, target, index))
					}
				}
			}
		}
	case "array.init":
		// Handle array literal initialization
		// WARNING: Arrays are currently allocated on the stack, which means:
		// - They cannot be returned from functions (dangling pointers)
		// - They cannot be stored in longer-lived variables
		// - They decay to pointers when passed around
		// TODO: Implement heap allocation for arrays to support proper value semantics
		if len(inst.Operands) > 0 {
			varName := g.getVariableName(inst.ID)
			arrayLength := len(inst.Operands)
			// Store array length for later use in len() and bounds checking
			g.arrayLengths[inst.ID] = arrayLength
			
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
				g.output.WriteString(fmt.Sprintf("  // WARNING: Stack-allocated array, cannot be returned or stored long-term\n"))
				g.output.WriteString(fmt.Sprintf("  %s %s[%d];\n", elementType, varName, arrayLength))
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
				g.output.WriteString(fmt.Sprintf("  // WARNING: Stack-allocated array, cannot be returned or stored long-term\n"))
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
						putFunc := g.getMapPutFunction(keyType, valueType)
						if putFunc != "" {
							g.output.WriteString(fmt.Sprintf("  %s(%s, %s, %s);\n", putFunc, varName, key, value))
						} else {
							// Unsupported type combination - report error
							g.errors = append(g.errors, fmt.Sprintf("unsupported map type combination: map<%s,%s>", keyType, valueType))
							g.output.WriteString(fmt.Sprintf("  // ERROR: Unsupported map type %s\n", mapType))
						}
					}
				}
			}
		}

		// Mark this variable as a map (legacy tracking by name)
		g.mapVars[varName] = true
		// Also track by value ID and type (more reliable)
		if inst.ID != mir.InvalidValue {
			g.mapTypes[inst.ID] = inst.Type
		}
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
						
						// Default to int if still no type, but warn about it
						if fieldType == "" || fieldType == "<inferred>" {
							g.errors = append(g.errors, fmt.Sprintf("could not infer type for struct field '%s' in struct.init, defaulting to int (this may cause incorrect behavior)", fieldName))
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
							// For non-primitive types, we can't use the primitive setters
							// This should have been caught earlier, but fail loudly here
							if !g.isPrimitiveType(fieldType) {
								g.errors = append(g.errors, fmt.Sprintf("cannot set struct field '%s' with type %s: only primitive types are supported", fieldName, fieldType))
								// Fall back to int to prevent compilation errors
								g.output.WriteString(fmt.Sprintf("  // ERROR: Cannot set field %s with type %s, using int setter (WRONG)\n", fieldName, fieldType))
								g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field(%s, \"%s\", %s); // WRONG TYPE\n", varName, fieldName, fieldValue))
							} else {
								// Default to int for unknown primitive types
								g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field(%s, \"%s\", %s);\n", varName, fieldName, fieldValue))
							}
						}
					}
				}
			} else {
				// Positional field format: [field1Value, field2Value, ...]
				// We need to know the struct field names and order
				// For now, we'll use generic field names based on position
				// This is a limitation - ideally we'd look up the struct definition
				// but that information isn't easily accessible in the C generator
				for i, fieldValueOp := range inst.Operands {
					// Skip first operand if it's the struct type name
					if i == 0 && fieldValueOp.Kind == mir.OperandLiteral {
						// This might be the struct type name, skip it
						continue
					}
					
					fieldValue := g.getOperandValue(fieldValueOp)
					// Determine field type from the value operand
					fieldType := ""
					if fieldValueOp.Type != "" && fieldValueOp.Type != inferTypePlaceholder {
						fieldType = fieldValueOp.Type
					} else if fieldValueOp.Kind == mir.OperandValue {
						if storedType, ok := g.valueTypes[fieldValueOp.Value]; ok && storedType != "" {
							fieldType = storedType
						}
					}
					
					// Default to int if type not found
					if fieldType == "" {
						fieldType = "int"
					}
					
					// Use generic field name based on position
					fieldIndex := i
					if inst.Operands[0].Kind == mir.OperandLiteral {
						fieldIndex = i - 1 // Adjust if first operand was type name
					}
					fieldName := fmt.Sprintf("field%d", fieldIndex)
					
					// Use appropriate setter based on field type
					switch fieldType {
					case "string":
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
						g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field(%s, \"%s\", %s);\n", varName, fieldName, fieldValue))
					}
				}
			}
		}
	case "member":
		// Handle struct member access
		if len(inst.Operands) >= 2 {
			structVar := g.getOperandValue(inst.Operands[0])
			fieldName := inst.Operands[1].Literal
			varName := g.getVariableName(inst.ID)

			// Determine field type from instruction type
			// The type checker should have already substituted type parameters for generic structs
			// (e.g., Box<int>.value should have inst.Type = "int", not "T")
			fieldType := inst.Type
			
			// If type is not set or is a placeholder, try to infer it
			if fieldType == "" || fieldType == "<inferred>" || fieldType == inferTypePlaceholder {
				// Try to get the struct type and look up the field
				if inst.Operands[0].Kind == mir.OperandValue {
					if structType, ok := g.valueTypes[inst.Operands[0].Value]; ok && structType != "" && structType != inferTypePlaceholder {
						// Check if this is a generic struct (e.g., "Box<int>")
						if strings.Contains(structType, "<") && strings.Contains(structType, ">") {
							// Extract base name and type arguments
							baseName, typeArgs := g.extractGenericType(structType)
							if baseName != "" && len(typeArgs) > 0 {
								// For generic structs, we'd need struct field definitions to do substitution
								// But the type checker should have already done this, so inst.Type should be set
								// If it's not, we'll fall back to checking the operand type
							}
						}
					}
				}
				
				// Check operand type as fallback
				if (fieldType == "" || fieldType == "<inferred>" || fieldType == inferTypePlaceholder) && len(inst.Operands) > 0 {
					if inst.Operands[0].Type != "" && inst.Operands[0].Type != inferTypePlaceholder {
						// This is the struct type, not the field type, but we can use it as a hint
						// The real field type should come from inst.Type set by the type checker
					}
				}
				
				// Last resort: default to int (but this should rarely happen if type checker is working correctly)
				if fieldType == "" || fieldType == "<inferred>" || fieldType == inferTypePlaceholder {
					fieldType = "int"
				}
			}
			
			// Ensure we strip any generic syntax that might have leaked through
			// The type should already be the concrete type (e.g., "int" not "T" or "Box<int>")
			if strings.Contains(fieldType, "<") {
				// This shouldn't happen if type checker did its job, but handle it gracefully
				// Extract the inner type if it's a generic
				if strings.HasPrefix(fieldType, "array<") || strings.HasPrefix(fieldType, "[]<") {
					// Field type is an array, keep it as is
				} else {
					// Try to extract base type
					fieldType = strings.TrimSpace(fieldType)
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
	case "closure.create", "closure.capture", "closure.bind":
		// Closures are not yet supported in the C backend
		g.errors = append(g.errors, fmt.Sprintf("closures are not supported in the C backend (instruction: %s)", inst.Op))
		return fmt.Errorf("closures are not supported in the C backend: %s", inst.Op)
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
				// Track this string for cleanup (omni_await_string returns heap-allocated string)
				g.stringsToFree[inst.ID] = true
			case "float", "double":
				g.output.WriteString(fmt.Sprintf("  %s = omni_await_float(%s);\n", varName, promiseVar))
			case "bool":
				g.output.WriteString(fmt.Sprintf("  %s = omni_await_bool(%s);\n", varName, promiseVar))
			default:
				// For user-defined types, we cannot await them yet
				// Fail loudly instead of silently defaulting to string
				g.errors = append(g.errors, fmt.Sprintf("cannot await Promise<%s>: user-defined types are not supported in await expressions", resultType))
				// Still emit code to prevent compilation errors, but it will be wrong
				g.output.WriteString(fmt.Sprintf("  // ERROR: Cannot await user-defined type %s, defaulting to int (WRONG)\n", resultType))
				g.output.WriteString(fmt.Sprintf("  %s = omni_await_int(%s); // WRONG TYPE\n", varName, promiseVar))
				resultType = "int" // Store as int to prevent further type errors
			}
			// Store the result type
			g.valueTypes[inst.ID] = resultType
		}
	default:
		// Unknown instructions should cause a hard failure
		g.errors = append(g.errors, fmt.Sprintf("unsupported MIR instruction: %s (this indicates a missing implementation or invalid MIR)", inst.Op))
		return fmt.Errorf("unsupported MIR instruction: %s", inst.Op)
	}

	return nil
}

// generateTerminator generates C code for a terminator
func (g *CGenerator) generateTerminator(term *mir.Terminator, funcName string, returnType string) error {
	switch term.Op {
	case "ret":
		// Handle return statement
		if len(term.Operands) > 0 {
			// Track the returned value ID to exclude it from cleanup
			if term.Operands[0].Kind == mir.OperandValue {
				g.returnedValueID = term.Operands[0].Value
			}
			value := g.getOperandValue(term.Operands[0])
			// For async functions (Promise<T>), wrap the return value in a promise
			if strings.HasPrefix(returnType, "Promise<") {
				innerType := returnType[8 : len(returnType)-1] // Extract inner type from Promise<T>
				// Determine the promise creation function based on inner type
				var promiseFunc string
				switch innerType {
				case "int":
					promiseFunc = "omni_promise_create_int"
				case "string":
					promiseFunc = "omni_promise_create_string"
					// If returning a string, track the promise for cleanup (not the string itself)
					// The promise will be freed, but the string inside it is owned by the promise
				case "float", "double":
					promiseFunc = "omni_promise_create_float"
				case "bool":
					promiseFunc = "omni_promise_create_bool"
				default:
					// For user-defined types, we can't create promises yet
					// This should be caught earlier, but fail loudly here
					g.errors = append(g.errors, fmt.Sprintf("cannot create promise for user-defined type: %s", innerType))
					promiseFunc = "omni_promise_create_int" // Fallback to prevent compilation error
				}
				// Create promise and return it
				g.output.WriteString(fmt.Sprintf("  return %s(%s);\n", promiseFunc, value))
			} else {
				// For string return types, exclude the returned value from cleanup
				if returnType == "string" || returnType == "const char*" || returnType == "char*" {
					// The returned string is owned by the caller, don't free it
					if term.Operands[0].Kind == mir.OperandValue {
						delete(g.stringsToFree, term.Operands[0].Value)
					}
				}
				g.output.WriteString(fmt.Sprintf("  return %s;\n", value))
			}
		} else {
			// For main function (omni_main), return 0 instead of void return
			// Check if this is actually omni_main (mapped from main)
			if funcName == "omni_main" {
				g.output.WriteString("  return 0;\n")
			} else {
				// For void functions, use void return
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
		// Unknown terminators should cause a hard failure
		g.errors = append(g.errors, fmt.Sprintf("unsupported MIR terminator: %s (this indicates a missing implementation or invalid MIR)", term.Op))
		return fmt.Errorf("unsupported MIR terminator: %s", term.Op)
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
			// Track this temporary string for cleanup
			g.tempStringsToFree = append(g.tempStringsToFree, tempVar)
			return tempVar
		} else if operandType == "float" || operandType == "double" {
			// Convert float to string - use unique counter to avoid conflicts
			tempVar := fmt.Sprintf("temp_str_%d_%d", op.Value, g.output.Len())
			g.output.WriteString(fmt.Sprintf("  const char* %s = omni_float_to_string(%s);\n", tempVar, varName))
			// Track this temporary string for cleanup
			g.tempStringsToFree = append(g.tempStringsToFree, tempVar)
			return tempVar
		} else if operandType == "bool" {
			// Convert bool to string - use unique counter to avoid conflicts
			tempVar := fmt.Sprintf("temp_str_%d_%d", op.Value, g.output.Len())
			g.output.WriteString(fmt.Sprintf("  const char* %s = omni_bool_to_string(%s);\n", tempVar, varName))
			// Track this temporary string for cleanup
			g.tempStringsToFree = append(g.tempStringsToFree, tempVar)
			return tempVar
		} else {
			// Default: assume int - use unique counter to avoid conflicts
			tempVar := fmt.Sprintf("temp_str_%d_%d", op.Value, g.output.Len())
			g.output.WriteString(fmt.Sprintf("  const char* %s = omni_int_to_string(%s);\n", tempVar, varName))
			// Track this temporary string for cleanup
			g.tempStringsToFree = append(g.tempStringsToFree, tempVar)
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
		// Unknown type - this should not happen in valid programs
		// Report error instead of silently defaulting to int32_t
		g.errors = append(g.errors, fmt.Sprintf("unknown type: %s (cannot map to C type)", omniType))
		return "int32_t" // Temporary fallback to allow compilation to continue
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
	case "std.math.sin":
		return "omni_sin"
	case "std.math.cos":
		return "omni_cos"
	case "std.math.tan":
		return "omni_tan"
	case "std.math.asin":
		return "omni_asin"
	case "std.math.acos":
		return "omni_acos"
	case "std.math.atan":
		return "omni_atan"
	case "std.math.atan2":
		return "omni_atan2"
	case "std.math.exp":
		return "omni_exp"
	case "std.math.log":
		return "omni_log"
	case "std.math.log10":
		return "omni_log10"
	case "std.math.log2":
		return "omni_log2"
	case "std.math.sinh":
		return "omni_sinh"
	case "std.math.cosh":
		return "omni_cosh"
	case "std.math.tanh":
		return "omni_tanh"
	case "std.math.cbrt":
		return "omni_cbrt"
	case "std.math.trunc":
		return "omni_trunc"
	// Note: is_prime and fibonacci are implemented in OmniLang, not runtime
	case "std.math.is_prime":
		return "std_math_is_prime" // Will use OmniLang implementation
	case "std.math.fibonacci":
		return "std_math_fibonacci" // Will use OmniLang implementation

	// Collections functions
	case "std.collections.keys":
		return "omni_map_keys_string_int"
	case "std.collections.values":
		return "omni_map_values_string_int"
	case "std.collections.copy":
		return "omni_map_copy_string_int"
	case "std.collections.merge":
		return "omni_map_merge_string_int"
	// Set functions
	case "std.collections.set_create":
		return "omni_set_create"
	case "std.collections.set_add":
		return "omni_set_add"
	case "std.collections.set_remove":
		return "omni_set_remove"
	case "std.collections.set_contains":
		return "omni_set_contains"
	case "std.collections.set_size":
		return "omni_set_size"
	case "std.collections.set_clear":
		return "omni_set_clear"
	case "std.collections.set_union":
		return "omni_set_union"
	case "std.collections.set_intersection":
		return "omni_set_intersection"
	case "std.collections.set_difference":
		return "omni_set_difference"
	// Queue functions
	case "std.collections.queue_create":
		return "omni_queue_create"
	case "std.collections.queue_enqueue":
		return "omni_queue_enqueue"
	case "std.collections.queue_dequeue":
		return "omni_queue_dequeue"
	case "std.collections.queue_peek":
		return "omni_queue_peek"
	case "std.collections.queue_is_empty":
		return "omni_queue_is_empty"
	case "std.collections.queue_size":
		return "omni_queue_size"
	case "std.collections.queue_clear":
		return "omni_queue_clear"
	// Stack functions
	case "std.collections.stack_create":
		return "omni_stack_create"
	case "std.collections.stack_push":
		return "omni_stack_push"
	case "std.collections.stack_pop":
		return "omni_stack_pop"
	case "std.collections.stack_peek":
		return "omni_stack_peek"
	case "std.collections.stack_is_empty":
		return "omni_stack_is_empty"
	case "std.collections.stack_size":
		return "omni_stack_size"
	case "std.collections.stack_clear":
		return "omni_stack_clear"
	// Priority queue functions
	case "std.collections.priority_queue_create":
		return "omni_priority_queue_create"
	case "std.collections.priority_queue_insert":
		return "omni_priority_queue_insert"
	case "std.collections.priority_queue_extract_max":
		return "omni_priority_queue_extract_max"
	case "std.collections.priority_queue_peek":
		return "omni_priority_queue_peek"
	case "std.collections.priority_queue_is_empty":
		return "omni_priority_queue_is_empty"
	case "std.collections.priority_queue_size":
		return "omni_priority_queue_size"
	// Linked list functions
	case "std.collections.linked_list_create":
		return "omni_linked_list_create"
	case "std.collections.linked_list_append":
		return "omni_linked_list_append"
	case "std.collections.linked_list_prepend":
		return "omni_linked_list_prepend"
	case "std.collections.linked_list_insert":
		return "omni_linked_list_insert"
	case "std.collections.linked_list_remove":
		return "omni_linked_list_remove"
	case "std.collections.linked_list_get":
		return "omni_linked_list_get"
	case "std.collections.linked_list_set":
		return "omni_linked_list_set"
	case "std.collections.linked_list_size":
		return "omni_linked_list_size"
	case "std.collections.linked_list_is_empty":
		return "omni_linked_list_is_empty"
	case "std.collections.linked_list_clear":
		return "omni_linked_list_clear"
	// Binary tree functions
	case "std.collections.binary_tree_create":
		return "omni_binary_tree_create"
	case "std.collections.binary_tree_insert":
		return "omni_binary_tree_insert"
	case "std.collections.binary_tree_search":
		return "omni_binary_tree_search"
	case "std.collections.binary_tree_remove":
		return "omni_binary_tree_remove"
	case "std.collections.binary_tree_size":
		return "omni_binary_tree_size"
	case "std.collections.binary_tree_is_empty":
		return "omni_binary_tree_is_empty"
	case "std.collections.binary_tree_clear":
		return "omni_binary_tree_clear"
	// Network functions
	case "std.network.ip_parse":
		return "omni_ip_parse"
	case "std.network.ip_is_valid":
		return "omni_ip_is_valid"
	case "std.network.ip_is_private":
		return "omni_ip_is_private"
	case "std.network.ip_is_loopback":
		return "omni_ip_is_loopback"
	case "std.network.ip_to_string":
		return "omni_ip_to_string"
	case "std.network.url_parse":
		return "omni_url_parse"
	case "std.network.url_to_string":
		return "omni_url_to_string"
	case "std.network.url_is_valid":
		return "omni_url_is_valid"
	case "std.network.dns_lookup":
		return "omni_dns_lookup"
	case "std.network.dns_reverse_lookup":
		return "omni_dns_reverse_lookup"
	case "std.network.http_get":
		return "omni_http_get"
	case "std.network.http_post":
		return "omni_http_post"
	case "std.network.http_put":
		return "omni_http_put"
	case "std.network.http_delete":
		return "omni_http_delete"
	case "std.network.http_request":
		return "omni_http_request"
	case "std.network.socket_create":
		return "omni_socket_create"
	case "std.network.socket_connect":
		return "omni_socket_connect"
	case "std.network.socket_bind":
		return "omni_socket_bind"
	case "std.network.socket_listen":
		return "omni_socket_listen"
	case "std.network.socket_accept":
		return "omni_socket_accept"
	case "std.network.socket_send":
		return "omni_socket_send"
	case "std.network.socket_receive":
		return "omni_socket_receive"
	case "std.network.socket_close":
		return "omni_socket_close"
	case "std.network.network_is_connected":
		return "omni_network_is_connected"
	case "std.network.network_get_local_ip":
		return "omni_network_get_local_ip"
	case "std.network.network_ping":
		return "omni_network_ping"

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
	case "std.os.getenv":
		return "omni_getenv"
	case "std.os.setenv":
		return "omni_setenv"
	case "std.os.unsetenv":
		return "omni_unsetenv"
	case "std.os.getcwd":
		return "omni_getcwd"
	case "std.os.chdir":
		return "omni_chdir"
	case "std.os.mkdir":
		return "omni_mkdir"
	case "std.os.rmdir":
		return "omni_rmdir"
	case "std.os.remove":
		return "omni_remove"
	case "std.os.rename":
		return "omni_rename"
	case "std.os.copy":
		return "omni_copy"
	case "std.os.exists":
		return "omni_exists"
	case "std.os.is_file":
		return "omni_is_file"
	case "std.os.is_dir":
		return "omni_is_dir"
	case "std.os.args":
		return "omni_args_get"
	case "std.os.args_count":
		return "omni_args_count"
	case "std.os.has_flag":
		return "omni_args_has_flag"
	case "std.os.get_flag":
		return "omni_args_get_flag"
	case "std.os.positional_arg":
		return "omni_args_positional"
	case "os.args":
		return "omni_args_get"
	case "os.args_count":
		return "omni_args_count"
	case "os.has_flag":
		return "omni_args_has_flag"
	case "os.get_flag":
		return "omni_args_get_flag"
	case "os.positional_arg":
		return "omni_args_positional"
	case "std.os.getpid":
		return "omni_getpid"
	case "std.os.getppid":
		return "omni_getppid"
	case "os.getpid":
		return "omni_getpid"
	case "os.getppid":
		return "omni_getppid"
	case "std.string.is_alpha":
		return "omni_string_is_alpha"
	case "std.string.is_digit":
		return "omni_string_is_digit"
	case "std.string.is_alnum":
		return "omni_string_is_alnum"
	case "std.string.is_ascii":
		return "omni_string_is_ascii"
	case "std.string.is_upper":
		return "omni_string_is_upper"
	case "std.string.is_lower":
		return "omni_string_is_lower"
	case "string.is_alpha":
		return "omni_string_is_alpha"
	case "string.is_digit":
		return "omni_string_is_digit"
	case "string.is_alnum":
		return "omni_string_is_alnum"
	case "string.is_ascii":
		return "omni_string_is_ascii"
	case "string.is_upper":
		return "omni_string_is_upper"
	case "string.is_lower":
		return "omni_string_is_lower"
	case "std.string.encode_base64":
		return "omni_encode_base64"
	case "std.string.decode_base64":
		return "omni_decode_base64"
	case "std.string.encode_url":
		return "omni_encode_url"
	case "std.string.decode_url":
		return "omni_decode_url"
	case "std.string.escape_html":
		return "omni_escape_html"
	case "std.string.unescape_html":
		return "omni_unescape_html"
	case "std.string.escape_json":
		return "omni_escape_json"
	case "std.string.escape_shell":
		return "omni_escape_shell"
	case "string.encode_base64":
		return "omni_encode_base64"
	case "string.decode_base64":
		return "omni_decode_base64"
	case "string.encode_url":
		return "omni_encode_url"
	case "string.decode_url":
		return "omni_decode_url"
	case "string.escape_html":
		return "omni_escape_html"
	case "string.unescape_html":
		return "omni_unescape_html"
	case "string.escape_json":
		return "omni_escape_json"
	case "string.escape_shell":
		return "omni_escape_shell"
	case "std.time.now":
		return "omni_time_now_unix"
	case "std.time.unix_timestamp":
		return "omni_time_now_unix"
	case "std.time.unix_nano":
		return "omni_time_now_unix_nano"
	case "std.time.sleep_seconds":
		return "omni_time_sleep_seconds"
	case "std.time.sleep_milliseconds":
		return "omni_time_sleep_milliseconds"
	case "std.time.time_zone_offset":
		return "omni_time_zone_offset"
	case "std.time.time_zone_name":
		return "omni_time_zone_name"
	case "time.now":
		return "omni_time_now_unix"
	case "time.unix_timestamp":
		return "omni_time_now_unix"
	case "time.unix_nano":
		return "omni_time_now_unix_nano"
	case "time.sleep_seconds":
		return "omni_time_sleep_seconds"
	case "time.sleep_milliseconds":
		return "omni_time_sleep_milliseconds"
	case "time.time_zone_offset":
		return "omni_time_zone_offset"
	case "time.time_zone_name":
		return "omni_time_zone_name"

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

// hasRuntimeImplementation checks if a function has an actual runtime implementation
// This is used to verify that functions marked as runtime-provided actually exist
func (g *CGenerator) hasRuntimeImplementation(funcName string) bool {
	// Map of OmniLang function names to their runtime C function names
	// Only include functions that actually exist in omni_rt.h
	runtimeImplMap := map[string]string{
		// I/O functions
		"std.io.print":   "omni_print_string",
		"std.io.println": "omni_println_string",
		"io.print":       "omni_print_string",
		"io.println":     "omni_println_string",
		"std.io.read_line": "omni_read_line",
		"io.read_line":     "omni_read_line",
		
		// String functions
		"std.string.length":        "omni_strlen",
		"std.string.concat":        "omni_strcat",
		"std.string.substring":     "omni_substring",
		"std.string.char_at":        "omni_char_at",
		"std.string.starts_with":    "omni_starts_with",
		"std.string.ends_with":      "omni_ends_with",
		"std.string.contains":       "omni_contains",
		"std.string.index_of":       "omni_index_of",
		"std.string.last_index_of":  "omni_last_index_of",
		"std.string.trim":           "omni_trim",
		"std.string.to_upper":       "omni_to_upper",
		"std.string.to_lower":       "omni_to_lower",
		"std.string.equals":         "omni_string_equals",
		"std.string.compare":        "omni_string_compare",
		"string.length":             "omni_strlen",
		"string.concat":             "omni_strcat",
		"string.substring":          "omni_substring",
		"string.char_at":            "omni_char_at",
		"string.starts_with":        "omni_starts_with",
		"string.ends_with":          "omni_ends_with",
		"string.contains":           "omni_contains",
		"string.index_of":           "omni_index_of",
		"string.last_index_of":      "omni_last_index_of",
		"string.trim":               "omni_trim",
		"string.to_upper":           "omni_to_upper",
		"string.to_lower":           "omni_to_lower",
		"string.equals":             "omni_string_equals",
		"string.compare":            "omni_string_compare",
		
		// Math functions (only those with runtime implementations)
		"std.math.abs":       "omni_abs",
		"std.math.max":       "omni_max",
		"std.math.min":       "omni_min",
		"std.math.pow":       "omni_pow",
		"std.math.sqrt":      "omni_sqrt",
		"std.math.floor":     "omni_floor",
		"std.math.ceil":      "omni_ceil",
		"std.math.round":     "omni_round",
		"std.math.gcd":       "omni_gcd",
		"std.math.lcm":       "omni_lcm",
		"std.math.factorial": "omni_factorial",
		"std.math.sin":       "omni_sin",
		"std.math.cos":       "omni_cos",
		"std.math.tan":       "omni_tan",
		"std.math.asin":      "omni_asin",
		"std.math.acos":      "omni_acos",
		"std.math.atan":      "omni_atan",
		"std.math.atan2":     "omni_atan2",
		"std.math.exp":       "omni_exp",
		"std.math.log":       "omni_log",
		"std.math.log10":     "omni_log10",
		"std.math.log2":      "omni_log2",
		"std.math.sinh":      "omni_sinh",
		"std.math.cosh":      "omni_cosh",
		"std.math.tanh":      "omni_tanh",
		"std.math.cbrt":      "omni_cbrt",
		"std.math.trunc":     "omni_trunc",
		"math.abs":           "omni_abs",
		"math.max":           "omni_max",
		"math.min":           "omni_min",
		"math.pow":           "omni_pow",
		"math.sqrt":          "omni_sqrt",
		"math.floor":         "omni_floor",
		"math.ceil":          "omni_ceil",
		"math.round":         "omni_round",
		"math.gcd":           "omni_gcd",
		"math.lcm":           "omni_lcm",
		"math.factorial":     "omni_factorial",
		"math.sin":           "omni_sin",
		"math.cos":           "omni_cos",
		"math.tan":           "omni_tan",
		"math.asin":          "omni_asin",
		"math.acos":          "omni_acos",
		"math.atan":          "omni_atan",
		"math.atan2":         "omni_atan2",
		"math.exp":           "omni_exp",
		"math.log":           "omni_log",
		"math.log10":         "omni_log10",
		"math.log2":          "omni_log2",
		"math.sinh":          "omni_sinh",
		"math.cosh":          "omni_cosh",
		"math.tanh":          "omni_tanh",
		"math.cbrt":          "omni_cbrt",
		"math.trunc":         "omni_trunc",
		
		// Type conversion functions
		"std.int_to_string":   "omni_int_to_string",
		"std.float_to_string": "omni_float_to_string",
		"std.bool_to_string": "omni_bool_to_string",
		"std.string_to_int":   "omni_string_to_int",
		"std.string_to_float": "omni_string_to_float",
		"std.string_to_bool":  "omni_string_to_bool",
		
		// Logging functions
		"std.log.debug":     "omni_log_debug",
		"std.log.info":      "omni_log_info",
		"std.log.warn":      "omni_log_warn",
		"std.log.error":     "omni_log_error",
		"std.log.set_level": "omni_log_set_level",
		
		// File operations
		"std.file.open":   "omni_file_open",
		"std.file.close":  "omni_file_close",
		"std.file.read":   "omni_file_read",
		"std.file.write":  "omni_file_write",
		"std.file.seek":   "omni_file_seek",
		"std.file.tell":   "omni_file_tell",
		"std.file.exists": "omni_file_exists",
		"std.file.size":   "omni_file_size",
		"file.open":       "omni_file_open",
		"file.close":      "omni_file_close",
		"file.read":       "omni_file_read",
		"file.write":      "omni_file_write",
		"file.seek":       "omni_file_seek",
		"file.tell":       "omni_file_tell",
		"file.exists":    "omni_file_exists",
		"file.size":       "omni_file_size",
		"std.os.read_file":   "omni_read_file",
		"std.os.write_file":  "omni_write_file",
		"std.os.append_file": "omni_append_file",
		"os.read_file":       "omni_read_file",
		"os.write_file":     "omni_write_file",
		"os.append_file":    "omni_append_file",
		
		// System operations
		"std.os.exit":     "omni_exit",
		"std.os.getenv":   "omni_getenv",
		"std.os.setenv":   "omni_setenv",
		"std.os.unsetenv": "omni_unsetenv",
		"std.os.getcwd":   "omni_getcwd",
		"std.os.chdir":    "omni_chdir",
		"std.os.mkdir":    "omni_mkdir",
		"std.os.rmdir":    "omni_rmdir",
		"std.os.remove":   "omni_remove",
		"std.os.rename":   "omni_rename",
		"std.os.copy":     "omni_copy",
		"std.os.exists":   "omni_exists",
		"std.os.is_file":  "omni_is_file",
		"std.os.is_dir":   "omni_is_dir",
		"os.exit":         "omni_exit",
		"os.getenv":       "omni_getenv",
		"os.setenv":       "omni_setenv",
		"os.unsetenv":     "omni_unsetenv",
		"os.getcwd":       "omni_getcwd",
		"os.chdir":        "omni_chdir",
		"os.mkdir":        "omni_mkdir",
		"os.rmdir":        "omni_rmdir",
		"os.remove":       "omni_remove",
		"os.rename":       "omni_rename",
		"os.copy":         "omni_copy",
		"os.exists":       "omni_exists",
		"os.is_file":      "omni_is_file",
		"os.is_dir":       "omni_is_dir",
		
		// Testing functions
		"std.test.start": "omni_test_start",
		"std.test.end":   "omni_test_end",
		"std.assert":     "omni_assert",
		"std.assert_eq":  "omni_assert_eq_int", // Note: type-specific versions exist
		"test.start":     "omni_test_start",
		"test.end":       "omni_test_end",
		
		// String validation functions
		"std.string.is_alpha": "omni_string_is_alpha",
		"std.string.is_digit": "omni_string_is_digit",
		"std.string.is_alnum": "omni_string_is_alnum",
		"std.string.is_ascii": "omni_string_is_ascii",
		"std.string.is_upper": "omni_string_is_upper",
		"std.string.is_lower": "omni_string_is_lower",
		"string.is_alpha":     "omni_string_is_alpha",
		"string.is_digit":     "omni_string_is_digit",
		"string.is_alnum":     "omni_string_is_alnum",
		"string.is_ascii":     "omni_string_is_ascii",
		"string.is_upper":     "omni_string_is_upper",
		"string.is_lower":     "omni_string_is_lower",
		
		// String encoding/escaping functions
		"std.string.encode_base64": "omni_encode_base64",
		"std.string.decode_base64": "omni_decode_base64",
		"std.string.encode_url":    "omni_encode_url",
		"std.string.decode_url":    "omni_decode_url",
		"std.string.escape_html":   "omni_escape_html",
		"std.string.unescape_html": "omni_unescape_html",
		"std.string.escape_json":   "omni_escape_json",
		"std.string.escape_shell":  "omni_escape_shell",
		"string.encode_base64":     "omni_encode_base64",
		"string.decode_base64":     "omni_decode_base64",
		"string.encode_url":        "omni_encode_url",
		"string.decode_url":        "omni_decode_url",
		"string.escape_html":       "omni_escape_html",
		"string.unescape_html":     "omni_unescape_html",
		"string.escape_json":       "omni_escape_json",
		"string.escape_shell":      "omni_escape_shell",
		
		// Regex functions
		"std.string.matches":        "omni_string_matches",
		"std.string.find_match":     "omni_string_find_match",
		"std.string.find_all_matches": "omni_string_find_all_matches",
		"std.string.replace_regex":  "omni_string_replace_regex",
		"string.matches":             "omni_string_matches",
		"string.find_match":         "omni_string_find_match",
		"string.find_all_matches":   "omni_string_find_all_matches",
		"string.replace_regex":      "omni_string_replace_regex",
		
		// Time functions
		"std.time.now":              "omni_time_now_unix",
		"std.time.unix_timestamp":   "omni_time_now_unix",
		"std.time.unix_nano":        "omni_time_now_unix_nano",
		"std.time.sleep_seconds":   "omni_time_sleep_seconds",
		"std.time.sleep_milliseconds": "omni_time_sleep_milliseconds",
		"std.time.time_zone_offset": "omni_time_zone_offset",
		"std.time.time_zone_name":   "omni_time_zone_name",
		"std.time.time_from_unix":   "omni_time_from_unix",
		"std.time.time_to_unix":     "omni_time_to_unix",
		"std.time.time_to_string":   "omni_time_to_string",
		"std.time.time_from_string": "omni_time_from_string",
		"std.time.time_to_unix_nano": "omni_time_to_unix_nano",
		"std.time.duration_to_string": "omni_duration_to_string",
		"time.now":                  "omni_time_now_unix",
		"time.unix_timestamp":       "omni_time_now_unix",
		"time.unix_nano":            "omni_time_now_unix_nano",
		"time.sleep_seconds":        "omni_time_sleep_seconds",
		"time.sleep_milliseconds":   "omni_time_sleep_milliseconds",
		"time.time_zone_offset":     "omni_time_zone_offset",
		"time.time_zone_name":       "omni_time_zone_name",
		"time.time_from_unix":       "omni_time_from_unix",
		"time.time_to_unix":         "omni_time_to_unix",
		"time.time_to_string":       "omni_time_to_string",
		"time.time_from_string":     "omni_time_from_string",
		"time.time_to_unix_nano":    "omni_time_to_unix_nano",
		"time.duration_to_string":   "omni_duration_to_string",
		
		// Command-line argument functions
		"std.os.args":            "omni_args_get",
		"std.os.args_count":      "omni_args_count",
		"std.os.has_flag":        "omni_args_has_flag",
		"std.os.get_flag":        "omni_args_get_flag",
		"std.os.positional_arg":  "omni_args_positional",
		"os.args":                "omni_args_get",
		"os.args_count":          "omni_args_count",
		"os.has_flag":            "omni_args_has_flag",
		"os.get_flag":            "omni_args_get_flag",
		"os.positional_arg":      "omni_args_positional",
		
		// Process ID functions
		"std.os.getpid":  "omni_getpid",
		"std.os.getppid": "omni_getppid",
		"os.getpid":      "omni_getpid",
		"os.getppid":     "omni_getppid",
		
		// Collections functions
		"std.collections.keys":   "omni_map_keys_string_int",
		"std.collections.values": "omni_map_values_string_int",
		"std.collections.copy":   "omni_map_copy_string_int",
		"std.collections.merge":  "omni_map_merge_string_int",
		// Set functions
		"std.collections.set_create":       "omni_set_create",
		"std.collections.set_add":           "omni_set_add",
		"std.collections.set_remove":        "omni_set_remove",
		"std.collections.set_contains":      "omni_set_contains",
		"std.collections.set_size":          "omni_set_size",
		"std.collections.set_clear":         "omni_set_clear",
		"std.collections.set_union":         "omni_set_union",
		"std.collections.set_intersection": "omni_set_intersection",
		"std.collections.set_difference":   "omni_set_difference",
		// Queue functions
		"std.collections.queue_create":   "omni_queue_create",
		"std.collections.queue_enqueue":  "omni_queue_enqueue",
		"std.collections.queue_dequeue":  "omni_queue_dequeue",
		"std.collections.queue_peek":     "omni_queue_peek",
		"std.collections.queue_is_empty": "omni_queue_is_empty",
		"std.collections.queue_size":     "omni_queue_size",
		"std.collections.queue_clear":    "omni_queue_clear",
		// Stack functions
		"std.collections.stack_create":   "omni_stack_create",
		"std.collections.stack_push":      "omni_stack_push",
		"std.collections.stack_pop":       "omni_stack_pop",
		"std.collections.stack_peek":      "omni_stack_peek",
		"std.collections.stack_is_empty":  "omni_stack_is_empty",
		"std.collections.stack_size":      "omni_stack_size",
		"std.collections.stack_clear":     "omni_stack_clear",
		// Priority queue functions
		"std.collections.priority_queue_create":    "omni_priority_queue_create",
		"std.collections.priority_queue_insert":     "omni_priority_queue_insert",
		"std.collections.priority_queue_extract_max": "omni_priority_queue_extract_max",
		"std.collections.priority_queue_peek":       "omni_priority_queue_peek",
		"std.collections.priority_queue_is_empty":   "omni_priority_queue_is_empty",
		"std.collections.priority_queue_size":       "omni_priority_queue_size",
		// Linked list functions
		"std.collections.linked_list_create":   "omni_linked_list_create",
		"std.collections.linked_list_append":   "omni_linked_list_append",
		"std.collections.linked_list_prepend":  "omni_linked_list_prepend",
		"std.collections.linked_list_insert":   "omni_linked_list_insert",
		"std.collections.linked_list_remove":  "omni_linked_list_remove",
		"std.collections.linked_list_get":     "omni_linked_list_get",
		"std.collections.linked_list_set":      "omni_linked_list_set",
		"std.collections.linked_list_size":    "omni_linked_list_size",
		"std.collections.linked_list_is_empty": "omni_linked_list_is_empty",
		"std.collections.linked_list_clear":   "omni_linked_list_clear",
		// Binary tree functions
		"std.collections.binary_tree_create":   "omni_binary_tree_create",
		"std.collections.binary_tree_insert":   "omni_binary_tree_insert",
		"std.collections.binary_tree_search":   "omni_binary_tree_search",
		"std.collections.binary_tree_remove":   "omni_binary_tree_remove",
		"std.collections.binary_tree_size":    "omni_binary_tree_size",
		"std.collections.binary_tree_is_empty": "omni_binary_tree_is_empty",
		"std.collections.binary_tree_clear":    "omni_binary_tree_clear",
		// Network functions
		"std.network.ip_parse":         "omni_ip_parse",
		"std.network.ip_is_valid":      "omni_ip_is_valid",
		"std.network.ip_is_private":    "omni_ip_is_private",
		"std.network.ip_is_loopback":   "omni_ip_is_loopback",
		"std.network.ip_to_string":     "omni_ip_to_string",
		"std.network.url_parse":        "omni_url_parse",
		"std.network.url_to_string":    "omni_url_to_string",
		"std.network.url_is_valid":     "omni_url_is_valid",
		"std.network.dns_lookup":       "omni_dns_lookup",
		"std.network.dns_reverse_lookup": "omni_dns_reverse_lookup",
		"std.network.http_get":         "omni_http_get",
		"std.network.http_post":        "omni_http_post",
		"std.network.http_put":         "omni_http_put",
		"std.network.http_delete":      "omni_http_delete",
		"std.network.http_request":      "omni_http_request",
		"std.network.socket_create":    "omni_socket_create",
		"std.network.socket_connect":   "omni_socket_connect",
		"std.network.socket_bind":      "omni_socket_bind",
		"std.network.socket_listen":    "omni_socket_listen",
		"std.network.socket_accept":    "omni_socket_accept",
		"std.network.socket_send":       "omni_socket_send",
		"std.network.socket_receive":    "omni_socket_receive",
		"std.network.socket_close":     "omni_socket_close",
		"std.network.network_is_connected": "omni_network_is_connected",
		"std.network.network_get_local_ip":  "omni_network_get_local_ip",
		"std.network.network_ping":          "omni_network_ping",
	}
	
	_, exists := runtimeImplMap[funcName]
	return exists
}

// isRuntimeProvidedFunction checks if a function is provided by the runtime
// This should only return true for functions that actually have runtime implementations
func (g *CGenerator) isRuntimeProvidedFunction(funcName string) bool {
	// Only mark functions as runtime-provided if they actually have implementations
	if !g.hasRuntimeImplementation(funcName) {
		return false
	}
	
	// List of functions that are provided by the runtime
	// NOTE: This list should only include functions that:
	// 1. Have actual runtime implementations (checked by hasRuntimeImplementation)
	// 2. Should skip body generation (intrinsics)
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
		"std.math.factorial": true,
		"math.abs":           true,
		"math.max":           true,
		"math.min":           true,
		"math.pow":           true,
		"math.sqrt":          true,
		"math.floor":         true,
		"math.ceil":          true,
		"math.round":         true,
		"math.gcd":           true,
		"math.lcm":           true,
		"math.factorial":     true,
		"std.os.exit":        true,
		"os.exit":            true,
		"std.os.read_file":   true,
		"std.os.write_file":  true,
		"std.os.append_file": true,
		"os.read_file":       true,
		"os.write_file":      true,
		"os.append_file":     true,
		// File operations
		"file.open":         true,
		"file.close":        true,
		"file.read":         true,
		"file.write":        true,
		"file.seek":         true,
		"file.tell":         true,
		"file.exists":       true,
		"file.size":         true,
		"std.int_to_string":   true,
		"std.float_to_string": true,
		"std.bool_to_string":  true,
		"std.string_to_int":   true,
		"std.string_to_float": true,
		"std.string_to_bool":  true,
		"std.log.debug":       true,
		"std.log.info":        true,
		"std.log.warn":        true,
		"std.log.error":       true,
		"std.log.set_level":  true,
		"std.test.start":      true,
		"std.test.end":        true,
		"std.assert":          true,
		"test.start":          true,
		"test.end":            true,
		// Collections functions
		"std.collections.keys":   true,
		"std.collections.values": true,
		"std.collections.copy":   true,
		"std.collections.merge":  true,
		"std.collections.set_create":       true,
		"std.collections.set_add":           true,
		"std.collections.set_remove":        true,
		"std.collections.set_contains":      true,
		"std.collections.set_size":          true,
		"std.collections.set_clear":         true,
		"std.collections.set_union":         true,
		"std.collections.set_intersection": true,
		"std.collections.set_difference":   true,
		"std.collections.queue_create":   true,
		"std.collections.queue_enqueue":  true,
		"std.collections.queue_dequeue":  true,
		"std.collections.queue_peek":     true,
		"std.collections.queue_is_empty": true,
		"std.collections.queue_size":     true,
		"std.collections.queue_clear":    true,
		"std.collections.stack_create":   true,
		"std.collections.stack_push":     true,
		"std.collections.stack_pop":      true,
		"std.collections.stack_peek":     true,
		"std.collections.stack_is_empty": true,
		"std.collections.stack_size":     true,
		"std.collections.stack_clear":    true,
		"std.collections.priority_queue_create":    true,
		"std.collections.priority_queue_insert":     true,
		"std.collections.priority_queue_extract_max": true,
		"std.collections.priority_queue_peek":       true,
		"std.collections.priority_queue_is_empty":   true,
		"std.collections.priority_queue_size":       true,
		"std.collections.linked_list_create":   true,
		"std.collections.linked_list_append":   true,
		"std.collections.linked_list_prepend":  true,
		"std.collections.linked_list_insert":   true,
		"std.collections.linked_list_remove":  true,
		"std.collections.linked_list_get":     true,
		"std.collections.linked_list_set":      true,
		"std.collections.linked_list_size":    true,
		"std.collections.linked_list_is_empty": true,
		"std.collections.linked_list_clear":   true,
		"std.collections.binary_tree_create":   true,
		"std.collections.binary_tree_insert":   true,
		"std.collections.binary_tree_search":   true,
		"std.collections.binary_tree_remove":   true,
		"std.collections.binary_tree_size":    true,
		"std.collections.binary_tree_is_empty": true,
		"std.collections.binary_tree_clear":    true,
		// Network functions
		"std.network.ip_parse":         true,
		"std.network.ip_is_valid":      true,
		"std.network.ip_is_private":    true,
		"std.network.ip_is_loopback":   true,
		"std.network.ip_to_string":     true,
		"std.network.url_parse":        true,
		"std.network.url_to_string":    true,
		"std.network.url_is_valid":     true,
		"std.network.dns_lookup":       true,
		"std.network.dns_reverse_lookup": true,
		"std.network.http_get":         true,
		"std.network.http_post":        true,
		"std.network.http_put":         true,
		"std.network.http_delete":      true,
		"std.network.http_request":      true,
		"std.network.socket_create":    true,
		"std.network.socket_connect":   true,
		"std.network.socket_bind":      true,
		"std.network.socket_listen":    true,
		"std.network.socket_accept":    true,
		"std.network.socket_send":       true,
		"std.network.socket_receive":    true,
		"std.network.socket_close":     true,
		"std.network.network_is_connected": true,
		"std.network.network_get_local_ip":  true,
		"std.network.network_ping":          true,
	}

	return runtimeFunctions[funcName]
}

// isStdFunction checks if a function name looks like a standard library function
func (g *CGenerator) isStdFunction(funcName string) bool {
	return strings.HasPrefix(funcName, "std.") || 
		   strings.HasPrefix(funcName, "io.") ||
		   strings.HasPrefix(funcName, "math.") ||
		   strings.HasPrefix(funcName, "string.") ||
		   strings.HasPrefix(funcName, "array.") ||
		   strings.HasPrefix(funcName, "os.") ||
		   strings.HasPrefix(funcName, "collections.") ||
		   strings.HasPrefix(funcName, "file.") ||
		   strings.HasPrefix(funcName, "test.") ||
		   strings.HasPrefix(funcName, "network.")
}

// isStringReturningFunction checks if a function returns a heap-allocated string
// that needs to be freed by the caller
func (g *CGenerator) isStringReturningFunction(funcName string) bool {
	stringReturningFunctions := map[string]bool{
		"std.io.read_line":        true,
		"io.read_line":            true,
		"std.string.concat":       true,
		"std.string.substring":    true,
		"std.string.trim":         true,
		"std.string.to_upper":     true,
		"std.string.to_lower":     true,
		"std.int_to_string":       true,
		"std.float_to_string":     true,
		"std.bool_to_string":      true,
		"std.os.read_file":        true,
		"os.read_file":            true,
		"omni_read_line":          true,
		"omni_strcat":             true,
		"omni_substring":          true,
		"omni_trim":               true,
		"omni_to_upper":            true,
		"omni_to_lower":            true,
		"omni_int_to_string":      true,
		"omni_float_to_string":    true,
		"omni_bool_to_string":     true,
		"omni_read_file":          true,
		"omni_await_string":       true,
	}
	return stringReturningFunctions[funcName]
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

// extractMapTypes extracts key and value types from a map type string
func (g *CGenerator) extractMapTypes(mapType string) (keyType, valueType string) {
	if !strings.HasPrefix(mapType, "map<") || !strings.HasSuffix(mapType, ">") {
		return "string", "int" // Default fallback
	}
	inner := mapType[4 : len(mapType)-1] // Remove "map<" and ">"
	parts := strings.Split(inner, ",")
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return "string", "int" // Default fallback
}

// getMapPutFunction returns the appropriate map put function name for the given key and value types
func (g *CGenerator) getMapPutFunction(keyType, valueType string) string {
	// Normalize types (handle float/double, etc.)
	if valueType == "float" || valueType == "double" {
		valueType = "float"
	}
	if valueType == "bool" {
		valueType = "bool"
	}
	
	if keyType == "string" {
		switch valueType {
		case "int":
			return "omni_map_put_string_int"
		case "string":
			return "omni_map_put_string_string"
		case "float":
			return "omni_map_put_string_float"
		case "bool":
			return "omni_map_put_string_bool"
		}
	} else if keyType == "int" {
		switch valueType {
		case "int":
			return "omni_map_put_int_int"
		case "string":
			return "omni_map_put_int_string"
		case "float":
			return "omni_map_put_int_float"
		case "bool":
			return "omni_map_put_int_bool"
		}
	}
	return "" // Unsupported combination
}

// extractGenericType extracts the base name and type arguments from a generic type string
// e.g., "Box<int>" -> ("Box", ["int"]), "array<string>" -> ("array", ["string"])
func (g *CGenerator) extractGenericType(typeStr string) (baseName string, typeArgs []string) {
	if !strings.Contains(typeStr, "<") || !strings.HasSuffix(typeStr, ">") {
		return typeStr, nil
	}
	
	lessPos := strings.Index(typeStr, "<")
	baseName = typeStr[:lessPos]
	inner := typeStr[lessPos+1 : len(typeStr)-1] // Remove "<" and ">"
	
	// Split by comma, but handle nested generics
	typeArgs = g.splitGenericArgs(inner)
	return baseName, typeArgs
}

// splitGenericArgs splits generic arguments by comma, handling nested generics
func (g *CGenerator) splitGenericArgs(s string) []string {
	var args []string
	var current strings.Builder
	depth := 0
	
	for _, r := range s {
		switch r {
		case '<':
			depth++
			current.WriteRune(r)
		case '>':
			depth--
			current.WriteRune(r)
		case ',':
			if depth == 0 {
				args = append(args, strings.TrimSpace(current.String()))
				current.Reset()
			} else {
				current.WriteRune(r)
			}
		default:
			current.WriteRune(r)
		}
	}
	
	if current.Len() > 0 {
		args = append(args, strings.TrimSpace(current.String()))
	}
	
	return args
}

// getMapGetFunction returns the appropriate map get function name for the given key and value types
func (g *CGenerator) getMapGetFunction(keyType, valueType string) string {
	// Normalize types (handle float/double, etc.)
	if valueType == "float" || valueType == "double" {
		valueType = "float"
	}
	if valueType == "bool" {
		valueType = "bool"
	}
	
	if keyType == "string" {
		switch valueType {
		case "int":
			return "omni_map_get_string_int"
		case "string":
			return "omni_map_get_string_string"
		case "float":
			return "omni_map_get_string_float"
		case "bool":
			return "omni_map_get_string_bool"
		}
	} else if keyType == "int" {
		switch valueType {
		case "int":
			return "omni_map_get_int_int"
		case "string":
			return "omni_map_get_int_string"
		case "float":
			return "omni_map_get_int_float"
		case "bool":
			return "omni_map_get_int_bool"
		}
	}
	return "" // Unsupported combination
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
	// Find the main function to determine its return type
	var mainReturnType string = "int32_t" // Default
	for _, fn := range g.module.Functions {
		if fn.Name == "main" {
			mainReturnType = fn.ReturnType
			break
		}
	}
	
	g.output.WriteString("int main(int argc, char** argv) {\n")
	g.output.WriteString("    omni_args_init(argc, argv);\n")
	
	// Handle different return types
	if mainReturnType == "void" {
		g.output.WriteString("    omni_main();\n")
		g.output.WriteString("    printf(\"OmniLang program completed\\n\");\n")
		g.output.WriteString("    return 0;\n")
	} else if mainReturnType == "string" {
		g.output.WriteString("    const char* result = omni_main();\n")
		g.output.WriteString("    printf(\"OmniLang program result: %s\\n\", result ? result : \"(null)\");\n")
		g.output.WriteString("    return 0;\n")
	} else {
		// For int, float, bool, etc., use the mapped C type
		cReturnType := g.mapType(mainReturnType)
		g.output.WriteString(fmt.Sprintf("    %s result = omni_main();\n", cReturnType))
		if mainReturnType == "float" || mainReturnType == "double" {
			g.output.WriteString("    printf(\"OmniLang program result: %f\\n\", result);\n")
		} else {
			g.output.WriteString("    printf(\"OmniLang program result: %d\\n\", (int)result);\n")
		}
		g.output.WriteString("    return (int)result;\n")
	}
	
	g.output.WriteString("}\n")
}
