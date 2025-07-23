package routes

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"server/internal/usecases"
	"strconv"
	"strings"
)

type TakeWork struct {
	db      *sql.DB
	logger  *slog.Logger
	useCase *usecases.TakeWork
}

func NewTakeWork(
	db *sql.DB,
	logger *slog.Logger,
	useCase *usecases.TakeWork,
) *TakeWork {
	return &TakeWork{
		db:      db,
		logger:  logger,
		useCase: useCase,
	}
}

func (a *TakeWork) Mount(srv *http.ServeMux) {
	srv.HandleFunc("/work/take", a.handler)
}

func (a *TakeWork) handler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		a.writeError(writer, http.StatusBadRequest, errors.New("method POST expected"))
		return
	}

	params := request.URL.Query()

	kindStr := params.Get("kind")
	if kindStr == "" {
		a.writeError(writer, http.StatusBadRequest, errors.New("parameter 'kind' required"))
		return
	}
	kinds := strings.Split(kindStr, ",")

	limit := 1
	if params.Has("limit") {
		customLimit, err := strconv.Atoi(params.Get("limit"))
		if err != nil {
			a.writeError(writer, http.StatusBadRequest, errors.New("parameter 'limit' must be an integer"))
			return
		}
		limit = customLimit
	}

	tasks, err := a.useCase.Do(request.Context(), kinds, limit)
	if err != nil {
		a.writeError(writer, http.StatusInternalServerError, fmt.Errorf("useCase.Do: %w", err))
		return
	}

	a.writeSuccess(writer, tasks)
}

func (a *TakeWork) writeError(writer http.ResponseWriter, code int, err error) {
	if code >= http.StatusInternalServerError {
		a.logger.Error("get tasks use case failed", "error", err)
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

func (a *TakeWork) writeSuccess(writer http.ResponseWriter, tasks []usecases.TaskToWork) {
	result := make([]any, 0, len(tasks))

	for _, task := range tasks {
		result = append(result, map[string]any{
			"id":      task.ID,
			"payload": string(task.Payload),
		})
	}

	err := json.NewEncoder(writer).Encode(map[string]any{
		"success": true,
		"result":  result,
		"error":   nil,
	})
	if err != nil {
		a.logger.Error("json encode of success response failed", "error", err)
	}
}
