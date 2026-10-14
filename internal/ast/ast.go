package ast

import (
	"github.com/ScarfazeCompany/scar-script/internal/diag"
	"github.com/ScarfazeCompany/scar-script/internal/token"
)

type Node interface {
	Pos() diag.Position
}

type Stmt interface {
	Node
	stmtNode()
}

type Expr interface {
	Node
	exprNode()
}

type TypeExpr interface {
	Node
	typeNode()
}

// ---------- Expressions ----------

type IntLit struct {
	P     diag.Position
	Value int64
}

type FloatLit struct {
	P     diag.Position
	Value float64
}

type BoolLit struct {
	P     diag.Position
	Value bool
}

type Ident struct {
	P    diag.Position
	Name string
}

type StrPart struct {
	Lit  string // non-empty when Expr == nil
	Expr Expr   // interpolation expression
}

type StrLit struct {
	P     diag.Position
	Parts []StrPart
}

type ArrayLit struct {
	P     diag.Position
	Elems []Expr
}

type MapEntry struct {
	Key Expr // StrLit with a single literal part
	Val Expr
}

type MapLit struct {
	P       diag.Position
	Entries []MapEntry
}

type BinaryExpr struct {
	P  diag.Position
	Op token.Type
	L  Expr
	R  Expr
}

type UnaryExpr struct {
	P  diag.Position
	Op token.Type // token.MINUS or token.BANG
	X  Expr
}

type TernaryExpr struct {
	P    diag.Position
	Cond Expr
	A    Expr
	B    Expr
}

type RangeExpr struct {
	P  diag.Position
	Lo Expr
	Hi Expr
}

type CallExpr struct {
	P      diag.Position
	Callee Expr
	Args   []Expr
}

type MemberExpr struct {
	P       diag.Position // position of the base expression
	X       Expr
	Name    string
	NamePos diag.Position
}

type IndexExpr struct {
	P     diag.Position
	X     Expr
	Index Expr
}

// FuncLit is an anonymous function: `func(a: int) int { ... }`.
// It is only valid as the direct value of a declaration.
type FuncLit struct {
	P      diag.Position
	Params []Param
	Result TypeExpr // nil = void
	Body   *BlockStmt
}

type ReturnStmt struct {
	P      diag.Position
	Values []Expr // empty = bare `return` (void); multiple = tuple return
}

func (x *IntLit) Pos() diag.Position      { return x.P }
func (x *FloatLit) Pos() diag.Position    { return x.P }
func (x *BoolLit) Pos() diag.Position     { return x.P }
func (x *Ident) Pos() diag.Position       { return x.P }
func (x *StrLit) Pos() diag.Position      { return x.P }
func (x *ArrayLit) Pos() diag.Position    { return x.P }
func (x *MapLit) Pos() diag.Position      { return x.P }
func (x *BinaryExpr) Pos() diag.Position  { return x.P }
func (x *UnaryExpr) Pos() diag.Position   { return x.P }
func (x *TernaryExpr) Pos() diag.Position { return x.P }
func (x *RangeExpr) Pos() diag.Position   { return x.P }
func (x *CallExpr) Pos() diag.Position    { return x.P }
func (x *MemberExpr) Pos() diag.Position  { return x.P }
func (x *IndexExpr) Pos() diag.Position   { return x.P }
func (x *FuncLit) Pos() diag.Position     { return x.P }

func (*IntLit) exprNode()      {}
func (*FloatLit) exprNode()    {}
func (*BoolLit) exprNode()     {}
func (*Ident) exprNode()       {}
func (*StrLit) exprNode()      {}
func (*ArrayLit) exprNode()    {}
func (*MapLit) exprNode()      {}
func (*BinaryExpr) exprNode()  {}
func (*UnaryExpr) exprNode()   {}
func (*TernaryExpr) exprNode() {}
func (*RangeExpr) exprNode()   {}
func (*CallExpr) exprNode()    {}
func (*MemberExpr) exprNode()  {}
func (*IndexExpr) exprNode()   {}
func (*FuncLit) exprNode()     {}

// ---------- Type expressions ----------

type BasicType struct {
	P    diag.Position
	Kind token.Type // T_STR, T_INT, T_BOOL, T_ANY
}

type ArrayTypeExpr struct {
	P    diag.Position
	Elem TypeExpr
}

type MapTypeExpr struct {
	P   diag.Position
	Key TypeExpr
	Val TypeExpr
}

// TupleTypeExpr is a multiple-return type: `(int, bool)`.
type TupleTypeExpr struct {
	P     diag.Position
	Elems []TypeExpr
}

func (t *BasicType) Pos() diag.Position     { return t.P }
func (t *ArrayTypeExpr) Pos() diag.Position { return t.P }
func (t *MapTypeExpr) Pos() diag.Position   { return t.P }
func (t *TupleTypeExpr) Pos() diag.Position { return t.P }
func (*BasicType) typeNode()                {}
func (*ArrayTypeExpr) typeNode()            {}
func (*MapTypeExpr) typeNode()              {}
func (*TupleTypeExpr) typeNode()            {}

// ---------- Statements ----------

type Mutability int

const (
	Val Mutability = iota
	Mut
	Const
)

type DeclStmt struct {
	P       diag.Position
	Kind    Mutability
	Name    string
	NamePos diag.Position
	Type    TypeExpr // nil = inferred from Value (annotation is optional)
	Value   Expr
}

// MultiDeclStmt declares several names from a tuple-returning call:
// `val a, b = divide(10, 2)`. Annotation is not allowed (values come
// from the call's result types).
type MultiDeclStmt struct {
	P        diag.Position
	Kind     Mutability
	Names    []string
	NamePoss []diag.Position
	Value    Expr
}

// Param is one function parameter: `a: int`.
type Param struct {
	Name    string
	NamePos diag.Position
	Type    TypeExpr
}

// FuncDecl is a named top-level function:
// `func add(a: int, b: int) int { ... }`.
type FuncDecl struct {
	P       diag.Position
	Name    string
	NamePos diag.Position
	Params  []Param
	Result  TypeExpr // nil = void
	Body    *BlockStmt
}

type AssignStmt struct {
	P       diag.Position
	Name    string
	NamePos diag.Position
	Value   Expr
}

type IncDecStmt struct {
	P       diag.Position
	Name    string
	NamePos diag.Position
	Op      token.Type // PLUSPLUS or MINUSMINUS
}

type BlockStmt struct {
	P     diag.Position
	Stmts []Stmt
}

type IfStmt struct {
	P    diag.Position
	Cond Expr
	Then *BlockStmt
	Else Stmt // *BlockStmt, *IfStmt, or nil
}

type ForKind int

const (
	ForCond     ForKind = iota // for cond { }
	ForRange                   // for i in 1..5 { }
	ForIterable                // for x in xs { }
	ForTwo                     // for k, v in m { }
)

type ForStmt struct {
	P        diag.Position
	Kind     ForKind
	Var      string
	VarPos   diag.Position
	Var2     string
	Var2Pos  diag.Position
	Iterable Expr // range, iterable, or condition expression
	Body     *BlockStmt
}

type ExprStmt struct {
	P diag.Position
	X Expr
}

type Program struct {
	File  string
	Stmts []Stmt
}

func (s *DeclStmt) Pos() diag.Position      { return s.P }
func (s *MultiDeclStmt) Pos() diag.Position { return s.P }
func (s *AssignStmt) Pos() diag.Position    { return s.P }
func (s *IncDecStmt) Pos() diag.Position    { return s.P }
func (s *BlockStmt) Pos() diag.Position     { return s.P }
func (s *IfStmt) Pos() diag.Position        { return s.P }
func (s *ForStmt) Pos() diag.Position       { return s.P }
func (s *ExprStmt) Pos() diag.Position      { return s.P }
func (s *FuncDecl) Pos() diag.Position      { return s.P }
func (s *ReturnStmt) Pos() diag.Position    { return s.P }

func (*DeclStmt) stmtNode()      {}
func (*MultiDeclStmt) stmtNode() {}
func (*AssignStmt) stmtNode()    {}
func (*IncDecStmt) stmtNode()    {}
func (*BlockStmt) stmtNode()     {}
func (*IfStmt) stmtNode()        {}
func (*ForStmt) stmtNode()       {}
func (*ExprStmt) stmtNode()      {}
func (*FuncDecl) stmtNode()      {}
func (*ReturnStmt) stmtNode()    {}
