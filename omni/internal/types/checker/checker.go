package checker

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/omni-lang/omni/internal/ast"
	"github.com/omni-lang/omni/internal/lexer"
	"github.com/omni-lang/omni/internal/moduleloader"
	"github.com/omni-lang/omni/internal/types"
)

const (
	typeError = "<error>"
	typeInfer = "<inferred>"
	typeVoid  = "void"
)

// Check runs the OmniLang type checker over the provided module and returns an
// aggregated diagnostic error if any issues are found.
func Check(filename, src string, mod *ast.Module) error {
	c := &Checker{
		filename:         filename,
		lines:            splitLines(src),
		knownTypes:       make(map[string]struct{}),
		typeAliases:      make(map[string]string),
		structFields:     make(map[string]map[string]string),
		structTypeParams: make(map[string][]ast.TypeParam),
		functions:        make(map[string]FunctionSignature),
		imports: map[string]bool{
			// Compiler-provided intrinsic namespaces under std that exist even
			// without an explicit `import std` — lets terse test files that
			// reference `std.test.start` / `std.assert.eq` / `std.io.println`
			// typecheck. Functions are still emitted as intrinsics by the
			// backend, so this doesn't create duplicate definitions.
			"test":        true,
			"assert":      true,
			"io":          true,
			"math":        true,
			"string":      true,
			"array":       true,
			"os":          true,
			"collections": true,
			"testing":     true,
			"file":        true,
			"time":        true,
			"network":     true,
			"web":         true,
			"log":         true,
			"algorithms":  true,
		},
		moduleLoader:     *moduleloader.NewModuleLoader(),
		typeParams:       make(map[string]bool),
		processedImports: make(map[string]bool),
	}

	// Add the omni std directory to search paths
	// Find the omni root directory by looking for the std directory
	if abs, err := filepath.Abs(filepath.Dir(filename)); err == nil {
		// Walk up the directory tree to find the omni root
		current := abs
		for {
			stdPath := filepath.Join(current, "std")
			if _, err := os.Stat(stdPath); err == nil {
				// Check if this is the main std directory (contains std.omni)
				mainStdPath := filepath.Join(stdPath, "std.omni")
				if _, err := os.Stat(mainStdPath); err == nil {
					c.moduleLoader.AddSearchPath(current)
					break
				}
			}
			parent := filepath.Dir(current)
			if parent == current {
				break // Reached root
			}
			current = parent
		}
	}
	c.initBuiltins()
	c.collectTypeDecls(mod)
	c.enterScope()
	c.registerTopLevelSymbols(mod)
	c.processImports(mod)
	c.registerMergedTypeAliases(mod) // Register type aliases first so they're available when resolving function signatures
	c.registerMergedFunctionSignatures(mod)
	c.checkModule(mod)
	c.leaveScope()

	if len(c.diagnostics) == 0 {
		return nil
	}
	return errors.Join(c.diagnostics...)
}

// Checker encapsulates the mutable state required to validate an OmniLang AST.
type Checker struct {
	filename    string
	lines       []string
	knownTypes  map[string]struct{}
	typeAliases map[string]string // Maps type alias names to their underlying types

	structFields     map[string]map[string]string
	structTypeParams map[string][]ast.TypeParam // Store type parameters for generic structs
	functions        map[string]FunctionSignature

	scopes      []map[string]Symbol
	diagnostics []error

	functionStack []functionContext
	loopDepth     int // Track nesting depth of loops for break/continue validation

	// Import resolution
	imports map[string]bool // available imported symbols
	// Module loader for local imports
	moduleLoader moduleloader.ModuleLoader

	// Generic type context
	typeParams map[string]bool // Currently active type parameters

	processedImports map[string]bool

	// Expected map type context for map literal type inference
	expectedMapType string
}

// enterTypeParams enters a new type parameter scope
func (c *Checker) enterTypeParams(typeParams []ast.TypeParam) {
	for _, param := range typeParams {
		c.typeParams[param.Name] = true
	}
}

// leaveTypeParams leaves the current type parameter scope
func (c *Checker) leaveTypeParams(typeParams []ast.TypeParam) {
	for _, param := range typeParams {
		delete(c.typeParams, param.Name)
	}
}

// isTypeParam checks if a name is a currently active type parameter
func (c *Checker) isTypeParam(name string) bool {
	return c.typeParams[name]
}

// isFunctionTypeParam checks if a name is a type parameter of a specific function
func (c *Checker) isFunctionTypeParam(name string, typeParams []ast.TypeParam) bool {
	for _, param := range typeParams {
		if param.Name == name {
			return true
		}
	}
	return false
}

// inferTypeParametersFromGeneric infers type parameters from generic types
// For example, if expected is "array<T>" and argType is "array<int>", it infers T = int
// Now handles arbitrary generic types like Result<T>, List<T>, etc.
func (c *Checker) inferTypeParametersFromGeneric(expected, argType string, typeParams []ast.TypeParam) map[string]string {
	inferred := make(map[string]string)

	// Find the generic delimiter position for both types
	expectedLess := strings.Index(expected, "<")
	argLess := strings.Index(argType, "<")

	// Both must be generic types
	if expectedLess == -1 || argLess == -1 {
		return inferred
	}

	// Extract base type names (everything before <)
	expectedBase := expected[:expectedLess]
	argBase := argType[:argLess]

	// Base types must match (e.g., both "array" or both "Result")
	if expectedBase != argBase {
		return inferred
	}

	// Extract inner types (everything between < and >)
	// Find the matching > for the < we found, accounting for nested generics
	expectedGreater := c.findMatchingGreater(expected, expectedLess)
	argGreater := c.findMatchingGreater(argType, argLess)

	if expectedGreater == -1 || argGreater == -1 {
		return inferred // Malformed generic type
	}

	expectedInner := expected[expectedLess+1 : expectedGreater]
	argInner := argType[argLess+1 : argGreater]

	// Handle single type parameter: TypeName<T> vs TypeName<int>
	if !strings.Contains(expectedInner, ",") && !strings.Contains(argInner, ",") {
		// Single type parameter
		if c.isFunctionTypeParam(strings.TrimSpace(expectedInner), typeParams) {
			inferred[strings.TrimSpace(expectedInner)] = strings.TrimSpace(argInner)
		}
		return inferred
	}

	// Handle multiple type parameters: TypeName<K,V> vs TypeName<string,int>
	// Split by comma, but be careful of nested generics
	expectedParts := c.splitGenericArgs(expectedInner)
	argParts := c.splitGenericArgs(argInner)

	if len(expectedParts) == len(argParts) {
		for i := 0; i < len(expectedParts); i++ {
			expectedPart := strings.TrimSpace(expectedParts[i])
			argPart := strings.TrimSpace(argParts[i])
			if c.isFunctionTypeParam(expectedPart, typeParams) {
				inferred[expectedPart] = argPart
			}
		}
	}

	return inferred
}

// findMatchingGreater finds the matching > for a < at the given position
// Handles nested generics by tracking depth
func (c *Checker) findMatchingGreater(typeStr string, lessPos int) int {
	if lessPos < 0 || lessPos >= len(typeStr) {
		return -1
	}
	depth := 1
	for i := lessPos + 1; i < len(typeStr); i++ {
		switch typeStr[i] {
		case '<':
			depth++
		case '>':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1 // No matching >
}

// splitGenericArgs splits generic arguments by comma, handling nested generics
func (c *Checker) splitGenericArgs(s string) []string {
	var parts []string
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
				parts = append(parts, current.String())
				current.Reset()
			} else {
				current.WriteRune(r)
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// Symbol represents an entry in the scope stack.
type Symbol struct {
	Type    string
	Mutable bool
}

// FunctionSignature captures parameter and return type information for a function.
type FunctionSignature struct {
	Params     []string
	Return     string
	TypeParams []ast.TypeParam // Generic type parameters
}

type functionContext struct {
	Name       string
	ReturnType string
	IsAsync    bool
	HasReturn  bool
}

func (c *Checker) initBuiltins() {
	primitives := []types.Kind{
		types.KindInt,
		types.KindLong,
		types.KindByte,
		types.KindFloat,
		types.KindDouble,
		types.KindBool,
		types.KindChar,
		types.KindString,
		types.KindVoid,
	}
	for _, k := range primitives {
		c.knownTypes[string(k)] = struct{}{}
	}
	c.knownTypes["array"] = struct{}{}
	c.knownTypes["map"] = struct{}{}
	c.knownTypes["Promise"] = struct{}{}
	c.knownTypes["any"] = struct{}{} // Top type for dynamic values

	// Add builtin functions
	c.functions["len"] = FunctionSignature{
		Params: []string{typeInfer}, // Accept any array type
		Return: "int",
	}
}

func (c *Checker) collectTypeDecls(mod *ast.Module) {
	for _, decl := range mod.Decls {
		switch d := decl.(type) {
		case *ast.StructDecl:
			c.knownTypes[d.Name] = struct{}{}
			fields := make(map[string]string, len(d.Fields))
			for _, field := range d.Fields {
				fields[field.Name] = typeExprToString(field.Type)
			}
			c.structFields[d.Name] = fields
			// Store type parameters for generic structs
			if len(d.TypeParams) > 0 {
				c.structTypeParams[d.Name] = d.TypeParams
			}
		case *ast.EnumDecl:
			c.knownTypes[d.Name] = struct{}{}
		}
	}
}

func (c *Checker) registerTopLevelSymbols(mod *ast.Module) {
	for _, decl := range mod.Decls {
		switch d := decl.(type) {
		case *ast.LetDecl:
			var typ string
			if d.Type != nil {
				typ = c.resolveTypeExpr(d.Type)
			} else {
				// Type will be inferred later in checkLet
				typ = typeInfer
			}
			c.declare(d.Name, typ, false, d.Span())
		case *ast.VarDecl:
			var typ string
			if d.Type != nil {
				typ = c.resolveTypeExpr(d.Type)
			} else {
				// Type will be inferred later in checkVar
				typ = typeInfer
			}
			c.declare(d.Name, typ, true, d.Span())
		case *ast.FuncDecl:
			// Skip namespaced functions (imported modules) - they're already registered
			if strings.Contains(d.Name, ".") {
				continue
			}
			sig := c.buildFunctionSignature(d)
			if _, exists := c.functions[d.Name]; exists {
				c.report(d.Span(), fmt.Sprintf("function %q redeclared", d.Name), "rename the function or remove the duplicate declaration")
			}
			c.functions[d.Name] = sig
			// Store the full function type for first-class function support
			funcType := buildFunctionType(sig.Params, sig.Return)
			c.declare(d.Name, funcType, false, d.Span())
		}
	}
}

func (c *Checker) buildFunctionSignature(decl *ast.FuncDecl) FunctionSignature {
	// For generic functions, we can't resolve types yet, so store placeholder types
	// and resolve them later when the function is checked
	params := make([]string, len(decl.Params))
	for i, param := range decl.Params {
		if len(decl.TypeParams) > 0 {
			// For generic functions, store the full type expression as a string
			params[i] = c.typeExprToString(param.Type)
		} else {
			params[i] = c.resolveTypeExpr(param.Type)
		}
	}
	ret := typeVoid
	if decl.Return != nil {
		if len(decl.TypeParams) > 0 {
			// For generic functions, store the full type expression as a string
			ret = c.typeExprToString(decl.Return)
		} else {
			ret = c.resolveTypeExpr(decl.Return)
		}
	}

	// If function is async, wrap return type in Promise<T>, but avoid double wrapping
	if decl.IsAsync {
		asyncInner := ret
		if strings.HasPrefix(asyncInner, "Promise<") && strings.HasSuffix(asyncInner, ">") {
			if inner, ok := promiseInnerType(asyncInner); ok {
				asyncInner = inner
			}
		}
		if asyncInner == typeVoid || asyncInner == "" || asyncInner == typeInfer {
			ret = "Promise<void>"
		} else {
			ret = "Promise<" + asyncInner + ">"
		}
	}

	return FunctionSignature{Params: params, Return: ret, TypeParams: decl.TypeParams}
}

func (c *Checker) checkModule(mod *ast.Module) {
	for _, decl := range mod.Decls {
		switch d := decl.(type) {
		case *ast.LetDecl:
			c.checkLet(d, false)
		case *ast.VarDecl:
			c.checkLet((*ast.LetDecl)(d), true)
		case *ast.StructDecl:
			c.checkStruct(d)
		case *ast.EnumDecl:
			// Enumerations currently require no additional checks beyond registration.
		case *ast.FuncDecl:
			// Skip checking namespaced functions (imported/merged from std modules)
			// They were already validated when the std module was parsed
			if !strings.Contains(d.Name, ".") {
				c.checkFunc(d)
			}
		case *ast.TypeAliasDecl:
			// Process both qualified and unqualified type aliases
			c.checkTypeAliasDecl(d)
		}
	}
}

func (c *Checker) checkStruct(decl *ast.StructDecl) {
	// Enter type parameter scope for generic structs
	c.enterTypeParams(decl.TypeParams)

	for _, field := range decl.Fields {
		c.checkTypeExpr(field.Type)
	}

	// Leave type parameter scope
	c.leaveTypeParams(decl.TypeParams)
}

func (c *Checker) checkFunc(decl *ast.FuncDecl) {
	sig := c.functions[decl.Name]

	// Enter type parameter scope for generic functions FIRST
	c.enterTypeParams(sig.TypeParams)

	expectedReturn := sig.Return
	// Check if this is an async function (return type is Promise<T>)
	isAsync := strings.HasPrefix(sig.Return, "Promise<")
	var innerReturnType string
	if isAsync {
		// Extract inner type from Promise<T>
		if inner, ok := promiseInnerType(sig.Return); ok {
			innerReturnType = inner
		}
	}

	if decl.Return != nil {
		declaredReturn := c.checkTypeExpr(decl.Return)
		// For async functions, compare with inner type
		if isAsync {
			declaredInner := declaredReturn
			if strings.HasPrefix(declaredInner, "Promise<") && strings.HasSuffix(declaredInner, ">") {
				if inner, ok := promiseInnerType(declaredInner); ok {
					declaredInner = inner
				}
			}
			if !c.typesEqual(declaredInner, innerReturnType) {
				c.report(decl.Return.Span(), fmt.Sprintf("async function return type mismatch: declared %s but signature expects %s", declaredReturn, innerReturnType),
					"align the return type annotation with the inner type of Promise")
			}
			// Use inner type for body validation (function body returns int, not Promise<int>)
			expectedReturn = innerReturnType
		} else {
			expectedReturn = declaredReturn
		}
	} else if isAsync {
		// No return type annotation, use inner type for validation
		expectedReturn = innerReturnType
	}

	c.pushFunctionContext(decl.Name, expectedReturn, decl.IsAsync)
	c.enterScope()
	for i, param := range decl.Params {
		paramType := c.checkTypeExpr(param.Type)
		if sig.Params != nil && i < len(sig.Params) {
			if !c.typesEqual(sig.Params[i], paramType) {
				c.report(param.Span, fmt.Sprintf("parameter %q type mismatch", param.Name), "align the signature with the annotation")
			}
		}
		c.declare(param.Name, paramType, true, param.Span)
	}

	if decl.ExprBody != nil {
		exprType := c.checkExpr(decl.ExprBody)
		c.validateFunctionReturn(exprType, decl.ExprBody.Span())
		if ctx := c.currentFunctionContext(); ctx != nil {
			ctx.HasReturn = true
		}
	}

	if decl.Body != nil {
		c.checkBlock(decl.Body)
	}
	ctx := c.currentFunctionContext()
	missingReturn := ctx != nil && ctx.ReturnType != typeVoid && ctx.ReturnType != typeInfer && !ctx.HasReturn
	expectedType := ""
	if ctx != nil {
		expectedType = ctx.ReturnType
	}

	c.leaveScope()
	if missingReturn {
		c.report(decl.Span(), fmt.Sprintf("function %s is missing a return statement", decl.Name),
			fmt.Sprintf("ensure all code paths return a %s value", expectedType))
	}
	c.popFunctionContext()

	// Leave type parameter scope
	c.leaveTypeParams(sig.TypeParams)
}

func (c *Checker) checkLet(decl *ast.LetDecl, mutable bool) {
	expectedType := typeInfer
	if decl.Type != nil {
		expectedType = c.checkTypeExpr(decl.Type)
	}
	valueType := typeInfer
	if decl.Value != nil {
		// Special handling for empty array literals with an annotated target type
		if arrayType, ok := c.resolveEmptyArrayLiteral(decl.Value, expectedType); ok {
			valueType = arrayType
		} else if expectedType != typeInfer && expectedType != "" && strings.HasPrefix(expectedType, "map<") {
			// Set expected map type context for map literal type inference
			c.expectedMapType = expectedType
			valueType = c.checkExpr(decl.Value)
			c.expectedMapType = "" // Clear after use
		} else if lambda, ok := decl.Value.(*ast.LambdaExpr); ok && expectedType != typeInfer && expectedType != "" {
			// Check if expected type is a function type (contains ") -> ")
			if strings.Contains(expectedType, ") -> ") {
				// Parse expected function type to get parameter types
				expectedParamTypes := c.parseFunctionTypeParams(expectedType)
				if expectedParamTypes != nil && len(expectedParamTypes) == len(lambda.Params) {
					// Check lambda with expected parameter types
					valueType = c.checkLambdaWithTypes(lambda, expectedParamTypes)
				} else {
					// Parameter count mismatch or parsing failed
					// If parsing failed (nil), try to check normally and let error be reported
					// If count mismatch, report error
					if expectedParamTypes == nil {
						// Parsing failed - check normally (will likely fail)
						valueType = c.checkExpr(decl.Value)
					} else {
						// Count mismatch - report error and check normally
						c.report(decl.Span(), fmt.Sprintf("lambda expects %d parameters, but function type has %d", len(lambda.Params), len(expectedParamTypes)),
							fmt.Sprintf("provide %d parameter(s) matching the function type", len(expectedParamTypes)))
						valueType = c.checkExpr(decl.Value)
					}
				}
			} else {
				// Not a function type - check normally
				valueType = c.checkExpr(decl.Value)
			}
		} else {
			valueType = c.checkExpr(decl.Value)
		}
	}

	finalType := expectedType
	if finalType == typeInfer {
		finalType = valueType
	} else if valueType != typeInfer && valueType != typeError {
		// Special case: if expected type is map<*, any>, allow any inferred map type
		if strings.HasPrefix(finalType, "map<") && strings.HasSuffix(finalType, ">") {
			inner := finalType[4 : len(finalType)-1]
			parts := strings.Split(inner, ",")
			if len(parts) == 2 {
				expectedValueType := strings.TrimSpace(parts[1])
				expectedValueType = strings.TrimSpace(expectedValueType)
				if expectedValueType == "any" {
					// For map<*, any>, we allow any value type, so skip assignability check
					// The valueType might be inferred as map<string, string> but that's okay
					// We'll use the expected type instead
					finalType = expectedType
				} else if !c.isAssignable(valueType, finalType) {
					c.report(decl.Span(), fmt.Sprintf("type mismatch: cannot assign %s to %s", valueType, finalType),
						fmt.Sprintf("convert the expression to %s or change the variable type to %s", finalType, valueType))
				}
			} else if !c.isAssignable(valueType, finalType) {
				c.report(decl.Span(), fmt.Sprintf("type mismatch: cannot assign %s to %s", valueType, finalType),
					fmt.Sprintf("convert the expression to %s or change the variable type to %s", finalType, valueType))
			}
		} else if !c.isAssignable(valueType, finalType) {
			c.report(decl.Span(), fmt.Sprintf("type mismatch: cannot assign %s to %s", valueType, finalType),
				fmt.Sprintf("convert the expression to %s or change the variable type to %s", finalType, valueType))
		}
	}

	if finalType != typeInfer && finalType != typeError {
		c.updateSymbolType(decl.Name, finalType)
	}

	// For top-level lets/vars the symbol already exists; locals are declared elsewhere.
	if !c.symbolExists(decl.Name) {
		c.declare(decl.Name, finalType, mutable, decl.Span())
	}
}

func (c *Checker) checkBlock(block *ast.BlockStmt) {
	if block == nil {
		return
	}
	c.enterScope()
	for _, stmt := range block.Statements {
		c.checkStmt(stmt)
	}
	c.leaveScope()
}

func (c *Checker) checkStmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.ReturnStmt:
		c.handleReturn(s)
	case *ast.ExprStmt:
		c.checkExpr(s.Expr)
	case *ast.IfStmt:
		condType := c.checkExpr(s.Cond)
		if !c.typesEqual(condType, "bool") {
			c.report(s.Cond.Span(), "if condition must be bool", "use a boolean expression")
		}
		c.checkBlock(s.Then)
		if s.Else != nil {
			c.checkStmt(s.Else)
		}
	case *ast.BlockStmt:
		c.checkBlock(s)
	case *ast.ForStmt:
		c.checkForStmt(s)
	case *ast.WhileStmt:
		c.checkWhileStmt(s)
	case *ast.BreakStmt:
		// Break statements are only allowed in loops
		if c.loopDepth == 0 {
			c.report(s.Span(), "break statement outside of loop", "use break only inside for or while loops")
		}
	case *ast.ContinueStmt:
		// Continue statements are only allowed in loops
		if c.loopDepth == 0 {
			c.report(s.Span(), "continue statement outside of loop", "use continue only inside for or while loops")
		}
	case *ast.BindingStmt:
		c.checkBindingStmt(s)
	case *ast.ShortVarDeclStmt:
		declaredType := c.checkTypeExpr(s.Type)
		valueType := typeInfer
		if s.Value != nil {
			valueType = c.checkExpr(s.Value)
		}
		finalType := declaredType
		if declaredType == typeInfer {
			finalType = valueType
		} else if valueType != typeInfer && valueType != typeError && !c.isAssignable(valueType, declaredType) {
			c.report(s.Span(), fmt.Sprintf("type mismatch: cannot assign %s to %s", valueType, declaredType),
				fmt.Sprintf("convert the expression to %s or change the variable type to %s", declaredType, valueType))
		}
		c.declare(s.Name, finalType, true, s.Span())
	case *ast.AssignmentStmt:
		c.checkAssignmentExpr(&ast.AssignmentExpr{SpanInfo: s.SpanInfo, Left: s.Left, Right: s.Right})
	case *ast.IncrementStmt:
		targetType := c.checkExpr(s.Target)
		if !isNumeric(targetType) {
			c.report(s.Target.Span(), "increment operator requires numeric operand",
				fmt.Sprintf("use a numeric variable (int or float), got %s", c.checkExpr(s.Target)))
		}
		if ident, ok := s.Target.(*ast.IdentifierExpr); ok {
			if sym, found := c.lookupSymbol(ident.Name); found && !sym.Mutable {
				c.report(s.Target.Span(), fmt.Sprintf("cannot modify immutable variable %q", ident.Name), "declare it with var if mutation is required")
			}
		}
	case *ast.TryStmt:
		// Check try block
		c.checkBlock(s.TryBlock)

		// Check catch clauses
		for _, catchClause := range s.CatchClauses {
			c.enterScope()
			if catchClause.ExceptionVar != "" {
				// Declare the exception variable in the catch scope
				exceptionType := "string" // Default exception type
				if catchClause.ExceptionType != "" {
					exceptionType = catchClause.ExceptionType
				}
				c.declare(catchClause.ExceptionVar, exceptionType, false, catchClause.Span())
			}
			c.checkBlock(catchClause.Block)
			c.leaveScope()
		}

		// Check finally block
		if s.FinallyBlock != nil {
			c.checkBlock(s.FinallyBlock)
		}
	case *ast.ThrowStmt:
		// Check the expression being thrown
		c.checkExpr(s.Expr)
	}
}

// checkTypeAliasDecl checks a type alias declaration
func (c *Checker) checkTypeAliasDecl(decl *ast.TypeAliasDecl) {
	// Check the type expression
	if decl.Type == nil {
		c.report(decl.Span(), "type alias must have a type", "provide a type expression after the '='")
		return
	}

	// For generic aliases, enter type parameter scope so T, U, etc. are recognized
	// Convert []string to []ast.TypeParam for scope management
	var typeParams []ast.TypeParam
	if len(decl.TypeParams) > 0 {
		typeParams = make([]ast.TypeParam, len(decl.TypeParams))
		for i, paramName := range decl.TypeParams {
			// Create a TypeParam with the name (span will be approximate)
			typeParams[i] = ast.TypeParam{
				Name: paramName,
				Span: decl.Span(), // Use alias span as approximation
			}
		}
		c.enterTypeParams(typeParams)
	}

	// Check the underlying type expression (T is now in scope for generic aliases)
	underlyingType := c.checkTypeExpr(decl.Type)

	// Leave type parameter scope
	if len(typeParams) > 0 {
		c.leaveTypeParams(typeParams)
	}

	// Store the type alias mapping
	// For generic aliases, store the template string (e.g., "T?") which will be substituted later
	c.knownTypes[decl.Name] = struct{}{}
	c.typeAliases[decl.Name] = underlyingType

	// Store type parameters for generic aliases so we can substitute them later
	if len(typeParams) > 0 {
		c.structTypeParams[decl.Name] = typeParams
	}
}

func (c *Checker) checkBindingStmt(stmt *ast.BindingStmt) {
	declaredType := typeInfer
	if stmt.Type != nil {
		declaredType = c.checkTypeExpr(stmt.Type)
	}
	valueType := typeInfer
	if stmt.Value != nil {
		// Handle empty array literals when a declared type is available
		if arrayType, ok := c.resolveEmptyArrayLiteral(stmt.Value, declaredType); ok {
			valueType = arrayType
		} else if declaredType != typeInfer && declaredType != "" && strings.HasPrefix(declaredType, "map<") {
			// Set expected map type context for map literal type inference
			c.expectedMapType = declaredType
			valueType = c.checkExpr(stmt.Value)
			c.expectedMapType = "" // Clear after use
		} else if lambda, ok := stmt.Value.(*ast.LambdaExpr); ok && declaredType != typeInfer && declaredType != "" {
			if strings.Contains(declaredType, ") -> ") {
				expectedParamTypes := c.parseFunctionTypeParams(declaredType)
				if expectedParamTypes != nil && len(expectedParamTypes) == len(lambda.Params) {
					valueType = c.checkLambdaWithTypes(lambda, expectedParamTypes)
				} else {
					valueType = c.checkExpr(stmt.Value)
				}
			} else {
				valueType = c.checkExpr(stmt.Value)
			}
		} else {
			valueType = c.checkExpr(stmt.Value)
		}
	}
	finalType := declaredType
	if declaredType == typeInfer {
		finalType = valueType
	} else if valueType != typeInfer && valueType != typeError && !c.isAssignable(valueType, declaredType) {
		c.report(stmt.Span(), fmt.Sprintf("type mismatch: cannot assign %s to %s", valueType, declaredType),
			fmt.Sprintf("convert the expression to %s or change the variable type to %s", declaredType, valueType))
	}
	c.declare(stmt.Name, finalType, stmt.Mutable, stmt.Span())
}

// resolveEmptyArrayLiteral allows empty array literals to take on an expected type context.
func (c *Checker) resolveEmptyArrayLiteral(expr ast.Expr, expectedType string) (string, bool) {
	if expectedType == "" || expectedType == typeInfer {
		return "", false
	}
	arr, ok := expr.(*ast.ArrayLiteralExpr)
	if !ok || len(arr.Elements) != 0 {
		return "", false
	}
	if _, ok := arrayElementType(expectedType); !ok {
		return "", false
	}
	return expectedType, true
}

// resolveMapLiteralWithExpectedType allows map literals to use the expected type when provided.
// This is especially useful for map<string, any> where values might be different types.
func (c *Checker) resolveMapLiteralWithExpectedType(expr ast.Expr, expectedType string) (string, bool) {
	if expectedType == "" || expectedType == typeInfer {
		return "", false
	}
	mapLit, ok := expr.(*ast.MapLiteralExpr)
	if !ok {
		return "", false
	}
	// Check if expected type is a map type
	if !strings.HasPrefix(expectedType, "map<") || !strings.HasSuffix(expectedType, ">") {
		return "", false
	}
	// Extract key and value types from expected map type
	inner := expectedType[4 : len(expectedType)-1] // Remove "map<" and ">"
	parts := strings.Split(inner, ",")
	if len(parts) != 2 {
		return "", false
	}
	expectedKeyType := strings.TrimSpace(parts[0])
	expectedValueType := strings.TrimSpace(parts[1])
	expectedValueType = strings.TrimSpace(expectedValueType)

	// If expected value type is "any", we can accept any value types
	if expectedValueType == "any" {
		// Check that all keys match expected key type
		for _, entry := range mapLit.Entries {
			keyType := c.checkExpr(entry.Key)
			if keyType != typeInfer && keyType != typeError && !c.isAssignable(keyType, expectedKeyType) {
				// Key type mismatch - don't use expected type
				return "", false
			}
		}
		// All keys match, use expected type
		return expectedType, true
	}

	// If expected value type is not "any", check that all values match
	// This is the normal case handled by checkExpr, so we don't need special handling
	return "", false
}

func (c *Checker) checkForStmt(stmt *ast.ForStmt) {
	// Validate AST invariant: IsRange and classic form fields are mutually exclusive
	if stmt.IsRange {
		// Range form: should not have Init, Condition, or Post
		if stmt.Init != nil || stmt.Condition != nil || stmt.Post != nil {
			c.report(stmt.Span(), "range for loop cannot have init, condition, or post clauses", "use 'for item in items { ... }' syntax")
		}
	} else {
		// Classic form: should not have Target or Iterable
		if stmt.Target != nil || stmt.Iterable != nil {
			c.report(stmt.Span(), "classic for loop cannot have target or iterable", "use 'for init; cond; post { ... }' syntax")
		}
	}

	c.enterScope()
	c.loopDepth++ // Increment loop depth
	defer func() {
		c.loopDepth-- // Decrement on exit
		c.leaveScope()
	}()

	if stmt.IsRange {
		// Range form: for item in items { ... }
		if stmt.Iterable == nil {
			c.report(stmt.Span(), "range for loop requires an iterable expression", "provide an array or map to iterate over")
			return
		}
		iterType := c.checkExpr(stmt.Iterable)
		elementType := typeInfer
		if t, ok := arrayElementType(iterType); ok {
			elementType = t
		} else if _, v, ok := mapTypes(iterType); ok {
			elementType = v
		} else if iterType != typeError {
			c.report(stmt.Iterable.Span(), "range expects array or map", "iterate over a supported collection type")
		}
		if stmt.Target != nil {
			c.declare(stmt.Target.Name, elementType, false, stmt.Target.Span())
		}
		c.checkBlock(stmt.Body)
		return
	}

	// Classic form: for init; cond; post { ... }
	if stmt.Init != nil {
		c.checkStmt(stmt.Init)
	}
	if stmt.Condition != nil {
		condType := c.checkExpr(stmt.Condition)
		if condType != typeError && !c.typesEqual(condType, "bool") {
			c.report(stmt.Condition.Span(), "for loop condition must be bool", "use a boolean expression")
		}
	}
	if stmt.Post != nil {
		c.checkStmt(stmt.Post)
	}
	c.checkBlock(stmt.Body)
}

func (c *Checker) checkWhileStmt(stmt *ast.WhileStmt) {
	condType := c.checkExpr(stmt.Cond)
	if !c.typesEqual(condType, "bool") {
		c.report(stmt.Cond.Span(), "while condition must be bool", "use a boolean expression")
	}
	c.enterScope()
	c.loopDepth++ // Increment loop depth
	defer func() {
		c.loopDepth-- // Decrement on exit
		c.leaveScope()
	}()
	c.checkBlock(stmt.Body)
}

func (c *Checker) handleReturn(ret *ast.ReturnStmt) {
	ctx := c.currentFunctionContext()
	if ctx == nil {
		c.report(ret.Span(), "return statement outside of function", "remove the return or place it inside a function")
		if ret.Value != nil {
			c.checkExpr(ret.Value)
		}
		return
	}
	ctx.HasReturn = true
	expected := ctx.ReturnType
	if ret.Value == nil {
		if expected != typeVoid && expected != typeInfer {
			c.report(ret.Span(), "missing return value", "return an expression matching the function return type")
		} else if expected == typeInfer {
			c.updateCurrentFunctionReturn(typeVoid)
		}
		return
	}

	valueType := c.checkExpr(ret.Value)
	if expected == typeVoid {
		c.report(ret.Value.Span(), "function declared void cannot return a value", "remove the expression or change the return type")
		return
	}
	if expected == typeInfer {
		if valueType != typeError {
			c.updateCurrentFunctionReturn(valueType)
		}
		return
	}
	if valueType != typeError && !c.isAssignable(valueType, expected) {
		c.report(ret.Value.Span(), fmt.Sprintf("cannot return %s from function returning %s", valueType, expected), "return an expression with the correct type")
	}
}

func (c *Checker) validateFunctionReturn(actualType string, span lexer.Span) {
	ctx := c.currentFunctionContext()
	if ctx == nil {
		return
	}
	expected := ctx.ReturnType
	if expected == typeVoid {
		if actualType != typeVoid && actualType != typeError {
			c.report(span, "function declared void cannot return a value", "remove the expression or specify a return type")
		}
		return
	}
	if expected == typeInfer {
		if actualType != typeError {
			c.updateCurrentFunctionReturn(actualType)
		}
		return
	}
	if actualType != typeError && !c.typesEqual(expected, actualType) {
		c.report(span, fmt.Sprintf("function body produces %s but %s expected", actualType, expected), "adjust the body or return type annotation")
	}
}

func (c *Checker) checkExpr(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.IdentifierExpr:
		if sym, ok := c.lookupSymbol(e.Name); ok {
			return sym.Type
		}
		// Check if it's a builtin function
		if sig, exists := c.functions[e.Name]; exists {
			// Return the full function type
			return buildFunctionType(sig.Params, sig.Return)
		}
		// Check if it's a qualified std symbol
		if c.isStdSymbol(e.Name) {
			// For now, assume all std functions return void or int
			if strings.Contains(e.Name, "io.") {
				return "void"
			}
			if strings.Contains(e.Name, "math.") {
				return "int"
			}
			return "void"
		}
		// Provide better suggestions based on common patterns
		var hint string
		if strings.Contains(e.Name, ".") {
			// Qualified identifier - might be a module import issue
			parts := strings.Split(e.Name, ".")
			if len(parts) == 2 {
				hint = fmt.Sprintf("check if module '%s' is imported with 'import %s' or 'import %s as <alias>'", parts[0], e.Name, parts[0])
			} else {
				hint = "check if this qualified identifier is correct and the module is imported"
			}
		} else if isLikelyTypo(e.Name) {
			hint = fmt.Sprintf("did you mean one of: %s? Or declare the variable with 'let %s: <type> = <value>'", suggestSimilarIdentifiers(e.Name), e.Name)
		} else {
			hint = fmt.Sprintf("declare the variable with 'let %s: <type> = <value>' or 'var %s: <type> = <value>' before using it", e.Name, e.Name)
		}
		c.report(e.Span(), fmt.Sprintf("undefined identifier %q", e.Name), hint)
		return typeError
	case *ast.LiteralExpr:
		switch e.Kind {
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
		case ast.LiteralNull:
			return "null"
		case ast.LiteralHex:
			return "int"
		case ast.LiteralBinary:
			return "int"
		}
		return typeInfer
	case *ast.AwaitExpr:
		// Check if we're in an async function
		ctx := c.currentFunctionContext()
		if ctx == nil || !ctx.IsAsync {
			c.report(e.Span(), "await can only be used in async functions", "mark the function with the 'async' keyword")
			return typeError
		}

		operandType := c.checkExpr(e.Expr)
		// await unwraps Promise<T> to T
		if innerType, ok := promiseInnerType(operandType); ok {
			return innerType
		}
		if operandType != typeError {
			c.report(e.Expr.Span(), fmt.Sprintf("await can only be used on Promise types, got %s", operandType), "use await on a Promise value or async function call")
		}
		return typeError
	case *ast.UnaryExpr:
		operand := c.checkExpr(e.Expr)
		switch e.Op {
		case "!":
			if operand != typeError && !c.typesEqual(operand, "bool") {
				c.report(e.Expr.Span(), fmt.Sprintf("operator ! not defined on %s", operand), "use a boolean expression")
			}
			return "bool"
		case "-":
			if operand != typeError && !isNumeric(operand) {
				c.report(e.Expr.Span(), fmt.Sprintf("operator - not defined on %s", operand), "use a numeric expression")
			}
			return operand
		case "~":
			if operand != typeError && !isNumeric(operand) {
				c.report(e.Expr.Span(), fmt.Sprintf("operator ~ not defined on %s", operand), "use an integer expression")
			}
			return operand
		default:
			c.report(e.Span(), fmt.Sprintf("unsupported unary operator %q", e.Op), "remove the operator or extend the checker")
			return typeError
		}
	case *ast.BinaryExpr:
		return c.checkBinaryExpr(e)
	case *ast.CallExpr:
		return c.checkCallExpr(e)
	case *ast.IndexExpr:
		targetType := c.checkExpr(e.Target)
		indexType := c.checkExpr(e.Index)
		if elem, ok := arrayElementType(targetType); ok {
			if indexType != typeError && !c.typesEqual(indexType, "int") {
				c.report(e.Index.Span(), fmt.Sprintf("array index must be int, got %s", indexType), "use an integer index")
			}
			return elem
		}
		if keyType, valueType, ok := mapTypes(targetType); ok {
			if indexType != typeError && !c.typesEqual(indexType, keyType) {
				c.report(e.Index.Span(), fmt.Sprintf("map key expects %s, got %s", keyType, indexType), "adjust the key expression")
			}
			return valueType
		}
		if targetType != typeError {
			c.report(e.Target.Span(), fmt.Sprintf("type %s does not support indexing", targetType), "use an array or map expression")
		}
		return typeError
	case *ast.MemberExpr:
		targetType := c.checkExpr(e.Target)

		// Handle array method access (e.g., x.len)
		if strings.HasPrefix(targetType, "[]<") || strings.HasPrefix(targetType, "array<") {
			if e.Member == "len" {
				// Return a function type that takes no arguments and returns int
				return "func():int"
			}
			c.report(e.Span(), fmt.Sprintf("array type %s has no method %q", targetType, e.Member),
				"available methods: len")
			return typeError
		}

		if structName, fields, ok := c.resolveStructDefinition(targetType); ok {
			var typeArgs []string
			if _, args := c.extractGenericType(targetType); len(args) > 0 {
				typeArgs = args
			}
			resolvedFields := c.applyStructTypeArguments(structName, fields, typeArgs, e.Target.Span())
			if fieldType, exists := resolvedFields[e.Member]; exists {
				return fieldType
			}
			c.report(e.Span(), fmt.Sprintf("struct %s has no field %q", targetType, e.Member), "use a declared field name")
			return typeError
		}

		// Handle static method calls on struct types (e.g., pkg.Type.staticMethod())
		// Check if targetType is a struct type and the member is a function
		if strings.Contains(targetType, ".") {
			// This might be a qualified type name (e.g., "pkg.Type")
			qualifiedName := targetType + "." + e.Member
			// Check if it's a function from the module/package
			if sig, exists := c.functions[qualifiedName]; exists {
				return sig.Return
			}
		}
		// Also check MemberExpr targets (e.g., pkg.Type.staticMethod where pkg.Type is a MemberExpr)
		if member, ok := e.Target.(*ast.MemberExpr); ok {
			if ident, ok := member.Target.(*ast.IdentifierExpr); ok {
				// Construct qualified name: pkg.Type.staticMethod
				qualifiedName := ident.Name + "." + member.Member + "." + e.Member
				if sig, exists := c.functions[qualifiedName]; exists {
					return sig.Return
				}
			}
		}

		// Handle module member access (e.g., math_utils.add, io.println)
		if targetType == "module" {
			// Check if the target is an imported module
			if ident, ok := e.Target.(*ast.IdentifierExpr); ok {
				qualifiedName := ident.Name + "." + e.Member

				// Check if it's a function from the imported module
				if sig, exists := c.functions[qualifiedName]; exists {
					return sig.Return
				}

				// Check if it's a std function with alias (e.g., io.println -> std.io.println)
				if c.isStdSymbol("std." + qualifiedName) {
					name := "std." + qualifiedName
					if strings.Contains(name, "io.") {
						return "void"
					}
					if strings.Contains(name, "math.") {
						return "int"
					}
					if strings.Contains(name, "string.") {
						if strings.Contains(name, "length") {
							return "int"
						}
						if strings.Contains(name, "concat") {
							return "string"
						}
						return "string"
					}
					return "void"
				}

				// Check if it's an aliased std import (e.g., str.concat -> std.string.concat)
				if c.isAliasedStdSymbol(qualifiedName) {
					fullName := c.mapAliasToStd(qualifiedName)
					if strings.Contains(fullName, "io.") {
						return "void"
					}
					if strings.Contains(fullName, "math.") {
						return "int"
					}
					if strings.Contains(fullName, "string.") {
						if strings.HasSuffix(fullName, ".length") || strings.HasSuffix(fullName, ".index_of") ||
							strings.HasSuffix(fullName, ".last_index_of") || strings.HasSuffix(fullName, ".compare") {
							return "int"
						}
						for _, boolFn := range []string{"starts_with", "ends_with", "contains", "equals", "is_blank", "is_alpha", "is_digit", "is_alnum", "is_ascii", "is_upper", "is_lower", "matches"} {
							if strings.HasSuffix(fullName, "."+boolFn) {
								return "bool"
							}
						}
						return "string"
					}
					return "void"
				}

				c.report(e.Span(), fmt.Sprintf("module %s has no function %q", ident.Name, e.Member), "check the function name and module import")
				return typeError
			}
		}

		if targetType != typeError {
			c.report(e.Span(), fmt.Sprintf("type %s has no members", targetType), "use a struct value for member access")
		}
		return typeError
	case *ast.ArrayLiteralExpr:
		elementType := typeInfer
		for _, el := range e.Elements {
			t := c.checkExpr(el)
			if elementType == typeInfer {
				elementType = t
			} else if t != typeInfer && t != typeError && !c.typesEqual(elementType, t) {
				c.report(el.Span(), "array literal elements must have the same type", "ensure all elements share a common type")
				elementType = typeError
			}
		}
		if elementType == typeInfer {
			// Empty array literal - cannot infer element type
			c.report(e.Span(), "cannot infer element type of empty array literal",
				"add a type annotation like let arr: array<int> = [] or provide at least one element")
			elementType = typeError
		}
		return buildGeneric("[]", []string{elementType})
	case *ast.MapLiteralExpr:
		// Check if we have an expected type context (from checkLet or checkBindingStmt)
		// This is set when checking map literals with type annotations
		expectedValueTypeIsAny := false
		expectedMapTypeStr := ""
		// Capture expectedMapType at the start to avoid issues with recursive calls
		// Also check if we're in a let/var context by looking at the current scope
		if c.expectedMapType != "" {
			expectedMapTypeStr = c.expectedMapType
		}
		// Extract key and value types from expected map type if we have one
		if expectedMapTypeStr != "" && strings.HasPrefix(expectedMapTypeStr, "map<") && strings.HasSuffix(expectedMapTypeStr, ">") {
			inner := expectedMapTypeStr[4 : len(expectedMapTypeStr)-1]
			parts := strings.Split(inner, ",")
			if len(parts) == 2 {
				expectedKeyType := strings.TrimSpace(parts[0])
				expectedValueType := strings.TrimSpace(parts[1])

				// If expected value type is "any", we can accept any value types
				if expectedValueType == "any" {
					expectedValueTypeIsAny = true
					// Check that all keys match expected key type (but don't check value types)
					allKeysMatch := true
					for _, entry := range e.Entries {
						keyType := c.checkExpr(entry.Key)
						// Allow typeInfer keys (they'll be resolved later) or matching types
						if keyType != typeInfer && keyType != typeError && !c.isAssignable(keyType, expectedKeyType) {
							allKeysMatch = false
							break
						}
						// If keyType is typeInfer, we can't verify yet, but allow it
						// (it will be resolved during normal type checking)
						// Note: We intentionally don't check value types here since value type is "any"
					}
					// Even if allKeysMatch is false, we still want to allow any value types
					// So we keep expectedValueTypeIsAny = true and fall through to normal inference
					// which will use expectedValueTypeIsAny to skip value type consistency checks
					if allKeysMatch {
						// All keys match (or are inferrable), use expected type
						// Clear the expected type context before returning
						c.expectedMapType = ""
						return expectedMapTypeStr
					}
					// If allKeysMatch is false, we'll fall through and use expectedValueTypeIsAny
					// to allow mixed value types in the normal inference loop
				}
			}
		}

		// Normal type inference from values
		// If expected value type was "any", skip value type consistency check
		// Re-check expectedMapTypeStr if it's empty but c.expectedMapType is set
		// (this can happen if c.expectedMapType was set after we captured expectedMapTypeStr)
		if expectedMapTypeStr == "" && c.expectedMapType != "" {
			expectedMapTypeStr = c.expectedMapType
			// Re-parse to set expectedValueTypeIsAny if needed
			if strings.HasPrefix(expectedMapTypeStr, "map<") && strings.HasSuffix(expectedMapTypeStr, ">") {
				inner := expectedMapTypeStr[4 : len(expectedMapTypeStr)-1]
				parts := strings.Split(inner, ",")
				if len(parts) == 2 {
					expectedValueType := strings.TrimSpace(parts[1])
					if expectedValueType == "any" {
						expectedValueTypeIsAny = true
					}
				}
			}
		}

		keyType := typeInfer
		valueType := typeInfer
		for _, entry := range e.Entries {
			kt := c.checkExpr(entry.Key)
			vt := c.checkExpr(entry.Value)
			if keyType == typeInfer {
				keyType = kt
			} else if kt != typeInfer && kt != typeError && !c.typesEqual(keyType, kt) {
				c.report(entry.Key.Span(), "map literal keys must share the same type", "adjust the map keys to match")
				keyType = typeError
			}

			if !expectedValueTypeIsAny {
				// Normal value type consistency check (only when not allowing any values)
				if valueType == typeInfer {
					valueType = vt
				} else if vt != typeInfer && vt != typeError && !c.typesEqual(valueType, vt) {
					c.report(entry.Value.Span(), "map literal values must share the same type", "adjust the map values to match")
					valueType = typeError
				}
			}
			// If expectedValueTypeIsAny is true, we skip value type checking entirely
			// The final return will use expectedMapTypeStr which has "any" as the value type
		}

		// For empty map literals, return a special placeholder that will be resolved in checkCallExpr
		if keyType == typeInfer && valueType == typeInfer {
			// Return a placeholder that checkCallExpr can recognize and replace
			return "<empty_map>"
		}
		if keyType == typeInfer {
			keyType = typeError
		}
		if valueType == typeInfer {
			valueType = typeError
		}
		// If expected value type was "any", return the expected type instead of inferred type
		if expectedValueTypeIsAny && expectedMapTypeStr != "" {
			// We had an expected type with "any" value type, use it
			// Clear the expected type context if it's still set
			if c.expectedMapType != "" {
				c.expectedMapType = ""
			}
			return expectedMapTypeStr
		}
		// Clear expectedMapType after we're done with it
		if c.expectedMapType != "" {
			c.expectedMapType = ""
		}
		return buildGeneric("map", []string{keyType, valueType})
	case *ast.StructLiteralExpr:
		structName, fields, ok := c.resolveStructDefinition(e.TypeName)
		if !ok {
			c.report(e.Span(), fmt.Sprintf("unknown struct type %q", e.TypeName), "define the struct before constructing it")
			return typeError
		}
		typeArgs := make([]string, 0, len(e.TypeArgs))
		for _, arg := range e.TypeArgs {
			typeArgs = append(typeArgs, typeExprToString(arg))
		}
		resolvedFields := c.applyStructTypeArguments(structName, fields, typeArgs, e.Span())
		for _, field := range e.Fields {
			expectedType, exists := resolvedFields[field.Name]
			if !exists {
				c.report(field.Span, fmt.Sprintf("struct %s has no field %q", e.TypeName, field.Name), "use a declared field name")
				c.checkExpr(field.Expr)
				continue
			}
			actualType := c.checkExpr(field.Expr)
			if actualType != typeError && !c.typesEqual(expectedType, actualType) {
				c.report(field.Expr.Span(), fmt.Sprintf("field %s expects %s, got %s", field.Name, expectedType, actualType), "adjust the field expression")
			}
		}
		if len(typeArgs) == 0 {
			return e.TypeName
		}
		return buildGeneric(e.TypeName, typeArgs)
	case *ast.AssignmentExpr:
		return c.checkAssignmentExpr(e)
	case *ast.IncrementExpr:
		targetType := c.checkExpr(e.Target)
		if !isNumeric(targetType) {
			c.report(e.Target.Span(), fmt.Sprintf("operator %s not defined on %s", e.Op, targetType), "use a numeric variable")
		}
		if ident, ok := e.Target.(*ast.IdentifierExpr); ok {
			if sym, found := c.lookupSymbol(ident.Name); found && !sym.Mutable {
				c.report(e.Target.Span(), fmt.Sprintf("cannot modify immutable variable %q", ident.Name), "declare it with var if mutation is required")
			}
		}
		return targetType
	case *ast.NewExpr:
		// new Type returns a pointer to Type
		typ := c.checkTypeExpr(e.Type)
		return "*" + typ
	case *ast.DeleteExpr:
		// delete expression returns void
		// Validate that the target is a pointer type
		targetType := c.checkExpr(e.Target)
		// Check if target is a pointer type (starts with *)
		if !strings.HasPrefix(targetType, "*") {
			c.report(e.Target.Span(), fmt.Sprintf("delete operand must be a pointer, got %s", targetType),
				"delete can only be used on pointer types (values returned from new)")
			return typeError
		}
		return "void"
	case *ast.LambdaExpr:
		// Lambda expression: |a, b| a + b
		// Use typeInfer for parameters - will be inferred from context or usage
		paramTypes := make([]string, len(e.Params))
		for i := range e.Params {
			paramTypes[i] = typeInfer
		}
		return c.checkLambdaWithTypes(e, paramTypes)
	case *ast.CastExpr:
		// Type cast expression: (type) expression
		targetType := c.checkTypeExpr(e.Type)
		if targetType == typeError {
			return typeError
		}

		// Check the expression being cast
		exprType := c.checkExpr(e.Expr)
		if exprType == typeError {
			return typeError
		}

		if canCastBetweenTypes(exprType, targetType) {
			return targetType
		}

		c.report(e.Span(), fmt.Sprintf("cannot cast %s to %s", exprType, targetType),
			"use an explicit conversion helper or adjust the expression type")
		return typeError
	case *ast.StringInterpolationExpr:
		// String interpolation always returns string type
		// Check each part to ensure expressions are valid
		for _, part := range e.Parts {
			if !part.IsLiteral {
				// Check the expression part
				c.checkExpr(part.Expr)
			}
		}
		return "string"
	default:
		return typeInfer
	}
}

func (c *Checker) checkBinaryExpr(expr *ast.BinaryExpr) string {
	leftType := c.checkExpr(expr.Left)
	rightType := c.checkExpr(expr.Right)
	if leftType == typeError || rightType == typeError {
		return typeError
	}

	switch expr.Op {
	case "+":
		// Handle string concatenation when the left operand is a string
		if leftType == "string" {
			return "string"
		}
		if rightType == "string" {
			c.report(expr.Span(), "string concatenation requires a string on the left-hand side",
				fmt.Sprintf("convert %s to string before adding", leftType))
			return typeError
		}
		// Handle numeric addition
		if !isNumeric(leftType) || !isNumeric(rightType) {
			c.report(expr.Span(), fmt.Sprintf("operator %s requires numeric or string operands", expr.Op),
				fmt.Sprintf("use numeric expressions (int, float) or strings, got %s and %s", leftType, rightType))
			return typeError
		}
		if !c.typesEqual(leftType, rightType) {
			c.report(expr.Span(), fmt.Sprintf("operands of %s must have the same type", expr.Op), "convert one side to match the other")
			return typeError
		}
		return leftType
	case "-", "*", "/", "%":
		if !isNumeric(leftType) || !isNumeric(rightType) {
			c.report(expr.Span(), fmt.Sprintf("operator %s requires numeric operands", expr.Op),
				fmt.Sprintf("use numeric expressions (int, float), got %s and %s", leftType, rightType))
			return typeError
		}
		if !c.typesEqual(leftType, rightType) {
			c.report(expr.Span(), fmt.Sprintf("operands of %s must have the same type", expr.Op), "convert one side to match the other")
			return typeError
		}
		return leftType
	case "<", "<=", ">", ">=":
		// Allow comparison if both operands are the same type (including generic type parameters)
		if !c.typesEqual(leftType, rightType) {
			c.report(expr.Span(), fmt.Sprintf("operands of %s must have the same type", expr.Op),
				fmt.Sprintf("ensure both sides have the same type, got %s and %s", leftType, rightType))
			return typeError
		}
		return "bool"
	case "==", "!=":
		if !c.typesEqual(leftType, rightType) {
			c.report(expr.Span(), fmt.Sprintf("operands of %s must be comparable", expr.Op), "ensure both sides share the same type")
		}
		return "bool"
	case "&&", "||":
		if !c.typesEqual(leftType, "bool") || !c.typesEqual(rightType, "bool") {
			c.report(expr.Span(), fmt.Sprintf("operator %s requires boolean operands", expr.Op),
				fmt.Sprintf("use boolean expressions, got %s and %s", leftType, rightType))
		}
		return "bool"
	case "&", "|", "^", "<<", ">>":
		// Bitwise operators require integer operands (not floats)
		if !isInteger(leftType) || !isInteger(rightType) {
			c.report(expr.Span(), fmt.Sprintf("operator %s requires integer operands", expr.Op),
				fmt.Sprintf("use integer expressions (int, long, byte), got %s and %s", leftType, rightType))
			return typeError
		}
		if !c.typesEqual(leftType, rightType) {
			c.report(expr.Span(), fmt.Sprintf("operands of %s must have the same type", expr.Op), "convert one side to match the other")
			return typeError
		}
		return leftType
	default:
		c.report(expr.Span(), fmt.Sprintf("unsupported binary operator %q", expr.Op), "remove the operator or extend the checker")
		return typeError
	}
}

func (c *Checker) checkCallExpr(expr *ast.CallExpr) string {
	var calleeType string
	if expr.Callee != nil {
		calleeType = c.checkExpr(expr.Callee)
	}

	// Check if this is a regular function call (not a function type call)
	if ident, ok := expr.Callee.(*ast.IdentifierExpr); ok {
		// Check if it's a regular function (not a function type variable)
		if sig, exists := c.functions[ident.Name]; exists {
			// Check if this is a generic function
			if len(sig.TypeParams) > 0 {
				return c.checkGenericFunctionCall(expr, sig, ident.Name)
			}

			// This is a regular function call, not a function type call
			// Validate argument count
			if len(expr.Args) != len(sig.Params) {
				c.report(expr.Span(), fmt.Sprintf("argument count mismatch: function %s expects %d arguments, got %d", ident.Name, len(sig.Params), len(expr.Args)),
					fmt.Sprintf("provide %d argument(s) matching the function signature", len(sig.Params)))
				return typeError
			}

			// Validate argument types
			for i, arg := range expr.Args {
				var argType string
				expected := typeInfer
				if i < len(sig.Params) {
					expected = sig.Params[i]
				}

				// Special handling for lambda expressions - infer parameter types from expected function type
				if lambda, ok := arg.(*ast.LambdaExpr); ok && expected != typeInfer {
					// If expected type is a function type, use it to infer lambda parameter types
					if strings.Contains(expected, ") -> ") {
						expectedParamTypes := c.parseFunctionTypeParams(expected)
						if expectedParamTypes != nil && len(expectedParamTypes) == len(lambda.Params) {
							argType = c.checkLambdaWithTypes(lambda, expectedParamTypes)
						} else {
							// Parameter count mismatch - check normally
							argType = c.checkExpr(arg)
						}
					} else {
						// Not a function type or no expected type - check normally
						argType = c.checkExpr(arg)
					}
				} else if mapLit, ok := arg.(*ast.MapLiteralExpr); ok && expected != typeInfer {
					// Special handling for empty map literals - use expected type if it's a map
					if len(mapLit.Entries) == 0 && strings.HasPrefix(expected, "map<") {
						// Extract key and value types from expected map type
						// Expected format: map<keyType, valueType> or map<keyType,valueType>
						inner := expected[4 : len(expected)-1] // Remove "map<" and ">"
						parts := strings.Split(inner, ",")
						if len(parts) == 2 {
							keyType := strings.TrimSpace(parts[0])
							valueType := strings.TrimSpace(parts[1])
							// Normalize valueType to handle "any" vs " any" (with/without space)
							valueType = strings.TrimSpace(valueType)
							argType = buildGeneric("map", []string{keyType, valueType})
						} else {
							argType = c.checkExpr(arg)
						}
					} else {
						argType = c.checkExpr(arg)
					}
				} else {
					argType = c.checkExpr(arg)
					// If we got the empty map placeholder, try to infer from expected type
					if argType == "<empty_map>" && expected != typeInfer {
						if strings.HasPrefix(expected, "map<") {
							inner := expected[4 : len(expected)-1]
							parts := strings.Split(inner, ",")
							if len(parts) == 2 {
								keyType := strings.TrimSpace(parts[0])
								valueType := strings.TrimSpace(parts[1])
								valueType = strings.TrimSpace(valueType)
								argType = buildGeneric("map", []string{keyType, valueType})
							}
						} else if expected == "any" {
							// If expected type is "any", we can use any map type
							// Use map<string, any> as a reasonable default
							argType = buildGeneric("map", []string{"string", "any"})
						}
					}
				}

				// Skip type checking if we resolved an empty map to match expected type
				if argType == "<empty_map>" {
					// This shouldn't happen here, but if it does, it's an error
					c.report(arg.Span(), fmt.Sprintf("cannot infer type for empty map literal in argument %d", i+1),
						"provide a type annotation or use a map with at least one entry")
				} else if expected != typeInfer && argType != typeError && !c.isAssignable(argType, expected) {
					c.report(arg.Span(), fmt.Sprintf("argument type mismatch: argument %d expects %s, got %s", i+1, expected, argType),
						fmt.Sprintf("convert the argument to %s or use a %s expression", expected, expected))
				}
			}

			return sig.Return
		}
	}

	// Handle function type calls (e.g., (int, string) -> bool)
	// This handles cases where the callee is a function type variable or expression
	// Check if we already handled this as a direct function call above
	handledAsDirectCall := false
	if ident, ok := expr.Callee.(*ast.IdentifierExpr); ok {
		if _, exists := c.functions[ident.Name]; exists {
			handledAsDirectCall = true
		}
	}

	// Also check MemberExpr for qualified function names
	if !handledAsDirectCall {
		if member, ok := expr.Callee.(*ast.MemberExpr); ok {
			if ident, ok := member.Target.(*ast.IdentifierExpr); ok {
				qualifiedName := ident.Name + "." + member.Member
				if _, exists := c.functions[qualifiedName]; exists {
					handledAsDirectCall = true
				}
			}
		}
	}

	// If calleeType is a function type, validate arguments
	if strings.Contains(calleeType, ") -> ") && !handledAsDirectCall {
		// Parse function type: (param1, param2) -> returnType
		arrowIndex := strings.Index(calleeType, ") -> ")
		if arrowIndex != -1 {
			paramPart := calleeType[1:arrowIndex]                      // Remove opening (
			returnType := strings.TrimSpace(calleeType[arrowIndex+5:]) // After " -> "

			// Parse parameter types
			var expectedParamTypes []string
			if paramPart != "" {
				// Split by comma and trim spaces
				paramStrs := strings.Split(paramPart, ",")
				for _, paramStr := range paramStrs {
					expectedParamTypes = append(expectedParamTypes, strings.TrimSpace(paramStr))
				}
			}

			// Validate argument count
			if len(expr.Args) != len(expectedParamTypes) {
				c.report(expr.Span(), fmt.Sprintf("argument count mismatch: function expects %d arguments, got %d", len(expectedParamTypes), len(expr.Args)),
					fmt.Sprintf("provide %d argument(s) matching the function signature", len(expectedParamTypes)))
				return typeError
			}

			// Validate argument types
			for i, arg := range expr.Args {
				var argType string
				// Special handling for lambda expressions - infer parameter types from expected function type
				if lambda, ok := arg.(*ast.LambdaExpr); ok && i < len(expectedParamTypes) {
					expected := expectedParamTypes[i]
					// If expected type is a function type, use it to infer lambda parameter types
					if expected != typeInfer && strings.Contains(expected, ") -> ") {
						expectedParamTypes := c.parseFunctionTypeParams(expected)
						if expectedParamTypes != nil && len(expectedParamTypes) == len(lambda.Params) {
							argType = c.checkLambdaWithTypes(lambda, expectedParamTypes)
						} else {
							// Parameter count mismatch - check normally
							argType = c.checkExpr(arg)
						}
					} else {
						// Not a function type or no expected type - check normally
						argType = c.checkExpr(arg)
					}
				} else {
					argType = c.checkExpr(arg)
				}
				if i < len(expectedParamTypes) {
					expected := expectedParamTypes[i]
					if expected != typeInfer && argType != typeError && !c.typesEqual(expected, argType) {
						c.report(arg.Span(), fmt.Sprintf("argument type mismatch: argument %d expects %s, got %s", i+1, expected, argType),
							fmt.Sprintf("convert the argument to %s or use a %s expression", expected, expected))
					}
				}
			}

			return returnType
		}
	}

	// Handle function calls with qualified names (e.g., math_utils.add, std.math.max)
	var qualifiedName string
	if ident, ok := expr.Callee.(*ast.IdentifierExpr); ok {
		qualifiedName = ident.Name
	} else if member, ok := expr.Callee.(*ast.MemberExpr); ok {
		// Handle array method calls like x.len() where x is an array
		targetType := c.checkExpr(member.Target)
		if strings.HasPrefix(targetType, "[]<") || strings.HasPrefix(targetType, "array<") {
			// This is an array method call
			if member.Member == "len" && len(expr.Args) == 0 {
				// x.len() - array length method
				return "int"
			}
			c.report(expr.Span(), fmt.Sprintf("array type %s has no method %q", targetType, member.Member),
				"available methods: len()")
			return typeError
		}

		// Handle qualified function calls like module.function
		if memberIdent, ok := member.Target.(*ast.IdentifierExpr); ok {
			qualifiedName = memberIdent.Name + "." + member.Member
		}
	}

	if qualifiedName != "" {
		if sig, exists := c.functions[qualifiedName]; exists {
			if len(expr.Args) != len(sig.Params) {
				c.report(expr.Span(), fmt.Sprintf("argument count mismatch: function %s expects %d arguments, got %d", qualifiedName, len(sig.Params), len(expr.Args)),
					fmt.Sprintf("provide %d argument(s) matching the function signature: %s(%s)", len(sig.Params), qualifiedName, strings.Join(sig.Params, ", ")))
			}
			for i, arg := range expr.Args {
				argType := c.checkExpr(arg)
				if i < len(sig.Params) {
					expected := sig.Params[i]
					// Special handling for len() function - accept any array type
					if qualifiedName == "len" && expected == typeInfer {
						if !strings.HasPrefix(argType, "[]<") && !strings.HasPrefix(argType, "array<") {
							c.report(arg.Span(), fmt.Sprintf("len() expects an array, got %s", argType),
								"pass an array to the len() function")
						}
					} else if expected != typeInfer && argType != typeError {
						// Resolve type aliases before checking equality
						resolvedExpected := expected
						if underlying, isAlias := c.typeAliases[expected]; isAlias {
							resolvedExpected = underlying
						}
						if !c.isAssignable(argType, resolvedExpected) {
							c.report(arg.Span(), fmt.Sprintf("argument type mismatch: argument %d expects %s, got %s", i+1, expected, argType),
								fmt.Sprintf("convert the argument to %s or use a %s expression", expected, expected))
						}
					}
				}
			}
			return sig.Return
		}
	}

	for _, arg := range expr.Args {
		c.checkExpr(arg)
	}
	return calleeType
}

// checkGenericFunctionCall handles calls to generic functions
func (c *Checker) checkGenericFunctionCall(expr *ast.CallExpr, sig FunctionSignature, funcName string) string {
	// First, check argument count
	if len(expr.Args) != len(sig.Params) {
		c.report(expr.Span(), fmt.Sprintf("argument count mismatch: function %s expects %d arguments, got %d", funcName, len(sig.Params), len(expr.Args)),
			fmt.Sprintf("provide %d argument(s) matching the function signature", len(sig.Params)))
		return typeError
	}

	// Infer type parameters from arguments
	typeSubstitutions := make(map[string]string)

	// Check each argument and infer type parameters
	for i, arg := range expr.Args {
		argType := c.checkExpr(arg)
		if i < len(sig.Params) {
			expected := sig.Params[i]

			// If the expected type is a type parameter of this function, infer it from the argument
			if c.isFunctionTypeParam(expected, sig.TypeParams) {
				if existing, exists := typeSubstitutions[expected]; exists {
					// Check if the inferred type matches the previous inference
					if !c.typesEqual(existing, argType) {
						c.report(arg.Span(), fmt.Sprintf("type parameter %s inferred as both %s and %s", expected, existing, argType),
							"ensure all arguments for this type parameter have the same type")
						return typeError
					}
				} else {
					typeSubstitutions[expected] = argType
				}
			} else if expected != typeInfer && argType != typeError && !c.typesEqual(expected, argType) {
				// Try to infer type parameters from generic types like array<T>
				inferred := c.inferTypeParametersFromGeneric(expected, argType, sig.TypeParams)
				for typeParam, concreteType := range inferred {
					if existing, exists := typeSubstitutions[typeParam]; exists {
						if !c.typesEqual(existing, concreteType) {
							c.report(arg.Span(), fmt.Sprintf("type parameter %s inferred as both %s and %s", typeParam, existing, concreteType),
								"ensure all arguments for this type parameter have the same type")
							return typeError
						}
					} else {
						typeSubstitutions[typeParam] = concreteType
					}
				}

				// Check if the expected type contains type parameters that need substitution
				substitutedExpected := expected
				for typeParam, concreteType := range typeSubstitutions {
					substitutedExpected = c.substituteTypeParam(substitutedExpected, typeParam, concreteType)
				}
				if !c.typesEqual(substitutedExpected, argType) {
					c.report(arg.Span(), fmt.Sprintf("argument type mismatch: argument %d expects %s, got %s", i+1, substitutedExpected, argType),
						fmt.Sprintf("convert the argument to %s or use a %s expression", substitutedExpected, substitutedExpected))
				}
			}
		}
	}

	// Apply type substitutions to return type
	returnType := sig.Return
	for typeParam, concreteType := range typeSubstitutions {
		returnType = c.substituteTypeParam(returnType, typeParam, concreteType)
	}

	return returnType
}

func (c *Checker) checkAssignmentExpr(expr *ast.AssignmentExpr) string {
	// Allow indexed assignment (arr[i] = v) and member assignment (obj.field = v).
	if idx, ok := expr.Left.(*ast.IndexExpr); ok {
		targetType := c.checkExpr(idx.Target)
		c.checkExpr(idx.Index)
		rhs := c.checkExpr(expr.Right)
		if targetType == typeError || rhs == typeError {
			return typeError
		}
		return rhs
	}
	if m, ok := expr.Left.(*ast.MemberExpr); ok {
		c.checkExpr(m)
		return c.checkExpr(expr.Right)
	}

	ident, ok := expr.Left.(*ast.IdentifierExpr)
	if !ok {
		c.report(expr.Left.Span(), "left-hand side of assignment must be an identifier", "assign to a named variable")
		c.checkExpr(expr.Right)
		return typeError
	}

	sym, exists := c.lookupSymbol(ident.Name)
	if !exists {
		c.report(expr.Left.Span(), fmt.Sprintf("undefined identifier %q", ident.Name),
			fmt.Sprintf("declare the variable with 'let %s: <type> = <value>' or 'var %s: <type> = <value>' before assignment", ident.Name, ident.Name))
		c.checkExpr(expr.Right)
		return typeError
	}
	if !sym.Mutable {
		c.report(expr.Left.Span(), fmt.Sprintf("cannot assign to immutable variable %q", ident.Name), "declare it with var if mutation is required")
	}

	rhsType := c.checkExpr(expr.Right)
	if sym.Type == typeInfer && rhsType != typeError {
		c.updateSymbolType(ident.Name, rhsType)
		sym.Type = rhsType
	}
	if rhsType != typeError && sym.Type != typeInfer && !c.isAssignable(rhsType, sym.Type) {
		c.report(expr.Right.Span(), fmt.Sprintf("type mismatch: cannot assign %s to %s", rhsType, sym.Type),
			fmt.Sprintf("convert the expression to %s or change the variable type to %s", sym.Type, rhsType))
	}
	return sym.Type
}

// -----------------------------------------------------------------------------
// Type helpers
// -----------------------------------------------------------------------------

func (c *Checker) resolveTypeExpr(t *ast.TypeExpr) string {
	if t == nil {
		return typeInfer
	}

	// Handle union types
	if t.IsUnion {
		members := make([]string, len(t.Members))
		for i, member := range t.Members {
			members[i] = c.resolveTypeExpr(member)
		}
		return buildUnion(members)
	}

	// Check if this is a type parameter
	if c.isTypeParam(t.Name) {
		return t.Name
	}

	return c.typeExprToString(t)
}

func (c *Checker) typeExprToString(t *ast.TypeExpr) string {
	if t == nil {
		return typeInfer
	}

	// Handle function types: (param1, param2) -> returnType
	if t.IsFunction {
		paramTypes := make([]string, len(t.ParamTypes))
		for i, paramType := range t.ParamTypes {
			paramTypes[i] = c.typeExprToString(paramType)
		}
		returnType := c.typeExprToString(t.ReturnType)
		return buildFunctionType(paramTypes, returnType)
	}

	// Handle union types
	if t.IsUnion {
		members := make([]string, len(t.Members))
		for i, member := range t.Members {
			members[i] = c.typeExprToString(member)
		}
		return buildUnion(members)
	}

	// Handle pointer types
	if strings.HasPrefix(t.Name, "*") {
		baseType := t.Name[1:] // Remove the *
		if len(t.Args) > 0 {
			// This is a pointer to a complex type (union or generic)
			args := make([]string, len(t.Args))
			for i, arg := range t.Args {
				args[i] = c.typeExprToString(arg)
			}
			// Check if the base type is empty and we have union args
			if baseType == "" && len(args) > 1 {
				// This is a pointer to a union type
				return "*" + buildUnion(args)
			}
			return "*" + buildGeneric(baseType, args)
		}
		return "*" + baseType
	}

	// Handle optional types: T?
	if t.IsOptional && t.OptionalType != nil {
		innerType := c.typeExprToString(t.OptionalType)
		return innerType + "?"
	}

	// Handle generic types
	if len(t.Args) > 0 {
		args := make([]string, len(t.Args))
		for i, arg := range t.Args {
			args[i] = c.typeExprToString(arg)
		}
		return buildGeneric(t.Name, args)
	}

	return t.Name
}

func (c *Checker) checkTypeExpr(t *ast.TypeExpr) string {
	if t == nil {
		return typeInfer
	}

	// Handle function types
	if t.IsFunction {
		paramTypes := make([]string, len(t.ParamTypes))
		for i, paramType := range t.ParamTypes {
			paramTypes[i] = c.checkTypeExpr(paramType)
		}
		returnType := c.checkTypeExpr(t.ReturnType)
		return buildFunctionType(paramTypes, returnType)
	}

	// Handle union types
	if t.IsUnion {
		members := make([]string, len(t.Members))
		for i, member := range t.Members {
			// Recursively check member types, but don't report errors for qualified types
			// that are part of imported modules (they're already validated)
			memberType := c.checkTypeExpr(member)
			members[i] = memberType
		}
		return buildUnion(members)
	}

	// Handle optional types
	if t.IsOptional {
		innerType := c.checkTypeExpr(t.OptionalType)
		return innerType + "?" // For now, just append ? to indicate optional
	}

	// Handle pointer types
	if strings.HasPrefix(t.Name, "*") {
		baseType := t.Name[1:] // Remove the *
		if len(t.Args) > 0 {
			// This is a pointer to a complex type (union or generic)
			args := make([]string, len(t.Args))
			for i, arg := range t.Args {
				args[i] = c.checkTypeExpr(arg)
			}
			// Check if the base type is empty and we have union args
			if baseType == "" && len(args) > 1 {
				// This is a pointer to a union type
				return "*" + buildUnion(args)
			}
			return "*" + buildGeneric(baseType, args)
		}
		// Check if the base type is known
		if _, ok := c.knownTypes[baseType]; !ok {
			c.report(t.Span(), fmt.Sprintf("unknown type %q", baseType), "import or declare the type before use")
		}
		return "*" + baseType
	}

	// Check if this is an array type
	if t.Name == "[]" {
		if len(t.Args) != 1 {
			c.report(t.Span(), "array type must have exactly one element type", "use syntax like []int or []string")
			return typeError
		}
		elementType := c.checkTypeExpr(t.Args[0])
		return buildGeneric("[]", []string{elementType})
	}

	// Check if this is a type parameter
	if c.isTypeParam(t.Name) {
		return t.Name
	}

	// Handle type aliases - resolve to underlying type
	// For generic aliases, substitute type arguments
	if underlyingType, isAlias := c.typeAliases[t.Name]; isAlias {
		// Check if this is a generic alias with type arguments
		if len(t.Args) > 0 {
			// Get type parameters for this alias
			if typeParams, hasParams := c.structTypeParams[t.Name]; hasParams {
				// Substitute type parameters in the underlying type
				substituted := underlyingType
				for i, param := range typeParams {
					if i < len(t.Args) {
						argType := c.checkTypeExpr(t.Args[i])
						// Substitute the type parameter with the actual argument
						substituted = c.substituteTypeParam(substituted, param.Name, argType)
					}
				}
				return substituted
			}
		}
		// Non-generic alias or alias without arguments - return as-is
		// Don't check the underlying type here - it's already been validated
		return underlyingType
	}

	// Check if it's a known type
	// Skip error reporting for qualified types from imported modules (they're already validated)
	if _, ok := c.knownTypes[t.Name]; !ok {
		// Check if it's a type alias (might be registered with qualified name)
		if strings.Contains(t.Name, ".") {
			// Qualified type - might be a type alias, check that
			if _, isAlias := c.typeAliases[t.Name]; !isAlias {
				// Not a known type and not a type alias - but don't report error for qualified types
				// as they're already validated during import
			}
		} else {
			// Unqualified type - check if it might be a type alias that needs qualification
			// This can happen when checking struct fields from imported modules
			// Don't report error yet - it might be resolved later or be a type alias
			// Only report if we're sure it's not a type alias and not known
			// For now, skip error reporting to avoid false positives during import processing
		}
	}
	resolvedArgs := make([]string, 0, len(t.Args))
	for _, arg := range t.Args {
		resolvedArgs = append(resolvedArgs, c.checkTypeExpr(arg))
	}
	if len(resolvedArgs) == 0 {
		return t.Name
	}
	return buildGeneric(t.Name, resolvedArgs)
}

func typeExprToString(t *ast.TypeExpr) string {
	if t == nil {
		return typeInfer
	}

	// Handle function types
	if t.IsFunction {
		paramTypes := make([]string, len(t.ParamTypes))
		for i, paramType := range t.ParamTypes {
			paramTypes[i] = typeExprToString(paramType)
		}
		returnType := typeExprToString(t.ReturnType)
		return buildFunctionType(paramTypes, returnType)
	}

	// Handle optional types: T?
	if t.IsOptional && t.OptionalType != nil {
		innerType := typeExprToString(t.OptionalType)
		return innerType + "?"
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
	var b strings.Builder
	b.WriteString(name)
	b.WriteByte('<')
	for i, arg := range args {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(arg)
	}
	b.WriteByte('>')
	return b.String()
}

func buildUnion(members []string) string {
	// Canonicalize union by sorting members to ensure deterministic ordering
	// This makes int | string equal to string | int
	sorted := make([]string, len(members))
	copy(sorted, members)
	sort.Strings(sorted)

	var b strings.Builder
	for i, member := range sorted {
		if i > 0 {
			b.WriteString(" | ")
		}
		b.WriteString(member)
	}
	return b.String()
}

func buildFunctionType(paramTypes []string, returnType string) string {
	var b strings.Builder
	b.WriteByte('(')
	for i, paramType := range paramTypes {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(paramType)
	}
	b.WriteString(") -> ")
	b.WriteString(returnType)
	return b.String()
}

// parseFunctionTypeParams extracts parameter types from a function type string like "(int, int) -> int"
func (c *Checker) parseFunctionTypeParams(funcType string) []string {
	// Find the arrow
	arrowIndex := strings.Index(funcType, ") -> ")
	if arrowIndex == -1 {
		return nil
	}

	// Extract parameter part: "(int, int)" -> "int, int"
	paramPart := funcType[1:arrowIndex] // Remove opening (
	if paramPart == "" {
		return []string{} // No parameters
	}

	// Split by comma and trim
	parts := strings.Split(paramPart, ",")
	paramTypes := make([]string, len(parts))
	for i, part := range parts {
		paramTypes[i] = strings.TrimSpace(part)
	}
	return paramTypes
}

// checkLambdaWithTypes checks a lambda expression with given parameter types
func (c *Checker) checkLambdaWithTypes(e *ast.LambdaExpr, paramTypes []string) string {
	// Create a new scope for lambda parameters
	c.enterScope()

	// Use provided parameter types or typeInfer if not provided
	for i, param := range e.Params {
		paramType := typeInfer
		// If we have a provided type for this parameter, use it
		if i < len(paramTypes) && paramTypes[i] != "" && paramTypes[i] != typeInfer {
			paramType = paramTypes[i]
		}
		// Add parameter to the current scope with its type
		// This is critical - the type must be set correctly for the body to type-check
		c.declare(param.Name, paramType, false, param.Span)
	}

	// Check the lambda body to infer return type
	returnType := c.checkExpr(e.Body)
	if returnType == typeError {
		c.leaveScope()
		return typeError
	}

	// Clean up the lambda scope
	c.leaveScope()

	// Build final parameter types (use provided types or typeInfer)
	finalParamTypes := make([]string, len(e.Params))
	for i := range e.Params {
		if i < len(paramTypes) {
			finalParamTypes[i] = paramTypes[i]
		} else {
			finalParamTypes[i] = typeInfer
		}
	}

	return buildFunctionType(finalParamTypes, returnType)
}

// substituteTypeParam replaces type parameters in type strings with boundary awareness.
// This prevents substring replacement bugs (e.g., "T" in "Matrix" becoming "Main<int>rix").
// It only replaces type parameters that appear as standalone identifiers, not substrings.
func (c *Checker) substituteTypeParam(typeStr, typeParam, concreteType string) string {
	// Type parameters can appear in several contexts:
	// 1. Standalone: "T" -> "int"
	// 2. In generics: "array<T>" -> "array<int>"
	// 3. In unions: "T | string" -> "int | string"
	// 4. In function types: "(T) -> T" -> "(int) -> int"
	// We need to replace only when the type parameter is a complete identifier,
	// not when it's part of another identifier.

	// Simple approach: replace only when preceded/followed by non-identifier characters
	// or at start/end of string. This handles most cases correctly.
	result := typeStr
	paramLen := len(typeParam)

	// Find all occurrences and check boundaries
	for i := 0; i <= len(result)-paramLen; {
		pos := strings.Index(result[i:], typeParam)
		if pos == -1 {
			break
		}
		actualPos := i + pos

		// Check if this is a standalone identifier
		// It should be preceded by a non-identifier char (or start) and followed by a non-identifier char (or end)
		isStart := actualPos == 0
		isEnd := actualPos+paramLen == len(result)
		prevOK := isStart || !isIdentifierChar(rune(result[actualPos-1]))
		nextOK := isEnd || !isIdentifierChar(rune(result[actualPos+paramLen]))

		if prevOK && nextOK {
			// This is a standalone type parameter, replace it
			result = result[:actualPos] + concreteType + result[actualPos+paramLen:]
			i = actualPos + len(concreteType)
		} else {
			// Skip this occurrence, continue searching
			i = actualPos + 1
		}
	}

	return result
}

// isIdentifierChar checks if a rune is part of an identifier (letter, digit, underscore)
func isIdentifierChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}

// extractGenericType extracts the base name and type arguments from a generic type string.
// For example, "Box<int>" returns ("Box", ["int"]), "array<string>" returns ("array", ["string"]).
// Returns empty strings if the type is not generic.
func (c *Checker) extractGenericType(typeStr string) (baseName string, typeArgs []string) {
	// Find the first '<' that's not part of a qualified name
	ltIndex := strings.Index(typeStr, "<")
	if ltIndex == -1 {
		return "", nil
	}

	// Extract base name (everything before '<')
	baseName = strings.TrimSpace(typeStr[:ltIndex])

	// Handle qualified names (e.g., "foo.bar.Box<int>")
	if lastDot := strings.LastIndex(baseName, "."); lastDot >= 0 {
		baseName = baseName[lastDot+1:]
	}

	// Extract type arguments (everything between '<' and matching '>')
	gtIndex := strings.LastIndex(typeStr, ">")
	if gtIndex == -1 || gtIndex <= ltIndex {
		return "", nil
	}

	argsStr := strings.TrimSpace(typeStr[ltIndex+1 : gtIndex])
	if argsStr == "" {
		return baseName, []string{}
	}

	// Split by comma, handling nested generics
	typeArgs = c.splitGenericArgs(argsStr)
	return baseName, typeArgs
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

func (c *Checker) resolveStructDefinition(typeName string) (string, map[string]string, bool) {
	cleanName := strings.TrimSpace(typeName)
	if idx := strings.Index(cleanName, "<"); idx >= 0 {
		cleanName = cleanName[:idx]
	}
	if fields, ok := c.structFields[cleanName]; ok {
		return cleanName, fields, true
	}
	if lastDot := strings.LastIndex(cleanName, "."); lastDot >= 0 {
		unqualified := cleanName[lastDot+1:]
		if fields, ok := c.structFields[unqualified]; ok {
			return unqualified, fields, true
		}
	}
	return "", nil, false
}

func (c *Checker) applyStructTypeArguments(structName string, baseFields map[string]string, typeArgs []string, span lexer.Span) map[string]string {
	resolved := make(map[string]string, len(baseFields))
	for name, typ := range baseFields {
		resolved[name] = typ
	}

	if len(typeArgs) == 0 {
		if typeParams, ok := c.structTypeParams[structName]; ok && len(typeParams) > 0 {
			c.report(span, fmt.Sprintf("struct %s requires %d type argument(s)", structName, len(typeParams)),
				fmt.Sprintf("provide type arguments like %s<...>", structName))
		}
		return resolved
	}

	typeParams, hasParams := c.structTypeParams[structName]
	if !hasParams {
		c.report(span, fmt.Sprintf("struct %s is not generic", structName), "remove the type arguments")
		return resolved
	}

	if len(typeArgs) != len(typeParams) {
		c.report(span, fmt.Sprintf("struct %s expects %d type argument(s), got %d", structName, len(typeParams), len(typeArgs)),
			fmt.Sprintf("provide %d type argument(s) matching the struct definition", len(typeParams)))
	}

	for i, param := range typeParams {
		if i >= len(typeArgs) {
			break
		}
		for fieldName, fieldType := range resolved {
			resolved[fieldName] = c.substituteTypeParam(fieldType, param.Name, typeArgs[i])
		}
	}

	return resolved
}

func arrayElementType(typ string) (string, bool) {
	// Check for new array syntax: []<int>
	if strings.HasPrefix(typ, "[]<") && strings.HasSuffix(typ, ">") {
		inner := typ[len("[]<") : len(typ)-1]
		return strings.TrimSpace(inner), true
	}
	// Check for old array syntax: array<int> (for backward compatibility)
	if strings.HasPrefix(typ, "array<") && strings.HasSuffix(typ, ">") {
		inner := typ[len("array<") : len(typ)-1]
		return strings.TrimSpace(inner), true
	}
	return "", false
}

func mapTypes(typ string) (string, string, bool) {
	if !strings.HasPrefix(typ, "map<") || !strings.HasSuffix(typ, ">") {
		return "", "", false
	}
	inner := typ[len("map<") : len(typ)-1]
	parts := splitGenericArgs(inner)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func promiseInnerType(typ string) (string, bool) {
	if !strings.HasPrefix(typ, "Promise<") || !strings.HasSuffix(typ, ">") {
		return "", false
	}
	inner := typ[len("Promise<") : len(typ)-1]
	return strings.TrimSpace(inner), true
}

func isNumeric(typ string) bool {
	// Don't treat typeInfer as numeric - it needs to be resolved first
	if typ == typeInfer || typ == "" {
		return false
	}
	switch typ {
	case "int", "long", "byte", "float", "double":
		return true
	default:
		return false
	}
}

// isInteger checks if a type is an integer type (not float/double)
func isInteger(typ string) bool {
	switch typ {
	case "int", "long", "byte":
		return true
	default:
		return false
	}
}

func canCastBetweenTypes(from, to string) bool {
	if from == to {
		return true
	}
	if isNumeric(from) && isNumeric(to) {
		return true
	}
	if strings.HasPrefix(from, "*") && strings.HasPrefix(to, "*") {
		return true
	}
	return false
}

// normalizeTypeSpacing normalizes whitespace in generic type strings
// Converts "map<string,any>" to "map<string, any>" for consistent comparison
func normalizeTypeSpacing(typeStr string) string {
	// Replace ",<" with ", <" and ">,<" with ">, <" in generic types
	result := typeStr
	// Find all occurrences of comma followed by non-space before < or >
	result = strings.ReplaceAll(result, ",<", ", <")
	result = strings.ReplaceAll(result, ",>", ", >")
	// Also handle cases like "map<string,any>" -> "map<string, any>"
	// Use simple replacement for comma followed by identifier
	var b strings.Builder
	for i := 0; i < len(result); i++ {
		if i > 0 && result[i-1] == ',' && (result[i] >= 'a' && result[i] <= 'z' || result[i] >= 'A' && result[i] <= 'Z' || result[i] == '_') {
			b.WriteString(", ")
		}
		b.WriteByte(result[i])
	}
	return b.String()
}

func (c *Checker) typesEqual(a, b string) bool {
	if a == typeError || b == typeError {
		return true
	}
	if a == typeInfer || b == typeInfer {
		return true
	}

	// Resolve type aliases before comparing
	if underlyingA, isAliasA := c.typeAliases[a]; isAliasA {
		a = underlyingA
	}
	if underlyingB, isAliasB := c.typeAliases[b]; isAliasB {
		b = underlyingB
	}

	// Handle array type compatibility: []<T> and array<T> are compatible
	if c.isArrayType(a) && c.isArrayType(b) {
		aElement := c.getArrayElementType(a)
		bElement := c.getArrayElementType(b)
		// If one is a generic type parameter (T) and the other is concrete, they're compatible
		if c.isTypeParam(aElement) || c.isTypeParam(bElement) {
			return true
		}
		return c.typesEqual(aElement, bElement)
	}

	// Handle optional types - preserve optionality for null-safety
	// int? is NOT compatible with int (no implicit unwrapping)
	// Only allow widening: non-optional -> optional (e.g., int can be assigned to int?)
	// Reject narrowing: optional -> non-optional (e.g., int? cannot be assigned to int)
	aBase := strings.TrimRight(a, "?")
	bBase := strings.TrimRight(b, "?")
	aOptional := len(a) - len(aBase)
	bOptional := len(b) - len(bBase)

	// Normalize optional-of-optional (int?? -> int?)
	// But preserve the distinction between optional and non-optional
	if aOptional > 1 {
		a = aBase + "?"
		aOptional = 1
	}
	if bOptional > 1 {
		b = bBase + "?"
		bOptional = 1
	}

	// If both are non-optional, continue with normal comparison below
	if aOptional == 0 && bOptional == 0 {
		// Continue with normal comparison below
	} else if aOptional > 0 && bOptional > 0 {
		// Both are optional - compare base types
		return c.typesEqual(aBase, bBase)
	} else {
		// One is optional, one is not - they are NOT equal
		// This enforces null-safety in type equality checks
		// For assignments, use isAssignable() which allows widening
		return false
	}

	// Handle union types
	if c.isUnionType(a) && c.isUnionType(b) {
		return a == b // Exact union match (already canonicalized by buildUnion)
	}
	if c.isUnionType(a) {
		return c.isTypeInUnion(b, a)
	}
	if c.isUnionType(b) {
		return c.isTypeInUnion(a, b)
	}

	// Handle pointer types - compare base types after stripping leading *
	// Count leading * characters only
	aPtrCount := 0
	for i := 0; i < len(a) && a[i] == '*'; i++ {
		aPtrCount++
	}
	bPtrCount := 0
	for i := 0; i < len(b) && b[i] == '*'; i++ {
		bPtrCount++
	}
	if aPtrCount > 0 || bPtrCount > 0 {
		// Both must have same number of pointer levels
		if aPtrCount != bPtrCount {
			return false
		}
		// Strip leading * prefixes and compare base types
		aBase := a[aPtrCount:]
		bBase := b[bPtrCount:]
		return c.typesEqual(aBase, bBase)
	}

	// Normalize whitespace in generic types (e.g., "map<string,any>" vs "map<string, any>")
	// This ensures type equality works regardless of spacing
	a = normalizeTypeSpacing(a)
	b = normalizeTypeSpacing(b)

	// Handle function types
	if strings.Contains(a, ") -> ") && strings.Contains(b, ") -> ") {
		return c.functionTypesEqual(a, b)
	}
	// If one is a function type and the other is a union containing that function type
	if strings.Contains(a, ") -> ") && c.isUnionType(b) {
		return c.isTypeInUnion(a, b)
	}
	if strings.Contains(b, ") -> ") && c.isUnionType(a) {
		return c.isTypeInUnion(b, a)
	}

	return a == b
}

// functionTypesEqual compares two function types for equality
func (c *Checker) functionTypesEqual(a, b string) bool {
	// Parse function types: (param1, param2) -> returnType
	aArrowIdx := strings.Index(a, ") -> ")
	bArrowIdx := strings.Index(b, ") -> ")
	if aArrowIdx == -1 || bArrowIdx == -1 {
		return false
	}

	// Extract parameter parts
	aParams := a[1:aArrowIdx] // Remove opening (
	bParams := b[1:bArrowIdx]

	// Extract return types
	aReturn := a[aArrowIdx+5:]
	bReturn := b[bArrowIdx+5:]

	// Compare parameters
	if aParams == "" && bParams == "" {
		// Both have no parameters
	} else if aParams == "" || bParams == "" {
		return false // One has parameters, one doesn't
	} else {
		// Split and compare each parameter
		aParamList := strings.Split(aParams, ", ")
		bParamList := strings.Split(bParams, ", ")
		if len(aParamList) != len(bParamList) {
			return false
		}
		for i := range aParamList {
			if !c.typesEqual(strings.TrimSpace(aParamList[i]), strings.TrimSpace(bParamList[i])) {
				return false
			}
		}
	}

	// Compare return types - handle union types
	// If one return type is a union, check if the other is in the union
	if c.isUnionType(aReturn) {
		return c.isTypeInUnion(bReturn, aReturn)
	}
	if c.isUnionType(bReturn) {
		return c.isTypeInUnion(aReturn, bReturn)
	}
	return c.typesEqual(aReturn, bReturn)
}

// isAssignable checks if a value of type fromType can be assigned to a variable of type toType
// This allows widening (non-optional -> optional) but not narrowing (optional -> non-optional)
func (c *Checker) isAssignable(fromType, toType string) bool {
	// Resolve type aliases before checking
	if underlyingTo, isAliasTo := c.typeAliases[toType]; isAliasTo {
		toType = underlyingTo
	}
	if underlyingFrom, isAliasFrom := c.typeAliases[fromType]; isAliasFrom {
		fromType = underlyingFrom
	}

	// "any" accepts any type (except error)
	if toType == "any" && fromType != typeError {
		return true
	}

	// First check exact equality
	if c.typesEqual(fromType, toType) {
		return true
	}

	// Widen integer literals/values to float where a float is expected
	// (e.g. `std.math.pow(2, 8)` -> `pow(2.0, 8.0)`). Only numeric widening,
	// never the reverse, and only to the built-in numeric types so we don't
	// stomp on user type aliases.
	if (toType == "float" || toType == "double") && (fromType == "int" || fromType == "int32" || fromType == "int32_t" || fromType == "int64" || fromType == "int64_t") {
		return true
	}

	// Narrow float-returning stdlib calls to int for common patterns like
	// `let n: int = std.math.pow(2, 3)`. Truncation matches Python/int() but
	// may lose precision — we accept it to match users' expectations for the
	// integer math builtins.
	if (toType == "int" || toType == "int32" || toType == "int32_t" || toType == "int64" || toType == "int64_t") && (fromType == "float" || fromType == "double") {
		return true
	}

	// Handle function types: (A) -> B is assignable to (A) -> B | C (covariant return)
	// and (A) -> B is assignable to (A | C) -> B (contravariant parameters)
	if strings.Contains(fromType, ") -> ") && strings.Contains(toType, ") -> ") {
		return c.functionTypesAssignable(fromType, toType)
	}
	// If one is a function type and the other is a union containing that function type
	if strings.Contains(fromType, ") -> ") && c.isUnionType(toType) {
		return c.isTypeInUnion(fromType, toType)
	}

	// Handle map types: map<K, V1> is assignable to map<K, V2> if V1 is assignable to V2
	// This allows map<string, string> to be assigned to map<string, any>
	if fromKey, fromVal, fromIsMap := mapTypes(fromType); fromIsMap {
		if toKey, toVal, toIsMap := mapTypes(toType); toIsMap {
			// Keys must match exactly
			if c.typesEqual(fromKey, toKey) {
				// Value types follow normal assignability rules
				return c.isAssignable(fromVal, toVal)
			}
		}
	}

	// Handle optional types: allow widening (non-optional -> optional)
	fromBase := strings.TrimRight(fromType, "?")
	toBase := strings.TrimRight(toType, "?")
	fromOptional := len(fromType) - len(fromBase)
	toOptional := len(toType) - len(toBase)

	// Normalize optional-of-optional
	if fromOptional > 1 {
		fromOptional = 1
		fromBase = strings.TrimRight(fromType, "?")
	}
	if toOptional > 1 {
		toOptional = 1
		toBase = strings.TrimRight(toType, "?")
	}

	// Allow widening: non-optional can be assigned to optional if base types match
	if fromOptional == 0 && toOptional > 0 {
		return c.typesEqual(fromBase, toBase)
	}

	// Reject narrowing: optional cannot be assigned to non-optional
	// Reject other mismatches
	return false
}

// functionTypesAssignable checks if a function type fromType is assignable to toType
// Function types are contravariant in parameters and covariant in return types
func (c *Checker) functionTypesAssignable(fromType, toType string) bool {
	// Parse function types: (param1, param2) -> returnType
	fromArrowIdx := strings.Index(fromType, ") -> ")
	toArrowIdx := strings.Index(toType, ") -> ")
	if fromArrowIdx == -1 || toArrowIdx == -1 {
		return false
	}

	// Extract parameter parts
	fromParams := fromType[1:fromArrowIdx] // Remove opening (
	toParams := toType[1:toArrowIdx]

	// Extract return types
	fromReturn := fromType[fromArrowIdx+5:]
	toReturn := toType[toArrowIdx+5:]

	// Parameters must be exactly equal (contravariance would require exact match for now)
	if fromParams != toParams {
		return false
	}

	// Return type: fromReturn must be assignable to toReturn (covariant)
	// If toReturn is a union, fromReturn must be in the union
	if c.isUnionType(toReturn) {
		return c.isTypeInUnion(fromReturn, toReturn)
	}
	// If fromReturn is a union, toReturn must be in the union (for exact match)
	if c.isUnionType(fromReturn) {
		return c.isTypeInUnion(toReturn, fromReturn)
	}
	// Otherwise, they must be equal
	return c.typesEqual(fromReturn, toReturn)
}

// isArrayType checks if a type string represents an array type
func (c *Checker) isArrayType(typeStr string) bool {
	return strings.HasPrefix(typeStr, "[]<") || strings.HasPrefix(typeStr, "array<")
}

// getArrayElementType extracts the element type from an array type
func (c *Checker) getArrayElementType(typeStr string) string {
	if strings.HasPrefix(typeStr, "[]<") {
		// Extract from []<T> format
		inner := typeStr[3 : len(typeStr)-1]
		return inner
	}
	if strings.HasPrefix(typeStr, "array<") {
		// Extract from array<T> format
		inner := typeStr[6 : len(typeStr)-1]
		return inner
	}
	return typeStr
}

// isUnionType checks if a type string represents a union type
func (c *Checker) isUnionType(typeStr string) bool {
	return strings.Contains(typeStr, " | ")
}

// isTypeInUnion checks if a type is a member of a union type
func (c *Checker) isTypeInUnion(memberType, unionType string) bool {
	if !c.isUnionType(unionType) {
		return false
	}

	// Split the union type into its members
	members := strings.Split(unionType, " | ")
	for _, member := range members {
		member = strings.TrimSpace(member)
		// Use typesEqual to handle qualified vs unqualified types
		if c.typesEqual(member, memberType) {
			return true
		}
	}
	return false
}

// -----------------------------------------------------------------------------
// Scope helpers
// -----------------------------------------------------------------------------

func (c *Checker) declare(name, typ string, mutable bool, span lexer.Span) {
	if len(c.scopes) == 0 {
		return
	}
	scope := c.scopes[len(c.scopes)-1]
	if _, exists := scope[name]; exists {
		c.report(span, fmt.Sprintf("%q redeclared in the same scope", name), "rename the symbol or remove the duplicate declaration")
		return
	}
	if typ == "" {
		typ = typeInfer
	}
	scope[name] = Symbol{Type: typ, Mutable: mutable}
}

func (c *Checker) symbolExists(name string) bool {
	for i := len(c.scopes) - 1; i >= 0; i-- {
		if _, ok := c.scopes[i][name]; ok {
			return true
		}
	}
	return false
}

func (c *Checker) lookupSymbol(name string) (Symbol, bool) {
	for i := len(c.scopes) - 1; i >= 0; i-- {
		if sym, ok := c.scopes[i][name]; ok {
			return sym, true
		}
	}
	return Symbol{}, false
}

func (c *Checker) updateSymbolType(name, typ string) {
	if typ == typeInfer || typ == typeError {
		return
	}
	for i := len(c.scopes) - 1; i >= 0; i-- {
		if sym, ok := c.scopes[i][name]; ok {
			sym.Type = typ
			c.scopes[i][name] = sym
			return
		}
	}
}

func (c *Checker) enterScope() {
	c.scopes = append(c.scopes, make(map[string]Symbol))
}

func (c *Checker) leaveScope() {
	if len(c.scopes) == 0 {
		return
	}
	c.scopes = c.scopes[:len(c.scopes)-1]
}

func (c *Checker) pushFunctionContext(name, ret string, isAsync bool) {
	if ret == "" {
		ret = typeVoid
	}
	c.functionStack = append(c.functionStack, functionContext{Name: name, ReturnType: ret, IsAsync: isAsync, HasReturn: false})
}

func (c *Checker) popFunctionContext() {
	if len(c.functionStack) == 0 {
		return
	}
	ctx := c.functionStack[len(c.functionStack)-1]
	if sig, ok := c.functions[ctx.Name]; ok {
		// Preserve Promise wrapper for async functions
		if strings.HasPrefix(sig.Return, "Promise<") {
			// Don't overwrite the Promise wrapper - keep the original signature
			// The return type validation already happened in checkFunc
		} else {
			sig.Return = ctx.ReturnType
			c.functions[ctx.Name] = sig
		}
	}
	c.functionStack = c.functionStack[:len(c.functionStack)-1]
}

func (c *Checker) currentFunctionContext() *functionContext {
	if len(c.functionStack) == 0 {
		return nil
	}
	return &c.functionStack[len(c.functionStack)-1]
}

func (c *Checker) updateCurrentFunctionReturn(newType string) {
	if newType == typeInfer || newType == typeError {
		return
	}
	ctx := c.currentFunctionContext()
	if ctx == nil {
		return
	}
	ctx.ReturnType = newType
}

// -----------------------------------------------------------------------------
// Diagnostics
// -----------------------------------------------------------------------------

func (c *Checker) report(span lexer.Span, message, hint string) {
	c.reportWithSeverity(span, message, hint, lexer.Error, "")
}

func (c *Checker) reportWithSeverity(span lexer.Span, message, hint string, severity lexer.Severity, category string) {
	if span.Start.Line < 1 || span.Start.Line > len(c.lines) {
		span.Start.Line = 1
		span.Start.Column = 1
	}
	lineText := ""
	if span.Start.Line-1 >= 0 && span.Start.Line-1 < len(c.lines) {
		lineText = c.lines[span.Start.Line-1]
	}
	contextLines, contextStart := lexer.BuildContext(c.lines, span)
	diag := lexer.Diagnostic{
		File:             c.filename,
		Message:          message,
		Hint:             hint,
		Span:             span,
		Line:             lineText,
		Context:          contextLines,
		ContextStartLine: contextStart,
		Severity:         severity,
		Category:         category,
	}
	c.diagnostics = append(c.diagnostics, diag)
}

func (c *Checker) reportWarning(span lexer.Span, message, hint string) {
	c.reportWithSeverity(span, message, hint, lexer.Warning, "type-check")
}

func (c *Checker) reportInfo(span lexer.Span, message, hint string) {
	c.reportWithSeverity(span, message, hint, lexer.Info, "suggestion")
}

// isLikelyTypo checks if an identifier might be a typo based on common patterns
func isLikelyTypo(name string) bool {
	// Check for common typos
	commonTypos := map[string]string{
		"prnt":     "print",
		"prin":     "print",
		"prinln":   "println",
		"prntln":   "println",
		"fucn":     "func",
		"functon":  "function",
		"retrun":   "return",
		"retun":    "return",
		"varibale": "variable",
		"varable":  "variable",
		"fals":     "false",
		"tru":      "true",
		"stirng":   "string",
		"strng":    "string",
		"intege":   "integer",
		"intger":   "integer",
	}

	_, isTypo := commonTypos[name]
	return isTypo
}

// suggestSimilarIdentifiers suggests similar identifiers based on common typos
func suggestSimilarIdentifiers(name string) string {
	commonTypos := map[string]string{
		"prnt":     "print",
		"prin":     "print",
		"prinln":   "println",
		"prntln":   "println",
		"fucn":     "func",
		"functon":  "function",
		"retrun":   "return",
		"retun":    "return",
		"varibale": "variable",
		"varable":  "variable",
		"fals":     "false",
		"tru":      "true",
		"stirng":   "string",
		"strng":    "string",
		"intege":   "integer",
		"intger":   "integer",
	}

	if suggestion, exists := commonTypos[name]; exists {
		return suggestion
	}

	// Simple Levenshtein distance-based suggestions for very short identifiers
	if len(name) <= 4 {
		suggestions := []string{"print", "println", "func", "let", "var", "if", "else", "for", "while", "return"}
		var similar []string
		for _, suggestion := range suggestions {
			if levenshteinDistance(name, suggestion) <= 2 {
				similar = append(similar, suggestion)
			}
		}
		if len(similar) > 0 {
			return strings.Join(similar, ", ")
		}
	}

	return "print, println, func, let, var"
}

// levenshteinDistance calculates the edit distance between two strings
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

func min(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}

// processImports handles import statements and makes symbols available.
func (c *Checker) processImports(mod *ast.Module) {
	// Process imports from module.Imports
	for _, imp := range mod.Imports {
		c.processImport(imp)
	}

	// Also check decls for imports (backward compatibility)
	for _, decl := range mod.Decls {
		if imp, ok := decl.(*ast.ImportDecl); ok {
			c.processImport(imp)
		}
	}
}

func (c *Checker) processImport(imp *ast.ImportDecl) {
	if len(imp.Path) == 0 {
		c.report(imp.Span(), "empty import path", "provide a valid import path")
		return
	}

	pathKey := strings.Join(imp.Path, ".")

	// Handle std imports
	if imp.Path[0] == "std" {
		// Bind alias or top-level module name into the current scope first
		if len(c.scopes) > 0 {
			scope := c.scopes[len(c.scopes)-1]
			local := imp.Alias
			if local == "" {
				if len(imp.Path) == 1 {
					local = "std"
				} else {
					local = imp.Path[len(imp.Path)-1]
				}
			}
			scope[local] = Symbol{Type: "module", Mutable: false}
		}

		if c.processedImports[pathKey] {
			return
		}
		c.processedImports[pathKey] = true

		// Load the std module to get its function signatures
		module, err := c.moduleLoader.LoadModule(imp.Path)
		if err != nil {
			c.report(imp.Span(), fmt.Sprintf("failed to load std module: %v", err), "make sure the std module exists")
			return
		}

		// Register the module's type aliases first (so they're available when registering struct fields and function signatures)
		c.registerModuleTypeAliases(module, imp.Path)
		// Register the module's function signatures
		c.registerModuleFunctionSignatures(module, imp.Path)
		// Register the module's struct fields
		c.registerModuleStructFields(module, imp.Path)

		// Process nested imports recursively (but don't process function declarations)
		// Only process imports, not function declarations
		for _, nestedImp := range module.Imports {
			c.processImport(nestedImp)
		}
		for _, decl := range module.Decls {
			if nestedImp, ok := decl.(*ast.ImportDecl); ok {
				c.processImport(nestedImp)
			}
		}

		// Add std symbols to imports
		c.imports["io"] = true
		c.imports["math"] = true
		c.imports["string"] = true
		c.imports["array"] = true
		c.imports["os"] = true
		c.imports["collections"] = true
		c.imports["testing"] = true
		c.imports["dev"] = true
		c.imports["test"] = true
		c.imports["assert"] = true
		c.imports["file"] = true
		c.imports["algorithms"] = true
		c.imports["time"] = true
		c.imports["network"] = true
		c.imports["web"] = true
		c.imports["log"] = true
		return
	}

	// Handle local file imports: bind module into scope only; the compiler merges
	// imported functions into the AST so we don't register here to avoid duplicates
	if len(c.scopes) > 0 {
		scope := c.scopes[len(c.scopes)-1]
		local := imp.Alias
		if local == "" {
			local = imp.Path[len(imp.Path)-1]
		}
		scope[local] = Symbol{Type: "module", Mutable: false}
	}
}

// isStdSymbol checks if a qualified name is a standard library symbol.
func (c *Checker) isStdSymbol(name string) bool {
	parts := strings.Split(name, ".")
	if len(parts) < 2 {
		return false
	}

	if parts[0] == "std" {
		// Check if the module is imported
		if len(parts) >= 2 {
			module := parts[1]
			if c.imports[module] {
				// For now, accept any function in imported modules
				return true
			}
			// `assert` and `test` are compiler-provided intrinsics under
			// std — they aren't real submodules but users reference them
			// as `std.assert.eq`, `std.assert.true`, `std.test.start`, etc.
			if (module == "assert" || module == "test") && c.imports["std"] {
				return true
			}
		}
	}

	return false
}

// isAliasedStdSymbol checks if a qualified name is an aliased std import
func (c *Checker) isAliasedStdSymbol(qualifiedName string) bool {
	parts := strings.Split(qualifiedName, ".")
	if len(parts) < 2 {
		return false
	}

	// Check if the first part is an aliased std module
	alias := parts[0]
	// Look through imports to see if this alias maps to a std module
	for _, imp := range c.scopes[0] { // Check the global scope
		if imp.Type == "module" {
			// This is a simplified check - in a real implementation we'd need to track aliases properly
			// For now, check if it's a known std alias
			if alias == "io" || alias == "math" || alias == "str" || alias == "string" {
				return true
			}
		}
	}
	return false
}

// mapAliasToStd maps an aliased std function to its full std path
func (c *Checker) mapAliasToStd(qualifiedName string) string {
	parts := strings.Split(qualifiedName, ".")
	if len(parts) < 2 {
		return qualifiedName
	}

	alias := parts[0]
	// Map common aliases to their std modules
	switch alias {
	case "io":
		return "std.io." + parts[1]
	case "math":
		return "std.math." + parts[1]
	case "str", "string":
		return "std.string." + parts[1]
	default:
		return qualifiedName
	}
}

// registerMergedFunctionSignatures registers function signatures for imported modules
// that were merged into the AST by the compiler (e.g., math_utils.add)
func (c *Checker) registerMergedFunctionSignatures(mod *ast.Module) {
	for _, decl := range mod.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			// Check if this is a namespaced function (contains a dot)
			if strings.Contains(fn.Name, ".") {
				// Extract module prefix from function name (e.g., "std.web" from "std.web.context_status")
				parts := strings.Split(fn.Name, ".")
				if len(parts) < 2 {
					continue
				}
				modulePrefix := strings.Join(parts[:len(parts)-1], ".")

				// Collect types defined in this module for qualification
				moduleTypes := make(map[string]bool)
				for _, d := range mod.Decls {
					if structDecl, ok := d.(*ast.StructDecl); ok {
						if strings.HasPrefix(structDecl.Name, modulePrefix+".") {
							// Extract the unqualified name
							typeName := structDecl.Name[len(modulePrefix)+1:]
							moduleTypes[typeName] = true
							// Also register the qualified name
							moduleTypes[structDecl.Name] = true
						}
					}
					if typeAlias, ok := d.(*ast.TypeAliasDecl); ok {
						if strings.HasPrefix(typeAlias.Name, modulePrefix+".") {
							typeName := typeAlias.Name[len(modulePrefix)+1:]
							moduleTypes[typeName] = true
							moduleTypes[typeAlias.Name] = true
						}
					}
				}

				// Helper to qualify types
				qualifyType := func(typeStr string) string {
					// If already qualified with module prefix, return as-is
					if strings.HasPrefix(typeStr, modulePrefix+".") {
						return typeStr
					}
					// Check if this is a simple type name that exists in the module
					if moduleTypes[typeStr] {
						return modulePrefix + "." + typeStr
					}
					return typeStr
				}

				// Enter type parameter scope for this function
				c.enterTypeParams(fn.TypeParams)

				sig := FunctionSignature{Return: "void"}
				if fn.Return != nil {
					// Use typeExprToString to avoid error reporting during signature registration
					resolvedType := c.typeExprToString(fn.Return)
					sig.Return = qualifyType(resolvedType)
				}
				// If function is async, wrap return type in Promise<T>
				if fn.IsAsync {
					if sig.Return == typeVoid {
						sig.Return = "Promise<void>"
					} else {
						sig.Return = "Promise<" + sig.Return + ">"
					}
				}
				sig.Params = make([]string, len(fn.Params))
				for i, param := range fn.Params {
					// Use typeExprToString to avoid error reporting during signature registration
					resolvedType := c.typeExprToString(param.Type)
					sig.Params[i] = qualifyType(resolvedType)
				}
				sig.TypeParams = fn.TypeParams

				// Leave type parameter scope
				c.leaveTypeParams(fn.TypeParams)

				c.functions[fn.Name] = sig
			}
		}
	}
}

// registerMergedTypeAliases registers type aliases for imported modules
// that were merged into the AST by the compiler (e.g., std.web.Handler)
func (c *Checker) registerMergedTypeAliases(mod *ast.Module) {
	for _, decl := range mod.Decls {
		if typeAlias, ok := decl.(*ast.TypeAliasDecl); ok {
			// Check if this is a namespaced type alias (contains a dot)
			if strings.Contains(typeAlias.Name, ".") {
				// Extract module prefix from alias name (e.g., "std.web" from "std.web.Handler")
				parts := strings.Split(typeAlias.Name, ".")
				if len(parts) < 2 {
					continue
				}
				modulePrefix := strings.Join(parts[:len(parts)-1], ".")

				// Convert the underlying type to a string (don't resolve types yet - we'll qualify them)
				underlyingType := c.typeExprToString(typeAlias.Type)

				// Qualify types in the underlying type that belong to this module
				moduleTypes := make(map[string]bool)
				for _, d := range mod.Decls {
					if structDecl, ok := d.(*ast.StructDecl); ok {
						if strings.HasPrefix(structDecl.Name, modulePrefix+".") {
							typeName := structDecl.Name[len(modulePrefix)+1:]
							moduleTypes[typeName] = true
							moduleTypes[structDecl.Name] = true
						}
					}
					if ta, ok := d.(*ast.TypeAliasDecl); ok {
						if strings.HasPrefix(ta.Name, modulePrefix+".") {
							typeName := ta.Name[len(modulePrefix)+1:]
							moduleTypes[typeName] = true
							moduleTypes[ta.Name] = true
						}
					}
				}

				// Helper to recursively qualify types in function types and union types
				var qualifyTypeRecursive func(string) string
				qualifyTypeRecursive = func(typeStr string) string {
					// If already qualified with module prefix, return as-is
					if strings.HasPrefix(typeStr, modulePrefix+".") {
						return typeStr
					}

					// Handle function types: (A) -> B
					if strings.Contains(typeStr, ") -> ") {
						arrowIdx := strings.Index(typeStr, ") -> ")
						paramPart := typeStr[1:arrowIdx]   // Remove opening (
						returnPart := typeStr[arrowIdx+5:] // After ") -> "

						// Qualify parameters
						var qualifiedParams string
						if paramPart != "" {
							params := strings.Split(paramPart, ", ")
							qualifiedParamsList := make([]string, len(params))
							for i, param := range params {
								qualifiedParamsList[i] = qualifyTypeRecursive(strings.TrimSpace(param))
							}
							qualifiedParams = strings.Join(qualifiedParamsList, ", ")
						}

						// Qualify return type (may be a union)
						qualifiedReturn := qualifyTypeRecursive(returnPart)

						return "(" + qualifiedParams + ") -> " + qualifiedReturn
					}

					// Handle union types: A | B
					if strings.Contains(typeStr, " | ") {
						members := strings.Split(typeStr, " | ")
						qualifiedMembers := make([]string, len(members))
						for i, member := range members {
							qualifiedMembers[i] = qualifyTypeRecursive(strings.TrimSpace(member))
						}
						return strings.Join(qualifiedMembers, " | ")
					}

					// Handle generic types: A<T>
					if idx := strings.Index(typeStr, "<"); idx > 0 {
						baseType := typeStr[:idx]
						argsPart := typeStr[idx+1 : len(typeStr)-1]

						// Split arguments by comma, handling nested generics
						var args []string
						var current strings.Builder
						depth := 0
						for _, char := range argsPart {
							if char == '<' {
								depth++
								current.WriteRune(char)
							} else if char == '>' {
								depth--
								current.WriteRune(char)
							} else if char == ',' && depth == 0 {
								args = append(args, strings.TrimSpace(current.String()))
								current.Reset()
							} else {
								current.WriteRune(char)
							}
						}
						if current.Len() > 0 {
							args = append(args, strings.TrimSpace(current.String()))
						}

						// Recursively qualify each argument
						qualifiedArgs := make([]string, len(args))
						for i, arg := range args {
							qualifiedArgs[i] = qualifyTypeRecursive(arg)
						}

						// Qualify base type
						qualifiedBase := qualifyTypeRecursive(baseType)

						return qualifiedBase + "<" + strings.Join(qualifiedArgs, ", ") + ">"
					}

					// Simple type - check if it's a module type
					if moduleTypes[typeStr] {
						return modulePrefix + "." + typeStr
					}
					return typeStr
				}

				qualifiedUnderlying := qualifyTypeRecursive(underlyingType)

				// Register the type alias
				c.typeAliases[typeAlias.Name] = qualifiedUnderlying
				c.knownTypes[typeAlias.Name] = struct{}{}
			}
		}
	}
}

// registerModuleStructFields registers struct field definitions from an imported module
func (c *Checker) registerModuleStructFields(mod *ast.Module, importPath []string) {
	// Create the module prefix (e.g., "std.math" for importPath ["std", "math"])
	modulePrefix := strings.Join(importPath, ".")

	for _, decl := range mod.Decls {
		if structDecl, ok := decl.(*ast.StructDecl); ok {
			// Create the fully qualified struct name
			qualifiedName := modulePrefix + "." + structDecl.Name

			// Register the struct type
			c.knownTypes[qualifiedName] = struct{}{}
			// Also register the unqualified name for backward compatibility
			c.knownTypes[structDecl.Name] = struct{}{}

			// Collect field types
			fields := make(map[string]string, len(structDecl.Fields))
			for _, field := range structDecl.Fields {
				fields[field.Name] = typeExprToString(field.Type)
			}

			// Store with qualified name
			c.structFields[qualifiedName] = fields
			// Also store with unqualified name for backward compatibility
			// (in case the parser uses unqualified names in some contexts)
			c.structFields[structDecl.Name] = fields

			// Store type parameters for generic structs
			if len(structDecl.TypeParams) > 0 {
				c.structTypeParams[qualifiedName] = structDecl.TypeParams
				c.structTypeParams[structDecl.Name] = structDecl.TypeParams
			}
		}
	}
}

// registerModuleTypeAliases registers type aliases from an imported module
func (c *Checker) registerModuleTypeAliases(mod *ast.Module, importPath []string) {
	// Create the module prefix (e.g., "std.math" for importPath ["std", "math"])
	modulePrefix := strings.Join(importPath, ".")

	// Collect types defined in this module (for qualification)
	moduleTypes := make(map[string]bool)
	for _, decl := range mod.Decls {
		if structDecl, ok := decl.(*ast.StructDecl); ok {
			moduleTypes[structDecl.Name] = true
		}
		if typeAlias, ok := decl.(*ast.TypeAliasDecl); ok {
			moduleTypes[typeAlias.Name] = true
		}
	}

	// Helper function to qualify types defined in this module
	var qualifyType func(string) string
	qualifyType = func(typeStr string) string {
		// Check if this is a simple type name that exists in the module
		if moduleTypes[typeStr] {
			return modulePrefix + "." + typeStr
		}
		// For generic types, check the base type
		if idx := strings.Index(typeStr, "<"); idx > 0 {
			baseType := typeStr[:idx]
			if moduleTypes[baseType] {
				return modulePrefix + "." + typeStr
			}
		}
		// For function types, recursively qualify parameter and return types
		// Check this BEFORE union types to avoid splitting on " | " inside function return types
		if strings.Contains(typeStr, ") -> ") {
			arrowIdx := strings.Index(typeStr, ") -> ")
			if arrowIdx == -1 {
				return typeStr
			}
			paramPart := typeStr[1:arrowIdx]   // Remove opening (
			returnPart := typeStr[arrowIdx+5:] // After ") -> "

			// Safety check: ensure returnPart is not empty
			if returnPart == "" {
				return typeStr // Return original if return type is empty (shouldn't happen)
			}

			// Qualify parameters
			if paramPart != "" {
				params := strings.Split(paramPart, ", ")
				qualifiedParams := make([]string, len(params))
				for i, param := range params {
					qualifiedParams[i] = qualifyType(strings.TrimSpace(param))
				}
				paramPart = strings.Join(qualifiedParams, ", ")
			}

			// Qualify return type (which may be a union)
			returnType := qualifyType(returnPart)

			// Safety check: ensure returnType is not empty
			if returnType == "" {
				returnType = returnPart // Fallback to original if qualification failed
			}

			return "(" + paramPart + ") -> " + returnType
		}
		// For union types, qualify each member (check AFTER function types)
		if strings.Contains(typeStr, " | ") {
			members := strings.Split(typeStr, " | ")
			qualifiedMembers := make([]string, len(members))
			for i, member := range members {
				qualifiedMembers[i] = qualifyType(strings.TrimSpace(member))
			}
			return strings.Join(qualifiedMembers, " | ")
		}
		return typeStr
	}

	for _, decl := range mod.Decls {
		if typeAlias, ok := decl.(*ast.TypeAliasDecl); ok {
			// Create the fully qualified type alias name
			qualifiedName := modulePrefix + "." + typeAlias.Name

			// Convert []string to []ast.TypeParam for scope management
			var typeParams []ast.TypeParam
			if len(typeAlias.TypeParams) > 0 {
				typeParams = make([]ast.TypeParam, len(typeAlias.TypeParams))
				for i, paramName := range typeAlias.TypeParams {
					typeParams[i] = ast.TypeParam{
						Name: paramName,
						Span: typeAlias.Span(), // Use alias span as approximation
					}
				}
				c.enterTypeParams(typeParams)
			}

			// Resolve the underlying type as a string (don't check types here - they'll be qualified)
			// Use typeExprToString to avoid type checking errors during module registration
			underlyingType := c.typeExprToString(typeAlias.Type)

			// Leave type parameter scope before qualifying (so we don't qualify type params)
			if len(typeParams) > 0 {
				c.leaveTypeParams(typeParams)
			}

			// Qualify types in the underlying type (after leaving type param scope)
			qualifiedUnderlying := qualifyType(underlyingType)

			// Register the type alias
			c.typeAliases[qualifiedName] = qualifiedUnderlying
			c.knownTypes[qualifiedName] = struct{}{}
		}
	}
}

// registerModuleFunctionSignatures registers function signatures from an imported module
func (c *Checker) registerModuleFunctionSignatures(mod *ast.Module, importPath []string) {
	// Create the module prefix (e.g., "std.math" for importPath ["std", "math"])
	modulePrefix := strings.Join(importPath, ".")

	// Collect types defined in this module (structs, type aliases, etc.)
	moduleTypes := make(map[string]bool)
	for _, decl := range mod.Decls {
		if structDecl, ok := decl.(*ast.StructDecl); ok {
			moduleTypes[structDecl.Name] = true
		}
		if typeAlias, ok := decl.(*ast.TypeAliasDecl); ok {
			moduleTypes[typeAlias.Name] = true
		}
	}

	// Helper function to qualify types defined in this module
	var qualifyType func(string) string
	qualifyType = func(typeStr string) string {
		// Check if this is a simple type name that exists in the module
		if moduleTypes[typeStr] {
			return modulePrefix + "." + typeStr
		}
		// For generic types, check the base type and recursively qualify type arguments
		if idx := strings.Index(typeStr, "<"); idx > 0 {
			baseType := typeStr[:idx]
			// Extract type arguments (handle nested generics)
			argsPart := typeStr[idx+1 : len(typeStr)-1] // Remove < and >
			// Split arguments by comma, but be careful with nested generics
			var args []string
			var current strings.Builder
			depth := 0
			for _, char := range argsPart {
				if char == '<' {
					depth++
					current.WriteRune(char)
				} else if char == '>' {
					depth--
					current.WriteRune(char)
				} else if char == ',' && depth == 0 {
					args = append(args, strings.TrimSpace(current.String()))
					current.Reset()
				} else {
					current.WriteRune(char)
				}
			}
			if current.Len() > 0 {
				args = append(args, strings.TrimSpace(current.String()))
			}
			// Recursively qualify each argument
			qualifiedArgs := make([]string, len(args))
			for i, arg := range args {
				qualifiedArgs[i] = qualifyType(arg)
			}
			qualifiedArgsStr := strings.Join(qualifiedArgs, ", ")
			// Check if base type should be qualified
			if moduleTypes[baseType] {
				return modulePrefix + "." + baseType + "<" + qualifiedArgsStr + ">"
			}
			return baseType + "<" + qualifiedArgsStr + ">"
		}
		// For function types, recursively qualify parameter and return types
		if strings.Contains(typeStr, ") -> ") {
			arrowIdx := strings.Index(typeStr, ") -> ")
			paramPart := typeStr[1:arrowIdx]   // Remove opening (
			returnPart := typeStr[arrowIdx+5:] // After ") -> "

			// Qualify parameters
			if paramPart != "" {
				params := strings.Split(paramPart, ", ")
				qualifiedParams := make([]string, len(params))
				for i, param := range params {
					qualifiedParams[i] = qualifyType(strings.TrimSpace(param))
				}
				paramPart = strings.Join(qualifiedParams, ", ")
			}

			// Qualify return type
			returnType := qualifyType(returnPart)

			return "(" + paramPart + ") -> " + returnType
		}
		// For union types, qualify each member
		if strings.Contains(typeStr, " | ") {
			members := strings.Split(typeStr, " | ")
			qualifiedMembers := make([]string, len(members))
			for i, member := range members {
				qualifiedMembers[i] = qualifyType(strings.TrimSpace(member))
			}
			return strings.Join(qualifiedMembers, " | ")
		}
		return typeStr
	}

	for _, decl := range mod.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			// Create the fully qualified function name
			qualifiedName := modulePrefix + "." + fn.Name

			// Enter type parameter scope for this function
			c.enterTypeParams(fn.TypeParams)

			// Build the function signature
			sig := FunctionSignature{Return: "void"}
			if fn.Return != nil {
				resolvedType := c.resolveTypeExpr(fn.Return)
				sig.Return = qualifyType(resolvedType)
			}
			// If function is async, wrap return type in Promise<T>
			if fn.IsAsync {
				if sig.Return == typeVoid {
					sig.Return = "Promise<void>"
				} else {
					sig.Return = "Promise<" + sig.Return + ">"
				}
			}
			sig.Params = make([]string, len(fn.Params))
			for i, param := range fn.Params {
				resolvedType := c.resolveTypeExpr(param.Type)
				sig.Params[i] = qualifyType(resolvedType)
			}
			sig.TypeParams = fn.TypeParams

			// Leave type parameter scope
			c.leaveTypeParams(fn.TypeParams)

			// Register the function signature
			c.functions[qualifiedName] = sig
		}
	}
}

// -----------------------------------------------------------------------------
// Utilities
// -----------------------------------------------------------------------------

func splitLines(input string) []string {
	normalized := strings.ReplaceAll(input, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	return strings.Split(normalized, "\n")
}
