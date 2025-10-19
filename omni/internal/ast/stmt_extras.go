package ast

import "github.com/omni-lang/omni/internal/lexer"

// ShortVarDeclStmt captures `name:Type = expr` used in for init clauses.
type ShortVarDeclStmt struct {
	SpanInfo lexer.Span
	Name     string
	Type     *TypeExpr
	Value    Expr
}

func (s *ShortVarDeclStmt) Span() lexer.Span { return s.SpanInfo }
func (s *ShortVarDeclStmt) node()            {}
func (s *ShortVarDeclStmt) stmt()            {}

// BindingStmt models `let`/`var` declarations inside blocks.
type BindingStmt struct {
	SpanInfo lexer.Span
	Mutable  bool
	Name     string
	Type     *TypeExpr
	Value    Expr
}

func (s *BindingStmt) Span() lexer.Span { return s.SpanInfo }
func (s *BindingStmt) node()            {}
func (s *BindingStmt) stmt()            {}

// AssignmentStmt wraps an assignment expression as a statement.
type AssignmentStmt struct {
	SpanInfo lexer.Span
	Left     Expr
	Right    Expr
}

func (s *AssignmentStmt) Span() lexer.Span { return s.SpanInfo }
func (s *AssignmentStmt) node()            {}
func (s *AssignmentStmt) stmt()            {}

// IncrementStmt represents postfix increments/decrements as statements.
type IncrementStmt struct {
	SpanInfo lexer.Span
	Target   Expr
	Op       string
}

func (s *IncrementStmt) Span() lexer.Span { return s.SpanInfo }
func (s *IncrementStmt) node()            {}
func (s *IncrementStmt) stmt()            {}
