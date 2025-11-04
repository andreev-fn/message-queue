package routes

import (
	"context"
	"log/slog"
	"net/http"

	"server/internal/routes/base"
	"server/internal/usecases"
	"server/pkg/httpmodels"
)

type ReleaseMessages struct {
	logger  *slog.Logger
	useCase *usecases.ReleaseMessages
}

func NewReleaseMessages(
	logger *slog.Logger,
	useCase *usecases.ReleaseMessages,
) *ReleaseMessages {
	return &ReleaseMessages{
		logger:  logger,
		useCase: useCase,
	}
}

func (a *ReleaseMessages) Mount(srv *http.ServeMux) {
	srv.Handle("/messages/release", base.NewTypedHandler(a.logger, a.handler))
}

func (a *ReleaseMessages) handler(
	ctx context.Context,
	req httpmodels.ReleaseRequest,
) (*httpmodels.OkResponse, *base.Error) {
	if err := a.useCase.Do(ctx, req); err != nil {
		return nil, base.ExtractKnownErrors(err)
	}

	return &httpmodels.OkResponse{Ok: true}, nil
}
