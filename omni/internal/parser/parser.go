package parser

import (
	"errors"
	"fmt"
	"strings"

	"github.com/omni-lang/omni/internal/ast"
	"github.com/omni-lang/omni/internal/lexer"
)

// transformTokensForNestedGenerics converts >> tokens to two > tokens in generic contexts
// Only treats < as generic delimiter when it follows an identifier in a type context
// (i.e., when the next token is likely a type argument, not a comparison)
func transformTokensForNestedGenerics(tokens []lexer.Token) []lexer.Token {
	var result []lexer.Token
	genericDepth := 0

	for i, token := range tokens {
		// Only treat < as generic delimiter when it follows an identifier AND
		// the next token suggests a type context (identifier, keyword, or type-related token)
		if token.Kind == lexer.TokenLess {
			// Check if previous token is an identifier
			if i > 0 {
				prevToken := tokens[i-1]
				if prevToken.Kind == lexer.TokenIdentifier {
					// Look ahead to see if this is likely a type context
					// If next token is identifier, keyword, or we're in a type-like position, it's generic
					if i+1 < len(tokens) {
						nextToken := tokens[i+1]
						// Type contexts: identifier (type name), keywords, or we're already in generics
						if nextToken.Kind == lexer.TokenIdentifier || 
						   nextToken.Kind == lexer.TokenLess || // nested generic
						   genericDepth > 0 { // already in generic context
							genericDepth++
						} else {
							// Not a generic delimiter (likely comparison), just pass through
							result = append(result, token)
							continue
						}
					} else {
						// End of tokens, not a generic
						result = append(result, token)
						continue
					}
				} else {
					// Not following identifier, not a generic delimiter
					result = append(result, token)
					continue
				}
			} else {
				// First token can't be a generic delimiter
				result = append(result, token)
				continue
			}
		} else if token.Kind == lexer.TokenGreater {
			if genericDepth > 0 {
				genericDepth--
			}
		}

		// Transform >> to two > tokens when in generic context
		if token.Kind == lexer.TokenRShift && genericDepth > 0 {
			// Create two separate > tokens with proper spans
			// First token: use the start position, advance column by 1 for non-zero width
			firstEnd := token.Span.Start
			firstEnd.Column++
			firstGreater := lexer.Token{
				Kind:   lexer.TokenGreater,
				Lexeme: ">",
				Span:   lexer.Span{Start: token.Span.Start, End: firstEnd},
			}
			// Second token: start where first ended, use original end
			secondStart := firstEnd
			secondGreater := lexer.Token{
				Kind:   lexer.TokenGreater,
				Lexeme: ">",
				Span:   lexer.Span{Start: secondStart, End: token.Span.End},
			}
			result = append(result, firstGreater, secondGreater)
		} else {
			result = append(result, token)
		}
	}

	return result
}

// Parse consumes the provided source and returns an abstract syntax tree.
func Parse(filename, input string) (*ast.Module, error) {
	tokens, err := lexer.LexAll(filename, input)
	if err != nil {
		return nil, err
	}
	// Transform >> tokens to two > tokens in generic contexts
	transformedTokens := transformTokensForNestedGenerics(tokens)
	p := &Parser{
		filename: filename,
		tokens:   transformedTokens,
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
	case lexer.TokenType:
		return p.parseTypeAliasDecl()
	case lexer.TokenAsync, lexer.TokenFunc:
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
	isAsync := false
	if p.match(lexer.TokenAsync) {
		isAsync = true
	}
	kw := p.expect(lexer.TokenFunc)
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
		return &ast.FuncDecl{SpanInfo: span, Name: nameTok.Lexeme, TypeParams: typeParams, Params: params, Return: retType, ExprBody: expr, IsAsync: isAsync}, nil
	}
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	span := lexer.Span{Start: kw.Span.Start, End: body.Span().End}
	return &ast.FuncDecl{SpanInfo: span, Name: nameTok.Lexeme, TypeParams: typeParams, Params: params, Return: retType, Body: body, IsAsync: isAsync}, nil
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
	case lexer.TokenWhile:
		return p.parseWhileStmt()
	case lexer.TokenBreak:
		return p.parseBreakStmt()
	case lexer.TokenContinue:
		return p.parseContinueStmt()
	case lexer.TokenLet:
		return p.parseBindingStmt(false)
	case lexer.TokenVar:
		return p.parseBindingStmt(true)
	case lexer.TokenLBrace:
		return p.parseBlock()
	case lexer.TokenTry:
		return p.parseTryStmt()
	case lexer.TokenThrow:
		return p.parseThrowStmt()
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

func (p *Parser) parseWhileStmt() (ast.Stmt, error) {
	tok := p.advance()
	cond, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	span := lexer.Span{Start: tok.Span.Start, End: body.Span().End}
	return &ast.WhileStmt{SpanInfo: span, Cond: cond, Body: body}, nil
}

func (p *Parser) parseBreakStmt() (ast.Stmt, error) {
	tok := p.advance()
	return &ast.BreakStmt{SpanInfo: tok.Span}, nil
}

func (p *Parser) parseContinueStmt() (ast.Stmt, error) {
	tok := p.advance()
	return &ast.ContinueStmt{SpanInfo: tok.Span}, nil
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
	// Handle infinite loop: for { ... }
	if p.peekKind() == lexer.TokenLBrace {
		body, err := p.parseBlock()
		if err != nil {
			return nil, err
		}
		span := lexer.Span{Start: tok.Span.Start, End: body.Span().End}
		return &ast.ForStmt{SpanInfo: span, Body: body}, nil
	}

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
	left, err := p.parseBitwiseOr()
	if err != nil {
		return nil, err
	}
	for p.match(lexer.TokenAndAnd) {
		op := p.previous().Lexeme
		right, err := p.parseBitwiseOr()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{SpanInfo: lexer.Span{Start: left.Span().Start, End: right.Span().End}, Left: left, Op: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parseBitwiseOr() (ast.Expr, error) {
	left, err := p.parseBitwiseXor()
	if err != nil {
		return nil, err
	}
	for p.match(lexer.TokenPipe) {
		op := p.previous().Lexeme
		right, err := p.parseBitwiseXor()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{SpanInfo: lexer.Span{Start: left.Span().Start, End: right.Span().End}, Left: left, Op: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parseBitwiseXor() (ast.Expr, error) {
	left, err := p.parseBitwiseAnd()
	if err != nil {
		return nil, err
	}
	for p.match(lexer.TokenCaret) {
		op := p.previous().Lexeme
		right, err := p.parseBitwiseAnd()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{SpanInfo: lexer.Span{Start: left.Span().Start, End: right.Span().End}, Left: left, Op: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parseBitwiseAnd() (ast.Expr, error) {
	left, err := p.parseEquality()
	if err != nil {
		return nil, err
	}
	for p.match(lexer.TokenAmpersand) {
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
	left, err := p.parseShift()
	if err != nil {
		return nil, err
	}
	for {
		switch p.peekKind() {
		case lexer.TokenLess, lexer.TokenLessEqual, lexer.TokenGreater, lexer.TokenGreaterEqual:
			op := p.advance()
			right, err := p.parseShift()
			if err != nil {
				return nil, err
			}
			left = &ast.BinaryExpr{SpanInfo: lexer.Span{Start: left.Span().Start, End: right.Span().End}, Left: left, Op: op.Lexeme, Right: right}
		default:
			return left, nil
		}
	}
}

func (p *Parser) parseShift() (ast.Expr, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	for {
		switch p.peekKind() {
		case lexer.TokenLShift, lexer.TokenRShift:
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
	case lexer.TokenAwait:
		awaitTok := p.advance()
		expr, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &ast.AwaitExpr{SpanInfo: lexer.Span{Start: awaitTok.Span.Start, End: expr.Span().End}, Expr: expr}, nil
	case lexer.TokenBang, lexer.TokenMinus, lexer.TokenTilde:
		op := p.advance()
		expr, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpr{SpanInfo: lexer.Span{Start: op.Span.Start, End: expr.Span().End}, Op: op.Lexeme, Expr: expr}, nil
	case lexer.TokenLParen:
		// Check if this is a type cast: (type) expression
		startPos := p.current().Span.Start
		savedPos := p.pos
		p.advance() // consume '('

		// Try to parse as a type expression
		typeExpr, err := p.parseTypeExpr()
		if err != nil {
			// Not a type cast, fall back to regular expression parsing
			p.pos = savedPos // restore position
			return p.parsePostfix()
		}

		// Check if next token is ')'
		if !p.match(lexer.TokenRParen) {
			// Not a type cast, fall back to regular expression parsing
			p.pos = savedPos // restore position
			return p.parsePostfix()
		}

		// Parse the expression to be cast
		expr, err := p.parseUnary()
		if err != nil {
			return nil, err
		}

		return &ast.CastExpr{
			SpanInfo: lexer.Span{Start: startPos, End: expr.Span().End},
			Type:     typeExpr,
			Expr:     expr,
		}, nil
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
	case lexer.TokenStringInterpolation:
		return p.parseStringInterpolation(tok)
	case lexer.TokenCharLiteral:
		return &ast.LiteralExpr{SpanInfo: tok.Span, Kind: ast.LiteralChar, Value: tok.Lexeme}, nil
	case lexer.TokenTrue, lexer.TokenFalse:
		return &ast.LiteralExpr{SpanInfo: tok.Span, Kind: ast.LiteralBool, Value: tok.Lexeme}, nil
	case lexer.TokenNullLiteral:
		return &ast.LiteralExpr{SpanInfo: tok.Span, Kind: ast.LiteralNull, Value: tok.Lexeme}, nil
	case lexer.TokenHexLiteral:
		return &ast.LiteralExpr{SpanInfo: tok.Span, Kind: ast.LiteralHex, Value: tok.Lexeme}, nil
	case lexer.TokenBinaryLiteral:
		return &ast.LiteralExpr{SpanInfo: tok.Span, Kind: ast.LiteralBinary, Value: tok.Lexeme}, nil
	case lexer.TokenLParen:
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		p.expect(lexer.TokenRParen)
		return expr, nil
	case lexer.TokenPipe:
		// Lambda expression: |a, b| a + b
		return p.parseLambda(tok)
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
	var typeName string
	var startSpan lexer.Span
	
	// Handle IdentifierExpr and MemberExpr (including multi-level qualified types like pkg.sub.Type)
	switch expr := base.(type) {
	case *ast.IdentifierExpr:
		typeName = expr.Name
		startSpan = expr.SpanInfo
	case *ast.MemberExpr:
		// Walk the chain of MemberExpr targets to build the full qualified name
		parts := []string{expr.Member}
		current := expr.Target
		found := false
		
		for !found {
			switch target := current.(type) {
			case *ast.IdentifierExpr:
				parts = append(parts, target.Name)
				// Reverse the parts since we built them backwards
				for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
					parts[i], parts[j] = parts[j], parts[i]
				}
				typeName = strings.Join(parts, ".")
				startSpan = expr.SpanInfo
				found = true
			case *ast.MemberExpr:
				parts = append(parts, target.Member)
				current = target.Target
			default:
				return nil, p.errorAtCurrent("struct literal must start with type identifier or qualified type")
			}
		}
	default:
		return nil, p.errorAtCurrent("struct literal must start with type identifier or qualified type")
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
	span := lexer.Span{Start: startSpan.Start, End: rbrace.Span.End}
	return &ast.StructLiteralExpr{SpanInfo: span, TypeName: typeName, Fields: fields}, nil
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
	// Accept IdentifierExpr and MemberExpr (including multi-level qualified types)
	_, isIdent := base.(*ast.IdentifierExpr)
	_, isMember := base.(*ast.MemberExpr)
	if !isIdent && !isMember {
		return false
	}
	// For MemberExpr, recursively check if the target is also valid
	if isMember {
		memberExpr := base.(*ast.MemberExpr)
		// The target should also be an identifier or member expression
		_, targetIsIdent := memberExpr.Target.(*ast.IdentifierExpr)
		_, targetIsMember := memberExpr.Target.(*ast.MemberExpr)
		if !targetIsIdent && !targetIsMember {
			return false
		}
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

// parseGenericTypeArgs parses generic type arguments with proper handling of nested generics
func (p *Parser) parseGenericTypeArgs() ([]*ast.TypeExpr, error) {
	args := []*ast.TypeExpr{}
	if !p.match(lexer.TokenGreater) {
		for {
			arg, err := p.parseTypeExprWithNestedGenerics()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
			if p.match(lexer.TokenComma) {
				continue
			}
			// Handle nested generics: array<array<int>> should parse as array<array<int>>
			// Check if we have >> (right shift) and convert to two separate > tokens
			if p.peekKind() == lexer.TokenRShift {
				// We have >> which should be parsed as two separate > tokens
				// The first > closes the inner generic, the second > closes the outer generic
				p.advance() // consume the RShift token
				return args, nil
			}
			p.expect(lexer.TokenGreater)
			return args, nil
		}
	}
	return args, nil
}

// parseTypeExprWithNestedGenerics is a specialized version that handles >> in generic contexts
func (p *Parser) parseTypeExprWithNestedGenerics() (*ast.TypeExpr, error) {
	// Parse the first type
	firstType, err := p.parseSingleTypeWithNestedGenerics()
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
			member, err := p.parseSingleTypeWithNestedGenerics()
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

// parseSingleTypeWithNestedGenerics handles >> tokens in generic contexts
func (p *Parser) parseSingleTypeWithNestedGenerics() (*ast.TypeExpr, error) {
	// Handle array types: []Type
	if p.match(lexer.TokenLBracket) {
		start := p.previous().Span.Start
		p.expect(lexer.TokenRBracket)
		elementType, err := p.parseSingleTypeWithNestedGenerics()
		if err != nil {
			return nil, err
		}
		span := lexer.Span{Start: start, End: elementType.SpanInfo.End}
		return &ast.TypeExpr{SpanInfo: span, Name: "[]", Args: []*ast.TypeExpr{elementType}}, nil
	}

	// Handle pointer types: *Type or *(Type)
	if p.match(lexer.TokenStar) {
		starToken := p.previous() // Capture the * token for correct span
		var baseType *ast.TypeExpr
		var err error

		if p.match(lexer.TokenLParen) {
			// Parse *(Type) - parentheses around type
			baseType, err = p.parseTypeExprWithNestedGenerics()
			if err != nil {
				return nil, err
			}
			p.expect(lexer.TokenRParen)
		} else {
			// Parse *Type - direct type
			baseType, err = p.parseSingleTypeWithNestedGenerics()
			if err != nil {
				return nil, err
			}
		}

		// Span should start at the * token, not the base type
		span := lexer.Span{Start: starToken.Span.Start, End: baseType.SpanInfo.End}
		
		// For union types, store the base type in Args
		if baseType.IsUnion {
			return &ast.TypeExpr{SpanInfo: span, Name: "*", Args: baseType.Members}, nil
		}
		
		// Preserve generic arguments and other flags from baseType
		// Create a new TypeExpr that wraps the base type with pointer semantics
		return &ast.TypeExpr{
			SpanInfo:   span,
			Name:       "*" + baseType.Name,
			Args:       baseType.Args,        // Preserve generic arguments
			IsOptional: baseType.IsOptional,  // Preserve optional flag
			OptionalType: baseType.OptionalType, // Preserve optional type if present
		}, nil
	}

	// Handle parentheses around types: (Type) or function types: (int, string) -> bool
	if p.match(lexer.TokenLParen) {
		start := p.previous().Span.Start

		// Try to parse as function type first
		paramTypes := []*ast.TypeExpr{}

		// Parse parameter types
		if !p.match(lexer.TokenRParen) {
			for {
				paramType, err := p.parseSingleTypeWithNestedGenerics()
				if err != nil {
					return nil, err
				}
				paramTypes = append(paramTypes, paramType)

				if p.match(lexer.TokenRParen) {
					break
				}
				if !p.match(lexer.TokenComma) {
					return nil, p.errorAtCurrent("expected ',' or ')' in function type")
				}
			}
		}

		// Check if this is a function type: (params) -> returnType
		if p.match(lexer.TokenArrow) {
			returnType, err := p.parseTypeExprWithNestedGenerics()
			if err != nil {
				return nil, err
			}
			span := lexer.Span{Start: start, End: returnType.SpanInfo.End}
			return &ast.TypeExpr{
				SpanInfo:   span,
				IsFunction: true,
				ParamTypes: paramTypes,
				ReturnType: returnType,
			}, nil
		}

		// If no -> found, this is just a parenthesized type
		if len(paramTypes) == 1 {
			return paramTypes[0], nil
		}

		// Multiple types without -> is an error
		return nil, p.errorAtCurrent("multiple types in parentheses must be followed by '->' for function type")
	}

	// Handle simple identifier types and qualified types (e.g., math.Point, pkg.sub.Type)
	nameTok := p.expect(lexer.TokenIdentifier)
	typeName := nameTok.Lexeme
	startSpan := nameTok.Span
	
	// Consume additional .-separated identifiers for qualified types
	for p.match(lexer.TokenDot) {
		nextIdent := p.expect(lexer.TokenIdentifier)
		typeName += "." + nextIdent.Lexeme
	}
	
	typeExpr := &ast.TypeExpr{SpanInfo: lexer.Span{Start: startSpan.Start, End: p.previous().Span.End}, Name: typeName}
	if p.match(lexer.TokenLess) {
		args, err := p.parseGenericTypeArgs()
		if err != nil {
			return nil, err
		}
		typeExpr.Args = args
		typeExpr.SpanInfo.End = p.previous().Span.End
	}

	// Check for optional type syntax: T?
	if p.match(lexer.TokenQuestion) {
		span := lexer.Span{Start: typeExpr.SpanInfo.Start, End: p.previous().Span.End}
		return &ast.TypeExpr{
			SpanInfo:     span,
			IsOptional:   true,
			OptionalType: typeExpr,
		}, nil
	}

	return typeExpr, nil
}

// parseSingleType parses a single type (not a union)
// This function now handles nested generics properly
func (p *Parser) parseSingleType() (*ast.TypeExpr, error) {
	// Handle array types: []Type
	if p.match(lexer.TokenLBracket) {
		start := p.previous().Span.Start
		p.expect(lexer.TokenRBracket)
		elementType, err := p.parseSingleType()
		if err != nil {
			return nil, err
		}
		span := lexer.Span{Start: start, End: elementType.SpanInfo.End}
		return &ast.TypeExpr{SpanInfo: span, Name: "[]", Args: []*ast.TypeExpr{elementType}}, nil
	}

	// Handle pointer types: *Type or *(Type)
	if p.match(lexer.TokenStar) {
		starToken := p.previous() // Capture the * token for correct span
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

		// Span should start at the * token, not the base type
		span := lexer.Span{Start: starToken.Span.Start, End: baseType.SpanInfo.End}
		
		// For union types, store the base type in Args
		if baseType.IsUnion {
			return &ast.TypeExpr{SpanInfo: span, Name: "*", Args: baseType.Members}, nil
		}
		
		// Preserve generic arguments and other flags from baseType
		return &ast.TypeExpr{
			SpanInfo:   span,
			Name:       "*" + baseType.Name,
			Args:       baseType.Args,        // Preserve generic arguments
			IsOptional: baseType.IsOptional,  // Preserve optional flag
			OptionalType: baseType.OptionalType, // Preserve optional type if present
		}, nil
	}

	// Handle parentheses around types: (Type) or function types: (int, string) -> bool
	if p.match(lexer.TokenLParen) {
		start := p.previous().Span.Start

		// Try to parse as function type first
		paramTypes := []*ast.TypeExpr{}

		// Parse parameter types
		if !p.match(lexer.TokenRParen) {
			for {
				paramType, err := p.parseSingleType()
				if err != nil {
					return nil, err
				}
				paramTypes = append(paramTypes, paramType)

				if p.match(lexer.TokenRParen) {
					break
				}
				if !p.match(lexer.TokenComma) {
					return nil, p.errorAtCurrent("expected ',' or ')' in function type")
				}
			}
		}

		// Check if this is a function type (has -> after )
		if p.match(lexer.TokenArrow) {
			// This is a function type: (params) -> returnType
			returnType, err := p.parseTypeExpr()
			if err != nil {
				return nil, err
			}
			span := lexer.Span{Start: start, End: returnType.SpanInfo.End}
			return &ast.TypeExpr{
				SpanInfo:   span,
				IsFunction: true,
				ParamTypes: paramTypes,
				ReturnType: returnType,
			}, nil
		}

		// If no -> found, this is just a parenthesized type
		if len(paramTypes) == 1 {
			return paramTypes[0], nil
		}

		// Multiple types without -> is an error
		return nil, p.errorAtCurrent("multiple types in parentheses must be followed by '->' for function type")
	}

	// Handle simple identifier types and qualified types (e.g., math.Point, pkg.sub.Type)
	nameTok := p.expect(lexer.TokenIdentifier)
	typeName := nameTok.Lexeme
	startSpan := nameTok.Span
	
	// Consume additional .-separated identifiers for qualified types
	for p.match(lexer.TokenDot) {
		nextIdent := p.expect(lexer.TokenIdentifier)
		typeName += "." + nextIdent.Lexeme
	}
	
	typeExpr := &ast.TypeExpr{SpanInfo: lexer.Span{Start: startSpan.Start, End: p.previous().Span.End}, Name: typeName}
	if p.match(lexer.TokenLess) {
		args, err := p.parseGenericTypeArgs()
		if err != nil {
			return nil, err
		}
		typeExpr.Args = args
		typeExpr.SpanInfo.End = p.previous().Span.End
	}

	// Check for optional type syntax: T?
	if p.match(lexer.TokenQuestion) {
		span := lexer.Span{Start: typeExpr.SpanInfo.Start, End: p.previous().Span.End}
		return &ast.TypeExpr{
			SpanInfo:     span,
			IsOptional:   true,
			OptionalType: typeExpr,
		}, nil
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
	contextLines, contextStart := lexer.BuildContext(p.lines, tok.Span)

	// Provide better hints based on the error message
	hint := p.generateHint(tok, message)

	return lexer.Diagnostic{
		File:             p.filename,
		Message:          message,
		Hint:             hint,
		Span:             tok.Span,
		Line:             lineText,
		Context:          contextLines,
		ContextStartLine: contextStart,
		Severity:         lexer.Error,
		Category:         "syntax",
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

func (p *Parser) parseLambda(startTok lexer.Token) (ast.Expr, error) {
	// Parse lambda parameters: |a, b|
	params := []ast.Param{}

	// Parse parameters until we hit the closing |
	for {
		if p.match(lexer.TokenPipe) {
			// End of parameters
			break
		}

		// Parse parameter name
		paramName := p.expect(lexer.TokenIdentifier)

		// For now, we'll infer parameter types
		// In a full implementation, we might support type annotations: |a: int, b: int|
		param := ast.Param{
			Name: paramName.Lexeme,
			Type: nil, // Type will be inferred
			Span: paramName.Span,
		}
		params = append(params, param)

		// Check for comma separator
		if p.match(lexer.TokenComma) {
			continue
		}

		// If we don't see a comma, we should see the closing |
		if p.peekKind() != lexer.TokenPipe {
			return nil, p.errorAtCurrent("expected ',' or '|' in lambda parameters")
		}
	}

	// Parse lambda body
	body, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	span := lexer.Span{Start: startTok.Span.Start, End: body.Span().End}
	return &ast.LambdaExpr{SpanInfo: span, Params: params, Body: body}, nil
}

// parseStringInterpolation parses a string interpolation token like "Hello, ${name}!"
func (p *Parser) parseStringInterpolation(tok lexer.Token) (ast.Expr, error) {
	// The token lexeme contains the full string including quotes
	// We need to parse the contents to extract literal parts and expressions
	content := tok.Lexeme
	if len(content) < 2 || content[0] != '"' || content[len(content)-1] != '"' {
		return nil, fmt.Errorf("invalid string interpolation token: %s", content)
	}

	// Remove the outer quotes
	content = content[1 : len(content)-1]

	var parts []ast.StringInterpolationPart
	current := ""

	for i := 0; i < len(content); i++ {
		if content[i] == '$' && i+1 < len(content) && content[i+1] == '{' {
			// Found an interpolation expression
			// Add current literal part if not empty
			if current != "" {
				parts = append(parts, ast.StringInterpolationPart{
					IsLiteral: true,
					Literal:   current,
					Span:      lexer.Span{Start: tok.Span.Start, End: tok.Span.End}, // TODO: calculate precise span
				})
				current = ""
			}

			// Find the closing brace
			i += 2 // skip "${"
			exprStart := i
			braceCount := 1
			for i < len(content) && braceCount > 0 {
				if content[i] == '{' {
					braceCount++
				} else if content[i] == '}' {
					braceCount--
				}
				i++
			}

			if braceCount > 0 {
				return nil, fmt.Errorf("unterminated interpolation expression in string")
			}

			// Parse the expression inside the braces
			exprContent := content[exprStart : i-1] // exclude the closing brace
			// Create a temporary parser to parse just this expression
			// Apply the same token transform as the main parser
			rawTokens, err := lexer.LexAll("", exprContent)
			if err != nil {
				return nil, fmt.Errorf("failed to parse interpolation expression: %v", err)
			}
			// Apply the same transformTokensForNestedGenerics transform
			tempTokens := transformTokensForNestedGenerics(rawTokens)
			tempParser := &Parser{
				filename: "interpolation",
				tokens:   tempTokens,
				lines:    []string{exprContent},
			}

			expr, err := tempParser.parseExpr()
			if err != nil {
				return nil, fmt.Errorf("failed to parse interpolation expression: %v", err)
			}
			// Verify that we consumed all tokens (enforce EOF)
			if tempParser.pos < len(tempParser.tokens) {
				return nil, fmt.Errorf("interpolation expression has trailing tokens")
			}

			parts = append(parts, ast.StringInterpolationPart{
				IsLiteral: false,
				Expr:      expr,
				Span:      lexer.Span{Start: tok.Span.Start, End: tok.Span.End}, // TODO: calculate precise span
			})
		} else {
			current += string(content[i])
		}
	}

	// Add remaining literal part if any
	if current != "" {
		parts = append(parts, ast.StringInterpolationPart{
			IsLiteral: true,
			Literal:   current,
			Span:      lexer.Span{Start: tok.Span.Start, End: tok.Span.End}, // TODO: calculate precise span
		})
	}

	return &ast.StringInterpolationExpr{
		SpanInfo: tok.Span,
		Parts:    parts,
	}, nil
}

// parseTryStmt parses a try-catch-finally statement
func (p *Parser) parseTryStmt() (ast.Stmt, error) {
	startPos := p.expect(lexer.TokenTry).Span.Start

	// Parse try block
	tryBlock, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	var catchClauses []*ast.CatchClause

	// Parse catch clauses
	for p.match(lexer.TokenCatch) {
		catchClause := &ast.CatchClause{SpanInfo: p.previous().Span}

		// Parse optional exception variable and type
		if p.match(lexer.TokenLParen) {
			if p.peekKind() == lexer.TokenIdentifier {
				catchClause.ExceptionVar = p.advance().Lexeme

				// Parse optional exception type
				if p.match(lexer.TokenColon) {
					if p.peekKind() == lexer.TokenIdentifier {
						catchClause.ExceptionType = p.advance().Lexeme
					} else {
						return nil, p.errorAtCurrent("expected exception type after colon")
					}
				}
			}

			if !p.match(lexer.TokenRParen) {
				return nil, p.errorAtCurrent("expected closing parenthesis in catch clause")
			}
		}

		// Parse catch block
		catchBlock, err := p.parseBlock()
		if err != nil {
			return nil, err
		}
		catchClause.Block = catchBlock

		catchClauses = append(catchClauses, catchClause)
	}

	// Parse optional finally block
	var finallyBlock *ast.BlockStmt
	if p.match(lexer.TokenFinally) {
		var err error
		finallyBlock, err = p.parseBlock()
		if err != nil {
			return nil, err
		}
	}

	if len(catchClauses) == 0 && finallyBlock == nil {
		return nil, p.errorAtCurrent("try statement must have at least one catch or finally block")
	}

	// Calculate end position
	endPos := tryBlock.Span().End
	if len(catchClauses) > 0 {
		endPos = catchClauses[len(catchClauses)-1].Block.Span().End
	}
	if finallyBlock != nil {
		endPos = finallyBlock.Span().End
	}

	return &ast.TryStmt{
		SpanInfo:     lexer.Span{Start: startPos, End: endPos},
		TryBlock:     tryBlock,
		CatchClauses: catchClauses,
		FinallyBlock: finallyBlock,
	}, nil
}

// parseThrowStmt parses a throw statement
func (p *Parser) parseThrowStmt() (ast.Stmt, error) {
	startPos := p.expect(lexer.TokenThrow).Span.Start

	// Parse the expression to throw
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	return &ast.ThrowStmt{
		SpanInfo: lexer.Span{Start: startPos, End: expr.Span().End},
		Expr:     expr,
	}, nil
}

// parseTypeAliasDecl parses a type alias declaration: type UserID = int
func (p *Parser) parseTypeAliasDecl() (ast.Decl, error) {
	startPos := p.expect(lexer.TokenType).Span.Start

	// Parse the type name
	nameTok := p.expect(lexer.TokenIdentifier)
	name := nameTok.Lexeme

	// Parse optional generic type parameters: type Container<T, U> = array<T>
	var typeParams []string
	if p.match(lexer.TokenLess) {
		for {
			if p.peekKind() == lexer.TokenIdentifier {
				paramTok := p.advance()
				typeParams = append(typeParams, paramTok.Lexeme)

				if p.match(lexer.TokenComma) {
					continue
				}
			}

			if p.match(lexer.TokenGreater) {
				break
			}

			return nil, p.errorAtCurrent("expected type parameter or closing bracket")
		}
	}

	// Parse the equals sign
	p.expect(lexer.TokenAssign)

	// Parse the type expression (which could be a union type)
	typeExpr, err := p.parseTypeExpr()
	if err != nil {
		return nil, err
	}

	endPos := typeExpr.Span().End

	return &ast.TypeAliasDecl{
		SpanInfo:   lexer.Span{Start: startPos, End: endPos},
		Name:       name,
		TypeParams: typeParams,
		Type:       *typeExpr,
	}, nil
}

func splitLines(input string) []string {
	normalized := strings.ReplaceAll(input, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	return strings.Split(normalized, "\n")
}
