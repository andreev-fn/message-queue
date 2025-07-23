package routes

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"server/internal/usecases"
)

type CheckTask struct {
	db      *sql.DB
	logger  *slog.Logger
	useCase *usecases.CheckTask
}

func NewCheckTask(
	db *sql.DB,
	logger *slog.Logger,
	useCase *usecases.CheckTask,
) *CheckTask {
	return &CheckTask{
		db:      db,
		logger:  logger,
		useCase: useCase,
	}
}

func (a *CheckTask) Mount(srv *http.ServeMux) {
	srv.HandleFunc("/task/check", a.handler)
}

func (a *CheckTask) handler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Add("Content-Type", "application/json")

	if request.Method != http.MethodGet {
		a.writeError(writer, http.StatusBadRequest, errors.New("method GET expected"))
		return
	}

	taskID := request.URL.Query().Get("id")
	if taskID == "" {
		a.writeError(writer, http.StatusBadRequest, errors.New("parameter 'id' required"))
		return
	}

	result, err := a.useCase.Do(request.Context(), taskID)
	if err != nil {
		a.writeError(writer, http.StatusInternalServerError, err)
		return
	}

	a.writeSuccess(writer, result)
}

func (a *CheckTask) writeError(writer http.ResponseWriter, code int, err error) {
	if code >= http.StatusInternalServerError {
		a.logger.Error("get tasks by id failed", "error", err)
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

func (a *CheckTask) writeSuccess(writer http.ResponseWriter, result *usecases.CheckTaskResult) {
	err := json.NewEncoder(writer).Encode(map[string]any{
		"success": true,
		"result": map[string]any{
			"id":           result.ID,
			"kind":         result.Kind,
			"created_at":   result.CreatedAt,
			"finalized_at": result.FinalizedAt,
			"status":       result.Status,
			"retries":      result.Retries,
			"payload":      result.Payload,
			"result":       result.Result,
		},
		"error": nil,
	})
	if err != nil {
		a.logger.Error("json encode of success response failed", "error", err)
	}
}
