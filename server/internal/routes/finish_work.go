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

type ErrorDTO struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Info    map[string]any `json:"additional_info"`
}

type SaveResultDTO struct {
	ID     string    `json:"id"`
	Report *string   `json:"report"`
	Error  *ErrorDTO `json:"error"`
}

func (d SaveResultDTO) Validate() error {
	if d.ID == "" {
		return errors.New("field 'id' is required")
	}
	if (d.Report != nil && d.Error != nil) || (d.Report == nil && d.Error == nil) {
		return errors.New("exactly one of 'error' or 'report' fields must be present")
	}
	if d.Error != nil && d.Error.Code == "" {
		return errors.New("error code is required")
	}
	return nil
}

type FinishWork struct {
	db      *sql.DB
	logger  *slog.Logger
	useCase *usecases.FinishWork
}

func NewFinishWork(
	db *sql.DB,
	logger *slog.Logger,
	useCase *usecases.FinishWork,
) *FinishWork {
	return &FinishWork{
		db:      db,
		logger:  logger,
		useCase: useCase,
	}
}

func (a *FinishWork) Mount(srv *http.ServeMux) {
	srv.HandleFunc("/work/finish", a.handler)
}

func (a *FinishWork) handler(writer http.ResponseWriter, request *http.Request) {
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

	var requestDTO SaveResultDTO
	err = json.Unmarshal(bodyBytes, &requestDTO)
	if err != nil {
		a.writeError(writer, http.StatusBadRequest, fmt.Errorf("json.Unmarshal: %w (%s)", err, string(bodyBytes)))
		return
	}

	if err := requestDTO.Validate(); err != nil {
		a.writeError(writer, http.StatusBadRequest, fmt.Errorf("requestDTO.Validate: %w", err))
		return
	}

	var errorCode *string
	if requestDTO.Error != nil {
		errorCode = &requestDTO.Error.Code
	}

	var report json.RawMessage
	if requestDTO.Report != nil {
		if err := json.Unmarshal([]byte(*requestDTO.Report), &report); err != nil {
			a.writeError(writer, http.StatusBadRequest, fmt.Errorf("json.Unmarshal: %w", err))
			return
		}
	}

	if err := a.useCase.Do(request.Context(), requestDTO.ID, report, errorCode); err != nil {
		a.writeError(writer, http.StatusInternalServerError, fmt.Errorf("useCase.Do: %w", err))
		return
	}

	a.writeSuccess(writer)
}

func (a *FinishWork) writeError(writer http.ResponseWriter, code int, err error) {
	if code >= http.StatusInternalServerError {
		a.logger.Error("save message result use case failed", "error", err)
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

func (a *FinishWork) writeSuccess(writer http.ResponseWriter) {
	err := json.NewEncoder(writer).Encode(map[string]any{
		"success": true,
		"error":   nil,
	})
	if err != nil {
		a.logger.Error("json encode of success response failed", "error", err)
	}
}
