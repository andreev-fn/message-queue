package base

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"server/pkg/httpmodels"
)

type ValidatableDTO interface {
	Validate() error
}

type TypedHandlerFunc[TI ValidatableDTO, TO any] = func(context.Context, TI) (TO, *Error)

type TypedHandler[TI ValidatableDTO, TO any] struct {
	logger      *slog.Logger
	handlerFunc TypedHandlerFunc[TI, TO]
}

func NewTypedHandler[TI ValidatableDTO, TO any](
	logger *slog.Logger,
	handlerFunc TypedHandlerFunc[TI, TO],
) *TypedHandler[TI, TO] {
	return &TypedHandler[TI, TO]{
		logger:      logger,
		handlerFunc: handlerFunc,
	}
}

func (a *TypedHandler[TI, TO]) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	respDTO, err := a.handleRequest(req)
	if err != nil {
		a.writeError(writer, err)
		return
	}
	a.writeSuccess(writer, respDTO)
}

func (a *TypedHandler[TI, TO]) handleRequest(req *http.Request) (TO, *Error) {
	var emptyResp TO

	if req.Method != http.MethodPost {
		return emptyResp, NewError(http.StatusMethodNotAllowed, errors.New("method not allowed"))
	}

	if req.Header.Get("Content-Type") != "application/json" {
		return emptyResp, NewError(http.StatusBadRequest, errors.New("json content type expected"))
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return emptyResp, NewError(http.StatusBadRequest, fmt.Errorf("io.ReadAll: %w", err))
	}

	dec := json.NewDecoder(bytes.NewReader(bodyBytes))
	dec.DisallowUnknownFields()

	var reqDTO TI
	if err := dec.Decode(&reqDTO); err != nil {
		return emptyResp, NewError(http.StatusBadRequest, fmt.Errorf("json.Unmarshal: %w", err))
	}

	if err := reqDTO.Validate(); err != nil {
		return emptyResp, NewError(http.StatusBadRequest, fmt.Errorf("reqDTO.Validate: %w", err))
	}

	return a.handlerFunc(req.Context(), reqDTO)
}

func (a *TypedHandler[TI, TO]) writeError(writer http.ResponseWriter, apiErr *Error) {
	if apiErr.StatusCode() >= http.StatusInternalServerError {
		a.logger.Error("request failed", "error", apiErr.Error())
	}

	writer.Header().Add("Content-Type", "application/json")
	writer.WriteHeader(apiErr.StatusCode())

	err := json.NewEncoder(writer).Encode(httpmodels.ErrorResponse{
		Code:    string(apiErr.ErrorCode()),
		Message: apiErr.Error(),
	})
	if err != nil {
		a.logger.Error("json encode of error response failed", "error", err)
	}
}

func (a *TypedHandler[TI, TO]) writeSuccess(writer http.ResponseWriter, respDTO TO) {
	writer.Header().Add("Content-Type", "application/json")

	err := json.NewEncoder(writer).Encode(respDTO)
	if err != nil {
		a.logger.Error("json encode of success response failed", "error", err)
	}
}
