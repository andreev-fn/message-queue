package base

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"server/pkg/httpmodels"
)

type ValidatableDTO interface {
	Validate() error
}

type TypedHandlerFunc[TI ValidatableDTO, TO any] = func(context.Context, TI) (TO, *httpmodels.Error)

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

func (a *TypedHandler[TI, TO]) handleRequest(req *http.Request) (TO, *httpmodels.Error) {
	var emptyResp TO

	if req.Method != http.MethodPost {
		return emptyResp, httpmodels.NewError(httpmodels.ErrorCodeRequestInvalid, "method not allowed")
	}

	if req.Header.Get("Content-Type") != "application/json" {
		return emptyResp, httpmodels.NewError(httpmodels.ErrorCodeRequestInvalid, "json content type expected")
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return emptyResp, httpmodels.NewError(
			httpmodels.ErrorCodeRequestInvalid,
			fmt.Sprintf("io.ReadAll: %v", err),
		)
	}

	dec := json.NewDecoder(bytes.NewReader(bodyBytes))
	dec.DisallowUnknownFields()

	var reqDTO TI
	if err := dec.Decode(&reqDTO); err != nil {
		return emptyResp, httpmodels.NewError(
			httpmodels.ErrorCodeRequestInvalid,
			fmt.Sprintf("json.Unmarshal: %v", err),
		)
	}

	if err := reqDTO.Validate(); err != nil {
		return emptyResp, httpmodels.NewError(
			httpmodels.ErrorCodeRequestInvalid,
			fmt.Sprintf("reqDTO.Validate: %v", err),
		)
	}

	return a.handlerFunc(req.Context(), reqDTO)
}

func (a *TypedHandler[TI, TO]) writeError(writer http.ResponseWriter, apiErr *httpmodels.Error) {
	statusCode := MapErrorCodeToStatusCode(apiErr.Code())
	if statusCode >= http.StatusInternalServerError {
		a.logger.Error("request failed", "error", apiErr.Error())
	}

	writer.Header().Add("Content-Type", "application/json")
	writer.WriteHeader(statusCode)

	err := json.NewEncoder(writer).Encode(httpmodels.ErrorResponse{
		Error: apiErr,
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
