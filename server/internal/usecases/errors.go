package usecases

import (
	"errors"
)

var ErrBatchSizeTooBig = errors.New("batch size limit exceeded")
var ErrDirectWriteToDLQNotAllowed = errors.New("writing directly to DLQ is not allowed")
