package routes

import (
	"context"
	"log/slog"
	"net/http"

	"server/internal/routes/base"
	"server/internal/usecases"
	"server/pkg/httpmodels"
)

type NackMessages struct {
	logger  *slog.Logger
	useCase *usecases.NackMessages
}

func NewNackMessages(
	logger *slog.Logger,
	useCase *usecases.NackMessages,
) *NackMessages {
	return &NackMessages{
		logger:  logger,
		useCase: useCase,
	}
}

func (a *NackMessages) Mount(srv *http.ServeMux) {
	srv.Handle("/messages/nack", base.NewTypedHandler(a.logger, a.handler))
}

func (a *NackMessages) handler(
	ctx context.Context,
	req httpmodels.NackRequest,
) (*httpmodels.OkResponse, *base.Error) {
	var nackParams []usecases.NackParams

	for _, param := range req {
		redeliver := true
		if param.Redeliver != nil {
			redeliver = *param.Redeliver
		}

		nackParams = append(nackParams, usecases.NackParams{
			ID:        param.ID,
			Redeliver: redeliver,
		})
	}

	if err := a.useCase.Do(ctx, nackParams); err != nil {
		return nil, base.ExtractKnownErrors(err)
	}

	return &httpmodels.OkResponse{Ok: true}, nil
}
