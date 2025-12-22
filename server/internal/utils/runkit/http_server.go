package runkit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type HTTPServer struct {
	Name   string
	Server *http.Server
	Logger *slog.Logger
}

func (r HTTPServer) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		r.Logger.Info(fmt.Sprintf("starting %s server", r.Name), "addr", r.Server.Addr)
		if err := r.Server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		r.Logger.Error(fmt.Sprintf("%s server failed", r.Name), "error", err)
		return err
	case <-ctx.Done():
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		r.Logger.Info(fmt.Sprintf("shutting down %s server", r.Name))
		err := r.Server.Shutdown(timeoutCtx)
		if err != nil {
			r.Logger.Error(fmt.Sprintf("shutting down %s server failed", r.Name), "error", err)
			return err
		}

		r.Logger.Info(fmt.Sprintf("%s server gracefully stopped", r.Name))
		return nil
	}
}
