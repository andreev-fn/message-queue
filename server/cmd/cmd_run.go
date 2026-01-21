package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"server/internal/appbuilder"
	"server/internal/utils/runkit"
)

func Run(app *appbuilder.App) {
	if err := PingDB(app.DB); err != nil {
		app.Logger.Error("database connection failed", "error", err)
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	err := runkit.Multiple{
		runkit.HTTPServer{
			Name: "API",
			Server: &http.Server{
				Addr:    fmt.Sprintf(":%d", app.Config.APIPort()),
				Handler: app.Router,
			},
			Logger: app.Logger,
		},
		runkit.Retrier{
			Fn:     app.EventBus,
			Name:   "event bus",
			Logger: app.Logger,
		},
		runkit.Retrier{
			Fn:     app.ResumeDelayed,
			Name:   "resume delayed",
			Logger: app.Logger,
		},
		runkit.Retrier{
			Fn:     app.ExpireProcessing,
			Name:   "expire processing",
			Logger: app.Logger,
		},
		runkit.Retrier{
			Fn:     app.ArchiveMessages,
			Name:   "message archivation",
			Logger: app.Logger,
		},
	}.Run(ctx)

	if err != nil {
		os.Exit(1)
	}
}
