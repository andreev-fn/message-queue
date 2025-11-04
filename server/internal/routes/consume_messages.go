package routes

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"server/internal/domain"
	"server/internal/routes/base"
	"server/internal/usecases"
	"server/pkg/httpmodels"
)

type ConsumeMessages struct {
	logger  *slog.Logger
	useCase *usecases.ConsumeMessages
}

func NewConsumeMessages(
	logger *slog.Logger,
	useCase *usecases.ConsumeMessages,
) *ConsumeMessages {
	return &ConsumeMessages{
		logger:  logger,
		useCase: useCase,
	}
}

func (a *ConsumeMessages) Mount(srv *http.ServeMux) {
	srv.Handle("/messages/consume", base.NewTypedHandler(a.logger, a.handler))
}

func (a *ConsumeMessages) handler(
	ctx context.Context,
	req httpmodels.ConsumeRequest,
) (httpmodels.ConsumeResponse, *base.Error) {
	queue, err := domain.NewQueueName(req.Queue)
	if err != nil {
		return nil, base.NewError(http.StatusBadRequest, err)
	}

	limit := 1
	if req.Limit != nil {
		limit = *req.Limit
	}

	poll := time.Duration(0)
	if req.Poll != nil {
		poll = time.Duration(*req.Poll) * time.Second
	}

	messages, err := a.useCase.Do(ctx, queue, limit, poll)
	if err != nil {
		return nil, base.ExtractKnownErrors(err)
	}

	resp := make([]httpmodels.ConsumeResponseItem, 0, len(messages))
	for _, msg := range messages {
		resp = append(resp, httpmodels.ConsumeResponseItem{
			ID:      msg.ID,
			Payload: msg.Payload,
		})
	}

	return resp, nil
}
