package token

// Type identifies the kind of a lexical token.
type Type string

const (
	EOF     Type = "EOF"
	NEWLINE Type = "NEWLINE"
	IDENT   Type = "IDENT"
	INT     Type = "INT"
	FLOAT   Type = "FLOAT"
	STRING  Type = "STRING"

	// Keywords
	VAL    Type = "val"
	MUT    Type = "mut"
	CONST  Type = "const"
	IF     Type = "if"
	ELSE   Type = "else"
	FOR    Type = "for"
	IN     Type = "in"
	AND    Type = "and"
	OR     Type = "or"
	TRUE   Type = "true"
	FALSE  Type = "false"
	MAP    Type = "map"
	T_STR  Type = "str"
	T_INT  Type = "int"
	T_BOOL Type = "bool"
	T_ANY  Type = "any"
	T_DEC  Type = "dec"
	FUNC   Type = "func"
	RETURN Type = "return"

	// Operators & punctuation
	ASSIGN     Type = "="
	EQ         Type = "=="
	NEQ        Type = "!="
	LT         Type = "<"
	GT         Type = ">"
	LTE        Type = "<="
	GTE        Type = ">="
	PLUS       Type = "+"
	MINUS      Type = "-"
	STAR       Type = "*"
	SLASH      Type = "/"
	PERCENT    Type = "%"
	BANG       Type = "!"
	PLUSPLUS   Type = "++"
	MINUSMINUS Type = "--"
	ARROW      Type = "->"
	DOT        Type = "."
	DOTDOT     Type = ".."
	QUESTION   Type = "?"
	COLON      Type = ":"
	COMMA      Type = ","
	LPAREN     Type = "("
	RPAREN     Type = ")"
	LBRACE     Type = "{"
	RBRACE     Type = "}"
	LBRACKET   Type = "["
	RBRACKET   Type = "]"
)

// Keywords maps identifier spellings to keyword token types.
var Keywords = map[string]Type{
	"val":    VAL,
	"mut":    MUT,
	"const":  CONST,
	"if":     IF,
	"else":   ELSE,
	"for":    FOR,
	"in":     IN,
	"and":    AND,
	"or":     OR,
	"true":   TRUE,
	"false":  FALSE,
	"map":    MAP,
	"str":    T_STR,
	"int":    T_INT,
	"bool":   T_BOOL,
	"any":    T_ANY,
	"dec":    T_DEC,
	"func":   FUNC,
	"return": RETURN,
}

// StrSeg is one part of a string literal token: either plain text or
// the source text of an interpolation expression like {name}.
type StrSeg struct {
	IsExpr bool
	Text   string
	Line   int
	Col    int
}

// Token is a single lexical token.
type Token struct {
	Type Type
	Lit  string // raw spelling (identifier name, number text, operator)
	Line int
	Col  int
	Segs []StrSeg // only for STRING tokens
}

func (t Token) String() string {
	if t.Type == STRING {
		return "STRING"
	}
	return string(t.Type)
}
