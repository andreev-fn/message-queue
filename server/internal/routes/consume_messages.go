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

type ConsumeMessagesDTO struct {
	Queue string `json:"queue"`
	Limit *int   `json:"limit"`
	Poll  *int   `json:"poll"`
}

func (d ConsumeMessagesDTO) Validate() error {
	if d.Queue == "" {
		return errors.New("parameter 'queue' required")
	}

	if d.Limit != nil && *d.Limit < 1 {
		return errors.New("parameter 'limit' must be greater than 0")
	}

	if d.Poll != nil && *d.Poll < 0 {
		return errors.New("parameter 'poll' must be >= 0")
	}

	return nil
}

type ConsumeMessages struct {
	db      *sql.DB
	logger  *slog.Logger
	useCase *usecases.ConsumeMessages
}

func NewConsumeMessages(
	db *sql.DB,
	logger *slog.Logger,
	useCase *usecases.ConsumeMessages,
) *ConsumeMessages {
	return &ConsumeMessages{
		db:      db,
		logger:  logger,
		useCase: useCase,
	}
}

func (a *ConsumeMessages) Mount(srv *http.ServeMux) {
	srv.HandleFunc("/messages/consume", a.handler)
}

func (a *ConsumeMessages) handler(writer http.ResponseWriter, request *http.Request) {
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

	var requestDTO ConsumeMessagesDTO
	err = json.Unmarshal(bodyBytes, &requestDTO)
	if err != nil {
		a.writeError(writer, http.StatusBadRequest, fmt.Errorf("json.Unmarshal: %w (%s)", err, string(bodyBytes)))
		return
	}

	if err := requestDTO.Validate(); err != nil {
		a.writeError(writer, http.StatusBadRequest, fmt.Errorf("requestDTO.Validate: %w", err))
		return
	}

	queue, err := domain.NewQueueName(requestDTO.Queue)
	if err != nil {
		a.writeError(writer, http.StatusBadRequest, fmt.Errorf("domain.NewQueueName: %w", err))
		return
	}

	limit := 1
	if requestDTO.Limit != nil {
		limit = *requestDTO.Limit
	}

	poll := time.Duration(0)
	if requestDTO.Poll != nil {
		poll = time.Duration(*requestDTO.Poll) * time.Second
	}

	messages, err := a.useCase.Do(request.Context(), queue, limit, poll)
	if err != nil {
		a.writeError(writer, http.StatusInternalServerError, fmt.Errorf("useCase.Do: %w", err))
		return
	}

	a.writeSuccess(writer, messages)
}

func (a *ConsumeMessages) writeError(writer http.ResponseWriter, code int, err error) {
	if code >= http.StatusInternalServerError {
		a.logger.Error("consume messages use case failed", "error", err)
	}

	writer.WriteHeader(code)

	err = json.NewEncoder(writer).Encode(map[string]any{
		"error": err.Error(),
	})
	if err != nil {
		a.logger.Error("json encode of error response failed", "error", err)
	}
}

func (a *ConsumeMessages) writeSuccess(writer http.ResponseWriter, messages []usecases.MessageToConsume) {
	result := make([]any, 0, len(messages))

	for _, message := range messages {
		result = append(result, map[string]any{
			"id":      message.ID,
			"payload": message.Payload,
		})
	}

	err := json.NewEncoder(writer).Encode(result)
	if err != nil {
		a.logger.Error("json encode of success response failed", "error", err)
	}
}
