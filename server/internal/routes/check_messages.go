package routes

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"server/internal/usecases"
)

type CheckMessages struct {
	db      *sql.DB
	logger  *slog.Logger
	useCase *usecases.CheckMessages
}

func NewCheckMessages(
	db *sql.DB,
	logger *slog.Logger,
	useCase *usecases.CheckMessages,
) *CheckMessages {
	return &CheckMessages{
		db:      db,
		logger:  logger,
		useCase: useCase,
	}
}

func (a *CheckMessages) Mount(srv *http.ServeMux) {
	srv.HandleFunc("/messages/check", a.handler)
}

func (a *CheckMessages) handler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Add("Content-Type", "application/json")

	if request.Method != http.MethodGet {
		a.writeError(writer, http.StatusBadRequest, errors.New("method GET expected"))
		return
	}

	msgID := request.URL.Query().Get("id")
	if msgID == "" {
		a.writeError(writer, http.StatusBadRequest, errors.New("parameter 'id' required"))
		return
	}

	result, err := a.useCase.Do(request.Context(), msgID)
	if err != nil {
		a.writeError(writer, http.StatusInternalServerError, err)
		return
	}

	a.writeSuccess(writer, result)
}

func (a *CheckMessages) writeError(writer http.ResponseWriter, code int, err error) {
	if code >= http.StatusInternalServerError {
		a.logger.Error("check messages use case failed", "error", err)
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

func (a *CheckMessages) writeSuccess(writer http.ResponseWriter, result *usecases.CheckMsgResult) {
	err := json.NewEncoder(writer).Encode(map[string]any{
		"success": true,
		"result": map[string]any{
			"id":           result.ID,
			"queue":        result.Queue,
			"created_at":   result.CreatedAt,
			"finalized_at": result.FinalizedAt,
			"status":       result.Status,
			"retries":      result.Retries,
			"payload":      result.Payload,
		},
		"error": nil,
	})
	if err != nil {
		a.logger.Error("json encode of success response failed", "error", err)
	}
}
