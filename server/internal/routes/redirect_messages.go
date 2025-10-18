package routes

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"server/internal/domain"
	"server/internal/usecases"
)

type RedirectMessagesDTO []struct {
	ID          string `json:"id"`
	Destination string `json:"destination"`
}

func (d RedirectMessagesDTO) Validate() error {
	if len(d) == 0 {
		return errors.New("at least one message must be specified")
	}

	for _, el := range d {
		if el.ID == "" {
			return errors.New("field 'id' must not be empty")
		}

		if el.Destination == "" {
			return errors.New("field 'destination' must not be empty")
		}
	}

	return nil
}

type RedirectMessages struct {
	db      *sql.DB
	logger  *slog.Logger
	useCase *usecases.RedirectMessages
}

func NewRedirectMessages(
	db *sql.DB,
	logger *slog.Logger,
	useCase *usecases.RedirectMessages,
) *RedirectMessages {
	return &RedirectMessages{
		db:      db,
		logger:  logger,
		useCase: useCase,
	}
}

func (a *RedirectMessages) Mount(srv *http.ServeMux) {
	srv.HandleFunc("/messages/redirect", a.handler)
}

func (a *RedirectMessages) handler(writer http.ResponseWriter, request *http.Request) {
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

	var requestDTO RedirectMessagesDTO
	err = json.Unmarshal(bodyBytes, &requestDTO)
	if err != nil {
		a.writeError(writer, http.StatusBadRequest, fmt.Errorf("json.Unmarshal: %w (%s)", err, string(bodyBytes)))
		return
	}

	if err := requestDTO.Validate(); err != nil {
		a.writeError(writer, http.StatusBadRequest, fmt.Errorf("requestDTO.Validate: %w", err))
		return
	}

	var redirectParams []usecases.RedirectParams
	for _, param := range requestDTO {
		destination, err := domain.NewQueueName(param.Destination)
		if err != nil {
			a.writeError(
				writer,
				http.StatusBadRequest,
				fmt.Errorf("domain.NewQueueName(%s): %w", param.Destination, err),
			)
			return
		}

		redirectParams = append(redirectParams, usecases.RedirectParams{
			ID:          param.ID,
			Destination: destination,
		})
	}

	if err := a.useCase.Do(request.Context(), redirectParams); err != nil {
		a.writeError(writer, http.StatusInternalServerError, fmt.Errorf("useCase.Do: %w", err))
		return
	}

	a.writeSuccess(writer)
}

func (a *RedirectMessages) writeError(writer http.ResponseWriter, code int, err error) {
	if code >= http.StatusInternalServerError {
		a.logger.Error("redirect messages use case failed", "error", err)
	}

	writer.WriteHeader(code)

	err = json.NewEncoder(writer).Encode(map[string]any{
		"error": err.Error(),
	})
	if err != nil {
		a.logger.Error("json encode of error response failed", "error", err)
	}
}

func (a *RedirectMessages) writeSuccess(writer http.ResponseWriter) {
	err := json.NewEncoder(writer).Encode(map[string]any{
		"ok": true,
	})
	if err != nil {
		a.logger.Error("json encode of success response failed", "error", err)
	}
}
