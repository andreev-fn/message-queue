package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"server/internal/appbuilder"
	"syscall"
	"time"
)

func Listen(app *appbuilder.App) {
	const ADDR = ":8060"

	if err := PingDB(app.DB); err != nil {
		app.Logger.Error("database connection failed", "error", err)
		return
	}

	srv := &http.Server{
		Addr:    ADDR,
		Handler: app.Router,
	}

	go func() {
		app.Logger.Info("start queue server...", "addr", ADDR)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			app.Logger.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit

	app.Logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		app.Logger.Error("shutting down failed", "error", err)
		return
	}

	app.Logger.Info("gracefully stopped")
}

func PingDB(db *sql.DB) error {
	pingCtx, closeCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer closeCtx()

	if err := db.PingContext(pingCtx); err != nil {
		return fmt.Errorf("db.PingContext: %w", err)
	}

	return nil
}
