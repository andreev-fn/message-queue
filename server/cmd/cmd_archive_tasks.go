package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"server/internal/appbuilder"
)

func ArchiveMessages(app *appbuilder.App) {
	const (
		limitTotal = 10000
		batchSize  = 1000
	)

	ctx, done := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer done()

	totalAffected := 0

	for totalAffected < limitTotal {
		target := min(batchSize, limitTotal-totalAffected)

		affected, err := app.ArchiveMessages.Do(ctx, target)
		if err != nil {
			app.Logger.Error("message archivation failed", "error", err)

			return
		}

		app.Logger.Info("batch of messages were archived", "count", affected)

		totalAffected += affected

		if affected < target {
			app.Logger.Info("all finalized messages archived")
			break
		}
	}

	app.Logger.Info("job done")
}
