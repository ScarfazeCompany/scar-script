package diag

import "fmt"

type Position struct {
	File string
	Line int
	Col  int
}

func (p Position) String() string {
	file := p.File
	if file == "" {
		file = "<input>"
	}
	return fmt.Sprintf("%s:%d:%d", file, p.Line, p.Col)
}

type Error struct {
	Pos Position
	Msg string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: error: %s", e.Pos, e.Msg)
}

func New(pos Position, format string, args ...any) *Error {
	return &Error{Pos: pos, Msg: fmt.Sprintf(format, args...)}
}
