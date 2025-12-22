package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"server/internal/appbuilder"
	"server/internal/utils/runkit"
)

func ResumeDelayed(app *appbuilder.App) {
	ctx, done := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer done()

	err := runkit.Retrier{
		Fn:     app.ResumeDelayed,
		Name:   "resume delayed",
		Logger: app.Logger,
	}.Run(ctx)

	if err != nil {
		os.Exit(1)
	}
}
