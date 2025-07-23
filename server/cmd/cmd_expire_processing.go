package main

import (
	"context"
	"os"
	"os/signal"
	"server/internal/appbuilder"
	"syscall"
)

func ExpireProcessing(app *appbuilder.App) {
	const (
		limitTotal = 10000
		batchSize  = 1000
	)

	ctx, done := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer done()

	totalAffected := 0

	for totalAffected < limitTotal {
		target := min(batchSize, limitTotal-totalAffected)

		affected, err := app.ExpireProcessing.Do(ctx, target)
		if err != nil {
			app.Logger.Error("expire processing failed", "error", err)

			return
		}

		app.Logger.Info("batch of tasks were expired", "count", affected)

		totalAffected += affected

		if affected < target {
			app.Logger.Info("all processing tasks expired")
			break
		}
	}

	app.Logger.Info("job done")
}
