package parser

import (
	"testing"

	"github.com/ScarfazeCompany/scar-script/internal/ast"
	"github.com/ScarfazeCompany/scar-script/internal/lexer"
)

func parseOK(t *testing.T, src string) *ast.Program {
	t.Helper()
	toks, lerr := lexer.Lex(src, "t.scar")
	if lerr != nil {
		t.Fatalf("lex error: %v", lerr)
	}
	prog, perr := Parse(toks, "t.scar")
	if perr != nil {
		t.Fatalf("parse error: %v", perr)
	}
	return prog
}

func parseErr(t *testing.T, src, msg string) {
	t.Helper()
	toks, lerr := lexer.Lex(src, "t.scar")
	if lerr != nil {
		t.Fatalf("lex error: %v", lerr)
	}
	_, perr := Parse(toks, "t.scar")
	if perr == nil {
		t.Fatalf("%q: expected error containing %q", src, msg)
	}
	if !contains(perr.Msg, msg) {
		t.Errorf("%q: err %q, want contains %q", src, perr.Msg, msg)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestParseDecl(t *testing.T) {
	prog := parseOK(t, "val tags: str[] = [\"a\", \"b\"]\nmut n: int = 0\nconst C: str = \"x\"\n")
	if len(prog.Stmts) != 3 {
		t.Fatalf("stmt count = %d", len(prog.Stmts))
	}
	d, ok := prog.Stmts[0].(*ast.DeclStmt)
	if !ok || d.Name != "tags" || d.Kind != ast.Val {
		t.Fatalf("stmt0 = %#v", prog.Stmts[0])
	}
	if _, ok := d.Type.(*ast.ArrayTypeExpr); !ok {
		t.Errorf("type = %#v", d.Type)
	}
	d2 := prog.Stmts[1].(*ast.DeclStmt)
	if d2.Kind != ast.Mut {
		t.Errorf("kind = %v", d2.Kind)
	}
	d3 := prog.Stmts[2].(*ast.DeclStmt)
	if d3.Kind != ast.Const {
		t.Errorf("kind = %v", d3.Kind)
	}
}

func TestParseAssignAndInc(t *testing.T) {
	prog := parseOK(t, "status = \"done\"\ncount++\n")
	a, ok := prog.Stmts[0].(*ast.AssignStmt)
	if !ok || a.Name != "status" {
		t.Fatalf("stmt0 = %#v", prog.Stmts[0])
	}
	i, ok := prog.Stmts[1].(*ast.IncDecStmt)
	if !ok || i.Name != "count" {
		t.Fatalf("stmt1 = %#v", prog.Stmts[1])
	}
}

func TestParseIfChain(t *testing.T) {
	prog := parseOK(t, "if a > 1 {\n  print(1)\n} else if b {\n  print(2)\n} else {\n  print(3)\n}\n")
	ifs, ok := prog.Stmts[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("stmt0 = %#v", prog.Stmts[0])
	}
	if len(ifs.Then.Stmts) != 1 {
		t.Errorf("then stmts = %d", len(ifs.Then.Stmts))
	}
	elseIf, ok := ifs.Else.(*ast.IfStmt)
	if !ok {
		t.Fatalf("else = %#v", ifs.Else)
	}
	if _, ok := elseIf.Else.(*ast.BlockStmt); !ok {
		t.Errorf("else-else = %#v", elseIf.Else)
	}
}

func TestParseForForms(t *testing.T) {
	prog := parseOK(t, `for i in 1..5 {
	print(i)
}
for tag in tags {
	print(tag)
}
for index, tag in tags {
	print(index)
}
for key, value in scores {
	print(key)
}
for count < 5 {
count++
}
`)
	if len(prog.Stmts) != 5 {
		t.Fatalf("stmt count = %d", len(prog.Stmts))
	}
	f1 := prog.Stmts[0].(*ast.ForStmt)
	if f1.Kind != ast.ForIterable || f1.Var != "i" {
		t.Errorf("f1 = %#v", f1)
	}
	if _, ok := f1.Iterable.(*ast.RangeExpr); !ok {
		t.Errorf("f1 iterable = %#v", f1.Iterable)
	}
	f3 := prog.Stmts[2].(*ast.ForStmt)
	if f3.Kind != ast.ForTwo || f3.Var != "index" || f3.Var2 != "tag" {
		t.Errorf("f3 = %#v", f3)
	}
	f5 := prog.Stmts[4].(*ast.ForStmt)
	if f5.Kind != ast.ForCond {
		t.Errorf("f5 kind = %v", f5.Kind)
	}
}

func TestPrecedence(t *testing.T) {
	prog := parseOK(t, "val x: int = 1 + 2 * 3\n")
	d := prog.Stmts[0].(*ast.DeclStmt)
	b, ok := d.Value.(*ast.BinaryExpr)
	if !ok || string(b.Op) != "+" {
		t.Fatalf("value = %#v", d.Value)
	}
	r, ok := b.R.(*ast.BinaryExpr)
	if !ok || string(r.Op) != "*" {
		t.Fatalf("rhs = %#v", b.R)
	}

	prog = parseOK(t, "val y: bool = a or b and c\n")
	d = prog.Stmts[0].(*ast.DeclStmt)
	b, ok = d.Value.(*ast.BinaryExpr)
	if !ok || string(b.Op) != "or" {
		t.Fatalf("value = %#v", d.Value)
	}
	if bb, ok := b.R.(*ast.BinaryExpr); !ok || string(bb.Op) != "and" {
		t.Errorf("rhs = %#v", b.R)
	}

	prog = parseOK(t, "val z: bool = a >= 18 ? true : false\n")
	d = prog.Stmts[0].(*ast.DeclStmt)
	if _, ok := d.Value.(*ast.TernaryExpr); !ok {
		t.Fatalf("value = %#v", d.Value)
	}

	prog = parseOK(t, "val w: int = 1..n + 1\n")
	d = prog.Stmts[0].(*ast.DeclStmt)
	rng, ok := d.Value.(*ast.RangeExpr)
	if !ok {
		t.Fatalf("value = %#v", d.Value)
	}
	if b, ok := rng.Hi.(*ast.BinaryExpr); !ok || string(b.Op) != "+" {
		t.Errorf("hi = %#v", rng.Hi)
	}
}

func TestParseStringsAndCollections(t *testing.T) {
	prog := parseOK(t, "val s: str = \"hi {name}!\"\nval m: map[str: int] = {\"a\": 1, b: 2}\nval e: map[str: int] = {}\n")
	d := prog.Stmts[0].(*ast.DeclStmt)
	s := d.Value.(*ast.StrLit)
	if len(s.Parts) != 3 || s.Parts[0].Lit != "hi " || s.Parts[1].Expr == nil || s.Parts[2].Lit != "!" {
		t.Errorf("parts = %#v", s.Parts)
	}
	d2 := prog.Stmts[1].(*ast.DeclStmt)
	m := d2.Value.(*ast.MapLit)
	if len(m.Entries) != 2 {
		t.Errorf("entries = %d", len(m.Entries))
	}
	d3 := prog.Stmts[2].(*ast.DeclStmt)
	if m3, ok := d3.Value.(*ast.MapLit); !ok || len(m3.Entries) != 0 {
		t.Errorf("empty map = %#v", d3.Value)
	}
}

func TestParseCallAndMember(t *testing.T) {
	prog := parseOK(t, "log.info(\"started\", port)\nitems.add(\"x\")\nprint(\"{a.b}\")\n")
	c := prog.Stmts[0].(*ast.ExprStmt).X.(*ast.CallExpr)
	m := c.Callee.(*ast.MemberExpr)
	if m.Name != "info" {
		t.Errorf("member = %q", m.Name)
	}
	if len(c.Args) != 2 {
		t.Errorf("args = %d", len(c.Args))
	}
	c2 := prog.Stmts[1].(*ast.ExprStmt).X.(*ast.CallExpr)
	if _, ok := c2.Callee.(*ast.MemberExpr); !ok {
		t.Errorf("callee = %#v", c2.Callee)
	}
}

func TestParseMultilineCall(t *testing.T) {
	parseOK(t, "print(\n  \"a\",\n  \"b\",\n)\n")
}

func TestParseErrors(t *testing.T) {
	cases := []struct{ src, msg string }{
		{"val x int = 1\n", "expected \"=\""},
		{"val x: int 5\n", "expected \"=\""},
		{"func f {}\n", "expected \"(\""},
		{"func f() \n", "expected type"},
		{"if x {\n", "unclosed"},
		{"print(1) print(2)\n", "expected newline"},
		{"val 5: int = 1\n", "expected identifier"},
		{"x = \n", "expected expression"},
		{"val x: frob = 1\n", "expected type"},
		{"for {\n", "expected map key"},
	}
	for _, c := range cases {
		parseErr(t, c.src, c.msg)
	}
}

func TestErrorPositions(t *testing.T) {
	toks, _ := lexer.Lex("val x: int = 1\nval y: = 2\n", "t.scar")
	_, perr := Parse(toks, "t.scar")
	if perr == nil {
		t.Fatal("expected error")
	}
	if perr.Pos.Line != 2 || perr.Pos.Col != 8 {
		t.Errorf("pos = %d:%d, want 2:8", perr.Pos.Line, perr.Pos.Col)
	}
}

func TestParseFuncDecl(t *testing.T) {
	prog := parseOK(t, "func add(a: int, b: int) int {\nreturn a + b\n}\n")
	fd, ok := prog.Stmts[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("stmt0 = %#v", prog.Stmts[0])
	}
	if fd.Name != "add" || len(fd.Params) != 2 {
		t.Errorf("name = %q, params = %d", fd.Name, len(fd.Params))
	}
	if fd.Params[0].Name != "a" || fd.Result == nil {
		t.Errorf("params/result = %#v", fd)
	}
	ret := fd.Body.Stmts[0].(*ast.ReturnStmt)
	if len(ret.Values) != 1 {
		t.Errorf("return values = %d", len(ret.Values))
	}
}

func TestParseFuncVoidAndArrow(t *testing.T) {
	prog := parseOK(t, "func greet(name: str) {\nprint(name)\n}\nfunc f() -> int {\nreturn 1\n}\n")
	fd := prog.Stmts[0].(*ast.FuncDecl)
	if fd.Result != nil {
		t.Errorf("void result = %#v", fd.Result)
	}
	fd2 := prog.Stmts[1].(*ast.FuncDecl)
	if fd2.Result == nil {
		t.Errorf("arrow result missing")
	}
}

func TestParseTupleResult(t *testing.T) {
	prog := parseOK(t, "func divide(a: int, b: int) (int, bool) {\nreturn 0, false\n}\n")
	fd := prog.Stmts[0].(*ast.FuncDecl)
	tup, ok := fd.Result.(*ast.TupleTypeExpr)
	if !ok || len(tup.Elems) != 2 {
		t.Fatalf("result = %#v", fd.Result)
	}
	ret := fd.Body.Stmts[0].(*ast.ReturnStmt)
	if len(ret.Values) != 2 {
		t.Errorf("return values = %d", len(ret.Values))
	}
}

func TestParseFuncLitAndMultiDecl(t *testing.T) {
	prog := parseOK(t, "val mul = func(x: int, y: int) int {\nreturn x * y\n}\nval a, b = f()\n")
	d := prog.Stmts[0].(*ast.DeclStmt)
	if d.Type != nil {
		t.Errorf("annotation should be nil for inference, got %#v", d.Type)
	}
	if _, ok := d.Value.(*ast.FuncLit); !ok {
		t.Errorf("value = %#v", d.Value)
	}
	md := prog.Stmts[1].(*ast.MultiDeclStmt)
	if len(md.Names) != 2 || md.Names[1] != "b" {
		t.Errorf("names = %v", md.Names)
	}
}

func TestParseBareReturn(t *testing.T) {
	prog := parseOK(t, "func f() {\nreturn\n}\n")
	fd := prog.Stmts[0].(*ast.FuncDecl)
	ret := fd.Body.Stmts[0].(*ast.ReturnStmt)
	if len(ret.Values) != 0 {
		t.Errorf("values = %d, want 0", len(ret.Values))
	}
}

func TestParseInferredDecl(t *testing.T) {
	prog := parseOK(t, "val x = 5\nval s = \"hi\"\nval d = 4.7\n")
	d := prog.Stmts[0].(*ast.DeclStmt)
	if d.Type != nil || d.Name != "x" {
		t.Errorf("decl = %#v", d)
	}
	if _, ok := prog.Stmts[2].(*ast.DeclStmt).Value.(*ast.FloatLit); !ok {
		t.Errorf("value = %#v", prog.Stmts[2].(*ast.DeclStmt).Value)
	}
}
