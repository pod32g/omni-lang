package ast

import (
	"testing"

	"github.com/omni-lang/omni/internal/lexer"
)

func TestSpan(t *testing.T) {
	// Test Span creation
	span := lexer.Span{
		Start: lexer.Position{Line: 1, Column: 1},
		End:   lexer.Position{Line: 1, Column: 10},
	}

	if span.Start.Line != 1 {
		t.Errorf("Expected Start.Line 1, got %d", span.Start.Line)
	}

	if span.Start.Column != 1 {
		t.Errorf("Expected Start.Column 1, got %d", span.Start.Column)
	}

	if span.End.Line != 1 {
		t.Errorf("Expected End.Line 1, got %d", span.End.Line)
	}

	if span.End.Column != 10 {
		t.Errorf("Expected End.Column 10, got %d", span.End.Column)
	}
}

func TestPosition(t *testing.T) {
	// Test Position creation
	pos := lexer.Position{Line: 5, Column: 20}

	if pos.Line != 5 {
		t.Errorf("Expected Line 5, got %d", pos.Line)
	}

	if pos.Column != 20 {
		t.Errorf("Expected Column 20, got %d", pos.Column)
	}
}

func TestModule(t *testing.T) {
	// Test Module creation
	module := &Module{
		Imports: []*ImportDecl{},
		Decls:   []Decl{},
	}

	if module == nil {
		t.Fatal("Expected non-nil module")
	}

	if len(module.Imports) != 0 {
		t.Errorf("Expected 0 imports, got %d", len(module.Imports))
	}

	if len(module.Decls) != 0 {
		t.Errorf("Expected 0 declarations, got %d", len(module.Decls))
	}
}

func TestImportDecl(t *testing.T) {
	// Test ImportDecl creation
	importDecl := &ImportDecl{
		Path:  []string{"std", "io"},
		Alias: "io",
	}

	if importDecl == nil {
		t.Fatal("Expected non-nil import declaration")
	}

	if len(importDecl.Path) != 2 {
		t.Errorf("Expected 2 path elements, got %d", len(importDecl.Path))
	}

	if importDecl.Alias != "io" {
		t.Errorf("Expected alias 'io', got '%s'", importDecl.Alias)
	}
}

func TestLetDecl(t *testing.T) {
	// Test LetDecl creation
	letDecl := &LetDecl{
		Name:  "x",
		Type:  &TypeExpr{Name: "int"},
		Value: &LiteralExpr{Kind: LiteralInt, Value: "42"},
	}

	if letDecl == nil {
		t.Fatal("Expected non-nil let declaration")
	}

	if letDecl.Name != "x" {
		t.Errorf("Expected name 'x', got '%s'", letDecl.Name)
	}

	if letDecl.Type == nil {
		t.Fatal("Expected non-nil type")
	}

	if letDecl.Value == nil {
		t.Fatal("Expected non-nil value")
	}
}

func TestVarDecl(t *testing.T) {
	// Test VarDecl creation
	varDecl := &VarDecl{
		Name:  "y",
		Type:  &TypeExpr{Name: "string"},
		Value: &LiteralExpr{Kind: LiteralString, Value: "hello"},
	}

	if varDecl == nil {
		t.Fatal("Expected non-nil var declaration")
	}

	if varDecl.Name != "y" {
		t.Errorf("Expected name 'y', got '%s'", varDecl.Name)
	}

	if varDecl.Type == nil {
		t.Fatal("Expected non-nil type")
	}

	if varDecl.Value == nil {
		t.Fatal("Expected non-nil value")
	}
}

func TestStructDecl(t *testing.T) {
	// Test StructDecl creation
	structDecl := &StructDecl{
		Name: "Point",
		Fields: []StructField{
			{Name: "x", Type: &TypeExpr{Name: "int"}},
			{Name: "y", Type: &TypeExpr{Name: "int"}},
		},
	}

	if structDecl == nil {
		t.Fatal("Expected non-nil struct declaration")
	}

	if structDecl.Name != "Point" {
		t.Errorf("Expected name 'Point', got '%s'", structDecl.Name)
	}

	if len(structDecl.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(structDecl.Fields))
	}
}

func TestEnumDecl(t *testing.T) {
	// Test EnumDecl creation
	enumDecl := &EnumDecl{
		Name: "Color",
		Variants: []EnumVariant{
			{Name: "Red"},
			{Name: "Green"},
			{Name: "Blue"},
		},
	}

	if enumDecl == nil {
		t.Fatal("Expected non-nil enum declaration")
	}

	if enumDecl.Name != "Color" {
		t.Errorf("Expected name 'Color', got '%s'", enumDecl.Name)
	}

	if len(enumDecl.Variants) != 3 {
		t.Errorf("Expected 3 variants, got %d", len(enumDecl.Variants))
	}
}

func TestFuncDecl(t *testing.T) {
	// Test FuncDecl creation
	funcDecl := &FuncDecl{
		Name: "add",
		Params: []Param{
			{Name: "a", Type: &TypeExpr{Name: "int"}},
			{Name: "b", Type: &TypeExpr{Name: "int"}},
		},
		Return: &TypeExpr{Name: "int"},
		Body:   &BlockStmt{},
	}

	if funcDecl == nil {
		t.Fatal("Expected non-nil function declaration")
	}

	if funcDecl.Name != "add" {
		t.Errorf("Expected name 'add', got '%s'", funcDecl.Name)
	}

	if len(funcDecl.Params) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(funcDecl.Params))
	}

	if funcDecl.Return == nil {
		t.Fatal("Expected non-nil return type")
	}

	if funcDecl.Body == nil {
		t.Fatal("Expected non-nil body")
	}
}

func TestBlockStmt(t *testing.T) {
	// Test BlockStmt creation
	blockStmt := &BlockStmt{
		Statements: []Stmt{},
	}

	if blockStmt == nil {
		t.Fatal("Expected non-nil block statement")
	}

	if len(blockStmt.Statements) != 0 {
		t.Errorf("Expected 0 statements, got %d", len(blockStmt.Statements))
	}
}

func TestReturnStmt(t *testing.T) {
	// Test ReturnStmt creation
	returnStmt := &ReturnStmt{
		Value: &LiteralExpr{Kind: LiteralInt, Value: "42"},
	}

	if returnStmt == nil {
		t.Fatal("Expected non-nil return statement")
	}

	if returnStmt.Value == nil {
		t.Fatal("Expected non-nil value")
	}
}

func TestExprStmt(t *testing.T) {
	// Test ExprStmt creation
	exprStmt := &ExprStmt{
		Expr: &LiteralExpr{Kind: LiteralInt, Value: "42"},
	}

	if exprStmt == nil {
		t.Fatal("Expected non-nil expression statement")
	}

	if exprStmt.Expr == nil {
		t.Fatal("Expected non-nil expression")
	}
}

func TestLiteralExpr(t *testing.T) {
	// Test LiteralExpr creation
	literalExpr := &LiteralExpr{
		Kind:  LiteralInt,
		Value: "42",
	}

	if literalExpr == nil {
		t.Fatal("Expected non-nil literal expression")
	}

	if literalExpr.Kind != LiteralInt {
		t.Errorf("Expected LiteralInt, got %v", literalExpr.Kind)
	}

	if literalExpr.Value != "42" {
		t.Errorf("Expected value '42', got '%s'", literalExpr.Value)
	}
}

func TestIdentifierExpr(t *testing.T) {
	// Test IdentifierExpr creation
	identifierExpr := &IdentifierExpr{
		Name: "x",
	}

	if identifierExpr == nil {
		t.Fatal("Expected non-nil identifier expression")
	}

	if identifierExpr.Name != "x" {
		t.Errorf("Expected name 'x', got '%s'", identifierExpr.Name)
	}
}

func TestBinaryExpr(t *testing.T) {
	// Test BinaryExpr creation
	binaryExpr := &BinaryExpr{
		Left:  &LiteralExpr{Kind: LiteralInt, Value: "1"},
		Op:    "+",
		Right: &LiteralExpr{Kind: LiteralInt, Value: "2"},
	}

	if binaryExpr == nil {
		t.Fatal("Expected non-nil binary expression")
	}

	if binaryExpr.Op != "+" {
		t.Errorf("Expected op '+', got '%s'", binaryExpr.Op)
	}

	if binaryExpr.Left == nil {
		t.Fatal("Expected non-nil left expression")
	}

	if binaryExpr.Right == nil {
		t.Fatal("Expected non-nil right expression")
	}
}

func TestTypeExpr(t *testing.T) {
	// Test TypeExpr creation
	typeExpr := &TypeExpr{
		Name: "int",
	}

	if typeExpr == nil {
		t.Fatal("Expected non-nil type expression")
	}

	if typeExpr.Name != "int" {
		t.Errorf("Expected name 'int', got '%s'", typeExpr.Name)
	}
}
