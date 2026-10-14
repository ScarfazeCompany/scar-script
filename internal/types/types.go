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
	}
	return "?"
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
