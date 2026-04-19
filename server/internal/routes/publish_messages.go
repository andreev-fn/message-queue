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
) (*httpmodels.PublishResponse, *httpmodels.Error) {
	return a.handler(ctx, req, true)
}

func (a *PublishMessages) prepareHandler(
	ctx context.Context,
	req httpmodels.PublishRequest,
) (*httpmodels.PublishResponse, *httpmodels.Error) {
	return a.handler(ctx, req, false)
}

func (a *PublishMessages) handler(
	ctx context.Context,
	req httpmodels.PublishRequest,
	autoRelease bool,
) (*httpmodels.PublishResponse, *httpmodels.Error) {
	mappedItems, mapItemErrors := base.MapBatchRequestItems(req, a.mapRequestItem)

	results, err := a.useCase.Do(ctx, mappedItems, autoRelease)
	if err != nil {
		return nil, base.ExtractKnownErrors(err)
	}

	return &httpmodels.PublishResponse{
		Results: base.MapBatchResults(mapItemErrors, results, a.mapResult),
	}, nil
}

func (a *PublishMessages) mapRequestItem(
	params httpmodels.PublishRequestItem,
) (usecases.NewMessageParams, *httpmodels.Error) {
	priority := 100
	if params.Priority != nil {
		priority = *params.Priority
	}

	queue, err := domain.NewQueueName(params.Queue)
	if err != nil {
		return usecases.NewMessageParams{}, httpmodels.NewError(httpmodels.ErrorCodeRequestInvalid, err.Error())
	}

	return usecases.NewMessageParams{
		Queue:    queue,
		Payload:  params.Payload,
		Priority: priority,
		StartAt:  params.StartAt,
	}, nil
}

func (a *PublishMessages) mapResult(result *usecases.NewMessageResult) *httpmodels.PublishedMessage {
	return &httpmodels.PublishedMessage{
		ID: result.ID,
	}
}
