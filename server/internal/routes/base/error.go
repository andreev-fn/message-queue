package base

type Error struct {
	status int
	err    error
	code   string
}

func NewError(status int, err error) *Error {
	return &Error{status, err, "unknown"}
}

func (e *Error) WithCode(code string) *Error {
	return &Error{e.status, e.err, code}
}

func (e *Error) Error() string {
	return e.err.Error()
}

func (e *Error) Unwrap() error {
	return e.err
}

func (e *Error) StatusCode() int {
	return e.status
}

func (e *Error) ErrorCode() string {
	return e.code
}
