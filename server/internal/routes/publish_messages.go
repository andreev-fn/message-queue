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

	"server/internal/domain"
	"server/internal/usecases"
)

type PublishMessagesDTO []struct {
	Queue    string `json:"queue"`
	Payload  string `json:"payload"`
	Priority *int   `json:"priority"`
	StartAt  string `json:"start_at"`

	parsedStartAt *time.Time
}

func (d PublishMessagesDTO) Validate() error {
	if len(d) == 0 {
		return errors.New("at least one message must be specified")
	}

	for _, el := range d {
		if el.Queue == "" {
			return errors.New("parameter 'queue' must be non-empty string")
		}

		if el.Priority != nil && (*el.Priority < 0 || *el.Priority > 255) {
			return errors.New("priority must be between 0 and 255")
		}

		if el.StartAt != "" {
			t, err := time.Parse(time.RFC3339, el.StartAt)
			if err != nil {
				return errors.New("can`t parse start_at, expected format is RFC3339")
			}
			el.parsedStartAt = &t
		}
	}

	return nil
}

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

func (a *PublishMessages) handler(writer http.ResponseWriter, request *http.Request, autoRelease bool) {
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

	var requestDTO PublishMessagesDTO
	err = json.Unmarshal(bodyBytes, &requestDTO)
	if err != nil {
		a.writeError(writer, http.StatusBadRequest, fmt.Errorf("json.Unmarshal: %w (%s)", err, string(bodyBytes)))
		return
	}

	if err := requestDTO.Validate(); err != nil {
		a.writeError(writer, http.StatusBadRequest, fmt.Errorf("requestDTO.Validate: %w", err))
		return
	}

	var newMessages []usecases.NewMessageParams
	for _, param := range requestDTO {
		priority := 100
		if param.Priority != nil {
			priority = *param.Priority
		}

		queue, err := domain.NewQueueName(param.Queue)
		if err != nil {
			a.writeError(writer, http.StatusBadRequest, fmt.Errorf("domain.NewQueueName: %w", err))
			return
		}

		newMessages = append(newMessages, usecases.NewMessageParams{
			Queue:    queue,
			Payload:  param.Payload,
			Priority: priority,
			StartAt:  param.parsedStartAt,
		})
	}

	msgIDs, err := a.useCase.Do(request.Context(), newMessages, autoRelease)
	if err != nil {
		a.writeError(writer, http.StatusInternalServerError, fmt.Errorf("useCase.Do: %w", err))
		return
	}

	a.writeSuccess(writer, msgIDs)
}

func (a *PublishMessages) writeError(writer http.ResponseWriter, code int, err error) {
	if code >= http.StatusInternalServerError {
		a.logger.Error("publish messages use case failed", "error", err)
	}

	writer.WriteHeader(code)

	err = json.NewEncoder(writer).Encode(map[string]any{
		"error": err.Error(),
	})
	if err != nil {
		a.logger.Error("json encode of error response failed", "error", err)
	}
}

func (a *PublishMessages) writeSuccess(writer http.ResponseWriter, msgIDs []string) {
	err := json.NewEncoder(writer).Encode(msgIDs)
	if err != nil {
		a.logger.Error("json encode of success response failed", "error", err)
	}
}
