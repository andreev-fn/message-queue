package routes

import (
	"context"
	"log/slog"
	"net/http"

	"server/internal/domain"
	"server/internal/routes/base"
	"server/internal/usecases"
	"server/pkg/httpmodels"
)

type PublishMessages struct {
	logger  *slog.Logger
	useCase *usecases.PublishMessages
}

func NewPublishMessages(
	logger *slog.Logger,
	useCase *usecases.PublishMessages,
) *PublishMessages {
	return &PublishMessages{
		logger:  logger,
		useCase: useCase,
	}
}

func (a *PublishMessages) Mount(srv *http.ServeMux) {
	srv.Handle("/messages/publish", base.NewTypedHandler(a.logger, a.publishHandler))
	srv.Handle("/messages/prepare", base.NewTypedHandler(a.logger, a.prepareHandler))
}

func (a *PublishMessages) publishHandler(
	ctx context.Context,
	req httpmodels.PublishRequest,
) (httpmodels.PublishResponse, *base.Error) {
	return a.handler(ctx, req, true)
}

func (a *PublishMessages) prepareHandler(
	ctx context.Context,
	req httpmodels.PublishRequest,
) (httpmodels.PublishResponse, *base.Error) {
	return a.handler(ctx, req, false)
}

func (a *PublishMessages) handler(
	ctx context.Context,
	req httpmodels.PublishRequest,
	autoRelease bool,
) (httpmodels.PublishResponse, *base.Error) {
	var newMessages []usecases.NewMessageParams
	for _, param := range req {
		priority := 100
		if param.Priority != nil {
			priority = *param.Priority
		}

		queue, err := domain.NewQueueName(param.Queue)
		if err != nil {
			return nil, base.NewError(http.StatusBadRequest, err)
		}

		newMessages = append(newMessages, usecases.NewMessageParams{
			Queue:    queue,
			Payload:  param.Payload,
			Priority: priority,
			StartAt:  param.StartAt,
		})
	}

	msgIDs, err := a.useCase.Do(ctx, newMessages, autoRelease)
	if err != nil {
		return nil, base.ExtractKnownErrors(err)
	}

	return msgIDs, nil
}
