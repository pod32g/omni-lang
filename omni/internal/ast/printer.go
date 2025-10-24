package ast

import (
	"bytes"
	"sort"
)

// Print returns a deterministic, human readable string representation of the AST.
func Print(mod *Module) string {
	var buf bytes.Buffer
	writer := printer{buf: &buf}
	writer.writeModule(mod)
	return buf.String()
}

type printer struct {
	buf   *bytes.Buffer
	level int
}

func (p *printer) writeModule(mod *Module) {
	p.writeLine("Module {")
	p.indent(func() {
		if len(mod.Imports) > 0 {
			p.writeLine("Imports [")
			p.indent(func() {
				for _, imp := range mod.Imports {
					p.writeImport(imp)
				}
			})
			p.writeLine("]")
		}
		if len(mod.Decls) > 0 {
			p.writeLine("Decls [")
			p.indent(func() {
				for _, decl := range mod.Decls {
					p.writeDecl(decl)
				}
			})
			p.writeLine("]")
		}
	})
	p.writeLine("}")
}

func (p *printer) writeImport(imp *ImportDecl) {
	p.writeLine("Import " + joinPath(imp.Path))
}

func (p *printer) writeDecl(decl Decl) {
	switch d := decl.(type) {
	case *LetDecl:
		p.writeLine("LetDecl {")
		p.indent(func() {
			p.writeLine("Name " + d.Name)
			if d.Type != nil {
				p.writeLine("Type " + p.formatType(d.Type))
			}
			if d.Value != nil {
				p.writeLine("Value")
				p.indent(func() { p.writeExpr(d.Value) })
			}
		})
		p.writeLine("}")
	case *VarDecl:
		p.writeLine("VarDecl {")
		p.indent(func() {
			p.writeLine("Name " + d.Name)
			if d.Type != nil {
				p.writeLine("Type " + p.formatType(d.Type))
			}
			if d.Value != nil {
				p.writeLine("Value")
				p.indent(func() { p.writeExpr(d.Value) })
			}
		})
		p.writeLine("}")
	case *StructDecl:
		p.writeLine("StructDecl {")
		p.indent(func() {
			p.writeLine("Name " + d.Name)
			if len(d.Fields) > 0 {
				p.writeLine("Fields [")
				p.indent(func() {
					for _, field := range d.Fields {
						p.writeLine(field.Name + ": " + p.formatType(field.Type))
					}
				})
				p.writeLine("]")
			}
		})
		p.writeLine("}")
	case *EnumDecl:
		p.writeLine("EnumDecl {")
		p.indent(func() {
			p.writeLine("Name " + d.Name)
			if len(d.Variants) > 0 {
				p.writeLine("Variants [")
				p.indent(func() {
					for _, v := range d.Variants {
						p.writeLine(v.Name)
					}
				})
				p.writeLine("]")
			}
		})
		p.writeLine("}")
	case *FuncDecl:
		p.writeLine("FuncDecl {")
		p.indent(func() {
			p.writeLine("Name " + d.Name)
			if len(d.Params) > 0 {
				p.writeLine("Params [")
				p.indent(func() {
					for _, param := range d.Params {
						p.writeLine(param.Name + ": " + p.formatType(param.Type))
					}
				})
				p.writeLine("]")
			}
			if d.Return != nil {
				p.writeLine("Return " + p.formatType(d.Return))
			}
			if d.ExprBody != nil {
				p.writeLine("ExprBody")
				p.indent(func() { p.writeExpr(d.ExprBody) })
			} else if d.Body != nil {
				p.writeLine("Body")
				p.indent(func() { p.writeBlock(d.Body) })
			}
		})
		p.writeLine("}")
	default:
		p.writeLine("<unknown decl>")
	}
}

func (p *printer) writeBlock(block *BlockStmt) {
	p.writeLine("Block {")
	p.indent(func() {
		for _, stmt := range block.Statements {
			p.writeStmt(stmt)
		}
	})
	p.writeLine("}")
}

func (p *printer) writeStmt(stmt Stmt) {
	switch s := stmt.(type) {
	case *ReturnStmt:
		p.writeLine("ReturnStmt {")
		p.indent(func() {
			if s.Value != nil {
				p.writeLine("Value")
				p.indent(func() { p.writeExpr(s.Value) })
			}
		})
		p.writeLine("}")
	case *ExprStmt:
		p.writeLine("ExprStmt")
		p.indent(func() { p.writeExpr(s.Expr) })
	case *IfStmt:
		p.writeLine("IfStmt {")
		p.indent(func() {
			p.writeLine("Cond")
			p.indent(func() { p.writeExpr(s.Cond) })
			p.writeLine("Then")
			p.indent(func() { p.writeBlock(s.Then) })
			if s.Else != nil {
				p.writeLine("Else")
				p.indent(func() { p.writeStmt(s.Else) })
			}
		})
		p.writeLine("}")
	case *BlockStmt:
		p.writeBlock(s)
	case *ForStmt:
		p.writeLine("ForStmt {")
		p.indent(func() {
			if s.IsRange {
				if s.Target != nil {
					p.writeLine("Target " + s.Target.Name)
				}
				p.writeLine("Iterable")
				p.indent(func() { p.writeExpr(s.Iterable) })
			} else {
				if s.Init != nil {
					p.writeLine("Init")
					p.indent(func() { p.writeStmt(s.Init) })
				}
				if s.Condition != nil {
					p.writeLine("Cond")
					p.indent(func() { p.writeExpr(s.Condition) })
				}
				if s.Post != nil {
					p.writeLine("Post")
					p.indent(func() { p.writeStmt(s.Post) })
				}
			}
			p.writeLine("Body")
			p.indent(func() { p.writeBlock(s.Body) })
		})
		p.writeLine("}")
	case *BindingStmt:
		mut := "let"
		if s.Mutable {
			mut = "var"
		}
		p.writeLine("BindingStmt " + mut + " {")
		p.indent(func() {
			p.writeLine("Name " + s.Name)
			if s.Type != nil {
				p.writeLine("Type " + p.formatType(s.Type))
			}
			if s.Value != nil {
				p.writeLine("Value")
				p.indent(func() { p.writeExpr(s.Value) })
			}
		})
		p.writeLine("}")
	case *ShortVarDeclStmt:
		p.writeLine("ShortVarDecl {")
		p.indent(func() {
			p.writeLine("Name " + s.Name)
			if s.Type != nil {
				p.writeLine("Type " + p.formatType(s.Type))
			}
			if s.Value != nil {
				p.writeLine("Value")
				p.indent(func() { p.writeExpr(s.Value) })
			}
		})
		p.writeLine("}")
	case *AssignmentStmt:
		p.writeLine("AssignmentStmt")
		p.indent(func() {
			p.writeExpr(s.Left)
			p.writeExpr(s.Right)
		})
	case *IncrementStmt:
		p.writeLine("IncrementStmt " + s.Op)
		p.indent(func() { p.writeExpr(s.Target) })
	default:
		p.writeLine("<unknown stmt>")
	}
}

func (p *printer) writeExpr(expr Expr) {
	switch e := expr.(type) {
	case *IdentifierExpr:
		p.writeLine("Identifier " + e.Name)
	case *LiteralExpr:
		p.writeLine("Literal " + string(e.Kind) + " " + e.Value)
	case *StringInterpolationExpr:
		p.writeLine("StringInterpolation")
		p.indent(func() {
			for _, part := range e.Parts {
				if part.IsLiteral {
					p.writeLine("LiteralPart " + part.Literal)
				} else {
					p.writeLine("ExprPart")
					p.indent(func() { p.writeExpr(part.Expr) })
				}
			}
		})
	case *UnaryExpr:
		p.writeLine("Unary " + e.Op)
		p.indent(func() { p.writeExpr(e.Expr) })
	case *BinaryExpr:
		p.writeLine("Binary " + e.Op)
		p.indent(func() {
			p.writeExpr(e.Left)
			p.writeExpr(e.Right)
		})
	case *CallExpr:
		p.writeLine("Call")
		p.indent(func() {
			p.writeLine("Callee")
			p.indent(func() { p.writeExpr(e.Callee) })
			if len(e.Args) > 0 {
				p.writeLine("Args [")
				p.indent(func() {
					for _, arg := range e.Args {
						p.writeExpr(arg)
					}
				})
				p.writeLine("]")
			}
		})
	case *IndexExpr:
		p.writeLine("Index")
		p.indent(func() {
			p.writeExpr(e.Target)
			p.writeExpr(e.Index)
		})
	case *MemberExpr:
		p.writeLine("Member " + e.Member)
		p.indent(func() { p.writeExpr(e.Target) })
	case *ArrayLiteralExpr:
		p.writeLine("ArrayLiteral [")
		p.indent(func() {
			for _, el := range e.Elements {
				p.writeExpr(el)
			}
		})
		p.writeLine("]")
	case *MapLiteralExpr:
		p.writeLine("MapLiteral {")
		p.indent(func() {
			sorted := make([]MapEntry, len(e.Entries))
			copy(sorted, e.Entries)
			sort.Slice(sorted, func(i, j int) bool {
				return sorted[i].Span.Start.Column < sorted[j].Span.Start.Column
			})
			for _, entry := range sorted {
				p.writeLine("Entry")
				p.indent(func() {
					p.writeExpr(entry.Key)
					p.writeExpr(entry.Value)
				})
			}
		})
		p.writeLine("}")
	case *StructLiteralExpr:
		p.writeLine("StructLiteral " + e.TypeName + " {")
		p.indent(func() {
			for _, field := range e.Fields {
				p.writeLine(field.Name)
				p.indent(func() { p.writeExpr(field.Expr) })
			}
		})
		p.writeLine("}")
	case *AssignmentExpr:
		p.writeLine("Assignment")
		p.indent(func() {
			p.writeExpr(e.Left)
			p.writeExpr(e.Right)
		})
	case *IncrementExpr:
		p.writeLine("Increment " + e.Op)
		p.indent(func() { p.writeExpr(e.Target) })
	default:
		p.writeLine("<unknown expr>")
	}
}

func (p *printer) formatType(t *TypeExpr) string {
	if t == nil {
		return "<nil>"
	}
	if len(t.Args) == 0 {
		return t.Name
	}
	var buf bytes.Buffer
	buf.WriteString(t.Name)
	buf.WriteByte('<')
	for i, arg := range t.Args {
		if i > 0 {
			buf.WriteString(",")
		}
		buf.WriteString(p.formatType(arg))
	}
	buf.WriteByte('>')
	return buf.String()
}

func (p *printer) indent(fn func()) {
	p.level++
	fn()
	p.level--
}

func (p *printer) writeLine(s string) {
	for i := 0; i < p.level; i++ {
		p.buf.WriteString("  ")
	}
	p.buf.WriteString(s)
	p.buf.WriteByte('\n')
}

func joinPath(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	res := parts[0]
	for _, p := range parts[1:] {
		res += "." + p
	}
	return res
}
