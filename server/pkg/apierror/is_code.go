package apierror

import "errors"

func IsCode(err error, code Code) bool {
	var apiErr Error
	if errors.As(err, &apiErr) {
		return apiErr.Code() == code
	}
	return false
}
