package routes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type CheckHealth struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewCheckHealth(
	db *sql.DB,
	logger *slog.Logger,
) *CheckHealth {
	return &CheckHealth{
		db:     db,
		logger: logger,
	}
}

func (a *CheckHealth) Mount(srv *http.ServeMux) {
	srv.HandleFunc("/health", a.handler)
}

func (a *CheckHealth) handler(writer http.ResponseWriter, request *http.Request) {
	if err := a.db.PingContext(request.Context()); err != nil {
		a.writeError(writer, http.StatusInternalServerError, fmt.Errorf("db.PingContext: %w", err))
		return
	}

	a.writeSuccess(writer)
}

func (a *CheckHealth) writeError(writer http.ResponseWriter, code int, err error) {
	a.logger.Error("health check failed", "error", err)

	writer.WriteHeader(code)

	err = json.NewEncoder(writer).Encode(map[string]any{
		"success": false,
	})
	if err != nil {
		a.logger.Error("json encode of error response failed", "error", err)
	}
}

func (a *CheckHealth) writeSuccess(writer http.ResponseWriter) {
	err := json.NewEncoder(writer).Encode(map[string]any{
		"success": true,
	})
	if err != nil {
		a.logger.Error("json encode of success response failed", "error", err)
	}
}
