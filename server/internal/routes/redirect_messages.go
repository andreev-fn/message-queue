package routes

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"server/internal/domain"
	"server/internal/routes/base"
	"server/internal/usecases"
	"server/pkg/httpmodels"
)

type RedirectMessages struct {
	logger  *slog.Logger
	useCase *usecases.RedirectMessages
}

func NewRedirectMessages(
	logger *slog.Logger,
	useCase *usecases.RedirectMessages,
) *RedirectMessages {
	return &RedirectMessages{
		logger:  logger,
		useCase: useCase,
	}
}

func (a *RedirectMessages) Mount(srv *http.ServeMux) {
	srv.Handle("/messages/redirect", base.NewTypedHandler(a.logger, a.handler))
}

func (a *RedirectMessages) handler(
	ctx context.Context,
	req httpmodels.RedirectRequest,
) (*httpmodels.OkResponse, *base.Error) {
	var redirectParams []usecases.RedirectParams

	for _, param := range req {
		destination, err := domain.NewQueueName(param.Destination)
		if err != nil {
			return nil, base.NewError(
				http.StatusBadRequest,
				fmt.Errorf("domain.NewQueueName(%s): %w", param.Destination, err),
			)
		}

		redirectParams = append(redirectParams, usecases.RedirectParams{
			ID:          param.ID,
			Destination: destination,
		})
	}

	if err := a.useCase.Do(ctx, redirectParams); err != nil {
		return nil, base.ExtractKnownErrors(err)
	}

	return &httpmodels.OkResponse{Ok: true}, nil
}
