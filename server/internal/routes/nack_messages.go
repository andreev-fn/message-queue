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

type NackMessagesDTO []struct {
	ID        string `json:"id"`
	Redeliver *bool  `json:"redeliver"`
}

func (d NackMessagesDTO) Validate() error {
	if len(d) == 0 {
		return errors.New("at least one message must be specified")
	}

	for _, el := range d {
		if el.ID == "" {
			return errors.New("field 'id' must not be empty")
		}
	}

	return nil
}

type NackMessages struct {
	db      *sql.DB
	logger  *slog.Logger
	useCase *usecases.NackMessages
}

func NewNackMessages(
	db *sql.DB,
	logger *slog.Logger,
	useCase *usecases.NackMessages,
) *NackMessages {
	return &NackMessages{
		db:      db,
		logger:  logger,
		useCase: useCase,
	}
}

func (a *NackMessages) Mount(srv *http.ServeMux) {
	srv.HandleFunc("/messages/nack", a.handler)
}

func (a *NackMessages) handler(writer http.ResponseWriter, request *http.Request) {
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

	var requestDTO NackMessagesDTO
	err = json.Unmarshal(bodyBytes, &requestDTO)
	if err != nil {
		a.writeError(writer, http.StatusBadRequest, fmt.Errorf("json.Unmarshal: %w (%s)", err, string(bodyBytes)))
		return
	}

	if err := requestDTO.Validate(); err != nil {
		a.writeError(writer, http.StatusBadRequest, fmt.Errorf("requestDTO.Validate: %w", err))
		return
	}

	var nackParams []usecases.NackParams
	for _, param := range requestDTO {
		redeliver := true
		if param.Redeliver != nil {
			redeliver = *param.Redeliver
		}
		nackParams = append(nackParams, usecases.NackParams{
			ID:        param.ID,
			Redeliver: redeliver,
		})
	}

	if err := a.useCase.Do(request.Context(), nackParams); err != nil {
		a.writeError(writer, http.StatusInternalServerError, fmt.Errorf("useCase.Do: %w", err))
		return
	}

	a.writeSuccess(writer)
}

func (a *NackMessages) writeError(writer http.ResponseWriter, code int, err error) {
	if code >= http.StatusInternalServerError {
		a.logger.Error("nack messages use case failed", "error", err)
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

func (a *NackMessages) writeSuccess(writer http.ResponseWriter) {
	err := json.NewEncoder(writer).Encode(map[string]any{
		"success": true,
		"error":   nil,
	})
	if err != nil {
		a.logger.Error("json encode of success response failed", "error", err)
	}
}
