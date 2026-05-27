package exitcode

import (
	"fmt"
	"sync/atomic"
)

var requested atomic.Int32

type Error struct {
	Code int
	Msg  string
}

func New(code int, format string, args ...any) Error {
	return Error{Code: code, Msg: fmt.Sprintf(format, args...)}
}

func (e Error) Error() string { return fmt.Sprintf("%s (exit code %d)", e.Msg, e.Code) }

func Request(code int) {
	if code != 0 {
		requested.Store(int32(code))
	}
}

func Requested() int { return int(requested.Load()) }
