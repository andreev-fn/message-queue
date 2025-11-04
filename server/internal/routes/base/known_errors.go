package base

import (
	"errors"
	"net/http"

	"server/internal/config"
	"server/internal/storage"
	"server/internal/usecases"
	"server/pkg/apierror"
)

func ExtractKnownErrors(err error) *Error {
	if errors.Is(err, usecases.ErrBatchSizeTooBig) {
		return NewError(http.StatusBadRequest, err)
	}

	if errors.Is(err, storage.ErrMsgNotFound) || errors.Is(err, storage.ErrArchivedMsgNotFound) {
		return NewError(http.StatusBadRequest, err).WithCode(apierror.CodeMessageNotFound)
	}

	var queueError config.QueueNotFoundError
	if errors.As(err, &queueError) {
		return NewError(http.StatusBadRequest, err).WithCode(apierror.CodeQueueNotFound)
	}

	return NewError(http.StatusInternalServerError, err)
}
