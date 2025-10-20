package parser

import (
	"errors"
	"fmt"
	"strings"

	"github.com/omni-lang/omni/internal/ast"
	"github.com/omni-lang/omni/internal/lexer"
)

// Parse consumes the provided source and returns an abstract syntax tree.
func Parse(filename, input string) (*ast.Module, error) {
	tokens, err := lexer.LexAll(filename, input)
	if err != nil {
		return nil, err
	}
	p := &Parser{
		filename: filename,
		tokens:   tokens,
		lines:    splitLines(input),
	}
	mod, err := p.parseModule()
	if err != nil {
		return mod, err
	}
	if len(p.diagnostics) > 0 {
		return mod, errors.Join(p.diagnostics...)
	}
	return mod, nil
}

// Parser implements a recursive descent parser for OmniLang.
type Parser struct {
	filename    string
	tokens      []lexer.Token
	pos         int
	lines       []string
	diagnostics []error
}

type parsePanic struct {
	diag error
}

func (p *Parser) parseModule() (*ast.Module, error) {
	defer func() {
		if r := recover(); r != nil {
			if pp, ok := r.(parsePanic); ok {
				p.addError(pp.diag)
			} else {
				panic(r)
			}
		}
	}()

	start := p.current().Span.Start
	module := &ast.Module{SpanInfo: lexer.Span{Start: start, End: start}}

	for p.peekKind() == lexer.TokenImport {
		if imp, ok := p.parseImportSafe(); ok {
			module.Imports = append(module.Imports, imp)
		}
	}

	for p.peekKind() != lexer.TokenEOF {
		decl, ok := p.parseDeclSafe()
		if !ok {
			continue
		}
		module.Decls = append(module.Decls, decl)
	}

	if len(module.Decls) > 0 {
		module.SpanInfo.End = module.Decls[len(module.Decls)-1].Span().End
	} else if len(module.Imports) > 0 {
		module.SpanInfo.End = module.Imports[len(module.Imports)-1].Span().End
	}

	eof := p.expect(lexer.TokenEOF)
	module.SpanInfo.End = eof.Span.End
	if len(p.diagnostics) > 0 {
		return module, errors.Join(p.diagnostics...)
	}
	return module, nil
}

func (p *Parser) parseImportSafe() (*ast.ImportDecl, bool) {
	defer func() {
		if r := recover(); r != nil {
			if pp, ok := r.(parsePanic); ok {
				p.addError(pp.diag)
				p.synchronizeDecl()
			} else {
				panic(r)
			}
		}
	}()
	imp, err := p.parseImport()
	if err != nil {
		p.addError(err)
		p.synchronizeDecl()
		return nil, false
	}
	return imp, true
}

func (p *Parser) parseDeclSafe() (ast.Decl, bool) {
	defer func() {
		if r := recover(); r != nil {
			if pp, ok := r.(parsePanic); ok {
				p.addError(pp.diag)
				p.synchronizeDecl()
			} else {
				panic(r)
			}
		}
	}()
	decl, err := p.parseDecl()
	if err != nil {
		p.addError(err)
		p.synchronizeDecl()
		return nil, false
	}
	return decl, true
}

func (p *Parser) parseImport() (*ast.ImportDecl, error) {
	tok := p.expect(lexer.TokenImport)
	parts := []string{}
	for {
		ident := p.expect(lexer.TokenIdentifier)
		parts = append(parts, ident.Lexeme)
		if !p.match(lexer.TokenDot) {
			break
		}
	}

	alias := ""
	// Optional: "as <alias>"
	if p.match(lexer.TokenAs) { // requires lexer to have TokenAs for 'as'
		a := p.expect(lexer.TokenIdentifier)
		alias = a.Lexeme
	}

	last := p.previous()
	span := lexer.Span{Start: tok.Span.Start, End: last.Span.End}
	return &ast.ImportDecl{SpanInfo: span, Path: parts, Alias: alias}, nil
}

func (p *Parser) parseDecl() (ast.Decl, error) {
	switch p.peekKind() {
	case lexer.TokenImport:
		return p.parseImport()
	case lexer.TokenLet:
		return p.parseLetDecl(false)
	case lexer.TokenVar:
		return p.parseLetDecl(true)
	case lexer.TokenStruct:
		return p.parseStructDecl()
	case lexer.TokenEnum:
		return p.parseEnumDecl()
	case lexer.TokenFunc:
		return p.parseFuncDecl()
	default:
		return nil, p.errorAtCurrent("unexpected token at top level: %s", p.peekKind())
	}
}

func (p *Parser) parseLetDecl(mutable bool) (ast.Decl, error) {
	kw := p.advance()
	nameTok := p.expect(lexer.TokenIdentifier)
	var typ *ast.TypeExpr
	if p.match(lexer.TokenColon) {
		t, err := p.parseTypeExpr()
		if err != nil {
			return nil, err
		}
		typ = t
	}
	p.expect(lexer.TokenAssign)
	value, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	span := lexer.Span{Start: kw.Span.Start, End: value.Span().End}
	if mutable {
		return &ast.VarDecl{SpanInfo: span, Name: nameTok.Lexeme, Type: typ, Value: value}, nil
	}
	return &ast.LetDecl{SpanInfo: span, Name: nameTok.Lexeme, Type: typ, Value: value}, nil
}

func (p *Parser) parseStructDecl() (ast.Decl, error) {
	kw := p.advance()
	nameTok := p.expect(lexer.TokenIdentifier)

	// Parse generic type parameters
	var typeParams []ast.TypeParam
	if p.match(lexer.TokenLess) {
		for {
			paramName := p.expect(lexer.TokenIdentifier)
			typeParams = append(typeParams, ast.TypeParam{Name: paramName.Lexeme, Span: paramName.Span})
			if p.match(lexer.TokenComma) {
				continue
			}
			p.expect(lexer.TokenGreater)
			break
		}
	}

	p.expect(lexer.TokenLBrace)
	fields := []ast.StructField{}
	for !p.match(lexer.TokenRBrace) {
		fieldName := p.expect(lexer.TokenIdentifier)
		p.expect(lexer.TokenColon)
		typ, err := p.parseTypeExpr()
		if err != nil {
			return nil, err
		}
		fields = append(fields, ast.StructField{Name: fieldName.Lexeme, Type: typ, Span: fieldName.Span})
		if p.match(lexer.TokenRBrace) {
			break
		}
	}
	span := lexer.Span{Start: kw.Span.Start, End: p.previous().Span.End}
	return &ast.StructDecl{SpanInfo: span, Name: nameTok.Lexeme, TypeParams: typeParams, Fields: fields}, nil
}

func (p *Parser) parseEnumDecl() (ast.Decl, error) {
	kw := p.advance()
	nameTok := p.expect(lexer.TokenIdentifier)
	p.expect(lexer.TokenLBrace)
	variants := []ast.EnumVariant{}
	for !p.match(lexer.TokenRBrace) {
		ident := p.expect(lexer.TokenIdentifier)
		variants = append(variants, ast.EnumVariant{Name: ident.Lexeme, Span: ident.Span})
		if p.match(lexer.TokenRBrace) {
			break
		}
	}
	span := lexer.Span{Start: kw.Span.Start, End: p.previous().Span.End}
	return &ast.EnumDecl{SpanInfo: span, Name: nameTok.Lexeme, Variants: variants}, nil
}

func (p *Parser) parseFuncDecl() (ast.Decl, error) {
	kw := p.advance()
	nameTok := p.expect(lexer.TokenIdentifier)

	// Parse generic type parameters
	var typeParams []ast.TypeParam
	if p.match(lexer.TokenLess) {
		for {
			paramName := p.expect(lexer.TokenIdentifier)
			typeParams = append(typeParams, ast.TypeParam{Name: paramName.Lexeme, Span: paramName.Span})
			if p.match(lexer.TokenComma) {
				continue
			}
			p.expect(lexer.TokenGreater)
			break
		}
	}

	p.expect(lexer.TokenLParen)
	params := []ast.Param{}
	if !p.match(lexer.TokenRParen) {
		for {
			paramName := p.expect(lexer.TokenIdentifier)
			p.expect(lexer.TokenColon)
			typ, err := p.parseTypeExpr()
			if err != nil {
				return nil, err
			}
			params = append(params, ast.Param{Name: paramName.Lexeme, Type: typ, Span: paramName.Span})
			if p.match(lexer.TokenComma) {
				continue
			}
			p.expect(lexer.TokenRParen)
			break
		}
	}
	var retType *ast.TypeExpr
	if p.match(lexer.TokenColon) {
		t, err := p.parseTypeExpr()
		if err != nil {
			return nil, err
		}
		retType = t
	}
	if p.match(lexer.TokenFatArrow) {
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		span := lexer.Span{Start: kw.Span.Start, End: expr.Span().End}
		return &ast.FuncDecl{SpanInfo: span, Name: nameTok.Lexeme, TypeParams: typeParams, Params: params, Return: retType, ExprBody: expr}, nil
	}
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	span := lexer.Span{Start: kw.Span.Start, End: body.Span().End}
	return &ast.FuncDecl{SpanInfo: span, Name: nameTok.Lexeme, TypeParams: typeParams, Params: params, Return: retType, Body: body}, nil
}

func (p *Parser) parseBlock() (*ast.BlockStmt, error) {
	lbrace := p.expect(lexer.TokenLBrace)
	stmts := []ast.Stmt{}
	for p.peekKind() != lexer.TokenRBrace && p.peekKind() != lexer.TokenEOF {
		stmt, ok := p.parseStmtSafe()
		if !ok {
			continue
		}
		stmts = append(stmts, stmt)
	}
	rbrace := p.expect(lexer.TokenRBrace)
	return &ast.BlockStmt{SpanInfo: lexer.Span{Start: lbrace.Span.Start, End: rbrace.Span.End}, Statements: stmts}, nil
}

func (p *Parser) parseStmtSafe() (ast.Stmt, bool) {
	defer func() {
		if r := recover(); r != nil {
			if pp, ok := r.(parsePanic); ok {
				p.addError(pp.diag)
				p.synchronizeStmt()
			} else {
				panic(r)
			}
		}
	}()
	stmt, err := p.parseStmt()
	if err != nil {
		p.addError(err)
		p.synchronizeStmt()
		return nil, false
	}
	return stmt, true
}

func (p *Parser) parseStmt() (ast.Stmt, error) {
	switch p.peekKind() {
	case lexer.TokenReturn:
		return p.parseReturnStmt()
	case lexer.TokenIf:
		return p.parseIfStmt()
	case lexer.TokenFor:
		return p.parseForStmt()
	case lexer.TokenLet:
		return p.parseBindingStmt(false)
	case lexer.TokenVar:
		return p.parseBindingStmt(true)
	default:
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		// Postfix increment/decrement as statements
		if inc, ok := expr.(*ast.IncrementExpr); ok {
			return &ast.IncrementStmt{SpanInfo: inc.SpanInfo, Target: inc.Target, Op: inc.Op}, nil
		}
		return &ast.ExprStmt{SpanInfo: expr.Span(), Expr: expr}, nil
	}
}

func (p *Parser) parseReturnStmt() (ast.Stmt, error) {
	tok := p.advance()
	if p.peekKind() == lexer.TokenRBrace {
		return &ast.ReturnStmt{SpanInfo: tok.Span}, nil
	}
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	span := lexer.Span{Start: tok.Span.Start, End: expr.Span().End}
	return &ast.ReturnStmt{SpanInfo: span, Value: expr}, nil
}

func (p *Parser) parseIfStmt() (ast.Stmt, error) {
	tok := p.advance()
	cond, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	thenBlock, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	var elseStmt ast.Stmt
	if p.match(lexer.TokenElse) {
		if p.peekKind() == lexer.TokenIf {
			elseStmt, err = p.parseIfStmt()
			if err != nil {
				return nil, err
			}
		} else {
			elseBlock, err := p.parseBlock()
			if err != nil {
				return nil, err
			}
			elseStmt = elseBlock
		}
	}
	end := thenBlock.Span().End
	if elseStmt != nil {
		end = elseStmt.Span().End
	}
	return &ast.IfStmt{SpanInfo: lexer.Span{Start: tok.Span.Start, End: end}, Cond: cond, Then: thenBlock, Else: elseStmt}, nil
}

func (p *Parser) parseForStmt() (ast.Stmt, error) {
	tok := p.advance()
	// detect range form: for ident in expr
	if p.peekKind() == lexer.TokenIdentifier && p.peekKindN(1) == lexer.TokenIn {
		targetTok := p.advance()
		p.advance() // consume 'in'
		iterable, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		body, err := p.parseBlock()
		if err != nil {
			return nil, err
		}
		span := lexer.Span{Start: tok.Span.Start, End: body.Span().End}
		return &ast.ForStmt{SpanInfo: span, Target: &ast.IdentifierExpr{Name: targetTok.Lexeme, SpanInfo: targetTok.Span}, Iterable: iterable, Body: body, IsRange: true}, nil
	}

	// Classic for loop grammar: for init; cond; post { ... }
	var init ast.Stmt
	if p.peekKind() != lexer.TokenSemicolon {
		stmt, err := p.parseForInit()
		if err != nil {
			return nil, err
		}
		init = stmt
	}
	p.expect(lexer.TokenSemicolon)

	var cond ast.Expr
	if p.peekKind() != lexer.TokenSemicolon {
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		cond = expr
	}
	p.expect(lexer.TokenSemicolon)

	var post ast.Stmt
	if p.peekKind() != lexer.TokenLBrace {
		stmt, err := p.parseForPost()
		if err != nil {
			return nil, err
		}
		post = stmt
	}

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	span := lexer.Span{Start: tok.Span.Start, End: body.Span().End}
	return &ast.ForStmt{SpanInfo: span, Init: init, Condition: cond, Post: post, Body: body}, nil
}

func (p *Parser) parseForInit() (ast.Stmt, error) {
	if p.peekKind() == lexer.TokenIdentifier && p.peekKindN(1) == lexer.TokenColon {
		nameTok := p.advance()
		p.advance() // colon
		typ, err := p.parseTypeExpr()
		if err != nil {
			return nil, err
		}
		p.expect(lexer.TokenAssign)
		value, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		span := lexer.Span{Start: nameTok.Span.Start, End: value.Span().End}
		return &ast.ShortVarDeclStmt{SpanInfo: span, Name: nameTok.Lexeme, Type: typ, Value: value}, nil
	}
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if assign, ok := expr.(*ast.AssignmentExpr); ok {
		return &ast.AssignmentStmt{SpanInfo: assign.SpanInfo, Left: assign.Left, Right: assign.Right}, nil
	}
	return &ast.ExprStmt{SpanInfo: expr.Span(), Expr: expr}, nil
}

func (p *Parser) parseForPost() (ast.Stmt, error) {
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	switch e := expr.(type) {
	case *ast.AssignmentExpr:
		return &ast.AssignmentStmt{SpanInfo: e.SpanInfo, Left: e.Left, Right: e.Right}, nil
	case *ast.IncrementExpr:
		return &ast.IncrementStmt{SpanInfo: e.SpanInfo, Target: e.Target, Op: e.Op}, nil
	default:
		return &ast.ExprStmt{SpanInfo: e.Span(), Expr: e}, nil
	}
}

func (p *Parser) parseBindingStmt(mutable bool) (ast.Stmt, error) {
	kw := p.advance()
	nameTok := p.expect(lexer.TokenIdentifier)
	var typ *ast.TypeExpr
	if p.match(lexer.TokenColon) {
		t, err := p.parseTypeExpr()
		if err != nil {
			return nil, err
		}
		typ = t
	}
	p.expect(lexer.TokenAssign)
	value, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	span := lexer.Span{Start: kw.Span.Start, End: value.Span().End}
	return &ast.BindingStmt{SpanInfo: span, Mutable: mutable, Name: nameTok.Lexeme, Type: typ, Value: value}, nil
}

// Expression parsing -------------------------------------------------------

func (p *Parser) parseExpr() (ast.Expr, error) {
	return p.parseAssignment()
}

func (p *Parser) parseAssignment() (ast.Expr, error) {
	left, err := p.parseLogicalOr()
	if err != nil {
		return nil, err
	}
	if p.match(lexer.TokenAssign) {
		right, err := p.parseAssignment()
		if err != nil {
			return nil, err
		}
		span := lexer.Span{Start: left.Span().Start, End: right.Span().End}
		return &ast.AssignmentExpr{SpanInfo: span, Left: left, Right: right}, nil
	}
	return left, nil
}

func (p *Parser) parseLogicalOr() (ast.Expr, error) {
	left, err := p.parseLogicalAnd()
	if err != nil {
		return nil, err
	}
	for p.match(lexer.TokenOrOr) {
		op := p.previous().Lexeme
		right, err := p.parseLogicalAnd()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{SpanInfo: lexer.Span{Start: left.Span().Start, End: right.Span().End}, Left: left, Op: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parseLogicalAnd() (ast.Expr, error) {
	left, err := p.parseEquality()
	if err != nil {
		return nil, err
	}
	for p.match(lexer.TokenAndAnd) {
		op := p.previous().Lexeme
		right, err := p.parseEquality()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{SpanInfo: lexer.Span{Start: left.Span().Start, End: right.Span().End}, Left: left, Op: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parseEquality() (ast.Expr, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}
	for p.match(lexer.TokenEqualEqual) || p.match(lexer.TokenBangEqual) {
		op := p.previous()
		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{SpanInfo: lexer.Span{Start: left.Span().Start, End: right.Span().End}, Left: left, Op: op.Lexeme, Right: right}
	}
	return left, nil
}

func (p *Parser) parseComparison() (ast.Expr, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	for {
		switch p.peekKind() {
		case lexer.TokenLess, lexer.TokenLessEqual, lexer.TokenGreater, lexer.TokenGreaterEqual:
			op := p.advance()
			right, err := p.parseTerm()
			if err != nil {
				return nil, err
			}
			left = &ast.BinaryExpr{SpanInfo: lexer.Span{Start: left.Span().Start, End: right.Span().End}, Left: left, Op: op.Lexeme, Right: right}
		default:
			return left, nil
		}
	}
}

func (p *Parser) parseTerm() (ast.Expr, error) {
	left, err := p.parseFactor()
	if err != nil {
		return nil, err
	}
	for {
		switch p.peekKind() {
		case lexer.TokenPlus, lexer.TokenMinus:
			op := p.advance()
			right, err := p.parseFactor()
			if err != nil {
				return nil, err
			}
			left = &ast.BinaryExpr{SpanInfo: lexer.Span{Start: left.Span().Start, End: right.Span().End}, Left: left, Op: op.Lexeme, Right: right}
		default:
			return left, nil
		}
	}
}

func (p *Parser) parseFactor() (ast.Expr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for {
		switch p.peekKind() {
		case lexer.TokenStar, lexer.TokenSlash, lexer.TokenPercent:
			op := p.advance()
			right, err := p.parseUnary()
			if err != nil {
				return nil, err
			}
			left = &ast.BinaryExpr{SpanInfo: lexer.Span{Start: left.Span().Start, End: right.Span().End}, Left: left, Op: op.Lexeme, Right: right}
		default:
			return left, nil
		}
	}
}

func (p *Parser) parseUnary() (ast.Expr, error) {
	switch p.peekKind() {
	case lexer.TokenBang, lexer.TokenMinus:
		op := p.advance()
		expr, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpr{SpanInfo: lexer.Span{Start: op.Span.Start, End: expr.Span().End}, Op: op.Lexeme, Expr: expr}, nil
	}
	return p.parsePostfix()
}

func (p *Parser) parsePostfix() (ast.Expr, error) {
	expr, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	for {
		switch p.peekKind() {
		case lexer.TokenLParen:
			expr, err = p.finishCall(expr)
			if err != nil {
				return nil, err
			}
		case lexer.TokenLBrace:
			if !p.isStructLiteralContext(expr) {
				return expr, nil
			}
			expr, err = p.parseStructLiteral(expr)
			if err != nil {
				return nil, err
			}
		case lexer.TokenLBracket:
			p.advance()
			index, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			rbrack := p.expect(lexer.TokenRBracket)
			expr = &ast.IndexExpr{SpanInfo: lexer.Span{Start: expr.Span().Start, End: rbrack.Span.End}, Target: expr, Index: index}
		case lexer.TokenDot:
			p.advance()
			member := p.expect(lexer.TokenIdentifier)
			expr = &ast.MemberExpr{SpanInfo: lexer.Span{Start: expr.Span().Start, End: member.Span.End}, Target: expr, Member: member.Lexeme}
		case lexer.TokenArrow:
			p.advance()
			member := p.expect(lexer.TokenIdentifier)
			expr = &ast.MemberExpr{SpanInfo: lexer.Span{Start: expr.Span().Start, End: member.Span.End}, Target: expr, Member: member.Lexeme}
		case lexer.TokenPlusPlus:
			tok := p.advance()
			expr = &ast.IncrementExpr{SpanInfo: lexer.Span{Start: expr.Span().Start, End: tok.Span.End}, Target: expr, Op: tok.Lexeme}
		case lexer.TokenMinusMinus:
			tok := p.advance()
			expr = &ast.IncrementExpr{SpanInfo: lexer.Span{Start: expr.Span().Start, End: tok.Span.End}, Target: expr, Op: tok.Lexeme}
		default:
			return expr, nil
		}
	}
}

func (p *Parser) finishCall(callee ast.Expr) (ast.Expr, error) {
	p.expect(lexer.TokenLParen)
	args := []ast.Expr{}
	if p.peekKind() != lexer.TokenRParen {
		for {
			arg, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
			if p.match(lexer.TokenComma) {
				continue
			}
			break
		}
	}
	rparen := p.expect(lexer.TokenRParen)
	span := lexer.Span{Start: callee.Span().Start, End: rparen.Span.End}
	return &ast.CallExpr{SpanInfo: span, Callee: callee, Args: args}, nil
}

func (p *Parser) parsePrimary() (ast.Expr, error) {
	tok := p.advance()
	switch tok.Kind {
	case lexer.TokenIdentifier:
		// Check if this is a std identifier that should be parsed as qualified
		if tok.Lexeme == "std" {
			return p.parseQualifiedIdentifier(tok)
		}
		return &ast.IdentifierExpr{SpanInfo: tok.Span, Name: tok.Lexeme}, nil
	case lexer.TokenIntLiteral:
		return &ast.LiteralExpr{SpanInfo: tok.Span, Kind: ast.LiteralInt, Value: tok.Lexeme}, nil
	case lexer.TokenFloatLiteral:
		return &ast.LiteralExpr{SpanInfo: tok.Span, Kind: ast.LiteralFloat, Value: tok.Lexeme}, nil
	case lexer.TokenStringLiteral:
		return &ast.LiteralExpr{SpanInfo: tok.Span, Kind: ast.LiteralString, Value: tok.Lexeme}, nil
	case lexer.TokenCharLiteral:
		return &ast.LiteralExpr{SpanInfo: tok.Span, Kind: ast.LiteralChar, Value: tok.Lexeme}, nil
	case lexer.TokenTrue, lexer.TokenFalse:
		return &ast.LiteralExpr{SpanInfo: tok.Span, Kind: ast.LiteralBool, Value: tok.Lexeme}, nil
	case lexer.TokenLParen:
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		p.expect(lexer.TokenRParen)
		return expr, nil
	case lexer.TokenLBracket:
		elems := []ast.Expr{}
		if !p.match(lexer.TokenRBracket) {
			for {
				expr, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				elems = append(elems, expr)
				if p.match(lexer.TokenComma) {
					continue
				}
				rbrack := p.expect(lexer.TokenRBracket)
				return &ast.ArrayLiteralExpr{SpanInfo: lexer.Span{Start: tok.Span.Start, End: rbrack.Span.End}, Elements: elems}, nil
			}
		}
		return &ast.ArrayLiteralExpr{SpanInfo: lexer.Span{Start: tok.Span.Start, End: p.previous().Span.End}, Elements: elems}, nil
	case lexer.TokenLBrace:
		return p.parseMapLiteral(tok)
	case lexer.TokenNew:
		return p.parseNewExpr(tok)
	case lexer.TokenDelete:
		return p.parseDeleteExpr(tok)
	default:
		return nil, p.errorAt(tok, "unexpected token %s", tok.Kind)
	}
}

func (p *Parser) parseStructLiteral(base ast.Expr) (ast.Expr, error) {
	ident, ok := base.(*ast.IdentifierExpr)
	if !ok {
		return nil, p.errorAtCurrent("struct literal must start with type identifier")
	}
	p.expect(lexer.TokenLBrace)
	fields := []ast.StructLiteralField{}
	if p.peekKind() != lexer.TokenRBrace {
		for {
			nameTok := p.expect(lexer.TokenIdentifier)
			p.expect(lexer.TokenColon)
			value, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			fields = append(fields, ast.StructLiteralField{Name: nameTok.Lexeme, Expr: value, Span: nameTok.Span})
			if p.match(lexer.TokenComma) {
				continue
			}
			break
		}
	}
	rbrace := p.expect(lexer.TokenRBrace)
	span := lexer.Span{Start: base.Span().Start, End: rbrace.Span.End}
	return &ast.StructLiteralExpr{SpanInfo: span, TypeName: ident.Name, Fields: fields}, nil
}

func (p *Parser) parseMapLiteral(lbrace lexer.Token) (ast.Expr, error) {
	entries := []ast.MapEntry{}
	if p.peekKind() != lexer.TokenRBrace {
		for {
			key, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			p.expect(lexer.TokenColon)
			value, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			entries = append(entries, ast.MapEntry{Key: key, Value: value, Span: lexer.Span{Start: key.Span().Start, End: value.Span().End}})
			if p.match(lexer.TokenComma) {
				continue
			}
			break
		}
	}
	rbrace := p.expect(lexer.TokenRBrace)
	return &ast.MapLiteralExpr{SpanInfo: lexer.Span{Start: lbrace.Span.Start, End: rbrace.Span.End}, Entries: entries}, nil
}

func (p *Parser) isStructLiteralContext(base ast.Expr) bool {
	if _, ok := base.(*ast.IdentifierExpr); !ok {
		return false
	}
	k1 := p.peekKindN(1)
	if k1 == lexer.TokenRBrace {
		return true
	}
	if k1 != lexer.TokenIdentifier {
		return false
	}
	k2 := p.peekKindN(2)
	return k2 == lexer.TokenColon
}

func (p *Parser) addError(err error) {
	if err == nil {
		return
	}
	p.diagnostics = append(p.diagnostics, err)
}

func (p *Parser) synchronizeDecl() {
	if p.peekKind() != lexer.TokenEOF {
		p.advance()
	}
	for {
		switch p.peekKind() {
		case lexer.TokenEOF, lexer.TokenFunc, lexer.TokenLet, lexer.TokenVar, lexer.TokenStruct, lexer.TokenEnum, lexer.TokenImport:
			return
		}
		p.advance()
	}
}

func (p *Parser) synchronizeStmt() {
	if p.peekKind() != lexer.TokenEOF {
		p.advance()
	}
	for {
		switch p.peekKind() {
		case lexer.TokenEOF, lexer.TokenRBrace, lexer.TokenReturn, lexer.TokenIf, lexer.TokenFor, lexer.TokenLet, lexer.TokenVar:
			return
		}
		p.advance()
	}
}

func (p *Parser) parseTypeExpr() (*ast.TypeExpr, error) {
	// Parse the first type
	firstType, err := p.parseSingleType()
	if err != nil {
		return nil, err
	}

	// Check if this is a union type (has | after it)
	if p.match(lexer.TokenPipe) {
		// This is a union type
		unionType := &ast.TypeExpr{
			SpanInfo: firstType.SpanInfo,
			IsUnion:  true,
			Members:  []*ast.TypeExpr{firstType},
		}

		// Parse additional union members
		for {
			member, err := p.parseSingleType()
			if err != nil {
				return nil, err
			}
			unionType.Members = append(unionType.Members, member)
			unionType.SpanInfo.End = member.SpanInfo.End

			if !p.match(lexer.TokenPipe) {
				break
			}
		}
		return unionType, nil
	}

	return firstType, nil
}

// parseSingleType parses a single type (not a union)
func (p *Parser) parseSingleType() (*ast.TypeExpr, error) {
	// Handle pointer types: *Type or *(Type)
	if p.match(lexer.TokenStar) {
		var baseType *ast.TypeExpr
		var err error

		if p.match(lexer.TokenLParen) {
			// Parse *(Type) - parentheses around type
			baseType, err = p.parseTypeExpr()
			if err != nil {
				return nil, err
			}
			p.expect(lexer.TokenRParen)
		} else {
			// Parse *Type - direct type
			baseType, err = p.parseSingleType()
			if err != nil {
				return nil, err
			}
		}

		span := lexer.Span{Start: p.previous().Span.Start, End: baseType.SpanInfo.End}
		// For union types, store the base type in Args
		if baseType.IsUnion {
			return &ast.TypeExpr{SpanInfo: span, Name: "*", Args: baseType.Members}, nil
		}
		return &ast.TypeExpr{SpanInfo: span, Name: "*" + baseType.Name}, nil
	}

	// Handle parentheses around types: (Type)
	if p.match(lexer.TokenLParen) {
		typ, err := p.parseTypeExpr()
		if err != nil {
			return nil, err
		}
		p.expect(lexer.TokenRParen)
		return typ, nil
	}

	nameTok := p.expect(lexer.TokenIdentifier)
	typeExpr := &ast.TypeExpr{SpanInfo: nameTok.Span, Name: nameTok.Lexeme}
	if p.match(lexer.TokenLess) {
		args := []*ast.TypeExpr{}
		if !p.match(lexer.TokenGreater) {
			for {
				arg, err := p.parseTypeExpr()
				if err != nil {
					return nil, err
				}
				args = append(args, arg)
				if p.match(lexer.TokenComma) {
					continue
				}
				gt := p.expect(lexer.TokenGreater)
				typeExpr.Args = args
				typeExpr.SpanInfo.End = gt.Span.End
				return typeExpr, nil
			}
		}
		typeExpr.Args = args
		typeExpr.SpanInfo.End = p.previous().Span.End
	}
	return typeExpr, nil
}

// Helpers ------------------------------------------------------------------

func (p *Parser) peekKind() lexer.Kind {
	if p.pos >= len(p.tokens) {
		return lexer.TokenEOF
	}
	return p.tokens[p.pos].Kind
}

func (p *Parser) peekKindN(n int) lexer.Kind {
	if p.pos+n >= len(p.tokens) {
		return lexer.TokenEOF
	}
	return p.tokens[p.pos+n].Kind
}

func (p *Parser) current() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Kind: lexer.TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) previous() lexer.Token {
	return p.tokens[p.pos-1]
}

func (p *Parser) advance() lexer.Token {
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return p.tokens[p.pos-1]
}

func (p *Parser) match(kind lexer.Kind) bool {
	if p.peekKind() == kind {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) expect(kind lexer.Kind) lexer.Token {
	if p.peekKind() != kind {
		panic(parsePanic{diag: p.errorAtCurrent("expected %s, found %s", kind, p.peekKind())})
	}
	return p.advance()
}

func (p *Parser) errorAtCurrent(format string, args ...interface{}) error {
	return p.errorAt(p.current(), format, args...)
}

func (p *Parser) errorAt(tok lexer.Token, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	lineText := ""
	if tok.Span.Start.Line-1 >= 0 && tok.Span.Start.Line-1 < len(p.lines) {
		lineText = p.lines[tok.Span.Start.Line-1]
	}

	// Provide better hints based on the error message
	hint := p.generateHint(tok, message)

	return lexer.Diagnostic{
		File:     p.filename,
		Message:  message,
		Hint:     hint,
		Span:     tok.Span,
		Line:     lineText,
		Severity: lexer.Error,
		Category: "syntax",
	}
}

// generateHint generates contextual hints based on the error message and token
func (p *Parser) generateHint(tok lexer.Token, message string) string {
	// Common syntax error patterns and their suggestions
	switch {
	case strings.Contains(message, "expected LBRACE"):
		return "add opening brace '{' to start a block"
	case strings.Contains(message, "expected RBRACE"):
		return "add closing brace '}' to end a block"
	case strings.Contains(message, "expected LPAREN"):
		return "add opening parenthesis '(' to start a function call or expression"
	case strings.Contains(message, "expected RPAREN"):
		return "add closing parenthesis ')' to end a function call or expression"
	case strings.Contains(message, "expected SEMICOLON"):
		return "add semicolon ';' to end the statement"
	case strings.Contains(message, "expected COLON"):
		return "add colon ':' for type annotation or after 'if' condition"
	case strings.Contains(message, "expected IDENTIFIER"):
		return "provide a valid identifier name"
	case strings.Contains(message, "unexpected token"):
		return "check the syntax around this token - it might be in the wrong place"
	case strings.Contains(message, "unexpected token at top level"):
		return "this statement must be inside a function or block"
	case strings.Contains(message, "expected LET"):
		return "use 'let' to declare a variable"
	case strings.Contains(message, "expected VAR"):
		return "use 'var' to declare a mutable variable"
	case strings.Contains(message, "expected FUNC"):
		return "use 'func' to declare a function"
	case strings.Contains(message, "expected IF"):
		return "use 'if' for conditional statements"
	case strings.Contains(message, "expected WHILE"):
		return "use 'while' for loops"
	case strings.Contains(message, "expected FOR"):
		return "use 'for' for loops"
	case strings.Contains(message, "expected RETURN"):
		return "use 'return' to return a value from a function"
	case strings.Contains(message, "expected IMPORT"):
		return "use 'import' to import modules"
	default:
		return "check syntax near this token"
	}
}

// parseQualifiedIdentifier parses an identifier and any following qualified parts.
func (p *Parser) parseQualifiedIdentifier(tok lexer.Token) (ast.Expr, error) {
	name := tok.Lexeme
	span := tok.Span

	// Check for qualified access (e.g., std.io.println)
	for p.peekKind() == lexer.TokenDot {
		p.advance() // consume the dot
		nextTok := p.advance()
		if nextTok.Kind != lexer.TokenIdentifier {
			return nil, p.errorAtCurrent("expected identifier after dot")
		}
		name += "." + nextTok.Lexeme
		span = lexer.Span{Start: span.Start, End: nextTok.Span.End}
	}

	return &ast.IdentifierExpr{SpanInfo: span, Name: name}, nil
}

func (p *Parser) parseNewExpr(tok lexer.Token) (ast.Expr, error) {
	// Parse: new Type or new (Type)
	var typ *ast.TypeExpr
	var err error

	if p.match(lexer.TokenLParen) {
		// Parse new (Type) - parentheses around type
		typ, err = p.parseTypeExpr()
		if err != nil {
			return nil, err
		}
		p.expect(lexer.TokenRParen)
	} else {
		// Parse new Type - direct type
		typ, err = p.parseTypeExpr()
		if err != nil {
			return nil, err
		}
	}

	span := lexer.Span{Start: tok.Span.Start, End: typ.SpanInfo.End}
	return &ast.NewExpr{SpanInfo: span, Type: typ}, nil
}

func (p *Parser) parseDeleteExpr(tok lexer.Token) (ast.Expr, error) {
	// Parse: delete expression
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	span := lexer.Span{Start: tok.Span.Start, End: expr.Span().End}
	return &ast.DeleteExpr{SpanInfo: span, Target: expr}, nil
}

func splitLines(input string) []string {
	normalized := strings.ReplaceAll(input, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	return strings.Split(normalized, "\n")
}
