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

// TypeExpr describes a type reference possibly containing generics or unions.
type TypeExpr struct {
	SpanInfo lexer.Span
	Name     string
	Args     []*TypeExpr
	// For union types: int | string | bool
	IsUnion bool
	Members []*TypeExpr // The types in the union
	// For function types: (int, string) -> bool
	IsFunction bool
	ParamTypes []*TypeExpr // Parameter types for function types
	ReturnType *TypeExpr   // Return type for function types
	// For optional types: Option<T> or T?
	IsOptional   bool
	OptionalType *TypeExpr // The wrapped type for Option<T>
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
	SpanInfo   lexer.Span
	Name       string
	TypeParams []TypeParam // Generic type parameters
	Fields     []StructField
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
	SpanInfo   lexer.Span
	Name       string
	TypeParams []TypeParam // Generic type parameters
	Params     []Param
	Return     *TypeExpr
	Body       *BlockStmt
	ExprBody   Expr // for fat arrow shorthand
	IsAsync    bool // async function
}

// TypeParam represents a generic type parameter.
type TypeParam struct {
	Name string
	Span lexer.Span
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
// ForStmt represents a for loop statement.
// INVARIANT: Either IsRange is true (range form: for item in items { ... })
// or IsRange is false (classic form: for init; cond; post { ... }).
// When IsRange is true, only Target, Iterable, and Body should be set.
// When IsRange is false, only Init, Condition, Post, and Body should be set.
// The parser must enforce this invariant.
type ForStmt struct {
	SpanInfo  lexer.Span
	Init      Stmt // optional - only used when IsRange is false
	Condition Expr // optional - only used when IsRange is false
	Post      Stmt // optional - only used when IsRange is false
	Target    *IdentifierExpr // only used when IsRange is true
	Iterable  Expr            // only used when IsRange is true
	Body      *BlockStmt
	IsRange   bool // true for range form, false for classic form
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

// WhileStmt expresses a while loop.
type WhileStmt struct {
	SpanInfo lexer.Span
	Cond     Expr
	Body     *BlockStmt
}

func (s *WhileStmt) Span() lexer.Span { return s.SpanInfo }
func (s *WhileStmt) node()            {}
func (s *WhileStmt) stmt()            {}

// BreakStmt exits a loop.
type BreakStmt struct {
	SpanInfo lexer.Span
}

func (s *BreakStmt) Span() lexer.Span { return s.SpanInfo }
func (s *BreakStmt) node()            {}
func (s *BreakStmt) stmt()            {}

// ContinueStmt skips to the next loop iteration.
type ContinueStmt struct {
	SpanInfo lexer.Span
}

func (s *ContinueStmt) Span() lexer.Span { return s.SpanInfo }
func (s *ContinueStmt) node()            {}
func (s *ContinueStmt) stmt()            {}

// TryStmt represents a try-catch-finally block.
type TryStmt struct {
	SpanInfo     lexer.Span
	TryBlock     *BlockStmt
	CatchClauses []*CatchClause
	FinallyBlock *BlockStmt
}

func (s *TryStmt) Span() lexer.Span { return s.SpanInfo }
func (s *TryStmt) node()            {}
func (s *TryStmt) stmt()            {}

// CatchClause represents a catch block with optional exception variable.
type CatchClause struct {
	SpanInfo      lexer.Span
	ExceptionVar  string // optional exception variable name
	ExceptionType string // optional exception type
	Block         *BlockStmt
}

func (c *CatchClause) Span() lexer.Span { return c.SpanInfo }
func (c *CatchClause) node()            {}
func (c *CatchClause) stmt()            {}

// ThrowStmt represents a throw statement.
type ThrowStmt struct {
	SpanInfo lexer.Span
	Expr     Expr
}

func (s *ThrowStmt) Span() lexer.Span { return s.SpanInfo }
func (s *ThrowStmt) node()            {}
func (s *ThrowStmt) stmt()            {}

// TypeAliasDecl represents a type alias declaration: type UserID = int
type TypeAliasDecl struct {
	SpanInfo   lexer.Span
	Name       string
	TypeParams []string // For generic type aliases
	Type       *TypeExpr // Pointer to avoid copying and allow updates to propagate
}

func (d *TypeAliasDecl) Span() lexer.Span { return d.SpanInfo }
func (d *TypeAliasDecl) node()            {}
func (d *TypeAliasDecl) decl()            {}

// GenericTypeExpr represents a generic type: T or Container<T>
type GenericTypeExpr struct {
	SpanInfo lexer.Span
	Name     string
	TypeArgs []TypeExpr
}

func (e *GenericTypeExpr) Span() lexer.Span { return e.SpanInfo }
func (e *GenericTypeExpr) node()            {}
func (e *GenericTypeExpr) typeExpr()        {}

// UnionTypeExpr represents a union type: int | string | bool
type UnionTypeExpr struct {
	SpanInfo lexer.Span
	Types    []TypeExpr
}

func (e *UnionTypeExpr) Span() lexer.Span { return e.SpanInfo }
func (e *UnionTypeExpr) node()            {}
func (e *UnionTypeExpr) typeExpr()        {}

// OptionalTypeExpr represents an optional type: T?
type OptionalTypeExpr struct {
	SpanInfo  lexer.Span
	InnerType TypeExpr
}

func (e *OptionalTypeExpr) Span() lexer.Span { return e.SpanInfo }
func (e *OptionalTypeExpr) node()            {}
func (e *OptionalTypeExpr) typeExpr()        {}

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
	LiteralNull   LiteralKind = "null"
	LiteralHex    LiteralKind = "hex"
	LiteralBinary LiteralKind = "binary"
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

// StringInterpolationExpr represents string interpolation: "Hello, ${name}!"
type StringInterpolationExpr struct {
	SpanInfo lexer.Span
	Parts    []StringInterpolationPart
}

func (e *StringInterpolationExpr) Span() lexer.Span { return e.SpanInfo }
func (e *StringInterpolationExpr) node()            {}
func (e *StringInterpolationExpr) expr()            {}

// StringInterpolationPart represents either a string literal or an expression in string interpolation
type StringInterpolationPart struct {
	IsLiteral bool   // true for string literals, false for expressions
	Literal   string // the string literal part
	Expr      Expr   // the expression part (when IsLiteral is false)
	Span      lexer.Span
}

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

// NewExpr models the new keyword for memory allocation.
type NewExpr struct {
	SpanInfo lexer.Span
	Type     *TypeExpr
}

func (e *NewExpr) Span() lexer.Span { return e.SpanInfo }
func (e *NewExpr) node()            {}
func (e *NewExpr) expr()            {}

// DeleteExpr models the delete keyword for memory deallocation.
type DeleteExpr struct {
	SpanInfo lexer.Span
	Target   Expr
}

func (e *DeleteExpr) Span() lexer.Span { return e.SpanInfo }
func (e *DeleteExpr) node()            {}
func (e *DeleteExpr) expr()            {}

// LambdaExpr models lambda/anonymous functions: |a, b| a + b
type LambdaExpr struct {
	SpanInfo lexer.Span
	Params   []Param
	Body     Expr
}

func (e *LambdaExpr) Span() lexer.Span { return e.SpanInfo }
func (e *LambdaExpr) node()            {}
func (e *LambdaExpr) expr()            {}

// CastExpr represents a type cast expression: (type) expression
type CastExpr struct {
	SpanInfo lexer.Span
	Type     *TypeExpr
	Expr     Expr
}

func (e *CastExpr) Span() lexer.Span { return e.SpanInfo }
func (e *CastExpr) node()            {}
func (e *CastExpr) expr()            {}

// AwaitExpr represents an await expression: await expression
type AwaitExpr struct {
	SpanInfo lexer.Span
	Expr     Expr
}

func (e *AwaitExpr) Span() lexer.Span { return e.SpanInfo }
func (e *AwaitExpr) node()            {}
func (e *AwaitExpr) expr()            {}
