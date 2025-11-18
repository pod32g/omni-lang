package ast

import (
	"strings"
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

func TestTypeExprWithGenerics(t *testing.T) {
	// Test TypeExpr with generic arguments
	typeExpr := &TypeExpr{
		Name: "List",
		Args: []*TypeExpr{
			{Name: "int"},
			{Name: "string"},
		},
	}

	if len(typeExpr.Args) != 2 {
		t.Errorf("Expected 2 generic arguments, got %d", len(typeExpr.Args))
	}

	if typeExpr.Args[0].Name != "int" {
		t.Errorf("Expected first arg 'int', got '%s'", typeExpr.Args[0].Name)
	}
}

func TestTypeExprWithUnion(t *testing.T) {
	// Test TypeExpr with union types
	typeExpr := &TypeExpr{
		IsUnion: true,
		Members: []*TypeExpr{
			{Name: "int"},
			{Name: "string"},
			{Name: "bool"},
		},
	}

	if !typeExpr.IsUnion {
		t.Error("Expected IsUnion to be true")
	}

	if len(typeExpr.Members) != 3 {
		t.Errorf("Expected 3 union members, got %d", len(typeExpr.Members))
	}
}

func TestTypeExprWithFunction(t *testing.T) {
	// Test TypeExpr with function type
	typeExpr := &TypeExpr{
		IsFunction: true,
		ParamTypes: []*TypeExpr{
			{Name: "int"},
			{Name: "string"},
		},
		ReturnType: &TypeExpr{Name: "bool"},
	}

	if !typeExpr.IsFunction {
		t.Error("Expected IsFunction to be true")
	}

	if len(typeExpr.ParamTypes) != 2 {
		t.Errorf("Expected 2 parameter types, got %d", len(typeExpr.ParamTypes))
	}

	if typeExpr.ReturnType == nil || typeExpr.ReturnType.Name != "bool" {
		t.Error("Expected return type 'bool'")
	}
}

func TestTypeExprWithOptional(t *testing.T) {
	// Test TypeExpr with optional type
	typeExpr := &TypeExpr{
		IsOptional:   true,
		OptionalType: &TypeExpr{Name: "int"},
	}

	if !typeExpr.IsOptional {
		t.Error("Expected IsOptional to be true")
	}

	if typeExpr.OptionalType == nil || typeExpr.OptionalType.Name != "int" {
		t.Error("Expected optional type 'int'")
	}
}

func TestStructField(t *testing.T) {
	// Test StructField creation
	field := StructField{
		Name: "x",
		Type: &TypeExpr{Name: "int"},
		Span: lexer.Span{Start: lexer.Position{Line: 1, Column: 1}, End: lexer.Position{Line: 1, Column: 10}},
	}

	if field.Name != "x" {
		t.Errorf("Expected name 'x', got '%s'", field.Name)
	}

	if field.Type == nil || field.Type.Name != "int" {
		t.Error("Expected type 'int'")
	}
}

func TestTypeParam(t *testing.T) {
	// Test TypeParam creation
	typeParam := TypeParam{
		Name: "T",
		Span: lexer.Span{Start: lexer.Position{Line: 1, Column: 1}, End: lexer.Position{Line: 1, Column: 2}},
	}

	if typeParam.Name != "T" {
		t.Errorf("Expected name 'T', got '%s'", typeParam.Name)
	}
}

func TestParam(t *testing.T) {
	// Test Param creation
	param := Param{
		Name: "x",
		Type: &TypeExpr{Name: "int"},
		Span: lexer.Span{Start: lexer.Position{Line: 1, Column: 1}, End: lexer.Position{Line: 1, Column: 10}},
	}

	if param.Name != "x" {
		t.Errorf("Expected name 'x', got '%s'", param.Name)
	}

	if param.Type == nil || param.Type.Name != "int" {
		t.Error("Expected type 'int'")
	}
}

func TestForStmtClassic(t *testing.T) {
	// Test classic for loop
	forStmt := &ForStmt{
		IsRange:   false,
		Init:      &ExprStmt{Expr: &LiteralExpr{Kind: LiteralInt, Value: "0"}},
		Condition: &LiteralExpr{Kind: LiteralBool, Value: "true"},
		Post:      &ExprStmt{Expr: &LiteralExpr{Kind: LiteralInt, Value: "1"}},
		Body:      &BlockStmt{},
	}

	if forStmt.IsRange {
		t.Error("Expected IsRange to be false for classic for loop")
	}

	if forStmt.Init == nil {
		t.Error("Expected non-nil Init")
	}

	if forStmt.Condition == nil {
		t.Error("Expected non-nil Condition")
	}
}

func TestForStmtRange(t *testing.T) {
	// Test range for loop
	forStmt := &ForStmt{
		IsRange: true,
		Target:  &IdentifierExpr{Name: "item"},
		Iterable: &IdentifierExpr{Name: "items"},
		Body:     &BlockStmt{},
	}

	if !forStmt.IsRange {
		t.Error("Expected IsRange to be true for range for loop")
	}

	if forStmt.Target == nil || forStmt.Target.Name != "item" {
		t.Error("Expected Target 'item'")
	}

	if forStmt.Iterable == nil {
		t.Error("Expected non-nil Iterable")
	}
}

func TestIfStmt(t *testing.T) {
	// Test IfStmt creation
	ifStmt := &IfStmt{
		Cond: &LiteralExpr{Kind: LiteralBool, Value: "true"},
		Then: &BlockStmt{},
		Else: nil,
	}

	if ifStmt.Cond == nil {
		t.Error("Expected non-nil condition")
	}

	if ifStmt.Then == nil {
		t.Error("Expected non-nil Then block")
	}

	// Test with else clause
	ifStmt.Else = &BlockStmt{}
	if ifStmt.Else == nil {
		t.Error("Expected non-nil Else block")
	}
}

func TestWhileStmt(t *testing.T) {
	// Test WhileStmt creation
	whileStmt := &WhileStmt{
		Cond: &LiteralExpr{Kind: LiteralBool, Value: "true"},
		Body: &BlockStmt{},
	}

	if whileStmt.Cond == nil {
		t.Error("Expected non-nil condition")
	}

	if whileStmt.Body == nil {
		t.Error("Expected non-nil body")
	}
}

func TestBreakStmt(t *testing.T) {
	// Test BreakStmt creation
	breakStmt := &BreakStmt{
		SpanInfo: lexer.Span{Start: lexer.Position{Line: 1, Column: 1}, End: lexer.Position{Line: 1, Column: 6}},
	}

	if breakStmt == nil {
		t.Fatal("Expected non-nil break statement")
	}

	var stmt Stmt = breakStmt
	if stmt == nil {
		t.Error("Expected BreakStmt to implement Stmt interface")
	}
}

func TestContinueStmt(t *testing.T) {
	// Test ContinueStmt creation
	continueStmt := &ContinueStmt{
		SpanInfo: lexer.Span{Start: lexer.Position{Line: 1, Column: 1}, End: lexer.Position{Line: 1, Column: 9}},
	}

	if continueStmt == nil {
		t.Fatal("Expected non-nil continue statement")
	}

	var stmt Stmt = continueStmt
	if stmt == nil {
		t.Error("Expected ContinueStmt to implement Stmt interface")
	}
}

func TestTryStmt(t *testing.T) {
	// Test TryStmt creation
	tryStmt := &TryStmt{
		TryBlock: &BlockStmt{},
		CatchClauses: []*CatchClause{
			{
				ExceptionVar:  "e",
				ExceptionType: "Error",
				Block:         &BlockStmt{},
			},
		},
		FinallyBlock: &BlockStmt{},
	}

	if tryStmt.TryBlock == nil {
		t.Error("Expected non-nil TryBlock")
	}

	if len(tryStmt.CatchClauses) != 1 {
		t.Errorf("Expected 1 catch clause, got %d", len(tryStmt.CatchClauses))
	}

	if tryStmt.FinallyBlock == nil {
		t.Error("Expected non-nil FinallyBlock")
	}
}

func TestCatchClause(t *testing.T) {
	// Test CatchClause creation
	catchClause := &CatchClause{
		ExceptionVar:  "e",
		ExceptionType: "Error",
		Block:         &BlockStmt{},
	}

	if catchClause.ExceptionVar != "e" {
		t.Errorf("Expected ExceptionVar 'e', got '%s'", catchClause.ExceptionVar)
	}

	if catchClause.ExceptionType != "Error" {
		t.Errorf("Expected ExceptionType 'Error', got '%s'", catchClause.ExceptionType)
	}

	if catchClause.Block == nil {
		t.Error("Expected non-nil Block")
	}
}

func TestThrowStmt(t *testing.T) {
	// Test ThrowStmt creation
	throwStmt := &ThrowStmt{
		Expr: &LiteralExpr{Kind: LiteralString, Value: "\"error\""},
	}

	if throwStmt.Expr == nil {
		t.Error("Expected non-nil expression")
	}

	var stmt Stmt = throwStmt
	if stmt == nil {
		t.Error("Expected ThrowStmt to implement Stmt interface")
	}
}

func TestTypeAliasDecl(t *testing.T) {
	// Test TypeAliasDecl creation
	typeAlias := &TypeAliasDecl{
		Name: "UserID",
		Type: &TypeExpr{Name: "int"},
	}

	if typeAlias.Name != "UserID" {
		t.Errorf("Expected name 'UserID', got '%s'", typeAlias.Name)
	}

	if typeAlias.Type == nil || typeAlias.Type.Name != "int" {
		t.Error("Expected type 'int'")
	}

	// Test with generic type parameters
	typeAlias.TypeParams = []string{"T"}
	if len(typeAlias.TypeParams) != 1 {
		t.Errorf("Expected 1 type parameter, got %d", len(typeAlias.TypeParams))
	}
}

func TestUnaryExpr(t *testing.T) {
	// Test UnaryExpr creation
	unaryExpr := &UnaryExpr{
		Op:   "-",
		Expr: &LiteralExpr{Kind: LiteralInt, Value: "42"},
	}

	if unaryExpr.Op != "-" {
		t.Errorf("Expected op '-', got '%s'", unaryExpr.Op)
	}

	if unaryExpr.Expr == nil {
		t.Error("Expected non-nil expression")
	}

	var expr Expr = unaryExpr
	if expr == nil {
		t.Error("Expected UnaryExpr to implement Expr interface")
	}
}

func TestCallExpr(t *testing.T) {
	// Test CallExpr creation
	callExpr := &CallExpr{
		Callee: &IdentifierExpr{Name: "println"},
		Args: []Expr{
			&LiteralExpr{Kind: LiteralString, Value: "\"hello\""},
			&LiteralExpr{Kind: LiteralInt, Value: "42"},
		},
	}

	if callExpr.Callee == nil {
		t.Error("Expected non-nil callee")
	}

	if len(callExpr.Args) != 2 {
		t.Errorf("Expected 2 arguments, got %d", len(callExpr.Args))
	}

	var expr Expr = callExpr
	if expr == nil {
		t.Error("Expected CallExpr to implement Expr interface")
	}
}

func TestIndexExpr(t *testing.T) {
	// Test IndexExpr creation
	indexExpr := &IndexExpr{
		Target: &IdentifierExpr{Name: "arr"},
		Index:  &LiteralExpr{Kind: LiteralInt, Value: "0"},
	}

	if indexExpr.Target == nil {
		t.Error("Expected non-nil target")
	}

	if indexExpr.Index == nil {
		t.Error("Expected non-nil index")
	}

	var expr Expr = indexExpr
	if expr == nil {
		t.Error("Expected IndexExpr to implement Expr interface")
	}
}

func TestMemberExpr(t *testing.T) {
	// Test MemberExpr creation
	memberExpr := &MemberExpr{
		Target: &IdentifierExpr{Name: "obj"},
		Member: "field",
	}

	if memberExpr.Target == nil {
		t.Error("Expected non-nil target")
	}

	if memberExpr.Member != "field" {
		t.Errorf("Expected member 'field', got '%s'", memberExpr.Member)
	}

	var expr Expr = memberExpr
	if expr == nil {
		t.Error("Expected MemberExpr to implement Expr interface")
	}
}

func TestArrayLiteralExpr(t *testing.T) {
	// Test ArrayLiteralExpr creation
	arrayExpr := &ArrayLiteralExpr{
		Elements: []Expr{
			&LiteralExpr{Kind: LiteralInt, Value: "1"},
			&LiteralExpr{Kind: LiteralInt, Value: "2"},
			&LiteralExpr{Kind: LiteralInt, Value: "3"},
		},
	}

	if len(arrayExpr.Elements) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(arrayExpr.Elements))
	}

	var expr Expr = arrayExpr
	if expr == nil {
		t.Error("Expected ArrayLiteralExpr to implement Expr interface")
	}
}

func TestMapLiteralExpr(t *testing.T) {
	// Test MapLiteralExpr creation
	mapExpr := &MapLiteralExpr{
		Entries: []MapEntry{
			{
				Key:   &LiteralExpr{Kind: LiteralString, Value: "\"key1\""},
				Value: &LiteralExpr{Kind: LiteralInt, Value: "1"},
			},
			{
				Key:   &LiteralExpr{Kind: LiteralString, Value: "\"key2\""},
				Value: &LiteralExpr{Kind: LiteralInt, Value: "2"},
			},
		},
	}

	if len(mapExpr.Entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(mapExpr.Entries))
	}

	var expr Expr = mapExpr
	if expr == nil {
		t.Error("Expected MapLiteralExpr to implement Expr interface")
	}
}

func TestStructLiteralExpr(t *testing.T) {
	// Test StructLiteralExpr creation
	structExpr := &StructLiteralExpr{
		TypeName: "Point",
		Fields: []StructLiteralField{
			{Name: "x", Expr: &LiteralExpr{Kind: LiteralInt, Value: "1"}},
			{Name: "y", Expr: &LiteralExpr{Kind: LiteralInt, Value: "2"}},
		},
	}

	if structExpr.TypeName != "Point" {
		t.Errorf("Expected type name 'Point', got '%s'", structExpr.TypeName)
	}

	if len(structExpr.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(structExpr.Fields))
	}

	var expr Expr = structExpr
	if expr == nil {
		t.Error("Expected StructLiteralExpr to implement Expr interface")
	}
}

func TestAssignmentExpr(t *testing.T) {
	// Test AssignmentExpr creation
	assignExpr := &AssignmentExpr{
		Left:  &IdentifierExpr{Name: "x"},
		Right: &LiteralExpr{Kind: LiteralInt, Value: "42"},
	}

	if assignExpr.Left == nil {
		t.Error("Expected non-nil left expression")
	}

	if assignExpr.Right == nil {
		t.Error("Expected non-nil right expression")
	}

	var expr Expr = assignExpr
	if expr == nil {
		t.Error("Expected AssignmentExpr to implement Expr interface")
	}
}

func TestIncrementExpr(t *testing.T) {
	// Test IncrementExpr creation
	incExpr := &IncrementExpr{
		Target: &IdentifierExpr{Name: "x"},
		Op:     "++",
	}

	if incExpr.Target == nil {
		t.Error("Expected non-nil target")
	}

	if incExpr.Op != "++" {
		t.Errorf("Expected op '++', got '%s'", incExpr.Op)
	}

	var expr Expr = incExpr
	if expr == nil {
		t.Error("Expected IncrementExpr to implement Expr interface")
	}
}

func TestNewExpr(t *testing.T) {
	// Test NewExpr creation
	newExpr := &NewExpr{
		Type: &TypeExpr{Name: "int"},
	}

	if newExpr.Type == nil || newExpr.Type.Name != "int" {
		t.Error("Expected type 'int'")
	}

	var expr Expr = newExpr
	if expr == nil {
		t.Error("Expected NewExpr to implement Expr interface")
	}
}

func TestDeleteExpr(t *testing.T) {
	// Test DeleteExpr creation
	deleteExpr := &DeleteExpr{
		Target: &IdentifierExpr{Name: "ptr"},
	}

	if deleteExpr.Target == nil {
		t.Error("Expected non-nil target")
	}

	var expr Expr = deleteExpr
	if expr == nil {
		t.Error("Expected DeleteExpr to implement Expr interface")
	}
}

func TestLambdaExpr(t *testing.T) {
	// Test LambdaExpr creation
	lambdaExpr := &LambdaExpr{
		Params: []Param{
			{Name: "a", Type: &TypeExpr{Name: "int"}},
			{Name: "b", Type: &TypeExpr{Name: "int"}},
		},
		Body: &BinaryExpr{
			Left:  &IdentifierExpr{Name: "a"},
			Op:    "+",
			Right: &IdentifierExpr{Name: "b"},
		},
	}

	if len(lambdaExpr.Params) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(lambdaExpr.Params))
	}

	if lambdaExpr.Body == nil {
		t.Error("Expected non-nil body")
	}

	var expr Expr = lambdaExpr
	if expr == nil {
		t.Error("Expected LambdaExpr to implement Expr interface")
	}
}

func TestCastExpr(t *testing.T) {
	// Test CastExpr creation
	castExpr := &CastExpr{
		Type: &TypeExpr{Name: "int"},
		Expr: &LiteralExpr{Kind: LiteralFloat, Value: "3.14"},
	}

	if castExpr.Type == nil || castExpr.Type.Name != "int" {
		t.Error("Expected type 'int'")
	}

	if castExpr.Expr == nil {
		t.Error("Expected non-nil expression")
	}

	var expr Expr = castExpr
	if expr == nil {
		t.Error("Expected CastExpr to implement Expr interface")
	}
}

func TestAwaitExpr(t *testing.T) {
	// Test AwaitExpr creation
	awaitExpr := &AwaitExpr{
		Expr: &CallExpr{
			Callee: &IdentifierExpr{Name: "asyncFunc"},
			Args:   []Expr{},
		},
	}

	if awaitExpr.Expr == nil {
		t.Error("Expected non-nil expression")
	}

	var expr Expr = awaitExpr
	if expr == nil {
		t.Error("Expected AwaitExpr to implement Expr interface")
	}
}

func TestStringInterpolationExpr(t *testing.T) {
	// Test StringInterpolationExpr creation
	interpExpr := &StringInterpolationExpr{
		Parts: []StringInterpolationPart{
			{IsLiteral: true, Literal: "Hello, "},
			{IsLiteral: false, Expr: &IdentifierExpr{Name: "name"}},
			{IsLiteral: true, Literal: "!"},
		},
	}

	if len(interpExpr.Parts) != 3 {
		t.Errorf("Expected 3 parts, got %d", len(interpExpr.Parts))
	}

	if !interpExpr.Parts[0].IsLiteral {
		t.Error("Expected first part to be literal")
	}

	if interpExpr.Parts[1].IsLiteral {
		t.Error("Expected second part to be expression")
	}

	var expr Expr = interpExpr
	if expr == nil {
		t.Error("Expected StringInterpolationExpr to implement Expr interface")
	}
}

func TestShortVarDeclStmt(t *testing.T) {
	// Test ShortVarDeclStmt creation
	shortVar := &ShortVarDeclStmt{
		Name:  "x",
		Type:  &TypeExpr{Name: "int"},
		Value: &LiteralExpr{Kind: LiteralInt, Value: "42"},
	}

	if shortVar.Name != "x" {
		t.Errorf("Expected name 'x', got '%s'", shortVar.Name)
	}

	if shortVar.Type == nil {
		t.Error("Expected non-nil type")
	}

	if shortVar.Value == nil {
		t.Error("Expected non-nil value")
	}

	var stmt Stmt = shortVar
	if stmt == nil {
		t.Error("Expected ShortVarDeclStmt to implement Stmt interface")
	}
}

func TestBindingStmt(t *testing.T) {
	// Test BindingStmt creation (let)
	bindingStmt := &BindingStmt{
		Mutable: false,
		Name:    "x",
		Type:    &TypeExpr{Name: "int"},
		Value:   &LiteralExpr{Kind: LiteralInt, Value: "42"},
	}

	if bindingStmt.Mutable {
		t.Error("Expected Mutable to be false for let")
	}

	if bindingStmt.Name != "x" {
		t.Errorf("Expected name 'x', got '%s'", bindingStmt.Name)
	}

	// Test var binding
	bindingStmt.Mutable = true
	if !bindingStmt.Mutable {
		t.Error("Expected Mutable to be true for var")
	}

	var stmt Stmt = bindingStmt
	if stmt == nil {
		t.Error("Expected BindingStmt to implement Stmt interface")
	}
}

func TestAssignmentStmt(t *testing.T) {
	// Test AssignmentStmt creation
	assignStmt := &AssignmentStmt{
		Left:  &IdentifierExpr{Name: "x"},
		Right: &LiteralExpr{Kind: LiteralInt, Value: "42"},
	}

	if assignStmt.Left == nil {
		t.Error("Expected non-nil left expression")
	}

	if assignStmt.Right == nil {
		t.Error("Expected non-nil right expression")
	}

	var stmt Stmt = assignStmt
	if stmt == nil {
		t.Error("Expected AssignmentStmt to implement Stmt interface")
	}
}

func TestIncrementStmt(t *testing.T) {
	// Test IncrementStmt creation
	incStmt := &IncrementStmt{
		Target: &IdentifierExpr{Name: "x"},
		Op:     "++",
	}

	if incStmt.Target == nil {
		t.Error("Expected non-nil target")
	}

	if incStmt.Op != "++" {
		t.Errorf("Expected op '++', got '%s'", incStmt.Op)
	}

	var stmt Stmt = incStmt
	if stmt == nil {
		t.Error("Expected IncrementStmt to implement Stmt interface")
	}
}

func TestFuncDeclWithTypeParams(t *testing.T) {
	// Test FuncDecl with generic type parameters
	funcDecl := &FuncDecl{
		Name: "identity",
		TypeParams: []TypeParam{
			{Name: "T"},
		},
		Params: []Param{
			{Name: "x", Type: &TypeExpr{Name: "T"}},
		},
		Return: &TypeExpr{Name: "T"},
		Body:   &BlockStmt{},
	}

	if len(funcDecl.TypeParams) != 1 {
		t.Errorf("Expected 1 type parameter, got %d", len(funcDecl.TypeParams))
	}

	if funcDecl.TypeParams[0].Name != "T" {
		t.Errorf("Expected type parameter 'T', got '%s'", funcDecl.TypeParams[0].Name)
	}
}

func TestFuncDeclAsync(t *testing.T) {
	// Test async FuncDecl
	funcDecl := &FuncDecl{
		Name:    "asyncFunc",
		IsAsync: true,
		Return:  &TypeExpr{Name: "Promise", Args: []*TypeExpr{{Name: "int"}}},
		Body:    &BlockStmt{},
	}

	if !funcDecl.IsAsync {
		t.Error("Expected IsAsync to be true")
	}
}

func TestFuncDeclExprBody(t *testing.T) {
	// Test FuncDecl with expression body
	funcDecl := &FuncDecl{
		Name:     "add",
		Params:   []Param{{Name: "a", Type: &TypeExpr{Name: "int"}}, {Name: "b", Type: &TypeExpr{Name: "int"}}},
		Return:   &TypeExpr{Name: "int"},
		ExprBody: &BinaryExpr{Left: &IdentifierExpr{Name: "a"}, Op: "+", Right: &IdentifierExpr{Name: "b"}},
	}

	if funcDecl.ExprBody == nil {
		t.Error("Expected non-nil ExprBody")
	}

	if funcDecl.Body != nil {
		t.Error("Expected nil Body when ExprBody is set")
	}
}

func TestStructDeclWithTypeParams(t *testing.T) {
	// Test StructDecl with generic type parameters
	structDecl := &StructDecl{
		Name: "Box",
		TypeParams: []TypeParam{
			{Name: "T"},
		},
		Fields: []StructField{
			{Name: "value", Type: &TypeExpr{Name: "T"}},
		},
	}

	if len(structDecl.TypeParams) != 1 {
		t.Errorf("Expected 1 type parameter, got %d", len(structDecl.TypeParams))
	}

	if structDecl.TypeParams[0].Name != "T" {
		t.Errorf("Expected type parameter 'T', got '%s'", structDecl.TypeParams[0].Name)
	}
}

func TestEnumVariant(t *testing.T) {
	// Test EnumVariant creation
	variant := EnumVariant{
		Name: "Red",
		Span: lexer.Span{Start: lexer.Position{Line: 1, Column: 1}, End: lexer.Position{Line: 1, Column: 4}},
	}

	if variant.Name != "Red" {
		t.Errorf("Expected name 'Red', got '%s'", variant.Name)
	}
}

func TestMapEntry(t *testing.T) {
	// Test MapEntry creation
	entry := MapEntry{
		Key:   &LiteralExpr{Kind: LiteralString, Value: "\"key\""},
		Value: &LiteralExpr{Kind: LiteralInt, Value: "42"},
		Span:  lexer.Span{Start: lexer.Position{Line: 1, Column: 1}, End: lexer.Position{Line: 1, Column: 10}},
	}

	if entry.Key == nil {
		t.Error("Expected non-nil key")
	}

	if entry.Value == nil {
		t.Error("Expected non-nil value")
	}
}

func TestStructLiteralField(t *testing.T) {
	// Test StructLiteralField creation
	field := StructLiteralField{
		Name: "x",
		Expr: &LiteralExpr{Kind: LiteralInt, Value: "1"},
		Span: lexer.Span{Start: lexer.Position{Line: 1, Column: 1}, End: lexer.Position{Line: 1, Column: 5}},
	}

	if field.Name != "x" {
		t.Errorf("Expected name 'x', got '%s'", field.Name)
	}

	if field.Expr == nil {
		t.Error("Expected non-nil expression")
	}
}

func TestStringInterpolationPart(t *testing.T) {
	// Test StringInterpolationPart creation
	part := StringInterpolationPart{
		IsLiteral: true,
		Literal:   "Hello",
		Span:      lexer.Span{Start: lexer.Position{Line: 1, Column: 1}, End: lexer.Position{Line: 1, Column: 6}},
	}

	if !part.IsLiteral {
		t.Error("Expected IsLiteral to be true")
	}

	if part.Literal != "Hello" {
		t.Errorf("Expected literal 'Hello', got '%s'", part.Literal)
	}

	// Test expression part
	part = StringInterpolationPart{
		IsLiteral: false,
		Expr:      &IdentifierExpr{Name: "name"},
		Span:      lexer.Span{Start: lexer.Position{Line: 1, Column: 1}, End: lexer.Position{Line: 1, Column: 5}},
	}

	if part.IsLiteral {
		t.Error("Expected IsLiteral to be false")
	}

	if part.Expr == nil {
		t.Error("Expected non-nil expression")
	}
}

func TestLiteralExprAllKinds(t *testing.T) {
	// Test all literal kinds
	kinds := []LiteralKind{
		LiteralInt,
		LiteralFloat,
		LiteralString,
		LiteralChar,
		LiteralBool,
		LiteralNull,
		LiteralHex,
		LiteralBinary,
	}

	for _, kind := range kinds {
		lit := &LiteralExpr{
			Kind:  kind,
			Value: "test",
		}

		if lit.Kind != kind {
			t.Errorf("Expected kind %v, got %v", kind, lit.Kind)
		}
	}
}

// Integration tests for AST module

func TestASTPrinterModule(t *testing.T) {
	// Test AST printer with a complete module
	module := &Module{
		Imports: []*ImportDecl{
			{Path: []string{"std", "io"}, Alias: "io"},
		},
		Decls: []Decl{
			&FuncDecl{
				Name: "main",
				Return: &TypeExpr{Name: "int"},
				Body: &BlockStmt{
					Statements: []Stmt{
						&ReturnStmt{
							Value: &LiteralExpr{Kind: LiteralInt, Value: "42"},
						},
					},
				},
			},
		},
	}

	output := Print(module)
	if output == "" {
		t.Error("Expected non-empty printer output")
	}

	if !strings.Contains(output, "Module") {
		t.Error("Expected output to contain 'Module'")
	}

	if !strings.Contains(output, "main") {
		t.Error("Expected output to contain 'main'")
	}
}

func TestASTPrinterComplexTypes(t *testing.T) {
	// Test AST printer with complex type expressions
	module := &Module{
		Decls: []Decl{
			&LetDecl{
				Name: "x",
				Type: &TypeExpr{
					Name: "List",
					Args: []*TypeExpr{
						{Name: "int"},
					},
				},
				Value: &ArrayLiteralExpr{
					Elements: []Expr{
						&LiteralExpr{Kind: LiteralInt, Value: "1"},
						&LiteralExpr{Kind: LiteralInt, Value: "2"},
					},
				},
			},
		},
	}

	output := Print(module)
	if output == "" {
		t.Error("Expected non-empty printer output")
	}
}

func TestASTPrinterUnionTypes(t *testing.T) {
	// Test AST printer with union types
	module := &Module{
		Decls: []Decl{
			&LetDecl{
				Name: "value",
				Type: &TypeExpr{
					IsUnion: true,
					Members: []*TypeExpr{
						{Name: "int"},
						{Name: "string"},
					},
				},
				Value: &LiteralExpr{Kind: LiteralInt, Value: "42"},
			},
		},
	}

	output := Print(module)
	if output == "" {
		t.Error("Expected non-empty printer output")
	}
}

func TestASTPrinterFunctionTypes(t *testing.T) {
	// Test AST printer with function types
	module := &Module{
		Decls: []Decl{
			&LetDecl{
				Name: "fn",
				Type: &TypeExpr{
					IsFunction: true,
					ParamTypes: []*TypeExpr{
						{Name: "int"},
						{Name: "string"},
					},
					ReturnType: &TypeExpr{Name: "bool"},
				},
				Value: &LambdaExpr{
					Params: []Param{
						{Name: "a", Type: &TypeExpr{Name: "int"}},
						{Name: "b", Type: &TypeExpr{Name: "string"}},
					},
					Body: &LiteralExpr{Kind: LiteralBool, Value: "true"},
				},
			},
		},
	}

	output := Print(module)
	if output == "" {
		t.Error("Expected non-empty printer output")
	}
}

func TestASTPrinterStructLiteral(t *testing.T) {
	// Test AST printer with struct literal
	module := &Module{
		Decls: []Decl{
			&StructDecl{
				Name: "Point",
				Fields: []StructField{
					{Name: "x", Type: &TypeExpr{Name: "int"}},
					{Name: "y", Type: &TypeExpr{Name: "int"}},
				},
			},
			&LetDecl{
				Name: "p",
				Type: &TypeExpr{Name: "Point"},
				Value: &StructLiteralExpr{
					TypeName: "Point",
					Fields: []StructLiteralField{
						{Name: "x", Expr: &LiteralExpr{Kind: LiteralInt, Value: "1"}},
						{Name: "y", Expr: &LiteralExpr{Kind: LiteralInt, Value: "2"}},
					},
				},
			},
		},
	}

	output := Print(module)
	if output == "" {
		t.Error("Expected non-empty printer output")
	}
}

func TestASTPrinterMapLiteral(t *testing.T) {
	// Test AST printer with map literal
	module := &Module{
		Decls: []Decl{
			&LetDecl{
				Name: "m",
				Type: &TypeExpr{Name: "map", Args: []*TypeExpr{{Name: "string"}, {Name: "int"}}},
				Value: &MapLiteralExpr{
					Entries: []MapEntry{
						{
							Key:   &LiteralExpr{Kind: LiteralString, Value: "\"a\""},
							Value: &LiteralExpr{Kind: LiteralInt, Value: "1"},
						},
						{
							Key:   &LiteralExpr{Kind: LiteralString, Value: "\"b\""},
							Value: &LiteralExpr{Kind: LiteralInt, Value: "2"},
						},
					},
				},
			},
		},
	}

	output := Print(module)
	if output == "" {
		t.Error("Expected non-empty printer output")
	}
}

func TestASTPrinterControlFlow(t *testing.T) {
	// Test AST printer with control flow statements
	module := &Module{
		Decls: []Decl{
			&FuncDecl{
				Name: "test",
				Return: &TypeExpr{Name: "int"},
				Body: &BlockStmt{
					Statements: []Stmt{
						&IfStmt{
							Cond: &LiteralExpr{Kind: LiteralBool, Value: "true"},
							Then: &BlockStmt{
								Statements: []Stmt{
									&ReturnStmt{Value: &LiteralExpr{Kind: LiteralInt, Value: "1"}},
								},
							},
							Else: &BlockStmt{
								Statements: []Stmt{
									&ReturnStmt{Value: &LiteralExpr{Kind: LiteralInt, Value: "0"}},
								},
							},
						},
						&ForStmt{
							IsRange: true,
							Target:  &IdentifierExpr{Name: "item"},
							Iterable: &IdentifierExpr{Name: "items"},
							Body: &BlockStmt{
								Statements: []Stmt{
									&ExprStmt{Expr: &CallExpr{
										Callee: &IdentifierExpr{Name: "println"},
										Args:   []Expr{&IdentifierExpr{Name: "item"}},
									}},
								},
							},
						},
						&WhileStmt{
							Cond: &LiteralExpr{Kind: LiteralBool, Value: "true"},
							Body: &BlockStmt{
								Statements: []Stmt{
									&BreakStmt{},
								},
							},
						},
					},
				},
			},
		},
	}

	output := Print(module)
	if output == "" {
		t.Error("Expected non-empty printer output")
	}
}

func TestASTPrinterTryCatch(t *testing.T) {
	// Test AST printer with try-catch
	module := &Module{
		Decls: []Decl{
			&FuncDecl{
				Name: "test",
				Return: &TypeExpr{Name: "int"},
				Body: &BlockStmt{
					Statements: []Stmt{
						&TryStmt{
							TryBlock: &BlockStmt{
								Statements: []Stmt{
									&ThrowStmt{Expr: &LiteralExpr{Kind: LiteralString, Value: "\"error\""}},
								},
							},
							CatchClauses: []*CatchClause{
								{
									ExceptionVar:  "e",
									ExceptionType: "Error",
									Block: &BlockStmt{
										Statements: []Stmt{
											&ReturnStmt{Value: &LiteralExpr{Kind: LiteralInt, Value: "0"}},
										},
									},
								},
							},
							FinallyBlock: &BlockStmt{
								Statements: []Stmt{},
							},
						},
					},
				},
			},
		},
	}

	output := Print(module)
	if output == "" {
		t.Error("Expected non-empty printer output")
	}
}

func TestASTPrinterStringInterpolation(t *testing.T) {
	// Test AST printer with string interpolation
	module := &Module{
		Decls: []Decl{
			&LetDecl{
				Name: "msg",
				Type: &TypeExpr{Name: "string"},
				Value: &StringInterpolationExpr{
					Parts: []StringInterpolationPart{
						{IsLiteral: true, Literal: "Hello, "},
						{IsLiteral: false, Expr: &IdentifierExpr{Name: "name"}},
						{IsLiteral: true, Literal: "!"},
					},
				},
			},
		},
	}

	output := Print(module)
	if output == "" {
		t.Error("Expected non-empty printer output")
	}
}

func TestASTNodeRelationships(t *testing.T) {
	// Test AST node relationships and spans
	span := lexer.Span{
		Start: lexer.Position{Line: 1, Column: 1},
		End:   lexer.Position{Line: 1, Column: 10},
	}

	// Test that all nodes properly implement interfaces
	var node Node = &Module{SpanInfo: span}
	if node.Span() != span {
		t.Error("Module Span() failed")
	}

	var decl Decl = &LetDecl{SpanInfo: span}
	if decl.Span() != span {
		t.Error("LetDecl Span() failed")
	}

	var stmt Stmt = &BlockStmt{SpanInfo: span}
	if stmt.Span() != span {
		t.Error("BlockStmt Span() failed")
	}

	var expr Expr = &LiteralExpr{SpanInfo: span}
	if expr.Span() != span {
		t.Error("LiteralExpr Span() failed")
	}
}

func TestASTComplexNestedStructures(t *testing.T) {
	// Test AST with deeply nested structures
	module := &Module{
		Decls: []Decl{
			&FuncDecl{
				Name: "complex",
				TypeParams: []TypeParam{{Name: "T"}},
				Params: []Param{
					{Name: "x", Type: &TypeExpr{Name: "T"}},
				},
				Return: &TypeExpr{
					IsOptional: true,
					OptionalType: &TypeExpr{
						Name: "array",
						Args: []*TypeExpr{
							{Name: "T"},
						},
					},
				},
				Body: &BlockStmt{
					Statements: []Stmt{
						&IfStmt{
							Cond: &BinaryExpr{
								Left:  &IdentifierExpr{Name: "x"},
								Op:    "==",
								Right: &LiteralExpr{Kind: LiteralNull, Value: "null"},
							},
							Then: &BlockStmt{
								Statements: []Stmt{
									&ReturnStmt{Value: &LiteralExpr{Kind: LiteralNull, Value: "null"}},
								},
							},
							Else: &BlockStmt{
								Statements: []Stmt{
									&ReturnStmt{
										Value: &ArrayLiteralExpr{
											Elements: []Expr{
												&IdentifierExpr{Name: "x"},
											},
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

	output := Print(module)
	if output == "" {
		t.Error("Expected non-empty printer output")
	}

	// Verify the structure
	if len(module.Decls) != 1 {
		t.Errorf("Expected 1 declaration, got %d", len(module.Decls))
	}

	funcDecl, ok := module.Decls[0].(*FuncDecl)
	if !ok {
		t.Fatal("Expected FuncDecl")
	}

	if len(funcDecl.TypeParams) != 1 {
		t.Errorf("Expected 1 type parameter, got %d", len(funcDecl.TypeParams))
	}

	if funcDecl.Return == nil || !funcDecl.Return.IsOptional {
		t.Error("Expected optional return type")
	}
}
