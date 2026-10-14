package lexer

import (
	"strings"
	"testing"

	"github.com/ScarfazeCompany/scar-script/internal/token"
)

func lexOK(t *testing.T, src string) []token.Token {
	t.Helper()
	toks, err := Lex(src, "t.scar")
	if err != nil {
		t.Fatalf("lex error: %v", err)
	}
	return toks
}

func types(toks []token.Token) []token.Type {
	out := make([]token.Type, len(toks))
	for i, tk := range toks {
		out[i] = tk.Type
	}
	return out
}

func expectTypes(t *testing.T, toks []token.Token, want ...token.Type) {
	t.Helper()
	got := types(toks)
	if len(got) != len(want) {
		t.Fatalf("token count = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("tok[%d] = %s, want %s", i, got[i], want[i])
		}
	}
}

func TestKeywordsAndIdents(t *testing.T) {
	toks := lexOK(t, "val status: str = 1")
	expectTypes(t, toks,
		token.VAL, token.IDENT, token.COLON, token.T_STR, token.ASSIGN, token.INT, token.EOF)
	if toks[1].Lit != "status" {
		t.Errorf("ident lit = %q", toks[1].Lit)
	}
	if toks[5].Lit != "1" {
		t.Errorf("int lit = %q", toks[5].Lit)
	}
}

func TestFuncKeywords(t *testing.T) {
	toks := lexOK(t, "func add(a: dec) int { return a }")
	expectTypes(t, toks,
		token.FUNC, token.IDENT, token.LPAREN, token.IDENT, token.COLON, token.T_DEC,
		token.RPAREN, token.T_INT, token.LBRACE, token.RETURN, token.IDENT,
		token.RBRACE, token.EOF)
}

func TestFloatLiterals(t *testing.T) {
	toks := lexOK(t, "3.14 0.5 4.7")
	expectTypes(t, toks, token.FLOAT, token.FLOAT, token.FLOAT, token.EOF)
	for i, want := range []string{"3.14", "0.5", "4.7"} {
		if toks[i].Lit != want {
			t.Errorf("tok[%d] lit = %q, want %q", i, toks[i].Lit, want)
		}
	}
}

func TestRangeNotFloat(t *testing.T) {
	toks := lexOK(t, "for i in 1..5 { }")
	expectTypes(t, toks,
		token.FOR, token.IDENT, token.IN, token.INT, token.DOTDOT, token.INT,
		token.LBRACE, token.RBRACE, token.EOF)
}

func TestTrailingDotNotFloat(t *testing.T) {
	toks := lexOK(t, "1.")
	expectTypes(t, toks, token.INT, token.DOT, token.EOF)
}

func TestOperators(t *testing.T) {
	toks := lexOK(t, "== != <= >= ++ -- .. + - ! ? : .")
	expectTypes(t, toks,
		token.EQ, token.NEQ, token.LTE, token.GTE, token.PLUSPLUS, token.MINUSMINUS,
		token.DOTDOT, token.PLUS, token.MINUS, token.BANG, token.QUESTION, token.COLON,
		token.DOT, token.EOF)
}

func TestCommentsAndNewlines(t *testing.T) {
	toks := lexOK(t, "a // comment\nb\n")
	expectTypes(t, toks, token.IDENT, token.NEWLINE, token.IDENT, token.NEWLINE, token.EOF)
}

func TestPlainString(t *testing.T) {
	toks := lexOK(t, `"hello world"`)
	expectTypes(t, toks, token.STRING, token.EOF)
	if len(toks[0].Segs) != 1 || toks[0].Segs[0].Text != "hello world" || toks[0].Segs[0].IsExpr {
		t.Errorf("segs = %+v", toks[0].Segs)
	}
}

func TestInterpolation(t *testing.T) {
	toks := lexOK(t, `"User {name} ki age {age} hai"`)
	segs := toks[0].Segs
	if len(segs) != 5 {
		t.Fatalf("seg count = %d: %+v", len(segs), segs)
	}
	if segs[0].IsExpr || segs[0].Text != "User " {
		t.Errorf("seg0 = %+v", segs[0])
	}
	if !segs[1].IsExpr || segs[1].Text != "name" {
		t.Errorf("seg1 = %+v", segs[1])
	}
	if !segs[3].IsExpr || segs[3].Text != "age" {
		t.Errorf("seg3 = %+v", segs[3])
	}
	if segs[2].IsExpr || segs[2].Text != " ki age " {
		t.Errorf("seg2 = %+v", segs[2])
	}
}

func TestEscapes(t *testing.T) {
	toks := lexOK(t, `"a\nb\t\"q\" \{x\}"`)
	segs := toks[0].Segs
	if len(segs) != 1 {
		t.Fatalf("segs = %+v", segs)
	}
	want := "a\nb\t\"q\" {x}"
	if segs[0].Text != want {
		t.Errorf("text = %q, want %q", segs[0].Text, want)
	}
}

func TestNestedInterpolation(t *testing.T) {
	toks := lexOK(t, `"{a + (1 == 1 ? 2 : 3)}"`)
	segs := toks[0].Segs
	if len(segs) != 1 || !segs[0].IsExpr || segs[0].Text != "a + (1 == 1 ? 2 : 3)" {
		t.Errorf("segs = %+v", segs)
	}
}

func TestPositions(t *testing.T) {
	toks := lexOK(t, "val\n  status")
	// val at 1:1, newline at 1:4, status at 2:3
	if toks[0].Line != 1 || toks[0].Col != 1 {
		t.Errorf("val pos = %d:%d", toks[0].Line, toks[0].Col)
	}
	if toks[2].Line != 2 || toks[2].Col != 3 {
		t.Errorf("status pos = %d:%d", toks[2].Line, toks[2].Col)
	}
}

func TestLexErrors(t *testing.T) {
	cases := []struct {
		src string
		msg string
	}{
		{`"abc`, "unterminated string literal"},
		{`"ab\z"`, "unknown escape"},
		{`"a {b"`, "unterminated interpolation"},
		{`"a {}"`, "empty interpolation"},
		{`@`, "unexpected character"},
		{"123456789012345678901234567890", "out of range"},
	}
	for _, c := range cases {
		_, err := Lex(c.src, "t.scar")
		if err == nil {
			t.Errorf("%q: expected error", c.src)
			continue
		}
		if !strings.Contains(err.Msg, c.msg) {
			t.Errorf("%q: err %q, want contains %q", c.src, err.Msg, c.msg)
		}
	}
}
