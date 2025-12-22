package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"server/internal/appbuilder"
	"server/internal/utils/runkit"
)

func ArchiveMessages(app *appbuilder.App) {
	ctx, done := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer done()

	err := runkit.Retrier{
		Fn:     app.ArchiveMessages,
		Name:   "message archivation",
		Logger: app.Logger,
	}.Run(ctx)

	if err != nil {
		os.Exit(1)
	}
}
