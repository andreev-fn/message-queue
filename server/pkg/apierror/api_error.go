package apierror

type Error struct {
	code    Code
	message string
}

func NewError(code Code, message string) Error {
	return Error{code, message}
}

func (e Error) Error() string {
	return e.message
}

func (e Error) Code() Code {
	return e.code
}
