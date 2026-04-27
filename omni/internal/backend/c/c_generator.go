package cbackend

import (
	"fmt"
	"sort"
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
	// arrayLengthExprs holds C expressions (variable names, usually) that
	// evaluate to the runtime length of an array value. Populated for
	// array function parameters via the synthetic __len_<name> companion
	// param the codegen inserts. Looked up after arrayLengths misses.
	arrayLengthExprs map[mir.ValueID]string
	// userFuncArrayParams maps a user-defined function name (mangled
	// or unmangled) to the indices of its array<T> parameters. Used by
	// call sites to know where to inject the synthetic length argument.
	userFuncArrayParams map[string][]int
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
	// Track every value ID returned by this function. A single function can
	// have multiple return sites, and each returned heap value is caller-owned.
	returnedValueIDs map[mir.ValueID]bool
	// Track which variables were declared at the top of the function
	declaredVariables map[mir.ValueID]bool
	// currentDeferFn is the MIR function currently being generated. The defer
	// emitters read it to name per-site context structs and thunks
	// consistently between the pre-function thunk emission and the body
	// emission of defer.push.
	currentDeferFn *mir.Function
	// tupleStructs records every tuple type the program references; each
	// distinct type gets a typedef emitted once at file scope via
	// ensureTupleStruct + writeTupleStructDecls. Value is the C typedef
	// name. Keyed by the full OmniLang type string (e.g.
	// "tuple<int,bool>").
	tupleStructs map[string]string
	// tupleSplits records tuples that aren't actually stored as a
	// struct — specifically, the synthetic (value, ok) tuple that
	// chan.recv.ok produces. Keyed by the tuple-producing instruction's
	// ValueID, value is the per-component C variable names in order.
	// tuple.extract reads from these names instead of from a field.
	tupleSplits map[mir.ValueID][]string
}

// NewCGenerator creates a new C code generator
func NewCGenerator(module *mir.Module) *CGenerator {
	return &CGenerator{
		module:              module,
		optLevel:            "2", // Default to standard optimization
		debugInfo:           false,
		sourceFile:          "",
		variables:           make(map[mir.ValueID]string),
		phiVars:             make(map[mir.ValueID]bool),
		mutableVars:         make(map[mir.ValueID]bool),
		mapVars:             make(map[string]bool),
		mapTypes:            make(map[mir.ValueID]string),
		arrayLengths:        make(map[mir.ValueID]int),
		arrayLengthExprs:    make(map[mir.ValueID]string),
		userFuncArrayParams: make(map[string][]int),
		sourceMap:           make(map[string]int),
		lineMap:             make(map[int]string),
		valueTypes:          make(map[mir.ValueID]string),
		errors:              []string{},
		stringsToFree:       make(map[mir.ValueID]bool),
		promisesToFree:      make(map[mir.ValueID]bool),
		tempStringsToFree:   []string{},
		returnedValueID:     mir.InvalidValue,
		returnedValueIDs:    make(map[mir.ValueID]bool),
		declaredVariables:   make(map[mir.ValueID]bool),
		tupleStructs:        make(map[string]string),
		tupleSplits:         make(map[mir.ValueID][]string),
	}
}

// NewCGeneratorWithOptLevel creates a new C code generator with specified optimization level
func NewCGeneratorWithOptLevel(module *mir.Module, optLevel string) *CGenerator {
	return &CGenerator{
		module:              module,
		optLevel:            optLevel,
		debugInfo:           false,
		sourceFile:          "",
		variables:           make(map[mir.ValueID]string),
		phiVars:             make(map[mir.ValueID]bool),
		mutableVars:         make(map[mir.ValueID]bool),
		mapVars:             make(map[string]bool),
		mapTypes:            make(map[mir.ValueID]string),
		arrayLengths:        make(map[mir.ValueID]int),
		arrayLengthExprs:    make(map[mir.ValueID]string),
		userFuncArrayParams: make(map[string][]int),
		sourceMap:           make(map[string]int),
		lineMap:             make(map[int]string),
		valueTypes:          make(map[mir.ValueID]string),
		errors:              []string{},
		stringsToFree:       make(map[mir.ValueID]bool),
		promisesToFree:      make(map[mir.ValueID]bool),
		tempStringsToFree:   []string{},
		returnedValueID:     mir.InvalidValue,
		returnedValueIDs:    make(map[mir.ValueID]bool),
		declaredVariables:   make(map[mir.ValueID]bool),
		tupleStructs:        make(map[string]string),
		tupleSplits:         make(map[mir.ValueID][]string),
	}
}

// NewCGeneratorWithDebug creates a new C code generator with debug information
func NewCGeneratorWithDebug(module *mir.Module, optLevel string, debugInfo bool, sourceFile string) *CGenerator {
	return &CGenerator{
		module:              module,
		optLevel:            optLevel,
		debugInfo:           debugInfo,
		sourceFile:          sourceFile,
		variables:           make(map[mir.ValueID]string),
		phiVars:             make(map[mir.ValueID]bool),
		mutableVars:         make(map[mir.ValueID]bool),
		mapVars:             make(map[string]bool),
		mapTypes:            make(map[mir.ValueID]string),
		arrayLengths:        make(map[mir.ValueID]int),
		arrayLengthExprs:    make(map[mir.ValueID]string),
		userFuncArrayParams: make(map[string][]int),
		sourceMap:           make(map[string]int),
		lineMap:             make(map[int]string),
		valueTypes:          make(map[mir.ValueID]string),
		errors:              []string{},
		stringsToFree:       make(map[mir.ValueID]bool),
		promisesToFree:      make(map[mir.ValueID]bool),
		tempStringsToFree:   []string{},
		returnedValueID:     mir.InvalidValue,
		returnedValueIDs:    make(map[mir.ValueID]bool),
		declaredVariables:   make(map[mir.ValueID]bool),
		tupleStructs:        make(map[string]string),
		tupleSplits:         make(map[mir.ValueID][]string),
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
	// Pre-compute the indices of array<T> parameters for every user-
	// defined function so call-site emission knows where to splice in
	// the synthetic length argument.
	for _, fn := range g.module.Functions {
		var idx []int
		for i, p := range fn.Params {
			if isArrayParamType(p.Type) {
				idx = append(idx, i)
			}
		}
		if len(idx) > 0 {
			g.userFuncArrayParams[fn.Name] = idx
			g.userFuncArrayParams[g.mapFunctionName(fn.Name)] = idx
		}
	}
	// Intrinsics that take array<T> at parameter slot 0 also need the
	// synthetic length argument splice (sorts, searches, aggregates).
	// std.algorithms.linear_search / binary_search / count_occurrences
	// take (arr, target) so the array is still at index 0; the trailing
	// scalar follows the synthesized length cleanly.
	for _, name := range []string{
		"std.algorithms.bubble_sort",
		"std.algorithms.selection_sort",
		"std.algorithms.insertion_sort",
		"std.algorithms.linear_search",
		"std.algorithms.binary_search",
		"std.algorithms.find_max",
		"std.algorithms.find_min",
		"std.algorithms.count_occurrences",
		"std.algorithms.reverse",
		"std.algorithms.rotate",
	} {
		g.userFuncArrayParams[name] = []int{0}
	}

	g.writeHeader()
	g.writeStdLibFunctions()

	// Eager-scan every function's return type and every instruction's
	// result type for tuple shapes so writeTupleStructDecls can emit the
	// typedefs up-front. Without this pre-scan, mapType would populate
	// tupleStructs lazily during function-declaration emission — by which
	// point the struct would be referenced before it was defined.
	g.collectTupleStructsFromModule()
	g.writeTupleStructDecls()

	// Generate function declarations first
	g.writeFunctionDeclarations()

	// Emit the interface method lookup table (keyed by concrete-type + method
	// name) so iface.call dispatch can resolve the right function pointer at
	// runtime. Must come AFTER function declarations so the table can take
	// addresses of the generated methods.
	g.writeInterfaceDispatchSupport()

	// Then generate function definitions
	for _, fn := range g.module.Functions {
		// Emit defer thunks for this function ahead of the function body so
		// the body can reference them as static file-scope symbols.
		g.writeDeferThunksForFunction(fn)
		if err := g.generateFunction(fn); err != nil {
			return "", err
		}
	}

	g.writeMain()

	// Check for errors collected during code generation
	// Separate true errors from warnings so WARNING-prefixed diagnostics don't
	// fail compilation on their own.
	var hardErrors []string
	for _, e := range g.errors {
		if !strings.HasPrefix(e, "WARNING") {
			hardErrors = append(hardErrors, e)
		}
	}
	if len(hardErrors) > 0 {
		return "", fmt.Errorf("code generation errors:\n%s", strings.Join(hardErrors, "\n"))
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
			returnType = "int32_t" // Always use int32_t for omni_main to match runtime (even if async)
		} else if strings.HasPrefix(fn.ReturnType, "Promise<") {
			// For async functions (Promise<T>), the function should return omni_promise_t*
			// The Promise is created at the call site, but the function itself returns the promise
			returnType = "omni_promise_t*"
		}

		// Handle function pointer return types
		if strings.Contains(fn.ReturnType, ") -> ") {
			// This is a function pointer return type - need special handling
			g.output.WriteString(g.generateCompleteFunctionSignature(fn.ReturnType, funcName, fn.Params))
			g.output.WriteString(";\n")
		} else {
			g.output.WriteString(fmt.Sprintf("%s %s(", returnType, funcName))

			// Generate parameters. Mirrors the definition: each
			// array<T> parameter gets a synthetic int32_t length
			// companion right after it. See isArrayParamType.
			first := true
			for _, param := range fn.Params {
				if !first {
					g.output.WriteString(", ")
				}
				first = false
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
					if isArrayParamType(param.Type) {
						g.output.WriteString(fmt.Sprintf(", int32_t %s", arrayLenParamName(param.Name)))
					}
				}
			}
			g.output.WriteString(");\n")
		}
	}
	g.output.WriteString("\n")
}

// writeDeferThunksForFunction scans a function for defer.push instructions
// and, for each one, emits a file-scope context struct + thunk function that
// unpacks the snapshotted args and invokes the target. Emitted just before
// the enclosing function's definition.
func (g *CGenerator) writeDeferThunksForFunction(fn *mir.Function) {
	for _, block := range fn.Blocks {
		for _, inst := range block.Instructions {
			switch inst.Op {
			case "defer.push":
				g.writeDeferThunkNamed(fn, inst)
			case "defer.push.func":
				g.writeDeferThunkFunc(fn, inst)
			case "defer.push.iface":
				g.writeDeferThunkIface(fn, inst)
			case "spawn":
				g.writeSpawnThunk(fn, inst)
			}
		}
	}
}

// writeSpawnThunk emits a per-call-site pthread thunk for `spawn fn(args)`.
// The thunk signature matches `void* (*)(void*)` so it can be handed to
// pthread_create. It unpacks the heap-allocated context (which holds
// snapshotted argument values), invokes the target by its mangled C name,
// frees the context, and returns NULL — exactly the same shape as the
// per-site defer thunks. Lambda-typed (function-value) spawn callees use
// a `void* fn` slot that's cast to the right signature at call time.
func (g *CGenerator) writeSpawnThunk(fn *mir.Function, inst mir.Instruction) {
	if len(inst.Operands) == 0 {
		return
	}
	calleeOp := inst.Operands[0]
	argOps := inst.Operands[1:]
	ctxStructName := g.spawnCtxStructName(fn, inst.ID)
	thunkName := g.spawnThunkName(fn, inst.ID)

	// Context struct: holds either a function pointer (for func-valued
	// callees) or nothing extra (for named callees), plus one field per
	// snapshotted argument.
	g.output.WriteString("typedef struct {\n")
	if calleeOp.Kind == mir.OperandValue {
		g.output.WriteString("    void* fn;\n")
	}
	if calleeOp.Kind != mir.OperandValue && len(argOps) == 0 {
		g.output.WriteString("    char __omni_spawn_empty;\n")
	}
	for i, op := range argOps {
		g.output.WriteString(fmt.Sprintf("    %s a%d;\n", g.mapType(op.Type), i))
	}
	g.output.WriteString(fmt.Sprintf("} %s;\n", ctxStructName))

	g.output.WriteString(fmt.Sprintf("static void* %s(void* __raw) {\n", thunkName))
	g.output.WriteString(fmt.Sprintf("    %s* __ctx = (%s*)__raw;\n", ctxStructName, ctxStructName))

	switch calleeOp.Kind {
	case mir.OperandLiteral:
		calleeCName := g.mapFunctionName(calleeOp.Literal)
		g.output.WriteString(fmt.Sprintf("    %s(", calleeCName))
		for i := range argOps {
			if i > 0 {
				g.output.WriteString(", ")
			}
			g.output.WriteString(fmt.Sprintf("__ctx->a%d", i))
		}
		g.output.WriteString(");\n")
	case mir.OperandValue:
		// Function-value callee: cast __ctx->fn to the right signature
		// using the operand's MIR func type. Return type defaults to void
		// because spawn discards results.
		retType := "void"
		if r, ok := parseFuncReturnType(calleeOp.Type); ok {
			retType = r
		}
		retC := g.mapType(retType)
		var argCs []string
		for _, op := range argOps {
			argCs = append(argCs, g.mapType(op.Type))
		}
		g.output.WriteString(fmt.Sprintf("    ((%s(*)(%s))__ctx->fn)(",
			retC, strings.Join(argCs, ", ")))
		for i := range argOps {
			if i > 0 {
				g.output.WriteString(", ")
			}
			g.output.WriteString(fmt.Sprintf("__ctx->a%d", i))
		}
		g.output.WriteString(");\n")
	}
	g.output.WriteString("    free(__ctx);\n")
	g.output.WriteString("    return NULL;\n")
	g.output.WriteString("}\n\n")
}

func (g *CGenerator) spawnCtxStructName(fn *mir.Function, id mir.ValueID) string {
	return fmt.Sprintf("__omni_spawn_ctx_%s_%d_t", g.mapFunctionName(fn.Name), int(id))
}

func (g *CGenerator) spawnThunkName(fn *mir.Function, id mir.ValueID) string {
	return fmt.Sprintf("__omni_spawn_thunk_%s_%d", g.mapFunctionName(fn.Name), int(id))
}

// writeDeferThunkNamed handles `defer f(args)` / `defer Type.method(recv, args)`
// where the callee is a static symbol. Operand layout: [calleeLit, args...].
func (g *CGenerator) writeDeferThunkNamed(fn *mir.Function, inst mir.Instruction) {
	if len(inst.Operands) == 0 {
		return
	}
	calleeOp := inst.Operands[0]
	if calleeOp.Kind != mir.OperandLiteral {
		return
	}
	argOps := inst.Operands[1:]
	ctxStructName := g.deferCtxStructName(fn, inst.ID)
	thunkName := g.deferThunkName(fn, inst.ID)
	calleeCName := g.mapFunctionName(calleeOp.Literal)

	g.emitDeferCtxStruct(ctxStructName, nil, argOps)
	g.output.WriteString(fmt.Sprintf("static void %s(void* __raw) {\n", thunkName))
	g.output.WriteString(fmt.Sprintf("    %s* __ctx = (%s*)__raw;\n", ctxStructName, ctxStructName))
	g.output.WriteString(fmt.Sprintf("    %s(", calleeCName))
	for i := range argOps {
		if i > 0 {
			g.output.WriteString(", ")
		}
		g.output.WriteString(fmt.Sprintf("__ctx->a%d", i))
	}
	g.output.WriteString(");\n")
	g.output.WriteString("    free(__ctx);\n")
	g.output.WriteString("}\n\n")
}

// writeDeferThunkFunc handles `defer fn(args)` where fn is a function-valued
// local. Operand layout: [fnValue, args...]. The ctx stores the callable as
// a void* (the actual C type is a function pointer, but function-pointer
// syntax with embedded field names is awkward in a struct field position)
// and the thunk casts it to the right signature before invoking.
func (g *CGenerator) writeDeferThunkFunc(fn *mir.Function, inst mir.Instruction) {
	if len(inst.Operands) == 0 {
		return
	}
	fnOp := inst.Operands[0]
	argOps := inst.Operands[1:]
	ctxStructName := g.deferCtxStructName(fn, inst.ID)
	thunkName := g.deferThunkName(fn, inst.ID)

	g.output.WriteString(fmt.Sprintf("typedef struct {\n"))
	g.output.WriteString("    void* fn;\n")
	for i, op := range argOps {
		g.output.WriteString(fmt.Sprintf("    %s a%d;\n", g.mapType(op.Type), i))
	}
	g.output.WriteString(fmt.Sprintf("} %s;\n", ctxStructName))

	retType := "void"
	if r, ok := parseFuncReturnType(fnOp.Type); ok {
		retType = r
	}
	retC := g.mapType(retType)

	g.output.WriteString(fmt.Sprintf("static void %s(void* __raw) {\n", thunkName))
	g.output.WriteString(fmt.Sprintf("    %s* __ctx = (%s*)__raw;\n", ctxStructName, ctxStructName))
	g.output.WriteString(fmt.Sprintf("    ((%s(*)(", retC))
	for i, op := range argOps {
		if i > 0 {
			g.output.WriteString(", ")
		}
		g.output.WriteString(g.mapType(op.Type))
	}
	g.output.WriteString("))__ctx->fn)(")
	for i := range argOps {
		if i > 0 {
			g.output.WriteString(", ")
		}
		g.output.WriteString(fmt.Sprintf("__ctx->a%d", i))
	}
	g.output.WriteString(");\n")
	g.output.WriteString("    free(__ctx);\n")
	g.output.WriteString("}\n\n")
}

// writeDeferThunkIface handles `defer x.m(args)` where x is interface-typed.
// Operand layout: [ifaceLit, methodLit, recv, args...]. Dispatch is via the
// existing omni_method_lookup helper — the same mechanism iface.call uses.
func (g *CGenerator) writeDeferThunkIface(fn *mir.Function, inst mir.Instruction) {
	if len(inst.Operands) < 3 {
		return
	}
	ifaceOp := inst.Operands[0]
	methodOp := inst.Operands[1]
	if ifaceOp.Kind != mir.OperandLiteral || methodOp.Kind != mir.OperandLiteral {
		return
	}
	argOps := inst.Operands[3:]
	ctxStructName := g.deferCtxStructName(fn, inst.ID)
	thunkName := g.deferThunkName(fn, inst.ID)

	g.output.WriteString(fmt.Sprintf("typedef struct {\n"))
	g.output.WriteString("    omni_struct_t* recv;\n")
	for i, op := range argOps {
		g.output.WriteString(fmt.Sprintf("    %s a%d;\n", g.mapType(op.Type), i))
	}
	g.output.WriteString(fmt.Sprintf("} %s;\n", ctxStructName))

	// Look up the method's signature from the module's interface metadata so
	// we can cast the void* returned by omni_method_lookup to the precise
	// function-pointer type.
	var method *mir.InterfaceMethod
	for _, iface := range g.module.Interfaces {
		if iface.Name != ifaceOp.Literal {
			continue
		}
		for i, m := range iface.Methods {
			if m.Name == methodOp.Literal {
				method = &iface.Methods[i]
				break
			}
		}
	}
	retC := "void"
	var argCs []string
	if method != nil {
		retC = g.mapType(method.ReturnType)
		for _, pt := range method.ParamTypes {
			argCs = append(argCs, g.mapType(pt))
		}
	} else {
		// Fallback: types from the actual operands. Safe but less precise.
		for _, op := range argOps {
			argCs = append(argCs, g.mapType(op.Type))
		}
	}
	castParams := append([]string{"omni_struct_t*"}, argCs...)

	g.output.WriteString(fmt.Sprintf("static void %s(void* __raw) {\n", thunkName))
	g.output.WriteString(fmt.Sprintf("    %s* __ctx = (%s*)__raw;\n", ctxStructName, ctxStructName))
	g.output.WriteString(fmt.Sprintf("    ((%s(*)(%s))omni_method_lookup(omni_struct_get_type_name(__ctx->recv), \"%s\"))(__ctx->recv",
		retC, strings.Join(castParams, ", "), methodOp.Literal))
	for i := range argOps {
		g.output.WriteString(fmt.Sprintf(", __ctx->a%d", i))
	}
	g.output.WriteString(");\n")
	g.output.WriteString("    free(__ctx);\n")
	g.output.WriteString("}\n\n")
}

// emitDeferCtxStruct writes a `typedef struct { ... } <name>;` for the named
// defer kind. leadingFields is an optional list of pre-arg fields (none for
// the plain named kind); argOps is the list of value operands whose types
// become `a0`, `a1`, etc.
func (g *CGenerator) emitDeferCtxStruct(name string, leadingFields []string, argOps []mir.Operand) {
	g.output.WriteString(fmt.Sprintf("typedef struct {\n"))
	if len(leadingFields) == 0 && len(argOps) == 0 {
		// C forbids zero-length structs in strict modes; include a padding
		// byte so the declaration is always legal.
		g.output.WriteString("    char __omni_defer_empty;\n")
	}
	for _, f := range leadingFields {
		g.output.WriteString(fmt.Sprintf("    %s\n", f))
	}
	for i, op := range argOps {
		g.output.WriteString(fmt.Sprintf("    %s a%d;\n", g.mapType(op.Type), i))
	}
	g.output.WriteString(fmt.Sprintf("} %s;\n", name))
}

// parseFuncReturnType extracts the return-type substring from a MIR function
// type `(p1, p2) -> ret`. Returns ("", false) if the input does not look
// like one — callers treat that as "unknown return type" and default to void.
func parseFuncReturnType(funcType string) (string, bool) {
	idx := strings.Index(funcType, ") -> ")
	if idx == -1 {
		return "", false
	}
	return strings.TrimSpace(funcType[idx+len(") -> "):]), true
}

// functionHasDefer returns true if the given function contains at least one
// defer.push instruction. The C backend uses this to decide whether to emit
// the per-function omni_defer_frame_t variable.
func functionHasDefer(fn *mir.Function) bool {
	for _, block := range fn.Blocks {
		for _, inst := range block.Instructions {
			switch inst.Op {
			case "defer.push", "defer.push.func", "defer.push.iface":
				return true
			}
		}
	}
	return false
}

func (g *CGenerator) deferCtxStructName(fn *mir.Function, id mir.ValueID) string {
	return fmt.Sprintf("__omni_defer_ctx_%s_%d_t", g.mapFunctionName(fn.Name), int(id))
}

func (g *CGenerator) deferThunkName(fn *mir.Function, id mir.ValueID) string {
	return fmt.Sprintf("__omni_defer_thunk_%s_%d", g.mapFunctionName(fn.Name), int(id))
}

// emitDeferPush emits the C code for a defer.push instruction: allocate a
// context struct, fill it from the operands' current values, and register it
// with the enclosing function's omni_defer_frame.
func (g *CGenerator) emitDeferPush(inst *mir.Instruction) error {
	if len(inst.Operands) == 0 {
		return fmt.Errorf("defer.push: missing callee operand")
	}
	argOps := inst.Operands[1:]
	ctxStructName := g.deferCtxStructName(g.currentDeferFn, inst.ID)
	thunkName := g.deferThunkName(g.currentDeferFn, inst.ID)

	g.output.WriteString("  {\n")
	g.output.WriteString(fmt.Sprintf("    %s* __ctx = (%s*)malloc(sizeof(*__ctx));\n", ctxStructName, ctxStructName))
	for i, op := range argOps {
		g.output.WriteString(fmt.Sprintf("    __ctx->a%d = %s;\n", i, g.getOperandValue(op)))
	}
	g.output.WriteString(fmt.Sprintf("    omni_defer_push(&__omni_defer_frame, %s, __ctx);\n", thunkName))
	g.output.WriteString("  }\n")
	return nil
}

// emitDeferPushFunc emits the defer.push.func site: snapshot the function
// value (as a void*) and the args into a ctx, push the site's thunk.
func (g *CGenerator) emitDeferPushFunc(inst *mir.Instruction) error {
	if len(inst.Operands) == 0 {
		return fmt.Errorf("defer.push.func: missing callee operand")
	}
	fnOp := inst.Operands[0]
	argOps := inst.Operands[1:]
	ctxStructName := g.deferCtxStructName(g.currentDeferFn, inst.ID)
	thunkName := g.deferThunkName(g.currentDeferFn, inst.ID)

	g.output.WriteString("  {\n")
	g.output.WriteString(fmt.Sprintf("    %s* __ctx = (%s*)malloc(sizeof(*__ctx));\n", ctxStructName, ctxStructName))
	g.output.WriteString(fmt.Sprintf("    __ctx->fn = (void*)%s;\n", g.getOperandValue(fnOp)))
	for i, op := range argOps {
		g.output.WriteString(fmt.Sprintf("    __ctx->a%d = %s;\n", i, g.getOperandValue(op)))
	}
	g.output.WriteString(fmt.Sprintf("    omni_defer_push(&__omni_defer_frame, %s, __ctx);\n", thunkName))
	g.output.WriteString("  }\n")
	return nil
}

// emitDeferPushIface emits the defer.push.iface site: snapshot the receiver
// + args and push the site's thunk (which does the runtime method lookup).
func (g *CGenerator) emitDeferPushIface(inst *mir.Instruction) error {
	if len(inst.Operands) < 3 {
		return fmt.Errorf("defer.push.iface: expected at least ifaceName, methodName, receiver")
	}
	recvOp := inst.Operands[2]
	argOps := inst.Operands[3:]
	ctxStructName := g.deferCtxStructName(g.currentDeferFn, inst.ID)
	thunkName := g.deferThunkName(g.currentDeferFn, inst.ID)

	g.output.WriteString("  {\n")
	g.output.WriteString(fmt.Sprintf("    %s* __ctx = (%s*)malloc(sizeof(*__ctx));\n", ctxStructName, ctxStructName))
	g.output.WriteString(fmt.Sprintf("    __ctx->recv = %s;\n", g.getOperandValue(recvOp)))
	for i, op := range argOps {
		g.output.WriteString(fmt.Sprintf("    __ctx->a%d = %s;\n", i, g.getOperandValue(op)))
	}
	g.output.WriteString(fmt.Sprintf("    omni_defer_push(&__omni_defer_frame, %s, __ctx);\n", thunkName))
	g.output.WriteString("  }\n")
	return nil
}

// emitDeferRun emits the flush call: `omni_defer_run_all(&__omni_defer_frame);`.
// The MIR builder places this immediately before every `ret` terminator, so
// it naturally runs in LIFO order ahead of the actual return.
func (g *CGenerator) emitDeferRun() {
	g.output.WriteString("  omni_defer_run_all(&__omni_defer_frame);\n")
}

// emitSliceAppend lowers `append(slice, elem)` into a call to
// omni_slice_append. The runtime grows-or-reuses the heap allocation and
// returns the (possibly new) data pointer, which we cast back to the slice's
// element type and store into the result variable. We pass `&tmp` as the
// element address so we can append by-value through a generic memcpy without
// the runtime caring about the element type.
func (g *CGenerator) emitSliceAppend(inst *mir.Instruction) error {
	if len(inst.Operands) != 2 {
		return fmt.Errorf("slice.append: expected (slice, elem) operands, got %d", len(inst.Operands))
	}
	sliceOp := inst.Operands[0]
	elemOp := inst.Operands[1]

	elemTypeStr := g.elementTypeOf(sliceOp.Type)
	elemC := g.mapType(elemTypeStr)
	if !g.isPrimitiveType(elemTypeStr) && !strings.Contains(elemTypeStr, "<") && !strings.Contains(elemTypeStr, "(") {
		elemC = "omni_struct_t*"
	}

	varName := g.getVariableName(inst.ID)
	sliceExpr := g.getOperandValue(sliceOp)
	elemExpr := g.getOperandValue(elemOp)

	// Compound block scopes the temporary so back-to-back appends don't
	// fight over a `__elem` name. Casting append's return back to elemC* is
	// safe because the runtime preserves elem_size from the original
	// allocation — same backing layout, possibly new pointer.
	g.output.WriteString("  {\n")
	g.output.WriteString(fmt.Sprintf("    %s __elem = %s;\n", elemC, elemExpr))
	if g.declaredVariables[inst.ID] {
		g.output.WriteString(fmt.Sprintf("    %s = (%s*)omni_slice_append(%s, &__elem);\n", varName, elemC, sliceExpr))
	} else {
		g.output.WriteString(fmt.Sprintf("    %s* %s = (%s*)omni_slice_append(%s, &__elem);\n", elemC, varName, elemC, sliceExpr))
		g.declaredVariables[inst.ID] = true
	}
	g.output.WriteString("  }\n")
	g.valueTypes[inst.ID] = inst.Type
	return nil
}

// emitSliceSlice lowers `target[low:high]` into a call to
// omni_slice_subslice. The MIR builder represents missing bounds with a
// literal "0" (low) or "-1" (high), so we just pass operands through to the
// runtime which interprets -1 as "len(target)".
func (g *CGenerator) emitSliceSlice(inst *mir.Instruction) error {
	if len(inst.Operands) != 3 {
		return fmt.Errorf("slice.slice: expected (target, low, high) operands, got %d", len(inst.Operands))
	}
	targetOp := inst.Operands[0]
	lowOp := inst.Operands[1]
	highOp := inst.Operands[2]

	elemTypeStr := g.elementTypeOf(targetOp.Type)
	elemC := g.mapType(elemTypeStr)
	if !g.isPrimitiveType(elemTypeStr) && !strings.Contains(elemTypeStr, "<") && !strings.Contains(elemTypeStr, "(") {
		elemC = "omni_struct_t*"
	}

	varName := g.getVariableName(inst.ID)
	tgtExpr := g.getOperandValue(targetOp)
	lowExpr := g.getOperandValue(lowOp)
	highExpr := g.getOperandValue(highOp)

	if g.declaredVariables[inst.ID] {
		g.output.WriteString(fmt.Sprintf("  %s = (%s*)omni_slice_subslice(%s, (int64_t)(%s), (int64_t)(%s));\n",
			varName, elemC, tgtExpr, lowExpr, highExpr))
	} else {
		g.output.WriteString(fmt.Sprintf("  %s* %s = (%s*)omni_slice_subslice(%s, (int64_t)(%s), (int64_t)(%s));\n",
			elemC, varName, elemC, tgtExpr, lowExpr, highExpr))
		g.declaredVariables[inst.ID] = true
	}
	g.valueTypes[inst.ID] = inst.Type
	return nil
}

// chanElementTypeOf extracts T from a channel type `chan<T>`. Falls back to
// "int" so the generated C is well-formed; the checker has already gated
// channel ops to actual channel-typed operands.
func (g *CGenerator) chanElementTypeOf(chanType string) string {
	if strings.HasPrefix(chanType, "chan<") && strings.HasSuffix(chanType, ">") {
		return chanType[5 : len(chanType)-1]
	}
	return "int"
}

// emitSpawn emits the call site for a `spawn fn(args)` instruction. The
// thunk has already been emitted at file scope by writeSpawnThunk; here we
// allocate the heap context, copy in snapshotted args, and hand it off to
// omni_spawn (a thin wrapper over pthread_create that detaches the thread).
func (g *CGenerator) emitSpawn(inst *mir.Instruction) error {
	if len(inst.Operands) == 0 {
		return fmt.Errorf("spawn: missing callee operand")
	}
	calleeOp := inst.Operands[0]
	argOps := inst.Operands[1:]
	ctxStructName := g.spawnCtxStructName(g.currentDeferFn, inst.ID)
	thunkName := g.spawnThunkName(g.currentDeferFn, inst.ID)

	g.output.WriteString("  {\n")
	g.output.WriteString(fmt.Sprintf("    %s* __ctx = (%s*)malloc(sizeof(*__ctx));\n", ctxStructName, ctxStructName))
	if calleeOp.Kind == mir.OperandValue {
		g.output.WriteString(fmt.Sprintf("    __ctx->fn = (void*)%s;\n", g.getOperandValue(calleeOp)))
	}
	for i, op := range argOps {
		g.output.WriteString(fmt.Sprintf("    __ctx->a%d = %s;\n", i, g.getOperandValue(op)))
	}
	g.output.WriteString(fmt.Sprintf("    omni_spawn(%s, __ctx);\n", thunkName))
	g.output.WriteString("  }\n")
	return nil
}

// emitChanMake lowers `make(chan T[, cap])` into omni_chan_make. Operand
// layout: [elemTypeLit, capValue?]. Cap defaults to 0 (which the runtime
// treats as a single-slot buffered channel — see omni_chan_make's comment).
func (g *CGenerator) emitChanMake(inst *mir.Instruction) error {
	if len(inst.Operands) == 0 {
		return fmt.Errorf("chan.make: missing element-type operand")
	}
	elemOp := inst.Operands[0]
	if elemOp.Kind != mir.OperandLiteral {
		return fmt.Errorf("chan.make: element type must be a literal operand")
	}
	elemC := g.mapType(elemOp.Literal)
	capExpr := "0"
	if len(inst.Operands) >= 2 {
		capExpr = g.getOperandValue(inst.Operands[1])
	}
	varName := g.getVariableName(inst.ID)
	if g.declaredVariables[inst.ID] {
		g.output.WriteString(fmt.Sprintf("  %s = omni_chan_make((int64_t)(%s), sizeof(%s));\n", varName, capExpr, elemC))
	} else {
		g.output.WriteString(fmt.Sprintf("  omni_chan_t* %s = omni_chan_make((int64_t)(%s), sizeof(%s));\n", varName, capExpr, elemC))
		g.declaredVariables[inst.ID] = true
	}
	g.valueTypes[inst.ID] = inst.Type
	return nil
}

// emitChanSend lowers `c <- v` into omni_chan_send. Snapshots the value
// into a typed temporary so we can pass &temp through the runtime's
// generic memcpy interface without caring about the element type.
func (g *CGenerator) emitChanSend(inst *mir.Instruction) error {
	if len(inst.Operands) != 2 {
		return fmt.Errorf("chan.send: expected (chan, value) operands, got %d", len(inst.Operands))
	}
	chOp := inst.Operands[0]
	valOp := inst.Operands[1]
	elemC := g.mapType(g.chanElementTypeOf(chOp.Type))
	chExpr := g.getOperandValue(chOp)
	valExpr := g.getOperandValue(valOp)
	g.output.WriteString("  {\n")
	g.output.WriteString(fmt.Sprintf("    %s __elem = %s;\n", elemC, valExpr))
	g.output.WriteString(fmt.Sprintf("    omni_chan_send(%s, &__elem);\n", chExpr))
	g.output.WriteString("  }\n")
	return nil
}

// emitChanRecv lowers `<-c` into omni_chan_recv. Declares a typed result
// variable and passes its address as the destination buffer.
func (g *CGenerator) emitChanRecv(inst *mir.Instruction) error {
	if len(inst.Operands) != 1 {
		return fmt.Errorf("chan.recv: expected (chan) operand, got %d", len(inst.Operands))
	}
	chOp := inst.Operands[0]
	elemC := g.mapType(g.chanElementTypeOf(chOp.Type))
	varName := g.getVariableName(inst.ID)
	chExpr := g.getOperandValue(chOp)
	if g.declaredVariables[inst.ID] {
		g.output.WriteString(fmt.Sprintf("  omni_chan_recv(%s, &%s);\n", chExpr, varName))
	} else {
		g.output.WriteString(fmt.Sprintf("  %s %s;\n", elemC, varName))
		g.output.WriteString(fmt.Sprintf("  omni_chan_recv(%s, &%s);\n", chExpr, varName))
		g.declaredVariables[inst.ID] = true
	}
	g.valueTypes[inst.ID] = inst.Type
	return nil
}

// emitChanRecvOk lowers `v, ok = <-c` via the runtime's ok-form helper.
// The op's ID maps to a pair of logical outputs: the received element and
// a bool flag. We synthesize two concrete C variables (<id>_val and
// <id>_ok) so the tuple-destructure pass can bind user-level `v` and `ok`
// to them.
func (g *CGenerator) emitChanRecvOk(inst *mir.Instruction) error {
	if len(inst.Operands) != 1 {
		return fmt.Errorf("chan.recv.ok: expected (chan) operand, got %d", len(inst.Operands))
	}
	chOp := inst.Operands[0]
	elemC := g.mapType(g.chanElementTypeOf(chOp.Type))
	baseName := g.getVariableName(inst.ID)
	valName := baseName + "_val"
	okName := baseName + "_ok"
	chExpr := g.getOperandValue(chOp)
	g.output.WriteString(fmt.Sprintf("  %s %s;\n", elemC, valName))
	g.output.WriteString(fmt.Sprintf("  int32_t %s;\n", okName))
	g.output.WriteString(fmt.Sprintf("  omni_chan_recv_ok(%s, &%s, &%s);\n", chExpr, valName, okName))
	// Register a synthetic split: subsequent tuple.extract ops on this
	// ValueID read from the pair of vars above instead of going through
	// a real tuple struct. Saves an unnecessary struct copy for the
	// ok-form hot path.
	g.tupleSplits[inst.ID] = []string{valName, okName}
	g.valueTypes[inst.ID] = inst.Type
	return nil
}

// emitSelect lowers the `select` MIR op into a runtime call. The op's
// operand layout is:
//
//	[0]: literal case count
//	Then 6 operands per case:
//	  [0] kind literal: "send" | "recv" | "recv.ok" | "default"
//	  [1] channel value (or "-" for default)
//	  [2] send-value (send only; "-" otherwise)
//	  [3] recv-dest SSA-id literal "%N" (recv / recv.ok only; "-" otherwise)
//	  [4] recv-ok-dest SSA-id literal (recv.ok only; "-" otherwise)
//	  [5] body-block literal (consumed by the cbr chain after this op)
//
// We emit a stack-allocated omni_select_case_t array, populate it per
// case (staging send values into typed temps so we can pass &temp),
// call omni_select, and store the chosen index into the SSA slot. The
// cbr chain the MIR builder emitted after the select dispatches to the
// correct body block.
func (g *CGenerator) emitSelect(inst *mir.Instruction) error {
	if len(inst.Operands) == 0 {
		return fmt.Errorf("select: missing case count")
	}
	countOp := inst.Operands[0]
	if countOp.Kind != mir.OperandLiteral {
		return fmt.Errorf("select: case count must be literal")
	}
	caseCount := 0
	if _, err := fmt.Sscanf(countOp.Literal, "%d", &caseCount); err != nil {
		return fmt.Errorf("select: invalid case count %q", countOp.Literal)
	}
	if 1+caseCount*6 != len(inst.Operands) {
		return fmt.Errorf("select: expected %d operands for %d cases, got %d", 1+caseCount*6, caseCount, len(inst.Operands))
	}

	// Parse the SSA-id literal "%N" the MIR builder emits for recv/ok
	// destinations into the corresponding C variable name. Returns ""
	// for the sentinel "-" so callers know the slot is unused.
	parseDestName := func(lit string) string {
		if lit == "-" || !strings.HasPrefix(lit, "%") {
			return ""
		}
		var id int
		if _, err := fmt.Sscanf(lit[1:], "%d", &id); err != nil {
			return ""
		}
		return g.getVariableName(mir.ValueID(id))
	}

	varName := g.getVariableName(inst.ID)

	// The recv-dest SSA ids the MIR builder allocated for this select
	// are never produced by a "defining" instruction, so the pre-decl
	// pass doesn't declare variables for them. Emit plain local decls
	// for each before the select body so the post-select handoff
	// assignment has something to write into.
	base := 1
	for i := 0; i < caseCount; i++ {
		kindOp := inst.Operands[base+0]
		chOp := inst.Operands[base+1]
		recvDestOp := inst.Operands[base+3]
		recvOkOp := inst.Operands[base+4]
		base += 6
		switch kindOp.Literal {
		case "recv", "recv.ok":
			destName := parseDestName(recvDestOp.Literal)
			if destName != "" {
				// Look up the recv-dest SSA id so we can check whether
				// pre-decl already declared it.
				var id mir.ValueID
				fmt.Sscanf(recvDestOp.Literal[1:], "%d", (*int)(&id))
				if !g.declaredVariables[id] {
					elemC := g.mapType(g.chanElementTypeOf(chOp.Type))
					g.output.WriteString(fmt.Sprintf("  %s %s = 0;\n", elemC, destName))
					g.declaredVariables[id] = true
				}
			}
			if kindOp.Literal == "recv.ok" {
				okName := parseDestName(recvOkOp.Literal)
				if okName != "" {
					var id mir.ValueID
					fmt.Sscanf(recvOkOp.Literal[1:], "%d", (*int)(&id))
					if !g.declaredVariables[id] {
						g.output.WriteString(fmt.Sprintf("  int32_t %s = 0;\n", okName))
						g.declaredVariables[id] = true
					}
				}
			}
		}
	}

	g.output.WriteString("  {\n")
	g.output.WriteString(fmt.Sprintf("    omni_select_case_t __omni_sel_cases[%d];\n", caseCount))

	base = 1
	sendTempIdx := 0
	for i := 0; i < caseCount; i++ {
		kindOp := inst.Operands[base+0]
		chOp := inst.Operands[base+1]
		sendOp := inst.Operands[base+2]
		recvDestOp := inst.Operands[base+3]
		recvOkOp := inst.Operands[base+4]
		// body-block at base+5 is consumed by the dispatching cbr.
		base += 6

		g.output.WriteString(fmt.Sprintf("    __omni_sel_cases[%d].kind = ", i))
		switch kindOp.Literal {
		case "send":
			g.output.WriteString("OMNI_SELECT_KIND_SEND;\n")
		case "recv":
			g.output.WriteString("OMNI_SELECT_KIND_RECV;\n")
		case "recv.ok":
			g.output.WriteString("OMNI_SELECT_KIND_RECV_OK;\n")
		case "default":
			g.output.WriteString("OMNI_SELECT_KIND_DEFAULT;\n")
		default:
			return fmt.Errorf("select: unknown case kind %q", kindOp.Literal)
		}

		if kindOp.Literal != "default" {
			g.output.WriteString(fmt.Sprintf("    __omni_sel_cases[%d].ch = %s;\n", i, g.getOperandValue(chOp)))
		} else {
			g.output.WriteString(fmt.Sprintf("    __omni_sel_cases[%d].ch = NULL;\n", i))
		}

		switch kindOp.Literal {
		case "send":
			elemC := g.mapType(g.chanElementTypeOf(chOp.Type))
			tempName := fmt.Sprintf("__omni_sel_send_%d", sendTempIdx)
			sendTempIdx++
			g.output.WriteString(fmt.Sprintf("    %s %s = %s;\n", elemC, tempName, g.getOperandValue(sendOp)))
			g.output.WriteString(fmt.Sprintf("    __omni_sel_cases[%d].send_value = &%s;\n", i, tempName))
			g.output.WriteString(fmt.Sprintf("    __omni_sel_cases[%d].recv_dest = NULL;\n", i))
			g.output.WriteString(fmt.Sprintf("    __omni_sel_cases[%d].recv_ok = NULL;\n", i))
		case "recv":
			destName := parseDestName(recvDestOp.Literal)
			if destName == "" {
				return fmt.Errorf("select: recv case %d missing destination", i)
			}
			elemC := g.mapType(g.chanElementTypeOf(chOp.Type))
			// Declare the destination up-front so the body block can
			// read it. Using a fresh, explicitly-declared local here
			// sidesteps the pre-decl pass for this SSA slot — the
			// assignment via &destName into omni_select supplies the
			// value.
			g.output.WriteString(fmt.Sprintf("    %s %s_recv;\n", elemC, destName))
			g.output.WriteString(fmt.Sprintf("    __omni_sel_cases[%d].send_value = NULL;\n", i))
			g.output.WriteString(fmt.Sprintf("    __omni_sel_cases[%d].recv_dest = &%s_recv;\n", i, destName))
			g.output.WriteString(fmt.Sprintf("    __omni_sel_cases[%d].recv_ok = NULL;\n", i))
		case "recv.ok":
			destName := parseDestName(recvDestOp.Literal)
			okName := parseDestName(recvOkOp.Literal)
			if destName == "" || okName == "" {
				return fmt.Errorf("select: recv.ok case %d missing destination(s)", i)
			}
			elemC := g.mapType(g.chanElementTypeOf(chOp.Type))
			g.output.WriteString(fmt.Sprintf("    %s %s_recv;\n", elemC, destName))
			g.output.WriteString(fmt.Sprintf("    int32_t %s_ok = 0;\n", okName))
			g.output.WriteString(fmt.Sprintf("    __omni_sel_cases[%d].send_value = NULL;\n", i))
			g.output.WriteString(fmt.Sprintf("    __omni_sel_cases[%d].recv_dest = &%s_recv;\n", i, destName))
			g.output.WriteString(fmt.Sprintf("    __omni_sel_cases[%d].recv_ok = &%s_ok;\n", i, okName))
		case "default":
			g.output.WriteString(fmt.Sprintf("    __omni_sel_cases[%d].send_value = NULL;\n", i))
			g.output.WriteString(fmt.Sprintf("    __omni_sel_cases[%d].recv_dest = NULL;\n", i))
			g.output.WriteString(fmt.Sprintf("    __omni_sel_cases[%d].recv_ok = NULL;\n", i))
		}
	}

	// Call omni_select and capture the chosen index into the SSA slot.
	if g.declaredVariables[inst.ID] {
		g.output.WriteString(fmt.Sprintf("    %s = omni_select(%d, __omni_sel_cases);\n", varName, caseCount))
	} else {
		g.output.WriteString(fmt.Sprintf("    int32_t %s = omni_select(%d, __omni_sel_cases);\n", varName, caseCount))
		g.declaredVariables[inst.ID] = true
	}

	// Copy recv/ok temps out to the SSA-id-named variables that the
	// body blocks read. We did this roundabout dance (temp → named var)
	// because the pre-decl pass already declares `int32_t v<N>;` for
	// the recv-dest SSA ids; we can't redeclare them here, so we use
	// `<name>_recv` as the handoff temp and then assign.
	base = 1
	for i := 0; i < caseCount; i++ {
		kindOp := inst.Operands[base+0]
		recvDestOp := inst.Operands[base+3]
		recvOkOp := inst.Operands[base+4]
		base += 6
		switch kindOp.Literal {
		case "recv":
			destName := parseDestName(recvDestOp.Literal)
			if destName != "" {
				g.output.WriteString(fmt.Sprintf("    if (%s == %d) %s = %s_recv;\n", varName, i, destName, destName))
			}
		case "recv.ok":
			destName := parseDestName(recvDestOp.Literal)
			okName := parseDestName(recvOkOp.Literal)
			if destName != "" {
				g.output.WriteString(fmt.Sprintf("    if (%s == %d) { %s = %s_recv; %s = %s_ok; }\n",
					varName, i, destName, destName, okName, okName))
			}
		}
	}
	g.output.WriteString("  }\n")
	g.valueTypes[inst.ID] = inst.Type
	return nil
}

// emitChanClose lowers `close(c)` into the runtime helper.
func (g *CGenerator) emitChanClose(inst *mir.Instruction) error {
	if len(inst.Operands) != 1 {
		return fmt.Errorf("chan.close: expected (chan) operand, got %d", len(inst.Operands))
	}
	chExpr := g.getOperandValue(inst.Operands[0])
	g.output.WriteString(fmt.Sprintf("  omni_chan_close(%s);\n", chExpr))
	return nil
}

// tupleStructName returns the C typedef name for a tuple type. Name is
// built from the component C types so the same shape always maps to the
// same struct (no aliasing by source-level name).
func (g *CGenerator) tupleStructName(tupleType string) string {
	if !strings.HasPrefix(tupleType, "tuple<") || !strings.HasSuffix(tupleType, ">") {
		return "omni_tuple_unknown_t"
	}
	inner := tupleType[len("tuple<") : len(tupleType)-1]
	parts := splitGenericTypeArgs(inner)
	var buf strings.Builder
	buf.WriteString("omni_tuple")
	for _, p := range parts {
		buf.WriteByte('_')
		buf.WriteString(sanitizeForIdent(g.mapType(strings.TrimSpace(p))))
	}
	buf.WriteString("_t")
	return buf.String()
}

// ensureTupleStruct registers a tuple type so its typedef gets emitted by
// writeTupleStructDecls. Idempotent — safe to call repeatedly. Keyed by
// the resulting C struct name (not the OmniLang tuple string), so that
// `tuple<int,bool>` and `tuple<int,int>` — which both map to the same C
// layout because `bool` and `int` are both int32_t — share one typedef
// instead of producing a clashing redefinition.
func (g *CGenerator) ensureTupleStruct(tupleType string) {
	name := g.tupleStructName(tupleType)
	if _, ok := g.tupleStructs[name]; ok {
		return
	}
	g.tupleStructs[name] = tupleType
}

// collectTupleStructsFromModule walks the whole module looking for tuple
// types in function return slots, parameters, and instruction result
// types. Pre-populates tupleStructs so writeTupleStructDecls can emit all
// typedefs at file scope before any user function references them.
func (g *CGenerator) collectTupleStructsFromModule() {
	consider := func(t string) {
		if strings.HasPrefix(t, "tuple<") && strings.HasSuffix(t, ">") {
			g.ensureTupleStruct(t)
		}
	}
	for _, fn := range g.module.Functions {
		consider(fn.ReturnType)
		for _, p := range fn.Params {
			consider(p.Type)
		}
		for _, block := range fn.Blocks {
			for _, inst := range block.Instructions {
				consider(inst.Type)
				for _, op := range inst.Operands {
					consider(op.Type)
				}
			}
		}
	}
}

// writeTupleStructDecls emits `typedef struct { T1 v0; T2 v1; ... } name_t;`
// for every tuple shape the module uses. Called once, after
// collectTupleStructsFromModule, before writeFunctionDeclarations.
func (g *CGenerator) writeTupleStructDecls() {
	if len(g.tupleStructs) == 0 {
		return
	}
	// Deterministic order keeps generated output diff-stable across runs.
	var names []string
	for n := range g.tupleStructs {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, name := range names {
		tupleType := g.tupleStructs[name]
		inner := tupleType[len("tuple<") : len(tupleType)-1]
		parts := splitGenericTypeArgs(inner)
		g.output.WriteString("typedef struct {\n")
		for i, p := range parts {
			g.output.WriteString(fmt.Sprintf("    %s v%d;\n", g.mapType(strings.TrimSpace(p)), i))
		}
		g.output.WriteString(fmt.Sprintf("} %s;\n", name))
	}
	g.output.WriteString("\n")
}

// sanitizeForIdent replaces characters not legal in C identifiers with '_'.
// Used when building tuple-struct names from component C types.
func sanitizeForIdent(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}

// splitGenericTypeArgs splits a comma-separated generic inner list,
// respecting angle-bracket nesting so `tuple<int, map<string,int>>`
// returns ["int", "map<string,int>"].
func splitGenericTypeArgs(inner string) []string {
	var parts []string
	depth := 0
	start := 0
	for i, r := range inner {
		switch r {
		case '<':
			depth++
		case '>':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, inner[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, inner[start:])
	return parts
}

// emitTupleNew lowers `tuple.new v0, v1, ...` into a compound literal
// assignment: the result is the tuple struct built from the operand
// values. Used at `return a, b;` sites.
func (g *CGenerator) emitTupleNew(inst *mir.Instruction) error {
	structName := g.tupleStructName(inst.Type)
	g.ensureTupleStruct(inst.Type)
	varName := g.getVariableName(inst.ID)
	var fields []string
	for i, op := range inst.Operands {
		fields = append(fields, fmt.Sprintf(".v%d = %s", i, g.getOperandValue(op)))
	}
	if g.declaredVariables[inst.ID] {
		g.output.WriteString(fmt.Sprintf("  %s = (%s){%s};\n", varName, structName, strings.Join(fields, ", ")))
	} else {
		g.output.WriteString(fmt.Sprintf("  %s %s = (%s){%s};\n", structName, varName, structName, strings.Join(fields, ", ")))
		g.declaredVariables[inst.ID] = true
	}
	g.valueTypes[inst.ID] = inst.Type
	return nil
}

// emitTupleExtract lowers `tuple.extract %tuple, N`. For regular tuples
// this is a struct field access. For synthetic split tuples (produced by
// chan.recv.ok — see tupleSplits), we read from the per-component
// variable registered there.
func (g *CGenerator) emitTupleExtract(inst *mir.Instruction) error {
	if len(inst.Operands) != 2 {
		return fmt.Errorf("tuple.extract: expected (tuple, index) operands, got %d", len(inst.Operands))
	}
	tupleOp := inst.Operands[0]
	idxOp := inst.Operands[1]
	if idxOp.Kind != mir.OperandLiteral {
		return fmt.Errorf("tuple.extract: index must be a literal")
	}
	idx := 0
	if _, err := fmt.Sscanf(idxOp.Literal, "%d", &idx); err != nil {
		return fmt.Errorf("tuple.extract: invalid index %q", idxOp.Literal)
	}

	elemC := g.mapType(inst.Type)
	varName := g.getVariableName(inst.ID)

	// Synthetic split: chan.recv.ok stored its (val, ok) pair as two
	// standalone C variables rather than a real struct. Route the
	// extract to the pre-registered names.
	if tupleOp.Kind == mir.OperandValue {
		if names, ok := g.tupleSplits[tupleOp.Value]; ok && idx < len(names) {
			g.emitExtractAssignment(varName, elemC, names[idx], inst.ID)
			return nil
		}
	}

	// Regular struct-backed tuple: produce `tupleVar.vN`.
	tupleExpr := g.getOperandValue(tupleOp)
	access := fmt.Sprintf("%s.v%d", tupleExpr, idx)
	g.emitExtractAssignment(varName, elemC, access, inst.ID)
	return nil
}

func (g *CGenerator) emitExtractAssignment(varName, varType, rhs string, id mir.ValueID) {
	if g.declaredVariables[id] {
		g.output.WriteString(fmt.Sprintf("  %s = %s;\n", varName, rhs))
	} else {
		g.output.WriteString(fmt.Sprintf("  %s %s = %s;\n", varType, varName, rhs))
		g.declaredVariables[id] = true
	}
}

// elementTypeOf extracts the element type string from an OmniLang array type
// like `[]<int>` or `array<int>`. Returns "int" if the input doesn't look
// like an array — that's a reasonable fallback that produces well-formed
// (if possibly wrong) C, leaving the actual diagnosis to the type checker.
func (g *CGenerator) elementTypeOf(arrayType string) string {
	if strings.HasPrefix(arrayType, "array<") && strings.HasSuffix(arrayType, ">") {
		return arrayType[6 : len(arrayType)-1]
	}
	if strings.HasPrefix(arrayType, "[]<") && strings.HasSuffix(arrayType, ">") {
		return arrayType[3 : len(arrayType)-1]
	}
	return "int"
}

// emitIfaceCall emits the C code for an iface.call MIR instruction. The op's
// operand layout is [ifaceNameLit, methodNameLit, receiver, args...]. We look
// up the concrete method pointer at runtime via omni_method_lookup using the
// receiver's type_name tag (set during struct.init), then cast to a typed
// function pointer using the interface method's declared signature.
func (g *CGenerator) emitIfaceCall(inst *mir.Instruction) error {
	if len(inst.Operands) < 3 {
		return fmt.Errorf("iface.call: expected at least 3 operands (iface, method, recv), got %d", len(inst.Operands))
	}
	ifaceOp := inst.Operands[0]
	methodOp := inst.Operands[1]
	if ifaceOp.Kind != mir.OperandLiteral || methodOp.Kind != mir.OperandLiteral {
		return fmt.Errorf("iface.call: iface and method must be literal operands")
	}
	ifaceName := ifaceOp.Literal
	methodName := methodOp.Literal

	// Find the matching interface method to reconstruct the C function pointer
	// signature (needed to cast the void* returned by omni_method_lookup).
	var methodSig *mir.InterfaceMethod
	for _, iface := range g.module.Interfaces {
		if iface.Name != ifaceName {
			continue
		}
		for i, m := range iface.Methods {
			if m.Name == methodName {
				methodSig = &iface.Methods[i]
				break
			}
		}
		if methodSig != nil {
			break
		}
	}
	if methodSig == nil {
		return fmt.Errorf("iface.call: interface %q has no method %q in MIR metadata", ifaceName, methodName)
	}

	// Build the C function-pointer cast: (<retC>(*)(omni_struct_t*, <argCs...>))
	retC := g.mapType(methodSig.ReturnType)
	var castParams []string
	castParams = append(castParams, "omni_struct_t*")
	for _, pt := range methodSig.ParamTypes {
		castParams = append(castParams, g.mapType(pt))
	}
	cast := fmt.Sprintf("%s(*)(%s)", retC, strings.Join(castParams, ", "))

	recvExpr := g.getOperandValue(inst.Operands[2])
	var argExprs []string
	for _, op := range inst.Operands[3:] {
		argExprs = append(argExprs, g.getOperandValue(op))
	}

	callArgs := []string{recvExpr}
	callArgs = append(callArgs, argExprs...)
	call := fmt.Sprintf("((%s)omni_method_lookup(omni_struct_get_type_name(%s), \"%s\"))(%s)",
		cast, recvExpr, methodName, strings.Join(callArgs, ", "))

	if inst.ID == mir.InvalidValue || retC == "void" {
		g.output.WriteString(fmt.Sprintf("  %s;\n", call))
		return nil
	}

	varName := g.getVariableName(inst.ID)
	g.output.WriteString(fmt.Sprintf("  %s = %s;\n", varName, call))
	return nil
}

// writeInterfaceDispatchSupport emits a program-wide method lookup table and
// the omni_method_lookup helper used to route iface.call instructions at
// runtime. The table lists every function whose MIR name matches the
// `<TypeName>.<methodName>` mangling (i.e. every method declared on a
// user-defined type, satisfying or not), keyed by concrete type name +
// method name. The helper does a linear scan; interface method calls are
// expected to be rare enough that a table large enough to matter doesn't
// arise in practice.
func (g *CGenerator) writeInterfaceDispatchSupport() {
	// Discover every (type, method, fn) triple by scanning MIR function names
	// for the `<TypeName>.<method>` mangling introduced by the Phase 1 method
	// support. This deliberately includes methods whether or not they
	// satisfy a declared interface — the table is a uniform runtime view of
	// the method set.
	type entry struct{ typeName, methodName, cName string }
	var entries []entry
	for _, fn := range g.module.Functions {
		// Only include `TypeName.method` mangled names — exactly one dot.
		// Module-qualified stdlib names like `std.io.println` (two dots)
		// are not user methods and must not appear in this table.
		if strings.Count(fn.Name, ".") != 1 {
			continue
		}
		dot := strings.Index(fn.Name, ".")
		typeName := fn.Name[:dot]
		methodName := fn.Name[dot+1:]
		if typeName == "" || methodName == "" || !isLikelyTypeName(typeName) {
			continue
		}
		// Also skip known stdlib aliases (e.g., `io.println`) that the MIR
		// sometimes collapses to a single-dot form — they're intrinsics, not
		// methods. We detect these by checking whether the resulting C name
		// would clash with a runtime-provided function.
		if g.isRuntimeProvidedFunction(fn.Name) {
			continue
		}
		entries = append(entries, entry{typeName, methodName, g.mapFunctionName(fn.Name)})
	}

	g.output.WriteString("// Interface method dispatch table (populated from user-defined methods)\n")
	g.output.WriteString("typedef struct {\n")
	g.output.WriteString("    const char* type_name;\n")
	g.output.WriteString("    const char* method_name;\n")
	g.output.WriteString("    void* fn;\n")
	g.output.WriteString("} omni_method_entry_t;\n\n")

	if len(entries) == 0 {
		// Emit a one-element dummy table so the C standard doesn't complain
		// about zero-length arrays and omni_method_lookup has something to
		// iterate over.
		g.output.WriteString("static const omni_method_entry_t omni_method_table[] = {\n")
		g.output.WriteString("    {NULL, NULL, NULL}\n")
		g.output.WriteString("};\n")
		g.output.WriteString("static const int omni_method_table_len = 0;\n\n")
	} else {
		g.output.WriteString("static const omni_method_entry_t omni_method_table[] = {\n")
		for _, e := range entries {
			g.output.WriteString(fmt.Sprintf("    {\"%s\", \"%s\", (void*)&%s},\n", e.typeName, e.methodName, e.cName))
		}
		g.output.WriteString("};\n")
		g.output.WriteString(fmt.Sprintf("static const int omni_method_table_len = %d;\n\n", len(entries)))
	}

	g.output.WriteString("static void* omni_method_lookup(const char* type_name, const char* method_name) {\n")
	g.output.WriteString("    if (!type_name || !method_name) return NULL;\n")
	g.output.WriteString("    for (int i = 0; i < omni_method_table_len; i++) {\n")
	g.output.WriteString("        if (strcmp(omni_method_table[i].type_name, type_name) == 0 &&\n")
	g.output.WriteString("            strcmp(omni_method_table[i].method_name, method_name) == 0) {\n")
	g.output.WriteString("            return omni_method_table[i].fn;\n")
	g.output.WriteString("        }\n")
	g.output.WriteString("    }\n")
	g.output.WriteString("    return NULL;\n")
	g.output.WriteString("}\n\n")
}

// isLikelyTypeName returns true if s looks like a user-defined type — i.e.
// it starts with a letter and contains only identifier characters. We use
// this as a cheap filter to avoid sweeping module-qualified function names
// into the method dispatch table.
func isLikelyTypeName(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || r == '_') {
				return false
			}
			continue
		}
		if !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}
	return true
}

func (g *CGenerator) functionNeedsReturnValue(funcName, originalReturnType string) bool {
	if funcName == "omni_main" {
		return true
	}
	return strings.TrimSpace(originalReturnType) != "" && strings.TrimSpace(originalReturnType) != "void"
}

func (g *CGenerator) returnSlotDeclaration(funcName, originalReturnType string) string {
	if funcName == "omni_main" {
		return "int32_t __omni_return_value = 0"
	}
	if strings.HasPrefix(originalReturnType, "Promise<") {
		return "omni_promise_t* __omni_return_value = NULL"
	}
	if strings.Contains(originalReturnType, ") -> ") {
		return g.mapFunctionTypeWithName(originalReturnType, "__omni_return_value") + " = NULL"
	}

	cType := g.mapType(originalReturnType)
	switch {
	case cType == "double":
		return cType + " __omni_return_value = 0.0"
	case cType == "int32_t" || cType == "int64_t":
		return cType + " __omni_return_value = 0"
	case strings.HasSuffix(cType, "*"):
		return cType + " __omni_return_value = NULL"
	default:
		return cType + " __omni_return_value = {0}"
	}
}

func (g *CGenerator) emitReturnThroughEpilogue(value string, funcName string, originalReturnType string) {
	if !g.functionNeedsReturnValue(funcName, originalReturnType) {
		g.output.WriteString("  goto __omni_epilogue;\n")
		return
	}

	if funcName == "omni_main" && strings.HasPrefix(originalReturnType, "Promise<") {
		innerType := originalReturnType[8 : len(originalReturnType)-1]
		switch innerType {
		case "int":
			g.output.WriteString(fmt.Sprintf("  __omni_return_value = %s;\n", value))
		case "string":
			g.output.WriteString("  // String return from main not supported, return 0\n")
			g.output.WriteString("  __omni_return_value = 0;\n")
		case "float", "double":
			g.output.WriteString(fmt.Sprintf("  __omni_return_value = (int32_t)%s;\n", value))
		case "bool":
			g.output.WriteString(fmt.Sprintf("  __omni_return_value = %s ? 0 : 1;\n", value))
		default:
			g.output.WriteString(fmt.Sprintf("  __omni_return_value = %s;\n", value))
		}
		g.output.WriteString("  goto __omni_epilogue;\n")
		return
	}

	if strings.HasPrefix(originalReturnType, "Promise<") {
		innerType := originalReturnType[8 : len(originalReturnType)-1]
		promiseFunc := "omni_promise_create_int"
		switch innerType {
		case "int":
			promiseFunc = "omni_promise_create_int"
		case "string":
			promiseFunc = "omni_promise_create_string"
		case "float", "double":
			promiseFunc = "omni_promise_create_float"
		case "bool":
			promiseFunc = "omni_promise_create_bool"
		default:
			g.errors = append(g.errors, fmt.Sprintf("cannot create promise for user-defined type: %s", innerType))
		}
		g.output.WriteString(fmt.Sprintf("  __omni_return_value = %s(%s);\n", promiseFunc, value))
		g.output.WriteString("  goto __omni_epilogue;\n")
		return
	}

	g.output.WriteString(fmt.Sprintf("  __omni_return_value = %s;\n", value))
	g.output.WriteString("  goto __omni_epilogue;\n")
}

func (g *CGenerator) emitReturnDefaultThroughEpilogue(funcName string, originalReturnType string) {
	if g.functionNeedsReturnValue(funcName, originalReturnType) {
		g.output.WriteString("  __omni_return_value = 0;\n")
	}
	g.output.WriteString("  goto __omni_epilogue;\n")
}

func (g *CGenerator) emitFunctionEpilogue(funcName string, originalReturnType string) {
	g.output.WriteString("  __omni_epilogue:\n")
	g.output.WriteString("  ;\n")

	if len(g.stringsToFree) > 0 {
		g.output.WriteString("  // Cleanup: free heap-allocated strings\n")
		var stringIDs []mir.ValueID
		for id := range g.stringsToFree {
			if g.returnedValueIDs[id] {
				continue
			}
			stringIDs = append(stringIDs, id)
		}
		for i := len(stringIDs) - 1; i >= 0; i-- {
			id := stringIDs[i]
			varName := g.getVariableName(id)
			g.output.WriteString(fmt.Sprintf("  if (%s != NULL) { free((void*)%s); %s = NULL; }\n", varName, varName, varName))
		}
	}

	if len(g.tempStringsToFree) > 0 {
		g.output.WriteString("  // Cleanup: free temporary string conversion variables\n")
		for i := len(g.tempStringsToFree) - 1; i >= 0; i-- {
			tempVar := g.tempStringsToFree[i]
			g.output.WriteString(fmt.Sprintf("  if (%s != NULL) { free((void*)%s); %s = NULL; }\n", tempVar, tempVar, tempVar))
		}
	}

	if len(g.promisesToFree) > 0 {
		g.output.WriteString("  // Cleanup: free promises\n")
		for id := range g.promisesToFree {
			if g.returnedValueIDs[id] {
				continue
			}
			varName := g.getVariableName(id)
			g.output.WriteString(fmt.Sprintf("  if (%s != NULL) { omni_promise_free(%s); %s = NULL; }\n", varName, varName, varName))
		}
	}

	if g.functionNeedsReturnValue(funcName, originalReturnType) {
		g.output.WriteString("  return __omni_return_value;\n")
	} else {
		g.output.WriteString("  return;\n")
	}
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
		// For async main, we need to await the promise and return int32_t
		// The runtime expects omni_main to return int32_t, not a promise
		if strings.HasPrefix(fn.ReturnType, "Promise<") {
			returnType = "int32_t"
		} else {
			returnType = "int32_t" // Always use int32_t for omni_main to match runtime
		}
	} else if strings.HasPrefix(fn.ReturnType, "Promise<") {
		// For async functions (Promise<T>), the function should return omni_promise_t*
		// The function body will create a promise and return it
		returnType = "omni_promise_t*"
	}

	// Handle function pointer return types
	if strings.Contains(fn.ReturnType, ") -> ") {
		// This is a function pointer return type - need special handling
		g.output.WriteString(g.generateCompleteFunctionSignature(fn.ReturnType, funcName, fn.Params))
		g.output.WriteString(" {\n")
	} else {
		g.output.WriteString(fmt.Sprintf("%s %s(", returnType, funcName))

		// Generate parameters. Each `array<T>` parameter gets a
		// synthetic `int32_t __omni_len_<name>` companion so the
		// callee can resolve len() at runtime — see
		// isArrayParamType / arrayLenParamName.
		first := true
		for _, param := range fn.Params {
			if !first {
				g.output.WriteString(", ")
			}
			first = false
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
				if isArrayParamType(param.Type) {
					g.output.WriteString(fmt.Sprintf(", int32_t %s", arrayLenParamName(param.Name)))
				}
			}
		}
		g.output.WriteString(") {\n")
	}

	// Track the current function so the defer emitters can name context
	// structs and thunks consistently. Emit the frame variable up front if
	// any defer.push appears in this function.
	g.currentDeferFn = fn
	if functionHasDefer(fn) {
		g.output.WriteString("  omni_defer_frame_t __omni_defer_frame = {0};\n")
	}

	// Reset maps for this function to avoid conflicts
	g.variables = make(map[mir.ValueID]string)
	g.phiVars = make(map[mir.ValueID]bool)
	g.mutableVars = make(map[mir.ValueID]bool)
	g.stringsToFree = make(map[mir.ValueID]bool)
	g.promisesToFree = make(map[mir.ValueID]bool)
	g.tempStringsToFree = []string{}
	g.returnedValueID = mir.InvalidValue
	g.returnedValueIDs = make(map[mir.ValueID]bool)
	g.declaredVariables = make(map[mir.ValueID]bool)
	// mapVars tracks names that were bound to map literals in the *current*
	// function; don't let it leak across functions (e.g. if one function
	// binds `values` to a map, another function's `values` array parameter
	// would be mis-detected as a map and indexed with omni_map_get_*).
	g.mapVars = make(map[string]bool)
	// mapTypes and valueTypes are keyed by SSA ValueID, which restarts per
	// function — leftover entries from a previous function could otherwise
	// drive wrong codegen (map dispatch on an array parameter, etc).
	g.mapTypes = make(map[mir.ValueID]string)
	g.valueTypes = make(map[mir.ValueID]string)
	g.arrayLengths = make(map[mir.ValueID]int)
	g.arrayLengthExprs = make(map[mir.ValueID]string)

	if g.functionNeedsReturnValue(funcName, fn.ReturnType) {
		g.output.WriteString(fmt.Sprintf("  %s;\n", g.returnSlotDeclaration(funcName, fn.ReturnType)))
	}

	// Map parameter SSA values to their names and types so downstream codegen
	// (e.g. string concat) can pick the right conversion helper instead of
	// defaulting to omni_int_to_string.
	for _, param := range fn.Params {
		g.variables[param.ID] = param.Name
		if param.Type != "" {
			g.valueTypes[param.ID] = param.Type
		}
		if isArrayParamType(param.Type) {
			g.arrayLengthExprs[param.ID] = arrayLenParamName(param.Name)
		}
	}

	// Pre-register every `assign` op's result SSA ID as pointing to the
	// target variable, before any block emits code. Without this,
	// terminators in blocks emitted before the block that contains the
	// assign (e.g. a while_exit lying between the header and the
	// mutating then-branch) look up the result ID, find no mapping, and
	// synthesize a fresh `v<N>` name that never gets initialized. We
	// chase the chain transitively so `assign` of an `assign` keeps
	// pointing to the original binding.
	assignedFrom := make(map[mir.ValueID]mir.ValueID)
	for _, block := range fn.Blocks {
		for _, inst := range block.Instructions {
			if inst.Op == "assign" && len(inst.Operands) >= 1 && inst.ID != mir.InvalidValue {
				if inst.Operands[0].Kind == mir.OperandValue {
					assignedFrom[inst.ID] = inst.Operands[0].Value
				}
			}
		}
	}
	resolveAssignRoot := func(id mir.ValueID) mir.ValueID {
		for {
			target, ok := assignedFrom[id]
			if !ok {
				return id
			}
			id = target
		}
	}
	for assignID := range assignedFrom {
		rootID := resolveAssignRoot(assignID)
		if name, ok := g.variables[rootID]; ok {
			g.variables[assignID] = name
		} else {
			g.variables[assignID] = fmt.Sprintf("v%d", int(rootID))
		}
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
				// Special case for read_line()/read_all() - always return string
				if inst.Op == "call" || inst.Op == "call.string" {
					if len(inst.Operands) > 0 && inst.Operands[0].Kind == mir.OperandLiteral {
						funcName := inst.Operands[0].Literal
						switch funcName {
						case "std.io.read_line", "io.read_line",
							"std.io.read_all", "io.read_all":
							varType = "const char*"
						case "std.string.char_at", "string.char_at":
							varType = "char"
						}
					}
				}
				// Special case for await - determine type from inst.Type or operand's Promise type
				if inst.Op == "await" {
					// First check inst.Type (should be the unwrapped type)
					if inst.Type == "string" {
						varType = "const char*"
					} else if inst.Type == "int" {
						varType = "int32_t"
					} else if inst.Type == "bool" {
						varType = "int32_t"
					} else if inst.Type == "float" || inst.Type == "double" {
						varType = "double"
					} else if strings.HasPrefix(inst.Type, "Promise<") {
						// If type is still Promise, extract inner type
						innerType := inst.Type[8 : len(inst.Type)-1]
						if innerType == "string" {
							varType = "const char*"
						} else if innerType == "int" {
							varType = "int32_t"
						} else if innerType == "bool" {
							varType = "int32_t"
						} else if innerType == "float" || innerType == "double" {
							varType = "double"
						} else {
							varType = "int32_t" // Default fallback
						}
					} else if len(inst.Operands) > 0 {
						// Try to infer from the operand's Promise type
						if inst.Operands[0].Kind == mir.OperandValue {
							operandID := inst.Operands[0].Value
							if operandInst, found := instructionMap[operandID]; found {
								if strings.HasPrefix(operandInst.Type, "Promise<") {
									innerType := operandInst.Type[8 : len(operandInst.Type)-1]
									if innerType == "string" {
										varType = "const char*"
									} else if innerType == "int" {
										varType = "int32_t"
									} else if innerType == "bool" {
										varType = "int32_t"
									} else if innerType == "float" || innerType == "double" {
										varType = "double"
									} else {
										varType = "int32_t"
									}
								} else if operandInst.Type != "" && operandInst.Type != inferTypePlaceholder {
									// Operand type might be the unwrapped type already
									varType = g.mapType(operandInst.Type)
								}
							}
						}
						// Also check operand's Type field
						if varType == "" && inst.Operands[0].Type != "" && inst.Operands[0].Type != inferTypePlaceholder {
							if strings.HasPrefix(inst.Operands[0].Type, "Promise<") {
								innerType := inst.Operands[0].Type[8 : len(inst.Operands[0].Type)-1]
								if innerType == "string" {
									varType = "const char*"
								} else if innerType == "int" {
									varType = "int32_t"
								} else if innerType == "bool" {
									varType = "int32_t"
								} else if innerType == "float" || innerType == "double" {
									varType = "double"
								} else {
									varType = "int32_t"
								}
							}
						}
					}
					// If still not determined, default based on common async I/O patterns
					if varType == "" {
						// Default to string for await (most common case for async I/O)
						varType = "const char*"
					}
				}
				// Special case for async function calls - they return Promise
				if inst.Op == "call" && strings.HasPrefix(inst.Type, "Promise<") {
					varType = "omni_promise_t*"
				}
				// File handles are FILE* pointers carried as intptr_t in C.
				// The Omni surface type remains int, but int32_t truncates
				// handles on 64-bit platforms.
				if inst.Op == "call" && len(inst.Operands) > 0 && inst.Operands[0].Kind == mir.OperandLiteral {
					if g.mapFunctionName(inst.Operands[0].Literal) == "omni_file_open" {
						varType = "intptr_t"
					}
				}
				// Some runtime intrinsics return int but lose their type during MIR
				// lowering (reported as void). When we see one with a consumed
				// result, pre-declare it as int32_t.
				if (inst.Op == "call" || inst.Op == "call.int" || inst.Op == "call.void") && (inst.Type == "" || inst.Type == "void") && len(inst.Operands) > 0 && inst.Operands[0].Kind == mir.OperandLiteral {
					fname := inst.Operands[0].Literal
					cname := g.mapFunctionName(fname)
					switch cname {
					case "omni_args_count", "omni_getpid", "omni_getppid",
						"omni_time_now_unix", "omni_time_now_unix_nano",
						"omni_ip_is_valid", "omni_ip_is_loopback",
						"omni_ip_is_private", "omni_ip_is_multicast",
						"omni_network_ping", "omni_network_is_connected",
						"omni_args_has_flag", "omni_socket_create",
						"omni_socket_bind", "omni_socket_listen", "omni_socket_accept",
						"omni_socket_send", "omni_socket_receive", "omni_socket_close",
						"omni_setenv", "omni_unsetenv", "omni_file_exists", "omni_delete_file",
						"omni_write_file", "omni_append_file",
						"omni_create_dir", "omni_remove_dir",
						"omni_chdir", "omni_mkdir", "omni_rmdir", "omni_remove",
						"omni_rename", "omni_copy", "omni_exists",
						"omni_is_file", "omni_is_dir",
						"omni_url_is_valid":
						varType = "int32_t"
					case "omni_ip_parse", "omni_network_get_local_ip":
						varType = "omni_ip_address_t*"
					case "omni_url_parse":
						varType = "omni_url_t*"
					case "omni_args_get_flag", "omni_args_get", "omni_ip_to_string",
						"omni_url_to_string", "omni_dns_reverse_lookup",
						"omni_getcwd", "omni_getenv", "omni_gethostname",
						"omni_read_file":
						varType = "const char*"
					}
				}
				// Network struct returns: when the MIR type heuristic
				// produced "URL"/"IPAddress"/"HTTPResponse" for a runtime
				// call, pin the C type to the concrete runtime struct
				// pointer so member access can use direct field reads.
				if inst.Op == "call" && len(inst.Operands) > 0 && inst.Operands[0].Kind == mir.OperandLiteral {
					switch g.mapFunctionName(inst.Operands[0].Literal) {
					case "omni_url_parse":
						if inst.Type == "URL" || strings.HasSuffix(inst.Type, ".URL") {
							varType = "omni_url_t*"
						}
					case "omni_ip_parse", "omni_network_get_local_ip":
						if inst.Type == "IPAddress" || strings.HasSuffix(inst.Type, ".IPAddress") {
							varType = "omni_ip_address_t*"
						}
					case "omni_http_get", "omni_http_post", "omni_http_put",
						"omni_http_delete", "omni_http_request":
						if inst.Type == "HTTPResponse" || strings.HasSuffix(inst.Type, ".HTTPResponse") {
							varType = "omni_http_response_t*"
						}
					}
				}
				// User-defined function calls that return a struct type often get
				// reported with inst.Type empty (the type checker doesn't always
				// qualify across module boundaries). If the call resolves to a
				// user function whose return type is a non-primitive, pre-declare
				// the result as omni_struct_t*.
				if (inst.Op == "call" || inst.Op == "call.int" || inst.Op == "call.void") && varType == "" && len(inst.Operands) > 0 && inst.Operands[0].Kind == mir.OperandLiteral {
					fname := inst.Operands[0].Literal
					// Skip runtime-provided intrinsics; the call-emission path
					// handles their specific return types.
					if g.isRuntimeProvidedFunction(fname) {
						goto skipUserFnCheck
					}
					for _, userFn := range g.module.Functions {
						if userFn.Name != fname {
							continue
						}
						rt := userFn.ReturnType
						switch {
						case rt == "" || rt == "void":
							// leave alone
						case rt == "string":
							varType = "const char*"
						case rt == "int" || rt == "int32" || rt == "int32_t":
							varType = "int32_t"
						case rt == "int64" || rt == "int64_t":
							varType = "int64_t"
						case rt == "bool":
							varType = "int32_t"
						case rt == "float" || rt == "double":
							varType = "double"
						case !g.isPrimitiveType(rt) && !strings.Contains(rt, "<") && !strings.Contains(rt, "("):
							varType = "omni_struct_t*"
						default:
							varType = g.mapType(rt)
						}
						break
					}
				skipUserFnCheck:
				}
				// Special case for member access - infer type from field name for known structs
				if inst.Op == "member" && len(inst.Operands) >= 2 {
					fieldName := inst.Operands[1].Literal
					// Special handling for known HTTPResponse fields
					if fieldName == "body" || fieldName == "status_text" {
						varType = "const char*"
					} else if fieldName == "status_code" {
						varType = "int32_t"
					} else if inst.Type != "" && inst.Type != inferTypePlaceholder {
						// Use inst.Type if available
						varType = g.mapType(inst.Type)
					}
					// URL string fields and IPAddress fields don't always
					// have inst.Type set after the MIR builder lowers them
					// (the type checker may strip the field type). Look at
					// the producing instruction's MIR type for the source
					// struct to fill in the gap.
					if (varType == "" || varType == "int32_t") && len(inst.Operands) >= 1 && inst.Operands[0].Kind == mir.OperandValue {
						srcType := ""
						if srcInst, ok := instructionMap[inst.Operands[0].Value]; ok {
							srcType = srcInst.Type
						}
						if srcType == "" {
							srcType = g.valueTypes[inst.Operands[0].Value]
						}
						if srcType == "URL" || strings.HasSuffix(srcType, ".URL") {
							switch fieldName {
							case "scheme", "host", "path", "query", "fragment":
								varType = "const char*"
							case "port":
								varType = "int32_t"
							}
						} else if srcType == "IPAddress" || strings.HasSuffix(srcType, ".IPAddress") {
							switch fieldName {
							case "address":
								varType = "const char*"
							case "is_ipv4", "is_ipv6":
								varType = "int32_t"
							}
						}
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
						// Special case: char_at returns char, not a struct
						if inst.Op == "call" || inst.Op == "call.string" {
							if len(inst.Operands) > 0 && inst.Operands[0].Kind == mir.OperandLiteral {
								funcName := inst.Operands[0].Literal
								if funcName == "std.string.char_at" || funcName == "string.char_at" {
									varType = "char"
								}
							}
						}
						if varType == "" {
							isStruct := !g.isPrimitiveType(resultType) && !strings.Contains(resultType, "<") && !strings.Contains(resultType, "(")
							if isStruct {
								varType = "omni_struct_t*"
							} else {
								varType = g.mapType(inst.Type)
							}
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
				// Special case for function pointer types
				if strings.Contains(inst.Type, ") -> ") {
					// Use mapFunctionTypeWithName to generate correct function pointer type
					// This already includes the variable name, so we'll handle it specially in the declaration
					varType = g.mapFunctionTypeWithName(inst.Type, varName)
				}
				// Default: use the instruction type
				if varType == "" {
					// Handle <infer> type - try to infer from context or use default
					if inst.Type == "<infer>" || inst.Type == inferTypePlaceholder || inst.Type == "" {
						// For member access, try to infer from field name if it's a known struct field
						if inst.Op == "member" && len(inst.Operands) >= 2 {
							fieldName := inst.Operands[1].Literal
							// Special handling for known HTTPResponse fields
							if fieldName == "body" || fieldName == "status_text" {
								varType = "const char*"
							} else if fieldName == "status_code" {
								varType = "int32_t"
							} else {
								// Try to get type from valueTypes map
								if storedType, ok := g.valueTypes[inst.ID]; ok && storedType != "" && storedType != "<infer>" && storedType != inferTypePlaceholder {
									varType = g.mapType(storedType)
								} else {
									// Default to int if we can't infer
									varType = "int32_t"
								}
							}
						} else {
							// Try to get type from valueTypes map
							if storedType, ok := g.valueTypes[inst.ID]; ok && storedType != "" && storedType != "<infer>" && storedType != inferTypePlaceholder {
								varType = g.mapType(storedType)
							} else {
								// Default to int if we can't infer
								varType = "int32_t"
							}
						}
					} else {
						varType = g.mapType(inst.Type)
					}
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
					// Check if this is a function pointer type (already includes variable name)
					if strings.Contains(varType, "(*") && strings.Contains(varType, ")(") && strings.Contains(varType, varName) {
						// Function pointer type already includes variable name, just output type with semicolon
						g.output.WriteString(fmt.Sprintf("  %s;\n", varType))
					} else if strings.HasSuffix(strings.TrimSpace(varType), "*") {
						g.output.WriteString(fmt.Sprintf("  %s %s = NULL;\n", varType, varName))
					} else {
						g.output.WriteString(fmt.Sprintf("  %s %s;\n", varType, varName))
					}
				}
				// Mark this variable as declared
				g.declaredVariables[id] = true
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

	g.emitFunctionEpilogue(funcName, fn.ReturnType)

	g.output.WriteString("}\n\n")
	return nil
}

// generateBlock generates C code for a basic block
func (g *CGenerator) generateBlock(block *mir.BasicBlock, fn *mir.Function) error {
	funcName := fn.Name
	if funcName == "main" {
		funcName = "omni_main"
	}
	// Generate block label. The entry block gets `entry:` too so self
	// tail-calls can `goto entry;` after reassigning parameters; that's
	// half of the TCO story (the other half is cross-function tail calls
	// staying as `return f(args);` for clang's sibling-call optimization).
	g.output.WriteString(fmt.Sprintf("  %s:\n", block.Name))
	g.output.WriteString("  ;\n")

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

	// Detect a tail call: the block ends with `ret %callId` and the last
	// instruction is `call` producing %callId with a literal callee. We
	// then emit either `goto entry;` (after reassigning params) for
	// self-recursion, or `return f(args);` for cross-function tail calls
	// — the latter shape is what clang's -foptimize-sibling-calls
	// recognizes.
	tailIdx := -1
	if isTail, idx := tailCallIndex(block); isTail {
		tailIdx = idx
	}

	// Generate instructions, skipping the tail-call instruction itself
	// (we'll emit its specialized form below).
	for i, inst := range block.Instructions {
		if i == tailIdx {
			continue
		}
		if err := g.generateInstruction(&inst); err != nil {
			return err
		}
	}

	if tailIdx >= 0 {
		g.emitTailCall(block.Instructions[tailIdx], fn)
		return nil
	}

	// Generate terminator
	if err := g.generateTerminator(&block.Terminator, funcName, fn.ReturnType); err != nil {
		return err
	}

	return nil
}

// tailCallIndex returns (true, idx) if block ends with `ret %callId` where
// the previous instruction at idx produces %callId via a static-callee
// `call`. Conservative: skips iface.call, func.call, intrinsics with
// special lowering, and Promise-returning calls.
func tailCallIndex(block *mir.BasicBlock) (bool, int) {
	if block.Terminator.Op != "ret" || len(block.Terminator.Operands) != 1 {
		return false, -1
	}
	retOp := block.Terminator.Operands[0]
	if retOp.Kind != mir.OperandValue {
		return false, -1
	}
	for i := len(block.Instructions) - 1; i >= 0; i-- {
		inst := block.Instructions[i]
		if inst.ID != retOp.Value {
			continue
		}
		switch inst.Op {
		case "call", "call.int", "call.string", "call.bool":
			// proceed
		default:
			return false, -1
		}
		if len(inst.Operands) == 0 || inst.Operands[0].Kind != mir.OperandLiteral {
			return false, -1
		}
		callee := inst.Operands[0].Literal
		// Skip stdlib / module-qualified callees: their lowering is
		// special-cased throughout the C backend, and the natural shape
		// `omni_println_string(...)` doesn't need TCO anyway.
		if strings.Contains(callee, ".") {
			return false, -1
		}
		// Builtins like `len` get bespoke call-site emission (length
		// hint, element-size argument, etc). Never tail-call them — the
		// generic `return f(args);` shape would lose those arguments.
		if callee == "len" || callee == "append" {
			return false, -1
		}
		// Skip Promise returns — the regular call path wraps these.
		if strings.HasPrefix(inst.Type, "Promise<") {
			return false, -1
		}
		// Must be the LAST non-terminator instruction. If anything
		// non-trivial sits between the call and the ret, we'd lose those
		// side effects.
		if i != len(block.Instructions)-1 {
			return false, -1
		}
		return true, i
	}
	return false, -1
}

// emitTailCall writes the C tail-call form for a tail-position `call`
// instruction. Self-recursion becomes parameter reassignment + goto entry.
// Cross-function calls still flow through the shared epilogue so normal
// cleanup runs before returning to the caller.
func (g *CGenerator) emitTailCall(inst mir.Instruction, fn *mir.Function) {
	callee := inst.Operands[0].Literal
	argOps := inst.Operands[1:]

	if callee == fn.Name && len(argOps) == len(fn.Params) {
		// Self-recursive: stage new args into temporaries first so we
		// don't clobber a parameter that another arg expression reads.
		// Array params also need the synthetic length companion staged
		// so the rebind into entry: keeps len() working.
		g.output.WriteString("  {\n")
		for i, op := range argOps {
			pType := g.mapType(fn.Params[i].Type)
			g.output.WriteString(fmt.Sprintf("    %s __omni_tco_a%d = %s;\n", pType, i, g.getOperandValue(op)))
			if isArrayParamType(fn.Params[i].Type) {
				g.output.WriteString(fmt.Sprintf("    int32_t __omni_tco_a%d_len = %s;\n", i, g.getOperandLengthExpr(op)))
			}
		}
		for i, p := range fn.Params {
			paramName := g.formatParamRef(p)
			g.output.WriteString(fmt.Sprintf("    %s = __omni_tco_a%d;\n", paramName, i))
			if isArrayParamType(p.Type) {
				g.output.WriteString(fmt.Sprintf("    %s = __omni_tco_a%d_len;\n", arrayLenParamName(p.Name), i))
			}
		}
		g.output.WriteString("    goto entry;\n")
		g.output.WriteString("  }\n")
		return
	}

	// Cross-function: stage the result and let the function epilogue handle
	// cleanup before returning to the caller.
	calleeC := g.mapFunctionName(callee)
	arrayParamSet := map[int]bool{}
	if idxs, ok := g.userFuncArrayParams[callee]; ok {
		for _, i := range idxs {
			arrayParamSet[i] = true
		}
	}
	var argExprs []string
	for i, op := range argOps {
		argExprs = append(argExprs, g.getOperandValue(op))
		if arrayParamSet[i] {
			argExprs = append(argExprs, g.getOperandLengthExpr(op))
		}
	}
	currentFuncName := g.mapFunctionName(fn.Name)
	if fn.Name == "main" {
		currentFuncName = "omni_main"
	}
	g.emitReturnThroughEpilogue(fmt.Sprintf("%s(%s)", calleeC, strings.Join(argExprs, ", ")), currentFuncName, fn.ReturnType)
}

// formatParamRef returns the C identifier the function body uses to read
// param p. Param values flow through SSA value IDs the same way as any
// other binding, so we just look up the variable name we'd have given
// inst.ID == p.ID.
func (g *CGenerator) formatParamRef(p mir.Param) string {
	return g.getVariableName(p.ID)
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
	case "iface.call":
		if err := g.emitIfaceCall(inst); err != nil {
			return err
		}
		return nil
	case "defer.push":
		if err := g.emitDeferPush(inst); err != nil {
			return err
		}
		return nil
	case "defer.push.func":
		if err := g.emitDeferPushFunc(inst); err != nil {
			return err
		}
		return nil
	case "defer.push.iface":
		if err := g.emitDeferPushIface(inst); err != nil {
			return err
		}
		return nil
	case "defer.run":
		g.emitDeferRun()
		return nil
	case "slice.append":
		if err := g.emitSliceAppend(inst); err != nil {
			return err
		}
		return nil
	case "slice.slice":
		if err := g.emitSliceSlice(inst); err != nil {
			return err
		}
		return nil
	case "spawn":
		if err := g.emitSpawn(inst); err != nil {
			return err
		}
		return nil
	case "chan.make":
		if err := g.emitChanMake(inst); err != nil {
			return err
		}
		return nil
	case "chan.send":
		if err := g.emitChanSend(inst); err != nil {
			return err
		}
		return nil
	case "chan.recv":
		if err := g.emitChanRecv(inst); err != nil {
			return err
		}
		return nil
	case "chan.recv.ok":
		if err := g.emitChanRecvOk(inst); err != nil {
			return err
		}
		return nil
	case "chan.close":
		if err := g.emitChanClose(inst); err != nil {
			return err
		}
		return nil
	case "tuple.new":
		if err := g.emitTupleNew(inst); err != nil {
			return err
		}
		return nil
	case "tuple.extract":
		if err := g.emitTupleExtract(inst); err != nil {
			return err
		}
		return nil
	case "select":
		if err := g.emitSelect(inst); err != nil {
			return err
		}
		return nil
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
			case "char":
				// Char literals come through as 'a' / '\n' / etc. — C
				// already understands that syntax (yields an int with the
				// code-point value), so emit it verbatim.
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
			if !g.declaredVariables[inst.ID] {
				g.output.WriteString(fmt.Sprintf("  int32_t %s = !%s;\n", varName, operand))
				g.declaredVariables[inst.ID] = true
			} else {
				g.output.WriteString(fmt.Sprintf("  %s = !%s;\n", varName, operand))
			}
			g.valueTypes[inst.ID] = "bool"
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
			// Handle <infer> type
			targetType := "int32_t" // Default
			if inst.Type != "<infer>" && inst.Type != inferTypePlaceholder {
				targetType = g.mapType(inst.Type)
			} else if storedType, ok := g.valueTypes[inst.ID]; ok && storedType != "" && storedType != "<infer>" && storedType != inferTypePlaceholder {
				targetType = g.mapType(storedType)
			}
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
				// Get array length: prefer compile-time tracking, fall
				// back to the runtime length expression we register for
				// function-parameter arrays.
				arrayLength := -1
				lengthExpr := ""
				if inst.Operands[1].Kind == mir.OperandValue {
					arrayOperandID := inst.Operands[1].Value
					if length, ok := g.arrayLengths[arrayOperandID]; ok {
						arrayLength = length
					} else if expr, ok := g.arrayLengthExprs[arrayOperandID]; ok {
						lengthExpr = expr
					} else {
						g.errors = append(g.errors, fmt.Sprintf("WARNING: array length not known for variable %s (ID: %d) - len() requires compile-time known array length or explicit length parameter", arrayVar, arrayOperandID))
						arrayLength = -1
					}
				}
				// Get array type to determine element size
				arrayType := "int" // Default element type
				if inst.Operands[1].Kind == mir.OperandValue {
					arrayOperandID := inst.Operands[1].Value
					if arrType, ok := g.valueTypes[arrayOperandID]; ok && arrType != "" && arrType != "<infer>" {
						// Extract element type from array type (e.g., "[]<int>" -> "int")
						if strings.HasPrefix(arrType, "[]<") && strings.HasSuffix(arrType, ">") {
							arrayType = arrType[3 : len(arrType)-1]
						} else if strings.HasPrefix(arrType, "array<") && strings.HasSuffix(arrType, ">") {
							arrayType = arrType[6 : len(arrType)-1]
						}
						// Handle cases where element type is still <infer>
						if arrayType == "<infer>" || arrayType == "" {
							arrayType = "int"
						}
					}
				}
				// Map element type to C type to get element size
				elementCType := g.mapType(arrayType)
				elementSize := "sizeof(" + elementCType + ")"
				// Prefer the runtime length expression (function-param
				// arrays) when we have one; fall back to the compile-time
				// length otherwise.
				if lengthExpr != "" {
					g.output.WriteString(fmt.Sprintf("  %s = omni_len((void*)%s, %s, %s);\n",
						varName, arrayVar, elementSize, lengthExpr))
				} else {
					if arrayLength < 0 {
						g.output.WriteString(fmt.Sprintf("  // WARNING: Array length unknown for %s, len() may return incorrect value\n", arrayVar))
					}
					g.output.WriteString(fmt.Sprintf("  %s = omni_len((void*)%s, %s, %d);\n",
						varName, arrayVar, elementSize, arrayLength))
				}
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
			if (funcName == "std.io.eprint" || funcName == "io.eprint") && len(inst.Operands) >= 2 {
				g.emitPrintTo(inst.Operands[1], false, true)
				return nil
			}
			if funcName == "std.io.eprintln" || funcName == "io.eprintln" {
				if len(inst.Operands) >= 2 {
					g.emitPrintTo(inst.Operands[1], true, true)
				} else {
					g.output.WriteString("  omni_eprintln_string(\"\");\n")
				}
				return nil
			}
			if funcName == "std.io.flush" || funcName == "io.flush" {
				g.output.WriteString("  omni_io_flush();\n")
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
			if funcName == "std.io.read_all" || funcName == "io.read_all" {
				if inst.ID != mir.InvalidValue {
					varName := g.getVariableName(inst.ID)
					g.output.WriteString(fmt.Sprintf("  %s = omni_read_all();\n", varName))
					g.valueTypes[inst.ID] = "string"
					g.stringsToFree[inst.ID] = true
				} else {
					g.output.WriteString("  omni_read_all();\n")
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

			// Check for web framework functions early (before void check)
			isWebContextFunc := cFuncName == "omni_context_text" || cFuncName == "omni_context_json" || cFuncName == "omni_context_file" ||
				cFuncName == "omni_context_html" || cFuncName == "omni_context_redirect" ||
				funcName == "std.web.context_text" || funcName == "std.web.context_json" || funcName == "std.web.context_file" ||
				funcName == "std.web.context_html" || funcName == "std.web.context_redirect"
			isWebContextParamFunc := cFuncName == "omni_context_param" || funcName == "std.web.context_param" ||
				cFuncName == "omni_context_query" || funcName == "std.web.context_query" ||
				cFuncName == "omni_context_header" || funcName == "std.web.context_header" ||
				cFuncName == "omni_context_get_cookie" || funcName == "std.web.context_get_cookie"
			isWebContextSingleArgStringFunc := cFuncName == "omni_context_body" || funcName == "std.web.context_body"
			isWebContextFilesFunc := cFuncName == "omni_context_body_form" || funcName == "std.web.context_body_form" ||
				cFuncName == "omni_context_files" || funcName == "std.web.context_files"
			isWebContextBodyFunc := cFuncName == "omni_context_body_json" || funcName == "std.web.context_body_json"
			isServerCreateFunc := cFuncName == "omni_server_create" || funcName == "std.web.server_create"
			isServerListenFunc := cFuncName == "omni_server_listen" || funcName == "std.web.server_listen"
			isServerGroupFunc := cFuncName == "omni_server_group" || funcName == "std.web.server_group"
			isArrayGetFunc := cFuncName == "omni_array_get_int" || funcName == "std.array.get"
			isArraySetFunc := cFuncName == "omni_array_set_int" || funcName == "std.array.set"
			isAssertEqCall := funcName == "std.assert.eq" || funcName == "assert.eq" || cFuncName == "omni_assert_eq"
			// std.array.append/prepend/insert/remove/reverse/slice/concat/fill/copy
			// route through emitStdArrayIntOp for int / string element types; the
			// rest fall back to the legacy passthrough that returns the input
			// array unchanged. std.algorithms.shuffle / unique share the same
			// element-typed runtime layout (omni_array_<int|str>_<op>) so they
			// piggy-back on the same dispatcher.
			isArrayPassthroughFunc := funcName == "std.array.append" || funcName == "std.array.prepend" ||
				funcName == "std.array.insert" || funcName == "std.array.remove" ||
				funcName == "std.array.reverse" || funcName == "std.array.slice" ||
				funcName == "std.array.concat" || funcName == "std.array.fill" ||
				funcName == "std.array.copy" ||
				funcName == "std.algorithms.shuffle" || funcName == "std.algorithms.unique"
			isArrayContainsFunc := funcName == "std.array.contains"
			isArrayIndexOfFunc := funcName == "std.array.index_of"
			// Variable-length-result string ops: the runtime writes the
			// element count to a stack-allocated companion. Routed
			// through a dedicated emitter to keep the standard call path
			// clean of out-pointer plumbing.
			isStringVarLenArrayFunc := funcName == "std.string.split" ||
				funcName == "std.string.split_lines" ||
				funcName == "std.string.split_words" ||
				funcName == "std.string.find_all"
			isStringJoinFunc := funcName == "std.string.join"
			isWebBoolIntrinsic := cFuncName == "omni_validate_string" || cFuncName == "omni_validate_int" ||
				cFuncName == "omni_validate_email" || cFuncName == "omni_validate_url" ||
				funcName == "std.web.validate_string" || funcName == "std.web.validate_int" ||
				funcName == "std.web.validate_email" || funcName == "std.web.validate_url"
			isWebStringIntrinsic := cFuncName == "omni_sanitize_html" || cFuncName == "omni_sanitize_sql" ||
				funcName == "std.web.sanitize_html" || funcName == "std.web.sanitize_sql"

			// Special case for HTTP functions - they always return HTTPResponse struct
			isHTTPFunc := funcName == "std.network.http_get" || funcName == "std.network.http_post" || funcName == "std.network.http_put" || funcName == "std.network.http_delete" || funcName == "std.network.http_request" ||
				cFuncName == "omni_http_get" || cFuncName == "omni_http_post" || cFuncName == "omni_http_put" || cFuncName == "omni_http_delete" || cFuncName == "omni_http_request"

			// Handle web context functions BEFORE void check (they return values)
			// Note: inst.Operands[0] is the function name (literal), actual args start at [1]
			if isWebContextFunc && inst.ID != mir.InvalidValue {
				varName := g.getVariableName(inst.ID)
				if len(inst.Operands) >= 4 {
					// Function has 3 arguments: ctx, arg1, arg2 (e.g. context_redirect)
					ctx := g.getOperandValue(inst.Operands[1])
					a1 := g.getOperandValue(inst.Operands[2])
					a2 := g.getOperandValue(inst.Operands[3])
					if !g.declaredVariables[inst.ID] {
						g.output.WriteString(fmt.Sprintf("  omni_struct_t* %s = %s(%s, %s, %s);\n", varName, cFuncName, ctx, a1, a2))
						g.declaredVariables[inst.ID] = true
					} else {
						g.output.WriteString(fmt.Sprintf("  %s = %s(%s, %s, %s);\n", varName, cFuncName, ctx, a1, a2))
					}
					g.valueTypes[inst.ID] = inst.Type
					return nil
				} else if len(inst.Operands) >= 3 {
					// Function has 2 arguments: ctx and arg
					ctx := g.getOperandValue(inst.Operands[1])
					arg := g.getOperandValue(inst.Operands[2])
					if !g.declaredVariables[inst.ID] {
						g.output.WriteString(fmt.Sprintf("  omni_struct_t* %s = %s(%s, %s);\n", varName, cFuncName, ctx, arg))
						g.declaredVariables[inst.ID] = true
					} else {
						g.output.WriteString(fmt.Sprintf("  %s = %s(%s, %s);\n", varName, cFuncName, ctx, arg))
					}
					g.valueTypes[inst.ID] = inst.Type
					return nil
				} else if len(inst.Operands) >= 2 {
					// Function has 1 argument: ctx
					ctx := g.getOperandValue(inst.Operands[1])
					if !g.declaredVariables[inst.ID] {
						g.output.WriteString(fmt.Sprintf("  omni_struct_t* %s = %s(%s);\n", varName, cFuncName, ctx))
						g.declaredVariables[inst.ID] = true
					} else {
						g.output.WriteString(fmt.Sprintf("  %s = %s(%s);\n", varName, cFuncName, ctx))
					}
					g.valueTypes[inst.ID] = inst.Type
					return nil
				}
			} else if isWebContextParamFunc && inst.ID != mir.InvalidValue && len(inst.Operands) >= 3 {
				varName := g.getVariableName(inst.ID)
				ctx := g.getOperandValue(inst.Operands[1])
				name := g.getOperandValue(inst.Operands[2])
				if !g.declaredVariables[inst.ID] {
					g.output.WriteString(fmt.Sprintf("  const char* %s = %s(%s, %s);\n", varName, cFuncName, ctx, name))
					g.declaredVariables[inst.ID] = true
				} else {
					g.output.WriteString(fmt.Sprintf("  %s = %s(%s, %s);\n", varName, cFuncName, ctx, name))
				}
				// omni_context_param returns const char*, ensure type is recorded correctly
				g.valueTypes[inst.ID] = "const char*"
				return nil
			} else if isWebContextBodyFunc && inst.ID != mir.InvalidValue && len(inst.Operands) >= 2 {
				varName := g.getVariableName(inst.ID)
				ctx := g.getOperandValue(inst.Operands[1])
				if !g.declaredVariables[inst.ID] {
					g.output.WriteString(fmt.Sprintf("  void* %s = %s(%s);\n", varName, cFuncName, ctx))
					g.declaredVariables[inst.ID] = true
				} else {
					g.output.WriteString(fmt.Sprintf("  %s = %s(%s);\n", varName, cFuncName, ctx))
				}
				g.valueTypes[inst.ID] = inst.Type
				return nil
			} else if isWebContextSingleArgStringFunc && inst.ID != mir.InvalidValue && len(inst.Operands) >= 2 {
				varName := g.getVariableName(inst.ID)
				ctx := g.getOperandValue(inst.Operands[1])
				if !g.declaredVariables[inst.ID] {
					g.output.WriteString(fmt.Sprintf("  const char* %s = %s(%s);\n", varName, cFuncName, ctx))
					g.declaredVariables[inst.ID] = true
				} else {
					g.output.WriteString(fmt.Sprintf("  %s = %s(%s);\n", varName, cFuncName, ctx))
				}
				g.valueTypes[inst.ID] = "const char*"
				return nil
			} else if isWebContextFilesFunc && inst.ID != mir.InvalidValue && len(inst.Operands) >= 2 {
				varName := g.getVariableName(inst.ID)
				ctx := g.getOperandValue(inst.Operands[1])
				returnType := "omni_array_t*"
				if cFuncName == "omni_context_body_form" {
					returnType = "omni_map_t*"
				}
				if !g.declaredVariables[inst.ID] {
					g.output.WriteString(fmt.Sprintf("  %s %s = %s(%s);\n", returnType, varName, cFuncName, ctx))
					g.declaredVariables[inst.ID] = true
				} else {
					g.output.WriteString(fmt.Sprintf("  %s = %s(%s);\n", varName, cFuncName, ctx))
				}
				g.valueTypes[inst.ID] = returnType
				return nil
			} else if isServerCreateFunc && inst.ID != mir.InvalidValue && len(inst.Operands) >= 3 {
				varName := g.getVariableName(inst.ID)
				port := g.getOperandValue(inst.Operands[1])
				options := g.getOperandValue(inst.Operands[2])
				if g.declaredVariables[inst.ID] {
					g.output.WriteString(fmt.Sprintf("  %s = omni_server_create(%s, %s);\n", varName, port, options))
				} else {
					g.output.WriteString(fmt.Sprintf("  omni_server_t* %s = omni_server_create(%s, %s);\n", varName, port, options))
					g.declaredVariables[inst.ID] = true
				}
				g.valueTypes[inst.ID] = inst.Type
				return nil
			} else if isServerListenFunc && inst.ID != mir.InvalidValue && len(inst.Operands) >= 2 {
				varName := g.getVariableName(inst.ID)
				server := g.getOperandValue(inst.Operands[1])
				if !g.declaredVariables[inst.ID] {
					g.output.WriteString(fmt.Sprintf("  int32_t %s = %s(%s);\n", varName, cFuncName, server))
					g.declaredVariables[inst.ID] = true
				} else {
					g.output.WriteString(fmt.Sprintf("  %s = %s(%s);\n", varName, cFuncName, server))
				}
				g.valueTypes[inst.ID] = "bool"
				return nil
			} else if (isWebBoolIntrinsic || isWebStringIntrinsic) && inst.ID != mir.InvalidValue && len(inst.Operands) >= 2 {
				varName := g.getVariableName(inst.ID)
				var args []string
				for _, arg := range inst.Operands[1:] {
					args = append(args, g.getOperandValue(arg))
				}
				cType := "int32_t"
				recordType := "bool"
				if isWebStringIntrinsic {
					cType = "const char*"
					recordType = "const char*"
				}
				if !g.declaredVariables[inst.ID] {
					g.output.WriteString(fmt.Sprintf("  %s %s = %s(%s);\n", cType, varName, cFuncName, strings.Join(args, ", ")))
					g.declaredVariables[inst.ID] = true
				} else {
					g.output.WriteString(fmt.Sprintf("  %s = %s(%s);\n", varName, cFuncName, strings.Join(args, ", ")))
				}
				g.valueTypes[inst.ID] = recordType
				return nil
			} else if isArrayPassthroughFunc && inst.ID != mir.InvalidValue && len(inst.Operands) >= 2 {
				if g.emitStdArrayIntOp(inst, funcName) {
					return nil
				}
				// Fall through to legacy passthrough for non-int element types.
				varName := g.getVariableName(inst.ID)
				arr := g.getOperandValue(inst.Operands[1])
				elemCType := "int32_t"
				if inst.Operands[1].Kind == mir.OperandValue {
					if t, ok := g.valueTypes[inst.Operands[1].Value]; ok && t != "" {
						if strings.HasPrefix(t, "array<") && strings.HasSuffix(t, ">") {
							elemCType = g.mapType(t[6 : len(t)-1])
						} else if strings.HasPrefix(t, "[]<") && strings.HasSuffix(t, ">") {
							elemCType = g.mapType(t[3 : len(t)-1])
						}
					}
				}
				if g.declaredVariables[inst.ID] {
					g.output.WriteString(fmt.Sprintf("  %s = %s;\n", varName, arr))
				} else {
					g.output.WriteString(fmt.Sprintf("  %s* %s = %s;\n", elemCType, varName, arr))
					g.declaredVariables[inst.ID] = true
				}
				if inst.Operands[1].Kind == mir.OperandValue {
					if t, ok := g.valueTypes[inst.Operands[1].Value]; ok {
						g.valueTypes[inst.ID] = t
					}
				}
				return nil
			} else if isArrayContainsFunc && inst.ID != mir.InvalidValue {
				if g.emitStdArrayIntOp(inst, funcName) {
					return nil
				}
				varName := g.getVariableName(inst.ID)
				if g.declaredVariables[inst.ID] {
					g.output.WriteString(fmt.Sprintf("  %s = 0;\n", varName))
				} else {
					g.output.WriteString(fmt.Sprintf("  int32_t %s = 0;\n", varName))
					g.declaredVariables[inst.ID] = true
				}
				return nil
			} else if isArrayIndexOfFunc && inst.ID != mir.InvalidValue {
				if g.emitStdArrayIntOp(inst, funcName) {
					return nil
				}
				varName := g.getVariableName(inst.ID)
				if g.declaredVariables[inst.ID] {
					g.output.WriteString(fmt.Sprintf("  %s = -1;\n", varName))
				} else {
					g.output.WriteString(fmt.Sprintf("  int32_t %s = -1;\n", varName))
					g.declaredVariables[inst.ID] = true
				}
				return nil
			} else if isStringVarLenArrayFunc && inst.ID != mir.InvalidValue && len(inst.Operands) >= 2 {
				// Variable-length string-array result. Allocate the
				// companion length variable, pass its address to the
				// runtime, register it as the result's runtime length.
				varName := g.getVariableName(inst.ID)
				lenVar := fmt.Sprintf("__omni_strarr_len_%d", inst.ID)
				g.output.WriteString(fmt.Sprintf("  int32_t %s = 0;\n", lenVar))
				resultC := "const char**"
				resultType := "array<string>"
				if funcName == "std.string.find_all" {
					resultC = "int32_t*"
					resultType = "array<int>"
				}
				cFn := ""
				switch funcName {
				case "std.string.split":
					cFn = "omni_string_split"
				case "std.string.split_lines":
					cFn = "omni_string_split_lines"
				case "std.string.split_words":
					cFn = "omni_string_split_words"
				case "std.string.find_all":
					cFn = "omni_string_find_all"
				}
				// Emit args: first the input string, optionally the
				// delimiter / substring, then &lenVar.
				var args []string
				for _, op := range inst.Operands[1:] {
					args = append(args, g.getOperandValue(op))
				}
				args = append(args, "&"+lenVar)
				expr := fmt.Sprintf("%s(%s)", cFn, strings.Join(args, ", "))
				if g.declaredVariables[inst.ID] {
					g.output.WriteString(fmt.Sprintf("  %s = %s;\n", varName, expr))
				} else {
					g.output.WriteString(fmt.Sprintf("  %s %s = %s;\n", resultC, varName, expr))
					g.declaredVariables[inst.ID] = true
				}
				g.arrayLengthExprs[inst.ID] = lenVar
				g.valueTypes[inst.ID] = resultType
				return nil
			} else if isStringJoinFunc && inst.ID != mir.InvalidValue && len(inst.Operands) >= 3 {
				// std.string.join(parts, sep): pass parts' runtime length
				// alongside it (matches the array-param ABI).
				varName := g.getVariableName(inst.ID)
				parts := g.getOperandValue(inst.Operands[1])
				partsLen := g.getOperandLengthExpr(inst.Operands[1])
				sep := g.getOperandValue(inst.Operands[2])
				expr := fmt.Sprintf("omni_string_join(%s, %s, %s)", parts, partsLen, sep)
				if g.declaredVariables[inst.ID] {
					g.output.WriteString(fmt.Sprintf("  %s = %s;\n", varName, expr))
				} else {
					g.output.WriteString(fmt.Sprintf("  const char* %s = %s;\n", varName, expr))
					g.declaredVariables[inst.ID] = true
				}
				g.valueTypes[inst.ID] = "string"
				g.stringsToFree[inst.ID] = true
				return nil
			} else if isAssertEqCall && len(inst.Operands) >= 4 {
				// Dispatch to type-specific omni_assert_eq_{int,string,float}.
				expectedOp := inst.Operands[1]
				actualOp := inst.Operands[2]
				msgOp := inst.Operands[3]
				typeOf := func(op mir.Operand) string {
					if op.Type != "" && op.Type != inferTypePlaceholder {
						return op.Type
					}
					if op.Kind == mir.OperandValue {
						if t, ok := g.valueTypes[op.Value]; ok {
							return t
						}
					}
					return ""
				}
				eType := typeOf(expectedOp)
				aType := typeOf(actualOp)
				argType := eType
				if argType == "" || argType == inferTypePlaceholder {
					argType = aType
				}
				helper := "omni_assert_eq_int"
				switch argType {
				case "string", "const char*", "char*":
					helper = "omni_assert_eq_string"
				case "float", "double":
					helper = "omni_assert_eq_float"
				}
				expected := g.getOperandValue(expectedOp)
				actual := g.getOperandValue(actualOp)
				msg := g.getOperandValue(msgOp)
				g.output.WriteString(fmt.Sprintf("  %s(%s, %s, %s);\n", helper, expected, actual, msg))
				return nil
			} else if isArrayGetFunc && inst.ID != mir.InvalidValue && len(inst.Operands) >= 3 {
				varName := g.getVariableName(inst.ID)
				arr := g.getOperandValue(inst.Operands[1])
				idx := g.getOperandValue(inst.Operands[2])
				lengthExpr := g.getOperandLengthExpr(inst.Operands[1])
				if !g.declaredVariables[inst.ID] {
					g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_array_get_int(%s, %s, %s);\n", varName, arr, idx, lengthExpr))
					g.declaredVariables[inst.ID] = true
				} else {
					g.output.WriteString(fmt.Sprintf("  %s = omni_array_get_int(%s, %s, %s);\n", varName, arr, idx, lengthExpr))
				}
				g.valueTypes[inst.ID] = "int"
				return nil
			} else if isArraySetFunc && len(inst.Operands) >= 4 {
				arr := g.getOperandValue(inst.Operands[1])
				idx := g.getOperandValue(inst.Operands[2])
				val := g.getOperandValue(inst.Operands[3])
				lengthExpr := g.getOperandLengthExpr(inst.Operands[1])
				g.output.WriteString(fmt.Sprintf("  omni_array_set_int(%s, %s, %s, %s);\n", arr, idx, val, lengthExpr))
				return nil
			} else if isServerGroupFunc && inst.ID != mir.InvalidValue && len(inst.Operands) >= 3 {
				varName := g.getVariableName(inst.ID)
				server := g.getOperandValue(inst.Operands[1])
				prefix := g.getOperandValue(inst.Operands[2])
				if !g.declaredVariables[inst.ID] {
					g.output.WriteString(fmt.Sprintf("  omni_struct_t* %s = omni_server_group(%s, %s);\n", varName, server, prefix))
					g.declaredVariables[inst.ID] = true
				} else {
					g.output.WriteString(fmt.Sprintf("  %s = omni_server_group(%s, %s);\n", varName, server, prefix))
				}
				g.valueTypes[inst.ID] = inst.Type
				return nil
			}

			if isHTTPFunc && inst.ID != mir.InvalidValue {
				varName := g.getVariableName(inst.ID)
				if (funcName == "std.network.http_get" || cFuncName == "omni_http_get") && len(inst.Operands) >= 2 {
					url := g.getOperandValue(inst.Operands[1])
					g.output.WriteString(fmt.Sprintf("  omni_http_response_t* %s = omni_http_get(%s);\n", varName, url))
					if inst.Type != "" {
						g.valueTypes[inst.ID] = inst.Type
					} else {
						g.valueTypes[inst.ID] = "HTTPResponse"
					}
				} else if (funcName == "std.network.http_post" || cFuncName == "omni_http_post") && len(inst.Operands) >= 3 {
					url := g.getOperandValue(inst.Operands[1])
					body := g.getOperandValue(inst.Operands[2])
					g.output.WriteString(fmt.Sprintf("  omni_http_response_t* %s = omni_http_post(%s, %s);\n", varName, url, body))
					if inst.Type != "" {
						g.valueTypes[inst.ID] = inst.Type
					} else {
						g.valueTypes[inst.ID] = "HTTPResponse"
					}
				} else if (funcName == "std.network.http_put" || cFuncName == "omni_http_put") && len(inst.Operands) >= 3 {
					url := g.getOperandValue(inst.Operands[1])
					body := g.getOperandValue(inst.Operands[2])
					g.output.WriteString(fmt.Sprintf("  omni_http_response_t* %s = omni_http_put(%s, %s);\n", varName, url, body))
					if inst.Type != "" {
						g.valueTypes[inst.ID] = inst.Type
					} else {
						g.valueTypes[inst.ID] = "HTTPResponse"
					}
				} else if (funcName == "std.network.http_delete" || cFuncName == "omni_http_delete") && len(inst.Operands) >= 2 {
					url := g.getOperandValue(inst.Operands[1])
					g.output.WriteString(fmt.Sprintf("  omni_http_response_t* %s = omni_http_delete(%s);\n", varName, url))
					if inst.Type != "" {
						g.valueTypes[inst.ID] = inst.Type
					} else {
						g.valueTypes[inst.ID] = "HTTPResponse"
					}
				} else if (funcName == "std.network.http_request" || cFuncName == "omni_http_request") && len(inst.Operands) >= 2 {
					req := g.getOperandValue(inst.Operands[1])
					g.output.WriteString(fmt.Sprintf("  omni_http_response_t* %s = omni_http_request(%s);\n", varName, req))
					if inst.Type != "" {
						g.valueTypes[inst.ID] = inst.Type
					} else {
						g.valueTypes[inst.ID] = "HTTPResponse"
					}
				}
				return nil
			}

			// Check for server_create even if inst.ID is InvalidValue (might be used later)
			// This can happen if the MIR builder doesn't assign an ID but the result is used
			if isServerCreateFunc && len(inst.Operands) >= 3 {
				// Try to find the variable that should hold the server
				// For now, if inst.ID is InvalidValue, we can't assign it, but we should still call it
				// This is a MIR builder issue, but we can work around it by checking if there's a later use
				if inst.ID != mir.InvalidValue {
					varName := g.getVariableName(inst.ID)
					port := g.getOperandValue(inst.Operands[1])
					options := g.getOperandValue(inst.Operands[2])
					g.output.WriteString(fmt.Sprintf("  omni_server_t* %s = omni_server_create(%s, %s);\n", varName, port, options))
					g.valueTypes[inst.ID] = inst.Type
					return nil
				} else {
					// inst.ID is InvalidValue - this shouldn't happen for let statements
					// But if it does, we need to generate a temporary variable
					// For now, just call without assignment (this will cause errors if used later)
					port := g.getOperandValue(inst.Operands[1])
					options := g.getOperandValue(inst.Operands[2])
					g.output.WriteString(fmt.Sprintf("  omni_server_create(%s, %s);\n", port, options))
					// This is a bug - the result should be assigned
					g.errors = append(g.errors, fmt.Sprintf("WARNING: omni_server_create called without assignment (inst.ID is InvalidValue)"))
					return nil
				}
			}

			// Handle void function calls differently.
			// If the MIR says void but a result is used (inst.ID valid AND variable
			// is pre-declared from the function-scope sweep), assign anyway and
			// assume int32_t — this covers runtime intrinsics whose return type
			// was lost during type checking (e.g. omni_args_count, omni_getpid).
			// Handle void function calls differently.
			// Special case: a small set of runtime intrinsics actually return int
			// but lose their return type in the MIR. When these appear with a
			// valid inst.ID that's already been pre-declared (i.e. a later
			// instruction will reference the result), emit the assignment.
			valueReturningVoidIntrinsics := map[string]bool{
				"omni_args_count":           true,
				"omni_getpid":               true,
				"omni_getppid":              true,
				"omni_time_now_unix":        true,
				"omni_time_now_unix_nano":   true,
				"omni_ip_is_valid":          true,
				"omni_ip_is_loopback":       true,
				"omni_ip_is_private":        true,
				"omni_ip_is_multicast":      true,
				"omni_network_ping":         true,
				"omni_network_is_connected": true,
				"omni_ip_parse":             true,
				"omni_network_get_local_ip": true,
				"omni_url_parse":            true,
				"omni_args_has_flag":        true,
				"omni_args_get_flag":        true,
				"omni_args_get":             true,
				"omni_ip_to_string":         true,
				"omni_url_to_string":        true,
				"omni_dns_reverse_lookup":   true,
				"omni_getcwd":               true,
				"omni_getenv":               true,
				"omni_gethostname":          true,
				"omni_setenv":               true,
				"omni_unsetenv":             true,
				"omni_file_exists":          true,
				"omni_delete_file":          true,
				"omni_write_file":           true,
				"omni_append_file":          true,
				"omni_read_file":            true,
				"omni_create_dir":           true,
				"omni_remove_dir":           true,
				"omni_chdir":                true,
				"omni_mkdir":                true,
				"omni_rmdir":                true,
				"omni_remove":               true,
				"omni_rename":               true,
				"omni_copy":                 true,
				"omni_exists":               true,
				"omni_is_file":              true,
				"omni_is_dir":               true,
				"omni_url_is_valid":         true,
				"omni_socket_create":        true,
				"omni_socket_bind":          true,
				"omni_socket_listen":        true,
				"omni_socket_accept":        true,
				"omni_socket_send":          true,
				"omni_socket_receive":       true,
				"omni_socket_close":         true,
			}
			if inst.Type == "void" {
				if inst.ID != mir.InvalidValue && g.declaredVariables[inst.ID] && valueReturningVoidIntrinsics[cFuncName] {
					varName := g.getVariableName(inst.ID)
					g.output.WriteString(fmt.Sprintf("  %s = %s(", varName, cFuncName))
				} else {
					g.output.WriteString(fmt.Sprintf("  %s(", cFuncName))
				}
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
					// Async functions already return omni_promise_t*, so just assign (variable already declared)
					g.output.WriteString(fmt.Sprintf("  %s = %s(", varName, cFuncName))
					// Add arguments
					for i, arg := range inst.Operands[1:] {
						if i > 0 {
							g.output.WriteString(", ")
						}
						g.output.WriteString(g.getOperandValue(arg))
					}
					g.output.WriteString(");\n")
					// Store the Promise type in valueTypes so await can find it
					g.valueTypes[inst.ID] = inst.Type
					// Track promises for cleanup (especially string promises)
					innerType := inst.Type[8 : len(inst.Type)-1] // Extract inner type
					if innerType == "string" {
						g.promisesToFree[inst.ID] = true
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
					// Check for server_listen which returns bool
					if (funcName == "std.web.server_listen" || cFuncName == "omni_server_listen") && len(inst.Operands) >= 2 {
						varName := g.getVariableName(inst.ID)
						server := g.getOperandValue(inst.Operands[1])
						if !g.declaredVariables[inst.ID] {
							g.output.WriteString(fmt.Sprintf("  int32_t %s = %s(%s);\n", varName, cFuncName, server))
							g.declaredVariables[inst.ID] = true
						} else {
							g.output.WriteString(fmt.Sprintf("  %s = %s(%s);\n", varName, cFuncName, server))
						}
						g.valueTypes[inst.ID] = "bool"
						return nil
					}
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
						g.valueTypes[inst.ID] = inst.Type
					} else if funcName == "omni_url_parse" && inst.Type != "" && strings.Contains(inst.Type, "URL") {
						// URL functions return omni_url_t*
						varName := g.getVariableName(inst.ID)
						if len(inst.Operands) >= 2 {
							urlStr := g.getOperandValue(inst.Operands[1])
							g.output.WriteString(fmt.Sprintf("  omni_url_t* %s = omni_url_parse(%s);\n", varName, urlStr))
							g.valueTypes[inst.ID] = inst.Type
						}
					} else if funcName == "std.network.http_get" || funcName == "std.network.http_post" || funcName == "std.network.http_put" || funcName == "std.network.http_delete" || funcName == "std.network.http_request" ||
						cFuncName == "omni_http_get" || cFuncName == "omni_http_post" || cFuncName == "omni_http_put" || cFuncName == "omni_http_delete" || cFuncName == "omni_http_request" {
						// HTTP functions return omni_http_response_t*
						varName := g.getVariableName(inst.ID)
						if (funcName == "std.network.http_get" || cFuncName == "omni_http_get") && len(inst.Operands) >= 2 {
							url := g.getOperandValue(inst.Operands[1])
							g.output.WriteString(fmt.Sprintf("  omni_http_response_t* %s = omni_http_get(%s);\n", varName, url))
							g.valueTypes[inst.ID] = inst.Type
						} else if (funcName == "std.network.http_post" || cFuncName == "omni_http_post") && len(inst.Operands) >= 3 {
							url := g.getOperandValue(inst.Operands[1])
							body := g.getOperandValue(inst.Operands[2])
							g.output.WriteString(fmt.Sprintf("  omni_http_response_t* %s = omni_http_post(%s, %s);\n", varName, url, body))
							g.valueTypes[inst.ID] = inst.Type
						} else if (funcName == "std.network.http_put" || cFuncName == "omni_http_put") && len(inst.Operands) >= 3 {
							url := g.getOperandValue(inst.Operands[1])
							body := g.getOperandValue(inst.Operands[2])
							g.output.WriteString(fmt.Sprintf("  omni_http_response_t* %s = omni_http_put(%s, %s);\n", varName, url, body))
							g.valueTypes[inst.ID] = inst.Type
						} else if (funcName == "std.network.http_delete" || cFuncName == "omni_http_delete") && len(inst.Operands) >= 2 {
							url := g.getOperandValue(inst.Operands[1])
							g.output.WriteString(fmt.Sprintf("  omni_http_response_t* %s = omni_http_delete(%s);\n", varName, url))
							g.valueTypes[inst.ID] = inst.Type
						} else if (funcName == "std.network.http_request" || cFuncName == "omni_http_request") && len(inst.Operands) >= 2 {
							req := g.getOperandValue(inst.Operands[1])
							g.output.WriteString(fmt.Sprintf("  omni_http_response_t* %s = omni_http_request(%s);\n", varName, req))
							g.valueTypes[inst.ID] = inst.Type
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
						if (funcName == "std.time.now" || funcName == "time.now") && inst.Type != "" && strings.Contains(inst.Type, "Time") {
							varName := g.getVariableName(inst.ID)
							yearVar := fmt.Sprintf("_year_%d", inst.ID)
							monthVar := fmt.Sprintf("_month_%d", inst.ID)
							dayVar := fmt.Sprintf("_day_%d", inst.ID)
							hourVar := fmt.Sprintf("_hour_%d", inst.ID)
							minuteVar := fmt.Sprintf("_minute_%d", inst.ID)
							secondVar := fmt.Sprintf("_second_%d", inst.ID)
							nanosecondVar := fmt.Sprintf("_nanosecond_%d", inst.ID)

							g.output.WriteString(fmt.Sprintf("  int32_t %s, %s, %s, %s, %s, %s, %s;\n",
								yearVar, monthVar, dayVar, hourVar, minuteVar, secondVar, nanosecondVar))
							g.output.WriteString(fmt.Sprintf("  omni_time_from_unix(omni_time_now_unix(), &%s, &%s, &%s, &%s, &%s, &%s, &%s);\n",
								yearVar, monthVar, dayVar, hourVar, minuteVar, secondVar, nanosecondVar))

							g.output.WriteString(fmt.Sprintf("  %s = omni_struct_create();\n", varName))
							g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field(%s, \"year\", %s);\n", varName, yearVar))
							g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field(%s, \"month\", %s);\n", varName, monthVar))
							g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field(%s, \"day\", %s);\n", varName, dayVar))
							g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field(%s, \"hour\", %s);\n", varName, hourVar))
							g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field(%s, \"minute\", %s);\n", varName, minuteVar))
							g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field(%s, \"second\", %s);\n", varName, secondVar))
							g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field(%s, \"nanosecond\", %s);\n", varName, nanosecondVar))
						} else if cFuncName == "omni_time_from_unix" && inst.Type != "" && strings.Contains(inst.Type, "Time") {
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
						} else if cFuncName == "omni_time_from_string" && inst.Type != "" && strings.Contains(inst.Type, "Time") {
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
						} else if cFuncName == "omni_time_to_unix" && len(inst.Operands) >= 2 {
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
						} else if cFuncName == "omni_time_to_string" && len(inst.Operands) >= 2 {
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
						} else if cFuncName == "omni_time_to_unix_nano" && len(inst.Operands) >= 2 {
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
						} else if cFuncName == "omni_duration_to_string" && len(inst.Operands) >= 2 {
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
							// Regular function call - check if we need to declare the variable
							// For HTTP functions that return structs, we need special handling
							// Check both OmniLang function names and C function names
							isHTTPFunc := funcName == "std.network.http_get" || funcName == "std.network.http_post" || funcName == "std.network.http_put" || funcName == "std.network.http_delete" || funcName == "std.network.http_request" ||
								cFuncName == "omni_http_get" || cFuncName == "omni_http_post" || cFuncName == "omni_http_put" || cFuncName == "omni_http_delete" || cFuncName == "omni_http_request"

							// Check for web framework functions that return structs
							isWebContextFunc := cFuncName == "omni_context_text" || cFuncName == "omni_context_json" || cFuncName == "omni_context_file" ||
								funcName == "std.web.context_text" || funcName == "std.web.context_json" || funcName == "std.web.context_file"
							isWebContextParamFunc := cFuncName == "omni_context_param" || funcName == "std.web.context_param" ||
								cFuncName == "omni_context_query" || funcName == "std.web.context_query" ||
								cFuncName == "omni_context_header" || funcName == "std.web.context_header" ||
								cFuncName == "omni_context_get_cookie" || funcName == "std.web.context_get_cookie"
							isWebContextSingleArgStringFunc := cFuncName == "omni_context_body" || funcName == "std.web.context_body"
							isWebContextFilesFunc := cFuncName == "omni_context_body_form" || funcName == "std.web.context_body_form" ||
								cFuncName == "omni_context_files" || funcName == "std.web.context_files"
							isWebContextBodyFunc := cFuncName == "omni_context_body_json" || funcName == "std.web.context_body_json"
							isServerCreateFunc := cFuncName == "omni_server_create" || funcName == "std.web.server_create"

							if isHTTPFunc {
								// HTTP functions return omni_http_response_t* - declare inline
								if (funcName == "std.network.http_get" || cFuncName == "omni_http_get") && len(inst.Operands) >= 2 {
									url := g.getOperandValue(inst.Operands[1])
									g.output.WriteString(fmt.Sprintf("  omni_http_response_t* %s = omni_http_get(%s);\n", varName, url))
									g.valueTypes[inst.ID] = inst.Type
								} else if (funcName == "std.network.http_post" || cFuncName == "omni_http_post") && len(inst.Operands) >= 3 {
									url := g.getOperandValue(inst.Operands[1])
									body := g.getOperandValue(inst.Operands[2])
									g.output.WriteString(fmt.Sprintf("  omni_http_response_t* %s = omni_http_post(%s, %s);\n", varName, url, body))
									g.valueTypes[inst.ID] = inst.Type
								} else if (funcName == "std.network.http_put" || cFuncName == "omni_http_put") && len(inst.Operands) >= 3 {
									url := g.getOperandValue(inst.Operands[1])
									body := g.getOperandValue(inst.Operands[2])
									g.output.WriteString(fmt.Sprintf("  omni_http_response_t* %s = omni_http_put(%s, %s);\n", varName, url, body))
									g.valueTypes[inst.ID] = inst.Type
								} else if (funcName == "std.network.http_delete" || cFuncName == "omni_http_delete") && len(inst.Operands) >= 2 {
									url := g.getOperandValue(inst.Operands[1])
									g.output.WriteString(fmt.Sprintf("  omni_http_response_t* %s = omni_http_delete(%s);\n", varName, url))
									g.valueTypes[inst.ID] = inst.Type
								} else if (funcName == "std.network.http_request" || cFuncName == "omni_http_request") && len(inst.Operands) >= 2 {
									req := g.getOperandValue(inst.Operands[1])
									g.output.WriteString(fmt.Sprintf("  omni_http_response_t* %s = omni_http_request(%s);\n", varName, req))
									g.valueTypes[inst.ID] = inst.Type
								} else {
									// Fallback: assign to already declared variable
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
							} else if isServerCreateFunc && len(inst.Operands) >= 3 {
								// omni_server_create returns omni_server_t* - declare inline
								port := g.getOperandValue(inst.Operands[1])
								options := g.getOperandValue(inst.Operands[2])
								g.output.WriteString(fmt.Sprintf("  omni_server_t* %s = omni_server_create(%s, %s);\n", varName, port, options))
								g.valueTypes[inst.ID] = inst.Type
							} else if isWebContextFunc {
								// Context functions that return omni_struct_t* - declare inline if not already declared
								// Always assign the return value, even if inst.ID is InvalidValue
								// (the return value might be used in a return statement)
								if inst.ID != mir.InvalidValue {
									varName := g.getVariableName(inst.ID)
									if len(inst.Operands) >= 2 {
										ctx := g.getOperandValue(inst.Operands[0])
										arg := g.getOperandValue(inst.Operands[1])
										if !g.declaredVariables[inst.ID] {
											g.output.WriteString(fmt.Sprintf("  omni_struct_t* %s = %s(%s, %s);\n", varName, cFuncName, ctx, arg))
											g.declaredVariables[inst.ID] = true
										} else {
											g.output.WriteString(fmt.Sprintf("  %s = %s(%s, %s);\n", varName, cFuncName, ctx, arg))
										}
										g.valueTypes[inst.ID] = inst.Type
									} else if len(inst.Operands) >= 1 {
										// Some context functions might only take ctx
										ctx := g.getOperandValue(inst.Operands[0])
										if !g.declaredVariables[inst.ID] {
											g.output.WriteString(fmt.Sprintf("  omni_struct_t* %s = %s(%s);\n", varName, cFuncName, ctx))
											g.declaredVariables[inst.ID] = true
										} else {
											g.output.WriteString(fmt.Sprintf("  %s = %s(%s);\n", varName, cFuncName, ctx))
										}
										g.valueTypes[inst.ID] = inst.Type
									}
								} else {
									// inst.ID is InvalidValue, but we still need to assign the return value
									// This can happen if the MIR doesn't assign the result but it's used in a return
									// Generate a temporary variable name based on the instruction index or use a default
									// For now, just call without assignment - this is a MIR builder issue
									if len(inst.Operands) >= 2 {
										ctx := g.getOperandValue(inst.Operands[0])
										arg := g.getOperandValue(inst.Operands[1])
										g.output.WriteString(fmt.Sprintf("  %s(%s, %s);\n", cFuncName, ctx, arg))
									} else if len(inst.Operands) >= 1 {
										ctx := g.getOperandValue(inst.Operands[0])
										g.output.WriteString(fmt.Sprintf("  %s(%s);\n", cFuncName, ctx))
									}
								}
							} else if isWebContextParamFunc && len(inst.Operands) >= 2 {
								// omni_context_param returns const char* - declare inline if not already declared
								ctx := g.getOperandValue(inst.Operands[0])
								name := g.getOperandValue(inst.Operands[1])
								if !g.declaredVariables[inst.ID] {
									g.output.WriteString(fmt.Sprintf("  const char* %s = %s(%s, %s);\n", varName, cFuncName, ctx, name))
									g.declaredVariables[inst.ID] = true
								} else {
									g.output.WriteString(fmt.Sprintf("  %s = %s(%s, %s);\n", varName, cFuncName, ctx, name))
								}
								g.valueTypes[inst.ID] = inst.Type
							} else if isWebContextSingleArgStringFunc && len(inst.Operands) >= 1 {
								ctx := g.getOperandValue(inst.Operands[0])
								if !g.declaredVariables[inst.ID] {
									g.output.WriteString(fmt.Sprintf("  const char* %s = %s(%s);\n", varName, cFuncName, ctx))
									g.declaredVariables[inst.ID] = true
								} else {
									g.output.WriteString(fmt.Sprintf("  %s = %s(%s);\n", varName, cFuncName, ctx))
								}
								g.valueTypes[inst.ID] = "const char*"
							} else if isWebContextFilesFunc && len(inst.Operands) >= 1 {
								ctx := g.getOperandValue(inst.Operands[0])
								returnType := "omni_array_t*"
								if cFuncName == "omni_context_body_form" {
									returnType = "omni_map_t*"
								}
								if !g.declaredVariables[inst.ID] {
									g.output.WriteString(fmt.Sprintf("  %s %s = %s(%s);\n", returnType, varName, cFuncName, ctx))
									g.declaredVariables[inst.ID] = true
								} else {
									g.output.WriteString(fmt.Sprintf("  %s = %s(%s);\n", varName, cFuncName, ctx))
								}
								g.valueTypes[inst.ID] = returnType
							} else if isWebContextBodyFunc && len(inst.Operands) >= 1 {
								// omni_context_body_json returns void* - declare inline if not already declared
								ctx := g.getOperandValue(inst.Operands[0])
								if !g.declaredVariables[inst.ID] {
									g.output.WriteString(fmt.Sprintf("  void* %s = %s(%s);\n", varName, cFuncName, ctx))
									g.declaredVariables[inst.ID] = true
								} else {
									g.output.WriteString(fmt.Sprintf("  %s = %s(%s);\n", varName, cFuncName, ctx))
								}
								g.valueTypes[inst.ID] = inst.Type
							} else {
								// Regular function call - assign to already declared variable
								// But first check if variable is declared - if not, declare it
								if !g.declaredVariables[inst.ID] {
									// Determine the C type for the return value
									cType := g.mapType(inst.Type)
									g.output.WriteString(fmt.Sprintf("  %s %s = %s(", cType, varName, cFuncName))
									g.declaredVariables[inst.ID] = true
								} else {
									g.output.WriteString(fmt.Sprintf("  %s = %s(", varName, cFuncName))
								}
								// Add arguments. For each user-defined function whose
								// signature includes array<T> parameters, splice in the
								// synthetic length companion right after each array arg.
								arrayParamSet := map[int]bool{}
								if idxs, ok := g.userFuncArrayParams[funcName]; ok {
									for _, i := range idxs {
										arrayParamSet[i] = true
									}
								}
								wroteAny := false
								for i, arg := range inst.Operands[1:] {
									if wroteAny {
										g.output.WriteString(", ")
									}
									wroteAny = true
									g.output.WriteString(g.getOperandValue(arg))
									if arrayParamSet[i] {
										g.output.WriteString(", ")
										g.output.WriteString(g.getOperandLengthExpr(arg))
									}
								}
								g.output.WriteString(");\n")
								g.valueTypes[inst.ID] = inst.Type
								// Some runtime functions return known concrete C
								// structs (omni_url_t*, omni_ip_address_t*, etc.).
								// When inst.Type is empty/inferred, member access
								// downstream needs to know the OmniLang struct
								// type to pick direct field access vs the generic
								// omni_struct_t* getter; record it explicitly here.
								if existing := g.valueTypes[inst.ID]; existing == "" || existing == "<infer>" || existing == inferTypePlaceholder {
									switch cFuncName {
									case "omni_url_parse":
										g.valueTypes[inst.ID] = "URL"
									case "omni_ip_parse", "omni_network_get_local_ip":
										g.valueTypes[inst.ID] = "IPAddress"
									case "omni_http_get", "omni_http_post", "omni_http_put",
										"omni_http_delete", "omni_http_request":
										g.valueTypes[inst.ID] = "HTTPResponse"
									}
								}
								// Propagate array length for length-preserving
								// intrinsics. Sorts and reverse return a fresh
								// array of the same size as their input, so the
								// caller can len() the result without losing
								// track. Without this, a sort feeding a search
								// would lose the length and produce -1.
								if isLengthPreservingArrayIntrinsic(funcName) && len(inst.Operands) > 1 {
									if expr := g.getOperandLengthExpr(inst.Operands[1]); expr != "-1" {
										g.arrayLengthExprs[inst.ID] = expr
									}
								}
								// User-defined (and other) functions returning
								// array<T> don't carry a length companion through
								// the C return ABI. Read the length back from the
								// slice header that omni_slice_make wrote. Safe
								// because omni_slice_len_real returns 0 when the
								// pointer lacks a header (NULL-safe too).
								if _, already := g.arrayLengthExprs[inst.ID]; !already {
									if isArrayParamType(inst.Type) {
										g.arrayLengthExprs[inst.ID] = fmt.Sprintf("(int32_t)omni_slice_len_real((void*)%s)", varName)
									}
								}
							}
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
				// Get the element type from the array type
				var elementType string
				if inst.Operands[0].Kind == mir.OperandValue {
					arrayOperandID := inst.Operands[0].Value
					if arrType, ok := g.valueTypes[arrayOperandID]; ok {
						// Extract element type from array type (e.g., "array<string>" -> "string")
						if strings.HasPrefix(arrType, "array<") && strings.HasSuffix(arrType, ">") {
							elementType = arrType[6 : len(arrType)-1]
						} else if strings.HasPrefix(arrType, "[]<") && strings.HasSuffix(arrType, ">") {
							elementType = arrType[3 : len(arrType)-1]
						}
					}
				}
				// Fallback to result type if element type not found
				if elementType == "" {
					elementType = inst.Type
				}

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
				} else if elementType == "string" {
					// For string arrays, use direct indexing (strings are const char*)
					arrayLength := -1
					if inst.Operands[0].Kind == mir.OperandValue {
						arrayOperandID := inst.Operands[0].Value
						if length, ok := g.arrayLengths[arrayOperandID]; ok {
							arrayLength = length
						}
					}
					if arrayLength >= 0 {
						// Use bounds checking for string arrays
						g.output.WriteString(fmt.Sprintf("  if (%s < 0 || %s >= %d) { fprintf(stderr, \"Array index out of bounds: %%d (length: %%d)\\n\", %s, %d); exit(1); }\n",
							index, index, arrayLength, index, arrayLength))
					}
					g.output.WriteString(fmt.Sprintf("  %s = %s[%s];\n", varName, target, index))
				} else if elementType == "int" || elementType == "int32" {
					// For int arrays, use runtime function with bounds checking if length is known
					arrayLength := -1
					lengthExpr := ""
					if inst.Operands[0].Kind == mir.OperandValue {
						arrayOperandID := inst.Operands[0].Value
						if length, ok := g.arrayLengths[arrayOperandID]; ok {
							arrayLength = length
						} else if expr, ok := g.arrayLengthExprs[arrayOperandID]; ok {
							lengthExpr = expr
						}
					}
					if arrayLength >= 0 {
						g.output.WriteString(fmt.Sprintf("  %s = omni_array_get_int(%s, %s, %d);\n",
							varName, target, index, arrayLength))
					} else if lengthExpr != "" {
						g.output.WriteString(fmt.Sprintf("  %s = omni_array_get_int(%s, %s, %s);\n",
							varName, target, index, lengthExpr))
					} else {
						// Length unknown (might be parameter) - still use runtime function but with -1
						// This will cause a runtime error rather than silent memory corruption
						g.errors = append(g.errors, fmt.Sprintf("WARNING: array length not known for indexing %s (ID: %d) - bounds checking disabled, may cause memory corruption", target, inst.Operands[0].Value))
						g.output.WriteString(fmt.Sprintf("  // WARNING: Array length unknown, bounds checking disabled\n"))
						g.output.WriteString(fmt.Sprintf("  %s = %s[%s]; // UNSAFE: No bounds check\n", varName, target, index))
					}
				} else {
					// For other types (float, bool, etc.), use direct indexing
					arrayLength := -1
					if inst.Operands[0].Kind == mir.OperandValue {
						arrayOperandID := inst.Operands[0].Value
						if length, ok := g.arrayLengths[arrayOperandID]; ok {
							arrayLength = length
						}
					}
					if arrayLength >= 0 {
						// Use bounds checking
						g.output.WriteString(fmt.Sprintf("  if (%s < 0 || %s >= %d) { fprintf(stderr, \"Array index out of bounds: %%d (length: %%d)\\n\", %s, %d); exit(1); }\n",
							index, index, arrayLength, index, arrayLength))
					}
					g.output.WriteString(fmt.Sprintf("  %s = %s[%s];\n", varName, target, index))
				}
			}
		}
	case "array.init":
		// Heap-allocate the backing storage via omni_slice_make so the
		// pointer carries a runtime length/capacity header. This unlocks
		// append() and slicing, and incidentally fixes the long-standing
		// "can't return arrays" bug since the storage outlives the frame.
		varName := g.getVariableName(inst.ID)
		arrayLength := len(inst.Operands)
		g.arrayLengths[inst.ID] = arrayLength

		// Extract element type from array type ("array<int>" / "[]<int>" /
		// fallback "int").
		var elementTypeStr string
		if strings.HasPrefix(inst.Type, "array<") && strings.HasSuffix(inst.Type, ">") {
			elementTypeStr = inst.Type[6 : len(inst.Type)-1]
		} else if strings.HasPrefix(inst.Type, "[]<") && strings.HasSuffix(inst.Type, ">") {
			elementTypeStr = inst.Type[3 : len(inst.Type)-1]
		} else {
			elementTypeStr = inst.Type
		}
		if elementTypeStr == "" || elementTypeStr == inferTypePlaceholder || elementTypeStr == "<infer>" {
			elementTypeStr = "int"
		}

		// Struct elements are tracked as omni_struct_t* so the runtime can
		// store pointers to per-element heap structs.
		isStruct := !g.isPrimitiveType(elementTypeStr) && !strings.Contains(elementTypeStr, "<") && !strings.Contains(elementTypeStr, "(")
		var elementType string
		if isStruct {
			elementType = "omni_struct_t*"
		} else {
			elementType = g.mapType(elementTypeStr)
		}

		// Empty literal: still allocate so subsequent appends find a header.
		if arrayLength == 0 {
			if g.declaredVariables[inst.ID] {
				g.output.WriteString(fmt.Sprintf("  %s = (%s*)omni_slice_make(0, 0, sizeof(%s));\n", varName, elementType, elementType))
			} else {
				g.output.WriteString(fmt.Sprintf("  %s* %s = (%s*)omni_slice_make(0, 0, sizeof(%s));\n", elementType, varName, elementType, elementType))
				g.declaredVariables[inst.ID] = true
			}
			g.valueTypes[inst.ID] = inst.Type
			return nil
		}

		// Non-empty literal: allocate then assign each element by index.
		if g.declaredVariables[inst.ID] {
			g.output.WriteString(fmt.Sprintf("  %s = (%s*)omni_slice_make(%d, %d, sizeof(%s));\n", varName, elementType, arrayLength, arrayLength, elementType))
		} else {
			g.output.WriteString(fmt.Sprintf("  %s* %s = (%s*)omni_slice_make(%d, %d, sizeof(%s));\n", elementType, varName, elementType, arrayLength, arrayLength, elementType))
			g.declaredVariables[inst.ID] = true
		}
		for i, op := range inst.Operands {
			g.output.WriteString(fmt.Sprintf("  %s[%d] = %s;\n", varName, i, g.getOperandValue(op)))
		}
		g.valueTypes[inst.ID] = inst.Type
	case "map.init":
		// Handle map initialization
		varName := g.getVariableName(inst.ID)
		// Assign to already declared variable
		g.output.WriteString(fmt.Sprintf("  %s = omni_map_create();\n", varName))

		// First, check if the map has mixed value types by examining all entries
		// This needs to be done before processing entries so we know to use "any" functions
		hasMixedValueTypes := false
		seenValueTypes := make(map[string]bool)
		// Build instruction map for this function to look up value types
		// We need to get fn from the context - it's passed to generateBlock
		// For now, we'll search through all functions in the module
		instructionMap := make(map[mir.ValueID]*mir.Instruction)
		if g.module != nil {
			for _, moduleFn := range g.module.Functions {
				for _, block := range moduleFn.Blocks {
					for idx := range block.Instructions {
						instPtr := &block.Instructions[idx]
						if instPtr.ID != mir.InvalidValue {
							instructionMap[instPtr.ID] = instPtr
						}
					}
				}
			}
		}
		for j := 1; j < len(inst.Operands); j += 2 {
			if j+1 < len(inst.Operands) {
				var entryValueType string
				// Try to get type from operand's Type field first
				if inst.Operands[j+1].Type != "" && inst.Operands[j+1].Type != inferTypePlaceholder {
					entryValueType = inst.Operands[j+1].Type
				} else if inst.Operands[j+1].Kind == mir.OperandValue {
					// Look up the type from valueTypes map
					if storedType, ok := g.valueTypes[inst.Operands[j+1].Value]; ok && storedType != "" && storedType != inferTypePlaceholder {
						entryValueType = storedType
					} else if valueInst, found := instructionMap[inst.Operands[j+1].Value]; found {
						// Look up the instruction that produces this value
						if valueInst.Type != "" && valueInst.Type != inferTypePlaceholder {
							entryValueType = valueInst.Type
						} else if valueInst.Op == "const" && len(valueInst.Operands) > 0 {
							// For const instructions, infer from literal
							if valueInst.Operands[0].Kind == mir.OperandLiteral {
								literal := valueInst.Operands[0].Literal
								if strings.HasPrefix(literal, "\"") && strings.HasSuffix(literal, "\"") {
									entryValueType = "string"
								} else if _, err := strconv.ParseInt(literal, 10, 64); err == nil {
									entryValueType = "int"
								} else if _, err := strconv.ParseFloat(literal, 64); err == nil {
									entryValueType = "float"
								}
							}
						}
					}
				} else if inst.Operands[j+1].Kind == mir.OperandLiteral {
					// For literals, infer type from the literal value
					literal := inst.Operands[j+1].Literal
					if strings.HasPrefix(literal, "\"") && strings.HasSuffix(literal, "\"") {
						entryValueType = "string"
					} else if _, err := strconv.ParseInt(literal, 10, 64); err == nil {
						entryValueType = "int"
					} else if _, err := strconv.ParseFloat(literal, 64); err == nil {
						entryValueType = "float"
					}
				}
				if entryValueType != "" {
					if len(seenValueTypes) > 0 && !seenValueTypes[entryValueType] {
						hasMixedValueTypes = true
						break
					}
					seenValueTypes[entryValueType] = true
				}
			}
		}

		// Process key-value pairs from operands
		for i := 0; i < len(inst.Operands); i += 2 {
			if i+1 < len(inst.Operands) {
				key := g.getOperandValue(inst.Operands[i])
				value := g.getOperandValue(inst.Operands[i+1])

				// Determine key and value types from the map type
				// Prefer the type from mapTypes (set by type checker) over inst.Type (from MIR builder)
				mapType := inst.Type
				if storedType, ok := g.mapTypes[inst.ID]; ok && storedType != "" {
					mapType = storedType
				}
				// If mapType is still not set or is inferred, try to get it from inst.Type
				if mapType == "" || mapType == inferTypePlaceholder {
					mapType = inst.Type
				}
				if strings.HasPrefix(mapType, "map<") && strings.HasSuffix(mapType, ">") {
					inner := mapType[4 : len(mapType)-1] // Remove "map<" and ">"
					parts := strings.Split(inner, ",")
					if len(parts) == 2 {
						keyType := strings.TrimSpace(parts[0])
						valueType := strings.TrimSpace(parts[1])

						// Get the actual runtime types of the operands
						// Always determine actual types from operands, not from map type
						actualKeyType := keyType
						actualValueType := valueType

						// Determine actual key type from operand
						if inst.Operands[i].Type != "" && inst.Operands[i].Type != inferTypePlaceholder {
							actualKeyType = inst.Operands[i].Type
						} else if inst.Operands[i].Kind == mir.OperandValue {
							// Look up the type from valueTypes map
							if storedType, ok := g.valueTypes[inst.Operands[i].Value]; ok && storedType != "" && storedType != inferTypePlaceholder {
								actualKeyType = storedType
							} else if valueInst, found := instructionMap[inst.Operands[i].Value]; found {
								if valueInst.Type != "" && valueInst.Type != inferTypePlaceholder {
									actualKeyType = valueInst.Type
								}
							}
						}

						// Determine actual value type from operand (always, not just when valueType is "any")
						if inst.Operands[i+1].Type != "" && inst.Operands[i+1].Type != inferTypePlaceholder {
							actualValueType = inst.Operands[i+1].Type
						} else if inst.Operands[i+1].Kind == mir.OperandValue {
							// Look up the type from valueTypes map
							if storedType, ok := g.valueTypes[inst.Operands[i+1].Value]; ok && storedType != "" && storedType != inferTypePlaceholder {
								actualValueType = storedType
							} else if valueInst, found := instructionMap[inst.Operands[i+1].Value]; found {
								if valueInst.Type != "" && valueInst.Type != inferTypePlaceholder {
									actualValueType = valueInst.Type
								} else if valueInst.Op == "const" && len(valueInst.Operands) > 0 {
									// For const instructions, infer from literal
									if valueInst.Operands[0].Kind == mir.OperandLiteral {
										literal := valueInst.Operands[0].Literal
										if strings.HasPrefix(literal, "\"") && strings.HasSuffix(literal, "\"") {
											actualValueType = "string"
										} else if _, err := strconv.ParseInt(literal, 10, 64); err == nil {
											actualValueType = "int"
										} else if _, err := strconv.ParseFloat(literal, 64); err == nil {
											actualValueType = "float"
										}
									}
								}
							}
						} else if inst.Operands[i+1].Kind == mir.OperandLiteral {
							// For literals, infer type from the literal value
							literal := inst.Operands[i+1].Literal
							if strings.HasPrefix(literal, "\"") && strings.HasSuffix(literal, "\"") {
								actualValueType = "string"
							} else if _, err := strconv.ParseInt(literal, 10, 64); err == nil {
								actualValueType = "int"
							} else if _, err := strconv.ParseFloat(literal, 64); err == nil {
								actualValueType = "float"
							}
						}

						// hasMixedValueTypes is already computed above for all entries

						// Also check if the current value type doesn't match the map's inferred value type
						// This handles the case where the map type is inferred incorrectly from the first entry
						// Get the actual type of the current value
						currentValueType := actualValueType
						if currentValueType == "" || currentValueType == inferTypePlaceholder {
							// Try to get from operand
							if inst.Operands[i+1].Type != "" && inst.Operands[i+1].Type != inferTypePlaceholder {
								currentValueType = inst.Operands[i+1].Type
							} else if inst.Operands[i+1].Kind == mir.OperandValue {
								if storedType, ok := g.valueTypes[inst.Operands[i+1].Value]; ok && storedType != "" && storedType != inferTypePlaceholder {
									currentValueType = storedType
								} else if valueInst, found := instructionMap[inst.Operands[i+1].Value]; found {
									if valueInst.Type != "" && valueInst.Type != inferTypePlaceholder {
										currentValueType = valueInst.Type
									}
								}
							}
						}

						// If the current value type doesn't match the map's value type, we have mixed types
						// OR if we already detected mixed types, use "any"
						if !hasMixedValueTypes && valueType != "any" && currentValueType != "" && currentValueType != valueType && currentValueType != inferTypePlaceholder {
							hasMixedValueTypes = true
						}

						// Generate appropriate put function call based on types
						// Use "any" as the value type if the map has mixed value types or if valueType is "any"
						mapValueType := valueType
						if hasMixedValueTypes || valueType == "any" {
							// Use "any" as the value type for the function selection
							mapValueType = "any"
						}
						putFunc := g.getMapPutFunction(keyType, mapValueType)
						if putFunc != "" {
							// Check if we need to pass type constants for any types
							if strings.Contains(putFunc, "_any") || strings.Contains(putFunc, "any_") {
								// Need to pass type constants
								if keyType == "any" {
									keyTypeConst := g.getOmniTypeConstant(actualKeyType)
									if valueType == "any" {
										valueTypeConst := g.getOmniTypeConstant(actualValueType)
										// Both are any: omni_map_put_any_any(map, key, key_type, value, value_type)
										g.output.WriteString(fmt.Sprintf("  %s(%s, %s, %s, %s, %s);\n", putFunc, varName, key, keyTypeConst, value, valueTypeConst))
									} else {
										// Key is any, value is not: omni_map_put_any_*(map, key, key_type, value)
										g.output.WriteString(fmt.Sprintf("  %s(%s, %s, %s, %s);\n", putFunc, varName, key, keyTypeConst, value))
									}
								} else if valueType == "any" {
									valueTypeConst := g.getOmniTypeConstant(actualValueType)
									// Value is any, key is not: omni_map_put_*_any(map, key, value, value_type)
									g.output.WriteString(fmt.Sprintf("  %s(%s, %s, %s, %s);\n", putFunc, varName, key, value, valueTypeConst))
								}
							} else {
								// Standard function call without type constants
								g.output.WriteString(fmt.Sprintf("  %s(%s, %s, %s);\n", putFunc, varName, key, value))
							}
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
		// Tag the value with its OmniLang type so interface method dispatch
		// can resolve the concrete implementation at runtime.
		if inst.Type != "" && inst.Type != inferTypePlaceholder && !strings.Contains(inst.Type, "<") {
			g.output.WriteString(fmt.Sprintf("  omni_struct_set_type_name(%s, \"%s\");\n", varName, inst.Type))
		}

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
							// Check for complex types
							if strings.HasPrefix(fieldType, "array<") || strings.HasPrefix(fieldType, "[]<") {
								// Array type - extract element type
								var elementTypeStr string
								if strings.HasPrefix(fieldType, "array<") {
									elementTypeStr = fieldType[6 : len(fieldType)-1]
								} else if strings.HasPrefix(fieldType, "[]<") {
									elementTypeStr = fieldType[3 : len(fieldType)-1]
								}
								elementTypeConstant := g.getOmniTypeConstant(elementTypeStr)
								// Get array length if available
								arrayLength := -1
								if fieldValueOp.Kind == mir.OperandValue {
									if length, ok := g.arrayLengths[fieldValueOp.Value]; ok {
										arrayLength = length
									}
								}
								if arrayLength < 0 {
									arrayLength = 0 // Default to 0 if unknown
								}
								g.output.WriteString(fmt.Sprintf("  omni_struct_set_array_field(%s, \"%s\", %s, %s, %d);\n", varName, fieldName, fieldValue, elementTypeConstant, arrayLength))
							} else if strings.HasPrefix(fieldType, "map<") {
								// Map type
								g.output.WriteString(fmt.Sprintf("  omni_struct_set_map_field(%s, \"%s\", %s);\n", varName, fieldName, fieldValue))
							} else if fieldType == "null" || fieldValue == "NULL" || fieldValue == "nil" || (fieldValueOp.Kind == mir.OperandLiteral && fieldValueOp.Literal == "nil") {
								// Null/nil value
								g.output.WriteString(fmt.Sprintf("  omni_struct_set_null_field(%s, \"%s\");\n", varName, fieldName))
							} else if !g.isPrimitiveType(fieldType) && !strings.Contains(fieldType, "<") && !strings.Contains(fieldType, "(") {
								// Struct type (not a primitive, not an array, not a map, not a function)
								g.output.WriteString(fmt.Sprintf("  omni_struct_set_struct_field(%s, \"%s\", %s);\n", varName, fieldName, fieldValue))
							} else if !g.isPrimitiveType(fieldType) {
								// Unknown complex type - try to handle as struct
								g.output.WriteString(fmt.Sprintf("  // WARNING: Unknown complex type %s for field %s, treating as struct\n", fieldType, fieldName))
								g.output.WriteString(fmt.Sprintf("  omni_struct_set_struct_field(%s, \"%s\", %s);\n", varName, fieldName, fieldValue))
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
						// Check for complex types
						if strings.HasPrefix(fieldType, "array<") || strings.HasPrefix(fieldType, "[]<") {
							// Array type - extract element type
							var elementTypeStr string
							if strings.HasPrefix(fieldType, "array<") {
								elementTypeStr = fieldType[6 : len(fieldType)-1]
							} else if strings.HasPrefix(fieldType, "[]<") {
								elementTypeStr = fieldType[3 : len(fieldType)-1]
							}
							elementTypeConstant := g.getOmniTypeConstant(elementTypeStr)
							// Get array length if available
							arrayLength := -1
							if fieldValueOp.Kind == mir.OperandValue {
								if length, ok := g.arrayLengths[fieldValueOp.Value]; ok {
									arrayLength = length
								}
							}
							if arrayLength < 0 {
								arrayLength = 0 // Default to 0 if unknown
							}
							g.output.WriteString(fmt.Sprintf("  omni_struct_set_array_field(%s, \"%s\", %s, %s, %d);\n", varName, fieldName, fieldValue, elementTypeConstant, arrayLength))
						} else if strings.HasPrefix(fieldType, "map<") {
							// Map type
							g.output.WriteString(fmt.Sprintf("  omni_struct_set_map_field(%s, \"%s\", %s);\n", varName, fieldName, fieldValue))
						} else if fieldType == "null" || fieldValue == "NULL" || fieldValue == "nil" || (fieldValueOp.Kind == mir.OperandLiteral && fieldValueOp.Literal == "nil") {
							// Null/nil value
							g.output.WriteString(fmt.Sprintf("  omni_struct_set_null_field(%s, \"%s\");\n", varName, fieldName))
						} else if !g.isPrimitiveType(fieldType) && !strings.Contains(fieldType, "<") && !strings.Contains(fieldType, "(") {
							// Struct type (not a primitive, not an array, not a map, not a function)
							g.output.WriteString(fmt.Sprintf("  omni_struct_set_struct_field(%s, \"%s\", %s);\n", varName, fieldName, fieldValue))
						} else {
							// Default to int for unknown primitive types
							g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field(%s, \"%s\", %s);\n", varName, fieldName, fieldValue))
						}
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
			// Special handling for known HTTPResponse fields - use field name to infer type
			fieldType := inst.Type
			if fieldType == "" || fieldType == "<inferred>" || fieldType == inferTypePlaceholder {
				// Infer from common field names in std structs (HTTPRequest/HTTPResponse/URL/Context)
				switch fieldName {
				case "body", "status_text", "method", "url", "path", "host", "scheme", "query", "fragment", "address", "content_type", "name", "filename":
					fieldType = "string"
				case "status_code", "port", "socket", "size":
					fieldType = "int"
				case "is_ipv4", "is_ipv6":
					fieldType = "bool"
				case "request", "response":
					fieldType = "omni_struct_t*"
				}
			}

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

			// Special handling for HTTPResponse struct - use direct field access
			// Check if the struct variable is an HTTPResponse type
			structType := ""
			if inst.Operands[0].Kind == mir.OperandValue {
				if storedType, ok := g.valueTypes[inst.Operands[0].Value]; ok && storedType != "" {
					structType = storedType
				}
			}
			// Check if this is an HTTPResponse by checking the variable's stored type
			// HTTP functions return HTTPResponse type, which gets stored in valueTypes
			isHTTPResponse := strings.Contains(structType, "HTTPResponse")

			// Also check the operand's type field if structType wasn't found
			if !isHTTPResponse && inst.Operands[0].Type != "" {
				isHTTPResponse = strings.Contains(inst.Operands[0].Type, "HTTPResponse")
			}

			// Last resort: only assume HTTPResponse for status_code/status_text fields since
			// `body` is shared with Context; using ->body on an opaque Context pointer is invalid C.
			if !isHTTPResponse && (fieldName == "status_code" || fieldName == "status_text") {
				if inst.Operands[0].Kind == mir.OperandValue && structType == "" {
					isHTTPResponse = true
				}
			}

			// URL is a concrete C struct (omni_url_t) returned by
			// omni_url_parse — use direct field access.
			isURL := strings.Contains(structType, "URL") && !strings.Contains(structType, "HTTPResponse") && !strings.Contains(structType, "URLs")
			if !isURL && inst.Operands[0].Type != "" && strings.Contains(inst.Operands[0].Type, "URL") {
				isURL = true
			}
			if isURL {
				switch fieldName {
				case "scheme", "host", "path", "query", "fragment":
					if _, alreadyDeclared := g.declaredVariables[inst.ID]; alreadyDeclared {
						g.output.WriteString(fmt.Sprintf("  %s = %s->%s;\n", varName, structVar, fieldName))
					} else {
						g.output.WriteString(fmt.Sprintf("  const char* %s = %s->%s;\n", varName, structVar, fieldName))
					}
					g.valueTypes[inst.ID] = "string"
					return nil
				case "port":
					if _, alreadyDeclared := g.declaredVariables[inst.ID]; alreadyDeclared {
						g.output.WriteString(fmt.Sprintf("  %s = %s->port;\n", varName, structVar))
					} else {
						g.output.WriteString(fmt.Sprintf("  int32_t %s = %s->port;\n", varName, structVar))
					}
					g.valueTypes[inst.ID] = "int"
					return nil
				}
			}

			// IPAddress is a concrete C struct (omni_ip_address_t) returned
			// by omni_ip_parse — use direct field access.
			isIPAddress := strings.Contains(structType, "IPAddress")
			if !isIPAddress && inst.Operands[0].Type != "" && strings.Contains(inst.Operands[0].Type, "IPAddress") {
				isIPAddress = true
			}
			if isIPAddress {
				switch fieldName {
				case "address":
					if _, alreadyDeclared := g.declaredVariables[inst.ID]; alreadyDeclared {
						g.output.WriteString(fmt.Sprintf("  %s = %s->address;\n", varName, structVar))
					} else {
						g.output.WriteString(fmt.Sprintf("  const char* %s = %s->address;\n", varName, structVar))
					}
					g.valueTypes[inst.ID] = "string"
					return nil
				case "is_ipv4", "is_ipv6":
					if _, alreadyDeclared := g.declaredVariables[inst.ID]; alreadyDeclared {
						g.output.WriteString(fmt.Sprintf("  %s = %s->%s;\n", varName, structVar, fieldName))
					} else {
						g.output.WriteString(fmt.Sprintf("  int32_t %s = %s->%s;\n", varName, structVar, fieldName))
					}
					g.valueTypes[inst.ID] = "bool"
					return nil
				}
			}

			if isHTTPResponse {
				// HTTPResponse is a concrete struct - use direct field access
				if fieldName == "status_code" {
					if _, alreadyDeclared := g.declaredVariables[inst.ID]; alreadyDeclared {
						g.output.WriteString(fmt.Sprintf("  %s = %s->status_code;\n", varName, structVar))
					} else {
						g.output.WriteString(fmt.Sprintf("  int32_t %s = %s->status_code;\n", varName, structVar))
					}
					g.valueTypes[inst.ID] = "int"
				} else if fieldName == "body" {
					if _, alreadyDeclared := g.declaredVariables[inst.ID]; alreadyDeclared {
						g.output.WriteString(fmt.Sprintf("  %s = %s->body;\n", varName, structVar))
					} else {
						g.output.WriteString(fmt.Sprintf("  const char* %s = %s->body;\n", varName, structVar))
					}
					g.valueTypes[inst.ID] = "string"
				} else if fieldName == "status_text" {
					if _, alreadyDeclared := g.declaredVariables[inst.ID]; alreadyDeclared {
						g.output.WriteString(fmt.Sprintf("  %s = %s->status_text;\n", varName, structVar))
					} else {
						g.output.WriteString(fmt.Sprintf("  const char* %s = %s->status_text;\n", varName, structVar))
					}
					g.valueTypes[inst.ID] = "string"
				} else {
					// Unknown field - fall back to generic accessor (shouldn't happen)
					if _, alreadyDeclared := g.declaredVariables[inst.ID]; alreadyDeclared {
						g.output.WriteString(fmt.Sprintf("  %s = omni_struct_get_int_field((omni_struct_t*)%s, \"%s\");\n", varName, structVar, fieldName))
					} else {
						g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_int_field((omni_struct_t*)%s, \"%s\");\n", varName, structVar, fieldName))
					}
					g.valueTypes[inst.ID] = "int"
				}
			} else if fieldName == "body" || fieldName == "status_text" {
				// Generic struct with body/status_text fields - use generic getter
				if _, alreadyDeclared := g.declaredVariables[inst.ID]; alreadyDeclared {
					g.output.WriteString(fmt.Sprintf("  %s = omni_struct_get_string_field(%s, \"%s\");\n", varName, structVar, fieldName))
				} else {
					g.output.WriteString(fmt.Sprintf("  const char* %s = omni_struct_get_string_field(%s, \"%s\");\n", varName, structVar, fieldName))
				}
				g.valueTypes[inst.ID] = "string"
			} else if fieldName == "status_code" {
				// Generic struct with status_code field - use generic getter
				if _, alreadyDeclared := g.declaredVariables[inst.ID]; alreadyDeclared {
					g.output.WriteString(fmt.Sprintf("  %s = omni_struct_get_int_field(%s, \"%s\");\n", varName, structVar, fieldName))
				} else {
					g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_int_field(%s, \"%s\");\n", varName, structVar, fieldName))
				}
				g.valueTypes[inst.ID] = "int"
			} else if _, alreadyDeclared := g.declaredVariables[inst.ID]; alreadyDeclared {
				// Variable already declared at top - just assign
				if fieldName == "body" || fieldName == "status_text" {
					g.output.WriteString(fmt.Sprintf("  %s = omni_struct_get_string_field(%s, \"%s\");\n", varName, structVar, fieldName))
					g.valueTypes[inst.ID] = "string"
				} else if fieldName == "status_code" {
					g.output.WriteString(fmt.Sprintf("  %s = omni_struct_get_int_field(%s, \"%s\");\n", varName, structVar, fieldName))
					g.valueTypes[inst.ID] = "int"
				} else {
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
					case "omni_struct_t*":
						g.output.WriteString(fmt.Sprintf("  %s = omni_struct_get_struct_field(%s, \"%s\");\n", varName, structVar, fieldName))
						g.valueTypes[inst.ID] = "omni_struct_t*"
					default:
						if !g.isPrimitiveType(fieldType) && fieldType != "" && !strings.Contains(fieldType, "<") && !strings.Contains(fieldType, "(") {
							g.output.WriteString(fmt.Sprintf("  %s = omni_struct_get_struct_field(%s, \"%s\");\n", varName, structVar, fieldName))
							g.valueTypes[inst.ID] = "omni_struct_t*"
						} else {
							g.output.WriteString(fmt.Sprintf("  %s = omni_struct_get_int_field(%s, \"%s\");\n", varName, structVar, fieldName))
							g.valueTypes[inst.ID] = "int"
						}
					}
				}
			} else {
				// Variable not declared - declare inline
				if fieldName == "body" || fieldName == "status_text" {
					g.output.WriteString(fmt.Sprintf("  const char* %s = omni_struct_get_string_field(%s, \"%s\");\n", varName, structVar, fieldName))
					g.valueTypes[inst.ID] = "string"
				} else if fieldName == "status_code" {
					g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_int_field(%s, \"%s\");\n", varName, structVar, fieldName))
					g.valueTypes[inst.ID] = "int"
				} else {
					// Use appropriate getter based on field type
					switch fieldType {
					case "string":
						g.output.WriteString(fmt.Sprintf("  const char* %s = omni_struct_get_string_field(%s, \"%s\");\n", varName, structVar, fieldName))
						g.valueTypes[inst.ID] = "string"
					case "float", "double":
						g.output.WriteString(fmt.Sprintf("  double %s = omni_struct_get_float_field(%s, \"%s\");\n", varName, structVar, fieldName))
						g.valueTypes[inst.ID] = fieldType
					case "bool":
						g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_bool_field(%s, \"%s\");\n", varName, structVar, fieldName))
						g.valueTypes[inst.ID] = "bool"
					case "omni_struct_t*":
						g.output.WriteString(fmt.Sprintf("  omni_struct_t* %s = omni_struct_get_struct_field(%s, \"%s\");\n", varName, structVar, fieldName))
						g.valueTypes[inst.ID] = "omni_struct_t*"
					default:
						if !g.isPrimitiveType(fieldType) && fieldType != "" && !strings.Contains(fieldType, "<") && !strings.Contains(fieldType, "(") {
							g.output.WriteString(fmt.Sprintf("  omni_struct_t* %s = omni_struct_get_struct_field(%s, \"%s\");\n", varName, structVar, fieldName))
							g.valueTypes[inst.ID] = "omni_struct_t*"
						} else {
							g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_struct_get_int_field(%s, \"%s\");\n", varName, structVar, fieldName))
							g.valueTypes[inst.ID] = "int"
						}
					}
				}
			}
		}
	case "index.assign":
		// Handle array element assignment: arr[i] = value.
		// For map element assignment use the map_set_* runtime helpers; for
		// arrays use a plain C subscript assignment (bounds checking is the
		// caller's responsibility).
		if len(inst.Operands) >= 3 {
			target := g.getOperandValue(inst.Operands[0])
			index := g.getOperandValue(inst.Operands[1])
			value := g.getOperandValue(inst.Operands[2])
			isMap := false
			if inst.Operands[0].Kind == mir.OperandValue {
				if mt, ok := g.mapTypes[inst.Operands[0].Value]; ok && strings.HasPrefix(mt, "map<") {
					isMap = true
				} else if st, ok := g.valueTypes[inst.Operands[0].Value]; ok && strings.HasPrefix(st, "map<") {
					isMap = true
				}
			}
			if isMap {
				mapType := "map<string,int>"
				if st, ok := g.valueTypes[inst.Operands[0].Value]; ok && strings.HasPrefix(st, "map<") {
					mapType = st
				}
				keyType, valueType := g.extractMapTypes(mapType)
				setFn := g.getMapSetFunction(keyType, valueType)
				if setFn != "" {
					g.output.WriteString(fmt.Sprintf("  %s(%s, %s, %s);\n", setFn, target, index, value))
				}
			} else {
				g.output.WriteString(fmt.Sprintf("  %s[%s] = %s;\n", target, index, value))
			}
		}
	case "member.assign":
		// Handle struct field assignment: obj.field = value
		if len(inst.Operands) >= 3 {
			structVar := g.getOperandValue(inst.Operands[0])
			fieldName := inst.Operands[1].Literal
			fieldValue := g.getOperandValue(inst.Operands[2])

			// Determine field type from instruction type or operand type
			fieldType := inst.Type
			if fieldType == "" || fieldType == "<inferred>" || fieldType == inferTypePlaceholder {
				if inst.Operands[2].Type != "" && inst.Operands[2].Type != inferTypePlaceholder {
					fieldType = inst.Operands[2].Type
				} else if inst.Operands[2].Kind == mir.OperandValue {
					if storedType, ok := g.valueTypes[inst.Operands[2].Value]; ok && storedType != "" {
						fieldType = storedType
					}
				}
			}

			// Default to int if type not found
			if fieldType == "" || fieldType == "<inferred>" || fieldType == inferTypePlaceholder {
				fieldType = "int"
			}

			// Special handling for HTTPResponse struct - use direct field access
			structType := ""
			if inst.Operands[0].Kind == mir.OperandValue {
				if storedType, ok := g.valueTypes[inst.Operands[0].Value]; ok && storedType != "" {
					structType = storedType
				}
			}
			isHTTPResponse := strings.Contains(structType, "HTTPResponse")
			if !isHTTPResponse && inst.Operands[0].Type != "" {
				isHTTPResponse = strings.Contains(inst.Operands[0].Type, "HTTPResponse")
			}

			// Use direct field assignment for HTTPResponse
			if isHTTPResponse {
				if fieldName == "status_code" {
					g.output.WriteString(fmt.Sprintf("  %s->status_code = %s;\n", structVar, fieldValue))
				} else if fieldName == "body" {
					g.output.WriteString(fmt.Sprintf("  %s->body = %s;\n", structVar, fieldValue))
				} else if fieldName == "status_text" {
					g.output.WriteString(fmt.Sprintf("  %s->status_text = %s;\n", structVar, fieldValue))
				} else {
					// Unknown field - fall back to generic setter
					switch fieldType {
					case "string":
						g.output.WriteString(fmt.Sprintf("  omni_struct_set_string_field((omni_struct_t*)%s, \"%s\", %s);\n", structVar, fieldName, fieldValue))
					case "float", "double":
						g.output.WriteString(fmt.Sprintf("  omni_struct_set_float_field((omni_struct_t*)%s, \"%s\", %s);\n", structVar, fieldName, fieldValue))
					case "bool":
						g.output.WriteString(fmt.Sprintf("  omni_struct_set_bool_field((omni_struct_t*)%s, \"%s\", %s);\n", structVar, fieldName, fieldValue))
					default:
						g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field((omni_struct_t*)%s, \"%s\", %s);\n", structVar, fieldName, fieldValue))
					}
				}
			} else {
				// Generic struct - use generic setters
				switch fieldType {
				case "string":
					// Handle string literals specially
					if inst.Operands[2].Kind == mir.OperandLiteral && strings.HasPrefix(inst.Operands[2].Literal, "\"") && strings.HasSuffix(inst.Operands[2].Literal, "\"") {
						g.output.WriteString(fmt.Sprintf("  omni_struct_set_string_field(%s, \"%s\", %s);\n", structVar, fieldName, inst.Operands[2].Literal))
					} else {
						g.output.WriteString(fmt.Sprintf("  omni_struct_set_string_field(%s, \"%s\", %s);\n", structVar, fieldName, fieldValue))
					}
				case "float", "double":
					g.output.WriteString(fmt.Sprintf("  omni_struct_set_float_field(%s, \"%s\", %s);\n", structVar, fieldName, fieldValue))
				case "bool":
					g.output.WriteString(fmt.Sprintf("  omni_struct_set_bool_field(%s, \"%s\", %s);\n", structVar, fieldName, fieldValue))
				default:
					g.output.WriteString(fmt.Sprintf("  omni_struct_set_int_field(%s, \"%s\", %s);\n", structVar, fieldName, fieldValue))
				}
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

			// Struct-vs-struct: compare as pointers (identity). Only fall into the
			// string-comparison path when at least one operand is actually a string.
			isLeftStruct := !g.isPrimitiveType(leftType) && !strings.Contains(leftType, "<") && !strings.Contains(leftType, "(") && leftType != "" && leftType != "string" && leftType != "const char*" && leftType != "char*"
			isRightStruct := !g.isPrimitiveType(rightType) && !strings.Contains(rightType, "<") && !strings.Contains(rightType, "(") && rightType != "" && rightType != "string" && rightType != "const char*" && rightType != "char*"
			bothStructs := isLeftStruct && isRightStruct

			// If either operand is a string, use string comparison function
			if (leftType == "string" || rightType == "string") && !bothStructs {
				// Convert both operands to strings if needed
				leftStr := left
				rightStr := right
				if leftType != "string" && leftType != "const char*" && leftType != "char*" {
					leftStr = g.convertOperandToString(inst.Operands[0])
				}
				if rightType != "string" && rightType != "const char*" && rightType != "char*" {
					rightStr = g.convertOperandToString(inst.Operands[1])
				}
				switch inst.Op {
				case "cmp.eq":
					g.output.WriteString(fmt.Sprintf("  %s = omni_string_equals(%s, %s) ? 1 : 0;\n",
						varName, leftStr, rightStr))
				case "cmp.neq":
					g.output.WriteString(fmt.Sprintf("  %s = omni_string_equals(%s, %s) ? 0 : 1;\n",
						varName, leftStr, rightStr))
				case "cmp.lt":
					g.output.WriteString(fmt.Sprintf("  %s = (omni_string_compare(%s, %s) < 0) ? 1 : 0;\n",
						varName, leftStr, rightStr))
				case "cmp.lte":
					g.output.WriteString(fmt.Sprintf("  %s = (omni_string_compare(%s, %s) <= 0) ? 1 : 0;\n",
						varName, leftStr, rightStr))
				case "cmp.gt":
					g.output.WriteString(fmt.Sprintf("  %s = (omni_string_compare(%s, %s) > 0) ? 1 : 0;\n",
						varName, leftStr, rightStr))
				case "cmp.gte":
					g.output.WriteString(fmt.Sprintf("  %s = (omni_string_compare(%s, %s) >= 0) ? 1 : 0;\n",
						varName, leftStr, rightStr))
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
			g.output.WriteString(fmt.Sprintf("  intptr_t %s = omni_file_open(%s, %s);\n",
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
			message := g.convertOperandToString(inst.Operands[1])
			g.output.WriteString(fmt.Sprintf("  omni_assert(%s, %s);\n", condition, message))
		}
	case "assert.eq":
		// Handle equality assertion
		if len(inst.Operands) >= 3 {
			expectedOp := inst.Operands[0]
			actualOp := inst.Operands[1]
			messageOp := inst.Operands[2]

			// Determine the actual types of the operands
			expectedType := expectedOp.Type
			actualType := actualOp.Type
			if expectedType == "" || expectedType == inferTypePlaceholder {
				if expectedOp.Kind == mir.OperandValue {
					if storedType, ok := g.valueTypes[expectedOp.Value]; ok && storedType != "" {
						expectedType = storedType
					}
				}
			}
			if actualType == "" || actualType == inferTypePlaceholder {
				if actualOp.Kind == mir.OperandValue {
					if storedType, ok := g.valueTypes[actualOp.Value]; ok && storedType != "" {
						actualType = storedType
					}
				}
			}

			// Use the instruction type if operand types are unknown
			if expectedType == "" || expectedType == inferTypePlaceholder {
				expectedType = inst.Type
			}
			if actualType == "" || actualType == inferTypePlaceholder {
				actualType = inst.Type
			}

			expected := g.getOperandValue(expectedOp)
			actual := g.getOperandValue(actualOp)
			message := g.convertOperandToString(messageOp)

			// Use appropriate assertion function based on operand types
			// Prefer string comparison if either operand is a string
			if expectedType == "string" || actualType == "string" {
				// Convert both to strings if needed
				expectedStr := g.convertOperandToString(expectedOp)
				actualStr := g.convertOperandToString(actualOp)
				g.output.WriteString(fmt.Sprintf("  omni_assert_eq_string(%s, %s, %s);\n", expectedStr, actualStr, message))
			} else if expectedType == "int" || actualType == "int" || expectedType == "bool" || actualType == "bool" {
				// For int/bool comparisons, ensure both are ints
				g.output.WriteString(fmt.Sprintf("  omni_assert_eq_int(%s, %s, %s);\n", expected, actual, message))
			} else if expectedType == "float" || expectedType == "double" || actualType == "float" || actualType == "double" {
				// For float comparisons
				g.output.WriteString(fmt.Sprintf("  omni_assert_eq_float(%s, %s, %s);\n", expected, actual, message))
			} else {
				// For structs or unknown types, convert to strings for comparison
				expectedStr := g.convertOperandToString(expectedOp)
				actualStr := g.convertOperandToString(actualOp)
				g.output.WriteString(fmt.Sprintf("  omni_assert_eq_string(%s, %s, %s);\n", expectedStr, actualStr, message))
			}
		}
	case "assert.true":
		// Handle true assertion
		if len(inst.Operands) >= 2 {
			condition := g.getOperandValue(inst.Operands[0])
			message := g.convertOperandToString(inst.Operands[1])
			g.output.WriteString(fmt.Sprintf("  omni_assert_true(%s, %s);\n", condition, message))
		}
	case "assert.false":
		// Handle false assertion
		if len(inst.Operands) >= 2 {
			condition := g.getOperandValue(inst.Operands[0])
			message := g.convertOperandToString(inst.Operands[1])
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
			// Variable is already declared at the top, just assign the function pointer
			g.output.WriteString(fmt.Sprintf("  %s = %s;\n", varName, mappedName))
		}
	case "func.assign":
		// Handle function assignment: func_var = function_name
		if len(inst.Operands) >= 1 {
			funcName := g.getOperandValue(inst.Operands[0])
			varName := g.getVariableName(inst.ID)
			// Variable is already declared at the top, just assign the function pointer
			g.output.WriteString(fmt.Sprintf("  %s = %s;\n", varName, funcName))
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
	case "std.io.eprint":
		if len(inst.Operands) >= 1 {
			g.emitPrintTo(inst.Operands[0], false, true)
		}
	case "std.io.eprintln":
		if len(inst.Operands) >= 1 {
			g.emitPrintTo(inst.Operands[0], true, true)
		} else {
			g.output.WriteString("  omni_eprintln_string(\"\");\n")
		}
	case "std.io.flush":
		g.output.WriteString("  omni_io_flush();\n")
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

			// Ensure variable is declared if it wasn't declared at the top
			// This can happen if the await instruction wasn't collected in allVariables
			// We'll determine the type first, then check if we need to declare it

			// Determine the await function based on result type
			// The inst.Type should already be the unwrapped type (e.g., "string" not "Promise<string>")
			resultType := inst.Type

			// If type is empty or unknown, try to infer from the operand's promise type
			if resultType == "" || resultType == "<inferred>" || resultType == "<infer>" || resultType == inferTypePlaceholder {
				// Try to get type from operand's promise type
				if len(inst.Operands) > 0 && inst.Operands[0].Kind == mir.OperandValue {
					operandID := inst.Operands[0].Value
					if storedType, ok := g.valueTypes[operandID]; ok {
						// Extract inner type from Promise<T>
						if strings.HasPrefix(storedType, "Promise<") && strings.HasSuffix(storedType, ">") {
							innerType := storedType[len("Promise<") : len(storedType)-1]
							// Only use it if it's not an infer placeholder
							if innerType != "<infer>" && innerType != "<inferred>" && innerType != inferTypePlaceholder && innerType != "" {
								resultType = innerType
							}
						}
					}
					// Also check the operand's type field
					if resultType == "" || resultType == "<inferred>" || resultType == "<infer>" || resultType == inferTypePlaceholder {
						if inst.Operands[0].Type != "" && inst.Operands[0].Type != inferTypePlaceholder {
							if strings.HasPrefix(inst.Operands[0].Type, "Promise<") && strings.HasSuffix(inst.Operands[0].Type, ">") {
								innerType := inst.Operands[0].Type[len("Promise<") : len(inst.Operands[0].Type)-1]
								if innerType != "<infer>" && innerType != "<inferred>" && innerType != inferTypePlaceholder && innerType != "" {
									resultType = innerType
								}
							}
						}
					}
				}
			}

			// Determine C type for the result
			cType := "int32_t" // default
			switch resultType {
			case "int":
				cType = "int32_t"
			case "string":
				cType = "const char*"
			case "float", "double":
				cType = "double"
			case "bool":
				cType = "int32_t"
			}

			// Check if variable was declared at the top of the function
			// If not, declare it inline before assignment
			// We check by seeing if the variable name exists in g.variables for this ID
			// But actually, g.variables is populated during getVariableName, so we need a different check
			// We'll use a simple heuristic: if the variable wasn't in allVariables during declaration,
			// we need to declare it now. We can track this by checking if we've output a declaration.
			// Actually, a simpler approach: always check if we need to declare by looking at the
			// instruction's presence in the initial variable collection. Since we can't access
			// allVariables here, we'll use a workaround: declare inline if the variable
			// assignment would be the first use (which we can't easily detect).
			// Better approach: track declared variables in a set during declaration phase,
			// then check that set here. But for now, let's just always declare await variables
			// inline to be safe, since they might not have been collected.

			// Check if variable was already declared at the top of the function
			// If not, declare it inline with initialization
			needsDeclaration := !g.declaredVariables[inst.ID]

			switch resultType {
			case "int":
				if needsDeclaration {
					g.output.WriteString(fmt.Sprintf("  %s %s = omni_await_int(%s);\n", cType, varName, promiseVar))
					g.declaredVariables[inst.ID] = true
				} else {
					g.output.WriteString(fmt.Sprintf("  %s = omni_await_int(%s);\n", varName, promiseVar))
				}
			case "string":
				if needsDeclaration {
					g.output.WriteString(fmt.Sprintf("  %s %s = omni_await_string(%s);\n", cType, varName, promiseVar))
					g.declaredVariables[inst.ID] = true
				} else {
					g.output.WriteString(fmt.Sprintf("  %s = omni_await_string(%s);\n", varName, promiseVar))
				}
				// Track this string for cleanup (omni_await_string returns heap-allocated string)
				g.stringsToFree[inst.ID] = true
			case "float", "double":
				if needsDeclaration {
					g.output.WriteString(fmt.Sprintf("  %s %s = omni_await_float(%s);\n", cType, varName, promiseVar))
					g.declaredVariables[inst.ID] = true
				} else {
					g.output.WriteString(fmt.Sprintf("  %s = omni_await_float(%s);\n", varName, promiseVar))
				}
			case "bool":
				if needsDeclaration {
					g.output.WriteString(fmt.Sprintf("  %s %s = omni_await_bool(%s);\n", cType, varName, promiseVar))
					g.declaredVariables[inst.ID] = true
				} else {
					g.output.WriteString(fmt.Sprintf("  %s = omni_await_bool(%s);\n", varName, promiseVar))
				}
			default:
				// Check if resultType is still an infer placeholder - this means we couldn't determine the type
				needsDecl := !g.declaredVariables[inst.ID]
				if resultType == "<infer>" || resultType == "<inferred>" || resultType == inferTypePlaceholder || resultType == "" {
					// Try one more time to get it from the function signature if this is awaiting a direct call
					// For now, we'll need to fail with a more helpful error
					g.errors = append(g.errors, fmt.Sprintf("cannot await Promise: type could not be inferred. Ensure async functions have explicit return types (e.g., async func f():int instead of async func f())"))
					// Default to int as a fallback to allow compilation to continue
					g.output.WriteString(fmt.Sprintf("  // ERROR: Type inference failed, defaulting to int\n"))
					if needsDecl {
						g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_await_int(%s); // INFERRED TYPE\n", varName, promiseVar))
						g.declaredVariables[inst.ID] = true
					} else {
						g.output.WriteString(fmt.Sprintf("  %s = omni_await_int(%s); // INFERRED TYPE\n", varName, promiseVar))
					}
					resultType = "int"
				} else {
					// For user-defined types, we cannot await them yet
					// Fail loudly instead of silently defaulting to string
					g.errors = append(g.errors, fmt.Sprintf("cannot await Promise<%s>: user-defined types are not supported in await expressions", resultType))
					// Still emit code to prevent compilation errors, but it will be wrong
					g.output.WriteString(fmt.Sprintf("  // ERROR: Cannot await user-defined type %s, defaulting to int (WRONG)\n", resultType))
					if needsDecl {
						g.output.WriteString(fmt.Sprintf("  int32_t %s = omni_await_int(%s); // WRONG TYPE\n", varName, promiseVar))
						g.declaredVariables[inst.ID] = true
					} else {
						g.output.WriteString(fmt.Sprintf("  %s = omni_await_int(%s); // WRONG TYPE\n", varName, promiseVar))
					}
					resultType = "int" // Store as int to prevent further type errors
				}
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
// funcName is the C function name (e.g., "omni_main")
// originalReturnType is the original MIR return type (e.g., "Promise<int>" or "int")
func (g *CGenerator) generateTerminator(term *mir.Terminator, funcName string, originalReturnType string) error {
	switch term.Op {
	case "ret":
		// Handle return statement
		if len(term.Operands) > 0 {
			// Track the returned value ID to exclude it from cleanup
			if term.Operands[0].Kind == mir.OperandValue {
				g.returnedValueID = term.Operands[0].Value
				g.returnedValueIDs[term.Operands[0].Value] = true
			}
			value := g.getOperandValue(term.Operands[0])
			if originalReturnType == "string" || originalReturnType == "const char*" || originalReturnType == "char*" {
				if term.Operands[0].Kind == mir.OperandValue {
					delete(g.stringsToFree, term.Operands[0].Value)
				}
			}
			g.emitReturnThroughEpilogue(value, funcName, originalReturnType)
		} else {
			g.emitReturnDefaultThroughEpilogue(funcName, originalReturnType)
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
// getOperandLengthExpr returns a C expression for the length of an
// array-valued operand: a literal int when we tracked it at compile
// time, or a variable name when the array is a function parameter.
// Falls back to "-1" when the length is unknown — call sites should
// avoid emitting calls into runtime intrinsics in that case.
func (g *CGenerator) getOperandLengthExpr(op mir.Operand) string {
	if op.Kind != mir.OperandValue {
		return "-1"
	}
	if length, ok := g.arrayLengths[op.Value]; ok {
		return fmt.Sprintf("%d", length)
	}
	if expr, ok := g.arrayLengthExprs[op.Value]; ok {
		return expr
	}
	return "-1"
}

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
		// First check for string types
		if operandType == "string" || operandType == "const char*" || operandType == "char*" {
			// Already a string, return as is
			return varName
		}

		// Check for struct/pointer types (must check before primitive types)
		isStructType := strings.Contains(operandType, "struct") ||
			strings.Contains(operandType, "omni_struct") ||
			strings.Contains(operandType, "omni_server") ||
			strings.Contains(operandType, "omni_map") ||
			strings.Contains(operandType, "omni_array") ||
			(strings.HasSuffix(operandType, "*") && !g.isPrimitiveType(operandType) && !strings.Contains(operandType, "<") && !strings.Contains(operandType, "(")) ||
			(!g.isPrimitiveType(operandType) && !strings.Contains(operandType, "<") && !strings.Contains(operandType, "(") && operandType != "" && !strings.Contains(operandType, "int") && !strings.Contains(operandType, "float") && !strings.Contains(operandType, "double") && !strings.Contains(operandType, "bool"))

		if isStructType {
			// Struct or pointer type - convert pointer to string representation
			// For now, use a placeholder string since we can't easily convert structs to strings
			tempVar := fmt.Sprintf("temp_str_%d_%d", op.Value, g.output.Len())
			g.output.WriteString(fmt.Sprintf("  const char* %s = \"<struct>\";\n", tempVar))
			// Track this temporary string for cleanup (though it's a literal, so no cleanup needed)
			return tempVar
		}

		// Check for primitive numeric types
		if operandType == "int" || operandType == "int32_t" || operandType == "int64_t" {
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
		} else if operandType == "" || operandType == inferTypePlaceholder {
			// Unknown type - use placeholder for safety
			tempVar := fmt.Sprintf("temp_str_%d_%d", op.Value, g.output.Len())
			g.output.WriteString(fmt.Sprintf("  const char* %s = \"<unknown>\";\n", tempVar))
			return tempVar
		} else {
			// Unknown type - use placeholder instead of assuming int
			tempVar := fmt.Sprintf("temp_str_%d_%d", op.Value, g.output.Len())
			g.output.WriteString(fmt.Sprintf("  const char* %s = \"<unknown>\";\n", tempVar))
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
	g.emitPrintTo(op, newline, false)
}

// emitPrintTo is the stderr-aware version of emitPrint.
func (g *CGenerator) emitPrintTo(op mir.Operand, newline, stderr bool) {
	var funcName string
	switch {
	case stderr && newline:
		funcName = "omni_eprintln_string"
	case stderr:
		funcName = "omni_eprint_string"
	case newline:
		funcName = "omni_println_string"
	default:
		funcName = "omni_print_string"
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
	// Normalize the type string (trim whitespace)
	omniType = strings.TrimSpace(omniType)

	// Handle <infer> type - default to int
	// Also handle "infer" (without brackets) which can come from array element extraction
	// Check for various forms of infer type
	if omniType == "<infer>" || omniType == inferTypePlaceholder || omniType == "infer" || omniType == "" ||
		omniType == "<inferred>" || strings.HasPrefix(omniType, "<infer") {
		return "int32_t"
	}

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
		// Handle <infer> element type
		if elementType == "<infer>" || elementType == inferTypePlaceholder || elementType == "infer" {
			elementType = "int"
		}
		// Arrays in C are represented as pointers to the element type
		return g.mapType(elementType) + "*"
	}
	// Handle old array syntax: array<ElementType>
	if strings.HasPrefix(omniType, "array<") && strings.HasSuffix(omniType, ">") {
		elementType := omniType[6 : len(omniType)-1]
		// Handle <infer> element type
		if elementType == "<infer>" || elementType == inferTypePlaceholder || elementType == "infer" {
			elementType = "int"
		}
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

	// std.collections compound types — each backed by a typedef'd
	// runtime struct in omni_rt.h. Element types don't change the C
	// carrier today (the runtime holds int32_t elements internally).
	if strings.HasPrefix(omniType, "queue<") && strings.HasSuffix(omniType, ">") {
		return "omni_queue_t*"
	}
	if strings.HasPrefix(omniType, "stack<") && strings.HasSuffix(omniType, ">") {
		return "omni_stack_t*"
	}
	if strings.HasPrefix(omniType, "set<") && strings.HasSuffix(omniType, ">") {
		return "omni_set_t*"
	}
	if strings.HasPrefix(omniType, "linked_list<") && strings.HasSuffix(omniType, ">") {
		return "omni_linked_list_t*"
	}
	if strings.HasPrefix(omniType, "binary_tree<") && strings.HasSuffix(omniType, ">") {
		return "omni_binary_tree_t*"
	}
	if strings.HasPrefix(omniType, "priority_queue<") && strings.HasSuffix(omniType, ">") {
		return "omni_priority_queue_t*"
	}

	// Handle channel types: chan<T> — represented at runtime by an
	// omni_chan_t* (a pthread-backed bounded ring buffer). Element types
	// don't widen the C type because the channel stores raw bytes.
	if strings.HasPrefix(omniType, "chan<") && strings.HasSuffix(omniType, ">") {
		return "omni_chan_t*"
	}

	// Handle tuple types: tuple<T1,T2,...> — represented as a
	// program-unique struct emitted up-front. Multi-return functions use
	// this as their C return type so call sites can read the components
	// via field access.
	if strings.HasPrefix(omniType, "tuple<") && strings.HasSuffix(omniType, ">") {
		g.ensureTupleStruct(omniType)
		return g.tupleStructName(omniType)
	}

	// Handle struct types: struct<Field1Type,Field2Type,...>
	if strings.HasPrefix(omniType, "struct<") && strings.HasSuffix(omniType, ">") {
		return "omni_struct_t*"
	}

	// Handle special web framework types
	if omniType == "std.web.Server" || omniType == "Server" {
		return "omni_server_t*"
	}
	if omniType == "std.web.Handler" || omniType == "Handler" {
		return "void*" // Function pointer
	}
	if omniType == "std.web.Middleware" || omniType == "Middleware" {
		return "void*" // Function pointer
	}
	if omniType == "any" {
		return "void*"
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
	case "char":
		// Chars hold Unicode code points; the backend stores them as
		// int32_t so std.char_code / std.char_from_code are identity
		// casts.
		return "int32_t"
	case "ptr":
		return "void*"
	default:
		// Check again for infer types that might have been missed
		if omniType == "<infer>" || omniType == inferTypePlaceholder || omniType == "infer" || omniType == "" ||
			omniType == "<inferred>" || strings.HasPrefix(omniType, "<infer") {
			return "int32_t"
		}
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

	// std.collections — basic map operations on map<string, int> /
	// map<int, int>. Keys/values/copy/merge are generic wrappers whose
	// runtime intrinsics need additional (buffer, size) arguments;
	// they remain on the loaded-body path.
	case "std.collections.size":
		return "omni_map_size"
	case "std.collections.get":
		return "omni_map_get_string_int"
	case "std.collections.set":
		return "omni_map_put_string_int"
	case "std.collections.has":
		return "omni_map_has_string"
	case "std.collections.remove":
		return "omni_map_remove_string"
	case "std.collections.clear":
		return "omni_map_clear"
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

	// Web framework functions
	case "std.web.server_create":
		return "omni_server_create"
	case "std.web.server_listen":
		return "omni_server_listen"
	case "std.web.server_listen_tls":
		return "omni_server_listen_tls"
	case "std.web.server_close":
		return "omni_server_close"
	case "std.web.server_graceful_shutdown":
		return "omni_server_graceful_shutdown"
	case "std.web.context_param":
		return "omni_context_param"
	case "std.web.context_query":
		return "omni_context_query"
	case "std.web.context_query_all":
		return "omni_context_query_all"
	case "std.web.context_header":
		return "omni_context_header"
	case "std.web.context_set_header":
		return "omni_context_set_header"
	case "std.web.context_status":
		return "omni_context_status"
	case "std.web.context_html":
		return "omni_context_html"
	case "std.web.context_redirect":
		return "omni_context_redirect"
	case "std.web.context_cookie":
		return "omni_context_cookie"
	case "std.web.context_get_cookie":
		return "omni_context_get_cookie"
	case "std.web.context_body":
		return "omni_context_body"
	case "std.web.context_set_state":
		return "omni_context_set_state"
	case "std.web.context_get_state":
		return "omni_context_get_state"
	case "std.web.context_file":
		return "omni_context_file"
	case "std.web.context_body_json":
		return "omni_context_body_json"
	case "std.web.context_text":
		return "omni_context_text"
	case "std.web.context_json":
		return "omni_context_json"
	case "std.web.server_get":
		return "omni_server_get"
	case "std.web.server_post":
		return "omni_server_post"
	case "std.web.server_put":
		return "omni_server_put"
	case "std.web.server_delete":
		return "omni_server_delete"
	case "std.web.context_body_form":
		return "omni_context_body_form"
	case "std.web.context_files":
		return "omni_context_files"
	case "std.web.server_websocket":
		return "omni_server_websocket"
	case "std.web.websocket_send":
		return "omni_websocket_send"
	case "std.web.websocket_receive":
		return "omni_websocket_receive"
	case "std.web.websocket_close":
		return "omni_websocket_close"
	case "std.web.validate_string":
		return "omni_validate_string"
	case "std.web.validate_int":
		return "omni_validate_int"
	case "std.web.validate_email":
		return "omni_validate_email"
	case "std.web.validate_url":
		return "omni_validate_url"
	case "std.web.sanitize_html":
		return "omni_sanitize_html"
	case "std.web.sanitize_sql":
		return "omni_sanitize_sql"
	case "std.web.server_patch":
		return "omni_server_patch"
	case "std.web.server_all":
		return "omni_server_all"
	case "std.web.server_route":
		return "omni_server_route"
	case "std.web.server_group":
		return "omni_server_group"
	case "std.web.server_use":
		return "omni_server_use"
	case "std.web.server_use_before":
		return "omni_server_use_before"
	case "std.web.server_use_after":
		return "omni_server_use_after"
	case "std.web.group_get":
		return "omni_group_get"
	case "std.web.group_post":
		return "omni_group_post"
	case "std.web.group_use":
		return "omni_group_use"
	case "std.web.middleware_logger":
		return "omni_middleware_logger"
	case "std.web.middleware_cors":
		return "omni_middleware_cors"
	case "std.web.middleware_json_parser":
		return "omni_middleware_json_parser"
	case "std.web.middleware_form_parser":
		return "omni_middleware_form_parser"
	case "std.web.middleware_multipart_parser_impl":
		return "omni_middleware_multipart_parser_impl"
	case "std.web.middleware_multipart_parser":
		return "omni_middleware_multipart_parser"
	case "std.web.middleware_static_impl":
		return "omni_middleware_static_impl"
	case "std.web.middleware_static":
		return "omni_middleware_static"
	case "std.web.template_render":
		return "omni_template_render"
	case "std.web.template_load":
		return "omni_template_load"
	case "std.web.template_cache_enable":
		return "omni_template_cache_enable"
	case "std.web.test_client_create":
		return "omni_test_client_create"
	case "std.web.test_client_get":
		return "omni_test_client_get"
	case "std.web.test_client_post":
		return "omni_test_client_post"
	case "std.web.test_response_status":
		return "omni_test_response_status"
	case "std.web.test_response_body":
		return "omni_test_response_body"
	case "std.web.test_response_headers":
		return "omni_test_response_headers"
	case "std.web.test_response_json":
		return "omni_test_response_json"
	case "std.web.omni_http_parse_request":
		return "omni_http_parse_request"
	case "std.web.omni_http_build_response":
		return "omni_http_build_response"
	case "std.web.omni_http_parse_query":
		return "omni_http_parse_query"
	case "std.web.omni_http_match_path":
		return "omni_http_match_path"
	case "std.web.omni_json_parse":
		return "omni_json_parse"
	case "std.web.omni_json_stringify":
		return "omni_json_stringify"
	case "std.web.omni_http_parse_form_urlencoded":
		return "omni_http_parse_form_urlencoded"
	case "std.web.omni_http_parse_multipart":
		return "omni_http_parse_multipart"
	case "std.web.omni_file_upload_save":
		return "omni_file_upload_save"
	case "std.web.omni_file_upload_validate":
		return "omni_file_upload_validate"
	case "std.web.omni_file_read_binary":
		return "omni_file_read_binary"
	case "std.web.omni_file_get_mime_type":
		return "omni_file_get_mime_type"
	case "std.web.omni_file_get_size":
		return "omni_file_get_size"
	case "std.web.omni_http_compress_gzip":
		return "omni_http_compress_gzip"
	case "std.web.omni_http_decompress_gzip":
		return "omni_http_decompress_gzip"
	case "std.web.omni_websocket_handshake":
		return "omni_websocket_handshake"
	case "std.web.omni_websocket_frame_create":
		return "omni_websocket_frame_create"
	case "std.web.omni_websocket_frame_parse":
		return "omni_websocket_frame_parse"
	case "std.web.omni_server_connection_pool_create":
		return "omni_server_connection_pool_create"
	case "std.web.omni_server_connection_pool_acquire":
		return "omni_server_connection_pool_acquire"
	case "std.web.omni_server_connection_pool_release":
		return "omni_server_connection_pool_release"
	case "std.web.omni_server_thread_pool_create":
		return "omni_server_thread_pool_create"
	case "std.web.omni_server_thread_pool_submit":
		return "omni_server_thread_pool_submit"
	case "std.web.omni_server_set_timeout":
		return "omni_server_set_timeout"
	case "std.web.omni_server_set_max_request_size":
		return "omni_server_set_max_request_size"
	case "std.web.omni_server_set_max_headers_size":
		return "omni_server_set_max_headers_size"
	case "std.web.omni_server_create":
		return "omni_server_create"

	// IO functions
	case "std.io.print":
		return "omni_print_string"
	case "std.io.println":
		return "omni_println_string"
	case "std.io.eprint":
		return "omni_eprint_string"
	case "std.io.eprintln":
		return "omni_eprintln_string"
	case "std.io.flush":
		return "omni_io_flush"
	case "std.io.read_line":
		return "omni_read_line"
	case "std.io.read_all":
		return "omni_read_all"

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
	case "std.string.trim_left":
		return "omni_trim_left"
	case "std.string.trim_right":
		return "omni_trim_right"
	case "std.string.trim_all":
		return "omni_trim_all"
	case "std.string.to_upper":
		return "omni_to_upper"
	case "std.string.to_lower":
		return "omni_to_lower"
	case "std.string.to_title":
		return "omni_to_title"
	case "std.string.capitalize":
		return "omni_capitalize"
	case "std.string.reverse":
		return "omni_string_reverse"
	case "std.string.equals":
		return "omni_string_equals"
	case "std.string.compare":
		return "omni_string_compare"
	case "std.string.equals_ignore_case":
		return "omni_string_equals_ignore_case"
	case "std.string.compare_ignore_case":
		return "omni_string_compare_ignore_case"
	case "std.string.count_occurrences":
		return "omni_count_occurrences"
	case "std.string.count_lines":
		return "omni_count_lines"
	case "std.string.count_words":
		return "omni_count_words"
	case "std.string.is_empty":
		return "omni_string_is_empty"
	case "std.string.join":
		return "omni_string_join"
	case "std.string.replace", "std.string.replace_all":
		return "omni_string_replace_all"
	case "std.string.replace_first":
		return "omni_string_replace_first"
	case "std.string.replace_last":
		return "omni_string_replace_last"
	case "std.algorithms.euclidean_distance":
		return "omni_euclidean_distance"
	case "std.algorithms.manhattan_distance":
		return "omni_manhattan_distance"
	case "std.algorithms.levenshtein_distance":
		return "omni_levenshtein_distance"
	case "std.algorithms.bubble_sort":
		return "omni_bubble_sort"
	case "std.algorithms.selection_sort":
		return "omni_selection_sort"
	case "std.algorithms.insertion_sort":
		return "omni_insertion_sort"
	case "std.algorithms.linear_search":
		return "omni_linear_search"
	case "std.algorithms.binary_search":
		return "omni_binary_search"
	case "std.algorithms.find_max":
		return "omni_array_find_max"
	case "std.algorithms.find_min":
		return "omni_array_find_min"
	case "std.algorithms.count_occurrences":
		return "omni_array_count_occurrences"
	case "std.algorithms.reverse":
		return "omni_array_reverse"
	case "std.algorithms.rotate":
		return "omni_array_rotate"
	case "std.math.random_seed":
		return "omni_random_seed"
	case "std.math.random_int":
		return "omni_random_int"

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
	case "std.os.write_file":
		return "omni_write_file"
	case "std.os.append_file":
		return "omni_append_file"
	case "std.os.read_file":
		return "omni_read_file"
	case "std.os.delete_file":
		return "omni_delete_file"
	case "std.os.create_dir":
		return "omni_create_dir"
	case "std.os.remove_dir":
		return "omni_remove_dir"
	case "std.os.file_exists":
		return "omni_file_exists"
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
	case "std.time.time_from_unix":
		return "omni_time_from_unix"
	case "std.time.time_to_unix":
		return "omni_time_to_unix"
	case "std.time.time_to_string":
		return "omni_time_to_string"
	case "std.time.time_from_string":
		return "omni_time_from_string"
	case "std.time.time_to_unix_nano":
		return "omni_time_to_unix_nano"
	case "std.time.duration_to_string":
		return "omni_duration_to_string"
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
	case "time.time_from_unix":
		return "omni_time_from_unix"
	case "time.time_to_unix":
		return "omni_time_to_unix"
	case "time.time_to_string":
		return "omni_time_to_string"
	case "time.time_from_string":
		return "omni_time_from_string"
	case "time.time_to_unix_nano":
		return "omni_time_to_unix_nano"
	case "time.duration_to_string":
		return "omni_duration_to_string"

	// Utility functions
	case "std.assert":
		return "omni_assert"
	case "std.panic":
		return "omni_panic"
	case "std.int_to_string", "std.string.int_to_string":
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
	case "std.char_code":
		return "omni_char_code"
	case "std.char_from_code":
		return "omni_char_from_code"
	case "std.char_to_string":
		return "omni_char_to_string"
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
		"std.io.print":     "omni_print_string",
		"std.io.println":   "omni_println_string",
		"io.print":         "omni_print_string",
		"io.println":       "omni_println_string",
		"std.io.eprint":    "omni_eprint_string",
		"std.io.eprintln":  "omni_eprintln_string",
		"io.eprint":        "omni_eprint_string",
		"io.eprintln":      "omni_eprintln_string",
		"std.io.flush":     "omni_io_flush",
		"io.flush":         "omni_io_flush",
		"std.io.read_line": "omni_read_line",
		"io.read_line":     "omni_read_line",
		"std.io.read_all":  "omni_read_all",
		"io.read_all":      "omni_read_all",

		// String functions
		"std.string.length":                   "omni_strlen",
		"std.string.concat":                   "omni_strcat",
		"std.string.substring":                "omni_substring",
		"std.string.char_at":                  "omni_char_at",
		"std.string.starts_with":              "omni_starts_with",
		"std.string.ends_with":                "omni_ends_with",
		"std.string.contains":                 "omni_contains",
		"std.string.index_of":                 "omni_index_of",
		"std.string.last_index_of":            "omni_last_index_of",
		"std.string.trim":                     "omni_trim",
		"std.string.trim_left":                "omni_trim_left",
		"std.string.trim_right":               "omni_trim_right",
		"std.string.trim_all":                 "omni_trim_all",
		"std.string.to_upper":                 "omni_to_upper",
		"std.string.to_lower":                 "omni_to_lower",
		"std.string.to_title":                 "omni_to_title",
		"std.string.capitalize":               "omni_capitalize",
		"std.string.reverse":                  "omni_string_reverse",
		"std.string.equals":                   "omni_string_equals",
		"std.string.compare":                  "omni_string_compare",
		"std.string.equals_ignore_case":       "omni_string_equals_ignore_case",
		"std.string.compare_ignore_case":      "omni_string_compare_ignore_case",
		"std.string.count_occurrences":        "omni_count_occurrences",
		"std.string.count_lines":              "omni_count_lines",
		"std.string.count_words":              "omni_count_words",
		"std.string.is_empty":                 "omni_string_is_empty",
		"std.string.join":                     "omni_string_join",
		"std.string.replace":                  "omni_string_replace_all",
		"std.string.replace_all":              "omni_string_replace_all",
		"std.string.replace_first":            "omni_string_replace_first",
		"std.string.replace_last":             "omni_string_replace_last",
		"std.algorithms.euclidean_distance":   "omni_euclidean_distance",
		"std.algorithms.manhattan_distance":   "omni_manhattan_distance",
		"std.algorithms.levenshtein_distance": "omni_levenshtein_distance",
		"std.algorithms.bubble_sort":          "omni_bubble_sort",
		"std.algorithms.selection_sort":       "omni_selection_sort",
		"std.algorithms.insertion_sort":       "omni_insertion_sort",
		"std.algorithms.linear_search":        "omni_linear_search",
		"std.algorithms.binary_search":        "omni_binary_search",
		"std.algorithms.find_max":             "omni_array_find_max",
		"std.algorithms.find_min":             "omni_array_find_min",
		"std.algorithms.count_occurrences":    "omni_array_count_occurrences",
		"std.algorithms.reverse":              "omni_array_reverse",
		"std.algorithms.rotate":               "omni_array_rotate",
		"std.math.random_seed":                "omni_random_seed",
		"std.math.random_int":                 "omni_random_int",
		"string.length":                       "omni_strlen",
		"string.concat":                       "omni_strcat",
		"string.substring":                    "omni_substring",
		"string.char_at":                      "omni_char_at",
		"string.starts_with":                  "omni_starts_with",
		"string.ends_with":                    "omni_ends_with",
		"string.contains":                     "omni_contains",
		"string.index_of":                     "omni_index_of",
		"string.last_index_of":                "omni_last_index_of",
		"string.trim":                         "omni_trim",
		"string.to_upper":                     "omni_to_upper",
		"string.to_lower":                     "omni_to_lower",
		"string.equals":                       "omni_string_equals",
		"string.compare":                      "omni_string_compare",

		// Array functions
		"std.array.length": "omni_array_length",
		"std.array.get":    "omni_array_get_int",
		"std.array.set":    "omni_array_set_int",

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
		"std.int_to_string":        "omni_int_to_string",
		"std.string.int_to_string": "omni_int_to_string",
		"std.float_to_string":      "omni_float_to_string",
		"std.bool_to_string":       "omni_bool_to_string",
		"std.string_to_int":        "omni_string_to_int",
		"std.string_to_float":      "omni_string_to_float",
		"std.string_to_bool":       "omni_string_to_bool",
		"std.char_code":            "omni_char_code",
		"std.char_from_code":       "omni_char_from_code",
		"std.char_to_string":       "omni_char_to_string",

		// Logging functions
		"std.log.debug":     "omni_log_debug",
		"std.log.info":      "omni_log_info",
		"std.log.warn":      "omni_log_warn",
		"std.log.error":     "omni_log_error",
		"std.log.set_level": "omni_log_set_level",

		// File operations
		"std.file.open":      "omni_file_open",
		"std.file.close":     "omni_file_close",
		"std.file.read":      "omni_file_read",
		"std.file.write":     "omni_file_write",
		"std.file.seek":      "omni_file_seek",
		"std.file.tell":      "omni_file_tell",
		"std.file.exists":    "omni_file_exists",
		"std.file.size":      "omni_file_size",
		"file.open":          "omni_file_open",
		"file.close":         "omni_file_close",
		"file.read":          "omni_file_read",
		"file.write":         "omni_file_write",
		"file.seek":          "omni_file_seek",
		"file.tell":          "omni_file_tell",
		"file.exists":        "omni_file_exists",
		"file.size":          "omni_file_size",
		"std.os.read_file":   "omni_read_file",
		"std.os.write_file":  "omni_write_file",
		"std.os.append_file": "omni_append_file",
		"os.read_file":       "omni_read_file",
		"os.write_file":      "omni_write_file",
		"os.append_file":     "omni_append_file",

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
		"std.string.matches":          "omni_string_matches",
		"std.string.find_match":       "omni_string_find_match",
		"std.string.find_all_matches": "omni_string_find_all_matches",
		"std.string.replace_regex":    "omni_string_replace_regex",
		"string.matches":              "omni_string_matches",
		"string.find_match":           "omni_string_find_match",
		"string.find_all_matches":     "omni_string_find_all_matches",
		"string.replace_regex":        "omni_string_replace_regex",

		// Time functions
		"std.time.now":                "omni_time_now_unix",
		"std.time.unix_timestamp":     "omni_time_now_unix",
		"std.time.unix_nano":          "omni_time_now_unix_nano",
		"std.time.sleep_seconds":      "omni_time_sleep_seconds",
		"std.time.sleep_milliseconds": "omni_time_sleep_milliseconds",
		"std.time.time_zone_offset":   "omni_time_zone_offset",
		"std.time.time_zone_name":     "omni_time_zone_name",
		"std.time.time_from_unix":     "omni_time_from_unix",
		"std.time.time_to_unix":       "omni_time_to_unix",
		"std.time.time_to_string":     "omni_time_to_string",
		"std.time.time_from_string":   "omni_time_from_string",
		"std.time.time_to_unix_nano":  "omni_time_to_unix_nano",
		"std.time.duration_to_string": "omni_duration_to_string",
		"time.now":                    "omni_time_now_unix",
		"time.unix_timestamp":         "omni_time_now_unix",
		"time.unix_nano":              "omni_time_now_unix_nano",
		"time.sleep_seconds":          "omni_time_sleep_seconds",
		"time.sleep_milliseconds":     "omni_time_sleep_milliseconds",
		"time.time_zone_offset":       "omni_time_zone_offset",
		"time.time_zone_name":         "omni_time_zone_name",
		"time.time_from_unix":         "omni_time_from_unix",
		"time.time_to_unix":           "omni_time_to_unix",
		"time.time_to_string":         "omni_time_to_string",
		"time.time_from_string":       "omni_time_from_string",
		"time.time_to_unix_nano":      "omni_time_to_unix_nano",
		"time.duration_to_string":     "omni_duration_to_string",

		// Command-line argument functions
		"std.os.args":           "omni_args_get",
		"std.os.args_count":     "omni_args_count",
		"std.os.has_flag":       "omni_args_has_flag",
		"std.os.get_flag":       "omni_args_get_flag",
		"std.os.positional_arg": "omni_args_positional",
		"os.args":               "omni_args_get",
		"os.args_count":         "omni_args_count",
		"os.has_flag":           "omni_args_has_flag",
		"os.get_flag":           "omni_args_get_flag",
		"os.positional_arg":     "omni_args_positional",

		// Process ID functions
		"std.os.getpid":  "omni_getpid",
		"std.os.getppid": "omni_getppid",
		"os.getpid":      "omni_getpid",
		"os.getppid":     "omni_getppid",

		// Collections functions
		"std.collections.size":   "omni_map_size",
		"std.collections.get":    "omni_map_get_string_int",
		"std.collections.set":    "omni_map_put_string_int",
		"std.collections.has":    "omni_map_has_string",
		"std.collections.remove": "omni_map_remove_string",
		"std.collections.clear":  "omni_map_clear",
		"std.collections.copy":   "omni_map_copy_string_int",
		"std.collections.merge":  "omni_map_merge_string_int",
		// Set functions
		"std.collections.set_create":       "omni_set_create",
		"std.collections.set_add":          "omni_set_add",
		"std.collections.set_remove":       "omni_set_remove",
		"std.collections.set_contains":     "omni_set_contains",
		"std.collections.set_size":         "omni_set_size",
		"std.collections.set_clear":        "omni_set_clear",
		"std.collections.set_union":        "omni_set_union",
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
		"std.collections.stack_push":     "omni_stack_push",
		"std.collections.stack_pop":      "omni_stack_pop",
		"std.collections.stack_peek":     "omni_stack_peek",
		"std.collections.stack_is_empty": "omni_stack_is_empty",
		"std.collections.stack_size":     "omni_stack_size",
		"std.collections.stack_clear":    "omni_stack_clear",
		// Priority queue functions
		"std.collections.priority_queue_create":      "omni_priority_queue_create",
		"std.collections.priority_queue_insert":      "omni_priority_queue_insert",
		"std.collections.priority_queue_extract_max": "omni_priority_queue_extract_max",
		"std.collections.priority_queue_peek":        "omni_priority_queue_peek",
		"std.collections.priority_queue_is_empty":    "omni_priority_queue_is_empty",
		"std.collections.priority_queue_size":        "omni_priority_queue_size",
		// Linked list functions
		"std.collections.linked_list_create":   "omni_linked_list_create",
		"std.collections.linked_list_append":   "omni_linked_list_append",
		"std.collections.linked_list_prepend":  "omni_linked_list_prepend",
		"std.collections.linked_list_insert":   "omni_linked_list_insert",
		"std.collections.linked_list_remove":   "omni_linked_list_remove",
		"std.collections.linked_list_get":      "omni_linked_list_get",
		"std.collections.linked_list_set":      "omni_linked_list_set",
		"std.collections.linked_list_size":     "omni_linked_list_size",
		"std.collections.linked_list_is_empty": "omni_linked_list_is_empty",
		"std.collections.linked_list_clear":    "omni_linked_list_clear",
		// Binary tree functions
		"std.collections.binary_tree_create":   "omni_binary_tree_create",
		"std.collections.binary_tree_insert":   "omni_binary_tree_insert",
		"std.collections.binary_tree_search":   "omni_binary_tree_search",
		"std.collections.binary_tree_remove":   "omni_binary_tree_remove",
		"std.collections.binary_tree_size":     "omni_binary_tree_size",
		"std.collections.binary_tree_is_empty": "omni_binary_tree_is_empty",
		"std.collections.binary_tree_clear":    "omni_binary_tree_clear",
		// Network functions
		"std.network.ip_parse":             "omni_ip_parse",
		"std.network.ip_is_valid":          "omni_ip_is_valid",
		"std.network.ip_is_private":        "omni_ip_is_private",
		"std.network.ip_is_loopback":       "omni_ip_is_loopback",
		"std.network.ip_to_string":         "omni_ip_to_string",
		"std.network.url_parse":            "omni_url_parse",
		"std.network.url_to_string":        "omni_url_to_string",
		"std.network.url_is_valid":         "omni_url_is_valid",
		"std.network.dns_lookup":           "omni_dns_lookup",
		"std.network.dns_reverse_lookup":   "omni_dns_reverse_lookup",
		"std.network.http_get":             "omni_http_get",
		"std.network.http_post":            "omni_http_post",
		"std.network.http_put":             "omni_http_put",
		"std.network.http_delete":          "omni_http_delete",
		"std.network.http_request":         "omni_http_request",
		"std.network.socket_create":        "omni_socket_create",
		"std.network.socket_connect":       "omni_socket_connect",
		"std.network.socket_bind":          "omni_socket_bind",
		"std.network.socket_listen":        "omni_socket_listen",
		"std.network.socket_accept":        "omni_socket_accept",
		"std.network.socket_send":          "omni_socket_send",
		"std.network.socket_receive":       "omni_socket_receive",
		"std.network.socket_close":         "omni_socket_close",
		"std.network.network_is_connected": "omni_network_is_connected",
		"std.network.network_get_local_ip": "omni_network_get_local_ip",
		"std.network.network_ping":         "omni_network_ping",

		// Web framework functions
		"std.web.server_create":                       "omni_server_create",
		"std.web.server_listen":                       "omni_server_listen",
		"std.web.server_listen_tls":                   "omni_server_listen_tls",
		"std.web.server_close":                        "omni_server_close",
		"std.web.server_graceful_shutdown":            "omni_server_graceful_shutdown",
		"std.web.server_get":                          "omni_server_get",
		"std.web.server_post":                         "omni_server_post",
		"std.web.server_put":                          "omni_server_put",
		"std.web.server_delete":                       "omni_server_delete",
		"std.web.server_patch":                        "omni_server_patch",
		"std.web.server_all":                          "omni_server_all",
		"std.web.server_route":                        "omni_server_route",
		"std.web.server_group":                        "omni_server_group",
		"std.web.server_use":                          "omni_server_use",
		"std.web.server_use_before":                   "omni_server_use_before",
		"std.web.server_use_after":                    "omni_server_use_after",
		"std.web.group_get":                           "omni_group_get",
		"std.web.group_post":                          "omni_group_post",
		"std.web.group_use":                           "omni_group_use",
		"std.web.context_param":                       "omni_context_param",
		"std.web.context_query":                       "omni_context_query",
		"std.web.context_query_all":                   "omni_context_query_all",
		"std.web.context_header":                      "omni_context_header",
		"std.web.context_set_header":                  "omni_context_set_header",
		"std.web.context_status":                      "omni_context_status",
		"std.web.context_html":                        "omni_context_html",
		"std.web.context_redirect":                    "omni_context_redirect",
		"std.web.context_cookie":                      "omni_context_cookie",
		"std.web.context_get_cookie":                  "omni_context_get_cookie",
		"std.web.context_body":                        "omni_context_body",
		"std.web.context_set_state":                   "omni_context_set_state",
		"std.web.context_get_state":                   "omni_context_get_state",
		"std.web.context_file":                        "omni_context_file",
		"std.web.context_body_json":                   "omni_context_body_json",
		"std.web.context_text":                        "omni_context_text",
		"std.web.context_json":                        "omni_context_json",
		"std.web.context_body_form":                   "omni_context_body_form",
		"std.web.context_files":                       "omni_context_files",
		"std.web.middleware_logger":                   "omni_middleware_logger",
		"std.web.middleware_cors":                     "omni_middleware_cors",
		"std.web.middleware_json_parser":              "omni_middleware_json_parser",
		"std.web.middleware_form_parser":              "omni_middleware_form_parser",
		"std.web.middleware_multipart_parser_impl":    "omni_middleware_multipart_parser_impl",
		"std.web.middleware_multipart_parser":         "omni_middleware_multipart_parser",
		"std.web.middleware_static_impl":              "omni_middleware_static_impl",
		"std.web.middleware_static":                   "omni_middleware_static",
		"std.web.template_render":                     "omni_template_render",
		"std.web.template_load":                       "omni_template_load",
		"std.web.template_cache_enable":               "omni_template_cache_enable",
		"std.web.validate_request":                    "omni_validate_request",
		"std.web.test_client_create":                  "omni_test_client_create",
		"std.web.test_client_get":                     "omni_test_client_get",
		"std.web.test_client_post":                    "omni_test_client_post",
		"std.web.test_response_status":                "omni_test_response_status",
		"std.web.test_response_body":                  "omni_test_response_body",
		"std.web.test_response_headers":               "omni_test_response_headers",
		"std.web.test_response_json":                  "omni_test_response_json",
		"std.web.server_websocket":                    "omni_server_websocket",
		"std.web.websocket_send":                      "omni_websocket_send",
		"std.web.websocket_receive":                   "omni_websocket_receive",
		"std.web.websocket_close":                     "omni_websocket_close",
		"std.web.validate_string":                     "omni_validate_string",
		"std.web.validate_int":                        "omni_validate_int",
		"std.web.validate_email":                      "omni_validate_email",
		"std.web.validate_url":                        "omni_validate_url",
		"std.web.sanitize_html":                       "omni_sanitize_html",
		"std.web.sanitize_sql":                        "omni_sanitize_sql",
		"std.web.omni_http_parse_request":             "omni_http_parse_request",
		"std.web.omni_http_build_response":            "omni_http_build_response",
		"std.web.omni_http_parse_query":               "omni_http_parse_query",
		"std.web.omni_http_match_path":                "omni_http_match_path",
		"std.web.omni_json_parse":                     "omni_json_parse",
		"std.web.omni_json_stringify":                 "omni_json_stringify",
		"std.web.omni_http_parse_form_urlencoded":     "omni_http_parse_form_urlencoded",
		"std.web.omni_http_parse_multipart":           "omni_http_parse_multipart",
		"std.web.omni_file_upload_save":               "omni_file_upload_save",
		"std.web.omni_file_upload_validate":           "omni_file_upload_validate",
		"std.web.omni_file_read_binary":               "omni_file_read_binary",
		"std.web.omni_file_get_mime_type":             "omni_file_get_mime_type",
		"std.web.omni_file_get_size":                  "omni_file_get_size",
		"std.web.omni_http_compress_gzip":             "omni_http_compress_gzip",
		"std.web.omni_http_decompress_gzip":           "omni_http_decompress_gzip",
		"std.web.omni_websocket_handshake":            "omni_websocket_handshake",
		"std.web.omni_websocket_frame_create":         "omni_websocket_frame_create",
		"std.web.omni_websocket_frame_parse":          "omni_websocket_frame_parse",
		"std.web.omni_server_connection_pool_create":  "omni_server_connection_pool_create",
		"std.web.omni_server_connection_pool_acquire": "omni_server_connection_pool_acquire",
		"std.web.omni_server_connection_pool_release": "omni_server_connection_pool_release",
		"std.web.omni_server_thread_pool_create":      "omni_server_thread_pool_create",
		"std.web.omni_server_thread_pool_submit":      "omni_server_thread_pool_submit",
		"std.web.omni_server_set_timeout":             "omni_server_set_timeout",
		"std.web.omni_server_set_max_request_size":    "omni_server_set_max_request_size",
		"std.web.omni_server_set_max_headers_size":    "omni_server_set_max_headers_size",
		"std.web.omni_server_create":                  "omni_server_create",
		// Memory management functions
		"std.panic":   "omni_panic",
		"std.malloc":  "omni_malloc",
		"std.free":    "omni_free",
		"std.realloc": "omni_realloc",
	}

	if _, exists := runtimeImplMap[funcName]; exists {
		return true
	}
	// The module loader rewrites imports so a std sub-module like std.web
	// may end up with a function name like `web.server_create` instead of
	// `std.web.server_create`. Try the std-prefixed alias too.
	if !strings.HasPrefix(funcName, "std.") {
		if _, exists := runtimeImplMap["std."+funcName]; exists {
			return true
		}
	}
	return false
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
		"std.io.print":                        true,
		"std.io.println":                      true,
		"io.print":                            true,
		"io.println":                          true,
		"std.string.length":                   true,
		"std.string.concat":                   true,
		"std.string.substring":                true,
		"std.string.char_at":                  true,
		"std.string.starts_with":              true,
		"std.string.ends_with":                true,
		"std.string.contains":                 true,
		"std.string.index_of":                 true,
		"std.string.last_index_of":            true,
		"std.string.trim":                     true,
		"std.string.trim_left":                true,
		"std.string.trim_right":               true,
		"std.string.trim_all":                 true,
		"std.string.to_upper":                 true,
		"std.string.to_lower":                 true,
		"std.string.to_title":                 true,
		"std.string.capitalize":               true,
		"std.string.reverse":                  true,
		"std.string.equals":                   true,
		"std.string.compare":                  true,
		"std.string.equals_ignore_case":       true,
		"std.string.compare_ignore_case":      true,
		"std.string.count_occurrences":        true,
		"std.string.count_lines":              true,
		"std.string.count_words":              true,
		"std.string.is_empty":                 true,
		"std.string.join":                     true,
		"std.string.replace":                  true,
		"std.string.replace_all":              true,
		"std.string.replace_first":            true,
		"std.string.replace_last":             true,
		"std.string.split":                    true,
		"std.string.split_lines":              true,
		"std.string.split_words":              true,
		"std.string.find_all":                 true,
		"std.algorithms.euclidean_distance":   true,
		"std.algorithms.manhattan_distance":   true,
		"std.algorithms.levenshtein_distance": true,
		"std.algorithms.bubble_sort":          true,
		"std.algorithms.selection_sort":       true,
		"std.algorithms.insertion_sort":       true,
		"std.algorithms.linear_search":        true,
		"std.algorithms.binary_search":        true,
		"std.algorithms.find_max":             true,
		"std.algorithms.find_min":             true,
		"std.algorithms.count_occurrences":    true,
		"std.algorithms.reverse":              true,
		"std.algorithms.rotate":               true,
		"std.math.random_seed":                true,
		"std.math.random_int":                 true,
		"string.length":                       true,
		"string.concat":                       true,
		"string.substring":                    true,
		"string.char_at":                      true,
		"string.starts_with":                  true,
		"string.ends_with":                    true,
		"string.contains":                     true,
		"string.index_of":                     true,
		"string.last_index_of":                true,
		"string.trim":                         true,
		"string.to_upper":                     true,
		"string.to_lower":                     true,
		"string.equals":                       true,
		"string.compare":                      true,
		"std.math.abs":                        true,
		"std.math.max":                        true,
		"std.math.min":                        true,
		"std.math.pow":                        true,
		"std.math.sqrt":                       true,
		"std.math.floor":                      true,
		"std.math.ceil":                       true,
		"std.math.round":                      true,
		"std.math.gcd":                        true,
		"std.math.lcm":                        true,
		"std.math.factorial":                  true,
		"math.abs":                            true,
		"math.max":                            true,
		"math.min":                            true,
		"math.pow":                            true,
		"math.sqrt":                           true,
		"math.floor":                          true,
		"math.ceil":                           true,
		"math.round":                          true,
		"math.gcd":                            true,
		"math.lcm":                            true,
		"math.factorial":                      true,
		"std.os.exit":                         true,
		"os.exit":                             true,
		"std.os.read_file":                    true,
		"std.os.write_file":                   true,
		"std.os.append_file":                  true,
		"os.read_file":                        true,
		"os.write_file":                       true,
		"os.append_file":                      true,
		// Time operations
		"std.time.now":                true,
		"std.time.unix_timestamp":     true,
		"std.time.unix_nano":          true,
		"std.time.sleep_seconds":      true,
		"std.time.sleep_milliseconds": true,
		"std.time.time_zone_offset":   true,
		"std.time.time_zone_name":     true,
		"std.time.time_from_unix":     true,
		"std.time.time_from_string":   true,
		"std.time.time_to_unix":       true,
		"std.time.time_to_string":     true,
		"std.time.time_to_unix_nano":  true,
		"std.time.duration_to_string": true,
		"time.now":                    true,
		"time.unix_timestamp":         true,
		"time.unix_nano":              true,
		"time.sleep_seconds":          true,
		"time.sleep_milliseconds":     true,
		"time.time_zone_offset":       true,
		"time.time_zone_name":         true,
		"time.time_from_unix":         true,
		"time.time_from_string":       true,
		"time.time_to_unix":           true,
		"time.time_to_string":         true,
		"time.time_to_unix_nano":      true,
		"time.duration_to_string":     true,
		// File operations
		"std.file.open":            true,
		"std.file.close":           true,
		"std.file.read":            true,
		"std.file.write":           true,
		"std.file.seek":            true,
		"std.file.tell":            true,
		"std.file.exists":          true,
		"std.file.size":            true,
		"file.open":                true,
		"file.close":               true,
		"file.read":                true,
		"file.write":               true,
		"file.seek":                true,
		"file.tell":                true,
		"file.exists":              true,
		"file.size":                true,
		"std.int_to_string":        true,
		"std.string.int_to_string": true,
		"std.float_to_string":      true,
		"std.bool_to_string":       true,
		"std.string_to_int":        true,
		"std.string_to_float":      true,
		"std.string_to_bool":       true,
		"std.char_code":            true,
		"std.char_from_code":       true,
		"std.char_to_string":       true,
		"std.log.debug":            true,
		"std.log.info":             true,
		"std.log.warn":             true,
		"std.log.error":            true,
		"std.log.set_level":        true,
		"std.test.start":           true,
		"std.test.end":             true,
		"std.assert":               true,
		"test.start":               true,
		"test.end":                 true,
		// Collections functions
		"std.collections.size":                       true,
		"std.collections.get":                        true,
		"std.collections.set":                        true,
		"std.collections.has":                        true,
		"std.collections.remove":                     true,
		"std.collections.clear":                      true,
		"std.collections.copy":                       true,
		"std.collections.merge":                      true,
		"std.collections.set_create":                 true,
		"std.collections.set_add":                    true,
		"std.collections.set_remove":                 true,
		"std.collections.set_contains":               true,
		"std.collections.set_size":                   true,
		"std.collections.set_clear":                  true,
		"std.collections.set_union":                  true,
		"std.collections.set_intersection":           true,
		"std.collections.set_difference":             true,
		"std.collections.queue_create":               true,
		"std.collections.queue_enqueue":              true,
		"std.collections.queue_dequeue":              true,
		"std.collections.queue_peek":                 true,
		"std.collections.queue_is_empty":             true,
		"std.collections.queue_size":                 true,
		"std.collections.queue_clear":                true,
		"std.collections.stack_create":               true,
		"std.collections.stack_push":                 true,
		"std.collections.stack_pop":                  true,
		"std.collections.stack_peek":                 true,
		"std.collections.stack_is_empty":             true,
		"std.collections.stack_size":                 true,
		"std.collections.stack_clear":                true,
		"std.collections.priority_queue_create":      true,
		"std.collections.priority_queue_insert":      true,
		"std.collections.priority_queue_extract_max": true,
		"std.collections.priority_queue_peek":        true,
		"std.collections.priority_queue_is_empty":    true,
		"std.collections.priority_queue_size":        true,
		"std.collections.linked_list_create":         true,
		"std.collections.linked_list_append":         true,
		"std.collections.linked_list_prepend":        true,
		"std.collections.linked_list_insert":         true,
		"std.collections.linked_list_remove":         true,
		"std.collections.linked_list_get":            true,
		"std.collections.linked_list_set":            true,
		"std.collections.linked_list_size":           true,
		"std.collections.linked_list_is_empty":       true,
		"std.collections.linked_list_clear":          true,
		"std.collections.binary_tree_create":         true,
		"std.collections.binary_tree_insert":         true,
		"std.collections.binary_tree_search":         true,
		"std.collections.binary_tree_remove":         true,
		"std.collections.binary_tree_size":           true,
		"std.collections.binary_tree_is_empty":       true,
		"std.collections.binary_tree_clear":          true,
		// Network functions
		"std.network.ip_parse":             true,
		"std.network.ip_is_valid":          true,
		"std.network.ip_is_private":        true,
		"std.network.ip_is_loopback":       true,
		"std.network.ip_to_string":         true,
		"std.network.url_parse":            true,
		"std.network.url_to_string":        true,
		"std.network.url_is_valid":         true,
		"std.network.dns_lookup":           true,
		"std.network.dns_reverse_lookup":   true,
		"std.network.http_get":             true,
		"std.network.http_post":            true,
		"std.network.http_put":             true,
		"std.network.http_delete":          true,
		"std.network.http_request":         true,
		"std.network.socket_create":        true,
		"std.network.socket_connect":       true,
		"std.network.socket_bind":          true,
		"std.network.socket_listen":        true,
		"std.network.socket_accept":        true,
		"std.network.socket_send":          true,
		"std.network.socket_receive":       true,
		"std.network.socket_close":         true,
		"std.network.network_is_connected": true,
		"std.network.network_get_local_ip": true,
		"std.network.network_ping":         true,
		// Web framework functions
		"std.web.server_create":                    true,
		"std.web.server_listen":                    true,
		"std.web.server_listen_tls":                true,
		"std.web.server_close":                     true,
		"std.web.server_graceful_shutdown":         true,
		"std.web.server_get":                       true,
		"std.web.server_post":                      true,
		"std.web.server_put":                       true,
		"std.web.server_delete":                    true,
		"std.web.server_patch":                     true,
		"std.web.server_all":                       true,
		"std.web.server_route":                     true,
		"std.web.server_group":                     true,
		"std.web.server_use":                       true,
		"std.web.server_use_before":                true,
		"std.web.server_use_after":                 true,
		"std.web.group_get":                        true,
		"std.web.group_post":                       true,
		"std.web.group_use":                        true,
		"std.web.context_param":                    true,
		"std.web.context_query":                    true,
		"std.web.context_query_all":                true,
		"std.web.context_header":                   true,
		"std.web.context_set_header":               true,
		"std.web.context_status":                   true,
		"std.web.context_html":                     true,
		"std.web.context_redirect":                 true,
		"std.web.context_cookie":                   true,
		"std.web.context_get_cookie":               true,
		"std.web.context_body":                     true,
		"std.web.context_set_state":                true,
		"std.web.context_get_state":                true,
		"std.web.context_file":                     true,
		"std.web.context_body_json":                true,
		"std.web.context_text":                     true,
		"std.web.context_json":                     true,
		"std.web.context_body_form":                true,
		"std.web.context_files":                    true,
		"std.web.middleware_logger":                true,
		"std.web.middleware_cors":                  true,
		"std.web.middleware_json_parser":           true,
		"std.web.middleware_form_parser":           true,
		"std.web.middleware_multipart_parser_impl": true,
		"std.web.middleware_multipart_parser":      true,
		"std.web.middleware_static_impl":           true,
		"std.web.middleware_static":                true,
		"std.web.template_render":                  true,
		"std.web.template_load":                    true,
		"std.web.template_cache_enable":            true,
		"std.web.validate_request":                 true,
		"std.web.test_client_create":               true,
		"std.web.test_client_get":                  true,
		"std.web.test_client_post":                 true,
		"std.web.test_response_status":             true,
		"std.web.test_response_body":               true,
		"std.web.test_response_headers":            true,
		"std.web.test_response_json":               true,
		"std.web.server_websocket":                 true,
		"std.web.websocket_send":                   true,
		"std.web.websocket_receive":                true,
		"std.web.websocket_close":                  true,
		"std.web.validate_string":                  true,
		"std.web.validate_int":                     true,
		"std.web.validate_email":                   true,
		"std.web.validate_url":                     true,
		"std.web.sanitize_html":                    true,
		"std.web.sanitize_sql":                     true,
		// Memory management functions
		"std.panic":   true,
		"std.malloc":  true,
		"std.free":    true,
		"std.realloc": true,
	}

	if runtimeFunctions[funcName] {
		return true
	}
	// Allow the same aliasing as hasRuntimeImplementation: if a function is
	// named without the `std.` prefix (as the module loader sometimes
	// produces), also match its `std.`-prefixed form.
	if !strings.HasPrefix(funcName, "std.") && runtimeFunctions["std."+funcName] {
		return true
	}
	return false
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
		"std.io.read_line":         true,
		"io.read_line":             true,
		"std.io.read_all":          true,
		"io.read_all":              true,
		"std.string.concat":        true,
		"std.string.substring":     true,
		"std.string.trim":          true,
		"std.string.trim_left":     true,
		"std.string.trim_right":    true,
		"std.string.trim_all":      true,
		"std.string.to_upper":      true,
		"std.string.to_lower":      true,
		"std.string.to_title":      true,
		"std.string.capitalize":    true,
		"std.string.reverse":       true,
		"std.string.join":          true,
		"std.string.replace":       true,
		"std.string.replace_all":   true,
		"std.string.replace_first": true,
		"std.string.replace_last":  true,
		"std.int_to_string":        true,
		"std.float_to_string":      true,
		"std.bool_to_string":       true,
		"std.os.read_file":         true,
		"os.read_file":             true,
		"std.char_to_string":       true,
		"omni_char_to_string":      true,
		"omni_read_line":           true,
		"omni_strcat":              true,
		"omni_substring":           true,
		"omni_trim":                true,
		"omni_to_upper":            true,
		"omni_to_lower":            true,
		"omni_int_to_string":       true,
		"omni_float_to_string":     true,
		"omni_bool_to_string":      true,
		"omni_read_file":           true,
		"omni_await_string":        true,
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
// isArrayParamType reports whether a function-parameter type spelled in
// MIR (e.g. "array<int>", "[]<string>") needs the implicit length
// companion the codegen attaches. Function-pointer types containing a
// parenthesized arrow are explicitly excluded so we don't mistake
// e.g. "(array<int>) -> int" for an array.
func isArrayParamType(omniType string) bool {
	t := strings.TrimSpace(omniType)
	if t == "" {
		return false
	}
	if strings.Contains(t, ") -> ") {
		return false
	}
	return strings.HasPrefix(t, "array<") || strings.HasPrefix(t, "[]<")
}

// arrayLenParamName returns the C name of the implicit length companion
// emitted alongside an array parameter named `name`.
func arrayLenParamName(name string) string {
	return "__omni_len_" + name
}

// emitStdArrayIntOp lowers a std.array.* call on a known-element-type
// array to its specialized runtime intrinsic. Currently routes int and
// string element types; other types fall through and the caller drops
// back to the legacy placeholder branches.
func (g *CGenerator) emitStdArrayIntOp(inst *mir.Instruction, funcName string) bool {
	if inst.ID == mir.InvalidValue || len(inst.Operands) < 2 {
		return false
	}
	arrOp := inst.Operands[1]
	if arrOp.Kind != mir.OperandValue {
		return false
	}
	elemType := ""
	if t, ok := g.valueTypes[arrOp.Value]; ok && t != "" {
		if strings.HasPrefix(t, "array<") && strings.HasSuffix(t, ">") {
			elemType = strings.TrimSpace(t[6 : len(t)-1])
		} else if strings.HasPrefix(t, "[]<") && strings.HasSuffix(t, ">") {
			elemType = strings.TrimSpace(t[3 : len(t)-1])
		}
	}
	// Pick the runtime suffix and the C carrier type for the result.
	// Treat unknown / inferred element types as int — that's what the
	// call sites have historically defaulted to and most arrays in
	// practice are int32_t-sized.
	var rtSuffix, ptrCType, retArrType string
	switch elemType {
	case "string":
		rtSuffix = "str"
		ptrCType = "const char**"
		retArrType = "array<string>"
	case "int", "", "<infer>":
		rtSuffix = "int"
		ptrCType = "int32_t*"
		retArrType = "array<int>"
	default:
		return false
	}

	varName := g.getVariableName(inst.ID)
	arr := g.getOperandValue(arrOp)
	arrLen := g.getOperandLengthExpr(arrOp)

	declOrAssign := func(cType, expr string) {
		if g.declaredVariables[inst.ID] {
			g.output.WriteString(fmt.Sprintf("  %s = %s;\n", varName, expr))
		} else {
			g.output.WriteString(fmt.Sprintf("  %s %s = %s;\n", cType, varName, expr))
			g.declaredVariables[inst.ID] = true
		}
	}

	switch funcName {
	case "std.array.contains":
		if len(inst.Operands) < 3 {
			return false
		}
		val := g.getOperandValue(inst.Operands[2])
		declOrAssign("int32_t", fmt.Sprintf("omni_array_%s_contains(%s, %s, %s)", rtSuffix, arr, arrLen, val))
		g.valueTypes[inst.ID] = "bool"
		return true
	case "std.array.index_of":
		if len(inst.Operands) < 3 {
			return false
		}
		val := g.getOperandValue(inst.Operands[2])
		declOrAssign("int32_t", fmt.Sprintf("omni_array_%s_index_of(%s, %s, %s)", rtSuffix, arr, arrLen, val))
		g.valueTypes[inst.ID] = "int"
		return true
	case "std.array.append":
		if len(inst.Operands) < 3 {
			return false
		}
		val := g.getOperandValue(inst.Operands[2])
		declOrAssign(ptrCType, fmt.Sprintf("omni_array_%s_append(%s, %s, %s)", rtSuffix, arr, arrLen, val))
		g.arrayLengthExprs[inst.ID] = fmt.Sprintf("(%s) + 1", arrLen)
		g.valueTypes[inst.ID] = retArrType
		return true
	case "std.array.prepend":
		if len(inst.Operands) < 3 {
			return false
		}
		val := g.getOperandValue(inst.Operands[2])
		declOrAssign(ptrCType, fmt.Sprintf("omni_array_%s_prepend(%s, %s, %s)", rtSuffix, arr, arrLen, val))
		g.arrayLengthExprs[inst.ID] = fmt.Sprintf("(%s) + 1", arrLen)
		g.valueTypes[inst.ID] = retArrType
		return true
	case "std.array.insert":
		if len(inst.Operands) < 4 {
			return false
		}
		idx := g.getOperandValue(inst.Operands[2])
		val := g.getOperandValue(inst.Operands[3])
		declOrAssign(ptrCType, fmt.Sprintf("omni_array_%s_insert(%s, %s, %s, %s)", rtSuffix, arr, arrLen, idx, val))
		g.arrayLengthExprs[inst.ID] = fmt.Sprintf("(%s) + 1", arrLen)
		g.valueTypes[inst.ID] = retArrType
		return true
	case "std.array.remove":
		if len(inst.Operands) < 3 {
			return false
		}
		idx := g.getOperandValue(inst.Operands[2])
		declOrAssign(ptrCType, fmt.Sprintf("omni_array_%s_remove(%s, %s, %s)", rtSuffix, arr, arrLen, idx))
		g.arrayLengthExprs[inst.ID] = fmt.Sprintf("(%s) - 1", arrLen)
		g.valueTypes[inst.ID] = retArrType
		return true
	case "std.array.concat":
		if len(inst.Operands) < 3 {
			return false
		}
		bOp := inst.Operands[2]
		bArr := g.getOperandValue(bOp)
		bLen := g.getOperandLengthExpr(bOp)
		declOrAssign(ptrCType, fmt.Sprintf("omni_array_%s_concat(%s, %s, %s, %s)", rtSuffix, arr, arrLen, bArr, bLen))
		g.arrayLengthExprs[inst.ID] = fmt.Sprintf("(%s) + (%s)", arrLen, bLen)
		g.valueTypes[inst.ID] = retArrType
		return true
	case "std.array.slice":
		if len(inst.Operands) < 4 {
			return false
		}
		start := g.getOperandValue(inst.Operands[2])
		end := g.getOperandValue(inst.Operands[3])
		declOrAssign(ptrCType, fmt.Sprintf("omni_array_%s_slice(%s, %s, %s, %s)", rtSuffix, arr, arrLen, start, end))
		g.arrayLengthExprs[inst.ID] = fmt.Sprintf("(%s) - (%s)", end, start)
		g.valueTypes[inst.ID] = retArrType
		return true
	case "std.algorithms.shuffle":
		// Length-preserving Fisher–Yates. Forward the input length onto
		// the result so a downstream len() / index keeps working.
		declOrAssign(ptrCType, fmt.Sprintf("omni_array_%s_shuffle(%s, %s)", rtSuffix, arr, arrLen))
		g.arrayLengthExprs[inst.ID] = arrLen
		g.valueTypes[inst.ID] = retArrType
		return true
	case "std.algorithms.unique":
		// Variable-length output: the runtime writes the result count
		// to an int32_t we allocate next to the array variable. We
		// register that variable as the array's runtime length so
		// len(unique(arr)) works downstream.
		lenVar := fmt.Sprintf("__omni_unique_len_%d", inst.ID)
		g.output.WriteString(fmt.Sprintf("  int32_t %s = 0;\n", lenVar))
		declOrAssign(ptrCType, fmt.Sprintf("omni_array_%s_unique(%s, %s, &%s)", rtSuffix, arr, arrLen, lenVar))
		g.arrayLengthExprs[inst.ID] = lenVar
		g.valueTypes[inst.ID] = retArrType
		return true
	}
	return false
}

// isLengthPreservingArrayIntrinsic reports whether `name` is an
// std.algorithms intrinsic that returns a fresh array of the same
// length as its first input. Used at call sites to forward the
// runtime length to the result so a downstream len() / search call
// keeps working.
func isLengthPreservingArrayIntrinsic(name string) bool {
	switch name {
	case "std.algorithms.bubble_sort",
		"std.algorithms.selection_sort",
		"std.algorithms.insertion_sort",
		"std.algorithms.reverse",
		"std.algorithms.rotate":
		return true
	}
	return false
}

func (g *CGenerator) isPrimitiveType(omniType string) bool {
	switch omniType {
	case "int", "float", "double", "string", "void", "void*", "bool", "ptr", "char":
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

// getOmniTypeConstant returns the OMNI_TYPE_* constant for a given OmniLang type
func (g *CGenerator) getOmniTypeConstant(omniType string) string {
	// Normalize type
	omniType = strings.TrimSpace(omniType)

	// Handle generic types
	if strings.HasPrefix(omniType, "map<") {
		return "OMNI_TYPE_MAP"
	}
	if strings.HasPrefix(omniType, "array<") || strings.HasPrefix(omniType, "[]<") {
		return "OMNI_TYPE_ARRAY"
	}
	if strings.HasPrefix(omniType, "Promise<") {
		// Promise is not a map value type, but handle it
		return "OMNI_TYPE_ANY"
	}

	// Handle primitive types
	switch omniType {
	case "int", "int32_t", "int32":
		return "OMNI_TYPE_INT"
	case "string", "const char*", "char*":
		return "OMNI_TYPE_STRING"
	case "float", "double":
		return "OMNI_TYPE_FLOAT"
	case "bool":
		return "OMNI_TYPE_BOOL"
	case "any":
		return "OMNI_TYPE_ANY"
	default:
		// Unknown type - assume struct or other complex type
		// Check if it's a known struct type
		if !g.isPrimitiveType(omniType) && !strings.Contains(omniType, "(") && !strings.Contains(omniType, "<") {
			return "OMNI_TYPE_STRUCT"
		}
		return "OMNI_TYPE_ANY" // Fallback
	}
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
	if keyType == "float" || keyType == "double" {
		keyType = "float"
	}
	if keyType == "bool" {
		keyType = "bool"
	}

	// Handle any type combinations
	if keyType == "any" && valueType == "any" {
		return "omni_map_put_any_any"
	}
	if keyType == "any" {
		switch valueType {
		case "int":
			return "omni_map_put_any_int"
		case "string":
			return "omni_map_put_any_string"
		case "float":
			return "omni_map_put_any_float"
		case "bool":
			return "omni_map_put_any_bool"
		default:
			// For other value types with any key, use any_any
			return "omni_map_put_any_any"
		}
	}
	if valueType == "any" {
		if keyType == "string" {
			return "omni_map_put_string_any"
		} else if keyType == "int" {
			return "omni_map_put_int_any"
		} else {
			// For other key types with any value, use any_any
			return "omni_map_put_any_any"
		}
	}

	// Handle standard type combinations
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
		default:
			// For complex value types (maps, arrays, structs), use string_any
			return "omni_map_put_string_any"
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
		default:
			// For complex value types, use int_any
			return "omni_map_put_int_any"
		}
	} else {
		// For complex key types (maps, arrays, structs), use any_* functions
		if valueType == "int" {
			return "omni_map_put_any_int"
		} else if valueType == "string" {
			return "omni_map_put_any_string"
		} else if valueType == "float" {
			return "omni_map_put_any_float"
		} else if valueType == "bool" {
			return "omni_map_put_any_bool"
		} else {
			return "omni_map_put_any_any"
		}
	}
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

// getMapSetFunction returns the appropriate map set function name for the given key and value types.
func (g *CGenerator) getMapSetFunction(keyType, valueType string) string {
	if valueType == "double" {
		valueType = "float"
	}
	if keyType == "string" {
		switch valueType {
		case "int":
			return "omni_map_set_string_int"
		case "string":
			return "omni_map_set_string_string"
		case "float":
			return "omni_map_set_string_float"
		case "bool":
			return "omni_map_set_string_bool"
		}
	} else if keyType == "int" {
		switch valueType {
		case "int":
			return "omni_map_set_int_int"
		case "string":
			return "omni_map_set_int_string"
		case "float":
			return "omni_map_set_int_float"
		case "bool":
			return "omni_map_set_int_bool"
		}
	}
	return ""
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
	hasMain := false
	for _, fn := range g.module.Functions {
		if fn.Name == "main" {
			mainReturnType = fn.ReturnType
			hasMain = true
			break
		}
	}

	g.output.WriteString("int main(int argc, char** argv) {\n")
	g.output.WriteString("    omni_args_init(argc, argv);\n")

	// Library files (no main) still get a trivial entry point so `omnic`
	// produces a runnable binary — useful for test harnesses that compile
	// each file individually.
	if !hasMain {
		g.output.WriteString("    return 0;\n")
		g.output.WriteString("}\n")
		return
	}

	// Handle Promise return types (async main) - unwrap to inner type
	if strings.HasPrefix(mainReturnType, "Promise<") {
		innerType := mainReturnType[8 : len(mainReturnType)-1]
		mainReturnType = innerType
	}

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
		// omni_main always returns int32_t (even for async main)
		g.output.WriteString("    int32_t result = omni_main();\n")
		if mainReturnType == "float" || mainReturnType == "double" {
			g.output.WriteString("    printf(\"OmniLang program result: %f\\n\", (double)result);\n")
		} else {
			g.output.WriteString("    printf(\"OmniLang program result: %d\\n\", result);\n")
		}
		g.output.WriteString("    return result;\n")
	}

	g.output.WriteString("}\n")
}
