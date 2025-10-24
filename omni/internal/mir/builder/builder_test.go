package builder

import (
	"testing"

	"github.com/omni-lang/omni/internal/ast"
)

func TestBuildModule(t *testing.T) {
	// Test building a module from AST
	module := &ast.Module{
		Imports: []*ast.ImportDecl{},
		Decls:   []ast.Decl{},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if result == nil {
		t.Fatal("BuildModule returned nil")
	}

	if len(result.Functions) != 0 {
		t.Errorf("Expected 0 functions initially, got %d", len(result.Functions))
	}
}

func TestBuildModuleWithFunction(t *testing.T) {
	// Test building a module with a function
	funcDecl := &ast.FuncDecl{
		Name: "test_func",
		Params: []ast.Param{
			{Name: "x", Type: &ast.TypeExpr{Name: "int"}},
		},
		Return: &ast.TypeExpr{Name: "int"},
		Body: &ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.ReturnStmt{
					Value: &ast.LiteralExpr{Kind: ast.LiteralInt, Value: "42"},
				},
			},
		},
	}

	module := &ast.Module{
		Imports: []*ast.ImportDecl{},
		Decls:   []ast.Decl{funcDecl},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if result == nil {
		t.Fatal("BuildModule returned nil")
	}

	if len(result.Functions) != 1 {
		t.Errorf("Expected 1 function, got %d", len(result.Functions))
	}

	if result.Functions[0].Name != "test_func" {
		t.Errorf("Expected function name 'test_func', got '%s'", result.Functions[0].Name)
	}
}

func TestBuildModuleWithImport(t *testing.T) {
	// Test building a module with imports
	importDecl := &ast.ImportDecl{
		Path:  []string{"std", "io"},
		Alias: "io",
	}

	module := &ast.Module{
		Imports: []*ast.ImportDecl{importDecl},
		Decls:   []ast.Decl{},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if result == nil {
		t.Fatal("BuildModule returned nil")
	}

	// Should have no functions but should process imports
	if len(result.Functions) != 0 {
		t.Errorf("Expected 0 functions, got %d", len(result.Functions))
	}
}

func TestBuildModuleWithStruct(t *testing.T) {
	// Test building a module with a struct
	structDecl := &ast.StructDecl{
		Name: "Point",
		Fields: []ast.StructField{
			{Name: "x", Type: &ast.TypeExpr{Name: "int"}},
			{Name: "y", Type: &ast.TypeExpr{Name: "int"}},
		},
	}

	module := &ast.Module{
		Imports: []*ast.ImportDecl{},
		Decls:   []ast.Decl{structDecl},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if result == nil {
		t.Fatal("BuildModule returned nil")
	}

	// Should have no functions but should process struct
	if len(result.Functions) != 0 {
		t.Errorf("Expected 0 functions, got %d", len(result.Functions))
	}
}

func TestBuildModuleWithEnum(t *testing.T) {
	// Test building a module with an enum
	enumDecl := &ast.EnumDecl{
		Name: "Color",
		Variants: []ast.EnumVariant{
			{Name: "Red"},
			{Name: "Green"},
			{Name: "Blue"},
		},
	}

	module := &ast.Module{
		Imports: []*ast.ImportDecl{},
		Decls:   []ast.Decl{enumDecl},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if result == nil {
		t.Fatal("BuildModule returned nil")
	}

	// Should have no functions but should process enum
	if len(result.Functions) != 0 {
		t.Errorf("Expected 0 functions, got %d", len(result.Functions))
	}
}

func TestBuildModuleWithVariable(t *testing.T) {
	// Test building a module with a variable declaration
	varDecl := &ast.VarDecl{
		Name:  "x",
		Type:  &ast.TypeExpr{Name: "int"},
		Value: &ast.LiteralExpr{Kind: ast.LiteralInt, Value: "42"},
	}

	module := &ast.Module{
		Imports: []*ast.ImportDecl{},
		Decls:   []ast.Decl{varDecl},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if result == nil {
		t.Fatal("BuildModule returned nil")
	}

	// Should have no functions but should process variable
	if len(result.Functions) != 0 {
		t.Errorf("Expected 0 functions, got %d", len(result.Functions))
	}
}

func TestBuildModuleWithLet(t *testing.T) {
	// Test building a module with a let declaration
	letDecl := &ast.LetDecl{
		Name:  "y",
		Type:  &ast.TypeExpr{Name: "string"},
		Value: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "hello"},
	}

	module := &ast.Module{
		Imports: []*ast.ImportDecl{},
		Decls:   []ast.Decl{letDecl},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if result == nil {
		t.Fatal("BuildModule returned nil")
	}

	// Should have no functions but should process let
	if len(result.Functions) != 0 {
		t.Errorf("Expected 0 functions, got %d", len(result.Functions))
	}
}

func TestBuildModuleWithTypeAlias(t *testing.T) {
	// Test building a module with a type alias
	typeAliasDecl := &ast.TypeAliasDecl{
		Name: "MyInt",
		Type: ast.TypeExpr{Name: "int"},
	}

	module := &ast.Module{
		Imports: []*ast.ImportDecl{},
		Decls:   []ast.Decl{typeAliasDecl},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if result == nil {
		t.Fatal("BuildModule returned nil")
	}

	// Should have no functions but should process type alias
	if len(result.Functions) != 0 {
		t.Errorf("Expected 0 functions, got %d", len(result.Functions))
	}
}

func TestBuildModuleWithMultipleDeclarations(t *testing.T) {
	// Test building a module with multiple declarations
	funcDecl := &ast.FuncDecl{
		Name: "add",
		Params: []ast.Param{
			{Name: "a", Type: &ast.TypeExpr{Name: "int"}},
			{Name: "b", Type: &ast.TypeExpr{Name: "int"}},
		},
		Return: &ast.TypeExpr{Name: "int"},
		Body: &ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.ReturnStmt{
					Value: &ast.BinaryExpr{
						Left:  &ast.IdentifierExpr{Name: "a"},
						Op:    "+",
						Right: &ast.IdentifierExpr{Name: "b"},
					},
				},
			},
		},
	}

	varDecl := &ast.VarDecl{
		Name:  "x",
		Type:  &ast.TypeExpr{Name: "int"},
		Value: &ast.LiteralExpr{Kind: ast.LiteralInt, Value: "42"},
	}

	module := &ast.Module{
		Imports: []*ast.ImportDecl{},
		Decls:   []ast.Decl{funcDecl, varDecl},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if result == nil {
		t.Fatal("BuildModule returned nil")
	}

	// Should have 1 function
	if len(result.Functions) != 1 {
		t.Errorf("Expected 1 function, got %d", len(result.Functions))
	}

	if result.Functions[0].Name != "add" {
		t.Errorf("Expected function name 'add', got '%s'", result.Functions[0].Name)
	}
}
