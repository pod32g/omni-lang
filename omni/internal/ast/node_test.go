package ast

import (
	"testing"

	"github.com/omni-lang/omni/internal/lexer"
)

func TestNodeMethods(t *testing.T) {
	// Test Span methods for all node types
	span := lexer.Span{Start: lexer.Position{Line: 1, Column: 1}, End: lexer.Position{Line: 1, Column: 10}}

	t.Run("ImportDecl", func(t *testing.T) {
		decl := &ImportDecl{
			SpanInfo: span,
			Path:     []string{"std", "io"},
			Alias:    "io",
		}

		if decl.Span() != span {
			t.Errorf("Expected span %v, got %v", span, decl.Span())
		}

		// Test interface methods
		var node Node = decl
		if node.Span() != span {
			t.Errorf("Expected span %v, got %v", span, node.Span())
		}

		var declNode Decl = decl
		if declNode.Span() != span {
			t.Errorf("Expected span %v, got %v", span, declNode.Span())
		}
	})

	t.Run("LetDecl", func(t *testing.T) {
		decl := &LetDecl{
			SpanInfo: span,
			Name:     "x",
			Type:     &TypeExpr{Name: "int"},
			Value:    &LiteralExpr{Kind: LiteralInt, Value: "42"},
		}

		if decl.Span() != span {
			t.Errorf("Expected span %v, got %v", span, decl.Span())
		}

		// Test interface methods
		var node Node = decl
		if node.Span() != span {
			t.Errorf("Expected span %v, got %v", span, node.Span())
		}

		var declNode Decl = decl
		if declNode.Span() != span {
			t.Errorf("Expected span %v, got %v", span, declNode.Span())
		}
	})

	t.Run("VarDecl", func(t *testing.T) {
		decl := &VarDecl{
			SpanInfo: span,
			Name:     "y",
			Type:     &TypeExpr{Name: "string"},
			Value:    &LiteralExpr{Kind: LiteralString, Value: "\"hello\""},
		}

		if decl.Span() != span {
			t.Errorf("Expected span %v, got %v", span, decl.Span())
		}

		// Test interface methods
		var node Node = decl
		if node.Span() != span {
			t.Errorf("Expected span %v, got %v", span, node.Span())
		}

		var declNode Decl = decl
		if declNode.Span() != span {
			t.Errorf("Expected span %v, got %v", span, declNode.Span())
		}
	})

	t.Run("StructDecl", func(t *testing.T) {
		decl := &StructDecl{
			SpanInfo: span,
			Name:     "Point",
			Fields:   []StructField{{Name: "x", Type: &TypeExpr{Name: "int"}}},
		}

		if decl.Span() != span {
			t.Errorf("Expected span %v, got %v", span, decl.Span())
		}

		// Test interface methods
		var node Node = decl
		if node.Span() != span {
			t.Errorf("Expected span %v, got %v", span, node.Span())
		}

		var declNode Decl = decl
		if declNode.Span() != span {
			t.Errorf("Expected span %v, got %v", span, declNode.Span())
		}
	})

	t.Run("EnumDecl", func(t *testing.T) {
		decl := &EnumDecl{
			SpanInfo: span,
			Name:     "Color",
			Variants: []EnumVariant{{Name: "Red"}, {Name: "Green"}, {Name: "Blue"}},
		}

		if decl.Span() != span {
			t.Errorf("Expected span %v, got %v", span, decl.Span())
		}

		// Test interface methods
		var node Node = decl
		if node.Span() != span {
			t.Errorf("Expected span %v, got %v", span, node.Span())
		}

		var declNode Decl = decl
		if declNode.Span() != span {
			t.Errorf("Expected span %v, got %v", span, declNode.Span())
		}
	})

	t.Run("FuncDecl", func(t *testing.T) {
		decl := &FuncDecl{
			SpanInfo: span,
			Name:     "main",
			Params:   []Param{{Name: "x", Type: &TypeExpr{Name: "int"}}},
			Return:   &TypeExpr{Name: "int"},
			Body:     &BlockStmt{},
		}

		if decl.Span() != span {
			t.Errorf("Expected span %v, got %v", span, decl.Span())
		}

		// Test interface methods
		var node Node = decl
		if node.Span() != span {
			t.Errorf("Expected span %v, got %v", span, node.Span())
		}

		var declNode Decl = decl
		if declNode.Span() != span {
			t.Errorf("Expected span %v, got %v", span, declNode.Span())
		}
	})

	t.Run("TypeAliasDecl", func(t *testing.T) {
		decl := &TypeAliasDecl{
			SpanInfo: span,
			Name:     "UserID",
			Type:     &TypeExpr{Name: "int"},
		}

		if decl.Span() != span {
			t.Errorf("Expected span %v, got %v", span, decl.Span())
		}

		// Test interface methods
		var node Node = decl
		if node.Span() != span {
			t.Errorf("Expected span %v, got %v", span, node.Span())
		}

		var declNode Decl = decl
		if declNode.Span() != span {
			t.Errorf("Expected span %v, got %v", span, declNode.Span())
		}
	})

	t.Run("BlockStmt", func(t *testing.T) {
		stmt := &BlockStmt{
			SpanInfo:   span,
			Statements: []Stmt{},
		}

		if stmt.Span() != span {
			t.Errorf("Expected span %v, got %v", span, stmt.Span())
		}

		// Test interface methods
		var node Node = stmt
		if node.Span() != span {
			t.Errorf("Expected span %v, got %v", span, node.Span())
		}

		var stmtNode Stmt = stmt
		if stmtNode.Span() != span {
			t.Errorf("Expected span %v, got %v", span, stmtNode.Span())
		}
	})

	t.Run("ExprStmt", func(t *testing.T) {
		stmt := &ExprStmt{
			SpanInfo: span,
			Expr:     &LiteralExpr{Kind: LiteralInt, Value: "42"},
		}

		if stmt.Span() != span {
			t.Errorf("Expected span %v, got %v", span, stmt.Span())
		}

		// Test interface methods
		var node Node = stmt
		if node.Span() != span {
			t.Errorf("Expected span %v, got %v", span, node.Span())
		}

		var stmtNode Stmt = stmt
		if stmtNode.Span() != span {
			t.Errorf("Expected span %v, got %v", span, stmtNode.Span())
		}
	})

	t.Run("IfStmt", func(t *testing.T) {
		stmt := &IfStmt{
			SpanInfo: span,
			Cond:     &LiteralExpr{Kind: LiteralBool, Value: "true"},
			Then:     &BlockStmt{},
			Else:     nil,
		}

		if stmt.Span() != span {
			t.Errorf("Expected span %v, got %v", span, stmt.Span())
		}

		// Test interface methods
		var node Node = stmt
		if node.Span() != span {
			t.Errorf("Expected span %v, got %v", span, node.Span())
		}

		var stmtNode Stmt = stmt
		if stmtNode.Span() != span {
			t.Errorf("Expected span %v, got %v", span, stmtNode.Span())
		}
	})

	t.Run("ForStmt", func(t *testing.T) {
		stmt := &ForStmt{
			SpanInfo:  span,
			Init:      &ExprStmt{Expr: &LiteralExpr{Kind: LiteralInt, Value: "0"}},
			Condition: &LiteralExpr{Kind: LiteralBool, Value: "true"},
			Post:      &ExprStmt{Expr: &LiteralExpr{Kind: LiteralInt, Value: "1"}},
			Body:      &BlockStmt{},
		}

		if stmt.Span() != span {
			t.Errorf("Expected span %v, got %v", span, stmt.Span())
		}

		// Test interface methods
		var node Node = stmt
		if node.Span() != span {
			t.Errorf("Expected span %v, got %v", span, node.Span())
		}

		var stmtNode Stmt = stmt
		if stmtNode.Span() != span {
			t.Errorf("Expected span %v, got %v", span, stmtNode.Span())
		}
	})

	t.Run("ReturnStmt", func(t *testing.T) {
		stmt := &ReturnStmt{
			SpanInfo: span,
			Value:    &LiteralExpr{Kind: LiteralInt, Value: "42"},
		}

		if stmt.Span() != span {
			t.Errorf("Expected span %v, got %v", span, stmt.Span())
		}

		// Test interface methods
		var node Node = stmt
		if node.Span() != span {
			t.Errorf("Expected span %v, got %v", span, node.Span())
		}

		var stmtNode Stmt = stmt
		if stmtNode.Span() != span {
			t.Errorf("Expected span %v, got %v", span, stmtNode.Span())
		}
	})

	t.Run("LiteralExpr", func(t *testing.T) {
		expr := &LiteralExpr{
			SpanInfo: span,
			Kind:     LiteralInt,
			Value:    "42",
		}

		if expr.Span() != span {
			t.Errorf("Expected span %v, got %v", span, expr.Span())
		}

		// Test interface methods
		var node Node = expr
		if node.Span() != span {
			t.Errorf("Expected span %v, got %v", span, node.Span())
		}

		var exprNode Expr = expr
		if exprNode.Span() != span {
			t.Errorf("Expected span %v, got %v", span, exprNode.Span())
		}
	})

	t.Run("BinaryExpr", func(t *testing.T) {
		expr := &BinaryExpr{
			SpanInfo: span,
			Left:     &LiteralExpr{Kind: LiteralInt, Value: "1"},
			Op:       "+",
			Right:    &LiteralExpr{Kind: LiteralInt, Value: "2"},
		}

		if expr.Span() != span {
			t.Errorf("Expected span %v, got %v", span, expr.Span())
		}

		// Test interface methods
		var node Node = expr
		if node.Span() != span {
			t.Errorf("Expected span %v, got %v", span, node.Span())
		}

		var exprNode Expr = expr
		if exprNode.Span() != span {
			t.Errorf("Expected span %v, got %v", span, exprNode.Span())
		}
	})

	t.Run("UnaryExpr", func(t *testing.T) {
		expr := &UnaryExpr{
			SpanInfo: span,
			Op:       "-",
			Expr:     &LiteralExpr{Kind: LiteralInt, Value: "42"},
		}

		if expr.Span() != span {
			t.Errorf("Expected span %v, got %v", span, expr.Span())
		}

		// Test interface methods
		var node Node = expr
		if node.Span() != span {
			t.Errorf("Expected span %v, got %v", span, node.Span())
		}

		var exprNode Expr = expr
		if exprNode.Span() != span {
			t.Errorf("Expected span %v, got %v", span, exprNode.Span())
		}
	})

	t.Run("CallExpr", func(t *testing.T) {
		expr := &CallExpr{
			SpanInfo: span,
			Callee:   &IdentifierExpr{Name: "println"},
			Args:     []Expr{&LiteralExpr{Kind: LiteralString, Value: "\"hello\""}},
		}

		if expr.Span() != span {
			t.Errorf("Expected span %v, got %v", span, expr.Span())
		}

		// Test interface methods
		var node Node = expr
		if node.Span() != span {
			t.Errorf("Expected span %v, got %v", span, node.Span())
		}

		var exprNode Expr = expr
		if exprNode.Span() != span {
			t.Errorf("Expected span %v, got %v", span, exprNode.Span())
		}
	})

	t.Run("IdentifierExpr", func(t *testing.T) {
		expr := &IdentifierExpr{
			SpanInfo: span,
			Name:     "x",
		}

		if expr.Span() != span {
			t.Errorf("Expected span %v, got %v", span, expr.Span())
		}

		// Test interface methods
		var node Node = expr
		if node.Span() != span {
			t.Errorf("Expected span %v, got %v", span, node.Span())
		}

		var exprNode Expr = expr
		if exprNode.Span() != span {
			t.Errorf("Expected span %v, got %v", span, exprNode.Span())
		}
	})

	t.Run("MemberExpr", func(t *testing.T) {
		expr := &MemberExpr{
			SpanInfo: span,
			Target:   &IdentifierExpr{Name: "obj"},
			Member:   "field",
		}

		if expr.Span() != span {
			t.Errorf("Expected span %v, got %v", span, expr.Span())
		}

		// Test interface methods
		var node Node = expr
		if node.Span() != span {
			t.Errorf("Expected span %v, got %v", span, node.Span())
		}

		var exprNode Expr = expr
		if exprNode.Span() != span {
			t.Errorf("Expected span %v, got %v", span, exprNode.Span())
		}
	})

	t.Run("IndexExpr", func(t *testing.T) {
		expr := &IndexExpr{
			SpanInfo: span,
			Target:   &IdentifierExpr{Name: "arr"},
			Index:    &LiteralExpr{Kind: LiteralInt, Value: "0"},
		}

		if expr.Span() != span {
			t.Errorf("Expected span %v, got %v", span, expr.Span())
		}

		// Test interface methods
		var node Node = expr
		if node.Span() != span {
			t.Errorf("Expected span %v, got %v", span, node.Span())
		}

		var exprNode Expr = expr
		if exprNode.Span() != span {
			t.Errorf("Expected span %v, got %v", span, exprNode.Span())
		}
	})

	t.Run("AssignmentExpr", func(t *testing.T) {
		expr := &AssignmentExpr{
			SpanInfo: span,
			Left:     &IdentifierExpr{Name: "x"},
			Right:    &LiteralExpr{Kind: LiteralInt, Value: "42"},
		}

		if expr.Span() != span {
			t.Errorf("Expected span %v, got %v", span, expr.Span())
		}

		// Test interface methods
		var node Node = expr
		if node.Span() != span {
			t.Errorf("Expected span %v, got %v", span, node.Span())
		}

		var exprNode Expr = expr
		if exprNode.Span() != span {
			t.Errorf("Expected span %v, got %v", span, exprNode.Span())
		}
	})

	t.Run("TypeExpr", func(t *testing.T) {
		typeExpr := &TypeExpr{
			SpanInfo: span,
			Name:     "int",
		}

		if typeExpr.Span() != span {
			t.Errorf("Expected span %v, got %v", span, typeExpr.Span())
		}

		// Test interface methods
		var node Node = typeExpr
		if node.Span() != span {
			t.Errorf("Expected span %v, got %v", span, node.Span())
		}

		var typeExprNode TypeExpr = *typeExpr
		if typeExprNode.Span() != span {
			t.Errorf("Expected span %v, got %v", span, typeExprNode.Span())
		}
	})

	t.Run("GenericTypeExpr", func(t *testing.T) {
		typeExpr := &GenericTypeExpr{
			SpanInfo: span,
			Name:     "Array",
			TypeArgs: []TypeExpr{{Name: "int"}},
		}

		if typeExpr.Span() != span {
			t.Errorf("Expected span %v, got %v", span, typeExpr.Span())
		}

		// Test interface methods
		var node Node = typeExpr
		if node.Span() != span {
			t.Errorf("Expected span %v, got %v", span, node.Span())
		}
	})

	t.Run("UnionTypeExpr", func(t *testing.T) {
		typeExpr := &UnionTypeExpr{
			SpanInfo: span,
			Types:    []TypeExpr{{Name: "int"}, {Name: "string"}},
		}

		if typeExpr.Span() != span {
			t.Errorf("Expected span %v, got %v", span, typeExpr.Span())
		}

		// Test interface methods
		var node Node = typeExpr
		if node.Span() != span {
			t.Errorf("Expected span %v, got %v", span, node.Span())
		}
	})

	t.Run("OptionalTypeExpr", func(t *testing.T) {
		typeExpr := &OptionalTypeExpr{
			SpanInfo:  span,
			InnerType: TypeExpr{Name: "int"},
		}

		if typeExpr.Span() != span {
			t.Errorf("Expected span %v, got %v", span, typeExpr.Span())
		}

		// Test interface methods
		var node Node = typeExpr
		if node.Span() != span {
			t.Errorf("Expected span %v, got %v", span, node.Span())
		}
	})

	t.Run("StringInterpolationExpr", func(t *testing.T) {
		expr := &StringInterpolationExpr{
			SpanInfo: span,
			Parts:    []StringInterpolationPart{{IsLiteral: true, Literal: "Hello"}},
		}

		if expr.Span() != span {
			t.Errorf("Expected span %v, got %v", span, expr.Span())
		}

		// Test interface methods
		var node Node = expr
		if node.Span() != span {
			t.Errorf("Expected span %v, got %v", span, node.Span())
		}

		var exprNode Expr = expr
		if exprNode.Span() != span {
			t.Errorf("Expected span %v, got %v", span, exprNode.Span())
		}
	})

	t.Run("TryStmt", func(t *testing.T) {
		stmt := &TryStmt{
			SpanInfo:     span,
			TryBlock:     &BlockStmt{},
			CatchClauses: []*CatchClause{},
		}

		if stmt.Span() != span {
			t.Errorf("Expected span %v, got %v", span, stmt.Span())
		}

		// Test interface methods
		var node Node = stmt
		if node.Span() != span {
			t.Errorf("Expected span %v, got %v", span, node.Span())
		}

		var stmtNode Stmt = stmt
		if stmtNode.Span() != span {
			t.Errorf("Expected span %v, got %v", span, stmtNode.Span())
		}
	})

	t.Run("ThrowStmt", func(t *testing.T) {
		stmt := &ThrowStmt{
			SpanInfo: span,
			Expr:     &LiteralExpr{Kind: LiteralString, Value: "\"error\""},
		}

		if stmt.Span() != span {
			t.Errorf("Expected span %v, got %v", span, stmt.Span())
		}

		// Test interface methods
		var node Node = stmt
		if node.Span() != span {
			t.Errorf("Expected span %v, got %v", span, node.Span())
		}

		var stmtNode Stmt = stmt
		if stmtNode.Span() != span {
			t.Errorf("Expected span %v, got %v", span, stmtNode.Span())
		}
	})
}
