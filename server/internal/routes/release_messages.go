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

type ReleaseMessagesDTO []string

func (d ReleaseMessagesDTO) Validate() error {
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

type ReleaseMessages struct {
	db      *sql.DB
	logger  *slog.Logger
	useCase *usecases.ReleaseMessages
}

func NewReleaseMessages(
	db *sql.DB,
	logger *slog.Logger,
	useCase *usecases.ReleaseMessages,
) *ReleaseMessages {
	return &ReleaseMessages{
		db:      db,
		logger:  logger,
		useCase: useCase,
	}
}

func (a *ReleaseMessages) Mount(srv *http.ServeMux) {
	srv.HandleFunc("/messages/release", a.handler)
}

func (a *ReleaseMessages) handler(writer http.ResponseWriter, request *http.Request) {
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

	var requestDTO ReleaseMessagesDTO
	err = json.Unmarshal(bodyBytes, &requestDTO)
	if err != nil {
		a.writeError(writer, http.StatusBadRequest, fmt.Errorf("json.Unmarshal: %w (%s)", err, string(bodyBytes)))
		return
	}

	if err := requestDTO.Validate(); err != nil {
		a.writeError(writer, http.StatusBadRequest, fmt.Errorf("requestDTO.Validate: %w", err))
		return
	}

	if err := a.useCase.Do(request.Context(), requestDTO); err != nil {
		a.writeError(writer, http.StatusInternalServerError, err)
		return
	}

	a.writeSuccess(writer)
}

func (a *ReleaseMessages) writeError(writer http.ResponseWriter, code int, err error) {
	if code >= http.StatusInternalServerError {
		a.logger.Error("release messages use case failed", "error", err)
	}

	writer.WriteHeader(code)

	err = json.NewEncoder(writer).Encode(map[string]any{
		"error": err.Error(),
	})
	if err != nil {
		a.logger.Error("json encode of error response failed", "error", err)
	}
}

func (a *ReleaseMessages) writeSuccess(writer http.ResponseWriter) {
	err := json.NewEncoder(writer).Encode(map[string]any{
		"ok": true,
	})
	if err != nil {
		a.logger.Error("json encode of success response failed", "error", err)
	}
}
