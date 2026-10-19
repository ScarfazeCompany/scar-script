package checker

import (
	"strings"
	"testing"

	"github.com/ScarfazeCompany/scar-script/internal/ast"
	"github.com/ScarfazeCompany/scar-script/internal/diag"
	"github.com/ScarfazeCompany/scar-script/internal/lexer"
	"github.com/ScarfazeCompany/scar-script/internal/parser"
)

func check(t *testing.T, src string) *diag.Error {
	t.Helper()
	toks, lerr := lexer.Lex(src, "t.scar")
	if lerr != nil {
		t.Fatalf("lex error: %v", lerr)
	}
	prog, perr := parser.Parse(toks, "t.scar")
	if perr != nil {
		t.Fatalf("parse error: %v", perr)
	}
	_, err := Check(prog)
	return err
}

func checkOK(t *testing.T, src string) {
	t.Helper()
	if err := check(t, src); err != nil {
		t.Errorf("unexpected error:\n%s\nsource:\n%s", err, src)
	}
}

func checkErr(t *testing.T, src, msg string) {
	t.Helper()
	err := check(t, src)
	if err == nil {
		t.Errorf("expected error containing %q, got nil\nsource:\n%s", msg, src)
		return
	}
	if !strings.Contains(err.Msg, msg) {
		t.Errorf("err %q, want contains %q\nsource:\n%s", err.Msg, msg, src)
	}
}

const specProgram = `
mut status: str = "pending"
status = "completed"
const API_URL: str = "https://api.scar.com"
val tags: str[] = ["fast", "backend", "express"]
val scores: map[str: int] = {"zawiyar": 100, "kashif": 95}
val rawData: any = { "status": 200 }
val name: str = "Zawiyar"
val age: int = 25
print("Hello World")
print("User {name} ki age {age} hai.")
print(name, age, true)
log.info("Server started on port 8080")
log.error("Database connection failed")
val isAdult: bool = age >= 18 ? true : false
val fallback: str = rawData.status or "default_status"
if age >= 18 and isAdult {
    print("Access Granted")
} else if age < 18 {
    print("Underage")
} else {
    print("Access Denied")
}
for i in 1..5 {
    print("Count: {i}")
}
for tag in tags {
    print("Tag: {tag}")
}
for index, tag in tags {
    print("{index}: {tag}")
}
for key, value in scores {
    print("{key} scored {value}")
}
mut count: int = 0
for count < 5 {
    print(count)
    count++
}
val clean: str = name.trim()
val upper: str = name.toUpper()
val isMatch: bool = name.contains("Zaw")
mut items: str[] = ["apple", "banana"]
items.add("orange")
items.remove(0)
val length: int = items.len()
`

func TestSpecProgramTypeChecks(t *testing.T) {
	checkOK(t, specProgram)
}

func TestValidPrograms(t *testing.T) {
	cases := []string{
		"val x: int = 1 + 2 * 3\n",
		"val s: str = \"a\" + \"b\"\n",
		"const N: int = -5\n",
		"val b: bool = !(1 > 2)\n",
		"val a: any = 5\nval b: int = a or 7\n",
		"val m: map[str: int] = {}\nprint(m.len())\n",
		"val xs: any = []\n",
		"val m: any = {}\n",
		"if true {\nval x: int = 1\nprint(x)\n}\n",
	}
	for _, src := range cases {
		checkOK(t, src)
	}
	checkErr(t, "val x: int = 1..5\n", "range expressions can only be used in 'for' loops")
}

func TestImmutability(t *testing.T) {
	checkErr(t, "val x: int = 1\nx = 2\n", "immutable variable")
	checkErr(t, "const C: int = 1\nC = 2\n", "constant")
	checkErr(t, "val x: int = 1\nx++\n", "immutable variable")
	checkErr(t, "val xs: str[] = [\"a\"]\nxs.add(\"b\")\n", "immutable array")
	checkOK(t, "mut x: int = 1\nx = 2\nx++\n")
	checkOK(t, "mut xs: str[] = [\"a\"]\nxs.add(\"b\")\nxs.remove(0)\n")
}

func TestTypeErrors(t *testing.T) {
	cases := []struct{ src, msg string }{
		{"val x: int = \"hi\"\n", "cannot use value of type"},
		{"val x: int = 1\nval y: str = x\n", "cannot use value of type"},
		{"mut s: str = \"a\"\ns = 5\n", "cannot use value of type"},
		{"val x: int = print(\"a\")\n", "void"},
		{"val x: any = 1\nval y: str = x\n", "cannot use value of type"},
		{"if 5 {\nprint(1)\n}\n", "condition must be bool"},
		{"val a: int = 1\nval b: str = \"x\"\nval c: bool = a and b\n", "'and' requires bool"},
		{"val a: int = 1\nval b: str = \"x\"\nval c: str = a + b\n", "'+' is not defined"},
		{"val c: bool = 1 > \"a\"\n", "not defined"},
		{"val c: bool = 1 == \"a\"\n", "cannot compare"},
		{"val x: str[] = [\"a\", 1]\n", "same type"},
		{"print([])\n", "cannot infer"},
		{"val m: map[str: int] = {a: 1, b: \"x\"}\n", "same type"},
		{"val x: map[bool: int] = {a: 1}\n", "map key must be"},
		{"val b: bool = 1..2\n", "range expressions"},
		{"print(1..2)\n", "range expressions"},
		{"undefinedVar\n", "undefined variable"},
		{"undefinedVar()\n", "undefined function"},
		{"val x: int = 1\nx\n", "expression has no effect"},
		{"val x: int = 1\nx.foo()\n", "has no method"},
		{"val x: str = \"a\"\nx.add(1)\n", "str has no method"},
		{"val x: str = \"a\"\nprint(x.trim)\n", "must be called with"},
		{"val x: any = {}\nx.foo()\n", "cannot call methods on"},
		{"val x: int = 5()\n", "not callable"},
		{"print(\"a\") + 1\n", "cannot use a void value"},
		{"val x: str[] = [\"a\"]\nx.add(5)\n", "cannot add"},
		{"mut xs: str[] = [\"a\"]\nxs.remove(\"b\")\n", "remove expects an int"},
		{"val n: int = 5\nn.trim()\n", "has no method"},
		{"val m: map[str: int] = {a: 1}\nm.trim()\n", "map has no method"},
		{"const C: str = \"a\" + \"b\"\n", "const requires a literal"},
		{"const C: str[] = [\"a\"]\n", "const must be of type"},
		{"print()\nprint(\"{print(1)}\")\n", "cannot interpolate a void value"},
		{"val x: int = 1\nval x: int = 2\n", "already declared"},
		{"val print: int = 1\n", "builtin name"},
		{"val s: str = \"a\"\nfor s {\nprint(1)\n}\n", "loop condition must be bool"},
		{"val n: int = 5\nfor n {\nprint(1)\n}\n", "loop condition must be bool"},
		{"for i in 5 {\nprint(i)\n}\n", "cannot iterate over"},
		{"for a, b in 5 {\nprint(a)\n}\n", "cannot iterate with two variables"},
		{"mut m: map[str: int] = {a: 1}\nm[1]\n", "map index must be"},
	}
	for _, c := range cases {
		checkErr(t, c.src, c.msg)
	}
}

func TestUndefinedAndReserved(t *testing.T) {
	checkErr(t, "val x: int = y\n", "undefined variable")
	checkErr(t, "val x: int = print\n", "'print' is a builtin")
	checkErr(t, "val x: int = log\n", "'log' is a builtin")
	checkErr(t, "log.info(1)\nlog.warn(1)\n", "log has no method")
	checkErr(t, "log()\n", "builtin")
	checkErr(t, "mut x: int = 1\nx = print(1)\n", "void")
}

func TestScoping(t *testing.T) {
	checkErr(t, "if true {\nval x: int = 1\n}\nprint(x)\n", "undefined variable")
	checkOK(t, "val x: int = 1\nif true {\nval x: str = \"a\"\nprint(x)\n}\nprint(x)\n")
	checkErr(t, "if true {\nval x: int = 1\nval x: int = 2\n}\n", "already declared")
	checkOK(t, "for i in 1..3 {\nval i: int = 5\nprint(i)\n}\n")
}

func TestForLoopVars(t *testing.T) {
	checkOK(t, "for i in 1..3 {\ni++\nprint(i)\n}\n")
	checkOK(t, "val xs: int[] = [1]\nfor a, b in xs {\nprint(a, b)\n}\n")
	checkOK(t, "val m: map[str: int] = {a: 1}\nfor k in m {\nprint(k)\n}\n")
	checkErr(t, "val xs: int[] = [1]\nfor a, a in xs {\nprint(a)\n}\n", "already declared")
	checkOK(t, "mut count: int = 0\nfor count < 3 {\ncount++\n}\n")
}

func TestMemberAccessOnAnyAndMap(t *testing.T) {
	checkOK(t, "val raw: any = {status: 200}\nval s: str = raw.status or \"ok\"\n")
	checkOK(t, "val m: map[str: int] = {a: 1}\nprint(m.a)\n")
	checkOK(t, "val m: map[str: int] = {a: 1}\nprint(m[\"a\"])\n")
	checkErr(t, "val m: map[str: int] = {a: 1}\nprint(m[1])\n", "map index must be")
	checkErr(t, "val xs: int[] = [1]\nprint(xs[\"a\"])\n", "array index must be int")
	checkErr(t, "val xs: int[] = [1]\nprint(xs.foo)\n", "has no field")
}

func TestErrorPositionInChecker(t *testing.T) {
	err := check(t, "val x: int = 1\nval y: str = x\n")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Pos.Line != 2 || err.Pos.Col != 14 {
		t.Errorf("pos = %d:%d, want 2:14", err.Pos.Line, err.Pos.Col)
	}
}

func TestReservedNames(t *testing.T) {
	checkErr(t, "val type: int = 1\n", "reserved word")
	checkErr(t, "val range: int = 1\n", "reserved word")
	checkErr(t, "val _hidden: int = 1\n", "cannot start with '_'")
	checkOK(t, "val fmt: int = 1\nprint(fmt)\n")
}

func TestUnusedTracking(t *testing.T) {
	res, err := Check(mustParse(t, "val used: int = 1\nval unused: int = 2\nprint(used)\nconst C: int = 3\nmut m: int = 0\nm = 5\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := map[string]bool{}
	for _, names := range res.Unused {
		for _, n := range names {
			got[n] = true
		}
	}
	if len(got) != 2 || !got["unused"] || !got["m"] {
		t.Errorf("unused = %v, want {unused, m}", got)
	}
}

func TestFunctions(t *testing.T) {
	checkOK(t, "func add(a: int, b: int) int {\nreturn a + b\n}\nval x: int = add(1, 2)\nprint(x)\n")
	checkOK(t, "func greet(name: str) {\nprint(\"Hello, {name}!\")\n}\ngreet(\"Zawiyar\")\n")
	checkOK(t, "func f() int {\nreturn 1\n}\nfunc g() int {\nreturn f()\n}\nprint(g())\n")
	// forward reference & recursion
	checkOK(t, "func fact(n: int) int {\nif n <= 1 {\nreturn 1\n}\nreturn n * fact(n - 1)\n}\nprint(fact(5))\n")
	checkOK(t, "func a() int {\nreturn b()\n}\nfunc b() int {\nreturn 7\n}\nprint(a())\n")
	// function can access top-level constants and immutable variables
	checkOK(t, "val x: int = 1\nfunc f() int {\nreturn x\n}\nprint(f())\n")
	// annotation optional via inference
	checkOK(t, "val x = 5\nprint(x)\n")
	checkOK(t, "val x = \"hi\"\nprint(x)\n")
}

func TestFunctionErrors(t *testing.T) {
	cases := []struct{ src, msg string }{
		{"func f() int {\nprint(1)\n}\n", "missing return"},
		{"func f() int {\nif true {\nreturn 1\n}\n}\n", "missing return"},
		{"func f() int {\nreturn 1\nprint(2)\n}\n", "unreachable code"},
		{"func f() int {\nreturn \"a\"\n}\n", "cannot return value of type"},
		{"func f() int {\nreturn\n}\n", "missing return value"},
		{"func f() {\nreturn 5\n}\n", "void function cannot return a value"},
		{"func f(a: int) int {\nreturn a\n}\nf(\"x\")\n", "argument 1 expects int, got str"},
		{"func f(a: int) int {\nreturn a\n}\nf()\n", "expected 1 argument(s), got 0"},
		{"func f(a: int, a: int) int {\nreturn a\n}\n", "duplicate parameter"},
		{"if true {\nfunc f() int {\nreturn 1\n}\n}\n", "top level"},
		{"func f() int {\nfunc g() int {\nreturn 1\n}\nreturn g()\n}\n", "top level"},
		{"return\n", "return outside of a function"},
		{"func f() int {\nreturn 1\n}\nval x: int = f\n", "is a function — call it"},
		{"func f() int {\nreturn 1\n}\nf\n", "is a function — call it"},
		{"val f: int = 1\nf()\n", "not a function"},
		{"func add(a: int) int {\nreturn a\n}\nval add: int = 1\n", "already declared as a function"},
	}
	for _, c := range cases {
		checkErr(t, c.src, c.msg)
	}
}

func TestMultipleReturns(t *testing.T) {
	const div = "func divide(a: int, b: int) (int, bool) {\nif b == 0 {\nreturn 0, false\n}\nreturn a / b, true\n}\n"
	checkOK(t, div+"val res, ok = divide(10, 2)\nprint(res, ok)\n")
	checkOK(t, div+"val a, b = divide(1, 2)\nval c, d = divide(3, 4)\nprint(a, b, c, d)\n")
	checkOK(t, div+"mut res, ok = divide(10, 2)\nres = 5\nprint(res, ok)\n")
	checkOK(t, div+"func pass() (int, bool) {\nreturn divide(4, 2)\n}\nval x, y = pass()\nprint(x, y)\n")

	checkErr(t, div+"val x = divide(1, 2)\n", "multiple values")
	checkErr(t, div+"val a, b, c = divide(1, 2)\n", "expected 2 return values, got 3")
	checkErr(t, div+"val a, b = divide(1, 2).foo\n", "multiple declarations require a function call")
	checkErr(t, div+"print(divide(1, 2))\n", "multiple values")
	checkErr(t, "func f() int {\nreturn 1, 2\n}\n", "function returns a single")
	checkErr(t, "func f() (int, int) {\nreturn 1\n}\n", "expected 2 return values, got 1")
	checkErr(t, "func f() (int, int) {\nreturn 1, \"a\"\n}\n", "return value 2 expects int")
	checkErr(t, "val a, b = 5\n", "multiple declarations require a function call")
	checkErr(t, "func f() int {\nreturn 1\n}\nconst a, b = f()\n", "const cannot declare multiple")
	checkErr(t, "val a: (int, int) = 1\n", "tuple types can only be used")
}

func TestAnonymousFunctions(t *testing.T) {
	checkOK(t, "val mul = func(x: int, y: int) int {\nreturn x * y\n}\nprint(mul(4, 5))\n")
	checkOK(t, "val dbl = func(x: int) int {\nreturn x * 2\n}\nval r: int = dbl(21)\nprint(r)\n")
	checkOK(t, "func apply(f: int) int {\nreturn f\n}\nprint(apply(3))\n")
	// closure capture
	checkOK(t, "val n: int = 10\nval addn = func(x: int) int {\nreturn x + n\n}\nprint(addn(5))\n")
	checkErr(t, "val f = func(x: int) int {\nreturn x\n}\nprint(f(\"a\"))\n", "argument 1 expects int")
	checkErr(t, "val x = func() {}\nval y: int = x\n", "cannot use value of type")
}

func TestDecType(t *testing.T) {
	checkOK(t, "val x: dec = 4.7\nprint(x)\n")
	checkOK(t, "val x: dec = 4\nprint(x)\n") // int literal widens
	checkOK(t, "val x: dec = -4\nprint(x)\n")
	checkOK(t, "val y: dec = 4.7 + 1\nprint(y)\n")
	checkOK(t, "val y: dec = 1 + 4.7\nprint(y)\n")
	checkOK(t, "val i: int = 3\nval y: dec = 4.7 + i\nprint(y)\n")
	checkOK(t, "val b: bool = 4.7 > 2\nprint(b)\n")
	checkErr(t, "val x: int = 4.7 - 0.7\n", "cannot use value of type")

	checkErr(t, "val x: int = 4.7\n", "cannot use value of type")
	checkErr(t, "val i: int = 3\nval d: dec = i\n", "cannot use value of type")
	checkErr(t, "val x: dec = \"a\"\n", "cannot use value of type")
	checkErr(t, "val x: dec = 4.7 % 2\n", "requires int operands")
	checkErr(t, "mut x: dec = 1.5\nx++\n", "requires an int variable")
	checkErr(t, "val x: dec = 1 + 2\n", "cannot use value of type")
}

func TestModuleCalls(t *testing.T) {
	checkOK(t, "val sq: dec = math.sqrt(16)\nprint(sq)\n")
	checkOK(t, "val pw: dec = math.pow(2, 3)\nprint(pw)\n")
	checkOK(t, "val a: int = math.abs(-45)\nprint(a)\n")
	checkOK(t, "val r: int = math.round(4.7)\nprint(r)\n")
	checkOK(t, "val n: int = math.rand(1, 100)\nprint(n)\n")
	checkOK(t, "file.write(\"t.txt\", \"data\")\nval c: str = file.read(\"t.txt\")\nprint(c)\n")
	checkOK(t, "file.append(\"t.txt\", \"more\")\n")
	checkOK(t, "if file.exists(\"t.txt\") {\nfile.delete(\"t.txt\")\n}\n")
	checkOK(t, "val p: str = sys.env(\"PORT\") or \"8080\"\nprint(p)\n")
	checkOK(t, "val now: str = time.now()\nprint(now)\n")
	checkOK(t, "time.sleep(1)\n")

	checkErr(t, "val x: dec = math.sqrt(\"a\")\n", "argument 1 expects dec, got str")
	checkErr(t, "val x: int = math.abs(1.5)\n", "argument 1 expects int, got dec")
	checkErr(t, "math.frob(1)\n", "math has no function")
	checkErr(t, "file.read(1)\n", "argument 1 expects str, got int")
	checkErr(t, "time.sleep(\"a\")\n", "argument 1 expects int, got str")
	checkErr(t, "val x: int = math\n", "is a module")
	checkErr(t, "math()\n", "is a module")
	checkErr(t, "val math: int = 1\n", "module name")
	checkErr(t, "func math(a: int) int {\nreturn a\n}\n", "module name")
}

func TestFloatAndFunctionCodegenRoundTrip(t *testing.T) {
	// full pipeline check happens in codegen/driver tests; here we only
	// ensure the checker resolves every expression type it needs.
	src := "func calc(p: dec) dec {\nreturn p * 1.5\n}\nval x: dec = calc(2.5)\nprint(x)\n"
	checkOK(t, src)
}

func mustParse(t *testing.T, src string) *ast.Program {
	t.Helper()
	toks, lerr := lexer.Lex(src, "t.scar")
	if lerr != nil {
		t.Fatalf("lex error: %v", lerr)
	}
	prog, perr := parser.Parse(toks, "t.scar")
	if perr != nil {
		t.Fatalf("parse error: %v", perr)
	}
	return prog
}
