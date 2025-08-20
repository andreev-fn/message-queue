package routes

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"server/internal/usecases"
)

type ConfirmMessages struct {
	db      *sql.DB
	logger  *slog.Logger
	useCase *usecases.ConfirmMessages
}

func NewConfirmMessages(
	db *sql.DB,
	logger *slog.Logger,
	useCase *usecases.ConfirmMessages,
) *ConfirmMessages {
	return &ConfirmMessages{
		db:      db,
		logger:  logger,
		useCase: useCase,
	}
}

func (a *ConfirmMessages) Mount(srv *http.ServeMux) {
	srv.HandleFunc("/messages/confirm", a.handler)
}

func (a *ConfirmMessages) handler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Add("Content-Type", "application/json")

	if request.Method != http.MethodPost {
		a.writeError(writer, http.StatusBadRequest, errors.New("method POST expected"))
		return
	}

	msgID := request.URL.Query().Get("id")
	if msgID == "" {
		a.writeError(writer, http.StatusBadRequest, errors.New("parameter 'id' required"))
		return
	}

	if err := a.useCase.Do(request.Context(), msgID); err != nil {
		a.writeError(writer, http.StatusInternalServerError, err)
		return
	}

	a.writeSuccess(writer)
}

func (a *ConfirmMessages) writeError(writer http.ResponseWriter, code int, err error) {
	if code >= http.StatusInternalServerError {
		a.logger.Error("confirm messages use case failed", "error", err)
	}

	writer.WriteHeader(code)

	err = json.NewEncoder(writer).Encode(map[string]any{
		"success": false,
		"result":  nil,
		"error":   err.Error(),
	})
	if err != nil {
		a.logger.Error("json encode of error response failed", "error", err)
	}
}

func (a *ConfirmMessages) writeSuccess(writer http.ResponseWriter) {
	err := json.NewEncoder(writer).Encode(map[string]any{
		"success": true,
		"result":  nil,
		"error":   nil,
	})
	if err != nil {
		a.logger.Error("json encode of success response failed", "error", err)
	}
}
