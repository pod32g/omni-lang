package checker

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/omni-lang/omni/internal/ast"
	"github.com/omni-lang/omni/internal/lexer"
	"github.com/omni-lang/omni/internal/parser"
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
		filename:     filename,
		lines:        splitLines(src),
		knownTypes:   make(map[string]struct{}),
		structFields: make(map[string]map[string]string),
		functions:    make(map[string]FunctionSignature),
		imports:      make(map[string]bool),
		moduleLoader: *NewModuleLoader(),
	}
	c.initBuiltins()
	c.collectTypeDecls(mod)
	c.enterScope()
	c.processImports(mod)
	c.registerMergedFunctionSignatures(mod)
	c.registerTopLevelSymbols(mod)
	c.checkModule(mod)
	c.leaveScope()

	if len(c.diagnostics) == 0 {
		return nil
	}
	return errors.Join(c.diagnostics...)
}

// ModuleLoader handles loading and caching of imported modules.
type ModuleLoader struct {
	// Cache of loaded modules by import path
	cache map[string]*ast.Module
	// Search paths for finding modules
	searchPaths []string
}

// NewModuleLoader creates a new module loader with default search paths.
func NewModuleLoader() *ModuleLoader {
	return &ModuleLoader{
		cache:       make(map[string]*ast.Module),
		searchPaths: []string{".", "examples", "std"},
	}
}

// LoadModule loads a module by its import path.
func (ml *ModuleLoader) LoadModule(importPath []string) (*ast.Module, error) {
	pathKey := strings.Join(importPath, ".")

	// Check cache first
	if module, exists := ml.cache[pathKey]; exists {
		return module, nil
	}

	// Try to find the module file
	modulePath, err := ml.findModuleFile(importPath)
	if err != nil {
		return nil, err
	}

	// Read and parse the file
	content, err := os.ReadFile(modulePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read module %s: %w", pathKey, err)
	}

	module, err := parser.Parse(modulePath, string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse module %s: %w", pathKey, err)
	}

	// Cache the module
	ml.cache[pathKey] = module
	return module, nil
}

// findModuleFile searches for a module file in the search paths.
func (ml *ModuleLoader) findModuleFile(importPath []string) (string, error) {
	// Convert import path to file path
	moduleName := importPath[len(importPath)-1] // Last segment is the module name
	fileName := moduleName + ".omni"

	// Search in each search path
	for _, searchPath := range ml.searchPaths {
		fullPath := filepath.Join(searchPath, fileName)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, nil
		}
	}

	return "", fmt.Errorf("module %s not found in search paths: %v", strings.Join(importPath, "."), ml.searchPaths)
}

// Checker encapsulates the mutable state required to validate an OmniLang AST.
type Checker struct {
	filename   string
	lines      []string
	knownTypes map[string]struct{}

	structFields map[string]map[string]string
	functions    map[string]FunctionSignature

	scopes      []map[string]Symbol
	diagnostics []error

	functionStack []functionContext

	// Import resolution
	imports map[string]bool // available imported symbols
	// Module loader for local imports
	moduleLoader ModuleLoader
}

// Symbol represents an entry in the scope stack.
type Symbol struct {
	Type    string
	Mutable bool
}

// FunctionSignature captures parameter and return type information for a function.
type FunctionSignature struct {
	Params []string
	Return string
}

type functionContext struct {
	Name       string
	ReturnType string
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
		case *ast.EnumDecl:
			c.knownTypes[d.Name] = struct{}{}
		}
	}
}

func (c *Checker) registerTopLevelSymbols(mod *ast.Module) {
	for _, decl := range mod.Decls {
		switch d := decl.(type) {
		case *ast.LetDecl:
			typ := c.resolveTypeExpr(d.Type)
			c.declare(d.Name, typ, false, d.Span())
		case *ast.VarDecl:
			typ := c.resolveTypeExpr(d.Type)
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
			c.declare(d.Name, "func", false, d.Span())
		}
	}
}

func (c *Checker) buildFunctionSignature(decl *ast.FuncDecl) FunctionSignature {
	params := make([]string, len(decl.Params))
	for i, param := range decl.Params {
		params[i] = c.resolveTypeExpr(param.Type)
	}
	ret := typeVoid
	if decl.Return != nil {
		ret = c.resolveTypeExpr(decl.Return)
	}
	return FunctionSignature{Params: params, Return: ret}
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
			c.checkFunc(d)
		}
	}
}

func (c *Checker) checkStruct(decl *ast.StructDecl) {
	for _, field := range decl.Fields {
		c.checkTypeExpr(field.Type)
	}
}

func (c *Checker) checkFunc(decl *ast.FuncDecl) {
	sig := c.functions[decl.Name]
	expectedReturn := sig.Return
	if decl.Return != nil {
		expectedReturn = c.checkTypeExpr(decl.Return)
	}

	c.pushFunctionContext(decl.Name, expectedReturn)
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
	}

	if decl.Body != nil {
		c.checkBlock(decl.Body)
	}

	c.leaveScope()
	c.popFunctionContext()
}

func (c *Checker) checkLet(decl *ast.LetDecl, mutable bool) {
	expectedType := typeInfer
	if decl.Type != nil {
		expectedType = c.checkTypeExpr(decl.Type)
	}
	valueType := typeInfer
	if decl.Value != nil {
		valueType = c.checkExpr(decl.Value)
	}

	finalType := expectedType
	if finalType == typeInfer {
		finalType = valueType
	} else if valueType != typeInfer && valueType != typeError && !c.typesEqual(finalType, valueType) {
		c.report(decl.Span(), fmt.Sprintf("cannot assign %s to %s", valueType, finalType),
			fmt.Sprintf("convert the expression to %s or change the variable type to %s", finalType, valueType))
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
		} else if valueType != typeInfer && valueType != typeError && !c.typesEqual(declaredType, valueType) {
			c.report(s.Span(), fmt.Sprintf("cannot assign %s to %s", valueType, declaredType),
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
	}
}

func (c *Checker) checkBindingStmt(stmt *ast.BindingStmt) {
	declaredType := typeInfer
	if stmt.Type != nil {
		declaredType = c.checkTypeExpr(stmt.Type)
	}
	valueType := typeInfer
	if stmt.Value != nil {
		valueType = c.checkExpr(stmt.Value)
	}
	finalType := declaredType
	if declaredType == typeInfer {
		finalType = valueType
	} else if valueType != typeInfer && valueType != typeError && !c.typesEqual(declaredType, valueType) {
		c.report(stmt.Span(), fmt.Sprintf("cannot assign %s to %s", valueType, declaredType),
			fmt.Sprintf("convert the expression to %s or change the variable type to %s", declaredType, valueType))
	}
	c.declare(stmt.Name, finalType, stmt.Mutable, stmt.Span())
}

func (c *Checker) checkForStmt(stmt *ast.ForStmt) {
	c.enterScope()
	if stmt.IsRange {
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
		c.leaveScope()
		return
	}

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
	c.leaveScope()
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
	if valueType != typeError && !c.typesEqual(expected, valueType) {
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
		c.report(e.Span(), fmt.Sprintf("undefined identifier %q", e.Name),
			fmt.Sprintf("declare the variable with 'let %s: <type> = <value>' or 'var %s: <type> = <value>' before using it", e.Name, e.Name))
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
		}
		return typeInfer
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

		// Handle struct field access
		if fields, ok := c.structFields[targetType]; ok {
			if fieldType, exists := fields[e.Member]; exists {
				return fieldType
			}
			c.report(e.Span(), fmt.Sprintf("struct %s has no field %q", targetType, e.Member), "use a declared field name")
			return typeError
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
						if strings.Contains(fullName, "length") {
							return "int"
						}
						if strings.Contains(fullName, "concat") {
							return "string"
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
			elementType = typeError
		}
		return buildGeneric("array", []string{elementType})
	case *ast.MapLiteralExpr:
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
			if valueType == typeInfer {
				valueType = vt
			} else if vt != typeInfer && vt != typeError && !c.typesEqual(valueType, vt) {
				c.report(entry.Value.Span(), "map literal values must share the same type", "adjust the map values to match")
				valueType = typeError
			}
		}
		if keyType == typeInfer {
			keyType = typeError
		}
		if valueType == typeInfer {
			valueType = typeError
		}
		return buildGeneric("map", []string{keyType, valueType})
	case *ast.StructLiteralExpr:
		fields, ok := c.structFields[e.TypeName]
		if !ok {
			c.report(e.Span(), fmt.Sprintf("unknown struct type %q", e.TypeName), "define the struct before constructing it")
			return typeError
		}
		for _, field := range e.Fields {
			expectedType, exists := fields[field.Name]
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
		return e.TypeName
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
		// Handle string concatenation (string + string, string + int, int + string)
		if leftType == "string" || rightType == "string" {
			return "string"
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
		if !isNumeric(leftType) || !isNumeric(rightType) {
			c.report(expr.Span(), fmt.Sprintf("operator %s requires numeric operands", expr.Op),
				fmt.Sprintf("use numeric expressions (int, float), got %s and %s", leftType, rightType))
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

	// Handle function calls with qualified names (e.g., math_utils.add, std.math.max)
	var qualifiedName string
	if ident, ok := expr.Callee.(*ast.IdentifierExpr); ok {
		qualifiedName = ident.Name
	} else if member, ok := expr.Callee.(*ast.MemberExpr); ok {
		if memberIdent, ok := member.Target.(*ast.IdentifierExpr); ok {
			qualifiedName = memberIdent.Name + "." + member.Member
		}
	}

	if qualifiedName != "" {
		if sig, exists := c.functions[qualifiedName]; exists {
			if len(expr.Args) != len(sig.Params) {
				c.report(expr.Span(), fmt.Sprintf("function %s expects %d arguments, got %d", qualifiedName, len(sig.Params), len(expr.Args)),
					fmt.Sprintf("provide %d argument(s) matching the function signature: %s(%s)", len(sig.Params), qualifiedName, strings.Join(sig.Params, ", ")))
			}
			for i, arg := range expr.Args {
				argType := c.checkExpr(arg)
				if i < len(sig.Params) {
					expected := sig.Params[i]
					if expected != typeInfer && argType != typeError && !c.typesEqual(expected, argType) {
						c.report(arg.Span(), fmt.Sprintf("argument %d expects %s, got %s", i+1, expected, argType),
							fmt.Sprintf("convert the argument to %s or use a %s expression", expected, expected))
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

func (c *Checker) checkAssignmentExpr(expr *ast.AssignmentExpr) string {
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
	if rhsType != typeError && sym.Type != typeInfer && !c.typesEqual(sym.Type, rhsType) {
		c.report(expr.Right.Span(), fmt.Sprintf("cannot assign %s to %s", rhsType, sym.Type),
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
	return typeExprToString(t)
}

func (c *Checker) checkTypeExpr(t *ast.TypeExpr) string {
	if t == nil {
		return typeInfer
	}
	if _, ok := c.knownTypes[t.Name]; !ok {
		c.report(t.Span(), fmt.Sprintf("unknown type %q", t.Name), "import or declare the type before use")
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
			b.WriteByte(',')
		}
		b.WriteString(arg)
	}
	b.WriteByte('>')
	return b.String()
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

func arrayElementType(typ string) (string, bool) {
	if !strings.HasPrefix(typ, "array<") || !strings.HasSuffix(typ, ">") {
		return "", false
	}
	inner := typ[len("array<") : len(typ)-1]
	return strings.TrimSpace(inner), true
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

func isNumeric(typ string) bool {
	switch typ {
	case "int", "long", "byte", "float", "double":
		return true
	default:
		return false
	}
}

func (c *Checker) typesEqual(a, b string) bool {
	if a == typeError || b == typeError {
		return true
	}
	if a == typeInfer || b == typeInfer {
		return true
	}
	return a == b
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

func (c *Checker) pushFunctionContext(name, ret string) {
	if ret == "" {
		ret = typeVoid
	}
	c.functionStack = append(c.functionStack, functionContext{Name: name, ReturnType: ret})
}

func (c *Checker) popFunctionContext() {
	if len(c.functionStack) == 0 {
		return
	}
	ctx := c.functionStack[len(c.functionStack)-1]
	if sig, ok := c.functions[ctx.Name]; ok {
		sig.Return = ctx.ReturnType
		c.functions[ctx.Name] = sig
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
	if span.Start.Line < 1 || span.Start.Line > len(c.lines) {
		span.Start.Line = 1
		span.Start.Column = 1
	}
	lineText := ""
	if span.Start.Line-1 >= 0 && span.Start.Line-1 < len(c.lines) {
		lineText = c.lines[span.Start.Line-1]
	}
	diag := lexer.Diagnostic{
		File:    c.filename,
		Message: message,
		Hint:    hint,
		Span:    span,
		Line:    lineText,
	}
	c.diagnostics = append(c.diagnostics, diag)
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

	// Handle std imports
	if imp.Path[0] == "std" {
		// Add std symbols to imports
		c.imports["io"] = true
		c.imports["math"] = true
		c.imports["string"] = true
		c.imports["array"] = true
		c.imports["os"] = true
		c.imports["collections"] = true

		// Bind alias or top-level module name
		if len(c.scopes) > 0 {
			scope := c.scopes[len(c.scopes)-1]
			local := imp.Alias
			if local == "" {
				// default binding is the last segment (e.g., std.io -> io) or 'std' if just std
				if len(imp.Path) == 1 {
					local = "std"
				} else {
					local = imp.Path[len(imp.Path)-1]
				}
			}
			scope[local] = Symbol{Type: "module", Mutable: false}
		}
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
				sig := FunctionSignature{Return: "void"}
				if fn.Return != nil {
					sig.Return = typeExprToString(fn.Return)
				}
				sig.Params = make([]string, len(fn.Params))
				for i, param := range fn.Params {
					sig.Params[i] = typeExprToString(param.Type)
				}
				c.functions[fn.Name] = sig
			}
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
