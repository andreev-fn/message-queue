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

type AckMessagesDTO []struct {
	ID      string   `json:"id"`
	Release []string `json:"release"`
}

func (d AckMessagesDTO) Validate() error {
	for _, element := range d {
		if element.ID == "" {
			return errors.New("field 'id' must not be empty")
		}

		for _, id := range element.Release {
			if id == "" {
				return errors.New("every element inside 'release' must be non-empty string")
			}
		}
	}

	return nil
}

type AckMessages struct {
	db      *sql.DB
	logger  *slog.Logger
	useCase *usecases.AckMessages
}

func NewAckMessages(
	db *sql.DB,
	logger *slog.Logger,
	useCase *usecases.AckMessages,
) *AckMessages {
	return &AckMessages{
		db:      db,
		logger:  logger,
		useCase: useCase,
	}
}

func (a *AckMessages) Mount(srv *http.ServeMux) {
	srv.HandleFunc("/messages/ack", a.handler)
}

func (a *AckMessages) handler(writer http.ResponseWriter, request *http.Request) {
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

	var requestDTO AckMessagesDTO
	err = json.Unmarshal(bodyBytes, &requestDTO)
	if err != nil {
		a.writeError(writer, http.StatusBadRequest, fmt.Errorf("json.Unmarshal: %w (%s)", err, string(bodyBytes)))
		return
	}

	if err := requestDTO.Validate(); err != nil {
		a.writeError(writer, http.StatusBadRequest, fmt.Errorf("requestDTO.Validate: %w", err))
		return
	}

	var ackParams []usecases.AckParams
	for _, param := range requestDTO {
		ackParams = append(ackParams, usecases.AckParams{
			ID:      param.ID,
			Release: param.Release,
		})
	}

	if err := a.useCase.Do(request.Context(), ackParams); err != nil {
		a.writeError(writer, http.StatusInternalServerError, fmt.Errorf("useCase.Do: %w", err))
		return
	}

	a.writeSuccess(writer)
}

func (a *AckMessages) writeError(writer http.ResponseWriter, code int, err error) {
	if code >= http.StatusInternalServerError {
		a.logger.Error("ack messages use case failed", "error", err)
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

func (a *AckMessages) writeSuccess(writer http.ResponseWriter) {
	err := json.NewEncoder(writer).Encode(map[string]any{
		"success": true,
		"error":   nil,
	})
	if err != nil {
		a.logger.Error("json encode of success response failed", "error", err)
	}
}
