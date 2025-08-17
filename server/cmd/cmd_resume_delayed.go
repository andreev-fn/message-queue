package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"server/internal/appbuilder"
)

func ResumeDelayed(app *appbuilder.App) {
	const (
		limitTotal = 4000
		batchSize  = 250
	)

	ctx, done := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer done()

	totalAffected := 0

	for totalAffected < limitTotal {
		target := min(batchSize, limitTotal-totalAffected)

		affected, err := app.ResumeDelayed.Do(ctx, target)
		if err != nil {
			app.Logger.Error("retry processing failed", "error", err)
			if affected > 0 {
				app.Logger.Info("batch of messages was retried", "count", affected)
			}

			return
		}

		app.Logger.Info("batch of messages was retried", "count", affected)

		totalAffected += affected

		if affected < target {
			app.Logger.Info("all delayed messages retried")
			break
		}
	}

	app.Logger.Info("job done")
}
