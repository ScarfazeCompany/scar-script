package codegen

import (
	"strings"
	"testing"

	"github.com/ScarfazeCompany/scar-script/internal/checker"
	"github.com/ScarfazeCompany/scar-script/internal/lexer"
	"github.com/ScarfazeCompany/scar-script/internal/parser"
)

func gen(t *testing.T, src string) string {
	t.Helper()
	toks, lerr := lexer.Lex(src, "t.scar")
	if lerr != nil {
		t.Fatalf("lex error: %v", lerr)
	}
	prog, perr := parser.Parse(toks, "t.scar")
	if perr != nil {
		t.Fatalf("parse error: %v", perr)
	}
	res, cerr := checker.Check(prog)
	if cerr != nil {
		t.Fatalf("check error: %v", cerr)
	}
	out, gerr := Generate(prog, res)
	if gerr != nil {
		t.Fatalf("codegen error: %v", gerr)
	}
	return out
}

func containsSub(t *testing.T, out, want string) {
	t.Helper()
	if !strings.Contains(out, want) {
		t.Errorf("generated code missing %q\n---\n%s", want, out)
	}
}

func TestBasicDeclAndPrint(t *testing.T) {
	out := gen(t, "val name: str = \"Zawiyar\"\nprint(\"Hello {name}\")\n")
	containsSub(t, out, "func main() {")
	containsSub(t, out, `var name string = "Zawiyar"`)
	containsSub(t, out, `_scarfmt.Println("Hello " + _scar_str(name))`)
}

func TestUnusedVarEmitsBlankRead(t *testing.T) {
	out := gen(t, "val unused: int = 5\nval used: int = 1\nprint(used)\n")
	containsSub(t, out, "var unused int = 5")
	containsSub(t, out, "_ = unused")
	if strings.Contains(out, "_ = used") {
		t.Errorf("used variable should not get blank read:\n%s", out)
	}
}

func TestConstAndMut(t *testing.T) {
	out := gen(t, "const API: str = \"https://x\"\nmut n: int = 0\nn++\nprint(API, n)\n")
	containsSub(t, out, `const API string = "https://x"`)
	containsSub(t, out, "n++")
}

func TestIfChain(t *testing.T) {
	out := gen(t, "val age: int = 20\nif age >= 18 {\nprint(\"a\")\n} else if age < 18 {\nprint(\"b\")\n} else {\nprint(\"c\")\n}\n")
	containsSub(t, out, "if age >= 18 {")
	containsSub(t, out, "} else if age < 18 {")
	containsSub(t, out, "} else {")
}

func TestForForms(t *testing.T) {
	out := gen(t, `for i in 1..5 {
print(i)
}
val tags: str[] = ["a"]
for tag in tags {
print(tag)
}
for index, tag in tags {
print(index, tag)
}
val scores: map[str: int] = {a: 1}
for key, value in scores {
print(key, value)
}
mut count: int = 0
for count < 3 {
count++
}
`)
	containsSub(t, out, "for i := 1; i <= 5; i++ {")
	containsSub(t, out, "for _, tag := range tags {")
	containsSub(t, out, "for index, tag := range tags {")
	containsSub(t, out, "for key, value := range scores {")
	containsSub(t, out, "for count < 3 {")
}

func TestCollectionsAndHelpers(t *testing.T) {
	out := gen(t, `val scores: map[str: int] = {"zawiyar": 100}
mut items: str[] = ["a"]
items.add("b")
items.remove(0)
val n: int = items.len()
val s: str = "  x  "
print(s.trim().toUpper(), n, scores["zawiyar"], scores.zawiyar)
`)
	containsSub(t, out, `map[string]int{"zawiyar": 100}`)
	containsSub(t, out, `items = _scar_add(items, "b")`)
	containsSub(t, out, `items = _scar_remove(items, 0)`)
	containsSub(t, out, "len(items)")
	containsSub(t, out, `_scarstrings.TrimSpace(s)`)
	containsSub(t, out, `scores["zawiyar"]`)
}

func TestAnyMemberAndOr(t *testing.T) {
	out := gen(t, "val raw: any = {status: 200}\nval s: str = raw.status or \"default\"\nprint(s)\n")
	containsSub(t, out, `var raw any = map[string]int{"status": 200}`)
	containsSub(t, out, `_scar_or_default[string](_scar_field(raw, "status"), "default")`)
}

func TestTernary(t *testing.T) {
	out := gen(t, "val age: int = 20\nval ok: bool = age >= 18 ? true : false\nprint(ok)\n")
	containsSub(t, out, "func() bool {")
	containsSub(t, out, "if age >= 18 {")
	containsSub(t, out, "return true")
	containsSub(t, out, "return false")
}

func TestBoolOrShortCircuit(t *testing.T) {
	out := gen(t, "val a: bool = true\nval b: bool = false\nif a or b {\nprint(1)\n}\n")
	containsSub(t, out, "if a || b {")
}

func TestIntOrStrZeroCheck(t *testing.T) {
	out := gen(t, "val n: int = 0\nval m: int = n or 7\nval s: str = \"\"\nval t: str = s or \"d\"\nprint(m, t)\n")
	containsSub(t, out, "_scar_or[int](n, 7)")
	containsSub(t, out, `_scar_or[string](s, "d")`)
}

func TestPrecedenceParens(t *testing.T) {
	out := gen(t, "val x: int = (1 + 2) * 3\nprint(x)\n")
	containsSub(t, out, "(1 + 2) * 3")

	out = gen(t, "val x: int = 1 + 2 * 3\nprint(x)\n")
	containsSub(t, out, "1 + 2*3")

	out = gen(t, "val x: int = -(1 + 2)\nval y: bool = !(1 > 2)\nprint(x, y)\n")
	containsSub(t, out, "-(1 + 2)")
	containsSub(t, out, "!(1 > 2)")
}

func TestInterpolationPrecedence(t *testing.T) {
	// interpolated expression inside concatenation must be a function call
	out := gen(t, "val a: int = 1\nval s: str = \"x{a + 2}y\"\nprint(s)\n")
	containsSub(t, out, `_scar_str(a+2)`)
}

func TestEmptyLiteralsAdoptAnnotation(t *testing.T) {
	out := gen(t, "val xs: str[] = []\nval m: map[str: int] = {}\nprint(xs.len(), m.len())\n")
	containsSub(t, out, "var xs []string = []string{}")
	containsSub(t, out, "var m map[string]int = map[string]int{}")
}

func TestSpecProgramGenerates(t *testing.T) {
	src := `
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
print(clean, upper, isMatch, length, fallback, API_URL)
`
	out := gen(t, src)
	containsSub(t, out, "package main")
	containsSub(t, out, "func main() {")
	containsSub(t, out, "_scarstrings.TrimSpace(name)")
}

func TestFunctionEmission(t *testing.T) {
	out := gen(t, `func add(a: int, b: int) int {
return a + b
}
func greet(name: str) {
print("Hello, {name}!")
}
val x: int = add(1, 2)
greet("Z")
print(x)
`)
	containsSub(t, out, "func add(a int, b int) int {")
	containsSub(t, out, "func greet(name string) {")
	containsSub(t, out, "return a + b")
	containsSub(t, out, "var x int = add(1, 2)")
	containsSub(t, out, "greet(\"Z\")")
	// functions are emitted before main
	if strings.Index(out, "func add(") > strings.Index(out, "func main()") {
		t.Errorf("function emitted after main:\n%s", out)
	}
}

func TestMultipleReturnGen(t *testing.T) {
	out := gen(t, `func divide(a: int, b: int) (int, bool) {
if b == 0 {
return 0, false
}
return a / b, true
}
val res, ok = divide(10, 2)
print(res, ok)
`)
	containsSub(t, out, "func divide(a int, b int) (int, bool) {")
	containsSub(t, out, "res, ok := divide(10, 2)")
	containsSub(t, out, "return 0, false")
	containsSub(t, out, "return a / b, true")
}

func TestBareCallStatements(t *testing.T) {
	out := gen(t, `func div() (int, int) {
return 1, 2
}
func voidFn() {
print("x")
}
func single() int {
return 1
}
div()
voidFn()
single()
print(1)
`)
	if strings.Contains(out, "_ = div()") || strings.Contains(out, "_ = voidFn()") {
		t.Errorf("void/tuple calls must be bare statements:\n%s", out)
	}
	containsSub(t, out, "_ = single()")
}

func TestAnonymousFuncGen(t *testing.T) {
	out := gen(t, "val mul = func(x: int, y: int) int {\nreturn x * y\n}\nprint(mul(4, 5))\n")
	containsSub(t, out, "var mul = func(x int, y int) int {")
	containsSub(t, out, "return x * y")
	containsSub(t, out, "mul(4, 5)")
}

func TestFloatAndDecGen(t *testing.T) {
	out := gen(t, "val x: dec = 4.7\nval y: dec = 1 + 4.7\nval z: dec = 4.0\nprint(x, y, z)\n")
	containsSub(t, out, "var x float64 = 4.7")
	containsSub(t, out, "float64(1) + 4.7")
	containsSub(t, out, "var z float64 = 4.0")
}

func TestModuleGen(t *testing.T) {
	out := gen(t, `val sq: dec = math.sqrt(16)
val r: int = math.rand(1, 10)
val n: int = math.round(4.7)
time.sleep(1)
file.write("a.txt", "x")
print(sq, r, n, sys.env("PORT") or "8080")
`)
	containsSub(t, out, "_scarmath.Sqrt(16)")
	containsSub(t, out, "_scar_rand(1, 10)")
	containsSub(t, out, "_scar_round(4.7)")
	containsSub(t, out, "_scar_time_sleep(1)")
	containsSub(t, out, `_scar_file_write("a.txt", "x")`)
	containsSub(t, out, `_scaros.Getenv("PORT")`)
	containsSub(t, out, `_scarmath "math"`)
	containsSub(t, out, `_scartime "time"`)
	containsSub(t, out, `_scarrand "math/rand"`)
	containsSub(t, out, "_scar_time_now") // helpers emitted on demand
}

func TestNoUnusedImports(t *testing.T) {
	out := gen(t, "print(1)\n")
	for _, imp := range []string{`_scarmath "math"`, `_scartime "time"`, `_scarrand "math/rand"`, `_scarstrings "strings"`} {
		if strings.Contains(out, imp) {
			t.Errorf("unused import %q emitted:\n%s", imp, out)
		}
	}
}
