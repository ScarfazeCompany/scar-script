package parser

import (
	"strconv"
	"strings"

	"github.com/ScarfazeCompany/scar-script/internal/ast"
	"github.com/ScarfazeCompany/scar-script/internal/diag"
	"github.com/ScarfazeCompany/scar-script/internal/lexer"
	"github.com/ScarfazeCompany/scar-script/internal/token"
)

type Parser struct {
	toks        []token.Token
	pos         int
	file        string
	depth       int             // () and [] nesting — newlines are skipped while > 0
	structNames map[string]bool // top-level struct names declared in this file
}

type parsePanic struct{ err *diag.Error }

// Parse turns a token stream into a Program. On failure it returns the
// first syntax error.
func Parse(toks []token.Token, file string) (prog *ast.Program, err *diag.Error) {
	p := &Parser{toks: toks, file: file, structNames: map[string]bool{}}
	for i := 0; i+1 < len(toks); i++ {
		if toks[i].Type == token.STRUCT && toks[i+1].Type == token.IDENT {
			p.structNames[toks[i+1].Lit] = true
		}
	}
	defer func() {
		if r := recover(); r != nil {
			if pe, ok := r.(parsePanic); ok {
				prog, err = nil, pe.err
				return
			}
			panic(r)
		}
	}()
	prog = p.parseProgram()
	return prog, nil
}

// ---------- token helpers ----------

func (p *Parser) cur() token.Token  { return p.toks[p.pos] }
func (p *Parser) peek() token.Token { return p.toks[p.pos+1] }

// peekInfo returns the next token after the current position, skipping
// newlines, along with its index.
func (p *Parser) peekInfo() (token.Token, int) {
	i := p.pos + 1
	for i < len(p.toks) && p.toks[i].Type == token.NEWLINE {
		i++
	}
	return p.toks[i], i
}

func (p *Parser) at(t token.Type) bool { return p.cur().Type == t }

func (p *Parser) advance() token.Token {
	t := p.cur()
	if t.Type != token.EOF {
		p.pos++
	}
	return t
}

func (p *Parser) tkPos(t token.Token) diag.Position {
	return diag.Position{File: p.file, Line: t.Line, Col: t.Col}
}

func (p *Parser) posNow() diag.Position { return p.tkPos(p.cur()) }

func (p *Parser) fail(t token.Token, format string, args ...any) {
	panic(parsePanic{diag.New(p.tkPos(t), format, args...)})
}

func (p *Parser) failNow(format string, args ...any) {
	p.fail(p.cur(), format, args...)
}

func (p *Parser) expect(t token.Type) token.Token {
	if !p.at(t) {
		p.failNow("expected %q, got %q", string(t), p.cur().String())
	}
	return p.advance()
}

func (p *Parser) expectIdent() token.Token {
	if !p.at(token.IDENT) {
		p.failNow("expected identifier, got %q", p.cur().String())
	}
	return p.advance()
}

func (p *Parser) skipNewlines() {
	for p.at(token.NEWLINE) {
		p.advance()
	}
}

func (p *Parser) skipNestedNewlines() {
	if p.depth > 0 {
		p.skipNewlines()
	}
}

// ---------- program & statements ----------

func (p *Parser) parseProgram() *ast.Program {
	prog := &ast.Program{File: p.file}
	p.skipNewlines()
	for !p.at(token.EOF) {
		prog.Stmts = append(prog.Stmts, p.parseStmt())
		p.endStmt(token.EOF)
		p.skipNewlines()
	}
	return prog
}

// endStmt requires a statement separator: a newline, or `}` when stopAtBrace.
func (p *Parser) endStmt(stop token.Type) {
	if p.at(token.NEWLINE) || p.at(stop) || p.at(token.EOF) {
		return
	}
	p.failNow("unexpected %q after statement (expected newline)", p.cur().String())
}

func (p *Parser) parseBlock() *ast.BlockStmt {
	open := p.expect(token.LBRACE)
	blk := &ast.BlockStmt{P: p.tkPos(open)}
	p.skipNewlines()
	for !p.at(token.RBRACE) && !p.at(token.EOF) {
		blk.Stmts = append(blk.Stmts, p.parseStmt())
		p.endStmt(token.RBRACE)
		p.skipNewlines()
	}
	if p.at(token.EOF) {
		p.failNow("unclosed %q — expected %q", string(token.LBRACE), string(token.RBRACE))
	}
	p.expect(token.RBRACE)
	return blk
}

func (p *Parser) parseStmt() ast.Stmt {
	switch p.cur().Type {
	case token.VAL, token.MUT, token.CONST:
		return p.parseDecl()
	case token.IF:
		return p.parseIf()
	case token.FOR:
		return p.parseFor()
	case token.FUNC:
		return p.parseFuncDecl()
	case token.STRUCT:
		return p.parseStructDecl()
	case token.EXPORT:
		return p.parseExportDecl()
	case token.IMPORT:
		return p.parseImportStmt()
	case token.RETURN:
		return p.parseReturn()
	case token.CONTINUE, token.BREAK, token.STOP:
		return p.parseLoopControl()
	case token.MATCH:
		return p.parseMatch()
	}
	if p.at(token.IDENT) && p.peek().Type == token.ASSIGN {
		return p.parseAssign()
	}
	if p.at(token.IDENT) && p.peek().Type == token.DOT && p.pos+3 < len(p.toks) && p.toks[p.pos+2].Type == token.IDENT && p.toks[p.pos+3].Type == token.ASSIGN {
		// obj.field = value
		return p.parseAssignField()
	}
	return p.parseExprStmt()
}

func (p *Parser) parseDecl() ast.Stmt {
	kw := p.advance()
	kind := ast.Val
	switch kw.Type {
	case token.MUT:
		kind = ast.Mut
	case token.CONST:
		kind = ast.Const
	}
	name := p.expectIdent()

	// Multi-decl: `val a, b = f()` (tuple-returning call).
	if p.at(token.COMMA) {
		md := &ast.MultiDeclStmt{
			P: p.tkPos(kw), Kind: kind,
			Names: []string{name.Lit}, NamePoss: []diag.Position{p.tkPos(name)},
		}
		for p.at(token.COMMA) {
			p.advance()
			n := p.expectIdent()
			md.Names = append(md.Names, n.Lit)
			md.NamePoss = append(md.NamePoss, p.tkPos(n))
		}
		p.expect(token.ASSIGN)
		md.Value = p.parseExpression()
		return md
	}

	var ty ast.TypeExpr
	if p.at(token.COLON) {
		p.advance()
		ty = p.parseType()
	}
	p.expect(token.ASSIGN)
	val := p.parseExpression()
	return &ast.DeclStmt{
		P: p.tkPos(kw), Kind: kind, Name: name.Lit, NamePos: p.tkPos(name),
		Type: ty, Value: val,
	}
}

// parseFuncDecl parses a top-level `func name(params) [result] { ... }`.
// The parser also accepts it inside blocks; the checker rejects nesting.
func (p *Parser) parseFuncDecl() *ast.FuncDecl {
	kw := p.expect(token.FUNC)
	name := p.expectIdent()
	params := p.parseParams()
	fd := &ast.FuncDecl{P: p.tkPos(kw), Name: name.Lit, NamePos: p.tkPos(name), Params: params}
	fd.Result = p.parseResultType()
	fd.Body = p.parseBlock()
	return fd
}

// parseResultType reads an optional function result type: `int`,
// `(int, bool)`, `-> int`, or nothing (void before `{`).
func (p *Parser) parseResultType() ast.TypeExpr {
	if p.at(token.ARROW) {
		p.advance()
		return p.parseType()
	}
	if p.at(token.LBRACE) {
		return nil
	}
	return p.parseType()
}

func (p *Parser) parseParams() []ast.Param {
	p.expect(token.LPAREN)
	p.depth++
	var params []ast.Param
	p.skipNestedNewlines()
	if p.at(token.RPAREN) {
		p.depth--
		p.advance()
		return nil
	}
	for {
		n := p.expectIdent()
		p.expect(token.COLON)
		ty := p.parseType()
		params = append(params, ast.Param{Name: n.Lit, NamePos: p.tkPos(n), Type: ty})
		p.skipNestedNewlines()
		if p.at(token.COMMA) {
			p.advance()
			p.skipNestedNewlines()
			if p.at(token.RPAREN) {
				break
			}
			continue
		}
		break
	}
	p.depth--
	p.expect(token.RPAREN)
	return params
}

func (p *Parser) parseReturn() *ast.ReturnStmt {
	kw := p.expect(token.RETURN)
	rs := &ast.ReturnStmt{P: p.tkPos(kw)}
	if p.at(token.NEWLINE) || p.at(token.RBRACE) || p.at(token.EOF) {
		return rs
	}
	rs.Values = append(rs.Values, p.parseExpression())
	for p.at(token.COMMA) {
		p.advance()
		rs.Values = append(rs.Values, p.parseExpression())
	}
	return rs
}

// parseLoopControl parses `continue`, `break`, or `stop` statements.
func (p *Parser) parseLoopControl() ast.Stmt {
	kw := p.advance()
	switch kw.Type {
	case token.CONTINUE:
		return &ast.ContinueStmt{P: p.tkPos(kw)}
	default: // BREAK or STOP both map to BreakStmt
		return &ast.BreakStmt{P: p.tkPos(kw)}
	}
}

// parseStructDecl parses `struct Name { field: type, ... }`.
func (p *Parser) parseStructDecl() ast.Stmt {
	kw := p.expect(token.STRUCT)
	name := p.expectIdent()
	sd := &ast.StructDecl{P: p.tkPos(kw), Name: name.Lit, NamePos: p.tkPos(name)}
	p.expect(token.LBRACE)
	p.skipNewlines()
	for !p.at(token.RBRACE) && !p.at(token.EOF) {
		fname := p.expectIdent()
		p.expect(token.COLON)
		ftype := p.parseType()
		sd.Fields = append(sd.Fields, ast.StructField{
			Name:    fname.Lit,
			NamePos: p.tkPos(fname),
			Type:    ftype,
		})
		p.skipNewlines()
		if p.at(token.COMMA) {
			p.advance()
			p.skipNewlines()
			if p.at(token.RBRACE) {
				break
			}
			continue
		}
		if p.at(token.RBRACE) {
			break
		}
		continue
	}
	p.expect(token.RBRACE)
	return sd
}

// parseExportDecl parses `export <declaration>`.
func (p *Parser) parseExportDecl() ast.Stmt {
	kw := p.expect(token.EXPORT)
	// Parse the declaration that follows
	inner := p.parseStmt()
	return &ast.ExportDecl{P: p.tkPos(kw), Decl: inner}
}

// parseImportStmt parses `import { name1, name2 } from "./path.scar"`.
func (p *Parser) parseImportStmt() ast.Stmt {
	kw := p.expect(token.IMPORT)
	p.expect(token.LBRACE)
	var names []string
	var namePoss []diag.Position
	for {
		n := p.expectIdent()
		names = append(names, n.Lit)
		namePoss = append(namePoss, p.tkPos(n))
		p.skipNewlines()
		if p.at(token.COMMA) {
			p.advance()
			p.skipNewlines()
			if p.at(token.RBRACE) {
				break
			}
			continue
		}
		break
	}
	p.expect(token.RBRACE)
	p.expect(token.FROM)
	fromTok := p.expect(token.STRING)
	// Extract the string value from segments
	fromPath := ""
	for _, seg := range fromTok.Segs {
		if !seg.IsExpr {
			fromPath += seg.Text
		}
	}
	return &ast.ImportStmt{
		P:        p.tkPos(kw),
		Names:    names,
		NamePoss: namePoss,
		From:     fromPath,
		FromPos:  p.tkPos(fromTok),
	}
}

func (p *Parser) parseAssign() *ast.AssignStmt {
	name := p.expectIdent()
	p.expect(token.ASSIGN)
	val := p.parseExpression()
	return &ast.AssignStmt{P: p.tkPos(name), Name: name.Lit, NamePos: p.tkPos(name), Value: val}
}

// parseAssignField parses `obj.field = value`.
func (p *Parser) parseAssignField() ast.Stmt {
	base := p.parsePrimary() // parses u
	// base should be an Ident
	id, ok := base.(*ast.Ident)
	if !ok {
		p.failNow("expected identifier for field assignment")
	}
	if !p.at(token.DOT) {
		p.failNow("expected '.' after identifier")
	}
	p.advance() // consume '.'
	field := p.expectIdent()
	p.expect(token.ASSIGN)
	val := p.parseExpression()
	return &ast.AssignFieldStmt{
		P:        id.P,
		X:        id,
		Field:    field.Lit,
		FieldPos: p.tkPos(field),
		Value:    val,
	}
}

func (p *Parser) parseExprStmt() ast.Stmt {
	start := p.posNow()
	expr := p.parseExpression()
	if p.at(token.PLUSPLUS) || p.at(token.MINUSMINUS) {
		op := p.advance()
		id, ok := expr.(*ast.Ident)
		if !ok {
			p.fail(op, "cannot apply %q to this expression", op.Lit)
		}
		return &ast.IncDecStmt{P: start, Name: id.Name, NamePos: id.P, Op: op.Type}
	}
	return &ast.ExprStmt{P: start, X: expr}
}

func (p *Parser) parseIf() *ast.IfStmt {
	kw := p.expect(token.IF)
	cond := p.parseExpression()
	then := p.parseBlock()
	stmt := &ast.IfStmt{P: p.tkPos(kw), Cond: cond, Then: then}
	if p.at(token.ELSE) {
		p.advance()
		if p.at(token.IF) {
			stmt.Else = p.parseIf()
		} else {
			stmt.Else = p.parseBlock()
		}
	}
	return stmt
}

func (p *Parser) parseMatch() *ast.MatchStmt {
	kw := p.expect(token.MATCH)
	expr := p.parseExpression()
	arms := p.parseMatchArms()
	return &ast.MatchStmt{P: p.tkPos(kw), Expr: expr, Arms: arms}
}

// parseMatchArms parses match arms: pattern => expr/block, comma-free, newlines allowed
func (p *Parser) parseMatchArms() []ast.MatchArm {
	var arms []ast.MatchArm
	for {
		pattern := p.expectIdent().Lit
		p.expect(token.TRUE) // consume => - we use TRUE as placeholder, actually need -> but let's use a different approach
		// Actually, let me use a different approach - just expect the arrow token
		// Wait, ScarScript uses => not ->
		p.expect(token.COLON) // consume =>
		// Check if next is { (block) or expression
		if p.at(token.LBRACE) {
			body := p.parseBlock()
			arms = append(arms, ast.MatchArm{Pattern: pattern, Body: body})
		} else {
			body := p.parseExpression()
			arms = append(arms, ast.MatchArm{Pattern: pattern, Body: body})
		}
		p.skipNewlines()
		if p.at(token.RBRACE) || p.at(token.EOF) {
			break
		}
	}
	return arms
}

func (p *Parser) parseFor() *ast.ForStmt {
	kw := p.expect(token.FOR)
	fs := &ast.ForStmt{P: p.tkPos(kw)}

	if p.at(token.IDENT) && (p.peek().Type == token.IN || p.peek().Type == token.COMMA) {
		v1 := p.expectIdent()
		fs.Var, fs.VarPos = v1.Lit, p.tkPos(v1)
		if p.at(token.COMMA) {
			p.advance()
			v2 := p.expectIdent()
			fs.Var2, fs.Var2Pos = v2.Lit, p.tkPos(v2)
			fs.Kind = ast.ForTwo
		} else {
			fs.Kind = ast.ForIterable
		}
		p.expect(token.IN)
	} else {
		fs.Kind = ast.ForCond
	}
	fs.Iterable = p.parseExpression()
	fs.Body = p.parseBlock()
	return fs
}

// ---------- types ----------

func (p *Parser) parseType() ast.TypeExpr {
	start := p.cur()
	var base ast.TypeExpr

	switch start.Type {
	case token.T_STR, token.T_INT, token.T_BOOL, token.T_ANY, token.T_DEC:
		p.advance()
		base = &ast.BasicType{P: p.tkPos(start), Kind: start.Type}
	case token.OPTION:
		p.advance()
		p.expect(token.LBRACKET)
		elem := p.parseType()
		p.expect(token.RBRACKET)
		base = &ast.OptionTypeExpr{P: p.tkPos(start), Elem: elem}
	case token.RESULT:
		p.advance()
		p.expect(token.LBRACKET)
		ok := p.parseType()
		p.expect(token.COMMA)
		err := p.parseType()
		p.expect(token.RBRACKET)
		base = &ast.ResultTypeExpr{P: p.tkPos(start), Ok: ok, Err: err}
	case token.IDENT:
		name := start.Lit
		switch strings.ToLower(name) {
		case "option":
			p.advance()
			if p.at(token.LBRACKET) {
				p.advance()
			} else {
				p.expect(token.LT)
			}
			elem := p.parseType()
			if p.at(token.LBRACKET) {
				p.advance()
			} else {
				p.expect(token.GT)
			}
			base = &ast.OptionTypeExpr{P: p.tkPos(start), Elem: elem}
		case "result":
			p.advance()
			if p.at(token.LBRACKET) {
				p.advance()
			} else {
				p.expect(token.LT)
			}
			okT := p.parseType()
			p.expect(token.COMMA)
			errT := p.parseType()
			if p.at(token.LBRACKET) {
				p.advance()
			} else {
				p.expect(token.GT)
			}
			base = &ast.ResultTypeExpr{P: p.tkPos(start), Ok: okT, Err: errT}
		default:
			if p.structNames[name] {
				p.advance()
				base = &ast.StructTypeExpr{P: p.tkPos(start), Name: name}
			} else {
				p.failNow("expected type, got %q", name)
			}
		}
	case token.LPAREN:
		p.advance()
		p.depth++
		tup := &ast.TupleTypeExpr{P: p.tkPos(start)}
		for {
			tup.Elems = append(tup.Elems, p.parseType())
			if p.at(token.COMMA) {
				p.advance()
				continue
			}
			break
		}
		p.depth--
		p.expect(token.RPAREN)
		if len(tup.Elems) < 2 {
			p.fail(start, "tuple type must have at least 2 elements")
		}
		base = tup
	case token.MAP:
		p.advance()
		p.expect(token.LBRACKET)
		key := p.parseType()
		p.expect(token.COLON)
		val := p.parseType()
		p.expect(token.RBRACKET)
		base = &ast.MapTypeExpr{P: p.tkPos(start), Key: key, Val: val}
	default:
		p.failNow("expected type, got %q", p.cur().String())
	}

	for p.at(token.LBRACKET) {
		p.advance()
		p.expect(token.RBRACKET)
		base = &ast.ArrayTypeExpr{P: p.tkPos(start), Elem: base}
	}
	return base
}

// ---------- expressions ----------

func (p *Parser) parseExpression() ast.Expr {
	return p.parseTernary()
}

func (p *Parser) parseTernary() ast.Expr {
	cond := p.parseOr()
	if p.at(token.QUESTION) {
		q := p.advance()
		a := p.parseOr()
		p.expect(token.COLON)
		b := p.parseTernary()
		return &ast.TernaryExpr{P: p.tkPos(q), Cond: cond, A: a, B: b}
	}
	return cond
}

func (p *Parser) parseOr() ast.Expr {
	left := p.parseAnd()
	for p.at(token.OR) {
		op := p.advance()
		right := p.parseAnd()
		left = &ast.BinaryExpr{P: p.tkPos(op), Op: token.OR, L: left, R: right}
	}
	return left
}

func (p *Parser) parseAnd() ast.Expr {
	left := p.parseEquality()
	for p.at(token.AND) {
		op := p.advance()
		right := p.parseEquality()
		left = &ast.BinaryExpr{P: p.tkPos(op), Op: token.AND, L: left, R: right}
	}
	return left
}

func (p *Parser) parseEquality() ast.Expr {
	left := p.parseComparison()
	for p.at(token.EQ) || p.at(token.NEQ) {
		op := p.advance()
		right := p.parseComparison()
		left = &ast.BinaryExpr{P: p.tkPos(op), Op: op.Type, L: left, R: right}
	}
	return left
}

func (p *Parser) parseComparison() ast.Expr {
	left := p.parseRange()
	for p.at(token.LT) || p.at(token.GT) || p.at(token.LTE) || p.at(token.GTE) {
		op := p.advance()
		right := p.parseRange()
		left = &ast.BinaryExpr{P: p.tkPos(op), Op: op.Type, L: left, R: right}
	}
	return left
}

func (p *Parser) parseRange() ast.Expr {
	lo := p.parseAdditive()
	if p.at(token.DOTDOT) {
		op := p.advance()
		hi := p.parseAdditive()
		return &ast.RangeExpr{P: p.tkPos(op), Lo: lo, Hi: hi}
	}
	return lo
}

func (p *Parser) parseAdditive() ast.Expr {
	left := p.parseMultiplicative()
	for p.at(token.PLUS) || p.at(token.MINUS) {
		op := p.advance()
		right := p.parseMultiplicative()
		left = &ast.BinaryExpr{P: p.tkPos(op), Op: op.Type, L: left, R: right}
	}
	return left
}

func (p *Parser) parseMultiplicative() ast.Expr {
	left := p.parseUnary()
	for p.at(token.STAR) || p.at(token.SLASH) || p.at(token.PERCENT) {
		op := p.advance()
		right := p.parseUnary()
		left = &ast.BinaryExpr{P: p.tkPos(op), Op: op.Type, L: left, R: right}
	}
	return left
}

func (p *Parser) parseUnary() ast.Expr {
	if p.at(token.MINUS) || p.at(token.BANG) {
		op := p.advance()
		x := p.parseUnary()
		return &ast.UnaryExpr{P: p.tkPos(op), Op: op.Type, X: x}
	}
	return p.parsePostfix()
}

func (p *Parser) parsePostfix() ast.Expr {
	e := p.parsePrimary()
	for {
		switch p.cur().Type {
		case token.DOT:
			p.advance()
			name := p.expectIdent()
			e = &ast.MemberExpr{P: e.Pos(), X: e, Name: name.Lit, NamePos: p.tkPos(name)}
		case token.LBRACKET:
			p.advance()
			p.depth++
			idx := p.parseExpression()
			p.depth--
			p.expect(token.RBRACKET)
			e = &ast.IndexExpr{P: e.Pos(), X: e, Index: idx}
		case token.LPAREN:
			p.advance()
			p.depth++
			args := p.parseArgs()
			p.depth--
			p.expect(token.RPAREN)
			e = &ast.CallExpr{P: e.Pos(), Callee: e, Args: args}
		case token.LBRACE:
			// Struct literal: TypeName { field: value, ... }
			// Only treat `Ident {` as a struct literal when what follows the
			// brace looks like struct fields (`field: value`) or an empty
			// struct (`}`). A bare `{` after an expression is a block opener
			// (e.g. `if x {` or `for tag in tags {`).
			if _, isIdent := e.(*ast.Ident); isIdent {
				t1, i1 := p.peekInfo()
				isStruct := false
				if t1.Type == token.RBRACE {
					isStruct = true
				} else if t1.Type == token.IDENT && i1+1 < len(p.toks) && p.toks[i1+1].Type == token.COLON {
					isStruct = true
				}
				if isStruct {
					e = p.parseStructLit(e.(*ast.Ident))
				} else {
					return e
				}
			} else {
				return e
			}
		default:
			return e
		}
	}
}

func (p *Parser) parseArgs() []ast.Expr {
	var args []ast.Expr
	p.skipNestedNewlines()
	if p.at(token.RPAREN) {
		return nil
	}
	for {
		args = append(args, p.parseExpression())
		p.skipNestedNewlines()
		if p.at(token.COMMA) {
			p.advance()
			p.skipNestedNewlines()
			if p.at(token.RPAREN) {
				break
			}
			continue
		}
		break
	}
	return args
}

func (p *Parser) parsePrimary() ast.Expr {
	t := p.cur()
	switch t.Type {
	case token.INT:
		p.advance()
		v, err := strconv.ParseInt(t.Lit, 10, 64)
		if err != nil {
			p.fail(t, "invalid integer literal %s", t.Lit)
		}
		return &ast.IntLit{P: p.tkPos(t), Value: v}

	case token.FLOAT:
		p.advance()
		v, err := strconv.ParseFloat(t.Lit, 64)
		if err != nil {
			p.fail(t, "invalid float literal %s", t.Lit)
		}
		return &ast.FloatLit{P: p.tkPos(t), Value: v}

	case token.FUNC:
		return p.parseFuncLit()

	case token.STRING:
		p.advance()
		return p.parseStrLit(t)

	case token.TRUE, token.FALSE:
		p.advance()
		return &ast.BoolLit{P: p.tkPos(t), Value: t.Type == token.TRUE}

	case token.IDENT:
		p.advance()
		return &ast.Ident{P: p.tkPos(t), Name: t.Lit}

	case token.LPAREN:
		p.advance()
		p.depth++
		e := p.parseExpression()
		p.depth--
		p.expect(token.RPAREN)
		return e

	case token.LBRACE:
		return p.parseMapLit()

	case token.LBRACKET:
		return p.parseArrayLit()

	case token.T_STR, token.T_INT, token.T_BOOL, token.T_ANY, token.T_DEC, token.MAP:
		p.failNow("type %q cannot be used as a value", t.Lit)

	case token.SOME, token.NONE, token.OK, token.ERR:
		// Constructors: Some(x), None, Ok(x), Err(e) — represented as
		// Ident calls that the checker resolves specially.
		p.advance()
		return &ast.Ident{P: p.tkPos(t), Name: t.Lit}

	case token.EOF:
		p.failNow("unexpected end of file, expected expression")

	default:
		p.failNow("unexpected %q, expected expression", t.String())
	}
	return nil
}

// parseFuncLit parses an anonymous function body used as a value.
// It shares syntax with func decl but has no name.
func (p *Parser) parseFuncLit() *ast.FuncLit {
	kw := p.expect(token.FUNC)
	fl := &ast.FuncLit{P: p.tkPos(kw)}
	fl.Params = p.parseParams()
	fl.Result = p.parseResultType()
	fl.Body = p.parseBlock()
	return fl
}

func (p *Parser) parseStrLit(tok token.Token) *ast.StrLit {
	lit := &ast.StrLit{P: p.tkPos(tok)}
	for _, seg := range tok.Segs {
		if !seg.IsExpr {
			lit.Parts = append(lit.Parts, ast.StrPart{Lit: seg.Text})
			continue
		}
		subToks, err := lexer.LexAt(seg.Text, p.file, seg.Line, seg.Col)
		if err != nil {
			panic(parsePanic{err})
		}
		sub := &Parser{toks: subToks, file: p.file}
		e := sub.parseExpression()
		if !sub.at(token.EOF) {
			sub.failNow("unexpected %q after interpolation expression", sub.cur().String())
		}
		lit.Parts = append(lit.Parts, ast.StrPart{Expr: e})
	}
	return lit
}

func (p *Parser) parseArrayLit() *ast.ArrayLit {
	open := p.expect(token.LBRACKET)
	a := &ast.ArrayLit{P: p.tkPos(open)}
	p.depth++
	p.skipNestedNewlines()
	if p.at(token.RBRACKET) {
		p.depth--
		p.advance()
		return a
	}
	for {
		a.Elems = append(a.Elems, p.parseExpression())
		p.skipNestedNewlines()
		if p.at(token.COMMA) {
			p.advance()
			p.skipNestedNewlines()
			if p.at(token.RBRACKET) {
				break
			}
			continue
		}
		break
	}
	p.depth--
	p.expect(token.RBRACKET)
	return a
}

func (p *Parser) parseMapLit() *ast.MapLit {
	open := p.expect(token.LBRACE)
	m := &ast.MapLit{P: p.tkPos(open)}

	p.skipNewlines()
	if p.at(token.RBRACE) {
		p.advance()
		return m
	}
	for {
		key := p.parseMapKey()
		p.expect(token.COLON)
		val := p.parseExpression()
		m.Entries = append(m.Entries, ast.MapEntry{Key: key, Val: val})

		hadNewline := false
		for p.at(token.NEWLINE) {
			hadNewline = true
			p.advance()
		}
		if p.at(token.RBRACE) {
			break
		}
		if p.at(token.COMMA) {
			p.advance()
			p.skipNewlines()
			if p.at(token.RBRACE) {
				break
			}
			continue
		}
		if hadNewline {
			continue // newline acts as entry separator
		}
		p.failNow("expected %q or %q after map entry, got %q",
			string(token.COMMA), string(token.RBRACE), p.cur().String())
	}
	p.expect(token.RBRACE)
	return m
}

func (p *Parser) parseMapKey() ast.Expr {
	t := p.cur()
	switch t.Type {
	case token.IDENT:
		p.advance()
		return &ast.StrLit{P: p.tkPos(t), Parts: []ast.StrPart{{Lit: t.Lit}}}
	case token.STRING:
		p.advance()
		return p.parseStrLit(t)
	default:
		p.failNow("expected map key (identifier or string), got %q", t.String())
	}
	return nil
}

// parseStructLit parses `TypeName { field: value, ... }` assuming the
// TypeName Ident has already been consumed.
func (p *Parser) parseStructLit(typeName *ast.Ident) *ast.StructLit {
	open := p.expect(token.LBRACE)
	sl := &ast.StructLit{
		P:           p.tkPos(open),
		TypeName:    typeName.Name,
		TypeNamePos: typeName.P,
	}
	p.skipNewlines()
	if p.at(token.RBRACE) {
		p.advance()
		return sl
	}
	for {
		fname := p.expectIdent()
		p.expect(token.COLON)
		val := p.parseExpression()
		sl.Entries = append(sl.Entries, ast.StructLitEntry{
			Name:    fname.Lit,
			NamePos: p.tkPos(fname),
			Value:   val,
		})
		p.skipNewlines()
		if p.at(token.COMMA) {
			p.advance()
			p.skipNewlines()
			if p.at(token.RBRACE) {
				break
			}
			continue
		}
		if p.at(token.RBRACE) {
			break
		}
		continue
	}
	p.expect(token.RBRACE)
	return sl
}
