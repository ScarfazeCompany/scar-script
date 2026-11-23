package types

import (
	"fmt"
	"strings"
)

type Kind int

const (
	Str Kind = iota
	Int
	Bool
	Any
	Void
	Range
	Float // dec (64-bit float)
	StructKind
	OptionKind
	ResultKind
)

type Type interface {
	String() string
}

type Basic struct {
	K Kind
}

type Array struct {
	Elem Type
}

type Map struct {
	Key Type
	Val Type
}

// Struct is a user-defined struct type: `struct User { id: int, name: str }`.
type Struct struct {
	Name   string
	Fields []*StructField
}

// StructField is one field of a struct.
type StructField struct {
	Name string
	Type Type
}

// Option is `Option<T>`.
type Option struct {
	Elem Type
}

// Result is `Result<T, E>`.
type Result struct {
	Ok  Type
	Err Type
}

// Func is a function type: `func(int, bool) int`.
type Func struct {
	Params []Type
	Result Type // VoidType when no result
}

// Tuple is a multiple-return type: `(int, bool)`.
type Tuple struct {
	Elems []Type
}

var (
	StrType  = &Basic{K: Str}
	IntType  = &Basic{K: Int}
	BoolType = &Basic{K: Bool}
	AnyType  = &Basic{K: Any}
	VoidType = &Basic{K: Void}
	RangeT   = &Basic{K: Range}
	DecType  = &Basic{K: Float}
)

func (b *Basic) String() string {
	switch b.K {
	case Str:
		return "str"
	case Int:
		return "int"
	case Bool:
		return "bool"
	case Any:
		return "any"
	case Void:
		return "void"
	case Range:
		return "range"
	case Float:
		return "dec"
	case StructKind:
		return "struct"
	case OptionKind:
		return "option"
	case ResultKind:
		return "result"
	}
	return "?"
}

func (s *Struct) String() string {
	var sb strings.Builder
	sb.WriteString("struct ")
	sb.WriteString(s.Name)
	sb.WriteByte('{')
	for i, f := range s.Fields {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(f.Name)
		sb.WriteString(": ")
		sb.WriteString(f.Type.String())
	}
	sb.WriteByte('}')
	return sb.String()
}

func (o *Option) String() string {
	return "option[" + o.Elem.String() + "]"
}

func (r *Result) String() string {
	return "result[" + r.Ok.String() + ", " + r.Err.String() + "]"
}

func (a *Array) String() string { return a.Elem.String() + "[]" }

func (m *Map) String() string {
	return fmt.Sprintf("map[%s: %s]", m.Key, m.Val)
}

func (f *Func) String() string {
	var sb strings.Builder
	sb.WriteString("func(")
	for i, pt := range f.Params {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(pt.String())
	}
	sb.WriteString(") ")
	sb.WriteString(f.Result.String())
	return sb.String()
}

func (t *Tuple) String() string {
	var sb strings.Builder
	sb.WriteByte('(')
	for i, et := range t.Elems {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(et.String())
	}
	sb.WriteByte(')')
	return sb.String()
}

func Equal(a, b Type) bool {
	if a == b {
		return true
	}
	ab, aok := a.(*Basic)
	bb, bok := b.(*Basic)
	if aok && bok {
		return ab.K == bb.K
	}
	aa, aok := a.(*Array)
	ba, bok := b.(*Array)
	if aok && bok {
		return Equal(aa.Elem, ba.Elem)
	}
	am, aok := a.(*Map)
	bm, bok := b.(*Map)
	if aok && bok {
		return Equal(am.Key, bm.Key) && Equal(am.Val, bm.Val)
	}
	af, aok := a.(*Func)
	bf, bok := b.(*Func)
	if aok && bok {
		if len(af.Params) != len(bf.Params) || !Equal(af.Result, bf.Result) {
			return false
		}
		for i := range af.Params {
			if !Equal(af.Params[i], bf.Params[i]) {
				return false
			}
		}
		return true
	}
	at, aok := a.(*Tuple)
	bt, bok := b.(*Tuple)
	if aok && bok {
		if len(at.Elems) != len(bt.Elems) {
			return false
		}
		for i := range at.Elems {
			if !Equal(at.Elems[i], bt.Elems[i]) {
				return false
			}
		}
		return true
	}
	as_, aok := a.(*Struct)
	bs_, bok := b.(*Struct)
	if aok && bok {
		if as_.Name != bs_.Name || len(as_.Fields) != len(bs_.Fields) {
			return false
		}
		for i := range as_.Fields {
			if as_.Fields[i].Name != bs_.Fields[i].Name || !Equal(as_.Fields[i].Type, bs_.Fields[i].Type) {
				return false
			}
		}
		return true
	}
	ao, aok := a.(*Option)
	bo, bok := b.(*Option)
	if aok && bok {
		return Equal(ao.Elem, bo.Elem)
	}
	ar, aok := a.(*Result)
	br, bok := b.(*Result)
	if aok && bok {
		return Equal(ar.Ok, br.Ok) && Equal(ar.Err, br.Err)
	}
	return false
}

// Assignable reports whether a value of type src may be used where dst is
// expected. The only widening rule is T -> any.
func Assignable(dst, src Type) bool {
	if src == VoidType {
		return false
	}
	if dst == AnyType {
		return true
	}
	return Equal(dst, src)
}

// IsZeroCheckable reports whether the `or` default operator supports this
// type with a plain zero-value check (string/int/bool/dec).
func IsZeroCheckable(t Type) bool {
	b, ok := t.(*Basic)
	if !ok {
		return false
	}
	return b.K == Str || b.K == Int || b.K == Bool || b.K == Float
}

// IsNumeric reports whether t is int or dec.
func IsNumeric(t Type) bool {
	b, ok := t.(*Basic)
	return ok && (b.K == Int || b.K == Float)
}

// IsStruct reports whether t is a user-defined struct type.
func IsStruct(t Type) bool {
	_, ok := t.(*Struct)
	return ok
}

// IsOption reports whether t is Option<_>.
func IsOption(t Type) bool {
	_, ok := t.(*Option)
	return ok
}

// IsResult reports whether t is Result<_, _>.
func IsResult(t Type) bool {
	_, ok := t.(*Result)
	return ok
}

// OptionElem returns the element type of an Option<T>.
func OptionElem(t Type) Type {
	if o, ok := t.(*Option); ok {
		return o.Elem
	}
	return nil
}

// ResultOkType returns the Ok type of a Result<T, E>.
func ResultOkType(t Type) Type {
	if r, ok := t.(*Result); ok {
		return r.Ok
	}
	return nil
}

// ResultErrType returns the Err type of a Result<T, E>.
func ResultErrType(t Type) Type {
	if r, ok := t.(*Result); ok {
		return r.Err
	}
	return nil
}

// Promote returns the common type when two numeric types are mixed
// (int with dec promotes to dec); ok is false otherwise.
func Promote(a, b Type) (Type, bool) {
	if IsNumeric(a) && IsNumeric(b) {
		if Equal(a, b) {
			return a, true
		}
		return DecType, true
	}
	return nil, false
}
