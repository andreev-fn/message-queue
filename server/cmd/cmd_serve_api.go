package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"server/internal/appbuilder"
	"server/internal/utils/runkit"
)

func ServeAPI(app *appbuilder.App) {
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
				Addr:    ApiAddr,
				Handler: app.Router,
			},
			Logger: app.Logger,
		},
		runkit.Retrier{
			Fn:     app.EventBus,
			Name:   "event bus",
			Logger: app.Logger,
		},
	}.Run(ctx)

	if err != nil {
		os.Exit(1)
	}
}
