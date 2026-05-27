package exitcode

import "fmt"

type Error struct {
	Code int
	Msg  string
}

func New(code int, format string, args ...any) Error {
	return Error{Code: code, Msg: fmt.Sprintf(format, args...)}
}

func (e Error) Error() string { return e.Msg }
