package codegen

import (
	"bytes"
	"fmt"
	"go/format"
	"strconv"
	"strings"

	"github.com/ScarfazeCompany/scar-script/internal/ast"
	"github.com/ScarfazeCompany/scar-script/internal/checker"
	"github.com/ScarfazeCompany/scar-script/internal/diag"
	"github.com/ScarfazeCompany/scar-script/internal/token"
	"github.com/ScarfazeCompany/scar-script/internal/types"
)

// Generator emits a self-contained Go program from a checked ScarScript AST.
type Generator struct {
	buf         *bytes.Buffer
	ind         int
	file        string
	typ         map[ast.Expr]types.Type
	unused      map[ast.Node][]string
	needStrings bool
	needMath    bool
	needRand    bool
	needTime    bool
}

// Generate translates a checked program into gofmt-formatted Go source.
func Generate(prog *ast.Program, res *checker.Result) (string, *diag.Error) {
	g := &Generator{
		buf:    new(bytes.Buffer),
		file:   prog.File,
		typ:    res.Types,
		unused: res.Unused,
	}
	g.generate(prog) // pass 1: collect flags (e.g. needStrings)
	g.buf.Reset()
	g.generate(prog) // pass 2: emit with correct imports
	src := g.buf.Bytes()
	out, err := format.Source(src)
	if err != nil {
		return "", diag.New(diag.Position{File: g.file, Line: 1, Col: 1},
			"internal error: generated Go code does not compile: %v", err)
	}
	return string(out), nil
}

func (g *Generator) line(format string, args ...any) {
	g.buf.WriteString(strings.Repeat("\t", g.ind))
	fmt.Fprintf(g.buf, format, args...)
	g.buf.WriteByte('\n')
}

func (g *Generator) raw(s string) {
	g.buf.WriteString(s)
}

// capture runs fn with output redirected to a fresh buffer and returns
// what was written (used for multi-line function literals).
func (g *Generator) capture(fn func()) string {
	old := g.buf
	g.buf = new(bytes.Buffer)
	fn()
	out := g.buf.String()
	g.buf = old
	return out
}

func (g *Generator) generate(prog *ast.Program) {
	g.raw("package main\n\n")

	g.raw("import (\n")
	g.raw("\t_scarfmt \"fmt\"\n")
	g.raw("\t_scarlog \"log\"\n")
	g.raw("\t_scaros \"os\"\n")
	g.raw("\t_scarreflect \"reflect\"\n")
	g.raw("\t_scarstrconv \"strconv\"\n")
	if g.needStrings {
		g.raw("\t_scarstrings \"strings\"\n")
	}
	if g.needMath {
		g.raw("\t_scarmath \"math\"\n")
	}
	if g.needRand {
		g.raw("\t_scarrand \"math/rand\"\n")
	}
	if g.needTime {
		g.raw("\t_scartime \"time\"\n")
	}
	g.raw(")\n\n")

	g.raw(prelude)
	g.emitOptionalHelpers()

	// Top-level functions are emitted before main.
	for _, s := range prog.Stmts {
		if fd, ok := s.(*ast.FuncDecl); ok {
			g.funcDecl(fd)
		}
	}

	g.raw("\nfunc main() {\n")
	g.ind++
	for _, s := range prog.Stmts {
		if _, ok := s.(*ast.FuncDecl); ok {
			continue
		}
		g.stmt(s)
	}
	g.ind--
	g.raw("}\n")
}

// emitOptionalHelpers adds helpers that depend on conditionally imported
// packages (flags are collected during pass 1).
func (g *Generator) emitOptionalHelpers() {
	if g.needTime {
		g.raw(scarTimeHelpers)
	}
	if g.needRand {
		g.raw(scarRandHelper)
	}
}

// prelude is emitted verbatim (already formatted with tabs).
const prelude = `
func _scar_str(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case int:
		return _scarstrconv.Itoa(x)
	case bool:
		return _scarstrconv.FormatBool(x)
	case nil:
		return ""
	default:
		return _scarfmt.Sprint(v)
	}
}

func _scar_or[T comparable](left, right T) T {
	var zero T
	if left == zero {
		return right
	}
	return left
}

func _scar_or_default[T any](left any, right T) T {
	if v, ok := left.(T); ok && !_scar_empty(v) {
		return v
	}
	return right
}

func _scar_empty(v any) bool {
	if v == nil {
		return true
	}
	switch x := v.(type) {
	case string:
		return x == ""
	case int:
		return x == 0
	case bool:
		return !x
	}
	rv := _scarreflect.ValueOf(v)
	switch rv.Kind() {
	case _scarreflect.Slice, _scarreflect.Map:
		return rv.Len() == 0
	}
	return false
}

func _scar_field(m any, key string) any {
	if m == nil {
		return nil
	}
	rv := _scarreflect.ValueOf(m)
	if rv.Kind() != _scarreflect.Map || rv.Type().Key().Kind() != _scarreflect.String {
		panic(_scarfmt.Sprintf("scar: cannot read field %q from %T", key, m))
	}
	mv := rv.MapIndex(_scarreflect.ValueOf(key))
	if !mv.IsValid() {
		return nil
	}
	return mv.Interface()
}

func _scar_add[T any](s []T, v T) []T {
	return append(s, v)
}

func _scar_remove[T any](s []T, i int) []T {
	if i < 0 || i >= len(s) {
		panic(_scarfmt.Sprintf("scar: index %d out of range (len %d)", i, len(s)))
	}
	return append(s[:i], s[i+1:]...)
}

func _scar_index[T any](s []T, i int) T {
	if i < 0 || i >= len(s) {
		panic(_scarfmt.Sprintf("scar: index %d out of range (len %d)", i, len(s)))
	}
	return s[i]
}

var (
	_scar_logi = _scarlog.New(_scaros.Stderr, "[INFO] ", _scarlog.LstdFlags)
	_scar_loge = _scarlog.New(_scaros.Stderr, "[ERROR] ", _scarlog.LstdFlags)
)

func _scar_log_info(v ...any)  { _scar_logi.Println(v...) }
func _scar_log_error(v ...any) { _scar_loge.Println(v...) }

func _scar_abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func _scar_round(x float64) int {
	if x >= 0 {
		return int(x + 0.5)
	}
	return int(x - 0.5)
}

func _scar_file_write(path string, data string) {
	if err := _scaros.WriteFile(path, []byte(data), 0644); err != nil {
		panic(_scarfmt.Sprintf("scar: file.write: %v", err))
	}
}

func _scar_file_append(path string, data string) {
	f, err := _scaros.OpenFile(path, _scaros.O_APPEND|_scaros.O_CREATE|_scaros.O_WRONLY, 0644)
	if err != nil {
		panic(_scarfmt.Sprintf("scar: file.append: %v", err))
	}
	defer f.Close()
	if _, err := f.WriteString(data); err != nil {
		panic(_scarfmt.Sprintf("scar: file.append: %v", err))
	}
}

func _scar_file_read(path string) string {
	b, err := _scaros.ReadFile(path)
	if err != nil {
		panic(_scarfmt.Sprintf("scar: file.read: %v", err))
	}
	return string(b)
}

func _scar_file_exists(path string) bool {
	_, err := _scaros.Stat(path)
	return err == nil
}

func _scar_file_delete(path string) {
	if err := _scaros.Remove(path); err != nil {
		panic(_scarfmt.Sprintf("scar: file.delete: %v", err))
	}
}
`

// scarTimeHelpers is emitted only when time.* is used (needs the time import).
const scarTimeHelpers = `
func _scar_time_now() string {
	return _scartime.Now().Format("2006-01-02 15:04:05")
}

func _scar_time_sleep(ms int) {
	_scartime.Sleep(_scartime.Duration(ms) * _scartime.Millisecond)
}
`

// scarRandHelper is emitted only when math.rand is used.
const scarRandHelper = `
func _scar_rand(lo int, hi int) int {
	if hi < lo {
		panic(_scarfmt.Sprintf("scar: math.rand: hi %d is less than lo %d", hi, lo))
	}
	return lo + _scarrand.Intn(hi-lo+1)
}
`

// ---------- statements ----------

func (g *Generator) stmt(s ast.Stmt) {
	switch st := s.(type) {
	case *ast.DeclStmt:
		g.decl(st)
	case *ast.MultiDeclStmt:
		g.multiDecl(st)
	case *ast.AssignStmt:
		g.line("%s = %s", st.Name, g.expr(st.Value))
	case *ast.IncDecStmt:
		if st.Op == token.PLUSPLUS {
			g.line("%s++", st.Name)
		} else {
			g.line("%s--", st.Name)
		}
	case *ast.IfStmt:
		g.ifChain(st)
	case *ast.ForStmt:
		g.forStmt(st)
	case *ast.BlockStmt:
		g.block(st.Stmts)
	case *ast.ExprStmt:
		g.exprStmt(st)
	case *ast.ReturnStmt:
		g.returnStmt(st)
	}
}

func (g *Generator) isUnused(n ast.Node, name string) bool {
	for _, u := range g.unused[n] {
		if u == name {
			return true
		}
	}
	return false
}

func (g *Generator) decl(d *ast.DeclStmt) {
	kw := "var"
	if d.Kind == ast.Const {
		kw = "const"
	}
	if d.Type != nil {
		g.line("%s %s %s = %s", kw, d.Name, g.goTypeExpr(d.Type), g.expr(d.Value))
	} else {
		g.line("%s %s = %s", kw, d.Name, g.expr(d.Value))
	}
	if d.Kind != ast.Const && g.isUnused(d, d.Name) {
		g.line("_ = %s", d.Name)
	}
}

// multiDecl emits `a, b := f()` for tuple-returning calls.
func (g *Generator) multiDecl(d *ast.MultiDeclStmt) {
	g.line("%s := %s", strings.Join(d.Names, ", "), g.expr(d.Value))
	for _, n := range d.Names {
		if g.isUnused(d, n) {
			g.line("_ = %s", n)
		}
	}
}

func (g *Generator) returnStmt(r *ast.ReturnStmt) {
	if len(r.Values) == 0 {
		g.line("return")
		return
	}
	parts := make([]string, len(r.Values))
	for i, v := range r.Values {
		parts[i] = g.expr(v)
	}
	g.line("return %s", strings.Join(parts, ", "))
}

// funcDecl emits a top-level Go function.
func (g *Generator) funcDecl(fd *ast.FuncDecl) {
	g.line("func %s(%s)%s {", fd.Name, g.paramList(fd.Params), g.resultType(fd.Result))
	g.ind++
	for _, s := range fd.Body.Stmts {
		g.stmt(s)
	}
	g.ind--
	g.line("}")
}

func (g *Generator) paramList(params []ast.Param) string {
	parts := make([]string, len(params))
	for i, p := range params {
		parts[i] = p.Name + " " + g.goTypeExpr(p.Type)
	}
	return strings.Join(parts, ", ")
}

// resultType returns "" for void, " int" / " (int, bool)" otherwise.
func (g *Generator) resultType(te ast.TypeExpr) string {
	if te == nil {
		return ""
	}
	return " " + g.goTypeExpr(te)
}

func (g *Generator) block(stmts []ast.Stmt) {
	g.line("{")
	g.ind++
	for _, s := range stmts {
		g.stmt(s)
	}
	g.ind--
	g.line("}")
}

func (g *Generator) ifChain(s *ast.IfStmt) {
	g.line("if %s {", g.expr(s.Cond))
	g.ind++
	for _, st := range s.Then.Stmts {
		g.stmt(st)
	}
	g.ind--
	for {
		if s.Else == nil {
			g.line("}")
			return
		}
		if nxt, ok := s.Else.(*ast.IfStmt); ok {
			g.line("} else if %s {", g.expr(nxt.Cond))
			g.ind++
			for _, st := range nxt.Then.Stmts {
				g.stmt(st)
			}
			g.ind--
			s = nxt
			continue
		}
		blk := s.Else.(*ast.BlockStmt)
		g.line("} else {")
		g.ind++
		for _, st := range blk.Stmts {
			g.stmt(st)
		}
		g.ind--
		g.line("}")
		return
	}
}

func (g *Generator) forStmt(f *ast.ForStmt) {
	iter := g.expr(f.Iterable)
	if _, isLit := f.Iterable.(*ast.ArrayLit); isLit {
		iter = "(" + iter + ")"
	}
	if _, isLit := f.Iterable.(*ast.MapLit); isLit {
		iter = "(" + iter + ")"
	}

	switch f.Kind {
	case ast.ForCond:
		g.line("for %s {", iter)

	case ast.ForIterable:
		if rng, ok := f.Iterable.(*ast.RangeExpr); ok {
			lo, hi := g.expr(rng.Lo), g.expr(rng.Hi)
			g.line("for %s := %s; %s <= %s; %s++ {", f.Var, lo, f.Var, hi, f.Var)
		} else {
			switch g.typ[f.Iterable].(type) {
			case *types.Array:
				g.line("for _, %s := range %s {", f.Var, iter)
			default: // map
				g.line("for %s := range %s {", f.Var, iter)
			}
		}

	case ast.ForTwo:
		switch g.typ[f.Iterable].(type) {
		case *types.Array:
			g.line("for %s, %s := range %s {", f.Var, f.Var2, iter)
		default: // map
			g.line("for %s, %s := range %s {", f.Var, f.Var2, iter)
		}
	}

	g.ind++
	for _, st := range f.Body.Stmts {
		g.stmt(st)
	}
	if f.Var != "" && g.isUnused(f, f.Var) {
		g.line("_ = %s", f.Var)
	}
	if f.Var2 != "" && g.isUnused(f, f.Var2) {
		g.line("_ = %s", f.Var2)
	}
	g.ind--
	g.line("}")
}

func (g *Generator) exprStmt(s *ast.ExprStmt) {
	call, isCall := s.X.(*ast.CallExpr)
	ex := g.expr(s.X)
	if isCall && !g.isBareCall(call) {
		g.line("_ = %s", ex)
		return
	}
	g.raw(strings.Repeat("\t", g.ind))
	g.raw(ex)
	g.raw("\n")
}

// isBareCall reports whether the call can stand alone as a Go statement:
// Go allows discarding all results of a call expression, but `_ = f()`
// only works when f returns exactly one value.
func (g *Generator) isBareCall(c *ast.CallExpr) bool {
	t := g.typ[c]
	if t == nil {
		return false
	}
	if _, isTup := t.(*types.Tuple); isTup {
		return true
	}
	return t == types.VoidType
}

// ---------- expressions ----------

func (g *Generator) expr(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.IntLit:
		return strconv.FormatInt(x.Value, 10)
	case *ast.FloatLit:
		s := strconv.FormatFloat(x.Value, 'f', -1, 64)
		if !strings.Contains(s, ".") {
			s += ".0"
		}
		return s
	case *ast.BoolLit:
		if x.Value {
			return "true"
		}
		return "false"
	case *ast.Ident:
		return x.Name
	case *ast.StrLit:
		return g.strLit(x)
	case *ast.ArrayLit:
		return g.arrayLit(x)
	case *ast.MapLit:
		return g.mapLit(x)
	case *ast.BinaryExpr:
		return g.binary(x)
	case *ast.UnaryExpr:
		op := "-"
		if x.Op == token.BANG {
			op = "!"
		}
		inner := g.expr(x.X)
		if exprPrec(x.X) <= 8 {
			inner = "(" + inner + ")"
		}
		return op + inner
	case *ast.TernaryExpr:
		return g.ternary(x)
	case *ast.CallExpr:
		return g.call(x)
	case *ast.MemberExpr:
		return g.member(x)
	case *ast.IndexExpr:
		return g.index(x)
	case *ast.FuncLit:
		return g.capture(func() { g.funcLit(x) })
	case *ast.RangeExpr:
		return "0 /* internal: range outside for */"
	}
	return "/* internal: unknown expr */"
}

// funcLit emits an anonymous function as a Go function literal.
func (g *Generator) funcLit(f *ast.FuncLit) {
	g.line("func(%s)%s {", g.paramList(f.Params), g.resultType(f.Result))
	g.ind++
	for _, s := range f.Body.Stmts {
		g.stmt(s)
	}
	g.ind--
	g.line("}")
}

func (g *Generator) strLit(s *ast.StrLit) string {
	var pieces []string
	for _, p := range s.Parts {
		if p.Expr == nil {
			if p.Lit == "" {
				continue
			}
			pieces = append(pieces, strconv.Quote(p.Lit))
			continue
		}
		pieces = append(pieces, "_scar_str("+g.expr(p.Expr)+")")
	}
	if len(pieces) == 0 {
		return `""`
	}
	return strings.Join(pieces, " + ")
}

func (g *Generator) arrayLit(a *ast.ArrayLit) string {
	t, ok := g.typ[a].(*types.Array)
	if !ok {
		return "/* internal: unknown array type */"
	}
	var elems []string
	for _, el := range a.Elems {
		elems = append(elems, g.expr(el))
	}
	return "[]" + g.goType(t.Elem) + "{" + strings.Join(elems, ", ") + "}"
}

func (g *Generator) mapLit(m *ast.MapLit) string {
	t, ok := g.typ[m].(*types.Map)
	if !ok {
		return "/* internal: unknown map type */"
	}
	var entries []string
	for _, en := range m.Entries {
		entries = append(entries, g.expr(en.Key)+": "+g.expr(en.Val))
	}
	return "map[" + g.goType(t.Key) + "]" + g.goType(t.Val) + "{" + strings.Join(entries, ", ") + "}"
}

func (g *Generator) ternary(t *ast.TernaryExpr) string {
	ty := g.goType(g.typ[t])
	return "func() " + ty + " { if " + g.expr(t.Cond) + " { return " + g.expr(t.A) + " }; return " + g.expr(t.B) + " }()"
}

func exprPrec(e ast.Expr) int {
	b, ok := e.(*ast.BinaryExpr)
	if !ok {
		return 9
	}
	switch b.Op {
	case token.OR:
		return 1
	case token.AND:
		return 2
	case token.EQ, token.NEQ:
		return 3
	case token.LT, token.GT, token.LTE, token.GTE:
		return 4
	case token.PLUS, token.MINUS:
		return 6
	default:
		return 7
	}
}

func (g *Generator) operand(parent int, e ast.Expr, right bool) string {
	s := g.expr(e)
	p := exprPrec(e)
	if p < parent || (right && p == parent) {
		return "(" + s + ")"
	}
	return s
}

func (g *Generator) binary(b *ast.BinaryExpr) string {
	switch b.Op {
	case token.AND:
		return g.bin("&&", 2, b)
	case token.OR:
		lt := g.typ[b.L]
		if lt != nil && types.Equal(lt, types.BoolType) {
			return g.bin("||", 1, b)
		}
		res := g.goType(g.typ[b])
		if lt != nil && types.Equal(lt, types.AnyType) {
			return "_scar_or_default[" + res + "](" + g.expr(b.L) + ", " + g.expr(b.R) + ")"
		}
		return "_scar_or[" + res + "](" + g.expr(b.L) + ", " + g.expr(b.R) + ")"
	case token.PLUS, token.MINUS, token.STAR, token.SLASH, token.PERCENT,
		token.EQ, token.NEQ, token.LT, token.GT, token.LTE, token.GTE:
		return g.bin(string(b.Op), exprPrec(b), b)
	}
	return "/* internal: unknown operator */"
}

func (g *Generator) bin(op string, prec int, b *ast.BinaryExpr) string {
	l := g.operand(prec, b.L, false)
	r := g.operand(prec, b.R, true)
	// Mixed int/dec operands: promote the int side to float64 so Go
	// accepts the expression. (`%` and string ops never mix types.)
	lt, rt := g.typ[b.L], g.typ[b.R]
	if lt != nil && rt != nil {
		if types.Equal(lt, types.IntType) && types.Equal(rt, types.DecType) {
			l = "float64(" + l + ")"
		} else if types.Equal(lt, types.DecType) && types.Equal(rt, types.IntType) {
			r = "float64(" + r + ")"
		}
	}
	return l + " " + op + " " + r
}

func (g *Generator) call(c *ast.CallExpr) string {
	if id, ok := c.Callee.(*ast.Ident); ok {
		if id.Name == "print" {
			return "_scarfmt.Println(" + g.argList(c.Args) + ")"
		}
		// user function or variable holding a function
		return id.Name + "(" + g.argList(c.Args) + ")"
	}
	m, ok := c.Callee.(*ast.MemberExpr)
	if !ok {
		return "/* internal: not callable */"
	}
	if id, isIdent := m.X.(*ast.Ident); isIdent && id.Name == "log" {
		fn := "_scar_log_info"
		if m.Name == "error" {
			fn = "_scar_log_error"
		}
		return fn + "(" + g.argList(c.Args) + ")"
	}
	if id, isIdent := m.X.(*ast.Ident); isIdent && isModule(id.Name) {
		return g.moduleCall(id.Name, m.Name, g.argList(c.Args))
	}

	recv := g.expr(m.X)
	args := g.argList(c.Args)

	switch t := g.typ[m.X].(type) {
	case *types.Basic:
		switch t.K {
		case types.Str:
			switch m.Name {
			case "trim":
				g.needStrings = true
				return "_scarstrings.TrimSpace(" + recv + ")"
			case "toUpper":
				g.needStrings = true
				return "_scarstrings.ToUpper(" + recv + ")"
			case "contains":
				g.needStrings = true
				return "_scarstrings.Contains(" + recv + ", " + args + ")"
			case "len":
				return "len(" + recv + ")"
			}
		}
	case *types.Array:
		switch m.Name {
		case "add":
			return recv + " = _scar_add(" + recv + ", " + args + ")"
		case "remove":
			return recv + " = _scar_remove(" + recv + ", " + args + ")"
		case "len":
			return "len(" + recv + ")"
		}
	case *types.Map:
		if m.Name == "len" {
			return "len(" + recv + ")"
		}
	}
	return "/* internal: unknown method */"
}

func (g *Generator) argList(args []ast.Expr) string {
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = g.expr(a)
	}
	return strings.Join(parts, ", ")
}

func isModule(name string) bool {
	return name == "math" || name == "file" || name == "sys" || name == "time"
}

// moduleCall maps a built-in module function to its Go equivalent.
func (g *Generator) moduleCall(mod, name, args string) string {
	switch mod {
	case "math":
		switch name {
		case "sqrt":
			g.needMath = true
			return "_scarmath.Sqrt(" + args + ")"
		case "pow":
			g.needMath = true
			return "_scarmath.Pow(" + args + ")"
		case "abs":
			return "_scar_abs(" + args + ")"
		case "round":
			return "_scar_round(" + args + ")"
		case "rand":
			g.needRand = true
			return "_scar_rand(" + args + ")"
		}
	case "file":
		switch name {
		case "write", "append", "read", "exists", "delete":
			return "_scar_file_" + name + "(" + args + ")"
		}
	case "sys":
		if name == "env" {
			return "_scaros.Getenv(" + args + ")"
		}
	case "time":
		switch name {
		case "now":
			g.needTime = true
			return "_scar_time_now()"
		case "sleep":
			g.needTime = true
			return "_scar_time_sleep(" + args + ")"
		}
	}
	return "/* internal: unknown module function */"
}

func (g *Generator) member(m *ast.MemberExpr) string {
	base := g.expr(m.X)
	switch g.typ[m.X].(type) {
	case *types.Basic:
		return "_scar_field(" + base + ", " + strconv.Quote(m.Name) + ")"
	case *types.Map:
		return g.atomBase(m.X, base) + "[" + strconv.Quote(m.Name) + "]"
	}
	return "/* internal: unknown member */"
}

func (g *Generator) index(x *ast.IndexExpr) string {
	base := g.expr(x.X)
	idx := g.expr(x.Index)
	switch g.typ[x.X].(type) {
	case *types.Array:
		return "_scar_index[" + g.goType(g.typ[x]) + "](" + base + ", " + idx + ")"
	case *types.Map:
		return g.atomBase(x.X, base) + "[" + idx + "]"
	}
	return "/* internal: cannot index */"
}

// atomBase parenthesizes composite literals so that `T{...}[k]` parses.
func (g *Generator) atomBase(e ast.Expr, s string) string {
	switch e.(type) {
	case *ast.MapLit, *ast.ArrayLit, *ast.BinaryExpr, *ast.UnaryExpr:
		return "(" + s + ")"
	}
	return s
}

// ---------- Go type names ----------

func (g *Generator) goType(t types.Type) string {
	switch v := t.(type) {
	case *types.Basic:
		switch v.K {
		case types.Str:
			return "string"
		case types.Int:
			return "int"
		case types.Bool:
			return "bool"
		case types.Float:
			return "float64"
		default:
			return "any"
		}
	case *types.Array:
		return "[]" + g.goType(v.Elem)
	case *types.Map:
		return "map[" + g.goType(v.Key) + "]" + g.goType(v.Val)
	case *types.Func:
		parts := make([]string, len(v.Params))
		for i, p := range v.Params {
			parts[i] = g.goType(p)
		}
		res := ""
		if !types.Equal(v.Result, types.VoidType) {
			res = " " + g.goType(v.Result)
		}
		return "func(" + strings.Join(parts, ", ") + ")" + res
	case *types.Tuple:
		parts := make([]string, len(v.Elems))
		for i, e := range v.Elems {
			parts[i] = g.goType(e)
		}
		return "(" + strings.Join(parts, ", ") + ")"
	}
	return "any"
}

func (g *Generator) goTypeExpr(te ast.TypeExpr) string {
	switch t := te.(type) {
	case *ast.BasicType:
		switch t.Kind {
		case "str":
			return "string"
		case "int":
			return "int"
		case "bool":
			return "bool"
		case "dec":
			return "float64"
		default:
			return "any"
		}
	case *ast.ArrayTypeExpr:
		return "[]" + g.goTypeExpr(t.Elem)
	case *ast.MapTypeExpr:
		return "map[" + g.goTypeExpr(t.Key) + "]" + g.goTypeExpr(t.Val)
	case *ast.TupleTypeExpr:
		parts := make([]string, len(t.Elems))
		for i, e := range t.Elems {
			parts[i] = g.goTypeExpr(e)
		}
		return "(" + strings.Join(parts, ", ") + ")"
	}
	return "any"
}
