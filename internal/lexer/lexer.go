package lexer

import (
	"strconv"
	"strings"

	"github.com/ScarfazeCompany/scar-script/internal/diag"
	"github.com/ScarfazeCompany/scar-script/internal/token"
)

type Lexer struct {
	src  string
	file string
	pos  int
	line int
	col  int
	toks []token.Token
	err  *diag.Error
}

func New(src, file string) *Lexer {
	return &Lexer{src: src, file: file, line: 1, col: 1}
}

// Lex scans the whole input and returns all tokens (including a final EOF).
func Lex(src, file string) ([]token.Token, *diag.Error) {
	return LexAt(src, file, 1, 1)
}

// LexAt is like Lex but treats src[0] as being at the given line/col.
// It is used to re-lex interpolation expressions with correct positions.
func LexAt(src, file string, line, col int) ([]token.Token, *diag.Error) {
	l := New(src, file)
	l.line, l.col = line, col
	l.run()
	if l.err != nil {
		return nil, l.err
	}
	l.toks = append(l.toks, token.Token{Type: token.EOF, Line: l.line, Col: l.col})
	return l.toks, nil
}

func (l *Lexer) posNow() diag.Position {
	return diag.Position{File: l.file, Line: l.line, Col: l.col}
}

func (l *Lexer) fail(format string, args ...any) {
	if l.err == nil {
		l.err = diag.New(l.posNow(), format, args...)
	}
}

func (l *Lexer) peek(off int) byte {
	if l.pos+off >= len(l.src) {
		return 0
	}
	return l.src[l.pos+off]
}

func (l *Lexer) advance() byte {
	b := l.src[l.pos]
	l.pos++
	if b == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return b
}

func (l *Lexer) add(t token.Token) {
	l.toks = append(l.toks, t)
}

func (l *Lexer) run() {
	for l.pos < len(l.src) && l.err == nil {
		l.scanToken()
	}
}

func (l *Lexer) scanToken() {
	b := l.src[l.pos]

	switch {
	case b == ' ' || b == '\t' || b == '\r':
		l.advance()
		return

	case b == '\n':
		line, col := l.line, l.col
		l.advance()
		l.add(token.Token{Type: token.NEWLINE, Line: line, Col: col})
		return

	case b == '/' && l.peek(1) == '/':
		for l.pos < len(l.src) && l.src[l.pos] != '\n' {
			l.advance()
		}
		return

	case b >= '0' && b <= '9':
		l.scanNumber()
		return

	case isIdentStart(b):
		l.scanIdent()
		return

	case b == '"':
		l.scanString()
		return
	}

	if op, ok := l.scanOperator(); ok {
		l.add(op)
		return
	}

	l.fail("unexpected character %q", string(b))
	l.advance()
}

func (l *Lexer) scanNumber() {
	line, col := l.line, l.col
	start := l.pos
	for l.pos < len(l.src) && l.src[l.pos] >= '0' && l.src[l.pos] <= '9' {
		l.advance()
	}
	// Float only when '.' is immediately followed by a digit, so `1..5`
	// (int-dotdot-int) keeps scanning as a range operator.
	if l.pos < len(l.src) && l.src[l.pos] == '.' && l.peek(1) >= '0' && l.peek(1) <= '9' {
		l.advance() // '.'
		for l.pos < len(l.src) && l.src[l.pos] >= '0' && l.src[l.pos] <= '9' {
			l.advance()
		}
		lit := l.src[start:l.pos]
		if _, err := strconv.ParseFloat(lit, 64); err != nil {
			l.fail("float literal %s out of range", lit)
			return
		}
		l.add(token.Token{Type: token.FLOAT, Lit: lit, Line: line, Col: col})
		return
	}
	lit := l.src[start:l.pos]
	if _, err := strconv.ParseInt(lit, 10, 64); err != nil {
		l.fail("integer literal %s out of range", lit)
		return
	}
	l.add(token.Token{Type: token.INT, Lit: lit, Line: line, Col: col})
}

func isIdentStart(b byte) bool {
	return b == '_' || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func isIdentPart(b byte) bool {
	return isIdentStart(b) || (b >= '0' && b <= '9')
}

func (l *Lexer) scanIdent() {
	line, col := l.line, l.col
	start := l.pos
	for l.pos < len(l.src) && isIdentPart(l.src[l.pos]) {
		l.advance()
	}
	lit := l.src[start:l.pos]
	tt, isKw := token.Keywords[lit]
	if !isKw {
		tt = token.IDENT
	}
	l.add(token.Token{Type: tt, Lit: lit, Line: line, Col: col})
}

func (l *Lexer) scanOperator() (token.Token, bool) {
	line, col := l.line, l.col
	b := l.src[l.pos]

	two := func(next byte, yes, no token.Type) (token.Token, bool) {
		lit := string(b)
		if l.peek(1) == next {
			l.advance()
			l.advance()
			return token.Token{Type: yes, Lit: string([]byte{b, next}), Line: line, Col: col}, true
		}
		l.advance()
		return token.Token{Type: no, Lit: lit, Line: line, Col: col}, true
	}
	one := func(t token.Type) (token.Token, bool) {
		l.advance()
		return token.Token{Type: t, Lit: string(b), Line: line, Col: col}, true
	}

	switch b {
	case '=':
		return two('=', token.EQ, token.ASSIGN)
	case '!':
		return two('=', token.NEQ, token.BANG)
	case '<':
		return two('=', token.LTE, token.LT)
	case '>':
		return two('=', token.GTE, token.GT)
	case '+':
		return two('+', token.PLUSPLUS, token.PLUS)
	case '-':
		if l.peek(1) == '-' {
			return two('-', token.MINUSMINUS, token.MINUS)
		}
		if l.peek(1) == '>' {
			l.advance()
			l.advance()
			return token.Token{Type: token.ARROW, Lit: "->", Line: line, Col: col}, true
		}
		return two('-', token.MINUSMINUS, token.MINUS)
	case '.':
		if l.peek(1) == '.' && l.peek(2) == '.' {
			l.fail("unexpected token '...'")
			l.advance()
			return token.Token{}, false
		}
		return two('.', token.DOTDOT, token.DOT)
	case '*':
		return one(token.STAR)
	case '/':
		return one(token.SLASH)
	case '%':
		return one(token.PERCENT)
	case '?':
		return one(token.QUESTION)
	case ':':
		return one(token.COLON)
	case ',':
		return one(token.COMMA)
	case '(':
		return one(token.LPAREN)
	case ')':
		return one(token.RPAREN)
	case '{':
		return one(token.LBRACE)
	case '}':
		return one(token.RBRACE)
	case '[':
		return one(token.LBRACKET)
	case ']':
		return one(token.RBRACKET)
	}
	return token.Token{}, false
}

// scanString reads a single-line string literal, splitting it into literal
// and interpolation ({expr}) segments.
func (l *Lexer) scanString() {
	line, col := l.line, l.col
	l.add(l.lexString(line, col))
}

func (l *Lexer) lexString(line, col int) token.Token {
	l.advance() // opening quote

	var segs []token.StrSeg
	var lit strings.Builder

	flush := func() {
		if lit.Len() > 0 {
			segs = append(segs, token.StrSeg{Text: lit.String()})
			lit.Reset()
		}
	}

	for {
		if l.pos >= len(l.src) {
			l.fail("unterminated string literal")
			return token.Token{Type: token.STRING, Line: line, Col: col, Segs: segs}
		}
		b := l.src[l.pos]
		switch {
		case b == '"':
			l.advance()
			flush()
			return token.Token{Type: token.STRING, Line: line, Col: col, Segs: segs}

		case b == '\n':
			l.fail("unterminated string literal (newline in string)")
			return token.Token{Type: token.STRING, Line: line, Col: col, Segs: segs}

		case b == '\\':
			l.advance()
			if l.pos >= len(l.src) {
				l.fail("unterminated escape sequence")
				return token.Token{Type: token.STRING, Line: line, Col: col, Segs: segs}
			}
			e := l.advance()
			switch e {
			case 'n':
				lit.WriteByte('\n')
			case 't':
				lit.WriteByte('\t')
			case 'r':
				lit.WriteByte('\r')
			case '\\':
				lit.WriteByte('\\')
			case '"':
				lit.WriteByte('"')
			case '\'':
				lit.WriteByte('\'')
			case '{':
				lit.WriteByte('{')
			case '}':
				lit.WriteByte('}')
			default:
				l.fail("unknown escape sequence \\%s", string(e))
				return token.Token{Type: token.STRING, Line: line, Col: col, Segs: segs}
			}

		case b == '{':
			flush()
			el, ec := l.line, l.col
			l.advance() // consume '{'
			srcStart := l.pos
			depth := 1
			inStr := false
			for depth > 0 {
				if l.pos >= len(l.src) {
					l.fail("unterminated interpolation")
					return token.Token{Type: token.STRING, Line: line, Col: col, Segs: segs}
				}
				c := l.src[l.pos]
				if inStr {
					if c == '\\' {
						l.advance()
						if l.pos < len(l.src) {
							l.advance()
						}
						continue
					}
					if c == '"' {
						inStr = false
					}
					l.advance()
					continue
				}
				switch c {
				case '"':
					inStr = true
					l.advance()
				case '{':
					depth++
					l.advance()
				case '}':
					depth--
					if depth == 0 {
						expr := l.src[srcStart:l.pos]
						if strings.TrimSpace(expr) == "" {
							l.fail("empty interpolation expression")
							return token.Token{Type: token.STRING, Line: line, Col: col, Segs: segs}
						}
						segs = append(segs, token.StrSeg{IsExpr: true, Text: expr, Line: el, Col: ec})
						l.advance() // consume '}'
					} else {
						l.advance()
					}
				default:
					if c == '\n' {
						l.fail("unterminated interpolation (newline in string)")
						return token.Token{Type: token.STRING, Line: line, Col: col, Segs: segs}
					}
					l.advance()
				}
			}
		default:
			lit.WriteByte(b)
			l.advance()
		}
	}
}
