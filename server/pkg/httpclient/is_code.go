package httpclient

import (
	"errors"

	"server/pkg/httpmodels"
)

func IsCode(err error, code httpmodels.ErrorCode) bool {
	var apiErr *httpmodels.Error
	if errors.As(err, &apiErr) {
		return apiErr.Code() == code
	}
	return false
}
