package usecases

import (
	"errors"
)

var ErrBatchSizeTooBig = errors.New("batch size limit exceeded")
