package httpmodels

import (
	"encoding/json"
	"fmt"
)

type ErrorCode string

const (
	ErrorCodeUnknown          ErrorCode = "unknown"
	ErrorCodeMessageNotFound  ErrorCode = "message_not_found"
	ErrorCodeQueueNotFound    ErrorCode = "queue_not_found"
	ErrorCodeRequestInvalid   ErrorCode = "request_invalid"
	ErrorCodeBatchSizeTooBig  ErrorCode = "batch_size_too_big"
	ErrorCodeQueueNotWritable ErrorCode = "queue_not_writable"
)

type Error struct {
	code    ErrorCode
	message string
}

func NewError(code ErrorCode, message string) *Error {
	return &Error{code, message}
}

func (e *Error) Error() string {
	return e.message
}

func (e *Error) Code() ErrorCode {
	return e.code
}

type errorDTO struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(errorDTO{
		Code:    string(e.Code()),
		Message: e.message,
	})
}

func (e *Error) UnmarshalJSON(payload []byte) error {
	var dto errorDTO

	if err := json.Unmarshal(payload, &dto); err != nil {
		return fmt.Errorf("json.Unmarshal: %w; payload: %s", err, string(payload))
	}

	e.code = ErrorCode(dto.Code)
	e.message = dto.Message

	return nil
}
