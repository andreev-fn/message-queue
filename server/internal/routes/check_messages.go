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
)

type CheckMessagesDTO []string

func (d CheckMessagesDTO) Validate() error {
	if len(d) == 0 {
		return errors.New("at least one message must be specified")
	}

	for _, id := range d {
		if id == "" {
			return errors.New("id must not be empty string")
		}
	}

	return nil
}

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

	if request.Method != http.MethodPost {
		a.writeError(writer, http.StatusBadRequest, errors.New("method POST expected"))
		return
	}

	if request.Header.Get("Content-Type") != "application/json" {
		a.writeError(writer, http.StatusBadRequest, errors.New("json content type expected"))
		return
	}

	bodyBytes, err := io.ReadAll(request.Body)
	if err != nil {
		a.writeError(writer, http.StatusBadRequest, fmt.Errorf("io.ReadAll: %w", err))
		return
	}

	var requestDTO CheckMessagesDTO
	err = json.Unmarshal(bodyBytes, &requestDTO)
	if err != nil {
		a.writeError(writer, http.StatusBadRequest, fmt.Errorf("json.Unmarshal: %w (%s)", err, string(bodyBytes)))
		return
	}

	if err := requestDTO.Validate(); err != nil {
		a.writeError(writer, http.StatusBadRequest, fmt.Errorf("requestDTO.Validate: %w", err))
		return
	}

	result, err := a.useCase.Do(request.Context(), requestDTO)
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

func (a *CheckMessages) writeSuccess(writer http.ResponseWriter, result []usecases.CheckMsgResult) {
	messages := make([]any, 0, len(result))
	for _, msg := range result {
		history := make([]any, 0, len(msg.History))
		for _, chapter := range msg.History {
			history = append(history, map[string]any{
				"generation":    chapter.Generation,
				"queue":         chapter.Queue.String(),
				"redirected_at": chapter.RedirectedAt,
				"priority":      chapter.Priority,
				"retries":       chapter.Retries,
			})
		}

		messages = append(messages, map[string]any{
			"id":           msg.ID,
			"queue":        msg.Queue.String(),
			"created_at":   msg.CreatedAt,
			"finalized_at": msg.FinalizedAt,
			"status":       msg.Status,
			"priority":     msg.Priority,
			"retries":      msg.Retries,
			"generation":   msg.Generation,
			"payload":      msg.Payload,
			"history":      history,
		})
	}

	err := json.NewEncoder(writer).Encode(map[string]any{
		"success": true,
		"result":  messages,
		"error":   nil,
	})
	if err != nil {
		a.logger.Error("json encode of success response failed", "error", err)
	}
}
