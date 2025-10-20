package ast

import "github.com/omni-lang/omni/internal/lexer"

// Node represents any syntax tree node with an associated span.
type Node interface {
	Span() lexer.Span
	node()
}

// Module is the root of a compilation unit.
type Module struct {
	SpanInfo lexer.Span
	Imports  []*ImportDecl
	Decls    []Decl
}

// Span implements Node.
func (m *Module) Span() lexer.Span { return m.SpanInfo }
func (m *Module) node()            {}

// ImportDecl represents an import statement.
type ImportDecl struct {
	SpanInfo lexer.Span
	Path     []string
	// Alias, when non-empty, provides the local name used to reference the
	// imported module path (e.g., import std.io as io).
	Alias string
}

func (i *ImportDecl) Span() lexer.Span { return i.SpanInfo }
func (i *ImportDecl) node()            {}
func (i *ImportDecl) decl()            {}

// Decl captures top-level declarations.
type Decl interface {
	Node
	decl()
}

// TypeExpr describes a type reference possibly containing generics.
type TypeExpr struct {
	SpanInfo lexer.Span
	Name     string
	Args     []*TypeExpr
}

func (t *TypeExpr) Span() lexer.Span { return t.SpanInfo }
func (t *TypeExpr) node()            {}

// LetDecl models an immutable binding.
type LetDecl struct {
	SpanInfo lexer.Span
	Name     string
	Type     *TypeExpr
	Value    Expr
}

func (d *LetDecl) Span() lexer.Span { return d.SpanInfo }
func (d *LetDecl) node()            {}
func (d *LetDecl) decl()            {}

// VarDecl models a mutable binding.
type VarDecl struct {
	SpanInfo lexer.Span
	Name     string
	Type     *TypeExpr
	Value    Expr
}

func (d *VarDecl) Span() lexer.Span { return d.SpanInfo }
func (d *VarDecl) node()            {}
func (d *VarDecl) decl()            {}

// StructDecl defines a struct type.
type StructDecl struct {
	SpanInfo lexer.Span
	Name     string
	Fields   []StructField
}

func (d *StructDecl) Span() lexer.Span { return d.SpanInfo }
func (d *StructDecl) node()            {}
func (d *StructDecl) decl()            {}

// StructField models a field inside a struct.
type StructField struct {
	Name string
	Type *TypeExpr
	Span lexer.Span
}

// EnumDecl defines an enum with variants.
type EnumDecl struct {
	SpanInfo lexer.Span
	Name     string
	Variants []EnumVariant
}

func (d *EnumDecl) Span() lexer.Span { return d.SpanInfo }
func (d *EnumDecl) node()            {}
func (d *EnumDecl) decl()            {}

// EnumVariant describes a single enum case.
type EnumVariant struct {
	Name string
	Span lexer.Span
}

// FuncDecl describes a function definition.
type FuncDecl struct {
	SpanInfo lexer.Span
	Name     string
	Params   []Param
	Return   *TypeExpr
	Body     *BlockStmt
	ExprBody Expr // for fat arrow shorthand
}

func (d *FuncDecl) Span() lexer.Span { return d.SpanInfo }
func (d *FuncDecl) node()            {}
func (d *FuncDecl) decl()            {}

// Param represents a function parameter.
type Param struct {
	Name string
	Type *TypeExpr
	Span lexer.Span
}

// Stmt is a statement.
type Stmt interface {
	Node
	stmt()
}

// BlockStmt groups zero or more statements.
type BlockStmt struct {
	SpanInfo   lexer.Span
	Statements []Stmt
}

func (s *BlockStmt) Span() lexer.Span { return s.SpanInfo }
func (s *BlockStmt) node()            {}
func (s *BlockStmt) stmt()            {}

// ReturnStmt returns from a function.
type ReturnStmt struct {
	SpanInfo lexer.Span
	Value    Expr
}

func (s *ReturnStmt) Span() lexer.Span { return s.SpanInfo }
func (s *ReturnStmt) node()            {}
func (s *ReturnStmt) stmt()            {}

// ExprStmt wraps an expression as a statement.
type ExprStmt struct {
	SpanInfo lexer.Span
	Expr     Expr
}

func (s *ExprStmt) Span() lexer.Span { return s.SpanInfo }
func (s *ExprStmt) node()            {}
func (s *ExprStmt) stmt()            {}

// ForStmt handles both range and classic for loops.
type ForStmt struct {
	SpanInfo  lexer.Span
	Init      Stmt // optional
	Condition Expr // optional
	Post      Stmt // optional
	Target    *IdentifierExpr
	Iterable  Expr
	Body      *BlockStmt
	IsRange   bool
}

func (s *ForStmt) Span() lexer.Span { return s.SpanInfo }
func (s *ForStmt) node()            {}
func (s *ForStmt) stmt()            {}

// IfStmt expresses conditional logic.
type IfStmt struct {
	SpanInfo lexer.Span
	Cond     Expr
	Then     *BlockStmt
	Else     Stmt // either *BlockStmt or *IfStmt
}

func (s *IfStmt) Span() lexer.Span { return s.SpanInfo }
func (s *IfStmt) node()            {}
func (s *IfStmt) stmt()            {}

// Expr categories -----------------------------------------------------------

// Expr represents an expression node.
type Expr interface {
	Node
	expr()
}

// IdentifierExpr references a name.
type IdentifierExpr struct {
	SpanInfo lexer.Span
	Name     string
}

func (e *IdentifierExpr) Span() lexer.Span { return e.SpanInfo }
func (e *IdentifierExpr) node()            {}
func (e *IdentifierExpr) expr()            {}

// LiteralKind identifies literal classifications.
type LiteralKind string

const (
	LiteralInt    LiteralKind = "int"
	LiteralFloat  LiteralKind = "float"
	LiteralString LiteralKind = "string"
	LiteralChar   LiteralKind = "char"
	LiteralBool   LiteralKind = "bool"
)

// LiteralExpr stores literal values as raw lexemes.
type LiteralExpr struct {
	SpanInfo lexer.Span
	Kind     LiteralKind
	Value    string
}

func (e *LiteralExpr) Span() lexer.Span { return e.SpanInfo }
func (e *LiteralExpr) node()            {}
func (e *LiteralExpr) expr()            {}

// UnaryExpr applies an operator to a sub-expression.
type UnaryExpr struct {
	SpanInfo lexer.Span
	Op       string
	Expr     Expr
}

func (e *UnaryExpr) Span() lexer.Span { return e.SpanInfo }
func (e *UnaryExpr) node()            {}
func (e *UnaryExpr) expr()            {}

// BinaryExpr applies a binary operator.
type BinaryExpr struct {
	SpanInfo lexer.Span
	Left     Expr
	Op       string
	Right    Expr
}

func (e *BinaryExpr) Span() lexer.Span { return e.SpanInfo }
func (e *BinaryExpr) node()            {}
func (e *BinaryExpr) expr()            {}

// CallExpr invokes a call.
type CallExpr struct {
	SpanInfo lexer.Span
	Callee   Expr
	Args     []Expr
}

func (e *CallExpr) Span() lexer.Span { return e.SpanInfo }
func (e *CallExpr) node()            {}
func (e *CallExpr) expr()            {}

// IndexExpr models array/map indexing.
type IndexExpr struct {
	SpanInfo lexer.Span
	Target   Expr
	Index    Expr
}

func (e *IndexExpr) Span() lexer.Span { return e.SpanInfo }
func (e *IndexExpr) node()            {}
func (e *IndexExpr) expr()            {}

// MemberExpr models field access.
type MemberExpr struct {
	SpanInfo lexer.Span
	Target   Expr
	Member   string
}

func (e *MemberExpr) Span() lexer.Span { return e.SpanInfo }
func (e *MemberExpr) node()            {}
func (e *MemberExpr) expr()            {}

// ArrayLiteralExpr holds array literal entries.
type ArrayLiteralExpr struct {
	SpanInfo lexer.Span
	Elements []Expr
}

func (e *ArrayLiteralExpr) Span() lexer.Span { return e.SpanInfo }
func (e *ArrayLiteralExpr) node()            {}
func (e *ArrayLiteralExpr) expr()            {}

// MapEntry pairs key/value expressions.
type MapEntry struct {
	Key   Expr
	Value Expr
	Span  lexer.Span
}

// MapLiteralExpr holds map literal entries.
type MapLiteralExpr struct {
	SpanInfo lexer.Span
	Entries  []MapEntry
}

func (e *MapLiteralExpr) Span() lexer.Span { return e.SpanInfo }
func (e *MapLiteralExpr) node()            {}
func (e *MapLiteralExpr) expr()            {}

// StructLiteralField holds an assignment inside a struct literal.
type StructLiteralField struct {
	Name string
	Expr Expr
	Span lexer.Span
}

// StructLiteralExpr constructs struct values.
type StructLiteralExpr struct {
	SpanInfo lexer.Span
	TypeName string
	Fields   []StructLiteralField
}

func (e *StructLiteralExpr) Span() lexer.Span { return e.SpanInfo }
func (e *StructLiteralExpr) node()            {}
func (e *StructLiteralExpr) expr()            {}

// AssignmentExpr models simple assignments.
type AssignmentExpr struct {
	SpanInfo lexer.Span
	Left     Expr
	Right    Expr
}

func (e *AssignmentExpr) Span() lexer.Span { return e.SpanInfo }
func (e *AssignmentExpr) node()            {}
func (e *AssignmentExpr) expr()            {}

// IncrementExpr models postfix ++/-- used in for loops.
type IncrementExpr struct {
	SpanInfo lexer.Span
	Target   Expr
	Op       string
}

func (e *IncrementExpr) Span() lexer.Span { return e.SpanInfo }
func (e *IncrementExpr) node()            {}
func (e *IncrementExpr) expr()            {}
