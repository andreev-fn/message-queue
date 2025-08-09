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

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func() {
		if err := app.EventBus.Run(ctx); err != nil {
			app.Logger.Error("event bus stopped unexpectedly", "error", err)
		}
	}()

	<-ctx.Done()

	app.Logger.Info("shutting down server...")

	shutdownCtx, cancelShutdownCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdownCtx()

	if err := srv.Shutdown(shutdownCtx); err != nil {
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
