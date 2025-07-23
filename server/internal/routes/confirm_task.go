package routes

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"server/internal/usecases"
)

type ConfirmTask struct {
	db      *sql.DB
	logger  *slog.Logger
	useCase *usecases.ConfirmTask
}

func NewConfirmTask(
	db *sql.DB,
	logger *slog.Logger,
	useCase *usecases.ConfirmTask,
) *ConfirmTask {
	return &ConfirmTask{
		db:      db,
		logger:  logger,
		useCase: useCase,
	}
}

func (a *ConfirmTask) Mount(srv *http.ServeMux) {
	srv.HandleFunc("/task/confirm", a.handler)
}

func (a *ConfirmTask) handler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Add("Content-Type", "application/json")

	if request.Method != http.MethodPost {
		a.writeError(writer, http.StatusBadRequest, errors.New("method POST expected"))
		return
	}

	taskID := request.URL.Query().Get("id")
	if taskID == "" {
		a.writeError(writer, http.StatusBadRequest, errors.New("parameter 'id' required"))
		return
	}

	if err := a.useCase.Do(request.Context(), taskID); err != nil {
		a.writeError(writer, http.StatusInternalServerError, err)
		return
	}

	a.writeSuccess(writer)
}

func (a *ConfirmTask) writeError(writer http.ResponseWriter, code int, err error) {
	if code >= http.StatusInternalServerError {
		a.logger.Error("confirm task failed", "error", err)
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

func (a *ConfirmTask) writeSuccess(writer http.ResponseWriter) {
	err := json.NewEncoder(writer).Encode(map[string]any{
		"success": true,
		"result":  nil,
		"error":   nil,
	})
	if err != nil {
		a.logger.Error("json encode of success response failed", "error", err)
	}
}
