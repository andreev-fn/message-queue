package routes

import (
	"context"
	"log/slog"
	"net/http"

	"server/internal/routes/base"
	"server/internal/usecases"
	"server/pkg/httpmodels"
)

type AckMessages struct {
	logger  *slog.Logger
	useCase *usecases.AckMessages
}

func NewAckMessages(
	logger *slog.Logger,
	useCase *usecases.AckMessages,
) *AckMessages {
	return &AckMessages{
		logger:  logger,
		useCase: useCase,
	}
}

func (a *AckMessages) Mount(srv *http.ServeMux) {
	srv.Handle("/messages/ack", base.NewTypedHandler(a.logger, a.handler))
}

func (a *AckMessages) handler(
	ctx context.Context,
	req httpmodels.AckRequest,
) (*httpmodels.OkResponse, *base.Error) {
	var ackParams []usecases.AckParams
	for _, param := range req {
		ackParams = append(ackParams, usecases.AckParams{
			ID:      param.ID,
			Release: param.Release,
		})
	}

	if err := a.useCase.Do(ctx, ackParams); err != nil {
		return nil, base.NewError(http.StatusInternalServerError, err)
	}

	return &httpmodels.OkResponse{Ok: true}, nil
}
