package base

import (
	"server/pkg/apierror"
)

type Error struct {
	status int
	err    error
	code   apierror.Code
}

func NewError(status int, err error) *Error {
	return &Error{status, err, apierror.CodeUnknown}
}

func (e *Error) WithCode(code apierror.Code) *Error {
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

func (e *Error) ErrorCode() apierror.Code {
	return e.code
}
