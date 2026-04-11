package base

import (
	"errors"
	"net/http"

	"server/internal/config"
	"server/internal/storage"
	"server/internal/usecases"
	"server/pkg/httpmodels"
)

func ExtractKnownErrors(err error) *httpmodels.Error {
	if errors.Is(err, usecases.ErrBatchSizeTooBig) {
		return httpmodels.NewError(httpmodels.ErrorCodeBatchSizeTooBig, err.Error())
	}

	if errors.Is(err, usecases.ErrDirectWriteToDLQNotAllowed) {
		return httpmodels.NewError(httpmodels.ErrorCodeQueueNotWritable, err.Error())
	}

	if errors.Is(err, storage.ErrMsgNotFound) || errors.Is(err, storage.ErrArchivedMsgNotFound) {
		return httpmodels.NewError(httpmodels.ErrorCodeMessageNotFound, err.Error())
	}

	var queueError config.QueueNotFoundError
	if errors.As(err, &queueError) {
		return httpmodels.NewError(httpmodels.ErrorCodeQueueNotFound, err.Error())
	}

	return httpmodels.NewError(httpmodels.ErrorCodeUnknown, err.Error())
}

func MapErrorCodeToStatusCode(code httpmodels.ErrorCode) int {
	switch code {
	case httpmodels.ErrorCodeRequestInvalid, httpmodels.ErrorCodeBatchSizeTooBig, httpmodels.ErrorCodeQueueNotWritable:
		return http.StatusBadRequest
	case httpmodels.ErrorCodeMessageNotFound, httpmodels.ErrorCodeQueueNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
