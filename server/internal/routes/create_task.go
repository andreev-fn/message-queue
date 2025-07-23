package routes

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"server/internal/usecases"
	"time"
)

type CreateTask struct {
	db      *sql.DB
	logger  *slog.Logger
	useCase *usecases.CreateTask
}

func NewCreateTask(
	db *sql.DB,
	logger *slog.Logger,
	useCase *usecases.CreateTask,
) *CreateTask {
	return &CreateTask{
		db:      db,
		logger:  logger,
		useCase: useCase,
	}
}

func (a *CreateTask) Mount(srv *http.ServeMux) {
	srv.HandleFunc("/task/create", a.handler)
}

func (a *CreateTask) handler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		a.writeError(writer, errors.New("method POST expected"))
		return
	}

	if request.Header.Get("Content-Type") != "application/json" {
		a.writeError(writer, errors.New("json content type expected"))
		return
	}

	bodyBytes, err := io.ReadAll(request.Body)
	if err != nil {
		a.writeError(writer, fmt.Errorf("io.ReadAll: %w", err))
		return
	}

	var jobData struct {
		Kind        string          `json:"kind"`
		Payload     json.RawMessage `json:"payload"`
		Priority    *int            `json:"priority"`
		AutoConfirm bool            `json:"auto_confirm"`
		StartAt     string          `json:"start_at"`
	}
	err = json.Unmarshal(bodyBytes, &jobData)
	if err != nil {
		a.writeError(writer, fmt.Errorf("json.Unmarshal: %w (%s)", err, string(bodyBytes)))
		return
	}

	var startAt *time.Time
	if jobData.StartAt != "" {
		t, err := time.Parse(time.RFC3339, jobData.StartAt)
		if err != nil {
			a.writeError(writer, errors.New("can`t parse start_at, expected format is RFC3339"))
			return
		}
		startAt = &t
	}

	priority := 100
	if jobData.Priority != nil {
		priority = *jobData.Priority
		if priority < 0 || priority > 255 {
			a.writeError(writer, errors.New("priority must be between 0 and 255"))
			return
		}
	}

	taskID, err := a.useCase.Do(
		request.Context(),
		jobData.Kind,
		jobData.Payload,
		priority,
		jobData.AutoConfirm,
		startAt,
	)
	if err != nil {
		a.writeError(writer, fmt.Errorf("useCase.Do: %w", err))
		return
	}

	a.writeSuccess(taskID, writer)
}

func (a *CreateTask) writeError(writer http.ResponseWriter, err error) {
	writer.WriteHeader(http.StatusBadRequest)

	err = json.NewEncoder(writer).Encode(map[string]any{
		"success": false,
		"result":  nil,
		"error":   err.Error(),
	})
	if err != nil {
		a.logger.Error("json encode of error response failed", "error", err)
	}
}

func (a *CreateTask) writeSuccess(taskID string, writer http.ResponseWriter) {
	err := json.NewEncoder(writer).Encode(map[string]any{
		"success": true,
		"result": map[string]any{
			"id": taskID,
		},
		"error": nil,
	})
	if err != nil {
		a.logger.Error("json encode of success response failed", "error", err)
	}
}
