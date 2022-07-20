package ac

import "fmt"

type Error struct {
	Errno  int    `json:"errno"`
	Errmsg string `json:"errmsg,omitempty"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%d] %s", e.Errno, e.Errmsg)
}

func Errorf(errno int, format string, args ...interface{}) error {
	return &Error{Errno: errno, Errmsg: fmt.Sprintf(format, args...)}
}
