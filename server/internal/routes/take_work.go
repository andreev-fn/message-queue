package routes

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"server/internal/usecases"
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

	queueStr := params.Get("queue")
	if queueStr == "" {
		a.writeError(writer, http.StatusBadRequest, errors.New("parameter 'queue' required"))
		return
	}
	queues := strings.Split(queueStr, ",")

	limit := 1
	if params.Has("limit") {
		customLimit, err := strconv.Atoi(params.Get("limit"))
		if err != nil {
			a.writeError(writer, http.StatusBadRequest, errors.New("parameter 'limit' must be an integer"))
			return
		}
		if customLimit < 1 {
			a.writeError(writer, http.StatusBadRequest, errors.New("parameter 'limit' must be greater than 0"))
			return
		}
		limit = customLimit
	}

	poll := time.Duration(0)
	if params.Has("poll") {
		customPoll, err := strconv.Atoi(params.Get("poll"))
		if err != nil {
			a.writeError(writer, http.StatusBadRequest, errors.New("parameter 'poll' must be an integer"))
			return
		}
		if customPoll < 0 {
			a.writeError(writer, http.StatusBadRequest, errors.New("parameter 'poll' must be >= 0"))
			return
		}
		poll = time.Duration(customPoll) * time.Second
	}

	messages, err := a.useCase.Do(request.Context(), queues, limit, poll)
	if err != nil {
		a.writeError(writer, http.StatusInternalServerError, fmt.Errorf("useCase.Do: %w", err))
		return
	}

	a.writeSuccess(writer, messages)
}

func (a *TakeWork) writeError(writer http.ResponseWriter, code int, err error) {
	if code >= http.StatusInternalServerError {
		a.logger.Error("get messages use case failed", "error", err)
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

func (a *TakeWork) writeSuccess(writer http.ResponseWriter, messages []usecases.MessageToWork) {
	result := make([]any, 0, len(messages))

	for _, message := range messages {
		result = append(result, map[string]any{
			"id":      message.ID,
			"payload": string(message.Payload),
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
