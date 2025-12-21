package runkit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

type Retrier struct {
	Fn     Runnable
	Name   string
	Logger *slog.Logger
}

func (r Retrier) Run(ctx context.Context) error {
	const recoverDuration = 30 * time.Second
	const retryAttempts = 3

	attempt := 0

	for {
		startTime := time.Now()
		r.Logger.Info(fmt.Sprintf("starting %s process", r.Name))

		err := r.Fn.Run(ctx)

		if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			break
		}

		if time.Since(startTime) >= recoverDuration {
			attempt = 0
		}

		if attempt >= retryAttempts {
			r.Logger.Error(
				fmt.Sprintf("%s process failed", r.Name),
				"error", err,
			)
			return err
		}

		r.Logger.Error(
			fmt.Sprintf("%s process failed (will retry)", r.Name),
			"error", err,
		)

		retryDelay := time.Duration(attempt) * time.Second * 5
		attempt++

		select {
		case <-ctx.Done():
			break
		case <-time.After(retryDelay):
			continue
		}
	}

	r.Logger.Info(fmt.Sprintf("%s process exited", r.Name))
	return nil
}
