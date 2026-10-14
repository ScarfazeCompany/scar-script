package checker

import (
	"strings"

	"github.com/ScarfazeCompany/scar-script/internal/ast"
	"github.com/ScarfazeCompany/scar-script/internal/diag"
	"github.com/ScarfazeCompany/scar-script/internal/token"
	"github.com/ScarfazeCompany/scar-script/internal/types"
)

type varInfo struct {
	typ     types.Type
	mut     bool
	isConst bool
	used    bool
	name    string
	node    ast.Node // declaring statement (DeclStmt or ForStmt)
	pos     diag.Position
}

type scope struct {
	parent *scope
	vars   map[string]*varInfo
}

func (s *scope) lookup(name string) *varInfo {
	for sc := s; sc != nil; sc = sc.parent {
		if v, ok := sc.vars[name]; ok {
			return v
		}
	}
	return nil
}

type Checker struct {
	file       string
	scope      *scope
	types      map[ast.Expr]types.Type
	allVars    []*varInfo
	funcs      map[string]*types.Func // all top-level functions (phase 1)
	inFunc     int                    // >0 while checking a function body
	blockDepth int                    // nesting of blocks (top level == 0)
	curResult  types.Type             // result type of the current function
}

// Result carries everything the code generator needs after a successful
// check: the resolved type of every expression, and the variables that are
// never read (so codegen can emit blank reads for Go's "declared and not
// used" rule).
type Result struct {
	Types  map[ast.Expr]types.Type
	Unused map[ast.Node][]string
}

type checkPanic struct{ err *diag.Error }

// Check performs full static checking on a parsed program.
// It returns the first error found, or nil.
func Check(prog *ast.Program) (res *Result, err *diag.Error) {
	c := &Checker{
		file:  prog.File,
		scope: &scope{vars: map[string]*varInfo{}},
		types: map[ast.Expr]types.Type{},
		funcs: map[string]*types.Func{},
	}
	defer func() {
		if r := recover(); r != nil {
			if cp, ok := r.(checkPanic); ok {
				res, err = nil, cp.err
				return
			}
			panic(r)
		}
	}()
	// Phase 1: register every top-level function signature so calls work
	// before (or after) the declaration, and recursion works.
	for _, s := range prog.Stmts {
		if fd, ok := s.(*ast.FuncDecl); ok {
			c.registerFunc(fd)
		}
	}
	for _, s := range prog.Stmts {
		c.checkStmt(s)
	}

	unused := map[ast.Node][]string{}
	for _, v := range c.allVars {
		if !v.used && !v.isConst && v.node != nil {
			unused[v.node] = append(unused[v.node], v.name)
		}
	}
	return &Result{Types: c.types, Unused: unused}, nil
}

func (c *Checker) fail(pos diag.Position, format string, args ...any) {
	panic(checkPanic{diag.New(pos, format, args...)})
}

// isIntLit reports whether e is an (optionally negated) int literal —
// the only int value widened to dec implicitly.
func isIntLit(e ast.Expr) bool {
	switch v := e.(type) {
	case *ast.IntLit:
		return true
	case *ast.UnaryExpr:
		if v.Op == token.MINUS {
			_, ok := v.X.(*ast.IntLit)
			return ok
		}
	}
	return false
}

// assignableVal is types.Assignable plus the int-literal → dec widening.
func assignableVal(dst, src types.Type, e ast.Expr) bool {
	if types.Assignable(dst, src) {
		return true
	}
	if types.Equal(dst, types.DecType) && types.Equal(src, types.IntType) && isIntLit(e) {
		return true
	}
	return false
}

// isModule reports whether name is one of the built-in module prefixes.
func isModule(name string) bool {
	return name == "math" || name == "file" || name == "sys" || name == "time"
}

func isBuiltin(name string) bool {
	return name == "print" || name == "log" || isModule(name)
}

var goKeywords = map[string]bool{
	"break": true, "case": true, "chan": true, "continue": true,
	"default": true, "defer": true, "else": true, "fallthrough": true,
	"for": true, "func": true, "go": true, "goto": true, "if": true,
	"import": true, "map": true, "package": true, "range": true,
	"return": true, "select": true, "struct": true, "switch": true,
	"type": true, "var": true,
}

// checkName rejects builtin/module names, Go keywords, and `_`-prefixed
// identifiers for any declaration (variables, parameters, functions).
func (c *Checker) checkName(name string, pos diag.Position) {
	if isModule(name) {
		c.fail(pos, "%q is a module name and cannot be redeclared", name)
	}
	if name == "print" || name == "log" {
		c.fail(pos, "%q is a builtin name and cannot be redeclared", name)
	}
	if goKeywords[name] {
		c.fail(pos, "%q is a reserved word", name)
	}
	if len(name) > 0 && name[0] == '_' {
		c.fail(pos, "identifiers cannot start with '_'")
	}
}

func (c *Checker) declare(name string, v *varInfo, pos diag.Position, node ast.Node) {
	c.checkName(name, pos)
	if _, ok := c.funcs[name]; ok {
		c.fail(pos, "%q is already declared as a function", name)
	}
	if _, ok := c.scope.vars[name]; ok {
		c.fail(pos, "%q is already declared in this scope", name)
	}
	v.name = name
	v.node = node
	v.pos = pos
	c.scope.vars[name] = v
	c.allVars = append(c.allVars, v)
}

// registerFunc resolves and records a top-level function's signature.
func (c *Checker) registerFunc(fd *ast.FuncDecl) {
	c.checkName(fd.Name, fd.NamePos)
	if _, ok := c.funcs[fd.Name]; ok {
		c.fail(fd.NamePos, "%q is already declared", fd.Name)
	}
	ft := &types.Func{Result: types.VoidType, Params: make([]types.Type, len(fd.Params))}
	seen := map[string]bool{}
	for i, prm := range fd.Params {
		c.checkName(prm.Name, prm.NamePos)
		if seen[prm.Name] {
			c.fail(prm.NamePos, "duplicate parameter %q", prm.Name)
		}
		seen[prm.Name] = true
		ft.Params[i] = c.resolveType(prm.Type)
	}
	if fd.Result != nil {
		ft.Result = c.resolveResult(fd.Result)
	}
	c.funcs[fd.Name] = ft
}

func (c *Checker) push() { c.scope = &scope{parent: c.scope, vars: map[string]*varInfo{}} }

// pushIsolated starts a scope with no parent (named functions see only
// their own parameters, matching what the Go emitter can access).
func (c *Checker) pushIsolated() { c.scope = &scope{vars: map[string]*varInfo{}} }

func (c *Checker) pop() { c.scope = c.scope.parent }

// ---------- statements ----------

func (c *Checker) checkStmt(s ast.Stmt) {
	switch st := s.(type) {
	case *ast.DeclStmt:
		c.checkDecl(st)
	case *ast.MultiDeclStmt:
		c.checkMultiDecl(st)
	case *ast.AssignStmt:
		c.checkAssign(st)
	case *ast.IncDecStmt:
		c.checkIncDec(st)
	case *ast.IfStmt:
		c.checkIf(st)
	case *ast.ForStmt:
		c.checkFor(st)
	case *ast.BlockStmt:
		c.checkBlockInner(st)
	case *ast.ExprStmt:
		c.checkExprStmt(st)
	case *ast.FuncDecl:
		if c.blockDepth > 0 {
			c.fail(st.P, "functions can only be declared at the top level")
		}
		c.checkFuncDecl(st)
	case *ast.ReturnStmt:
		c.checkReturn(st)
	default:
		c.fail(s.Pos(), "internal: unknown statement %T", s)
	}
}

func (c *Checker) checkBlockInner(b *ast.BlockStmt) {
	c.push()
	c.blockDepth++
	c.checkStmtList(b.Stmts)
	c.blockDepth--
	c.pop()
}

// checkStmtList checks statements and rejects code after a terminating
// statement (dead code).
func (c *Checker) checkStmtList(stmts []ast.Stmt) {
	terminated := false
	for _, s := range stmts {
		if terminated {
			c.fail(s.Pos(), "unreachable code")
		}
		c.checkStmt(s)
		terminated = stmtReturns(s)
	}
}

func (c *Checker) checkDecl(d *ast.DeclStmt) {
	var ty, vt types.Type
	if d.Type != nil {
		ty = c.resolveType(d.Type)
		vt = c.checkExpr(d.Value, ty)
	} else {
		vt = c.checkExpr(d.Value, nil)
		if vt == types.RangeT {
			c.fail(d.Value.Pos(), "range expressions can only be used in 'for' loops")
		}
		if vt == types.VoidType {
			c.fail(d.Value.Pos(), "cannot assign a void value")
		}
		if _, isTup := vt.(*types.Tuple); isTup {
			c.fail(d.Value.Pos(), "function returns multiple values — use 'val a, b = %s(...)'",
				c.callName(d.Value))
		}
		ty = vt
	}
	if vt == types.VoidType {
		c.fail(d.Value.Pos(), "cannot assign a void value")
	}
	if d.Kind == ast.Const {
		if b, ok := ty.(*types.Basic); !ok || (b.K != types.Str && b.K != types.Int && b.K != types.Bool) {
			c.fail(d.NamePos, "const must be of type str, int, or bool")
		}
		if !c.isConstExpr(d.Value) {
			c.fail(d.Value.Pos(), "const requires a literal value (str, int, bool)")
		}
	}
	if !assignableVal(ty, vt, d.Value) {
		if vt == types.RangeT {
			c.fail(d.Value.Pos(), "range expressions can only be used in 'for' loops")
		}
		c.fail(d.Value.Pos(), "cannot use value of type %q where type %q is expected", vt, ty)
	}
	c.declare(d.Name, &varInfo{
		typ:     ty,
		mut:     d.Kind == ast.Mut,
		isConst: d.Kind == ast.Const,
	}, d.NamePos, d)
}

// callName renders the callee of a call expression for error messages.
func (c *Checker) callName(e ast.Expr) string {
	if call, ok := e.(*ast.CallExpr); ok {
		return c.calleeName(call.Callee)
	}
	return "f"
}

func (c *Checker) calleeName(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.MemberExpr:
		return c.calleeName(v.X) + "." + v.Name
	}
	return "f"
}

// checkMultiDecl handles `val a, b = f()` for tuple-returning calls.
func (c *Checker) checkMultiDecl(md *ast.MultiDeclStmt) {
	if md.Kind == ast.Const {
		c.fail(md.P, "const cannot declare multiple values")
	}
	if _, ok := md.Value.(*ast.CallExpr); !ok {
		c.fail(md.Value.Pos(), "multiple declarations require a function call")
	}
	vt := c.checkExpr(md.Value, nil)
	tup, isTup := vt.(*types.Tuple)
	if !isTup {
		c.fail(md.Value.Pos(), "expected a function returning multiple values, got %q", vt)
	}
	if len(tup.Elems) != len(md.Names) {
		c.fail(md.Value.Pos(), "expected %d return values, got %d", len(tup.Elems), len(md.Names))
	}
	for i, name := range md.Names {
		c.declare(name, &varInfo{typ: tup.Elems[i], mut: md.Kind == ast.Mut},
			md.NamePoss[i], md)
	}
}

func (c *Checker) isConstExpr(e ast.Expr) bool {
	switch v := e.(type) {
	case *ast.IntLit, *ast.BoolLit:
		return true
	case *ast.StrLit:
		for _, p := range v.Parts {
			if p.Expr != nil {
				return false
			}
		}
		return true
	case *ast.UnaryExpr:
		if v.Op == token.MINUS {
			_, ok := v.X.(*ast.IntLit)
			return ok
		}
		return false
	}
	return false
}

func (c *Checker) checkAssign(a *ast.AssignStmt) {
	info := c.lookupVar(a.Name, a.NamePos)
	if info.isConst {
		c.fail(a.NamePos, "cannot assign to constant %q", a.Name)
	}
	if !info.mut {
		c.fail(a.NamePos, "cannot assign to immutable variable %q (declare it with 'mut')", a.Name)
	}
	vt := c.checkExpr(a.Value, info.typ)
	if _, isTup := vt.(*types.Tuple); isTup {
		c.fail(a.Value.Pos(), "function returns multiple values — cannot assign to a single variable")
	}
	if !assignableVal(info.typ, vt, a.Value) {
		if vt == types.RangeT {
			c.fail(a.Value.Pos(), "range expressions can only be used in 'for' loops")
		}
		c.fail(a.Value.Pos(), "cannot use value of type %q where type %q is expected", vt, info.typ)
	}
}

func (c *Checker) lookupVar(name string, pos diag.Position) *varInfo {
	info := c.scope.lookup(name)
	if info == nil {
		if isBuiltin(name) {
			c.fail(pos, "%q is a builtin name", name)
		}
		c.fail(pos, "undefined variable %q", name)
	}
	return info
}

func (c *Checker) checkIncDec(s *ast.IncDecStmt) {
	info := c.lookupVar(s.Name, s.NamePos)
	if info.isConst {
		c.fail(s.NamePos, "cannot modify constant %q", s.Name)
	}
	if !info.mut {
		c.fail(s.NamePos, "cannot modify immutable variable %q (declare it with 'mut')", s.Name)
	}
	if !types.Equal(info.typ, types.IntType) {
		c.fail(s.NamePos, "%q requires an int variable, but %q has type %q", s.Op, s.Name, info.typ)
	}
}

func (c *Checker) checkIf(s *ast.IfStmt) {
	ct := c.checkExpr(s.Cond, nil)
	if !types.Equal(ct, types.BoolType) {
		c.fail(s.Cond.Pos(), "condition must be bool, got %q", ct)
	}
	c.checkBlockInner(s.Then)
	if s.Else != nil {
		c.checkStmt(s.Else)
	}
}

func (c *Checker) checkFor(f *ast.ForStmt) {
	c.push()
	defer c.pop()

	iterT := c.checkExpr(f.Iterable, nil)

	switch f.Kind {
	case ast.ForCond:
		if !types.Equal(iterT, types.BoolType) {
			c.fail(f.Iterable.Pos(), "loop condition must be bool, got %q", iterT)
		}

	case ast.ForIterable:
		vt := c.iteratedType(f.Iterable, iterT)
		c.declare(f.Var, &varInfo{typ: vt, mut: true}, f.VarPos, f)

	case ast.ForTwo:
		var t1, t2 types.Type
		switch it := iterT.(type) {
		case *types.Array:
			t1, t2 = types.IntType, it.Elem
		case *types.Map:
			t1, t2 = it.Key, it.Val
		default:
			c.fail(f.Iterable.Pos(), "cannot iterate with two variables over type %q", iterT)
		}
		c.declare(f.Var, &varInfo{typ: t1, mut: true}, f.VarPos, f)
		c.declare(f.Var2, &varInfo{typ: t2, mut: true}, f.Var2Pos, f)
	}

	c.checkBlockInner(f.Body)
}

func (c *Checker) iteratedType(e ast.Expr, t types.Type) types.Type {
	switch it := t.(type) {
	case *types.Basic:
		if it.K == types.Range {
			return types.IntType
		}
	case *types.Array:
		return it.Elem
	case *types.Map:
		return it.Key
	}
	c.fail(e.Pos(), "cannot iterate over type %q", t)
	return nil
}

func (c *Checker) checkExprStmt(s *ast.ExprStmt) {
	c.checkExpr(s.X, nil)
	if _, ok := s.X.(*ast.CallExpr); !ok {
		c.fail(s.X.Pos(), "expression has no effect — only calls can be used as statements")
	}
}

// ---------- functions ----------

func (c *Checker) checkFuncDecl(fd *ast.FuncDecl) {
	sig := c.funcs[fd.Name]
	c.inFunc++
	oldResult := c.curResult
	c.curResult = sig.Result
	oldScope := c.scope
	c.pushIsolated()
	c.declareParams(fd.Params)
	c.checkFuncBody(fd.Body, fd.NamePos, sig.Result)
	c.scope = oldScope
	c.curResult = oldResult
	c.inFunc--
}

func (c *Checker) checkFuncLit(fl *ast.FuncLit) types.Type {
	ft := &types.Func{Result: types.VoidType}
	for _, prm := range fl.Params {
		c.checkName(prm.Name, prm.NamePos)
		ft.Params = append(ft.Params, c.resolveType(prm.Type))
	}
	if fl.Result != nil {
		ft.Result = c.resolveResult(fl.Result)
	}
	c.inFunc++
	oldResult := c.curResult
	c.curResult = ft.Result
	c.push() // anonymous functions capture the enclosing scope
	c.declareParams(fl.Params)
	c.checkFuncBody(fl.Body, fl.P, ft.Result)
	c.pop()
	c.curResult = oldResult
	c.inFunc--
	return ft
}

// declareParams binds parameters in the current (already isolated) scope.
func (c *Checker) declareParams(params []ast.Param) {
	seen := map[string]bool{}
	for _, prm := range params {
		if seen[prm.Name] {
			c.fail(prm.NamePos, "duplicate parameter %q", prm.Name)
		}
		seen[prm.Name] = true
		c.declare(prm.Name, &varInfo{typ: c.resolveType(prm.Type), mut: false}, prm.NamePos, nil)
	}
}

// checkFuncBody checks statements of a function body and enforces the
// return rules: non-void functions must return on every path, and code
// after a return is dead.
func (c *Checker) checkFuncBody(body *ast.BlockStmt, namePos diag.Position, result types.Type) {
	c.blockDepth++
	c.checkStmtList(body.Stmts)
	c.blockDepth--
	if result != types.VoidType && !alwaysReturns(body.Stmts) {
		c.fail(namePos, "missing return — function must return %q on every path", result)
	}
}

func (c *Checker) checkReturn(r *ast.ReturnStmt) {
	if c.inFunc == 0 {
		c.fail(r.P, "return outside of a function")
	}
	res := c.curResult

	if res == types.VoidType {
		if len(r.Values) > 0 {
			c.fail(r.Values[0].Pos(), "void function cannot return a value")
		}
		return
	}

	if tup, ok := res.(*types.Tuple); ok {
		if len(r.Values) == 1 {
			// `return f()` — pass a whole tuple through.
			vt := c.checkExpr(r.Values[0], nil)
			if vtTup, isTup := vt.(*types.Tuple); isTup {
				if types.Equal(vtTup, tup) {
					return
				}
				c.fail(r.Values[0].Pos(), "cannot return %s — function must return %s", vt, tup)
			}
			c.fail(r.P, "expected %d return values, got %d", len(tup.Elems), len(r.Values))
		}
		if len(r.Values) != len(tup.Elems) {
			c.fail(r.P, "expected %d return values, got %d", len(tup.Elems), len(r.Values))
		}
		for i, v := range r.Values {
			vt := c.exprVal(v, tup.Elems[i])
			if !assignableVal(tup.Elems[i], vt, v) {
				c.fail(v.Pos(), "return value %d expects %s, got %s", i+1, tup.Elems[i], vt)
			}
		}
		return
	}

	if len(r.Values) == 0 {
		c.fail(r.P, "missing return value — function must return %q", res)
	}
	if len(r.Values) > 1 {
		c.fail(r.P, "function returns a single %q, got %d values", res, len(r.Values))
	}
	vt := c.exprVal(r.Values[0], res)
	if !assignableVal(res, vt, r.Values[0]) {
		if vt == types.RangeT {
			c.fail(r.Values[0].Pos(), "range expressions can only be used in 'for' loops")
		}
		c.fail(r.Values[0].Pos(), "cannot return value of type %q where type %q is expected", vt, res)
	}
}

// stmtReturns reports whether control never falls past s.
func stmtReturns(s ast.Stmt) bool {
	switch st := s.(type) {
	case *ast.ReturnStmt:
		return true
	case *ast.BlockStmt:
		return alwaysReturns(st.Stmts)
	case *ast.IfStmt:
		if st.Else == nil {
			return false
		}
		return alwaysReturns(st.Then.Stmts) && stmtReturns(st.Else)
	}
	return false
}

func alwaysReturns(stmts []ast.Stmt) bool {
	if len(stmts) == 0 {
		return false
	}
	return stmtReturns(stmts[len(stmts)-1])
}

// ---------- types ----------

func (c *Checker) resolveType(te ast.TypeExpr) types.Type {
	switch t := te.(type) {
	case *ast.BasicType:
		switch t.Kind {
		case "str":
			return types.StrType
		case "int":
			return types.IntType
		case "bool":
			return types.BoolType
		case "dec":
			return types.DecType
		default:
			return types.AnyType
		}
	case *ast.ArrayTypeExpr:
		return &types.Array{Elem: c.resolveType(t.Elem)}
	case *ast.MapTypeExpr:
		k := c.resolveType(t.Key)
		if !types.Equal(k, types.StrType) && !types.Equal(k, types.IntType) {
			c.fail(t.Key.Pos(), "map key must be 'str' or 'int', got %q", k)
		}
		return &types.Map{Key: k, Val: c.resolveType(t.Val)}
	case *ast.TupleTypeExpr:
		c.fail(t.P, "tuple types can only be used as a function return type")
	}
	c.fail(te.Pos(), "internal: unknown type %T", te)
	return nil
}

// resolveResult resolves a function result type; unlike resolveType it
// allows tuples `(int, bool)`.
func (c *Checker) resolveResult(te ast.TypeExpr) types.Type {
	if t, ok := te.(*ast.TupleTypeExpr); ok {
		tup := &types.Tuple{}
		for _, el := range t.Elems {
			tup.Elems = append(tup.Elems, c.resolveResult(el))
		}
		return tup
	}
	return c.resolveType(te)
}

// ---------- expressions ----------

// exprVal is checkExpr plus the rule that void/range/tuple values are not
// allowed in ordinary single-value positions.
func (c *Checker) exprVal(e ast.Expr, expected types.Type) types.Type {
	t := c.checkExpr(e, expected)
	if t == types.RangeT {
		c.fail(e.Pos(), "range expressions can only be used in 'for' loops")
	}
	if t == types.VoidType {
		c.fail(e.Pos(), "cannot use a void value")
	}
	if _, isTup := t.(*types.Tuple); isTup {
		c.fail(e.Pos(), "function returns multiple values — use 'val a, b = %s(...)'",
			c.callName(e))
	}
	return t
}

// checkExpr records the resolved type of every expression node; the code
// generator reads these back via Result.Types.
func (c *Checker) checkExpr(e ast.Expr, expected types.Type) types.Type {
	t := c.checkExprInner(e, expected)
	c.types[e] = t
	return t
}

func (c *Checker) checkExprInner(e ast.Expr, expected types.Type) types.Type {
	switch ex := e.(type) {
	case *ast.IntLit:
		return types.IntType
	case *ast.FloatLit:
		return types.DecType
	case *ast.BoolLit:
		return types.BoolType
	case *ast.StrLit:
		return c.checkStrLit(ex)
	case *ast.Ident:
		return c.checkIdent(ex)
	case *ast.ArrayLit:
		return c.checkArrayLit(ex, expected)
	case *ast.MapLit:
		return c.checkMapLit(ex, expected)
	case *ast.BinaryExpr:
		return c.checkBinary(ex)
	case *ast.UnaryExpr:
		return c.checkUnary(ex)
	case *ast.TernaryExpr:
		return c.checkTernary(ex, expected)
	case *ast.RangeExpr:
		return c.checkRange(ex)
	case *ast.CallExpr:
		return c.checkCall(ex)
	case *ast.MemberExpr:
		return c.checkMember(ex)
	case *ast.IndexExpr:
		return c.checkIndex(ex)
	case *ast.FuncLit:
		return c.checkFuncLit(ex)
	}
	c.fail(e.Pos(), "internal: unknown expression %T", e)
	return nil
}

func (c *Checker) checkStrLit(s *ast.StrLit) types.Type {
	for _, p := range s.Parts {
		if p.Expr == nil {
			continue
		}
		t := c.checkExpr(p.Expr, nil)
		if t == types.VoidType {
			c.fail(p.Expr.Pos(), "cannot interpolate a void value")
		}
		if t == types.RangeT {
			c.fail(p.Expr.Pos(), "range expressions can only be used in 'for' loops")
		}
	}
	return types.StrType
}

func (c *Checker) checkIdent(id *ast.Ident) types.Type {
	if id.Name == "print" {
		c.fail(id.P, "'print' is a builtin function — call it like print(...)")
	}
	if id.Name == "log" {
		c.fail(id.P, "'log' is a builtin — use log.info(...) or log.error(...)")
	}
	if isModule(id.Name) {
		c.fail(id.P, "%q is a module — use %s.<function>(...)", id.Name, id.Name)
	}
	info := c.scope.lookup(id.Name)
	if info == nil {
		if _, ok := c.funcs[id.Name]; ok {
			c.fail(id.P, "%q is a function — call it like %s(...)", id.Name, id.Name)
		}
		c.fail(id.P, "undefined variable %q", id.Name)
	}
	info.used = true
	return info.typ
}

func (c *Checker) checkArrayLit(a *ast.ArrayLit, expected types.Type) types.Type {
	if len(a.Elems) == 0 {
		if exp, ok := expected.(*types.Array); ok {
			return exp
		}
		if types.Equal(expected, types.AnyType) {
			return &types.Array{Elem: types.AnyType}
		}
		c.fail(a.P, "cannot infer type of empty array literal — add a type annotation")
	}
	t := c.exprVal(a.Elems[0], nil)
	for _, el := range a.Elems[1:] {
		et := c.exprVal(el, nil)
		if !types.Equal(t, et) {
			c.fail(el.Pos(), "array elements must have the same type, got %q and %q", t, et)
		}
	}
	return &types.Array{Elem: t}
}

func (c *Checker) checkMapLit(m *ast.MapLit, expected types.Type) types.Type {
	if len(m.Entries) == 0 {
		if exp, ok := expected.(*types.Map); ok {
			return exp
		}
		if types.Equal(expected, types.AnyType) {
			return &types.Map{Key: types.StrType, Val: types.AnyType}
		}
		c.fail(m.P, "cannot infer type of empty map literal — add a type annotation")
	}
	var vt types.Type
	for i, en := range m.Entries {
		c.exprVal(en.Key, nil) // keys are string literals (checked as str)
		t := c.exprVal(en.Val, nil)
		if i == 0 {
			vt = t
			continue
		}
		if !types.Equal(vt, t) {
			c.fail(en.Val.Pos(), "map values must have the same type, got %q and %q", vt, t)
		}
	}
	return &types.Map{Key: types.StrType, Val: vt}
}

func (c *Checker) checkBinary(b *ast.BinaryExpr) types.Type {
	lt := c.exprVal(b.L, nil)
	rt := c.exprVal(b.R, nil)

	isStr := func(t types.Type) bool { return types.Equal(t, types.StrType) }
	isBool := func(t types.Type) bool { return types.Equal(t, types.BoolType) }

	switch b.Op {
	case "+":
		if t, ok := types.Promote(lt, rt); ok {
			return t
		}
		if isStr(lt) && isStr(rt) {
			return types.StrType
		}
		c.fail(b.P, "operator '+' is not defined for %q and %q", lt, rt)
	case "-", "*", "/":
		if t, ok := types.Promote(lt, rt); ok {
			return t
		}
		c.fail(b.P, "operator %q requires numeric operands, got %q and %q", b.Op, lt, rt)
	case "%":
		if types.Equal(lt, types.IntType) && types.Equal(rt, types.IntType) {
			return types.IntType
		}
		c.fail(b.P, "operator %q requires int operands, got %q and %q", b.Op, lt, rt)
	case "<", ">", "<=", ">=":
		if _, ok := types.Promote(lt, rt); ok {
			return types.BoolType
		}
		if isStr(lt) && isStr(rt) {
			return types.BoolType
		}
		c.fail(b.P, "operator %q is not defined for %q and %q", b.Op, lt, rt)
	case "==", "!=":
		if types.Equal(lt, types.AnyType) || types.Equal(rt, types.AnyType) || types.Equal(lt, rt) {
			return types.BoolType
		}
		if _, ok := types.Promote(lt, rt); ok {
			return types.BoolType
		}
		c.fail(b.P, "cannot compare %q with %q", lt, rt)
	case "and":
		if isBool(lt) && isBool(rt) {
			return types.BoolType
		}
		c.fail(b.P, "'and' requires bool operands, got %q and %q", lt, rt)
	case "or":
		return c.checkOr(b, lt, rt)
	}
	c.fail(b.P, "internal: unknown operator %q", b.Op)
	return nil
}

func (c *Checker) checkOr(b *ast.BinaryExpr, lt, rt types.Type) types.Type {
	if types.Equal(lt, types.BoolType) && types.Equal(rt, types.BoolType) {
		return types.BoolType
	}
	if types.Equal(lt, types.AnyType) {
		return rt
	}
	if types.Equal(lt, rt) && types.IsZeroCheckable(lt) {
		return lt
	}
	c.fail(b.P, "'or' requires bool operands, matching str/int/dec operands, or 'any' on the left — got %q and %q", lt, rt)
	return nil
}

func (c *Checker) checkUnary(u *ast.UnaryExpr) types.Type {
	t := c.exprVal(u.X, nil)
	switch u.Op {
	case "-":
		if !types.IsNumeric(t) {
			c.fail(u.P, "unary '-' requires int or dec, got %q", t)
		}
		return t
	default: // "!"
		if !types.Equal(t, types.BoolType) {
			c.fail(u.P, "'!' requires bool, got %q", t)
		}
		return types.BoolType
	}
}

func (c *Checker) checkTernary(t *ast.TernaryExpr, expected types.Type) types.Type {
	ct := c.exprVal(t.Cond, nil)
	if !types.Equal(ct, types.BoolType) {
		c.fail(t.Cond.Pos(), "condition must be bool, got %q", ct)
	}
	a := c.exprVal(t.A, expected)
	b := c.exprVal(t.B, expected)
	if !types.Equal(a, b) {
		c.fail(t.P, "ternary branches must have the same type, got %q and %q", a, b)
	}
	return a
}

func (c *Checker) checkRange(r *ast.RangeExpr) types.Type {
	lo := c.exprVal(r.Lo, nil)
	hi := c.exprVal(r.Hi, nil)
	if !types.Equal(lo, types.IntType) || !types.Equal(hi, types.IntType) {
		c.fail(r.P, "range bounds must be int, got %q and %q", lo, hi)
	}
	return types.RangeT
}

func (c *Checker) checkIndex(x *ast.IndexExpr) types.Type {
	base := c.exprVal(x.X, nil)
	idx := c.exprVal(x.Index, nil)
	switch bt := base.(type) {
	case *types.Array:
		if !types.Equal(idx, types.IntType) {
			c.fail(x.Index.Pos(), "array index must be int, got %q", idx)
		}
		return bt.Elem
	case *types.Map:
		if !types.Equal(idx, bt.Key) {
			c.fail(x.Index.Pos(), "map index must be %q, got %q", bt.Key, idx)
		}
		return bt.Val
	}
	c.fail(x.P, "type %q cannot be indexed", base)
	return nil
}

func (c *Checker) checkMember(m *ast.MemberExpr) types.Type {
	base := c.checkExpr(m.X, nil)
	switch bt := base.(type) {
	case *types.Basic:
		if bt.K == types.Any {
			return types.AnyType
		}
	case *types.Map:
		if types.Equal(bt.Key, types.StrType) {
			if isMethod(c, m.Name, base) {
				c.fail(m.NamePos, "method %q must be called with ()", m.Name)
			}
			return bt.Val
		}
	}
	if isMethod(c, m.Name, base) {
		c.fail(m.NamePos, "method %q must be called with ()", m.Name)
	}
	c.fail(m.NamePos, "type %q has no field %q", base, m.Name)
	return nil
}

func isMethod(c *Checker, name string, recv types.Type) bool {
	if b, ok := recv.(*types.Basic); ok {
		switch b.K {
		case types.Str:
			return name == "trim" || name == "toUpper" || name == "contains" || name == "len"
		case types.Int, types.Bool:
			return false
		}
	}
	if _, ok := recv.(*types.Array); ok {
		return name == "add" || name == "remove" || name == "len"
	}
	if _, ok := recv.(*types.Map); ok {
		return name == "len"
	}
	return false
}

func (c *Checker) checkCall(call *ast.CallExpr) types.Type {
	switch callee := call.Callee.(type) {
	case *ast.Ident:
		return c.checkBuiltinCall(callee, call)
	case *ast.MemberExpr:
		return c.checkMemberCall(callee, call)
	}
	c.fail(call.Callee.Pos(), "expression of this type is not callable")
	return nil
}

func (c *Checker) checkBuiltinCall(id *ast.Ident, call *ast.CallExpr) types.Type {
	if id.Name == "print" {
		for _, a := range call.Args {
			c.exprVal(a, nil)
		}
		return types.VoidType
	}
	if info := c.scope.lookup(id.Name); info != nil {
		ft, ok := info.typ.(*types.Func)
		if !ok {
			c.fail(id.P, "variable %q of type %q is not a function", id.Name, info.typ)
		}
		info.used = true
		c.checkArgs(call, ft.Params)
		return ft.Result
	}
	if sig, ok := c.funcs[id.Name]; ok {
		c.checkArgs(call, sig.Params)
		return sig.Result
	}
	if isModule(id.Name) {
		c.fail(id.P, "%q is a module — use %s.<function>(...)", id.Name, id.Name)
	}
	if isBuiltin(id.Name) {
		c.fail(id.P, "builtin %q cannot be called like this", id.Name)
	}
	c.fail(id.P, "undefined function %q", id.Name)
	return nil
}

// checkArgs type-checks a call's arguments against the parameter list.
// Errors use the spec's format: `argument 1 expects int, got str`.
func (c *Checker) checkArgs(call *ast.CallExpr, params []types.Type) {
	if len(call.Args) != len(params) {
		c.fail(call.P, "expected %d argument(s), got %d", len(params), len(call.Args))
	}
	for i, a := range call.Args {
		t := c.exprVal(a, params[i])
		if !assignableVal(params[i], t, a) {
			c.fail(a.Pos(), "argument %d expects %s, got %s", i+1, params[i], t)
		}
	}
}

// modFunc is one built-in module function's signature.
type modFunc struct {
	name   string
	params []types.Type
	result types.Type
}

// modules lists the built-in math/file/sys/time functions in a stable
// order (used for "available:" error messages too).
var modules = map[string][]modFunc{
	"math": {
		{"sqrt", []types.Type{types.DecType}, types.DecType},
		{"pow", []types.Type{types.DecType, types.DecType}, types.DecType},
		{"abs", []types.Type{types.IntType}, types.IntType},
		{"round", []types.Type{types.DecType}, types.IntType},
		{"rand", []types.Type{types.IntType, types.IntType}, types.IntType},
	},
	"file": {
		{"write", []types.Type{types.StrType, types.StrType}, types.VoidType},
		{"append", []types.Type{types.StrType, types.StrType}, types.VoidType},
		{"read", []types.Type{types.StrType}, types.StrType},
		{"exists", []types.Type{types.StrType}, types.BoolType},
		{"delete", []types.Type{types.StrType}, types.VoidType},
	},
	"sys": {
		{"env", []types.Type{types.StrType}, types.StrType},
	},
	"time": {
		{"now", nil, types.StrType},
		{"sleep", []types.Type{types.IntType}, types.VoidType},
	},
}

func (c *Checker) checkModuleCall(mod string, m *ast.MemberExpr, call *ast.CallExpr) types.Type {
	fns := modules[mod]
	for _, fn := range fns {
		if fn.name == m.Name {
			c.checkArgs(call, fn.params)
			return fn.result
		}
	}
	names := make([]string, len(fns))
	for i, fn := range fns {
		names[i] = fn.name
	}
	c.fail(m.NamePos, "%s has no function %q (available: %s)",
		mod, m.Name, strings.Join(names, ", "))
	return nil
}

func (c *Checker) checkMemberCall(m *ast.MemberExpr, call *ast.CallExpr) types.Type {
	// math.* / file.* / sys.* / time.*
	if id, ok := m.X.(*ast.Ident); ok && isModule(id.Name) {
		return c.checkModuleCall(id.Name, m, call)
	}

	// log.info / log.error
	if id, ok := m.X.(*ast.Ident); ok && id.Name == "log" && c.scope.lookup("log") == nil {
		if m.Name != "info" && m.Name != "error" {
			c.fail(m.NamePos, "log has no method %q (available: info, error)", m.Name)
		}
		for _, a := range call.Args {
			c.exprVal(a, nil)
		}
		return types.VoidType
	}

	recv := c.checkExpr(m.X, nil)

	if types.Equal(recv, types.AnyType) {
		c.fail(m.NamePos, "cannot call methods on %q — narrow the type first", recv)
	}

	if b, ok := recv.(*types.Basic); ok && b.K == types.Str {
		return c.checkStrMethod(recv, m, call)
	}
	if arr, ok := recv.(*types.Array); ok {
		return c.checkArrayMethod(arr, m, call)
	}
	if _, ok := recv.(*types.Map); ok {
		return c.checkMapMethod(recv, m, call)
	}
	c.fail(m.NamePos, "type %q has no method %q", recv, m.Name)
	return nil
}

func (c *Checker) nargs(call *ast.CallExpr, want int) {
	if len(call.Args) != want {
		c.fail(call.P, "expected %d argument(s), got %d", want, len(call.Args))
	}
}

func (c *Checker) checkStrMethod(recv types.Type, m *ast.MemberExpr, call *ast.CallExpr) types.Type {
	switch m.Name {
	case "trim", "toUpper":
		c.nargs(call, 0)
		return types.StrType
	case "contains":
		c.nargs(call, 1)
		t := c.exprVal(call.Args[0], nil)
		if !types.Equal(t, types.StrType) {
			c.fail(call.Args[0].Pos(), "contains expects a str argument, got %q", t)
		}
		return types.BoolType
	case "len":
		c.nargs(call, 0)
		return types.IntType
	}
	c.fail(m.NamePos, "str has no method %q (available: trim, toUpper, contains, len)", m.Name)
	return nil
}

func (c *Checker) checkArrayMethod(recvArr *types.Array, m *ast.MemberExpr, call *ast.CallExpr) types.Type {
	mutRecv := func() {
		id, ok := m.X.(*ast.Ident)
		if !ok {
			c.fail(m.P, "the receiver of %q must be a variable", m.Name)
		}
		info := c.scope.lookup(id.Name)
		if info == nil || !info.mut {
			c.fail(m.P, "cannot modify immutable array %q (declare it with 'mut')", id.Name)
		}
	}

	switch m.Name {
	case "add":
		c.nargs(call, 1)
		t := c.exprVal(call.Args[0], nil)
		if !types.Equal(t, recvArr.Elem) {
			c.fail(call.Args[0].Pos(), "cannot add %q to array of %q", t, recvArr.Elem)
		}
		mutRecv()
		return types.VoidType
	case "remove":
		c.nargs(call, 1)
		t := c.exprVal(call.Args[0], nil)
		if !types.Equal(t, types.IntType) {
			c.fail(call.Args[0].Pos(), "remove expects an int index, got %q", t)
		}
		mutRecv()
		return types.VoidType
	case "len":
		c.nargs(call, 0)
		return types.IntType
	}
	c.fail(m.NamePos, "array has no method %q (available: add, remove, len)", m.Name)
	return nil
}

func (c *Checker) checkMapMethod(recv types.Type, m *ast.MemberExpr, call *ast.CallExpr) types.Type {
	if m.Name == "len" {
		c.nargs(call, 0)
		return types.IntType
	}
	c.fail(m.NamePos, "map has no method %q (available: len)", m.Name)
	return nil
}
