package routes

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"server/internal/usecases"
)

type PublishMessages struct {
	db      *sql.DB
	logger  *slog.Logger
	useCase *usecases.PublishMessages
}

func NewPublishMessages(
	db *sql.DB,
	logger *slog.Logger,
	useCase *usecases.PublishMessages,
) *PublishMessages {
	return &PublishMessages{
		db:      db,
		logger:  logger,
		useCase: useCase,
	}
}

func (a *PublishMessages) Mount(srv *http.ServeMux) {
	srv.HandleFunc("/messages/publish", func(w http.ResponseWriter, r *http.Request) {
		a.handler(w, r, true)
	})
	srv.HandleFunc("/messages/prepare", func(w http.ResponseWriter, r *http.Request) {
		a.handler(w, r, false)
	})
}

func (a *PublishMessages) handler(writer http.ResponseWriter, request *http.Request, autoConfirm bool) {
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

	var msgData struct {
		Queue    string          `json:"queue"`
		Payload  json.RawMessage `json:"payload"`
		Priority *int            `json:"priority"`
		StartAt  string          `json:"start_at"`
	}
	err = json.Unmarshal(bodyBytes, &msgData)
	if err != nil {
		a.writeError(writer, fmt.Errorf("json.Unmarshal: %w (%s)", err, string(bodyBytes)))
		return
	}

	var startAt *time.Time
	if msgData.StartAt != "" {
		t, err := time.Parse(time.RFC3339, msgData.StartAt)
		if err != nil {
			a.writeError(writer, errors.New("can`t parse start_at, expected format is RFC3339"))
			return
		}
		startAt = &t
	}

	priority := 100
	if msgData.Priority != nil {
		priority = *msgData.Priority
		if priority < 0 || priority > 255 {
			a.writeError(writer, errors.New("priority must be between 0 and 255"))
			return
		}
	}

	msgID, err := a.useCase.Do(
		request.Context(),
		msgData.Queue,
		msgData.Payload,
		priority,
		autoConfirm,
		startAt,
	)
	if err != nil {
		a.writeError(writer, fmt.Errorf("useCase.Do: %w", err))
		return
	}

	a.writeSuccess(msgID, writer)
}

func (a *PublishMessages) writeError(writer http.ResponseWriter, err error) {
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

func (a *PublishMessages) writeSuccess(msgID string, writer http.ResponseWriter) {
	err := json.NewEncoder(writer).Encode(map[string]any{
		"success": true,
		"result": map[string]any{
			"id": msgID,
		},
		"error": nil,
	})
	if err != nil {
		a.logger.Error("json encode of success response failed", "error", err)
	}
}
