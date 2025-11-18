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
		Type: &ast.TypeExpr{Name: "int"},
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

// Test statement lowering functions

func TestLowerIfStmt(t *testing.T) {
	// Test lowering if statement
	module := &ast.Module{
		Decls: []ast.Decl{
			&ast.FuncDecl{
				Name: "test",
				Return: &ast.TypeExpr{Name: "int"},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.IfStmt{
							Cond: &ast.LiteralExpr{Kind: ast.LiteralBool, Value: "true"},
							Then: &ast.BlockStmt{
								Statements: []ast.Stmt{
									&ast.ReturnStmt{Value: &ast.LiteralExpr{Kind: ast.LiteralInt, Value: "1"}},
								},
							},
							Else: &ast.BlockStmt{
								Statements: []ast.Stmt{
									&ast.ReturnStmt{Value: &ast.LiteralExpr{Kind: ast.LiteralInt, Value: "0"}},
								},
							},
						},
					},
				},
			},
		},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if len(result.Functions) != 1 {
		t.Fatalf("Expected 1 function, got %d", len(result.Functions))
	}

	// Should have multiple blocks for if/else
	if len(result.Functions[0].Blocks) < 2 {
		t.Errorf("Expected at least 2 blocks for if/else, got %d", len(result.Functions[0].Blocks))
	}
}

func TestLowerForStmtRange(t *testing.T) {
	// Test lowering range for loop
	module := &ast.Module{
		Decls: []ast.Decl{
			&ast.FuncDecl{
				Name: "test",
				Return: &ast.TypeExpr{Name: "void"},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.BindingStmt{
							Mutable: false,
							Name:    "items",
							Type:    &ast.TypeExpr{Name: "array", Args: []*ast.TypeExpr{{Name: "int"}}},
							Value: &ast.ArrayLiteralExpr{
								Elements: []ast.Expr{
									&ast.LiteralExpr{Kind: ast.LiteralInt, Value: "1"},
									&ast.LiteralExpr{Kind: ast.LiteralInt, Value: "2"},
								},
							},
						},
						&ast.ForStmt{
							IsRange: true,
							Target:  &ast.IdentifierExpr{Name: "item"},
							Iterable: &ast.IdentifierExpr{Name: "items"},
							Body: &ast.BlockStmt{
								Statements: []ast.Stmt{
									&ast.ExprStmt{Expr: &ast.CallExpr{
										Callee: &ast.IdentifierExpr{Name: "println"},
										Args:   []ast.Expr{&ast.IdentifierExpr{Name: "item"}},
									}},
								},
							},
						},
					},
				},
			},
		},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if len(result.Functions) != 1 {
		t.Fatalf("Expected 1 function, got %d", len(result.Functions))
	}
}

func TestLowerForStmtClassic(t *testing.T) {
	// Test lowering classic for loop
	module := &ast.Module{
		Decls: []ast.Decl{
			&ast.FuncDecl{
				Name: "test",
				Return: &ast.TypeExpr{Name: "void"},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.ForStmt{
							IsRange: false,
							Init: &ast.BindingStmt{
								Mutable: true,
								Name:    "i",
								Type:    &ast.TypeExpr{Name: "int"},
								Value:   &ast.LiteralExpr{Kind: ast.LiteralInt, Value: "0"},
							},
							Condition: &ast.BinaryExpr{
								Left:  &ast.IdentifierExpr{Name: "i"},
								Op:    "<",
								Right: &ast.LiteralExpr{Kind: ast.LiteralInt, Value: "10"},
							},
							Post: &ast.IncrementStmt{
								Target: &ast.IdentifierExpr{Name: "i"},
								Op:     "++",
							},
							Body: &ast.BlockStmt{
								Statements: []ast.Stmt{},
							},
						},
					},
				},
			},
		},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if len(result.Functions) != 1 {
		t.Fatalf("Expected 1 function, got %d", len(result.Functions))
	}
}

func TestLowerWhileStmt(t *testing.T) {
	// Test lowering while statement
	module := &ast.Module{
		Decls: []ast.Decl{
			&ast.FuncDecl{
				Name: "test",
				Return: &ast.TypeExpr{Name: "void"},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.WhileStmt{
							Cond: &ast.LiteralExpr{Kind: ast.LiteralBool, Value: "true"},
							Body: &ast.BlockStmt{
								Statements: []ast.Stmt{
									&ast.BreakStmt{},
								},
							},
						},
					},
				},
			},
		},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if len(result.Functions) != 1 {
		t.Fatalf("Expected 1 function, got %d", len(result.Functions))
	}
}

func TestLowerBreakContinueStmt(t *testing.T) {
	// Test lowering break and continue statements
	module := &ast.Module{
		Decls: []ast.Decl{
			&ast.FuncDecl{
				Name: "test",
				Return: &ast.TypeExpr{Name: "void"},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.WhileStmt{
							Cond: &ast.LiteralExpr{Kind: ast.LiteralBool, Value: "true"},
							Body: &ast.BlockStmt{
								Statements: []ast.Stmt{
									&ast.BreakStmt{},
									&ast.ContinueStmt{},
								},
							},
						},
					},
				},
			},
		},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if len(result.Functions) != 1 {
		t.Fatalf("Expected 1 function, got %d", len(result.Functions))
	}
}

func TestLowerTryStmt(t *testing.T) {
	// Test lowering try-catch statement
	module := &ast.Module{
		Decls: []ast.Decl{
			&ast.FuncDecl{
				Name: "test",
				Return: &ast.TypeExpr{Name: "int"},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.TryStmt{
							TryBlock: &ast.BlockStmt{
								Statements: []ast.Stmt{
									&ast.ThrowStmt{Expr: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "\"error\""}},
								},
							},
							CatchClauses: []*ast.CatchClause{
								{
									ExceptionVar:  "e",
									ExceptionType: "Error",
									Block: &ast.BlockStmt{
										Statements: []ast.Stmt{
											&ast.ReturnStmt{Value: &ast.LiteralExpr{Kind: ast.LiteralInt, Value: "0"}},
										},
									},
								},
							},
							FinallyBlock: &ast.BlockStmt{
								Statements: []ast.Stmt{
									&ast.ReturnStmt{Value: &ast.LiteralExpr{Kind: ast.LiteralInt, Value: "1"}},
								},
							},
						},
					},
				},
			},
		},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if len(result.Functions) != 1 {
		t.Fatalf("Expected 1 function, got %d", len(result.Functions))
	}
}

func TestLowerIncrementStmt(t *testing.T) {
	// Test lowering increment statement
	module := &ast.Module{
		Decls: []ast.Decl{
			&ast.FuncDecl{
				Name: "test",
				Return: &ast.TypeExpr{Name: "void"},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.BindingStmt{
							Mutable: true,
							Name:    "x",
							Type:    &ast.TypeExpr{Name: "int"},
							Value:   &ast.LiteralExpr{Kind: ast.LiteralInt, Value: "0"},
						},
						&ast.IncrementStmt{
							Target: &ast.IdentifierExpr{Name: "x"},
							Op:     "++",
						},
					},
				},
			},
		},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if len(result.Functions) != 1 {
		t.Fatalf("Expected 1 function, got %d", len(result.Functions))
	}
}

// Test expression lowering functions

func TestEmitBinaryExpr(t *testing.T) {
	// Test emitting binary expressions
	module := &ast.Module{
		Decls: []ast.Decl{
			&ast.FuncDecl{
				Name: "test",
				Return: &ast.TypeExpr{Name: "int"},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.ReturnStmt{
							Value: &ast.BinaryExpr{
								Left:  &ast.LiteralExpr{Kind: ast.LiteralInt, Value: "1"},
								Op:    "+",
								Right: &ast.LiteralExpr{Kind: ast.LiteralInt, Value: "2"},
							},
						},
					},
				},
			},
		},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if len(result.Functions) != 1 {
		t.Fatalf("Expected 1 function, got %d", len(result.Functions))
	}
}

func TestEmitUnaryExpr(t *testing.T) {
	// Test emitting unary expressions
	module := &ast.Module{
		Decls: []ast.Decl{
			&ast.FuncDecl{
				Name: "test",
				Return: &ast.TypeExpr{Name: "int"},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.ReturnStmt{
							Value: &ast.UnaryExpr{
								Op:   "-",
								Expr: &ast.LiteralExpr{Kind: ast.LiteralInt, Value: "42"},
							},
						},
					},
				},
			},
		},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if len(result.Functions) != 1 {
		t.Fatalf("Expected 1 function, got %d", len(result.Functions))
	}
}

func TestEmitCallExpr(t *testing.T) {
	// Test emitting call expressions
	module := &ast.Module{
		Decls: []ast.Decl{
			&ast.FuncDecl{
				Name: "test",
				Return: &ast.TypeExpr{Name: "void"},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.ExprStmt{
							Expr: &ast.CallExpr{
								Callee: &ast.IdentifierExpr{Name: "println"},
								Args: []ast.Expr{
									&ast.LiteralExpr{Kind: ast.LiteralString, Value: "\"hello\""},
								},
							},
						},
					},
				},
			},
		},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if len(result.Functions) != 1 {
		t.Fatalf("Expected 1 function, got %d", len(result.Functions))
	}
}

func TestEmitMemberAccess(t *testing.T) {
	// Test emitting member access expressions
	module := &ast.Module{
		Decls: []ast.Decl{
			&ast.StructDecl{
				Name: "Point",
				Fields: []ast.StructField{
					{Name: "x", Type: &ast.TypeExpr{Name: "int"}},
				},
			},
			&ast.FuncDecl{
				Name: "test",
				Return: &ast.TypeExpr{Name: "int"},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.BindingStmt{
							Mutable: false,
							Name:    "p",
							Type:    &ast.TypeExpr{Name: "Point"},
							Value: &ast.StructLiteralExpr{
								TypeName: "Point",
								Fields: []ast.StructLiteralField{
									{Name: "x", Expr: &ast.LiteralExpr{Kind: ast.LiteralInt, Value: "1"}},
								},
							},
						},
						&ast.ReturnStmt{
							Value: &ast.MemberExpr{
								Target: &ast.IdentifierExpr{Name: "p"},
								Member: "x",
							},
						},
					},
				},
			},
		},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if len(result.Functions) != 1 {
		t.Fatalf("Expected 1 function, got %d", len(result.Functions))
	}
}

func TestEmitArrayLiteral(t *testing.T) {
	// Test emitting array literal expressions
	module := &ast.Module{
		Decls: []ast.Decl{
			&ast.FuncDecl{
				Name: "test",
				Return: &ast.TypeExpr{Name: "array", Args: []*ast.TypeExpr{{Name: "int"}}},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.ReturnStmt{
							Value: &ast.ArrayLiteralExpr{
								Elements: []ast.Expr{
									&ast.LiteralExpr{Kind: ast.LiteralInt, Value: "1"},
									&ast.LiteralExpr{Kind: ast.LiteralInt, Value: "2"},
									&ast.LiteralExpr{Kind: ast.LiteralInt, Value: "3"},
								},
							},
						},
					},
				},
			},
		},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if len(result.Functions) != 1 {
		t.Fatalf("Expected 1 function, got %d", len(result.Functions))
	}
}

func TestEmitMapLiteral(t *testing.T) {
	// Test emitting map literal expressions
	module := &ast.Module{
		Decls: []ast.Decl{
			&ast.FuncDecl{
				Name: "test",
				Return: &ast.TypeExpr{Name: "map", Args: []*ast.TypeExpr{{Name: "string"}, {Name: "int"}}},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.ReturnStmt{
							Value: &ast.MapLiteralExpr{
								Entries: []ast.MapEntry{
									{
										Key:   &ast.LiteralExpr{Kind: ast.LiteralString, Value: "\"a\""},
										Value: &ast.LiteralExpr{Kind: ast.LiteralInt, Value: "1"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if len(result.Functions) != 1 {
		t.Fatalf("Expected 1 function, got %d", len(result.Functions))
	}
}

func TestEmitStructLiteral(t *testing.T) {
	// Test emitting struct literal expressions
	module := &ast.Module{
		Decls: []ast.Decl{
			&ast.StructDecl{
				Name: "Point",
				Fields: []ast.StructField{
					{Name: "x", Type: &ast.TypeExpr{Name: "int"}},
					{Name: "y", Type: &ast.TypeExpr{Name: "int"}},
				},
			},
			&ast.FuncDecl{
				Name: "test",
				Return: &ast.TypeExpr{Name: "Point"},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.ReturnStmt{
							Value: &ast.StructLiteralExpr{
								TypeName: "Point",
								Fields: []ast.StructLiteralField{
									{Name: "x", Expr: &ast.LiteralExpr{Kind: ast.LiteralInt, Value: "1"}},
									{Name: "y", Expr: &ast.LiteralExpr{Kind: ast.LiteralInt, Value: "2"}},
								},
							},
						},
					},
				},
			},
		},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if len(result.Functions) != 1 {
		t.Fatalf("Expected 1 function, got %d", len(result.Functions))
	}
}

func TestEmitIndexAccess(t *testing.T) {
	// Test emitting index access expressions
	module := &ast.Module{
		Decls: []ast.Decl{
			&ast.FuncDecl{
				Name: "test",
				Return: &ast.TypeExpr{Name: "int"},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.BindingStmt{
							Mutable: false,
							Name:    "arr",
							Type:    &ast.TypeExpr{Name: "array", Args: []*ast.TypeExpr{{Name: "int"}}},
							Value: &ast.ArrayLiteralExpr{
								Elements: []ast.Expr{
									&ast.LiteralExpr{Kind: ast.LiteralInt, Value: "1"},
								},
							},
						},
						&ast.ReturnStmt{
							Value: &ast.IndexExpr{
								Target: &ast.IdentifierExpr{Name: "arr"},
								Index:  &ast.LiteralExpr{Kind: ast.LiteralInt, Value: "0"},
							},
						},
					},
				},
			},
		},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if len(result.Functions) != 1 {
		t.Fatalf("Expected 1 function, got %d", len(result.Functions))
	}
}

func TestEmitLambda(t *testing.T) {
	// Test emitting lambda expressions
	module := &ast.Module{
		Decls: []ast.Decl{
			&ast.FuncDecl{
				Name: "test",
				Return: &ast.TypeExpr{
					IsFunction: true,
					ParamTypes: []*ast.TypeExpr{{Name: "int"}, {Name: "int"}},
					ReturnType: &ast.TypeExpr{Name: "int"},
				},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.ReturnStmt{
							Value: &ast.LambdaExpr{
								Params: []ast.Param{
									{Name: "a", Type: &ast.TypeExpr{Name: "int"}},
									{Name: "b", Type: &ast.TypeExpr{Name: "int"}},
								},
								Body: &ast.BinaryExpr{
									Left:  &ast.IdentifierExpr{Name: "a"},
									Op:    "+",
									Right: &ast.IdentifierExpr{Name: "b"},
								},
							},
						},
					},
				},
			},
		},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	// Should have the main function plus lambda function
	if len(result.Functions) < 1 {
		t.Fatalf("Expected at least 1 function, got %d", len(result.Functions))
	}
}

func TestEmitCast(t *testing.T) {
	// Test emitting cast expressions
	module := &ast.Module{
		Decls: []ast.Decl{
			&ast.FuncDecl{
				Name: "test",
				Return: &ast.TypeExpr{Name: "int"},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.ReturnStmt{
							Value: &ast.CastExpr{
								Type: &ast.TypeExpr{Name: "int"},
								Expr: &ast.LiteralExpr{Kind: ast.LiteralFloat, Value: "3.14"},
							},
						},
					},
				},
			},
		},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if len(result.Functions) != 1 {
		t.Fatalf("Expected 1 function, got %d", len(result.Functions))
	}
}

func TestEmitStringInterpolation(t *testing.T) {
	// Test emitting string interpolation expressions
	module := &ast.Module{
		Decls: []ast.Decl{
			&ast.FuncDecl{
				Name: "test",
				Return: &ast.TypeExpr{Name: "string"},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.BindingStmt{
							Mutable: false,
							Name:    "name",
							Type:    &ast.TypeExpr{Name: "string"},
							Value:   &ast.LiteralExpr{Kind: ast.LiteralString, Value: "\"World\""},
						},
						&ast.ReturnStmt{
							Value: &ast.StringInterpolationExpr{
								Parts: []ast.StringInterpolationPart{
									{IsLiteral: true, Literal: "Hello, "},
									{IsLiteral: false, Expr: &ast.IdentifierExpr{Name: "name"}},
									{IsLiteral: true, Literal: "!"},
								},
							},
						},
					},
				},
			},
		},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if len(result.Functions) != 1 {
		t.Fatalf("Expected 1 function, got %d", len(result.Functions))
	}
}

func TestEmitAwait(t *testing.T) {
	// Test emitting await expressions
	module := &ast.Module{
		Decls: []ast.Decl{
			&ast.FuncDecl{
				Name:    "test",
				IsAsync: true,
				Return:  &ast.TypeExpr{Name: "int"},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.ReturnStmt{
							Value: &ast.AwaitExpr{
								Expr: &ast.CallExpr{
									Callee: &ast.IdentifierExpr{Name: "asyncFunc"},
									Args:   []ast.Expr{},
								},
							},
						},
					},
				},
			},
		},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if len(result.Functions) != 1 {
		t.Fatalf("Expected 1 function, got %d", len(result.Functions))
	}
}

func TestEmitAllBinaryOps(t *testing.T) {
	// Test emitting all binary operators
	ops := []string{"+", "-", "*", "/", "%", "==", "!=", "<", ">", "<=", ">=", "&&", "||"}

	for _, op := range ops {
		module := &ast.Module{
			Decls: []ast.Decl{
				&ast.FuncDecl{
					Name: "test",
					Return: &ast.TypeExpr{Name: "int"},
					Body: &ast.BlockStmt{
						Statements: []ast.Stmt{
							&ast.ReturnStmt{
								Value: &ast.BinaryExpr{
									Left:  &ast.LiteralExpr{Kind: ast.LiteralInt, Value: "1"},
									Op:    op,
									Right: &ast.LiteralExpr{Kind: ast.LiteralInt, Value: "2"},
								},
							},
						},
					},
				},
			},
		}

		result, err := BuildModule(module)
		if err != nil {
			t.Fatalf("BuildModule failed for op %s: %v", op, err)
		}

		if len(result.Functions) != 1 {
			t.Fatalf("Expected 1 function for op %s, got %d", op, len(result.Functions))
		}
	}
}

func TestEmitAllUnaryOps(t *testing.T) {
	// Test emitting all unary operators
	ops := []string{"-", "!", "~"}

	for _, op := range ops {
		module := &ast.Module{
			Decls: []ast.Decl{
				&ast.FuncDecl{
					Name: "test",
					Return: &ast.TypeExpr{Name: "int"},
					Body: &ast.BlockStmt{
						Statements: []ast.Stmt{
							&ast.ReturnStmt{
								Value: &ast.UnaryExpr{
									Op:   op,
									Expr: &ast.LiteralExpr{Kind: ast.LiteralInt, Value: "42"},
								},
							},
						},
					},
				},
			},
		}

		result, err := BuildModule(module)
		if err != nil {
			t.Fatalf("BuildModule failed for op %s: %v", op, err)
		}

		if len(result.Functions) != 1 {
			t.Fatalf("Expected 1 function for op %s, got %d", op, len(result.Functions))
		}
	}
}

// Test helper functions

func TestMapBinaryOp(t *testing.T) {
	// Test mapBinaryOp helper function
	testCases := []struct {
		op     string
		expect string
	}{
		{"+", "add"},
		{"-", "sub"},
		{"*", "mul"},
		{"/", "div"},
		{"%", "mod"},
		{"==", "cmp.eq"},
		{"!=", "cmp.neq"},
		{"<", "cmp.lt"},
		{">", "cmp.gt"},
		{"<=", "cmp.lte"},
		{">=", "cmp.gte"},
		{"&&", "and"},
		{"||", "or"},
	}

	for _, tc := range testCases {
		result := mapBinaryOp(tc.op)
		if result != tc.expect {
			t.Errorf("mapBinaryOp(%q) = %q, expected %q", tc.op, result, tc.expect)
		}
	}
}

func TestIsComparison(t *testing.T) {
	// Test isComparison helper function
	comparisonOps := []string{"==", "!=", "<", ">", "<=", ">="}
	nonComparisonOps := []string{"+", "-", "*", "/", "%", "&&", "||"}

	for _, op := range comparisonOps {
		if !isComparison(op) {
			t.Errorf("Expected isComparison(%q) to be true", op)
		}
	}

	for _, op := range nonComparisonOps {
		if isComparison(op) {
			t.Errorf("Expected isComparison(%q) to be false", op)
		}
	}
}

func TestIsLogical(t *testing.T) {
	// Test isLogical helper function
	logicalOps := []string{"&&", "||"}
	nonLogicalOps := []string{"+", "-", "*", "/", "%", "==", "!=", "<", ">"}

	for _, op := range logicalOps {
		if !isLogical(op) {
			t.Errorf("Expected isLogical(%q) to be true", op)
		}
	}

	for _, op := range nonLogicalOps {
		if isLogical(op) {
			t.Errorf("Expected isLogical(%q) to be false", op)
		}
	}
}

func TestTypeExprToString(t *testing.T) {
	// Test typeExprToString helper function
	testCases := []struct {
		expr   *ast.TypeExpr
		expect string
	}{
		{&ast.TypeExpr{Name: "int"}, "int"},
		{&ast.TypeExpr{Name: "string"}, "string"},
		{
			&ast.TypeExpr{
				Name: "array",
				Args: []*ast.TypeExpr{{Name: "int"}},
			},
			"array<int>",
		},
		{
			&ast.TypeExpr{
				Name: "map",
				Args: []*ast.TypeExpr{{Name: "string"}, {Name: "int"}},
			},
			"map<string,int>",
		},
		{
			&ast.TypeExpr{
				IsFunction: true,
				ParamTypes: []*ast.TypeExpr{{Name: "int"}, {Name: "string"}},
				ReturnType: &ast.TypeExpr{Name: "bool"},
			},
			"(int, string) -> bool",
		},
	}

	for _, tc := range testCases {
		result := typeExprToString(tc.expr)
		if result != tc.expect {
			t.Errorf("typeExprToString(%v) = %q, expected %q", tc.expr, result, tc.expect)
		}
	}
}

// Integration tests

func TestCompleteFunctionLowering(t *testing.T) {
	// Test complete function lowering with complex control flow
	module := &ast.Module{
		Decls: []ast.Decl{
			&ast.FuncDecl{
				Name: "complex",
				Params: []ast.Param{
					{Name: "x", Type: &ast.TypeExpr{Name: "int"}},
				},
				Return: &ast.TypeExpr{Name: "int"},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.IfStmt{
							Cond: &ast.BinaryExpr{
								Left:  &ast.IdentifierExpr{Name: "x"},
								Op:    ">",
								Right: &ast.LiteralExpr{Kind: ast.LiteralInt, Value: "0"},
							},
							Then: &ast.BlockStmt{
								Statements: []ast.Stmt{
									&ast.ReturnStmt{Value: &ast.IdentifierExpr{Name: "x"}},
								},
							},
							Else: &ast.BlockStmt{
								Statements: []ast.Stmt{
									&ast.ReturnStmt{
										Value: &ast.UnaryExpr{
											Op:   "-",
											Expr: &ast.IdentifierExpr{Name: "x"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if len(result.Functions) != 1 {
		t.Fatalf("Expected 1 function, got %d", len(result.Functions))
	}

	fn := result.Functions[0]
	if len(fn.Blocks) < 2 {
		t.Errorf("Expected at least 2 blocks, got %d", len(fn.Blocks))
	}
}

func TestAsyncFunctionLowering(t *testing.T) {
	// Test async function lowering
	module := &ast.Module{
		Decls: []ast.Decl{
			&ast.FuncDecl{
				Name:    "asyncFunc",
				IsAsync: true,
				Return:  &ast.TypeExpr{Name: "int"},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.ReturnStmt{
							Value: &ast.LiteralExpr{Kind: ast.LiteralInt, Value: "42"},
						},
					},
				},
			},
		},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if len(result.Functions) != 1 {
		t.Fatalf("Expected 1 function, got %d", len(result.Functions))
	}

	fn := result.Functions[0]
	if fn.ReturnType != "Promise<int>" {
		t.Errorf("Expected return type 'Promise<int>', got %q", fn.ReturnType)
	}
}

func TestFunctionWithExprBody(t *testing.T) {
	// Test function with expression body
	module := &ast.Module{
		Decls: []ast.Decl{
			&ast.FuncDecl{
				Name: "add",
				Params: []ast.Param{
					{Name: "a", Type: &ast.TypeExpr{Name: "int"}},
					{Name: "b", Type: &ast.TypeExpr{Name: "int"}},
				},
				Return: &ast.TypeExpr{Name: "int"},
				ExprBody: &ast.BinaryExpr{
					Left:  &ast.IdentifierExpr{Name: "a"},
					Op:    "+",
					Right: &ast.IdentifierExpr{Name: "b"},
				},
			},
		},
	}

	result, err := BuildModule(module)
	if err != nil {
		t.Fatalf("BuildModule failed: %v", err)
	}

	if len(result.Functions) != 1 {
		t.Fatalf("Expected 1 function, got %d", len(result.Functions))
	}

	fn := result.Functions[0]
	if len(fn.Blocks) != 1 {
		t.Errorf("Expected 1 block for expr body, got %d", len(fn.Blocks))
	}
}
